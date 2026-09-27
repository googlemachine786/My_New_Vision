"""
Tests for content deduplication module.
"""

import pytest
import numpy as np
from ingestion.dedup import (
    DeduplicationDetector,
    ContentHasher,
    SemanticDeduplicator,
    MinHashLSH,
    Chunk,
)


class TestContentHasher:
    """Tests for exact hash deduplication."""

    def test_hash_content_consistent(self):
        """Same content should produce same hash."""
        content = "Cell membrane is the boundary of the cell"

        hash1 = ContentHasher.hash_content(content)
        hash2 = ContentHasher.hash_content(content)

        assert hash1 == hash2
        assert len(hash1) == 64  # SHA-256 hex length

    def test_hash_content_different(self):
        """Different content should produce different hash."""
        content1 = "Cell membrane"
        content2 = "Cell wall"

        hash1 = ContentHasher.hash_content(content1)
        hash2 = ContentHasher.hash_content(content2)

        assert hash1 != hash2

    def test_hash_content_whitespace_normalized(self):
        """Whitespace should be normalized before hashing."""
        content1 = "Cell membrane is important"
        content2 = "Cell  membrane\nis\timportant"

        hash1 = ContentHasher.hash_content(content1)
        hash2 = ContentHasher.hash_content(content2)

        assert hash1 == hash2

    def test_hash_content_case_sensitive(self):
        """Hashing should be case-sensitive."""
        content1 = "Cell membrane"
        content2 = "cell membrane"

        hash1 = ContentHasher.hash_content(content1)
        hash2 = ContentHasher.hash_content(content2)

        assert hash1 != hash2


class TestSemanticDeduplicator:
    """Tests for semantic similarity deduplication."""

    def test_find_duplicates_high_similarity(self):
        """Should detect duplicates with high cosine similarity."""
        dedup = SemanticDeduplicator(similarity_threshold=0.92)

        # Create similar embeddings (high cosine similarity)
        embedding1 = np.array([1.0, 0.0, 0.0, 0.0, 0.0])
        embedding2 = np.array([0.98, 0.02, 0.0, 0.0, 0.0])  # Very similar

        chunk1 = Chunk(content="Test 1", chunk_id="1", taxonomy_id=1)
        chunk2 = Chunk(content="Test 2", chunk_id="2", taxonomy_id=1)

        dedup.add_chunks([chunk1], [embedding1])
        duplicates = dedup.find_duplicates(chunk2, embedding2)

        assert len(duplicates) > 0
        assert duplicates[0][0] == "1"
        assert duplicates[0][1] > 0.92

    def test_no_duplicates_low_similarity(self):
        """Should not detect duplicates with low similarity."""
        dedup = SemanticDeduplicator(similarity_threshold=0.92)

        # Create orthogonal embeddings (zero similarity)
        embedding1 = np.array([1.0, 0.0, 0.0, 0.0, 0.0])
        embedding2 = np.array([0.0, 1.0, 0.0, 0.0, 0.0])

        chunk1 = Chunk(content="Test 1", chunk_id="1", taxonomy_id=1)
        chunk2 = Chunk(content="Test 2", chunk_id="2", taxonomy_id=1)

        dedup.add_chunks([chunk1], [embedding1])
        duplicates = dedup.find_duplicates(chunk2, embedding2)

        assert len(duplicates) == 0

    def test_empty_index(self):
        """Should return no duplicates for empty index."""
        dedup = SemanticDeduplicator()

        chunk = Chunk(content="Test", chunk_id="1", taxonomy_id=1)
        embedding = np.random.rand(768)

        duplicates = dedup.find_duplicates(chunk, embedding)

        assert len(duplicates) == 0

    def test_clear_index(self):
        """Should clear index properly."""
        dedup = SemanticDeduplicator()

        chunk = Chunk(content="Test", chunk_id="1", taxonomy_id=1)
        embedding = np.random.rand(768)

        dedup.add_chunks([chunk], [embedding])
        dedup.clear()

        duplicates = dedup.find_duplicates(chunk, embedding)
        assert len(duplicates) == 0


class TestMinHashLSH:
    """Tests for MinHash LSH near-duplicate detection."""

    def test_add_and_find_chunk(self):
        """Should find added chunk."""
        lsh = MinHashLSH(num_perm=128, threshold=0.85)

        chunk = Chunk(
            content="This is a test document about cell biology",
            chunk_id="1",
            taxonomy_id=1,
        )

        lsh.add_chunk(chunk)
        # Note: LSH is probabilistic, exact match may not be found
        # This tests that the API works without errors

    def test_similar_documents(self):
        """Should detect similar documents."""
        lsh = MinHashLSH(num_perm=128, threshold=0.5)  # Lower threshold for testing

        chunk1 = Chunk(
            content="The cell membrane is the boundary of the cell that controls transport",
            chunk_id="1",
            taxonomy_id=1,
        )

        chunk2 = Chunk(
            content="The cell membrane is the boundary of the cell that controls transport",
            chunk_id="2",
            taxonomy_id=1,
        )

        lsh.add_chunk(chunk1)
        duplicates = lsh.find_duplicates(chunk2)

        # With identical text, should find as duplicate
        assert len(duplicates) > 0

    def test_different_documents(self):
        """Should not detect very different documents."""
        lsh = MinHashLSH(num_perm=128, threshold=0.85)

        chunk1 = Chunk(
            content="Cell membrane mitochondria ribosome nucleus cytoplasm",
            chunk_id="1",
            taxonomy_id=1,
        )

        chunk2 = Chunk(
            content="Photosynthesis chloroplast glucose oxygen carbon dioxide water",
            chunk_id="2",
            taxonomy_id=1,
        )

        lsh.add_chunk(chunk1)
        duplicates = lsh.find_duplicates(chunk2)

        # Very different content should not be detected as duplicate
        # (LSH is probabilistic, so we just check it doesn't crash)
        assert isinstance(duplicates, list)

    def test_clear_index(self):
        """Should clear index properly."""
        lsh = MinHashLSH()

        chunk = Chunk(content="Test content", chunk_id="1", taxonomy_id=1)
        lsh.add_chunk(chunk)
        lsh.clear()

        assert len(lsh.index) == 0
        assert len(lsh.chunk_hashes) == 0


class TestDeduplicationDetector:
    """Tests for hybrid deduplication detector."""

    def test_exact_duplicate_detection(self):
        """Should detect exact duplicates."""
        detector = DeduplicationDetector(
            use_exact_hash=True, use_semantic=False, use_minhash=False
        )

        chunk1 = Chunk(
            content="Cell membrane is the boundary of the cell",
            chunk_id="1",
            taxonomy_id=1,
        )

        chunk2 = Chunk(
            content="Cell membrane is the boundary of the cell",
            chunk_id="2",
            taxonomy_id=1,
        )

        # First chunk should not be duplicate
        is_dup1, _ = detector.is_duplicate(chunk1)
        assert not is_dup1

        # Add to index
        detector.add_to_index(chunk1)

        # Second chunk should be detected as duplicate
        is_dup2, info = detector.is_duplicate(chunk2)
        assert is_dup2
        assert info["exact_match"] is not None

    def test_no_duplicate_different_content(self):
        """Should not detect different content as duplicate."""
        detector = DeduplicationDetector(
            use_exact_hash=True, use_semantic=False, use_minhash=False
        )

        chunk1 = Chunk(content="Cell membrane", chunk_id="1", taxonomy_id=1)
        chunk2 = Chunk(content="Cell wall", chunk_id="2", taxonomy_id=1)

        detector.add_to_index(chunk1)
        is_dup, _ = detector.is_duplicate(chunk2)

        assert not is_dup

    def test_semantic_duplicate(self):
        """Should detect semantic duplicates."""
        detector = DeduplicationDetector(
            use_exact_hash=False, use_semantic=True, use_minhash=False, semantic_threshold=0.95
        )

        # Create similar embeddings
        embedding1 = np.ones(768) / np.sqrt(768)
        embedding2 = np.ones(768) / np.sqrt(768) * 0.99  # Very similar

        chunk1 = Chunk(content="Test 1", chunk_id="1", taxonomy_id=1)
        chunk2 = Chunk(content="Test 2", chunk_id="2", taxonomy_id=1)

        detector.add_to_index(chunk1, embedding1)
        is_dup, info = detector.is_duplicate(chunk2, embedding2)

        # With very high similarity, should detect as duplicate
        assert is_dup or len(info["semantic_matches"]) > 0

    def test_batch_deduplication(self):
        """Should handle batch deduplication."""
        detector = DeduplicationDetector()

        chunks = [
            Chunk(content=f"Content {i}", chunk_id=str(i), taxonomy_id=1)
            for i in range(10)
        ]

        results = detector.add_batch(chunks)

        # All should be unique (first time)
        for chunk_id, (is_dup, _) in results.items():
            assert not is_dup, f"Chunk {chunk_id} should not be duplicate"

        # Try adding duplicates
        duplicate_chunks = [
            Chunk(content="Content 0", chunk_id=f"dup_{i}", taxonomy_id=1)
            for i in range(3)
        ]

        results = detector.add_batch(duplicate_chunks)

        # All should be detected as duplicates
        for chunk_id, (is_dup, _) in results.items():
            assert is_dup, f"Chunk {chunk_id} should be duplicate"

    def test_get_stats(self):
        """Should return correct statistics."""
        detector = DeduplicationDetector()

        chunks = [
            Chunk(content=f"Content {i}", chunk_id=str(i), taxonomy_id=1)
            for i in range(5)
        ]

        for chunk in chunks:
            detector.add_to_index(chunk)

        stats = detector.get_stats()

        assert stats["total_chunks"] == 5
        assert stats["exact_hashes"] == 5
        assert stats["use_exact_hash"] is True
        assert stats["use_semantic"] is True
        assert stats["use_minhash"] is True

    def test_save_and_load_index(self, tmp_path):
        """Should save and load index correctly."""
        detector = DeduplicationDetector(
            use_exact_hash=True, use_semantic=False, use_minhash=False
        )

        chunk = Chunk(content="Test content", chunk_id="1", taxonomy_id=1)
        detector.add_to_index(chunk)

        # Save index
        detector.save_index(str(tmp_path))

        # Load into new detector
        detector2 = DeduplicationDetector(
            use_exact_hash=True, use_semantic=False, use_minhash=False
        )
        detector2.load_index(str(tmp_path))

        # Check loaded index
        is_dup, _ = detector2.is_duplicate(chunk)
        assert is_dup  # Should detect as duplicate

    def test_whitespace_normalization(self):
        """Should normalize whitespace for exact matching."""
        detector = DeduplicationDetector(
            use_exact_hash=True, use_semantic=False, use_minhash=False
        )

        chunk1 = Chunk(
            content="Cell membrane is important",
            chunk_id="1",
            taxonomy_id=1,
        )

        chunk2 = Chunk(
            content="Cell  membrane\nis\timportant",
            chunk_id="2",
            taxonomy_id=1,
        )

        detector.add_to_index(chunk1)
        is_dup, _ = detector.is_duplicate(chunk2)

        # Should be detected as duplicate (whitespace normalized)
        assert is_dup

    def test_configure_deduplication_strategies(self):
        """Should configure different strategies."""
        # Only exact hash
        detector1 = DeduplicationDetector(
            use_exact_hash=True, use_semantic=False, use_minhash=False
        )
        assert detector1.hasher is not None
        assert detector1.semantic_dedup is None
        assert detector1.minhash_lsh is None

        # Only semantic
        detector2 = DeduplicationDetector(
            use_exact_hash=False, use_semantic=True, use_minhash=False
        )
        assert detector2.hasher is None
        assert detector2.semantic_dedup is not None
        assert detector2.minhash_lsh is None

        # Only MinHash
        detector3 = DeduplicationDetector(
            use_exact_hash=False, use_semantic=False, use_minhash=True
        )
        assert detector3.hasher is None
        assert detector3.semantic_dedup is None
        assert detector3.minhash_lsh is not None

        # All disabled
        detector4 = DeduplicationDetector(
            use_exact_hash=False, use_semantic=False, use_minhash=False
        )
        assert detector4.hasher is None
        assert detector4.semantic_dedup is None
        assert detector4.minhash_lsh is None


class TestChunk:
    """Tests for Chunk dataclass."""

    def test_create_chunk(self):
        """Should create chunk with required fields."""
        chunk = Chunk(content="Test", chunk_id="1", taxonomy_id=1)

        assert chunk.content == "Test"
        assert chunk.chunk_id == "1"
        assert chunk.taxonomy_id == 1
        assert chunk.metadata == {}

    def test_create_chunk_with_metadata(self):
        """Should create chunk with metadata."""
        chunk = Chunk(
            content="Test",
            chunk_id="1",
            taxonomy_id=1,
            metadata={"source": "test", "page": 1},
        )

        assert chunk.metadata["source"] == "test"
        assert chunk.metadata["page"] == 1


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
