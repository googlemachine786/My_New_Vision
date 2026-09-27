# Embedding Service

Standalone Go HTTP service for generating text embeddings using Google Vertex AI `text-embedding-005` model.

## Features

- **Vertex AI Integration**: Uses `text-embedding-005` model with 768 dimensions
- **Task Type Support**: `RETRIEVAL_DOCUMENT` for indexing, `RETRIEVAL_QUERY` for search
- **Batch Processing**: Up to 5 texts per request (configurable)
- **Exponential Backoff Retry**: 5 attempts max with configurable backoff
- **OpenTelemetry Tracing**: Distributed tracing support
- **Structured Logging**: JSON logs with zerolog, request ID correlation
- **Request Validation**: Text length limits, batch size validation
- **Health & Stats Endpoints**: Service monitoring

## Quick Start

### Prerequisites

- Go 1.23+
- Google Cloud project with Vertex AI API enabled
- Service account with `roles/aiplatform.user`
- `gcloud` CLI authenticated (for local development)

### Setup

1. **Clone and navigate**:
   ```bash
   cd services/embedding-service
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Configure environment**:
   ```bash
   cp .env.example .env
   # Edit .env with your GCP project ID and settings
   ```

4. **Authenticate with GCP** (local development):
   ```bash
   gcloud auth application-default login
   ```

5. **Run the service**:
   ```bash
   go run main.go
   ```

The service will start on `http://localhost:8081`.

## API Endpoints

### POST /embed

Generate embeddings for one or more texts.

**Request**:
```json
{
  "texts": ["What is photosynthesis?", "Plants convert light to energy"],
  "task_type": "RETRIEVAL_QUERY"
}
```

**Response**:
```json
{
  "data": {
    "embeddings": [
      [0.012, -0.034, 0.056, ...],
      [-0.021, 0.045, -0.012, ...]
    ],
    "dimensions": 768
  }
}
```

**Parameters**:
- `texts` (array, required): 1-5 text strings to embed
- `task_type` (string, optional): `RETRIEVAL_DOCUMENT` or `RETRIEVAL_QUERY` (default: `RETRIEVAL_QUERY`)

**Validation**:
- Maximum 5 texts per batch (configurable via `MAX_TEXTS_PER_BATCH`)
- Maximum 20,000 characters per text (configurable via `MAX_TEXT_LENGTH`)

**Error Response**:
```json
{
  "error": {
    "code": "invalid_json",
    "message": "Request body must be valid JSON"
  }
}
```

### GET /health

Check service health and Vertex AI connectivity.

**Response** (healthy):
```json
{
  "status": "healthy",
  "timestamp": "2026-04-03T10:00:00Z",
  "service": "embedding-service",
  "vertex_ai": "healthy"
}
```

**Response** (degraded):
```json
{
  "status": "degraded",
  "timestamp": "2026-04-03T10:00:00Z",
  "service": "embedding-service",
  "vertex_ai": "unhealthy"
}
```

### GET /stats

Get service statistics.

**Response**:
```json
{
  "service": "embedding-service",
  "config": {
    "model": "text-embedding-005",
    "dimension": 768,
    "max_texts_per_batch": 5,
    "max_text_length": 20000,
    "max_retries": 5
  },
  "vertex": {
    "project": "my-project",
    "location": "us-central1",
    "model": "text-embedding-005",
    "dimension": 768,
    "request_count": 1234,
    "error_count": 5,
    "avg_latency_ms": 145.2,
    "last_request": "2026-04-03T10:00:00Z"
  }
}
```

### GET /version

Get service version information.

**Response**:
```json
{
  "version": "dev",
  "git_commit": "unknown",
  "build_time": "unknown"
}
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `invalid_json` | 400 | Request body is not valid JSON |
| `missing_texts` | 400 | No texts provided |
| `batch_too_large` | 400 | Too many texts in batch |
| `empty_text` | 400 | One of the texts is empty |
| `text_too_long` | 400 | Text exceeds max length |
| `invalid_task_type` | 400 | Invalid task_type value |
| `vertex_ai_error` | 502 | Vertex AI API call failed |
| `method_not_allowed` | 405 | Wrong HTTP method |
| `not_found` | 404 | Endpoint not found |
| `internal_error` | 500 | Internal server error |

## Environment Variables

See `.env.example` for all available configuration options.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `VERTEX_PROJECT` | Yes | - | GCP project ID |
| `VERTEX_LOCATION` | No | `us-central1` | GCP region |
| `VERTEX_MODEL` | No | `text-embedding-005` | Embedding model |
| `VERTEX_DIMENSION` | No | `768` | Embedding dimension |
| `PORT` | No | `8081` | HTTP server port |
| `MAX_RETRIES` | No | `5` | Max retry attempts |
| `MAX_TEXTS_PER_BATCH` | No | `5` | Max texts per batch |
| `MAX_TEXT_LENGTH` | No | `20000` | Max chars per text |
| `LOG_LEVEL` | No | `info` | Log level |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | No | - | OTLP endpoint |

## Testing

### Manual Testing with curl

```bash
# Health check
curl http://localhost:8081/health

# Single text embedding
curl -X POST http://localhost:8081/embed \
  -H "Content-Type: application/json" \
  -d '{"texts": ["What is machine learning?"], "task_type": "RETRIEVAL_QUERY"}'

# Batch embedding
curl -X POST http://localhost:8081/embed \
  -H "Content-Type: application/json" \
  -d '{"texts": ["Text one", "Text two", "Text three"]}'

# Stats
curl http://localhost:8081/stats
```

### Build for Production

```bash
# Build with version info
go build -ldflags "-X main.Version=1.0.0 -X main.GitCommit=$(git rev-parse HEAD) -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o embedding-service

# Run
./embedding-service
```

## Deployment

### Cloud Run

```bash
gcloud run deploy embedding-service \
  --source . \
  --platform managed \
  --region us-central1 \
  --set-env-vars VERTEX_PROJECT=my-project,VERTEX_LOCATION=us-central1 \
  --service-account embedding-sa@my-project.iam.gserviceaccount.com
```

### Docker

```bash
docker build -t embedding-service .
docker run -p 8081:8081 \
  -e VERTEX_PROJECT=my-project \
  -e VERTEX_LOCATION=us-central1 \
  -e GOOGLE_APPLICATION_CREDENTIALS=/key.json \
  -v /path/to/key.json:/key.json:ro \
  embedding-service
```

## Architecture

```
Request -> Request ID Middleware -> Logging Middleware -> Recovery Middleware -> Router
                                                                              |-> POST /embed -> Vertex AI Client
                                                                              |-> GET /health
                                                                              |-> GET /stats
                                                                              |-> GET /version
```

### Middleware Chain

1. **Recovery**: Catches panics, returns 500
2. **Request ID**: Injects UUID into context and response header
3. **Logging**: Logs request method, path, latency, and request ID

### Vertex AI Client

- Uses `cloud.google.com/go/vertexai` Go SDK
- Exponential backoff retry (configurable)
- Automatic retry on transient errors (429, 500, 503)
- Request/response statistics tracking

## Observability

### Logging

All logs include `request_id` for correlation. Example:
```json
{
  "level": "info",
  "method": "POST",
  "path": "/embed",
  "request_id": "abc-123-def",
  "latency_ms": 145,
  "message": "Request completed"
}
```

### Tracing

When `OTEL_EXPORTER_OTLP_ENDPOINT` is set, the service exports traces to the configured collector. Spans include:
- `vertex.embed_text` - Single text embedding
- `vertex.embed_batch` - Batch embedding
- `vertex.health_check` - Health check

## License

Private - All rights reserved

---

## See Also

### Related Services
- 📦 [API Gateway](../api-gateway/) — Main HTTP gateway with CAG orchestration
- 📦 [Vector Search Service](../vector-search-service/) — Hybrid search (dense + sparse + RRF)
- 📦 [Query Understanding Service](../query-understanding-service/) — Query rewriting and sanitization

### Architecture & Design
- 📄 [Complete Architecture](../../docs/architecture/MICROSERVICES_COMPLETE.md) — Full microservices architecture
- 📄 [CAG Architecture](../../docs/caching/CAG_ARCHITECTURE.md) — 5-layer caching design
- 📄 [Main README](../../README.md) — Project overview
- 📄 [Complete Documentation](../../docs/COMPLETE_DOCUMENTATION.md) — Full technical reference (~1500 lines)

### Technical Guides
- 📖 [Embedding Guide](../../docs/guides/guide_embedding.md) — Embedding models and strategies
- 📖 [Semantic Chunking](../../docs/guides/guide_semantic_chunking.md) — Semantic chunking strategies
- 📖 [Retrieval Guide](../../docs/guides/guide_retrieval.md) — Retrieval strategies

### Audits & Performance
- 📊 [Latency Audit](../../docs/audits/audit_langchaingo_latency.md) — Pipeline latency analysis
- 📊 [Cost Audit](../../docs/audits/audit_langchaingo_cost.md) — Cost analysis
- 📊 [Chunking Audit](../../docs/audits/audit_langchaingo_chunking.md) — Chunking analysis
- 📊 [Semantic Chunking Audit](../../docs/audits/audit_langchaingo_semantic_chunking.md) — Semantic chunking

### Quality & Security
- 🛡️ [Defensive Fixes](../../DEFENSIVE_FIXES.md) — Defensive programming improvements
- 🛡️ [Code Quality](../../docs/quality/CODE_QUALITY_IMPROVEMENTS.md) — Code quality improvements

### Deployment
- 🚀 [Docker Compose](../../docker-compose.yml) — Local deployment
- 🚀 [Supabase Guide](../../supabase/README.md) — Cloud deployment
- 🚀 [Terraform/GCP](../../terraform/) — Enterprise deployment

Private - Visionary RAG Pipeline
