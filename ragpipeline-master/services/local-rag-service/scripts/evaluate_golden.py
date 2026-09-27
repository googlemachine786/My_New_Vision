"""Evaluate the local RAG service against data/golden_qa_dataset.jsonl.

Metrics:
  - retrieval hit rate: does any retrieved chunk contain key answer terms?
  - answer keyword overlap: fraction of golden-answer content words present
    in the generated answer.

Usage: python -m scripts.evaluate_golden [service_url]
"""

import json
import re
import sys
from pathlib import Path

import httpx

STOP = set(
    "the a an of in and or to is are was were it its this that by for with as "
    "on at from into through which their they them there these those be been "
    "has have had do does did can could will would should".split()
)


def content_words(text: str) -> set[str]:
    return {w for w in re.findall(r"[a-z]+", text.lower()) if w not in STOP and len(w) > 2}


def main() -> None:
    url = sys.argv[1] if len(sys.argv) > 1 else "http://localhost:8080"
    golden = (
        Path(__file__).resolve().parent.parent.parent.parent
        / "data"
        / "golden_qa_dataset.jsonl"
    )
    rows = [json.loads(l) for l in golden.read_text().splitlines() if l.strip()]

    hits = 0
    overlaps = []
    failures = []
    for r in rows:
        resp = httpx.post(
            f"{url}/api/v1/query",
            json={"query": r["question"], "grade": str(r["grade"]), "subject": r["subject"]},
            timeout=120,
        )
        resp.raise_for_status()
        data = resp.json()
        gold_terms = content_words(r["answer"])
        retrieved_text = " ".join(s["content"] for s in data.get("sources", []))
        answer_terms = content_words(data.get("answer", ""))
        retr_overlap = len(gold_terms & content_words(retrieved_text)) / max(len(gold_terms), 1)
        ans_overlap = len(gold_terms & answer_terms) / max(len(gold_terms), 1)
        hit = retr_overlap >= 0.5
        hits += hit
        overlaps.append(ans_overlap)
        status = "PASS" if hit else "MISS"
        if not hit:
            failures.append(r["question"])
        print(f"[{status}] retr={retr_overlap:.2f} ans={ans_overlap:.2f} | {r['question']}")

    n = len(rows)
    print(f"\nRetrieval hit rate : {hits}/{n} = {hits/n:.1%}")
    print(f"Answer overlap avg : {sum(overlaps)/n:.1%}")
    if failures:
        print("Misses:", *failures, sep="\n  - ")


if __name__ == "__main__":
    main()
