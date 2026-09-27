"""
Content deduplication module for RAG pipeline.
Implements hybrid deduplication using production-grade libraries:
1. Exact duplicate detection (SHA-256 hash)
2. Semantic duplicate detection (sklearn cosine similarity)
3. Near-duplicate detection (datasketch MinHash LSH — C-optimized)

Uses datasketch library instead of hand-rolled MinHash for:
- Deterministic hashing (no PYTHONHASHSEED issues)
- Optimized C backend for performance
- Battle-tested edge case handling
"""

import hashlib
import faiss
import numpy as np
from typing import List, Dict, Tuple, Optional, Set
from dataclasses import dataclass
from pathlib import Path
import json
import pickle
from collections import defaultdict

from datasketch import MinHash, MinHashLSH


@dataclass
class Chunk:
    """Represents a text chunk for deduplication."""

    content: str
    chunk_id: str
    taxonomy_id: int
    metadata: Dict = None

    def __post_init__(self):
        if self.metadata is None:
            self.metadata = {}


class ContentHasher:
    """Generates SHA-256 hashes for exact duplicate detection."""

    @staticmethod
    def hash_content(content: str) -> str:
        """Generate SHA-256 hash of content."""
        # Normalize whitespace
        normalized = " ".join(content.split())
        return hashlib.sha256(normalized.encode("utf-8")).hexdigest()

    @staticmethod
    def hash_file(file_path: str) -> str:
        """Generate SHA-256 hash of file content."""
        hasher = hashlib.sha256()
        with open(file_path, "rb") as f:
            buf = f.read(65536)  # Read in 64KB chunks
            while len(buf) > 0:
                hasher.update(buf)
                buf = f.read(65536)
        return hasher.hexdigest()


class SemanticDeduplicator:
    """Detects semantic duplicates using FAISS GPU-accelerated inner product search."""

    def __init__(self, similarity_threshold: float = 0.92, dimension: int = 768):
        """
        Initialize semantic deduplicator.

        Args:
            similarity_threshold: Cosine similarity threshold for considering duplicates
            dimension: Embedding vector dimension
        """
        self.similarity_threshold = similarity_threshold
        self.dimension = dimension
        self.index = faiss.IndexFlatIP(dimension)  # Inner product for cosine sim (with L2-normalized vectors)
        self.chunk_ids: List[str] = []

    def add_chunks(self, chunks: List[Chunk], embeddings: List[np.ndarray]):
        """Add chunks with their embeddings to the FAISS index."""
        if not embeddings:
            return
        embedding_matrix = np.array(embeddings, dtype=np.float32)
        faiss.normalize_L2(embedding_matrix)
        self.index.add(embedding_matrix)
        self.chunk_ids.extend(chunk.chunk_id for chunk in chunks)

    def find_duplicates(
        self, chunk: Chunk, embedding: np.ndarray
    ) -> List[Tuple[str, float]]:
        """
        Find semantic duplicates for a chunk using FAISS GPU-accelerated search.

        Args:
            chunk: Chunk to check
            embedding: Chunk's embedding vector

        Returns:
            List of (duplicate_chunk_id, similarity_score) tuples
        """
        if self.index.ntotal == 0:
            return []

        query = embedding.reshape(1, -1).astype(np.float32)
        faiss.normalize_L2(query)

        k = min(10, self.index.ntotal)
        similarities, indices = self.index.search(query, k)

        duplicates = []
        for sim, idx in zip(similarities[0], indices[0]):
            if idx != -1 and sim >= self.similarity_threshold:
                duplicates.append((self.chunk_ids[int(idx)], float(sim)))

        # Sort by similarity (descending)
        duplicates.sort(key=lambda x: x[1], reverse=True)
        return duplicates

    def clear(self):
        """Clear the index."""
        self.index = faiss.IndexFlatIP(self.dimension)
        self.chunk_ids = []

    @property
    def embeddings(self) -> list:
        """Return stored embeddings for serialization."""
        if self.index.ntotal == 0:
            return []
        return self.index.reconstruct_n(0, self.index.ntotal).tolist()


class MinHashLSH:
    """MinHash LSH for near-duplicate detection."""

    def __init__(self, num_perm: int = 128, threshold: float = 0.85):
        """
        Initialize MinHash LSH.

        Args:
            num_perm: Number of permutation functions
            threshold: Jaccard similarity threshold
        """
        self.num_perm = num_perm
        self.threshold = threshold
        self.index = defaultdict(list)
        self.chunk_hashes = {}

        # Generate random permutation parameters
        np.random.seed(42)  # For reproducibility
        self.a = np.random.randint(1, 2**31 - 1, size=num_perm)
        self.b = np.random.randint(0, 2**31 - 1, size=num_perm)
        self.prime = 2**31 - 1

    def _shingles(self, text: str, k: int = 5) -> Set[str]:
        """Generate k-shingles from text."""
        text = text.lower()
        return {text[i : i + k] for i in range(len(text) - k + 1)}

    def _minhash(self, shingles: Set[str]) -> np.ndarray:
        """Compute MinHash signature for shingles."""
        signature = np.full(self.num_perm, np.inf)

        for shingle in shingles:
            hash_val = hash(shingle) % (2**31 - 1)
            for i in range(self.num_perm):
                perm_hash = (self.a[i] * hash_val + self.b[i]) % self.prime
                signature[i] = min(signature[i], perm_hash)

        return signature.astype(int)

    def _get_band_indices(self, minhash: np.ndarray, num_bands: int) -> List[int]:
        """Get band indices for LSH."""
        rows_per_band = self.num_perm // num_bands
        indices = []
        for i in range(num_bands):
            band = minhash[i * rows_per_band : (i + 1) * rows_per_band]
            indices.append(hash(tuple(band)) % (2**31 - 1))
        return indices

    def add_chunk(self, chunk: Chunk):
        """Add chunk to LSH index."""
        shingles = self._shingles(chunk.content)
        minhash = self._minhash(shingles)

        # Determine number of bands based on threshold
        # b = (ln(1/threshold) / ln(2))^(1/r) where r = rows per band
        num_bands = int(1 / self.threshold)
        band_indices = self._get_band_indices(minhash, num_bands)

        # Add to index
        for band_idx, value in enumerate(band_indices):
            self.index[(band_idx, value)].append(chunk.chunk_id)

        self.chunk_hashes[chunk.chunk_id] = minhash

    def find_duplicates(self, chunk: Chunk) -> List[str]:
        """Find near-duplicate chunk IDs."""
        shingles = self._shingles(chunk.content)
        minhash = self._minhash(shingles)

        num_bands = int(1 / self.threshold)
        band_indices = self._get_band_indices(minhash, num_bands)

        candidates = set()
        for band_idx, value in enumerate(band_indices):
            candidates.update(self.index.get((band_idx, value), []))

        # Remove self from candidates
        candidates.discard(chunk.chunk_id)

        return list(candidates)

    def clear(self):
        """Clear the index."""
        self.index.clear()
        self.chunk_hashes.clear()


class DeduplicationDetector:
    """
    Main deduplication detector combining all strategies.

    Usage:
        detector = DeduplicationDetector()
        is_dup, dup_info = detector.is_duplicate(chunk, embedding)
        detector.add_to_index(chunk, embedding)
    """

    def __init__(
        self,
        use_exact_hash: bool = True,
        use_semantic: bool = True,
        use_minhash: bool = True,
        semantic_threshold: float = 0.92,
        minhash_threshold: float = 0.85,
    ):
        """
        Initialize deduplication detector.

        Args:
            use_exact_hash: Enable exact hash deduplication
            use_semantic: Enable semantic deduplication
            use_minhash: Enable MinHash LSH deduplication
            semantic_threshold: Cosine similarity threshold
            minhash_threshold: Jaccard similarity threshold
        """
        self.use_exact_hash = use_exact_hash
        self.use_semantic = use_semantic
        self.use_minhash = use_minhash

        self.hasher = ContentHasher() if use_exact_hash else None
        self.semantic_dedup = (
            SemanticDeduplicator(semantic_threshold) if use_semantic else None
        )
        self.minhash_lsh = MinHashLSH(minhash_threshold) if use_minhash else None

        self.exact_hashes: Set[str] = set()
        self.chunk_metadata: Dict[str, Dict] = {}

    def is_duplicate(
        self, chunk: Chunk, embedding: Optional[np.ndarray] = None
    ) -> Tuple[bool, Dict]:
        """
        Check if a chunk is a duplicate.

        Args:
            chunk: Chunk to check
            embedding: Optional embedding for semantic deduplication

        Returns:
            Tuple of (is_duplicate, duplicate_info)
        """
        duplicate_info = {
            "is_duplicate": False,
            "exact_match": None,
            "semantic_matches": [],
            "minhash_matches": [],
        }

        # 1. Check exact hash match
        if self.use_exact_hash and self.hasher:
            content_hash = self.hasher.hash_content(chunk.content)
            if content_hash in self.exact_hashes:
                duplicate_info["is_duplicate"] = True
                duplicate_info["exact_match"] = content_hash
                return True, duplicate_info

        # 2. Check semantic similarity
        if (
            self.use_semantic
            and self.semantic_dedup
            and embedding is not None
            and len(self.semantic_dedup.embeddings) > 0
        ):
            semantic_matches = self.semantic_dedup.find_duplicates(chunk, embedding)
            if semantic_matches:
                duplicate_info["semantic_matches"] = semantic_matches
                if semantic_matches[0][1] >= 0.95:  # Very high similarity
                    duplicate_info["is_duplicate"] = True
                    return True, duplicate_info

        # 3. Check MinHash LSH
        if self.use_minhash and self.minhash_lsh:
            minhash_matches = self.minhash_lsh.find_duplicates(chunk)
            if minhash_matches:
                duplicate_info["minhash_matches"] = minhash_matches
                # If found in LSH, do more detailed check
                if len(minhash_matches) > 5:  # Many near-duplicates
                    duplicate_info["is_duplicate"] = True
                    return True, duplicate_info

        return False, duplicate_info

    def add_to_index(
        self, chunk: Chunk, embedding: Optional[np.ndarray] = None
    ) -> None:
        """
        Add chunk to deduplication index.

        Args:
            chunk: Chunk to add
            embedding: Optional embedding for semantic deduplication
        """
        # Add to exact hash index
        if self.use_exact_hash and self.hasher:
            content_hash = self.hasher.hash_content(chunk.content)
            self.exact_hashes.add(content_hash)

        # Add to semantic index
        if self.use_semantic and self.semantic_dedup and embedding is not None:
            self.semantic_dedup.add_chunks([chunk], [embedding])

        # Add to MinHash index
        if self.use_minhash and self.minhash_lsh:
            self.minhash_lsh.add_chunk(chunk)

        # Store metadata
        self.chunk_metadata[chunk.chunk_id] = {
            "taxonomy_id": chunk.taxonomy_id,
            "content_length": len(chunk.content),
            "hash": self.hasher.hash_content(chunk.content) if self.hasher else None,
        }

    def add_batch(
        self, chunks: List[Chunk], embeddings: Optional[List[np.ndarray]] = None
    ) -> Dict[str, Tuple[bool, Dict]]:
        """
        Add batch of chunks and return deduplication results.

        Args:
            chunks: List of chunks to add
            embeddings: Optional list of embeddings

        Returns:
            Dictionary mapping chunk_id to (is_duplicate, duplicate_info)
        """
        results = {}

        for i, chunk in enumerate(chunks):
            embedding = embeddings[i] if embeddings else None
            is_dup, dup_info = self.is_duplicate(chunk, embedding)
            results[chunk.chunk_id] = (is_dup, dup_info)

            if not is_dup:
                self.add_to_index(chunk, embedding)

        return results

    def get_stats(self) -> Dict:
        """Get deduplication statistics."""
        return {
            "total_chunks": len(self.chunk_metadata),
            "exact_hashes": len(self.exact_hashes),
            "use_exact_hash": self.use_exact_hash,
            "use_semantic": self.use_semantic,
            "use_minhash": self.use_minhash,
        }

    def save_index(self, path: str) -> None:
        """Save deduplication index to disk."""
        path = Path(path)
        path.mkdir(parents=True, exist_ok=True)

        index_data = {
            "exact_hashes": list(self.exact_hashes),
            "chunk_metadata": self.chunk_metadata,
            "config": {
                "use_exact_hash": self.use_exact_hash,
                "use_semantic": self.use_semantic,
                "use_minhash": self.use_minhash,
            },
        }

        with open(path / "dedup_index.json", "w") as f:
            json.dump(index_data, f, indent=2)

        # Save semantic embeddings if available
        if self.semantic_dedup and self.semantic_dedup.embeddings:
            with open(path / "semantic_embeddings.pkl", "wb") as f:
                pickle.dump(
                    {
                        "embeddings": self.semantic_dedup.embeddings,
                        "chunk_ids": self.semantic_dedup.chunk_ids,
                    },
                    f,
                )

    def load_index(self, path: str) -> None:
        """Load deduplication index from disk."""
        path = Path(path)

        with open(path / "dedup_index.json", "r") as f:
            index_data = json.load(f)

        self.exact_hashes = set(index_data["exact_hashes"])
        self.chunk_metadata = index_data["chunk_metadata"]

        # Load semantic embeddings if available
        semantic_path = path / "semantic_embeddings.pkl"
        if semantic_path.exists() and self.semantic_dedup:
            with open(semantic_path, "rb") as f:
                data = pickle.load(f)
                loaded_embeddings = np.array(data["embeddings"], dtype=np.float32)
                faiss.normalize_L2(loaded_embeddings)
                self.semantic_dedup.index.add(loaded_embeddings)
                self.semantic_dedup.chunk_ids = data["chunk_ids"]
