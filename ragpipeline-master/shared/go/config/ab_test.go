// Package config provides A/B testing configuration for RAG validation.
// Ported from Python ab_test_config.py with equivalent functionality.
package config

import "strconv"

// RAGConfigAB holds RAG configuration for A/B testing.
type RAGConfigAB struct {
	Name                string
	ChunkSize           int
	ChunkOverlap        int
	EmbedModel          string
	TopK                int
	RRFK                float64
	SimilarityThreshold float64
	RetrievalStrategy   string
	Temperature         float64
	MaxTokens           int
}

// String returns a human-readable config summary.
func (c *RAGConfigAB) String() string {
	embedName := c.EmbedModel
	if len(embedName) > 15 {
		embedName = embedName[:15]
	}
	return c.Name + ": chunk=" + strconv.Itoa(c.ChunkSize) +
		", overlap=" + strconv.Itoa(c.ChunkOverlap) +
		", top_k=" + strconv.Itoa(c.TopK) +
		", embed=" + embedName
}

// ConfigOld returns the old (suboptimal) configuration for A/B testing.
func ConfigOld() *RAGConfigAB {
	return &RAGConfigAB{
		Name:                "OLD (suboptimal)",
		ChunkSize:           512,
		ChunkOverlap:        77,
		EmbedModel:          "nomic-embed-text",
		TopK:                3,
		RRFK:                60.0,
		SimilarityThreshold: 0.5,
		RetrievalStrategy:   "hybrid",
		Temperature:         0.7,
		MaxTokens:           150,
	}
}

// ConfigNew returns the new (optimal) configuration for A/B testing.
// Based on grid search optimization results.
func ConfigNew() *RAGConfigAB {
	return &RAGConfigAB{
		Name:                "NEW (optimal)",
		ChunkSize:           400,
		ChunkOverlap:        150,
		EmbedModel:          "sentence-transformers/all-MiniLM-L6-v2",
		TopK:                5,
		RRFK:                60.0,
		SimilarityThreshold: 0.36,
		RetrievalStrategy:   "dense",
		Temperature:         0.7,
		MaxTokens:           150,
	}
}

// GoldenQuestions holds the golden QA dataset for A/B testing.
var GoldenQuestions = []string{
	"What is photosynthesis and how do plants make food?",
	"What are the parts of a cell and their functions?",
	"What is force and what are its effects on motion?",
	"What is combustion and what does it produce?",
	"What are microorganisms and where can they be found?",
	"What is friction and how does it affect moving objects?",
	"How do humans pollute the air and what are the effects?",
	"What is sound and how is it produced?",
	"What is light and how does reflection work?",
	"What are the properties of metals vs non-metals?",
	"What is irrigation and what are its different methods?",
	"What are kharif and rabi crops? Give examples.",
	"What is the structure of a plant cell?",
	"What is the law of reflection?",
	"How does electroplating work and what is it used for?",
	"What causes earthquakes and how are they measured?",
	"What is deforestation and what are its consequences?",
	"What are fossil fuels and how are they formed?",
	"What is the difference between rolling and sliding friction?",
	"What are the different types of microorganisms?",
}

// MetricWeights holds the weights for composite score calculation.
var MetricWeights = map[string]float64{
	"rouge_l":             0.06,
	"bert_score":          0.12,
	"semantic_similarity": 0.08,
	"ndcg":                0.14,
	"recall_at_k":         0.13,
	"context_precision":   0.12,
	"faithfulness":        0.18,
}

// TestResult holds the result for a single test question.
type TestResult struct {
	Question             string  `json:"question"`
	ConfigName           string  `json:"config_name"`
	RecallAtK            float64 `json:"recall_at_k"`
	ContextPrecision     float64 `json:"context_precision"`
	NDCG                 float64 `json:"ndcg"`
	ROUGEL               float64 `json:"rouge_l"`
	BERTScore            float64 `json:"bert_score"`
	SemanticSimilarity   float64 `json:"semantic_similarity"`
	Faithfulness         float64 `json:"faithfulness"`
	RetrievalLatencyMs   float64 `json:"retrieval_latency_ms"`
	GenerationLatencyMs  float64 `json:"generation_latency_ms"`
	TotalLatencyMs       float64 `json:"total_latency_ms"`
	CompositeScore       float64 `json:"composite_score"`
	RetrievedChunks      int     `json:"retrieved_chunks"`
	AnswerLength         int     `json:"answer_length"`
	Error                string  `json:"error,omitempty"`
}

// MetricStats holds aggregated statistics for a metric.
type MetricStats struct {
	Mean float64 `json:"mean"`
	Std  float64 `json:"std"`
	Min  float64 `json:"min"`
	Max  float64 `json:"max"`
}

// ABTestReport holds the aggregated A/B test report.
type ABTestReport struct {
	Old          map[string]MetricStats `json:"old"`
	New          map[string]MetricStats `json:"new"`
	Improvements map[string]float64     `json:"improvements"`
}
