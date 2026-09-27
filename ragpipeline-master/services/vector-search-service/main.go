// Package main provides the HTTP server entry point for the vector search service.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/visionary/ragpipeline/services/vector-search-service/cache"
	"github.com/visionary/ragpipeline/services/vector-search-service/config"
	"github.com/visionary/ragpipeline/services/vector-search-service/db"
	"github.com/visionary/ragpipeline/services/vector-search-service/handler"
	"github.com/visionary/ragpipeline/services/vector-search-service/middleware"
	"github.com/visionary/ragpipeline/services/vector-search-service/search"

	"github.com/visionary/ragpipeline/pkg/contextkeys"
)

// Version information (set via ldflags).
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

func main() {
	// Initialize structured logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs

	log.Info().Msg("Starting Vector Search Service")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize database connection
	ctx := context.Background()
	store, err := db.NewStore(ctx, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize database")
	}
	defer store.Close()

	// Initialize RRF configuration
	rrfCfg := search.RRFConfig{
		K:            cfg.RRFK,
		DenseWeight:  cfg.RRFDenseWeight,
		SparseWeight: cfg.RRFSparseWeight,
	}

	// Initialize searchers
	hybridSearcher := search.NewHybridSearcher(store, cfg, &rrfCfg)

	// Initialize Redis client and search cache
	var searchCache *cache.SearchCache
	redisURL := getEnv("REDIS_URL", "redis://localhost:6379/0")
	searchCacheEnabled := getEnvBool("SEARCH_CACHE_ENABLED", true)

	if searchCacheEnabled {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to parse Redis URL, search cache disabled")
			searchCacheEnabled = false
		} else {
			redisClient := redis.NewClient(opt)
			if err := redisClient.Ping(ctx).Err(); err != nil {
				log.Warn().Err(err).Msg("Failed to connect to Redis, search cache disabled")
				searchCacheEnabled = false
			} else {
				searchCacheTTL := getDurationEnv("SEARCH_CACHE_TTL", 30*time.Minute)
				searchCacheMaxSize := getIntEnv("SEARCH_CACHE_MAX_SIZE", 3000)
				searchCache = cache.NewSearchCache(redisClient,
					cache.WithCacheTTL(searchCacheTTL),
					cache.WithCacheMaxSize(searchCacheMaxSize),
				)
				log.Info().
					Str("redis_url", redisURL).
					Dur("ttl", searchCacheTTL).
					Int("max_size", searchCacheMaxSize).
					Msg("Search results cache initialized")
			}
		}
	} else {
		log.Info().Msg("Search results cache disabled")
	}

	// Create handler
	searchHandler := handler.NewSearchHandler(handler.SearchHandlerDeps{
		HybridSearcher: hybridSearcher,
		Config:         cfg,
		Store:          store,
		SearchCache:    searchCache,
	})

	// Set version in handler
	handler.Version = Version

	// Create router
	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/search", searchHandler.Search).Methods("POST", "OPTIONS")
	router.HandleFunc("/search/dense", searchHandler.DenseSearch).Methods("POST", "OPTIONS")
	router.HandleFunc("/search/sparse", searchHandler.SparseSearch).Methods("POST", "OPTIONS")
	router.HandleFunc("/health", searchHandler.Health).Methods("GET")
	router.HandleFunc("/stats", searchHandler.Stats).Methods("GET")

	// Version endpoint
	router.HandleFunc("/version", versionHandler).Methods("GET")

	// Prometheus metrics endpoint
	router.Handle("/metrics", middleware.MetricsHandler()).Methods("GET")

	// Fallback routes
	router.NotFoundHandler = http.HandlerFunc(searchHandler.NotFound)
	router.MethodNotAllowedHandler = http.HandlerFunc(searchHandler.MethodNotAllowed)

	// Apply middleware chain
	finalHandler := recoveryMiddleware(requestIDMiddleware(loggingMiddleware(middleware.PrometheusMiddleware(router))))

	// Create HTTP server
	addr := ":" + cfg.Port
	server := &http.Server{
		Addr:         addr,
		Handler:      finalHandler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Info().Str("addr", addr).Msg("HTTP server starting")

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	log.Info().
		Str("version", Version).
		Str("port", cfg.Port).
		Int("embedding_dim", cfg.EmbeddingDim).
		Float64("rrf_k", cfg.RRFK).
		Int("default_top_k", cfg.DefaultTopK).
		Str("query_timeout", cfg.QueryTimeout.String()).
		Msg("Vector search service ready")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down vector search service...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server shutdown failed")
	}

	log.Info().Msg("Vector search service stopped")
}

// versionHandler returns version information.
// P2-22: Uses json.NewEncoder instead of fmt.Fprintf for JSON safety.
func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"version":    Version,
		"git_commit": GitCommit,
		"build_time": BuildTime,
	})
}

// requestIDMiddleware injects a unique request ID into the request context and response header.
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		w.Header().Set("X-Request-ID", requestID)
		ctx := contextkeys.WithRequestID(r.Context(), requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// loggingMiddleware logs request details and latency.
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		requestID, _ := contextkeys.RequestIDFromContext(r.Context())
		if requestID == "" {
			requestID = "unknown"
		}

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("request_id", requestID).
			Str("remote_addr", r.RemoteAddr).
			Str("user_agent", r.UserAgent()).
			Msg("Request started")

		next.ServeHTTP(w, r)

		log.Info().
			Str("method", r.Method).
			Str("path", r.URL.Path).
			Str("request_id", requestID).
			Int64("latency_ms", time.Since(start).Milliseconds()).
			Msg("Request completed")
	})
}

// recoveryMiddleware recovers from panics and returns 500.
func recoveryMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				requestID, _ := contextkeys.RequestIDFromContext(r.Context())
				if requestID == "" {
					requestID = "unknown"
				}

				log.Error().
					Interface("panic", rec).
					Str("request_id", requestID).
					Str("path", r.URL.Path).
					Msg("Recovered from panic")

				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Request-ID", requestID)
				w.WriteHeader(http.StatusInternalServerError)

				fmt.Fprintf(w, `{"error":{"code":"internal_error","message":"An internal error occurred"}}`)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

// getEnv returns the value of an environment variable or a default.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDurationEnv parses a duration from environment or returns default.
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
		log.Warn().Str("key", key).Str("value", value).Msg("Invalid duration, using default")
	}
	return defaultValue
}

// getIntEnv parses an int from environment or returns default.
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		v, err := strconv.Atoi(strings.TrimSpace(value))
		if err == nil {
			return v
		}
		log.Warn().Str("key", key).Str("value", value).Msg("Invalid int, using default")
	}
	return defaultValue
}

// getEnvBool parses a bool from environment or returns default.
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		default:
			log.Warn().Str("key", key).Str("value", value).Msg("Invalid bool, using default")
			return defaultValue
		}
	}
	return defaultValue
}
