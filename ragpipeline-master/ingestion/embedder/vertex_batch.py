"""
Vertex AI Batch Embedder

Creates embeddings using Vertex AI text-embedding-005 model.

Configuration:
- Model: text-embedding-005 (NOT gecko@003 - deprecated)
- Task type: RETRIEVAL_DOCUMENT (ingestion) - NOT RETRIEVAL_QUERY
- Output dimensionality: 768 (explicit)
- Batch size: ≤5 texts per API call (Vertex AI limit)
- Retry: tenacity with exponential backoff (max 5 attempts)
"""

import math
from typing import List, Optional, Dict, Any
from dataclasses import dataclass
import structlog

from google.cloud import aiplatform
from google.cloud.aiplatform import telemetry
from tenacity import (
    retry,
    stop_after_attempt,
    wait_exponential,
    retry_if_exception_type,
)

logger = structlog.get_logger()


@dataclass
class EmbeddingConfig:
    """Vertex AI embedding configuration.
    
    Attributes:
        project_id: GCP project ID
        location: GCP region (e.g., "asia-south1")
        model: Model name (default: "text-embedding-005")
        task_type: Task type (default: "RETRIEVAL_DOCUMENT")
        output_dimensionality: Output dimensions (default: 768)
        batch_size: Texts per API call (default: 5, max: 5)
    """
    project_id: str
    location: str
    model: str = "text-embedding-005"
    task_type: str = "RETRIEVAL_DOCUMENT"
    output_dimensionality: int = 768
    batch_size: int = 5
    
    def __post_init__(self):
        """Validate configuration."""
        if self.batch_size > 5:
            logger.warning("Batch size > 5 may cause API errors",
                          batch_size=self.batch_size)
        
        if self.output_dimensionality != 768:
            logger.warning("Non-standard dimensionality may affect retrieval",
                          dimensionality=self.output_dimensionality)


class VertexAIEmbedder:
    """Vertex AI embedding client with batching and retry logic.
    
    Attributes:
        config: EmbeddingConfig instance
        client: Vertex AI TextEmbeddingModel instance
    """
    
    def __init__(self, config: EmbeddingConfig):
        """Initialize embedder.
        
        Args:
            config: EmbeddingConfig with project and model settings
        """
        self.config = config
        
        # Initialize Vertex AI
        aiplatform.init(
            project=config.project_id,
            location=config.location,
        )
        
        # Load model
        self.client = aiplatform.TextEmbeddingModel.from_pretrained(
            config.model
        )
        
        logger.info("Vertex AI embedder initialized",
                   project=config.project_id,
                   location=config.location,
                   model=config.model,
                   task_type=config.task_type,
                   dimensionality=config.output_dimensionality)
    
    def embed_texts(self, texts: List[str]) -> List[List[float]]:
        """Embed multiple texts with batching.
        
        Args:
            texts: List of texts to embed
        
        Returns:
            List of embedding vectors (768-dimensional)
        """
        if not texts:
            return []
        
        # Calculate number of batches
        num_batches = math.ceil(len(texts) / self.config.batch_size)
        
        logger.info("Starting batch embedding",
                   total_texts=len(texts),
                   batch_size=self.config.batch_size,
                   num_batches=num_batches)
        
        all_embeddings = []
        
        for i in range(num_batches):
            start_idx = i * self.config.batch_size
            end_idx = min(start_idx + self.config.batch_size, len(texts))
            batch = texts[start_idx:end_idx]
            
            # Embed batch with retry
            batch_embeddings = self._embed_batch_with_retry(batch)
            all_embeddings.extend(batch_embeddings)
            
            logger.debug("Batch embedded",
                        batch_num=i + 1,
                        batch_size=len(batch),
                        embeddings_shape=(len(batch_embeddings), 768))
        
        # Validate all embeddings
        for i, emb in enumerate(all_embeddings):
            if len(emb) != self.config.output_dimensionality:
                logger.error("Invalid embedding dimension",
                            index=i,
                            expected=self.config.output_dimensionality,
                            actual=len(emb))
                raise ValueError(
                    f"Embedding {i} has {len(emb)} dims, expected {self.config.output_dimensionality}"
                )
        
        logger.info("Batch embedding complete",
                   total_embeddings=len(all_embeddings),
                   shape=(len(all_embeddings), self.config.output_dimensionality))
        
        return all_embeddings
    
    @retry(
        stop=stop_after_attempt(5),
        wait=wait_exponential(multiplier=1, min=2, max=60),
        retry=retry_if_exception_type(Exception),
        reraise=True,
    )
    def _embed_batch_with_retry(self, texts: List[str]) -> List[List[float]]:
        """Embed a single batch with retry logic.
        
        Args:
            texts: Batch of texts (≤5)
        
        Returns:
            List of embedding vectors
        """
        try:
            with telemetry.tool_context_manager(self._get_telemetry_context()):
                embeddings = self.client.get_embeddings(
                    texts,
                    task_type=self.config.task_type,
                    output_dimensionality=self.config.output_dimensionality,
                )
                
                return [emb.values for emb in embeddings]
                
        except Exception as e:
            logger.warning("Embedding batch failed, retrying",
                          error=str(e),
                          batch_size=len(texts))
            raise
    
    def _get_telemetry_context(self) -> str:
        """Get telemetry context for Vertex AI.
        
        Returns:
            Telemetry context string
        """
        return "visionary-rag-ingestion"
    
    def embed_single(self, text: str) -> List[float]:
        """Embed a single text.
        
        Args:
            text: Text to embed
        
        Returns:
            Embedding vector (768-dimensional)
        """
        embeddings = self.embed_texts([text])
        return embeddings[0] if embeddings else []


def embed_batch(
    texts: List[str],
    project_id: str,
    location: str,
    model: str = "text-embedding-005",
    task_type: str = "RETRIEVAL_DOCUMENT",
    output_dimensionality: int = 768,
    batch_size: int = 5,
) -> List[List[float]]:
    """Embed a batch of texts using Vertex AI.
    
    Convenience function that creates an embedder and embeds texts.
    
    Args:
        texts: List of texts to embed
        project_id: GCP project ID
        location: GCP region
        model: Model name (default: "text-embedding-005")
        task_type: Task type (default: "RETRIEVAL_DOCUMENT")
        output_dimensionality: Output dimensions (default: 768)
        batch_size: Texts per API call (default: 5)
    
    Returns:
        List of embedding vectors
    
    Examples:
        >>> embeddings = embed_batch(
        ...     ["The cell membrane", "Photosynthesis"],
        ...     project_id="my-project",
        ...     location="asia-south1"
        ... )
        >>> len(embeddings[0])
        768
    """
    config = EmbeddingConfig(
        project_id=project_id,
        location=location,
        model=model,
        task_type=task_type,
        output_dimensionality=output_dimensionality,
        batch_size=batch_size,
    )
    
    embedder = VertexAIEmbedder(config)
    return embedder.embed_texts(texts)


def validate_embeddings(
    embeddings: List[List[float]],
    expected_count: int,
    expected_dims: int = 768,
) -> Dict[str, Any]:
    """Validate embedding output.
    
    Args:
        embeddings: List of embedding vectors
        expected_count: Expected number of embeddings
        expected_dims: Expected dimensions (default: 768)
    
    Returns:
        Dict with validation results
    """
    # Check count
    count_valid = len(embeddings) == expected_count
    
    # Check dimensions
    dims_valid = all(len(emb) == expected_dims for emb in embeddings)
    
    # Check for NaN or Inf
    import math
    no_nan = all(
        all(not math.isnan(v) for v in emb)
        for emb in embeddings
    )
    no_inf = all(
        all(not math.isinf(v) for v in emb)
        for emb in embeddings
    )
    
    # Check for zero vectors
    no_zeros = all(
        any(v != 0.0 for v in emb)
        for emb in embeddings
    )
    
    return {
        "count": len(embeddings),
        "expected_count": expected_count,
        "count_valid": count_valid,
        "dimensions": expected_dims,
        "dims_valid": dims_valid,
        "no_nan": no_nan,
        "no_inf": no_inf,
        "no_zeros": no_zeros,
        "valid": count_valid and dims_valid and no_nan and no_inf and no_zeros,
    }


# Example usage:
if __name__ == "__main__":
    import os
    
    # Test embedding (requires GCP credentials)
    project_id = os.getenv("GOOGLE_CLOUD_PROJECT")
    location = os.getenv("GOOGLE_CLOUD_REGION", "asia-south1")
    
    if not project_id:
        print("Set GOOGLE_CLOUD_PROJECT environment variable to test")
        exit(1)
    
    test_texts = [
        "The cell membrane controls what enters and exits the cell.",
        "Photosynthesis converts CO₂ and H₂O into glucose and oxygen.",
        "Water is essential for all known forms of life.",
    ]
    
    embeddings = embed_batch(
        test_texts,
        project_id=project_id,
        location=location,
    )
    
    print(f"Created {len(embeddings)} embeddings")
    
    for i, emb in enumerate(embeddings):
        print(f"\nEmbedding {i+1}: {len(emb)} dimensions")
        print(f"First 10 values: {emb[:10]}")
    
    # Validate
    validation = validate_embeddings(embeddings, len(test_texts))
    print(f"\nValidation: {validation}")
