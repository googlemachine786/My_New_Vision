// Package config provides configuration management for the embedding service.
// All configuration is loaded from environment variables with sensible defaults.
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Config holds all configuration for the embedding service.
type Config struct {
	// Server
	Port        string
	Environment string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	// Vertex AI
	VertexProject   string
	VertexLocation  string
	VertexModel     string
	VertexDimension int
	VertexTimeout   time.Duration

	// Task types
	DocumentTaskType string
	QueryTaskType    string

	// Retry
	MaxRetries    int
	InitialBackoff time.Duration
	MaxBackoff    time.Duration

	// Observability
	OTelEndpoint    string
	OTelServiceName string
	OTelSampleRate  float64
	LogLevel        string

	// Request limits
	MaxTextsPerBatch int
	MaxTextLength    int

	// Cache (CAG)
	RedisURL             string
	EmbeddingCacheEnabled bool
	EmbeddingCacheTTL    time.Duration
	EmbeddingCacheMaxSize int
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	// Load .env file if it exists (development only)
	_ = godotenv.Load(".env")
	_ = godotenv.Load(".env.local")

	cfg := &Config{
		Port:             getEnv("PORT", "8081"),
		Environment:      getEnv("ENVIRONMENT", "production"),
		ReadTimeout:      getDurationEnv("READ_TIMEOUT", 10*time.Second),
		WriteTimeout:     getDurationEnv("WRITE_TIMEOUT", 35*time.Second),
		IdleTimeout:      getDurationEnv("IDLE_TIMEOUT", 120*time.Second),

		VertexProject:    getEnv("VERTEX_PROJECT", ""),
		VertexLocation:   getEnv("VERTEX_LOCATION", "us-central1"),
		VertexModel:      getEnv("VERTEX_MODEL", "text-embedding-005"),
		VertexDimension:  getIntEnv("VERTEX_DIMENSION", 768),
		VertexTimeout:    getDurationEnv("VERTEX_TIMEOUT", 30*time.Second),

		DocumentTaskType: getEnv("DOCUMENT_TASK_TYPE", "RETRIEVAL_DOCUMENT"),
		QueryTaskType:    getEnv("QUERY_TASK_TYPE", "RETRIEVAL_QUERY"),

		MaxRetries:       getIntEnv("MAX_RETRIES", 5),
		InitialBackoff:   getDurationEnv("INITIAL_BACKOFF", 100*time.Millisecond),
		MaxBackoff:       getDurationEnv("MAX_BACKOFF", 5*time.Second),

		OTelEndpoint:     getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		OTelServiceName:  getEnv("OTEL_SERVICE_NAME", "embedding-service"),
		OTelSampleRate:   getFloatEnv("OTEL_SAMPLE_RATE", 1.0),
		LogLevel:         getEnv("LOG_LEVEL", "info"),

		MaxTextsPerBatch: getIntEnv("MAX_TEXTS_PER_BATCH", 5),
		MaxTextLength:    getIntEnv("MAX_TEXT_LENGTH", 20000),

		RedisURL:             getEnv("REDIS_URL", "redis://localhost:6379/0"),
		EmbeddingCacheEnabled: getEnvBool("EMBEDDING_CACHE_ENABLED", true),
		EmbeddingCacheTTL:    getDurationEnv("EMBEDDING_CACHE_TTL", 24*time.Hour),
		EmbeddingCacheMaxSize: getIntEnv("EMBEDDING_CACHE_MAX_SIZE", 10000),
	}

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	// Set log level
	level, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		log.Warn().Str("level", cfg.LogLevel).Msg("Invalid log level, using info")
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure log output format
	if cfg.Environment != "production" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}

	log.Info().
		Str("environment", cfg.Environment).
		Str("port", cfg.Port).
		Str("vertex_project", cfg.VertexProject).
		Str("vertex_location", cfg.VertexLocation).
		Str("model", cfg.VertexModel).
		Int("dimension", cfg.VertexDimension).
		Int("max_retries", cfg.MaxRetries).
		Int("max_texts_per_batch", cfg.MaxTextsPerBatch).
		Int("max_text_length", cfg.MaxTextLength).
		Msg("Configuration loaded")

	return cfg, nil
}

// Validate checks that all required configuration fields are set.
func (c *Config) Validate() error {
	if c.VertexProject == "" {
		return fmt.Errorf("VERTEX_PROJECT is required")
	}
	if c.VertexLocation == "" {
		return fmt.Errorf("VERTEX_LOCATION is required")
	}
	if c.MaxRetries < 1 || c.MaxRetries > 10 {
		return fmt.Errorf("MAX_RETRIES must be between 1 and 10")
	}
	if c.MaxTextsPerBatch < 1 || c.MaxTextsPerBatch > 10 {
		return fmt.Errorf("MAX_TEXTS_PER_BATCH must be between 1 and 10")
	}
	if c.MaxTextLength < 100 || c.MaxTextLength > 100000 {
		return fmt.Errorf("MAX_TEXT_LENGTH must be between 100 and 100000")
	}
	return nil
}

// VertexAIEndpoint returns the Vertex AI endpoint URL.
func (c *Config) VertexAIEndpoint() string {
	return fmt.Sprintf("%s-aiplatform.googleapis.com", c.VertexLocation)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if d, err := time.ParseDuration(value); err == nil {
			return d
		}
		log.Warn().Str("key", key).Str("value", value).Msg("Invalid duration, using default")
	}
	return defaultValue
}

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

func getFloatEnv(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		v, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
		if err == nil {
			return v
		}
		log.Warn().Str("key", key).Str("value", value).Msg("Invalid float, using default")
	}
	return defaultValue
}

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
