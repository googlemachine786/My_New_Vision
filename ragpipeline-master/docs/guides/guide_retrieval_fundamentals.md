# RETRIEVAL FUNDAMENTALS: HOW RAG SEARCH WORKS
**January 15, 2026** | **18 min read** | **Ailog Team**

*Master the basics of retrieval in RAG systems: embeddings, vector search, chunking, and indexing for relevant results.*

Retrieval is the beating heart of any RAG (Retrieval-Augmented Generation) system. Without effective search, even the best LLM in the world will produce off-topic or incomplete answers. This guide walks you through a deep understanding of retrieval mechanisms, from theory to practical implementation.

## WHY RETRIEVAL IS CRITICAL IN A RAG SYSTEM
A RAG system works in two stages: first retrieving relevant documents (retrieval), then generating a response based on those documents (generation). The quality of the final response directly depends on the quality of retrieved documents.

Imagine an assistant that needs to answer "What is your return policy?" If retrieval brings back pages about shipping conditions instead of the return policy, the LLM will generate an incorrect answer or invent a fictional policy.

## THE THREE PILLARS OF RETRIEVAL
1. **Representation:** How to transform text into mathematical vectors
2. **Indexing:** How to organize these vectors for fast search
3. **Search:** How to find the most relevant documents

## UNDERSTANDING EMBEDDINGS
Embeddings are vector representations of text. Each word, sentence, or document is transformed into a vector of numbers (typically 384 to 1536 dimensions) that captures its semantic meaning.

### HOW EMBEDDINGS WORK
```python
from sentence_transformers import SentenceTransformer

# Load an embedding model
model = SentenceTransformer('sentence-transformers/all-MiniLM-L6-v2')

# Create embeddings
texts = [
    "How do I return a product?",
    "What is the refund policy?",
    "Store opening hours"
]

embeddings = model.encode(texts)

# Calculate similarity
from sklearn.metrics.pairwise import cosine_similarity
similarities = cosine_similarity(embeddings)

print("Similarity 'return' vs 'refund':", similarities[0][1])  # ~0.85
print("Similarity 'return' vs 'hours':", similarities[0][2])   # ~0.25
```
The first two sentences, although worded differently, have high similarity because they deal with the same topic. The third is semantically distant.

### CHOOSING YOUR EMBEDDING MODEL
| Model | Dimensions | Performance | Speed | Recommended Use |
| :--- | :--- | :--- | :--- | :--- |
| `all-MiniLM-L6-v2` | 384 | Good | Fast | Prototyping, high volumes |
| `all-mpnet-base-v2` | 768 | Very good | Medium | General production |
| `text-embedding-3-small` | 1536 | Excellent | Fast (API) | Production with API budget |
| `text-embedding-3-large` | 3072 | State of the art | Medium (API) | Critical high-precision cases |
| `multilingual-e5-large` | 1024 | Excellent multilingual | Medium | FR/EN/multilingual content |

For multilingual projects, prioritize models trained on diverse corpora:
```python
# Excellent choice for multilingual
model = SentenceTransformer('intfloat/multilingual-e5-large')

# Prefix required for E5
query = "query: How does the warranty work?"
documents = ["passage: The warranty covers manufacturing defects for 2 years..."]
```

## CHUNKING: INTELLIGENTLY SPLITTING DOCUMENTS
Chunking is the art of splitting documents into appropriately sized pieces. Too large, and the chunk contains noise. Too small, and it loses context.

### CHUNKING STRATEGIES

#### 1. FIXED-SIZE CHUNKING
The simplest method: split every X characters with overlap.
```python
from langchain.text_splitter import RecursiveCharacterTextSplitter

splitter = RecursiveCharacterTextSplitter(
    chunk_size=500,        # Target size
    chunk_overlap=50,      # Overlap to preserve context
    separators=["\n\n", "\n", ". ", " ", ""]
)

document = """
Return Policy

You have 30 days to return an unused product in its original packaging.

Return Procedure:
1. Log in to your customer account
2. Select the relevant order
3. Click "Request a return"
4. Print the return label

Return shipping costs are your responsibility unless the product is defective.

Refund

Once the return is received and validated, the refund is processed within 5 business days to the payment method used for the purchase.
"""

chunks = splitter.split_text(document)
for i, chunk in enumerate(chunks):
    print(f"Chunk {i+1}: {chunk[:100]}...")
```

#### 2. SEMANTIC CHUNKING
More sophisticated: split at natural text boundaries (paragraphs, sections).
```python
from langchain.text_splitter import MarkdownTextSplitter

md_splitter = MarkdownTextSplitter(
    chunk_size=500,
    chunk_overlap=0
)

# Respects Markdown structure
chunks = md_splitter.split_text(markdown_document)
```

#### 3. SENTENCE CHUNKING WITH SLIDING WINDOW
Ideal for FAQs and short content:
```python
import nltk
nltk.download('punkt')

def chunk_by_sentences(text, sentences_per_chunk=3, overlap=1):
    sentences = nltk.sent_tokenize(text)
    chunks = []

    for i in range(0, len(sentences), sentences_per_chunk - overlap):
        chunk = " ".join(sentences[i:i + sentences_per_chunk])
        chunks.append(chunk)

    return chunks
```

### STRATEGY COMPARISON TABLE
| Strategy | Advantages | Disadvantages | Use Case |
| :--- | :--- | :--- | :--- |
| Fixed size | Simple, predictable | Cuts mid-idea | Homogeneous documents |
| Semantic | Preserves meaning | More complex | Structured documentation |
| By sentence | Fine precision | Sometimes too short chunks | FAQ, support |
| Hierarchical | Parent context preserved | Increased complexity | Technical documentation |

## INDEXING WITH VECTOR DATABASES
Once embeddings are created, they need to be stored and indexed for fast search. Vector databases are optimized for this task.

### QDRANT: IMPLEMENTATION EXAMPLE
```python
from qdrant_client import QdrantClient
from qdrant_client.models import Distance, VectorParams, PointStruct

# Connection
client = QdrantClient(host="localhost", port=6333)

# Create a collection
client.create_collection(
    collection_name="knowledge_base",
    vectors_config=VectorParams(
        size=384,  # Dimension of your embeddings
        distance=Distance.COSINE
    )
)

# Index documents
points = [
    PointStruct(
        id=i,
        vector=embedding.tolist(),
        payload={
            "text": chunk,
            "source": "return_policy.md",
            "category": "support"
        }
    )
    for i, (embedding, chunk) in enumerate(zip(embeddings, chunks))
]

client.upsert(
    collection_name="knowledge_base",
    points=points
)
```

### VECTOR SEARCH
```python
def search(query: str, top_k: int = 5):
    # Encode the query
    query_embedding = model.encode(query)

    # Search
    results = client.search(
        collection_name="knowledge_base",
        query_vector=query_embedding.tolist(),
        limit=top_k
    )

    return [
        {
            "text": hit.payload["text"],
            "score": hit.score,
            "source": hit.payload["source"]
        }
        for hit in results
    ]

# Example
results = search("How do I get a refund?")
for r in results:
    print(f"Score: {r['score']:.3f} - {r['text'][:100]}...")
```

### SIMILARITY METRICS
The choice of metric impacts search results.

#### COSINE SIMILARITY
The most widely used. Measures the angle between two vectors, regardless of their magnitude.
```python
import numpy as np

def cosine_similarity(a, b):
    return np.dot(a, b) / (np.linalg.norm(a) * np.linalg.norm(b))
```
*Advantages:* Insensitive to original text length  
*Disadvantages:* May miss magnitude nuances

#### DOT PRODUCT
Faster, but sensitive to vector magnitude.
```python
def dot_product(a, b):
    return np.dot(a, b)
```
*Advantages:* Faster to compute  
*Disadvantages:* Requires normalized vectors to be comparable to cosine

#### EUCLIDEAN DISTANCE
Measures the "as the crow flies" distance between two points.
```python
def euclidean_distance(a, b):
    return np.linalg.norm(a - b)
```
*Advantages:* Geometrically intuitive  
*Disadvantages:* Sensitive to outliers and dimensionality

## OPTIMIZING RETRIEVAL

### 1. QUERY EXPANSION
Enrich the user query to improve recall:
```python
def expand_query(query: str, llm) -> list[str]:
    prompt = f"""
    Generate 3 reformulations of this question to improve search:
    Original question: {query}

    Reformulations:
    """

    expansions = llm.generate(prompt)
    return [query] + expansions

# Search with all variants
def search_expanded(query: str, top_k: int = 5):
    queries = expand_query(query, llm)
    all_results = []

    for q in queries:
        results = search(q, top_k=top_k)
        all_results.extend(results)

    # Deduplicate and re-score
    return deduplicate_and_rerank(all_results)
```

### 2. RERANKING
Use a reranking model to refine results:
```python
from sentence_transformers import CrossEncoder

reranker = CrossEncoder('cross-encoder/ms-marco-MiniLM-L-6-v2')

def rerank(query: str, documents: list[str], top_k: int = 3):
    pairs = [[query, doc] for doc in documents]
    scores = reranker.predict(pairs)

    # Sort by descending score
    ranked = sorted(zip(documents, scores), key=lambda x: x[1], reverse=True)
    return ranked[:top_k]

# Complete pipeline
def search_with_rerank(query: str):
    # 1. Initial search (high recall)
    initial_results = search(query, top_k=20)

    # 2. Reranking (high precision)
    documents = [r["text"] for r in initial_results]
    reranked = rerank(query, documents, top_k=5)

    return reranked
```

### 3. METADATA FILTERING
Combine vector search with classic filters:
```python
from qdrant_client.models import Filter, FieldCondition, MatchValue

def search_filtered(query: str, category: str = None, top_k: int = 5):
    query_embedding = model.encode(query)

    # Build filter
    filter_conditions = None
    if category:
        filter_conditions = Filter(
            must=[
                FieldCondition(
                    key="category",
                    match=MatchValue(value=category)
                )
            ]
        )

    results = client.search(
        collection_name="knowledge_base",
        query_vector=query_embedding.tolist(),
        query_filter=filter_conditions,
        limit=top_k
    )

    return results

# Search only in "support" category
results = search_filtered("return policy", category="support")
```

## EVALUATING RETRIEVAL QUALITY
To measure your retrieval system's effectiveness, use these metrics:

### RECALL@K
Proportion of relevant documents found among the top k results.
```python
def recall_at_k(retrieved: list, relevant: list, k: int) -> float:
    retrieved_k = set(retrieved[:k])
    relevant_set = set(relevant)

    return len(retrieved_k & relevant_set) / len(relevant_set)
```

### MRR (MEAN RECIPROCAL RANK)
Average position of the first relevant document.
```python
def mrr(queries_results: list[tuple[list, list]]) -> float:
    reciprocal_ranks = []

    for retrieved, relevant in queries_results:
        for i, doc in enumerate(retrieved):
            if doc in relevant:
                reciprocal_ranks.append(1 / (i + 1))
                break
        else:
            reciprocal_ranks.append(0)

    return sum(reciprocal_ranks) / len(reciprocal_ranks)
```

### NDCG (NORMALIZED DISCOUNTED CUMULATIVE GAIN)
Takes into account result order and relevance scores.
```python
import numpy as np

def ndcg_at_k(relevances: list[float], k: int) -> float:
    relevances = np.array(relevances[:k])

    # DCG
    discounts = np.log2(np.arange(2, len(relevances) + 2))
    dcg = np.sum(relevances / discounts)

    # IDCG (ideal DCG)
    ideal_relevances = np.sort(relevances)[::-1]
    idcg = np.sum(ideal_relevances / discounts)

    return dcg / idcg if idcg > 0 else 0
```

## COMMON PITFALLS AND SOLUTIONS

1. **CHUNKS TOO LARGE**  
   *Symptom:* Retrieval returns vaguely relevant but imprecise documents.  
   *Solution:* Reduce chunk size or use hierarchical chunking.

2. **DOMAIN VOCABULARY**  
   *Symptom:* Business terms are not well understood by embeddings.  
   *Solution:* Fine-tune the embedding model or use a synonym vocabulary.
   ```python
   synonyms = {
       "ticket": ["request", "inquiry", "incident"],
       "KB": ["knowledge base", "documentation"],
   }

   def expand_with_synonyms(query: str) -> str:
       for term, syns in synonyms.items():
           if term.lower() in query.lower():
               query += " " + " ".join(syns)
       return query
   ```

3. **AMBIGUOUS QUERIES**  
   *Symptom:* "Problem with my order" returns too many different results.  
   *Solution:* Use conversational context or ask for clarification.

4. **COLD START**  
   *Symptom:* Little data at startup, irrelevant retrieval.  
   *Solution:* Enrich with synthetic data or generated FAQs.

## PRODUCTION ARCHITECTURE
For a production retrieval system, here's a recommended architecture:
```
+-------------------------------------------------------------+
|                        API Gateway                           |
+-----------------------------+-------------------------------+
                              |
+-----------------------------v-------------------------------+
|                   Query Processor                            |
|  - Normalization                                             |
|  - Language detection                                        |
|  - Query expansion                                           |
+-----------------------------+-------------------------------+
                              |
             +----------------+----------------+
             v                                 v
+-------------------------+      +-------------------------+
|  Dense Search           |      |  Sparse Search          |
|   (Qdrant)              |      |   (BM25)                |
+-----------+-------------+      +-----------+-------------+
            |                                |
            +----------------+---------------+
                             v
              +-------------------------+
              |  Fusion/Rerank          |
              +-----------+-------------+
                          v
              +-------------------------+
              |   LLM Context           |
              +-------------------------+
```

## NEXT STEPS
Now that you've mastered retrieval fundamentals, dive deeper with our specialized guides:
* **Dense Retrieval: Semantic Search with Embeddings** - Dive into advanced embeddings
* **Sparse Retrieval and BM25** - Discover when lexical search excels
* **Hybrid Fusion** - Combine the best of both worlds

For a comprehensive RAG overview, check our *Complete Introduction to RAG*.

---

## PUT IT INTO PRACTICE WITH AILOG
Implementing a performant retrieval system takes time and expertise. With Ailog, get a turnkey RAG infrastructure:
* Intelligent chunking optimized for your content type
* Multilingual embedding models (native French/English)
* Automatic reranking for ultra-precise results
* Sovereign hosting in France, GDPR compliant

*Try Ailog for free and deploy your first RAG assistant in 3 minutes.*

## TAGS
`RAG` `retrieval` `embeddings` `vector search` `chunking`
