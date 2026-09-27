// Package main provides the HTTP server entry point for the embedding service.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/visionary/ragpipeline/services/embedding-service/cache"
	"github.com/visionary/ragpipeline/services/embedding-service/config"
	"github.com/visionary/ragpipeline/services/embedding-service/handler"
	"github.com/visionary/ragpipeline/services/embedding-service/middleware"
	"github.com/visionary/ragpipeline/services/embedding-service/vertex"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"

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

	log.Info().Msg("Starting Embedding Service")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// Initialize OpenTelemetry tracer
	tracerShutdown := initTracer(context.Background(), cfg)
	defer tracerShutdown()

	// Initialize Vertex AI client
	ctx := context.Background()
	vertexClient, err := vertex.NewClient(
		ctx,
		cfg.VertexProject,
		cfg.VertexLocation,
		cfg.VertexModel,
		cfg.VertexDimension,
		cfg.VertexTimeout,
	)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Vertex AI client")
	}
	defer vertexClient.Close()

	// Initialize Redis client and embedding cache
	var embeddingCache *cache.EmbeddingCache
	if cfg.EmbeddingCacheEnabled {
		opt, err := redis.ParseURL(cfg.RedisURL)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to parse Redis URL, embedding cache disabled")
			cfg.EmbeddingCacheEnabled = false
		} else {
			redisClient := redis.NewClient(opt)
			// FIX P0: Use background context with timeout for startup health check
			pingCtx, pingCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer pingCancel()
			if err := redisClient.Ping(pingCtx).Err(); err != nil {
				log.Warn().Err(err).Msg("Failed to connect to Redis, embedding cache disabled")
				cfg.EmbeddingCacheEnabled = false
			} else {
				embeddingCache = cache.NewEmbeddingCache(redisClient,
					cache.WithCacheTTL(cfg.EmbeddingCacheTTL),
					cache.WithCacheMaxSize(cfg.EmbeddingCacheMaxSize),
				)
				log.Info().
					Str("redis_url", cfg.RedisURL).
					Dur("ttl", cfg.EmbeddingCacheTTL).
					Int("max_size", cfg.EmbeddingCacheMaxSize).
					Msg("Embedding cache initialized")
			}
		}
	} else {
		log.Info().Msg("Embedding cache disabled")
	}

	// Create handler
	embHandler := handler.NewEmbeddingHandler(cfg, vertexClient, embeddingCache)

	// Create router
	router := mux.NewRouter()

	// API routes
	router.HandleFunc("/embed", embHandler.Embed).Methods("POST", "OPTIONS")
	router.HandleFunc("/health", embHandler.Health).Methods("GET")
	router.HandleFunc("/stats", embHandler.Stats).Methods("GET")

	// Version endpoint
	router.HandleFunc("/version", versionHandler).Methods("GET")

	// Prometheus metrics endpoint
	router.Handle("/metrics", middleware.MetricsHandler()).Methods("GET")

	// Fallback routes
	router.NotFoundHandler = http.HandlerFunc(embHandler.NotFound)
	router.MethodNotAllowedHandler = http.HandlerFunc(embHandler.MethodNotAllowed)

	// Apply middleware chain
	finalHandler := recoveryMiddleware(requestIDMiddleware(loggingMiddleware(corsMiddleware(middleware.PrometheusMiddleware(router)))))

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
		Str("model", cfg.VertexModel).
		Msg("Embedding service ready")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down embedding service...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("Server shutdown failed")
	}

	log.Info().Msg("Embedding service stopped")
}

// initTracer initializes OpenTelemetry tracer.
func initTracer(ctx context.Context, cfg *config.Config) func() {
	if cfg.OTelEndpoint == "" {
		log.Info().Msg("OpenTelemetry endpoint not configured, skipping tracer init")
		return func() {}
	}

	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(cfg.OTelEndpoint),
		// FIX P1: Use TLS for trace data; remove WithInsecure()
	)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create OTLP exporter, continuing without tracing")
		return func() {}
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(cfg.OTelServiceName),
			semconv.ServiceVersion(Version),
		),
	)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to create OT resource")
		return func() {}
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(cfg.OTelSampleRate),
		)),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("Failed to shutdown tracer")
		}
	}

	log.Info().
		Str("endpoint", cfg.OTelEndpoint).
		Float64("sample_rate", cfg.OTelSampleRate).
		Msg("OpenTelemetry tracer initialized")

	return shutdown
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
		ctx := context.WithValue(r.Context(), "request_id", requestID)
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

// corsMiddleware adds CORS headers to responses.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
		w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
		w.Header().Set("Access-Control-Max-Age", "86400")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// GetTracerFromContext extracts the request ID from context for use in tracing.
func GetTracerFromContext(ctx context.Context) string {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		return requestID
	}
	return "unknown"
}

// AddRequestIDAttribute adds the request ID to an OpenTelemetry span.
func AddRequestIDAttribute(ctx context.Context, span trace.Span) {
	if requestID, ok := ctx.Value("request_id").(string); ok {
		span.SetAttributes(attribute.String("request_id", requestID))
	}
}
