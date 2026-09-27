package config

import (
	"testing"
	"time"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			cfg: Config{
				Server: ServerConfig{
					Port: 8080,
				},
				Services: ServiceConfig{
					EmbeddingServiceURL:    "http://embedding:8081",
					VectorSearchServiceURL: "http://vector-search:8082",
					QueryUnderstandingURL:  "http://query-understanding:8083",
					LLMServiceURL:          "http://llm:8084",
				},
				Redis: RedisConfig{
					Addr: "localhost:6379",
				},
				Auth: AuthConfig{
					JWTSecret: "test-secret",
				},
				RateLimit: RateLimitConfig{
					RequestsPerMinute: 60,
					BurstSize:         10,
				},
				Timeouts: TimeoutConfig{
					EmbeddingService:    5 * time.Second,
					VectorSearchService: 10 * time.Second,
					QueryUnderstanding:  3 * time.Second,
					LLMService:          30 * time.Second,
					TotalRequest:        60 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "missing embedding service URL",
			cfg: Config{
				Server:   ServerConfig{Port: 8080},
				Services: ServiceConfig{},
				Redis:    RedisConfig{Addr: "localhost:6379"},
				Auth:     AuthConfig{JWTSecret: "test-secret"},
				RateLimit: RateLimitConfig{RequestsPerMinute: 60, BurstSize: 10},
				Timeouts: TimeoutConfig{
					EmbeddingService: 5*time.Second, VectorSearchService: 10*time.Second,
					QueryUnderstanding: 3*time.Second, LLMService: 30*time.Second, TotalRequest: 120*time.Second,
				},
			},
			wantErr: true,
			errMsg:  "embedding_service",
		},
		{
			name: "missing vector search service URL",
			cfg: Config{
				Server: ServerConfig{Port: 8080},
				Services: ServiceConfig{
					EmbeddingServiceURL: "http://embedding:8081",
				},
				Redis: RedisConfig{Addr: "localhost:6379"},
				Auth:  AuthConfig{JWTSecret: "test-secret"},
				RateLimit: RateLimitConfig{RequestsPerMinute: 60, BurstSize: 10},
				Timeouts: TimeoutConfig{
					EmbeddingService: 5*time.Second, VectorSearchService: 10*time.Second,
					QueryUnderstanding: 3*time.Second, LLMService: 30*time.Second, TotalRequest: 120*time.Second,
				},
			},
			wantErr: true,
			errMsg:  "vector_search",
		},
		{
			name: "missing Redis address",
			cfg: Config{
				Server: ServerConfig{Port: 8080},
				Services: ServiceConfig{
					EmbeddingServiceURL:    "http://embedding:8081",
					VectorSearchServiceURL: "http://vector-search:8082",
					QueryUnderstandingURL:  "http://query:8083",
					LLMServiceURL:          "http://llm:8084",
				},
				Redis: RedisConfig{},
				Auth:  AuthConfig{JWTSecret: "test-secret"},
				RateLimit: RateLimitConfig{RequestsPerMinute: 60, BurstSize: 10},
				Timeouts: TimeoutConfig{
					EmbeddingService: 5*time.Second, VectorSearchService: 10*time.Second,
					QueryUnderstanding: 3*time.Second, LLMService: 30*time.Second, TotalRequest: 120*time.Second,
				},
			},
			wantErr: true,
			errMsg:  "REDIS_ADDR",
		},
		{
			name: "missing JWT secret",
			cfg: Config{
				Server: ServerConfig{Port: 8080},
				Services: ServiceConfig{
					EmbeddingServiceURL:    "http://embedding:8081",
					VectorSearchServiceURL: "http://vector-search:8082",
					QueryUnderstandingURL:  "http://query:8083",
					LLMServiceURL:          "http://llm:8084",
				},
				Redis: RedisConfig{Addr: "localhost:6379"},
				Auth:  AuthConfig{},
				RateLimit: RateLimitConfig{RequestsPerMinute: 60, BurstSize: 10},
				Timeouts: TimeoutConfig{
					EmbeddingService: 5*time.Second, VectorSearchService: 10*time.Second,
					QueryUnderstanding: 3*time.Second, LLMService: 30*time.Second, TotalRequest: 120*time.Second,
				},
			},
			wantErr: true,
			errMsg:  "JWT_SECRET",
		},
		{
			name: "invalid port",
			cfg: Config{
				Server: ServerConfig{Port: -1},
				Services: ServiceConfig{
					EmbeddingServiceURL:    "http://embedding:8081",
					VectorSearchServiceURL: "http://vector-search:8082",
					QueryUnderstandingURL:  "http://query:8083",
					LLMServiceURL:          "http://llm:8084",
				},
				Redis: RedisConfig{Addr: "localhost:6379"},
				Auth:  AuthConfig{JWTSecret: "test-secret"},
				RateLimit: RateLimitConfig{RequestsPerMinute: 60, BurstSize: 10},
				Timeouts: TimeoutConfig{
					EmbeddingService: 5*time.Second, VectorSearchService: 10*time.Second,
					QueryUnderstanding: 3*time.Second, LLMService: 30*time.Second, TotalRequest: 120*time.Second,
				},
			},
			wantErr: true,
			errMsg:  "port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()

			if tt.wantErr {
				if err == nil {
					t.Errorf("Config.Validate() = nil, want error containing %q", tt.errMsg)
					return
				}
				if !containsStr(err.Error(), tt.errMsg) {
					t.Errorf("Config.Validate() error = %q, want containing %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("Config.Validate() = %v, want nil", err)
				}
			}
		})
	}
}

func TestServerConfig_PortValidation(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		wantValid   bool
	}{
		{"standard port 8080", 8080, true},
		{"port 80", 80, true},
		{"port 443", 443, true},
		{"zero port", 0, false},
		{"negative port", -1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{
				Server:   ServerConfig{Port: tt.port},
				Services: ServiceConfig{
					EmbeddingServiceURL:    "http://embedding:8081",
					VectorSearchServiceURL: "http://vector-search:8082",
					QueryUnderstandingURL:  "http://query:8083",
					LLMServiceURL:          "http://llm:8084",
				},
				Redis: RedisConfig{Addr: "localhost:6379"},
				Auth:  AuthConfig{JWTSecret: "test-secret"},
				RateLimit: RateLimitConfig{
					RequestsPerMinute: 60,
					BurstSize:         10,
				},
				Timeouts: TimeoutConfig{
					EmbeddingService:    5 * time.Second,
					VectorSearchService: 10 * time.Second,
					QueryUnderstanding:  3 * time.Second,
					LLMService:          30 * time.Second,
					TotalRequest:        120 * time.Second,
				},
			}
			err := cfg.Validate()
			if tt.wantValid && err != nil {
				t.Errorf("Config.Validate() = %v, want nil", err)
			}
			if !tt.wantValid && err == nil {
				t.Error("Config.Validate() = nil, want error")
			}
		})
	}
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
