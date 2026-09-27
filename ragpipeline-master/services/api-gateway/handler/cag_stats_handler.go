// Package handler provides HTTP handlers for CAG cache statistics.
package handler

import (
	"net/http"
	"time"

	"github.com/visionary/ragpipeline/services/api-gateway/cache"
	"github.com/visionary/ragpipeline/pkg/types"
)

// CAGStatsHandler handles the CAG cache statistics endpoint
type CAGStatsHandler struct {
	CAGOrchestrator *cache.CAGOrchestrator
	StartTime       time.Time
}

// NewCAGStatsHandler creates a new CAG stats handler
func NewCAGStatsHandler(cagOrchestrator *cache.CAGOrchestrator) *CAGStatsHandler {
	return &CAGStatsHandler{
		CAGOrchestrator: cagOrchestrator,
		StartTime:       time.Now(),
	}
}

// ServeHTTP handles GET /cag/stats
func (h *CAGStatsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "method_not_allowed", Message: "Only GET is allowed"},
		})
		return
	}

	response := map[string]interface{}{
		"uptime": time.Since(h.StartTime).Round(time.Second).String(),
	}

	if h.CAGOrchestrator != nil {
		summary := h.CAGOrchestrator.GetSummary()
		response["cag"] = summary

		// Add individual cache stats
		cacheDetails := make(map[string]interface{}, 5) // 5 cache layers

		if h.CAGOrchestrator.ExactCache != nil {
			cacheDetails["exact_match"] = h.CAGOrchestrator.ExactCache.Stats()
		}
		if h.CAGOrchestrator.SemanticCache != nil {
			cacheDetails["semantic"] = h.CAGOrchestrator.SemanticCache.Stats()
			cacheDetails["semantic_cost_saved"] = h.CAGOrchestrator.SemanticCache.CostSaved()
		}
		if h.CAGOrchestrator.EmbeddingCache != nil {
			cacheDetails["embedding"] = h.CAGOrchestrator.EmbeddingCache.Stats()
			cacheDetails["embedding_cost_saved"] = h.CAGOrchestrator.EmbeddingCache.CostSaved()
		}
		if h.CAGOrchestrator.TemplateCache != nil {
			cacheDetails["template"] = h.CAGOrchestrator.TemplateCache.Stats()
			cacheDetails["template_cost_saved"] = h.CAGOrchestrator.TemplateCache.CostSaved()
		}

		response["cache_details"] = cacheDetails
	}

	writeJSON(w, http.StatusOK, response)
}
