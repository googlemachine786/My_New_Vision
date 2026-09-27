# Go Ingestion Pipeline - Implementation Plan

**Priority:** P0-Critical  
**Effort:** 2-3 weeks  
**Owner:** Backend Team (Go)  
**Start Date:** April 1, 2026  
**Target Completion:** April 18, 2026

---

## Overview

This document provides a detailed implementation plan for the Go ingestion pipeline, which is currently missing (only README exists). The Go ingestion pipeline will handle chunking, keyword extraction, embedding, and database writing, while delegating PDF parsing to Python.

---

## Architecture

### Hybrid Approach

```
┌─────────────────────────────────────────────────────────────────┐
│                    HYBRID INGESTION FLOW                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  PDF File                                                       │
│     │                                                           │
│     ▼                                                           │
│  ┌─────────────────────────────────────┐                       │
│  │  Python Parser (Existing)           │                       │
│  │  - PyMuPDF 5-pass extraction        │                       │
│  │  - Font calibration                 │                       │
│  │  - Table extraction                 │                       │
│  │  - Heading mapping                  │                       │
│  │  - Formula detection                │                       │
│  │  - Output: JSON                     │                       │
│  └────────────────┬────────────────────┘                       │
│                   │                                             │
│                   ▼                                             │
│  ┌─────────────────────────────────────┐                       │
│  │  Go Pipeline (NEW)                  │                       │
│  │  ┌───────────────────────────────┐  │                       │
│  │  │ 1. JSON Parser                │  │                       │
│  │  │    - Parse Python output      │  │                       │
│  │  │    - Validate schema          │  │                       │
│  │  │    - Enrich metadata          │  │                       │
│  │  └───────────────┬───────────────┘  │                       │
│  │                  │                    │                       │
│  │  ┌───────────────▼───────────────┐  │                       │
│  │  │ 2. Parent-Child Chunker       │  │                       │
│  │  │    - 1500 char parents        │  │                       │
│  │  │    - 512 char children        │  │                       │
│  │  │    - 77 char overlap (15%)    │  │                       │
│  │  │    - Atomic tables            │  │                       │
│  │  └───────────────┬───────────────┘  │                       │
│  │                  │                    │                       │
│  │  ┌───────────────▼───────────────┐  │                       │
│  │  │ 3. Keyword Extractor          │  │                       │
│  │  │    - Gemini API call          │  │                       │
│  │  │    - Top 8 bigrams            │  │                       │
│  │  │    - CBSE-optimized           │  │                       │
│  │  └───────────────┬───────────────┘  │                       │
│  │                  │                    │                       │
│  │  ┌───────────────▼───────────────┐  │                       │
│  │  │ 4. Vertex AI Embedder         │  │                       │
│  │  │    - Batch embedding (n=5)    │  │                       │
│  │  │    - text-embedding-005       │  │                       │
│  │  │    - 768 dimensions           │  │                       │
│  │  │    - Retry with backoff       │  │                       │
│  │  └───────────────┬───────────────┘  │                       │
│  │                  │                    │                       │
│  │  ┌───────────────▼───────────────┐  │                       │
│  │  │ 5. AlloyDB Writer             │  │                       │
│  │  │    - Transactional insert     │  │                       │
│  │  │    - Parent + children atomic │  │                       │
│  │  │    - DLQ on failure           │  │                       │
│  │  │    - Concurrent inserts       │  │                       │
│  │  └───────────────────────────────┘  │                       │
│  └─────────────────────────────────────┘                       │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Implementation Tasks

### Task 1: Project Structure Setup
**Duration:** 1 day  
**Owner:** Backend Engineer

**Files to Create:**
```bash
cd ingestion-go

# Create directory structure
mkdir -p cmd/ingestion
mkdir -p parser
mkdir -p chunker
mkdir -p keywords
mkdir -p embedder
mkdir -p writer
mkdir -p dlq
mkdir -p tests
mkdir -p config
mkdir -p metrics
```

**go.mod Updates:**
```go
module github.com/visionary/ragpipeline/ingestion-go

go 1.22

require (
    github.com/jackc/pgx/v5 v5.5.0
    github.com/jackc/pgx/v5/pgxpool v5.5.0
    github.com/jackc/pgx/v5/pgtype v5.5.0
    github.com/pgvector/pgvector-go v0.2.0
    github.com/redis/go-redis/v9 v9.5.0
    cloud.google.com/go/vertexai v0.5.0
    cloud.google.com/go/secretmanager v1.11.4
    github.com/rs/zerolog v1.31.0
    go.opentelemetry.io/otel v1.24.0
    go.opentelemetry.io/otel/trace v1.24.0
    go.opentelemetry.io/otel/metric v1.24.0
    github.com/stretchr/testify v1.9.0
    github.com/google/uuid v1.6.0
    github.com/sethvargo/go-retry v0.2.4
)
```

**Environment Template (.env.local.example for ingestion-go):**
```bash
# Go Ingestion Configuration
GOOGLE_CLOUD_PROJECT=your-project-id
GOOGLE_CLOUD_LOCATION=asia-south1
ALLOYDB_DSN="host=127.0.0.1 port=6432 dbname=visionary user=visionary password=changeme"
REDIS_ADDR="localhost:6379"
REDIS_PASSWORD=""
REDIS_DB=0
LOG_LEVEL=info
ENVIRONMENT=production

# Ingestion Tuning
PARENT_CHUNK_SIZE=1500
CHILD_CHUNK_SIZE=512
CHUNK_OVERLAP=77
EMBEDDING_BATCH_SIZE=5
DB_INSERT_CONCURRENCY=10
MAX_RETRIES=5
```

---

### Task 2: JSON Parser (parser/json_parser.go)
**Duration:** 2 days  
**Owner:** Backend Engineer

**Purpose:** Parse JSON output from Python parser

**Implementation:**

```go
// parser/json_parser.go
package parser

import (
    "encoding/json"
    "fmt"
    "io"
    "time"
)

// PageElement represents a parsed element from Python parser.
type PageElement struct {
    PageNumber   int       `json:"page_number"`
    Text         string    `json:"text"`
    Chapter      string    `json:"chapter"`
    Section      string    `json:"section"`
    ContentType  string    `json:"content_type"` // "prose", "table", "figure", "formula"
    IsTable      bool      `json:"is_table"`
    Grade        int       `json:"grade"`
    Subject      string    `json:"subject"`
    TaxonomyID   int       `json:"taxonomy_id"`
    Metadata     Metadata  `json:"metadata,omitempty"`
}

// Metadata contains additional element metadata.
type Metadata struct {
    Fonts       []string  `json:"fonts,omitempty"`
    FontSizes   []float64 `json:"font_sizes,omitempty"`
    BBox        []float64 `json:"bbox,omitempty"`
    TableRows   int       `json:"table_rows,omitempty"`
    TableCols   int       `json:"table_cols,omitempty"`
    FormulaType string    `json:"formula_type,omitempty"`
}

// ParseJSONOutput parses JSON output from Python parser.
func ParseJSONOutput(r io.Reader) ([]PageElement, error) {
    var elements []PageElement
    decoder := json.NewDecoder(r)
    
    if err := decoder.Decode(&elements); err != nil {
        return nil, fmt.Errorf("failed to decode JSON: %w", err)
    }
    
    // Validate elements
    for i, elem := range elements {
        if err := validateElement(elem); err != nil {
            return nil, fmt.Errorf("invalid element at index %d: %w", i, err)
        }
    }
    
    return elements, nil
}

// validateElement validates a parsed element.
func validateElement(elem PageElement) error {
    if elem.PageNumber < 1 {
        return fmt.Errorf("invalid page_number: %d", elem.PageNumber)
    }
    if elem.Text == "" {
        return fmt.Errorf("empty text")
    }
    if elem.Grade < 6 || elem.Grade > 8 {
        return fmt.Errorf("invalid grade: %d", elem.Grade)
    }
    if elem.TaxonomyID == 0 {
        return fmt.Errorf("missing taxonomy_id")
    }
    return nil
}

// ParseFile parses a JSON file from Python parser.
func ParseFile(path string) ([]PageElement, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()
    
    return ParseJSONOutput(file)
}
```

**Tests (parser/json_parser_test.go):**
```go
package parser_test

import (
    "strings"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/visionary/ragpipeline/ingestion-go/parser"
)

func TestParseJSONOutput_Valid(t *testing.T) {
    jsonStr := `[
        {
            "page_number": 1,
            "text": "Cell membrane is the boundary of the cell",
            "chapter": "Cell Structure",
            "section": "The Cell Membrane",
            "content_type": "prose",
            "is_table": false,
            "grade": 7,
            "subject": "Science",
            "taxonomy_id": 42
        }
    ]`
    
    elements, err := parser.ParseJSONOutput(strings.NewReader(jsonStr))
    
    assert.NoError(t, err)
    assert.Len(t, elements, 1)
    assert.Equal(t, 1, elements[0].PageNumber)
    assert.Equal(t, "Cell membrane is the boundary of the cell", elements[0].Text)
}

func TestParseJSONOutput_Invalid(t *testing.T) {
    tests := []struct {
        name    string
        jsonStr string
        wantErr bool
    }{
        {
            name:    "empty text",
            jsonStr: `{"page_number": 1, "text": "", "grade": 7, "taxonomy_id": 1}`,
            wantErr: true,
        },
        {
            name:    "invalid grade",
            jsonStr: `{"page_number": 1, "text": "test", "grade": 10, "taxonomy_id": 1}`,
            wantErr: true,
        },
        {
            name:    "missing taxonomy_id",
            jsonStr: `{"page_number": 1, "text": "test", "grade": 7}`,
            wantErr: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := parser.ParseJSONOutput(strings.NewReader(tt.jsonStr))
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

---

### Task 3: Parent-Child Chunker (chunker/parent_child.go)
**Duration:** 3 days  
**Owner:** Backend Engineer

**Purpose:** Split elements into parent (1500 chars) and child (512 chars) chunks

**Implementation:**

```go
// chunker/parent_child.go
package chunker

import (
    "github.com/visionary/ragpipeline/ingestion-go/parser"
    "github.com/google/uuid"
)

const (
    DefaultParentSize   = 1500
    DefaultChildSize    = 512
    DefaultOverlap      = 77 // 15% of child size
    MinChunkSize        = 50  // Minimum chunk to keep
)

// ParentChunk represents a parent chunk.
type ParentChunk struct {
    ID           uuid.UUID
    TaxonomyID   int
    Grade        int
    Subject      string
    Chapter      string
    Section      string
    Content      string
    Keywords     []string // Populated later
    PageNumbers  []int
    ContentType  string
    Metadata     map[string]interface{}
}

// ChildChunk represents a child chunk.
type ChildChunk struct {
    ID           uuid.UUID
    ParentID     uuid.UUID
    TaxonomyID   int
    Grade        int
    Subject      string
    Chapter      string
    Section      string
    Content      string
    Embedding    []float32 // Populated later
    ChunkIndex   int       // Position within parent
    PageNumbers  []int
    ContentType  string
}

// ChunkElements creates parent-child chunks from parsed elements.
func ChunkElements(elements []parser.PageElement, config *ChunkConfig) ([]ParentChunk, []ChildChunk) {
    if config == nil {
        config = DefaultConfig()
    }
    
    var parents []ParentChunk
    var children []ChildChunk
    
    // Group elements by page and type
    groups := groupElementsByType(elements)
    
    for _, group := range groups {
        // Tables: atomic (parent == child)
        if group[0].IsTable {
            parent, child := createTableChunk(group)
            parents = append(parents, parent)
            children = append(children, child)
            continue
        }
        
        // Prose: recursive splitting
        groupParents, groupChildren := createProseChunks(group, config)
        parents = append(parents, groupParents...)
        children = append(children, groupChildren...)
    }
    
    return parents, children
}

// groupElementsByType groups consecutive elements by type.
func groupElementsByType(elements []parser.PageElement) [][]parser.PageElement {
    if len(elements) == 0 {
        return nil
    }
    
    var groups [][]parser.PageElement
    currentGroup := []parser.PageElement{elements[0]}
    
    for i := 1; i < len(elements); i++ {
        if elements[i].IsTable == elements[i-1].IsTable &&
           elements[i].ContentType == elements[i-1].ContentType {
            currentGroup = append(currentGroup, elements[i])
        } else {
            groups = append(groups, currentGroup)
            currentGroup = []parser.PageElement{elements[i]}
        }
    }
    groups = append(groups, currentGroup)
    
    return groups
}

// createTableChunk creates atomic table chunk.
func createTableChunk(elements []parser.PageElement) (ParentChunk, ChildChunk) {
    parentID := uuid.New()
    
    // Concatenate table text
    var content string
    var pages []int
    for _, elem := range elements {
        content += elem.Text + "\n\n"
        pages = append(pages, elem.PageNumber)
    }
    
    parent := ParentChunk{
        ID:          parentID,
        TaxonomyID:  elements[0].TaxonomyID,
        Grade:       elements[0].Grade,
        Subject:     elements[0].Subject,
        Chapter:     elements[0].Chapter,
        Section:     elements[0].Section,
        Content:     content,
        PageNumbers: pages,
        ContentType: "table",
    }
    
    child := ChildChunk{
        ID:          parentID, // Same as parent for tables
        ParentID:    parentID,
        TaxonomyID:  elements[0].TaxonomyID,
        Grade:       elements[0].Grade,
        Subject:     elements[0].Subject,
        Chapter:     elements[0].Chapter,
        Section:     elements[0].Section,
        Content:     content,
        ChunkIndex:  0,
        PageNumbers: pages,
        ContentType: "table",
    }
    
    return parent, child
}

// createProseChunks creates parent-child chunks for prose.
func createProseChunks(elements []parser.PageElement, config *ChunkConfig) ([]ParentChunk, []ChildChunk) {
    // Concatenate prose
    var fullText string
    var pages []int
    for _, elem := range elements {
        fullText += elem.Text + " "
        pages = append(pages, elem.PageNumber)
    }
    
    // Create parent chunks (1500 chars)
    parentTexts := splitText(fullText, config.ParentSize, config.Overlap)
    
    var parents []ParentChunk
    var children []ChildChunk
    
    for _, parentText := range parentTexts {
        parentID := uuid.New()
        
        parent := ParentChunk{
            ID:          parentID,
            TaxonomyID:  elements[0].TaxonomyID,
            Grade:       elements[0].Grade,
            Subject:     elements[0].Subject,
            Chapter:     elements[0].Chapter,
            Section:     elements[0].Section,
            Content:     parentText,
            PageNumbers: uniqueInts(pages),
            ContentType: "prose",
        }
        parents = append(parents, parent)
        
        // Create child chunks (512 chars)
        childTexts := splitText(parentText, config.ChildSize, config.Overlap)
        for i, childText := range childTexts {
            if len(childText) < MinChunkSize {
                continue
            }
            
            child := ChildChunk{
                ID:          uuid.New(),
                ParentID:    parentID,
                TaxonomyID:  elements[0].TaxonomyID,
                Grade:       elements[0].Grade,
                Subject:     elements[0].Subject,
                Chapter:     elements[0].Chapter,
                Section:     elements[0].Section,
                Content:     childText,
                ChunkIndex:  i,
                PageNumbers: uniqueInts(pages),
                ContentType: "prose",
            }
            children = append(children, child)
        }
    }
    
    return parents, children
}

// splitText splits text into chunks with overlap.
func splitText(text string, chunkSize, overlap int) []string {
    if len(text) <= chunkSize {
        return []string{text}
    }
    
    var chunks []string
    start := 0
    
    for start < len(text) {
        end := start + chunkSize
        if end > len(text) {
            end = len(text)
        }
        
        // Try to break at sentence boundary
        if end < len(text) {
            for i := end; i > start && i > start+chunkSize/2; i-- {
                if text[i] == '.' || text[i] == '!' || text[i] == '?' {
                    end = i + 1
                    break
                }
            }
        }
        
        chunks = append(chunks, text[start:end])
        start = end - overlap
        if start < 0 {
            start = 0
        }
    }
    
    return chunks
}

// ChunkConfig holds chunking configuration.
type ChunkConfig struct {
    ParentSize int
    ChildSize  int
    Overlap    int
}

// DefaultConfig returns default chunking configuration.
func DefaultConfig() *ChunkConfig {
    return &ChunkConfig{
        ParentSize: DefaultParentSize,
        ChildSize:  DefaultChildSize,
        Overlap:    DefaultOverlap,
    }
}

// uniqueInts returns unique integers from a slice.
func uniqueInts(nums []int) []int {
    seen := make(map[int]bool)
    var result []int
    for _, n := range nums {
        if !seen[n] {
            seen[n] = true
            result = append(result, n)
        }
    }
    return result
}
```

**Tests:**
```go
// chunker/parent_child_test.go
package chunker_test

import (
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/visionary/ragpipeline/ingestion-go/chunker"
    "github.com/visionary/ragpipeline/ingestion-go/parser"
)

func TestChunkElements_Prose(t *testing.T) {
    elements := []parser.PageElement{
        {
            PageNumber:  1,
            Text:        "Cell membrane is the boundary of the cell. " + strings.Repeat("It controls what enters and leaves. ", 50),
            Chapter:     "Cell Structure",
            Section:     "The Cell Membrane",
            ContentType: "prose",
            IsTable:     false,
            Grade:       7,
            Subject:     "Science",
            TaxonomyID:  42,
        },
    }
    
    parents, children := chunker.ChunkElements(elements, nil)
    
    assert.NotEmpty(t, parents)
    assert.NotEmpty(t, children)
    
    // Verify parent size
    for _, p := range parents {
        assert.LessOrEqual(t, len(p.Content), chunker.DefaultParentSize+100)
    }
    
    // Verify child size
    for _, c := range children {
        assert.LessOrEqual(t, len(c.Content), chunker.DefaultChildSize+50)
    }
}

func TestChunkElements_Table(t *testing.T) {
    elements := []parser.PageElement{
        {
            PageNumber:  1,
            Text:        "Organelle | Function\nMitochondria | Powerhouse",
            Chapter:     "Cell Structure",
            Section:     "Organelles",
            ContentType: "table",
            IsTable:     true,
            Grade:       7,
            Subject:     "Science",
            TaxonomyID:  42,
        },
    }
    
    parents, children := chunker.ChunkElements(elements, nil)
    
    assert.Len(t, parents, 1)
    assert.Len(t, children, 1)
    assert.Equal(t, parents[0].ID, children[0].ID) // Atomic for tables
}
```

---

### Task 4: Keyword Extractor (keywords/gemini_extractor.go)
**Duration:** 2 days  
**Owner:** Backend Engineer

**Purpose:** Extract keywords using Gemini API

**Implementation:**

```go
// keywords/gemini_extractor.go
package keywords

import (
    "context"
    "fmt"
    "strings"
    
    "cloud.google.com/go/vertexai/genai"
    "github.com/rs/zerolog/log"
)

const (
    DefaultTopK        = 8  // CBSE-optimized
    DefaultMaxBigrams  = 2
    ExtractionPrompt   = `Extract %d key educational terms (1-2 words each) from this text. 
Focus on curriculum-relevant terminology for grade %d %s.
Return ONLY the terms, one per line, no explanations.`
)

// Extractor extracts keywords using Gemini API.
type Extractor struct {
    client *genai.Client
    model  *genai.GenerativeModel
    topK   int
}

// NewExtractor creates a new keyword extractor.
func NewExtractor(ctx context.Context, projectID, location string) (*Extractor, error) {
    client, err := genai.NewClient(ctx, projectID, location)
    if err != nil {
        return nil, fmt.Errorf("failed to create Vertex AI client: %w", err)
    }
    
    model := client.GenerativeModel("gemini-1.5-flash")
    model.SetTemperature(0.1) // Low temperature for consistency
    
    return &Extractor{
        client: client,
        model:  model,
        topK:   DefaultTopK,
    }, nil
}

// ExtractKeywords extracts keywords from text.
func (e *Extractor) ExtractKeywords(ctx context.Context, text string, grade int, subject string) ([]string, error) {
    prompt := fmt.Sprintf(ExtractionPrompt, e.topK, grade, subject)
    fullPrompt := fmt.Sprintf("%s\n\nText:\n%s", prompt, truncate(text, 2000))
    
    resp, err := e.model.GenerateContent(ctx, genai.Text(fullPrompt))
    if err != nil {
        return nil, fmt.Errorf("Gemini API error: %w", err)
    }
    
    if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
        return nil, fmt.Errorf("no keywords extracted")
    }
    
    // Parse response
    keywords := parseKeywords(resp.Candidates[0].Content.Parts[0])
    return keywords[:min(len(keywords), e.topK)], nil
}

// parseKeywords parses Gemini response into keyword list.
func parseKeywords(part genai.Part) []string {
    text := part.(genai.Text)
    lines := strings.Split(string(text), "\n")
    
    var keywords []string
    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line != "" && !strings.HasPrefix(line, "#") {
            keywords = append(keywords, line)
        }
    }
    
    return keywords
}

// ExtractKeywordsBatch extracts keywords for multiple parent chunks.
func (e *Extractor) ExtractKeywordsBatch(
    ctx context.Context,
    parents []chunker.ParentChunk,
    concurrency int,
) error {
    sem := make(chan struct{}, concurrency)
    errChan := make(chan error, len(parents))
    
    for _, parent := range parents {
        sem <- struct{}{}
        go func(p chunker.ParentChunk) {
            defer func() { <-sem }()
            
            keywords, err := e.ExtractKeywords(ctx, p.Content, p.Grade, p.Subject)
            if err != nil {
                log.Error().Err(err).Int("grade", p.Grade).Msg("Keyword extraction failed")
                errChan <- err
                return
            }
            
            p.Keywords = keywords
        }(parent)
    }
    
    // Wait for all to complete
    for i := 0; i < len(parents); i++ {
        select {
        case err := <-errChan:
            if err != nil {
                return err
            }
        case <-ctx.Done():
            return ctx.Err()
        }
    }
    
    return nil
}

// FallbackExtractor provides simple TF-IDF fallback.
type FallbackExtractor struct{}

// ExtractKeywords extracts keywords using simple frequency.
func (f *FallbackExtractor) ExtractKeywords(text string, topK int) []string {
    // Simple word frequency
    words := strings.Fields(strings.ToLower(text))
    freq := make(map[string]int)
    
    stopWords := map[string]bool{
        "the": true, "a": true, "an": true, "and": true, "or": true,
        "is": true, "are": true, "was": true, "were": true,
    }
    
    for _, word := range words {
        if !stopWords[word] && len(word) > 2 {
            freq[word]++
        }
    }
    
    // Sort by frequency
    type kv struct {
        Key   string
        Value int
    }
    
    var sorted []kv
    for k, v := range freq {
        sorted = append(sorted, kv{k, v})
    }
    
    sort.Slice(sorted, func(i, j int) bool {
        return sorted[i].Value > sorted[j].Value
    })
    
    // Return top K
    var keywords []string
    for i := 0; i < min(len(sorted), topK); i++ {
        keywords = append(keywords, sorted[i].Key)
    }
    
    return keywords
}

func truncate(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen] + "..."
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}
```

---

### Task 5: Vertex AI Embedder (embedder/vertex_batch.go)
**Duration:** 2 days  
**Owner:** Backend Engineer

**Implementation:**

```go
// embedder/vertex_batch.go
package embedder

import (
    "context"
    "fmt"
    "time"
    
    "cloud.google.com/go/vertexai/genai"
    "github.com/rs/zerolog/log"
    "github.com/sethvargo/go-retry"
)

const (
    DefaultBatchSize    = 5   // Vertex AI limit
    DefaultMaxRetries   = 5
    DefaultInitialDelay = 2 * time.Second
    DefaultMaxDelay     = 60 * time.Second
    EmbeddingModel      = "text-embedding-005"
    EmbeddingDimensions = 768
)

// Embedder handles Vertex AI embeddings.
type Embedder struct {
    client    *genai.Client
    model     *genai.EmbeddingModel
    batchSize int
}

// NewEmbedder creates a new embedder.
func NewEmbedder(ctx context.Context, projectID, location string) (*Embedder, error) {
    client, err := genai.NewClient(ctx, projectID, location)
    if err != nil {
        return nil, fmt.Errorf("failed to create Vertex AI client: %w", err)
    }
    
    model, err := client.EmbeddingModel(EmbeddingModel)
    if err != nil {
        return nil, fmt.Errorf("failed to load embedding model: %w", err)
    }
    
    return &Embedder{
        client:    client,
        model:     model,
        batchSize: DefaultBatchSize,
    }, nil
}

// EmbedDocuments embeds multiple documents with batching.
func (e *Embedder) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
    var allEmbeddings [][]float32
    
    // Process in batches
    for i := 0; i < len(texts); i += e.batchSize {
        end := i + e.batchSize
        if end > len(texts) {
            end = len(texts)
        }
        
        batch := texts[i:end]
        embeddings, err := e.embedBatch(ctx, batch)
        if err != nil {
            return nil, fmt.Errorf("batch embedding failed at index %d: %w", i, err)
        }
        
        allEmbeddings = append(allEmbeddings, embeddings...)
    }
    
    return allEmbeddings, nil
}

// embedBatch embeds a single batch of texts.
func (e *Embedder) embedBatch(ctx context.Context, texts []string) ([][]float32, error) {
    var result [][]float32
    
    // Retry with exponential backoff
    err := retry.Do(ctx, retry.WithMaxRetries(DefaultMaxRetries, retry.NewExponential(DefaultInitialDelay)), func(ctx context.Context) error {
        req := &genai.BatchEmbedContentsRequest{
            Requests: make([]*genai.EmbedContentRequest, len(texts)),
        }
        
        for i, text := range texts {
            req.Requests[i] = &genai.EmbedContentRequest{
                Content: &genai.Content{Parts: []genai.Part{genai.Text(text)}},
                TaskType: genai.TaskTypeRetrievalDocument,
            }
        }
        
        resp, err := e.model.BatchEmbedContents(ctx, req)
        if err != nil {
            return retry.RetryableError(err)
        }
        
        for _, emb := range resp.Embeddings {
            result = append(result, emb.Values)
        }
        
        return nil
    })
    
    if err != nil {
        return nil, err
    }
    
    return result, nil
}

// EmbedQuery embeds a single query (for retrieval).
func (e *Embedder) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
    req := &genai.EmbedContentRequest{
        Content:  &genai.Content{Parts: []genai.Part{genai.Text(query)}},
        TaskType: genai.TaskTypeRetrievalQuery,
    }
    
    resp, err := e.model.EmbedContent(ctx, req)
    if err != nil {
        return nil, err
    }
    
    return resp.Embedding.Values, nil
}
```

---

### Task 6: AlloyDB Writer (writer/alloydb_writer.go)
**Duration:** 3 days  
**Owner:** Backend Engineer

**Implementation:**

```go
// writer/alloydb_writer.go
package writer

import (
    "context"
    "fmt"
    
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/jackc/pgx/v5/pgtype"
    "github.com/pgvector/pgvector-go"
    "github.com/visionary/ragpipeline/ingestion-go/chunker"
    "github.com/rs/zerolog/log"
)

// Writer writes chunks to AlloyDB.
type Writer struct {
    pool *pgxpool.Pool
}

// NewWriter creates a new database writer.
func NewWriter(ctx context.Context, dsn string) (*Writer, error) {
    config, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to parse DSN: %w", err)
    }
    
    // Configure connection pool
    config.MaxConns = 18
    config.MinConns = 4
    
    pool, err := pgxpool.NewWithConfig(ctx, config)
    if err != nil {
        return nil, fmt.Errorf("failed to create pool: %w", err)
    }
    
    // Test connection
    if err := pool.Ping(ctx); err != nil {
        return nil, fmt.Errorf("failed to ping database: %w", err)
    }
    
    return &Writer{pool: pool}, nil
}

// WriteBatch writes parent and child chunks to database.
func (w *Writer) WriteBatch(ctx context.Context, parents []chunker.ParentChunk, children []chunker.ChildChunk, concurrency int) (success, failure int, err error) {
    sem := make(chan struct{}, concurrency)
    successChan := make(chan int, len(parents))
    errorChan := make(chan error, len(parents))
    
    for _, parent := range parents {
        sem <- struct{}{}
        go func(p chunker.ParentChunk) {
            defer func() { <-sem }()
            
            // Get children for this parent
            var parentChildren []chunker.ChildChunk
            for _, c := range children {
                if c.ParentID == p.ID {
                    parentChildren = append(parentChildren, c)
                }
            }
            
            // Insert with transaction
            err := w.insertParentWithChildren(ctx, p, parentChildren)
            if err != nil {
                log.Error().Err(err).Str("parent_id", p.ID.String()).Msg("Insert failed")
                errorChan <- err
                return
            }
            
            successChan <- 1
        }(parent)
    }
    
    // Collect results
    for i := 0; i < len(parents); i++ {
        select {
        case <-successChan:
            success++
        case err := <-errorChan:
            failure++
            log.Error().Err(err).Msg("Batch write error")
        case <-ctx.Done():
            return success, failure, ctx.Err()
        }
    }
    
    return success, failure, nil
}

// insertParentWithChildren inserts parent and children atomically.
func (w *Writer) insertParentWithChildren(ctx context.Context, parent chunker.ParentChunk, children []chunker.ChildChunk) error {
    return w.pool.BeginFunc(ctx, func(tx pgx.Tx) error {
        // Insert parent
        err := w.insertParentTx(ctx, tx, parent)
        if err != nil {
            return fmt.Errorf("parent insert failed: %w", err)
        }
        
        // Insert children
        for _, child := range children {
            err := w.insertChildTx(ctx, tx, child)
            if err != nil {
                return fmt.Errorf("child insert failed: %w", err)
            }
        }
        
        return nil
    })
}

// insertParentTx inserts a parent chunk within a transaction.
func (w *Writer) insertParentTx(ctx context.Context, tx pgx.Tx, parent chunker.ParentChunk) error {
    query := `
        INSERT INTO parent_chunks (
            id, taxonomy_id, grade, subject, chapter, section,
            content, keywords, page_numbers, content_type, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
    `
    
    _, err := tx.Exec(ctx, query,
        parent.ID,
        parent.TaxonomyID,
        parent.Grade,
        parent.Subject,
        parent.Chapter,
        parent.Section,
        parent.Content,
        parent.Keywords,
        parent.PageNumbers,
        parent.ContentType,
    )
    
    return err
}

// insertChildTx inserts a child chunk within a transaction.
func (w *Writer) insertChildTx(ctx context.Context, tx pgx.Tx, child chunker.ChildChunk) error {
    query := `
        INSERT INTO child_chunks (
            id, parent_id, taxonomy_id, grade, subject, chapter, section,
            content, embedding, chunk_index, page_numbers, content_type, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, NOW())
    `
    
    var embedding pgvector.Vector
    if len(child.Embedding) > 0 {
        embedding = pgvector.NewVector(child.Embedding)
    }
    
    _, err := tx.Exec(ctx, query,
        child.ID,
        child.ParentID,
        child.TaxonomyID,
        child.Grade,
        child.Subject,
        child.Chapter,
        child.Section,
        child.Content,
        embedding,
        child.ChunkIndex,
        child.PageNumbers,
        child.ContentType,
    )
    
    return err
}

// Close closes the database pool.
func (w *Writer) Close() {
    w.pool.Close()
}
```

---

### Task 7: CLI Entry Point (cmd/ingestion/main.go)
**Duration:** 2 days  
**Owner:** Backend Engineer

**Implementation:**

```go
// cmd/ingestion/main.go
package main

import (
    "context"
    "flag"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
    
    "github.com/visionary/ragpipeline/ingestion-go/parser"
    "github.com/visionary/ragpipeline/ingestion-go/chunker"
    "github.com/visionary/ragpipeline/ingestion-go/keywords"
    "github.com/visionary/ragpipeline/ingestion-go/embedder"
    "github.com/visionary/ragpipeline/ingestion-go/writer"
)

func main() {
    // Parse flags
    pdfPath := flag.String("pdf", "", "Path to PDF file (required)")
    grade := flag.Int("grade", 0, "Grade level 6-8 (required)")
    subject := flag.String("subject", "", "Subject name (required)")
    taxonomyID := flag.Int("taxonomy-id", 0, "Reference to cbse_taxonomy (required)")
    projectID := flag.String("project-id", os.Getenv("GOOGLE_CLOUD_PROJECT"), "GCP project ID")
    location := flag.String("location", "asia-south1", "GCP region")
    dsn := flag.String("alloydb-dsn", os.Getenv("ALLOYDB_DSN"), "Database connection string")
    batchSize := flag.Int("batch-size", 5, "Embedding batch size")
    concurrency := flag.Int("concurrency", 10, "Concurrent DB inserts")
    pythonParser := flag.String("python-parser", "", "Path to Python parser script")
    flag.Parse()
    
    // Validate required flags
    if *pdfPath == "" || *grade == 0 || *subject == "" || *taxonomyID == 0 {
        flag.Usage()
        os.Exit(1)
    }
    
    // Initialize logger
    zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
    if os.Getenv("ENVIRONMENT") != "production" {
        log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
    }
    
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    
    // Handle shutdown signals
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigChan
        log.Info().Msg("Shutting down...")
        cancel()
    }()
    
    // Run ingestion
    if err := runIngestion(ctx, *pdfPath, *grade, *subject, *taxonomyID, *projectID, *location, *dsn, *batchSize, *concurrency, *pythonParser); err != nil {
        log.Fatal().Err(err).Msg("Ingestion failed")
    }
    
    log.Info().Msg("Ingestion completed successfully")
}

func runIngestion(ctx context.Context, pdfPath string, grade int, subject string, taxonomyID int, projectID, location, dsn string, batchSize, concurrency int, pythonParser string) error {
    startTime := time.Now()
    
    // Step 1: Parse PDF (Python)
    log.Info().Str("pdf", pdfPath).Msg("Step 1: Parsing PDF with Python")
    elements, err := parseWithPython(ctx, pdfPath, grade, subject, taxonomyID, pythonParser)
    if err != nil {
        return fmt.Errorf("parsing failed: %w", err)
    }
    log.Info().Int("elements", len(elements)).Msg("Parsed elements")
    
    // Step 2: Create parent-child chunks
    log.Info().Msg("Step 2: Creating parent-child chunks")
    parents, children := chunker.ChunkElements(elements, nil)
    log.Info().Int("parents", len(parents)).Int("children", len(children)).Msg("Chunks created")
    
    // Step 3: Extract keywords
    log.Info().Msg("Step 3: Extracting keywords with Gemini")
    keywordExtractor, err := keywords.NewExtractor(ctx, projectID, location)
    if err != nil {
        return fmt.Errorf("keyword extractor init failed: %w", err)
    }
    if err := keywordExtractor.ExtractKeywordsBatch(ctx, parents, concurrency); err != nil {
        return fmt.Errorf("keyword extraction failed: %w", err)
    }
    log.Info().Msg("Keywords extracted")
    
    // Step 4: Generate embeddings
    log.Info().Msg("Step 4: Generating embeddings with Vertex AI")
    embedder, err := embedder.NewEmbedder(ctx, projectID, location)
    if err != nil {
        return fmt.Errorf("embedder init failed: %w", err)
    }
    
    childTexts := make([]string, len(children))
    for i, child := range children {
        childTexts[i] = child.Content
    }
    
    embeddings, err := embedder.EmbedDocuments(ctx, childTexts)
    if err != nil {
        return fmt.Errorf("embedding failed: %w", err)
    }
    
    // Attach embeddings to children
    for i := range children {
        children[i].Embedding = embeddings[i]
    }
    log.Info().Msg("Embeddings generated")
    
    // Step 5: Write to database
    log.Info().Msg("Step 5: Writing to AlloyDB")
    dbWriter, err := writer.NewWriter(ctx, dsn)
    if err != nil {
        return fmt.Errorf("writer init failed: %w", err)
    }
    defer dbWriter.Close()
    
    success, failure, err := dbWriter.WriteBatch(ctx, parents, children, concurrency)
    if err != nil {
        return fmt.Errorf("database write failed: %w", err)
    }
    
    log.Info().
        Int("success", success).
        Int("failure", failure).
        Dur("duration", time.Since(startTime)).
        Msg("Ingestion complete")
    
    return nil
}

// parseWithPython calls Python parser and returns elements.
func parseWithPython(ctx context.Context, pdfPath string, grade int, subject string, taxonomyID int, pythonParser string) ([]parser.PageElement, error) {
    // Implementation: exec.Command to run Python parser
    // For now, assume JSON output file exists
    jsonPath := pdfPath + ".json"
    return parser.ParseFile(jsonPath)
}
```

---

### Task 8: Testing
**Duration:** 3 days  
**Owner:** QA Engineer

**Test Coverage Requirements:**
- Unit tests: 90%+ coverage for all packages
- Integration tests: End-to-end ingestion flow
- Performance tests: ≥50 pages/min throughput

**Test Files:**
```
ingestion-go/tests/
├── parser_test.go
├── chunker_test.go
├── keywords_test.go
├── embedder_test.go
├── writer_test.go
└── integration_test.go
```

---

## Acceptance Criteria

- [ ] All 8 tasks completed
- [ ] 90%+ test coverage
- [ ] Performance: ≥50 pages/min
- [ ] Error handling with DLQ routing
- [ ] Graceful degradation to Python parser
- [ ] Documentation complete
- [ ] Code review approved
- [ ] Integration tests passing

---

## Dependencies

- Python parser must remain compatible
- Database schema must be applied
- GCP credentials configured
- Vertex AI API enabled

---

## Risks & Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| Python parser output changes | High | Schema validation, versioning |
| Vertex AI API rate limits | Medium | Batching, retry logic |
| Database connection exhaustion | Medium | PgBouncer, connection pooling |
| Memory issues with large PDFs | Medium | Streaming, chunking |

---

**Next Steps:**
1. Review and approve this plan
2. Assign engineers to tasks
3. Set up development environment
4. Begin Task 1 (Project Structure)
