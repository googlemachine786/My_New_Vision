package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	// FeedbackAggKeyPrefix is the Redis key prefix for aggregated feedback
	FeedbackAggKeyPrefix = "rag:feedback:agg:"
	// FeedbackStatsTTL is the TTL for feedback stats
	FeedbackStatsTTL = 30 * 24 * time.Hour // 30 days
	// FeedbackReviewTTL is the TTL for feedback review entries
	FeedbackReviewTTL = 30 * 24 * time.Hour // 30 days
)

// FeedbackStats represents aggregated feedback statistics
type FeedbackStats struct {
	ThumbsUp       int64             `json:"thumbs_up"`
	ThumbsDown     int64             `json:"thumbs_down"`
	AccuracyRate   float64           `json:"accuracy_rate"`
	TotalFeedback  int64             `json:"total_feedback"`
	ByGrade        map[string]*Grade `json:"by_grade,omitempty"`
	BySubject      map[string]*Subject `json:"by_subject,omitempty"`
	Trending       []TrendingPoint   `json:"trending,omitempty"`
}

// Grade represents feedback stats for a specific grade
type Grade struct {
	ThumbsUp     int64   `json:"thumbs_up"`
	ThumbsDown   int64   `json:"thumbs_down"`
	AccuracyRate float64 `json:"accuracy_rate"`
}

// Subject represents feedback stats for a specific subject
type Subject struct {
	ThumbsUp     int64   `json:"thumbs_up"`
	ThumbsDown   int64   `json:"thumbs_down"`
	AccuracyRate float64 `json:"accuracy_rate"`
}

// TrendingPoint represents a point in the trending timeline
type TrendingPoint struct {
	Date         string  `json:"date"`
	ThumbsUp     int64   `json:"thumbs_up"`
	ThumbsDown   int64   `json:"thumbs_down"`
	AccuracyRate float64 `json:"accuracy_rate"`
}

// FeedbackAggregator aggregates feedback by topic, grade, and subject
type FeedbackAggregator struct {
	redisClient *redis.Client
}

// NewFeedbackAggregator creates a new feedback aggregator
func NewFeedbackAggregator(redisClient *redis.Client) *FeedbackAggregator {
	return &FeedbackAggregator{
		redisClient: redisClient,
	}
}

// RecordFeedback records feedback and updates aggregated stats
func (a *FeedbackAggregator) RecordFeedback(ctx context.Context, thumbsUp bool, grade, subject string) error {
	// Update global counters
	if err := a.incrementCounter(ctx, feedbackAggKey("global", thumbsUp)); err != nil {
		return fmt.Errorf("failed to update global feedback counter: %w", err)
	}

	// Update per-grade counters
	if grade != "" {
		if err := a.incrementCounter(ctx, feedbackAggKey(fmt.Sprintf("grade:%s", grade), thumbsUp)); err != nil {
			log.Warn().Err(err).Str("grade", grade).Msg("Failed to update grade feedback counter")
		}
	}

	// Update per-subject counters
	if subject != "" {
		if err := a.incrementCounter(ctx, feedbackAggKey(fmt.Sprintf("subject:%s", subject), thumbsUp)); err != nil {
			log.Warn().Err(err).Str("subject", subject).Msg("Failed to update subject feedback counter")
		}
	}

	// Update daily trending
	today := time.Now().UTC().Format("2006-01-02")
	if err := a.incrementCounter(ctx, feedbackAggKey(fmt.Sprintf("daily:%s", today), thumbsUp)); err != nil {
		log.Warn().Err(err).Str("date", today).Msg("Failed to update daily feedback counter")
	}

	return nil
}

// GetStats returns aggregated feedback statistics
func (a *FeedbackAggregator) GetStats(ctx context.Context) (*FeedbackStats, error) {
	stats := &FeedbackStats{
		ByGrade:   make(map[string]*Grade),
		BySubject: make(map[string]*Subject),
	}

	// Get global stats
	globalUp, _ := a.getCounter(ctx, feedbackAggKey("global", true))
	globalDown, _ := a.getCounter(ctx, feedbackAggKey("global", false))
	stats.ThumbsUp = globalUp
	stats.ThumbsDown = globalDown
	stats.TotalFeedback = globalUp + globalDown
	if stats.TotalFeedback > 0 {
		stats.AccuracyRate = float64(globalUp) / float64(stats.TotalFeedback)
	}

	// Get per-grade stats
	grades := []string{"6", "7", "8", "9", "10", "11", "12"}
	for _, grade := range grades {
		key := fmt.Sprintf("grade:%s", grade)
		up, _ := a.getCounter(ctx, feedbackAggKey(key, true))
		down, _ := a.getCounter(ctx, feedbackAggKey(key, false))
		if up > 0 || down > 0 {
			total := up + down
			stats.ByGrade[grade] = &Grade{
				ThumbsUp:     up,
				ThumbsDown:   down,
				AccuracyRate: float64(up) / float64(total),
			}
		}
	}

	// Get per-subject stats
	subjects := []string{"mathematics", "science", "language_arts", "history"}
	for _, subject := range subjects {
		key := fmt.Sprintf("subject:%s", subject)
		up, _ := a.getCounter(ctx, feedbackAggKey(key, true))
		down, _ := a.getCounter(ctx, feedbackAggKey(key, false))
		if up > 0 || down > 0 {
			total := up + down
			stats.BySubject[subject] = &Subject{
				ThumbsUp:     up,
				ThumbsDown:   down,
				AccuracyRate: float64(up) / float64(total),
			}
		}
	}

	// Get trending (last 7 days)
	stats.Trending = a.getTrending(ctx, 7)

	return stats, nil
}

// incrementCounter increments a Redis counter
func (a *FeedbackAggregator) incrementCounter(ctx context.Context, key string) error {
	pipe := a.redisClient.Pipeline()
	pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, FeedbackStatsTTL)
	_, err := pipe.Exec(ctx)
	return err
}

// getCounter gets the value of a Redis counter
func (a *FeedbackAggregator) getCounter(ctx context.Context, key string) (int64, error) {
	val, err := a.redisClient.Get(ctx, key).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return val, err
}

// getTrending returns trending data for the last N days
func (a *FeedbackAggregator) getTrending(ctx context.Context, days int) []TrendingPoint {
	var trending []TrendingPoint
	now := time.Now().UTC()

	for i := days - 1; i >= 0; i-- {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		key := fmt.Sprintf("daily:%s", date)

		up, _ := a.getCounter(ctx, feedbackAggKey(key, true))
		down, _ := a.getCounter(ctx, feedbackAggKey(key, false))

		total := up + down
		accuracyRate := float64(0)
		if total > 0 {
			accuracyRate = float64(up) / float64(total)
		}

		trending = append(trending, TrendingPoint{
			Date:         date,
			ThumbsUp:     up,
			ThumbsDown:   down,
			AccuracyRate: accuracyRate,
		})
	}

	return trending
}

// ServeHTTP handles GET /feedback/stats
func (a *FeedbackAggregator) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, map[string]string{
			"error": "Only GET is allowed",
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	stats, err := a.GetStats(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get feedback stats")
		writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve feedback stats",
		})
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// feedbackAggKey generates the Redis key for aggregated feedback
func feedbackAggKey(category string, thumbsUp bool) string {
	scoreType := "up"
	if !thumbsUp {
		scoreType = "down"
	}
	return fmt.Sprintf("%s%s:%s", FeedbackAggKeyPrefix, category, scoreType)
}

// FormatAccuracy formats an accuracy rate as a percentage string
func FormatAccuracy(rate float64) string {
	return strconv.FormatFloat(rate*100, 'f', 1, 64) + "%"
}

// StoreFeedbackWithContext stores feedback with full query/response/context (Story A)
func (a *FeedbackAggregator) StoreFeedbackWithContext(ctx context.Context, feedbackID, query, response string, retrievedContext []string, grade, subject string) error {
	contextData := map[string]interface{}{
		"feedback_id":       feedbackID,
		"query":             query,
		"response":          response,
		"retrieved_context": retrievedContext,
		"grade":             grade,
		"subject":           subject,
		"timestamp":         time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(contextData)
	if err != nil {
		return fmt.Errorf("failed to marshal feedback context: %w", err)
	}

	key := fmt.Sprintf("rag:feedback:context:%s", feedbackID)
	if err := a.redisClient.Set(ctx, key, data, FeedbackReviewTTL).Err(); err != nil {
		return fmt.Errorf("failed to store feedback context: %w", err)
	}

	return nil
}

// GetReviewItems retrieves thumbs_down items for review for a given date (Story A)
func (a *FeedbackAggregator) GetReviewItems(ctx context.Context, date string) ([]map[string]string, error) {
	reviewKey := fmt.Sprintf("rag:feedback:review:%s", date)

	// Get all items from the sorted set (score 0 to +inf)
	items, err := a.redisClient.ZRange(ctx, reviewKey, 0, -1).Result()
	if err != nil {
		if err == redis.Nil {
			return []map[string]string{}, nil
		}
		return nil, fmt.Errorf("failed to get review items: %w", err)
	}

	var reviews []map[string]string
	for _, item := range items {
		var review map[string]string
		if err := json.Unmarshal([]byte(item), &review); err != nil {
			log.Warn().Err(err).Msg("Failed to unmarshal review item")
			continue
		}
		reviews = append(reviews, review)
	}

	return reviews, nil
}

// GetFeedbackContext retrieves the full context for a feedback entry (Story A)
func (a *FeedbackAggregator) GetFeedbackContext(ctx context.Context, feedbackID string) (map[string]interface{}, error) {
	key := fmt.Sprintf("rag:feedback:context:%s", feedbackID)

	data, err := a.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("feedback context not found")
		}
		return nil, fmt.Errorf("failed to get feedback context: %w", err)
	}

	var context map[string]interface{}
	if err := json.Unmarshal([]byte(data), &context); err != nil {
		return nil, fmt.Errorf("failed to unmarshal feedback context: %w", err)
	}

	return context, nil
}
