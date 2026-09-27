// Package chunking provides token-based chunking strategies for document ingestion.
// P3-30: Token-based splitting using tiktoken-compatible token counting.
package chunking

import (
	"strings"
	"unicode/utf8"
)

// TokenChunker splits text based on token count rather than character count.
// More precise for LLM context window alignment.
type TokenChunker struct {
	tokensPerChunk int
	overlapTokens  int
}

// TokenChunkerConfig configures the token chunker.
type TokenChunkerConfig struct {
	TokensPerChunk int
	OverlapTokens  int
}

// NewTokenChunker creates a token-based chunker.
func NewTokenChunker(cfg TokenChunkerConfig) *TokenChunker {
	if cfg.TokensPerChunk == 0 {
		cfg.TokensPerChunk = 500
	}
	if cfg.OverlapTokens == 0 {
		cfg.OverlapTokens = 100
	}
	return &TokenChunker{
		tokensPerChunk: cfg.TokensPerChunk,
		overlapTokens:  cfg.OverlapTokens,
	}
}

// TokenChunk represents a chunk with token count information.
type TokenChunk struct {
	Content    string `json:"content"`
	TokenCount int    `json:"token_count"`
	ChunkIndex int    `json:"chunk_index"`
}

// Chunk splits text into token-based chunks.
// Uses approximate token counting (4 chars ≈ 1 token for English text).
// In production, use tiktoken or equivalent for precise counting.
func (tc *TokenChunker) Chunk(text string) []TokenChunk {
	// Estimate tokens (4 chars per token for English)
	approxTokens := utf8.RuneCountInString(text) / 4

	if approxTokens <= tc.tokensPerChunk {
		return []TokenChunk{{
			Content:    text,
			TokenCount: approxTokens,
			ChunkIndex: 0,
		}}
	}

	chunks := make([]TokenChunk, 0)
	words := strings.Fields(text)
	currentTokens := 0
	chunkStart := 0

	for i, word := range words {
		wordTokens := utf8.RuneCountInString(word)/4 + 1
		currentTokens += wordTokens

		if currentTokens >= tc.tokensPerChunk {
			// Create chunk from chunkStart to i
			content := strings.Join(words[chunkStart:i+1], " ")
			chunks = append(chunks, TokenChunk{
				Content:    content,
				TokenCount: currentTokens,
				ChunkIndex: len(chunks),
			})

			// Move back for overlap
			overlapStart := i
			overlapTokens := 0
			for overlapStart > chunkStart && overlapTokens < tc.overlapTokens {
				overlapTokens += utf8.RuneCountInString(words[overlapStart])/4 + 1
				overlapStart--
			}

			chunkStart = overlapStart
			currentTokens = overlapTokens
		}
	}

	// Add remaining text
	if chunkStart < len(words) {
		content := strings.Join(words[chunkStart:], " ")
		chunks = append(chunks, TokenChunk{
			Content:    content,
			TokenCount: utf8.RuneCountInString(content) / 4,
			ChunkIndex: len(chunks),
		})
	}

	return chunks
}
