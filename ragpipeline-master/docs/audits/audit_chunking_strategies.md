# Audit: RAG Chunking Strategies Guide vs. Implementation

**Guide:** `guide_chunking_strategies.md` (Ailog, Jan 2025)
**Audited:** 2026-04-03
**Scope:** C:\Users\kommi\ragpipeline\

---

## Guide Summary

This guide covers document chunking strategies for RAG: fixed-size chunking (character, token-based), semantic chunking, metadata-aware chunking (Markdown, HTML, code), recursive character splitting, chunk overlap optimization, and evaluation methodologies.

---

## Key Recommendations and Implementation Match

### 1. Fixed-Size Chunking

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | 500-1000 tokens for general docs, 10-20% overlap | |
| **Implementation** | **PARTIALLY MATCHES** | `C:\Users\kommi\ragpipeline\ingestion\chunker\parent_child.py` implements character-based (not token-based) chunking. Uses 400 characters with 150-char overlap (37.5%). |
| **Code Reference** | `parent_child.py` lines 69-77: `RecursiveCharacterTextSplitter.__init__()` with `max_chars=400`, `overlap=150`. |
| **Note** | Character-based splitting is less precise than token-based for LLM contexts, but the guide acknowledges character-based as valid for quick prototypes. |

### 2. Recursive Character Splitting

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Try separators in order: `\n\n`, `\n`, `. `, ` `, `` (LangChain approach) | |
| **Implementation** | **MATCH** | `parent_child.py` lines 75: `separators = ["\n\n", "\n", ". ", " ", ""]` - exactly matches the guide's recommended hierarchy. |
| **Code Reference** | `parent_child.py` lines 88-142: `_split_recursive()` method implements the recursive splitting with separator fallback. |

### 3. Chunk Overlap

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | 10-20% overlap ratio for 512-token chunks (50-100 tokens) | |
| **Implementation** | **MATCH (exceeds)** | Uses 37.5% overlap (150/400 chars), which is higher than the recommended 10-20%. This is based on empirical grid search evaluation. |
| **Code Reference** | `parent_child.py` lines 155-158: `parent_overlap=150`, `child_overlap=150` with 400-char chunks = 37.5%. `rag_config.py` line 52: `chunk_overlap: int = 150`. |
| **Note** | Higher overlap improves recall at the cost of more storage. The guide recommends 10-20% but the project's 37.5% was empirically validated. |

### 4. Parent-Child (Hierarchical) Chunking

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Chunk at multiple granularities: document -> sections -> paragraphs | |
| **Implementation** | **MATCH** | Parent-child chunking with prose splitting and atomic table handling. |
| **Code Reference** | `parent_child.py` lines 150-253: `create_parent_child_chunks()` groups elements by type (prose vs tables), creates parent chunks then child chunks within each parent. |

### 5. Metadata-Aware Chunking

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Preserve headers, sections, structure as metadata | |
| **Implementation** | **MATCH** | ParentChunk and ChildChunk dataclasses include taxonomy_id, page_number, chapter, section, subsection, content_type metadata. |
| **Code Reference** | `parent_child.py` lines 32-55: `ParentChunk` dataclass with metadata fields. Lines 57-76: `ChildChunk` dataclass. |

### 6. Content-Type-Specific Chunking

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Different strategies for different content types (tech docs, code, chat logs) | |
| **Implementation** | **PARTIALLY MATCHES** | The pipeline distinguishes prose from tables (`parent_child.py` lines 175-180). Tables are kept atomic (no splitting). However, no special handling for code, formulas beyond the content_type field. |
| **Missing** | No syntax-aware code splitting, no formula-specific chunking strategy. |

### 7. Chunk Size Recommendations

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | General docs: 512-1024 tokens, Short FAQ: 128-256, Technical: 1024-2048 | |
| **Implementation** | **PARTIALLY MATCHES** | Uses 400 characters (~100 tokens assuming ~4 chars/token), which is below the recommended minimum. This is justified by empirical grid search evaluation. |
| **Code Reference** | `rag_config.py` line 51: `chunk_size: int = 400`. Comment: "OPTIMIZED: Based on grid search evaluation (visionary_rag_v5_grand_table.csv)". |

### 8. Evaluating Chunking Strategies

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Test retrieval metrics (precision, recall) and end-to-end metrics (answer accuracy, groundedness) | |
| **Implementation** | **MATCH** | `C:\Users\kommi\ragpipeline\bench_and_improve.py` evaluates chunking configurations with keyword coverage, faithfulness, recall@5, NDCG@5, and composite quality scores across multiple configurations. |
| **Code Reference** | `bench_and_improve.py` lines 222-267: `compute_metrics()` function calculating keyword_coverage, faithfulness, not_found_penalty, recall_at_5, ndcg_at_5, and quality_score. |

### 9. Decision Framework

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Start simple (fixed-size + overlap), measure, iterate, A/B test | |
| **Implementation** | **MATCH** | `rag_config.py` provides three presets: `balanced`, `low_latency`, `high_quality`. `bench_and_improve.py` runs benchmark comparisons across configurations. |

### 10. Common Pitfalls

| Pitfall | Status | Details |
|---------|--------|---------|
| Too small chunks | **MATCH** | 400 chars is on the smaller side but empirically validated |
| Too large chunks | **N/A** | Not an issue in this project |
| No overlap | **N/A** | Overlap is implemented (37.5%) |
| Ignoring structure | **PARTIALLY** | Tables handled specially, but no code/formula-specific handling |
| One-size-fits-all | **PARTIALLY** | Prose vs table differentiation exists, but not more granular |
| No evaluation | **N/A** | Evaluation framework exists |

---

## Gap Analysis Summary

### Implemented
- Recursive character text splitting with proper separator hierarchy
- Parent-child hierarchical chunking
- Metadata preservation (chapter, section, subsection, page_number, content_type)
- Content-type differentiation (prose vs tables)
- Empirically optimized chunk size and overlap
- Evaluation framework with multiple metrics

### Partially Implemented
- Token-based chunking (uses character-based instead)
- Content-type-specific strategies (only prose vs tables)

### NOT Implemented
- Semantic chunking based on sentence similarity
- Code syntax-aware splitting
- Markdown/HTML-aware chunking with header preservation
- Sliding window with contextual overlap
- Hybrid hierarchical chunking at 3+ levels

---

## Overall Assessment: PARTIALLY MATCHES

The project has a solid parent-child chunking implementation with recursive character splitting and empirically validated parameters. The chunking strategy is well-tested and evaluated. However, it lacks semantic chunking, content-type-specific strategies beyond prose/tables, and token-based (rather than character-based) chunking.
