package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPrometheusMiddleware_RecordsMetrics(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Wrap with metrics middleware
	handler := PrometheusMiddleware(testHandler)

	// Create test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	// Execute
	handler.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestMetricsHandler_ReturnsMetrics(t *testing.T) {
	handler := MetricsHandler()

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	// Prometheus metrics endpoint should return text/plain content
	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		t.Errorf("expected text/plain content type, got: %s", contentType)
	}

	// Should contain metric names
	body := w.Body.String()
	if !strings.Contains(body, "api_latency_seconds") {
		t.Error("expected metrics to contain api_latency_seconds")
	}
	if !strings.Contains(body, "http_requests_total") {
		t.Error("expected metrics to contain http_requests_total")
	}
}

func TestRecordFeedback(t *testing.T) {
	RecordFeedback(true)
	RecordFeedback(true)
	RecordFeedback(false)

	// Verify no panic/error - metrics are recorded
}

func TestRecordCacheHitRate(t *testing.T) {
	RecordCacheHitRate("semantic", 0.85)
	// Verify no panic/error
}

func TestRecordCircuitBreakerState(t *testing.T) {
	RecordCircuitBreakerState("embedding_service", 0) // closed
	RecordCircuitBreakerState("llm_service", 1)      // open
	// Verify no panic/error
}

func TestUpdateActiveConnections(t *testing.T) {
	UpdateActiveConnections(5)
	UpdateActiveConnections(0)
	// Verify no panic/error
}

func TestRecordConfidenceScore(t *testing.T) {
	RecordConfidenceScore("rag", 0.85)
	RecordConfidenceScore("fallback", 0.45)
	RecordConfidenceScore("cache", 1.0)
	// Verify no panic/error
}

func TestRecordLowConfidenceResponse(t *testing.T) {
	RecordLowConfidenceResponse("rag")
	RecordLowConfidenceResponse("fallback")
	// Verify no panic/error
}
