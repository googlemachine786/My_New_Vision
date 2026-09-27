# Prioritized Code Optimizations — Unimplemented Audit Findings

**Generated:** 2026-04-06
**Source:** All audit files in `docs/audits/`, `docs/guides/`, `.planning/codebase/`, `docs/quality/`, `docs/archive/`
**Scope:** Actionable code optimizations NOT yet implemented in the codebase

---

## P0 — Critical (Must Fix)

### 1. Add Dedicated Cross-Encoder Reranker to Active Pipeline

| Field | Details |
|-------|---------|
| **Source** | `audit_reranking.md`, `guide_reranking.md`, `audit_advanced_retrieval_strategies.md` |
| **Target** | `orchestrator/rag/` or `orchestrator/handler/rag_handler.go` |
| **Current State** | `orchestrator/reranker/` package exists with 6 reranker implementations (vertex, cross_encoder, onnx, ollama) but audits confirm they are **not wired into the active query pipeline**. Pipeline only uses RRF fusion for ranking. |
| **Recommended Fix** | Wire `orchestrator/reranker/cross_encoder.go` (or `onnx_reranker.go`) into the RAG handler pipeline: retrieve topK*20 → rerank with cross-encoder → return topK. Add reranking stage between hybrid search and LLM generation. |
| **Priority** | P0 |
| **Impact** | +10-25% retrieval accuracy (audit_reranking.md: "easiest way to gain another 10-25% accuracy"). Reduces hallucinations, improves answer quality. |

### 2. Parallel Dense + Sparse Search Execution

| Field | Details |
|-------|---------|
| **Source** | `audit_rag_latency.md` |
| **Target** | `orchestrator/retrieval/hybrid_search.go` (or equivalent retrieval module) |
| **Current State** | Dense search and sparse search run **sequentially** (lines 48-55 per audit). No `asyncio.gather()` or `go func()` parallel execution. |
| **Recommended Fix** | Use goroutines to run dense and sparse searches concurrently, then wait for both results before RRF fusion. Example: launch both searches in separate goroutines, use `sync.WaitGroup` or `errgroup.Group`, merge results. |
| **Priority** | P0 |
| **Impact** | ~50% reduction in retrieval latency (search time dominated by slower of two searches instead of sum). Total pipeline latency reduction of 15-30ms. |

### 3. Streaming LLM Responses (End-to-End)

| Field | Details |
|-------|---------|
| **Source** | `audit_rag_latency.md`, `guide_latency_reduction.md` |
| **Target** | `orchestrator/handler/rag_handler.go` + `orchestrator/llm/gemini/` |
| **Current State** | Audit states "No streaming implementation found." LLM judge uses batch prediction. Some `gemini_client.go` files have SSE streaming but it's not confirmed wired through the main RAG handler end-to-end for query responses. |
| **Recommended Fix** | Ensure `GenerateStream()` is called in the main query handler path, tokens are yielded via SSE to the client immediately. Set `stream=True` in LLM calls. Implement token-by-token yielding with proper error handling mid-stream. |
| **Priority** | P0 |
| **Impact** | Time-to-first-token (TTFT) drops from ~350ms (full response) to ~150ms. Dramatically improves perceived latency. |

### 4. User Rate Limiting / Quota System

| Field | Details |
|-------|---------|
| **Source** | `audit_cost_optimization.md`, `guide_cost_optimization.md` |
| **Target** | `orchestrator/middleware/` (new file) or `orchestrator/handler/` |
| **Current State** | No per-user query limits, no sliding window rate limiter, no abuse prevention. Rate limiting exists in `services/api-gateway/middleware/rate_limit.go` but the orchestrator (active Go service) has no equivalent. |
| **Recommended Fix** | Implement per-user rate limiter using Redis-backed sliding window or token bucket. Add middleware that checks query count per user per time window before allowing RAG query. Return 429 when exceeded. |
| **Priority** | P0 |
| **Impact** | Prevents cost explosion from abuse. Protects against quota exhaustion. Critical for production multi-tenant usage. |

---

## P1 — High Priority (Should Fix)

### 5. Explicit HNSW Index Parameter Tuning

| Field | Details |
|-------|---------|
| **Source** | `audit_rag_latency.md`, `audit_vector_databases.md` |
| **Target** | `supabase/migrations/002_hnsw_optimization.sql` + runtime query config |
| **Current State** | HNSW index exists with `m=16, ef_construction=64` (confirmed in `002_hnsw_optimization.sql`). However, **no runtime `hnsw.ef_search` parameter** is configured. The guide recommends tuning `ef` at query time for speed/recall trade-off. |
| **Recommended Fix** | Add `SET LOCAL hnsw.ef_search = 40;` (or calibrated value) before vector search queries in `hybrid_search.go`. Create migration to test higher `ef_construction` values (100-200) for better index quality. |
| **Priority** | P1 |
| **Impact** | 10x faster vector search with proper HNSW tuning (per `audit_rag_latency.md`). |

### 6. Actual Cost-per-Query Tracking in USD

| Field | Details |
|-------|---------|
| **Source** | `audit_cost_optimization.md`, `audit_monitoring.md` |
| **Target** | `orchestrator/observability/metrics.go` |
| **Current State** | Generic `rag_cost_usd` counter exists conceptually but is **not wired to actual cost calculations**. No token counting for embedding/LLM calls multiplied by per-model pricing. |
| **Recommended Fix** | Add token counting in embedding and LLM calls. Multiply by model-specific pricing (e.g., $0.025/1K for Vertex AI embeddings, $0.075/1K chars for Gemini). Increment Prometheus counter with actual USD cost per request. |
| **Priority** | P1 |
| **Impact** | Real-time cost visibility. Enables cost alerts. Critical for budget management. |

### 7. Prometheus Alerting Rules

| Field | Details |
|-------|---------|
| **Source** | `audit_monitoring.md`, `guide_monitoring.md` |
| **Target** | New file: `orchestrator/observability/prometheus_alerts.yml` + Docker Compose |
| **Current State** | No Prometheus alerting rules configuration found. No YAML alerting rules. |
| **Recommended Fix** | Create alerting rules YAML file with thresholds: p95 latency > 500ms, error rate > 5%, cache hit rate < 50%, cost > $X/day. Wire into `docker-compose.yml` Prometheus service configuration. |
| **Priority** | P1 |
| **Impact** | Proactive issue detection before users are impacted. Cost anomaly alerts. |

### 8. Grafana Dashboard Configurations

| Field | Details |
|-------|---------|
| **Source** | `audit_monitoring.md`, `guide_monitoring.md` |
| **Target** | New file: `orchestrator/observability/grafana_dashboards/` (JSON files) |
| **Current State** | No Grafana dashboard configurations (JSON) found. No PromQL query definitions. |
| **Recommended Fix** | Create Grafana dashboard JSON files with panels for: latency histogram, error rate over time, cache hit rate, cost per day, requests per second, chunks retrieved distribution. Provision via `docker-compose.yml`. |
| **Priority** | P1 |
| **Impact** | Visual observability. Faster debugging and performance tuning. |

### 9. Semantic Similarity Caching (Not Just Exact Match)

| Field | Details |
|-------|---------|
| **Source** | `audit_cost_optimization.md`, `audit_vector_databases.md` |
| **Target** | `orchestrator/cache/response_cache.go` (SemanticCache stub) |
| **Current State** | `SemanticCache` stub exists at lines 134-154, marked TODO. Only exact-match caching by SHA256 of query embeddings is active. |
| **Recommended Fix** | Implement semantic caching: store query embeddings with responses. On new query, compute embedding similarity to cached queries. If similarity > 0.95 threshold, return cached response. Use Redis vector similarity or in-memory FAISS. |
| **Priority** | P1 |
| **Impact** | Additional 20-30% cache hit rate beyond exact-match caching. Significant cost savings. |

### 10. A/B Testing Framework

| Field | Details |
|-------|---------|
| **Source** | `audit_monitoring.md`, `guide_monitoring.md` |
| **Target** | New module: `orchestrator/abtesting/` |
| **Current State** | No A/B testing framework found. `ab_test_config.py` exists at root but is a standalone config script, not integrated into the pipeline. |
| **Recommended Fix** | Implement user variant assignment middleware. Log which config variant (e.g., chunking strategy, reranker model, top_k) each query used. Add statistical analysis endpoint. Wire `ab_test_config.py` into the Go orchestrator. |
| **Priority** | P1 |
| **Impact** | Data-driven optimization. Measure impact of configuration changes on real users. |

### 11. Query Expansion for Recall Improvement

| Field | Details |
|-------|---------|
| **Source** | `audit_query_expansion.md` |
| **Target** | New module: `orchestrator/retrieval/query_expansion.go` |
| **Current State** | Self-query retrieval extracts metadata filters but does NOT expand queries for recall improvement. No synonym expansion, no multi-query retrieval, no HyDE, no sub-query decomposition. |
| **Recommended Fix** | Implement at minimum: LLM-based query rewriting to generate 2-3 alternative phrasings → search with all variations → deduplicate and rank by combined score. Start with multi-query retrieval (highest impact, lowest complexity). |
| **Priority** | P1 |
| **Impact** | +30-50% recall improvement (per `audit_query_expansion.md`). "Low-cost, high-impact" optimization. |

### 12. Live Retrieval Quality Metrics (nDCG, Precision@K as Prometheus Metrics)

| Field | Details |
|-------|---------|
| **Source** | `audit_monitoring.md` |
| **Target** | `orchestrator/observability/metrics.go` |
| **Current State** | Recall@K, MRR, NDCG implemented in `rrf.go` but only used in **offline** benchmark evaluation (`bench_and_improve.py`). Not exposed as live Prometheus metrics. |
| **Recommended Fix** | When ground truth is available (golden QA dataset, user feedback), calculate nDCG@5 in real-time and expose as `rag_relevance_ndcg` Prometheus gauge. Track over time to detect retrieval quality regression. |
| **Priority** | P1 |
| **Impact** | Detect retrieval quality degradation in real-time. Measure impact of pipeline changes on actual retrieval quality. |

---

## P2 — Medium Priority (Nice to Have)

### 13. Contextual Compression Before LLM

| Field | Details |
|-------|---------|
| **Source** | `audit_advanced_retrieval_strategies.md`, `guide_advanced_retrieval_strategies.md` |
| **Target** | `orchestrator/llm/context_compressor.go` (new) |
| **Current State** | Full chunk content is passed to the LLM without compression. No contextual compression implementation. |
| **Recommended Fix** | Add LLM-based or heuristic-based extraction of only relevant sentences/paragraphs from retrieved chunks before passing to generation model. Or use a cheap summarization model to compress each chunk to 50 words. |
| **Priority** | P2 |
| **Impact** | 50-70% reduction in context tokens sent to LLM. Lower costs, potentially faster generation. |

### 14. Maximal Marginal Relevance (MMR) for Diversity

| Field | Details |
|-------|---------|
| **Source** | `audit_advanced_retrieval_strategies.md`, `guide_advanced_retrieval_strategies.md` |
| **Target** | `orchestrator/retrieval/mmr.go` (new) |
| **Current State** | No MMR implementation. No diversity scoring in retrieval pipeline. |
| **Recommended Fix** | Implement MMR algorithm: after retrieval, select top-K documents balancing relevance vs. diversity (lambda parameter). Replace redundant highly-similar chunks with less-similar alternatives. |
| **Priority** | P2 |
| **Impact** | More diverse context for LLM generation. Reduces redundant information in prompts. |

### 15. Multi-Query Retrieval with Score Aggregation

| Field | Details |
|-------|---------|
| **Source** | `audit_query_expansion.md` |
| **Target** | `orchestrator/retrieval/multi_query.go` (new) |
| **Current State** | Each query results in a single embedding and single search. No parallel search with multiple query variations. |
| **Recommended Fix** | After query rewriting (or LLM-generated variations), embed each variation, search for each, deduplicate results by parent_id, aggregate scores across all query variations. |
| **Priority** | P2 |
| **Impact** | +20-40% recall for ambiguous or complex queries. |

### 16. Domain-Specific Embedding Model Fine-Tuning

| Field | Details |
|-------|---------|
| **Source** | `audit_embedding_models.md` |
| **Target** | New training pipeline script: `scripts/fine tune_embeddings.py` |
| **Current State** | No fine-tuning of embedding models. Uses pre-trained models without domain adaptation for CBSE Science content. |
| **Recommended Fix** | Create training pairs from textbook content (question-answer, section-summary pairs). Fine-tune `all-MiniLM-L6-v2` with `MultipleNegativesRankingLoss` using `sentence_transformers`. Expected +10-30% retrieval improvement for domain-specific content. |
| **Priority** | P2 |
| **Impact** | +10-15% quality improvement for CBSE Science domain (per `audit_embedding_models.md`). |

### 17. Upgrade to Higher-MTEB Embedding Models

| Field | Details |
|-------|---------|
| **Source** | `audit_embedding_models.md` |
| **Target** | `orchestrator/embed/` + `ingestion/embedder/` |
| **Current State** | Uses `all-MiniLM-L6-v2` (MTEB: 56.3, rank 8 "fast prototyping") for queries. Uses Vertex AI `text-embedding-005` for documents. Both below top-tier models (Gemini 68.3, Qwen3 70.58, Cohere v4 65.2). |
| **Recommended Fix** | Migrate query embeddings to `text-embedding-004` or `text-embedding-005` (same as documents) for consistency and quality. Note: requires re-embedding all documents or dual-index migration. |
| **Priority** | P2 |
| **Impact** | +5-10% retrieval quality improvement. Better semantic understanding. |

### 18. golangci-lint Configuration and Enforcement

| Field | Details |
|-------|---------|
| **Source** | `GO_PERFORMANCE_TESTING_REVIEW.md` (C3) |
| **Target** | Root `.golangci.yml` |
| **Current State** | No `.golangci.yml` found. No automated static analysis. |
| **Recommended Fix** | Create `.golangci.yml` enabling: `errcheck`, `govet`, `staticcheck`, `goimports`, `revive`, `gosec`, `ineffassign`, `misspell`, `unconvert`. Add `make lint` target. Integrate into CI. |
| **Priority** | P2 |
| **Impact** | Catches bugs before they reach production. Enforces code quality standards. |

### 19. Benchmark Functions for Hot Paths

| Field | Details |
|-------|---------|
| **Source** | `GO_PERFORMANCE_TESTING_REVIEW.md` (R1) |
| **Target** | `*_test.go` files in `orchestrator/`, `shared/go/` |
| **Current State** | No `Benchmark*` functions found anywhere in the codebase. |
| **Recommended Fix** | Add benchmarks for: `CosineSimilarity` (768-dim vectors), `ReciprocalRankFuse`, `CalculateRecallAtK`, circuit breaker `AllowRequest`/`RecordFailure`, cache key builders. Add `make bench` target with `-benchmem`. |
| **Priority** | P2 |
| **Impact** | Quantifiable performance baselines. Prevent regression during refactoring. |

### 20. Race Detector Integration

| Field | Details |
|-------|---------|
| **Source** | `GO_PERFORMANCE_TESTING_REVIEW.md` (R3) |
| **Target** | `Makefile` + CI pipeline |
| **Current State** | No evidence of race detector usage in Makefile, CI, or documentation. |
| **Recommended Fix** | Add `make test-race: go test -race ./...` target. Run race detector on all concurrent code paths (circuit breaker, cache layers, goroutine-heavy retrieval). |
| **Priority** | P2 |
| **Impact** | Detect data races and goroutine leaks before production. |

### 21. goroutine Leak Detection (goleak)

| Field | Details |
|-------|---------|
| **Source** | `GO_PERFORMANCE_TESTING_REVIEW.md` (R2) |
| **Target** | `tests/integration_test.go` `TestMain` |
| **Current State** | No `TestMain` with `goleak.VerifyTestMain(m)`. Integration test `TestMain` only prints messages. |
| **Recommended Fix** | Add `go.uber.org/goleak` dependency. Wrap `TestMain` with `goleak.VerifyTestMain(m)`. |
| **Priority** | P2 |
| **Impact** | Detect goroutine leaks that cause memory growth over time. |

### 22. JSON Response Safety (fmt.Fprintf → json.Marshal)

| Field | Details |
|-------|---------|
| **Source** | `GO_PERFORMANCE_TESTING_REVIEW.md` (R10) |
| **Target** | `services/api-gateway/main.go:239`, `services/vector-search-service/main.go:130`, `services/embedding-service/main.go:173` |
| **Current State** | Uses `fmt.Fprintf(w, {"version":"%s"...}, Version, ...)` — produces invalid JSON if version strings contain special characters. |
| **Recommended Fix** | Replace with `json.NewEncoder(w).Encode(map[string]string{"version": Version, ...})`. |
| **Priority** | P2 |
| **Impact** | Prevents invalid JSON responses. Security hardening. |

### 23. Typed Context Keys

| Field | Details |
|-------|---------|
| **Source** | `GO_PERFORMANCE_TESTING_REVIEW.md` (H6) |
| **Target** | `services/vector-search-service/main.go`, `services/embedding-service/main.go`, `services/query-understanding-service/main.go` |
| **Current State** | All use string `"request_id"` as context key. Any package using same string key would collide. |
| **Recommended Fix** | Define `type contextKey string; const contextKeyRequestID contextKey = "request_id"`. Use typed key in `context.WithValue()` and retrieval. |
| **Priority** | P2 |
| **Impact** | Prevents context key collisions. Type-safe context values. |

---

## P3 — Low Priority (Consider for Future)

### 24. Semantic Chunking as Alternative Strategy

| Field | Details |
|-------|---------|
| **Source** | `audit_semantic_chunking.md` |
| **Target** | `ingestion/chunker/semantic_chunker.py` (new) |
| **Current State** | Only character-based recursive splitting. No semantic similarity-based chunking. |
| **Recommended Fix** | Add `SemanticChunker` using sentence embedding similarity with breakpoint threshold. Run A/B test against fixed-size chunking on CBSE Science content to validate if +15-30% retrieval quality justifies 100x compute cost. |
| **Priority** | P3 |
| **Impact** | Potential +15-30% retrieval improvement for narrative content. High compute cost. |

### 25. HyDE (Hypothetical Document Embeddings)

| Field | Details |
|-------|---------|
| **Source** | `audit_query_expansion.md` |
| **Target** | `orchestrator/retrieval/hyde.go` (new) |
| **Current State** | No HyDE implementation. |
| **Recommended Fix** | Generate hypothetical answer with LLM → embed hypothetical answer → search for similar documents. Useful for complex conceptual queries. |
| **Priority** | P3 |
| **Impact** | +10-20% recall for complex/conceptual questions. Adds LLM call cost and latency. |

### 26. Step-Back Prompting

| Field | Details |
|-------|---------|
| **Source** | `audit_query_expansion.md` |
| **Target** | `orchestrator/llm/stepback.go` (new) |
| **Current State** | No step-back prompting found. |
| **Recommended Fix** | Generate broader question from specific query → search both specific and broad queries → combine results. |
| **Priority** | P3 |
| **Impact** | Better recall for queries that need foundational context. |

### 27. Sub-Query Decomposition

| Field | Details |
|-------|---------|
| **Source** | `audit_query_expansion.md` |
| **Target** | `orchestrator/retrieval/decomposition.go` (new) |
| **Current State** | No query decomposition found. |
| **Recommended Fix** | LLM breaks complex queries into 2-3 sub-questions → retrieve for each → deduplicate and merge results. |
| **Priority** | P3 |
| **Impact** | Better recall for multi-part questions. |

### 28. Cascading Reranking (Multi-Tier)

| Field | Details |
|-------|---------|
| **Source** | `audit_reranking.md`, `guide_reranking.md` |
| **Target** | `orchestrator/reranker/cascading_reranker.go` (new) |
| **Current State** | Only one ranking mechanism: RRF fusion. No cascading with multiple quality tiers. |
| **Recommended Fix** | Stage 1: Fast retrieval (100 candidates) → Stage 2: TinyBERT rerank (top 20) → Stage 3: MiniLM rerank (top 5). Balance cost and quality. |
| **Priority** | P3 |
| **Impact** | Best retrieval quality with cost control. Adds 100-300ms latency. |

### 29. Query-Adaptive Reranking

| Field | Details |
|-------|---------|
| **Source** | `audit_reranking.md` |
| **Target** | `orchestrator/retrieval/query_classifier.go` (new) |
| **Current State** | No query-type classification for reranking. |
| **Recommended Fix** | Classify queries as factual/semantic/complex → use different rerankers per type (keyword signals for factual, cross-encoder for semantic, LLM for complex). |
| **Priority** | P3 |
| **Impact** | Optimal reranking per query type. Better quality/cost balance. |

### 30. Token-Based Chunking (vs Character-Based)

| Field | Details |
|-------|---------|
| **Source** | `audit_chunking_strategies.md` |
| **Target** | `ingestion/chunker/token_chunker.py` (new) |
| **Current State** | Uses character-based splitting (400 chars, 150 overlap). Less precise than token-based for LLM contexts. |
| **Recommended Fix** | Add token-based chunker using `tiktoken` or `tokenizers` library. Split at 500-token boundaries with 100-token overlap. Compare quality against character-based. |
| **Priority** | P3 |
| **Impact** | More precise chunk sizes for LLM context windows. Better alignment with model tokenization. |

### 31. Code/Formula-Specific Chunking Strategies

| Field | Details |
|-------|---------|
| **Source** | `audit_chunking_strategies.md` |
| **Target** | `ingestion/chunker/special_content.py` (new) |
| **Current State** | Only prose vs table differentiation. No syntax-aware code splitting, no formula-specific chunking. |
| **Recommended Fix** | Detect code blocks and formulas during parsing → apply syntax-aware splitting that preserves完整性 (e.g., don't split mid-function, keep formulas atomic). |
| **Priority** | P3 |
| **Impact** | Better retrieval for technical content. Preserves semantic completeness of code/formulas. |

### 32. Embedding Migration Tooling

| Field | Details |
|-------|---------|
| **Source** | `audit_embedding_models.md` |
| **Target** | `scripts/migrate_embeddings.py` (new) |
| **Current State** | No embedding migration tooling. No hybrid search between old and new embedding indices. |
| **Recommended Fix** | Build migration tool that: re-embeds documents with new model in batches, validates quality on golden dataset, supports dual-index gradual migration with rollback. |
| **Priority** | P3 |
| **Impact** | Safe path to upgrade embedding models without service disruption. |

### 33. Backup and Recovery Automation

| Field | Details |
|-------|---------|
| **Source** | `audit_vector_databases.md` |
| **Target** | `scripts/backup_db.py`, `scripts/restore_db.py` |
| **Current State** | No explicit backup/recovery implementation. Relies on AlloyDB's native backup capabilities. |
| **Recommended Fix** | Create automated backup scripts with pg_dump, scheduled via cron or Cloud Scheduler. Implement restore procedure with validation. |
| **Priority** | P3 |
| **Impact** | Disaster recovery capability. Compliance requirement. |

---

## Summary Statistics

| Priority | Count | Category Focus |
|----------|-------|----------------|
| **P0 Critical** | 4 | Reranker integration, parallel search, streaming, rate limiting |
| **P1 High** | 8 | HNSW tuning, cost tracking, alerting, dashboards, semantic cache, A/B testing, query expansion, live metrics |
| **P2 Medium** | 10 | Contextual compression, MMR, multi-query, fine-tuning, model upgrade, linting, benchmarks, race detection, goleak, JSON safety, typed context |
| **P3 Low** | 10 | Semantic chunking, HyDE, step-back, decomposition, cascading reranking, adaptive reranking, token chunking, special content, migration, backup |
| **Total** | **32** | |

---

## Quick Wins (Highest Impact per Effort)

| # | Optimization | Effort | Impact | Why |
|---|-------------|--------|--------|-----|
| 1 | **#1 Wire reranker into pipeline** | Low (code exists, just wiring) | High (+10-25% accuracy) | Code already in `orchestrator/reranker/`, not connected |
| 2 | **#2 Parallel search** | Low (go routine refactor) | Medium (30-50% retrieval speedup) | Simple concurrency change |
| 3 | **#4 Rate limiting** | Medium | High (cost control) | Redis-based sliding window |
| 4 | **#5 HNSW ef_search** | Very Low (one SQL line) | Medium (10x search speedup) | `SET LOCAL hnsw.ef_search` |
| 5 | **#11 Query expansion** | Medium | High (+30-50% recall) | LLM generates variations, parallel search |

## Largest Gaps

| # | Gap | Why It Matters |
|---|-----|----------------|
| 1 | **No reranker in active pipeline** | Single largest accuracy gap. Reranking is "easiest way to gain 10-25% accuracy" |
| 2 | **No rate limiting in orchestrator** | Cost explosion risk, no abuse prevention |
| 3 | **No streaming end-to-end** | TTFT ~350ms instead of ~150ms, poor user experience |
| 4 | **No cost tracking** | Flying blind on spending |
| 5 | **No alerting/dashboards** | No visibility into pipeline health |
