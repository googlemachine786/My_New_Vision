# Complete Infrastructure Audit — Visionary RAG Pipeline

**Date:** 2026-04-03  
**Scope:** Full infrastructure requirements for functional deployment + science textbook testing  
**Focus:** GCP services, local dependencies, env vars, service topology, data requirements, end-to-end test path

---

## 1. GCP APIs to Enable

Before any deployment, enable these APIs in your GCP project:

| API | Service Name | Why Needed |
|-----|-------------|------------|
| **Vertex AI API** | `aiplatform.googleapis.com` | Embeddings (`text-embedding-005`) + LLM (`gemini-2.0-flash`) |
| **Cloud Run API** | `run.googleapis.com` | Host Go orchestrator microservice |
| **Cloud Run Jobs API** | `run.googleapis.com` | Python ingestion pipeline jobs |
| **Cloud Build API** | `cloudbuild.googleapis.com` | Build container images |
| **Artifact Registry API** | `artifactregistry.googleapis.com` | Store Docker images |
| **Secret Manager API** | `secretmanager.googleapis.com` | Store DB passwords, JWT secrets, service account keys |
| **Cloud VPC Access API** | `vpcaccess.googleapis.com` | VPC connector for Cloud Run → AlloyDB/Redis |
| **Compute Engine API** | `compute.googleapis.com` | VPC network, subnets |
| **AlloyDB API** | `alloydb.googleapis.com` | Managed PostgreSQL with pgvector + ScaNN |
| **Memorystore Redis API** | `redis.googleapis.com` | Managed Redis 7.x for sessions/caching |
| **Cloud Trace API** | `cloudtrace.googleapis.com` | Distributed tracing (optional, `ENABLE_CLOUD_TRACE`) |
| **Cloud Monitoring API** | `monitoring.googleapis.com` | Metrics/alerting (optional, `ENABLE_CLOUD_MONITORING`) |
| **IAM API** | `iam.googleapis.com` | Service account permissions |
| **Cloud Resource Manager API** | `cloudresourcemanager.googleapis.com` | Project-level IAM bindings |

**One-liner to enable all:**
```bash
gcloud services enable \
  aiplatform.googleapis.com \
  run.googleapis.com \
  cloudbuild.googleapis.com \
  artifactregistry.googleapis.com \
  secretmanager.googleapis.com \
  vpcaccess.googleapis.com \
  compute.googleapis.com \
  alloydb.googleapis.com \
  redis.googleapis.com \
  cloudtrace.googleapis.com \
  monitoring.googleapis.com \
  iam.googleapis.com \
  cloudresourcemanager.googleapis.com
```

---

## 2. GCP Resources to Create

### 2.1 Vertex AI Models (No creation needed — just access)

| Model | Purpose | Env Var | Region |
|-------|---------|---------|--------|
| `text-embedding-005` | Embedding generation (768-dim) | `VERTEX_MODEL` | `us-central1` or `asia-south1` |
| `gemini-2.0-flash` | Query rewriting, intent classification, LLM answer generation | `GEMINI_MODEL` | Same as above |

**Required IAM role:** `roles/aiplatform.user` on the Cloud Run service account.

### 2.2 AlloyDB Cluster + Instance

| Resource | Value | Notes |
|----------|-------|-------|
| Cluster ID | `visionary-rag-cluster` | REGIONAL deployment |
| Instance ID | `visionary-rag-primary` | `n2-standard-8` (8 vCPU, 32GB RAM) |
| Database | `visionary` (created by Terraform) | PostgreSQL 15 |
| User | `visionary` | Initial user set in Terraform |
| Extensions required | `vector`, `btree_gin`, `pgcrypto` | Applied via `schema/v2_production.sql` |
| Tables | `cbse_taxonomy`, `parent_chunks`, `child_chunks`, `ingestion_dlq`, `ai_feedback_loop` | Schema v2.0 |

**Critical:** AlloyDB is required for ScaNN vector index — Cloud SQL with pgvector will NOT support the ScaNN index defined in the schema. If using Cloud SQL, the schema must be modified to use `vector_cosine_ops` with an `ivfflat` or `hnsw` index instead.

### 2.3 Memorystore Redis 7.x

| Resource | Value | Notes |
|----------|-------|-------|
| Instance name | `visionary-rag-session` | |
| Tier | `STANDARD_HA` | High availability |
| Memory | 4 GB | maxmemory set to 2.5GB (60% cap) |
| Version | `REDIS_7_0` | |
| Eviction policy | `allkeys-lru` | Prevents OOM |
| Databases used | DB 0 (API Gateway cache), DB 1 (embedding cache), DB 2 (search cache) | Multiple DB indices |

### 2.4 Cloud Run Services

| Service | Container | Port | Min Instances | Max Instances |
|---------|-----------|------|---------------|---------------|
| `visionary-rag-orchestrator` | Go API Gateway | 8080 | 2 | 100 |
| `visionary-rag-ingestion` (Job) | Python ingestion | N/A | 0 | 1 (per execution) |

### 2.5 Artifact Registry

| Resource | Value |
|----------|-------|
| Repository ID | `visionary-rag-images` |
| Format | Docker |
| Location | `asia-south1` |

### 2.6 Secret Manager (4 Secrets Required)

| Secret ID | Contains | Used By |
|-----------|----------|---------|
| `alloydb-dsn` | `postgresql://visionary:PASSWORD@ALLOYDB_IP:5432/visionary` | Cloud Run orchestrator, ingestion job |
| `redis-addr` | `REDIS_IP:6379` | Cloud Run orchestrator |
| `vertex-sa` | Service account JSON key for Vertex AI | Embedding service, query understanding |
| `jwt-secret` | Random string (strong secret) | API Gateway JWT auth |

### 2.7 VPC Network

| Resource | Value |
|----------|-------|
| VPC | `visionary-rag-vpc` (custom, not default) |
| Subnet | `visionary-rag-subnet` — `10.8.0.0/24` |
| VPC Connector | `visionary-rag-vpc-connector` (e2-standard, 2-10 instances) |

### 2.8 Service Account

The Cloud Run orchestrator needs a service account with these roles:
- `roles/aiplatform.user` — Vertex AI access
- `roles/secretmanager.secretAccessor` — Read secrets
- `roles/cloudtrace.agent` — Cloud Trace (optional)

---

## 3. Local Services to Run (Development Mode)

For local development without GCP, you need these services running:

### 3.1 Redis (Local)

```bash
# Via Docker (recommended)
docker run -d --name redis -p 6379:6379 redis:7-alpine

# Or via brew (macOS)
brew install redis && redis-server

# Or via choco (Windows)
choco install redis-64
```

**Connection:** `redis://localhost:6379` (used across DB 0, 1, 2)

### 3.2 PostgreSQL with pgvector (Local)

```bash
# Via Docker (pgvector image)
docker run -d --name postgres \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=ragdb \
  -p 5432:5432 \
  pgvector/pgvector:pg15
```

**Connection:** `postgresql://postgres:postgres@localhost:5432/ragdb`

**After starting, apply schema:**
```bash
psql postgresql://postgres:postgres@localhost:5432/ragdb -f schema/v2_production.sql
```

### 3.3 Ollama (Optional — Local Embeddings/LLM)

For fully offline development without GCP:

```bash
# Install Ollama
curl -fsSL https://ollama.com/install.sh | sh

# Pull models
ollama pull nomic-embed-text
ollama pull llama3.2:3b
```

**Connection:** `http://localhost:11434`

### 3.4 Google Cloud SDK (For Vertex AI Auth)

```bash
# Install gcloud CLI
# Then authenticate for Application Default Credentials:
gcloud auth application-default login
```

This is required for the embedding service to call Vertex AI without a service account JSON file in local dev.

---

## 4. Environment Variables — Complete List

### 4.1 API Gateway (`.env` or `.env.local`)

```bash
# Server
SERVER_PORT=8080
SERVER_ENV=development
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=120s
SERVER_IDLE_TIMEOUT=30s
SERVER_SHUTDOWN_TIMEOUT=40s

# Downstream Service URLs
EMBEDDING_SERVICE_URL=http://localhost:8081
VECTOR_SEARCH_SERVICE_URL=http://localhost:8082
QUERY_UNDERSTANDING_URL=http://localhost:8083
LLM_SERVICE_URL=http://localhost:8084

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10

# Auth
JWT_SECRET=your-strong-secret-key-here

# Rate Limiting
RATE_LIMIT_RPM=100
RATE_LIMIT_BURST=20

# Timeouts
TIMEOUT_EMBEDDING=10s
TIMEOUT_VECTOR_SEARCH=15s
TIMEOUT_QUERY_UNDERSTANDING=5s
TIMEOUT_LLM=60s
TIMEOUT_TOTAL=120s

# CAG (Cache Augmented Generation)
CAG_ENABLED=true
SEMANTIC_CACHE_ENABLED=true
EMBEDDING_CACHE_ENABLED=true
SEARCH_CACHE_ENABLED=true
TEMPLATE_CACHE_ENABLED=true
EXACT_CACHE_TTL=1h
SEMANTIC_CACHE_TTL=6h
SEMANTIC_CACHE_THRESHOLD=0.95
EMBEDDING_CACHE_TTL=24h
TEMPLATE_CACHE_TTL=12h
EMBEDDING_SERVICE_REDIS_URL=redis://localhost:6379/0
CACHE_INVALIDATE_ON_INGEST=true
```

### 4.2 Embedding Service

```bash
# Server
PORT=8081
ENVIRONMENT=development
READ_TIMEOUT=10s
WRITE_TIMEOUT=35s
IDLE_TIMEOUT=120s

# Vertex AI (Required)
VERTEX_PROJECT=your-gcp-project-id
VERTEX_LOCATION=us-central1
VERTEX_MODEL=text-embedding-005
VERTEX_DIMENSION=768
VERTEX_TIMEOUT=30s

# Task Types
DOCUMENT_TASK_TYPE=RETRIEVAL_DOCUMENT
QUERY_TASK_TYPE=RETRIEVAL_QUERY

# Retry
MAX_RETRIES=5
INITIAL_BACKOFF=100ms
MAX_BACKOFF=5s

# Request Limits
MAX_TEXTS_PER_BATCH=5
MAX_TEXT_LENGTH=20000

# Redis (for embedding cache)
REDIS_URL=redis://localhost:6379/1
EMBEDDING_CACHE_ENABLED=true
EMBEDDING_CACHE_TTL=24h
EMBEDDING_CACHE_MAX_SIZE=10000

# Observability (optional)
OTEL_EXPORTER_OTLP_ENDPOINT=
OTEL_SERVICE_NAME=embedding-service
OTEL_SAMPLE_RATE=1.0
LOG_LEVEL=debug

# GCP Auth (one of):
# - gcloud auth application-default login (local dev)
# - GOOGLE_APPLICATION_CREDENTIALS=/path/to/sa-key.json (production)
```

### 4.3 Vector Search Service

```bash
# Server
PORT=8082
ENVIRONMENT=development
READ_TIMEOUT=10s
WRITE_TIMEOUT=60s
IDLE_TIMEOUT=120s

# Database (Required)
DATABASE_URL=postgresql://postgres:postgres@localhost:5432/ragdb
DB_MAX_CONNS=25
DB_MIN_CONNS=5
DB_MAX_CONN_LIFETIME=30m
DB_MAX_CONN_IDLE_TIME=5m
DB_HEALTH_CHECK_PERIOD=30s

# Search
DEFAULT_TOP_K=10
MAX_TOP_K=100
QUERY_TIMEOUT=30s
SIMILARITY_THRESHOLD=0.36

# Dense Search
DENSE_CHILD_TABLE=child_chunks
DENSE_PARENT_TABLE=parent_chunks
EMBEDDING_COLUMN=embedding
EMBEDDING_DIM=768

# Sparse Search
SPARSE_TABLE=parent_chunks
KEYWORDS_COLUMN=extracted_keywords
SPARSE_OVERLAP_LIMIT=200

# RRF Fusion
RRF_K=60.0
RRF_DENSE_WEIGHT=1.0
RRF_SPARSE_WEIGHT=1.0

# Redis (for search cache)
REDIS_URL=redis://localhost:6379/2
SEARCH_CACHE_ENABLED=true
SEARCH_CACHE_TTL=30m
SEARCH_CACHE_MAX_SIZE=3000

# Observability (optional)
OTEL_EXPORTER_OTLP_ENDPOINT=
OTEL_SERVICE_NAME=vector-search-service
LOG_LEVEL=info
```

### 4.4 Query Understanding Service

```bash
# Server
PORT=8083
READ_TIMEOUT=10s
WRITE_TIMEOUT=60s
IDLE_TIMEOUT=60s

# Gemini/Vertex AI (Required)
GEMINI_PROJECT_ID=your-gcp-project-id
GEMINI_LOCATION=us-central1
GEMINI_MODEL=gemini-2.0-flash
GEMINI_TEMPERATURE=0.1
GEMINI_MAX_TOKENS=512
GEMINI_TOP_P=0.95
GEMINI_TOP_K=40
GEMINI_REQUEST_TIMEOUT=30s
GEMINI_MAX_RETRIES=3
GEMINI_RATE_LIMIT=10
GEMINI_RATE_BURST=20

# GCP Auth
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service-account.json

# Session
MAX_HISTORY_TURNS=10

# Prompt Versions
PROMPT_REWRITE_VERSION=v1
PROMPT_PARSE_FILTERS_VERSION=v1
PROMPT_CLASSIFY_INTENT_VERSION=v1
```

### 4.5 Root-level `.env.local` (Python ingestion + shared config)

```bash
# Supabase (alternative to AlloyDB for local dev)
SUPABASE_URL=https://your-project.supabase.co
SUPABASE_ANON_KEY=your-anon-key
SUPABASE_SERVICE_KEY=your-service-key
DATABASE_URL=postgresql://postgres:[PASSWORD]@db.your-project.supabase.co:5432/postgres

# GCP / Vertex AI
GOOGLE_CLOUD_PROJECT=your-gcp-project
GOOGLE_CLOUD_LOCATION=asia-south1
GOOGLE_APPLICATION_CREDENTIALS=/path/to/service_account.json

# Embedding model
EMBED_MODEL=sentence-transformers/all-MiniLM-L6-v2
EMBED_DIMENSION=384

# Ollama (local alternative)
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_EMBED_MODEL=nomic-embed-text
OLLAMA_LLM_MODEL=llama3.2:3b

# LLM Generation
LLM_TEMPERATURE=0.75
MAX_TOKENS=150

# App
APP_MODE=local
LOG_LEVEL=info
ENVIRONMENT=development
REQUEST_TIMEOUT=450ms

# Retrieval (optimized from grid search)
TOP_K=5
SIMILARITY_THRESHOLD=0.36
RETRIEVAL_STRATEGY=dense
RRF_K=60.0
BM25_K1=2.05

# Chunking
PARENT_MAX_CHARS=400
CHILD_MAX_CHARS=400
CHILD_OVERLAP=150

# Cache/Session
CACHE_TTL_SECONDS=300
SESSION_TTL_SECONDS=2100

# JWT
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRATION=24h

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=50

# Feature Flags
USE_LOCAL_EMBEDDINGS=false
USE_LOCAL_LLM=false
ENABLE_CLOUD_TRACE=false
ENABLE_CLOUD_MONITORING=false
```

### 4.6 Ingestion Pipeline (Python)

```bash
# Passed as CLI args or env vars to pipeline.py
GOOGLE_CLOUD_PROJECT=your-gcp-project
ALLOYDB_DSN=postgresql://visionary:PASSWORD@ALLOYDB_IP:5432/visionary
VERTEX_PROJECT=your-gcp-project
VERTEX_LOCATION=asia-south1
```

---

## 5. Service Startup Order

### 5.1 Docker Compose (All-in-One)

```bash
# Single command — docker-compose handles dependencies
docker-compose up -d

# Order enforced by depends_on:
# 1. Redis (healthcheck: redis-cli ping)
# 2. PostgreSQL (healthcheck: pg_isready)
# 3. Embedding Service (depends on Redis)
# 4. Vector Search Service (depends on PostgreSQL + Redis)
# 5. Query Understanding Service (no dependencies, but needs GCP auth)
# 6. API Gateway (depends on all above)
```

### 5.2 Manual Local Startup

```bash
# Step 1: Start infrastructure
docker run -d --name redis -p 6379:6379 redis:7-alpine
docker run -d --name postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=ragdb -p 5432:5432 pgvector/pgvector:pg15

# Step 2: Wait for health
sleep 10
redis-cli ping  # Should return PONG
pg_isready -h localhost -U postgres  # Should return "accepting connections"

# Step 3: Apply database schema
psql postgresql://postgres:postgres@localhost:5432/ragdb -f schema/v2_population.sql

# Step 4: GCP authentication (for Vertex AI)
gcloud auth application-default login

# Step 5: Start services (order matters)
# Terminal 1: Embedding Service
cd services/embedding-service && cp .env.example .env && go run main.go

# Terminal 2: Vector Search Service
cd services/vector-search-service && cp .env.example .env && go run main.go

# Terminal 3: Query Understanding Service
cd services/query-understanding-service && cp .env.example .env && go run main.go

# Terminal 4: API Gateway
cd services/api-gateway && cp .env.example .env && go run main.go
```

### 5.3 Production (GCP via Terraform)

```bash
# Step 1: Apply Terraform (creates VPC, AlloyDB, Redis, Cloud Run, Secrets)
cd terraform && terraform apply

# Step 2: Apply database schema
psql $ALLOYDB_DSN -f ../schema/v2_production.sql

# Step 3: Populate secrets
echo -n "postgresql://visionary:PASSWORD@IP:5432/visionary" | gcloud secrets versions add alloydb-dsn --data-file=-
echo -n "REDIS_IP:6379" | gcloud secrets versions add redis-addr --data-file=-
# ... (see PLAN.md T1.9)

# Step 4: Build and push container images
docker build -t asia-south1-docker.pkg.dev/PROJECT/visionary-rag-images/orchestrator:latest -f services/api-gateway/Dockerfile .
docker push asia-south1-docker.pkg.dev/PROJECT/visionary-rag-images/orchestrator:latest

# Step 5: Deploy Cloud Run services
terraform apply (Cloud Run auto-deploys on image update)
```

---

## 6. End-to-End Test with Science Textbooks

### 6.1 Prerequisites

1. All 4 Go services running (ports 8080-8083)
2. Redis running on port 6379
3. PostgreSQL with pgvector running on port 5432
4. Schema applied (`schema/v2_production.sql`)
5. GCP authentication (`gcloud auth application-default login`)
6. Science textbook PDFs available

### 6.2 Step 1: Ingest Science Textbooks

```bash
# Navigate to ingestion
cd ingestion

# Run pipeline for a CBSE Science textbook
python pipeline.py \
  --pdf /path/to/NCERT_Science_Class8.pdf \
  --grade 8 \
  --subject "Science" \
  --taxonomy-id 35 \
  --project-id your-gcp-project \
  --location asia-south1 \
  --alloydb-dsn "postgresql://postgres:postgres@localhost:5432/ragdb" \
  --verbose
```

**Expected output:**
```
Pipeline Complete
==================================================
Pages processed: 120
Elements created: 450
Parent chunks: 380
Child chunks: 1520
Keywords extracted: 3040
Embeddings created: 1520
DB inserts success: 1520
DB inserts failure: 0
```

### 6.3 Step 2: Verify Data in Database

```bash
psql postgresql://postgres:postgres@localhost:5432/ragdb -c "
  SELECT grade, subject, chapter, COUNT(*) as chapters
  FROM cbse_taxonomy 
  GROUP BY grade, subject;
"

psql postgresql://postgres:postgres@localhost:5432/ragdb -c "
  SELECT COUNT(*) as parent_chunks FROM parent_chunks;
  SELECT COUNT(*) as child_chunks FROM child_chunks;
"
```

### 6.4 Step 3: Query the RAG Pipeline

```bash
# Test health
curl http://localhost:8080/health

# Query with grade/subject filter
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{
    "query": "What is photosynthesis?",
    "grade": "8",
    "subject": "Science",
    "stream": false
  }'
```

**Expected response:**
```json
{
  "session_id": "...",
  "request_id": "...",
  "answer": "Photosynthesis is the process by which plants...",
  "sources": [
    {
      "parent_id": "uuid-here",
      "content": "Photosynthesis is the process by which green plants...",
      "score": 0.82
    }
  ],
  "total_tokens": 45
}
```

### 6.5 Step 4: Streaming Query

```bash
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test-token" \
  -d '{
    "query": "Explain the water cycle for class 7",
    "grade": "7",
    "subject": "Science",
    "stream": true
  }'
```

### 6.6 Step 5: Run Evaluation Scripts

```bash
# Activate Python venv
.\venv\Scripts\Activate.ps1  # Windows
source venv/bin/activate      # Linux/Mac

# End-to-end test
python test_e2e.py

# Science-specific test (Class 8)
python test_science_class8.py

# Full RAG evaluation
python evaluate_complete_rag.py
```

### 6.7 Science Textbook Data Sources

The system is pre-configured for **CBSE NCERT textbooks** (Grades 6-8 Science):

| Grade | Subject | Taxonomy IDs | Chapters |
|-------|---------|-------------|----------|
| 6 | Science | 1-16 | Food, Components of Food, Fibre to Fabric, ..., Garbage In Garbage Out |
| 7 | Science | 17-34 | Nutrition in Plants/Animals, Heat, Acids Bases Salts, ..., Wastewater Story |
| 8 | Science | 35-52 | Crop Production, Microorganisms, Synthetic Fibres, ..., Pollution of Air and Water |

**To add your own PDFs:**
1. Place PDF in `data/textbooks/`
2. Find or create taxonomy entry in `cbse_taxonomy` table
3. Run `python pipeline.py --pdf data/textbooks/your-book.pdf --grade X --subject "Science" --taxonomy-id Y`

### 6.8 Quick Smoke Test (No Textbooks Needed)

If you want to test the services are running without ingesting data:

```bash
# Health checks
curl http://localhost:8080/health    # API Gateway
curl http://localhost:8081/health    # Embedding Service
curl http://localhost:8082/health    # Vector Search Service
curl http://localhost:8083/health    # Query Understanding Service

# Test embedding directly
curl -X POST http://localhost:8081/embed \
  -H "Content-Type: application/json" \
  -d '{"text": "What is gravity?"}'

# Test query understanding
curl -X POST http://localhost:8083/rewrite \
  -H "Content-Type: application/json" \
  -d '{"query": "tell me about cells", "history": []}'
```

---

## 7. Architecture Diagram — Complete Service Topology

```
                          ┌─────────────────┐
                          │   Client/Browser │
                          └────────┬────────┘
                                   │ HTTP
                                   ▼
                    ┌──────────────────────────┐
                    │   API Gateway (:8080)    │
                    │  - JWT Auth              │
                    │  - Rate Limiting         │
                    │  - CAG Multi-layer Cache │
                    │  - Session Management    │
                    │  - Request Tracing       │
                    └─────┬──────┬──────┬──────┘
                          │      │      │
              ┌───────────┘      │      └───────────┐
              ▼                  ▼                  ▼
    ┌─────────────────┐ ┌──────────────────┐ ┌──────────────────────┐
    │Embedding Service│ │ Vector Search    │ │ Query Understanding  │
    │    (:8081)      │ │ Service (:8082)  │ │ Service (:8083)      │
    │                 │ │                  │ │                      │
    │→ Vertex AI      │ │→ PostgreSQL      │ │→ Gemini 2.0 Flash    │
    │  text-embedding │ │  pgvector/ScaNN  │ │  (Vertex AI)         │
    │  -005           │ │  - parent_chunks │ │                      │
    │→ Redis DB:1     │ │  - child_chunks  │ │→ Rate Limiter        │
    │  (embed cache)  │ │→ Redis DB:2      │ │→ Circuit Breaker     │
    │                 │ │  (search cache)  │ │                      │
    └─────────────────┘ └──────────────────┘ └──────────────────────┘
              │                  │                  │
              ▼                  ▼                  ▼
    ┌──────────────────────────────────────────────────────┐
    │              Vertex AI (GCP)                         │
    │  - text-embedding-005 (768-dim)                      │
    │  - gemini-2.0-flash                                  │
    └──────────────────────────────────────────────────────┘

    ┌─────────────────┐    ┌──────────────────┐
    │ Redis (:6379)   │    │ PostgreSQL (:5432)│
    │  DB:0 API Gate  │    │  - cbse_taxonomy │
    │  DB:1 Embeddings│    │  - parent_chunks │
    │  DB:2 Search    │    │  - child_chunks  │
    │                 │    │  - ingestion_dlq │
    │                 │    │  - ai_feedback   │
    └─────────────────┘    └──────────────────┘

    ┌──────────────────────────────────────────┐
    │  Python Ingestion Pipeline (CLI/Job)     │
    │  1. PDF Parse (PyMuPDF + pdfplumber)     │
    │  2. Font Calibration                     │
    │  3. Table Extraction                     │
    │  4. Heading Mapping                      │
    │  5. Parent-Child Chunking (400/150)      │
    │  6. Keyword Extraction (YAKE)            │
    │  7. Vertex AI Embedding                  │
    │  8. AlloyDB/PostgreSQL Write              │
    └──────────────────────────────────────────┘
```

---

## 8. Cost Estimates (GCP Production)

| Resource | Monthly Cost (asia-south1) | Notes |
|----------|---------------------------|-------|
| AlloyDB (n2-standard-8, REGIONAL) | ~$350-450 | 8 vCPU, 32GB RAM, storage extra |
| Memorystore Redis 7.x (4GB, HA) | ~$90 | STANDARD_HA tier |
| Cloud Run (2 min instances) | ~$30-60 | 2 vCPU, 2Gi, depends on traffic |
| Vertex AI Embeddings | ~$0.025/1K requests | text-embedding-005 |
| Vertex AI Gemini 2.0 Flash | ~$0.075/1K input chars | Very cheap for Flash |
| Artifact Registry | ~$0-5 | Small storage |
| Secret Manager | ~$0-6 | 4 secrets |
| **Total (baseline)** | **~$475-611/month** | Before traffic costs |

---

## 9. Checklist — Everything You Need

### GCP Setup
- [ ] GCP project created with billing enabled
- [ ] All 14 APIs enabled (see Section 1)
- [ ] Service account created with `aiplatform.user`, `secretmanager.secretAccessor`, `cloudtrace.agent` roles
- [ ] `gcloud auth application-default login` completed
- [ ] Terraform state bucket created (`visionary-rag-terraform-state`)

### Infrastructure (Terraform)
- [ ] `terraform/terraform.tfvars` created with `gcp_project_id`, `region`, `alloydb_initial_password`
- [ ] `terraform apply` completed successfully
- [ ] VPC, subnet, connector created
- [ ] AlloyDB cluster + instance running
- [ ] Memorystore Redis running
- [ ] Cloud Run service deployed
- [ ] Artifact Registry repository created
- [ ] 4 secrets populated in Secret Manager

### Database
- [ ] Schema applied: `psql $ALLOYDB_DSN -f schema/v2_production.sql`
- [ ] Taxonomy verified: `SELECT COUNT(*) FROM cbse_taxonomy` (should be 52)
- [ ] Extensions verified: `vector`, `btree_gin`, `pgcrypto`

### Local Development
- [ ] Redis running on localhost:6379
- [ ] PostgreSQL with pgvector running on localhost:5432
- [ ] Schema applied to local DB
- [ ] GCP authenticated (`gcloud auth application-default login`)
- [ ] All 4 `.env` files created from `.env.example` templates
- [ ] All 4 Go services running (`make run` or manual)

### Data
- [ ] Science textbook PDFs obtained (NCERT Class 6-8)
- [ ] Ingestion pipeline run for each textbook
- [ ] Data verified in database (parent_chunks, child_chunks counts)

### Testing
- [ ] Health endpoints return 200 on all 4 services
- [ ] Embedding service returns 768-dim vectors
- [ ] Vector search returns results for test queries
- [ ] Full query pipeline returns answers with sources
- [ ] `python test_e2e.py` passes
- [ ] `python test_science_class8.py` passes

---

## 10. Common Pitfalls

| Problem | Cause | Fix |
|---------|-------|-----|
| `VERTEX_PROJECT is required` | Missing env var | Set `VERTEX_PROJECT` in embedding service `.env` |
| `DATABASE_URL is required` | Missing env var | Set `DATABASE_URL` in vector search `.env` |
| `ScaNN index not supported` | Using Cloud SQL | Use AlloyDB or change schema to `ivfflat`/`hnsw` |
| `Failed to connect to Redis` | Redis not running | `docker run -d -p 6379:6379 redis:7-alpine` |
| `401 Unauthorized` on /query | JWT secret mismatch | Use same `JWT_SECRET` in API Gateway `.env` |
| `total timeout must be > sum` | Timeout misconfiguration | Ensure `TIMEOUT_TOTAL` > sum of all individual timeouts |
| `JWT_SECRET must not be 'change-me-in-production'` | Default secret | Set a strong random secret |
| Vertex AI 403 | API not enabled or no permissions | Enable `aiplatform.googleapis.com` and grant `roles/aiplatform.user` |
| Embedding dimension mismatch | `VERTEX_DIMENSION` vs schema | Ensure both are 768 (text-embedding-005) |
| Python ingestion fails on PDF | PDF is image-based (scanned) | Use OCR-first or ensure text-based PDFs |
