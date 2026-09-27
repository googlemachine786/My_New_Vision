package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPrometheusMiddleware_RecordsMetrics(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	handler := PrometheusMiddleware(testHandler)

	req := httptest.NewRequest(http.MethodPost, "/classify-intent", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

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

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/plain") {
		t.Errorf("expected text/plain content type, got: %s", contentType)
	}

	body := w.Body.String()
	if !strings.Contains(body, "api_latency_seconds") {
		t.Error("expected metrics to contain api_latency_seconds")
	}
}

func TestRecordLLMLatency(t *testing.T) {
	RecordLLMLatency("intent_classification", 1.2)
	RecordLLMLatency("query_rewrite", 0.8)
	RecordLLMLatency("filter_parse", 0.5)
	// Verify no panic/error
}

func TestRecordIntentConfidence(t *testing.T) {
	RecordIntentConfidence("information_seeking", 0.85)
	RecordIntentConfidence("problem_solving", 0.45)
	// Verify no panic/error
}

func TestRecordCacheHitRate(t *testing.T) {
	RecordCacheHitRate("query_cache", 0.70)
	// Verify no panic/error
}

func TestRecordCircuitBreakerState(t *testing.T) {
	RecordCircuitBreakerState("llm_service", 0)
	RecordCircuitBreakerState("llm_service", 1)
	// Verify no panic/error
}
