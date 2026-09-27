"""SQLite-backed vector + document store with in-memory search indices.

Local-mode stand-in for AlloyDB/Supabase pgvector. Persists chunks and
embeddings in SQLite; serves dense search from a numpy matrix and sparse
search from a BM25 index (rank-bm25), the same hybrid the Go
vector-search-service performs against pgvector + BM25.
"""

from __future__ import annotations

import json
import re
import sqlite3
import threading
from pathlib import Path

import numpy as np
from rank_bm25 import BM25Okapi

_TOKEN = re.compile(r"[a-z0-9]+")


def tokenize(text: str) -> list[str]:
    return _TOKEN.findall(text.lower())


class VectorStore:
    def __init__(self, db_path: str, dimension: int = 384):
        Path(db_path).parent.mkdir(parents=True, exist_ok=True)
        self._db_path = db_path
        self._dim = dimension
        self._lock = threading.Lock()
        self._init_db()
        # In-memory indices
        self._ids: list[str] = []
        self._matrix: np.ndarray = np.zeros((0, dimension), dtype=np.float32)
        self._bm25: BM25Okapi | None = None
        self._chunks: dict[str, dict] = {}
        self._parents: dict[str, dict] = {}
        self._load()

    def _conn(self) -> sqlite3.Connection:
        conn = sqlite3.connect(self._db_path)
        conn.execute("PRAGMA journal_mode=WAL")
        return conn

    def _init_db(self) -> None:
        with self._conn() as c:
            c.executescript(
                """
                CREATE TABLE IF NOT EXISTS parents (
                    parent_id TEXT PRIMARY KEY,
                    title     TEXT NOT NULL,
                    content   TEXT NOT NULL,
                    metadata  TEXT NOT NULL DEFAULT '{}'
                );
                CREATE TABLE IF NOT EXISTS chunks (
                    chunk_id  TEXT PRIMARY KEY,
                    parent_id TEXT NOT NULL REFERENCES parents(parent_id),
                    content   TEXT NOT NULL,
                    metadata  TEXT NOT NULL DEFAULT '{}',
                    embedding BLOB NOT NULL
                );
                CREATE TABLE IF NOT EXISTS feedback (
                    id          INTEGER PRIMARY KEY AUTOINCREMENT,
                    created_at  TEXT DEFAULT CURRENT_TIMESTAMP,
                    payload     TEXT NOT NULL
                );
                """
            )

    def _load(self) -> None:
        with self._conn() as c:
            rows = c.execute(
                "SELECT chunk_id, parent_id, content, metadata, embedding FROM chunks"
            ).fetchall()
            prows = c.execute(
                "SELECT parent_id, title, content, metadata FROM parents"
            ).fetchall()
        self._parents = {
            r[0]: {"title": r[1], "content": r[2], "metadata": json.loads(r[3])}
            for r in prows
        }
        self._ids = [r[0] for r in rows]
        self._chunks = {
            r[0]: {"parent_id": r[1], "content": r[2], "metadata": json.loads(r[3])}
            for r in rows
        }
        if rows:
            self._matrix = np.vstack(
                [np.frombuffer(r[4], dtype=np.float32) for r in rows]
            )
            self._bm25 = BM25Okapi([tokenize(r[2]) for r in rows])
        else:
            self._matrix = np.zeros((0, self._dim), dtype=np.float32)
            self._bm25 = None

    # ------------------------------------------------------------------ write

    def upsert(
        self, parents: list, children: list, embeddings: np.ndarray
    ) -> None:
        with self._lock, self._conn() as c:
            for p in parents:
                c.execute(
                    "INSERT OR REPLACE INTO parents VALUES (?,?,?,?)",
                    (p.parent_id, p.title, p.content, json.dumps(p.metadata)),
                )
            for child, emb in zip(children, embeddings):
                c.execute(
                    "INSERT OR REPLACE INTO chunks VALUES (?,?,?,?,?)",
                    (
                        child.chunk_id,
                        child.parent_id,
                        child.content,
                        json.dumps(child.metadata),
                        emb.astype(np.float32).tobytes(),
                    ),
                )
        with self._lock:
            self._load()

    def save_feedback(self, payload: dict) -> int:
        with self._lock, self._conn() as c:
            cur = c.execute(
                "INSERT INTO feedback (payload) VALUES (?)", (json.dumps(payload),)
            )
            return int(cur.lastrowid)

    # ------------------------------------------------------------------- read

    @property
    def size(self) -> int:
        return len(self._ids)

    def chunk(self, chunk_id: str) -> dict:
        return self._chunks[chunk_id]

    def parent(self, parent_id: str) -> dict | None:
        return self._parents.get(parent_id)

    def _filter_mask(self, grade: str | None, subject: str | None) -> np.ndarray | None:
        if not grade and not subject:
            return None
        mask = np.ones(len(self._ids), dtype=bool)
        for i, cid in enumerate(self._ids):
            md = self._chunks[cid]["metadata"]
            if grade and str(md.get("grade", "")) not in ("", str(grade)):
                mask[i] = False
            if subject and md.get("subject", "").lower() not in ("", subject.lower()):
                mask[i] = False
        return mask

    def dense_search(
        self, query_emb: np.ndarray, top_k: int,
        grade: str | None = None, subject: str | None = None,
    ) -> list[tuple[str, float]]:
        if self.size == 0:
            return []
        sims = self._matrix @ query_emb.astype(np.float32)
        mask = self._filter_mask(grade, subject)
        if mask is not None:
            sims = np.where(mask, sims, -1.0)
        order = np.argsort(-sims)[:top_k]
        return [(self._ids[i], float(sims[i])) for i in order if sims[i] > -1.0]

    def sparse_search(
        self, query: str, top_k: int,
        grade: str | None = None, subject: str | None = None,
    ) -> list[tuple[str, float]]:
        if self._bm25 is None:
            return []
        scores = np.asarray(self._bm25.get_scores(tokenize(query)), dtype=np.float32)
        mask = self._filter_mask(grade, subject)
        if mask is not None:
            scores = np.where(mask, scores, -1.0)
        order = np.argsort(-scores)[:top_k]
        return [
            (self._ids[i], float(scores[i])) for i in order if scores[i] > 0
        ]

    def stats(self) -> dict:
        grades: dict[str, int] = {}
        chapters: set[str] = set()
        for c in self._chunks.values():
            g = str(c["metadata"].get("grade", "?"))
            grades[g] = grades.get(g, 0) + 1
            chapters.add(c["metadata"].get("chapter", "?"))
        return {
            "total_chunks": self.size,
            "total_parents": len(self._parents),
            "chunks_by_grade": grades,
            "chapters": sorted(chapters),
        }
