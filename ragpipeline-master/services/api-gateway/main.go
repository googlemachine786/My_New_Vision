package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/visionary/ragpipeline/services/api-gateway/cache"
	"github.com/visionary/ragpipeline/services/api-gateway/client"
	"github.com/visionary/ragpipeline/services/api-gateway/config"
	sessionctx "github.com/visionary/ragpipeline/services/api-gateway/context"
	"github.com/visionary/ragpipeline/services/api-gateway/handler"
	"github.com/visionary/ragpipeline/services/api-gateway/middleware"
	"github.com/visionary/ragpipeline/services/api-gateway/session"

	"github.com/gorilla/mux"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Initialize structured logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if os.Getenv("SERVER_ENV") == "production" {
		log.Logger = log.Output(os.Stdout)
	} else {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	}

	log.Info().Msg("Starting API Gateway...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	log.Info().
		Int("port", cfg.Server.Port).
		Str("environment", cfg.Server.Environment).
		Msg("Configuration loaded")

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: 2,
	})

	// Test Redis connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Redis")
	}
	cancel()

	log.Info().Str("addr", cfg.Redis.Addr).Msg("Connected to Redis")

	// Initialize service clients
	embeddingClient := client.NewServiceClient(
		"embedding_service",
		cfg.Services.EmbeddingServiceURL,
		cfg.Timeouts.EmbeddingService,
	)

	vectorSearchClient := client.NewServiceClient(
		"vector_search",
		cfg.Services.VectorSearchServiceURL,
		cfg.Timeouts.VectorSearchService,
	)

	queryUnderstandingClient := client.NewServiceClient(
		"query_understanding",
		cfg.Services.QueryUnderstandingURL,
		cfg.Timeouts.QueryUnderstanding,
	)

	llmClient := client.NewServiceClient(
		"llm_service",
		cfg.Services.LLMServiceURL,
		cfg.Timeouts.LLMService,
	)

	// Initialize cache
	responseCache := cache.NewResponseCache(
		redisClient,
		cache.WithCacheTTL(time.Hour),
		cache.WithCacheMaxSize(1000),
	)

	// Initialize CAG caches
	semanticCache := cache.NewSemanticCache(
		redisClient,
		cache.WithCacheTTL(6*time.Hour),
		cache.WithCacheMaxSize(5000),
		cache.WithSimilarityThreshold(0.95),
	)

	templateCache := cache.NewTemplateCache(
		redisClient,
		cache.WithCacheTTL(12*time.Hour),
		cache.WithCacheMaxSize(500),
	)

	// CAG multi-layer cache initialized (embedding cache managed by embedding-service)

	// Initialize CAG orchestrator
	cagOrchestrator := cache.NewCAGOrchestrator(
		responseCache,
		semanticCache,
		templateCache,
	)

	log.Info().
		Bool("exact_cache_enabled", true).
		Bool("semantic_cache_enabled", true).
		Bool("template_cache_enabled", true).
		Msg("CAG multi-layer cache initialized")

	// Initialize session manager
	sessionManager := session.NewSessionManager(
		redisClient,
		30*time.Minute, // Default TTL
		20,             // Default max turns
	)

	// Initialize middleware
	recoveryMiddleware := middleware.NewRecoveryMiddleware()
	requestIDMiddleware := middleware.NewRequestIDMiddleware()
	loggingMiddleware := middleware.NewLoggingMiddleware()
	corsMiddleware := middleware.NewCORSMiddleware()
	authMiddleware := middleware.NewAuthMiddleware(cfg.Auth.JWTSecret)
	rateLimiterMiddleware := middleware.NewRateLimiterMiddleware(
		redisClient,
		middleware.WithRateLimit(cfg.RateLimit.RequestsPerMinute),
		middleware.WithBurstSize(cfg.RateLimit.BurstSize),
	)

	// Story C: Initialize fallback handler BEFORE QueryHandler (dependency)
	fallbackHandler := handler.NewFallbackHandler(llmClient)

	// Story D: Initialize re-explain handler
	reexplainHandler := handler.NewReexplainHandler(llmClient, redisClient)

	// Story B: Initialize quality pipeline
	qualityPipeline := handler.NewQualityPipeline(redisClient)

	// Story G: Initialize session context registry (before query handler)
	contextRegistry := sessionctx.NewSessionRegistry(redisClient)

	// Initialize handlers
	healthHandler := handler.NewHealthHandler(
		embeddingClient,
		vectorSearchClient,
		queryUnderstandingClient,
		func(ctx context.Context) error {
			// FIX P0: Use background context with timeout for health checks
			hctx, hcancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer hcancel()
			return redisClient.Ping(hctx).Err()
		},
	)

	queryHandler := handler.NewQueryHandler(handler.QueryHandlerDeps{
		EmbeddingClient:          embeddingClient,
		VectorSearchClient:       vectorSearchClient,
		QueryUnderstandingClient: queryUnderstandingClient,
		LLMClient:                llmClient,
		Cache:                    responseCache,
		SessionManager:           sessionManager,
		CAGOrchestrator:          cagOrchestrator,
		FallbackHandler:          fallbackHandler, // FIX P0: Now properly initialized
		ContextRegistry:          contextRegistry,
		NLIClient:                handler.NewNLIClient(""),
	})

	statsHandler := handler.NewStatsHandler(
		embeddingClient,
		vectorSearchClient,
		queryUnderstandingClient,
		llmClient,
		responseCache,
		rateLimiterMiddleware,
	)

	cagStatsHandler := handler.NewCAGStatsHandler(cagOrchestrator)

	feedbackHandler := handler.NewFeedbackHandler(redisClient)

	// Initialize clarification handler
	clarificationHandler := handler.NewClarificationHandler()

	// Initialize feedback aggregator
	feedbackAggregator := handler.NewFeedbackAggregator(redisClient)

	// Story G: Initialize context handler
	contextHandler := handler.NewContextHandler(contextRegistry)

	// Initialize WebSocket handler
	wsHandler := handler.NewWebSocketHandler(handler.WebSocketHandlerDeps{
		EmbeddingClient:          embeddingClient,
		VectorSearchClient:       vectorSearchClient,
		QueryUnderstandingClient: queryUnderstandingClient,
		LLMClient:                llmClient,
		CAGOrchestrator:          cagOrchestrator,
		SessionManager:           sessionManager,
		RedisClient:              redisClient,
	})

	// Setup router
	router := mux.NewRouter()

	// Public routes (no auth required)
	router.HandleFunc("/health", healthHandler.ServeHTTP).Methods("GET")
	router.HandleFunc("/version", versionHandler).Methods("GET")
	router.Handle("/metrics", middleware.MetricsHandler()).Methods("GET")

	// API routes (auth required)
	api := router.PathPrefix("/").Subrouter()
	api.HandleFunc("/query", queryHandler.ServeHTTP).Methods("POST")
	api.HandleFunc("/ws/chat", wsHandler.ServeHTTP).Methods("GET")
	api.HandleFunc("/stats", statsHandler.ServeHTTP).Methods("GET")
	api.HandleFunc("/cag/stats", cagStatsHandler.ServeHTTP).Methods("GET")
	api.HandleFunc("/feedback", feedbackHandler.ServeHTTP).Methods("POST")
	api.HandleFunc("/feedback/stats", feedbackAggregator.ServeHTTP).Methods("GET")
	api.HandleFunc("/clarify", clarificationHandler.ServeHTTP).Methods("POST")
	api.HandleFunc("/clarify/select", clarificationHandler.HandleSelection).Methods("POST")
	// Story B: Quality pipeline export
	api.HandleFunc("/quality/export", qualityPipeline.ServeHTTP).Methods("GET")
	// Story C: Fallback endpoint (for direct testing)
	api.HandleFunc("/fallback", fallbackHandler.ServeHTTP).Methods("POST")
	// Story D: Re-explain endpoint
	api.HandleFunc("/reexplain", reexplainHandler.ServeHTTP).Methods("POST")
	// Story G: Session context endpoints
	api.HandleFunc("/context/{sessionID}", contextHandler.GetContext).Methods("GET")
	api.HandleFunc("/context/{sessionID}", contextHandler.UpdateContext).Methods("PUT")

	// Apply middleware chain.
	// Each Handler() call wraps the current handler with middleware.
	// The LAST wrap call creates the OUTERMOST middleware (executes first).
	// Correct order: Recovery → RequestID → Logging → CORS → RateLimit → Auth → Metrics → Router
	var handler http.Handler = router
	handler = recoveryMiddleware.Handler(handler)         // Innermost: catches panics from all below
	handler = requestIDMiddleware.Handler(handler)        // Adds request ID early
	handler = loggingMiddleware.Handler(handler)          // Logs all requests
	handler = corsMiddleware.Handler(handler)             // Handles CORS preflight
	handler = rateLimiterMiddleware.Handler(handler)      // Rate limits before expensive auth
	handler = authMiddleware.Handler(handler)             // Auth before hitting router
	handler = middleware.PrometheusMiddleware(handler)    // Metrics tracking

	// Create HTTP server
	server := &http.Server{
		Addr:         ":" + strconv.Itoa(cfg.Server.Port),
		Handler:      handler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		log.Info().Int("port", cfg.Server.Port).Msg("API Gateway started")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("Server forced to shutdown")
	}

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		log.Error().Err(err).Msg("Error closing Redis connection")
	}

	log.Info().Msg("Server stopped")
}

// versionHandler handles GET /version
func versionHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"version":"` + Version + `","commit":"` + Commit + `","build_time":"` + BuildTime + `"}`))
}

// getEnv returns the value of an environment variable or a default.
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
