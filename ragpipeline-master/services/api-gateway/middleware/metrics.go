// Package middleware provides Prometheus metrics collection for the API Gateway.
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
			Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
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

	// feedbackScoreTotal tracks thumbs up/down feedback counts.
	feedbackScoreTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "feedback_score_total",
			Help: "Total feedback scores (thumbs up/down)",
		},
		[]string{"score_type"},
	)

	// cacheHitRate tracks cache hit rate per cache layer.
	cacheHitRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_hit_rate",
			Help: "Cache hit rate (0-1) per cache layer",
		},
		[]string{"cache_layer"},
	)

	// circuitBreakerState tracks circuit breaker state per service.
	circuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Circuit breaker state per service (0=closed, 1=open, 2=half-open)",
		},
		[]string{"service"},
	)

	// activeConnections tracks active WebSocket connections.
	activeConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active WebSocket connections",
		},
	)

	// responseConfidenceScore tracks the distribution of response confidence scores (Story E).
	responseConfidenceScore = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "response_confidence_score",
			Help:    "Distribution of response confidence scores (0.0 to 1.0)",
			Buckets: []float64{0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
		},
		[]string{"source"}, // "rag", "fallback", "cache"
	)

	// lowConfidenceResponsesTotal counts responses with confidence below threshold (Story E).
	lowConfidenceResponsesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "low_confidence_responses_total",
			Help: "Total number of responses with confidence below 0.6 threshold",
		},
		[]string{"source"},
	)

	// ragGroundingScore tracks the distribution of RAG grounding validation scores (Story J).
	ragGroundingScore = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "rag_grounding_score",
			Help:    "Distribution of RAG grounding validation scores (0.0 to 1.0)",
			Buckets: []float64{0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0},
		},
		nil,
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

// RecordFeedback records a feedback metric
func RecordFeedback(thumbsUp bool) {
	if thumbsUp {
		feedbackScoreTotal.WithLabelValues("thumbs_up").Inc()
	} else {
		feedbackScoreTotal.WithLabelValues("thumbs_down").Inc()
	}
}

// RecordCacheHitRate records the cache hit rate for a specific layer
func RecordCacheHitRate(layer string, rate float64) {
	cacheHitRate.WithLabelValues(layer).Set(rate)
}

// RecordCircuitBreakerState records the circuit breaker state for a service
func RecordCircuitBreakerState(service string, state int) {
	circuitBreakerState.WithLabelValues(service).Set(float64(state))
}

// UpdateActiveConnections updates the active WebSocket connections gauge
func UpdateActiveConnections(count int64) {
	activeConnections.Set(float64(count))
}

// RecordConfidenceScore records a response confidence score (Story E)
func RecordConfidenceScore(source string, score float64) {
	responseConfidenceScore.WithLabelValues(source).Observe(score)
}

// RecordLowConfidenceResponse increments the low confidence response counter (Story E)
func RecordLowConfidenceResponse(source string) {
	lowConfidenceResponsesTotal.WithLabelValues(source).Inc()
}

// RecordGroundingScore records a RAG grounding validation score (Story J)
func RecordGroundingScore(score float64) {
	ragGroundingScore.WithLabelValues().Observe(score)
}
