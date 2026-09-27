package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"query-understanding-service/classifier"
	"query-understanding-service/parser"
	"query-understanding-service/prompts"
	"query-understanding-service/rewriter"
	"query-understanding-service/sanitizer"
)

// Services holds references to the query understanding services.
type Services struct {
	Rewrite  *rewriter.Service
	Parse    *parser.Service
	Classify *classifier.Service
}

// QueryHandler handles HTTP requests for the query understanding endpoints.
type QueryHandler struct {
	services *Services
	logger   *log.Logger
}

// NewQueryHandler creates a new query handler.
func NewQueryHandler(services *Services, logger *log.Logger) *QueryHandler {
	return &QueryHandler{
		services: services,
		logger:   logger,
	}
}

// RegisterRoutes registers all handler endpoints on the mux.
func (h *QueryHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /rewrite", h.handleRewrite)
	mux.HandleFunc("POST /parse-filters", h.handleParseFilters)
	mux.HandleFunc("POST /classify-intent", h.handleClassifyIntent)
	mux.HandleFunc("GET /health", h.handleHealth)
	mux.HandleFunc("GET /ready", h.handleReady)
}

// --- Rewrite Handler ---

type rewriteRequest struct {
	Query   string                   `json:"query"`
	History []map[string]interface{} `json:"history,omitempty"`
}

type historyEntry struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type rewriteDataResponse struct {
	Data *rewriter.RewriteResponse `json:"data"`
}

func (h *QueryHandler) handleRewrite(w http.ResponseWriter, r *http.Request) {
	var req rewriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body", err.Error())
		return
	}

	// Validate query
	if err := sanitizer.SanitizeQuery(req.Query); err != nil {
		writeError(w, http.StatusBadRequest, "invalid query", err.Error())
		return
	}

	// Sanitize query for prompt inclusion
	safeQuery := sanitizer.SanitizeForPrompt(req.Query)

	// Validate and convert history
	var history []prompts.HistoryMessage
	if len(req.History) > 0 {
		// Convert to map format for validation
		historyMaps := make([]map[string]string, 0, len(req.History))
		for _, entry := range req.History {
			if roleStr, ok := entry["role"].(string); ok {
				if contentStr, ok := entry["content"].(string); ok {
					historyMaps = append(historyMaps, map[string]string{
						"role":    roleStr,
						"content": contentStr,
					})
				}
			}
		}

		if err := sanitizer.ValidateHistory(historyMaps); err != nil {
			writeError(w, http.StatusBadRequest, "invalid conversation history", err.Error())
			return
		}

		for _, entry := range historyMaps {
			history = append(history, prompts.HistoryMessage{
				Role:    entry["role"],
				Content: sanitizer.SanitizeForPrompt(entry["content"]),
			})
		}
	}

	rewriteReq := &rewriter.RewriteRequest{
		Query:   safeQuery,
		History: history,
	}

	result, err := h.services.Rewrite.Rewrite(r.Context(), rewriteReq)
	if err != nil {
		h.logger.Printf("rewrite error: %v", err)
		writeError(w, http.StatusInternalServerError, "rewrite failed", "internal error occurred")
		return
	}

	writeJSON(w, http.StatusOK, rewriteDataResponse{Data: result})
}

// --- Parse Filters Handler ---

type parseFiltersRequest struct {
	Query string `json:"query"`
}

type parseFiltersDataResponse struct {
	Data *parser.ParseResponse `json:"data"`
}

func (h *QueryHandler) handleParseFilters(w http.ResponseWriter, r *http.Request) {
	var req parseFiltersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body", err.Error())
		return
	}

	// Validate query
	if err := sanitizer.SanitizeQuery(req.Query); err != nil {
		writeError(w, http.StatusBadRequest, "invalid query", err.Error())
		return
	}

	// Sanitize query for prompt inclusion
	safeQuery := sanitizer.SanitizeForPrompt(req.Query)

	parseReq := &parser.ParseRequest{
		Query: safeQuery,
	}

	result, err := h.services.Parse.Parse(r.Context(), parseReq)
	if err != nil {
		h.logger.Printf("parse filters error: %v", err)
		writeError(w, http.StatusInternalServerError, "filter parsing failed", "internal error occurred")
		return
	}

	writeJSON(w, http.StatusOK, parseFiltersDataResponse{Data: result})
}

// --- Classify Intent Handler ---

type classifyIntentRequest struct {
	Query string `json:"query"`
}

type classifyIntentDataResponse struct {
	Data *classifier.ClassifyResponse `json:"data"`
}

func (h *QueryHandler) handleClassifyIntent(w http.ResponseWriter, r *http.Request) {
	var req classifyIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body", err.Error())
		return
	}

	// Validate query
	if err := sanitizer.SanitizeQuery(req.Query); err != nil {
		writeError(w, http.StatusBadRequest, "invalid query", err.Error())
		return
	}

	// Sanitize query for prompt inclusion
	safeQuery := sanitizer.SanitizeForPrompt(req.Query)

	classifyReq := &classifier.ClassifyRequest{
		Query: safeQuery,
	}

	result, err := h.services.Classify.Classify(r.Context(), classifyReq)
	if err != nil {
		h.logger.Printf("classify intent error: %v", err)
		writeError(w, http.StatusInternalServerError, "intent classification failed", "internal error occurred")
		return
	}

	// Include needs_clarification flag in response
	needsClarification := result.Confidence < 0.7

	response := map[string]interface{}{
		"data":             result,
		"needs_clarification": needsClarification,
	}

	writeJSON(w, http.StatusOK, response)
}

// --- Health Check Handler ---

type healthDataResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

func (h *QueryHandler) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, healthDataResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "v0.1.0",
	})
}

// --- Readiness Check Handler ---

type readyDataResponse struct {
	Ready    bool   `json:"ready"`
	Services string `json:"services"`
}

func (h *QueryHandler) handleReady(w http.ResponseWriter, r *http.Request) {
	ready := h.services.Rewrite != nil && h.services.Parse != nil && h.services.Classify != nil
	services := "all"
	if !ready {
		services = "degraded"
	}

	writeJSON(w, http.StatusOK, readyDataResponse{
		Ready:    ready,
		Services: services,
	})
}

// --- Helper Functions ---

type errorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string, details string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	resp := errorResponse{
		Error: message,
	}
	if details != "" {
		resp.Details = details
	}
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("failed to encode error response: %v", err)
	}
}
