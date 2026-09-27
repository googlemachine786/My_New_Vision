# Audit: Retrieval Fundamentals Guide vs. Implementation

**Guide:** `guide_retrieval_fundamentals.md` (Ailog, Jan 2026)
**Audited:** 2026-04-03
**Scope:** C:\Users\kommi\ragpipeline\

---

## Guide Summary

This guide covers the three pillars of retrieval: representation (embeddings), indexing (vector databases), and search (similarity metrics). It covers embedding model selection, chunking strategies, vector database usage (Qdrant), similarity metrics (cosine, dot product, Euclidean), query expansion, reranking, metadata filtering, and evaluation metrics (Recall@K, MRR, NDCG).

---

## Key Recommendations and Implementation Match

### 1. Embeddings as Vector Representations

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Transform text into mathematical vectors (384-1536 dimensions) | |
| **Implementation** | **MATCH** | Uses `sentence-transformers/all-MiniLM-L6-v2` (384 dimensions) in `rag_config.py` line 54. Ingestion uses Vertex AI `text-embedding-005` (768 dimensions) in `C:\Users\kommi\ragpipeline\ingestion\embedder\vertex_batch.py`. |
| **Code Reference** | `rag_config.py` line 54: `embed_model: str = "sentence-transformers/all-MiniLM-L6-v2"`, `embed_dimension: int = 384`. |

### 2. Embedding Model Selection

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Choose based on use case: all-MiniLM for prototyping, all-mpnet for production, text-embedding-3-large for high precision | |
| **Implementation** | **MATCH** | Uses `all-MiniLM-L6-v2` for query embedding (fast, good for prototyping). Uses Vertex AI `text-embedding-005` for document embeddings during ingestion (production quality). |
| **Code Reference** | `rag_config.py` line 54, `vertex_batch.py` line 13: `model: str = "text-embedding-005"`. |

### 3. Chunking Strategies

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Fixed-size, semantic, sentence-based, hierarchical | |
| **Implementation** | **MATCH** | Implements parent-child hierarchical chunking with recursive character splitting. See `audit_chunking_strategies.md` for full details. |
| **Code Reference** | `C:\Users\kommi\ragpipeline\ingestion\chunker\parent_child.py`. |

### 4. Vector Database Usage

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Store and index embeddings for fast search | |
| **Implementation** | **MATCH** | Uses pgvector (PostgreSQL extension) on AlloyDB. Child chunks stored with embeddings, parent chunks store metadata. |
| **Code Reference** | `C:\Users\kommi\ragpipeline\orchestrator\retrieval\hybrid_search.go` lines 95-135: pgvector cosine similarity search using `<=>` operator. |

### 5. Similarity Metrics

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Cosine similarity (most common), dot product (faster for normalized), Euclidean distance | |
| **Implementation** | **MATCH** | Uses cosine distance (`<=>` operator in pgvector) for dense search. Embeddings are normalized during encoding. |
| **Code Reference** | `hybrid_search.go` line 103: `cc.embedding <=> $1 AS cosine_dist`. `self_query_retriever.py` line 122: `normalize_embeddings=True`. |

### 6. Query Expansion

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Enrich user query to improve recall | |
| **Implementation** | **PARTIALLY MATCHES** | The `ClarificationGenerator` in `C:\Users\kommi\ragpipeline\orchestrator\llm\clarification_generator.go` analyzes query ambiguity and suggests refinements. Self-query parser extracts filters from queries. However, no multi-query expansion or HyDE. |
| **Code Reference** | `clarification_generator.go` lines 51-80: `AnalyzeQuery()` for ambiguity detection. |

### 7. Reranking

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Use cross-encoder to refine initial results | |
| **Implementation** | **DOES NOT MATCH** | No cross-encoder reranker. RRF fusion serves as the ranking mechanism. See `audit_reranking.md` for full details. |

### 8. Metadata Filtering

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Combine vector search with classic filters (category, date, etc.) | |
| **Implementation** | **MATCH** | Self-query retrieval extracts metadata filters (grade, chapter, content_type) and applies them as SQL WHERE clauses before vector search. Hybrid search filters by `taxonomy_id`. |
| **Code Reference** | `hybrid_search.go` line 108: `WHERE pc.taxonomy_id = $2`. `self_query_retriever.py` lines 155-180: SQL WHERE clause generation with metadata filters. |

### 9. Evaluation: Recall@K

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Proportion of relevant documents found in top-k | |
| **Implementation** | **MATCH** | Implemented in both Go and Python. |
| **Code Reference** | `C:\Users\kommi\ragpipeline\orchestrator\retrieval\rrf.go` lines 57-75: `CalculateRecallAtK`. `bench_and_improve.py` lines 214-220: `recall_at_k`. |

### 10. Evaluation: MRR

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Average position of first relevant document | |
| **Implementation** | **MATCH** | Implemented in Go. |
| **Code Reference** | `rrf.go` lines 78-94: `CalculateMRR`. |

### 11. Evaluation: NDCG

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Takes into account result order and relevance scores | |
| **Implementation** | **MATCH** | Implemented in both Go and Python. |
| **Code Reference** | `rrf.go` lines 97-121: `CalculateNDCG`. `bench_and_improve.py` lines 196-211: `ndcg_at_k`. |

### 12. Common Pitfalls

| Pitfall | Status | Details |
|---------|--------|---------|
| Chunks too large | **N/A** | Uses 400-char chunks (well within range) |
| Domain vocabulary | **PARTIALLY** | Keyword extraction with YAKE helps, but no synonym expansion |
| Ambiguous queries | **MATCH** | `ClarificationGenerator` detects ambiguity and suggests refinements |
| Cold start | **DOES NOT MATCH** | No synthetic data enrichment or FAQ generation for cold start |

### 13. Production Architecture

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | API Gateway -> Query Processor -> Dense Search + Sparse Search -> Fusion/Rerank -> LLM Context | |
| **Implementation** | **MATCH** | The project architecture follows this pattern: Query -> Clarification/Self-Query -> Hybrid Search (Dense + Sparse) -> RRF Fusion -> LLM Generation. |
| **Code Reference** | Architecture spans: `clarification_generator.go` (query processing), `hybrid_search.go` (dense+sparse), `rrf.go` (fusion). |

---

## Gap Analysis Summary

### Implemented
- Embedding models (all-MiniLM for queries, Vertex AI for documents)
- pgvector-based vector storage with cosine similarity search
- Hierarchical chunking with metadata preservation
- Metadata filtering (taxonomy_id, grade, chapter, content_type)
- Evaluation metrics (Recall@K, MRR, NDCG)
- Production architecture following the recommended pattern
- Query ambiguity detection

### Partially Implemented
- Query expansion (ambiguity analysis exists, but no multi-query expansion)
- Domain vocabulary (YAKE keyword extraction, but no synonym expansion)

### NOT Implemented
- Cross-encoder reranking
- Cold start mitigation (synthetic data/FAQs)
- HyDE retrieval
- Synonym-based query expansion

---

## Overall Assessment: MATCH

The project aligns well with retrieval fundamentals. It implements embedding-based search, vector storage, proper chunking, metadata filtering, and all three key evaluation metrics. The production architecture follows the recommended pattern. The main gaps are in reranking (covered separately) and advanced query expansion techniques.
