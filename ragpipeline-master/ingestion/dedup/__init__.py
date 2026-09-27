"""
Deduplication module for RAG pipeline.
Provides hybrid deduplication: exact hash, semantic, and MinHash LSH.
"""

from .detector import (
    DeduplicationDetector,
    ContentHasher,
    SemanticDeduplicator,
    MinHashLSH,
    Chunk,
)

__all__ = [
    "DeduplicationDetector",
    "ContentHasher",
    "SemanticDeduplicator",
    "MinHashLSH",
    "Chunk",
]
