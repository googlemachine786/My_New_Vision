// Package config provides configuration management for the vector search service.
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

// Config holds all configuration for the vector search service.
type Config struct {
	// Server
	Port         string
	Environment  string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration

	// Database
	DatabaseURL       string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration

	// Search
	DefaultTopK     int
	MaxTopK         int
	QueryTimeout    time.Duration
	SimilarityThreshold float64

	// Dense Search
	DenseChildTable  string
	DenseParentTable string
	EmbeddingColumn  string
	EmbeddingDim     int

	// Sparse Search
	SparseTable        string
	KeywordsColumn     string
	SparseOverlapLimit int

	// RRF Fusion
	RRFK            float64
	RRFDenseWeight  float64
	RRFSparseWeight float64

	// Observability
	OTelEndpoint    string
	OTelServiceName string
	OTelSampleRate  float64
	LogLevel        string

	// Performance
	EnableExplainAnalyze bool
	EnableQueryMetrics   bool

	// Reranker
	RerankerServiceURL string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	// Load .env file if it exists (development only)
	_ = godotenv.Load(".env")
	_ = godotenv.Load(".env.local")

	cfg := &Config{
		// Server
		Port:         getEnv("PORT", "8082"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		ReadTimeout:  getDurationEnv("READ_TIMEOUT", 10*time.Second),
		WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 60*time.Second),
		IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", 120*time.Second),

		// Database
		DatabaseURL:       getEnv("DATABASE_URL", ""),
		MaxConns:          int32(getIntEnv("DB_MAX_CONNS", 25)),
		MinConns:          int32(getIntEnv("DB_MIN_CONNS", 5)),
		MaxConnLifetime:   getDurationEnv("DB_MAX_CONN_LIFETIME", 30*time.Minute),
		MaxConnIdleTime:   getDurationEnv("DB_MAX_CONN_IDLE_TIME", 5*time.Minute),
		HealthCheckPeriod: getDurationEnv("DB_HEALTH_CHECK_PERIOD", 30*time.Second),

		// Search
		DefaultTopK:         getIntEnv("DEFAULT_TOP_K", 10),
		MaxTopK:             getIntEnv("MAX_TOP_K", 100),
		QueryTimeout:        getDurationEnv("QUERY_TIMEOUT", 30*time.Second),
		SimilarityThreshold: getFloatEnv("SIMILARITY_THRESHOLD", 0.36),

		// Dense Search
		DenseChildTable:  getEnv("DENSE_CHILD_TABLE", "child_chunks"),
		DenseParentTable: getEnv("DENSE_PARENT_TABLE", "parent_chunks"),
		EmbeddingColumn:  getEnv("EMBEDDING_COLUMN", "embedding"),
		EmbeddingDim:     getIntEnv("EMBEDDING_DIM", 768),

		// Sparse Search
		SparseTable:        getEnv("SPARSE_TABLE", "parent_chunks"),
		KeywordsColumn:     getEnv("KEYWORDS_COLUMN", "extracted_keywords"),
		SparseOverlapLimit: getIntEnv("SPARSE_OVERLAP_LIMIT", 200),

		// RRF Fusion
		RRFK:            getFloatEnv("RRF_K", 60.0),
		RRFDenseWeight:  getFloatEnv("RRF_DENSE_WEIGHT", 1.0),
		RRFSparseWeight: getFloatEnv("RRF_SPARSE_WEIGHT", 1.0),

		// Observability
		OTelEndpoint:    getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		OTelServiceName: getEnv("OTEL_SERVICE_NAME", "vector-search-service"),
		OTelSampleRate:  getFloatEnv("OTEL_SAMPLE_RATE", 1.0),
		LogLevel:        getEnv("LOG_LEVEL", "info"),

		// Performance
		EnableExplainAnalyze: getBoolEnv("ENABLE_EXPLAIN_ANALYZE", false),
		EnableQueryMetrics:   getBoolEnv("ENABLE_QUERY_METRICS", true),

		// Reranker
		RerankerServiceURL: getEnv("RERANKER_SERVICE_URL", ""),
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
		Int("max_conns", int(cfg.MaxConns)).
		Int("default_top_k", cfg.DefaultTopK).
		Float64("rrf_k", cfg.RRFK).
		Str("query_timeout", cfg.QueryTimeout.String()).
		Msg("Configuration loaded")

	return cfg, nil
}

// Validate checks that all required configuration fields are set.
func (c *Config) Validate() error {
	if c.DatabaseURL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}
	if c.MaxConns < 1 {
		return fmt.Errorf("DB_MAX_CONNS must be at least 1")
	}
	if c.MinConns < 0 {
		return fmt.Errorf("DB_MIN_CONNS must be non-negative")
	}
	if c.MinConns > c.MaxConns {
		return fmt.Errorf("DB_MIN_CONNS cannot exceed DB_MAX_CONNS")
	}
	if c.DefaultTopK < 1 {
		return fmt.Errorf("DEFAULT_TOP_K must be at least 1")
	}
	if c.MaxTopK < c.DefaultTopK {
		return fmt.Errorf("MAX_TOP_K must be at least DEFAULT_TOP_K")
	}
	if c.QueryTimeout < 1*time.Second {
		return fmt.Errorf("QUERY_TIMEOUT must be at least 1s")
	}
	if c.RRFK <= 0 {
		return fmt.Errorf("RRF_K must be positive")
	}
	if c.EmbeddingDim < 1 {
		return fmt.Errorf("EMBEDDING_DIM must be positive")
	}
	return nil
}

// DenseSearchLimit returns the maximum number of results to fetch from dense search.
// We fetch more than top_k to allow for deduplication after RRF fusion.
func (c *Config) DenseSearchLimit(topK int) int {
	limit := topK * 20
	if limit > c.MaxTopK*10 {
		return c.MaxTopK * 10
	}
	return limit
}

// SparseSearchLimit returns the maximum number of results to fetch from sparse search.
func (c *Config) SparseSearchLimit(topK int) int {
	limit := topK * 20
	if limit > c.SparseOverlapLimit {
		return c.SparseOverlapLimit
	}
	return limit
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

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
		log.Warn().Str("key", key).Str("value", value).Msg("Invalid bool, using default")
	}
	return defaultValue
}
