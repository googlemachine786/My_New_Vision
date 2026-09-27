# Audit: Advanced Retrieval Strategies Guide vs. Implementation

**Guide:** `guide_advanced_retrieval_strategies.md` (Ailog, Feb 2025)
**Audited:** 2026-04-03
**Scope:** C:\Users\kommi\ragpipeline\

---

## Guide Summary

This guide covers advanced retrieval strategies beyond basic semantic similarity search, including hybrid search (BM25 + vector), query expansion, Maximal Marginal Relevance (MMR), parent-child retrieval, ensemble retrieval, self-query retrieval, multi-stage retrieval, and contextual compression.

---

## Key Recommendations and Implementation Match

### 1. Hybrid Search (BM25 + Vector)

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Combine semantic (vector) and lexical (keyword/BM25) search | |
| **Implementation** | **MATCH** | Go implementation in `C:\Users\kommi\ragpipeline\orchestrator\retrieval\hybrid_search.go` performs dense search (pgvector cosine similarity) + sparse search (GIN keyword intersection) with RRF fusion. |
| **Code Reference** | `hybrid_search.go` lines 43-76: `HybridSearcher.Search()` runs both `denseSearch()` and `sparseSearch()`, then calls `rrfFuse()` with k=60. |
| **RRF Implementation** | **MATCH** | `C:\Users\kommi\ragpipeline\orchestrator\retrieval\rrf.go` implements Reciprocal Rank Fusion with standard k=60 parameter. |
| **Keyword Extraction** | **MATCH** | `hybrid_search.go` lines 188-223: `ExtractKeywords()` performs stopword-filtered unigram+bigram extraction, limited to 8 keywords. |

### 2. Query Expansion

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Generate multiple query variations, HyDE, query decomposition | |
| **Implementation** | **PARTIALLY MATCHES** | The `ClarificationGenerator` in `C:\Users\kommi\ragpipeline\orchestrator\llm\clarification_generator.go` performs query ambiguity analysis and suggests refinements, but does NOT implement multi-query retrieval, HyDE, or sub-query decomposition. |
| **Missing** | No HyDE (Hypothetical Document Embeddings), no multi-query generation for parallel retrieval, no query decomposition into sub-queries. |

### 3. Maximal Marginal Relevance (MMR)

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Reduce redundancy by balancing relevance vs. diversity | |
| **Implementation** | **DOES NOT MATCH** | No MMR implementation found anywhere in the codebase. No lambda parameter for relevance/diversity tradeoff. |
| **Missing** | No `mmr()` function, no diversity scoring in retrieval pipeline. |

### 4. Parent-Child Retrieval

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Retrieve small child chunks for precision, return parent for context | |
| **Implementation** | **MATCH** | `C:\Users\kommi\ragpipeline\ingestion\chunker\parent_child.py` implements parent-child chunking with optimized parameters: 400-char chunks, 150-char overlap (37.5%). Tables are atomic (parent_id == child_id). |
| **Code Reference** | `parent_child.py` lines 150-210: `create_parent_child_chunks()` creates both parent and child chunks with proper mapping. `hybrid_search.go` joins `child_chunks` to `parent_chunks` to return parent content alongside child results. |

### 5. Ensemble Retrieval

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Combine multiple retrievers with weighted scores | |
| **Implementation** | **PARTIALLY MATCHES** | The RRF fusion in `hybrid_search.go` combines dense and sparse retrievers, which is a form of ensemble. However, there is no configurable weighting system (alpha parameter) as described in the guide's `EnsembleRetriever` class. |

### 6. Self-Query Retrieval

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Extract structured filters from natural language queries | |
| **Implementation** | **MATCH** | `C:\Users\kommi\ragpipeline\retrievers\self_query_retriever.py` and `self_query_parser.py` implement LLM-based filter extraction using Ollama (llama3.2:3b). Extracts grade, chapter_number, chapter_name, content_type, subject, and page range filters. |
| **Code Reference** | `self_query_retriever.py` lines 60-120: `SelfQueryRetriever.search()` pipeline: extract filters -> build SQL -> embed -> execute filtered vector search -> return results. |

### 7. Multi-Stage Retrieval (Retrieval + Reranking)

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Fast retrieval overfetch, then accurate reranking | |
| **Implementation** | **PARTIALLY MATCHES** | The hybrid search retrieves `topK*20` candidates before RRF fusion and dedup to top-K. However, there is NO cross-encoder reranker (e.g., ms-marco-MiniLM) or LLM-based reranking stage. |
| **Missing** | No dedicated reranker model. No cascading reranking stages. |

### 8. Contextual Compression

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Use LLM to extract only relevant parts from retrieved chunks | |
| **Implementation** | **DOES NOT MATCH** | No contextual compression implementation found. Full chunk content is passed to the LLM without compression. |

### 9. Retrieval Strategy Decision Framework

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Start simple, add hybrid if specific terms, add query expansion if ambiguous | |
| **Implementation** | **PARTIALLY MATCHES** | `C:\Users\kommi\ragpipeline\rag_config.py` provides presets: `low_latency`, `high_quality`, `balanced`. Supports `retrieval_strategy` field with values "dense", "hybrid", "bm25". No automated decision based on query analysis. |

---

## Gap Analysis Summary

### Implemented
- Hybrid search (dense + sparse with RRF fusion)
- Parent-child chunking and retrieval
- Self-query retrieval with LLM filter extraction
- Query ambiguity analysis (partial query expansion)
- Configurable retrieval strategies via RAGConfig

### Partially Implemented
- Query expansion (only ambiguity analysis, no HyDE/multi-query/decomposition)
- Ensemble retrieval (RRF fusion exists, but no weighted retriever ensemble)
- Multi-stage retrieval (overfetch exists, but no reranker)

### NOT Implemented
- Maximal Marginal Relevance (MMR)
- Contextual compression
- HyDE (Hypothetical Document Embeddings)
- Query decomposition into sub-queries
- Cascading reranking with multiple quality tiers
- Query-adaptive reranking (different rerankers per query type)

---

## Overall Assessment: PARTIALLY MATCHES

The project implements hybrid search, parent-child retrieval, and self-query retrieval well. The hybrid search with RRF fusion is a production-grade implementation. However, it lacks MMR, contextual compression, HyDE, query decomposition, and any form of cross-encoder reranking, which are all significant recommendations in this guide.
