package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/visionary/ragpipeline/services/api-gateway/client"
	"github.com/visionary/ragpipeline/pkg/types"

	"github.com/rs/zerolog/log"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	EmbeddingClient    *client.ServiceClient
	VectorSearchClient *client.ServiceClient
	QueryUnderstandingClient *client.ServiceClient
	RedisPinger       func(context.Context) error
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(
	embeddingClient, vectorSearchClient, queryUnderstandingClient *client.ServiceClient,
	redisPinger func(context.Context) error,
) *HealthHandler {
	return &HealthHandler{
		EmbeddingClient:    embeddingClient,
		VectorSearchClient: vectorSearchClient,
		QueryUnderstandingClient: queryUnderstandingClient,
		RedisPinger:        redisPinger,
	}
}

// ServeHTTP handles GET /health
func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "method_not_allowed", Message: "Only GET is allowed"},
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]string, 4) // embedding, vector, llm, query
	overallStatus := "healthy"

	// Check Embedding Service
	if h.checkServiceHealth(ctx, h.EmbeddingClient, "/health") {
		services["embedding_service"] = "healthy"
	} else {
		services["embedding_service"] = "unhealthy"
		overallStatus = "degraded"
	}

	// Check Vector Search Service
	if h.checkServiceHealth(ctx, h.VectorSearchClient, "/health") {
		services["vector_search"] = "healthy"
	} else {
		services["vector_search"] = "unhealthy"
		overallStatus = "degraded"
	}

	// Check Query Understanding Service
	if h.checkServiceHealth(ctx, h.QueryUnderstandingClient, "/health") {
		services["query_understanding"] = "healthy"
	} else {
		services["query_understanding"] = "unhealthy"
		overallStatus = "degraded"
	}

	// Check Redis
	if h.RedisPinger != nil {
		if err := h.RedisPinger(ctx); err != nil {
			services["redis"] = "unhealthy"
			// Redis being down is more serious
			if overallStatus == "healthy" {
				overallStatus = "degraded"
			}
		} else {
			services["redis"] = "healthy"
		}
	} else {
		services["redis"] = "unknown"
	}

	// If all services are unhealthy, set overall to unhealthy
	allUnhealthy := true
	for _, status := range services {
		if status == "healthy" {
			allUnhealthy = false
			break
		}
	}
	if allUnhealthy {
		overallStatus = "unhealthy"
	}

	response := types.HealthResponse{
		Status:    overallStatus,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Services:  services,
	}

	// Always return 200 OK with status field indicating health.
	// HTTP 206 is for byte-range requests, not health status.
	// Load balancers and health check probes understand 200 + status field.
	statusCode := http.StatusOK
	if overallStatus == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	writeJSON(w, statusCode, response)
}

func (h *HealthHandler) checkServiceHealth(ctx context.Context, svcClient *client.ServiceClient, path string) bool {
	if svcClient == nil {
		return false
	}

	resp, err := svcClient.Get(ctx, path)
	if err != nil {
		log.Debug().Err(err).Str("service", svcClient.Name()).Msg("Health check failed")
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Error().Err(err).Msg("Failed to encode JSON response")
	}
}
