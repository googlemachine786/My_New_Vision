// Package handler provides HTTP handlers for the API Gateway.
package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/visionary/ragpipeline/pkg/types"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	// QualityFlaggedKeyPrefix is the Redis key prefix for flagged quality items
	QualityFlaggedKeyPrefix = "quality:flagged:"
	// QualityDatasetKeyPrefix is the Redis key prefix for weekly fine-tuning datasets
	QualityDatasetKeyPrefix = "quality:dataset:"
	// QualityFlaggedTTL is the TTL for flagged items (90 days)
	QualityFlaggedTTL = 90 * 24 * time.Hour
	// QualityDatasetTTL is the TTL for weekly datasets (1 year)
	QualityDatasetTTL = 365 * 24 * time.Hour
	// LowConfidenceThreshold is the confidence threshold below which responses are flagged
	LowConfidenceThreshold = 0.6
)

// FlaggedItem represents a response flagged for quality review (Story B)
type FlaggedItem struct {
	Query            string   `json:"query"`
	Response         string   `json:"response"`
	RetrievedContext []string `json:"retrieved_context"`
	ConfidenceScore  float64  `json:"confidence_score"`
	FeedbackReason   string   `json:"feedback_reason,omitempty"`
	ThumbsDownCount  int      `json:"thumbs_down_count"`
	Grade            string   `json:"grade,omitempty"`
	Subject          string   `json:"subject,omitempty"`
	Timestamp        string   `json:"timestamp"`
}

// QualityPipeline manages automated quality monitoring (Story B)
type QualityPipeline struct {
	redisClient *redis.Client
}

// NewQualityPipeline creates a new quality pipeline handler
func NewQualityPipeline(redisClient *redis.Client) *QualityPipeline {
	return &QualityPipeline{
		redisClient: redisClient,
	}
}

// AutoFlagResponse automatically flags a response if it has low confidence or thumbs_down
func (q *QualityPipeline) AutoFlagResponse(ctx context.Context, query, response string, retrievedContext []string, confidenceScore float64, grade, subject string) error {
	shouldFlag := confidenceScore < LowConfidenceThreshold

	if !shouldFlag {
		return nil
	}

	flaggedItem := FlaggedItem{
		Query:            query,
		Response:         response,
		RetrievedContext: retrievedContext,
		ConfidenceScore:  confidenceScore,
		Grade:            grade,
		Subject:          subject,
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
	}

	return q.addFlaggedItem(ctx, flaggedItem)
}

// FlagFromFeedback flags a response based on thumbs_down feedback
func (q *QualityPipeline) FlagFromFeedback(ctx context.Context, query, response string, retrievedContext []string, feedbackReason string, thumbsDownCount int, grade, subject string) error {
	// Auto-flag if 2+ thumbs_down or low confidence
	if thumbsDownCount < 2 {
		return nil
	}

	flaggedItem := FlaggedItem{
		Query:            query,
		Response:         response,
		RetrievedContext: retrievedContext,
		FeedbackReason:   feedbackReason,
		ThumbsDownCount:  thumbsDownCount,
		Grade:            grade,
		Subject:          subject,
		Timestamp:        time.Now().UTC().Format(time.RFC3339),
	}

	return q.addFlaggedItem(ctx, flaggedItem)
}

// addFlaggedItem adds an item to the daily flagged items set
func (q *QualityPipeline) addFlaggedItem(ctx context.Context, item FlaggedItem) error {
	data, err := json.Marshal(item)
	if err != nil {
		return fmt.Errorf("failed to marshal flagged item: %w", err)
	}

	date := time.Now().UTC().Format("2006-01-02")
	key := fmt.Sprintf("%s%s", QualityFlaggedKeyPrefix, date)

	// Add to sorted set with timestamp score
	if err := q.redisClient.ZAdd(ctx, key, redis.Z{
		Score:  float64(time.Now().UTC().Unix()),
		Member: data,
	}).Err(); err != nil {
		return fmt.Errorf("failed to add flagged item: %w", err)
	}

	// Set TTL
	q.redisClient.Expire(ctx, key, QualityFlaggedTTL)

	log.Info().
		Str("query", item.Query).
		Float64("confidence", item.ConfidenceScore).
		Str("date", date).
		Msg("Response auto-flagged for quality review")

	return nil
}

// GenerateWeeklyDataset generates a fine-tuning dataset from flagged items (Story B)
func (q *QualityPipeline) GenerateWeeklyDataset(ctx context.Context, week string) ([]types.QualityExportItem, error) {
	// Parse week string (YYYY-WW format)
	year, weekNum, err := parseWeekString(week)
	if err != nil {
		return nil, fmt.Errorf("invalid week format (expected YYYY-WW): %w", err)
	}

	// Calculate date range for the week
	startDate := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	startDate = startDate.AddDate(0, 0, (weekNum-1)*7)
	endDate := startDate.AddDate(0, 0, 7)

	var exportItems []types.QualityExportItem

	// Collect flagged items for each day in the week
	for d := startDate; d.Before(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		key := fmt.Sprintf("%s%s", QualityFlaggedKeyPrefix, dateStr)

		items, err := q.redisClient.ZRange(ctx, key, 0, -1).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			log.Warn().Err(err).Str("date", dateStr).Msg("Failed to get flagged items")
			continue
		}

		for _, itemStr := range items {
			var flagged FlaggedItem
			if err := json.Unmarshal([]byte(itemStr), &flagged); err != nil {
				log.Warn().Err(err).Msg("Failed to unmarshal flagged item")
				continue
			}

			exportItems = append(exportItems, types.QualityExportItem{
				Query:            flagged.Query,
				Response:         flagged.Response,
				RetrievedContext: flagged.RetrievedContext,
				FeedbackReason:   flagged.FeedbackReason,
				ConfidenceScore:  flagged.ConfidenceScore,
				Timestamp:        flagged.Timestamp,
			})
		}
	}

	// Store the dataset for future retrieval
	if len(exportItems) > 0 {
		datasetKey := fmt.Sprintf("%s%s", QualityDatasetKeyPrefix, week)
		data, err := json.Marshal(exportItems)
		if err != nil {
			log.Error().Err(err).Str("week", week).Msg("Failed to marshal weekly dataset")
		} else {
			if err := q.redisClient.Set(ctx, datasetKey, data, QualityDatasetTTL).Err(); err != nil {
				log.Warn().Err(err).Str("week", week).Msg("Failed to store weekly dataset")
			}
		}
	}

	log.Info().
		Str("week", week).
		Int("items", len(exportItems)).
		Msg("Generated weekly quality dataset")

	return exportItems, nil
}

// GetWeeklyDataset retrieves a previously generated weekly dataset
func (q *QualityPipeline) GetWeeklyDataset(ctx context.Context, week string) ([]types.QualityExportItem, error) {
	datasetKey := fmt.Sprintf("%s%s", QualityDatasetKeyPrefix, week)

	data, err := q.redisClient.Get(ctx, datasetKey).Result()
	if err != nil {
		if err == redis.Nil {
			// Dataset not pre-generated, generate on demand
			return q.GenerateWeeklyDataset(ctx, week)
		}
		return nil, fmt.Errorf("failed to get weekly dataset: %w", err)
	}

	var items []types.QualityExportItem
	if err := json.Unmarshal([]byte(data), &items); err != nil {
		return nil, fmt.Errorf("failed to unmarshal weekly dataset: %w", err)
	}

	return items, nil
}

// ServeHTTP handles GET /quality/export?week=YYYY-WW (Story B)
func (q *QualityPipeline) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeJSON(w, http.StatusMethodNotAllowed, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "method_not_allowed", Message: "Only GET is allowed"},
		})
		return
	}

	week := r.URL.Query().Get("week")
	if week == "" {
		// Default to current week
		now := time.Now().UTC()
		_, wNum := now.ISOWeek()
		week = fmt.Sprintf("%d-W%02d", now.Year(), wNum)
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	items, err := q.GetWeeklyDataset(ctx, week)
	if err != nil {
		log.Error().Err(err).Str("week", week).Msg("Failed to get weekly dataset")
		writeJSON(w, http.StatusInternalServerError, types.ErrorResponse{
			Error: types.ErrorDetail{Code: "internal_error", Message: "Failed to export quality dataset"},
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"week":  week,
		"count": len(items),
		"items": items,
	})
}

// parseWeekString parses a YYYY-WW format string
func parseWeekString(week string) (int, int, error) {
	if len(week) < 7 {
		return 0, 0, fmt.Errorf("week string too short")
	}

	// Handle both "YYYY-WW" and "YYYY-WWW" formats
	var year int
	var weekNum int

	// Find the 'W' separator
	wIndex := -1
	for i := 0; i < len(week); i++ {
		if week[i] == 'W' || week[i] == 'w' {
			wIndex = i
			break
		}
	}

	if wIndex <= 0 {
		return 0, 0, fmt.Errorf("invalid week format: missing W separator")
	}

	// Parse year
	yearStr := week[:wIndex]
	// Remove trailing dash if present
	if yearStr[len(yearStr)-1] == '-' {
		yearStr = yearStr[:len(yearStr)-1]
	}

	_, err := fmt.Sscanf(yearStr, "%d", &year)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid year in week string: %w", err)
	}

	// Parse week number
	weekStr := week[wIndex+1:]
	_, err = fmt.Sscanf(weekStr, "%d", &weekNum)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid week number: %w", err)
	}

	if weekNum < 1 || weekNum > 53 {
		return 0, 0, fmt.Errorf("week number must be between 1 and 53")
	}

	return year, weekNum, nil
}

// writeJSON is a helper to write JSON responses (defined in other handler files)
// This is a forward declaration - the actual implementation is in the handler package
