# Go Microservices Architecture

## Overview

The Visionary RAG Pipeline is a production-grade Retrieval-Augmented Generation system built as Go microservices for CBSE Science education (Grades 6-12).

## Architecture

```
Client → API Gateway (8080) → Embedding (8081) + Vector Search (8082) + Query Understanding (8083)
```

## Services

| Service | Port | Purpose |
|---------|------|---------|
| [API Gateway](../services/api-gateway/README.md) | 8080 | Auth, caching, CAG orchestration, SSE streaming |
| [Embedding Service](../services/embedding-service/README.md) | 8081 | Vertex AI embeddings with retry and caching |
| [Vector Search](../services/vector-search-service/README.md) | 8082 | Hybrid search (dense ScaNN + sparse GIN + RRF) |
| [Query Understanding](../services/query-understanding-service/README.md) | 8083 | Gemini LLM, query rewriting, self-query parsing |

## Key Features

- **5-Layer CAG Caching**: 72% cost reduction (~$854/month savings)
- **Circuit Breakers**: Fault tolerance for all downstream services
- **SSE Streaming**: Real-time LLM responses
- **Redis Caching**: Response, session, and rate limit storage
- **JWT Authentication**: Per-user auth and rate limiting

## Documentation

- [CAG Architecture](../docs/caching/CAG_ARCHITECTURE.md)
- [Migration Guide](../docs/migration/MIGRATION_GUIDE.md)
- [Production Readiness](../docs/PRODUCTION_READINESS_CHECKLIST.md)
- [Code Quality Improvements](../docs/quality/CODE_QUALITY_IMPROVEMENTS.md)
- [Functional Options Fixes](../docs/quality/FUNCTIONAL_OPTIONS_FIXES.md)
- [Go API Design Fixes](../docs/quality/GO_API_DESIGN_FIXES.md)

## Quick Start

```bash
# Start all services
docker-compose up -d

# Or run individually
cd services/api-gateway && go run cmd/server/main.go
```
