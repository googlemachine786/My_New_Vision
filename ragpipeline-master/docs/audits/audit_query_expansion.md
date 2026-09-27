# Audit: Query Expansion Guide vs. Implementation

**Guide:** `guide_query_expansion.md` (Ailog, Nov 2025)
**Audited:** 2026-04-03
**Scope:** C:\Users\kommi\ragpipeline\

---

## Guide Summary

This guide covers query expansion techniques to improve recall by 40%: synonym expansion, LLM query rewriting, multi-query retrieval, HyDE (Hypothetical Document Embeddings), step-back prompting, and sub-query decomposition.

---

## Key Recommendations and Implementation Match

### 1. Synonym Expansion

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Expand queries with synonyms using WordNet or custom dictionaries | |
| **Implementation** | **DOES NOT MATCH** | No synonym expansion found. The `ExtractKeywords()` function in `hybrid_search.go` uses stopword filtering but no synonym replacement or addition. |
| **Missing** | No WordNet usage, no synonym dictionary, no query variation generation via synonyms. |

### 2. LLM Query Rewriting

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Use LLM to generate 3 alternative phrasings of the user's query | |
| **Implementation** | **PARTIALLY MATCHES** | The `ClarificationGenerator` in `C:\Users\kommi\ragpipeline\orchestrator\llm\clarification_generator.go` analyzes query ambiguity and generates `suggested_refinements` and `clarification_questions`. However, these are for asking the user for clarification, not for expanding the search query. The `RefineQuery()` method (line 200) incorporates user clarification into the query, which is a form of rewriting but requires user interaction. |
| **Code Reference** | `clarification_generator.go` lines 51-80: `AnalyzeQuery()`. Lines 200-204: `RefineQuery()`. |
| **Missing** | No automatic LLM-generated query variations for parallel retrieval. |

### 3. Multi-Query Retrieval

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Search with all query variations, deduplicate, and rank by frequency/combined score | |
| **Implementation** | **DOES NOT MATCH** | No multi-query retrieval. Each query results in a single embedding and single search. |
| **Missing** | No `expand_with_llm()` function, no parallel search with multiple queries, no score aggregation across query variations. |

### 4. HyDE (Hypothetical Document Embeddings)

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Generate hypothetical answer with LLM, embed the hypothetical answer (not the query), search for similar documents | |
| **Implementation** | **DOES NOT MATCH** | No HyDE implementation found. |
| **Missing** | No hypothetical answer generation, no embedding of hypothetical documents for search. |

### 5. Step-Back Prompting

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Generate a broader, more general question first, then search both specific and broad queries | |
| **Implementation** | **DOES NOT MATCH** | No step-back prompting found. |
| **Missing** | No broader question generation, no dual search (specific + broad). |

### 6. Sub-Query Decomposition

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Break complex queries into 2-3 simpler sub-questions, retrieve for each, deduplicate results | |
| **Implementation** | **DOES NOT MATCH** | No query decomposition found. |
| **Missing** | No LLM-based sub-question generation, no per-sub-question retrieval, no result aggregation. |

### 7. LangChain MultiQueryRetriever

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Use LangChain's `MultiQueryRetriever` for automatic query expansion | |
| **Implementation** | **DOES NOT MATCH** | No LangChain usage in the project. |

### 8. Evaluation of Query Expansion

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Measure recall improvement: baseline vs expanded | |
| **Implementation** | **PARTIALLY MATCHES** | The project evaluates retrieval quality (`bench_and_improve.py`, `rrf.go`) but does not compare baseline vs expanded query retrieval. |
| **Missing** | No side-by-side comparison of recall with and without query expansion. |

### 9. Self-Query as Query Enhancement

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Extract structured information from queries for better search | |
| **Implementation** | **MATCH** | The `SelfQueryRetriever` extracts metadata filters (grade, chapter, content_type) from natural language queries, which improves search precision. This is a form of query enhancement, though different from the expansion techniques in the guide. |
| **Code Reference** | `C:\Users\kommi\ragpipeline\retrievers\self_query_retriever.py` lines 60-120. `self_query_parser.py`: LLM-based filter extraction using Ollama. |

---

## Gap Analysis Summary

### Implemented
- Self-query retrieval (metadata filter extraction from queries)
- Query ambiguity analysis with suggested refinements
- Query refinement via user clarification (`RefineQuery()`)

### Partially Implemented
- Query rewriting (only via user clarification, not automatic)
- Evaluation (exists but not for query expansion specifically)

### NOT Implemented
- Synonym expansion
- Automatic LLM query rewriting (generating variations without user input)
- Multi-query retrieval (searching with multiple variations)
- HyDE (Hypothetical Document Embeddings)
- Step-back prompting
- Sub-query decomposition
- LangChain MultiQueryRetriever

---

## Overall Assessment: DOES NOT MATCH

This is a significant gap. The project has self-query retrieval (which extracts filters) but lacks all forms of query expansion for recall improvement. The guide recommends query expansion as "low-cost, high-impact" that can "boost recall by 30-50% instantly." The only query enhancement in the project is metadata filter extraction, which improves precision but does not improve recall. Adding LLM-based query rewriting for multi-query retrieval would be the highest-impact addition from this guide.
