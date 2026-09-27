# Audit: Semantic Chunking Guide vs. Implementation

**Guide:** `guide_semantic_chunking.md` (Ailog, Nov 2025)
**Audited:** 2026-04-03
**Scope:** C:\Users\kommi\ragpipeline\

---

## Guide Summary

This guide focuses specifically on semantic chunking: splitting documents based on meaning rather than length. It covers embedding consecutive sentences, calculating cosine similarity between them, and splitting where similarity drops (topic change). It also covers LangChain's SemanticChunker, multi-level semantic chunking, hybrid approaches, and performance considerations.

---

## Key Recommendations and Implementation Match

### 1. Semantic Similarity-Based Splitting

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Embed each sentence, calculate similarity between consecutive sentences, split where similarity drops below threshold | |
| **Implementation** | **DOES NOT MATCH** | No semantic similarity-based chunking found. The project uses `RecursiveCharacterTextSplitter` (character-based) rather than sentence embedding similarity. |
| **Missing** | No sentence embedding, no cosine similarity between consecutive sentences, no `similarity_threshold` for split detection. |

### 2. LangChain SemanticChunker

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Use `SemanticChunker` from langchain with breakpoint_threshold_type ("percentile" or "standard_deviation") | |
| **Implementation** | **DOES NOT MATCH** | No LangChain SemanticChunker usage. |
| **Missing** | No `from langchain.text_splitter import SemanticChunker`. |

### 3. Multi-Level Semantic Chunking

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | First split semantically, then split large chunks by paragraphs, merge small chunks | |
| **Implementation** | **PARTIALLY MATCHES** | The project does have a two-level approach: first split prose vs tables, then apply recursive character splitting. However, this is structural, not semantic. |
| **Code Reference** | `parent_child.py` lines 175-180: prose vs table grouping. No semantic similarity-based initial split. |

### 4. Hybrid Approach

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Semantic split first, then merge small chunks and split large ones to target size | |
| **Implementation** | **DOES NOT MATCH** | No semantic-first splitting. The project uses recursive character splitting directly. |
| **Missing** | No semantic pre-splitting before size-based refinement. |

### 5. LlamaIndex Semantic Splitter

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Use `SemanticSplitterNodeParser` with buffer_size and breakpoint_percentile_threshold | |
| **Implementation** | **DOES NOT MATCH** | No LlamaIndex usage. |

### 6. When to Use Semantic Chunking

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Use when: documents have clear topic transitions, high precision needed, narrative content. Use fixed-size when: speed critical, uniform documents, budget limited. | |
| **Implementation** | **PARTIALLY MATCHES** | The project uses fixed-size chunking, which the guide says is appropriate when "speed is critical" and "budget is limited." The choice is justified by empirical grid search evaluation showing 400-char chunks achieve 0.8386 composite score. |
| **Code Reference** | `rag_config.py` comments reference grid search validation. |

### 7. Performance Considerations

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Semantic chunking costs 100-500ms per document vs ~1ms for fixed-size. Requires embedding every sentence. | |
| **Implementation** | **N/A** | Since semantic chunking is not used, this is not applicable. The project's character-based chunking is fast (~1ms per document). |

### 8. Evaluation of Semantic vs Fixed

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Test retrieval quality (MRR) with semantic vs fixed chunking on test queries | |
| **Implementation** | **PARTIALLY MATCHES** | The project evaluates chunking quality extensively with `bench_and_improve.py`, but only tests different configurations of the fixed-size approach (chunk_size, overlap, top_k). Does not compare against semantic chunking as a baseline. |
| **Code Reference** | `bench_and_improve.py` lines 330-355: `CONFIGS` array with different chunk_size and overlap values. No semantic chunking config. |

---

## Gap Analysis Summary

### Implemented
- Fast character-based chunking (alternative approach recommended by guide)
- Parent-child hierarchical chunking
- Extensive evaluation of chunking configurations

### Partially Implemented
- Evaluation framework (tests configurations but not semantic chunking)

### NOT Implemented
- Sentence embedding for semantic similarity calculation
- Similarity threshold-based split detection
- LangChain SemanticChunker
- LlamaIndex SemanticSplitterNodeParser
- Multi-level semantic chunking
- Hybrid semantic + size-based chunking
- Breakpoint percentile/standard deviation threshold methods
- Comparison benchmark: semantic vs fixed-size chunking

---

## Overall Assessment: DOES NOT MATCH

This is a clear gap. The project uses exclusively character-based (recursive) chunking with no semantic awareness. The guide recommends semantic chunking for documents with "clear topic transitions" and "high precision retrieval," which could benefit CBSE Science textbook content. The guide notes semantic chunking "typically improves retrieval by 15-30% but costs 100x more compute." The project has chosen speed over this potential quality gain. Adding semantic chunking as an alternative strategy (alongside the existing fixed-size approach) would allow A/B testing to determine if the quality gain justifies the compute cost for this specific domain.
