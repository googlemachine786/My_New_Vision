# RAG COST OPTIMIZATION: CUT SPENDING BY 90%

**Date:** November 12, 2025  
**Read Time:** 11 min  
**Author:** Ailog Research Team  
**Level:** Intermediate  

**Summary:** Reduce RAG costs from $10k to $1k/month: smart chunking, caching, model selection, and batch processing.

---

## COST BREAKDOWN (TYPICAL RAG)
Per 1M queries:
* Embeddings: $100
* Vector DB: $200
* LLM calls: $5,000
* **Total: $5,300**

---

## 1. REDUCE EMBEDDING COSTS
Use smaller models:
```python
# Before: text-embedding-3-large
# Cost: $0.13 / 1M tokens
# Dimensions: 3072

# After: text-embedding-3-small
# Cost: $0.02 / 1M tokens (6.5x cheaper)
# Dimensions: 1536
# Performance: -5% accuracy for most use cases

import openai

embeddings = openai.Embedding.create(
    input=texts,
    model="text-embedding-3-small"  # 6.5x cheaper
)
```

Or use open-source models:
```python
# Free embeddings (self-hosted)
from sentence_transformers import SentenceTransformer

model = SentenceTransformer('BAAI/bge-small-en-v1.5')
embeddings = model.encode(texts)  # $0 cost
```

---

## 2. SMART CHUNKING
Fewer chunks = lower costs:
```python
# Before: 500-token chunks -> 10,000 chunks
chunk_size = 500
# Embedding cost: $100
# Storage cost: $50

# After: 800-token chunks -> 6,250 chunks (37.5% fewer)
chunk_size = 800
# Embedding cost: $65 (-35%)
# Storage cost: $32 (-36%)

# Trade-off: Slightly less precise, but huge savings
```

---

## 3. AGGRESSIVE CACHING
Cache everything:
```python
import redis
import hashlib

redis_client = redis.Redis(host='localhost', port=6379)

def cached_rag(query):
    # Check cache (90% hit rate -> 90% cost savings)
    cache_key = f"rag:{hashlib.md5(query.encode()).hexdigest()}"
    cached = redis_client.get(cache_key)

    if cached:
        return cached.decode()  # $0 cost

    # Cache miss - do full RAG
    response = expensive_rag_pipeline(query)

    # Store for 24 hours
    redis_client.setex(cache_key, 86400, response)

    return response

# With 90% cache hit rate:
# Before: $5,300/month
# After: $530/month (-90%)
```

---

## 4. USE SMALLER LLMS
```python
# Before: GPT-4 Turbo
# Cost: $10/1M input tokens, $30/1M output tokens

# After: GPT-4o-mini
# Cost: $0.15/1M input, $0.60/1M output (60x cheaper)
# Performance: 80-90% as good for most RAG tasks

import openai

response = openai.ChatCompletion.create(
    model="gpt-4o-mini",  # 60x cheaper
    messages=[...]
)

# Or even cheaper: GPT-3.5 Turbo
# Cost: $0.50/1M input, $1.50/1M output
```

---

## 5. REDUCE CONTEXT SIZE
Fewer tokens to LLM = lower cost:
```python
# Before: Send top 10 docs (5000 tokens)
context = "\n\n".join(retrieve(query, k=10))
# Cost: 5000 tokens * $10/1M = $0.05 per query

# After: Send top 3 docs (1500 tokens)
context = "\n\n".join(retrieve(query, k=3))
# Cost: 1500 tokens * $10/1M = $0.015 per query (-70%)

# Or summarize context first
def compress_context(docs):
    summaries = []
    for doc in docs:
        summary = openai.ChatCompletion.create(
            model="gpt-4o-mini",  # Cheap model for summarization
            messages=[{
                "role": "user",
                "content": f"Summarize in 50 words: {doc}"
            }]
        )
        summaries.append(summary.choices[0].message.content)

    return "\n\n".join(summaries)
```

---

## 6. BATCH PROCESSING
Process multiple queries together:
```python
# Instead of 1000 individual API calls
for query in queries:
    embed(query)  # 1000 calls

# Batch embed
batch_embeddings = openai.Embedding.create(
    input=queries,  # Single call
    model="text-embedding-3-small"
)

# Savings: Reduced latency overhead
```

---

## 7. SELF-HOSTED VECTOR DB
```python
# Before: Pinecone
# Cost: $70/month for 1M vectors

# After: Qdrant (self-hosted)
# Cost: $20/month (DigitalOcean droplet)
# Savings: $50/month (-71%)

docker run -p 6333:6333 qdrant/qdrant
```

---

## 8. LAZY RERANKING
Only rerank when necessary:
```python
def smart_rerank(query, candidates):
    # If top result has high score, skip reranking
    if candidates[0].score > 0.9:
        return candidates[:5]  # Skip expensive reranking

    # Otherwise, rerank
    return rerank(query, candidates)

# Savings: 50% fewer reranking calls
```

---

## 9. USER QUOTAS
Prevent abuse:
```python
import time

user_quotas = {}  # {user_id: [timestamp, timestamp, ...]}

def rate_limit(user_id, max_queries=100, window=3600):
    now = time.time()

    # Remove old queries outside window
    if user_id in user_quotas:
        user_quotas[user_id] = [
            ts for ts in user_quotas[user_id]
            if now - ts < window
        ]
    else:
        user_quotas[user_id] = []

    # Check limit
    if len(user_quotas[user_id]) >= max_queries:
        raise Exception("Rate limit exceeded")

    # Add current query
    user_quotas[user_id].append(now)
```

---

## 10. MONITORING & ALERTS
Track costs in real-time:
```python
import prometheus_client

# Track costs
embedding_cost = prometheus_client.Counter(
    'rag_embedding_cost_usd',
    'Total embedding API costs'
)

llm_cost = prometheus_client.Counter(
    'rag_llm_cost_usd',
    'Total LLM API costs'
)

def track_embedding_cost(tokens):
    cost = tokens / 1_000_000 * 0.02  # $0.02/1M tokens
    embedding_cost.inc(cost)

def track_llm_cost(input_tokens, output_tokens):
    cost = (input_tokens / 1_000_000 * 0.15) + (output_tokens / 1_000_000 * 0.60)
    llm_cost.inc(cost)

# Set alerts when cost > $1000/day
```

---

## COMPLETE COST OPTIMIZATION
```python
@cached  # 90% cache hit
def optimized_rag(query):
    # 1. Cheap embeddings
    query_emb = open_source_embed(query)  # Free

    # 2. Efficient retrieval (fewer docs)
    docs = vector_db.search(query_emb, limit=3)  # Not 10

    # 3. Smart reranking (only if needed)
    if docs[0].score < 0.9:
        docs = fast_rerank(query, docs)  # TinyBERT, not GPT-4

    # 4. Compress context
    context = compress_context(docs)  # 500 tokens, not 5000

    # 5. Cheap LLM
    response = openai.ChatCompletion.create(
        model="gpt-4o-mini",  # 60x cheaper
        messages=[{
            "role": "user",
            "content": f"Context: {context}\n\nQ: {query}"
        }]
    )

    return response.choices[0].message.content

# Cost reduction:
# - Embeddings: -100% (self-hosted)
# - Vector DB: -71% (self-hosted)
# - LLM: -60% (smaller model)
# - Cache: -90% (fewer calls)
# Total: ~95% cost reduction
```
Smart optimizations can cut RAG costs by 90%+ without sacrificing quality.

---

**TAGS**  
`optimization` `cost` `budget` `efficiency`

**RELATED POSTS**
* **CONTEXT WINDOW OPTIMIZATION: MANAGING TOKEN LIMITS** (Advanced | 11 min) - Strategies for fitting more information in limited context windows: compression, summarization, smart selection, and window management techniques.
* **EVALUATING A RAG SYSTEM: METRICS AND METHODOLOGIES** (Advanced | 23 min) - Complete guide to measuring your RAG performance: faithfulness, relevancy, recall, and automated evaluation frameworks.
* **CACHING STRATEGIES TO REDUCE RAG LATENCY AND COST** (Intermediate | 10 min) - Cut costs by 80%: implement semantic caching, embedding caching, and response caching for production RAG.
