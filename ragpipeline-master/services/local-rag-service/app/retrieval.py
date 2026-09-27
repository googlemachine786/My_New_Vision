"""Hybrid retrieval: dense (cosine) + sparse (BM25) fused with RRF (k=60),
similarity threshold 0.36, optional cross-encoder reranking.

Mirrors services/vector-search-service/search/{dense,sparse,rrf,hybrid}.go.
"""

from __future__ import annotations

from dataclasses import dataclass

from .config import settings
from .embedder import Embedder
from .reranker import Reranker
from .store import VectorStore


@dataclass
class RetrievedChunk:
    chunk_id: str
    parent_id: str
    content: str
    metadata: dict
    score: float


def rrf_fuse(
    ranked_lists: list[list[str]], k: float = 60.0
) -> dict[str, float]:
    """Reciprocal Rank Fusion: score(d) = sum(1 / (k + rank_i(d)))."""
    fused: dict[str, float] = {}
    for ranking in ranked_lists:
        for rank, doc_id in enumerate(ranking, start=1):
            fused[doc_id] = fused.get(doc_id, 0.0) + 1.0 / (k + rank)
    return fused


class Retriever:
    def __init__(self, store: VectorStore, embedder: Embedder, reranker: Reranker | None):
        self.store = store
        self.embedder = embedder
        self.reranker = reranker

    def retrieve(
        self,
        query: str,
        top_k: int | None = None,
        grade: str | None = None,
        subject: str | None = None,
        strategy: str | None = None,
    ) -> list[RetrievedChunk]:
        top_k = top_k or settings.top_k
        strategy = strategy or settings.retrieval_strategy

        query_emb = self.embedder.embed_one(query)
        dense = self.store.dense_search(
            query_emb, settings.top_k_dense, grade, subject
        )
        dense_scores = dict(dense)

        if strategy == "dense":
            candidates = [d for d, s in dense if s >= settings.similarity_threshold]
            scores = {d: dense_scores[d] for d in candidates}
        else:
            sparse = self.store.sparse_search(
                query, settings.top_k_sparse, grade, subject
            )
            if strategy == "bm25":
                candidates = [d for d, _ in sparse]
                scores = dict(sparse)
            else:  # hybrid: RRF fusion of both rankings
                fused = rrf_fuse(
                    [[d for d, _ in dense], [d for d, _ in sparse]], settings.rrf_k
                )
                # Apply the dense similarity threshold as a quality floor when
                # a chunk was only found by dense search.
                candidates = sorted(fused, key=fused.get, reverse=True)
                sparse_set = {d for d, _ in sparse}
                candidates = [
                    d
                    for d in candidates
                    if d in sparse_set
                    or dense_scores.get(d, 0.0) >= settings.similarity_threshold
                ]
                scores = {d: fused[d] for d in candidates}

        # Cross-encoder rerank of the top candidates
        pool = candidates[: settings.rerank_candidates]
        if self.reranker is not None and settings.rerank_enabled and pool:
            texts = [self.store.chunk(c)["content"] for c in pool]
            reranked = self.reranker.rerank(query, texts, top_k)
            final = [(pool[i], score) for i, score in reranked]
        else:
            final = [(c, scores[c]) for c in pool[:top_k]]

        results: list[RetrievedChunk] = []
        for chunk_id, score in final:
            c = self.store.chunk(chunk_id)
            results.append(
                RetrievedChunk(
                    chunk_id=chunk_id,
                    parent_id=c["parent_id"],
                    content=c["content"],
                    metadata=c["metadata"],
                    score=round(float(score), 6),
                )
            )
        return results
