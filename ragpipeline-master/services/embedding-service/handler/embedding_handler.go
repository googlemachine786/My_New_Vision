// Package handler provides HTTP handlers for the embedding service API.
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/visionary/ragpipeline/services/embedding-service/cache"
	"github.com/visionary/ragpipeline/services/embedding-service/config"
	"github.com/visionary/ragpipeline/services/embedding-service/vertex"
)

// EmbeddingHandler handles embedding API requests.
type EmbeddingHandler struct {
	cfg             *config.Config
	client          *vertex.Client
	embeddingCache  *cache.EmbeddingCache
}

// NewEmbeddingHandler creates a new embedding handler.
func NewEmbeddingHandler(cfg *config.Config, client *vertex.Client, embeddingCache *cache.EmbeddingCache) *EmbeddingHandler {
	return &EmbeddingHandler{
		cfg:            cfg,
		client:         client,
		embeddingCache: embeddingCache,
	}
}

// EmbedRequest represents the JSON request body for embedding.
type EmbedRequest struct {
	Texts    []string `json:"texts"`
	TaskType string   `json:"task_type,omitempty"` // RETRIEVAL_DOCUMENT or RETRIEVAL_QUERY
}

// EmbedResponse represents the JSON response body for embedding.
type EmbedResponse struct {
	Embeddings [][]float64 `json:"embeddings"`
	Dimensions int         `json:"dimensions"`
}

// ErrorResponse represents a structured error response.
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error code and message.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// SuccessResponse wraps a successful response.
type SuccessResponse struct {
	Data interface{} `json:"data"`
}

// Embed handles POST /embed requests.
func (h *EmbeddingHandler) Embed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Only POST method is allowed")
		return
	}

	var req EmbedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_json", "Request body must be valid JSON")
		return
	}

	// Validate texts
	if len(req.Texts) == 0 {
		h.writeError(w, http.StatusBadRequest, "missing_texts", "At least one text is required")
		return
	}

	if len(req.Texts) > h.cfg.MaxTextsPerBatch {
		h.writeError(w, http.StatusBadRequest, "batch_too_large",
			fmt.Sprintf("Maximum %d texts per batch allowed", h.cfg.MaxTextsPerBatch))
		return
	}

	// Validate text lengths
	for i, text := range req.Texts {
		if len(text) == 0 {
			h.writeError(w, http.StatusBadRequest, "empty_text",
				fmt.Sprintf("Text at index %d is empty", i))
			return
		}
		if len(text) > h.cfg.MaxTextLength {
			h.writeError(w, http.StatusBadRequest, "text_too_long",
				fmt.Sprintf("Text at index %d exceeds maximum length of %d", i, h.cfg.MaxTextLength))
			return
		}
	}

	// Determine task type
	taskType := vertex.TaskTypeQuery
	if req.TaskType != "" {
		switch req.TaskType {
		case "RETRIEVAL_DOCUMENT":
			taskType = vertex.TaskTypeDocument
		case "RETRIEVAL_QUERY":
			taskType = vertex.TaskTypeQuery
		default:
			h.writeError(w, http.StatusBadRequest, "invalid_task_type",
				"Task type must be RETRIEVAL_DOCUMENT or RETRIEVAL_QUERY")
			return
		}
	}

	// Generate embeddings with caching
	result := EmbedResponse{
		Dimensions: h.cfg.VertexDimension,
		Embeddings: make([][]float64, 0, len(req.Texts)),
	}

	for _, text := range req.Texts {
		// Check cache first for single-text requests
		if h.embeddingCache != nil && len(req.Texts) == 1 {
			cached, err := h.embeddingCache.Get(r.Context(), text, req.TaskType)
			if err == nil && cached != nil {
				// Cache hit - skip Vertex AI call
				result.Embeddings = append(result.Embeddings, cached.Embedding)
				continue
			}
		}

		// Cache miss or batch - call Vertex AI
		var embedding *vertex.EmbedResponse
		var err error

		if len(req.Texts) == 1 {
			embedding, err = h.client.EmbedText(r.Context(), text, taskType)
		} else {
			var embeddings []*vertex.EmbedResponse
			embeddings, err = h.client.EmbedBatch(r.Context(), req.Texts, taskType)
			if err == nil {
				for i, emb := range embeddings {
					result.Embeddings = append(result.Embeddings, emb.Embedding)

					// Cache each embedding
					if h.embeddingCache != nil {
						h.embeddingCache.Set(r.Context(), req.Texts[i], req.TaskType, &cache.CachedEmbedding{
							Embedding:  emb.Embedding,
							Dimensions: h.cfg.VertexDimension,
							TaskType:   req.TaskType,
							Text:       req.Texts[i],
							CreatedAt:  time.Now(),
						})
					}
				}
			}
			if err != nil {
				h.writeError(w, http.StatusBadGateway, "vertex_ai_error",
					"Failed to generate embeddings: "+err.Error())
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(SuccessResponse{Data: result})
			return
		}

		if err != nil {
			h.writeError(w, http.StatusBadGateway, "vertex_ai_error",
				"Failed to generate embedding: "+err.Error())
			return
		}

		// Cache the result
		if h.embeddingCache != nil {
			h.embeddingCache.Set(r.Context(), text, req.TaskType, &cache.CachedEmbedding{
				Embedding:  embedding.Embedding,
				Dimensions: h.cfg.VertexDimension,
				TaskType:   req.TaskType,
				Text:       text,
				CreatedAt:  time.Now(),
			})
		}

		result.Embeddings = append(result.Embeddings, embedding.Embedding)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(SuccessResponse{Data: result})
}

// Health handles GET /health requests.
func (h *EmbeddingHandler) Health(w http.ResponseWriter, r *http.Request) {
	type HealthStatus struct {
		Status    string `json:"status"`
		Timestamp string `json:"timestamp"`
		Service   string `json:"service"`
		VertexAI  string `json:"vertex_ai"`
	}

	status := HealthStatus{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Service:   "embedding-service",
		VertexAI:  "unknown",
	}

	// Check Vertex AI connectivity
	if err := h.client.HealthCheck(r.Context()); err != nil {
		status.VertexAI = "unhealthy"
		status.Status = "degraded"
	} else {
		status.VertexAI = "healthy"
	}

	w.Header().Set("Content-Type", "application/json")
	if status.Status != "healthy" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(status)
}

// Stats handles GET /stats requests.
func (h *EmbeddingHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"service": "embedding-service",
		"config": map[string]interface{}{
			"model":               h.cfg.VertexModel,
			"dimension":           h.cfg.VertexDimension,
			"max_texts_per_batch": h.cfg.MaxTextsPerBatch,
			"max_text_length":     h.cfg.MaxTextLength,
			"max_retries":         h.cfg.MaxRetries,
		},
		"vertex": h.client.Stats(),
	}

	if h.embeddingCache != nil {
		stats["embedding_cache"] = h.embeddingCache.Stats()
		stats["embedding_cache_cost_saved"] = h.embeddingCache.CostSaved()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// NotFound handles 404 responses.
func (h *EmbeddingHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotFound, "not_found", "The requested resource was not found")
}

// MethodNotAllowed handles 405 responses.
func (h *EmbeddingHandler) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "The HTTP method is not allowed for this endpoint")
}

// writeError writes a structured JSON error response.
func (h *EmbeddingHandler) writeError(w http.ResponseWriter, statusCode int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}
