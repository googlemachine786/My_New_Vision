"""Ingest all markdown documents from data/corpus/ into the local store.

Usage:
    python -m scripts.ingest_corpus            # ingest bundled corpus
    python -m scripts.ingest_corpus /path/dir  # ingest a custom directory
"""

import sys
import time
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from app.chunker import parse_markdown_document  # noqa: E402
from app.config import settings  # noqa: E402
from app.embedder import Embedder  # noqa: E402
from app.store import VectorStore  # noqa: E402


def main() -> None:
    corpus_dir = Path(
        sys.argv[1] if len(sys.argv) > 1
        else Path(__file__).resolve().parent.parent / "data" / "corpus"
    )
    files = sorted(list(corpus_dir.glob("*.md")) + list(corpus_dir.glob("*.txt")))
    if not files:
        print(f"No .md/.txt files found in {corpus_dir}")
        sys.exit(1)

    t0 = time.time()
    store = VectorStore(settings.db_path, settings.embed_dimension)
    embedder = Embedder(settings.embed_model_path, settings.embed_tokenizer_path)

    total_p = total_c = 0
    for f in files:
        parents, children = parse_markdown_document(
            f.stem, f.read_text(encoding="utf-8"), {},
            settings.chunk_size, settings.chunk_overlap,
        )
        if not children:
            continue
        embs = embedder.embed([c.content for c in children])
        store.upsert(parents, children, embs)
        total_p += len(parents)
        total_c += len(children)
        print(f"  {f.name}: {len(parents)} sections -> {len(children)} chunks")

    print(
        f"\nIngested {len(files)} docs, {total_p} parents, {total_c} chunks "
        f"in {time.time()-t0:.1f}s. Store now has {store.size} chunks at {settings.db_path}"
    )


if __name__ == "__main__":
    main()
