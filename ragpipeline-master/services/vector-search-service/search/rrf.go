package search

import (
	"sort"

	"github.com/google/uuid"

	"github.com/visionary/ragpipeline/services/vector-search-service/db"
)

// RRFConfig holds configuration for Reciprocal Rank Fusion.
type RRFConfig struct {
	// K is the RRF constant (typically 60).
	// Controls the balance between high-ranking and low-ranking results.
	K float64

	// DenseWeight is the weight applied to dense search scores.
	DenseWeight float64

	// SparseWeight is the weight applied to sparse search scores.
	SparseWeight float64

	// BM25Weight is the weight applied to BM25 scores.
	BM25Weight float64
}

// DefaultRRFConfig returns the default RRF configuration.
func DefaultRRFConfig() RRFConfig {
	return RRFConfig{
		K:            60.0,
		DenseWeight:  1.0,
		SparseWeight: 1.0,
		BM25Weight:   1.0,
	}
}

// rrfEntry holds intermediate RRF data for a single result.
type rrfEntry struct {
	parentID    uuid.UUID
	denseRank   int
	sparseRank  int
	bm25Rank    int
	denseScore  float64
	sparseScore float64
	bm25Score   float64
	rrfScore    float64
	denseResult *db.DenseSearchResult
	sparseResult *db.SparseSearchResult
}

// ReciprocalRankFuse performs Reciprocal Rank Fusion on dense, sparse, and BM25 results.
//
// RRF Formula: score = sum(1 / (k + rank_i)) for each ranking system i
// where k is a constant (typically 60) that dampens the impact of high ranks.
//
// The algorithm:
// 1. Initialize scores for each unique result
// 2. Add dense scores: 1/(k + dense_rank) * dense_weight
// 3. Add sparse scores: 1/(k + sparse_rank) * sparse_weight
// 4. Add BM25 scores: 1/(k + bm25_rank) * bm25_weight
// 5. Sort by combined RRF score (descending)
//
// k=60 is empirically shown to work well across many retrieval scenarios,
// providing a good balance between the ranking systems.
func ReciprocalRankFuse(
	denseResults []db.DenseSearchResult,
	sparseResults []db.SparseSearchResult,
	bm25Results []BM25Result,
	cfg RRFConfig,
) []rrfEntry {
	// Map from parent_id to RRF entry
	entries := make(map[uuid.UUID]*rrfEntry)

	// Process dense results
	for rank, result := range denseResults {
		entry := &rrfEntry{
			parentID:    result.ParentID,
			denseRank:   rank + 1,
			sparseRank:  0,
			bm25Rank:    0,
			denseScore:  1.0 / (cfg.K + float64(rank+1)),
			sparseScore: 0,
			bm25Score:   0,
			denseResult: &result,
		}
		entry.rrfScore = entry.denseScore * cfg.DenseWeight
		entries[result.ParentID] = entry
	}

	// Process sparse results
	for rank, result := range sparseResults {
		sparseScore := 1.0 / (cfg.K + float64(rank+1))

		if entry, exists := entries[result.ParentID]; exists {
			// Merge: add sparse score to existing entry
			entry.sparseRank = rank + 1
			entry.sparseScore = sparseScore
			entry.sparseResult = &result
			entry.rrfScore += sparseScore * cfg.SparseWeight
		} else {
			// New entry from sparse only
			entry := &rrfEntry{
				parentID:     result.ParentID,
				denseRank:    0,
				sparseRank:   rank + 1,
				bm25Rank:     0,
				denseScore:   0,
				sparseScore:  sparseScore,
				bm25Score:    0,
				sparseResult: &result,
			}
			entry.rrfScore = sparseScore * cfg.SparseWeight
			entries[result.ParentID] = entry
		}
	}

	// Process BM25 results
	for rank, result := range bm25Results {
		parentID, err := uuid.Parse(result.DocID)
		if err != nil {
			continue
		}
		bm25Score := 1.0 / (cfg.K + float64(rank+1))

		if entry, exists := entries[parentID]; exists {
			// Merge: add BM25 score to existing entry
			entry.bm25Rank = rank + 1
			entry.bm25Score = bm25Score
			entry.rrfScore += bm25Score * cfg.BM25Weight
		} else {
			// New entry from BM25 only
			entry := &rrfEntry{
				parentID:    parentID,
				denseRank:   0,
				sparseRank:  0,
				bm25Rank:    rank + 1,
				denseScore:  0,
				sparseScore: 0,
				bm25Score:   bm25Score,
			}
			entry.rrfScore = bm25Score * cfg.BM25Weight
			entries[parentID] = entry
		}
	}

	// Convert map to slice
	fused := make([]rrfEntry, 0, len(entries))
	for _, entry := range entries {
		fused = append(fused, *entry)
	}

	// Sort by RRF score descending, then by dense score as tiebreaker
	sort.Slice(fused, func(i, j int) bool {
		if fused[i].rrfScore != fused[j].rrfScore {
			return fused[i].rrfScore > fused[j].rrfScore
		}
		// Tiebreaker: prefer lower cosine distance (higher similarity)
		if fused[i].denseResult != nil && fused[j].denseResult != nil {
			return fused[i].denseResult.CosineDist < fused[j].denseResult.CosineDist
		}
		// If only one has dense result, prefer that one
		if fused[i].denseResult != nil {
			return true
		}
		return false
	})

	return fused
}

// DeduplicateAndLimit removes duplicate parent chunks and limits to topK results.
func DeduplicateAndLimit(fused []rrfEntry, topK int) []rrfEntry {
	if topK <= 0 || len(fused) == 0 {
		return nil
	}

	seen := make(map[uuid.UUID]struct{}, len(fused))
	results := make([]rrfEntry, 0, topK)

	for _, entry := range fused {
		if len(results) >= topK {
			break
		}

		if _, exists := seen[entry.parentID]; !exists {
			seen[entry.parentID] = struct{}{}
			results = append(results, entry)
		}
	}

	return results
}

// ToSearchResults converts RRF entries to the unified SearchResult format.
func ToSearchResults(fused []rrfEntry) []SearchResult {
	results := make([]SearchResult, 0, len(fused))

	for _, entry := range fused {
		result := SearchResult{
			ParentID: entry.parentID.String(),
			RRFScore: entry.rrfScore,
			Metadata: make(map[string]interface{}, 4), // page, chapter, section, grade
		}

		// Prefer content from dense result (has join), fallback to sparse
		if entry.denseResult != nil {
			result.Content = entry.denseResult.Content
			result.ParentContent = entry.denseResult.Content
			if entry.denseResult.PageNumber != nil {
				result.Metadata["page_number"] = *entry.denseResult.PageNumber
			}
			result.Metadata["content_type"] = entry.denseResult.ContentType
			result.Metadata["chapter"] = entry.denseResult.Chapter
			result.Metadata["section"] = entry.denseResult.Section
			result.Metadata["subsection"] = entry.denseResult.Subsection
			result.Metadata["grade"] = entry.denseResult.Grade
			result.Metadata["subject"] = entry.denseResult.Subject
			result.DenseRank = &entry.denseRank
			cosineDist := entry.denseResult.CosineDist
			result.DenseScore = &cosineDist
		}

		if entry.sparseResult != nil {
			// Fill in any missing fields from sparse result
			if result.Content == "" {
				result.Content = entry.sparseResult.Content
				result.ParentContent = entry.sparseResult.Content
			}
			if _, ok := result.Metadata["page_number"]; !ok && entry.sparseResult.PageNumber != nil {
				result.Metadata["page_number"] = *entry.sparseResult.PageNumber
			}
			if _, ok := result.Metadata["content_type"]; !ok {
				result.Metadata["content_type"] = entry.sparseResult.ContentType
			}
			if _, ok := result.Metadata["chapter"]; !ok {
				result.Metadata["chapter"] = entry.sparseResult.Chapter
			}
			if _, ok := result.Metadata["section"]; !ok {
				result.Metadata["section"] = entry.sparseResult.Section
			}
			if _, ok := result.Metadata["subsection"]; !ok {
				result.Metadata["subsection"] = entry.sparseResult.Subsection
			}
			if _, ok := result.Metadata["grade"]; !ok {
				result.Metadata["grade"] = entry.sparseResult.Grade
			}
			if _, ok := result.Metadata["subject"]; !ok {
				result.Metadata["subject"] = entry.sparseResult.Subject
			}
			result.SparseRank = &entry.sparseRank
			sparseScore := entry.sparseResult.KeywordScore
			result.SparseScore = &sparseScore
		}

		// Add BM25 scores if available
		if entry.bm25Rank > 0 {
			bm25Rank := entry.bm25Rank
			result.BM25Rank = &bm25Rank
			bm25Score := entry.bm25Score
			result.BM25Score = &bm25Score
		}

		results = append(results, result)
	}

	return results
}
