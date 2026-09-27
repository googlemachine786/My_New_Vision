// Package search provides BM25 scoring for keyword-based retrieval.
package search

import (
	"math"
	"strings"

	"github.com/visionary/ragpipeline/services/vector-search-service/index"
)

// BM25Scorer implements Okapi BM25 scoring over an inverted keyword index.
type BM25Scorer struct {
	index  *index.KeywordIndex
	k1     float64 // term frequency saturation parameter (default 1.5)
	b      float64 // document length normalization parameter (default 0.75)
	avgDL  float64 // average document length
}

// BM25Result holds a single BM25 scoring result.
type BM25Result struct {
	DocID    string
	Score    float64
	Rank     int
}

// NewBM25Scorer creates a new BM25 scorer with the given keyword index.
func NewBM25Scorer(idx *index.KeywordIndex, k1, b float64) *BM25Scorer {
	if k1 <= 0 {
		k1 = 1.5
	}
	if b < 0 || b > 1 {
		b = 0.75
	}

	avgDL := idx.AverageDocLength()

	return &BM25Scorer{
		index: idx,
		k1:    k1,
		b:     b,
		avgDL: avgDL,
	}
}

// Score computes BM25 scores for all documents matching the given query terms.
// Results are sorted by score in descending order.
func (s *BM25Scorer) Score(queryTerms []string) []BM25Result {
	if len(queryTerms) == 0 || s.index == nil {
		return nil
	}

	// Collect all matching doc IDs
	docScores := make(map[string]float64)

	for _, term := range queryTerms {
		normalizedTerm := strings.ToLower(term)
		df := s.index.DocFrequency(normalizedTerm)
		n := s.index.NumDocs()

		// IDF component: log((N - df + 0.5) / (df + 0.5))
		idf := math.Log((float64(n) - float64(df) + 0.5) / (float64(df) + 0.5))
		if idf < 0 {
			idf = 0
		}

		// Get postings for this term
		postings := s.index.GetPostings(normalizedTerm)
		for docID, tf := range postings {
			docLen := s.index.DocLength(docID)

			// TF component: (tf * (k1 + 1)) / (tf + k1 * (1 - b + b * |d|/avgdl))
			num := float64(tf) * (s.k1 + 1.0)
			denom := float64(tf) + s.k1*(1.0-s.b+s.b*float64(docLen)/s.avgDL)
			tfComponent := num / denom

			docScores[docID] += idf * tfComponent
		}
	}

	// Convert map to sorted slice
	results := make([]BM25Result, 0, len(docScores))
	for docID, score := range docScores {
		results = append(results, BM25Result{
			DocID: docID,
			Score: score,
		})
	}

	// Sort by score descending
	bm25Sort(results)

	// Assign ranks
	for i := range results {
		results[i].Rank = i + 1
	}

	return results
}

// ScoreWithQuery normalizes the query into terms and scores documents.
func (s *BM25Scorer) ScoreWithQuery(query string) []BM25Result {
	terms := tokenize(query)
	return s.Score(terms)
}

// tokenize splits a query into lowercase tokens (words).
func tokenize(query string) []string {
	fields := strings.Fields(strings.ToLower(query))
	tokens := make([]string, 0, len(fields))
	for _, f := range fields {
		// Strip common punctuation
		f = strings.Trim(f, ".,;:!?\"'()[]{}")
		if len(f) > 0 {
			tokens = append(tokens, f)
		}
	}
	return tokens
}

// bm25Sort sorts results by score descending using a simple insertion sort.
// For small result sets (typical in search), this is efficient enough.
func bm25Sort(results []BM25Result) {
	for i := 1; i < len(results); i++ {
		key := results[i]
		j := i - 1
		for j >= 0 && results[j].Score < key.Score {
			results[j+1] = results[j]
			j--
		}
		results[j+1] = key
	}
}
