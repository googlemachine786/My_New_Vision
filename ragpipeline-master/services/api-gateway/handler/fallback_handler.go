// Package handler provides HTTP handlers for the API Gateway.
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/visionary/ragpipeline/pkg/types"
	"github.com/visionary/ragpipeline/services/api-gateway/client"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
)

var (
	// fallbackRateCounter tracks the rate of fallback invocations (Story C)
	fallbackRateCounter = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "rag_fallback_rate_total",
			Help: "Total number of fallback LLM invocations when retrieval returned zero results",
		},
	)
)

// FallbackHandler handles fallback strategy when vector search returns zero results (Story C)
type FallbackHandler struct {
	LLMClient *client.ServiceClient
}

// NewFallbackHandler creates a new fallback handler
func NewFallbackHandler(llmClient *client.ServiceClient) *FallbackHandler {
	return &FallbackHandler{
		LLMClient: llmClient,
	}
}

// GenerateFallbackResponse generates a response using general LLM knowledge when retrieval fails
func (h *FallbackHandler) GenerateFallbackResponse(ctx context.Context, query, grade, subject string) (string, int, error) {
	// Increment fallback counter
	fallbackRateCounter.Inc()

	// Build general knowledge prompt with disclaimer
	prompt := h.buildFallbackPrompt(query, grade, subject)

	type llmRequest struct {
		Query  string `json:"query"`
		Stream bool   `json:"stream"`
	}

	type llmResponse struct {
		Answer      string `json:"answer"`
		TotalTokens int    `json:"total_tokens"`
	}

	var result llmResponse
	body, err := h.LLMClient.PostJSONRaw(ctx, "/generate", llmRequest{
		Query:  prompt,
		Stream: false,
	})
	if err != nil {
		return "", 0, fmt.Errorf("fallback LLM request failed: %w", err)
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", 0, fmt.Errorf("failed to decode fallback LLM response: %w", err)
	}

	log.Info().
		Str("query", query).
		Str("grade", grade).
		Str("subject", subject).
		Int("tokens", result.TotalTokens).
		Msg("Fallback response generated")

	return result.Answer, result.TotalTokens, nil
}

// buildFallbackPrompt constructs a system prompt for general knowledge responses
func (h *FallbackHandler) buildFallbackPrompt(query, grade, subject string) string {
	var sb strings.Builder

	sb.WriteString("You are an educational assistant helping students learn. ")

	if grade != "" {
		sb.WriteString(fmt.Sprintf("The student is in Grade %s. ", grade))
	}
	if subject != "" {
		sb.WriteString(fmt.Sprintf("The subject is %s. ", subject))
	}

	sb.WriteString("\n\n")
	sb.WriteString("IMPORTANT: I don't have specific curriculum data for this question, so I'm answering based on my general knowledge. ")
	sb.WriteString("Please note that this answer may not perfectly match your specific curriculum.\n\n")
	sb.WriteString(fmt.Sprintf("Question: %s\n\n", query))
	sb.WriteString("Please provide a clear, educational answer appropriate for the student's level. ")
	sb.WriteString("If you're uncertain about any part of the answer, mention that explicitly.")

	return sb.String()
}

// ShouldUseFallback determines if fallback should be triggered based on search results
func ShouldUseFallback(sources []types.Source) bool {
	// Trigger fallback when vector search returns zero results
	return len(sources) == 0
}

// ServeHTTP handles POST /fallback (direct fallback endpoint for testing)
func (h *FallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "method_not_allowed", Message: "Only POST is allowed"},
		})
		return
	}

	var req struct {
		Query   string `json:"query"`
		Grade   string `json:"grade,omitempty"`
		Subject string `json:"subject,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "invalid_request", Message: "Invalid JSON body"},
		})
		return
	}

	if req.Query == "" {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "validation_error", Message: "query is required"},
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	answer, tokens, err := h.GenerateFallbackResponse(ctx, req.Query, req.Grade, req.Subject)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate fallback response")
		writeJSON(w, http.StatusBadGateway, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "fallback_failed", Message: "Failed to generate fallback response"},
		})
		return
	}

	wasFallback := true
	writeJSON(w, http.StatusOK, types.QueryResponse{
		Answer:      answer,
		TotalTokens: tokens,
		WasFallback: &wasFallback,
	})
}

// RecordFallbackMetric records the fallback rate metric (non-blocking)
func RecordFallbackMetric() {
	go func() {
		fallbackRateCounter.Inc()
	}()
}
