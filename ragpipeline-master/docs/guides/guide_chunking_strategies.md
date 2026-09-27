# RAG CHUNKING STRATEGIES 2025: OPTIMAL CHUNK SIZES & TECHNIQUES

**January 25, 2025** | 15 min read | Ailog Research Team

Master document chunking for RAG: optimal chunk sizes (512-1024 tokens), overlap strategies, semantic vs fixed-size splitting. Improve retrieval by 25%+.

## TL;DR
* Chunk size matters: 500-1000 tokens balances context and precision
* Semantic chunking (splitting by meaning) beats fixed-size for quality (+15-30% retrieval accuracy)
* Overlap (10-20%) prevents losing context at boundaries
* Best for most use cases: Recursive text splitter with 512 tokens, 50 token overlap
* Try it now: Test different strategies on Ailog's platform

## THE CHUNKING PROBLEM
Most documents are too long to:
* Embed as a single vector (context window limits)
* Use entirely as LLM context (token limits)
* Retrieve with precision (too much irrelevant information)

Chunking splits documents into smaller, manageable pieces while preserving semantic meaning.

## WHY CHUNKING MATTERS
Poor chunking leads to:
* Split context: Important information broken across chunks
* Irrelevant retrieval: Chunks contain mix of relevant and irrelevant content
* Lost context: Chunk boundaries cut off critical information
* Poor generation: LLM lacks complete context to answer accurately

Good chunking enables:
* Precise retrieval: Find exactly the relevant information
* Complete context: Chunks contain full thoughts or concepts
* Efficient token usage: No wasted context on irrelevant text
* Better answers: LLM has what it needs, nothing more

## FIXED-SIZE CHUNKING

### CHARACTER-BASED
Split text every N characters.

```python
def chunk_by_chars(text, chunk_size=1000, overlap=200):
    chunks = []
    start = 0
    while start < len(text):
        end = start + chunk_size
        chunks.append(text[start:end])
        start += chunk_size - overlap
    return chunks
```
**Pros:**
* Simple implementation
* Predictable chunk sizes
* Fast processing

**Cons:**
* Splits mid-word, mid-sentence
* Ignores semantic boundaries
* Breaks code, tables, lists

**Use when:**
* Quick prototype needed
* Text structure is homogeneous
* Precision not critical

### TOKEN-BASED
Split by token count (matches model tokenization).

```python
import tiktoken

def chunk_by_tokens(text, chunk_size=512, overlap=50):
    encoding = tiktoken.get_encoding("cl100k_base")
    tokens = encoding.encode(text)

    chunks = []
    start = 0
    while start < len(tokens):
        end = start + chunk_size
        chunk_tokens = tokens[start:end]
        chunks.append(encoding.decode(chunk_tokens))
        start += chunk_size - overlap

    return chunks
```
**Pros:**
* Respects token limits precisely
* Works with any embedding model
* Predictable embedding costs

**Cons:**
* Still ignores semantic boundaries
* Tokenization overhead
* May split important context

**Use when:**
* Strict token budget
* Token count is critical (API costs)
* Embedding model has hard token limits

### RECOMMENDED FIXED SIZES
| Use Case | Chunk Size | Overlap | Rationale |
|---|---|---|---|
| Short FAQ | 128-256 tokens | 0-20 | Minimal context needed |
| General docs | 512-1024 tokens | 50-100 | Balance precision and context |
| Technical docs | 1024-2048 tokens | 100-200 | More context for complex topics |
| Code | 256-512 tokens | 50-100 | Preserve function/class context |

## SEMANTIC CHUNKING
Split at natural semantic boundaries.

### SENTENCE-BASED
Split at sentence boundaries.

```python
import nltk
nltk.download('punkt')

def chunk_by_sentences(text, sentences_per_chunk=5):
    sentences = nltk.sent_tokenize(text)
    chunks = []

    for i in range(0, len(sentences), sentences_per_chunk):
        chunk = ' '.join(sentences[i:i + sentences_per_chunk])
        chunks.append(chunk)

    return chunks
```
**Pros:**
* Respects sentence boundaries
* More readable chunks
* Preserves complete thoughts

**Cons:**
* Variable chunk sizes
* Sentence detection can fail
* May not group related sentences

**Use when:**
* Readability is important
* Sentences are self-contained
* General narrative text

### PARAGRAPH-BASED
Split at paragraph breaks.

```python
def chunk_by_paragraphs(text, paragraphs_per_chunk=2):
    paragraphs = text.split('\n\n')
    chunks = []

    for i in range(0, len(paragraphs), paragraphs_per_chunk):
        chunk = '\n\n'.join(paragraphs[i:i + paragraphs_per_chunk])
        chunks.append(chunk)

    return chunks
```
**Pros:**
* Respects document structure
* Keeps related content together
* Natural reading units

**Cons:**
* Highly variable sizes
* Depends on formatting
* Long paragraphs still problematic

**Use when:**
* Well-formatted documents
* Paragraphs represent complete ideas
* Blog posts, articles

### RECURSIVE CHARACTER SPLITTING
LangChain's approach: try splits in order of preference.

```python
from langchain.text_splitter import RecursiveCharacterTextSplitter

splitter = RecursiveCharacterTextSplitter(
    chunk_size=1000,
    chunk_overlap=200,
    separators=["\n\n", "\n", ". ", " ", ""]
)

chunks = splitter.split_text(text)
```
**Split hierarchy:**
1. Double newline (paragraphs)
2. Single newline (lines)
3. Period + space (sentences)
4. Space (words)
5. Character

**Pros:**
* Respects document structure when possible
* Falls back gracefully
* Balances semantics and size

**Cons:**
* Still somewhat arbitrary
* May not capture true semantic units
* Configuration required

**Use when:**
* General-purpose chunking
* Mixed document types
* Good default choice

## METADATA-AWARE CHUNKING
Use document structure to inform chunking.

### MARKDOWN CHUNKING
Split by headers, preserving hierarchy.

```python
def chunk_markdown(text):
    chunks = []
    current_h1 = ""
    current_h2 = ""
    current_chunk = []

    for line in text.split('\n'):
        if line.startswith('# '):
            if current_chunk:
                chunks.append({
                    'content': '\n'.join(current_chunk),
                    'h1': current_h1,
                    'h2': current_h2
                })
                current_chunk = []
            current_h1 = line[2:]

        elif line.startswith('## '):
            if current_chunk:
                chunks.append({
                    'content': '\n'.join(current_chunk),
                    'h1': current_h1,
                    'h2': current_h2
                })
                current_chunk = []
            current_h2 = line[3:]

        current_chunk.append(line)

    if current_chunk:
        chunks.append({
            'content': '\n'.join(current_chunk),
            'h1': current_h1,
            'h2': current_h2
        })

    return chunks
```
**Metadata benefits:**
* Headers provide context for search
* Can filter by section
* Better relevance scoring

### HTML/XML CHUNKING
Split by semantic HTML tags.

```python
from bs4 import BeautifulSoup

def chunk_html(html):
    soup = BeautifulSoup(html, 'html.parser')
    chunks = []

    # Split by sections
    for section in soup.find_all(['section', 'article', 'div']):
        if section.get('class') in ['content', 'main']:
            chunks.append({
                'content': section.get_text(),
                'tag': section.name,
                'class': section.get('class')
            })

    return chunks
```

### CODE CHUNKING
Split by function/class boundaries.

```python
import ast

def chunk_python_code(code):
    tree = ast.parse(code)
    chunks = []

    for node in ast.walk(tree):
        if isinstance(node, (ast.FunctionDef, ast.ClassDef)):
            chunk_lines = code.split('\n')[node.lineno-1:node.end_lineno]
            chunks.append({
                'content': '\n'.join(chunk_lines),
                'type': type(node).__name__,
                'name': node.name
            })

    return chunks
```
**Pros:**
* Preserves logical units (functions, classes)
* Metadata aids discovery
* Natural code boundaries

**Cons:**
* Language-specific parsing
* Complex implementation
* May miss cross-function context

## ADVANCED CHUNKING TECHNIQUES

### SEMANTIC SIMILARITY-BASED
Group sentences by semantic similarity.

```python
from sentence_transformers import SentenceTransformer
from sklearn.cluster import AgglomerativeClustering

def semantic_chunking(text, model, max_chunk_size=512):
    sentences = nltk.sent_tokenize(text)
    embeddings = model.encode(sentences)

    # Cluster similar sentences
    clustering = AgglomerativeClustering(
        n_clusters=None,
        distance_threshold=0.5
    )
    labels = clustering.fit_predict(embeddings)

    # Group sentences by cluster
    chunks = {}
    for sent, label in zip(sentences, labels):
        chunks.setdefault(label, []).append(sent)

    return [' '.join(sents) for sents in chunks.values()]
```
**Pros:**
* Truly semantic grouping
* Handles topic shifts
* Optimal information density

**Cons:**
* Computationally expensive
* Requires embedding model
* Complex to tune

### SLIDING WINDOW WITH CONTEXTUAL OVERLAP
Add surrounding context to each chunk.

```python
def sliding_window_chunk(text, window_size=512, context_size=128):
    tokens = tokenize(text)
    chunks = []

    for i in range(0, len(tokens), window_size):
        # Main window
        start = max(0, i - context_size)
        end = min(len(tokens), i + window_size + context_size)

        chunk = {
            'content': detokenize(tokens[i:i+window_size]),
            'context': detokenize(tokens[start:end]),
            'position': i
        }
        chunks.append(chunk)

    return chunks
```
**Pros:**
* Each chunk has surrounding context
* Reduces information loss
* Better for cross-boundary queries

**Cons:**
* Larger storage requirements
* More embeddings needed
* Potential redundancy

### HYBRID HIERARCHICAL CHUNKING
Chunk at multiple granularities.

```python
def hierarchical_chunk(document):
    # Level 1: Document
    doc_embedding = embed(document['content'])

    # Level 2: Sections
    sections = split_by_headers(document['content'])
    section_embeddings = [embed(s) for s in sections]

    # Level 3: Paragraphs
    paragraph_chunks = []
    for section in sections:
        paragraphs = section.split('\n\n')
        paragraph_chunks.extend([
            {'content': p, 'section': section} for p in paragraphs
        ])
    para_embeddings = [embed(p['content']) for p in paragraph_chunks]

    return {
        'document': {'embedding': doc_embedding, 'content': document},
        'sections': [{'embedding': e, 'content': s} for e, s in zip(section_embeddings, sections)],
        'paragraphs': [{'embedding': e, **p} for e, p in zip(para_embeddings, paragraph_chunks)]
    }
```
**Retrieval strategy:**
1. Search at document level
2. If match found, search within sections
3. Finally retrieve specific paragraphs

**Pros:**
* Multiple levels of granularity
* Coarse-to-fine retrieval
* Better context preservation

**Cons:**
* Complex implementation
* More storage needed
* Slower indexing

## CHUNK OVERLAP

### WHY OVERLAP?
Without overlap:
Chunk 1: "...the database schema includes user tables"
Chunk 2: "with columns for email and password..."
Query: "database user email" might miss both chunks

With overlap:
Chunk 1: "...the database schema includes user tables with columns for..."
Chunk 2: "...user tables with columns for email and password..."
Now "user tables with columns" appears in both, improving recall.

### OPTIMAL OVERLAP
| Chunk Size | Recommended Overlap | Ratio |
|---|---|---|
| 128 tokens | 10-20 tokens | 8-15% |
| 512 tokens | 50-100 tokens | 10-20% |
| 1024 tokens | 100-200 tokens | 10-20% |
| 2048 tokens | 200-400 tokens | 10-20% |

**Trade-offs:**
* More overlap: Better recall, more storage, slower search
* Less overlap: Less storage, faster search, may miss context

## CHUNKING FOR DIFFERENT CONTENT TYPES

**TECHNICAL DOCUMENTATION**
```python
# Recommended: Markdown-aware, preserve code blocks
chunk_size = 1024
overlap = 150
preserve_code_blocks = True
preserve_tables = True
```

**CUSTOMER SUPPORT TICKETS**
```python
# Recommended: Fixed-size with moderate overlap
chunk_size = 512
overlap = 100
split_by_turns = True  # Each Q&A turn
```

**RESEARCH PAPERS**
```python
# Recommended: Section-based with citations
split_by_sections = True
preserve_citations = True
chunk_size = 1024
```

**CODE REPOSITORIES**
```python
# Recommended: Syntactic splitting
split_by_functions = True
include_docstrings = True
chunk_size = 512
```

**CHAT LOGS**
```python
# Recommended: Message-based
chunk_by_messages = True
messages_per_chunk = 10
preserve_threading = True
```

## EVALUATING CHUNKING STRATEGIES

### RETRIEVAL METRICS
Test with query set:
```python
def evaluate_chunking(queries, ground_truth, chunking_fn):
    chunks = chunking_fn(documents)
    embeddings = embed(chunks)

    precision_scores = []
    recall_scores = []

    for query, expected_docs in zip(queries, ground_truth):
        retrieved = search(embed(query), embeddings, k=5)
        precision = len(set(retrieved) & set(expected_docs)) / len(retrieved)
        recall = len(set(retrieved) & set(expected_docs)) / len(expected_docs)

        precision_scores.append(precision)
        recall_scores.append(recall)

    return {
        'precision': np.mean(precision_scores),
        'recall': np.mean(recall_scores)
    }
```

### END-TO-END METRICS
Test full RAG pipeline:
* Answer accuracy
* Context utilization (how much of retrieved context is used)
* Answer groundedness (faithfulness to chunks)

## PRACTICAL RECOMMENDATIONS

### DECISION FRAMEWORK
1. Start simple: Fixed-size with overlap (512 tokens, 100 overlap)
2. Measure performance: Use evaluation metrics
3. Identify failures: Where does retrieval fail?
4. Iterate: Try semantic or metadata-aware chunking
5. A/B test: Compare strategies on real queries

### COMMON PATTERNS
**90% of use cases:**
* Recursive character splitting
* 512-1024 token chunks
* 10-20% overlap

**Structured documents:**
* Markdown/HTML-aware chunking
* Preserve metadata (headers, sections)
* Variable sizes OK

**Code:**
* Syntax-aware splitting
* Include docstrings with functions
* Smaller chunks (256-512)

**Hybrid search:**
* Multiple chunk sizes
* Hierarchical retrieval
* Worth the complexity for high-value apps

### COMMON PITFALLS
1. Too small chunks: Lose context, fragmented retrieval
2. Too large chunks: Irrelevant information, token waste
3. No overlap: Miss boundary-spanning queries
4. Ignoring structure: Arbitrary splits in tables, code, lists
5. One-size-fits-all: Different content needs different strategies
6. No evaluation: Guessing instead of measuring

> Expert Tip from Ailog: In production with 10M+ documents, we've found that starting with 512-token chunks and 10% overlap works for 80% of use cases. Only optimize further if you see retrieval failures in your evaluation metrics. The biggest mistake is over-engineering chunking before measuring actual performance. Start simple, measure, iterate.

## TRY CHUNKING STRATEGIES ON AILOG
Want to test different chunking approaches without writing code?

Ailog's platform lets you:
* Upload documents and compare chunking strategies side-by-side
* Test semantic vs fixed-size chunking instantly
* Visualize chunk boundaries and overlap
* Benchmark retrieval quality with real queries
* Deploy the best strategy to production in one click

Try it free -> No credit card required.

## NEXT STEPS
With documents properly chunked, the next step is selecting and configuring a vector database to store and search embeddings efficiently. This is covered in the next guide on vector databases.

## TAGS
chunking | document processing | retrieval | optimization | semantic chunking | langchain

## RELATED POSTS
**GuidesAdvanced**
HIERARCHICAL CHUNKING: PRESERVING DOCUMENT STRUCTURE
Hierarchical chunking maintains parent-child relationships in your documents. Learn how to implement this advanced technique to improve RAG retrieval quality. 11 min read

**GuidesBeginner**
FIXED-SIZE CHUNKING: FAST AND RELIABLE
Master the basics: implement fixed-size chunking with overlaps for consistent, predictable RAG performance. 7 min read

**GuidesAdvanced**
SEMANTIC CHUNKING FOR BETTER RETRIEVAL
Split documents intelligently based on meaning, not just length. Learn semantic chunking techniques for RAG. 12 min read
