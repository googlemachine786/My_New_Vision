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

	req := httptest.NewRequest(http.MethodPost, "/search", nil)
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

func TestRecordSearchLatency(t *testing.T) {
	RecordSearchLatency("hybrid", 0.05)
	RecordSearchLatency("dense", 0.03)
	RecordSearchLatency("sparse", 0.02)
	// Verify no panic/error
}

func TestRecordCacheHitRate(t *testing.T) {
	RecordCacheHitRate("search", 0.60)
	// Verify no panic/error
}
