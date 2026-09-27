"""Cross-encoder reranker using the repo-bundled ms-marco-MiniLM-L-12-v2 ONNX model.

Same model the reranker-service (FlashRank) uses; loaded directly with
onnxruntime so local mode needs no sidecar.
"""

from __future__ import annotations

import threading

import numpy as np
import onnxruntime as ort
from tokenizers import Tokenizer


class Reranker:
    def __init__(self, model_path: str, tokenizer_path: str, max_length: int = 512):
        self._tokenizer = Tokenizer.from_file(tokenizer_path)
        self._tokenizer.enable_padding()
        self._tokenizer.enable_truncation(max_length)
        opts = ort.SessionOptions()
        opts.intra_op_num_threads = 2
        self._session = ort.InferenceSession(
            model_path, sess_options=opts, providers=["CPUExecutionProvider"]
        )
        self._input_names = {i.name for i in self._session.get_inputs()}
        self._lock = threading.Lock()

    def score(self, query: str, passages: list[str], batch_size: int = 8) -> np.ndarray:
        """Return relevance scores (sigmoid logits) for each (query, passage) pair."""
        scores: list[float] = []
        for i in range(0, len(passages), batch_size):
            batch = passages[i : i + batch_size]
            enc = self._tokenizer.encode_batch([(query, p) for p in batch])
            ids = np.array([e.ids for e in enc], dtype=np.int64)
            mask = np.array([e.attention_mask for e in enc], dtype=np.int64)
            feed = {"input_ids": ids, "attention_mask": mask}
            if "token_type_ids" in self._input_names:
                feed["token_type_ids"] = np.array(
                    [e.type_ids for e in enc], dtype=np.int64
                )
            with self._lock:
                logits = self._session.run(None, feed)[0]
            logits = np.asarray(logits, dtype=np.float32).reshape(len(batch), -1)[:, 0]
            scores.extend((1.0 / (1.0 + np.exp(-logits))).tolist())
        return np.asarray(scores, dtype=np.float32)

    def rerank(
        self, query: str, passages: list[str], top_k: int
    ) -> list[tuple[int, float]]:
        """Return [(original_index, score)] sorted by score desc, truncated to top_k."""
        if not passages:
            return []
        s = self.score(query, passages)
        order = np.argsort(-s)[:top_k]
        return [(int(i), float(s[i])) for i in order]
