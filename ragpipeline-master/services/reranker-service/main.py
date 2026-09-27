"""FlashRank reranker sidecar service.

Provides a FastAPI /rerank endpoint that uses FlashRank for
production-quality cross-encoder reranking.
"""

from fastapi import FastAPI
from pydantic import BaseModel
from flashrank import Ranker, RerankRequest
from typing import List, Optional

app = FastAPI(title="Reranker Service", version="1.0.0")

# Load the model once at startup.
# ms-marco-MiniLM-L-12-v2 is the best cross-encoder reranker (~34MB).
ranker = Ranker(model_name="ms-marco-MiniLM-L-12-v2", cache_dir="/app/models")


class Passage(BaseModel):
    id: str
    text: str
    meta: Optional[dict] = None


class RerankRequest(BaseModel):
    query: str
    passages: List[Passage]


class RerankResult(BaseModel):
    id: str
    text: str
    score: float
    meta: Optional[dict] = None


class RerankResponse(BaseModel):
    results: List[RerankResult]


@app.post("/rerank", response_model=RerankResponse)
async def rerank(req: RerankRequest) -> RerankResponse:
    passages = [
        {"id": p.id, "text": p.text, "meta": p.meta} for p in req.passages
    ]
    rerank_req = RerankRequest(query=req.query, passages=passages)
    results = ranker.rerank(rerank_req)
    return RerankResponse(
        results=[
            RerankResult(
                id=r["id"],
                text=r["text"],
                score=r["score"],
                meta=r.get("meta"),
            )
            for r in results
        ]
    )


@app.get("/health")
async def health() -> dict:
    return {"status": "healthy", "model": "ms-marco-MiniLM-L-6-v2"}
