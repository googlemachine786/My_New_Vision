package search

import (
	"errors"
	"strconv"
)

// Sentinel errors for the search package.
var (
	ErrEmptyEmbedding       = errors.New("embedding vector is empty")
	ErrEmptyKeywords        = errors.New("keywords list is empty")
	ErrNoSearchCriteria     = errors.New("at least one of embedding or keywords must be provided")
	ErrInvalidTopK          = errors.New("top_k must be positive")
	ErrTopKExceedsMax       = errors.New("top_k exceeds maximum allowed value")
)

// ErrEmbeddingDimensionMismatch is returned when the embedding dimension doesn't match the expected dimension.
type ErrEmbeddingDimensionMismatch struct {
	Expected int
	Got      int
}

func (e ErrEmbeddingDimensionMismatch) Error() string {
	return "embedding dimension mismatch: expected " + strconv.Itoa(e.Expected) + ", got " + strconv.Itoa(e.Got)
}
