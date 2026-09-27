# SEMANTIC CHUNKING FOR BETTER RETRIEVAL
**Date:** November 8, 2025 | **Read Time:** 12 min | **Author:** Ailog Research Team  
*Split documents intelligently based on meaning, not just length. Learn semantic chunking techniques for RAG.*

## THE PROBLEM WITH FIXED-SIZE CHUNKING
Traditional chunking splits text every N characters or tokens:
* Breaks sentences mid-thought
* Separates related content
* No context awareness

Semantic chunking splits based on meaning, not length.

## HOW SEMANTIC CHUNKING WORKS
1. Embed each sentence using a sentence encoder
2. Calculate similarity between consecutive sentences
3. Split where similarity drops (topic change)

```python
from sentence_transformers import SentenceTransformer
import numpy as np

model = SentenceTransformer('all-MiniLM-L6-v2')

def semantic_chunk(text, similarity_threshold=0.5):
    # Split into sentences
    sentences = text.split('. ')

    # Embed all sentences
    embeddings = model.encode(sentences)

    # Calculate cosine similarity between consecutive sentences
    chunks = []
    current_chunk = [sentences[0]]

    for i in range(1, len(sentences)):
        similarity = np.dot(embeddings[i-1], embeddings[i]) / (
            np.linalg.norm(embeddings[i-1]) * np.linalg.norm(embeddings[i])
        )

        if similarity < similarity_threshold:
            # Topic changed - start new chunk
            chunks.append('. '.join(current_chunk))
            current_chunk = [sentences[i]]
        else:
            current_chunk.append(sentences[i])

    # Add final chunk
    chunks.append('. '.join(current_chunk))

    return chunks
```

## LANGCHAIN SEMANTIC CHUNKING (2025)
LangChain now includes built-in semantic chunking:

```python
from langchain.text_splitter import SemanticChunker
from langchain_openai.embeddings import OpenAIEmbeddings

text_splitter = SemanticChunker(
    OpenAIEmbeddings(),
    breakpoint_threshold_type="percentile",  # or "standard_deviation"
    breakpoint_threshold_amount=95
)

chunks = text_splitter.create_documents([long_text])
```

## ADVANCED: MULTI-LEVEL SEMANTIC CHUNKING
Combine semantic splits with size constraints:

```python
def smart_semantic_chunk(text, max_chunk_size=1000, min_chunk_size=200):
    # First: semantic split
    semantic_chunks = semantic_chunk(text)

    final_chunks = []

    for chunk in semantic_chunks:
        # If chunk too large, further split
        if len(chunk) > max_chunk_size:
            # Split by paragraphs within this semantic section
            paragraphs = chunk.split('\n\n')
            sub_chunk = ""

            for para in paragraphs:
                if len(sub_chunk) + len(para) < max_chunk_size:
                    sub_chunk += para + "\n\n"
                else:
                    final_chunks.append(sub_chunk.strip())
                    sub_chunk = para + "\n\n"

            if sub_chunk:
                final_chunks.append(sub_chunk.strip())

        # If chunk too small, merge with next
        elif len(chunk) < min_chunk_size and final_chunks:
            final_chunks[-1] += "\n\n" + chunk
        else:
            final_chunks.append(chunk)

    return final_chunks
```

## LLAMAINDEX SEMANTIC SPLITTER

```python
from llama_index.node_parser import SemanticSplitterNodeParser
from llama_index.embeddings import OpenAIEmbedding

embed_model = OpenAIEmbedding()

splitter = SemanticSplitterNodeParser(
    buffer_size=1,  # Sentences to group
    breakpoint_percentile_threshold=95,
    embed_model=embed_model
)

nodes = splitter.get_nodes_from_documents(documents)
```

## WHEN TO USE SEMANTIC CHUNKING
**Use semantic chunking when:**
* Documents have clear topic transitions
* You need high precision retrieval
* Content is narrative or explanatory
* You can afford the compute cost

**Stick to fixed-size when:**
* Speed is critical
* Documents are very uniform
* Budget is limited
* Content is tabular or structured

## PERFORMANCE CONSIDERATIONS
**Embedding cost:**
* Semantic chunking requires embedding every sentence
* For a 10,000-word document: ~300 sentences to embed
* Consider caching embeddings

**Speed comparison (November 2025):**
* Fixed-size: ~1ms per document
* Semantic: ~100-500ms per document (depending on model)

## HYBRID APPROACH: BEST OF BOTH WORLDS

```python
def hybrid_chunk(text, target_size=500):
    # 1. Semantic split first
    semantic_chunks = semantic_chunk(text, similarity_threshold=0.6)

    # 2. Merge small chunks, split large ones
    final_chunks = []
    buffer = ""

    for chunk in semantic_chunks:
        if len(buffer) + len(chunk) < target_size * 1.5:
            buffer += "\n\n" + chunk if buffer else chunk
        else:
            if buffer:
                final_chunks.append(buffer)
            buffer = chunk

    if buffer:
        final_chunks.append(buffer)

    return final_chunks
```

## EVALUATION
Test retrieval quality with semantic vs fixed chunking:

```python
# Your test queries
queries = [
    "How does photosynthesis work?",
    "What are the benefits of exercise?"
]

# Compare retrieval accuracy
semantic_results = evaluate_chunking(semantic_chunks, queries)
fixed_results = evaluate_chunking(fixed_chunks, queries)

print(f"Semantic MRR: {semantic_results['mrr']}")
print(f"Fixed MRR: {fixed_results['mrr']}")
```

> Semantic chunking typically improves retrieval by 15-30% but costs 100x more compute. Choose based on your accuracy/cost tradeoff.

---
### TAGS
`chunking` `semantic` `nlp` `embeddings`

### RELATED POSTS
* **RAG CHUNKING STRATEGIES 2025: OPTIMAL CHUNK SIZES & TECHNIQUES** *(Guides/Intermediate | 15 min read)*  
  Master document chunking for RAG: optimal chunk sizes (512-1024 tokens), overlap strategies, semantic vs fixed-size splitting. Improve retrieval by 25%+.
* **HIERARCHICAL CHUNKING: PRESERVING DOCUMENT STRUCTURE** *(Guides/Advanced | 11 min read)*  
  Hierarchical chunking maintains parent-child relationships in your documents. Learn how to implement this advanced technique to improve RAG retrieval quality.
* **FIXED-SIZE CHUNKING: FAST AND RELIABLE** *(Guides/Beginner | 7 min read)*  
  Master the basics: implement fixed-size chunking with overlaps for consistent, predictable RAG performance.
