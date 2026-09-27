package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/visionary/ragpipeline/services/vector-search-service/config"
	"github.com/visionary/ragpipeline/services/vector-search-service/db"
	"github.com/visionary/ragpipeline/services/vector-search-service/index"
)

// HybridSearcher orchestrates hybrid search with dense, sparse, BM25, and RRF fusion.
type HybridSearcher struct {
	denseSearcher  *DenseSearcher
	sparseSearcher *SparseSearcher
	bm25Scorer     *BM25Scorer
	keywordIndex   *index.KeywordIndex
	cfg            *config.Config
	rrfConfig      RRFConfig
}

// NewHybridSearcher creates a new hybrid searcher.
func NewHybridSearcher(
	store *db.Store,
	cfg *config.Config,
	rrfCfg *RRFConfig,
) *HybridSearcher {
	if rrfCfg == nil {
		defaultCfg := DefaultRRFConfig()
		rrfCfg = &defaultCfg
	}

	// Build keyword index from sparse search data
	keywordIdx := index.NewKeywordIndex()

	return &HybridSearcher{
		denseSearcher:  NewDenseSearcher(store, cfg),
		sparseSearcher: NewSparseSearcher(store, cfg),
		bm25Scorer:     NewBM25Scorer(keywordIdx, 1.5, 0.75),
		keywordIndex:   keywordIdx,
		cfg:            cfg,
		rrfConfig:      *rrfCfg,
	}
}

// IndexDocument adds a document to the BM25 keyword index.
func (h *HybridSearcher) IndexDocument(docID, content string) {
	h.keywordIndex.AddDocument(docID, content)
}

// Search performs hybrid search with dense vector similarity, sparse keyword matching,
// BM25 scoring, and RRF fusion to combine results.
//
// The search flow:
// 1. Dense search: ScaNN cosine similarity on child_chunks embedding vectors
// 2. Sparse search: GIN index intersection on parent_chunks extracted_keywords
// 3. BM25 scoring: Okapi BM25 over keyword index
// 4. RRF Fusion: Reciprocal Rank Fusion with k=60 to merge all three rankings
// 5. Deduplication: Remove duplicate parent chunks
// 6. Limit: Return top-k results
func (h *HybridSearcher) Search(ctx context.Context, query *db.SearchQuery) (*SearchResponse, error) {
	start := time.Now()
	stats := &db.QueryStats{}

	if query.TopK < 1 {
		query.TopK = h.cfg.DefaultTopK
	}
	if query.TopK > h.cfg.MaxTopK {
		query.TopK = h.cfg.MaxTopK
	}

	// Determine search mode based on input
	hasEmbedding := len(query.Embedding) > 0
	hasKeywords := len(query.Keywords) > 0

	if !hasEmbedding && !hasKeywords {
		return nil, ErrNoSearchCriteria
	}

	var denseResults []db.DenseSearchResult
	var sparseResults []db.SparseSearchResult
	var denseErr, sparseErr error

	// Execute dense and sparse search in parallel conceptually
	// (sequentially here, but could be parallelized with goroutines)

	// Dense search (if embedding provided)
	if hasEmbedding {
		denseStart := time.Now()
		denseResults, _, denseErr = h.denseSearcher.Search(ctx, query)
		stats.DenseTimeMS = time.Since(denseStart).Milliseconds()
		stats.DenseResults = len(denseResults)

		if denseErr != nil {
			log.Warn().Err(denseErr).Msg("Dense search failed, continuing with sparse only")
			denseResults = nil
		}
	}

	// Sparse search (if keywords provided)
	if hasKeywords {
		sparseStart := time.Now()
		sparseResults, _, sparseErr = h.sparseSearcher.Search(ctx, query)
		stats.SparseTimeMS = time.Since(sparseStart).Milliseconds()
		stats.SparseResults = len(sparseResults)

		if sparseErr != nil {
			log.Warn().Err(sparseErr).Msg("Sparse search failed, continuing with dense only")
			sparseResults = nil
		}
	}

	// BM25 search (if keywords provided)
	var bm25Results []BM25Result
	var bm25Err error
	if hasKeywords && h.bm25Scorer != nil {
		bm25Start := time.Now()
		bm25Results = h.bm25Scorer.ScoreWithQuery(strings.Join(query.Keywords, " "))
		stats.BM25TimeMS = time.Since(bm25Start).Milliseconds()
		stats.BM25Results = len(bm25Results)
	}

	// If all searches failed, return error
	if denseErr != nil && sparseErr != nil && bm25Err != nil {
		return nil, fmt.Errorf("all search types failed: dense=%w, sparse=%w, bm25=%w", denseErr, sparseErr, bm25Err)
	}

	// If only one search type was used/succeeded, return those results directly
	if !hasEmbedding && sparseErr == nil && bm25Err == nil && len(bm25Results) == 0 {
		searchResults := sparseOnlyToResults(sparseResults)
		stats.FusedResults = len(searchResults)
		stats.TotalTimeMS = time.Since(start).Milliseconds()

		return &SearchResponse{
			Results:      searchResults,
			TotalResults: len(searchResults),
			SearchTimeMS: stats.TotalTimeMS,
			Stats:        stats,
		}, nil
	}

	if !hasKeywords && denseErr == nil {
		searchResults := denseOnlyToResults(denseResults, query.TopK)
		stats.FusedResults = len(searchResults)
		stats.TotalTimeMS = time.Since(start).Milliseconds()

		return &SearchResponse{
			Results:      searchResults,
			TotalResults: len(searchResults),
			SearchTimeMS: stats.TotalTimeMS,
			Stats:        stats,
		}, nil
	}

	// RRF Fusion (3-way: dense + sparse + BM25)
	rrfStart := time.Now()
	fused := ReciprocalRankFuse(denseResults, sparseResults, bm25Results, h.rrfConfig)
	stats.RRFTimeMS = time.Since(rrfStart).Milliseconds()

	// Deduplicate and limit
	deduped := DeduplicateAndLimit(fused, query.TopK)
	stats.FusedResults = len(deduped)

	// Convert to unified search results
	searchResults := ToSearchResults(deduped)
	stats.TotalTimeMS = time.Since(start).Milliseconds()

	log.Info().
		Int("dense_results", stats.DenseResults).
		Int("sparse_results", stats.SparseResults).
		Int("fused_results", stats.FusedResults).
		Int64("total_time_ms", stats.TotalTimeMS).
		Msg("Hybrid search completed")

	return &SearchResponse{
		Results:      searchResults,
		TotalResults: len(searchResults),
		SearchTimeMS: stats.TotalTimeMS,
		Stats:        stats,
	}, nil
}

// DenseOnly performs pure dense vector search without sparse or RRF.
func (h *HybridSearcher) DenseOnly(ctx context.Context, query *db.SearchQuery) (*SearchResponse, error) {
	if len(query.Embedding) == 0 {
		return nil, ErrEmptyEmbedding
	}

	if query.TopK < 1 {
		query.TopK = h.cfg.DefaultTopK
	}

	denseResults, duration, err := h.denseSearcher.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	searchResults := denseOnlyToResults(denseResults, query.TopK)

	return &SearchResponse{
		Results:      searchResults,
		TotalResults: len(searchResults),
		SearchTimeMS: duration.Milliseconds(),
	}, nil
}

// SparseOnly performs pure sparse keyword search without dense or RRF.
func (h *HybridSearcher) SparseOnly(ctx context.Context, query *db.SearchQuery) (*SearchResponse, error) {
	if len(query.Keywords) == 0 {
		return nil, ErrEmptyKeywords
	}

	if query.TopK < 1 {
		query.TopK = h.cfg.DefaultTopK
	}

	sparseResults, duration, err := h.sparseSearcher.Search(ctx, query)
	if err != nil {
		return nil, err
	}

	searchResults := sparseOnlyToResults(sparseResults)

	return &SearchResponse{
		Results:      searchResults,
		TotalResults: len(searchResults),
		SearchTimeMS: duration.Milliseconds(),
	}, nil
}

// denseOnlyToResults converts dense search results to unified SearchResult format.
func denseOnlyToResults(results []db.DenseSearchResult, topK int) []SearchResult {
	if topK > 0 && len(results) > topK {
		results = results[:topK]
	}

	searchResults := make([]SearchResult, 0, len(results))
	for i, r := range results {
		denseRank := i + 1
		cosineDist := r.CosineDist
		searchResults = append(searchResults, SearchResult{
			ParentID:      r.ParentID.String(),
			Content:       r.ParentContent,
			ParentContent: r.ParentContent,
			DenseScore:    &cosineDist,
			DenseRank:     &denseRank,
			RRFScore:      1.0 / (60.0 + float64(denseRank)),
			Metadata: map[string]interface{}{
				"page_number":  r.PageNumber,
				"content_type": r.ContentType,
				"chapter":      r.Chapter,
				"section":      r.Section,
				"subsection":   r.Subsection,
				"grade":        r.Grade,
				"subject":      r.Subject,
			},
		})
	}

	return searchResults
}

// sparseOnlyToResults converts sparse search results to unified SearchResult format.
func sparseOnlyToResults(results []db.SparseSearchResult) []SearchResult {
	searchResults := make([]SearchResult, 0, len(results))
	for i, r := range results {
		sparseRank := i + 1
		keywordScore := r.KeywordScore
		searchResults = append(searchResults, SearchResult{
			ParentID:      r.ParentID.String(),
			Content:       r.Content,
			ParentContent: r.Content,
			SparseScore:   &keywordScore,
			SparseRank:    &sparseRank,
			RRFScore:      1.0 / (60.0 + float64(sparseRank)),
			Metadata: map[string]interface{}{
				"page_number":  r.PageNumber,
				"content_type": r.ContentType,
				"chapter":      r.Chapter,
				"section":      r.Section,
				"subsection":   r.Subsection,
				"grade":        r.Grade,
				"subject":      r.Subject,
				"match_count":  r.MatchCount,
			},
		})
	}

	return searchResults
}
