// Package search provides hybrid search orchestration for RAG retrieval,
// including dense vector search, sparse keyword search, and RRF fusion.
package search

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/visionary/ragpipeline/services/vector-search-service/config"
	"github.com/visionary/ragpipeline/services/vector-search-service/db"
)

// SearchResult represents a unified search result with scores from all sources.
type SearchResult struct {
	ParentID      string                 `json:"parent_id"`
	Content       string                 `json:"content"`
	ParentContent string                 `json:"parent_content"`
	Metadata      map[string]interface{} `json:"metadata"`
	DenseScore    *float64               `json:"dense_score,omitempty"`
	SparseScore   *float64               `json:"sparse_score,omitempty"`
	BM25Score     *float64               `json:"bm25_score,omitempty"`
	RRFScore      float64                `json:"rrf_score"`
	DenseRank     *int                   `json:"dense_rank,omitempty"`
	SparseRank    *int                   `json:"sparse_rank,omitempty"`
	BM25Rank      *int                   `json:"bm25_rank,omitempty"`
}

// SearchResponse represents the complete search response.
type SearchResponse struct {
	Results      []SearchResult `json:"results"`
	TotalResults int            `json:"total_results"`
	SearchTimeMS int64          `json:"search_time_ms"`
	Stats        *db.QueryStats `json:"stats,omitempty"`
}

// DenseSearcher performs dense vector similarity search.
type DenseSearcher struct {
	store *db.Store
	cfg   *config.Config
}

// NewDenseSearcher creates a new dense searcher.
func NewDenseSearcher(store *db.Store, cfg *config.Config) *DenseSearcher {
	return &DenseSearcher{
		store: store,
		cfg:   cfg,
	}
}

// Search performs dense vector similarity search using ScaNN cosine distance.
// Returns results ranked by cosine similarity (lower distance = higher similarity).
func (d *DenseSearcher) Search(ctx context.Context, query *db.SearchQuery) ([]db.DenseSearchResult, time.Duration, error) {
	start := time.Now()

	if len(query.Embedding) == 0 {
		return nil, 0, ErrEmptyEmbedding
	}

	if len(query.Embedding) != d.cfg.EmbeddingDim {
		return nil, 0, ErrEmbeddingDimensionMismatch{
			Expected: d.cfg.EmbeddingDim,
			Got:      len(query.Embedding),
		}
	}

	if query.TopK < 1 {
		query.TopK = d.cfg.DefaultTopK
	}

	results, duration, err := d.store.DenseSearch(ctx, query)
	if err != nil {
		log.Error().
			Err(err).
			Int("top_k", query.TopK).
			Int("embedding_dim", len(query.Embedding)).
			Msg("Dense search failed")
		return nil, time.Since(start), err
	}

	log.Info().
		Int("results", len(results)).
		Int64("duration_ms", duration.Milliseconds()).
		Msg("Dense search completed")

	return results, time.Since(start), nil
}

// ValidateEmbedding checks that the embedding vector is valid.
func (d *DenseSearcher) ValidateEmbedding(embedding []float32) error {
	if len(embedding) == 0 {
		return ErrEmptyEmbedding
	}
	if len(embedding) != d.cfg.EmbeddingDim {
		return ErrEmbeddingDimensionMismatch{
			Expected: d.cfg.EmbeddingDim,
			Got:      len(embedding),
		}
	}
	return nil
}
