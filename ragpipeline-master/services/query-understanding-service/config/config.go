package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Server  ServerConfig
	Gemini  GeminiConfig
	Prompts PromptConfig
}

type ServerConfig struct {
	Port            int
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	MaxHistoryTurns int
}

type GeminiConfig struct {
	ProjectID     string
	Location      string
	Model         string
	Temperature   float32
	MaxTokens     int
	TopP          float32
	TopK          int
	RequestTimeout time.Duration
	JSONKeyPath   string // Optional: path to service account JSON key
	MaxRetries    int    // Maximum retry attempts for LLM calls
	RateLimitPerSecond int // Rate limit for LLM API calls
	RateLimitBurst     int // Burst size for rate limiter
}

type PromptConfig struct {
	RewriteVersion         string
	ParseFiltersVersion    string
	ClassifyIntentVersion  string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	_ = godotenv.Load()

	cfg := &Config{
		Server: ServerConfig{
			Port:            getEnvInt("PORT", 8083),
			ReadTimeout:     getEnvDuration("READ_TIMEOUT", 10*time.Second),
			WriteTimeout:    getEnvDuration("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:     getEnvDuration("IDLE_TIMEOUT", 60*time.Second),
			MaxHistoryTurns: getEnvInt("MAX_HISTORY_TURNS", 10),
		},
		Gemini: GeminiConfig{
			ProjectID:          getEnv("GEMINI_PROJECT_ID", ""),
			Location:           getEnv("GEMINI_LOCATION", "us-central1"),
			Model:              getEnv("GEMINI_MODEL", "gemini-2.0-flash"),
			Temperature:        getEnvFloat32("GEMINI_TEMPERATURE", 0.1),
			MaxTokens:          getEnvInt("GEMINI_MAX_TOKENS", 512),
			TopP:               getEnvFloat32("GEMINI_TOP_P", 0.95),
			TopK:               getEnvInt("GEMINI_TOP_K", 40),
			RequestTimeout:     getEnvDuration("GEMINI_REQUEST_TIMEOUT", 30*time.Second),
			JSONKeyPath:        getEnv("GOOGLE_APPLICATION_CREDENTIALS", ""),
			MaxRetries:         getEnvInt("GEMINI_MAX_RETRIES", 3),
			RateLimitPerSecond: getEnvInt("GEMINI_RATE_LIMIT", 10),
			RateLimitBurst:     getEnvInt("GEMINI_RATE_BURST", 20),
		},
		Prompts: PromptConfig{
			RewriteVersion:      getEnv("PROMPT_REWRITE_VERSION", "v1"),
			ParseFiltersVersion: getEnv("PROMPT_PARSE_FILTERS_VERSION", "v1"),
			ClassifyIntentVersion: getEnv("PROMPT_CLASSIFY_INTENT_VERSION", "v1"),
		},
	}

	// Validate required fields
	if cfg.Gemini.ProjectID == "" {
		return nil, fmt.Errorf("GEMINI_PROJECT_ID is required")
	}

	return cfg, nil
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvFloat32(key string, defaultVal float32) float32 {
	if val := os.Getenv(key); val != "" {
		if f, err := strconv.ParseFloat(val, 32); err == nil {
			return float32(f)
		}
	}
	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val := os.Getenv(key); val != "" {
		if d, err := time.ParseDuration(val); err == nil {
			return d
		}
	}
	return defaultVal
}

// ValidateSubject checks if a subject is in the allowed list.
func ValidateSubject(subject string) bool {
	allowed := map[string]bool{
		"math": true, "mathematics": true,
		"physics": true, "chemistry": true, "biology": true,
		"english": true, "history": true, "geography": true,
		"science": true, "computer science": true, "computers": true,
		"economics": true, "accountancy": true, "business studies": true,
	}
	return allowed[strings.ToLower(strings.TrimSpace(subject))]
}

// ValidateGrade checks if a grade is in the valid range (1-12).
func ValidateGrade(grade string) bool {
	if g, err := strconv.Atoi(grade); err == nil {
		return g >= 1 && g <= 12
	}
	return false
}
