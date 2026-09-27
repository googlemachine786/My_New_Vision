# Science Class 8 PDF Test Report

**Date**: March 27, 2026  
**PDF**: science class 8.pdf (32 MB, 265 pages)  
**Dataset**: science_dataset.xlsx (470 Q&A pairs)

---

## ✅ File Verification

### PDF File
- **Status**: ✅ Found
- **Size**: 32.04 MB
- **Pages**: 265
- **Format**: CBSE Science Class 8 Textbook

### Golden Dataset
- **Status**: ✅ Found
- **Format**: Excel (.xlsx)
- **Rows**: 470 Q&A pairs
- **Columns**: `x_data` (question), `y_data` (answer)

---

## 📊 PDF Content Analysis

### First Page Preview
```
SCIENCE
TEXTBOOK FOR CLASS VIII
SCIENCE
2018-19
```

### Expected Chapters (CBSE Class 8 Science)
1. Crop Production and Management
2. Microorganisms: Friend and Foe
3. Synthetic Fibres and Plastics
4. Materials: Metals and Non-Metals
5. Coal and Petroleum
6. Combustion and Flame
7. Conservation of Plants and Animals
8. Cell — Structure and Functions
9. Reproduction in Animals
10. Reaching the Age of Adolescence
11. Force and Pressure
12. Friction
13. Sound
14. Chemical Effects of Electric Current
15. Some Natural Phenomena
16. Light
17. Stars and the Solar System
18. Pollution of Air and Water

---

## 🧪 Test Results

### 1. File Verification ✅
```
✅ PDF found: science class 8.pdf (32.04 MB)
✅ Dataset found: science_dataset.xlsx (470 rows)
✅ Dataset columns: ['x_data', 'y_data']
```

### 2. PDF Accessibility ✅
```
✅ PDF opened successfully
✅ Pages: 265
✅ Text extraction working
```

### 3. Dataset Structure ✅
```
✅ Total questions: 470
✅ Format: Question (x_data) + Answer (y_data)
✅ Ready for RAG evaluation
```

---

## 📈 Sample Questions from Dataset

Based on the dataset structure, expected questions include:

1. **Biology**
   - What is photosynthesis?
   - What are the parts of a cell?
   - How do plants reproduce?

2. **Chemistry**
   - What is combustion?
   - What are synthetic fibres?
   - What is the difference between metals and non-metals?

3. **Physics**
   - What is force?
   - How is sound produced?
   - What is friction?

4. **Environmental Science**
   - How to conserve plants and animals?
   - What causes air pollution?
   - How to reduce pollution?

---

## 🎯 Next Steps for Full Testing

### Step 1: Ingest PDF into Vector Store
```bash
cd ingestion-go
go run cmd/ingestion/main.go \
  --pdf "science class 8.pdf" \
  --grade 8 \
  --subject Science \
  --taxonomy-id 35
```

### Step 2: Test RAG Queries
```python
queries = [
    "What is photosynthesis?",
    "What are the parts of a cell?",
    "What is force and pressure?",
]

for query in queries:
    response = rag_chain.call(query)
    print(f"Q: {query}")
    print(f"A: {response}")
```

### Step 3: Evaluate Against Golden Dataset
```python
from datasets import load_dataset

# Load golden dataset
dataset = pd.read_excel("science_dataset.xlsx")

# Evaluate each question
for idx, row in dataset.iterrows():
    question = row['x_data']
    golden_answer = row['y_data']
    
    # Get RAG response
    rag_answer = rag_chain.call(question)
    
    # Calculate metrics
    metrics = evaluate(rag_answer, golden_answer)
```

---

## 📊 Expected Metrics

### Retrieval Metrics
- **Recall@5**: ≥0.85 (target)
- **MRR**: ≥0.70 (target)
- **NDCG@10**: ≥0.80 (target)

### Generation Metrics
- **BLEU**: ≥0.30 (target)
- **ROUGE-L**: ≥0.40 (target)
- **BERTScore**: ≥0.70 (target)

---

## 🚀 Pipeline Status

| Component | Status | Ready |
|-----------|--------|-------|
| **PDF Ingestion** | ✅ File verified | Yes |
| **Dataset Loading** | ✅ Loaded (470 rows) | Yes |
| **Text Extraction** | ✅ Working | Yes |
| **Vector Store** | ⏳ Pending ingestion | Soon |
| **RAG Queries** | ⏳ Pending setup | Soon |
| **Evaluation** | ⏳ Pending metrics | Soon |

---

## 📝 Test Commands

### Quick Test (Local)
```bash
# 1. Start local services
docker-compose up -d

# 2. Ingest PDF
cd ingestion-go
go run cmd/ingestion/main.go --pdf "science class 8.pdf" --grade 8

# 3. Run queries
python test_science_class8.py
```

### Full Test (with Evaluation)
```bash
python test_science_class8.py --evaluate --dataset science_dataset.xlsx
```

---

## ✅ Summary

**What Works Now:**
- ✅ PDF file verified (32 MB, 265 pages)
- ✅ Golden dataset loaded (470 Q&A pairs)
- ✅ Text extraction working
- ✅ Pipeline ready for ingestion

**What's Next:**
1. Ingest PDF into vector store (FAISS for local)
2. Test RAG queries against ingested content
3. Evaluate all 470 questions from dataset
4. Calculate retrieval and generation metrics

**Status**: 🟡 **Ready for full pipeline testing!**

---

**Test Script**: `test_science_class8.py`  
**Results**: Saved to `test_results_science_class8.json`  
**Next Run**: After PDF ingestion into vector store
