// Package config provides centralized RAG configuration with presets and validation.
// Ported from Python rag_config.py with equivalent functionality.
package config

import (
	"fmt"
	"os"
	"strconv"
)

// RetrievalStrategy represents the type of retrieval to use.
type RetrievalStrategy string

const (
	Dense  RetrievalStrategy = "dense"
	Hybrid RetrievalStrategy = "hybrid"
	BM25   RetrievalStrategy = "bm25"
)

// RAGConfig holds production RAG configuration optimized for CBSE Science (Grades 6-8).
// All parameters are optimized based on empirical evaluation data.
type RAGConfig struct {
	// Chunking (MOST CRITICAL - 40% of performance variance)
	ChunkSize    int `json:"chunk_size"`
	ChunkOverlap int `json:"chunk_overlap"`

	// Embedding (35% of performance variance)
	EmbedModel     string `json:"embed_model"`
	EmbedDimension int    `json:"embed_dimension"`

	// Retrieval (15% of performance variance)
	RetrievalStrategy RetrievalStrategy `json:"retrieval_strategy"`
	TopK              int               `json:"top_k"`
	RRFK              float64           `json:"rrf_k"`
	SimilarityThreshold float64         `json:"similarity_threshold"`
	BM25K1            float64           `json:"bm25_k1"`

	// Generation (10% of performance variance)
	Temperature float64 `json:"temperature"`
	MaxTokens   int     `json:"max_tokens"`

	// Performance
	BatchSize        int  `json:"batch_size"`
	CacheEnabled     bool `json:"cache_enabled"`
	CacheTTLSeconds  int  `json:"cache_ttl_seconds"`

	// Metadata
	Version string `json:"version"`
}

// Default returns the default production configuration (balanced).
func Default() *RAGConfig {
	return Balanced()
}

// Balanced returns balanced configuration (same as default).
// Performance: 0.8386 composite score, 547ms P95 latency
// Best overall trade-off between quality and speed.
func Balanced() *RAGConfig {
	return &RAGConfig{
		ChunkSize:           400,
		ChunkOverlap:        150,
		EmbedModel:          "sentence-transformers/all-MiniLM-L6-v2",
		EmbedDimension:      384,
		RetrievalStrategy:   Dense,
		TopK:                5,
		RRFK:                60,
		SimilarityThreshold: 0.364,
		BM25K1:              2.05,
		Temperature:         0.75,
		MaxTokens:           143,
		BatchSize:           8,
		CacheEnabled:        true,
		CacheTTLSeconds:     300,
		Version:             "5.0-balanced",
	}
}

// LowLatency returns configuration optimized for speed.
// Performance: 0.8329 composite score, 305ms P95 latency
// Trade-off: -0.7% quality for -44% latency
func LowLatency() *RAGConfig {
	return &RAGConfig{
		ChunkSize:           400,
		ChunkOverlap:        200,
		EmbedModel:          "sentence-transformers/all-MiniLM-L6-v2",
		EmbedDimension:      384,
		RetrievalStrategy:   Dense,
		TopK:                8,
		RRFK:                60,
		SimilarityThreshold: 0.35,
		BM25K1:              2.05,
		Temperature:         0.7,
		MaxTokens:           100,
		BatchSize:           8,
		CacheEnabled:        true,
		CacheTTLSeconds:     300,
		Version:             "5.0-low-latency",
	}
}

// HighQuality returns configuration optimized for accuracy.
// Performance: 0.8374 composite score, 705ms P95 latency
// Trade-off: +37% latency for -0.1% quality
func HighQuality() *RAGConfig {
	return &RAGConfig{
		ChunkSize:           400,
		ChunkOverlap:        150,
		EmbedModel:          "sentence-transformers/all-MiniLM-L6-v2",
		EmbedDimension:      384,
		RetrievalStrategy:   Dense,
		TopK:                7,
		RRFK:                60,
		SimilarityThreshold: 0.32,
		BM25K1:              2.05,
		Temperature:         0.63,
		MaxTokens:           146,
		BatchSize:           4,
		CacheEnabled:        true,
		CacheTTLSeconds:     300,
		Version:             "5.0-high-quality",
	}
}

// FromEnv loads configuration from environment variables.
func FromEnv() *RAGConfig {
	cfg := Default()

	if val := os.Getenv("PARENT_MAX_CHARS"); val != "" {
		if size, err := strconv.Atoi(val); err == nil {
			cfg.ChunkSize = size
		}
	}

	if val := os.Getenv("CHILD_OVERLAP"); val != "" {
		if overlap, err := strconv.Atoi(val); err == nil {
			cfg.ChunkOverlap = overlap
		}
	}

	if val := os.Getenv("EMBED_MODEL"); val != "" {
		cfg.EmbedModel = val
	}

	if val := os.Getenv("RETRIEVAL_STRATEGY"); val != "" {
		cfg.RetrievalStrategy = RetrievalStrategy(val)
	}

	if val := os.Getenv("TOP_K"); val != "" {
		if topK, err := strconv.Atoi(val); err == nil {
			cfg.TopK = topK
		}
	}

	if val := os.Getenv("RRF_K"); val != "" {
		if rrfK, err := strconv.ParseFloat(val, 64); err == nil {
			cfg.RRFK = rrfK
		}
	}

	if val := os.Getenv("SIMILARITY_THRESHOLD"); val != "" {
		if threshold, err := strconv.ParseFloat(val, 64); err == nil {
			cfg.SimilarityThreshold = threshold
		}
	}

	if val := os.Getenv("BM25_K1"); val != "" {
		if bm25K1, err := strconv.ParseFloat(val, 64); err == nil {
			cfg.BM25K1 = bm25K1
		}
	}

	if val := os.Getenv("LLM_TEMPERATURE"); val != "" {
		if temp, err := strconv.ParseFloat(val, 64); err == nil {
			cfg.Temperature = temp
		}
	}

	if val := os.Getenv("MAX_TOKENS"); val != "" {
		if maxTokens, err := strconv.Atoi(val); err == nil {
			cfg.MaxTokens = maxTokens
		}
	}

	return cfg
}

// Validate validates configuration parameters.
// Returns (isValid, validationErrors).
func (c *RAGConfig) Validate() (bool, []string) {
	var errs []string

	// Chunking validation
	if c.ChunkSize < 100 {
		errs = append(errs, fmt.Sprintf("chunk_size (%d) too small, minimum 100", c.ChunkSize))
	}
	if c.ChunkSize > 2000 {
		errs = append(errs, fmt.Sprintf("chunk_size (%d) too large, maximum 2000", c.ChunkSize))
	}
	if c.ChunkOverlap >= c.ChunkSize {
		errs = append(errs, fmt.Sprintf("chunk_overlap (%d) must be < chunk_size (%d)", c.ChunkOverlap, c.ChunkSize))
	}
	if c.ChunkOverlap < 0 {
		errs = append(errs, "chunk_overlap must be non-negative")
	}

	// Overlap ratio check (optimal: 30-40%)
	overlapRatio := float64(c.ChunkOverlap) / float64(c.ChunkSize)
	if overlapRatio < 0.2 || overlapRatio > 0.5 {
		errs = append(errs, fmt.Sprintf("overlap ratio (%.0f%%) outside recommended range (20-50%%)", overlapRatio*100))
	}

	// Embedding validation
	if c.EmbedDimension <= 0 {
		errs = append(errs, "embed_dimension must be positive")
	}

	// Retrieval validation
	if c.TopK < 1 {
		errs = append(errs, "top_k must be >= 1")
	}
	if c.TopK > 20 {
		errs = append(errs, fmt.Sprintf("top_k (%d) too large, consider <= 10", c.TopK))
	}
	if c.RRFK <= 0 {
		errs = append(errs, "rrf_k must be positive")
	}
	if c.SimilarityThreshold < 0.0 || c.SimilarityThreshold > 1.0 {
		errs = append(errs, "similarity_threshold must be between 0 and 1")
	}
	if c.BM25K1 <= 0 {
		errs = append(errs, "bm25_k1 must be positive")
	}

	// Generation validation
	if c.Temperature < 0.0 || c.Temperature > 2.0 {
		errs = append(errs, "temperature must be between 0 and 2")
	}
	if c.MaxTokens <= 0 {
		errs = append(errs, "max_tokens must be positive")
	}
	if c.MaxTokens > 4096 {
		errs = append(errs, fmt.Sprintf("max_tokens (%d) exceeds model limit", c.MaxTokens))
	}

	// Performance validation
	if c.BatchSize < 1 {
		errs = append(errs, "batch_size must be >= 1")
	}
	if c.BatchSize > 64 {
		errs = append(errs, fmt.Sprintf("batch_size (%d) too large for most GPUs", c.BatchSize))
	}

	return len(errs) == 0, errs
}

// String returns a human-readable configuration summary.
func (c *RAGConfig) String() string {
	return fmt.Sprintf(
		"RAGConfig v%s\n"+
			"==================================================\n"+
			"Chunking:     %d chars (+%d overlap)\n"+
			"Embedding:    %s (%dd)\n"+
			"Retrieval:    %s (top_k=%d, rrf_k=%.0f)\n"+
			"Similarity:   threshold=%.3f\n"+
			"Generation:   temp=%.2f, max_tokens=%d\n"+
			"Performance:  batch_size=%d, cache=%v",
		c.Version,
		c.ChunkSize, c.ChunkOverlap,
		c.EmbedModel, c.EmbedDimension,
		c.RetrievalStrategy, c.TopK, c.RRFK,
		c.SimilarityThreshold,
		c.Temperature, c.MaxTokens,
		c.BatchSize, c.CacheEnabled,
	)
}

// GetConfig returns configuration by preset name.
func GetConfig(preset string) (*RAGConfig, error) {
	switch preset {
	case "balanced":
		return Balanced(), nil
	case "low_latency":
		return LowLatency(), nil
	case "high_quality":
		return HighQuality(), nil
	default:
		return nil, fmt.Errorf("unknown preset: %s. Available: balanced, low_latency, high_quality", preset)
	}
}
