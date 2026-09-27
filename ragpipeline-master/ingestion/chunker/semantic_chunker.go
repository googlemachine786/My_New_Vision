// Package chunking provides semantic chunking strategies for document ingestion.
// P3-24: Semantic chunking using sentence embedding similarity with breakpoint threshold.
package chunking

import (
	"math"
	"regexp"
	"strings"
)

// SemanticChunker splits text based on semantic similarity between sentences.
// Uses cosine similarity of sentence embeddings to detect topic boundaries.
type SemanticChunker struct {
	breakpointPercentileThreshold float64 // Threshold for detecting chunk boundaries (0.0-1.0)
	minChunkSize                 int     // Minimum sentences per chunk
	maxChunkSize                 int     // Maximum sentences per chunk
}

// SemanticChunkerConfig configures the semantic chunker.
type SemanticChunkerConfig struct {
	BreakpointPercentileThreshold float64
	MinChunkSize                 int
	MaxChunkSize                 int
}

// NewSemanticChunker creates a semantic chunker.
func NewSemanticChunker(cfg SemanticChunkerConfig) *SemanticChunker {
	if cfg.BreakpointPercentileThreshold == 0 {
		cfg.BreakpointPercentileThreshold = 0.95
	}
	if cfg.MinChunkSize == 0 {
		cfg.MinChunkSize = 3
	}
	if cfg.MaxChunkSize == 0 {
		cfg.MaxChunkSize = 20
	}
	return &SemanticChunker{
		breakpointPercentileThreshold: cfg.BreakpointPercentileThreshold,
		minChunkSize:                 cfg.MinChunkSize,
		maxChunkSize:                 cfg.MaxChunkSize,
	}
}

// SemanticChunk represents a chunk with semantic boundary information.
type SemanticChunk struct {
	Content        string   `json:"content"`
	SentenceCount  int      `json:"sentence_count"`
	BoundaryScore float64  `json:"boundary_score"` // Similarity drop at boundary
}

// Chunk splits text into semantically coherent chunks.
// In production, embeddings would be computed via an embedding model.
// This implementation uses a simplified heuristic-based approach that
// can be replaced with actual embedding similarity.
func (sc *SemanticChunker) Chunk(text string) []SemanticChunk {
	// Split into sentences
	sentences := splitIntoSentences(text)
	if len(sentences) <= sc.minChunkSize {
		return []SemanticChunk{{
			Content:       text,
			SentenceCount: len(sentences),
			BoundaryScore: 0,
		}}
	}

	// In production: compute embeddings for each sentence
	// embeddings := computeEmbeddings(sentences)
	// For now, use keyword-based similarity as proxy
	similarities := computeHeuristicSimilarity(sentences)

	// Find breakpoints (large drops in similarity)
	breakpoints := findBreakpoints(similarities, sc.breakpointPercentileThreshold)

	// Create chunks from breakpoints
	chunks := make([]SemanticChunk, 0)
	start := 0
	for _, bp := range breakpoints {
		end := bp.index + 1
		if end-start < sc.minChunkSize {
			continue
		}
		if end-start > sc.maxChunkSize {
			end = start + sc.maxChunkSize
		}

		content := strings.Join(sentences[start:end], " ")
		chunks = append(chunks, SemanticChunk{
			Content:       content,
			SentenceCount: end - start,
			BoundaryScore: bp.score,
		})
		start = end
	}

	// Add remaining sentences as final chunk
	if start < len(sentences) {
		end := start + sc.maxChunkSize
		if end > len(sentences) {
			end = len(sentences)
		}
		content := strings.Join(sentences[start:end], " ")
		chunks = append(chunks, SemanticChunk{
			Content:       content,
			SentenceCount: end - start,
			BoundaryScore: 0,
		})
	}

	return chunks
}

// sentenceSplitRegex matches sentence-ending punctuation followed by whitespace or end of string.
// Handles periods, exclamation marks, and question marks.
var sentenceSplitRegex = regexp.MustCompile(`[.!?]+\s*`)

// splitIntoSentences splits text into sentences using regex-based sentence boundary detection.
func splitIntoSentences(text string) []string {
	// Split on sentence-ending punctuation (. ! ?) with optional trailing whitespace
	raw := sentenceSplitRegex.Split(text, -1)

	var sentences []string
	for _, s := range raw {
		s = strings.TrimSpace(s)
		if len(s) > 10 {
			sentences = append(sentences, s)
		}
	}

	if len(sentences) == 0 {
		sentences = append(sentences, text)
	}

	return sentences
}

// computeHeuristicSimilarity computes a proxy for sentence similarity
// using keyword overlap. In production, replace with actual embeddings.
func computeHeuristicSimilarity(sentences []string) []SimilarityPair {
	pairs := make([]SimilarityPair, 0, len(sentences)-1)

	for i := 0; i < len(sentences)-1; i++ {
		sim := keywordOverlap(sentences[i], sentences[i+1])
		pairs = append(pairs, SimilarityPair{index: i, score: sim})
	}

	return pairs
}

// keywordOverlap calculates keyword overlap between two sentences.
func keywordOverlap(s1, s2 string) float64 {
	words1 := make(map[string]bool)
	for _, w := range strings.Fields(strings.ToLower(s1)) {
		if len(w) > 3 {
			words1[w] = true
		}
	}

	if len(words1) == 0 {
		return 1.0
	}

	overlap := 0
	for _, w := range strings.Fields(strings.ToLower(s2)) {
		if len(w) > 3 && words1[w] {
			overlap++
		}
	}

	return float64(overlap) / float64(len(words1))
}

// SimilarityPair represents similarity between adjacent sentences.
type SimilarityPair struct {
	index int
	score float64
}

// findBreakpoints finds sentences with large similarity drops.
func findBreakpoints(pairs []SimilarityPair, threshold float64) []SimilarityPair {
	if len(pairs) == 0 {
		return nil
	}

	// Calculate mean and stddev
	mean := 0.0
	for _, p := range pairs {
		mean += p.score
	}
	mean /= float64(len(pairs))

	variance := 0.0
	for _, p := range pairs {
		diff := p.score - mean
		variance += diff * diff
	}
	stddev := math.Sqrt(variance / float64(len(pairs)))

	// Find breakpoints below threshold
	breakpoints := make([]SimilarityPair, 0)
	cutoff := mean - threshold*stddev

	for _, p := range pairs {
		if p.score < cutoff {
			breakpoints = append(breakpoints, p)
		}
	}

	return breakpoints
}
