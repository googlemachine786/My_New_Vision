# 🔍 RAG Pipeline Technical Audit Findings

**Audit Date:** March 27, 2026  
**Auditor:** Senior Software Engineer (12+ years distributed systems)  
**Scope:** Complete technical audit against Visionary_Production_RAG_GCP_v2.docx and rag_evaluation_metrics.pdf requirements  

---

## Executive Summary

This audit identifies **47 gaps** between the requirements specification (Visionary_Production_RAG_GCP_v2.docx, rag_evaluation_metrics.pdf) and the current implementation. The implementation shows strong foundational work with ~12,800 lines of code across 50+ files, but has **critical production readiness gaps** that must be addressed before deployment.

### Overall Assessment

| Category | Completion | Status |
|----------|------------|--------|
| **Infrastructure (Terraform)** | 85% | 🟡 Partial |
| **Database Schema** | 95% | 🟢 Complete |
| **Go Ingestion Pipeline** | 70% | 🟡 Partial |
| **Go Orchestration Layer** | 75% | 🟡 Partial |
| **Python Ingestion (Reference)** | 90% | 🟢 Complete |
| **Evaluation Metrics** | 40% | 🔴 Incomplete |
| **Quality Loop** | 60% | 🟡 Partial |
| **Observability** | 20% | 🔴 Missing |
| **Security** | 50% | 🟡 Partial |
| **Documentation** | 95% | 🟢 Complete |

### Critical Findings Summary

| Severity | Count | Must-Fix Before Production |
|----------|-------|---------------------------|
| **CRITICAL** | 8 | ✅ YES - All 8 |
| **HIGH** | 12 | ✅ YES - All 12 |
| **MEDIUM** | 18 | ⚠️ Recommended |
| **LOW** | 9 | 📋 Nice-to-have |

---

## Critical Gaps (Must-Have for Production)

### CRIT-01: No Actual LLM Generation Implementation

**Requirement (v2.docx §6.2):**
> "Stream tokens via SSE from Gemini 1.5/2.0 Flash with sub-350ms TTFT"

**Current Implementation:**
```go
// orchestrator/handler/rag_handler.go:230
func (h *RAGHandler) createGeminiStream(ctx context.Context, prompt string) (<-chan string, error) {
    // In production, use the actual Gemini streaming API
    // For now, return a mock stream
    stream := make(chan string, 100)
    go func() {
        defer close(stream)
        response := "This is a mock response. In production, this would stream from Gemini 1.5 Flash."
        // ...
    }()
    return stream, nil
}
```

**Gap:** The SSE streaming handler uses **mock responses** instead of actual Gemini API calls. The `google.golang.org/genai` SDK is listed in requirements but not imported or used.

**Impact:** 
- Cannot serve real student queries
- TTFT SLA (sub-500ms) is unmeasurable
- No actual generation occurs in production

**Fix Required:**
```go
import "google.golang.org/genai"

func (h *RAGHandler) createGeminiStream(ctx context.Context, prompt string) (<-chan string, error) {
    client, err := genai.NewClient(ctx, &genai.ClientConfig{
        Project:  h.cfg.VertexProject,
        Location: h.cfg.VertexLocation,
    })
    if err != nil {
        return nil, err
    }
    
    resp, err := client.Models.GenerateContentStream(ctx, h.cfg.GenerationModel, prompt, nil)
    // Handle streaming response...
}
```

**Files to Modify:** `orchestrator/handler/rag_handler.go`, `orchestrator/go.mod`

---

### CRIT-02: No Query Rewriting Implementation

**Requirement (v2.docx §6.2):**
> "Query rewriting with coreference resolution using Gemini Flash (<30ms)"

**Current Implementation:**
```go
// orchestrator/handler/rag_handler.go:172
func (h *RAGHandler) rewriteQuery(ctx context.Context, query string, history []session.Turn) (string, error) {
    // Build prompt for query rewriting
    // ...
    // Call Gemini Flash for rewriting
    // In production, use the actual Gemini API call
    // For now, return original query (rewriting is optional)
    _ = prompt
    return query, nil  // ← NO-OP
}
```

**Gap:** Query rewriting is a **no-op** - returns the original query unchanged. This breaks conversational context handling (e.g., "What about photosynthesis?" after asking about plants).

**Impact:**
- Multi-turn conversations fail
- Coreference resolution missing ("it", "they" not resolved)
- Student experience degraded for follow-up questions

**Fix Required:** Implement actual Gemini Flash call for query rewriting with conversation history.

**Files to Modify:** `orchestrator/handler/rag_handler.go`

---

### CRIT-03: Evaluation Metrics Have Zero Data Coverage

**Requirement (rag_evaluation_metrics.pdf Page 3):**
> "A robust RAG evaluation stack must combine metrics across all stages into a coordinated multi-dimensional assessment"

**Current Implementation:**
```python
# evaluate_complete_rag.py:273
metrics = evaluator.evaluate_all(
    chunks=[],      # ← EMPTY
    embeddings=[],  # ← EMPTY
    qa_pairs=[],    # ← EMPTY
    results=test_results if isinstance(test_results, list) else []
)

# evaluate_complete_rag.py:136
def calculate_chunk_coherence_score(chunks, embeddings):
    if not chunks or not embeddings:
        return 0.0  # ← ALWAYS RETURNS 0
```

**Gap:** All 29 metrics are **implemented as functions** but execute with **empty data**. Current evaluation results:
- Chunking metrics: 0.00 (8/8 metrics)
- Retrieval metrics: Hardcoded values (0.75, 0.68, etc.)
- Generation metrics: 0.00-0.27

**Impact:**
- Cannot measure actual pipeline quality
- No baseline for improvement
- Production deployment without quality gates

**Fix Required:**
1. Integrate with actual ingestion pipeline to load real chunks
2. Create golden QA dataset (200+ questions per grade)
3. Wire evaluation to actual retrieval results

**Files to Modify:** `evaluate_complete_rag.py`, `test_continuous_improvement.py`

---

### CRIT-04: No OpenTelemetry Instrumentation

**Requirement (v2.docx §13.2):**
> "All metrics flow to Cloud Monitoring via OpenTelemetry Go SDK. Distributed traces link embed → retrieve → generate stages in Cloud Trace."

**Current Implementation:**
```go
// orchestrator/observability/ directory exists but EMPTY
// No OTel imports in any Go file
// No span creation in request lifecycle
```

**Gap:** The `orchestrator/observability/` directory exists but contains **no instrumentation code**. No traces, no metrics, no distributed tracing.

**Impact:**
- Cannot debug latency issues
- No TTFT decomposition (which stage is slow?)
- No production observability
- Cannot meet SLA monitoring requirements

**Fix Required:**
```go
import "go.opentelemetry.io/otel"

func (h *RAGHandler) Query(w http.ResponseWriter, r *http.Request) {
    ctx, span := otel.Tracer("visionary").Start(r.Context(), "rag.query")
    defer span.End()
    
    // Instrument each stage
    embedCtx, embedSpan := otel.Tracer("visionary").Start(ctx, "vertex.embed")
    embedding, err := h.vertex.EmbedQuery(embedCtx, query)
    embedSpan.End()
    // ...
}
```

**Files to Create:** `orchestrator/observability/tracer.go`, `orchestrator/observability/metrics.go`  
**Files to Modify:** `orchestrator/handler/rag_handler.go`, `orchestrator/cmd/server/main.go`

---

### CRIT-05: No Actual Hybrid Search in Python Pipeline

**Requirement (v2.docx §6.3):**
> "Hybrid search: ScaNN dense + GIN keyword intersection combined via RRF fusion (k=60)"

**Current Implementation:**
```python
# ask_query.py:38
def retrieve_context(self, query: str, top_k: int = 3):
    # Simple keyword-based retrieval for demo
    # In production, use actual vector similarity
    query_words = set(query.lower().split())
    scored_chunks = []
    for chunk in self.chunks:
        chunk_words = set(chunk["content"].lower().split())
        overlap = len(query_words & chunk_words)  # ← KEYWORD ONLY
```

**Gap:** Python test scripts use **keyword overlap** instead of actual hybrid search. The Go implementation exists (`orchestrator/retrieval/hybrid_search.go`) but is not integrated with the evaluation pipeline.

**Impact:**
- Evaluation results don't reflect production retrieval
- Dense vector search not tested
- RRF fusion not validated

**Fix Required:** Integrate Go hybrid search with Python evaluation or implement Python equivalent using `pgvector` client.

**Files to Modify:** `evaluate_complete_rag.py`, `test_continuous_improvement.py`

---

### CRIT-06: LLM Judge Untested and Incomplete

**Requirement (v2.docx §12.2):**
> "The LLM Judge requires the exact retrieval context that produced the rejected response to generate meaningfully corrected responses."

**Current Implementation:**
```python
# quality_loop/judge.py:160
async def fetch_context_chunks(db_connection, chunk_uuids: List[str]) -> List[Dict]:
    # SQL: SELECT content, page_number FROM child_chunks WHERE child_id = ANY($1)
    # In production, use actual database connection
    
    # Placeholder implementation
    chunks = []
    for uuid in chunk_uuids:
        chunks.append({
            'content': f"Content for chunk {uuid}",  # ← MOCK DATA
            'page': 42,
        })
    return chunks
```

**Gap:** The LLM Judge has **placeholder implementations** for critical database operations. No actual DB connection, no UUID[] array reconstruction.

**Impact:**
- Quality loop cannot process real feedback
- No golden responses generated for tuning
- Vertex AI fine-tuning pipeline blocked

**Fix Required:** Implement actual `psycopg3` async connection and UUID[] array queries.

**Files to Modify:** `quality_loop/judge.py`

---

### CRIT-07: No Response Caching Layer

**Requirement (v2.docx §8.2):**
> "Cache frequent query embeddings and common retrievals to reduce latency and cost"

**Current Implementation:**
- No Redis caching in query path
- Every query re-computes embeddings
- No semantic caching for similar queries

**Gap:** Redis is configured for session storage only. No response caching, no embedding cache, no retrieval cache.

**Impact:**
- Repeated queries cost 3-5× more (re-embedding + re-retrieval)
- Latency 2-3× higher for common queries
- Cannot handle traffic spikes efficiently

**Fix Required:**
```python
# Add Redis caching layer
from redis import asyncio as redis

class ResponseCache:
    def __init__(self, redis_client, ttl=300):
        self.redis = redis_client
        self.ttl = ttl
    
    async def get(self, query_embedding: List[float]) -> Optional[str]:
        # Semantic similarity search in cache
        pass
    
    async def set(self, query_embedding: List[float], response: str):
        await self.redis.setex(f"cache:{hash(query_embedding)}", self.ttl, response)
```

**Files to Create:** `orchestrator/cache/response_cache.go`  
**Files to Modify:** `orchestrator/handler/rag_handler.go`

---

### CRIT-08: No Load Testing or Performance Validation

**Requirement (v2.docx §11.2):**
> "Load Test Acceptance Criteria: p99 TTFT < 500ms at 1,000 concurrent users"

**Current Implementation:**
- No k6 or Locust load tests
- No concurrent user simulation
- Performance metrics from single-user tests only

**Gap:** Zero load testing infrastructure. No validation of 1,000 concurrent user requirement.

**Impact:**
- Production may fail under load
- TTFT SLA unvalidated
- Connection pool sizing untested

**Fix Required:** Create k6 load test scripts with gradual ramp-up to 1,000 VUs.

**Files to Create:** `eval/load_test.js` (k6 script), `eval/performance_tests.py`

---

## High Priority Gaps (Should-Have)

### HIGH-01: No DOCX/Markdown Ingestion Support

**Requirement (v2.docx §3.2):**
> "Standard extraction tools destroy educational document structure. The five-pass pipeline applies progressive structural resolution."

**Current Implementation:**
```python
# ingestion/pipeline.py:57
parser.add_argument("--pdf", required=True, help="Path to PDF file")
# No --docx or --markdown flags
```

**Gap:** Only PDF ingestion implemented. CBSE also distributes content in DOCX and Markdown formats.

**Impact:**
- Cannot ingest non-PDF educational materials
- Manual conversion required
- Limited content sources

**Files to Create:** `ingestion/parser/docx_parser.py`, `ingestion/parser/markdown_parser.py`

---

### HIGH-02: No Content Deduplication

**Requirement:** Implicit in production requirements (no duplicate content in RAG systems)

**Current Implementation:**
- No similarity check before insertion
- Same chunk can be ingested multiple times
- No deduplication across different PDFs

**Gap:** Ingesting the same chapter from multiple sources creates duplicate chunks.

**Impact:**
- Wasted embedding API costs
- Degraded retrieval quality (duplicate results)
- Larger database than necessary

**Fix Required:** Add L2 norm similarity check before insertion.

**Files to Modify:** `ingestion/writer/alloydb_writer.go`, `ingestion-go/writer/alloydb_writer.go`

---

### HIGH-03: No Incremental Ingestion

**Requirement:** Production requirement for content updates

**Current Implementation:**
```python
# ingestion/pipeline.py - Full re-ingestion only
async def run(self):
    # Always processes entire PDF
    # No change detection
    # No upsert logic
```

**Gap:** Cannot update individual chapters. Must re-ingest entire textbook for single-page changes.

**Impact:**
- High operational cost for content updates
- Downtime during re-ingestion
- Version control complexity

**Files to Modify:** `ingestion/pipeline.py`, `ingestion/writer/alloydb_writer.py`

---

### HIGH-04: No Database Migration Tooling

**Requirement:** Production requirement for schema evolution

**Current Implementation:**
```bash
# Manual schema application
psql $ALLOYDB_DSN -f schema/v2_production.sql
```

**Gap:** No Alembic, Flyway, or golang-migrate. Schema changes require manual intervention.

**Impact:**
- Schema drift between environments
- No rollback capability
- Manual deployment errors

**Files to Create:** `migrations/001_initial_schema.sql`, `migrations/002_add_indexes.sql`

---

### HIGH-05: No Read Replicas Configured

**Requirement (v2.docx §8.1):**
> "AlloyDB REGIONAL cluster with 99.99% SLA"

**Current Implementation:**
```hcl
# terraform/main.tf
resource "google_alloydb_instance" "primary" {
  instance_id = "visionary-rag-primary"
  instance_type = "PRIMARY"
  # No read replica configured
}
```

**Gap:** Single primary instance. No read replicas for query scaling.

**Impact:**
- Single point of failure (despite REGIONAL)
- Query load not distributed
- Cannot scale read-heavy retrieval workload

**Files to Modify:** `terraform/main.tf`

---

### HIGH-06: No PgBouncer Sidecar Deployment

**Requirement (v2.docx §8.1):**
> "PgBouncer deployed as sidecar to Cloud Run with transaction pooling"

**Current Implementation:**
```ini
# orchestrator/pgbouncer/pgbouncer.ini - Config exists
[databases]
visionary = host=ALLOYDB_PRIVATE_IP port=5432 dbname=visionary
```

**Gap:** Config file exists but **not deployed** as Cloud Run sidecar. Go app connects directly to AlloyDB.

**Impact:**
- Connection pool exhaustion at 400+ concurrent users
- AlloyDB connection limit breach
- No connection reuse

**Files to Modify:** `terraform/main.tf` (add sidecar config), `orchestrator/config/config.go`

---

### HIGH-07: No JWT Authentication Middleware

**Requirement (v2.docx §9.2):**
> "JWT claim carries taxonomy_id. Orchestrator enforces taxonomy_id filter on every AlloyDB query."

**Current Implementation:**
```go
// orchestrator/handler/rag_handler.go:260
func (h *RAGHandler) extractClaims(r *http.Request) (*JWTClaims, error) {
    // JWT parsing implemented
    // BUT: No middleware to protect routes
}
```

**Gap:** JWT extraction exists but no middleware. `/query` endpoint can be called without authentication in local dev.

**Impact:**
- Grade-level isolation bypassed
- Unauthorized access possible
- Student data exposure risk

**Files to Create:** `orchestrator/middleware/auth.go`  
**Files to Modify:** `orchestrator/handler/router.go`

---

### HIGH-08: No Dead Letter Queue Processing

**Requirement (v2.docx §5):**
> "DLQ table for failed ingestion attempts with retry logic"

**Current Implementation:**
```sql
-- schema/v2_production.sql
CREATE TABLE ingestion_dlq (
  dlq_id BIGSERIAL PRIMARY KEY,
  payload JSONB NOT NULL,
  error_message TEXT NOT NULL,
  retry_count INTEGER NOT NULL DEFAULT 0,
  failed BOOLEAN NOT NULL DEFAULT TRUE,
  -- Index exists
);
```

**Gap:** DLQ table exists but **no retry processor**. Failed inserts never retried.

**Impact:**
- Data loss on transient failures
- Manual intervention required
- No automatic recovery

**Files to Create:** `ingestion/dlq/retry_processor.py`

---

### HIGH-09: No Feedback Score Update Endpoint

**Requirement (v2.docx §12.1):**
> "Student gives thumbs down (feedback_score = -1)"

**Current Implementation:**
```go
// ai_feedback_loop table has feedback_score column
// But no HTTP endpoint to update it
```

**Gap:** Feedback is logged but students cannot update scores (thumbs up/down) after receiving response.

**Impact:**
- Quality loop has no feedback signals
- Cannot collect training data
- No user satisfaction tracking

**Files to Create:** `orchestrator/handler/feedback_handler.go`

---

### HIGH-10: No Cloud Scheduler for Quality Loop

**Requirement (v2.docx §12):**
> "Cloud Scheduler triggers every 6 hours"

**Current Implementation:**
```python
# quality_loop/judge.py - Can run manually
# No Cloud Scheduler job defined
```

**Gap:** Quality loop can run manually but not scheduled. No automated processing.

**Impact:**
- Golden responses not generated automatically
- Tuning data stale
- Manual operational burden

**Files to Create:** `terraform/quality_loop_scheduler.tf`

---

### HIGH-11: No Response Streaming in Python Test Scripts

**Requirement (v2.docx §6.2):**
> "Stream tokens via SSE with http.Flusher per token"

**Current Implementation:**
```python
# ask_query.py:85
def generate_answer(self, query: str, context: str):
    response = requests.post(
        f"{self.ollama_url}/api/generate",
        json={"model": "llama3.2:3b", "prompt": prompt, "stream": False}  # ← NOT STREAMING
    )
```

**Gap:** Test scripts use batch responses, not streaming. Cannot measure TTFT.

**Impact:**
- TTFT metrics inaccurate
- User experience different from production
- Performance testing invalid

**Files to Modify:** `ask_query.py`, `test_continuous_improvement.py`

---

### HIGH-12: No Taxonomy Filtering in Evaluation

**Requirement (v2.docx §9.3):**
> "Grade 6 student cannot retrieve Grade 8 content — enforced at DB layer"

**Current Implementation:**
```python
# test_continuous_improvement.py:150
def retrieve_chunks(self, query: str, top_k: int):
    # No taxonomy_id filter
    # All chunks searched regardless of grade
```

**Gap:** Evaluation doesn't enforce grade-level isolation.

**Impact:**
- Evaluation results don't reflect production behavior
- Cross-grade contamination possible in tests
- False positive retrieval metrics

**Files to Modify:** `test_continuous_improvement.py`, `evaluate_complete_rag.py`

---

## Medium Priority Gaps (Could-Have)

### MED-01: No Subsection Detection in Metadata

**Gap:** `ParsedElement.subsection` always `None`. Only chapter/section detected.

**Impact:** Reduced retrieval precision for specific topics.

**Files to Modify:** `ingestion/parser/heading_mapper.py`

---

### MED-02: No Equation Detection

**Gap:** Only chemical formulas detected. Math equations (LaTeX) not handled.

**Impact:** Physics/math content not properly indexed.

**Files to Modify:** `ingestion/parser/formula_detector.py`

---

### MED-03: No Image/Diagram Extraction

**Gap:** Figures and diagrams ignored. Only text extracted.

**Impact:** Visual educational content lost.

**Files to Create:** `ingestion/parser/image_extractor.py`

---

### MED-04: No Semantic Chunking Alternative

**Gap:** Only fixed-size recursive chunking. No semantic chunking (e.g., `semchunk`).

**Impact:** Topic boundaries may not align with chunk boundaries.

**Files to Create:** `ingestion/chunker/semantic_chunker.py`

---

### MED-05: No RAPTOR Hierarchical Chunking

**Gap:** No tree summarization for multi-hop queries.

**Impact:** Complex queries spanning multiple sections may fail.

**Reference:** RAPTOR paper (Sarthi et al., 2024)

---

### MED-06: No BERTScore Library Integration

**Gap:** `calculate_bertscore_f1()` uses character-level F1, not actual BERT embeddings.

**Impact:** Generation quality metrics less accurate.

**Fix:** Integrate `bert-score` Python package.

**Files to Modify:** `evaluate_complete_rag.py`

---

### MED-07: No RAGAS Library Integration

**Gap:** Custom implementations of Context Recall/Precision instead of `ragas` package.

**Impact:** Metrics may not match industry standard.

**Files to Modify:** `evaluate_complete_rag.py`

---

### MED-08: No Evaluation on Real Science PDF

**Gap:** Tests use mock data, not `science class 8.pdf`.

**Impact:** Evaluation results not representative of actual performance.

**Files to Modify:** `test_continuous_improvement.py`

---

### MED-09: No Prompt Templates Defined

**Gap:** Prompts hardcoded in handler. No template management.

**Impact:** Prompt changes require code deployment.

**Files to Create:** `orchestrator/prompts/templates.go`

---

### MED-10: No Golden Dataset Integration

**Gap:** `qa_pairs=[]` in evaluation. No golden Q&A dataset loaded.

**Impact:** Cannot measure answer correctness accurately.

**Files to Create:** `data/golden_qa_dataset.jsonl`

---

### MED-11: No Noise Robustness Testing

**Requirement (rag_evaluation_metrics.pdf Page 34):**
> "Inject 1-3 topically unrelated documents into top-k retrieved set"

**Gap:** No robustness stress tests.

**Impact:** Unknown how system handles noisy retrieval.

**Files to Create:** `eval/robustness_tests.py`

---

### MED-12: No Counterfactual Testing

**Requirement (rag_evaluation_metrics.pdf Page 34):**
> "Replace key factual claims with plausible but wrong alternatives"

**Gap:** No counterfactual injection tests.

**Impact:** Unknown if LLM uses context or parametric memory.

**Files to Create:** `eval/counterfactual_tests.py`

---

### MED-13: No Cross-Chunk Conflict Testing

**Gap:** No tests for contradictory information in retrieved chunks.

**Impact:** Unknown how system handles conflicting sources.

**Files to Create:** `eval/conflict_tests.py`

---

### MED-14: No Cost Tracking Per Query

**Gap:** No Vertex AI token usage tracking.

**Impact:** Cannot optimize for cost efficiency.

**Files to Modify:** `orchestrator/observability/metrics.go`

---

### MED-15: No Cache Hit Rate Metrics

**Gap:** No caching, so no hit rate tracking.

**Impact:** Cannot measure caching effectiveness.

**Files to Create:** `orchestrator/cache/metrics.go`

---

### MED-16: No Coverage Rate Tracking

**Gap:** No tracking of % queries answered vs abstained.

**Impact:** Unknown if corpus is sufficient.

**Files to Modify:** `orchestrator/handler/rag_handler.go`

---

### MED-17: No Go PDF Parser (Hybrid Dependency)

**Gap:** Go ingestion depends on Python parser. Not a single binary.

**Impact:** More complex deployment, two runtime dependencies.

**Files to Create:** `ingestion-go/parser/pdf_parser.go` (using `pdfcpu`)

---

### MED-18: No Terraform State Backend Configured

**Gap:** `backend "gcs"` commented out in `main.tf`.

**Impact:** Terraform state stored locally, not versioned.

**Files to Modify:** `terraform/main.tf`

---

## Low Priority Gaps (Nice-to-Have)

### LOW-01: No Chunk Size Distribution Analysis

**Gap:** `chunk_size_stats` calculated but not analyzed for optimization.

---

### LOW-02: No Information Density Optimization

**Gap:** Density metric exists but no threshold-based chunking.

---

### LOW-03: No Overlap Leakage Measurement

**Gap:** Metric exists but not used to tune overlap parameters.

---

### LOW-04: No Redundancy Score Monitoring

**Gap:** No alerts for high redundancy (>0.5).

---

### LOW-05: No MRR Trend Analysis

**Gap:** MRR calculated but not tracked over time.

---

### LOW-06: No NDCG Delta Monitoring

**Gap:** Reranking improvement not tracked.

---

### LOW-07: No Latency Trend Dashboards

**Gap:** No Grafana/Cloud Monitoring dashboards.

---

### LOW-08: No Token Usage Breakdown

**Gap:** Total tokens tracked but not split by prompt/completion.

---

### LOW-09: No Query Complexity Analysis

**Gap:** No classification of queries by difficulty.

---

## Hardcoded Values That Should Be Configurable

| File | Line | Hardcoded Value | Should Be Config |
|------|------|-----------------|------------------|
| `orchestrator/handler/rag_handler.go` | 45 | `450ms` timeout | `REQUEST_TIMEOUT_MS` |
| `orchestrator/handler/rag_handler.go` | 230 | `20ms` streaming delay | Remove (mock only) |
| `orchestrator/retrieval/hybrid_search.go` | 200 | `k=60` for RRF | `RRF_K` |
| `orchestrator/retrieval/hybrid_search.go` | 75 | `topK*20` dense limit | `DENSE_SEARCH_LIMIT` |
| `ingestion/chunker/parent_child.go` | 250 | `1500` parent chars | `PARENT_MAX_CHARS` |
| `ingestion/chunker/parent_child.go` | 251 | `512` child chars | `CHILD_MAX_CHARS` |
| `ingestion/chunker/parent_child.go` | 252 | `77` child overlap | `CHILD_OVERLAP` |
| `evaluate_complete_rag.py` | 180 | `0.75` hardcoded NDCG | Remove (calculate) |
| `quality_loop/judge.py` | 45 | `gemini-2.0-flash` | `JUDGE_MODEL` |
| `test_continuous_improvement.py` | 50 | `top_k=3,5` | Configurable per test |

---

## Missing Error Handling, Logging, and Monitoring

### Error Handling Gaps

| Component | Missing Error Handling | Risk |
|-----------|----------------------|------|
| `orchestrator/handler/rag_handler.go` | No retry on Vertex AI timeout | Request fails on transient errors |
| `ingestion/writer/alloydb_writer.py` | No circuit breaker on DB failures | Cascade failures under load |
| `quality_loop/judge.py` | No fallback on LLM Judge failure | Quality loop stalls |
| `ingestion/embedder/vertex_batch.py` | No exponential backoff | API rate limit violations |

### Logging Gaps

| Component | Missing Logging | Impact |
|-----------|----------------|--------|
| `orchestrator/handler/rag_handler.go` | No structured query logging | Cannot debug failed queries |
| `orchestrator/retrieval/hybrid_search.go` | No retrieval debug logs | Cannot tune search quality |
| `ingestion/pipeline.py` | No progress logging for large PDFs | Unknown ingestion status |
| `quality_loop/judge.py` | No judge decision logging | Cannot audit golden responses |

### Monitoring Gaps

| Metric | Missing Monitor | SLA Impact |
|--------|----------------|------------|
| TTFT p99 | No Cloud Monitoring alert | SLA violations undetected |
| Error rate | No PagerDuty integration | Outages not alerted |
| Recall@5 | No quality degradation alert | Quality drift undetected |
| DLQ depth | No queue depth alert | Backlog grows unnoticed |
| Redis memory | No OOM warning | Session loss risk |

---

## Scalability and Production-Readiness Gaps

### Scalability Issues

| Issue | Current State | Production Risk |
|-------|--------------|-----------------|
| **Connection Pooling** | Direct DB connections | Pool exhaustion at 400+ users |
| **Response Caching** | No caching | 3-5× higher latency for repeated queries |
| **Horizontal Scaling** | Cloud Run min=2 | Scale-up latency may breach TTFT |
| **Database Read Scaling** | Single primary | Query latency increases with load |
| **Embedding Batching** | Batch size=5 | Underutilized API quota |

### Production-Readiness Issues

| Issue | Status | Blocker For |
|-------|--------|-------------|
| **Load Testing** | Not done | Production deployment |
| **Disaster Recovery** | No runbook | Incident response |
| **Rollback Plan** | Not defined | Safe deployment |
| **On-Call Rotation** | Not set up | 24/7 support |
| **Runbooks** | Missing | Operational readiness |

---

## Specific Code Files That Need Changes

### Critical Files (Must Fix)

| File | Lines to Change | Priority | Effort |
|------|----------------|----------|--------|
| `orchestrator/handler/rag_handler.go` | ~200 | CRITICAL | 2 days |
| `orchestrator/observability/tracer.go` | NEW FILE | CRITICAL | 1 day |
| `evaluate_complete_rag.py` | ~150 | CRITICAL | 2 days |
| `quality_loop/judge.py` | ~100 | CRITICAL | 1 day |
| `orchestrator/cache/response_cache.go` | NEW FILE | CRITICAL | 1 day |

### High Priority Files

| File | Lines to Change | Priority | Effort |
|------|----------------|----------|--------|
| `ingestion/pipeline.py` | ~80 | HIGH | 1 day |
| `ingestion/writer/alloydb_writer.go` | ~60 | HIGH | 0.5 days |
| `terraform/main.tf` | ~100 | HIGH | 1 day |
| `orchestrator/middleware/auth.go` | NEW FILE | HIGH | 0.5 days |
| `eval/load_test.js` | NEW FILE | HIGH | 1 day |

### Medium Priority Files

| File | Lines to Change | Priority | Effort |
|------|----------------|----------|--------|
| `ingestion/parser/docx_parser.py` | NEW FILE | MEDIUM | 1 day |
| `ingestion/parser/markdown_parser.py` | NEW FILE | MEDIUM | 0.5 days |
| `evaluate_complete_rag.py` (BERTScore) | ~30 | MEDIUM | 0.5 days |
| `orchestrator/prompts/templates.go` | NEW FILE | MEDIUM | 0.5 days |

---

## Recommended Priority Order for Fixes

### Phase 0: Critical Production Blockers (Week 1)

**Goal:** Make pipeline functional and measurable

1. **CRIT-01:** Implement actual Gemini streaming (2 days)
2. **CRIT-04:** Add OpenTelemetry instrumentation (1 day)
3. **CRIT-03:** Wire evaluation to real data (2 days)
4. **CRIT-06:** Fix LLM Judge DB integration (1 day)

**Exit Criteria:**
- ✅ Real queries return real answers
- ✅ TTFT measurable via Cloud Trace
- ✅ Evaluation metrics reflect actual quality

### Phase 1: High Priority Production Readiness (Week 2)

**Goal:** Meet production requirements

5. **CRIT-02:** Implement query rewriting (0.5 days)
6. **CRIT-05:** Integrate hybrid search with evaluation (1 day)
7. **CRIT-07:** Add response caching (1 day)
8. **CRIT-08:** Create load tests (1 day)
9. **HIGH-06:** Deploy PgBouncer sidecar (0.5 days)
10. **HIGH-07:** Add JWT middleware (0.5 days)

**Exit Criteria:**
- ✅ p99 TTFT < 500ms at 100 concurrent users
- ✅ Cache hit rate > 30% for common queries
- ✅ Grade-level isolation enforced

### Phase 2: Quality and Observability (Week 3)

**Goal:** Production-grade observability and quality

11. **HIGH-01:** Add DOCX/Markdown ingestion (1 day)
12. **HIGH-02:** Implement deduplication (0.5 days)
13. **HIGH-09:** Add feedback endpoint (0.5 days)
14. **HIGH-10:** Schedule quality loop (0.5 days)
15. **MED-06:** Integrate BERTScore library (0.5 days)
16. **MED-07:** Integrate RAGAS library (0.5 days)

**Exit Criteria:**
- ✅ Multi-format ingestion working
- ✅ Feedback loop automated
- ✅ Industry-standard metrics

### Phase 3: Scalability and Hardening (Week 4)

**Goal:** Scale to 1,000 concurrent users

17. **HIGH-05:** Add read replicas (0.5 days)
18. **HIGH-03:** Implement incremental ingestion (1 day)
19. **HIGH-04:** Add database migrations (0.5 days)
20. **MED-09:** Create prompt templates (0.5 days)
21. **MED-10:** Create golden dataset (1 day)
22. **MED-11/12/13:** Add robustness tests (1 day)

**Exit Criteria:**
- ✅ p99 TTFT < 500ms at 1,000 concurrent users
- ✅ Content updates without full re-ingestion
- ✅ Quality ≥ 0.75 on golden dataset

---

## Summary Statistics

| Metric | Count |
|--------|-------|
| **Total Gaps Identified** | 47 |
| **Critical Gaps** | 8 |
| **High Priority Gaps** | 12 |
| **Medium Priority Gaps** | 18 |
| **Low Priority Gaps** | 9 |
| **Files to Create** | 15 |
| **Files to Modify** | 20 |
| **Estimated Fix Effort** | 25 days |
| **Minimum Viable Production** | 10 days (Phase 0 + Phase 1) |

---

## Conclusion

The RAG pipeline implementation demonstrates **strong architectural foundations** with well-documented code, comprehensive schema design, and thoughtful Go-first architecture. However, **critical production gaps** exist in:

1. **Actual LLM integration** (mock responses instead of Gemini)
2. **Observability** (no OpenTelemetry instrumentation)
3. **Evaluation validity** (metrics with zero data coverage)
4. **Quality loop** (placeholder DB operations)

**Recommendation:** Complete Phase 0 (Critical Blockers) and Phase 1 (High Priority) before any production deployment. This requires **~10 days of focused engineering effort** to achieve minimum viable production readiness.

**Risk Assessment:**
- **Deploy Now:** High risk of production failure, unmeasurable SLA, poor student experience
- **After Phase 1:** Moderate risk, measurable SLA, functional pipeline
- **After Phase 3:** Low risk, production-grade system ready for 1,000+ concurrent users

---

**Audit Completed:** March 27, 2026  
**Next Review:** After Phase 1 completion  
**Audit Owner:** Engineering Team
