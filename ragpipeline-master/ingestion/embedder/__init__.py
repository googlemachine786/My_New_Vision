"""
Visionary RAG Pipeline - Embedder Module
Vertex AI Batch Embedding for Dense Retrieval
"""

from .vertex_batch import embed_batch, VertexAIEmbedder, EmbeddingConfig

__all__ = [
    "embed_batch",
    "VertexAIEmbedder",
    "EmbeddingConfig",
]
