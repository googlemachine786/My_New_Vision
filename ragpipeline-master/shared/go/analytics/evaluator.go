// Package analytics provides RAG evaluation metrics including Recall@K, MRR, and NDCG.
// Ported from Python eval/recall_eval.py with equivalent functionality.
package analytics

import (
	"math"
	"slices"
)

// QAPair represents a question-answer pair for evaluation.
type QAPair struct {
	Question       string   `json:"question"`
	Answer         string   `json:"answer"`
	Grade          int      `json:"grade"`
	Subject        string   `json:"subject"`
	GoldChunkIDs   []string `json:"gold_chunk_ids"`
}

// CalculateRecallAtK calculates Recall@K metric.
// Recall@K measures the proportion of relevant items retrieved in top-K results.
func CalculateRecallAtK(retrieved []string, relevant []string, k int) float64 {
	if len(retrieved) == 0 || len(relevant) == 0 {
		return 0.0
	}

	// Limit to top-k
	if k < len(retrieved) {
		retrieved = retrieved[:k]
	}

	// Count hits
	relevantSet := make(map[string]struct{}, len(relevant))
	for _, r := range relevant {
		relevantSet[r] = struct{}{}
	}

	hits := 0
	for _, r := range retrieved {
		if _, ok := relevantSet[r]; ok {
			hits++
		}
	}

	return float64(hits) / float64(len(relevant))
}

// CalculateMRR calculates Mean Reciprocal Rank.
// MRR measures the rank position of the first relevant item.
func CalculateMRR(retrieved []string, relevant []string) float64 {
	if len(retrieved) == 0 || len(relevant) == 0 {
		return 0.0
	}

	relevantSet := make(map[string]struct{}, len(relevant))
	for _, r := range relevant {
		relevantSet[r] = struct{}{}
	}

	// Find first relevant result
	for i, r := range retrieved {
		if _, ok := relevantSet[r]; ok {
			return 1.0 / float64(i+1)
		}
	}

	return 0.0
}

// CalculateNDCG calculates Normalized Discounted Cumulative Gain.
// NDCG measures the quality of ranking with position-based discounting.
func CalculateNDCG(retrieved []string, relevant []string) float64 {
	if len(retrieved) == 0 || len(relevant) == 0 {
		return 0.0
	}

	relevantSet := make(map[string]struct{}, len(relevant))
	for _, r := range relevant {
		relevantSet[r] = struct{}{}
	}

	// DCG (Discounted Cumulative Gain)
	dcg := 0.0
	for i, r := range retrieved {
		if _, ok := relevantSet[r]; ok {
			dcg += 1.0 / math.Log2(float64(i+2))
		}
	}

	// Ideal DCG
	idealCount := len(relevant)
	if idealCount > len(retrieved) {
		idealCount = len(retrieved)
	}
	
	idcg := 0.0
	for i := 0; i < idealCount; i++ {
		idcg += 1.0 / math.Log2(float64(i+2))
	}

	if idcg == 0 {
		return 0.0
	}

	return dcg / idcg
}

// RetrievalMetrics holds all retrieval evaluation metrics.
type RetrievalMetrics struct {
	RecallAtK map[int]float64 `json:"recall_at_k"`
	MRR       float64         `json:"mrr"`
	NDCG      float64         `json:"ndcg"`
	NumQueries int            `json:"num_queries"`
}

// EvaluateRetrieval evaluates retrieval quality for multiple QA pairs.
func EvaluateRetrieval(qaPairs []QAPair, retrievalFunc func(query string, grade int, subject string) []string, topK []int, verbose bool) *RetrievalMetrics {
	// Accumulators for metrics
	recallAccum := make(map[int][]float64)
	mrrAccum := make([]float64, 0, len(qaPairs))
	ndcgAccum := make([]float64, 0, len(qaPairs))

	// Initialize recall accumulator slices
	for _, k := range topK {
		recallAccum[k] = make([]float64, 0, len(qaPairs))
	}

	for i, qa := range qaPairs {
		if verbose && (i+1)%10 == 0 {
			// Note: In production, use proper logging
			// fmt.Printf("Evaluating %d/%d...\n", i+1, len(qaPairs))
		}

		// Retrieve chunks
		retrievedIDs := retrievalFunc(qa.Question, qa.Grade, qa.Subject)

		// Calculate metrics
		for _, k := range topK {
			recall := CalculateRecallAtK(retrievedIDs, qa.GoldChunkIDs, k)
			recallAccum[k] = append(recallAccum[k], recall)
		}

		mrrAccum = append(mrrAccum, CalculateMRR(retrievedIDs, qa.GoldChunkIDs))
		ndcgAccum = append(ndcgAccum, CalculateNDCG(retrievedIDs, qa.GoldChunkIDs))
	}

	// Aggregate metrics into final result
	aggregated := &RetrievalMetrics{
		RecallAtK: make(map[int]float64),
		MRR:       0.0,
		NDCG:      0.0,
		NumQueries: len(qaPairs),
	}

	for k, values := range recallAccum {
		aggregated.RecallAtK[k] = mean(values)
	}

	if len(mrrAccum) > 0 {
		aggregated.MRR = mean(mrrAccum)
	}

	if len(ndcgAccum) > 0 {
		aggregated.NDCG = mean(ndcgAccum)
	}

	return aggregated
}

// mean calculates the mean of a slice of float64 values.
func mean(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// StdDev calculates the standard deviation of a slice of float64 values.
func StdDev(values []float64) float64 {
	if len(values) <= 1 {
		return 0.0
	}
	m := mean(values)
	sumSquares := 0.0
	for _, v := range values {
		diff := v - m
		sumSquares += diff * diff
	}
	return math.Sqrt(sumSquares / float64(len(values)-1))
}

// Min returns the minimum value in a slice of float64.
func Min(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	return slices.Min(values)
}

// Max returns the maximum value in a slice of float64.
func Max(values []float64) float64 {
	if len(values) == 0 {
		return 0.0
	}
	max := values[0]
	for _, v := range values[1:] {
		if v > max {
			max = v
		}
	}
	return max
}
