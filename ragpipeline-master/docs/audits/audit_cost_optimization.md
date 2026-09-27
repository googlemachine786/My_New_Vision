# Audit: Cost Optimization Guide vs. Implementation

**Guide:** `guide_cost_optimization.md` (Ailog, Nov 2025)
**Audited:** 2026-04-03
**Scope:** C:\Users\kommi\ragpipeline\

---

## Guide Summary

This guide covers reducing RAG costs by 90% through: smaller embedding models, smart chunking (fewer chunks), aggressive caching, smaller LLMs, reduced context size, batch processing, self-hosted vector DB, lazy reranking, user quotas, and real-time cost monitoring with Prometheus/Grafana.

---

## Key Recommendations and Implementation Match

### 1. Reduce Embedding Costs

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Use smaller/cheaper models (text-embedding-3-small vs large, or open-source) | |
| **Implementation** | **MATCH** | Uses `all-MiniLM-L6-v2` (free, self-hosted) for query embeddings. Uses Vertex AI `text-embedding-005` for document embeddings (moderate cost). |
| **Code Reference** | `rag_config.py` line 54: `embed_model: str = "sentence-transformers/all-MiniLM-L6-v2"`. |

### 2. Smart Chunking

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Larger chunks = fewer embeddings = lower costs (800-token vs 500-token saves 35%) | |
| **Implementation** | **PARTIALLY MATCHES** | Uses 400-character chunks which are smaller than both guide recommendations. This increases embedding count but improves precision. Parent-child chunking means both parent and child are embedded. |
| **Note** | This is a deliberate trade-off: more embeddings for better retrieval quality. The grid search validated 400-char as optimal. |

### 3. Aggressive Caching

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Cache everything with Redis (90% hit rate -> 90% cost savings) | |
| **Implementation** | **MATCH** | Redis response caching with configurable TTL (300 seconds default). Caches by query embedding hash. |
| **Code Reference** | `C:\Users\kommi\ragpipeline\orchestrator\cache\response_cache.go` lines 30-50. `rag_config.py` line 63: `cache_enabled: bool = True`, `cache_ttl_seconds: int = 300`. |
| **Missing** | No semantic similarity caching (stub exists but marked TODO). No query text caching (only embedding hash). |

### 4. Use Smaller LLMs

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | GPT-4o-mini instead of GPT-4 Turbo (60x cheaper) | |
| **Implementation** | **MATCH** | Uses Gemini 2.0 Flash for LLM generation and LLM judge, which is a cost-efficient model. Uses llama3.2:3b (local, free) for self-query parsing and clarification. |
| **Code Reference** | `judge.go` line 57: `model: "gemini-2.0-flash"`. `bench_and_improve.py` line 34: `LLM_MODEL = "llama3.2:3b"`. |

### 5. Reduce Context Size

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Send top 3 docs (1500 tokens) instead of top 10 (5000 tokens) = 70% savings | |
| **Implementation** | **MATCH** | Default top_k=5, which is already conservative. Max tokens limited to 143-150. |
| **Code Reference** | `rag_config.py` line 57: `top_k: int = 5`. Line 69: `max_tokens: int = 150`. |

### 6. Batch Processing

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Batch embed multiple queries together instead of individual API calls | |
| **Implementation** | **MATCH** | Vertex AI embedder batches texts (max 5 per API call, which is the Vertex AI limit). |
| **Code Reference** | `vertex_batch.py` lines 76-96: `embed_texts()` with batch size of 5. Lines 104-109: `_embed_batch_with_retry()` for batch API calls. |

### 7. Self-Hosted Vector DB

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Self-host Qdrant instead of Pinecone ($20 vs $70/month) | |
| **Implementation** | **PARTIALLY MATCHES** | Uses AlloyDB (managed Google Cloud PostgreSQL with pgvector). Not self-hosted but uses existing PostgreSQL infrastructure rather than a dedicated vector DB service. |
| **Note** | AlloyDB is managed, so there is a cost, but it serves dual purpose (relational + vector). |

### 8. Lazy Reranking

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Skip reranking if top result has high confidence score (>0.9) | |
| **Implementation** | **DOES NOT MATCH** | No reranking exists, so no lazy reranking either. |

### 9. User Quotas / Rate Limiting

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Rate limit users to prevent abuse (max queries per window) | |
| **Implementation** | **DOES NOT MATCH** | No rate limiting or quota system found in the codebase. |
| **Missing** | No per-user query limits, no sliding window rate limiter, no abuse prevention. |

### 10. Monitoring and Alerts

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Track costs in real-time with Prometheus counters, set alerts for thresholds | |
| **Implementation** | **PARTIALLY MATCHES** | OpenTelemetry metrics track requests, errors, cache hits/misses, latency, chunks retrieved. Cost tracking exists conceptually but is not wired to actual cost calculations. |
| **Code Reference** | `metrics.go`: counters for requests, errors, cache hits/misses. Histograms for TTFT, latency, chunks. |
| **Missing** | No actual cost-per-query calculation in USD. No Prometheus alerting rules configured. No Grafana dashboards. No cost threshold alerts. |

### 11. Complete Cost Optimization Pipeline

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Combine all optimizations: caching + cheap embeddings + fewer docs + smart reranking + cheap LLM = ~95% reduction | |
| **Implementation** | **PARTIALLY MATCHES** | Several optimizations are in place (caching, cheap query embeddings, small context, small LLM), but they are not assembled into a single cost-optimized pipeline. |

---

## Gap Analysis Summary

### Implemented
- Free embedding model for queries (all-MiniLM-L6-v2)
- Redis response caching with configurable TTL
- Gemini 2.0 Flash for cost-efficient generation
- Local llama3.2:3b for self-query parsing
- Conservative top_k=5 and max_tokens=150
- Batch embedding operations
- OpenTelemetry metrics infrastructure

### Partially Implemented
- Cost tracking (metrics exist but not wired to actual costs)
- Vector DB cost (AlloyDB managed, not self-hosted)

### NOT Implemented
- User quotas / rate limiting
- Actual cost-per-query calculation in USD
- Prometheus alerting rules
- Grafana dashboards
- Semantic caching (stub exists)
- Lazy reranking (no reranker exists)
- Context compression before LLM

---

## Overall Assessment: PARTIALLY MATCHES

The project implements several cost optimization strategies well: free query embeddings, Redis caching, cost-efficient LLM models, and conservative context sizes. The biggest gaps are in user rate limiting (abuse prevention), actual cost tracking in USD, and alerting infrastructure. The cost optimization is good but not comprehensive.
