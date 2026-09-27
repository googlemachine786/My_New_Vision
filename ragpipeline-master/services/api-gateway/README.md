# API Gateway - RAG Microservices

Production-ready API Gateway for the RAG (Retrieval-Augmented Generation) microservices architecture.

## Overview

This API Gateway is the orchestration layer for a 4-service microservices architecture:

- **API Gateway** (port 8080) - This service
- **Embedding Service** (port 8081) - Vertex AI embeddings
- **Vector Search Service** (port 8082) - Hybrid search (dense + sparse + RRF)
- **Query Understanding Service** (port 8083) - Query rewriting, self-query parsing

## Architecture

```
Client -> API Gateway -> [Embedding Service -> Vector Search Service -> LLM Service]
                         -> Query Understanding Service (for query rewriting)
                         -> Redis (caching, sessions, rate limiting)
```

## Features

- **Request Orchestration**: Coordinates all downstream services for the `/query` endpoint
- **SSE Streaming**: Real-time response streaming with Server-Sent Events
- **Response Caching**: Redis-backed cache with SHA-256 keyed lookups
- **Session Management**: Redis-backed conversation history (20 turns max)
- **Circuit Breakers**: Per-service circuit breakers prevent cascading failures
- **Rate Limiting**: Distributed token bucket rate limiting (100 req/min per user)
- **JWT Authentication**: Bearer token validation with configurable skip paths
- **Health Checks**: Aggregated health monitoring of all downstream services
- **Observability**: Structured logging, request IDs, OpenTelemetry tracing

## Quick Start

### Prerequisites

- Go 1.21+
- Redis 6+
- All downstream services running

### Installation

```bash
cd services/api-gateway

# Copy and configure environment
cp .env.example .env

# Download dependencies
go mod tidy

# Build
go build -o api-gateway -ldflags="-X main.Version=1.0.0 -X main.Commit=$(git rev-parse HEAD) -X main.BuildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Run
./api-gateway
```

### Docker

```bash
docker build -t api-gateway .
docker run -p 8080:8080 --env-file .env api-gateway
```

## API Endpoints

### POST /query

Main query endpoint with optional SSE streaming.

**Request:**
```json
{
  "query": "What is quantum physics?",
  "session_id": "optional-uuid",
  "grade": "10",
  "subject": "physics",
  "stream": true
}
```

**Response (Streaming):**
```
event: start
data: {"session_id": "uuid", "request_id": "req-123"}

event: chunk
data: {"content": "Quantum physics is..."}

event: end
data: {"sources": [...], "total_tokens": 150}
```

**Response (Non-streaming):**
```json
{
  "session_id": "uuid",
  "request_id": "req-123",
  "answer": "Quantum physics is...",
  "sources": [...],
  "total_tokens": 150
}
```

### GET /health

Aggregated health check.

```json
{
  "status": "healthy",
  "timestamp": "2026-04-03T10:30:00Z",
  "services": {
    "embedding_service": "healthy",
    "vector_search": "healthy",
    "query_understanding": "healthy",
    "redis": "healthy"
  }
}
```

### GET /stats

Service statistics.

```json
{
  "request_count": 1500,
  "error_count": 12,
  "average_latency": "250ms",
  "cache_hit_rate": 35.5,
  "circuit_breakers": {...},
  "rate_limiting": {...},
  "uptime": "24h"
}
```

### POST /feedback

Submit user feedback.

```json
{
  "query": "What is quantum physics?",
  "response_id": "resp-123",
  "session_id": "session-456",
  "thumbs_up": true,
  "rating": 5,
  "comment": "Great answer!"
}
```

### GET /version

Build information.

```json
{
  "version": "1.0.0",
  "commit": "abc123",
  "build_time": "2026-04-03T10:00:00Z"
}
```

## Configuration

All configuration is via environment variables. See `.env.example` for all options.

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_PORT` | 8080 | HTTP server port |
| `SERVER_ENV` | development | Environment (development/staging/production) |
| `EMBEDDING_SERVICE_URL` | http://localhost:8081 | Embedding service URL |
| `VECTOR_SEARCH_SERVICE_URL` | http://localhost:8082 | Vector search service URL |
| `QUERY_UNDERSTANDING_URL` | http://localhost:8083 | Query understanding service URL |
| `LLM_SERVICE_URL` | http://localhost:8084 | LLM service URL |
| `REDIS_ADDR` | localhost:6379 | Redis address |
| `JWT_SECRET` | change-me-in-production | JWT signing secret |
| `RATE_LIMIT_RPM` | 100 | Requests per minute per user |

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `unauthorized` | 401 | Invalid or missing JWT token |
| `rate_limit_exceeded` | 429 | Too many requests |
| `validation_error` | 400 | Invalid request parameters |
| `embedding_failed` | 502 | Embedding service error |
| `search_failed` | 502 | Vector search service error |
| `llm_failed` | 502 | LLM service error |
| `service_unavailable` | 503 | Downstream service unavailable |
| `internal_error` | 500 | Internal server error |

## Middleware Chain

Requests pass through the following middleware in order:

1. **Recovery** - Catches panics, returns 500
2. **Request ID** - Adds X-Request-ID header
3. **Logging** - Structured JSON logging
4. **CORS** - Cross-origin headers
5. **Rate Limiting** - Token bucket algorithm
6. **Authentication** - JWT validation

## Circuit Breaker

Each downstream service has an independent circuit breaker:

- **Closed**: Normal operation
- **Open**: Failing, requests rejected immediately
- **Half-Open**: Testing recovery

Configuration:
- Trip after 5 consecutive failures
- Reset after 60 seconds
- Allow 3 test requests in half-open state

## Security

- JWT Bearer token validation
- Rate limiting per user
- Request ID propagation
- No sensitive data in logs
- HTTPS via reverse proxy in production

## Monitoring

### Metrics

- Request count per service
- Error rate per service
- Average latency per service
- Cache hit/miss rate
- Circuit breaker states
- Rate limit statistics

### Logging

Structured JSON logs with:
- Request ID
- User ID (if authenticated)
- HTTP method, path, status
- Duration
- Remote address

### Tracing

OpenTelemetry spans for each downstream service call.

## Project Structure

```
services/api-gateway/
├── main.go                    # Entry point
├── version.go                 # Build info
├── go.mod                     # Dependencies
├── .env.example               # Environment template
├── config/
│   ├── config.go              # Configuration loading
│   └── env_helpers.go         # Environment parsing
├── client/
│   ├── service_client.go      # Downstream service client
│   └── circuit_breaker.go     # Circuit breaker pattern
├── handler/
│   ├── query_handler.go       # Main query handler
│   ├── health_handler.go      # Health check
│   ├── stats_handler.go       # Stats endpoint
│   └── feedback_handler.go    # Feedback endpoint
├── cache/
│   └── response_cache.go      # Response caching
├── session/
│   └── session_manager.go     # Session management
└── middleware/
    ├── auth.go                # JWT authentication
    ├── rate_limit.go          # Rate limiting
    ├── cors.go                # CORS headers
    ├── request_id.go          # Request ID
    ├── logging.go             # Structured logging
    └── recovery.go            # Panic recovery
```

## License

Private - All rights reserved

---

## See Also

### Related Services
- 📦 [Embedding Service](../embedding-service/) — Vertex AI embeddings with caching
- 📦 [Vector Search Service](../vector-search-service/) — Hybrid search (dense + sparse + RRF)
- 📦 [Query Understanding Service](../query-understanding-service/) — Query rewriting and sanitization

### Architecture & Design
- 📄 [Complete Architecture](../../docs/architecture/MICROSERVICES_COMPLETE.md) — Full microservices architecture
- 📄 [CAG Architecture](../../docs/caching/CAG_ARCHITECTURE.md) — 5-layer caching design
- 📄 [CAG Implementation](../../docs/caching/CAG_IMPLEMENTATION_COMPLETE.md) — Implementation complete
- 📄 [Main README](../../README.md) — Project overview
- 📄 [Complete Documentation](../../docs/COMPLETE_DOCUMENTATION.md) — Full technical reference (~1500 lines)

### Audits & Performance
- 📊 [Latency Audit](../../docs/audits/audit_langchaingo_latency.md) — Pipeline latency analysis
- 📊 [Cost Audit](../../docs/audits/audit_langchaingo_cost.md) — Cost analysis with caching impact
- 📊 [Redundancy Audit](../../docs/audits/REDUNDANCY_AUDIT.md) — Code redundancy findings
- 📊 [Monitoring Audit](../../docs/audits/monitoring_audit.md) — Monitoring infrastructure

### Technical Guides
- 📖 [Retrieval Guide](../../docs/guides/guide_retrieval.md) — Retrieval strategies
- 📖 [Cost Optimization](../../docs/guides/guide_cost_optimization.md) — Cost reduction techniques
- 📖 [Latency Reduction](../../docs/guides/guide_latency_reduction.md) — Performance optimization

### Quality & Security
- 🛡️ [Defensive Fixes](../../DEFENSIVE_FIXES.md) — Defensive programming improvements
- 🛡️ [API Design Fixes](../../GO_API_DESIGN_FIXES.md) — API design improvements
- 🛡️ [Code Quality](../../docs/quality/CODE_QUALITY_IMPROVEMENTS.md) — Code quality improvements

### Deployment
- 🚀 [Docker Compose](../../docker-compose.yml) — Local deployment
- 🚀 [Supabase Guide](../../supabase/README.md) — Cloud deployment
- 🚀 [Terraform/GCP](../../terraform/) — Enterprise deployment
