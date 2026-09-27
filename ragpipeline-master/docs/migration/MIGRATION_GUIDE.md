# RAG Pipeline - Codebase Reorganization Guide

## Overview
This document describes the reorganization of the RAG pipeline codebase from a cluttered monolith into a clean microservices architecture.

## What Changed

### Phase 1: Documentation Cleanup ✅
**45+ redundant .md files archived to `docs/archive/`**

**Kept in root:**
- `README.md` - Main project documentation
- `REDUNDANCY_AUDIT.md` - Code audit findings
- `REORGANIZATION_PLAN.md` - This reorganization plan
- `.env.local.example` - Environment configuration template
- `PLAN.md` - Active development plan

**Archived (38 files):**
- All audit reports (AUDIT_*.md, COMPREHENSIVE_*.md, ENTERPRISE_*.md)
- All test reports (TEST_REPORT*.md, FINAL_TEST_*.md)
- All implementation summaries (IMPLEMENTATION_*.md, FINAL_*.md)
- All progress reports (STANDUP_*.md, DEVELOPMENT_*.md)
- Configuration docs (CONFIG_*.md, OPTIMAL_CONFIG_*.md)
- Self-query docs (SELF_QUERY_*.md)
- LangchainGo docs (LANGCHAINGO_*.md)
- Phase/backlog files (PHASE*.md)
- Other historical reports

### Phase 2: Data File Organization ✅
**Moved stray files to `data/` directory:**
- `science class 8.pdf` → `data/`
- `science_dataset.xlsx` → `data/`
- `visionary_rag_v5_grand_table.csv` → `data/`
- Jupyter notebooks → `docs/archive/`
- Word documents → `docs/archive/`

### Phase 3: Stray Files Removed ✅
- `$null` → deleted (PowerShell artifact)
- `echo` → deleted (command artifact)
- `-p` → deleted (flag artifact)

## Current Directory Structure

```
ragpipeline/
├── .env.local.example          # Environment config template
├── README.md                   # Main documentation
├── REDUNDANCY_AUDIT.md        # Audit findings
├── REORGANIZATION_PLAN.md     # Reorganization details
├── PLAN.md                     # Active development plan
│
├── services/                   # 🆕 Microservices
│   ├── embedding-service/     # ✅ Vertex AI embeddings (port 8081)
│   ├── query-understanding-service/  # ✅ Query rewriting & self-query (port 8083)
│   ├── vector-search-service/ # 🔄 TODO: Hybrid search (port 8082)
│   └── api-gateway/           # 🔄 TODO: External HTTP interface
│
├── shared/                     # 🆕 Shared Go libraries
│   └── go/
│       ├── models/            # Common data structures
│       ├── config/            # Configuration loader
│       ├── middleware/        # HTTP middleware
│       └── errors/            # Error handling
│
├── orchestrator/               # ⚠️ Legacy monolith (to be deprecated)
│   ├── cmd/server/            # Main entry point
│   ├── handler/               # HTTP handlers
│   ├── retrieval/             # Search logic → move to vector-search-service
│   ├── db/                    # Database queries → move to vector-search-service
│   ├── embed/                 # Stub embed → DELETE (use embedding-service)
│   ├── llm/                   # LLM client → move to query-understanding
│   ├── cache/                 # Response cache → move to api-gateway
│   ├── session/               # Session management → move to api-gateway
│   └── config/                # Configuration
│
├── ingestion/                  # ✅ Python ingestion (KEEP - production ready)
│   ├── pipeline.py
│   ├── parser/
│   ├── chunker/
│   ├── embedder/
│   ├── keywords/
│   └── writer/
│
├── ingestion-go/               # ❌ DELETE - stub implementation
│
├── retrievers/                 # Python self-query retrieval
│
├── quality_loop/               # Feedback quality analysis
│
├── eval/                       # Evaluation scripts
│
├── tests/                      # All test files
│
├── frontend/                   # React frontend
│
├── terraform/                  # GCP infrastructure
├── schema/                     # Database schemas
├── supabase/                   # Supabase migrations
├── scripts/                    # Utility scripts
├── data/                       # Test data files
└── docs/
    └── archive/               # Archived documentation (38 files)
```

## Next Steps: Complete Microservices Migration

### Step 1: Create Vector Search Service
**Source:** `orchestrator/retrieval/` + `orchestrator/db/`
**Target:** `services/vector-search-service/`

```bash
# TODO: Execute this migration
1. Copy orchestrator/retrieval/* → services/vector-search-service/search/
2. Copy orchestrator/db/* → services/vector-search-service/db/
3. Consolidate 3 RRF implementations into 1
4. Update import paths
5. Create go.mod with proper dependencies
6. Test service independently
```

### Step 2: Create API Gateway
**Source:** `orchestrator/handler/` + `orchestrator/cmd/server/`
**Target:** `services/api-gateway/`

```bash
# TODO: Execute this migration
1. Copy orchestrator/handler/* → services/api-gateway/handler/
2. Copy orchestrator/cache/ → services/api-gateway/cache/
3. Copy orchestrator/session/ → services/api-gateway/session/
4. Create main.go with route definitions
5. Add service-to-service HTTP client
6. Create go.mod
```

### Step 3: Move LLM Client
**Source:** `orchestrator/llm/`
**Target:** `services/query-understanding-service/llm/` (already exists)

```bash
# TODO: Execute this migration
1. Copy orchestrator/llm/* → services/query-understanding-service/llm/
2. Merge with existing implementation
3. Update imports
```

### Step 4: Create Shared Libraries
**Target:** `shared/go/`

```bash
# TODO: Create these shared packages
1. shared/go/models/ - Chunk, SearchResult, QueryRequest, etc.
2. shared/go/config/ - Environment variable loader
3. shared/go/middleware/ - Logging, recovery, request ID
4. shared/go/errors/ - Error response formatting
```

### Step 5: Delete Legacy Code
```bash
# TODO: After all services are operational
1. Delete orchestrator/ directory
2. Delete ingestion-go/ directory
3. Delete orchestrator/embed/ (stub implementation)
4. Update README to reflect new architecture
```

## Service Communication

### Current (Monolith)
```
HTTP Request → Orchestrator → (embed + search + llm + cache + session)
                ↑ Everything in one process
```

### Target (Microservices)
```
HTTP Request → API Gateway (8080)
                  ↓
          Embedding Service (8081)
                  ↓
          Vector Search Service (8082)
                  ↓
          Query Understanding Service (8083)
                  ↓
          Response → Cache → Client
```

**Communication Pattern:**
- Synchronous HTTP calls between services
- Shared request ID via `X-Request-ID` header
- Common error response format: `{"error": {"code": "...", "message": "..."}}`
- Service discovery via environment variables (e.g., `EMBEDDING_SERVICE_URL`)

## Port Assignments

| Service | Port | Purpose |
|---------|------|---------|
| API Gateway | 8080 | External HTTP interface |
| Embedding Service | 8081 | Vertex AI embeddings |
| Vector Search Service | 8082 | Hybrid search (dense + sparse + RRF) |
| Query Understanding | 8083 | Query rewriting, self-query parsing |

## Configuration Management

Each service reads from environment variables:
- Service-specific config (e.g., `EMBEDDING_MODEL`, `DB_HOST`)
- Shared config via `shared/go/config/` package
- `.env` files for local development
- GCP Secret Manager for production

## How to Run Services Locally

### Embedding Service (Already Working)
```bash
cd services/embedding-service
cp .env.example .env
# Edit .env with your GCP credentials
go mod download
go run main.go
# Service starts on http://localhost:8081
```

### Query Understanding Service (Already Working)
```bash
cd services/query-understanding-service
cp .env.example .env
# Edit .env with your GCP credentials
go mod download
go run main.go
# Service starts on http://localhost:8083
```

### Python Ingestion (Production Ready)
```bash
# Use venv
venv\Scripts\activate  # On Windows
pip install -r requirements.txt

# Run ingestion
python -m ingestion.pipeline --config config.yaml
```

## Testing Strategy

### Unit Tests
- Each service has its own `*_test.go` files
- Run: `go test ./...` in each service directory

### Integration Tests
- Test service-to-service communication
- Test database queries
- Test external API calls (Vertex AI)

### End-to-End Tests
- Full query flow through all services
- Full ingestion pipeline
- Use existing test files in `tests/`

## Migration Checklist

- [x] Archive redundant documentation
- [x] Organize data files
- [x] Create directory structure
- [x] Create embedding service
- [x] Create query understanding service
- [ ] Create vector search service
- [ ] Create API gateway
- [ ] Create shared libraries
- [ ] Move code from orchestrator to services
- [ ] Update all import paths
- [ ] Test each service independently
- [ ] Test service-to-service communication
- [ ] Delete orchestrator/ directory
- [ ] Delete ingestion-go/ directory
- [ ] Update README.md
- [ ] Update deployment scripts
- [ ] Update docker-compose.yml
- [ ] Update terraform configs

## Benefits of This Reorganization

1. **Clear Separation of Concerns**: Each service has a single responsibility
2. **Independent Deployment**: Services can be deployed and scaled independently
3. **Reduced Clutter**: Root directory is clean and understandable
4. **No Duplicate Code**: Single source of truth for each feature
5. **Easier Testing**: Test services in isolation
6. **Better Team Collaboration**: Different teams can work on different services
7. **Technology Flexibility**: Each service can use the best tool for its job

## Rollback Plan

If the reorganization causes issues:
1. All archived docs are in `docs/archive/` - can be restored
2. Git history is preserved - can revert commits
3. Old orchestrator/ remains until all services are tested
4. Parallel run: old and new can coexist during transition

## Questions?

Refer to:
- `REDUNDANCY_AUDIT.md` - Detailed audit findings
- `REORGANIZATION_PLAN.md` - Full reorganization strategy
- `README.md` - Updated project overview
- Service-specific `README.md` files in each service directory
