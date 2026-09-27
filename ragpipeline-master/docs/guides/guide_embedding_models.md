# BEST EMBEDDING MODELS 2025: MTEB SCORES & LEADERBOARD (COHERE, OPENAI, BGE)

**January 16, 2026** | **11 min read** | **Ailog Research Team**

Compare MTEB scores for top embedding models: Cohere embed-v4 (65.2), OpenAI text-3-large (64.6), BGE-M3 (63.0). Full leaderboard with pricing.

## MTEB LEADERBOARD 2025 - TOP EMBEDDING MODELS

Quick reference table with MTEB scores for all major embedding models:

| Rank | Model | MTEB Score | Dimensions | Price/1M tokens | Best For |
|------|-------|------------|------------|-----------------|----------|
| 1 | Gemini-embedding-001 | 68.3 | 3072 | ~$0.004/1K | Overall best, multilingual |
| 2 | Qwen3-Embedding-8B | 70.58* | 4096 | Free | Best open-source |
| 3 | Voyage-3-large | 66.8 | 1536 | $0.12 | Domain tuning |
| 4 | Cohere embed-v4 | 65.2 | 1024 | $0.10 | Enterprise, noisy data |
| 5 | OpenAI text-embedding-3-large | 64.6 | 3072 | $0.13 | General purpose |
| 6 | BGE-M3 | 63.0 | 1024 | Free | Self-hosted budget |
| 7 | Nomic-embed-text-v1.5 | 59.4 | 768 | $0.05 | Budget option |
| 8 | all-MiniLM-L6-v2 | 56.3 | 384 | Free | Fast prototyping |

*Qwen3-Embedding-8B scores 70.58 on MTEB Multilingual leaderboard. Last updated: January 2026. Source: MTEB Leaderboard*

---

## EMBEDDING MODEL LANDSCAPE (2025)

The embedding space has evolved dramatically. Here's what's leading:

### TOP MODELS BY MTEB SCORE

**1. Gemini-embedding-001 (NEW #1)**
* **Dimensions:** 3072
* **MTEB Score:** 68.32 (+5.81 over competitors)
* **Cost:** ~$0.004 per 1K tokens
* **Best for:** Overall best, multilingual (100+ languages)

**2. Qwen3-Embedding-8B (Best Open-Source)**
* **Dimensions:** 4096
* **MTEB Score:** 70.58 (multilingual leaderboard)
* **Cost:** Free (Apache 2.0 license)
* **Best for:** Self-hosted, privacy-first, multilingual

**3. Voyage-3-large**
* **Dimensions:** 1536
* **MTEB Score:** 66.8
* **Cost:** $0.12 per 1M tokens
* **Best for:** Domain-specific tuning

**4. Cohere embed-v4**
* **Dimensions:** 1024
* **MTEB Score:** 65.2
* **Cost:** $0.10 per 1M tokens
* **Best for:** Enterprise, noisy real-world data

**5. OpenAI text-embedding-3-large**
* **Dimensions:** 3072 (configurable down to 256)
* **MTEB Score:** 64.6
* **Cost:** $0.13 per 1M tokens
* **Best for:** General purpose, existing OpenAI stack

### KEY DECISION FACTORS

**1. ACCURACY VS COST**
```python
# High accuracy: OpenAI or Cohere
from openai import OpenAI
client = OpenAI()

response = client.embeddings.create(
    model="text-embedding-3-large",
    input="Your text here"
)
embedding = response.data[0].embedding

# Budget option: Open-source
from sentence_transformers import SentenceTransformer
model = SentenceTransformer('BAAI/bge-large-en-v1.5')
embedding = model.encode("Your text here")
```

**2. DIMENSION SIZE**
Smaller = faster, cheaper storage, but less accurate
```python
# OpenAI: Configurable dimensions
response = client.embeddings.create(
    model="text-embedding-3-large",
    input="text",
    dimensions=512  # vs default 3072
)
```

**3. LANGUAGE SUPPORT**
Multilingual leaders:
* Cohere embed-v4: 100+ languages
* BGE-M3: 100+ languages
* OpenAI text-embedding-3-large: 100+ languages

**4. DOMAIN SPECIALIZATION**
* **Code:** OpenAI text-embedding-3-small, Voyage code-2
* **Legal:** Fine-tuned BGE on legal corpus
* **Medical:** BioGPT embeddings, PubMedBERT

### BENCHMARKING YOUR USE CASE

Don't trust generic benchmarks - test on YOUR data:
```python
from sentence_transformers import SentenceTransformer, util

def benchmark_model(model_name, queries, documents):
    model = SentenceTransformer(model_name)

    # Embed
    query_embs = model.encode(queries)
    doc_embs = model.encode(documents)

    # Calculate similarities
    similarities = util.cos_sim(query_embs, doc_embs)

    return similarities

# Test multiple models
models = [
    "text-embedding-3-large",
    "BAAI/bge-large-en-v1.5",
    "sentence-transformers/all-MiniLM-L6-v2"
]

for model in models:
    scores = benchmark_model(model, test_queries, test_docs)
    print(f"{model}: {scores.mean()}")
```

### MATRYOSHKA EMBEDDINGS (2025-2026)

New models support variable dimensions from the same embedding:
```python
# Generate once at full dimension
full_embedding = model.encode(text, dimension=1024)

# Truncate later as needed
small_embedding = full_embedding[:256]  # Just use first 256
medium_embedding = full_embedding[:512]

# Quality degrades gracefully, not catastrophically
```
**Models supporting this:**
* OpenAI text-embedding-3-*
* Nomic embed-v1.5
* Jina embeddings v2

### FINE-TUNING FOR YOUR DOMAIN
```python
from sentence_transformers import SentenceTransformer, InputExample, losses
from torch.utils.data import DataLoader

# Load base model
model = SentenceTransformer('BAAI/bge-base-en-v1.5')

# Create training examples
train_examples = [
    InputExample(texts=['query', 'positive_doc', 'negative_doc'])
]

train_dataloader = DataLoader(train_examples, shuffle=True, batch_size=16)

# Fine-tune
train_loss = losses.MultipleNegativesRankingLoss(model)
model.fit(
    train_objectives=[(train_dataloader, train_loss)],
    epochs=1,
    warmup_steps=100
)
```

### COST ANALYSIS (1M DOCUMENTS)

| Model | Embedding Cost | Storage | Inference |
|-------|----------------|---------|-----------|
| Gemini-embedding-001 | ~$40 | $50/month | ~$0.004/1K queries |
| OpenAI-3-large | $130 | $50/month | $0.13/1M queries |
| Cohere v4 | $100 | $50/month | $0.10/1M queries |
| Qwen3-8B (self-hosted) | $0 | $50/month | GPU: $100/month |
| all-MiniLM | $0 | $20/month | CPU: $20/month |

### RECOMMENDATIONS BY USE CASE

* **Startup/MVP:** `all-MiniLM-L6-v2` (free, fast)
* **Production (quality matters):** `Cohere embed-v4` or `OpenAI text-embedding-3-large`
* **Production (budget matters):** `BGE-M3` self-hosted
* **Multilingual:** `Cohere embed-v4` or `BGE-M3`
* **Code search:** `Voyage code-2` or `OpenAI text-embedding-3-small`
* **Privacy-critical:** `BGE-M3` (MIT license, self-hosted)
* **Enterprise (noisy data):** `Cohere embed-v4`

### MIGRATION STRATEGY

Changing embeddings requires re-embedding everything:
```python
# Gradual migration
def hybrid_search(query, old_index, new_index, alpha=0.5):
    # Search both indices
    old_results = old_index.search(old_model.encode(query))
    new_results = new_index.search(new_model.encode(query))

    # Blend results
    return blend_rankings(old_results, new_results, alpha)
```

The embedding model is your RAG foundation. Choose wisely, benchmark thoroughly, and be ready to upgrade as models improve.

### FAQ

**Which embedding model is best for RAG?**
Cohere embed-v4 leads with 65.2 on MTEB, followed closely by OpenAI text-embedding-3-large (64.6). For open-source, BGE-M3 (63.0) offers excellent performance at no cost.

**Is OpenAI text-embedding-3-large worth the cost?**
Yes. With a 64.6 MTEB score, it's among the top performers and integrates seamlessly with the OpenAI ecosystem. Consider Cohere for slightly better multilingual performance.

**What's the best free embedding model?**
BGE-M3 is the top open-source choice with 63.0 on MTEB, supporting 100+ languages. For English-only use cases, all-MiniLM-L6-v2 offers fast, lightweight embeddings.

**How do I choose between embedding models?**
Consider: (1) accuracy requirements, (2) language support, (3) cost constraints, (4) latency needs. Benchmark on YOUR data - generic scores don't always translate to your domain.

**Should I fine-tune my embedding model?**
Fine-tuning shows +10-30% gains for specialized domains (legal, medical, code). Start with a pre-trained model, then fine-tune if generic performance is insufficient.

---

**TAGS**
`embeddings` `models` `benchmarks` `mteb` `openai` `cohere` `bge-m3` `2025` `leaderboard`

**RELATED POSTS**
* **EMBEDDINGS: THE FOUNDATION OF SEMANTIC SEARCH** (12 min read) - Deep dive into embedding models, vector representations, and how to choose the right embedding strategy for your RAG system.
* **MULTILINGUAL EMBEDDINGS FOR GLOBAL RAG** (11 min read) - Build RAG systems that work across languages using multilingual embedding models and cross-lingual retrieval.
* **FINE-TUNE EMBEDDINGS FOR YOUR DOMAIN** (14 min read) - Boost RAG retrieval accuracy by 30-50% with domain-specific fine-tuning. Learn to create custom embeddings for your documents and queries.
