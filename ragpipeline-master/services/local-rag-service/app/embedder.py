"""ONNX MiniLM-L6-v2 embedder (384-dim), fully offline.

Matches rag_config.py: embed_model = sentence-transformers/all-MiniLM-L6-v2.
Uses the INT8-quantized ONNX export bundled under models/.
"""

from __future__ import annotations

import threading

import numpy as np
import onnxruntime as ort
from tokenizers import Tokenizer


class Embedder:
    def __init__(self, model_path: str, tokenizer_path: str, max_length: int = 256):
        self._tokenizer = Tokenizer.from_file(tokenizer_path)
        self._tokenizer.enable_padding()
        self._tokenizer.enable_truncation(max_length)
        opts = ort.SessionOptions()
        opts.intra_op_num_threads = 2
        self._session = ort.InferenceSession(
            model_path, sess_options=opts, providers=["CPUExecutionProvider"]
        )
        self._lock = threading.Lock()

    @property
    def dimension(self) -> int:
        return 384

    def embed(self, texts: list[str], batch_size: int = 16) -> np.ndarray:
        """Return L2-normalized embeddings, shape (n, 384)."""
        out: list[np.ndarray] = []
        for i in range(0, len(texts), batch_size):
            batch = texts[i : i + batch_size]
            enc = self._tokenizer.encode_batch(batch)
            ids = np.array([e.ids for e in enc], dtype=np.int64)
            mask = np.array([e.attention_mask for e in enc], dtype=np.int64)
            ttype = np.zeros_like(ids)
            with self._lock:
                vecs = self._session.run(
                    ["sentence_embedding"],
                    {"input_ids": ids, "attention_mask": mask, "token_type_ids": ttype},
                )[0]
            out.append(vecs)
        embs = np.vstack(out).astype(np.float32)
        norms = np.linalg.norm(embs, axis=1, keepdims=True)
        norms[norms == 0] = 1.0
        return embs / norms

    def embed_one(self, text: str) -> np.ndarray:
        return self.embed([text])[0]
