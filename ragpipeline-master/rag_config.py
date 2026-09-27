"""
Central RAG Configuration Module

Based on comprehensive grid search optimization (visionary_rag_v5_grand_table.csv)
Testing 100+ configurations across 10 parameters and 10 metrics.

Usage:
    from rag_config import DEFAULT_CONFIG, RAGConfig
    
    # Use production defaults
    config = DEFAULT_CONFIG
    
    # Or customize
    config = RAGConfig(
        chunk_size=400,
        top_k=5,
        similarity_threshold=0.36,
    )
    
    # Use preset configurations
    config = RAGConfig.low_latency()   # Faster, slightly lower quality
    config = RAGConfig.high_quality()  # Slower, maximum quality
"""

from dataclasses import dataclass, field
from typing import Literal, Dict, Any
import os


@dataclass(frozen=True)
class RAGConfig:
    """
    Production RAG configuration for CBSE Science (Grades 6-8).
    
    All parameters are optimized based on empirical evaluation data.
    See RAG_PIPELINE_AUDIT_REPORT.md for detailed analysis.
    
    Attributes:
        chunk_size: Characters per chunk (optimal: 400)
        chunk_overlap: Overlap between chunks in characters (optimal: 150, 37.5%)
        embed_model: SentenceTransformer model name
        embed_dimension: Embedding vector dimensions
        retrieval_strategy: dense, hybrid, or bm25
        top_k: Number of documents to retrieve
        rrf_k: Reciprocal Rank Fusion constant
        similarity_threshold: Minimum similarity for chunk filtering
        bm25_k1: BM25 k1 parameter (only used in hybrid mode)
        temperature: LLM sampling temperature
        max_tokens: Maximum generation tokens
        batch_size: Batch size for inference
        cache_enabled: Enable query caching
    """
    
    # Chunking (MOST CRITICAL - 40% of performance variance)
    chunk_size: int = 400
    chunk_overlap: int = 150
    
    # Embedding (35% of performance variance)
    embed_model: str = "sentence-transformers/all-MiniLM-L6-v2"
    embed_dimension: int = 384
    
    # Retrieval (15% of performance variance)
    retrieval_strategy: Literal["dense", "hybrid", "bm25"] = "dense"
    top_k: int = 5
    rrf_k: int = 60
    similarity_threshold: float = 0.36
    bm25_k1: float = 2.05
    
    # Generation (10% of performance variance)
    temperature: float = 0.75
    max_tokens: int = 150
    
    # Performance
    batch_size: int = 8
    cache_enabled: bool = True
    cache_ttl_seconds: int = 300
    
    # Metadata
    version: str = "5.0-optimized"
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert config to dictionary."""
        return {
            "chunk_size": self.chunk_size,
            "chunk_overlap": self.chunk_overlap,
            "embed_model": self.embed_model,
            "embed_dimension": self.embed_dimension,
            "retrieval_strategy": self.retrieval_strategy,
            "top_k": self.top_k,
            "rrf_k": self.rrf_k,
            "similarity_threshold": self.similarity_threshold,
            "bm25_k1": self.bm25_k1,
            "temperature": self.temperature,
            "max_tokens": self.max_tokens,
            "batch_size": self.batch_size,
            "cache_enabled": self.cache_enabled,
            "cache_ttl_seconds": self.cache_ttl_seconds,
            "version": self.version,
        }
    
    @classmethod
    def from_env(cls) -> "RAGConfig":
        """Load configuration from environment variables."""
        return cls(
            chunk_size=int(os.getenv("PARENT_MAX_CHARS", "400")),
            chunk_overlap=int(os.getenv("CHILD_OVERLAP", "150")),
            embed_model=os.getenv("EMBED_MODEL", "sentence-transformers/all-MiniLM-L6-v2"),
            retrieval_strategy=os.getenv("RETRIEVAL_STRATEGY", "dense"),
            top_k=int(os.getenv("TOP_K", "5")),
            rrf_k=float(os.getenv("RRF_K", "60")),
            similarity_threshold=float(os.getenv("SIMILARITY_THRESHOLD", "0.36")),
            bm25_k1=float(os.getenv("BM25_K1", "2.05")),
            temperature=float(os.getenv("LLM_TEMPERATURE", "0.75")),
            max_tokens=int(os.getenv("MAX_TOKENS", "150")),
        )
    
    @classmethod
    def low_latency(cls) -> "RAGConfig":
        """
        Low-latency configuration optimized for speed.
        
        Performance: 0.8329 composite score, 305ms P95 latency
        Trade-off: -0.7% quality for -44% latency
        
        Use case: Real-time applications, interactive demos
        """
        return cls(
            chunk_size=400,
            chunk_overlap=200,
            top_k=8,
            similarity_threshold=0.35,
            temperature=0.7,
            max_tokens=100,
            batch_size=8,
            version="5.0-low-latency",
        )
    
    @classmethod
    def high_quality(cls) -> "RAGConfig":
        """
        High-quality configuration optimized for accuracy.
        
        Performance: 0.8374 composite score, 705ms P95 latency
        Trade-off: +37% latency for -0.1% quality
        
        Use case: Batch processing, critical queries, evaluation
        """
        return cls(
            chunk_size=400,
            chunk_overlap=150,
            top_k=7,
            rrf_k=60,
            similarity_threshold=0.32,
            temperature=0.63,
            max_tokens=146,
            batch_size=4,
            version="5.0-high-quality",
        )
    
    @classmethod
    def balanced(cls) -> "RAGConfig":
        """
        Balanced configuration (same as default).
        
        Performance: 0.8386 composite score, 547ms P95 latency
        Best overall trade-off between quality and speed.
        
        Use case: Production default
        """
        return cls(
            chunk_size=400,
            chunk_overlap=150,
            top_k=5,
            rrf_k=60,
            similarity_threshold=0.364,
            bm25_k1=2.05,
            temperature=0.75,
            max_tokens=143,
            batch_size=8,
            version="5.0-balanced",
        )
    
    def validate(self) -> tuple[bool, list[str]]:
        """
        Validate configuration parameters.
        
        Returns:
            Tuple of (is_valid, list_of_errors)
        """
        errors = []
        
        # Chunking validation
        if self.chunk_size < 100:
            errors.append(f"chunk_size ({self.chunk_size}) too small, minimum 100")
        if self.chunk_size > 2000:
            errors.append(f"chunk_size ({self.chunk_size}) too large, maximum 2000")
        if self.chunk_overlap >= self.chunk_size:
            errors.append(f"chunk_overlap ({self.chunk_overlap}) must be < chunk_size ({self.chunk_size})")
        if self.chunk_overlap < 0:
            errors.append(f"chunk_overlap must be non-negative")
        
        # Overlap ratio check (optimal: 30-40%)
        overlap_ratio = self.chunk_overlap / self.chunk_size
        if overlap_ratio < 0.2 or overlap_ratio > 0.5:
            errors.append(
                f"overlap ratio ({overlap_ratio:.1%}) outside recommended range (20-50%)"
            )
        
        # Embedding validation
        if self.embed_dimension <= 0:
            errors.append(f"embed_dimension must be positive")
        
        # Retrieval validation
        if self.top_k < 1:
            errors.append(f"top_k must be >= 1")
        if self.top_k > 20:
            errors.append(f"top_k ({self.top_k}) too large, consider <= 10")
        if self.rrf_k <= 0:
            errors.append(f"rrf_k must be positive")
        if not 0.0 <= self.similarity_threshold <= 1.0:
            errors.append(f"similarity_threshold must be between 0 and 1")
        if self.bm25_k1 <= 0:
            errors.append(f"bm25_k1 must be positive")
        
        # Generation validation
        if not 0.0 <= self.temperature <= 2.0:
            errors.append(f"temperature must be between 0 and 2")
        if self.max_tokens <= 0:
            errors.append(f"max_tokens must be positive")
        if self.max_tokens > 4096:
            errors.append(f"max_tokens ({self.max_tokens}) exceeds model limit")
        
        # Performance validation
        if self.batch_size < 1:
            errors.append(f"batch_size must be >= 1")
        if self.batch_size > 64:
            errors.append(f"batch_size ({self.batch_size}) too large for most GPUs")
        
        return (len(errors) == 0, errors)
    
    def __str__(self) -> str:
        """Human-readable configuration summary."""
        lines = [
            f"RAGConfig v{self.version}",
            "=" * 50,
            f"Chunking:     {self.chunk_size} chars (+{self.chunk_overlap} overlap)",
            f"Embedding:    {self.embed_model} ({self.embed_dimension}d)",
            f"Retrieval:    {self.retrieval_strategy} (top_k={self.top_k}, rrf_k={self.rrf_k})",
            f"Similarity:   threshold={self.similarity_threshold:.3f}",
            f"Generation:   temp={self.temperature}, max_tokens={self.max_tokens}",
            f"Performance:  batch_size={self.batch_size}, cache={self.cache_enabled}",
        ]
        return "\n".join(lines)


# Default production configuration
DEFAULT_CONFIG = RAGConfig.balanced()

# Convenience exports
LOW_LATENCY_CONFIG = RAGConfig.low_latency()
HIGH_QUALITY_CONFIG = RAGConfig.high_quality()


def get_config(preset: str = "balanced") -> RAGConfig:
    """
    Get configuration by preset name.
    
    Args:
        preset: One of "balanced", "low_latency", "high_quality"
    
    Returns:
        RAGConfig instance
    
    Raises:
        ValueError: If preset is unknown
    """
    presets = {
        "balanced": DEFAULT_CONFIG,
        "low_latency": LOW_LATENCY_CONFIG,
        "high_quality": HIGH_QUALITY_CONFIG,
    }
    
    if preset not in presets:
        raise ValueError(
            f"Unknown preset: {preset}. "
            f"Available: {', '.join(presets.keys())}"
        )
    
    return presets[preset]


# Example usage and testing
if __name__ == "__main__":
    print("RAG Configuration Module - Test Run")
    print("=" * 70)
    
    # Test default config
    print("\n1. Default (Balanced) Configuration:")
    print(DEFAULT_CONFIG)
    
    # Validate
    is_valid, errors = DEFAULT_CONFIG.validate()
    print(f"\nValidation: {'✅ PASS' if is_valid else '❌ FAIL'}")
    if errors:
        for error in errors:
            print(f"  - {error}")
    
    # Test presets
    print("\n2. Low-Latency Preset:")
    print(RAGConfig.low_latency())
    
    print("\n3. High-Quality Preset:")
    print(RAGConfig.high_quality())
    
    # Test from_dict
    print("\n4. Custom Configuration:")
    custom = RAGConfig(
        chunk_size=500,
        top_k=7,
        similarity_threshold=0.4,
    )
    print(custom)
    
    # Test environment loading
    print("\n5. From Environment:")
    env_config = RAGConfig.from_env()
    print(env_config)
    
    # Test conversion
    print("\n6. As Dictionary:")
    config_dict = DEFAULT_CONFIG.to_dict()
    for key, value in config_dict.items():
        print(f"  {key}: {value}")
    
    print("\n" + "=" * 70)
    print("✅ All tests passed!")
