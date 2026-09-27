"""Local RAG Service — drop-in stand-in for the Go API Gateway (port 8080).

Implements the same HTTP contract (pkg/types/types.go):
  GET  /health
  GET  /version
  POST /api/v1/query      {query, session_id?, grade?, subject?, stream}
  POST /api/v1/feedback
  GET  /api/v1/stats
  POST /api/v1/ingest     (local-mode extra: ingest markdown/text documents)

Runs fully offline: ONNX MiniLM embeddings + BM25 + RRF hybrid retrieval +
bundled cross-encoder reranker + extractive (or LLM, if configured) answers.
"""

from __future__ import annotations

import hashlib
import json
import time
import uuid
from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import StreamingResponse
from pydantic import BaseModel, Field

from .answerer import Answerer
from .chunker import parse_markdown_document
from .config import settings
from .embedder import Embedder
from .reranker import Reranker
from .retrieval import Retriever
from .store import VectorStore

VERSION = "1.0.0-local"

# ----------------------------------------------------------------- lifecycle

state: dict = {}


@asynccontextmanager
async def lifespan(app: FastAPI):
    t0 = time.time()
    store = VectorStore(settings.db_path, settings.embed_dimension)
    embedder = Embedder(settings.embed_model_path, settings.embed_tokenizer_path)
    reranker = None
    if settings.rerank_enabled:
        try:
            reranker = Reranker(
                settings.reranker_model_path, settings.reranker_tokenizer_path
            )
        except Exception as exc:  # model missing → degrade gracefully
            print(f"[local-rag] reranker unavailable: {exc}")
    state["store"] = store
    state["retriever"] = Retriever(store, embedder, reranker)
    state["embedder"] = embedder
    state["answerer"] = Answerer(store)
    state["cache"] = {}
    state["metrics"] = {"queries": 0, "cache_hits": 0, "started_at": time.time()}
    print(
        f"[local-rag] ready in {time.time()-t0:.1f}s — "
        f"{store.size} chunks indexed, mode={settings.app_mode}, "
        f"llm={settings.llm_provider}, rerank={reranker is not None}"
    )
    yield


app = FastAPI(title="Visionary Local RAG Service", version=VERSION, lifespan=lifespan)
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

# -------------------------------------------------------------------- models


class QueryRequest(BaseModel):
    query: str = Field(min_length=1, max_length=10000)
    session_id: str | None = None
    grade: str | None = None
    subject: str | None = None
    stream: bool = False
    top_k: int | None = None


class IngestDocument(BaseModel):
    doc_id: str
    text: str
    metadata: dict = Field(default_factory=dict)


class IngestRequest(BaseModel):
    documents: list[IngestDocument]


class FeedbackRequest(BaseModel):
    query: str
    session_id: str | None = None
    response_id: str | None = None
    rating: int | None = None
    helpful: bool | None = None
    comment: str | None = None


# ----------------------------------------------------------------- endpoints


@app.get("/health")
def health():
    return {
        "status": "healthy",
        "version": VERSION,
        "mode": settings.app_mode,
        "llm_provider": settings.llm_provider,
        "indexed_chunks": state["store"].size,
    }


@app.get("/version")
def version():
    return {"version": VERSION, "service": "local-rag-service"}


def _run_query(req: QueryRequest) -> dict:
    store: VectorStore = state["store"]
    retriever: Retriever = state["retriever"]
    answerer: Answerer = state["answerer"]

    session_id = req.session_id or str(uuid.uuid4())
    request_id = str(uuid.uuid4())

    # Layer-1 exact-match cache (SHA-256 of query+filters), TTL-bounded
    key = hashlib.sha256(
        f"{req.query}|{req.grade}|{req.subject}|{req.top_k}".encode()
    ).hexdigest()
    cache = state["cache"]
    now = time.time()
    state["metrics"]["queries"] += 1
    if key in cache and now - cache[key]["ts"] < settings.cache_ttl_seconds:
        state["metrics"]["cache_hits"] += 1
        cached = dict(cache[key]["resp"])
        cached.update({"session_id": session_id, "request_id": request_id})
        return cached

    chunks = retriever.retrieve(
        req.query, top_k=req.top_k, grade=req.grade, subject=req.subject
    )
    answer, tokens = answerer.answer(req.query, chunks)
    confidence = round(float(chunks[0].score), 4) if chunks else 0.0

    resp = {
        "session_id": session_id,
        "request_id": request_id,
        "answer": answer,
        "sources": [
            {"parent_id": c.parent_id, "content": c.content, "score": c.score}
            for c in chunks
        ],
        "total_tokens": tokens,
        "confidence_score": confidence,
        "was_fallback": len(chunks) == 0,
    }
    if len(cache) < settings.cache_max_entries:
        cache[key] = {"ts": now, "resp": resp}
    return resp


@app.post("/api/v1/query")
def query(req: QueryRequest):
    if not req.query.strip():
        raise HTTPException(400, detail={"code": "INVALID_REQUEST", "message": "query is required"})
    resp = _run_query(req)

    if not req.stream:
        return resp

    def sse():
        yield "event: start\ndata: " + json.dumps(
            {"session_id": resp["session_id"], "request_id": resp["request_id"]}
        ) + "\n\n"
        words = resp["answer"].split(" ")
        for i in range(0, len(words), 8):
            chunk = " ".join(words[i : i + 8])
            if i + 8 < len(words):
                chunk += " "
            yield "event: chunk\ndata: " + json.dumps({"content": chunk}) + "\n\n"
        yield "event: end\ndata: " + json.dumps(
            {"sources": resp["sources"], "total_tokens": resp["total_tokens"]}
        ) + "\n\n"

    return StreamingResponse(sse(), media_type="text/event-stream")


@app.post("/api/v1/ingest")
def ingest(req: IngestRequest):
    store: VectorStore = state["store"]
    embedder: Embedder = state["embedder"]
    total_parents = total_children = 0
    for doc in req.documents:
        parents, children = parse_markdown_document(
            doc.doc_id,
            doc.text,
            doc.metadata,
            settings.chunk_size,
            settings.chunk_overlap,
        )
        if not children:
            continue
        embeddings = embedder.embed([c.content for c in children])
        store.upsert(parents, children, embeddings)
        total_parents += len(parents)
        total_children += len(children)
    state["cache"].clear()
    return {
        "status": "ok",
        "documents": len(req.documents),
        "parents_indexed": total_parents,
        "chunks_indexed": total_children,
        "total_chunks": store.size,
    }


@app.post("/api/v1/feedback")
def feedback(req: FeedbackRequest):
    fid = state["store"].save_feedback(req.model_dump())
    return {"status": "recorded", "feedback_id": fid}


@app.get("/api/v1/stats")
def stats():
    m = state["metrics"]
    s = state["store"].stats()
    hit_rate = m["cache_hits"] / m["queries"] if m["queries"] else 0.0
    return {
        "uptime_seconds": round(time.time() - m["started_at"], 1),
        "total_queries": m["queries"],
        "cache_hit_rate": round(hit_rate, 3),
        "index": s,
        "config": {
            "retrieval_strategy": settings.retrieval_strategy,
            "top_k": settings.top_k,
            "rrf_k": settings.rrf_k,
            "similarity_threshold": settings.similarity_threshold,
            "chunk_size": settings.chunk_size,
            "chunk_overlap": settings.chunk_overlap,
            "llm_provider": settings.llm_provider,
        },
    }


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host=settings.host, port=settings.port)
