# Complete Documentation Map & Backlink Analysis

**Generated:** 2026-04-06
**Scope:** ALL `.md` files across docs/, .planning/, root, services/, supabase/, frontend/
**Total Files Found:** 122 markdown files

---

## 1. COMPLETE FILE INVENTORY

### Root Level (2 files)

| File | Title/Main Topic | Key Sections |
|------|-----------------|--------------|
| `README.md` | Visionary RAG Pipeline - Main Overview | System Architecture, Services Overview, CAG, Circuit Breaker, Hybrid Search, Quick Start, Repo Structure, Testing, Performance, Security, GCP Deployment, Go Code Quality, Contributing |
| `CONTRIBUTING.md` | Contributing Guidelines | Code of Conduct, Prerequisites, Dev Setup, Architecture, Workflow, Testing, PR Process, Project Structure |

### .planning/ (5 files)

| File | Title/Main Topic | Key Sections |
|------|-----------------|--------------|
| `.planning/codebase/RAG_PIPELINE_AUDIT.md` | Enterprise Production Readiness Audit | Core Pipeline Architecture, File Organization, Dependencies, DB Schema, Testing Coverage, Deployment, Observability, Security, Error Handling |
| `.planning/codebase/INFRASTRUCTURE_AUDIT.md` | Complete Infrastructure Audit | GCP APIs, Resources, Local Services, Env Vars, Service Startup Order, E2E Testing, Architecture Topology, Cost Estimates, Checklists |
| `.planning/codebase/GO_PERFORMANCE_TESTING_REVIEW.md` | Go Performance & Testing Review | Critical Issues (string concat, byte conversions, linting, testify), High Priority (capacity hints, bubble sort, unsafe assertions), Recommendations (benchmarks, goleak, race detection) |
| `.planning/research/SUMMARY.md` | Production RAG System on GCP Research | AlloyDB Config, Memorystore Redis, Cloud Run, pgvector/ScaNN, Vertex AI Embeddings, Go Libraries, Python Libraries, Architecture Validation |
| `.planning/research/GO_REQUIREMENTS_SUMMARY.md` | Go Implementation Requirements Summary | Current Go Plans, Python-to-Go Migration Feasibility, Go Libraries/Patterns, Architecture Decisions, Migration Path |

### docs/ Root (1 file)

| File | Title/Main Topic | Key Sections |
|------|-----------------|--------------|
| `docs/COMPLETE_DOCUMENTATION.md` | Complete Technical Reference | System Architecture, Microservices Overview, Core Patterns, Repo Structure, Go Modules, Testing Strategy, Performance Optimizations, Security, Configuration, Deployment, API Reference, DB Schema, Monitoring, Troubleshooting |

### docs/architecture/ (3 files)

| File | Title/Main Topic | Key Sections |
|------|-----------------|--------------|
| `docs/architecture/MICROSERVICES_COMPLETE.md` | Microservices Implementation Complete | Executive Summary, Complete Architecture, Service Details (API Gateway, Embedding, Vector Search, Query Understanding), How to Run, API Usage Examples, Production Checklist, Next Steps |
| `docs/architecture/MICROSERVICES_IMPLEMENTATION_PLAN.md` | Microservices Implementation Assessment & Plan | Current State Assessment, Implementation Strategy, Phases 1-4, Architecture Diagram, Cost Analysis, Success Criteria |
| `docs/architecture/MICROSERVICES_STATUS.md` | Microservices Architecture Status | Documentation Cleanup, Data Organization, Microservices Created, Architecture Overview, Service Details, Port Assignments, Benefits Achieved |

### docs/archive/ (52 files)

| File | Title/Main Topic | Key Sections |
|------|-----------------|--------------|
| `docs/archive/ADVANCED_TESTING_SUMMARY.md` | Advanced Testing Summary | Test results and summaries |
| `docs/archive/AUDIT_AND_IMPLEMENTATION_SUMMARY.md` | Audit & Implementation Summary | Combined audit findings and implementation status |
| `docs/archive/AUDIT_FINDINGS.md` | Initial Audit Findings | Code quality issues, architectural concerns |
| `docs/archive/AUDIT_FIX_PROGRESS.md` | Audit Fix Progress | Tracking of fixes applied |
| `docs/archive/AUDIT_IMPLEMENTATION_SUMMARY.md` | Audit Implementation Summary | Implementation of audit recommendations |
| `docs/archive/AUTO_DETECT_SETUP.md` | Auto-Detect Setup Guide | Automatic environment configuration |
| `docs/archive/BUILD_AND_TEST_STATUS.md` | Build & Test Status | Build and test results |
| `docs/archive/COMPLETE_AUDIT_FIX_SUMMARY.md` | Complete Audit Fix Summary | Comprehensive fix summary |
| `docs/archive/COMPLETE_RAG_EVALUATION_SUMMARY.md` | Complete RAG Evaluation Summary | Full RAG evaluation results |
| `docs/archive/COMPREHENSIVE_CODE_AUDIT.md` | Comprehensive Code Audit | Full codebase audit |
| `docs/archive/CONFIG_IMPLEMENTATION_SUMMARY.md` | Config Implementation Summary | Configuration changes |
| `docs/archive/CONFIG_QUICK_REFERENCE.md` | Config Quick Reference | Quick reference for configuration |
| `docs/archive/CREDENTIALS_REQUIRED.md` | Credentials Required | Required credentials list |
| `docs/archive/CRITICAL_FIXES.md` | Critical Fixes | Critical bug fixes applied |
| `docs/archive/DEFENSIVE_FIXES.md` | Defensive Fixes | Defensive programming fixes |
| `docs/archive/DEVELOPMENT_SESSION_SUMMARY.md` | Development Session Summary | Session notes |
| `docs/archive/ENTERPRISE_AUDIT_COMPLETE.md` | Enterprise Audit Complete | Enterprise-grade audit results |
| `docs/archive/ENTERPRISE_HARDENING_PLAN.md` | Enterprise Hardening Plan | Hardening recommendations |
| `docs/archive/ENTERPRISE_PRODUCTION_AUDIT.md` | Enterprise Production Audit | Production readiness audit |
| `docs/archive/ENTERPRISE_PRODUCTION_ROADMAP.md` | Enterprise Production Roadmap | Roadmap for production |
| `docs/archive/FINAL_IMPLEMENTATION_COMPLETE.md` | Final Implementation Complete | Final implementation status |
| `docs/archive/FINAL_IMPLEMENTATION_SUMMARY.md` | Final Implementation Summary | Implementation summary |
| `docs/archive/FINAL_TEST_REPORT.md` | Final Test Report | Test results |
| `docs/archive/FINAL_TEST_SUMMARY.md` | Final Test Summary | Test summary |
| `docs/archive/FUNCTIONAL_OPTIONS_FIXES.md` | Functional Options Fixes | Functional options pattern fixes |
| `docs/archive/GAP_ANALYSIS.md` | Gap Analysis | Feature gap analysis |
| `docs/archive/GO_API_DESIGN_FIXES.md` | Go API Design Fixes | API design fixes |
| `docs/archive/GO_FIRST_SUMMARY.md` | Go-First Approach Summary | Go-first strategy summary |
| `docs/archive/GO_INGESTION_IMPLEMENTATION_PLAN.md` | Go Ingestion Implementation Plan | Go ingestion pipeline plan |
| `docs/archive/HOW_TO_RUN_OLLAMA_TEST.md` | How to Run Ollama Test | Ollama testing guide |
| `docs/archive/IMPLEMENTATION_PROGRESS.md` | Implementation Progress | Progress tracking |
| `docs/archive/IMPLEMENTATION_SUMMARY.md` | Implementation Summary | Implementation overview |
| `docs/archive/IMPROVEMENT_LOOP_RESULTS.md` | Improvement Loop Results | Results from improvement cycles |
| `docs/archive/LANGCHAINGO_INTEGRATION.md` | LangChainGo Integration | LangChainGo integration guide |
| `docs/archive/LANGCHAINGO_USAGE.md` | LangChainGo Usage | LangChainGo usage patterns |
| `docs/archive/LOCAL_DEVELOPMENT_SETUP.md` | Local Development Setup | Local setup instructions |
| `docs/archive/METRIC_IMPROVEMENTS.md` | Metric Improvements | Performance metric improvements |
| `docs/archive/OPTIMAL_CONFIG_QUICKSTART.md` | Optimal Config Quickstart | Quick start with optimal config |
| `docs/archive/PHASE_1_MVP_BACKLOG.md` | Phase 1 MVP Backlog | MVP backlog |
| `docs/archive/PHASE1_DEPLOYMENT.md` | Phase 1 Deployment | Phase 1 deployment guide |
| `docs/archive/PLAN.md` | Implementation Plan | Active development plan |
| `docs/archive/PRODUCTION_READINESS_CHECKLIST.md` | Production Readiness Checklist | Production checklist |
| `docs/archive/QUICKSTART_LOCAL.md` | Quickstart Local | Quick start guide |
| `docs/archive/RAG_EVALUATION_METRICS.md` | RAG Evaluation Metrics | Evaluation framework |
| `docs/archive/RAG_PIPELINE_AUDIT_REPORT.md` | RAG Pipeline Audit Report | Audit report |
| `docs/archive/RAGAS_OPTIMIZATION_COMPLETE.md` | RAGAS Optimization Complete | RAGAS optimization results |
| `docs/archive/README_ENTERPRISE.md` | Enterprise README | Enterprise overview |
| `docs/archive/SELF_QUERY_FUNCTIONAL.md` | Self-Query Functional | Self-query feature status |
| `docs/archive/SELF_QUERY_IMPLEMENTATION_PLAN.md` | Self-Query Implementation Plan | Self-query plan |
| `docs/archive/SELF_QUERY_RESULTS.md` | Self-Query Results | Self-query test results |
| `docs/archive/STANDUP_PROGRESS.md` | Standup Progress | Progress logs |
| `docs/archive/TEST_REPORT_SCIENCE_CLASS8.md` | Test Report Science Class 8 | Class 8 science test results |
| `docs/archive/TEST_REPORT.md` | Test Report | General test results |
| `docs/archive/USER_STORY_ALIGNMENT.md` | User Story Alignment | Requirements traceability |
| `docs/archive/visionary_rag_implementation_plan.md` | Visionary RAG Implementation Plan | Historical implementation plan |

### docs/audits/ (16 files)

| File | Title/Main Topic | Key Sections |
|------|-----------------|--------------|
| `docs/audits/audit_advanced_retrieval_strategies.md` | Audit: Advanced Retrieval Strategies | Ailog guide audit findings |
| `docs/audits/audit_chunking_strategies.md` | Audit: Chunking Strategies | Chunking audit findings |
| `docs/audits/audit_cost_optimization.md` | Audit: Cost Optimization | Cost optimization audit |
| `docs/audits/audit_embedding_models.md` | Audit: Embedding Models | Embedding model audit |
| `docs/audits/audit_langchaingo_advanced_retrieval.md` | Audit: LangChainGo Advanced Retrieval | LangChainGo retrieval capabilities |
| `docs/audits/audit_langchaingo_chunking.md` | Audit: LangChainGo Chunking | LangChainGo chunking analysis |
| `docs/audits/audit_langchaingo_latency.md` | Audit: LangChainGo Latency | LangChainGo latency analysis |
| `docs/audits/audit_langchaingo_query_expansion.md` | Audit: LangChainGo Query Expansion | LangChainGo query expansion |
| `docs/audits/audit_langchaingo_reranking.md` | Audit: LangChainGo Reranking | LangChainGo reranking analysis |
| `docs/audits/audit_langchaingo_semantic_chunking.md` | Audit: LangChainGo Semantic Chunking | LangChainGo semantic chunking |
| `docs/audits/audit_monitoring.md` | Audit: Monitoring | Monitoring audit findings |
| `docs/audits/audit_query_expansion.md` | Audit: Query Expansion | Query expansion audit |
| `docs/audits/audit_rag_latency.md` | Audit: RAG Latency | RAG latency audit |
| `docs/audits/audit_reranking.md` | Audit: Reranking | Reranking audit |
| `docs/audits/audit_retrieval_fundamentals.md` | Audit: Retrieval Fundamentals | Retrieval fundamentals audit |
| `docs/audits/audit_vector_databases.md` | Audit: Vector Databases | Vector DB audit |
| `docs/audits/LANGCHAINGO_AUDIT.md` | LangChainGo Audit & Integration Plan | LangChainGo capabilities vs current implementation, migration priority matrix, phased implementation plan |
| `docs/audits/LANGCHAINGO_VS_AILOG_GUIDES_AUDIT.md` | LangChainGo vs Ailog Guides | LangChainGo capability audit across 11 Ailog guides, gap analysis, code examples, verdicts per guide |
| `docs/audits/REDUNDANCY_AUDIT.md` | Redundancy Audit Report | Documentation redundancy, ingestion pipeline duplication, RRF implementations, embedding clients, config duplication, CLI scripts, test files, dead code, inconsistent implementations |

### docs/caching/ (2 files)

| File | Title/Main Topic | Key Sections |
|------|-----------------|--------------|
| `docs/caching/CAG_ARCHITECTURE.md` | Cache Augmented Generation Architecture | Cost Centers, 5-Layer CAG Architecture (Exact Match, Semantic, Embedding, Search Results, Template), Cache Invalidation, Metrics, Configuration, Expected Impact, Redis Architecture |
| `docs/caching/CAG_IMPLEMENTATION_COMPLETE.md` | CAG Implementation Complete | Cost Impact (before/after), 5-Layer Details, CAG Orchestrator, Cache Statistics Endpoint, Configuration, Redis Architecture, Cache Invalidation, Error Handling, Performance Impact, Files Created/Modified |

### docs/guides/ (11 files)

| File | Title/Main Topic | Key Sections |
|------|-----------------|--------------|
| `docs/guides/guide_advanced_retrieval_strategies.md` | Guide: Advanced Retrieval Strategies | Hybrid search, HyDE, MMR, contextual compression, self-query, query decomposition |
| `docs/guides/guide_chunking_strategies.md` | Guide: Chunking Strategies | Recursive splitting, token splitting, semantic chunking, document-structure splitting |
| `docs/guides/guide_cost_optimization.md` | Guide: Cost Optimization | Free vs paid embeddings, LLM cost management, caching, batch operations, rate limiting |
| `docs/guides/guide_embedding_models.md` | Guide: Embedding Models | Model selection, MTEB leaderboard, fine-tuning, benchmarking, cost-performance tradeoffs |
| `docs/guides/guide_monitoring.md` | Guide: Monitoring | Key metrics, instrumentation, dashboards, alerting |
| `docs/guides/guide_query_expansion.md` | Guide: Query Expansion | Synonym expansion, LLM rewriting, multi-query retrieval, HyDE, step-back prompting |
| `docs/guides/guide_rag_latency.md` | Guide: RAG Latency | Streaming, parallel retrieval, HNSW tuning, caching, TTFT reduction |
| `docs/guides/guide_reranking.md` | Guide: Reranking | Cross-encoder reranking, Cohere API, LLM-based reranking, two-stage retrieval |
| `docs/guides/guide_retrieval_fundamentals.md` | Guide: Retrieval Fundamentals | Embeddings, chunking, vector databases, similarity metrics, evaluation metrics (Recall@K, MRR, NDCG) |
| `docs/guides/guide_semantic_chunking.md` | Guide: Semantic Chunking | Embedding-based chunking, topic boundary detection, section-aware splitting |
| `docs/guides/guide_vector_databases.md` | Guide: Vector Databases | Pinecone vs Qdrant vs Weaviate vs Milvus vs Chroma vs pgvector comparison, indexing strategies (HNSW, IVF, FLAT), metadata filtering, distance metrics, cost optimization |

### docs/migration/ (6 files)

| File | Title/Main Topic | Key Sections |
|------|-----------------|--------------|
| `docs/migration/LANGCHAINGO_MIGRATION_COMPLETE.md` | LangChainGo Migration Complete | LLM Client with Streaming, Output Parsers, Prompt Templates, API Endpoints, Integration Guide, Performance Improvements, What Was NOT Changed |
| `docs/migration/MIGRATION_GUIDE.md` | Codebase Reorganization Guide | Documentation Cleanup, Data Organization, Microservices Created, Directory Structure, Next Steps, Service Communication, Port Assignments, Migration Checklist |
| `docs/migration/PYTHON_TO_GO_BUILD_VERIFICATION.md` | Python-to-Go Build Verification | Build results for all converted components, issues found & fixed, module structure, dependency analysis, integration guide |
| `docs/migration/PYTHON_TO_GO_CONVERSION_PLAN.md` | Python-to-Go Conversion Plan | Library Compatibility Analysis, Conversion Priority Matrix, 7-Phase Detailed Plan, Components That Should Stay Python, Post-Conversion Architecture |
| `docs/migration/PYTHON_TO_GO_CONVERSION_STATUS.md` | Python-to-Go Conversion Status Report | 5 Components Converted (Config, A/B Test, Evaluation, LLM Judge, Shared Utils), Library Compatibility Analysis, Testing Status, Migration Guide |
| `docs/migration/RERANKING_GO_IMPLEMENTATION.md` | Go-Based Cross-Encoder Reranking | 4 Options (Vertex AI Endpoint, Cloud Run + ONNX, Vertex AI Embeddings, Ollama), Comparison Matrix, Dockerfile, Deployment Steps, Architecture Diagram, Expected Impact |

### docs/quality/ (1 file)

| File | Title/Main Topic | Key Sections |
|------|-----------------|--------------|
| `docs/quality/CODE_QUALITY_IMPROVEMENTS.md` | Code Quality Improvements Summary | Retry Logic, Data Race Fix, Prompt Injection Prevention, Rate Limiting, History Validation, Concurrent Request Limiting, Circuit Breaker, Request ID Tracing, Graceful Shutdown |

### docs/reorganization/ (1 file)

| File | Title/Main Topic | Key Sections |
|------|-----------------|--------------|
| `docs/reorganization/GO_PROJECT_REORGANIZATION.md` | Go Project Reorganization Plan | Current State Assessment (4/10 maturity), Target Directory Structure, File Mapping, Unified Import Paths, go.work Workspace, Makefile |

### data/ (2 files)

| File | Title/Main Topic | Key Sections |
|------|-----------------|--------------|
| `data/FINAL_TEST_RESULTS.md` | Final Test Results | Ingestion Results (265 pages, 1590 chunks), Query Results (5 queries), Performance Metrics, 100% Pass |
| `data/OLLAMA_REAL_TEST_RESULTS.md` | Real Ollama Test Results | Ingestion (50 chunks, 768 dim), Query Results (3 queries, ~12s latency), 100% Pass |

### services/ (4 README files)

| File | Title/Main Topic | Key Sections |
|------|-----------------|--------------|
| `services/api-gateway/README.md` | API Gateway Documentation | Architecture, Features (Orchestration, SSE, Caching, Sessions, Circuit Breakers), API Endpoints, Configuration, Error Codes, Middleware Chain, Project Structure |
| `services/embedding-service/README.md` | Embedding Service Documentation | Features, API Endpoints (/embed, /health, /stats, /version), Error Codes, Environment Variables, Testing, Deployment, Architecture |
| `services/query-understanding-service/README.md` | Query Understanding Service Documentation | API Endpoints (/rewrite, /parse-filters, /classify-intent), Architecture, Configuration, Fallback Behavior, Prompt Versioning, Intent Categories, Filter Schema |
| `services/vector-search-service/README.md` | Vector Search Service Documentation | Hybrid Search Architecture, API Reference (/search, /search/dense, /search/sparse), RRF Algorithm, Database Schema Requirements, Error Handling |

### supabase/ (1 file)

| File | Title/Main Topic | Key Sections |
|------|-----------------|--------------|
| `supabase/README.md` | Production-Grade Vector Database Schema | Key Features, Migration Files, Deployment, Schema Overview (9 tables), RPC Functions, Python Client Example, Indexes, RLS, Views & Materialized Views, Maintenance |

### frontend/ (1 file)

| File | Title/Main Topic | Key Sections |
|------|-----------------|--------------|
| `frontend/README.md` | React + TypeScript + Vite Setup | Vite plugins, ESLint configuration, React Compiler notes (standard Vite template README) |

---

## 2. EXISTING BACKLINKS (Cross-References Found)

### README.md references:
- `docs/COMPLETE_DOCUMENTATION.md` ✅
- `docs/architecture/MICROSERVICES_COMPLETE.md` ✅
- `docs/caching/CAG_ARCHITECTURE.md` ✅
- `docs/audits/` ✅
- `docs/guides/` ✅
- `docs/migration/` ✅
- `docs/quality/` ✅

### CONTRIBUTING.md references:
- `docs/architecture/` ✅
- `docs/caching/` ✅
- `README.md` ✅
- Service `README.md` files ✅

### docs/COMPLETE_DOCUMENTATION.md references:
- (Self-contained, few external links)

### docs/architecture/MICROSERVICES_COMPLETE.md references:
- `MICROSERVICES_IMPLEMENTATION_PLAN.md` ✅
- `MICROSERVICES_STATUS.md` ✅
- `MIGRATION_GUIDE.md` ✅
- `CODE_QUALITY_IMPROVEMENTS.md` ✅
- `REDUNDANCY_AUDIT.md` ✅
- `REORGANIZATION_COMPLETE.md` ✅

### docs/caching/CAG_IMPLEMENTATION_COMPLETE.md references:
- `CAG_ARCHITECTURE.md` ✅

### docs/migration/MIGRATION_GUIDE.md references:
- `REDUNDANCY_AUDIT.md` ✅
- `REORGANIZATION_PLAN.md` ✅
- `README.md` ✅
- Service `README.md` files ✅

### docs/migration/PYTHON_TO_GO_CONVERSION_STATUS.md references:
- `PYTHON_TO_GO_CONVERSION_PLAN.md` ✅
- `PYTHON_TO_GO_BUILD_VERIFICATION.md` ✅

### docs/audits/LANGCHAINGO_AUDIT.md references:
- (Self-contained analysis)

### docs/audits/REDUNDANCY_AUDIT.md references:
- `.planning/codebase/RAG_PIPELINE_AUDIT.md` ✅

### supabase/README.md references:
- (Self-contained)

### services/*/README.md files:
- Reference `.env.example` files ✅
- Reference sibling service ports ✅

### .planning/codebase/INFRASTRUCTURE_AUDIT.md references:
- `PLAN.md` ✅
- `schema/v2_production.sql` ✅

---

## 3. MISSING BACKLINKS (Should Reference But Don't)

### CRITICAL Missing Backlinks:

| Source File | Should Link To | Why |
|-------------|---------------|-----|
| `README.md` | `docs/migration/MIGRATION_GUIDE.md` | Main entry point should link to migration guide |
| `README.md` | `docs/audits/REDUNDANCY_AUDIT.md` | Should reference audit findings |
| `README.md` | `supabase/README.md` | Database setup is critical |
| `README.md` | `docs/migration/PYTHON_TO_GO_CONVERSION_PLAN.md` | Architecture decision doc |
| `README.md` | `docs/audits/LANGCHAINGO_AUDIT.md` | Major library decision |
| `README.md` | `data/FINAL_TEST_RESULTS.md` | Proof of working system |
| `docs/COMPLETE_DOCUMENTATION.md` | `docs/caching/CAG_ARCHITECTURE.md` | CAG is core feature |
| `docs/COMPLETE_DOCUMENTATION.md` | `docs/audits/REDUNDANCY_AUDIT.md` | Codebase state |
| `docs/COMPLETE_DOCUMENTATION.md` | `supabase/README.md` | DB schema reference |
| `docs/architecture/MICROSERVICES_COMPLETE.md` | `services/api-gateway/README.md` | Service docs |
| `docs/architecture/MICROSERVICES_COMPLETE.md` | `services/embedding-service/README.md` | Service docs |
| `docs/architecture/MICROSERVICES_COMPLETE.md` | `services/vector-search-service/README.md` | Service docs |
| `docs/architecture/MICROSERVICES_COMPLETE.md` | `services/query-understanding-service/README.md` | Service docs |
| `docs/architecture/MICROSERVICES_IMPLEMENTATION_PLAN.md` | `docs/architecture/MICROSERVICES_STATUS.md` | Status follows plan |
| `docs/architecture/MICROSERVICES_IMPLEMENTATION_PLAN.md` | `docs/migration/MIGRATION_GUIDE.md` | Migration follows plan |
| `docs/caching/CAG_ARCHITECTURE.md` | `docs/caching/CAG_IMPLEMENTATION_COMPLETE.md` | Architecture → Implementation |
| `docs/caching/CAG_IMPLEMENTATION_COMPLETE.md` | `docs/caching/CAG_ARCHITECTURE.md` | Implementation → Architecture |
| `docs/guides/guide_retrieval_fundamentals.md` | `docs/guides/guide_vector_databases.md` | Fundamentals → DB selection |
| `docs/guides/guide_retrieval_fundamentals.md` | `docs/guides/guide_embedding_models.md` | Fundamentals → Embedding selection |
| `docs/guides/guide_retrieval_fundamentals.md` | `docs/guides/guide_advanced_retrieval_strategies.md` | Fundamentals → Advanced strategies |
| `docs/guides/guide_reranking.md` | `docs/migration/RERANKING_GO_IMPLEMENTATION.md` | Guide → Implementation |
| `docs/guides/guide_chunking_strategies.md` | `docs/guides/guide_semantic_chunking.md` | Basic → Advanced chunking |
| `docs/guides/guide_query_expansion.md` | `docs/guides/guide_advanced_retrieval_strategies.md` | Expansion is advanced retrieval |
| `docs/migration/MIGRATION_GUIDE.md` | `docs/audits/REDUNDANCY_AUDIT.md` | Migration triggered by audit |
| `docs/migration/PYTHON_TO_GO_CONVERSION_PLAN.md` | `docs/migration/PYTHON_TO_GO_CONVERSION_STATUS.md` | Plan → Status |
| `docs/migration/PYTHON_TO_GO_CONVERSION_STATUS.md` | `.planning/research/GO_REQUIREMENTS_SUMMARY.md` | Status ties to requirements |
| `docs/audits/LANGCHAINGO_AUDIT.md` | `docs/migration/LANGCHAINGO_MIGRATION_COMPLETE.md` | Audit → Migration |
| `docs/audits/LANGCHAINGO_VS_AILOG_GUIDES_AUDIT.md` | `docs/audits/LANGCHAINGO_AUDIT.md` | Related audits |
| `docs/audits/REDUNDANCY_AUDIT.md` | `docs/reorganization/GO_PROJECT_REORGANIZATION.md` | Audit → Reorganization plan |
| `docs/quality/CODE_QUALITY_IMPROVEMENTS.md` | `docs/audits/REDUNDANCY_AUDIT.md` | Quality follows audit |
| `docs/reorganization/GO_PROJECT_REORGANIZATION.md` | `docs/audits/REDUNDANCY_AUDIT.md` | Reorganization follows audit |
| `docs/reorganization/GO_PROJECT_REORGANIZATION.md` | `docs/migration/MIGRATION_GUIDE.md` | Reorganization → Migration |
| `supabase/README.md` | `README.md` | Supabase links to main project |
| `supabase/README.md` | `schema/v2_production.sql` | Schema reference |
| `services/api-gateway/README.md` | `services/embedding-service/README.md` | Downstream service |
| `services/api-gateway/README.md` | `services/vector-search-service/README.md` | Downstream service |
| `services/api-gateway/README.md` | `services/query-understanding-service/README.md` | Downstream service |
| `services/embedding-service/README.md` | `services/api-gateway/README.md` | Consumer service |
| `services/vector-search-service/README.md` | `services/api-gateway/README.md` | Consumer service |
| `services/query-understanding-service/README.md` | `services/api-gateway/README.md` | Consumer service |
| `.planning/codebase/RAG_PIPELINE_AUDIT.md` | `.planning/codebase/INFRASTRUCTURE_AUDIT.md` | Related audits |
| `.planning/codebase/RAG_PIPELINE_AUDIT.md` | `.planning/codebase/GO_PERFORMANCE_TESTING_REVIEW.md` | Related audits |
| `.planning/codebase/INFRASTRUCTURE_AUDIT.md` | `.planning/research/SUMMARY.md` | Infrastructure ties to research |
| `.planning/research/GO_REQUIREMENTS_SUMMARY.md` | `.planning/research/SUMMARY.md` | Requirements from research |
| `CONTRIBUTING.md` | `docs/guides/` | Should link to guides |
| `CONTRIBUTING.md` | `docs/quality/CODE_QUALITY_IMPROVEMENTS.md` | Quality standards |
| `data/FINAL_TEST_RESULTS.md` | `data/OLLAMA_REAL_TEST_RESULTS.md` | Related test results |

---

## 4. DOCUMENTATION CLUSTER MAP

### Cluster 1: Core Documentation (Active)
```
README.md
├── docs/COMPLETE_DOCUMENTATION.md
├── CONTRIBUTING.md
└── docs/architecture/
    ├── MICROSERVICES_COMPLETE.md
    ├── MICROSERVICES_IMPLEMENTATION_PLAN.md
    └── MICROSERVICES_STATUS.md
```

### Cluster 2: Caching
```
docs/caching/
├── CAG_ARCHITECTURE.md ←→ CAG_IMPLEMENTATION_COMPLETE.md (bidirectional)
```

### Cluster 3: Guides (Ailog Research)
```
docs/guides/
├── guide_retrieval_fundamentals.md (entry point)
├── guide_embedding_models.md
├── guide_vector_databases.md
├── guide_chunking_strategies.md ←→ guide_semantic_chunking.md
├── guide_advanced_retrieval_strategies.md
├── guide_query_expansion.md
├── guide_reranking.md
├── guide_rag_latency.md
├── guide_cost_optimization.md
└── guide_monitoring.md
```

### Cluster 4: Audits
```
docs/audits/
├── REDUNDANCY_AUDIT.md (central)
├── LANGCHAINGO_AUDIT.md
├── LANGCHAINGO_VS_AILOG_GUIDES_AUDIT.md
├── audit_*.md (11 Ailog guide audits)
```

### Cluster 5: Migration
```
docs/migration/
├── MIGRATION_GUIDE.md (central)
├── LANGCHAINGO_MIGRATION_COMPLETE.md
├── PYTHON_TO_GO_CONVERSION_PLAN.md ←→ PYTHON_TO_GO_CONVERSION_STATUS.md
├── PYTHON_TO_GO_BUILD_VERIFICATION.md
└── RERANKING_GO_IMPLEMENTATION.md
```

### Cluster 6: Planning & Research
```
.planning/
├── codebase/
│   ├── RAG_PIPELINE_AUDIT.md
│   ├── INFRASTRUCTURE_AUDIT.md
│   └── GO_PERFORMANCE_TESTING_REVIEW.md
└── research/
    ├── SUMMARY.md
    └── GO_REQUIREMENTS_SUMMARY.md
```

### Cluster 7: Service Documentation
```
services/
├── api-gateway/README.md (orchestrator)
├── embedding-service/README.md
├── vector-search-service/README.md
└── query-understanding-service/README.md
```

### Cluster 8: Infrastructure
```
supabase/README.md
schema/v2_production.sql
terraform/ (no README)
docker-compose.yml
```

### Cluster 9: Archive (52 files - historical, not actively maintained)
```
docs/archive/ (all 52 files)
```

### Cluster 10: Test Results
```
data/
├── FINAL_TEST_RESULTS.md
└── OLLAMA_REAL_TEST_RESULTS.md
```

### Cluster 11: Quality & Reorganization
```
docs/quality/CODE_QUALITY_IMPROVEMENTS.md
docs/reorganization/GO_PROJECT_REORGANIZATION.md
```

---

## 5. RECOMMENDED BACKLINK ADDITIONS (Priority Order)

### P0 - Critical Navigation (README.md)
1. `README.md` → Add link to `supabase/README.md` in "Quick Start" section
2. `README.md` → Add link to `docs/migration/MIGRATION_GUIDE.md` in "Documentation" table
3. `README.md` → Add link to `data/FINAL_TEST_RESULTS.md` in "Current Status" section

### P1 - Architecture Clusters
4. `docs/caching/CAG_ARCHITECTURE.md` → Add link to `CAG_IMPLEMENTATION_COMPLETE.md` in "Next Steps"
5. `docs/caching/CAG_IMPLEMENTATION_COMPLETE.md` → Add link to `CAG_ARCHITECTURE.md` in "Overview"
6. `docs/guides/guide_retrieval_fundamentals.md` → Add "Next Steps" links to all advanced guides
7. `docs/audits/REDUNDANCY_AUDIT.md` → Add link to `docs/reorganization/GO_PROJECT_REORGANIZATION.md`

### P2 - Plan → Status → Verification Chains
8. `docs/migration/PYTHON_TO_GO_CONVERSION_PLAN.md` → Link to `PYTHON_TO_GO_CONVERSION_STATUS.md`
9. `docs/migration/PYTHON_TO_GO_CONVERSION_STATUS.md` → Link to `PYTHON_TO_GO_BUILD_VERIFICATION.md`
10. `docs/audits/LANGCHAINGO_AUDIT.md` → Link to `docs/migration/LANGCHAINGO_MIGRATION_COMPLETE.md`

### P3 - Service Interconnectivity
11. All 4 service README files → Cross-link to each other as "Related Services"
12. `services/api-gateway/README.md` → Link to all 3 downstream service READMEs

### P4 - Planning Documents
13. `.planning/codebase/RAG_PIPELINE_AUDIT.md` → Link to `INFRASTRUCTURE_AUDIT.md`
14. `.planning/codebase/INFRASTRUCTURE_AUDIT.md` → Link to `.planning/research/SUMMARY.md`

---

## 6. FILES RECOMMENDED FOR ARCHIVAL

These files are superseded or redundant:

| File | Superseded By | Reason |
|------|--------------|--------|
| `data/FINAL_TEST_RESULTS.md` | `data/OLLAMA_REAL_TEST_RESULTS.md` | Ollama test is more real |
| `docs/audits/audit_*.md` (11 files) | `docs/audits/LANGCHAINGO_VS_AILOG_GUIDES_AUDIT.md` | Consolidated audit exists |
| `docs/archive/` (52 files) | Already in archive | Consider deleting if no historical value needed |
| `frontend/README.md` | N/A | Standard Vite template, no project-specific content |

---

## 7. SUMMARY STATISTICS

| Metric | Count |
|--------|-------|
| **Total .md files** | 122 |
| **Root level** | 2 |
| **docs/ (all subdirs)** | 103 |
| **.planning/** | 5 |
| **services/** | 4 |
| **supabase/** | 1 |
| **frontend/** | 1 |
| **data/** | 2 |
| **Active (non-archive)** | 70 |
| **Archived** | 52 |
| **Existing cross-references** | ~25 |
| **Missing cross-references** | ~55 |
| **Recommended backlink additions** | 14 |

---

*End of Documentation Map*
