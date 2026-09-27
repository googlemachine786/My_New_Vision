package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestFeedbackAggregator_RecordFeedback(t *testing.T) {
	// Skip if Redis is not available
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping test")
	}

	// Clean up test keys
	defer func() {
		rdb.Del(ctx, "rag:feedback:agg:global:up", "rag:feedback:agg:global:down")
	}()

	aggregator := NewFeedbackAggregator(rdb)

	// Record thumbs up
	if err := aggregator.RecordFeedback(ctx, true, "", ""); err != nil {
		t.Fatalf("RecordFeedback(true) error: %v", err)
	}

	// Record thumbs down
	if err := aggregator.RecordFeedback(ctx, false, "", ""); err != nil {
		t.Fatalf("RecordFeedback(false) error: %v", err)
	}

	// Verify counters
	up, _ := aggregator.getCounter(ctx, feedbackAggKey("global", true))
	down, _ := aggregator.getCounter(ctx, feedbackAggKey("global", false))

	if up != 1 {
		t.Errorf("expected thumbs_up counter = 1, got %d", up)
	}

	if down != 1 {
		t.Errorf("expected thumbs_down counter = 1, got %d", down)
	}
}

func TestFeedbackAggregator_GetStats(t *testing.T) {
	// Skip if Redis is not available
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping test")
	}

	// Clean up test keys
	defer func() {
		rdb.Del(ctx, "rag:feedback:agg:global:up", "rag:feedback:agg:global:down")
	}()

	aggregator := NewFeedbackAggregator(rdb)

	// Record some feedback
	_ = aggregator.RecordFeedback(ctx, true, "", "")
	_ = aggregator.RecordFeedback(ctx, true, "", "")
	_ = aggregator.RecordFeedback(ctx, false, "", "")

	// Get stats
	stats, err := aggregator.GetStats(ctx)
	if err != nil {
		t.Fatalf("GetStats() error: %v", err)
	}

	if stats.ThumbsUp != 2 {
		t.Errorf("expected thumbs_up = 2, got %d", stats.ThumbsUp)
	}

	if stats.ThumbsDown != 1 {
		t.Errorf("expected thumbs_down = 1, got %d", stats.ThumbsDown)
	}

	if stats.TotalFeedback != 3 {
		t.Errorf("expected total_feedback = 3, got %d", stats.TotalFeedback)
	}

	expectedRate := float64(2) / float64(3)
	if stats.AccuracyRate < expectedRate-0.01 || stats.AccuracyRate > expectedRate+0.01 {
		t.Errorf("expected accuracy_rate ~= %f, got %f", expectedRate, stats.AccuracyRate)
	}
}

func TestFeedbackAggregator_ServeHTTP(t *testing.T) {
	// Skip if Redis is not available
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis not available, skipping test")
	}

	aggregator := NewFeedbackAggregator(rdb)

	req := httptest.NewRequest(http.MethodGet, "/feedback/stats", nil)
	w := httptest.NewRecorder()

	aggregator.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var stats FeedbackStats
	if err := json.Unmarshal(w.Body.Bytes(), &stats); err != nil {
		t.Fatalf("failed to unmarshal stats: %v", err)
	}

	// by_grade and by_subject are empty when no feedback recorded
	// (nil maps are valid JSON as null, empty maps as {})
	if stats.TotalFeedback != 0 {
		t.Errorf("expected 0 total feedbacks, got %d", stats.TotalFeedback)
	}
}

func TestFeedbackAggregator_ServeHTTP_InvalidMethod(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	aggregator := NewFeedbackAggregator(rdb)

	req := httptest.NewRequest(http.MethodPost, "/feedback/stats", nil)
	w := httptest.NewRecorder()

	aggregator.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func TestFormatAccuracy(t *testing.T) {
	tests := []struct {
		rate     float64
		expected string
	}{
		{0.0, "0.0%"},
		{0.5, "50.0%"},
		{0.95, "95.0%"},
		{1.0, "100.0%"},
		{0.123, "12.3%"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			got := FormatAccuracy(tt.rate)
			if got != tt.expected {
				t.Errorf("FormatAccuracy(%f) = %s, want %s", tt.rate, got, tt.expected)
			}
		})
	}
}
