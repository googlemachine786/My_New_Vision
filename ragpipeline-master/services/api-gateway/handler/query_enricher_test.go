package handler

import (
	"testing"

	"github.com/visionary/ragpipeline/services/api-gateway/context"
)

func TestBuildPromptPrefix_FullContext(t *testing.T) {
	eq := &EnrichedQuery{
		OriginalQuery:  "What is photosynthesis?",
		Grade:          "10",
		Board:          "CBSE",
		Language:       "English",
		CurrentSubject: "Biology",
		CurrentChapter: "Life Processes",
	}

	result := BuildPromptPrefix(eq)

	if result == "" {
		t.Fatal("expected non-empty prompt prefix")
	}
}

func TestBuildPromptPrefix_MinimalContext(t *testing.T) {
	eq := &EnrichedQuery{
		OriginalQuery: "What is photosynthesis?",
		Grade:         "8",
	}

	result := BuildPromptPrefix(eq)
	if result == "" {
		t.Fatal("expected non-empty prompt prefix")
	}
}

func TestBuildPromptPrefix_WithErrors(t *testing.T) {
	eq := &EnrichedQuery{
		OriginalQuery: "What is force?",
		Grade:         "9",
		PracticeErrors: []context.PracticeError{
			{Topic: "physics", ErrorMsg: "unit conversion", Count: 2},
		},
	}

	result := BuildPromptPrefix(eq)
	if result == "" {
		t.Fatal("expected non-empty prompt prefix")
	}
}

func TestQueryEnricher_Enrich_NoRegistry(t *testing.T) {
	enricher := NewQueryEnricher(nil)

	result, err := enricher.Enrich(nil, "sess1", "user1", "test query", "10")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.OriginalQuery != "test query" {
		t.Errorf("expected original query 'test query', got '%s'", result.OriginalQuery)
	}

	if result.EnrichedPrompt == "" {
		t.Error("expected non-empty enriched prompt")
	}
}

func TestCoalesceString(t *testing.T) {
	tests := []struct {
		input    []string
		expected string
	}{
		{[]string{"a", "b"}, "a"},
		{[]string{"", "b"}, "b"},
		{[]string{"", "", "c"}, "c"},
		{[]string{"", ""}, ""},
	}

	for _, tt := range tests {
		result := coalesceString(tt.input...)
		if result != tt.expected {
			t.Errorf("coalesceString(%v) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
