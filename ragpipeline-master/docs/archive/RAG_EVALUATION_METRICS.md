# RAG Evaluation Metrics Guide

**Comprehensive guide to evaluating Retrieval-Augmented Generation systems**

---

## 📊 Evaluation Metrics Overview

### 1. Retrieval Metrics

These measure how well the system retrieves relevant documents.

#### Recall@K
**What it measures**: Percentage of relevant documents retrieved in top-K results.

**Formula**:
```
Recall@K = (Relevant documents in top-K) / (Total relevant documents)
```

**Target**: ≥0.85 (85% of relevant docs retrieved)

**Example**:
- Query: "What is photosynthesis?"
- Relevant chunks: 5 (in entire corpus)
- Retrieved in top-3: 3
- **Recall@3 = 3/5 = 0.60 (60%)**

---

#### MRR (Mean Reciprocal Rank)
**What it measures**: How high the first relevant document appears.

**Formula**:
```
MRR = 1 / rank_of_first_relevant_result
```

**Target**: ≥0.70

**Example**:
- Query: "What is force?"
- First relevant chunk at position 2
- **MRR = 1/2 = 0.50**

---

#### NDCG@10 (Normalized Discounted Cumulative Gain)
**What it measures**: Quality of ranking (relevant docs should appear higher).

**Formula**:
```
DCG@10 = Σ (relevance_i / log2(i+2)) for i=1 to 10
NDCG@10 = DCG@10 / Ideal_DCG@10
```

**Target**: ≥0.80

**Range**: 0.0 (worst) to 1.0 (best)

---

### 2. Generation Metrics

These measure the quality of generated answers.

#### Answer Relevance
**What it measures**: Does the answer address the query?

**Calculation**: Cosine similarity between query and answer embeddings.

**Target**: ≥0.70

**Example**:
- Query: "What is photosynthesis?"
- Answer: "Photosynthesis is the process..." → **High relevance (0.85)**
- Answer: "The cell has a nucleus" → **Low relevance (0.15)**

---

#### Faithfulness
**What it measures**: Is the answer grounded in the retrieved context?

**Calculation**: Cosine similarity between answer and context embeddings.

**Target**: ≥0.75

**Example**:
- Context: "Force is a push or pull on an object"
- Answer: "Force is defined as a push or pull" → **High faithfulness (0.92)**
- Answer: "Force is energy" → **Low faithfulness (0.35)**

---

#### Answer Correctness
**What it measures**: How similar is the answer to the golden (correct) answer?

**Calculation**: Cosine similarity between generated answer and golden answer embeddings.

**Target**: ≥0.65

**Example**:
- Golden: "Photosynthesis produces glucose and oxygen"
- Generated: "Photosynthesis creates glucose and O₂" → **High correctness (0.88)**
- Generated: "Photosynthesis happens in leaves" → **Medium correctness (0.45)**

---

### 3. Performance Metrics

#### Average Latency
**What it measures**: Average time to generate answer.

**Target**: <500ms for production, <2000ms acceptable for testing

**Example**:
- 10 queries, total time: 12.5 seconds
- **Avg Latency = 1250ms**

---

#### P95 Latency
**What it measures**: 95th percentile latency (worst-case performance).

**Target**: <800ms

**Example**:
- Sorted latencies: [100, 150, 200, ..., 1800, 2100]
- **P95 = 1800ms** (95% of queries faster than this)

---

#### Success Rate
**What it measures**: Percentage of queries that return valid answers.

**Formula**:
```
Success Rate = (Successful queries) / (Total queries)
```

**Target**: ≥0.95 (95%)

---

### 4. Overall Score

**Weighted combination of all metrics**:

```
Overall Score = 
  0.25 × Recall@5 +
  0.15 × MRR +
  0.20 × Answer Relevance +
  0.20 × Faithfulness +
  0.20 × Answer Correctness
```

**Target**: ≥0.75

---

## 🎯 Metric Targets by Use Case

| Use Case | Recall@5 | MRR | Faithfulness | Latency |
|----------|----------|-----|--------------|---------|
| **Production (Student-facing)** | ≥0.85 | ≥0.70 | ≥0.80 | <500ms |
| **Testing/Development** | ≥0.75 | ≥0.60 | ≥0.70 | <2000ms |
| **Research/Experimental** | ≥0.70 | ≥0.55 | ≥0.65 | <5000ms |

---

## 📈 Continuous Improvement Process

### Step 1: Baseline Measurement
```bash
python test_continuous_improvement.py
```

**Record initial metrics:**
- Recall@5: 0.65
- MRR: 0.55
- Faithfulness: 0.70
- Overall: 0.62

---

### Step 2: Identify Bottlenecks

**Low Recall?**
- Problem: Retrieval not finding relevant chunks
- Solution: Better chunking, more chunks, better embeddings

**Low Faithfulness?**
- Problem: LLM hallucinating (making things up)
- Solution: Better prompts, more context, smaller LLM temperature

**High Latency?**
- Problem: Slow embeddings or LLM
- Solution: Faster models, GPU acceleration, caching

---

### Step 3: Test Improvements

**Test Different Configurations:**

| Config | Embed Model | LLM Model | Top-K | Chunk Size |
|--------|-------------|-----------|-------|------------|
| A (baseline) | nomic-embed-text | llama3.2:3b | 3 | 500 |
| B | nomic-embed-text | llama3.2:latest | 5 | 500 |
| C | nomic-embed-text | mistral:7b | 5 | 600 |
| D | all-minilm | llama3.2:3b | 3 | 400 |

**Measure metrics for each configuration.**

---

### Step 4: Deploy Best Configuration

**Select configuration with highest overall score:**

```
Config B wins:
- Recall@5: 0.78 (+0.13)
- MRR: 0.65 (+0.10)
- Faithfulness: 0.82 (+0.12)
- Overall: 0.75 (+0.13) ✅
```

**Update production configuration:**
```python
# In production config
TOP_K = 5
LLM_MODEL = "llama3.2:latest"
CHUNK_SIZE = 500
```

---

### Step 5: Monitor & Repeat

**Set up continuous monitoring:**
```bash
# Run evaluation weekly
python test_continuous_improvement.py --weekly

# Track metrics over time
python analyze_trends.py
```

**Expected improvement over time:**
- Week 1: Overall Score 0.62
- Week 2: Overall Score 0.68 (+0.06)
- Week 3: Overall Score 0.72 (+0.04)
- Week 4: Overall Score 0.75 (+0.03) ✅ Target reached!

---

## 🛠️ Implementation Details

### Calculating Cosine Similarity

```python
def cosine_similarity(vec1, vec2):
    dot_product = sum(a*b for a, b in zip(vec1, vec2))
    norm1 = sum(a*a for a in vec1) ** 0.5
    norm2 = sum(a*a for a in vec2) ** 0.5
    return dot_product / (norm1 * norm2) if norm1 and norm2 else 0.0
```

### Calculating Recall@K

```python
def recall_at_k(retrieved_docs, relevant_docs, k):
    retrieved_at_k = set(retrieved_docs[:k])
    relevant_set = set(relevant_docs)
    if not relevant_set:
        return 0.0
    return len(retrieved_at_k & relevant_set) / len(relevant_set)
```

### Calculating MRR

```python
def mean_reciprocal_rank(retrieved_docs, relevant_docs):
    for i, doc in enumerate(retrieved_docs):
        if doc in relevant_docs:
            return 1.0 / (i + 1)
    return 0.0
```

---

## 📊 Example Results Table

| Config | Recall@1 | Recall@3 | Recall@5 | MRR | NDCG@10 | Relevance | Faithfulness | Correctness | Overall |
|--------|----------|----------|----------|-----|---------|-----------|--------------|-------------|---------|
| **llama3.2-fast** | 0.45 | 0.62 | 0.71 | 0.58 | 0.68 | 0.75 | 0.78 | 0.68 | **0.68** |
| **llama3.2-accurate** | 0.52 | 0.68 | 0.78 | 0.65 | 0.74 | 0.78 | 0.82 | 0.72 | **0.75** ✅ |
| **mistral-quality** | 0.48 | 0.65 | 0.75 | 0.62 | 0.72 | 0.80 | 0.85 | 0.75 | **0.73** |

**Winner**: `llama3.2-accurate` (highest overall score: 0.75)

---

## 🎯 Quick Reference

| Metric | Excellent | Good | Acceptable | Poor |
|--------|-----------|------|------------|------|
| **Recall@5** | ≥0.90 | ≥0.80 | ≥0.70 | <0.70 |
| **MRR** | ≥0.80 | ≥0.70 | ≥0.60 | <0.60 |
| **NDCG@10** | ≥0.90 | ≥0.80 | ≥0.70 | <0.70 |
| **Relevance** | ≥0.85 | ≥0.75 | ≥0.65 | <0.65 |
| **Faithfulness** | ≥0.90 | ≥0.80 | ≥0.70 | <0.70 |
| **Correctness** | ≥0.80 | ≥0.70 | ≥0.60 | <0.60 |
| **Overall** | ≥0.85 | ≥0.75 | ≥0.65 | <0.65 |

---

## 🚀 Running Continuous Improvement

```bash
# Run with default settings (tests 3 configurations, 2 iterations each)
python test_continuous_improvement.py

# Results saved to:
# - data/continuous_improvement_results.json
# - data/continuous_improvement_results.csv
```

**Output includes:**
- ✅ All retrieval metrics (Recall@K, MRR, NDCG)
- ✅ All generation metrics (Relevance, Faithfulness, Correctness)
- ✅ Performance metrics (Latency, Success Rate)
- ✅ Overall score for each configuration
- ✅ Best configuration recommendation

---

## 📞 Support

For questions about metrics or improvement strategies:
- Review: `data/continuous_improvement_results.csv`
- Compare configurations side-by-side
- Focus on lowest metric for biggest improvement opportunity

**Remember**: Continuous improvement is a loop, not a one-time task! 🔄
