"""
Visionary RAG Pipeline - Keywords Module
YAKE Keyword Extraction for Sparse Retrieval
"""

from .yake_extractor import extract_keywords, YakeKeywordExtractor

__all__ = [
    "extract_keywords",
    "YakeKeywordExtractor",
]
