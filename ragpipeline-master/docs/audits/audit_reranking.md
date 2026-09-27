# Audit: Reranking Guide vs. Implementation

**Guide:** `guide_reranking.md` (Ailog, Feb 2025)
**Audited:** 2026-04-03
**Scope:** C:\Users\kommi\ragpipeline\

---

## Guide Summary

This guide covers reranking as a second-pass scoring mechanism for retrieved documents, focusing on cross-encoders (ms-marco-MiniLM, TinyBERT), Cohere Rerank API, LLM-based reranking, FlashRank, and hybrid reranking. It recommends a two-stage retrieval pipeline: fast retrieval (high recall) followed by accurate reranking (high precision).

---

## Key Recommendations and Implementation Match

### 1. Two-Stage Retrieval Architecture

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Query -> [Stage 1: Retrieval] -> 100 candidates -> [Stage 2: Reranking] -> 10 best | |
| **Implementation** | **PARTIALLY MATCHES** | The hybrid search in `C:\Users\kommi\ragpipeline\orchestrator\retrieval\hybrid_search.go` line 54-55 retrieves `topK*20` candidates before RRF fusion. However, there is NO dedicated cross-encoder or LLM reranking stage. RRF fusion serves as the "reranking" mechanism. |
| **Code Reference** | `hybrid_search.go` lines 48-55: `denseSearch(ctx, embedding, taxonomyID, topK*20)` and `sparseSearch(ctx, keywords, taxonomyID, topK*20)`. |

### 2. Cross-Encoder Models

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Use cross-encoders like `ms-marco-MiniLM-L6-v2` for query-document pair scoring | |
| **Implementation** | **DOES NOT MATCH** | No cross-encoder model is used anywhere in the codebase. No `sentence_transformers.CrossEncoder` import or usage found. |
| **Missing** | No cross-encoder model loading, no `model.predict()` for query-document pairs, no reranking by cross-encoder scores. |

### 3. Cohere Rerank API

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Managed reranking service ($1/1000 queries) | |
| **Implementation** | **DOES NOT MATCH** | No Cohere API integration found. |
| **Missing** | No `cohere` client, no `co.rerank()` calls. |

### 4. LLM-Based Reranking

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Use LLM to judge relevance (binary, scoring, or comparative) | |
| **Implementation** | **DOES NOT MATCH** | While the `LLMJudge` in `C:\Users\kommi\ragpipeline\orchestrator\quality\judge.go` evaluates negative feedback, it does NOT rerank retrieved candidates. It generates golden responses for feedback loops, not retrieval reranking. |
| **Missing** | No LLM reranking of retrieval candidates, no relevance scoring per document, no comparative ranking. |

### 5. FlashRank

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Efficient local reranking with FlashRank | |
| **Implementation** | **DOES NOT MATCH** | No FlashRank usage found in the codebase. |

### 6. Hybrid Reranking

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Combine multiple signals (vector_score + bm25_score + cross_encoder_score) with weights | |
| **Implementation** | **PARTIALLY MATCHES** | RRF fusion combines dense and sparse rankings, which is a form of hybrid scoring. However, there is no cross-encoder score component and no configurable weighting. |
| **Code Reference** | `rrf.go` lines 17-49: `rrfFuse()` combines dense and sparse scores via `1/(k + rank)`. No third signal (cross-encoder) is added. |

### 7. Top-K Reranking (Overfetch)

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Retrieve 3-5x the final k for reranking | |
| **Implementation** | **MATCH** | The hybrid search retrieves `topK*20` candidates from both dense and sparse search before RRF fusion. For top_k=5, this means 100 candidates are retrieved. |
| **Code Reference** | `hybrid_search.go` lines 48, 55: `topK*20` limit on both searches. |

### 8. Cascading Reranking

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Multiple reranking stages: fast reranker -> accurate reranker | |
| **Implementation** | **DOES NOT MATCH** | Only one reranking mechanism exists: RRF fusion. No cascading with multiple quality tiers. |

### 9. Query-Adaptive Reranking

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Different reranking strategy based on query type (factual, semantic, complex) | |
| **Implementation** | **DOES NOT MATCH** | No query-type classification for reranking. The `ClarificationGenerator` analyzes ambiguity but does not route to different rerankers. |

### 10. Performance Optimization

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Batch reranking, caching, async operations | |
| **Implementation** | **PARTIALLY MATCHES** | The project has response caching (`C:\Users\kommi\ragpipeline\orchestrator\cache\response_cache.go`) but it caches full RAG responses, not reranking scores. No batch reranking since there is no reranker. |

### 11. Evaluation Metrics

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Precision@k, NDCG@k, MRR, A/B testing | |
| **Implementation** | **MATCH** | `C:\Users\kommi\ragpipeline\orchestrator\retrieval\rrf.go` lines 57-111 implement `CalculateRecallAtK`, `CalculateMRR`, and `CalculateNDCG`. `bench_and_improve.py` computes keyword_coverage, faithfulness, recall_at_5, ndcg_at_5. |
| **Code Reference** | `rrf.go` lines 57-75: `CalculateRecallAtK`, lines 78-94: `CalculateMRR`, lines 97-121: `CalculateNDCG`. |

### 12. Cost-Benefit Analysis

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Measure reranking latency impact (50-500ms) vs quality gain | |
| **Implementation** | **DOES NOT MATCH** | No reranking latency tracking since there is no separate reranking stage. The observability module tracks total request latency but not per-stage breakdown for reranking. |

---

## Gap Analysis Summary

### Implemented
- Overfetch before fusion (topK*20 candidates)
- RRF fusion as a ranking mechanism
- Evaluation metrics (Recall@K, MRR, NDCG)
- Response caching (but not reranking-specific)

### Partially Implemented
- Two-stage retrieval (overfetch + RRF, but no cross-encoder reranker)
- Hybrid scoring (dense + sparse RRF, but no cross-encoder component)

### NOT Implemented
- Cross-encoder reranking (ms-marco-MiniLM, TinyBERT, etc.)
- Cohere Rerank API integration
- LLM-based reranking of retrieval candidates
- FlashRank local reranking
- Cascading reranking (multi-tier)
- Query-adaptive reranking (different rerankers per query type)
- Reranking-specific caching
- Reranking latency/cost tracking

---

## Overall Assessment: DOES NOT MATCH

This is the largest gap in the project. The guide strongly recommends reranking as "the easiest way to gain another 10-25% accuracy" and the "optimization with the best effort-to-impact ratio." The project relies solely on RRF fusion for ranking, which is a lightweight ranking method but NOT a true reranker. There is no cross-encoder model, no LLM reranking, no Cohere integration, and no cascading reranking. Adding a cross-encoder reranker (even a small one like TinyBERT) would be the highest-impact single improvement to this pipeline.
