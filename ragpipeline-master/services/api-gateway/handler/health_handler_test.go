package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/visionary/ragpipeline/services/api-gateway/client"
	"github.com/visionary/ragpipeline/pkg/types"
)

func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	handler := &HealthHandler{}

	tests := []struct {
		name   string
		method string
	}{
		{"POST method", http.MethodPost},
		{"PUT method", http.MethodPut},
		{"DELETE method", http.MethodDelete},
		{"PATCH method", http.MethodPatch},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/health", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusMethodNotAllowed {
				t.Errorf("Health handler should return 405 for %s, got %d", tt.method, w.Code)
			}

			var resp types.ErrorResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("Failed to decode response: %v", err)
			}
			if resp.Error.Code != "method_not_allowed" {
				t.Errorf("Error code = %q, want %q", resp.Error.Code, "method_not_allowed")
			}
		})
	}
}

func TestHealthHandler_GetWithNilClients(t *testing.T) {
	handler := &HealthHandler{
		EmbeddingClient:    nil,
		VectorSearchClient: nil,
		QueryUnderstandingClient: nil,
		RedisPinger:        nil,
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Health handler with nil clients should return 503, got %d", w.Code)
	}

	var resp types.HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != "unhealthy" {
		t.Errorf("Status = %q, want %q", resp.Status, "unhealthy")
	}
	if resp.Services["embedding_service"] != "unhealthy" {
		t.Errorf("embedding_service = %q, want %q", resp.Services["embedding_service"], "unhealthy")
	}
	if resp.Services["redis"] != "unknown" {
		t.Errorf("redis = %q, want %q", resp.Services["redis"], "unknown")
	}
	if resp.Timestamp == "" {
		t.Error("Timestamp is empty")
	}
}

func TestHealthHandler_GetWithHealthyRedisPinger(t *testing.T) {
	handler := &HealthHandler{
		EmbeddingClient:    nil,
		VectorSearchClient: nil,
		QueryUnderstandingClient: nil,
		RedisPinger: func(ctx context.Context) error {
			return nil // Redis is healthy
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	var resp types.HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Services["redis"] != "healthy" {
		t.Errorf("redis = %q, want %q", resp.Services["redis"], "healthy")
	}
}

func TestHealthHandler_GetWithUnhealthyRedisPinger(t *testing.T) {
	handler := &HealthHandler{
		EmbeddingClient:    nil,
		VectorSearchClient: nil,
		QueryUnderstandingClient: nil,
		RedisPinger: func(ctx context.Context) error {
			return nil // Even if pinger is called, all other services are nil so overall unhealthy
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	// All services nil except Redis = degraded (not fully unhealthy)
	var resp types.HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp.Status != "degraded" {
		t.Errorf("Status = %q, want %q (Redis healthy but other services nil)", resp.Status, "degraded")
	}
}

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		data       interface{}
		wantStatus int
	}{
		{
			name:       "200 OK",
			status:     http.StatusOK,
			data:       types.HealthResponse{Status: "healthy"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "400 Bad Request",
			status:     http.StatusBadRequest,
			data:       types.ErrorResponse{Error: types.ErrorDetail{Code: "bad_request"}},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "500 Internal Server Error",
			status:     http.StatusInternalServerError,
			data:       types.ErrorResponse{Error: types.ErrorDetail{Code: "internal_error"}},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			writeJSON(w, tt.status, tt.data)

			if w.Code != tt.wantStatus {
				t.Errorf("writeJSON() status = %d, want %d", w.Code, tt.wantStatus)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json" {
				t.Errorf("Content-Type = %q, want %q", contentType, "application/json")
			}

			// Verify body is valid JSON
			var decoded interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &decoded); err != nil {
				t.Errorf("Response body is not valid JSON: %v", err)
			}
		})
	}
}

func TestNewHealthHandler(t *testing.T) {
	// Create real ServiceClient instances pointing to localhost
	embeddingClient := client.NewServiceClient("embedding-service", "http://localhost:8081", 10*time.Second)
	vectorClient := client.NewServiceClient("vector-search-service", "http://localhost:8082", 10*time.Second)
	queryClient := client.NewServiceClient("query-understanding-service", "http://localhost:8083", 10*time.Second)
	redisPinger := func(ctx context.Context) error { return nil }

	handler := NewHealthHandler(embeddingClient, vectorClient, queryClient, redisPinger)

	// Verify clients are set by checking their Name() method
	if handler.EmbeddingClient.Name() != "embedding-service" {
		t.Errorf("EmbeddingClient name = %q, want %q", handler.EmbeddingClient.Name(), "embedding-service")
	}
	if handler.VectorSearchClient.Name() != "vector-search-service" {
		t.Errorf("VectorSearchClient name = %q, want %q", handler.VectorSearchClient.Name(), "vector-search-service")
	}
	if handler.QueryUnderstandingClient.Name() != "query-understanding-service" {
		t.Errorf("QueryUnderstandingClient name = %q, want %q", handler.QueryUnderstandingClient.Name(), "query-understanding-service")
	}
	if handler.RedisPinger == nil {
		t.Error("RedisPinger is nil")
	}
}

// testServiceClient is a minimal mock for testing with nil client scenarios
type testServiceClient struct {
	name string
}

func (c *testServiceClient) Name() string {
	return c.name
}

// Timeout handler test
func TestHealthHandler_ResponseContainsTimestamp(t *testing.T) {
	handler := &HealthHandler{}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	var resp types.HealthResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify timestamp is in RFC3339 format
	_, err := time.Parse(time.RFC3339, resp.Timestamp)
	if err != nil {
		t.Errorf("Timestamp %q is not valid RFC3339: %v", resp.Timestamp, err)
	}
}
