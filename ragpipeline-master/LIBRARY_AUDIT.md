# Library Audit Report

**Scope:** `services/` (Go microservices) + `ingestion/` (Python pipeline)
**Date:** 2026-04-07
**Analyst:** Codebase Audit

---

## Executive Summary

This audit identifies **14 findings** across the Go services and Python ingestion pipeline where hand-rolled custom implementations should be replaced with production-grade libraries. The codebase shows strong engineering fundamentals but has several areas where battle-tested libraries would improve reliability, accuracy, and maintainability.

**Priority Breakdown:**
| Priority | Count | Description |
|----------|-------|-------------|
| P0 | 3 | Blocks production readiness |
| P1 | 7 | Degrades quality or increases maintenance burden |
| P2 | 4 | Nice-to-have improvements |

---

## Findings: Go Services

### 1. Custom BM25 Scorer (P1)

| Attribute | Value |
|-----------|-------|
| **File** | `services/vector-search-service/search/bm25_scorer.go` (lines 1–120) |
| **Current** | Hand-rolled Okapi BM25 with custom `KeywordIndex`, manual IDF/TF computation, insertion sort |
| **Should Use** | [`bleve`](https://github.com/blevesearch/bleve) or [`blevesearch/analysis`](https://github.com/blevesearch/bleve) with built-in BM25 scoring |
| **Type** | Go service |

**Details:**
- `BM25Scorer` (line 12) implements BM25 from scratch with a custom `KeywordIndex` data structure
- Tokenization is naive `strings.Fields` + punctuation trim (line 104–116) — no stemming, no stop-word removal
- Sorting uses O(n²) insertion sort (line 118) instead of `sort.Slice`
- The adjacent `KeywordIndex` (`search/index/keyword_index.go`) is also entirely hand-rolled

**Why it matters:** Bleve provides production-tested inverted indexing, token analyzers with stemming/stop-words, and optimized scoring. The current implementation will miss edge cases (Unicode normalization, CJK tokenization, compound words).

---

### 2. Custom Keyword Index (P1)

| Attribute | Value |
|-----------|-------|
| **File** | `services/vector-search-service/index/keyword_index.go` (lines 1–115) |
| **Current** | Hand-rolled inverted index with `map[string]map[string]int` postings |
| **Should Use** | [`bleve`](https://github.com/blevesearch/bleve) index or [`zoekt`](https://github.com/sourcegraph/zoekt) |
| **Type** | Go service |

**Details:**
- `KeywordIndex` (line 9) maintains three in-memory maps: postings, docLengths, docTerms
- `tokenize()` (line 103) uses the same naive `strings.Fields` approach — no language-aware tokenization
- No persistence: index is rebuilt on every service restart

---

### 3. Custom Sentence Splitting / Claim Extraction (P0)

| Attribute | Value |
|-----------|-------|
| **File** | `services/api-gateway/handler/grounding_validator.go` (lines 130–160) |
| **Current** | Hand-rolled `splitSentences()` that splits on `.`, `!`, `?` characters |
| **Should Use** | NLP library with proper sentence boundary detection. For Go: [`go-nlp`](https://github.com/james-bowman/nlp) or delegate to a Python NLI microservice |
| **Type** | Go service |

**Details:**
- `splitSentences()` (line 142) naively splits on punctuation — fails on abbreviations ("Dr.", "e.g.", "i.e."), decimal numbers ("3.14"), and ellipsis
- `extractClaims()` (line 125) uses simple length filtering (< 10 chars) and prefix matching for noise
- This directly impacts grounding validation accuracy — false claim boundaries produce incorrect grounding scores

**Why P0:** Grounding validation is a quality-critical production feature. Incorrect sentence splitting means claims are misidentified, leading to wrong grounding scores and potentially misleading users about response quality.

---

### 4. Custom Keyword Overlap Similarity (P0)

| Attribute | Value |
|-----------|-------|
| **File** | `services/api-gateway/handler/grounding_validator.go` (lines 195–217) |
| **Current** | Hand-rolled `keywordOverlapSimilarity()` using simple word overlap ratio |
| **Should Use** | Proper NLI model via a Python microservice (e.g., CrossEncoder with `cross-encoder/nli-deberta-v3-base`) or LLM-based verification |
| **Type** | Go service |

**Details:**
- `keywordOverlapSimilarity()` (line 195) computes a naive Jaccard-like overlap: `overlap / len(claimWords)`
- `isClaimGrounded()` (line 177) uses this as a proxy for semantic similarity when substring matching fails
- This is fundamentally flawed: "Apple is a fruit" and "Apple is a tech company" would score identically despite contradictory meanings

**Why P0:** Grounding validation is the system's fact-checking mechanism. Using keyword overlap instead of semantic understanding means contradictions can be falsely marked as "grounded."

---

### 5. Custom Cosine Similarity in Semantic Cache (P1)

| Attribute | Value |
|-----------|-------|
| **File** | `services/api-gateway/cache/semantic_cache.go` (lines 266–283) |
| **Current** | Hand-rolled `cosineSimilarity()` with manual dot product and norm computation |
| **Should Use** | [`gonum/gonum/mat`](https://pkg.go.dev/gonum.org/v1/gonum/mat) vector operations |
| **Type** | Go service |

**Details:**
- The function (line 267) loops over vectors computing dot product and norms manually
- No SIMD optimization, no vectorized operations
- Called in a tight loop during cache lookups (line 140) — performance-sensitive

**Why it matters:** While functionally correct, gonum provides optimized BLAS-level operations. For 768-dim embeddings scanned across thousands of cache entries, this is a measurable performance difference.

---

### 6. Custom NLI Prompt Templates (P1)

| Attribute | Value |
|-----------|-------|
| **File** | `services/query-understanding-service/prompts/nli_templates.go` (entire file) |
| **Current** | LLM-based NLI via prompt engineering (JSON output from Gemini) |
| **Should Use** | Dedicated NLI model (e.g., HuggingFace `cross-encoder/nli-deberta-v3-base`) served as a Python microservice |
| **Type** | Go service (prompts LLM) |

**Details:**
- `NLIClaimVerificationPrompt()` generates prompts asking an LLM to classify claims as SUPPORTED/CONTRADICTED/NOT_ENOUGH_INFO
- This is expensive ($0.002–0.01 per claim), slow (1–3s latency per call), and non-deterministic
- Production NLI systems use fine-tuned cross-encoder models that run in <50ms locally

---

### 7. Custom Rate Limiting (P2)

| Attribute | Value |
|-----------|-------|
| **File** | `services/api-gateway/middleware/rate_limit.go` (entire file) |
| **Current** | Custom token-bucket implementation with Redis + in-memory fallback |
| **Should Use** | [`go.uber.org/ratelimit`](https://pkg.go.dev/go.uber.org/ratelimit) or [`github.com/ulule/limiter`](https://github.com/ulule/limiter) |
| **Type** | Go service |

**Details:**
- `LocalRateLimiter` (line 42) implements token bucket from scratch
- Redis-based rate limiting uses fixed-window counters (line 119), not sliding window
- The implementation works but has edge cases around clock skew and burst handling

**Why P2:** The current implementation is functional. A production library would provide better sliding-window algorithms and more thorough testing, but this is not a blocker.

---

### 8. Manual RRF Fusion (P2)

| Attribute | Value |
|-----------|-------|
| **File** | `services/vector-search-service/search/rrf.go` (entire file) |
| **Current** | Hand-rolled Reciprocal Rank Fusion implementation |
| **Should Use** | RRF is a standard algorithm; hand-rolled is acceptable. Consider [`blevesearch/bleve`](https://github.com/blevesearch/bleve) for full hybrid search orchestration |
| **Type** | Go service |

**Details:**
- The RRF implementation (line 63) is correct and well-documented
- The tiebreaker logic (line 158) and deduplication (line 176) are sound
- This is actually a case where the custom implementation is reasonable

**Recommendation:** Keep as-is. RRF is a simple enough algorithm that a custom implementation is defensible.

---

## Findings: Python Ingestion Pipeline

### 9. Custom Text Chunking (P1)

| Attribute | Value |
|-----------|-------|
| **File** | `ingestion/chunker/parent_child.py` (lines 78–185) |
| **Current** | Custom `RecursiveCharacterTextSplitter` with separator-based splitting |
| **Should Use** | [`langchain.text_splitter.RecursiveCharacterTextSplitter`](https://python.langchain.com/docs/modules/data_connection/document_transformers/recursive_text_splitter) or [`llama_index.core.node_parser.TokenTextSplitter`](https://docs.llamaindex.ai/en/stable/api_reference/node_parsers/token/) |
| **Type** | Python component |

**Details:**
- `RecursiveCharacterTextSplitter` (line 78) is a from-scratch implementation that mirrors LangChain's design
- Overlap logic (line 147) copies trailing characters from previous chunk — this can split words mid-token
- No token-aware splitting (counts characters, not tokens) — critical issue since embeddings are token-based
- The implementation is well-parameterized (400 chars / 150 overlap from grid search) but the splitting logic doesn't respect tokenizer boundaries

**Why P1:** Character-based chunking without token awareness can split semantic units (words, subwords) across chunk boundaries, degrading embedding quality. LangChain/LlamaIndex chunkers respect token boundaries.

---

### 10. Custom MinHash LSH Deduplication (P1)

| Attribute | Value |
|-----------|-------|
| **File** | `ingestion/dedup/detector.py` (lines 115–195) |
| **Current** | Hand-rolled MinHash LSH with custom shingling and permutation |
| **Should Use** | [`datasketch`](https://ekzhu.github.io/datasketch/lsh.html) library for MinHash LSH |
| **Type** | Python component |

**Details:**
- `MinHashLSH` class (line 115) implements shingling, minhash computation, and banding from scratch
- Uses `hash()` Python builtin (line 144) which is non-deterministic across Python processes (PYTHONHASHSEED)
- `np.random.seed(42)` (line 131) for permutation parameters — works but datasketch provides optimized C implementations
- The `_shingles()` method (line 136) uses character 5-grams — no language-aware tokenization

**Why P1:** The datasketch library is production-tested, handles edge cases, and uses optimized C backends. The custom `hash()` usage introduces non-determinism risk in production.

---

### 11. Custom Semantic Deduplication with NumPy (P2)

| Attribute | Value |
|-----------|-------|
| **File** | `ingestion/dedup/detector.py` (lines 62–113) |
| **Current** | Manual cosine similarity via `np.dot` and `np.linalg.norm` in O(n²) loop |
| **Should Use** | [`sklearn.metrics.pairwise.cosine_similarity`](https://scikit-learn.org/stable/modules/generated/sklearn.metrics.pairwise.cosine_similarity.html) or FAISS for large-scale |
| **Type** | Python component |

**Details:**
- `SemanticDeduplicator.find_duplicates()` (line 75) loops over embeddings one-by-one
- For large corpora, this becomes O(n²) — FAISS would provide ANN search at scale
- The current approach works for small-scale but doesn't scale to production corpus sizes

---

### 12. Custom Formula Detection via Regex (P2)

| Attribute | Value |
|-----------|-------|
| **File** | `ingestion/parser/formula_detector.py` (entire file) |
| **Current** | Regex-based detection for chemical formulas, measurements, equations |
| **Should Use** | For production: a Python NLP microservice with a fine-tuned NER model, or accept regex as sufficient for the narrow CBSE science domain |
| **Type** | Python component |

**Details:**
- `CHEMICAL_FORMULA_REGEX` (line 36) matches patterns like H₂O, CO₂ — works for simple cases
- `EQUATION_REGEX` (line 55) matches "X = YZ" patterns — will produce false positives
- For the CBSE science textbook domain (limited scope), regex may be sufficient
- No handling of LaTeX markup, MathML, or complex chemical notation

**Why P2:** For the specific domain of CBSE textbooks (grades 6–8), regex-based detection covers ~90% of cases. A full NER model would be overkill unless the system needs to handle arbitrary scientific documents.

---

### 13. Custom Font Calibration / Heading Detection (P2)

| Attribute | Value |
|-----------|-------|
| **File** | `ingestion/parser/font_calibrator.py` + `ingestion/parser/heading_mapper.py` |
| **Current** | Statistical font size analysis (mode, 90th/97th percentiles) + regex patterns |
| **Should Use** | [`layoutparser`](https://layout-parser.github.io/) or [`pdfplumber`](https://github.com/jsvine/pdfplumber) structural analysis |
| **Type** | Python component |

**Details:**
- `calibrate_font_thresholds()` uses `statistics.mode()` for body text, percentiles for headings
- `_classify_heading_level()` (heading_mapper.py line 216) uses hardcoded thresholds
- The approach is clever (adaptive to each PDF) but fragile — depends on consistent typography
- Could benefit from layout-aware parsing that considers position, not just font size

---

### 14. Custom Table-to-Markdown Conversion (P2)

| Attribute | Value |
|-----------|-------|
| **File** | `ingestion/parser/table_extractor.py` (lines 83–125) |
| **Current** | Manual `_table_to_markdown()` with pipe-delimited string building |
| **Should Use** | [`pandas.DataFrame.to_markdown()`](https://pandas.pydata.org/docs/reference/api/pandas.DataFrame.to_markdown.html) or [`tabulate`](https://pypi.org/project/tabulate/) |
| **Type** | Python component |

**Details:**
- `_table_to_markdown()` (line 83) manually joins cells with `|` separators
- Doesn't handle multi-line cells, merged cells, or special character escaping beyond pipes
- `tabulate` library provides better markdown table formatting with alignment and edge cases handled

---

## Current Library Usage Assessment

### Go Services — Dependencies

| Library | Usage | Assessment |
|---------|-------|------------|
| `gorilla/mux` | HTTP routing | ✅ Standard, appropriate |
| `redis/go-redis/v9` | Redis client + caching | ✅ Production-grade |
| `pgvector/pgvector-go` | Vector search via pgvector | ✅ Appropriate for PostgreSQL |
| `jackc/pgx/v5` | PostgreSQL driver | ✅ Production-grade |
| `cenkalti/backoff/v4` | Retry logic | ✅ Production-grade |
| `rs/zerolog` | Structured logging | ✅ Production-grade |
| `prometheus/client_golang` | Metrics | ✅ Production-grade |
| `google/uuid` | UUID generation | ✅ Production-grade |
| `golang-jwt/jwt/v5` | JWT auth | ✅ Production-grade |
| `otel/*` | OpenTelemetry tracing | ✅ Production-grade |

**Missing Go libraries that should be added:**
| Library | Purpose |
|---------|---------|
| `blevesearch/bleve` | BM25/inverted index (replaces custom KeywordIndex + BM25Scorer) |
| `gonum/gonum/mat` | Vector math (replaces custom cosine similarity) |
| `go-nlp/nlp` or delegate to Python | NLI/grounding validation |

### Python Ingestion — Dependencies

| Library | Usage | Assessment |
|---------|-------|------------|
| `PyMuPDF` (fitz) | PDF text extraction, font detection | ✅ Production-grade |
| `pdfplumber` | Table extraction with bounding boxes | ✅ Production-grade |
| `yake` | Keyword extraction | ✅ Appropriate for domain |
| `google-cloud-aiplatform` | Vertex AI embeddings | ✅ Production-grade |
| `psycopg[binary]` | Async PostgreSQL driver | ✅ Production-grade |
| `pgvector` | Vector type support | ✅ Production-grade |
| `tenacity` | Retry logic | ✅ Production-grade |
| `structlog` | Structured logging | ✅ Production-grade |
| `pydantic` + `pydantic-settings` | Data validation | ✅ Production-grade |

**Missing Python libraries that should be added:**
| Library | Purpose |
|---------|---------|
| `langchain` or `llama-index` | Token-aware text chunking (replaces custom RecursiveCharacterTextSplitter) |
| `datasketch` | MinHash LSH deduplication (replaces custom implementation) |
| `scikit-learn` | Cosine similarity for dedup, vector operations |
| `faiss-cpu` | Scalable semantic deduplication for large corpora |
| `spaCy` or `nltk` | Proper sentence splitting, tokenization, stemming |
| `tabulate` | Table-to-markdown conversion (replaces manual conversion) |
| `layoutparser` | Layout-aware PDF parsing (improves heading detection) |

---

## Anti-Pattern Summary Matrix

| # | Anti-Pattern | Location | Priority | Type | Recommended Library |
|---|-------------|----------|----------|------|-------------------|
| 1 | Custom BM25 scorer | `search/bm25_scorer.go` | P1 | Go | bleve |
| 2 | Custom keyword index | `index/keyword_index.go` | P1 | Go | bleve |
| 3 | Custom sentence splitting | `grounding_validator.go:142` | **P0** | Go | go-nlp / Python NLI service |
| 4 | Custom keyword overlap similarity | `grounding_validator.go:195` | **P0** | Go | Python NLI cross-encoder |
| 5 | Custom cosine similarity | `semantic_cache.go:267` | P1 | Go | gonum/mat |
| 6 | LLM-based NLI prompts | `nli_templates.go` | P1 | Go | Python NLI microservice |
| 7 | Custom rate limiting | `rate_limit.go` | P2 | Go | ulule/limiter |
| 8 | Manual RRF fusion | `rrf.go` | P2 | Go | Keep as-is (acceptable) |
| 9 | Custom text chunking | `chunker/parent_child.py:78` | P1 | Python | langchain / llama-index |
| 10 | Custom MinHash LSH | `dedup/detector.py:115` | P1 | Python | datasketch |
| 11 | Custom semantic dedup | `dedup/detector.py:62` | P2 | Python | scikit-learn / FAISS |
| 12 | Custom formula detection | `formula_detector.py` | P2 | Python | NLP microservice (domain-dependent) |
| 13 | Custom font calibration | `font_calibrator.py` | P2 | Python | layoutparser |
| 14 | Custom table-to-markdown | `table_extractor.py:83` | P2 | Python | tabulate / pandas |

---

## Recommended Migration Path

### Phase 1: Critical (P0) — Block Production Readiness
1. **Replace grounding validator's claim extraction** with a proper NLP sentence splitter
   - Option A: Add `go-nlp` library for basic sentence boundary detection
   - Option B (recommended): Create a Python microservice with spaCy sentence segmentation + CrossEncoder NLI
2. **Replace keyword overlap similarity** with semantic similarity
   - Deploy a CrossEncoder model (`cross-encoder/nli-deberta-v3-base`) as a validation service
   - Or use LLM-based verification with structured output (current NLI prompt approach, but with proper model)

### Phase 2: Quality Improvement (P1)
3. **Replace BM25 + KeywordIndex** with bleve
4. **Replace custom cosine similarity** with gonum/mat
5. **Replace custom text chunking** with LangChain's RecursiveCharacterTextSplitter
6. **Replace custom MinHash LSH** with datasketch
7. **Replace LLM-based NLI prompts** with dedicated NLI model microservice

### Phase 3: Polish (P2)
8. **Replace rate limiter** with ulule/limiter
9. **Add FAISS** for scalable deduplication
10. **Replace formula detection** with NER model (if domain expands beyond CBSE)
11. **Replace font calibration** with layoutparser
12. **Replace table-to-markdown** with tabulate

---

## What's Done Well

- **Embedding service**: Properly uses Vertex AI `text-embedding-005` with batching and retry ✅
- **Keyword extraction**: Uses YAKE (production library) with domain-optimized parameters ✅
- **PDF parsing**: Uses PyMuPDF + pdfplumber (industry-standard libraries) ✅
- **Vector search**: Uses pgvector with ScaANN index (production-grade) ✅
- **Database layer**: Uses pgx with connection pooling and transactions ✅
- **Observability**: OpenTelemetry + Prometheus + structured logging throughout ✅
- **Retry logic**: Uses cenkalti/backoff and tenacity ✅
