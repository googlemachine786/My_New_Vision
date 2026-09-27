# Audit: RAG Latency Guide vs. Implementation

**Guide:** `guide_rag_latency.md` (Ailog, Nov 2025)
**Audited:** 2026-04-03
**Scope:** C:\Users\kommi\ragpipeline\

---

## Guide Summary

This guide covers reducing RAG latency from 2000ms to 200ms through parallel retrieval, streaming responses, approximate nearest neighbors (HNSW tuning), smaller reranking models, reduced context size, and edge caching.

---

## Key Recommendations and Implementation Match

### 1. Parallel Retrieval

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Run embedding + search concurrently with asyncio | |
| **Implementation** | **PARTIALLY MATCHES** | The Go-based retrieval is inherently concurrent (connection pooling via pgxpool). The Python ingestion pipeline uses async operations. However, query-time parallel retrieval (embedding + multiple index searches simultaneously) is not explicitly implemented. |
| **Code Reference** | `hybrid_search.go`: Sequential dense then sparse search (lines 48-55). Not parallelized. |
| **Missing** | No `asyncio.gather()` for simultaneous embedding + search. Dense and sparse searches are sequential, not parallel. |

### 2. Streaming Responses

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Stream LLM response tokens so user sees first token in ~150ms | |
| **Implementation** | **DOES NOT MATCH** | No streaming implementation found. The LLM judge in `judge.go` uses batch prediction (not streaming). The `clarification_generator.go` uses `Generate()` which is non-streaming. |
| **Missing** | No `stream=True` in LLM calls, no token-by-token yielding, no SSE (Server-Sent Events) endpoint. |

### 3. Approximate Nearest Neighbors (HNSW Tuning)

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Use HNSW with tuned parameters (M, ef_construct, ef) for 10x faster search | |
| **Implementation** | **PARTIALLY MATCHES** | Uses pgvector which supports HNSW. However, no explicit HNSW configuration parameters found. The guide recommends tuning M, ef_construct for index build and ef for query time. |
| **Missing** | No `CREATE INDEX ... WITH (m = 16, ef_construction = 100)` DDL. No `hnsw_ef` search parameter configuration. |

### 4. Smaller Reranking Models

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Use TinyBERT reranker for fast (~50ms) scoring of 20 docs | |
| **Implementation** | **DOES NOT MATCH** | No reranker of any kind exists. RRF fusion provides ranking without a dedicated model. |

### 5. Reduce Context Size

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Send fewer, shorter docs to LLM (5 short docs instead of 10 long) | |
| **Implementation** | **MATCH** | Default `top_k=5` in `rag_config.py` line 57. The `high_quality` preset uses `top_k=7`, `low_latency` uses `top_k=8`. Max tokens limited to 143-150. |
| **Code Reference** | `rag_config.py` line 57: `top_k: int = 5`. Lines 71, 79: `max_tokens: int = 143-150`. |

### 6. Edge Caching

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | CDN-level caching for popular queries (Cloudflare Workers example) | |
| **Implementation** | **PARTIALLY MATCHES** | Redis response caching exists (`response_cache.go`), but it is application-level, not edge/CDN-level. The cache uses SHA256 of query embeddings as keys. |
| **Code Reference** | `C:\Users\kommi\ragpipeline\orchestrator\cache\response_cache.go` lines 30-50. |
| **Missing** | No CDN/edge caching layer (Cloudflare, CloudFront). |

### 7. Complete Optimized Pipeline

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Cache check -> parallel embed+search -> fast rerank -> stream response (~200ms) | |
| **Implementation** | **PARTIALLY MATCHES** | The pipeline has caching and hybrid search but lacks parallel execution, reranking, and streaming. |
| **Missing** | No end-to-end optimized low-latency pipeline combining all techniques. |

### 8. Latency Breakdown Tracking

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Track embedding time, search time, rerank time, LLM time separately | |
| **Implementation** | **MATCH** | OpenTelemetry spans track individual stages: `SpanEmbedding`, `SpanHybridSearch`, `SpanRRFFusion`, `SpanGeneration`. TTFT and total latency tracked as histograms. |
| **Code Reference** | `C:\Users\kommi\ragpipeline\orchestrator\observability\tracer.go` lines 84-93: Span constants. `metrics.go` lines 36-42: TTFT and latency histograms. |

### 9. Target Latency Benchmarks

| Guide Target | Current Status | Match |
|--------------|---------------|-------|
| Embed query: 20ms (cached) | Uses local all-MiniLM (~50ms) | **PARTIALLY** |
| Vector search: 30ms (optimized) | pgvector with no HNSW tuning | **PARTIALLY** |
| Rerank: 50ms (parallel) | No reranker | **DOES NOT MATCH** |
| LLM generation: 100ms (streaming) | No streaming, Gemini batch | **DOES NOT MATCH** |
| Total: ~200ms | No measured baseline | **CANNOT DETERMINE** |

---

## Gap Analysis Summary

### Implemented
- Reduced context size (top_k=5, max_tokens=143-150)
- Redis response caching
- Per-stage latency tracking with OpenTelemetry
- Connection pooling for DB operations

### Partially Implemented
- Parallel retrieval (sequential dense+sparse, not parallel)
- Edge caching (application-level Redis, not CDN)
- Embedding speed (local model, but not cached for queries)
- HNSW (pgvector supports it, but not explicitly tuned)

### NOT Implemented
- Streaming LLM responses
- Dedicated fast reranker (TinyBERT)
- Explicit HNSW parameter tuning
- Parallel embedding + search execution
- CDN/edge caching
- End-to-end low-latency pipeline assembly

---

## Overall Assessment: PARTIALLY MATCHES

The project has foundational latency optimization pieces in place (caching, small context, telemetry) but is missing the highest-impact optimizations: streaming responses and parallel retrieval. The sequential dense+sparse search should be parallelized, and streaming would dramatically improve perceived latency. Without a reranker, the guide's reranking optimization recommendation cannot be assessed.
