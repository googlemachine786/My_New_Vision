"""
Visionary RAG Pipeline - Chunker Module
Parent-Child Chunking for RAG Retrieval

Strategy:
- Tables: Atomic (parent_id == child_id, no splitting)
- Prose: Parent splitter (1500 chars, 100 overlap)
         Child splitter (512 chars, 77 overlap = 15%)
"""

from .parent_child import (
    create_parent_child_chunks,
    ParentChunk,
    ChildChunk,
    RecursiveCharacterTextSplitter,
)

__all__ = [
    "create_parent_child_chunks",
    "ParentChunk",
    "ChildChunk",
    "RecursiveCharacterTextSplitter",
]
