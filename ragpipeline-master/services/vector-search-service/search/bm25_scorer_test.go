package search

import (
	"testing"

	"github.com/visionary/ragpipeline/services/vector-search-service/index"
)

func TestNewBM25Scorer(t *testing.T) {
	idx := index.NewKeywordIndex()
	scorer := NewBM25Scorer(idx, 1.5, 0.75)

	if scorer == nil {
		t.Fatal("expected non-nil scorer")
	}
	if scorer.k1 != 1.5 {
		t.Errorf("expected k1=1.5, got %f", scorer.k1)
	}
	if scorer.b != 0.75 {
		t.Errorf("expected b=0.75, got %f", scorer.b)
	}
}

func TestBM25Scorer_DefaultParams(t *testing.T) {
	idx := index.NewKeywordIndex()
	scorer := NewBM25Scorer(idx, 0, -1) // invalid params should use defaults

	if scorer.k1 != 1.5 {
		t.Errorf("expected default k1=1.5, got %f", scorer.k1)
	}
	if scorer.b != 0.75 {
		t.Errorf("expected default b=0.75, got %f", scorer.b)
	}
}

func TestBM25Scorer_EmptyQuery(t *testing.T) {
	idx := index.NewKeywordIndex()
	scorer := NewBM25Scorer(idx, 1.5, 0.75)

	results := scorer.Score(nil)
	if results != nil {
		t.Errorf("expected nil results for empty query, got %d items", len(results))
	}
}

func TestBM25Scorer_ScoreWithQuery(t *testing.T) {
	idx := index.NewKeywordIndex()
	idx.AddDocument("doc1", "the quick brown fox jumps over the lazy dog")
	idx.AddDocument("doc2", "the lazy dog sleeps all day")
	idx.AddDocument("doc3", "the quick rabbit hops around")

	scorer := NewBM25Scorer(idx, 1.5, 0.75)
	results := scorer.ScoreWithQuery("quick fox")

	if len(results) == 0 {
		t.Fatal("expected non-empty results")
	}

	// Results should be sorted by score descending
	for i := 1; i < len(results); i++ {
		if results[i].Score > results[i-1].Score {
			t.Errorf("results not sorted by score descending at index %d", i)
		}
	}

	// Ranks should be sequential
	for i, r := range results {
		if r.Rank != i+1 {
			t.Errorf("expected rank %d, got %d", i+1, r.Rank)
		}
	}
}

func TestBM25Scorer_NoMatches(t *testing.T) {
	idx := index.NewKeywordIndex()
	idx.AddDocument("doc1", "the quick brown fox")

	scorer := NewBM25Scorer(idx, 1.5, 0.75)
	results := scorer.ScoreWithQuery("quantum physics relativity")

	if len(results) != 0 {
		t.Errorf("expected 0 results for non-matching query, got %d", len(results))
	}
}

func TestBM25Sort(t *testing.T) {
	results := []BM25Result{
		{DocID: "a", Score: 0.5},
		{DocID: "b", Score: 1.5},
		{DocID: "c", Score: 1.0},
	}

	bm25Sort(results)

	if results[0].Score != 1.5 {
		t.Errorf("expected first score 1.5, got %f", results[0].Score)
	}
	if results[1].Score != 1.0 {
		t.Errorf("expected second score 1.0, got %f", results[1].Score)
	}
	if results[2].Score != 0.5 {
		t.Errorf("expected third score 0.5, got %f", results[2].Score)
	}
}

func TestTokenize(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"Hello World", []string{"hello", "world"}},
		{"Test, punctuation.", []string{"test", "punctuation"}},
		{"", nil},
	}

	for _, tt := range tests {
		result := tokenize(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("tokenize(%q) = %d tokens, want %d", tt.input, len(result), len(tt.expected))
			continue
		}
		for i, token := range result {
			if token != tt.expected[i] {
				t.Errorf("tokenize(%q)[%d] = %q, want %q", tt.input, i, token, tt.expected[i])
			}
		}
	}
}
