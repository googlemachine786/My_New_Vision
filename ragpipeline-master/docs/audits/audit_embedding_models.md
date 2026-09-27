# Audit: Embedding Models Guide vs. Implementation

**Guide:** `guide_embedding_models.md` (Ailog, Jan 2026)
**Audited:** 2026-04-03
**Scope:** C:\Users\kommi\ragpipeline\

---

## Guide Summary

This guide covers the embedding model landscape in 2025-2026, with MTEB leaderboard scores, model comparisons (Gemini, Qwen3, Voyage, Cohere, OpenAI, BGE-M3), dimension size trade-offs, language support, domain specialization, fine-tuning, Matryoshka embeddings, cost analysis, and migration strategies.

---

## Key Recommendations and Implementation Match

### 1. Model Selection Based on MTEB Scores

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Top models: Gemini-embedding-001 (68.3), Qwen3-8B (70.58), Voyage-3-large (66.8), Cohere v4 (65.2), OpenAI text-3-large (64.6), BGE-M3 (63.0) | |
| **Implementation** | **PARTIALLY MATCHES** | Uses `all-MiniLM-L6-v2` (MTEB: 56.3) for query embeddings and Vertex AI `text-embedding-005` for document embeddings. Both are below the top-tier models recommended by the guide. |
| **Code Reference** | `rag_config.py` line 54: `embed_model: str = "sentence-transformers/all-MiniLM-L6-v2"`. `vertex_batch.py` line 13: `model: str = "text-embedding-005"`. |
| **Note** | `all-MiniLM-L6-v2` is listed in the guide as "fast prototyping" (rank 8), not production quality. |

### 2. Accuracy vs. Cost Trade-off

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | High accuracy: OpenAI or Cohere. Budget: Open-source (BGE-M3, all-MiniLM) | |
| **Implementation** | **MATCH** | Uses free `all-MiniLM-L6-v2` for queries (cost-conscious) and Vertex AI for documents (production quality). This is a reasonable cost-quality trade-off. |
| **Code Reference** | `rag_config.py` line 54, `vertex_batch.py` line 13. |

### 3. Dimension Size

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Smaller = faster/cheaper but less accurate. Large = better quality but more storage. Configurable dimensions (Matryoshka). | |
| **Implementation** | **MATCH** | 384 dimensions for query model (all-MiniLM), 768 dimensions for document model (text-embedding-005). Moderate sizes balancing speed and quality. |
| **Code Reference** | `rag_config.py` line 55: `embed_dimension: int = 384`. `vertex_batch.py` line 43: `output_dimensionality: int = 768`. |

### 4. Language Support

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Multilingual models: Cohere v4 (100+), BGE-M3 (100+), OpenAI text-3-large (100+) | |
| **Implementation** | **DOES NOT MATCH** | `all-MiniLM-L6-v2` is English-only. `text-embedding-005` supports multiple languages but the project is focused on CBSE Science (English content). For the specific use case, this is acceptable. |
| **Missing** | No multilingual embedding model for queries. If the project needs to support non-English queries, this would be a gap. |

### 5. Domain Specialization

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Fine-tune embeddings for specialized domains (legal, medical, code) | |
| **Implementation** | **DOES NOT MATCH** | No fine-tuning of embedding models. Uses pre-trained models without domain adaptation. |
| **Missing** | No `sentence_transformers` fine-tuning pipeline, no domain-specific training examples, no MultipleNegativesRankingLoss training. |

### 6. Matryoshka Embeddings

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Generate once at full dimension, truncate as needed (OpenAI text-3-*, Nomic, Jina) | |
| **Implementation** | **DOES NOT MATCH** | Neither `all-MiniLM-L6-v2` nor `text-embedding-005` support Matryoshka embeddings. Fixed dimension output. |
| **Missing** | No variable-dimension embedding support. |

### 7. Fine-Tuning

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Fine-tune base model with domain-specific training data for +10-30% gains | |
| **Implementation** | **DOES NOT MATCH** | No fine-tuning pipeline exists. |
| **Missing** | No training examples creation, no `InputExample` datasets, no `MultipleNegativesRankingLoss` training loop. |

### 8. Cost Analysis

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Track embedding costs, storage costs, inference costs per model | |
| **Implementation** | **PARTIALLY MATCHES** | The `observability/metrics.go` tracks costs via Prometheus counters but does not specifically track embedding costs in USD. The cost tracking is generic (`rag_cost_usd` with component labels). |
| **Code Reference** | `guide_cost_optimization.md` references cost tracking. `metrics.go` has generic cost counters. |

### 9. Benchmarking on Own Data

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Test multiple models on YOUR data, don't trust generic benchmarks | |
| **Implementation** | **PARTIALLY MATCHES** | `bench_and_improve.py` evaluates retrieval quality with keyword coverage, faithfulness, recall, NDCG. However, it only tests one embedding model (`all-MiniLM-L6-v2`), not multiple models side-by-side. |
| **Code Reference** | `bench_and_improve.py` line 33: `EMBED_MODEL = "sentence-transformers/all-MiniLM-L6-v2"`. Only one model used across all configurations. |

### 10. Migration Strategy

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Changing embeddings requires re-embedding everything; use hybrid search during migration | |
| **Implementation** | **DOES NOT MATCH** | No embedding migration tooling. No hybrid search between old and new embedding indices. |
| **Missing** | No `hybrid_search()` function that blends results from old and new indices. No gradual migration support. |

### 11. Model Recommendations by Use Case

| Guide Recommendation | Project Status | Match |
|---------------------|----------------|-------|
| Startup/MVP: all-MiniLM-L6-v2 | Uses all-MiniLM-L6-v2 | **MATCH** |
| Production (quality): Cohere v4 or OpenAI text-3-large | Uses Vertex AI text-embedding-005 | **PARTIALLY** (Vertex AI is comparable quality tier) |
| Production (budget): BGE-M3 self-hosted | Not using BGE-M3 | **DOES NOT MATCH** |
| Multilingual: Cohere v4 or BGE-M3 | Not using multilingual models | **DOES NOT MATCH** |
| Code search: Voyage code-2 or OpenAI text-3-small | N/A for this project | **N/A** |
| Privacy-critical: BGE-M3 self-hosted | Not self-hosting for queries | **DOES NOT MATCH** |

---

## Gap Analysis Summary

### Implemented
- Free embedding model for queries (all-MiniLM-L6-v2)
- Production-quality model for documents (Vertex AI text-embedding-005)
- Reasonable dimension sizes (384 for queries, 768 for documents)
- Batch embedding with retry logic
- Evaluation framework for retrieval quality

### Partially Implemented
- Cost tracking (generic, not embedding-specific)
- Benchmarking (single model, not multi-model comparison)

### NOT Implemented
- Fine-tuning for domain-specific improvement
- Matryoshka/variable-dimension embeddings
- Multilingual query embeddings
- Embedding migration tooling
- Multi-model side-by-side benchmarking
- Top-tier MTEB models (Gemini, Qwen3, Cohere v4)

---

## Overall Assessment: PARTIALLY MATCHES

The embedding strategy is practical and cost-conscious, using a fast free model for queries and a production model for documents. However, it does not leverage top-tier MTEB models, lacks domain fine-tuning, has no multi-model benchmarking, and provides no migration path. For the CBSE Science use case, the current models are adequate but leave ~10-15% quality improvement on the table by not using higher-MTEB models.
