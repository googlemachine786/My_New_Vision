# QUERY EXPANSION: RETRIEVE MORE RELEVANT RESULTS

**Date:** November 14, 2025 | **Read Time:** 10 min | **Author:** Ailog Research Team

> Improve recall by 40%: expand user queries with synonyms, sub-queries, and LLM-generated variations.

## WHY QUERY EXPANSION?
**Problem:** User query is too short or uses different words  
**Example:**
* User: `"ML models"`
* Relevant docs use: `"machine learning algorithms"`, `"neural networks"`, `"deep learning"`

Query expansion rewrites queries to match more documents.

---

## METHOD 1: SYNONYM EXPANSION
```python
from nltk.corpus import wordnet

def expand_with_synonyms(query):
    words = query.split()
    expanded_queries = [query]  # Include original

    for word in words:
        synonyms = []
        for syn in wordnet.synsets(word):
            for lemma in syn.lemmas():
                if lemma.name() != word:
                    synonyms.append(lemma.name().replace('_', ' '))

        # Add synonym variations
        if synonyms:
            expanded = query.replace(word, synonyms[0])
            expanded_queries.append(expanded)

    return list(set(expanded_queries))

# Example
queries = expand_with_synonyms("fast car")
# ["fast car", "quick car", "fast automobile", "quick automobile"]
```

---

## METHOD 2: LLM QUERY REWRITING
```python
import openai

def expand_with_llm(query):
    response = openai.ChatCompletion.create(
        model="gpt-4-turbo",
        messages=[{
            "role": "system",
            "content": "Generate 3 alternative phrasings of the user's query. Output as JSON array."
        }, {
            "role": "user",
            "content": query
        }],
        response_format={"type": "json_object"}
    )

    variations = json.loads(response.choices[0].message.content)
    return [query] + variations["alternatives"]

# Example
queries = expand_with_llm("How to reduce costs?")
# [
#   "How to reduce costs?",
#   "What are cost reduction strategies?",
#   "Ways to lower expenses",
#   "Best practices for cutting costs"
# ]
```

---

## METHOD 3: MULTI-QUERY RETRIEVAL
Search with all variations, merge results:
```python
def multi_query_retrieval(query, vector_db):
    # Generate variations
    queries = expand_with_llm(query)

    # Retrieve for each
    all_results = []
    for q in queries:
        q_emb = embed(q)
        results = vector_db.search(q_emb, limit=20)
        all_results.extend(results)

    # Deduplicate and rank by frequency
    doc_scores = {}
    for doc in all_results:
        if doc.id not in doc_scores:
            doc_scores[doc.id] = 0
        doc_scores[doc.id] += doc.score

    # Sort by combined score
    ranked = sorted(doc_scores.items(), key=lambda x: x[1], reverse=True)

    return ranked[:10]
```

---

## METHOD 4: HYDE (HYPOTHETICAL DOCUMENT EMBEDDINGS)
Generate fake answer, search for it:
```python
def hyde_retrieval(query):
    # Generate hypothetical answer
    hypothetical_doc = openai.ChatCompletion.create(
        model="gpt-4-turbo",
        messages=[{
            "role": "system",
            "content": "Write a detailed answer to this question as if it were a Wikipedia article."
        }, {
            "role": "user",
            "content": query
        }]
    ).choices[0].message.content

    # Embed hypothetical doc (not the query!)
    doc_embedding = embed(hypothetical_doc)

    # Search for similar documents
    results = vector_db.search(doc_embedding, limit=10)

    return results
```

---

## METHOD 5: STEP-BACK PROMPTING
Ask broader question first:
```python
def step_back_expansion(query):
    # Generate broader question
    step_back = openai.ChatCompletion.create(
        model="gpt-4-turbo",
        messages=[{
            "role": "system",
            "content": "Given a specific question, generate a broader, more general question."
        }, {
            "role": "user",
            "content": query
        }]
    ).choices[0].message.content

    return [query, step_back]

# Example
queries = step_back_expansion("What is the capital of France?")
# [
#   "What is the capital of France?",
#   "What are capitals of European countries?"
# ]
```

---

## METHOD 6: SUB-QUERY DECOMPOSITION
Break complex queries into parts:
```python
def decompose_query(query):
    response = openai.ChatCompletion.create(
        model="gpt-4-turbo",
        messages=[{
            "role": "system",
            "content": "Break this complex question into 2-3 simpler sub-questions. Return JSON array."
        }, {
            "role": "user",
            "content": query
        }],
        response_format={"type": "json_object"}
    )

    sub_queries = json.loads(response.choices[0].message.content)["sub_questions"]

    # Retrieve for each sub-query
    all_results = []
    for sq in sub_queries:
        results = vector_db.search(embed(sq), limit=5)
        all_results.extend(results)

    return deduplicate(all_results)

# Example
sub_queries = decompose_query("How does photosynthesis affect climate change?")
# [
#   "What is photosynthesis?",
#   "How do plants remove CO2?",
#   "What is the relationship between CO2 and climate?"
# ]
```

---

## LANGCHAIN IMPLEMENTATION
```python
from langchain.retrievers.multi_query import MultiQueryRetriever
from langchain.llms import OpenAI

retriever = MultiQueryRetriever.from_llm(
    retriever=vector_store.as_retriever(),
    llm=OpenAI(temperature=0)
)

# Automatically expands query
docs = retriever.get_relevant_documents("How to train neural networks?")
```

---

## EVALUATION
```python
# Measure recall improvement
def evaluate_expansion(queries, ground_truth_docs):
    recall_baseline = []
    recall_expanded = []

    for query, relevant_docs in zip(queries, ground_truth_docs):
        # Baseline
        base_results = vector_db.search(embed(query), limit=10)
        base_recall = len(set(base_results) & set(relevant_docs)) / len(relevant_docs)
        recall_baseline.append(base_recall)

        # Expanded
        expanded_results = multi_query_retrieval(query, vector_db)
        exp_recall = len(set(expanded_results) & set(relevant_docs)) / len(relevant_docs)
        recall_expanded.append(exp_recall)

    print(f"Baseline recall: {np.mean(recall_baseline):.2f}")
    print(f"Expanded recall: {np.mean(recall_expanded):.2f}")
```

**Recommendation:** Query expansion is low-cost, high-impact. Boost recall by 30-50% instantly.

---

### TAGS
`retrieval` `query expansion` `recall` `search`

### RELATED POSTS
* **RETRIEVAL FUNDAMENTALS: HOW RAG SEARCH WORKS** *(Intermediate | 18 min read)*  
  Master the basics of retrieval in RAG systems: embeddings, vector search, chunking, and indexing for relevant results.
* **ADVANCED RETRIEVAL STRATEGIES FOR RAG** *(Advanced | 13 min read)*  
  Beyond basic similarity search: hybrid search, query expansion, MMR, and multi-stage retrieval for better RAG performance.
* **HYBRID SEARCH FOR RAG: BM25 + VECTOR SEARCH TUTORIAL (2025)** *(Intermediate | 10 min read)*  
  Boost RAG retrieval accuracy by 20-30% with hybrid search. Step-by-step tutorial combining BM25 keyword matching with vector search using Weaviate, Qdrant, or Pinecone.
