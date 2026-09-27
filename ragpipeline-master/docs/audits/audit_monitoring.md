# Audit: RAG Monitoring Guide vs. Implementation

**Guide:** `guide_monitoring.md` (Ailog, Nov 2025)
**Audited:** 2026-04-03
**Scope:** C:\Users\kommi\ragpipeline\

---

## Guide Summary

This guide covers monitoring and observability for production RAG systems: key metrics (latency, throughput, error rate, retrieval accuracy, answer quality, user feedback, costs), Prometheus + Grafana setup, OpenTelemetry distributed tracing, structured logging, alerting rules, user feedback collection, and A/B testing.

---

## Key Recommendations and Implementation Match

### 1. Key Metrics: Performance

| Metric | Status | Details |
|--------|--------|---------|
| **Latency (p50, p95, p99)** | **MATCH** | TTFT and total latency tracked as histograms with explicit bucket boundaries. |
| **Throughput (queries/second)** | **MATCH** | `rag_requests_total` counter can be used to derive throughput via `rate()`. |
| **Error rate** | **MATCH** | `rag_errors_total` counter with `stage` and `error_type` labels. |
| **Code Reference** | `C:\Users\kommi\ragpipeline\orchestrator\observability\metrics.go` lines 36-42: `ttftHistogram`, `latencyHistogram`. Lines 59-65: `requestsCounter`, `errorsCounter`. |

### 2. Key Metrics: Quality

| Metric | Status | Details |
|--------|--------|---------|
| **Retrieval accuracy** | **PARTIALLY** | Recall@K, MRR, NDCG implemented in `rrf.go` but not exposed as live metrics. They are used in benchmark evaluation, not real-time monitoring. |
| **Answer quality** | **DOES NOT MATCH** | No real-time answer quality metric. The LLM Judge in `judge.go` evaluates negative feedback but is not connected to monitoring. |
| **User feedback** | **MATCH** | Feedback loop table exists. The guide_monitoring.md shows a `/feedback` endpoint for collecting ratings. |
| **Code Reference** | `C:\Users\kommi\ragpipeline\orchestrator\retrieval\rrf.go`: evaluation metrics (not real-time). `orchestrator/quality/judge.go`: feedback evaluation. |

### 3. Key Metrics: Cost

| Metric | Status | Details |
|--------|--------|---------|
| **API costs per query** | **DOES NOT MATCH** | Cost tracking infrastructure exists (`cost_counter` in guide code, but not in actual `metrics.go`). No actual cost-per-query calculation. |
| **Storage costs** | **DOES NOT MATCH** | No storage cost tracking. |
| **Compute costs** | **DOES NOT MATCH** | No compute cost tracking. |

### 4. Prometheus + Grafana Setup

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Docker Compose with Prometheus and Grafana services | |
| **Implementation** | **PARTIALLY MATCHES** | `C:\Users\kommi\ragpipeline\docker-compose.yml` exists but its contents need to be checked. OpenTelemetry metrics are instrumented but the Prometheus/Grafana infrastructure setup is not verified. |
| **Code Reference** | `docker-compose.yml` at project root. |

### 5. Instrumenting RAG Pipeline

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Track embedding time, retrieval time, LLM time, costs separately | |
| **Implementation** | **MATCH** | OpenTelemetry spans for embedding, hybrid search, RRF fusion, generation. TTFT and total latency as histograms. |
| **Code Reference** | `metrics.go` lines 30-72: All metrics defined. `tracer.go` lines 84-93: Span stage names defined. `metrics.go` lines 117-141: `TimingHelper` for per-stage timing. |

### 6. Distributed Tracing with OpenTelemetry

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Trace the full RAG pipeline with spans for embedding, retrieval, reranking, generation | |
| **Implementation** | **MATCH** | OpenTelemetry tracer initialized with Cloud Trace exporter. Spans for all pipeline stages. |
| **Code Reference** | `tracer.go` lines 22-57: `InitTracer()` with OTLP exporter to `cloudtrace.googleapis.com:443`. Lines 60-65: `StartSpan()`. Lines 84-93: Stage span constants. |

### 7. Structured Logging

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Use structured logging (structlog) with query context, retrieved doc info, error details | |
| **Implementation** | **MATCH** | `structlog` used throughout the Python codebase. Go uses `zerolog` for structured logging. |
| **Code Reference** | `parent_child.py` line 19: `logger = structlog.get_logger()`. `vertex_batch.py` line 17: `logger = structlog.get_logger()`. `hybrid_search.go` line 12: `"github.com/rs/zerolog/log"`. |

### 8. User Feedback Collection

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | POST endpoint for feedback (rating, helpful boolean, comment), track in Prometheus | |
| **Implementation** | **MATCH** | Feedback loop implemented via `ai_feedback_loop` table. LLM Judge processes negative feedback. |
| **Code Reference** | `judge.go` lines 20-30: `FeedbackEvent` struct. `judge.go` lines 66-108: `JudgeFeedback()` method. |

### 9. Alerting Rules

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Prometheus alerting rules for high latency, high error rate, low relevance, high costs | |
| **Implementation** | **DOES NOT MATCH** | No Prometheus alerting rules configuration found. No YAML alerting rules. |
| **Missing** | No `prometheus_alerts.yml` or equivalent. No alert definitions for latency, error rate, relevance, or cost thresholds. |

### 10. Grafana Dashboard Queries

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Grafana dashboards with PromQL queries for latency, throughput, error rate, cost, satisfaction | |
| **Implementation** | **DOES NOT MATCH** | No Grafana dashboard configurations (JSON) found. |
| **Missing** | No Grafana dashboard JSON files, no PromQL query definitions. |

### 11. A/B Testing Framework

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Assign users to variants, track variant usage, log for analysis | |
| **Implementation** | **DOES NOT MATCH** | No A/B testing framework found. |
| **Missing** | No user variant assignment, no A/B test tracking, no statistical analysis of results. |

### 12. Cache Hit/Miss Tracking

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Track cache hit rate | |
| **Implementation** | **MATCH** | `rag_cache_hits_total` and `rag_cache_misses_total` counters. |
| **Code Reference** | `metrics.go` lines 50-56: `cacheHitsCounter`, `cacheMissesCounter`. Lines 145-153: `RecordCacheHit()`, `RecordCacheMiss()`. |

### 13. Retrieval Quality Tracking

| Aspect | Status | Details |
|--------|--------|---------|
| **Recommendation** | Calculate nDCG for queries with known ground truth, track over time | |
| **Implementation** | **PARTIALLY MATCHES** | NDCG calculation exists in `rrf.go` and `bench_and_improve.py`, but it is not tracked as a real-time Prometheus metric. It is used in offline evaluation only. |
| **Missing** | No `relevance_gauge` Prometheus metric for live nDCG tracking. |

---

## Gap Analysis Summary

### Implemented
- OpenTelemetry metrics (TTFT, latency, requests, errors, cache hits/misses)
- OpenTelemetry distributed tracing with Cloud Trace exporter
- Structured logging (structlog for Python, zerolog for Go)
- User feedback collection and LLM-based feedback evaluation
- Cache hit/miss tracking
- Per-stage timing helper

### Partially Implemented
- Retrieval quality metrics (calculated but not exposed as live metrics)
- Docker Compose setup (exists but Prometheus/Grafana services not verified)

### NOT Implemented
- Real-time cost tracking in USD
- Prometheus alerting rules
- Grafana dashboard configurations
- A/B testing framework
- Live relevance score tracking (nDCG, precision)
- Storage cost monitoring
- Answer quality real-time measurement

---

## Overall Assessment: PARTIALLY MATCHES

The project has a strong monitoring foundation with OpenTelemetry metrics, distributed tracing, and structured logging. Cache tracking and feedback collection are implemented. The main gaps are in alerting (no Prometheus rules defined), visualization (no Grafana dashboards), cost tracking (no actual cost calculations), and A/B testing. The observability infrastructure is well-designed but not fully operationalized with dashboards and alerts.
