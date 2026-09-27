# ADVANCED RETRIEVAL STRATEGIES FOR RAG

**February 5, 2025** | **13 min read** | **Ailog Research Team**

*Beyond basic similarity search: hybrid search, query expansion, MMR, and multi-stage retrieval for better RAG performance.*

---

### TL;DR
- **Hybrid search** (semantic + keyword) beats pure semantic by 20-35%
- **Query expansion** helps when queries are vague or use different terminology
- **MMR** reduces redundancy in retrieved results
- **Start simple:** Pure semantic -> Add hybrid -> Optimize with reranking
- **Test retrieval strategies side-by-side on Ailog**

---

### BEYOND SIMPLE SIMILARITY SEARCH
Basic RAG uses semantic similarity to retrieve documents. While effective, this approach has limitations:
- **Keyword blindness:** Misses exact term matches (product IDs, proper nouns)
- **Query-document mismatch:** Questions phrased differently than answers
- **Redundancy:** Retrieved chunks often contain similar information
- **Context insufficiency:** Top-k chunks may not provide complete context

Advanced retrieval strategies address these limitations.

---

### HYBRID SEARCH
Combines semantic (vector) and lexical (keyword) search.

#### BM25 + VECTOR SEARCH
BM25 (Best Matching 25): Statistical keyword ranking
```python
from rank_bm25 import BM25Okapi

# Index documents
tokenized_docs = [doc.split() for doc in documents]
bm25 = BM25Okapi(tokenized_docs)

# Keyword search
keyword_scores = bm25.get_scores(query.split())

# Vector search
vector_scores = cosine_similarity(query_embedding, doc_embeddings)

# Combine scores (weighted average)
alpha = 0.7  # Weight for vector search
final_scores = alpha * vector_scores + (1 - alpha) * keyword_scores

# Retrieve top-k
top_k_indices = np.argsort(final_scores)[-k:][::-1]
```

#### RECIPROCAL RANK FUSION (RRF)
Combine rankings from multiple retrievers.
```python
def reciprocal_rank_fusion(rankings_list, k=60):
    """
    rankings_list: List of ranked document IDs from different retrievers
    k: Constant (typically 60)
    """
    scores = {}

    for ranking in rankings_list:
        for rank, doc_id in enumerate(ranking, start=1):
            if doc_id not in scores:
                scores[doc_id] = 0
            scores[doc_id] += 1 / (k + rank)

    return sorted(scores.items(), key=lambda x: x[1], reverse=True)

# Example usage
vector_results = ["doc1", "doc3", "doc5", "doc2"]
bm25_results = ["doc2", "doc1", "doc4", "doc3"]

final_ranking = reciprocal_rank_fusion([vector_results, bm25_results])
# Result: [("doc1", score), ("doc2", score), ...]
```

#### WHEN TO USE HYBRID SEARCH
**Use hybrid when:**
- Queries contain specific terms (IDs, names, technical terms)
- Mix of semantic and exact matching needed
- Domain has specialized vocabulary

**Use vector-only when:**
- Natural language queries
- Synonym handling critical
- Multilingual search

**Benchmarks show:**
- Hybrid often outperforms either alone by 10-20%
- Especially effective for technical domains
- Critical for product search, code search

---

### QUERY EXPANSION
Reformulate or expand queries for better retrieval.

#### MULTI-QUERY GENERATION
Generate multiple query variations.
```python
def generate_query_variations(query, llm):
    prompt = f"""Given the user query, generate 3 variations that capture different aspects:

Original: {query}

Generate 3 variations:
1.
2.
3.
"""

    variations = llm.generate(prompt)
    all_queries = [query] + variations

    # Retrieve for each query
    all_results = []
    for q in all_queries:
        results = retrieve(q, k=5)
        all_results.extend(results)

    # Deduplicate and rerank
    unique_results = deduplicate(all_results)
    return rerank(unique_results, query)
```
**Benefits:** Captures multiple interpretations, Increases recall, Handles ambiguous queries
**Cost:** Multiple retrievals (slower, more expensive), LLM call for generation

#### HYDE (HYPOTHETICAL DOCUMENT EMBEDDINGS)
Generate a hypothetical answer, then search for it.
```python
def hyde_retrieval(query, llm, k=5):
    # Generate hypothetical answer
    prompt = f"""Write a passage that would answer this question:

Question: {query}

Passage:"""

    hypothetical_answer = llm.generate(prompt)

    # Embed and search using the hypothetical answer
    answer_embedding = embed(hypothetical_answer)
    results = vector_search(answer_embedding, k=k)

    return results
```
**Why it works:** Answers are semantically similar to answers (not questions), Bridges query-document gap, Effective when questions and answers are phrased differently
**When to use:** Question-answering systems, When queries are questions but documents are statements, Academic/research search

#### QUERY DECOMPOSITION
Break complex queries into sub-queries.
```python
def decompose_query(complex_query, llm):
    prompt = f"""Break this complex question into simpler sub-questions:

Question: {complex_query}

Sub-questions:
1.
2.
3.
"""

    sub_questions = llm.generate(prompt)

    # Retrieve for each sub-question
    all_contexts = []
    for sub_q in sub_questions:
        contexts = retrieve(sub_q, k=3)
        all_contexts.extend(contexts)

    # Generate final answer using all contexts
    final_answer = llm.generate(
        context=all_contexts,
        query=complex_query
    )

    return final_answer
```
**Use cases:** Multi-hop questions, Complex analytical queries, When single retrieval is insufficient

---

### MAXIMAL MARGINAL RELEVANCE (MMR)
Reduce redundancy in retrieved results.
```python
def mmr(query_embedding, doc_embeddings, documents, k=5, lambda_param=0.7):
    """
    Maximize relevance while minimizing similarity to already-selected docs.

    lambda_param: Tradeoff between relevance (1.0) and diversity (0.0)
    """
    selected = []
    remaining = list(range(len(documents)))

    while len(selected) < k and remaining:
        mmr_scores = []

        for i in remaining:
            # Relevance to query
            relevance = cosine_similarity(
                query_embedding,
                doc_embeddings[i]
            )

            # Max similarity to already selected docs
            if selected:
                similarities = [
                    cosine_similarity(doc_embeddings[i], doc_embeddings[j])
                    for j in selected
                ]
                max_sim = max(similarities)
            else:
                max_sim = 0

            # MMR score
            mmr_score = lambda_param * relevance - (1 - lambda_param) * max_sim
            mmr_scores.append((i, mmr_score))

        # Select best MMR score
        best = max(mmr_scores, key=lambda x: x[1])
        selected.append(best[0])
        remaining.remove(best[0])

    return [documents[i] for i in selected]
```
**Parameters:**
- `lambda_param = 1.0`: Pure relevance (no diversity)
- `lambda_param = 0.5`: Balance relevance and diversity
- `lambda_param = 0.0`: Maximum diversity

**Use when:** Retrieved chunks are very similar, Need diverse perspectives, Summarization tasks

---

### PARENT-CHILD RETRIEVAL
Retrieve small chunks, return larger context.
```python
class ParentChildRetriever:
    def __init__(self, documents):
        self.parents = []  # Original documents
        self.children = []  # Small chunks
        self.child_to_parent = {}  # Mapping

        for doc_id, doc in enumerate(documents):
            # Split into small chunks for precise retrieval
            chunks = split_document(doc, chunk_size=256)

            for chunk_id, chunk in enumerate(chunks):
                self.children.append(chunk)
                self.child_to_parent[len(self.children) - 1] = doc_id

            self.parents.append(doc)

        # Embed children for retrieval
        self.child_embeddings = embed_batch(self.children)

    def retrieve(self, query, k=3):
        # Search small chunks for precision
        query_emb = embed(query)
        child_indices = vector_search(query_emb, self.child_embeddings, k=k)

        # Return parent documents for context
        parent_indices = [self.child_to_parent[i] for i in child_indices]
        unique_parents = list(set(parent_indices))

        return [self.parents[i] for i in unique_parents]
```
**Benefits:** Precise retrieval (small chunks), Rich context (large documents), Best of both worlds
**Use when:** Need full context for generation, Documents have natural hierarchy (sections, paragraphs), Context window allows larger chunks

---

### ENSEMBLE RETRIEVAL
Combine multiple retrieval methods.
```python
class EnsembleRetriever:
    def __init__(self, retrievers, weights=None):
        self.retrievers = retrievers
        self.weights = weights or [1.0] * len(retrievers)

    def retrieve(self, query, k=5):
        all_results = []

        # Get results from each retriever
        for retriever, weight in zip(self.retrievers, self.weights):
            results = retriever.retrieve(query, k=k*2)  # Overfetch

            # Weight scores
            for doc, score in results:
                all_results.append((doc, score * weight))

        # Deduplicate and aggregate scores
        doc_scores = {}
        for doc, score in all_results:
            doc_id = doc['id']
            if doc_id not in doc_scores:
                doc_scores[doc_id] = {'doc': doc, 'score': 0}
            doc_scores[doc_id]['score'] += score

        # Sort and return top-k
        ranked = sorted(
            doc_scores.values(),
            key=lambda x: x['score'],
            reverse=True
        )

        return [item['doc'] for item in ranked[:k]]
```
**Example ensemble:**
```python
ensemble = EnsembleRetriever(
    retrievers=[
        VectorRetriever(embedding_model="openai"),
        BM25Retriever(),
        VectorRetriever(embedding_model="sentence-transformers")
    ],
    weights=[0.5, 0.3, 0.2]
)
```

---

### SELF-QUERY RETRIEVAL
Extract filters from natural language queries.
```python
def self_query_retrieval(query, llm, vector_db):
    # Extract structured query
    prompt = f"""Extract search filters from this query:

Query: {query}

Extract:
- search_text: Semantic search text
- filters: Metadata filters (dict)

Output (JSON):"""

    structured = llm.generate(prompt, format="json")

    # Example output:
    # {
    #   "search_text": "customer support best practices",
    #   "filters": {"department": "support", "date_range": "2024"}
    # }

    # Execute filtered search
    results = vector_db.query(
        text=structured['search_text'],
        filter=structured['filters'],
        k=5
    )

    return results
```
**Benefits:** Leverages metadata effectively, Natural language interface to filters, Better precision
**Use when:** Rich metadata available, Queries contain filterable attributes, Time-based, category-based, or attribute-based filtering needed

---

### MULTI-STAGE RETRIEVAL
Coarse-to-fine retrieval pipeline.
```python
class MultiStageRetriever:
    def __init__(self, fast_retriever, accurate_reranker):
        self.retriever = fast_retriever
        self.reranker = accurate_reranker

    def retrieve(self, query, k=5):
        # Stage 1: Fast retrieval (overfetch)
        candidates = self.retriever.retrieve(query, k=k*10)

        # Stage 2: Accurate reranking
        reranked = self.reranker.rerank(query, candidates)

        # Return top-k
        return reranked[:k]
```
**Stages:**
1. Retrieval (fast, high recall): 100 candidates
2. Reranking (accurate, expensive): Top 10
3. Optional: LLM-based refinement: Top 3

**Benefits:** Balance speed and accuracy, Cost-effective (expensive models on small candidate set), Higher quality results

---

### CONTEXTUAL COMPRESSION
Remove irrelevant parts from retrieved chunks.
```python
def compress_context(query, chunks, llm):
    compressed = []

    for chunk in chunks:
        prompt = f"""Extract only the parts relevant to the query:

Query: {query}

Document: {chunk}

Relevant extract:"""

        relevant_part = llm.generate(prompt, max_tokens=200)
        compressed.append(relevant_part)

    return compressed
```
**Benefits:** Reduce token usage, Fit more chunks in context window, Focus on relevant information
**Costs:** LLM calls (expensive), Additional latency
**Use when:** Token budget is tight, Retrieved chunks are long and partially relevant, Need to fit many sources

---

### CHOOSING A RETRIEVAL STRATEGY

#### DECISION FRAMEWORK
- **Start with:** Basic semantic search (vector similarity), k=3 to 5 chunks
- **Add hybrid search if:** Queries contain specific terms, Domain has specialized vocabulary, Performance improves in evaluation
- **Add query expansion if:** Queries are ambiguous, Recall is more important than precision, Willing to accept higher latency/cost
- **Add MMR if:** Retrieved chunks are redundant, Need diverse perspectives, Summarization or analysis tasks
- **Add reranking if:** Top-k results are not consistently relevant, Willing to trade latency for quality, Budget allows

#### PERFORMANCE IMPACT
| Strategy | Latency Impact | Cost Impact | Quality Gain |
|:---|:---|:---|:---|
| Hybrid search | +20-50ms | Low | +10-20% |
| Multi-query | +3x | High | +15-25% |
| HyDE | +LLM call | High | +10-30% |
| MMR | +10-50ms | Low | +5-15% |
| Parent-child | +0-20ms | Medium | +10-20% |
| Reranking | +50-200ms | Medium | +20-40% |

---

### PRACTICAL IMPLEMENTATION

#### LANGCHAIN EXAMPLE
```python
from langchain.retrievers import (
    EnsembleRetriever,
    ContextualCompressionRetriever
)
from langchain.retrievers.document_compressors import LLMChainExtractor

# Ensemble: Vector + BM25
vector_retriever = vector_db.as_retriever(search_kwargs={"k": 5})
bm25_retriever = BM25Retriever.from_documents(documents)

ensemble = EnsembleRetriever(
    retrievers=[vector_retriever, bm25_retriever],
    weights=[0.7, 0.3]
)

# Add compression
compressor = LLMChainExtractor.from_llm(llm)
compressed_retriever = ContextualCompressionRetriever(
    base_compressor=compressor,
    base_retriever=ensemble
)
```

#### LLAMAINDEX EXAMPLE
```python
from llama_index import VectorStoreIndex, SimpleKeywordTableIndex
from llama_index.retrievers import RouterRetriever

# Create retrievers
vector_retriever = VectorStoreIndex.from_documents(documents).as_retriever()
keyword_retriever = SimpleKeywordTableIndex.from_documents(documents).as_retriever()

# Router retriever (chooses based on query)
router = RouterRetriever(
    selector=llm_selector,
    retriever_dict={
        "vector": vector_retriever,
        "keyword": keyword_retriever
    }
)

# Query-dependent routing
results = router.retrieve("What are the system requirements?")
```

---

### MONITORING RETRIEVAL QUALITY
Track these metrics:

**Retrieval Metrics:**
- Precision@k: Relevant docs in top-k
- Recall@k: Retrieved relevant docs / all relevant docs
- MRR: Mean Reciprocal Rank of first relevant result

**User Metrics:**
- Answer quality ratings
- Follow-up question rate
- Task completion

**Technical Metrics:**
- Retrieval latency
- Cache hit rate
- Query throughput

> **Expert Tip from Ailog:** Hybrid search is the single highest-ROI retrieval improvement. Adding BM25 keyword search alongside semantic search consistently delivers 20-35% better results across domains. It's easier to implement than query expansion and more reliable than MMR tuning. If you only make one retrieval optimization, make it hybrid search. We implemented it in production and saw immediate quality gains with minimal complexity.

---

### TEST RETRIEVAL STRATEGIES ON AILOG
Compare retrieval methods with your documents:
Ailog lets you benchmark:
- Pure semantic vs hybrid search
- Different query expansion techniques
- MMR vs standard retrieval
- Custom metadata filtering

See real metrics: Precision@k, MRR, latency comparisons
*Start testing -> Free access to all retrieval strategies.*

---

### NEXT STEPS
Retrieved documents often need reranking to optimize for the specific query. The next guide covers reranking strategies and cross-encoder models to further improve RAG quality.

**TAGS:** `retrieval` `hybrid search` `query expansion` `MMR`

**RELATED POSTS:**
- **RETRIEVAL FUNDAMENTALS: HOW RAG SEARCH WORKS** (Intermediate) - Master the basics of retrieval in RAG systems: embeddings, vector search, chunking, and indexing for relevant results. (18 min read)
- **HYBRID FUSION: COMBINING DENSE AND SPARSE RETRIEVAL** - Master hybrid fusion to combine semantic and lexical search. RRF, weighted fusion, and optimal combination strategies explained.
- **HYBRID SEARCH FOR RAG: BM25 + VECTOR SEARCH TUTORIAL (2025)** (Intermediate) - Boost RAG retrieval accuracy by 20-30% with hybrid search. Step-by-step tutorial combining BM25 keyword matching with vector search using Weaviate, Qdrant, or Pinecone. (10 min read)
