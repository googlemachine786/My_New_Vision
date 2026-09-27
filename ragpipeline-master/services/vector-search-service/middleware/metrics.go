// Package middleware provides Prometheus metrics collection for the Vector Search Service.
package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// apiLatencySeconds tracks request duration as a histogram.
	apiLatencySeconds = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_latency_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0},
		},
		[]string{"method", "endpoint"},
	)

	// httpRequestsTotal counts total HTTP requests by endpoint and status.
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// searchLatency tracks vector search latency.
	searchLatency = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "search_latency_seconds",
			Help:    "Vector search latency in seconds",
			Buckets: []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"search_type"},
	)

	// cacheHitRate tracks cache hit rate per cache layer.
	cacheHitRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_hit_rate",
			Help: "Cache hit rate (0-1) per cache layer",
		},
		[]string{"cache_layer"},
	)

	// activeConnections tracks active connections (for consistency across services).
	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active connections",
		},
	)
)

// responseWriter wraps http.ResponseWriter to capture status codes
type metricsResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// WriteHeader captures the status code
func (rw *metricsResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// PrometheusMiddleware wraps HTTP handlers to record Prometheus metrics
func PrometheusMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &metricsResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(wrapped, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(wrapped.statusCode)
		endpoint := r.URL.Path

		apiLatencySeconds.WithLabelValues(r.Method, endpoint).Observe(duration)
		httpRequestsTotal.WithLabelValues(r.Method, endpoint, status).Inc()
	})
}

// MetricsHandler returns an HTTP handler for the /metrics endpoint
func MetricsHandler() http.Handler {
	return promhttp.Handler()
}

// RecordSearchLatency records vector search latency
func RecordSearchLatency(searchType string, duration float64) {
	searchLatency.WithLabelValues(searchType).Observe(duration)
}

// RecordCacheHitRate records the cache hit rate for a specific layer
func RecordCacheHitRate(layer string, rate float64) {
	cacheHitRate.WithLabelValues(layer).Set(rate)
}
