package handler

import (
	"net/http"
	"time"

	"github.com/visionary/ragpipeline/services/api-gateway/cache"
	"github.com/visionary/ragpipeline/services/api-gateway/client"
	"github.com/visionary/ragpipeline/services/api-gateway/middleware"
	"github.com/visionary/ragpipeline/pkg/types"
)

// StatsHandler handles the stats endpoint
type StatsHandler struct {
	EmbeddingClient    *client.ServiceClient
	VectorSearchClient *client.ServiceClient
	QueryUnderstandingClient *client.ServiceClient
	LLMClient          *client.ServiceClient
	Cache              *cache.ResponseCache
	RateLimiter        *middleware.RateLimiterMiddleware
	StartTime          time.Time
}

// NewStatsHandler creates a new stats handler
func NewStatsHandler(
	embeddingClient, vectorSearchClient, queryUnderstandingClient, llmClient *client.ServiceClient,
	cache *cache.ResponseCache,
	rateLimiter *middleware.RateLimiterMiddleware,
) *StatsHandler {
	return &StatsHandler{
		EmbeddingClient:    embeddingClient,
		VectorSearchClient: vectorSearchClient,
		QueryUnderstandingClient: queryUnderstandingClient,
		LLMClient:          llmClient,
		Cache:              cache,
		RateLimiter:        rateLimiter,
		StartTime:          time.Now(),
	}
}

// ServeHTTP handles GET /stats
func (h *StatsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "method_not_allowed", Message: "Only GET is allowed"},
		})
		return
	}

	// Aggregate metrics from all service clients
	var totalRequests, totalErrors int64
	var totalLatency time.Duration

	clients := []*client.ServiceClient{
		h.EmbeddingClient,
		h.VectorSearchClient,
		h.QueryUnderstandingClient,
		h.LLMClient,
	}

	for _, c := range clients {
		if c != nil {
			stats := c.Stats()
			if rc, ok := stats["request_count"].(int64); ok {
				totalRequests += rc
			}
			if ec, ok := stats["error_count"].(int64); ok {
				totalErrors += ec
			}
			if al, ok := stats["average_latency"].(string); ok {
				if d, err := time.ParseDuration(al); err == nil {
					totalLatency += d
				}
			}
		}
	}

	avgLatency := time.Duration(0)
	if totalRequests > 0 {
		avgLatency = totalLatency / time.Duration(totalRequests)
	}

	// Cache stats
	cacheStats := map[string]interface{}{}
	if h.Cache != nil {
		cacheStats = h.Cache.Stats()
	}

	// Circuit breaker states
	cbStates := make(map[string]interface{}, 4) // healthy, degraded, error, timeout
	if h.EmbeddingClient != nil {
		cbStates["embedding_service"] = h.EmbeddingClient.CircuitBreaker.Stats()
	}
	if h.VectorSearchClient != nil {
		cbStates["vector_search"] = h.VectorSearchClient.CircuitBreaker.Stats()
	}
	if h.QueryUnderstandingClient != nil {
		cbStates["query_understanding"] = h.QueryUnderstandingClient.CircuitBreaker.Stats()
	}
	if h.LLMClient != nil {
		cbStates["llm_service"] = h.LLMClient.CircuitBreaker.Stats()
	}

	// Rate limiter stats
	rateLimitStats := map[string]interface{}{}
	if h.RateLimiter != nil {
		rateLimitStats = h.RateLimiter.Stats()
	}

	cacheHitRate := float64(0)
	if hr, ok := cacheStats["hit_rate"].(float64); ok {
		cacheHitRate = hr
	}

	response := types.StatsResponse{
		RequestCount:    totalRequests,
		ErrorCount:      totalErrors,
		AverageLatency:  avgLatency.String(),
		CacheHitRate:    cacheHitRate,
		CircuitBreakers: cbStates,
		RateLimiting:    rateLimitStats,
		Uptime:          time.Since(h.StartTime).Round(time.Second).String(),
	}

	writeJSON(w, http.StatusOK, response)
}
