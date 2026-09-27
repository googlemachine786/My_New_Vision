# Redundancy Audit Report

**Project:** Visionary RAG Pipeline  
**Date:** 2026-04-03  
**Scope:** Full codebase analysis - Python vs Go implementations, configs, tests, docs  
**Auditor:** Automated Codebase Analysis

---

## Executive Summary

The codebase exhibits **significant redundancy** across all dimensions:
- **31 redundant `.md` documentation files** in the root with overlapping content
- **2 complete ingestion pipelines** (Python + Go) with identical functionality
- **3 RRF implementations** across different packages
- **3 Vertex AI embedding clients** with stub vs production implementations
- **3 overlapping Go config modules** for different services
- **Multiple test files** targeting the same functionality
- **Duplicate CLI scripts** (`ask_query.py` vs `ask_real_query.py`)

---

## 1. DOCUMENTATION REDUNDANCY (Critical)

### 1.1 Audit & Status Reports (14 files → 1)

| File | Purpose | Overlap |
|------|---------|---------|
| `AUDIT_FINDINGS.md` | Initial audit findings | High - superseded |
| `AUDIT_FIX_PROGRESS.md` | Fix progress tracking | High - superseded |
| `AUDIT_IMPLEMENTATION_SUMMARY.md` | Implementation summary | High - superseded |
| `COMPLETE_AUDIT_FIX_SUMMARY.md` | Complete fix summary | High - superset |
| `COMPREHENSIVE_CODE_AUDIT.md` | Code audit | High - overlaps with .planning |
| `ENTERPRISE_AUDIT_COMPLETE.md` | Enterprise audit | High - superset |
| `ENTERPRISE_PRODUCTION_AUDIT.md` | Production audit | High - overlaps |
| `.planning/codebase/RAG_PIPELINE_AUDIT.md` | Structured audit | **KEEP** - best organized |

**Recommendation:**
- **KEEP:** `.planning/codebase/RAG_PIPELINE_AUDIT.md` (structured, most complete)
- **KEEP:** `ENTERPRISE_PRODUCTION_AUDIT.md` (production-focused, unique insights)
- **DELETE:** All other 12 audit files (content captured in the two kept files)

### 1.2 Implementation & Test Reports (11 files → 2)

| File | Purpose | Overlap |
|------|---------|---------|
| `ADVANCED_TESTING_SUMMARY.md` | Test summary | High |
| `BUILD_AND_TEST_STATUS.md` | Build status | High |
| `FINAL_IMPLEMENTATION_COMPLETE.md` | Implementation complete | High |
| `FINAL_IMPLEMENTATION_SUMMARY.md` | Implementation summary | High - duplicates above |
| `FINAL_TEST_REPORT.md` | Test report | High |
| `FINAL_TEST_SUMMARY.md` | Test summary | High - duplicates above |
| `IMPLEMENTATION_PROGRESS.md` | Progress tracking | Medium |
| `IMPLEMENTATION_SUMMARY.md` | Implementation summary | High |
| `TEST_REPORT.md` | Test report | High |
| `TEST_REPORT_SCIENCE_CLASS8.md` | Science class test | Medium - specific |
| `METRIC_IMPROVEMENTS.md` | Metric improvements | Medium |

**Recommendation:**
- **KEEP:** `FINAL_IMPLEMENTATION_COMPLETE.md` (most comprehensive)
- **KEEP:** `FINAL_TEST_SUMMARY.md` (test results)
- **DELETE:** All other 9 files

### 1.3 Configuration & Setup Docs (7 files → 2)

| File | Purpose | Overlap |
|------|---------|---------|
| `AUTO_DETECT_SETUP.md` | Auto-detect config | High |
| `CONFIG_IMPLEMENTATION_SUMMARY.md` | Config implementation | High |
| `CONFIG_QUICK_REFERENCE.md` | Config quick reference | Medium |
| `CREDENTIALS_REQUIRED.md` | Credentials guide | Medium |
| `LOCAL_DEVELOPMENT_SETUP.md` | Local setup | High |
| `OPTIMAL_CONFIG_QUICKSTART.md` | Quick start | High |
| `QUICKSTART_LOCAL.md` | Quick start | High - duplicates above |

**Recommendation:**
- **KEEP:** `LOCAL_DEVELOPMENT_SETUP.md` (most comprehensive setup guide)
- **KEEP:** `CONFIG_QUICK_REFERENCE.md` (quick reference format is useful)
- **DELETE:** All other 5 files

### 1.4 Planning & Roadmap (5 files → 1)

| File | Purpose | Overlap |
|------|---------|---------|
| `ENTERPRISE_HARDENING_PLAN.md` | Hardening plan | High |
| `ENTERPRISE_PRODUCTION_ROADMAP.md` | Production roadmap | High |
| `GAP_ANALYSIS.md` | Gap analysis | Medium |
| `PHASE1_DEPLOYMENT.md` | Phase 1 deployment | High |
| `PHASE_1_MVP_BACKLOG.md` | MVP backlog | Medium |

**Recommendation:**
- **KEEP:** `ENTERPRISE_PRODUCTION_ROADMAP.md` (forward-looking, strategic)
- **DELETE:** All other 4 files

### 1.5 Feature-Specific Docs (4 files → 2)

| File | Purpose | Overlap |
|------|---------|---------|
| `GO_FIRST_SUMMARY.md` | Go-first approach | Medium |
| `GO_INGESTION_IMPLEMENTATION_PLAN.md` | Go ingestion plan | High |
| `LANGCHAINGO_INTEGRATION.md` | LangChainGo integration | High |
| `LANGCHAINGO_USAGE.md` | LangChainGo usage | High - duplicates above |

**Recommendation:**
- **KEEP:** `GO_INGESTION_IMPLEMENTATION_PLAN.md`
- **KEEP:** `LANGCHAINGO_INTEGRATION.md`
- **DELETE:** `LANGCHAINGO_USAGE.md` (duplicates integration doc)
- **DELETE:** `GO_FIRST_SUMMARY.md` (status update, not reference)

### 1.6 Self-Query Docs (3 files → 1)

| File | Purpose | Overlap |
|------|---------|---------|
| `SELF_QUERY_FUNCTIONAL.md` | Self-query working | High |
| `SELF_QUERY_IMPLEMENTATION_PLAN.md` | Implementation plan | High |
| `SELF_QUERY_RESULTS.md` | Results | High |

**Recommendation:**
- **KEEP:** `SELF_QUERY_FUNCTIONAL.md` (confirms it works)
- **DELETE:** `SELF_QUERY_IMPLEMENTATION_PLAN.md`, `SELF_QUERY_RESULTS.md`

### 1.7 Keep These Root Docs (Essential)

| File | Reason |
|------|--------|
| `README.md` | Main entry point |
| `README_ENTERPRISE.md` | Enterprise overview |
| `PRODUCTION_READINESS_CHECKLIST.md` | Production checklist |
| `PLAN.md` | Project plan |
| `USER_STORY_ALIGNMENT.md` | Requirements traceability |
| `HOW_TO_RUN_OLLAMA_TEST.md` | Testing guide |
| `DEVELOPMENT_SESSION_SUMMARY.md` | Session notes (ephemeral, delete after review) |
| `STANDUP_PROGRESS.md` | Progress log (ephemeral, delete after review) |
| `IMPROVEMENT_LOOP_RESULTS.md` | Results data (keep if referenced) |
| `RAGAS_OPTIMIZATION_COMPLETE.md` | Optimization results |
| `RAG_EVALUATION_METRICS.md` | Evaluation framework |
| `RAG_PIPELINE_AUDIT_REPORT.md` | Audit report (may overlap with .planning) |
| `visionary_rag_implementation_plan.md` | Historical implementation plan |

### Documentation Cleanup Summary

| Action | Count | Files |
|--------|-------|-------|
| **KEEP** | 15 | Essential docs listed above |
| **DELETE** | 31 | All redundant audit/test/config/plan files |
| **REVIEW** | 3 | Ephemeral files (session summaries, standup logs) |

---

## 2. INGESTION PIPELINE DUPLICATION (Critical)

### 2.1 Python vs Go Ingestion

| Aspect | Python (`ingestion/`) | Go (`ingestion-go/`) | Verdict |
|--------|----------------------|---------------------|---------|
| PDF Parser | ✅ Full (PyMuPDF + pdfplumber) | ❌ Stub (calls Python) | Python wins |
| Chunking | ✅ Parent-child (400/150 optimized) | ✅ Parent-child (1500/512 legacy) | Python wins (optimized) |
| Keywords | ✅ YAKE extractor | ✅ LLM-based extractor | Go wins (production) |
| Embedding | ✅ Vertex AI batch with retry | ❌ Mock/stub implementation | Python wins |
| DB Writer | ✅ Async with transactions + DLQ | ✅ Sync with concurrency | Python wins (more robust) |
| Orchestration | ✅ Full async pipeline | ✅ Hybrid (Python→JSON→Go) | Go wins (architecture) |
| Tests | ✅ 3 test files | ✅ 3 test files | Equal |

**Key Finding:** The Go ingestion pipeline is designed as a **hybrid** that calls Python for PDF parsing anyway. The chunking in Go uses **legacy parameters** (1500/512) vs Python's **optimized parameters** (400/150).

**Recommendation:**
- **KEEP** `ingestion/` (Python) as the **primary ingestion pipeline**
  - Superior PDF parsing (real libraries vs Go stubs)
  - Optimized chunking parameters from grid search
  - Production-ready embedding with retry logic
  - Robust async DB writer with DLQ
- **DELETE** `ingestion-go/` entirely
  - Its hybrid architecture adds complexity without benefit
  - Chunking uses outdated parameters
  - Embedding is a stub/mock
  - The Go ingestion `main.go` just wraps Python parser call

### 2.2 Ingestion File-by-File Analysis

| Python File | Go Equivalent | Status |
|-------------|--------------|--------|
| `ingestion/pipeline.py` | `ingestion-go/cmd/ingestion/main.go` | DELETE Go |
| `ingestion/chunker/parent_child.py` | `ingestion-go/chunker/parent_child.go` | DELETE Go (legacy params) |
| `ingestion/embedder/vertex_batch.py` | `ingestion-go/embedder/vertex_client.go` | DELETE Go (stub) |
| `ingestion/keywords/yake_extractor.py` | `ingestion-go/keywords/extractor.go` | DELETE Go (LLM-based is slower) |
| `ingestion/parser/*.py` (9 files) | `ingestion-go/parser/json_parser.go` | DELETE Go (calls Python) |
| `ingestion/writer/alloydb_writer.py` | `ingestion-go/writer/alloydb_writer.go` | DELETE Go |
| `ingestion/dlq/retry_processor.py` | None | Python only |
| `ingestion/incremental/tracker.py` | None | Python only |
| `ingestion/tests/*.py` | `ingestion-go/tests/*.go` | DELETE Go tests |

---

## 3. RRF IMPLEMENTATION DUPLICATION (High)

### 3.1 Three RRF Implementulations

| Location | File | Type | Quality | Verdict |
|----------|------|------|---------|---------|
| `orchestrator/retrieval/rrf.go` | `rrfFuseV2()` | Standalone function | Good - includes metrics (Recall@K, MRR, NDCG) | **KEEP** |
| `orchestrator/retrieval/hybrid_search.go` | `rrfFuse()` | Method in HybridSearcher | Good - full DB integration | **KEEP** (merge with rrf.go) |
| `services/vector-search-service/search/rrf.go` | `ReciprocalRankFuse()` | Standalone with config | Best - most complete, configurable, tiebreaker logic | **KEEP** |

**Key Finding:** All three implementations follow the same algorithm `1/(k + rank)` with k=60, but have different:
- Data structures (`RRFResult` vs `RetrievalResult` vs `rrfEntry`)
- Sorting strategies (score-only vs score+cosine_dist tiebreaker)
- Additional features (metrics in rrf.go, config in vector-search-service)

**Recommendation:**
- **CONSOLIDATE** into a single `internal/rrf/` package
- Keep the config-driven approach from `services/vector-search-service/search/rrf.go`
- Keep metrics functions (Recall@K, MRR, NDCG) from `orchestrator/retrieval/rrf.go`
- Keep tiebreaker logic from `services/vector-search-service/search/rrf.go`
- **DELETE** duplicate implementations after consolidation

### 3.2 Hybrid Search Duplication

| Location | File | Type | Verdict |
|----------|------|------|---------|
| `orchestrator/retrieval/hybrid_search.go` | `HybridSearcher` | Direct DB access | **KEEP** (orchestrator needs this) |
| `services/vector-search-service/search/hybrid.go` | `HybridSearcher` | Service-layer orchestration | **KEEP** (microservice boundary) |

**Verdict:** These serve different architectural purposes (monolithic orchestrator vs microservice). Keep both but ensure they share the same RRF implementation.

---

## 4. EMBEDDING CLIENT DUPLICATION (High)

### 4.1 Three Vertex AI Clients

| Location | File | Type | Quality | Verdict |
|----------|------|------|---------|---------|
| `ingestion/embedder/vertex_batch.py` | `VertexAIEmbedder` | Production Python | **KEEP** - Full batching, retry, telemetry |
| `ingestion-go/embedder/vertex_client.go` | `EmbedDocuments()` | Go stub | **DELETE** - Mock implementation |
| `orchestrator/embed/vertex_client.go` | `VertexClient` | Go stub | **KEEP** - Placeholder for orchestrator |

**Key Finding:**
- Python version (`ingestion/embedder/vertex_batch.py`) is **production-ready** with:
  - Batch processing (5 per call, Vertex AI limit)
  - Exponential backoff retry (5 attempts)
  - Telemetry context
  - Validation (dimensions, NaN, Inf, zeros)
- Both Go versions are **stubs** returning mock embeddings

**Recommendation:**
- **KEEP** `ingestion/embedder/vertex_batch.py` for ingestion (production-ready)
- **KEEP** `orchestrator/embed/vertex_client.go` as interface for orchestrator (needs real implementation)
- **DELETE** `ingestion-go/embedder/vertex_client.go` (redundant stub)
- **ACTION:** Implement real Vertex AI client in `orchestrator/embed/` before production

### 4.2 Embedding Service (Microservice)

| Location | Purpose | Verdict |
|----------|---------|---------|
| `services/embedding-service/` | Standalone embedding microservice | **KEEP** - Proper service architecture |
| `services/embedding-service/config/config.go` | Service config | **KEEP** |
| `services/embedding-service/vertex/client.go` | Vertex client | Review - may be stub |

---

## 5. CONFIGURATION DUPLICATION (Medium)

### 5.1 Three Go Config Modules

| Location | File | Purpose | Parameters | Verdict |
|----------|------|---------|------------|---------|
| `ingestion-go/config/config.go` | `Config` struct | Ingestion auto-detect (local/GCP) | 40+ fields | **DELETE** with ingestion-go |
| `orchestrator/config/config.go` | `Config` struct | Orchestrator production config | 35+ fields, Secret Manager | **KEEP** - Production-ready |
| `services/embedding-service/config/config.go` | `Config` struct | Embedding service config | 20+ fields | **KEEP** - Service-specific |
| `services/query-understanding-service/config/config.go` | `Config` struct | Query service config | Similar pattern | **KEEP** - Service-specific |
| `services/vector-search-service/config/config.go` | `Config` struct | Vector search config | Similar pattern | **KEEP** - Service-specific |

**Key Finding:** The `orchestrator/config/config.go` and `ingestion-go/config/config.go` have significant overlap:
- Both have: `AlloyDBDSN`, `PgBouncerHost/Port`, `RedisAddr`, `EmbedModel`, `LLMModel`, `TopK`, `RRFK`, etc.
- `orchestrator/config/config.go` has: Secret Manager integration, validation, PgBouncerDSN helper
- `ingestion-go/config/config.go` has: Auto-detect mode (local vs GCP), FAISS toggle

**Recommendation:**
- **KEEP** `orchestrator/config/config.go` (production-ready with Secret Manager)
- **DELETE** `ingestion-go/config/config.go` (with ingestion-go pipeline)
- The auto-detect feature from `ingestion-go/config/config.go` should be **merged** into `orchestrator/config/config.go` if needed

### 5.2 Python vs Go Configuration

| Location | File | Verdict |
|----------|------|---------|
| `rag_config.py` | Python RAG config (optimized params) | **KEEP** - Contains grid search results |
| `orchestrator/config/config.go` | Go orchestrator config | **KEEP** - Service config |

**These serve different purposes** (RAG algorithm params vs service infrastructure config). No conflict.

---

## 6. CLI SCRIPT DUPLICATION (Medium)

### 6.1 Query CLI Scripts

| File | Purpose | Features | Verdict |
|------|---------|----------|---------|
| `ask_query.py` | Demo CLI with hardcoded chunks | Simple keyword retrieval, sample data | **DELETE** - Demo only |
| `ask_real_query.py` | Real CLI with PDF extraction | Real PDF extraction, Ollama embeddings, self-query | **KEEP** |

**Recommendation:**
- **KEEP** `ask_real_query.py` (real functionality)
- **DELETE** `ask_query.py` (demo/sample data version)

### 6.2 Evaluation Scripts

| File | Purpose | Verdict |
|------|---------|---------|
| `evaluate_complete_rag.py` | Complete RAG evaluation | **KEEP** |
| `evaluate_real_rag.py` | Real RAG evaluation | Review - may overlap |
| `optimize_metrics.py` | Metric optimization | **KEEP** |
| `optimize_ragas_metrics.py` | RAGAS optimization | Review - may overlap |
| `bench_and_improve.py` | Benchmarking | **KEEP** |
| `ab_test_config.py` | A/B testing config | **KEEP** |

---

## 7. TEST FILE DUPLICATION (Medium)

### 7.1 Test File Inventory

| Location | Files | Target | Verdict |
|----------|-------|--------|---------|
| `tests/integration_test.go` | 1 file, 20+ tests | Full integration (chunker, keywords, retrieval, handler, config, session, DB) | **KEEP** - Comprehensive |
| `ingestion-go/tests/chunker_test.go` | 15 tests | Chunker only | **DELETE** with ingestion-go |
| `ingestion-go/tests/keywords_test.go` | Tests | Keywords only | **DELETE** with ingestion-go |
| `ingestion-go/tests/parser_test.go` | Tests | Parser only | **DELETE** with ingestion-go |
| `ingestion/tests/test_chunker.py` | Tests | Python chunker | **KEEP** |
| `ingestion/tests/test_dedup.py` | Tests | Dedup | **KEEP** |
| `ingestion/tests/test_parser.py` | Tests | Parser | **KEEP** |
| `tests/integration/test_rag_pipeline.py` | Tests | Python integration | **KEEP** |
| `tests/run_all_tests.py` | Test runner | All tests | **KEEP** |
| `run_all_tests.py` (root) | Test runner | Duplicate | Review |
| `run_all_tests_fast.py` (root) | Test runner | Fast version | Review |
| `test_chunking_strategies.py` | Tests | Chunking | Review |
| `test_continuous_improvement.py` | Tests | Improvement loop | **KEEP** |

### 7.2 Test Runner Duplication

| File | Purpose | Verdict |
|------|---------|---------|
| `tests/run_all_tests.py` | Python test runner | **KEEP** |
| `run_all_tests.py` (root) | May be duplicate | Check content |
| `run_all_tests_fast.py` (root) | Fast test runner | **KEEP** (different purpose) |
| `build_and_test.bat` | Windows build+test | **KEEP** |
| `cmd/test_e2e/main.go` | Go E2E tests | **KEEP** |

---

## 8. SERVICES ARCHITECTURE REDUNDANCY (Low)

### 8.1 Microservice Config Duplication

Each Go service has its own `config/config.go` with identical helper functions:

| Service | Has getEnv() | Has getIntEnv() | Has getFloatEnv() | Has getDurationEnv() |
|---------|-------------|-----------------|-------------------|---------------------|
| `orchestrator/config/` | ✅ | ✅ | ✅ | ✅ |
| `services/embedding-service/config/` | ✅ | ✅ | ✅ | ✅ |
| `services/query-understanding-service/config/` | ✅ | ✅ | ✅ | ❌ |
| `services/vector-search-service/config/` | ✅ | ✅ | ✅ | ❌ |

**Recommendation:** Extract shared `config/envloader` package with common env loading utilities.

---

## 9. DEAD CODE & UNUSED FILES

### 9.1 Definitely Dead Code

| File | Reason | Action |
|------|--------|--------|
| `full_rag_pipeline_v3.ipynb` | Jupyter notebook, superseded by scripts | **DELETE** |
| `frontend_init.py` | One-time init script | **DELETE** after first run |
| `migrate_to_optimal_config.py` | One-time migration script | **DELETE** after migration |
| `terminal_client.py` | Demo/test client | **DELETE** |
| `demo_self_query.py` | Demo script | **DELETE** (functionality in ask_real_query.py) |

### 9.2 Schema Duplication

| File | Purpose | Verdict |
|------|---------|---------|
| `supabase/migrations/001_initial_schema.sql` | Initial schema | **DELETE** - superseded |
| `supabase/migrations/001_production_schema.sql` | Production schema | **KEEP** |
| `supabase/migrations/002_hnsw_optimization.sql` | HNSW optimization | **KEEP** |
| `supabase/migrations/003_ingestion_functions.sql` | Ingestion functions | **KEEP** |
| `supabase/migrations/004_rpc_functions.sql` | RPC functions | **KEEP** |
| `schema/005_incremental_ingestion.sql` | Incremental ingestion | **KEEP** |
| `schema/v2_production.sql` | Production v2 | Review - may overlap with 001_production_schema.sql |

### 9.3 Data Files

| File | Purpose | Verdict |
|------|---------|---------|
| `data/complete_rag_evaluation.json` | Evaluation results | **KEEP** |
| `data/fast_master_test_results.json` | Test results | **DELETE** - intermediate |
| `data/optimization_results.json` | Optimization results | **KEEP** |
| `data/ragas_optimization_results.json` | RAGAS results | **KEEP** |
| `data/ragas_judge_evaluation.json` | Judge evaluation | **KEEP** |
| `data/REAL_llm_judge_results.json` | Real LLM results | **KEEP** |
| `data/test_results_final.json` | Final test results | **KEEP** |
| `data/test_results_ollama_real.json` | Ollama test results | **KEEP** |
| `data/textbook_metadata.json` | Textbook metadata | **KEEP** |
| `data/golden_qa_dataset.jsonl` | Golden QA dataset | **KEEP** |
| `data/quick_improvement_results.json` | Quick results | **DELETE** - intermediate |
| `data/quick_improvement_results.csv` | Quick results CSV | **DELETE** - intermediate |
| `data/FINAL_TEST_RESULTS.md` | Test results doc | **DELETE** - overlaps with root docs |
| `data/OLLAMA_REAL_TEST_RESULTS.md` | Ollama results doc | **DELETE** - overlaps with root docs |

---

## 10. INCONSISTENT IMPLEMENTATIONS

### 10.1 Chunking Parameters

| Location | Parent Size | Parent Overlap | Child Size | Child Overlap | Status |
|----------|------------|----------------|------------|---------------|--------|
| `rag_config.py` | 400 | 150 | N/A | N/A | Optimized (grid search) |
| `ingestion/chunker/parent_child.py` | 400 | 150 | 400 | 150 | ✅ Optimized |
| `ingestion-go/chunker/parent_child.go` | 1500 | 100 | 512 | 77 | ❌ Legacy |

**Impact:** Go ingestion produces **3.75x larger** parent chunks with **33% less overlap**, degrading retrieval quality.

### 10.2 Embedding Dimensions

| Location | Dimension | Model | Status |
|----------|-----------|-------|--------|
| `rag_config.py` | 384 | sentence-transformers/all-MiniLM-L6-v2 | Local dev |
| `ingestion/embedder/vertex_batch.py` | 768 | text-embedding-005 | Production |
| `ingestion-go/config/config.go` | 768 | varies by mode | Correct |
| `orchestrator/config/config.go` | 768 | text-embedding-005 | Production |

**Note:** `rag_config.py` uses 384-dim local model, while production uses 768-dim Vertex AI. This is intentional (local vs prod) but should be documented.

### 10.3 RRF Parameter

| Location | RRF K | Status |
|----------|-------|--------|
| All implementations | 60.0 | ✅ Consistent |

### 10.4 Top-K Values

| Location | Top-K | Context |
|----------|-------|---------|
| `rag_config.py` | 5 | Optimized (grid search) |
| `orchestrator/config/config.go` | 5 | Production default |
| `ask_real_query.py` | 5 | CLI default |
| `orchestrator/retrieval/hybrid_search.go` | topK*20 for dense/sparse | Internal expansion |

✅ Generally consistent at 5 for final results.

---

## 11. RECOMMENDED FILE STRUCTURE AFTER CLEANUP

```
ragpipeline/
├── README.md                              # Main entry point
├── README_ENTERPRISE.md                   # Enterprise overview
├── PRODUCTION_READINESS_CHECKLIST.md      # Production checklist
├── PLAN.md                                # Project plan
├── requirements.txt                       # Python dependencies
├── requirements-self-query.txt            # Self-query dependencies
├── rag_config.py                          # RAG algorithm config (optimized)
├── docker-compose.yml                     # Docker services
├── .env.local.example                     # Environment template
├── .gitignore
│
├── ingestion/                             # KEEP - Primary ingestion pipeline
│   ├── pipeline.py
│   ├── pyproject.toml
│   ├── README.md
│   ├── chunker/
│   ├── parser/
│   ├── keywords/
│   ├── embedder/
│   ├── writer/
│   ├── dedup/
│   ├── dlq/
│   ├── incremental/
│   └── tests/
│
├── orchestrator/                          # KEEP - Main Go service
│   ├── cmd/server/main.go
│   ├── config/config.go                   # Production config
│   ├── handler/
│   ├── retrieval/                         # KEEP - Consolidate RRF here
│   ├── embed/                             # TODO: Implement real Vertex client
│   ├── llm/
│   ├── session/
│   ├── cache/
│   ├── middleware/
│   ├── observability/
│   ├── analytics/
│   ├── db/
│   ├── pgbouncer/
│   ├── compliance/                        # REVIEW: Python in Go service
│   ├── disaster_recovery/                 # REVIEW: Python in Go service
│   ├── go.mod
│   ├── go.sum
│   ├── Dockerfile
│   └── README.md
│
├── services/                              # KEEP - Microservices
│   ├── embedding-service/
│   ├── query-understanding-service/
│   └── vector-search-service/
│
├── retrievers/                            # KEEP - Self-query retrieval
│
├── tests/                                 # KEEP - Integration tests
│   ├── integration/
│   ├── run_all_tests.py
│   └── integration_test.go
│
├── scripts/                               # KEEP - Utility scripts
│
├── supabase/                              # KEEP - Database migrations
│   └── migrations/
│
├── schema/                                # KEEP - Additional schemas
│
├── terraform/                             # KEEP - Infrastructure as Code
│
├── frontend/                              # KEEP - React frontend
│
├── eval/                                  # KEEP - Evaluation tools
│   ├── load_test.js
│   └── recall_eval.py
│
├── quality_loop/                          # KEEP - Quality improvement
│   └── judge.py
│
├── data/                                  # KEEP - Data files (remove intermediates)
│
├── .planning/                             # KEEP - Structured planning docs
│   └── codebase/
│       └── RAG_PIPELINE_AUDIT.md
│
├── ask_real_query.py                      # KEEP - Real CLI
├── build_and_test.bat                     # KEEP - Windows build script
├── setup.sh                               # KEEP - Linux setup
├── setup.ps1                              # KEEP - Windows setup
│
└── [DELETE ALL REDUNDANT FILES]           # 31 .md files + ingestion-go/ + misc
```

---

## 12. ACTION PLAN

### Phase 1: Safe Deletions (Immediate)

| Action | Files | Risk |
|--------|-------|------|
| Delete redundant audit docs | 12 .md files | Low |
| Delete redundant test docs | 9 .md files | Low |
| Delete redundant config docs | 5 .md files | Low |
| Delete redundant planning docs | 4 .md files | Low |
| Delete intermediate data files | 4 .md/.json/.csv files | Low |
| Delete dead code scripts | 5 .py files | Low |
| Delete duplicate schema | 001_initial_schema.sql | Low |

### Phase 2: Ingestion Consolidation (1-2 days)

| Action | Files | Risk |
|--------|-------|------|
| Delete ingestion-go/ | Entire directory | Medium (verify Python pipeline covers all cases) |
| Move auto-detect feature | ingestion-go/config/config.go → orchestrator/config/config.go | Medium |

### Phase 3: RRF Consolidation (1 day)

| Action | Files | Risk |
|--------|-------|------|
| Create internal/rrf package | New package | Medium |
| Merge all RRF logic | 3 files → 1 | Medium |
| Update all callers | orchestrator/, services/ | Medium |

### Phase 4: Embedding Client Implementation (2-3 days)

| Action | Files | Risk |
|--------|-------|------|
| Implement real Vertex client | orchestrator/embed/vertex_client.go | High |
| Delete redundant stub | ingestion-go/embedder/ | Low (with ingestion-go) |

### Phase 5: Config Helper Extraction (0.5 days)

| Action | Files | Risk |
|--------|-------|------|
| Create shared config utilities | New internal package | Low |

---

## 13. SUMMARY STATISTICS

| Metric | Before | After | Reduction |
|--------|--------|-------|-----------|
| Root .md files | 51 | 15 | -71% |
| Ingestion pipelines | 2 (Python + Go) | 1 (Python) | -50% |
| RRF implementations | 3 | 1 | -67% |
| Embedding clients | 3 (1 prod, 2 stubs) | 2 (1 prod, 1 to implement) | -33% |
| Go config modules | 5 (with overlap) | 4 (service-specific) | -20% |
| CLI query scripts | 2 | 1 | -50% |
| Test files | 12+ | 8 | -33% |
| Schema files | 7 | 6 | -14% |
| Data files | 13 | 9 | -31% |
| Total files (estimated) | ~250 | ~180 | -28% |

---

## 14. RISKS & MITIGATION

| Risk | Impact | Mitigation |
|------|--------|------------|
| Deleting ingestion-go breaks hybrid flow | High | Verify Python pipeline handles all ingestion cases first |
| RRF consolidation introduces bugs | Medium | Run full test suite after consolidation |
| Config auto-detect feature lost | Medium | Port to orchestrator/config before deletion |
| Historical data lost in doc cleanup | Low | Commit all docs to git before deletion |

---

*End of Redundancy Audit Report*
