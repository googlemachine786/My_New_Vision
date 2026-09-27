package config

import (
	"fmt"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the API Gateway
type Config struct {
	Server   ServerConfig
	Services ServiceConfig
	Redis    RedisConfig
	Auth     AuthConfig
	RateLimit RateLimitConfig
	Timeouts  TimeoutConfig
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
	ShutdownTimeout time.Duration
	Environment  string // development, staging, production
}

// ServiceConfig holds downstream service URLs
type ServiceConfig struct {
	EmbeddingServiceURL    string
	VectorSearchServiceURL string
	QueryUnderstandingURL  string
	LLMServiceURL          string // Gemini LLM service
	RerankerServiceURL     string // FlashRank reranker service
}

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	PoolSize int
}

// AuthConfig holds JWT authentication configuration
type AuthConfig struct {
	JWTSecret string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
}

// TimeoutConfig holds timeout configuration for downstream calls
type TimeoutConfig struct {
	EmbeddingService    time.Duration
	VectorSearchService time.Duration
	QueryUnderstanding  time.Duration
	LLMService          time.Duration
	TotalRequest        time.Duration // Must be > sum of all downstream timeouts
}

// Load loads configuration from environment variables with .env fallback
func Load() (*Config, error) {
	// Attempt to load .env file (ignore error if not present)
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnvInt("SERVER_PORT", 8080),
			ReadTimeout:     getEnvDuration("SERVER_READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getEnvDuration("SERVER_WRITE_TIMEOUT", 120*time.Second),
			IdleTimeout:     getEnvDuration("SERVER_IDLE_TIMEOUT", 30*time.Second),
			ShutdownTimeout: getEnvDuration("SERVER_SHUTDOWN_TIMEOUT", 40*time.Second),
			Environment:     getEnvString("SERVER_ENV", "development"),
		},
		Services: ServiceConfig{
			EmbeddingServiceURL:    getEnvString("EMBEDDING_SERVICE_URL", "http://localhost:8081"),
			VectorSearchServiceURL: getEnvString("VECTOR_SEARCH_SERVICE_URL", "http://localhost:8082"),
			QueryUnderstandingURL:  getEnvString("QUERY_UNDERSTANDING_URL", "http://localhost:8083"),
			LLMServiceURL:          getEnvString("LLM_SERVICE_URL", "http://localhost:8084"),
			RerankerServiceURL:     getEnvString("RERANKER_SERVICE_URL", "http://localhost:8085"),
		},
		Redis: RedisConfig{
			Addr:     getEnvString("REDIS_ADDR", "localhost:6379"),
			Password: getEnvString("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
			PoolSize: getEnvInt("REDIS_POOL_SIZE", 10),
		},
		Auth: AuthConfig{
			JWTSecret: getEnvString("JWT_SECRET", "change-me-in-production"),
		},
		RateLimit: RateLimitConfig{
			RequestsPerMinute: getEnvInt("RATE_LIMIT_RPM", 100),
			BurstSize:         getEnvInt("RATE_LIMIT_BURST", 20),
		},
		Timeouts: TimeoutConfig{
			EmbeddingService:    getEnvDuration("TIMEOUT_EMBEDDING", 10*time.Second),
			VectorSearchService: getEnvDuration("TIMEOUT_VECTOR_SEARCH", 15*time.Second),
			QueryUnderstanding:  getEnvDuration("TIMEOUT_QUERY_UNDERSTANDING", 5*time.Second),
			LLMService:          getEnvDuration("TIMEOUT_LLM", 60*time.Second),
			TotalRequest:        getEnvDuration("TIMEOUT_TOTAL", 120*time.Second),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

// Validate ensures configuration values are sane
func (c *Config) Validate() error {
	if c.Server.Port < 1 || c.Server.Port > 65535 {
		return fmt.Errorf("server port must be between 1 and 65535, got %d", c.Server.Port)
	}

	if c.Auth.JWTSecret == "" || c.Auth.JWTSecret == "change-me-in-production" {
		return fmt.Errorf("JWT_SECRET must be set to a strong secret (not 'change-me-in-production')")
	}

	if c.RateLimit.RequestsPerMinute <= 0 {
		return fmt.Errorf("rate limit requests per minute must be > 0")
	}

	if c.RateLimit.BurstSize <= 0 {
		return fmt.Errorf("rate limit burst size must be > 0")
	}

	// Validate timeouts are positive
	timeouts := []struct {
		name  string
		value time.Duration
	}{
		{"embedding", c.Timeouts.EmbeddingService},
		{"vector_search", c.Timeouts.VectorSearchService},
		{"query_understanding", c.Timeouts.QueryUnderstanding},
		{"llm", c.Timeouts.LLMService},
		{"total", c.Timeouts.TotalRequest},
	}

	for _, t := range timeouts {
		if t.value <= 0 {
			return fmt.Errorf("timeout %s must be > 0", t.name)
		}
	}

	// Total timeout must be greater than sum of individual timeouts
	sumTimeouts := c.Timeouts.EmbeddingService + c.Timeouts.VectorSearchService +
		c.Timeouts.QueryUnderstanding + c.Timeouts.LLMService

	if c.Timeouts.TotalRequest <= sumTimeouts {
		return fmt.Errorf("total timeout (%v) must be greater than sum of individual timeouts (%v)",
			c.Timeouts.TotalRequest, sumTimeouts)
	}

	// Validate service URLs
	urls := []struct {
		name     string
		url      string
		required bool
	}{
		{"embedding_service", c.Services.EmbeddingServiceURL, true},
		{"vector_search", c.Services.VectorSearchServiceURL, true},
		{"query_understanding", c.Services.QueryUnderstandingURL, true},
		{"reranker", c.Services.RerankerServiceURL, false}, // optional
	}

	for _, u := range urls {
		if u.required && u.url == "" {
			return fmt.Errorf("%s URL is required", u.name)
		}
	}

	// Redis is required
	if c.Redis.Addr == "" {
		return fmt.Errorf("REDIS_ADDR is required")
	}

	return nil
}
