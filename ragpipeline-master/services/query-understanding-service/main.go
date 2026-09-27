package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"query-understanding-service/classifier"
	"query-understanding-service/config"
	"query-understanding-service/handler"
	"query-understanding-service/llm"
	"query-understanding-service/middleware"
	"query-understanding-service/parser"
	"query-understanding-service/rewriter"
)

func main() {
	logger := log.New(os.Stdout, "[query-understanding] ", log.LstdFlags|log.Lshortfile)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatalf("Failed to load config: %v", err)
	}
	logger.Printf("Configuration loaded: port=%d, model=%s", cfg.Server.Port, cfg.Gemini.Model)

	// Initialize Gemini client
	ctx := context.Background()
	geminiClient, err := llm.NewClient(ctx, cfg.Gemini)
	if err != nil {
		logger.Fatalf("Failed to initialize Gemini client: %v", err)
	}
	defer geminiClient.Close()
	logger.Printf("Gemini client initialized: location=%s", cfg.Gemini.Location)

	// Initialize services
	rewriteService := rewriter.NewService(
		geminiClient,
		rewriter.WithMaxTurns(cfg.Server.MaxHistoryTurns),
		rewriter.WithPromptVersion(cfg.Prompts.RewriteVersion),
	)

	parseService := parser.NewService(
		geminiClient,
		cfg.Prompts.ParseFiltersVersion,
	)

	classifyService := classifier.NewService(
		geminiClient,
		cfg.Prompts.ClassifyIntentVersion,
	)

	logger.Printf("Services initialized: rewrite, parse-filters, classify-intent")

	// Set up HTTP handler
	services := &handler.Services{
		Rewrite:  rewriteService,
		Parse:    parseService,
		Classify: classifyService,
	}

	queryHandler := handler.NewQueryHandler(services, logger)
	mux := http.NewServeMux()
	queryHandler.RegisterRoutes(mux)

	// Register Prometheus metrics endpoint
	mux.Handle("GET /metrics", middleware.MetricsHandler())

	// Add middleware chain
	finalHandler := requestIDMiddleware(loggingMiddleware(concurrencyLimitMiddleware(mux, 100), logger))
	finalHandler = middleware.PrometheusMiddleware(finalHandler)

	// Create HTTP server
	addr := fmt.Sprintf(":%d", cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      finalHandler,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Printf("Starting HTTP server on %s", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Println("Shutting down server...")

	// Shutdown timeout must be longer than the longest request timeout
	// LLM requests can take up to RequestTimeout (30s), so shutdown needs 40s
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Println("Server stopped gracefully")
}

// requestIDMiddleware adds a unique request ID to each request.
func requestIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()
		w.Header().Set("X-Request-ID", requestID)
		r = r.WithContext(context.WithValue(r.Context(), "request_id", requestID))
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware logs each request with duration.
func loggingMiddleware(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		logger.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

// concurrencyLimitMiddleware limits the number of concurrent requests.
func concurrencyLimitMiddleware(next http.Handler, limit int) http.Handler {
	sem := make(chan struct{}, limit)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case sem <- struct{}{}:
			defer func() { <-sem }()
			next.ServeHTTP(w, r)
		case <-r.Context().Done():
			http.Error(w, "request cancelled", http.StatusServiceUnavailable)
		default:
			http.Error(w, "service overloaded", http.StatusServiceUnavailable)
			return
		}
	})
}

// generateRequestID creates a simple request ID.
func generateRequestID() string {
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}
