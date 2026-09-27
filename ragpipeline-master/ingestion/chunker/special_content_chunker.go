// Package chunking provides special content-aware chunking for code and formulas.
// P3-31: Syntax-aware splitting that preserves completeness of code/formulas.
package chunking

import (
	"regexp"
	"strings"
)

// SpecialContentChunker handles code blocks, formulas, and other special content.
// Preserves syntactic完整性 by not splitting mid-structure.
type SpecialContentChunker struct {
	baseChunker Chunker
	maxCodeLen  int
}

// Chunker interface for the base chunker.
type Chunker interface {
	Chunk(text string) []Chunk
}

// Chunk represents a chunk with content type information.
type Chunk struct {
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
	IsAtomic    bool   `json:"is_atomic"` // True if should not be split further
}

// NewSpecialContentChunker creates a special content chunker.
func NewSpecialContentChunker(baseChunker Chunker, maxCodeLen int) *SpecialContentChunker {
	if maxCodeLen == 0 {
		maxCodeLen = 1000
	}
	return &SpecialContentChunker{
		baseChunker: baseChunker,
		maxCodeLen:  maxCodeLen,
	}
}

// Chunk splits text while preserving special content blocks.
func (sc *SpecialContentChunker) Chunk(text string) []Chunk {
	// Detect and extract special content blocks
	blocks := sc.extractBlocks(text)

	chunks := make([]Chunk, 0)
	for _, block := range blocks {
		if block.IsAtomic || len(block.Content) <= sc.maxCodeLen {
			// Keep atomic blocks intact
			chunks = append(chunks, block)
		} else {
			// Split large blocks at safe boundaries
			subChunks := sc.splitSafely(block.Content, block.ContentType)
			for _, sub := range subChunks {
				chunks = append(chunks, Chunk{
					Content:     sub,
					ContentType: block.ContentType,
					IsAtomic:    false,
				})
			}
		}
	}

	return chunks
}

// Block represents a content block with type information.
type Block struct {
	Content     string
	ContentType string
	IsAtomic    bool
}

// extractBlocks detects and extracts special content blocks.
func (sc *SpecialContentChunker) extractBlocks(text string) []Block {
	blocks := make([]Block, 0)

	// Code block pattern: ```language\ncode\n```
	codeBlockRe := regexp.MustCompile("```[a-z]*\n(.*?)\n```")
	codeBlocks := codeBlockRe.FindAllStringSubmatch(text, -1)

	// Formula patterns: $$...$$, \(...\), \[...\]
	formulaBlockRe := regexp.MustCompile(`\$\$(.*?)\$\$`)
	formulaBlocks := formulaBlockRe.FindAllStringSubmatch(text, -1)

	// Remove special blocks from text for regular chunking
	remainingText := text
	for _, match := range codeBlocks {
		blocks = append(blocks, Block{
			Content:     match[0],
			ContentType: "code",
			IsAtomic:    true,
		})
		remainingText = strings.Replace(remainingText, match[0], "", 1)
	}

	for _, match := range formulaBlocks {
		blocks = append(blocks, Block{
			Content:     match[0],
			ContentType: "formula",
			IsAtomic:    true,
		})
		remainingText = strings.Replace(remainingText, match[0], "", 1)
	}

	// Chunk remaining text normally
	if strings.TrimSpace(remainingText) != "" {
		if sc.baseChunker != nil {
			baseChunks := sc.baseChunker.Chunk(remainingText)
			for _, c := range baseChunks {
				blocks = append(blocks, Block{
					Content:     c.Content,
					ContentType: "text",
					IsAtomic:    false,
				})
			}
		} else {
			blocks = append(blocks, Block{
				Content:     remainingText,
				ContentType: "text",
				IsAtomic:    false,
			})
		}
	}

	return blocks
}

// splitSafely splits content at safe boundaries.
func (sc *SpecialContentChunker) splitSafely(content, contentType string) []string {
	switch contentType {
	case "code":
		return sc.splitCodeSafely(content)
	case "formula":
		return []string{content} // Keep formulas atomic
	default:
		return []string{content}
	}
}

// splitCodeSafely splits code at function/class boundaries.
func (sc *SpecialContentChunker) splitCodeSafely(code string) []string {
	// Try to split at function boundaries (def, function, class, etc.)
	funcPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?m)^def `),
		regexp.MustCompile(`(?m)^class `),
		regexp.MustCompile(`(?m)^function `),
		regexp.MustCompile(`(?m)^\w+\s*\(.*\)\s*{`),
	}

	for _, pattern := range funcPatterns {
		matches := pattern.FindAllStringIndex(code, -1)
		if len(matches) > 1 {
			// Split at function boundaries
			chunks := make([]string, 0)
			for i, match := range matches {
				start := match[0]
				var end int
				if i+1 < len(matches) {
					end = matches[i+1][0]
				} else {
					end = len(code)
				}

				chunk := strings.TrimSpace(code[start:end])
				if len(chunk) > 0 {
					chunks = append(chunks, chunk)
				}
			}
			return chunks
		}
	}

	// Fallback: don't split
	return []string{code}
}
