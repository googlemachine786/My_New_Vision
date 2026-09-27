# Vector Search Service

A production-ready Go microservice for hybrid vector search in RAG pipelines. Implements dense vector similarity (ScaNN), sparse keyword search (GIN), and Reciprocal Rank Fusion (RRF) for high-quality retrieval.

## Features

- **Hybrid Search**: Combines dense vector similarity + sparse keyword matching with RRF fusion
- **Dense Search**: ScaNN cosine similarity on `child_chunks` table via pgvector
- **Sparse Search**: GIN index intersection on `parent_chunks.extracted_keywords` (TEXT[] array)
- **RRF Fusion**: Reciprocal Rank Fusion with configurable k (default 60) for result merging
- **Metadata Filtering**: Pre-filtering support on grade, subject, content_type, chapter
- **Connection Pooling**: Configurable pgx/v5 pool with health checks
- **Query Timeouts**: Configurable timeout to prevent long-running queries
- **Search Metrics**: Built-in statistics for search timing and result counts
- **Health Checks**: Database connectivity monitoring via `/health` endpoint

## Architecture

```
┌─────────────────────────────────────────────────┐
│                 HTTP Server (8082)               │
├─────────────────────────────────────────────────┤
│  POST /search        │ Hybrid dense + sparse    │
│  POST /search/dense  │ Pure vector similarity   │
│  POST /search/sparse │ Pure keyword search      │
│  GET  /health        │ Health check             │
│  GET  /stats         │ Service statistics       │
├─────────────────────────────────────────────────┤
│              Hybrid Search Orchestrator          │
│  ┌──────────────┐  ┌──────────────┐             │
│  │ DenseSearcher│  │SparseSearcher│             │
│  │  (ScaNN)     │  │   (GIN)      │             │
│  └──────┬───────┘  └──────┬───────┘             │
│         └────────┬─────────┘                    │
│                  ▼                              │
│         ┌──────────────┐                        │
│         │  RRF Fusion  │ (k=60)                │
│         └──────┬───────┘                        │
│                ▼                                │
│         ┌──────────────┐                        │
│         │ Deduplicate  │                        │
│         │  & Limit     │                        │
│         └──────────────┘                        │
├─────────────────────────────────────────────────┤
│              PostgreSQL / AlloyDB                │
│  ┌──────────────────┐  ┌──────────────────┐     │
│  │  child_chunks    │  │  parent_chunks   │     │
│  │  (embedding)     │  │  (keywords)      │     │
│  └──────────────────┘  └──────────────────┘     │
└─────────────────────────────────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.23+
- PostgreSQL 14+ with pgvector extension (or AlloyDB with ScaNN)
- GIN index on `parent_chunks.extracted_keywords`

### Installation

```bash
cd services/vector-search-service

# Download dependencies
go mod tidy

# Copy and configure environment
cp .env.example .env.local
# Edit .env.local with your database credentials
```

### Running

```bash
# Development
go run .

# With custom .env file
ENV_FILE=.env.local go run .

# Build and run binary
go build -o vector-search-service -ldflags="-X main.Version=1.0.0" .
./vector-search-service
```

### Environment Variables

See `.env.example` for all available configuration options.

Required variables:
- `DATABASE_URL` - PostgreSQL connection string

Key configuration:
- `EMBEDDING_DIM=768` - Must match your embedding model dimension
- `DEFAULT_TOP_K=10` - Default number of results
- `RRF_K=60.0` - RRF fusion parameter
- `QUERY_TIMEOUT=30s` - Maximum query duration

## API Reference

### POST /search

Hybrid vector search with dense embedding + sparse keywords and RRF fusion.

**Request:**

```json
{
  "embedding": [0.1, 0.2, ...],
  "keywords": ["machine", "learning"],
  "top_k": 10,
  "filters": {
    "grade": ["9", "10"],
    "subject": "physics",
    "content_type": ["text", "diagram"],
    "chapter": ["mechanics"]
  }
}
```

**Response:**

```json
{
  "data": {
    "results": [
      {
        "parent_id": "uuid-string",
        "content": "Parent chunk text content...",
        "parent_content": "Parent chunk text content...",
        "metadata": {
          "page_number": 5,
          "content_type": "text",
          "chapter": "mechanics",
          "section": "Newton's Laws",
          "subsection": "First Law",
          "grade": "9",
          "subject": "physics"
        },
        "dense_score": 0.15,
        "sparse_score": 3.0,
        "rrf_score": 0.0164,
        "dense_rank": 1,
        "sparse_rank": 2
      }
    ],
    "total_results": 10,
    "search_time_ms": 45,
    "stats": {
      "dense_time_ms": 20,
      "sparse_time_ms": 15,
      "rrf_time_ms": 2,
      "total_time_ms": 45,
      "dense_results": 200,
      "sparse_results": 50,
      "fused_results": 10
    }
  }
}
```

### POST /search/dense

Pure vector similarity search using ScaNN cosine distance.

**Request:**

```json
{
  "embedding": [0.1, 0.2, ...],
  "top_k": 10,
  "filters": {
    "grade": ["9"]
  }
}
```

**Response:** Same format as `/search`, but only `dense_score` and `dense_rank` are populated.

### POST /search/sparse

Pure keyword search using GIN index intersection.

**Request:**

```json
{
  "keywords": ["newton", "laws", "motion"],
  "top_k": 10
}
```

**Response:** Same format as `/search`, but only `sparse_score` and `sparse_rank` are populated.

### GET /health

Health check endpoint with database connectivity status.

**Response:**

```json
{
  "status": "healthy",
  "timestamp": "2026-04-03T12:00:00Z",
  "service": "vector-search-service",
  "version": "1.0.0",
  "database": "connected",
  "pool_stats": {
    "acquire_count": 100,
    "acquired_conns": 5,
    "idle_conns": 20,
    "max_conns": 25,
    "total_conns": 25
  }
}
```

### GET /stats

Service configuration and database pool statistics.

### GET /version

Build version information.

## RRF Algorithm

Reciprocal Rank Fusion combines results from multiple ranking systems using the formula:

```
RRF_score(d) = sum(1 / (k + rank_i(d))) for each ranking system i
```

Where:
- `k` = 60 (default, empirically optimized)
- `rank_i(d)` = rank of document d in ranking system i

The algorithm:
1. Fetch top-K results from dense search (ScaNN cosine similarity)
2. Fetch top-K results from sparse search (GIN keyword intersection)
3. Calculate RRF scores for each unique document
4. Sort by RRF score descending
5. Deduplicate by parent_id
6. Return top-K final results

## Database Schema Requirements

The service expects the following schema:

### child_chunks table
- `child_id` UUID (primary key)
- `parent_id` UUID (foreign key to parent_chunks)
- `embedding` vector(768) - ScaNN indexed
- `content` TEXT
- `page_number` INT
- `content_type` TEXT

### parent_chunks table
- `parent_id` UUID (primary key)
- `content` TEXT
- `extracted_keywords` TEXT[] - GIN indexed
- `page_number` INT
- `content_type` TEXT
- `chapter` TEXT
- `section` TEXT
- `subsection` TEXT
- `grade` TEXT
- `subject` TEXT

### Required Indexes

```sql
-- ScaNN index for dense vector search
CREATE INDEX ON child_chunks USING hnsw (embedding vector_cosine_ops);

-- GIN index for sparse keyword search
CREATE INDEX idx_parent_chunks_keywords ON parent_chunks USING GIN (extracted_keywords);

-- Metadata filter indexes (optional but recommended)
CREATE INDEX idx_parent_chunks_grade ON parent_chunks (grade);
CREATE INDEX idx_parent_chunks_subject ON parent_chunks (subject);
CREATE INDEX idx_parent_chunks_chapter ON parent_chunks (chapter);
```

## Error Handling

All errors follow a consistent format:

```json
{
  "error": {
    "code": "invalid_request",
    "message": "Request body must be valid JSON"
  }
}
```

Common error codes:
- `invalid_request` - Malformed JSON or missing fields
- `missing_criteria` - No embedding or keywords provided
- `invalid_embedding` - Wrong embedding dimension
- `missing_embedding` - Embedding required for dense search
- `missing_keywords` - Keywords required for sparse search
- `search_timeout` - Query exceeded timeout
- `search_error` - Internal search failure

## Development

### Running Tests

```bash
go test ./...
```

### Code Style

Follows standard Go conventions:
- Idiomatic Go naming and structure
- Error handling with wrapped errors
- Context propagation for cancellation
- Prepared statements for repeated queries

### Building for Production

```bash
go build -ldflags="-X main.Version=$(git describe --tags) -X main.GitCommit=$(git rev-parse HEAD)" .
```

## License

Internal use only - Visionary RAG Pipeline.

---

## See Also

### Related Services
- 📦 [API Gateway](../api-gateway/) — Main HTTP gateway with CAG orchestration
- 📦 [Embedding Service](../embedding-service/) — Vertex AI embeddings with caching
- 📦 [Query Understanding Service](../query-understanding-service/) — Query rewriting and sanitization

### Architecture & Design
- 📄 [Complete Architecture](../../docs/architecture/MICROSERVICES_COMPLETE.md) — Full microservices architecture
- 📄 [CAG Architecture](../../docs/caching/CAG_ARCHITECTURE.md) — 5-layer caching design
- 📄 [Main README](../../README.md) — Project overview
- 📄 [Complete Documentation](../../docs/COMPLETE_DOCUMENTATION.md) — Full technical reference (~1500 lines)

### Technical Guides
- 📖 [Retrieval Guide](../../docs/guides/guide_retrieval.md) — Retrieval strategies
- 📖 [Reranking Guide](../../docs/guides/guide_reranking.md) — Reranking strategies
- 📖 [Latency Reduction](../../docs/guides/guide_latency_reduction.md) — Performance optimization

### Audits & Performance
- 📊 [Advanced Retrieval Audit](../../docs/audits/audit_langchaingo_advanced_retrieval.md) — Retrieval analysis
- 📊 [Reranking Audit](../../docs/audits/audit_langchaingo_reranking.md) — Reranking analysis
- 📊 [Vector DB Audit](../../docs/audits/vector_db_audit.md) — Vector database comparison
- 📊 [Latency Audit](../../docs/audits/audit_langchaingo_latency.md) — Pipeline latency

### Quality & Security
- 🛡️ [Defensive Fixes](../../DEFENSIVE_FIXES.md) — Defensive programming improvements
- 🛡️ [Code Quality](../../docs/quality/CODE_QUALITY_IMPROVEMENTS.md) — Code quality improvements

### Deployment
- 🚀 [Docker Compose](../../docker-compose.yml) — Local deployment
- 🚀 [Supabase Guide](../../supabase/README.md) — Cloud deployment
- 🚀 [Terraform/GCP](../../terraform/) — Enterprise deployment
