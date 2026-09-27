package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/visionary/ragpipeline/pkg/types"
	"github.com/visionary/ragpipeline/services/api-gateway/middleware"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	FeedbackKeyPrefix       = "rag:feedback:"
	FeedbackContextKeyPrefix = "rag:feedback:context:"
	FeedbackReviewKeyPrefix = "rag:feedback:review:"
	FeedbackTTL             = 90 * 24 * time.Hour // 90 days
)

// FeedbackHandler handles user feedback submission
type FeedbackHandler struct {
	redisClient *redis.Client
	aggregator  *FeedbackAggregator
}

// NewFeedbackHandler creates a new feedback handler
func NewFeedbackHandler(redisClient *redis.Client) *FeedbackHandler {
	return &FeedbackHandler{
		redisClient: redisClient,
		aggregator:  NewFeedbackAggregator(redisClient),
	}
}

// ServeHTTP handles POST /feedback
func (h *FeedbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "method_not_allowed", Message: "Only POST is allowed"},
		})
		return
	}

	// Parse request
	var req types.FeedbackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "invalid_request", Message: "Invalid JSON body"},
		})
		return
	}

	if err := req.Validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "validation_error", Message: err.Error()},
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Store feedback
	feedbackID := uuid.New().String()
	feedback := FeedbackRecord{
		ID:           feedbackID,
		Query:        req.Query,
		ResponseID:   req.ResponseID,
		SessionID:    req.SessionID,
		ThumbsUp:     req.ThumbsUp,
		Rating:       req.Rating,
		Comment:      req.Comment,
		FeedbackText: req.FeedbackText, // Story A: store feedback_text
		Grade:        req.Grade,
		Subject:      req.Subject,
		Timestamp:    time.Now().UTC(),
	}

	data, err := json.Marshal(feedback)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal feedback")
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "internal_error", Message: "Failed to process feedback"},
		})
		return
	}

	// Store in Redis
	key := feedbackKey(feedbackID)
	if err := h.redisClient.Set(ctx, key, data, FeedbackTTL).Err(); err != nil {
		log.Error().Err(err).Msg("Failed to store feedback in Redis")
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "internal_error", Message: "Failed to store feedback"},
		})
		return
	}

	// Add to feedback index for querying
	indexKey := feedbackIndexKey()
	if err := h.redisClient.ZAdd(ctx, indexKey, redis.Z{
		Score:  float64(feedback.Timestamp.Unix()),
		Member: key,
	}).Err(); err != nil {
		log.Warn().Err(err).Msg("Failed to add to feedback index")
		// Non-fatal - feedback is still stored
	}

	// If session ID provided, link feedback to session
	if req.SessionID != "" {
		sessionFeedbackKey := sessionFeedbackKey(req.SessionID)
		if err := h.redisClient.LPush(ctx, sessionFeedbackKey, feedbackID).Err(); err != nil {
			log.Warn().Err(err).Msg("Failed to link feedback to session")
		}
	}

	// Story A: Store full context (query + response + retrieved context) with feedback
	// This is done asynchronously to not block the response
	if req.ResponseID != "" {
		go h.storeFeedbackContext(ctx, feedbackID, req)
	}

	// Story A: If thumbs_down, add to review queue in Redis
	thumbsUp := req.ThumbsUp != nil && *req.ThumbsUp
	if !thumbsUp {
		go h.addToReviewQueue(ctx, feedbackID, req)
	}

	// Record aggregated feedback (non-blocking)
	go func() {
		aggCtx, aggCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer aggCancel()
		if err := h.aggregator.RecordFeedback(aggCtx, thumbsUp, req.Grade, req.Subject); err != nil {
			log.Warn().Err(err).Msg("Failed to record aggregated feedback")
		}
	}()

	// Record Prometheus metric (non-blocking)
	go middleware.RecordFeedback(thumbsUp)

	log.Info().
		Str("feedback_id", feedbackID).
		Str("query", req.Query).
		Bool("thumbs_up", thumbsUp).
		Str("feedback_text", req.FeedbackText).
		Msg("Feedback received")

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"id":             feedbackID,
		"query":          req.Query,
		"thumbs_up":      req.ThumbsUp,
		"rating":         req.Rating,
		"feedback_text":  req.FeedbackText,
		"timestamp":      feedback.Timestamp.Format(time.RFC3339),
	})
}

// storeFeedbackContext stores full query + response + retrieved context with feedback (Story A)
func (h *FeedbackHandler) storeFeedbackContext(ctx context.Context, feedbackID string, req types.FeedbackRequest) {
	contextKey := feedbackContextKey(feedbackID)

	feedbackContext := types.FeedbackContext{
		FeedbackID: feedbackID,
		Query:      req.Query,
		Grade:      req.Grade,
		Subject:    req.Subject,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
	}

	// Note: In a full implementation, we would fetch the response and retrieved context
	// from the response cache or database using ResponseID. For now, we store what we have.
	data, err := json.Marshal(feedbackContext)
	if err != nil {
		log.Error().Err(err).Str("feedback_id", feedbackID).Msg("Failed to marshal feedback context")
		return
	}

	if err := h.redisClient.Set(ctx, contextKey, data, FeedbackTTL).Err(); err != nil {
		log.Error().Err(err).Str("feedback_id", feedbackID).Msg("Failed to store feedback context")
	}
}

// addToReviewQueue adds a thumbs_down feedback to the Redis review queue (Story A)
func (h *FeedbackHandler) addToReviewQueue(ctx context.Context, feedbackID string, req types.FeedbackRequest) {
	reviewKey := feedbackReviewKey(time.Now().UTC().Format("2006-01-02"))

	reviewEntry := map[string]string{
		"feedback_id":   feedbackID,
		"query":         req.Query,
		"feedback_text": req.FeedbackText,
		"comment":       req.Comment,
		"grade":         req.Grade,
		"subject":       req.Subject,
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(reviewEntry)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal review entry")
		return
	}

	// Add to sorted set ordered by timestamp for chronological review
	if err := h.redisClient.ZAdd(ctx, reviewKey, redis.Z{
		Score:  float64(time.Now().UTC().Unix()),
		Member: data,
	}).Err(); err != nil {
		log.Error().Err(err).Msg("Failed to add to review queue")
	}

	// Set TTL of 30 days on review queue
	h.redisClient.Expire(ctx, reviewKey, 30*24*time.Hour)
}

// FeedbackRecord represents stored feedback
type FeedbackRecord struct {
	ID           string    `json:"id"`
	Query        string    `json:"query"`
	ResponseID   string    `json:"response_id,omitempty"`
	SessionID    string    `json:"session_id,omitempty"`
	ThumbsUp     *bool     `json:"thumbs_up,omitempty"`
	Rating       string    `json:"rating,omitempty"`
	Comment      string    `json:"comment,omitempty"`
	FeedbackText string    `json:"feedback_text,omitempty"` // Story A: "What was wrong?" input
	Grade        string    `json:"grade,omitempty"`
	Subject      string    `json:"subject,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

func feedbackKey(id string) string {
	return FeedbackKeyPrefix + id
}

func feedbackContextKey(id string) string {
	return FeedbackContextKeyPrefix + id
}

func feedbackReviewKey(date string) string {
	return FeedbackReviewKeyPrefix + date
}

func feedbackIndexKey() string {
	return "rag:feedback:index"
}

func sessionFeedbackKey(sessionID string) string {
	return "rag:feedback:session:" + sessionID
}
