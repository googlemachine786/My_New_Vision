#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Visionary RAG - Full Pipeline Benchmark & Iterative Improvement
================================================================
Runs real queries through the Ollama pipeline, evaluates quality
on 5 dimensions, iterates configs, and prints a ranked leaderboard.

Usage:  python bench_and_improve.py
"""
import io
import sys
# Force UTF-8 output so emoji/unicode doesn't crash on Windows cp1252
if hasattr(sys.stdout, 'reconfigure'):
    sys.stdout.reconfigure(encoding='utf-8', errors='replace')
if hasattr(sys.stderr, 'reconfigure'):
    sys.stderr.reconfigure(encoding='utf-8', errors='replace')

import json
import time
import math
import requests
import fitz
import numpy as np
from pathlib import Path
from datetime import datetime
from typing import List, Dict, Tuple

# ──────────────────────────────────────────────────────────────
# CONFIG - OPTIMIZED based on grid search evaluation
# ──────────────────────────────────────────────────────────────
# See: visionary_rag_v5_grand_table.csv for full evaluation data
# Optimal configuration achieves 0.8386 composite score

OLLAMA_URL    = "http://localhost:11434"
EMBED_MODEL   = "sentence-transformers/all-MiniLM-L6-v2"  # OPTIMIZED: minilm (11% better)
LLM_MODEL     = "llama3.2:3b"
PDF_PATH      = Path("science class 8.pdf")
RESULTS_DIR   = Path("data")
CHUNK_SAMPLE  = 60
PAGE_STEP     = 5

# OPTIMIZED retrieval params (top configuration from grid search)
TOP_K         = 5           # Optimal: 5 (not 3 or 7+)
CHUNK_SIZE    = 400         # Optimal: 400 chars (not 500-1500)
CHUNK_OVERLAP = 150         # Optimal: 150 chars = 37.5% (not 77 = 15%)

# ──────────────────────────────────────────────────────────────
# GOLDEN QA DATASET  (10 questions, expected keywords per topic)
# ──────────────────────────────────────────────────────────────
GOLDEN_QA = [
    {
        "id": "Q01",
        "query": "What is photosynthesis?",
        "keywords": ["photosynthesis", "chlorophyll", "glucose", "sunlight", "oxygen", "carbon dioxide"],
        "chapter": "Crop Production and Management",
    },
    {
        "id": "Q02",
        "query": "What are the parts of a cell?",
        "keywords": ["nucleus", "cytoplasm", "cell membrane", "mitochondria", "cell wall"],
        "chapter": "Cell Structure and Functions",
    },
    {
        "id": "Q03",
        "query": "What is force and what are its effects?",
        "keywords": ["push", "pull", "motion", "shape", "newton", "magnitude"],
        "chapter": "Force and Pressure",
    },
    {
        "id": "Q04",
        "query": "What is combustion?",
        "keywords": ["combustion", "oxygen", "heat", "flame", "ignition", "fuel"],
        "chapter": "Combustion and Flame",
    },
    {
        "id": "Q05",
        "query": "What are microorganisms and where are they found?",
        "keywords": ["bacteria", "virus", "fungi", "protozoa", "algae", "microscopic"],
        "chapter": "Microorganisms",
    },
    {
        "id": "Q06",
        "query": "What is friction and how does it affect motion?",
        "keywords": ["friction", "surface", "force", "motion", "rolling", "sliding"],
        "chapter": "Friction",
    },
    {
        "id": "Q07",
        "query": "How do humans pollute air and what are the effects?",
        "keywords": ["pollution", "air", "fossil fuels", "smoke", "respiratory", "greenhouse"],
        "chapter": "Air and Water Pollution",
    },
    {
        "id": "Q08",
        "query": "What is sound and how is it produced?",
        "keywords": ["sound", "vibration", "medium", "wave", "frequency", "amplitude"],
        "chapter": "Sound",
    },
    {
        "id": "Q09",
        "query": "What is light and how does it reflect?",
        "keywords": ["light", "reflection", "incidence", "mirror", "angle", "refraction"],
        "chapter": "Light",
    },
    {
        "id": "Q10",
        "query": "What are the properties of metals and non-metals?",
        "keywords": ["metal", "non-metal", "conductor", "malleable", "ductile", "lustrous"],
        "chapter": "Materials: Metals and Non-metals",
    },
]

# ──────────────────────────────────────────────────────────────
# CONFIGS TO BENCHMARK - OPTIMIZED based on grid search
# ──────────────────────────────────────────────────────────────
# Baseline now uses optimal params from grid search (0.8386 composite)
# Compare variants to find best trade-off for your use case

CONFIGS = [
    {
        "name": "optimized-baseline",  # Top performer from grid search
        "llm": "llama3.2:3b",
        "top_k": 5,                    # Optimal: 5
        "chunk_size": 400,             # Optimal: 400 chars
        "chunk_overlap": 150,          # Optimal: 37.5%
        "prompt_style": "basic",
    },
    {
        "name": "low-latency",         # Faster inference (0.8329 @ 305ms)
        "llm": "llama3.2:3b",
        "top_k": 8,
        "chunk_size": 400,
        "chunk_overlap": 200,
        "prompt_style": "basic",
    },
    {
        "name": "high-quality",        # Maximum quality (0.8374 @ 705ms)
        "llm": "llama3.2:3b",
        "top_k": 7,
        "chunk_size": 400,
        "chunk_overlap": 150,
        "prompt_style": "tutor",
    },
]


# ══════════════════════════════════════════════════════════════
# HELPERS
# ══════════════════════════════════════════════════════════════

def cosine_sim(a: List[float], b: List[float]) -> float:
    na, nb = np.linalg.norm(a), np.linalg.norm(b)
    if na == 0 or nb == 0:
        return 0.0
    return float(np.dot(a, b) / (na * nb))


def get_embedding(text: str) -> List[float]:
    r = requests.post(
        f"{OLLAMA_URL}/api/embeddings",
        json={"model": EMBED_MODEL, "prompt": text},
        timeout=45,
    )
    r.raise_for_status()
    return r.json().get("embedding", [])


def generate(prompt: str, model: str, stream: bool = False) -> str:
    r = requests.post(
        f"{OLLAMA_URL}/api/generate",
        json={"model": model, "prompt": prompt, "stream": stream},
        timeout=120,
    )
    r.raise_for_status()
    return r.json().get("response", "")


def build_prompt(query: str, context: str, style: str) -> str:
    if style == "tutor":
        return f"""You are an expert CBSE Class 8 Science tutor.
Answer the student's question using ONLY the provided textbook context.
Be clear, accurate, and educational. Cite the page numbers.
If information is not in the context, say "This topic is outside the provided context."

TEXTBOOK CONTEXT:
{context}

STUDENT QUESTION: {query}

TUTOR ANSWER:"""
    else:  # basic
        return f"""Answer using ONLY this context. Cite page numbers.

Context:
{context}

Question: {query}

Answer:"""


# ══════════════════════════════════════════════════════════════
# PIPELINE CLASS
# ══════════════════════════════════════════════════════════════

class RAGPipeline:
    """Builds an in-memory index from the PDF, then answers queries."""

    def __init__(self, chunk_size: int = 500, page_step: int = PAGE_STEP):
        self.chunk_size = chunk_size
        self.page_step  = page_step
        self.chunks: List[Dict] = []
        self.embeddings: List[List[float]] = []

    def build_index(self) -> int:
        """Extract chunks from PDF and embed them."""
        if not PDF_PATH.exists():
            print(f"❌ PDF not found at: {PDF_PATH}")
            sys.exit(1)

        doc = fitz.open(PDF_PATH)
        total = len(doc)
        raw: List[Dict] = []

        for pnum in range(0, total, self.page_step):
            text = doc[pnum].get_text()
            if len(text.strip()) < 100:
                continue
            for i in range(0, len(text), self.chunk_size):
                chunk = text[i:i + self.chunk_size].strip()
                if len(chunk) > 60:
                    raw.append({"content": chunk, "page": pnum + 1})

        doc.close()

        sample = raw[:CHUNK_SAMPLE]
        self.chunks, self.embeddings = [], []

        for i, c in enumerate(sample):
            try:
                emb = get_embedding(c["content"])
                self.chunks.append(c)
                self.embeddings.append(emb)
            except Exception as e:
                pass
            if (i + 1) % 20 == 0:
                print(f"    → embedded {i+1}/{len(sample)} chunks")

        return len(self.chunks)

    def retrieve(self, query: str, top_k: int) -> List[Dict]:
        q_emb = get_embedding(query)
        scored = [(cosine_sim(q_emb, e), self.chunks[i])
                  for i, e in enumerate(self.embeddings)]
        scored.sort(key=lambda x: x[0], reverse=True)
        return [{"chunk": c, "score": s} for s, c in scored[:top_k]]

    def answer(self, query: str, top_k: int, model: str, style: str) -> Dict:
        t0 = time.time()
        hits = self.retrieve(query, top_k)
        retrieval_ms = (time.time() - t0) * 1000

        context = "\n\n".join(
            f"[Page {h['chunk']['page']}] {h['chunk']['content']}"
            for h in hits
        )
        prompt = build_prompt(query, context, style)

        t1 = time.time()
        answer_text = generate(prompt, model)
        gen_ms = (time.time() - t1) * 1000

        return {
            "answer": answer_text,
            "hits": hits,
            "retrieval_ms": retrieval_ms,
            "gen_ms": gen_ms,
            "total_ms": retrieval_ms + gen_ms,
        }


# ══════════════════════════════════════════════════════════════
# METRICS
# ══════════════════════════════════════════════════════════════

def keyword_coverage(answer: str, keywords: List[str]) -> float:
    """Fraction of expected topic keywords found in the answer."""
    a = answer.lower()
    return sum(1 for kw in keywords if kw.lower() in a) / max(len(keywords), 1)


def faithfulness_score(answer: str, context: str) -> float:
    """
    Approximate faithfulness: fraction of answer sentences that
    share at least one content word with the retrieved context.
    (proxy for full LLM-judge faithfulness without an extra LLM call)
    """
    ctx_words = set(context.lower().split())
    sentences = [s.strip() for s in answer.split(".") if len(s.strip()) > 10]
    if not sentences:
        return 0.0
    grounded = 0
    for sent in sentences:
        sent_words = set(sent.lower().split()) - {"the", "a", "is", "in", "of", "to", "and", "it"}
        if sent_words & ctx_words:
            grounded += 1
    return grounded / len(sentences)


def not_found_penalty(answer: str) -> float:
    """Return 0.0 if model admitted it couldn't find answer, else 1.0."""
    phrases = ["not found in context", "outside the provided context",
               "not in context", "cannot find", "no information"]
    low = answer.lower()
    return 0.0 if any(p in low for p in phrases) else 1.0


def dcg(relevance: List[float]) -> float:
    """Discounted Cumulative Gain."""
    return sum(r / math.log2(i + 2) for i, r in enumerate(relevance))


def ndcg_at_k(hits: List[Dict], keywords: List[str], k: int = 5) -> float:
    """Approximate NDCG@K using keyword overlap as relevance signal."""
    rels = []
    for h in hits[:k]:
        text = h["chunk"]["content"].lower()
        rel = sum(1 for kw in keywords if kw.lower() in text) / max(len(keywords), 1)
        rels.append(rel)
    ideal = sorted(rels, reverse=True)
    d = dcg(rels)
    i = dcg(ideal)
    return d / i if i > 0 else 0.0


def recall_at_k(hits: List[Dict], keywords: List[str], k: int = 5) -> float:
    """Recall@K: at least one keyword found in top-K chunks."""
    found = set()
    for h in hits[:k]:
        text = h["chunk"]["content"].lower()
        for kw in keywords:
            if kw.lower() in text:
                found.add(kw)
    return len(found) / max(len(keywords), 1)


def compute_metrics(result: Dict, qa: Dict) -> Dict:
    answer  = result["answer"]
    context = "\n".join(h["chunk"]["content"] for h in result["hits"])

    cov  = keyword_coverage(answer, qa["keywords"])
    fai  = faithfulness_score(answer, context)
    nfp  = not_found_penalty(answer)
    ndk  = ndcg_at_k(result["hits"], qa["keywords"])
    rck  = recall_at_k(result["hits"], qa["keywords"])

    # Composite quality score (weights tuned for CBSE tutoring)
    quality = (
        cov  * 0.35 +   # answer covers expected concepts
        fai  * 0.25 +   # answer grounded in context
        nfp  * 0.10 +   # model didn't punt
        rck  * 0.20 +   # relevant chunks retrieved
        ndk  * 0.10     # ranking quality
    )

    return {
        "keyword_coverage": round(cov, 3),
        "faithfulness":     round(fai, 3),
        "not_found_ok":     round(nfp, 3),
        "recall_at_5":      round(rck, 3),
        "ndcg_at_5":        round(ndk, 3),
        "quality_score":    round(quality, 3),
        "total_ms":         round(result["total_ms"]),
    }


# ══════════════════════════════════════════════════════════════
# BENCHMARK RUNNER
# ══════════════════════════════════════════════════════════════

def run_config(cfg: Dict) -> Dict:
    """Run all golden queries for one config and return aggregated metrics."""
    print(f"\n{'='*65}")
    print(f"  CONFIG: {cfg['name'].upper()}")
    print(f"  LLM={cfg['llm']}  top_k={cfg['top_k']}  "
          f"chunk={cfg['chunk_size']}  prompt={cfg['prompt_style']}")
    print(f"{'='*65}")

    # Build (or reuse) index
    pipeline = RAGPipeline(chunk_size=cfg["chunk_size"])
    n = pipeline.build_index()
    print(f"  [OK] Index built: {n} chunks\n")

    all_metrics = []
    per_query   = []

    for qa in GOLDEN_QA:
        print(f"  [Q] [{qa['id']}] {qa['query']}")
        try:
            result  = pipeline.answer(qa["query"], cfg["top_k"], cfg["llm"], cfg["prompt_style"])
            metrics = compute_metrics(result, qa)

            per_query.append({"qa": qa["id"], "query": qa["query"],
                               "answer": result["answer"][:200], **metrics})
            all_metrics.append(metrics)

            print(f"     quality={metrics['quality_score']:.3f}  "
                  f"kw-cov={metrics['keyword_coverage']:.2f}  "
                  f"recall@5={metrics['recall_at_5']:.2f}  "
                  f"faith={metrics['faithfulness']:.2f}  "
                  f"ndcg@5={metrics['ndcg_at_5']:.2f}  "
                  f"{metrics['total_ms']}ms")
        except Exception as e:
            print(f"     [ERR] {e}")
            per_query.append({"qa": qa["id"], "query": qa["query"],
                               "quality_score": 0, "error": str(e)})

    # Aggregate
    def avg(key): return round(sum(m.get(key, 0) for m in all_metrics) / max(len(all_metrics), 1), 3)

    agg = {
        "config":           cfg["name"],
        "avg_quality":      avg("quality_score"),
        "avg_kw_coverage":  avg("keyword_coverage"),
        "avg_faithfulness": avg("faithfulness"),
        "avg_recall_at_5":  avg("recall_at_5"),
        "avg_ndcg_at_5":    avg("ndcg_at_5"),
        "avg_latency_ms":   avg("total_ms"),
        "cfg_detail":       cfg,
        "per_query":        per_query,
        "timestamp":        datetime.now().isoformat(),
    }

    print(f"\n  [AVG] quality={agg['avg_quality']}  "
          f"kw={agg['avg_kw_coverage']}  "
          f"recall={agg['avg_recall_at_5']}  "
          f"faith={agg['avg_faithfulness']}  "
          f"ndcg={agg['avg_ndcg_at_5']}  "
          f"lat={agg['avg_latency_ms']}ms")

    return agg


def check_ollama() -> bool:
    global EMBED_MODEL, LLM_MODEL
    try:
        r = requests.get(f"{OLLAMA_URL}/api/tags", timeout=6)
        models = [m["name"] for m in r.json().get("models", [])]
        print(f"[OK] Ollama online | models: {models}")
        # Use prefix matching so 'nomic-embed-text:latest' matches 'nomic-embed-text'
        def has_model(prefix):
            return any(m == prefix or m.startswith(prefix + ":") for m in models)
        missing = [m for m in [EMBED_MODEL, LLM_MODEL] if not has_model(m)]
        if missing:
            print(f"[WARN] Missing models: {missing}")
            print(f"   Run: ollama pull {' '.join(missing)}")
            return False
        # Resolve actual tagged model names
        for m in models:
            if m == EMBED_MODEL or m.startswith(EMBED_MODEL + ":"):
                EMBED_MODEL = m
            if m == LLM_MODEL or m.startswith(LLM_MODEL + ":"):
                LLM_MODEL = m
        print(f"[OK] Using embed={EMBED_MODEL}  llm={LLM_MODEL}")
        return True
    except Exception as e:
        print(f"[ERR] Ollama not reachable: {e}")
        print(f"      Make sure Ollama is running: ollama serve")
        return False


def print_leaderboard(results: List[Dict]):
    ranked = sorted(results, key=lambda x: x["avg_quality"], reverse=True)

    print("\n" + "="*90)
    print("  === BENCHMARK LEADERBOARD ===")
    print("="*90)
    hdr = f"{'Rank':<6}{'Config':<22}{'Quality':>9}{'KW-Cov':>9}{'Recall@5':>10}{'Faith':>8}{'NDCG@5':>8}{'Lat(ms)':>9}"
    print(hdr)
    print("-"*90)

    medals = ["[1st]", "[2nd]", "[3rd]"]
    for i, r in enumerate(ranked):
        rank_label = medals[i] if i < 3 else f"  {i+1}. "
        print(f" {rank_label:<5} {r['config']:<22} "
              f"{r['avg_quality']:>8.3f} "
              f"{r['avg_kw_coverage']:>9.3f} "
              f"{r['avg_recall_at_5']:>9.3f} "
              f"{r['avg_faithfulness']:>8.3f} "
              f"{r['avg_ndcg_at_5']:>8.3f} "
              f"{r['avg_latency_ms']:>8.0f}")

    print("="*90)
    best = ranked[0]
    print(f"\n  [WINNER] {best['config']} -- Quality={best['avg_quality']}")
    best_cfg = best["cfg_detail"]
    print(f"     LLM={best_cfg['llm']}  top_k={best_cfg['top_k']}  "
          f"chunk_size={best_cfg['chunk_size']}  prompt={best_cfg['prompt_style']}")
    print()


def save_results(results: List[Dict]):
    RESULTS_DIR.mkdir(exist_ok=True)
    out_path = RESULTS_DIR / "benchmark_results.json"
    with open(out_path, "w", encoding="utf-8") as f:
        json.dump(results, f, indent=2)
    print(f"  [SAVED] Full results -> {out_path}")


# ══════════════════════════════════════════════════════════════
# MAIN
# ══════════════════════════════════════════════════════════════

def main():
    print("\n" + "="*65)
    print("  VISIONARY RAG -- FULL BENCHMARK & ITERATIVE IMPROVEMENT")
    print("="*65)
    print(f"  Date  : {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}")
    print(f"  PDF   : {PDF_PATH}")
    print(f"  Evals : {len(GOLDEN_QA)} golden queries")
    print(f"  Runs  : {len(CONFIGS)} configurations")
    print("="*65 + "\n")

    if not check_ollama():
        sys.exit(1)

    all_results = []
    t_start = time.time()

    for cfg in CONFIGS:
        agg = run_config(cfg)
        all_results.append(agg)

    elapsed = time.time() - t_start
    print(f"\n  Total wall-clock: {elapsed/60:.1f} min")

    print_leaderboard(all_results)
    save_results(all_results)

    print("  Done!\n")


if __name__ == "__main__":
    main()
