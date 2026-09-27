# Cache Augmented Generation (CAG) Architecture

## Overview

CAG dramatically reduces LLM costs by implementing **multi-layer caching** at every stage of the RAG pipeline. Instead of calling expensive APIs (Vertex AI embeddings, vector search, LLM generation) for every query, we cache intermediate results and reuse them for similar queries.

## Cost Centers in RAG Pipeline

| Stage | API | Cost per Call | Monthly Cost (10K queries/day) |
|-------|-----|---------------|--------------------------------|
| Embedding Generation | Vertex AI | $0.0001/1K tokens | ~$30/month |
| Vector Search | AlloyDB | $0.001/query | ~$300/month (compute) |
| Query Rewriting | Gemini | $0.00035/1K tokens | ~$10/month |
| **LLM Response Generation** | **Gemini** | **$0.0025/1K tokens** | **~$750/month** |
| **Total** | | | **~$1,090/month** |

**With CAG (80% hit rate):** ~$218/month (**80% cost reduction**)

---

## Multi-Layer CAG Architecture

```
Query → [Layer 1: Exact Match Cache] → Return if hit
         ↓
      [Layer 2: Semantic Cache] → Find similar queries (cosine similarity > 0.95)
         ↓ (if miss)
      [Layer 3: Embedding Cache] → Reuse embeddings for same text
         ↓
      [Layer 4: Search Results Cache] → Reuse retrieved chunks
         ↓
      [Layer 5: LLM Template Cache] → Cache response templates
         ↓ (if all miss)
      Call APIs → Cache all intermediate results → Return
```

---

## Layer 1: Exact Match Cache (Current)

**What it does:** Caches identical queries with same parameters

**Key:** `SHA256(query + grade + subject + filters)`  
**TTL:** 1 hour  
**Hit Condition:** Exact string match  
**Hit Rate:** ~15-20% (common questions)

**Already Implemented:** ✅ `services/api-gateway/cache/response_cache.go`

---

## Layer 2: Semantic Cache (NEW)

**What it does:** Finds semantically similar queries using embedding cosine similarity

**How it works:**
1. Generate embedding for incoming query
2. Search Redis vector index for similar cached queries
3. If cosine similarity > 0.95, return cached response
4. Otherwise, proceed with full pipeline

**Key:** Vector index in Redis (RediSearch)  
**TTL:** 6 hours  
**Hit Condition:** Cosine similarity > 0.95  
**Hit Rate:** ~40-50% (paraphrased questions)

**Example:**
```
Query 1: "What is quantum physics?"
Query 2: "Can you explain quantum physics?"
Query 3: "Tell me about quantum physics"

All 3 have similarity > 0.95 → Reuse same cached response
```

**Cost Saved:** Avoids embedding + search + LLM call

---

## Layer 3: Embedding Cache (NEW)

**What it does:** Caches embeddings to avoid redundant Vertex AI API calls

**How it works:**
1. Before calling Vertex AI, check if text embedding exists in cache
2. Key: `SHA256(text + task_type)`
3. If hit, return cached embedding (768-dim vector)
4. If miss, call Vertex AI and cache result

**Key:** `SHA256(text + task_type)`  
**TTL:** 24 hours (embeddings don't change)  
**Hit Rate:** ~60-70% (common phrases, repeated content)

**Example:**
```
"Ingestion pipeline" → Cached after first call
"Chunking strategy" → Cached after first call
Student queries often repeat same phrases
```

**Cost Saved:** $0.0001 per hit (Vertex AI embedding cost)

---

## Layer 4: Search Results Cache (NEW)

**What it does:** Caches retrieved chunks to avoid redundant vector database queries

**How it works:**
1. Before calling Vector Search Service, check cache
2. Key: `SHA256(embedding_hash + top_k + filters)`
3. If hit, return cached chunks
4. If miss, perform search and cache results

**Key:** `SHA256(embedding_hash + top_k + filters)`  
**TTL:** 30 minutes (chunks change with ingestion)  
**Hit Rate:** ~50-60% (similar queries retrieve same chunks)

**Cost Saved:** Database compute + I/O costs

---

## Layer 5: LLM Template Cache (NEW)

**What it does:** Caches LLM response templates for common query patterns

**How it works:**
1. Identify query intent (from Query Understanding Service)
2. For common intents, use cached response templates
3. Interpolate retrieved chunks into template
4. Only call LLM for novel responses

**Key:** `SHA256(intent + rewritten_query + chunk_hashes)`  
**TTL:** 12 hours  
**Hit Rate:** ~30-40% (explanatory queries)

**Example Templates:**
```
Intent: "definition"
Template: "{concept} is {definition}. It involves {key_aspects}. Examples include {examples}."

Intent: "comparison"
Template: "{A} and {B} differ in {differences}. While {A} has {A_features}, {B} has {B_features}."

Intent: "procedure"
Template: "To {goal}, follow these steps: {steps}. Important notes: {warnings}."
```

**Cost Saved:** $0.0025 per hit (Gemini LLM cost - most expensive!)

---

## Cache Invalidation Strategies

### 1. Time-Based (TTL)
- Embedding Cache: 24 hours
- Search Results: 30 minutes
- Response Cache: 1 hour
- Semantic Cache: 6 hours

### 2. Event-Based
- **Content Update:** Invalidate when new documents ingested
- **Schema Change:** Clear all caches on database migration
- **Model Update:** Invalidate when LLM model version changes

### 3. LRU Eviction
- Max entries per cache layer
- Evict least recently used when full
- Prevents memory exhaustion

### 4. Semantic Invalidation
- If cached response quality < threshold (from feedback)
- Auto-invalidates poorly rated cached responses

---

## Implementation Plan

### Phase 1: Embedding Cache
**File:** `services/embedding-service/cache/embedding_cache.go`
- Redis-backed embedding cache
- Key: `SHA256(text + task_type)`
- Store as Redis JSON array
- TTL: 24 hours

### Phase 2: Search Results Cache
**File:** `services/vector-search-service/cache/search_cache.go`
- Cache search results by embedding hash
- Key: `SHA256(embedding_hash + top_k + filters)`
- Store chunk IDs and scores
- TTL: 30 minutes

### Phase 3: Semantic Cache
**File:** `services/api-gateway/cache/semantic_cache.go`
- Redis vector index (RediSearch)
- Store query embeddings + responses
- Cosine similarity lookup
- Configurable threshold (default 0.95)

### Phase 4: LLM Template Cache
**File:** `services/api-gateway/cache/template_cache.go`
- Response templates by intent
- Template interpolation
- Fallback to full LLM if template unavailable

### Phase 5: CAG Orchestrator
**File:** `services/api-gateway/cache/cag_orchestrator.go`
- Coordinates all 5 cache layers
- Metrics collection
- Cache warming on startup
- Periodic cache cleanup

---

## Metrics & Monitoring

### Cache Metrics to Track
```go
type CAGMetrics struct {
    // Layer 1: Exact Match
    ExactMatchHits    int64
    ExactMatchMisses  int64
    
    // Layer 2: Semantic
    SemanticHits      int64
    SemanticMisses    int64
    AvgSimilarity     float64
    
    // Layer 3: Embedding
    EmbeddingCacheHits    int64
    EmbeddingCacheMisses  int64
    
    // Layer 4: Search
    SearchCacheHits    int64
    SearchCacheMisses  int64
    
    // Layer 5: Template
    TemplateHits    int64
    TemplateMisses  int64
    
    // Overall
    TotalCostSaved  float64  // USD
    TotalLLMCallsAvoided int64
}
```

### Dashboard
- Hit rate per layer
- Cost savings over time
- Cache size and eviction rate
- Average latency per cache hit vs miss
- Cache staleness (age of cached responses)

---

## Configuration

```env
# CAG Configuration
CAG_ENABLED=true

# Layer 1: Exact Match
EXACT_CACHE_TTL=1h
EXACT_CACHE_MAX_SIZE=1000

# Layer 2: Semantic
SEMANTIC_CACHE_ENABLED=true
SEMANTIC_CACHE_TTL=6h
SEMANTIC_CACHE_THRESHOLD=0.95
SEMANTIC_CACHE_MAX_SIZE=5000

# Layer 3: Embedding
EMBEDDING_CACHE_ENABLED=true
EMBEDDING_CACHE_TTL=24h
EMBEDDING_CACHE_MAX_SIZE=10000

# Layer 4: Search
SEARCH_CACHE_ENABLED=true
SEARCH_CACHE_TTL=30m
SEARCH_CACHE_MAX_SIZE=3000

# Layer 5: Template
TEMPLATE_CACHE_ENABLED=true
TEMPLATE_CACHE_TTL=12h
TEMPLATE_CACHE_MAX_SIZE=500

# Cache Invalidation
CACHE_INVALIDATE_ON_INGEST=true
CACHE_AUTO_INVALIDATE_LOW_QUALITY=true
CACHE_MIN_QUALITY_SCORE=0.7
```

---

## Expected Impact

### Cost Reduction
| Metric | Without CAG | With CAG (80% hit rate) | Savings |
|--------|-------------|-------------------------|---------|
| LLM Calls | 10,000/day | 2,000/day | 80% |
| Embedding Calls | 10,000/day | 3,000/day | 70% |
| Search Queries | 10,000/day | 4,000/day | 60% |
| **Monthly Cost** | **~$1,090** | **~$218** | **80%** |

### Latency Improvement
| Scenario | Without CAG | With CAG Hit | Improvement |
|----------|-------------|--------------|-------------|
| Full Pipeline | 2500ms | - | - |
| Exact Cache Hit | - | 50ms | 98% faster |
| Semantic Cache Hit | - | 150ms | 94% faster |
| Embedding Cache Hit | - | 2000ms | 20% faster |

### Scalability
- **Without CAG:** 100 concurrent users → 100 LLM calls
- **With CAG:** 100 concurrent users → 20 LLM calls (80% cached)
- **Result:** 5x more users with same infrastructure

---

## Redis Architecture

```
# Exact Match Cache
rag:cache:{sha256_hash} → CachedResponse (JSON, TTL 1h)

# Semantic Cache (RediSearch)
rag:semantic_cache FT.CREATE idx ON JSON ...
  - query_embedding: VECTOR (768-dim)
  - query_text: string
  - response: JSON
  - created_at: timestamp

# Embedding Cache
rag:embedding_cache:{sha256_hash} → [0.1, 0.2, ...] (JSON array, TTL 24h)

# Search Results Cache
rag:search_cache:{sha256_hash} → {chunk_ids: [...], scores: [...]} (JSON, TTL 30m)

# Template Cache
rag:template_cache:{intent}:{sha256_hash} → {template: "...", variables: [...]} (JSON, TTL 12h)

# Metrics
rag:metrics:cache_hits:{layer} → counter
rag:metrics:cache_misses:{layer} → counter
rag:metrics:cost_saved → float (USD)
```

---

## Next Steps

1. ✅ Architecture design (this document)
2. 🔲 Implement Embedding Cache (Layer 3)
3. 🔲 Implement Search Results Cache (Layer 4)
4. 🔲 Implement Semantic Cache (Layer 2)
5. 🔲 Implement Template Cache (Layer 5)
6. 🔲 Create CAG Orchestrator
7. 🔲 Update API Gateway query handler to use all layers
8. 🔲 Add metrics and monitoring
9. 🔲 Load test and measure actual hit rates
10. 🔲 Tune TTL and thresholds based on metrics

---

## Fallback Strategy

If any cache layer fails:
- **Non-critical:** Log error, skip that layer, continue
- **Cache read error:** Treat as miss, proceed to next layer
- **Cache write error:** Log error, continue without caching
- **Redis down:** All layers treated as misses, full pipeline

**Guarantee:** Cache failures never break the pipeline, only affect performance.
