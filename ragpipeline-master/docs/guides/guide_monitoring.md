# RAG MONITORING AND OBSERVABILITY
**Date:** November 11, 2025 | **Read Time:** 12 min read | **Author:** Ailog Research Team

Monitor RAG systems in production: track latency, costs, accuracy, and user satisfaction with metrics and dashboards.

## KEY METRICS TO TRACK
**Performance:**
* Latency (p50, p95, p99)
* Throughput (queries/second)
* Error rate

**Quality:**
* Retrieval accuracy
* Answer quality
* User feedback

**Cost:**
* API costs per query
* Storage costs
* Compute costs

## PROMETHEUS + GRAFANA SETUP
```yaml
# docker-compose.yml
version: '3.8'
services:
  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
```

## INSTRUMENTING RAG PIPELINE
```python
from prometheus_client import Counter, Histogram, Gauge, start_http_server
import time

# Metrics
query_counter = Counter('rag_queries_total', 'Total RAG queries')
latency_histogram = Histogram('rag_latency_seconds', 'RAG latency', buckets=[0.1, 0.5, 1, 2, 5])
error_counter = Counter('rag_errors_total', 'Total RAG errors', ['error_type'])
cost_counter = Counter('rag_cost_usd', 'Total costs', ['component'])
relevance_gauge = Gauge('rag_relevance_score', 'Average relevance score')

def instrumented_rag(query):
    query_counter.inc()
    start = time.time()

    try:
        # Embedding
        emb_start = time.time()
        embedding = embed(query)
        emb_time = time.time() - emb_start
        cost_counter.labels(component='embedding').inc(0.001)

        # Retrieval
        ret_start = time.time()
        docs = vector_db.search(embedding, limit=5)
        ret_time = time.time() - ret_start

        # LLM
        llm_start = time.time()
        response = llm_call(query, docs)
        llm_time = time.time() - llm_start
        cost_counter.labels(component='llm').inc(0.01)

        # Track latency
        total_time = time.time() - start
        latency_histogram.observe(total_time)

        # Log breakdown
        logger.info(f"Latency breakdown: emb={emb_time:.3f}s, ret={ret_time:.3f}s, llm={llm_time:.3f}s")

        return response

    except Exception as e:
        error_counter.labels(error_type=type(e).__name__).inc()
        raise

# Start metrics server
start_http_server(8000)
```

## TRACKING RETRIEVAL QUALITY
```python
from sklearn.metrics import ndcg_score
import numpy as np

def track_retrieval_quality(query, retrieved_docs, ground_truth_docs):
    # Calculate nDCG
    relevance_scores = []
    for doc in retrieved_docs:
        if doc.id in ground_truth_docs:
            relevance_scores.append(1)
        else:
            relevance_scores.append(0)

    ndcg = ndcg_score([relevance_scores], [range(len(relevance_scores))])

    # Update metric
    relevance_gauge.set(ndcg)

    # Log to database
    db.insert({
        "timestamp": time.time(),
        "query": query,
        "ndcg": ndcg,
        "retrieved_ids": [doc.id for doc in retrieved_docs]
    })
```

## USER FEEDBACK COLLECTION
```python
from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI()

class Feedback(BaseModel):
    query_id: str
    rating: int  # 1-5
    helpful: bool
    comment: str | None

feedback_counter = Counter(
    'rag_feedback_total',
    'User feedback',
    ['rating']
)

@app.post("/feedback")
def collect_feedback(feedback: Feedback):
    # Track in Prometheus
    feedback_counter.labels(rating=str(feedback.rating)).inc()

    # Store in database
    db.insert({
        "query_id": feedback.query_id,
        "rating": feedback.rating,
        "helpful": feedback.helpful,
        "comment": feedback.comment,
        "timestamp": time.time()
    })

    return {"status": "ok"}
```

## DISTRIBUTED TRACING WITH OPENTELEMETRY
```python
from opentelemetry import trace
from opentelemetry.sdk.trace import TracerProvider
from opentelemetry.sdk.trace.export import BatchSpanProcessor
from opentelemetry.exporter.jaeger.thrift import JaegerExporter

# Setup tracing
trace.set_tracer_provider(TracerProvider())
jaeger_exporter = JaegerExporter(
    agent_host_name="localhost",
    agent_port=6831,
)
trace.get_tracer_provider().add_span_processor(
    BatchSpanProcessor(jaeger_exporter)
)

tracer = trace.get_tracer(__name__)

def traced_rag(query):
    with tracer.start_as_current_span("rag_pipeline") as span:
        span.set_attribute("query", query)

        # Embedding
        with tracer.start_as_current_span("embedding"):
            embedding = embed(query)

        # Retrieval
        with tracer.start_as_current_span("retrieval"):
            docs = vector_db.search(embedding, limit=5)
            span.set_attribute("num_docs_retrieved", len(docs))

        # Reranking
        with tracer.start_as_current_span("reranking"):
            reranked = rerank(query, docs)

        # LLM
        with tracer.start_as_current_span("llm_generation"):
            response = llm_call(query, reranked)
            span.set_attribute("response_length", len(response))

        return response
```

## LOGGING BEST PRACTICES
```python
import structlog
import json

logger = structlog.get_logger()

def logged_rag(query, user_id):
    logger.info(
        "rag_query_started",
        query=query,
        user_id=user_id,
        timestamp=time.time()
    )

    try:
        # Retrieval
        docs = retrieve(query)
        logger.info(
            "documents_retrieved",
            num_docs=len(docs),
            doc_ids=[d.id for d in docs],
            scores=[d.score for d in docs]
        )

        # LLM
        response = llm_call(query, docs)
        logger.info(
            "rag_query_completed",
            response_length=len(response),
            tokens_used=estimate_tokens(query, docs, response)
        )

        return response

    except Exception as e:
        logger.error(
            "rag_query_failed",
            error=str(e),
            error_type=type(e).__name__
        )
        raise
```

## ALERTING RULES
```yaml
# prometheus_alerts.yml
groups:
  - name: rag_alerts
    interval: 30s
    rules:
      # High latency
      - alert: HighRAGLatency
        expr: histogram_quantile(0.95, rag_latency_seconds) > 2
        for: 5m
        annotations:
          summary: "RAG p95 latency > 2s"

      # High error rate
      - alert: HighErrorRate
        expr: rate(rag_errors_total[5m]) > 0.05
        for: 2m
        annotations:
          summary: "RAG error rate > 5%"

      # Low relevance
      - alert: LowRelevance
        expr: rag_relevance_score < 0.7
        for: 10m
        annotations:
          summary: "RAG relevance score < 0.7"

      # High costs
      - alert: HighDailyCost
        expr: increase(rag_cost_usd[24h]) > 100
        annotations:
          summary: "RAG costs > $100/day"
```

## DASHBOARD QUERIES (GRAFANA)
```promql
# Average latency over time
rate(rag_latency_seconds_sum[5m]) / rate(rag_latency_seconds_count[5m])

# Query throughput
rate(rag_queries_total[1m])

# Error rate
rate(rag_errors_total[5m]) / rate(rag_queries_total[5m])

# Cost per day
increase(rag_cost_usd[24h])

# User satisfaction
avg(rag_feedback_total{rating="5"}) / sum(rag_feedback_total)
```

## A/B TESTING FRAMEWORK
```python
import random

def ab_test_rag(query, user_id):
    # Assign user to variant
    variant = "A" if hash(user_id) % 2 == 0 else "B"

    if variant == "A":
        # Control: current RAG
        response = rag_pipeline_v1(query)
    else:
        # Treatment: new RAG
        response = rag_pipeline_v2(query)

    # Track variant
    variant_counter.labels(variant=variant).inc()

    # Log for analysis
    db.insert({
        "query": query,
        "user_id": user_id,
        "variant": variant,
        "timestamp": time.time()
    })

    return response, variant
```

**Conclusion:** Production RAG needs monitoring. Track everything, alert on anomalies, iterate based on data.

**Tags:** `optimization` `monitoring` `observability` `production`
