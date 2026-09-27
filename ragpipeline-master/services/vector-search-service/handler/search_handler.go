// Package handler provides HTTP handlers for the vector search service.
package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/visionary/ragpipeline/services/vector-search-service/cache"
	"github.com/visionary/ragpipeline/services/vector-search-service/config"
	"github.com/visionary/ragpipeline/services/vector-search-service/db"
	"github.com/visionary/ragpipeline/services/vector-search-service/search"
)

// SearchHandlerDeps holds the dependencies for SearchHandler.
type SearchHandlerDeps struct {
	HybridSearcher *search.HybridSearcher
	Config         *config.Config
	Store          *db.Store
	SearchCache    *cache.SearchCache // optional
}

// SearchHandler handles vector search HTTP requests.
type SearchHandler struct {
	hybridSearcher *search.HybridSearcher
	cfg            *config.Config
	store          *db.Store
	searchCache    *cache.SearchCache
}

// NewSearchHandler creates a new search handler.
func NewSearchHandler(deps SearchHandlerDeps) *SearchHandler {
	return &SearchHandler{
		hybridSearcher: deps.HybridSearcher,
		cfg:            deps.Config,
		store:          deps.Store,
		searchCache:    deps.SearchCache,
	}
}

// SearchRequest represents the incoming search request body.
type SearchRequest struct {
	Embedding []float32       `json:"embedding"`
	Keywords  []string        `json:"keywords"`
	TopK      int             `json:"top_k"`
	Filters   *FilterRequest  `json:"filters"`
}

// FilterRequest represents metadata filtering criteria in the request.
type FilterRequest struct {
	Grade       []string `json:"grade"`
	Subject     []string `json:"subject"`
	ContentType []string `json:"content_type"`
	Chapter     []string `json:"chapter"`
}

// ToSearchQuery converts the HTTP request to a database search query.
func (r *SearchRequest) ToSearchQuery() *db.SearchQuery {
	query := &db.SearchQuery{
		Embedding: r.Embedding,
		Keywords:  r.Keywords,
		TopK:      r.TopK,
	}

	if r.Filters != nil {
		query.Filters = &db.MetadataFilter{
			Grades:       r.Filters.Grade,
			Subjects:     r.Filters.Subject,
			ContentTypes: r.Filters.ContentType,
			Chapters:     r.Filters.Chapter,
		}
	}

	return query
}

// Search handles POST /search - Hybrid vector search with embedding + keywords.
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := getRequestID(r)

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", requestID)
		return
	}

	// Set default top_k if not provided
	if req.TopK < 1 {
		req.TopK = h.cfg.DefaultTopK
	}

	// Validate top_k
	if req.TopK > h.cfg.MaxTopK {
		respondError(w, http.StatusBadRequest, "invalid_top_k", "top_k exceeds maximum allowed value ("+strconv.Itoa(h.cfg.MaxTopK)+")", requestID)
		return
	}

	// Validate that at least one search criteria is provided
	if len(req.Embedding) == 0 && len(req.Keywords) == 0 {
		respondError(w, http.StatusBadRequest, "missing_criteria", "At least one of 'embedding' or 'keywords' must be provided", requestID)
		return
	}

	// Validate embedding dimension if provided
	if len(req.Embedding) > 0 && len(req.Embedding) != h.cfg.EmbeddingDim {
		respondError(w, http.StatusBadRequest, "invalid_embedding", "Embedding dimension mismatch: expected "+strconv.Itoa(h.cfg.EmbeddingDim)+", got "+strconv.Itoa(len(req.Embedding)), requestID)
		return
	}

	// Check search cache (only for dense searches with embeddings)
	if h.searchCache != nil && len(req.Embedding) > 0 {
		// Convert float32 embedding to float64 for cache key
		embedding64 := make([]float64, len(req.Embedding))
		for i, v := range req.Embedding {
			embedding64[i] = float64(v)
		}

		filters := make(map[string][]string, 4) // grade, subject, content_type, chapter
		if req.Filters != nil {
			if len(req.Filters.Grade) > 0 {
				filters["grade"] = req.Filters.Grade
			}
			if len(req.Filters.Subject) > 0 {
				filters["subject"] = req.Filters.Subject
			}
			if len(req.Filters.ContentType) > 0 {
				filters["content_type"] = req.Filters.ContentType
			}
			if len(req.Filters.Chapter) > 0 {
				filters["chapter"] = req.Filters.Chapter
			}
		}

		cached, err := h.searchCache.Get(r.Context(), embedding64, req.TopK, filters)
		if err == nil && cached != nil {
			// Cache hit - return cached results
			resp := map[string]interface{}{
				"data": map[string]interface{}{
					"results":        cached.Chunks,
					"total_results":  len(cached.ChunkIDs),
					"search_time_ms": 0,
					"cache_hit":      true,
				},
			}

			w.Header().Set("Cache-Control", "public, max-age=60")
			w.Header().Set("X-Request-ID", requestID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)

			log.Info().
				Str("request_id", requestID).
				Int("results", len(cached.ChunkIDs)).
				Int64("duration_ms", time.Since(start).Milliseconds()).
				Bool("cache_hit", true).
				Msg("Search request completed (cache hit)")
			return
		}
	}

	// Execute search
	query := req.ToSearchQuery()
	response, err := h.hybridSearcher.Search(r.Context(), query)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Search failed")

		if isTimeout(err) {
			respondError(w, http.StatusGatewayTimeout, "search_timeout", "Search query timed out", requestID)
			return
		}

		respondError(w, http.StatusInternalServerError, "search_error", "Search operation failed", requestID)
		return
	}

	// Cache the results (only for dense searches)
	if h.searchCache != nil && len(req.Embedding) > 0 {
		embedding64 := make([]float64, len(req.Embedding))
		for i, v := range req.Embedding {
			embedding64[i] = float64(v)
		}

		filters := make(map[string][]string, 4) // grade, subject, content_type, chapter
		if req.Filters != nil {
			if len(req.Filters.Grade) > 0 {
				filters["grade"] = req.Filters.Grade
			}
			if len(req.Filters.Subject) > 0 {
				filters["subject"] = req.Filters.Subject
			}
			if len(req.Filters.ContentType) > 0 {
				filters["content_type"] = req.Filters.ContentType
			}
			if len(req.Filters.Chapter) > 0 {
				filters["chapter"] = req.Filters.Chapter
			}
		}

		chunks := make([]cache.SearchResultChunk, 0, len(response.Results))
		chunkIDs := make([]string, 0, len(response.Results))
		scores := make([]float64, 0, len(response.Results))

		for _, r := range response.Results {
			chunkIDs = append(chunkIDs, r.ParentID)
			score := 0.0
			if r.RRFScore > 0 {
				score = r.RRFScore
			} else if r.DenseScore != nil {
				score = *r.DenseScore
			}
			scores = append(scores, score)

			chunks = append(chunks, cache.SearchResultChunk{
				ParentID: r.ParentID,
				Content:  r.Content,
				Score:    score,
				Metadata: r.Metadata,
			})
		}

		cachedResult := &cache.CachedSearchResult{
			ChunkIDs:  chunkIDs,
			Scores:    scores,
			Chunks:    chunks,
			TopK:      req.TopK,
			Filters:   "",
			CreatedAt: time.Now(),
		}

		// Cache asynchronously with detached background context
		// Using r.Context() here is unsafe because the request context
		// gets cancelled after the response is flushed to the client.
		go func() {
			cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := h.searchCache.Set(cacheCtx, embedding64, req.TopK, filters, cachedResult); err != nil {
				log.Debug().Err(err).Msg("Failed to cache search results")
			}
		}()
	}

	// Build response
	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"results":        response.Results,
			"total_results":  response.TotalResults,
			"search_time_ms": response.SearchTimeMS,
		},
	}

	// Include stats if enabled
	if h.cfg.EnableQueryMetrics && response.Stats != nil {
		resp["data"].(map[string]interface{})["stats"] = response.Stats
	}

	// Set cache headers for search results
	w.Header().Set("Cache-Control", "public, max-age=60") // Cache for 60 seconds
	w.Header().Set("X-Request-ID", requestID)
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)

	log.Info().
		Str("request_id", requestID).
		Int("results", response.TotalResults).
		Int64("duration_ms", time.Since(start).Milliseconds()).
		Msg("Search request completed")
}

// DenseSearch handles POST /search/dense - Pure vector similarity search.
func (h *SearchHandler) DenseSearch(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := getRequestID(r)

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", requestID)
		return
	}

	// Validate embedding is provided
	if len(req.Embedding) == 0 {
		respondError(w, http.StatusBadRequest, "missing_embedding", "'embedding' is required for dense search", requestID)
		return
	}

	// Validate embedding dimension
	if len(req.Embedding) != h.cfg.EmbeddingDim {
		respondError(w, http.StatusBadRequest, "invalid_embedding", "Embedding dimension mismatch: expected "+strconv.Itoa(h.cfg.EmbeddingDim)+", got "+strconv.Itoa(len(req.Embedding)), requestID)
		return
	}

	// Set default top_k if not provided
	if req.TopK < 1 {
		req.TopK = h.cfg.DefaultTopK
	}

	// Execute dense-only search
	query := req.ToSearchQuery()
	response, err := h.hybridSearcher.DenseOnly(r.Context(), query)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Dense search failed")

		if isTimeout(err) {
			respondError(w, http.StatusGatewayTimeout, "search_timeout", "Search query timed out", requestID)
			return
		}

		respondError(w, http.StatusInternalServerError, "search_error", "Dense search operation failed", requestID)
		return
	}

	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"results":        response.Results,
			"total_results":  response.TotalResults,
			"search_time_ms": response.SearchTimeMS,
		},
	}

	w.Header().Set("Cache-Control", "public, max-age=60")
	w.Header().Set("X-Request-ID", requestID)
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)

	log.Info().
		Str("request_id", requestID).
		Int("results", response.TotalResults).
		Int64("duration_ms", time.Since(start).Milliseconds()).
		Msg("Dense search request completed")
}

// SparseSearch handles POST /search/sparse - Pure keyword search.
func (h *SearchHandler) SparseSearch(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := getRequestID(r)

	var req SearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid_request", "Request body must be valid JSON", requestID)
		return
	}

	// Validate keywords are provided
	if len(req.Keywords) == 0 {
		respondError(w, http.StatusBadRequest, "missing_keywords", "'keywords' is required for sparse search", requestID)
		return
	}

	// Set default top_k if not provided
	if req.TopK < 1 {
		req.TopK = h.cfg.DefaultTopK
	}

	// Execute sparse-only search
	query := req.ToSearchQuery()
	response, err := h.hybridSearcher.SparseOnly(r.Context(), query)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Sparse search failed")

		if isTimeout(err) {
			respondError(w, http.StatusGatewayTimeout, "search_timeout", "Search query timed out", requestID)
			return
		}

		respondError(w, http.StatusInternalServerError, "search_error", "Sparse search operation failed", requestID)
		return
	}

	resp := map[string]interface{}{
		"data": map[string]interface{}{
			"results":        response.Results,
			"total_results":  response.TotalResults,
			"search_time_ms": response.SearchTimeMS,
		},
	}

	w.Header().Set("Cache-Control", "public, max-age=60")
	w.Header().Set("X-Request-ID", requestID)
	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)

	log.Info().
		Str("request_id", requestID).
		Int("results", response.TotalResults).
		Int64("duration_ms", time.Since(start).Milliseconds()).
		Msg("Sparse search request completed")
}

// Health handles GET /health - Health check endpoint.
func (h *SearchHandler) Health(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"service":   "vector-search-service",
		"version":   Version,
	}

	// Check database connectivity
	ctx := r.Context()
	if err := h.store.Ping(ctx); err != nil {
		health["status"] = "unhealthy"
		health["database"] = "disconnected"
		health["error"] = err.Error()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(health)
		return
	}

	health["database"] = "connected"
	health["pool_stats"] = h.store.Stats()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}

// Version is set via ldflags at build time.
var Version = "dev"

// Stats handles GET /stats - Search statistics and metrics.
func (h *SearchHandler) Stats(w http.ResponseWriter, r *http.Request) {
	stats := map[string]interface{}{
		"service": "vector-search-service",
		"database": h.store.Stats(),
		"config": map[string]interface{}{
			"default_top_k":          h.cfg.DefaultTopK,
			"max_top_k":              h.cfg.MaxTopK,
			"query_timeout":          h.cfg.QueryTimeout.String(),
			"rrf_k":                  h.cfg.RRFK,
			"embedding_dim":          h.cfg.EmbeddingDim,
			"similarity_threshold":   h.cfg.SimilarityThreshold,
			"enable_explain_analyze": h.cfg.EnableExplainAnalyze,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if h.searchCache != nil {
		stats["search_cache"] = h.searchCache.Stats()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stats)
}

// NotFound handles unmatched routes.
func (h *SearchHandler) NotFound(w http.ResponseWriter, r *http.Request) {
	respondError(w, http.StatusNotFound, "not_found", "The requested resource was not found", "")
}

// MethodNotAllowed handles disallowed HTTP methods.
func (h *SearchHandler) MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	respondError(w, http.StatusMethodNotAllowed, "method_not_allowed", "The HTTP method is not allowed for this endpoint", "")
}

// respondError sends a standardized error response.
func respondError(w http.ResponseWriter, statusCode int, code, message, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	if requestID != "" {
		w.Header().Set("X-Request-ID", requestID)
	}
	w.WriteHeader(statusCode)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
}

// getRequestID extracts or generates a request ID.
func getRequestID(r *http.Request) string {
	requestID := r.Header.Get("X-Request-ID")
	if requestID == "" {
		requestID = uuid.New().String()
	}
	return requestID
}

// isTimeout checks if the error is a timeout.
func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "timeout") || strings.Contains(msg, "context deadline exceeded") || strings.Contains(msg, "canceling statement due to user request")
}
