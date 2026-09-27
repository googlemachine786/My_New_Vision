// Package handler provides HTTP handlers for the API Gateway.
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/visionary/ragpipeline/pkg/types"
	"github.com/visionary/ragpipeline/services/api-gateway/client"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	// ReexplainSessionKeyPrefix is the Redis key prefix for re-explain session tracking
	ReexplainSessionKeyPrefix = "rag:reexplain:session:"
	// MaxReexplainAttempts is the maximum number of re-explain attempts per query (Story D)
	MaxReexplainAttempts = 3
	// ReexplainTTL is the TTL for re-explain session data (24 hours)
	ReexplainTTL = 24 * time.Hour
)

// ReexplainHandler handles re-explain concept differently requests (Story D)
type ReexplainHandler struct {
	LLMClient   *client.ServiceClient
	RedisClient *redis.Client
}

// NewReexplainHandler creates a new re-explain handler
func NewReexplainHandler(llmClient *client.ServiceClient, redisClient *redis.Client) *ReexplainHandler {
	return &ReexplainHandler{
		LLMClient:   llmClient,
		RedisClient: redisClient,
	}
}

// ServeHTTP handles POST /reexplain requests
func (h *ReexplainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "method_not_allowed", Message: "Only POST is allowed"},
		})
		return
	}

	var req types.ReexplainRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "invalid_request", Message: "Invalid JSON body"},
		})
		return
	}

	if err := h.validateRequest(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "validation_error", Message: err.Error()},
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Get current attempt count from session
	attemptNumber, err := h.getReexplainCount(ctx, req.SessionID, req.QueryID)
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get reexplain count, starting from 1")
		attemptNumber = 0
	}

	// Check if max attempts exceeded
	if attemptNumber >= MaxReexplainAttempts {
		// Trigger fallback strategy after 3 failures
		log.Info().
			Str("query_id", req.QueryID).
			Int("attempts", attemptNumber).
			Msg("Max re-explain attempts reached, triggering fallback")

		wasFallback := true
		writeJSON(w, http.StatusOK, types.ReexplainResponse{
			ReexplainedAnswer: "I see this is still unclear. Let me suggest: review the textbook chapter, ask your teacher for help, or try a related question. I'm here if you want to explore a different angle!",
			StrategyUsed:      "fallback_after_max_attempts",
			AttemptNumber:     attemptNumber,
			MaxAttempts:       MaxReexplainAttempts,
			WasFallback:       &wasFallback,
		})
		return
	}

	// Increment attempt counter
	attemptNumber++
	if err := h.incrementReexplainCount(ctx, req.SessionID, req.QueryID, attemptNumber); err != nil {
		log.Warn().Err(err).Msg("Failed to increment reexplain count")
	}

	// Select the next template strategy (cycle through different strategies)
	strategyIndex := (attemptNumber - 1) % 5 // 5 templates available
	strategy := h.selectStrategy(strategyIndex)

	// Generate the re-explained response
	answer, err := h.generateReexplain(ctx, req.OriginalQuery, req.PreviousResponse, strategy)
	if err != nil {
		log.Error().Err(err).Msg("Failed to generate reexplain response")
		writeJSON(w, http.StatusBadGateway, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "reexplain_failed", Message: "Failed to generate re-explained response"},
		})
		return
	}

	resp := types.ReexplainResponse{
		ReexplainedAnswer: answer,
		StrategyUsed:      string(strategy),
		AttemptNumber:     attemptNumber,
		MaxAttempts:       MaxReexplainAttempts,
	}

	log.Info().
		Str("query_id", req.QueryID).
		Int("attempt", attemptNumber).
		Str("strategy", string(strategy)).
		Msg("Re-explain response generated")

	writeJSON(w, http.StatusOK, resp)
}

// validateRequest validates the re-explain request
func (h *ReexplainHandler) validateRequest(req *types.ReexplainRequest) error {
	if req.QueryID == "" {
		return fmt.Errorf("query_id is required")
	}
	if req.OriginalQuery == "" {
		return fmt.Errorf("original_query is required")
	}
	if req.PreviousResponse == "" {
		return fmt.Errorf("previous_response is required")
	}
	return nil
}

// getReexplainCount retrieves the current re-explain attempt count for a session+query
func (h *ReexplainHandler) getReexplainCount(ctx context.Context, sessionID, queryID string) (int, error) {
	key := reexplainSessionKey(sessionID, queryID)
	count, err := h.RedisClient.Get(ctx, key).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get reexplain count: %w", err)
	}
	return count, nil
}

// incrementReexplainCount increments and stores the re-explain attempt count
func (h *ReexplainHandler) incrementReexplainCount(ctx context.Context, sessionID, queryID string, count int) error {
	key := reexplainSessionKey(sessionID, queryID)
	if err := h.RedisClient.Set(ctx, key, count, ReexplainTTL).Err(); err != nil {
		return fmt.Errorf("failed to set reexplain count: %w", err)
	}
	return nil
}

// selectStrategy selects the appropriate re-explain strategy based on attempt number
func (h *ReexplainHandler) selectStrategy(index int) types.ReexplainStrategy {
	// Import templates from query-understanding-service prompts
	// For now, use inline strategy names
	strategies := []types.ReexplainStrategy{
		"analogy_based",
		"step_by_step",
		"real_world_example",
		"simplified_language",
		"visual_description",
	}

	if index < 0 || index >= len(strategies) {
		return strategies[0]
	}
	return strategies[index]
}

// generateReexplain generates a re-explained response using the selected strategy
func (h *ReexplainHandler) generateReexplain(ctx context.Context, query, previousResponse string, strategy types.ReexplainStrategy) (string, error) {
	prompt := h.buildReexplainPrompt(query, previousResponse, strategy)

	type llmRequest struct {
		Query  string `json:"query"`
		Stream bool   `json:"stream"`
	}

	type llmResponse struct {
		Answer string `json:"answer"`
	}

	var result llmResponse
	body, err := h.LLMClient.PostJSONRaw(ctx, "/generate", llmRequest{
		Query:  prompt,
		Stream: false,
	})
	if err != nil {
		return "", fmt.Errorf("reexplain LLM request failed: %w", err)
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to decode reexplain LLM response: %w", err)
	}

	return result.Answer, nil
}

// buildReexplainPrompt constructs the prompt for re-explaining
func (h *ReexplainHandler) buildReexplainPrompt(query, previousResponse string, strategy types.ReexplainStrategy) string {
	// Build strategy-specific prompt
	switch strategy {
	case "analogy_based":
		return fmt.Sprintf(`You are an expert educator. The student didn't fully understand the previous explanation, so please re-explain the concept using a real-world analogy.

ORIGINAL QUESTION: %s

PREVIOUS EXPLANATION: %s

INSTRUCTIONS:
1. Identify the core concept(s) from the previous explanation
2. Create a relatable analogy from everyday life (sports, cooking, travel, etc.)
3. Map each part of the analogy back to the actual concept
4. Keep the analogy simple and memorable
5. After the analogy, briefly restate the key takeaway`, query, previousResponse)

	case "step_by_step":
		return fmt.Sprintf(`You are an expert educator. The student needs a more structured explanation. Please re-explain the concept by breaking it down into clear, numbered steps.

ORIGINAL QUESTION: %s

PREVIOUS EXPLANATION: %s

INSTRUCTIONS:
1. Identify the key components of the concept
2. Break them into 3-5 logical, sequential steps
3. Explain each step clearly with a brief header
4. Show how the steps connect to form the complete picture
5. Use bullet points for clarity`, query, previousResponse)

	case "real_world_example":
		return fmt.Sprintf(`You are an expert educator. The student learns better through concrete examples. Please re-explain using a different real-world example.

ORIGINAL QUESTION: %s

PREVIOUS EXPLANATION: %s

INSTRUCTIONS:
1. Identify the core concept
2. Choose a DIFFERENT domain for the example
3. Walk through the example step by step
4. Explicitly connect the example back to the abstract concept
5. Make it practical and memorable`, query, previousResponse)

	case "simplified_language":
		return fmt.Sprintf(`You are an expert educator. The student needs a simpler explanation. Please re-explain using simpler language, as if teaching a younger student.

ORIGINAL QUESTION: %s

PREVIOUS EXPLANATION: %s

INSTRUCTIONS:
1. Use shorter sentences (max 15-20 words each)
2. Avoid jargon - if you must use technical terms, define them immediately
3. Use "you" and "your" to make it conversational
4. Focus on the ONE most important idea
5. Use everyday words instead of academic language`, query, previousResponse)

	case "visual_description":
		return fmt.Sprintf(`You are an expert educator. The student learns better through visualization. Please re-explain by helping them picture the concept.

ORIGINAL QUESTION: %s

PREVIOUS EXPLANATION: %s

INSTRUCTIONS:
1. Start with "Imagine..." or "Picture this..."
2. Create a vivid mental image with sensory details
3. Walk through the visualization step by step
4. Connect what they're "seeing" to the actual concept
5. Use descriptive, colorful language`, query, previousResponse)

	default:
		return fmt.Sprintf(`Please re-explain this concept differently:

QUESTION: %s

PREVIOUS: %s

Provide a clearer explanation.`, query, previousResponse)
	}
}

// reexplainSessionKey generates the Redis key for re-explain session tracking
func reexplainSessionKey(sessionID, queryID string) string {
	if sessionID == "" {
		return ReexplainSessionKeyPrefix + queryID
	}
	return ReexplainSessionKeyPrefix + sessionID + ":" + queryID
}
