package search

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/visionary/ragpipeline/services/vector-search-service/config"
	"github.com/visionary/ragpipeline/services/vector-search-service/db"
)

// SparseSearcher performs sparse keyword search using GIN index intersection.
type SparseSearcher struct {
	store *db.Store
	cfg   *config.Config
}

// NewSparseSearcher creates a new sparse searcher.
func NewSparseSearcher(store *db.Store, cfg *config.Config) *SparseSearcher {
	return &SparseSearcher{
		store: store,
		cfg:   cfg,
	}
}

// Search performs sparse keyword search using GIN index on extracted_keywords.
// Returns results ranked by keyword match count (descending).
func (s *SparseSearcher) Search(ctx context.Context, query *db.SearchQuery) ([]db.SparseSearchResult, time.Duration, error) {
	start := time.Now()

	if len(query.Keywords) == 0 {
		return nil, 0, nil
	}

	if query.TopK < 1 {
		query.TopK = s.cfg.DefaultTopK
	}

	results, duration, err := s.store.SparseSearch(ctx, query)
	if err != nil {
		log.Error().
			Err(err).
			Strs("keywords", query.Keywords).
			Int("top_k", query.TopK).
			Msg("Sparse search failed")
		return nil, time.Since(start), err
	}

	log.Info().
		Int("results", len(results)).
		Int("keywords", len(query.Keywords)).
		Int64("duration_ms", duration.Milliseconds()).
		Msg("Sparse search completed")

	return results, time.Since(start), nil
}
