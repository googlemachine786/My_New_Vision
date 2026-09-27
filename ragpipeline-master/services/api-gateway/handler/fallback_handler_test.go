package handler

import (
	"testing"

	"github.com/visionary/ragpipeline/pkg/types"
)

func TestShouldUseFallback(t *testing.T) {
	tests := []struct {
		name    string
		sources []types.Source
		want    bool
	}{
		{
			name:    "empty sources triggers fallback",
			sources: []types.Source{},
			want:    true,
		},
		{
			name:    "nil sources triggers fallback",
			sources: nil,
			want:    true,
		},
		{
			name: "non-empty sources no fallback",
			sources: []types.Source{
				{Content: "test content", Score: 0.8},
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ShouldUseFallback(tt.sources)
			if got != tt.want {
				t.Errorf("ShouldUseFallback() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCalculateConfidenceScore(t *testing.T) {
	tests := []struct {
		name         string
		sources      []types.Source
		llmConfidence float64
		wantMin      float64
		wantMax      float64
	}{
		{
			name:    "no sources returns low confidence",
			sources: []types.Source{},
			wantMin: 0.0,
			wantMax: 0.3,
		},
		{
			name: "high score sources returns high confidence",
			sources: []types.Source{
				{Score: 0.9},
				{Score: 0.85},
			},
			llmConfidence: 0.0,
			wantMin:       0.7,
			wantMax:       1.0,
		},
		{
			name: "low score sources returns low confidence",
			sources: []types.Source{
				{Score: 0.3},
				{Score: 0.4},
			},
			llmConfidence: 0.0,
			wantMin:       0.0,
			wantMax:       0.5,
		},
		{
			name: "mixed sources with LLM confidence",
			sources: []types.Source{
				{Score: 0.8},
				{Score: 0.7},
			},
			llmConfidence: 0.9,
			wantMin:       0.7,
			wantMax:       1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateConfidenceScore(tt.sources, tt.llmConfidence)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("calculateConfidenceScore() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestCalculateConfidenceScore_Bounds(t *testing.T) {
	// Test that confidence is always between 0.0 and 1.0
	sources := []types.Source{
		{Score: 1.5}, // Above 1.0
	}
	got := calculateConfidenceScore(sources, 0.0)
	if got < 0.0 || got > 1.0 {
		t.Errorf("calculateConfidenceScore() = %v, want between 0.0 and 1.0", got)
	}

	// Test negative score
	sources = []types.Source{
		{Score: -0.5},
	}
	got = calculateConfidenceScore(sources, 0.0)
	if got < 0.0 || got > 1.0 {
		t.Errorf("calculateConfidenceScore() with negative score = %v, want between 0.0 and 1.0", got)
	}
}

func TestFallbackHandler_buildFallbackPrompt(t *testing.T) {
	handler := &FallbackHandler{}

	tests := []struct {
		name    string
		query   string
		grade   string
		subject string
	}{
		{
			name:    "all fields provided",
			query:   "What is photosynthesis?",
			grade:   "8",
			subject: "science",
		},
		{
			name:    "empty grade and subject",
			query:   "What is gravity?",
			grade:   "",
			subject: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := handler.buildFallbackPrompt(tt.query, tt.grade, tt.subject)
			if prompt == "" {
				t.Error("buildFallbackPrompt() returned empty string")
			}
			// Verify prompt contains the query
			if !containsStr(prompt, tt.query) {
				t.Errorf("buildFallbackPrompt() does not contain query %q", tt.query)
			}
			// Verify prompt contains disclaimer
			if !containsStr(prompt, "general knowledge") {
				t.Error("buildFallbackPrompt() does not contain general knowledge disclaimer")
			}
		})
	}
}

func TestRecordFallbackMetric(t *testing.T) {
	// Verify no panic
	RecordFallbackMetric()
}

// containsStr checks if s contains substr
func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStrHelper(s, substr))
}

func containsStrHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
