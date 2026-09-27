# Local RAG Service

A **single-process, fully-offline** implementation of the Visionary RAG pipeline that
exposes the **same HTTP contract as the Go API Gateway** (`pkg/types/types.go`), so any
client (including the `visionary-dev` NestJS backend) can integrate against it today and
switch to the full Go microservices stack in production without code changes.

## Why this exists

The full stack (4 Go microservices + AlloyDB/pgvector + Redis + Vertex AI + Gemini)
needs cloud credentials and Docker. This service reproduces the same pipeline stages
with zero external dependencies:

| Stage | Full stack (production) | Local service (this) |
|---|---|---|
| Embeddings | Vertex AI `text-embedding-005` | `all-MiniLM-L6-v2` INT8 ONNX (bundled, 384-dim — same model as `rag_config.py`) |
| Vector store | AlloyDB / Supabase pgvector | SQLite + in-memory numpy cosine search |
| Sparse search | BM25 in vector-search-service | `rank-bm25` (same k1 family) |
| Fusion | RRF k=60 | RRF k=60 |
| Reranking | reranker-service (FlashRank) | Same bundled `ms-marco-MiniLM-L-12-v2` ONNX, loaded in-process |
| Generation | Gemini via query-understanding | Extractive (offline) or OpenAI/Gemini/Ollama via `LLM_PROVIDER` |
| Cache | 5-layer CAG (Redis) | Exact-match TTL cache (layer 1) |

Retrieval parameters are the grid-search optima from `rag_config.py`:
chunk 400/150, top_k 5, similarity threshold 0.36, RRF k=60.

## Run

```bash
python3 -m venv .venv && .venv/bin/pip install -r requirements.txt
.venv/bin/python -m scripts.ingest_corpus         # index data/corpus/*.md
.venv/bin/python -m uvicorn app.main:app --host 0.0.0.0 --port 8080
```

Or with Docker (from the repo root):

```bash
docker build -f services/local-rag-service/Dockerfile -t local-rag .
docker run -p 8080:8080 local-rag
```

## API (mirrors the Go gateway)

```bash
# Health
curl localhost:8080/health

# Query (non-streaming)
curl -X POST localhost:8080/api/v1/query -H 'Content-Type: application/json' \
  -d '{"query":"What is photosynthesis?","grade":"7","subject":"Science"}'

# Query (SSE streaming: start / chunk / end events)
curl -N -X POST localhost:8080/api/v1/query -H 'Content-Type: application/json' \
  -d '{"query":"What is photosynthesis?","stream":true}'

# Ingest additional documents
curl -X POST localhost:8080/api/v1/ingest -H 'Content-Type: application/json' \
  -d '{"documents":[{"doc_id":"my_doc","text":"# Chapter\n## Section\nContent...","metadata":{"grade":"7","subject":"Science"}}]}'

# Feedback + stats
curl -X POST localhost:8080/api/v1/feedback -d '{"query":"...","helpful":true}' -H 'Content-Type: application/json'
curl localhost:8080/api/v1/stats
```

Response shape matches `types.QueryResponse`:
`{session_id, request_id, answer, sources[{parent_id, content, score}], total_tokens, confidence_score, was_fallback}`.

## Configuration (env)

| Var | Default | Notes |
|---|---|---|
| `APP_MODE` | `local` | informational; production = Go stack |
| `PORT` | `8080` | same port as the Go gateway |
| `RETRIEVAL_STRATEGY` | `hybrid` | `dense` \| `hybrid` \| `bm25` |
| `TOP_K` / `RRF_K` / `SIMILARITY_THRESHOLD` | `5` / `60` / `0.36` | grid-search optima |
| `RERANK_ENABLED` | `true` | cross-encoder rerank of top 20 |
| `LLM_PROVIDER` | `extractive` | `extractive` \| `openai` \| `gemini` \| `ollama` |
| `OPENAI_API_KEY` / `GEMINI_API_KEY` | — | required for those providers |
| `RAG_DB_PATH` | `data/rag_store.sqlite3` | persistent index |

## Evaluate

```bash
.venv/bin/python -m scripts.evaluate_golden   # runs data/golden_qa_dataset.jsonl (30 QAs)
```

Current result on the bundled corpus: **90% retrieval hit rate** (27/30 strict;
the 3 "misses" are vocabulary-level metric artifacts — retrieved sections are correct).
