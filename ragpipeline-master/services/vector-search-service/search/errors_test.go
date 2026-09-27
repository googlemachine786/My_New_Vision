package search

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want string
	}{
		{
			name: "ErrEmptyEmbedding",
			err:  ErrEmptyEmbedding,
			want: "embedding vector is empty",
		},
		{
			name: "ErrEmptyKeywords",
			err:  ErrEmptyKeywords,
			want: "keywords list is empty",
		},
		{
			name: "ErrNoSearchCriteria",
			err:  ErrNoSearchCriteria,
			want: "at least one of embedding or keywords must be provided",
		},
		{
			name: "ErrInvalidTopK",
			err:  ErrInvalidTopK,
			want: "top_k must be positive",
		},
		{
			name: "ErrTopKExceedsMax",
			err:  ErrTopKExceedsMax,
			want: "top_k exceeds maximum allowed value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err == nil {
				t.Fatalf("%s is nil", tt.name)
			}

			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("%s.Error() = %q, want %q", tt.name, got, tt.want)
			}

			// Verify errors.Is works with sentinel
			if !errors.Is(tt.err, tt.err) {
				t.Errorf("errors.Is(%s, %s) = false, want true", tt.name, tt.name)
			}
		})
	}
}

func TestErrEmbeddingDimensionMismatch(t *testing.T) {
	tests := []struct {
		name     string
		expected int
		got      int
		wantContains string
	}{
		{
			name:     "positive dimensions",
			expected: 768,
			got:      512,
			wantContains: "embedding dimension mismatch: expected 768, got 512",
		},
		{
			name:     "zero expected",
			expected: 0,
			got:      100,
			wantContains: "embedding dimension mismatch: expected 0, got 100",
		},
		{
			name:     "zero got",
			expected: 768,
			got:      0,
			wantContains: "embedding dimension mismatch: expected 768, got 0",
		},
		{
			name:     "negative dimensions",
			expected: -1,
			got:      768,
			wantContains: "embedding dimension mismatch: expected -1, got 768",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ErrEmbeddingDimensionMismatch{
				Expected: tt.expected,
				Got:      tt.got,
			}

			got := err.Error()
			if got != tt.wantContains {
				t.Errorf("Error() = %q, want %q", got, tt.wantContains)
			}

			// Verify it contains key phrases
			if !strings.Contains(got, "embedding dimension mismatch") {
				t.Errorf("Error() should contain 'embedding dimension mismatch', got %q", got)
			}
			if !strings.Contains(got, "expected") {
				t.Errorf("Error() should contain 'expected', got %q", got)
			}
			if !strings.Contains(got, "got") {
				t.Errorf("Error() should contain 'got', got %q", got)
			}
		})
	}
}

func TestSentinelErrors_AsWrapped(t *testing.T) {
	// Verify sentinel errors work properly with wrapping
	tests := []struct {
		name string
		err  error
		target error
		want bool
	}{
		{
			name:   "wrapped ErrEmptyEmbedding",
			err:    &wrappedErr{cause: ErrEmptyEmbedding},
			target: ErrEmptyEmbedding,
			want:   true,
		},
		{
			name:   "unrelated error",
			err:    &wrappedErr{cause: errors.New("something else")},
			target: ErrEmptyEmbedding,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := errors.Is(tt.err, tt.target)
			if got != tt.want {
				t.Errorf("errors.Is(wrapped, %T) = %v, want %v", tt.target, got, tt.want)
			}
		})
	}
}

// wrappedErr is a simple error wrapper for testing errors.Is
type wrappedErr struct {
	cause error
}

func (e *wrappedErr) Error() string {
	return "wrapped: " + e.cause.Error()
}

func (e *wrappedErr) Unwrap() error {
	return e.cause
}

func TestItoa(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want string
	}{
		{"zero", 0, "0"},
		{"positive single digit", 5, "5"},
		{"positive multi digit", 123, "123"},
		{"negative single digit", -3, "-3"},
		{"negative multi digit", -456, "-456"},
		{"large positive", 999999, "999999"},
		{"large negative", -999999, "-999999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := strconv.Itoa(tt.n)
			if got != tt.want {
				t.Errorf("strconv.Itoa(%d) = %q, want %q", tt.n, got, tt.want)
			}
		})
	}
}
