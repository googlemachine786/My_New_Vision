"""
Incremental ingestion module for RAG pipeline.
Provides change detection and versioned ingestion.
"""

from .tracker import (
    DocumentTracker,
    IncrementalIngester,
    DocumentVersion,
    ChangeDetection,
)

__all__ = [
    "DocumentTracker",
    "IncrementalIngester",
    "DocumentVersion",
    "ChangeDetection",
]
