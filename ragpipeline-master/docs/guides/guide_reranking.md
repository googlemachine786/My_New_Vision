# RERANKING FOR RAG: +40% ACCURACY WITH CROSS-ENCODERS (2025 GUIDE)

**February 10, 2025** | 11 min read | Ailog Research Team

Boost RAG accuracy by 40% using reranking. Complete guide to cross-encoders, Cohere Rerank API, and ColBERT for production retrieval systems.

---

### TL;DR
* **Reranking** = Second-pass scoring of retrieved docs for better precision
* **Cross-encoders** deliver 10-25% accuracy improvement over pure retrieval
* **Cohere Rerank API**: Easiest option ($1/1000 queries)
* **Self-hosted**: ms-marco cross-encoders (free, good quality)
* Compare rerankers on your data with Ailog

---

### THE RERANKING PROBLEM
Initial retrieval (vector search, BM25) casts a wide net to recall potentially relevant documents. However:
* **False positives:** Some retrieved chunks aren't actually relevant
* **Ranking quality:** Most relevant chunks may not be ranked first
* **Query-specific relevance:** Initial ranking doesn't account for query nuances

**Solution:** Rerank retrieved candidates with a more sophisticated model.

---

### TWO-STAGE RETRIEVAL
`Query -> [Stage 1: Retrieval] -> 100 candidates -> [Stage 2: Reranking] -> 10 best results -> [Stage 3: Generation] -> Answer`

**Why two stages?**
* **Retrieval:** Fast, scales to millions/billions of documents
* **Reranking:** Expensive but accurate, only on small candidate set
* **Best of both:** Speed + Quality

---

### RERANKING APPROACHES

#### CROSS-ENCODER MODELS
Unlike bi-encoders (embed query and document separately), cross-encoders process query and document together.

**Bi-encoder (Retrieval)**
```python
query_emb = embed(query)           # [768]
doc_emb = embed(document)          # [768]
score = cosine(query_emb, doc_emb)  # Similarity
```

**Cross-encoder (Reranking)**
```python
# Process together
input = f"[CLS] {query} [SEP] {document} [SEP]"
score = model(input)  # Direct relevance score
```

**Why cross-encoders are better:**
* Attention between query and document tokens
* Captures word-level interactions
* More accurate relevance scoring

**Why not use for retrieval:**
* Must score each query-document pair (O(n))
* Too slow for large collections
* No pre-computed embeddings

#### POPULAR CROSS-ENCODER MODELS

**ms-marco-MiniLM-L6-v2**
```python
from sentence_transformers import CrossEncoder

model = CrossEncoder('cross-encoder/ms-marco-MiniLM-L6-v2')

# Score query-document pairs
scores = model.predict([
    (query, doc1),
    (query, doc2),
    (query, doc3)
])

# Rerank by score
ranked_indices = np.argsort(scores)[::-1]
```
**Characteristics:**
* Size: 80MB
* Speed: ~50ms per batch
* Quality: Good for English
* Training: Trained on MS MARCO

**ms-marco-TinyBERT-L2-v2**
* Even smaller/faster
* Slight quality tradeoff
* Good for latency-critical apps

**mmarco-mMiniLMv2-L12-H384-v1**
* Multilingual support
* Similar performance to English models
* Supports 100+ languages

#### IMPLEMENTATION
```python
class RerankedRetriever:
    def __init__(self, base_retriever, reranker_model):
        self.retriever = base_retriever
        self.reranker = CrossEncoder(reranker_model)

    def retrieve(self, query, k=5, rerank_top_n=20):
        # Stage 1: Retrieve more candidates
        candidates = self.retriever.retrieve(query, k=rerank_top_n)

        # Stage 2: Rerank
        pairs = [(query, doc['content']) for doc in candidates]
        scores = self.reranker.predict(pairs)

        # Sort by reranker scores
        ranked_indices = np.argsort(scores)[::-1]
        reranked_docs = [candidates[i] for i in ranked_indices]

        # Return top-k
        return reranked_docs[:k]
```

#### LLM-BASED RERANKING
Use an LLM to judge relevance.

**Binary Relevance**
Ask LLM if document is relevant.
```python
def llm_rerank_binary(query, documents, llm):
    relevant_docs = []

    for doc in documents:
        prompt = f"""Is this document relevant to the query?

Query: {query}

Document: {doc}

Answer only 'yes' or 'no'."""

        response = llm.generate(prompt, max_tokens=5)

        if 'yes' in response.lower():
            relevant_docs.append(doc)

    return relevant_docs
```

**Scoring Relevance**
Get numerical relevance scores.
```python
def llm_rerank_score(query, documents, llm):
    scored_docs = []

    for doc in documents:
        prompt = f"""Rate the relevance of this document to the query on a scale of 1-10.

Query: {query}

Document: {doc}

Relevance score (1-10):"""

        score = int(llm.generate(prompt, max_tokens=5))
        scored_docs.append((doc, score))

    # Sort by score
    scored_docs.sort(key=lambda x: x[1], reverse=True)
    return [doc for doc, score in scored_docs]
```

**Comparative Ranking**
Compare documents pairwise or in batches.
```python
def llm_rerank_comparative(query, documents, llm):
    prompt = f"""Rank these documents by relevance to the query.

Query: {query}

Documents:
{format_documents(documents)}

Provide ranking (most to least relevant):"""

    ranking = llm.generate(prompt)
    ranked_docs = parse_ranking(ranking, documents)

    return ranked_docs
```
* **Pros:** Very accurate, Can handle nuanced relevance, Explains reasoning
* **Cons:** Expensive (LLM call per document or batch), Slow (hundreds of ms to seconds), May exceed context window with many docs
* **Use when:** Highest quality required, Cost/latency acceptable, Small candidate set (< 10 docs)

#### COHERE RERANK API
Managed reranking service.
```python
import cohere

co = cohere.Client(api_key="your-key")

def cohere_rerank(query, documents, top_n=5):
    response = co.rerank(
        query=query,
        documents=documents,
        top_n=top_n,
        model="rerank-english-v2.0"
    )

    return [doc.document for doc in response.results]
```
* **Models:** `rerank-english-v2.0` (English), `rerank-multilingual-v2.0` (100+ languages)
* **Pricing:** $1.00 per 1000 searches (as of 2025)
* **Pros:** Managed service, High quality, Multilingual
* **Cons:** API latency, Ongoing cost, Vendor dependency

#### FLASHRANK
Efficient local reranking.
```python
from flashrank import Ranker, RerankRequest

ranker = Ranker(model_name="ms-marco-MultiBERT-L-12")

def flashrank_rerank(query, documents, top_n=5):
    rerank_request = RerankRequest(
        query=query,
        passages=[{"text": doc} for doc in documents]
    )

    results = ranker.rerank(rerank_request)

    return [r.text for r in results[:top_n]]
```
* **Benefits:** Very fast (optimized inference), Self-hosted, No API costs

#### HYBRID RERANKING
Combine multiple signals.
```python
def hybrid_rerank(query, documents, weights=None):
    if weights is None:
        weights = {
            'vector_score': 0.3,
            'bm25_score': 0.2,
            'cross_encoder': 0.5
        }

    # Get scores from different models
    vector_scores = get_vector_scores(query, documents)
    bm25_scores = get_bm25_scores(query, documents)
    ce_scores = get_cross_encoder_scores(query, documents)

    # Normalize scores to [0, 1]
    vector_scores = normalize(vector_scores)
    bm25_scores = normalize(bm25_scores)
    ce_scores = normalize(ce_scores)

    # Weighted combination
    final_scores = (
        weights['vector_score'] * vector_scores +
        weights['bm25_score'] * bm25_scores +
        weights['cross_encoder'] * ce_scores
    )

    # Rank documents
    ranked_indices = np.argsort(final_scores)[::-1]
    return [documents[i] for i in ranked_indices]
```

---

### RERANKING STRATEGIES

#### TOP-K RERANKING
Rerank only top candidates from initial retrieval.
```python
# Retrieve top 20, rerank to get top 5
candidates = retriever.retrieve(query, k=20)
reranked = reranker.rerank(query, candidates, top_n=5)
```
* **Settings:** Retrieve: 3-5x the final k | Rerank: Final k needed
* **Example:** Need 5 final results -> Retrieve 20 candidates -> Rerank to top 5

#### CASCADING RERANKING
Multiple reranking stages with increasing accuracy.
```python
# Stage 1: Fast retrieval
candidates = fast_retriever.retrieve(query, k=100)

# Stage 2: Fast reranker
reranked_1 = tiny_reranker.rerank(query, candidates, top_n=20)

# Stage 3: Accurate reranker
reranked_2 = large_reranker.rerank(query, reranked_1, top_n=5)

return reranked_2
```
* **Use when:** Very large candidate sets, Multiple quality tiers needed, Optimizing cost/latency

#### QUERY-ADAPTIVE RERANKING
Different reranking based on query type.
```python
def adaptive_rerank(query, documents):
    query_type = classify_query(query)

    if query_type == "factual":
        # Use keyword signals
        return bm25_rerank(query, documents)

    elif query_type == "semantic":
        # Use cross-encoder
        return cross_encoder_rerank(query, documents)

    elif query_type == "complex":
        # Use LLM
        return llm_rerank(query, documents)
```

---

### PERFORMANCE OPTIMIZATION

#### BATCHING
Rerank multiple queries efficiently.
```python
# Bad: One at a time
for query in queries:
    rerank(query, docs)

# Good: Batched
pairs = [(q, doc) for q in queries for doc in docs]
scores = reranker.predict(pairs, batch_size=32)
```

#### CACHING
Cache reranking results.
```python
from functools import lru_cache
import hashlib

def cache_key(query, doc):
    return hashlib.md5(f"{query}:{doc}".encode()).hexdigest()

@lru_cache(maxsize=10000)
def cached_rerank_score(query, doc):
    return reranker.predict([(query, doc)])[0]
```

#### ASYNC RERANKING
Parallelize reranking calls.
```python
import asyncio

async def async_rerank_batch(query, documents):
    tasks = [
        rerank_async(query, doc)
        for doc in documents
    ]
    scores = await asyncio.gather(*tasks)
    return rank_by_scores(documents, scores)
```

---

### EVALUATION

#### METRICS
**Precision@k:** Relevant docs in top-k after reranking
```python
def precision_at_k(reranked_docs, relevant_docs, k):
    top_k = set(reranked_docs[:k])
    relevant = set(relevant_docs)
    return len(top_k & relevant) / k
```

**NDCG@k:** Normalized Discounted Cumulative Gain
```python
from sklearn.metrics import ndcg_score

def evaluate_reranking(predictions, relevance_labels, k=5):
    return ndcg_score([relevance_labels], [predictions], k=k)
```

**MRR:** Mean Reciprocal Rank
```python
def mrr(reranked_docs, relevant_docs):
    for i, doc in enumerate(reranked_docs, 1):
        if doc in relevant_docs:
            return 1 / i
    return 0
```

#### A/B TESTING
Compare reranking strategies.
```python
# Control: No reranking
control_results = retriever.retrieve(query, k=5)

# Treatment: With reranking
treatment_candidates = retriever.retrieve(query, k=20)
treatment_results = reranker.rerank(query, treatment_candidates, k=5)

# Measure: User satisfaction, answer quality
```

#### COST-BENEFIT ANALYSIS
| Reranker | Latency | Cost/1K | Quality | Best For |
|---|---|---|---|---|
| No reranking | 0ms | $0 | Baseline | Budget/speed critical |
| TinyBERT | +30ms | $0 (self-hosted) | +10% | Balanced |
| MiniLM | +50ms | $0 (self-hosted) | +20% | Quality-focused |
| Cohere | +100ms | $1 | +25% | Managed simplicity |
| LLM | +500ms | $5-20 | +30% | Highest quality |

---

### BEST PRACTICES
1. Always overfetch for reranking: Retrieve 3-5x the final k
2. Start with cross-encoder: MiniLM is a good default
3. Measure impact: A/B test reranking vs. no reranking
4. Tune retrieval count: Balance cost and recall
5. Consider query latency budget: Reranking adds 50-500ms
6. Monitor costs: LLM reranking can be expensive at scale

---

### CHOOSING A RERANKER
* **Prototyping:** `cross-encoder/ms-marco-MiniLM-L6-v2` (Easy to use, good quality)
* **Production (Cost-Sensitive):** `cross-encoder/ms-marco-TinyBERT-L2-v2` (Self-hosted, fast)
* **Production (Quality-Focused):** Cohere Rerank API (Highest quality, managed)
* **Multilingual:** `mmarco-mMiniLMv2-L12-H384-v1` or `cross-encoder/mmarco-mMiniLMv2-L12`
* **Highest Quality (Budget Available):** LLM-based reranking (GPT-4, Claude for best results)

> **Expert Tip from Ailog:** Reranking is high impact but not first priority. Get your chunking, embeddings, and retrieval right first - they're foundational. Once you have a working RAG system, reranking is the easiest way to gain another 10-25% accuracy. Start with Cohere Rerank API for zero-setup wins. We added reranking to production in one afternoon and immediately saw fewer hallucinations and better answer quality.

---

### TEST RERANKING ON AILOG
Compare reranking models with zero setup:
Ailog platform includes:
* Cohere Rerank, cross-encoders, LLM reranking
* Side-by-side quality comparison
* Latency and cost analysis
* A/B testing with real queries
*Test reranking free ->*

---

### FAQ

**WHAT IS THE DIFFERENCE BETWEEN RERANKING AND VECTOR SEARCH?**
Vector search (retrieval) compares embeddings to quickly find candidate documents among millions. Reranking comes after: it takes those candidates (typically 20-100) and re-scores them with a more accurate model (cross-encoder) that analyzes each query-document pair in depth. Vector search prioritizes speed, reranking prioritizes precision.

**DOES RERANKING SLOW DOWN RAG RESPONSES?**
Yes, reranking typically adds 50 to 500ms of latency depending on the model and number of documents. A lightweight cross-encoder like TinyBERT adds ~50ms for 20 documents. Cohere Rerank adds ~200ms (including API call). LLM-based reranking (GPT-4, Claude) can add 1-3 seconds. For most RAG applications, the quality gain far outweighs this delay.

**WHICH RERANKER SHOULD I USE FOR MULTILINGUAL CONTENT?**
For multilingual projects, use a model trained on multilingual data like `cross-encoder/mmarco-mMiniLMv2-L12` (open source) or the Cohere Rerank API which natively supports 100+ languages. Avoid models trained exclusively on English MS MARCO data: they lose quality on non-English content. The mMARCO models were specifically trained on multilingual data.

**IS RERANKING ESSENTIAL FOR A RAG SYSTEM?**
No, but it's highly recommended. A good RAG pipeline without reranking can work correctly if chunking and embeddings are well-calibrated. Reranking typically brings 10-25% additional precision and reduces hallucinations. It's the optimization with the best effort-to-impact ratio once the foundations are in place.

**HOW MUCH DOES RERANKING COST IN PRODUCTION?**
Self-hosted cross-encoders are free (excluding GPU compute). Cohere Rerank costs approximately $1 per 1,000 reranking requests. LLM-based reranking costs significantly more (input tokens x number of documents). For most projects, Cohere Rerank or a self-hosted cross-encoder offer the best value for money.

---

### NEXT STEPS
With retrieval and reranking optimized, it's critical to measure performance. The next guide covers evaluation metrics and methodologies for assessing RAG system quality.

**TAGS:** `reranking` `cross-encoder` `retrieval` `precision` `cohere` `colbert`
