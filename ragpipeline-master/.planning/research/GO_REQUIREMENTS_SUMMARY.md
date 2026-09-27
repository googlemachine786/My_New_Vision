# Go Implementation Requirements Summary

**Extracted from:** Visionary_Production_RAG_GCP_v2.docx (via related markdown documents)  
**Date:** March 27, 2026  
**Purpose:** Define ALL Go implementation requirements for the RAG pipeline

---

## Critical Finding: Architecture Decision

**The current architecture specifies Python for ingestion, NOT Go.**

However, based on the user's requirement to use Go "wherever possible," this document extracts:
1. What is **currently planned in Go** (Phase 3+ orchestration layer)
2. What **could be migrated to Go** (currently Python ingestion)
3. **Go libraries and patterns** mentioned or recommended

---

## Part 1: What IS Currently Planned in Go (Per Documentation)

### Orchestrator Layer (Phase 3)

| Component | File | Purpose | Go Libraries |
|-----------|------|---------|--------------|
| **HTTP Server** | `cmd/server/main.go` | SSE streaming, graceful shutdown | `gorilla/mux`, `net/http` |
| **RAG Handler** | `handler/rag_handler.go` | Query processing, JWT auth, SSE | `google.golang.org/genai` |
| **Database Pool** | `db/alloydb.go` | pgx connection pool | `pgx/v5`, `pgvector-go` |
| **Redis Store** | `session/redis_store.go` | Session management | `redis/go-redis/v9` |
| **Vertex Client** | `embed/vertex_client.go` | Embedding API client | `google.golang.org/genai` |
| **Hybrid Search** | `retrieval/hybrid_search.go` | ScaNN + GIN queries | `pgx/v5` |
| **RRF Fusion** | `retrieval/rrf.go` | Reciprocal Rank Fusion | Standard library (`sort`) |
| **Config Loader** | `config/config.go` | Secret Manager integration | `cloud.google.com/go/secretmanager` |
| **Observability** | `observability/otel.go` | OpenTelemetry tracing | `go.opentelemetry.io/otel` |

### Go Dependencies (from go.mod)

```go
require (
    github.com/gorilla/mux                      v1.8.x
    github.com/pgvector/pgvector-go             v0.2.0
    github.com/redis/go-redis/v9                v9.5.0
    google.golang.org/genai                     v0.5.0  // NOT deprecated vertexai package
    go.opentelemetry.io/otel                    v1.24.0
    go.opentelemetry.io/otel/trace              v1.24.0
    go.opentelemetry.io/otel/sdk                v1.24.0
    go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.24.0
    github.com/vmihailenco/msgpack/v5            // For Redis serialization
    gopkg.in/yaml.v3                             // Config parsing
)
```

**Critical Note:** Use `google.golang.org/genai` — the `cloud.google.com/go/vertexai` package is deprecated (June 2026).

---

## Part 2: What COULD Be Migrated to Go (Currently Python)

### Ingestion Pipeline (Phase 2 - Currently Python)

The documentation specifies **Python 3.11** for ingestion with these libraries:

| Component | Current Python Library | Go Alternative | Feasibility |
|-----------|----------------------|----------------|-------------|
| **PDF Parsing** | PyMuPDF (`fitz`), pdfplumber | `pdfcpu`, `unipdf` | ⚠️ MEDIUM - Go PDF libraries less mature |
| **Font Calibration** | Custom (fitz.Document) | `pdfcpu` extract fonts | ⚠️ MEDIUM |
| **Table Extraction** | pdfplumber.tables() | `unipdf` table detection | ⚠️ MEDIUM-HARD |
| **Heading Detection** | Custom font size logic | `pdfcpu` + custom logic | ✅ FEASIBLE |
| **Formula Detection** | Regex + span flags | Regex + PDF metadata | ✅ FEASIBLE |
| **Text Chunking** | RecursiveCharacterTextSplitter | Custom Go implementation | ✅ EASY |
| **Keyword Extraction** | YAKE (`pip install yake`) | ⚠️ NO DIRECT EQUIVALENT | ❌ HARD - See alternatives below |
| **Embedding** | Vertex AI Python SDK | `google.golang.org/genai` | ✅ EASY (same SDK) |
| **Database Write** | psycopg3 async | `pgx/v5` async | ✅ EASY (better in Go) |

### Recommendation for Go Migration

If migrating ingestion to Go, here's the recommended approach:

#### Option A: Full Go Implementation

```go
// PDF Parsing
import (
    "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"      // PDF manipulation
    "github.com/unidoc/unipdf/v3/extractor"    // Text extraction
    "github.com/unidoc/unipdf/v3/model"        // Low-level PDF access
)

// Text Chunking
// Implement RecursiveCharacterTextSplitter in Go:
type RecursiveCharacterTextSplitter struct {
    Separators []string
    ChunkSize  int
    Overlap    int
}

func (r *RecursiveCharacterTextSplitter) SplitText(text string) []string {
    // Implementation similar to Python langchain
}

// Keyword Extraction (CRITICAL - YAKE has no Go equivalent)
// Option 1: Use simple TF-IDF
import "github.com/kljensen/snowball"  // Stemming

// Option 2: Call YAKE as external service (HTTP API)
// Option 3: Implement simplified YAKE algorithm in Go
// Option 4: Use Vertex AI for keyword extraction (API call)

// Embedding
import "google.golang.org/genai"  // Same as orchestrator

// Database
import "github.com/jackc/pgx/v5"  // pgx v5
```

#### Option B: Hybrid Approach (Recommended)

Keep Python for PDF parsing (complex, library-dependent), use Go for:
- ✅ Text chunking
- ✅ Keyword extraction (if Go library found)
- ✅ Embedding (same genai SDK)
- ✅ Database writes (pgx is excellent)

**Architecture:**
```
PDF → Python (parse only) → JSON → Go (chunk, embed, write)
```

This gives you:
- Python's superior PDF libraries (PyMuPDF, pdfplumber)
- Go's superior concurrency and database handling
- Single language for 80% of the pipeline

---

## Part 3: Specific Go Libraries & Patterns

### PDF Parsing (If Implementing in Go)

```go
// Recommended: pdfcpu (active maintenance, good API)
go get github.com/pdfcpu/pdfcpu

// Usage example:
import (
    "github.com/pdfcpu/pdfcpu/pkg/pdfcpu"
    "github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// Extract text with font info
ctx, err := pdfcpu.ReadContextFile("input.pdf", "")
if err != nil {
    return err
}

// Access font information
for _, page := range ctx.PageMap {
    for _, font := range page.Fonts {
        // Font name, size, encoding
    }
}

// Alternative: UniDoc (commercial, more features)
go get github.com/unidoc/unipdf/v3
```

**Caveat:** Go PDF libraries are less mature than Python's PyMuPDF/pdfplumber for:
- Table detection
- Font size extraction
- Bounding box calculations

### Text Chunking (Go Implementation)

```go
// Implement RecursiveCharacterTextSplitter
package chunker

type RecursiveCharacterTextSplitter struct {
    Separators []string
    ChunkSize  int
    Overlap    int
    KeepSeparator bool
}

func NewParentChildSplitter() *RecursiveCharacterTextSplitter {
    return &RecursiveCharacterTextSplitter{
        Separators: []string{"\n\n", "\n", ". ", " ", ""},
        ChunkSize:  512,  // child size
        Overlap:    77,   // 15% of 512
    }
}

func (r *RecursiveCharacterTextSplitter) SplitText(text string) []string {
    // Implementation from langchain-go or custom
    // See: github.com/tmc/langchaingo/textsplitter
}
```

**Use langchain-go:**
```go
import "github.com/tmc/langchaingo/textsplitter"

splitter := textsplitter.NewRecursiveCharacterTextSplitter(
    textsplitter.WithChunkSize(512),
    textsplitter.WithChunkOverlap(77),
)
docs := splitter.SplitText(longText)
```

### Keyword Extraction (CRITICAL GAP)

**Problem:** YAKE has **no Go equivalent**. Options:

#### Option 1: TF-IDF (Simple, Fast)
```go
import "github.com/kljensen/snowball"  // For stemming

// Implement simple TF-IDF keyword extraction
func ExtractKeywords(text string, topN int) []string {
    // Tokenize
    // Stem words
    // Calculate TF-IDF
    // Return top N
}
```

#### Option 2: Vertex AI for Keywords (Recommended)
```go
import "google.golang.org/genai"

// Use Gemini to extract keywords
func ExtractKeywordsWithGemini(text string) ([]string, error) {
    prompt := fmt.Sprintf(`Extract 8 key concepts from this text as comma-separated bigrams:
%s`, text)
    
    resp, err := client.GenerateContent(ctx, &genai.Content{
        Parts: []genai.Part{genai.Text(prompt)},
    })
    
    // Parse response into []string
}
```

**Pros:** No library needed, uses existing Vertex AI integration  
**Cons:** API cost, latency (~200ms per parent chunk)

#### Option 3: Run YAKE as HTTP Service
```go
// Deploy YAKE as Cloud Run service
// Call from Go ingestion:
func ExtractKeywords(text string) ([]string, error) {
    resp, err := http.Post("http://yake-service/extract", 
        "application/json", 
        strings.NewReader(`{"text": "`+text+`"}`))
    
    // Parse JSON response
}
```

### Embedding (Go Client for Vertex AI)

**Already documented in orchestrator:**

```go
// embed/vertex_client.go
package embed

import (
    "context"
    "google.golang.org/genai"
)

type VertexEmbedClient struct {
    client *genai.Client
    model  string
}

func NewVertexEmbedClient(ctx context.Context, projectID, location string) (*VertexEmbedClient, error) {
    client, err := genai.NewClient(ctx, &genai.ClientConfig{
        Project:  projectID,
        Location: location,
    })
    
    return &VertexEmbedClient{
        client: client,
        model:  "text-embedding-005",
    }, nil
}

func (c *VertexEmbedClient) EmbedDocuments(ctx context.Context, texts []string) ([][]float32, error) {
    // Batch size ≤5 (Vertex AI limit)
    // Task type: RETRIEVAL_DOCUMENT
    // Output dimensionality: 768
    
    embeddings := make([][]float32, len(texts))
    
    for i := 0; i < len(texts); i += 5 {
        end := i + 5
        if end > len(texts) {
            end = len(texts)
        }
        
        batch := texts[i:end]
        resp, err := c.client.Models.EmbedContent(ctx, 
            c.model, 
            &genai.EmbedContentRequest{
                Contents: batch,
                TaskType: "RETRIEVAL_DOCUMENT",
                OutputDimensionality: 768,
            })
        
        if err != nil {
            return nil, err
        }
        
        for j, emb := range resp.Embeddings {
            embeddings[i+j] = emb.Values
        }
    }
    
    return embeddings, nil
}

func (c *VertexEmbedClient) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
    resp, err := c.client.Models.EmbedContent(ctx,
        c.model,
        &genai.EmbedContentRequest{
            Contents: []string{text},
            TaskType: "RETRIEVAL_QUERY",
            OutputDimensionality: 768,
        })
    
    if err != nil {
        return nil, err
    }
    
    return resp.Embeddings[0].Values, nil
}
```

### Database Writes (pgx v5)

```go
// db/alloydb.go
package db

import (
    "context"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/pgvector/pgvector-go"
)

type AlloyDBWriter struct {
    pool *pgxpool.Pool
}

func NewAlloyDBWriter(ctx context.Context, dsn string) (*AlloyDBWriter, error) {
    config, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, err
    }
    
    config.MaxConns = 18  // Match PgBouncer default_pool_size
    config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
        // Register pgvector type
        pgvector.RegisterTypes(ctx, conn)
        return nil
    }
    
    pool, err := pgxpool.NewWithConfig(ctx, config)
    if err != nil {
        return nil, err
    }
    
    return &AlloyDBWriter{pool: pool}, nil
}

// Transaction pattern (atomic parent+child insert)
func (w *AlloyDBWriter) InsertParentChild(ctx context.Context, 
    parent ParentChunk, children []ChildChunk) error {
    
    return w.pool.BeginFunc(ctx, func(tx pgx.Tx) error {
        // Insert parent
        var parentID uuid.UUID
        err := tx.QueryRow(ctx, `
            INSERT INTO parent_chunks (taxonomy_id, content, extracted_keywords, page_number, chapter, section, content_type)
            VALUES ($1, $2, $3, $4, $5, $6, $7)
            RETURNING parent_id`,
            parent.TaxonomyID, parent.Content, parent.Keywords, 
            parent.PageNumber, parent.Chapter, parent.Section, parent.ContentType,
        ).Scan(&parentID)
        
        if err != nil {
            return err
        }
        
        // Insert children
        for _, child := range children {
            _, err := tx.Exec(ctx, `
                INSERT INTO child_chunks (parent_id, taxonomy_id, content, embedding, page_number, content_type)
                VALUES ($1, $2, $3, $4, $5, $6)`,
                parentID, child.TaxonomyID, child.Content,
                pgvector.NewVector(child.Embedding), // pgvector-go conversion
                child.PageNumber, child.ContentType,
            )
            
            if err != nil {
                return err  // Triggers rollback
            }
        }
        
        return nil
    })
}

// DLQ routing on failure
func (w *AlloyDBWriter) InsertDLQ(ctx context.Context, payload string, err string) error {
    _, err := w.pool.Exec(ctx, `
        INSERT INTO ingestion_dlq (payload, error_message, created_at)
        VALUES ($1, $2, NOW())`,
        payload, err,
    )
    return err
}
```

### RRF Fusion (Already in Go)

```go
// retrieval/rrf.go
package retrieval

type ChunkResult struct {
    ChunkID   uuid.UUID
    ParentID  uuid.UUID
    Content   string
    Score     float64
    RankDense int
    RankSparse int
}

// rrfFuse performs Reciprocal Rank Fusion
// k=60 is the standard parameter
func rrfFuse(denseResults, sparseResults []ChunkResult, k float64) []ChunkResult {
    // Map parent_id -> ChunkResult (deduplicate)
    parentMap := make(map[uuid.UUID]*ChunkResult)
    
    // Rank dense results
    for i, result := range denseResults {
        result.RankDense = i + 1
        if existing, ok := parentMap[result.ParentID]; ok {
            existing.RankDense = min(existing.RankDense, result.RankDense)
        } else {
            parentMap[result.ParentID] = &result
        }
    }
    
    // Rank sparse results
    for i, result := range sparseResults {
        result.RankSparse = i + 1
        if existing, ok := parentMap[result.ParentID]; ok {
            existing.RankSparse = min(existing.RankSparse, result.RankSparse)
        } else {
            parentMap[result.ParentID] = &result
        }
    }
    
    // Calculate RRF score
    results := make([]ChunkResult, 0, len(parentMap))
    for _, result := range parentMap {
        result.Score = 0.0
        if result.RankDense > 0 {
            result.Score += 1.0 / (k + float64(result.RankDense))
        }
        if result.RankSparse > 0 {
            result.Score += 1.0 / (k + float64(result.RankSparse))
        }
        results = append(results, *result)
    }
    
    // Sort by RRF score (descending)
    sort.Slice(results, func(i, j int) bool {
        return results[i].Score > results[j].Score
    })
    
    return results
}
```

---

## Part 4: Architecture Decision Summary

### Current Architecture (Per Documentation)

```
┌─────────────────────────────────────────────────────────┐
│  Ingestion (Python 3.11)                                │
│  - PyMuPDF + pdfplumber for PDF parsing                 │
│  - YAKE for keyword extraction                          │
│  - Vertex AI Python SDK for embeddings                  │
│  - psycopg3 for async DB writes                         │
└─────────────────────────────────────────────────────────┘
                          ↓
                  AlloyDB (ScaNN + GIN)
                          ↓
┌─────────────────────────────────────────────────────────┐
│  Orchestrator (Go 1.22)                                 │
│  - gorilla/mux for HTTP                                 │
│  - pgx/v5 for database                                  │
│  - go-redis/v9 for session                              │
│  - google.golang.org/genai for Vertex AI                │
│  - RRF fusion in Go                                     │
└─────────────────────────────────────────────────────────┘
```

### Recommended for "Go Wherever Possible"

#### Option 1: Full Go Ingestion (Hard Mode)
```
PDF → Go (pdfcpu/unipdf) → Go (chunk) → Go (keywords via Gemini) → Go (embed) → Go (write)
```

**Pros:** Single language, better concurrency  
**Cons:** PDF parsing less mature, keyword extraction requires API calls

**Libraries:**
- PDF: `github.com/pdfcpu/pdfcpu`, `github.com/unidoc/unipdf/v3`
- Chunking: `github.com/tmc/langchaingo/textsplitter`
- Keywords: `google.golang.org/genai` (Gemini API) or custom TF-IDF
- Embedding: `google.golang.org/genai`
- Database: `github.com/jackc/pgx/v5`, `github.com/pgvector/pgvector-go`

#### Option 2: Hybrid (Recommended)
```
PDF → Python (parse only) → JSON → Go (chunk, embed, write)
```

**Pros:** Best of both worlds (Python PDF libs + Go performance)  
**Cons:** Two languages, inter-process communication

**Implementation:**
```python
# Python: parser.py
import fitz, json

def parse_pdf(pdf_path):
    doc = fitz.open(pdf_path)
    pages = []
    for page in doc:
        text = page.get_text()
        # Extract tables, fonts, etc.
        pages.append({...})
    return json.dumps(pages)  # Output JSON
```

```go
// Go: main.go
func main() {
    // Call Python parser
    cmd := exec.Command("python", "parser.py", "--pdf", pdfPath)
    output, _ := cmd.Output()
    
    var pages []Page
    json.Unmarshal(output, &pages)
    
    // Go: chunk, embed, write
    for _, page := range pages {
        chunks := chunker.Split(page.Text)
        embeddings := vertex.Embed(chunks)
        db.Write(chunks, embeddings)
    }
}
```

#### Option 3: Keep Current (Python Ingestion + Go Orchestrator)
```
Python Ingestion (batch) → AlloyDB → Go Orchestrator (real-time queries)
```

**Pros:** Each language does what it's best at  
**Cons:** Two codebases to maintain

**This is the current architecture and is production-ready.**

---

## Part 5: Specific Implementation Patterns

### Go Context Patterns

```go
// 450ms context deadline for requests
func (h *RAGHandler) Query(w http.ResponseWriter, r *http.Request) {
    ctx, cancel := context.WithTimeout(r.Context(), 450*time.Millisecond)
    defer cancel()
    
    // All operations use ctx
    session, err := h.redis.GetHistory(ctx, sessionID)
    embedding, err := h.vertex.EmbedQuery(ctx, query)
    results, err := h.db.HybridSearch(ctx, embedding, taxonomyID)
    
    // Goroutines use context.Background() to outlive request
    go h.session.AppendTurn(context.Background(), ...)
    go h.db.LogFeedback(context.Background(), ...)
}
```

### Go Retry Pattern (for Vertex AI)

```go
import "github.com/avast/retry-go/v4"

func (c *VertexClient) EmbedWithRetry(ctx context.Context, texts []string) ([][]float32, error) {
    var embeddings [][]float32
    
    err := retry.Do(
        func() error {
            var err error
            embeddings, err = c.EmbedDocuments(ctx, texts)
            return err
        },
        retry.Attempts(5),
        retry.Delay(2*time.Second),
        retry.MaxDelay(60*time.Second),
        retry.DelayType(retry.ExponentialDelay),
        retry.RetryIf(func(err error) bool {
            // Only retry on ResourceExhausted, not on invalid input
            return strings.Contains(err.Error(), "ResourceExhausted")
        }),
    )
    
    return embeddings, err
}
```

### Go Concurrency Pattern (Batch Processing)

```go
// Process chunks in parallel (with rate limiting)
func (p *IngestionPipeline) ProcessChunks(ctx context.Context, pages []Page) error {
    sem := make(chan struct{}, 10)  // Limit to 10 concurrent goroutines
    errChan := make(chan error, len(pages))
    wg := sync.WaitGroup{}
    
    for _, page := range pages {
        wg.Add(1)
        sem <- struct{}{}  // Acquire semaphore
        
        go func(page Page) {
            defer wg.Done()
            defer func() { <-sem }()  // Release semaphore
            
            chunks := p.chunker.Split(page.Text)
            embeddings, err := p.embedder.EmbedDocuments(ctx, chunks)
            if err != nil {
                errChan <- err
                return
            }
            
            err = p.db.InsertParentChild(ctx, chunks, embeddings)
            if err != nil {
                errChan <- err
            }
        }(page)
    }
    
    wg.Wait()
    close(errChan)
    
    // Collect errors
    for err := range errChan {
        if err != nil {
            return err
        }
    }
    
    return nil
}
```

---

## Part 6: Decision Matrix

| Component | Recommended Language | Rationale |
|-----------|---------------------|-----------|
| **PDF Parsing** | Python | PyMuPDF/pdfplumber are industry standard, Go alternatives less mature |
| **Text Chunking** | Go | Simple string manipulation, langchain-go available |
| **Keyword Extraction** | Go + Gemini API | YAKE has no Go equivalent; use Gemini or TF-IDF |
| **Embedding** | Go | Same genai SDK, better concurrency |
| **Database Writes** | Go | pgx/v5 is excellent, better than psycopg3 |
| **Orchestration** | Go | Already in Go, perfect fit |
| **Query Serving** | Go | Already in Go, SSE streaming works well |

---

## Part 7: Migration Path (If Choosing Full Go)

### Phase 1: Keep Python Parser, Migrate Rest to Go
```
Week 1: Python parser outputs JSON
Week 2: Go chunker implementation
Week 3: Go embedding + database writer
Week 4: Integration testing
```

### Phase 2: Evaluate Go PDF Libraries
```
Week 1: Prototype pdfcpu for text extraction
Week 2: Prototype unipdf for table detection
Week 3: Compare output quality vs PyMuPDF
Week 4: Decision: migrate or keep Python parser
```

### Phase 3: Keyword Extraction Solution
```
Option A: Implement TF-IDF in Go (1 week)
Option B: Use Gemini API for keywords (2 days)
Option C: Run YAKE as HTTP service (3 days)
```

---

## Conclusion

**Current Architecture Decision:** Python for ingestion, Go for orchestration is **sound** and production-ready.

**If migrating to Go:**
1. ✅ Easy: Chunking, embedding, database writes
2. ⚠️ Medium: PDF parsing (library maturity gap)
3. ❌ Hard: Keyword extraction (no YAKE equivalent)

**Recommended Hybrid Approach:**
```
Python (PDF parse → JSON) → Go (chunk, embed, write)
```

This gives you:
- Python's superior PDF libraries
- Go's superior concurrency and database handling
- Single language for 80% of the pipeline
- Easier maintenance than full Python or full Go

---

**Files to Create for Go Implementation:**
- `ingestion-go/` (new directory for Go ingestion)
  - `cmd/ingestion/main.go`
  - `parser/pdf_parser.go` (or call Python parser)
  - `chunker/recursive_splitter.go`
  - `keywords/extractor.go` (TF-IDF or Gemini)
  - `embedder/vertex_client.go` (reuse from orchestrator)
  - `writer/alloydb_writer.go` (reuse from orchestrator)
  - `go.mod`

**Dependencies to Add:**
```bash
go get github.com/jackc/pgx/v5
go get github.com/pgvector/pgvector-go
go get github.com/redis/go-redis/v9
go get google.golang.org/genai
go get github.com/tmc/langchaingo/textsplitter
go get github.com/pdfcpu/pdfcpu  # If implementing PDF parsing in Go
go get github.com/avast/retry-go/v4  # For retry logic
```
