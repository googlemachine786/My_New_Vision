# Langchaingo Integration Plan

**Refactoring entire pipeline to use github.com/tmc/langchaingo**

---

## 🎯 What Langchaingo Provides

### Core Components We'll Use

1. **Text Splitters** ✅
   - `textsplitter.RecursiveCharacterTextSplitter`
   - `textsplitter.ParentChildTextSplitter`
   - Better than our custom implementation

2. **Embeddings** ✅
   - `embeddings/vertexai` - Vertex AI embeddings
   - `embeddings/ollama` - Ollama embeddings
   - Unified interface

3. **Vector Stores** ✅
   - `vectorstores/pgvector` - PostgreSQL with pgvector
   - `vectorstores/faiss` - FAISS for local
   - Switch between local/GCP

4. **LLMs** ✅
   - `llms/vertexai` - Vertex AI (Gemini)
   - `llms/ollama` - Ollama
   - Unified interface

5. **RAG Chains** ✅
   - `chains/retrieval_qa` - RAG retrieval
   - `chains/conversational_retrieval` - Multi-turn
   - Better than custom implementation

6. **Document Loaders** ✅
   - `documentloaders/pdf` - PDF parsing
   - `documentloaders/text` - Text files
   - Replaces custom parser

---

## 📋 Refactoring Tasks

### Phase 1: Text Splitting (Day 1)
- [ ] Replace custom chunker with `textsplitter.ParentChildTextSplitter`
- [ ] Keep metadata extraction logic
- [ ] Test with sample documents

### Phase 2: Embeddings (Day 2)
- [ ] Use `embeddings/vertexai` for GCP mode
- [ ] Use `embeddings/ollama` for local mode
- [ ] Unified interface via `embeddings.Embedder`

### Phase 3: Vector Stores (Day 3)
- [ ] Use `vectorstores/pgvector` for GCP (AlloyDB)
- [ ] Use `vectorstores/faiss` for local
- [ ] Auto-switch based on mode

### Phase 4: LLM Integration (Day 4)
- [ ] Use `llms/vertexai` for Gemini
- [ ] Use `llms/ollama` for local LLM
- [ ] Streaming support

### Phase 5: RAG Chains (Day 5)
- [ ] Use `chains/retrieval_qa` for single-turn
- [ ] Use `chains/conversational_retrieval` for multi-turn
- [ ] Custom prompts for CBSE education

### Phase 6: Document Loaders (Day 6)
- [ ] Use `documentloaders/pdf` for PDF parsing
- [ ] Keep custom 5-pass parser for complex textbooks
- [ ] Hybrid approach

### Phase 7: Testing (Day 7)
- [ ] Unit tests for each component
- [ ] Integration tests
- [ ] End-to-end tests

---

## 🔄 Migration Strategy

### Keep (Custom Implementation)
- ✅ 5-pass PDF parser (better for CBSE textbooks)
- ✅ Auto-detection logic (local vs GCP)
- ✅ Configuration system
- ✅ JWT authentication
- ✅ Session management

### Replace with Langchaingo
- ❌ Custom text splitter → `textsplitter.ParentChildTextSplitter`
- ❌ Custom embeddings → `embeddings/vertexai` or `embeddings/ollama`
- ❌ Custom vector store → `vectorstores/pgvector` or `vectorstores/faiss`
- ❌ Custom RAG logic → `chains/retrieval_qa`

---

## 📊 Benefits

**Before (Custom):**
- 500+ lines of custom text splitting code
- 300+ lines of custom embedding code
- 400+ lines of custom vector store code
- 600+ lines of custom RAG logic
- **Total: ~1,800 lines**

**After (Langchaingo):**
- 50 lines using `textsplitter.ParentChildTextSplitter`
- 30 lines using `embeddings/vertexai`
- 40 lines using `vectorstores/pgvector`
- 100 lines using `chains/retrieval_qa`
- **Total: ~220 lines (88% reduction!)**

---

## 🚀 Implementation Order

1. **Start with embeddings** (easiest, most isolated)
2. **Add vector stores** (builds on embeddings)
3. **Add text splitters** (independent)
4. **Add LLMs** (independent)
5. **Add RAG chains** (uses all above)
6. **Integration testing** (end-to-end)

---

## 📝 Code Examples

### Before (Custom)
```go
// Custom embedding client
type VertexClient struct {
    client *genai.Client
    model  string
}

func (c *VertexClient) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
    // 50+ lines of custom code
}
```

### After (Langchaingo)
```go
// Langchaingo embedding
emb, err := vertexai.New(ctx, vertexai.WithModel("text-embedding-005"))
embeddings, err := emb.EmbedDocuments(ctx, []string{text})
// Done! 2 lines vs 50+
```

---

## ✅ Success Criteria

- [ ] All existing tests pass
- [ ] Code reduced by 80%+
- [ ] Performance same or better
- [ ] Local mode still works (Ollama + FAISS)
- [ ] GCP mode still works (Vertex AI + AlloyDB)
- [ ] Auto-detection preserved
- [ ] No breaking changes to API

---

**Let's refactor with langchaingo!** 🚀
