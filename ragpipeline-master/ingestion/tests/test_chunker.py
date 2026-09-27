"""
Tests for Chunker Module

Test cases:
- test_child_max_512_chars: all children ≤ 512
- test_parent_max_1500_chars: all parents ≤ 1500
- test_table_atomic_no_split: table chunk_id == parent_id
- test_overlap_15_percent: overlap ≈ 77 chars
"""

import pytest
import uuid

from chunker.parent_child import (
    create_parent_child_chunks,
    ParentChunk,
    ChildChunk,
    RecursiveCharacterTextSplitter,
    validate_chunks,
)
from parser.metadata_enricher import ParsedElement


class TestRecursiveCharacterTextSplitter:
    """Tests for recursive text splitting."""
    
    def test_split_by_paragraphs(self):
        """Test splitting by paragraph boundaries."""
        splitter = RecursiveCharacterTextSplitter(
            separators=["\n\n", "\n", ". ", " ", ""],
            max_chars=100,
            overlap=20,
        )
        
        text = "First paragraph.\n\nSecond paragraph.\n\nThird paragraph."
        chunks = splitter.split_text(text)
        
        # Should split at paragraph boundaries
        assert len(chunks) > 1
        
        # All chunks should fit within limit
        for chunk in chunks:
            assert len(chunk) <= 100
    
    def test_split_by_sentences(self):
        """Test splitting by sentence boundaries."""
        splitter = RecursiveCharacterTextSplitter(
            max_chars=50,
            overlap=10,
        )
        
        text = "First sentence. Second sentence. Third sentence."
        chunks = splitter.split_text(text)
        
        # Should split at sentence boundaries
        assert len(chunks) > 1
        
        # Chunks should contain complete sentences where possible
        for chunk in chunks:
            assert len(chunk) <= 50
    
    def test_split_long_word(self):
        """Test splitting long words (character-level)."""
        splitter = RecursiveCharacterTextSplitter(
            max_chars=10,
            overlap=0,
        )
        
        text = "supercalifragilisticexpialidocious"
        chunks = splitter.split_text(text)
        
        # Should split by characters
        assert len(chunks) > 1
        
        # All chunks within limit
        for chunk in chunks:
            assert len(chunk) <= 10
    
    def test_empty_text(self):
        """Test splitting empty text."""
        splitter = RecursiveCharacterTextSplitter()
        chunks = splitter.split_text("")
        assert chunks == []
    
    def test_short_text(self):
        """Test splitting text that fits in one chunk."""
        splitter = RecursiveCharacterTextSplitter(max_chars=100)
        text = "Short text"
        chunks = splitter.split_text(text)
        
        assert len(chunks) == 1
        assert chunks[0] == text


class TestParentChildChunking:
    """Tests for parent-child chunking."""
    
    def test_child_max_512_chars(self, sample_elements):
        """Test that all child chunks are ≤ 512 characters."""
        parents, children = create_parent_child_chunks(
            sample_elements,
            child_max_chars=512,
        )
        
        for child in children:
            assert len(child.content) <= 512, f"Child chunk exceeds 512 chars: {len(child.content)}"
    
    def test_parent_max_1500_chars(self, sample_elements):
        """Test that all parent chunks are ≤ 1500 characters."""
        parents, children = create_parent_child_chunks(
            sample_elements,
            parent_max_chars=1500,
        )
        
        for parent in parents:
            assert len(parent.content) <= 1500, f"Parent chunk exceeds 1500 chars: {len(parent.content)}"
    
    def test_table_atomic_no_split(self):
        """Test that tables are not split (parent_id == child_id)."""
        table_element = ParsedElement(
            text="| Header 1 | Header 2 |\n| --- | --- |\n| Cell 1 | Cell 2 |",
            page_number=1,
            taxonomy_id=1,
            grade=6,
            subject="Science",
            content_type="table",
            is_table=True,
        )
        
        parents, children = create_parent_child_chunks([table_element])
        
        # Should create one parent and one child
        assert len(parents) == 1
        assert len(children) == 1
        
        # Parent and child should have same ID (atomic)
        assert parents[0].parent_id == children[0].child_id
    
    def test_overlap_15_percent(self):
        """Test that child overlap is approximately 15% (77 chars)."""
        # Create long text that will split
        long_text = "This is a sentence. " * 100
        
        element = ParsedElement(
            text=long_text,
            page_number=1,
            taxonomy_id=1,
            grade=6,
            subject="Science",
        )
        
        parents, children = create_parent_child_chunks(
            [element],
            child_max_chars=512,
            child_overlap=77,  # 15% of 512
        )
        
        if len(children) > 1:
            # Check that there is overlap between consecutive children
            for i in range(1, len(children)):
                prev_child = children[i - 1]
                curr_child = children[i]
                
                # Get last 77 chars of previous child
                prev_end = prev_child.content[-77:]
                
                # Check if it appears at start of current child
                # (allowing for some variation due to sentence boundaries)
                assert prev_end[:50] in curr_child.content or \
                       curr_child.content[:50] in prev_end, \
                       "Overlap should exist between consecutive children"
    
    def test_parent_child_relationship(self, sample_elements):
        """Test that all children have valid parent references."""
        parents, children = create_parent_child_chunks(sample_elements)
        
        parent_ids = {p.parent_id for p in parents}
        
        for child in children:
            assert child.parent_id in parent_ids, \
                f"Child {child.child_id} references non-existent parent {child.parent_id}"
    
    def test_taxonomy_id_propagation(self, sample_elements):
        """Test that taxonomy_id is propagated to all chunks."""
        parents, children = create_parent_child_chunks(sample_elements)
        
        expected_taxonomy = sample_elements[0].taxonomy_id
        
        for parent in parents:
            assert parent.taxonomy_id == expected_taxonomy
        
        for child in children:
            assert child.taxonomy_id == expected_taxonomy


class TestChunkValidation:
    """Tests for chunk validation."""
    
    def test_validate_chunks_valid(self):
        """Test validation of valid chunks."""
        parent = ParentChunk(
            content="Test content",
            taxonomy_id=1,
            parent_id="parent-123",
        )
        
        child = ChildChunk(
            content="Test child",
            parent_id="parent-123",
            taxonomy_id=1,
            child_id="child-456",
        )
        
        validation = validate_chunks([parent], [child])
        
        assert validation["valid"] is True
        assert validation["orphaned_children"] == 0
        assert validation["oversized_children"] == 0
        assert validation["oversized_parents"] == 0
    
    def test_validate_orphaned_children(self):
        """Test detection of orphaned children."""
        child = ChildChunk(
            content="Orphaned child",
            parent_id="non-existent-parent",
            taxonomy_id=1,
        )
        
        validation = validate_chunks([], [child])
        
        assert validation["valid"] is False
        assert validation["orphaned_children"] == 1
    
    def test_validate_oversized_child(self):
        """Test detection of oversized child chunks."""
        parent = ParentChunk(
            content="Test",
            taxonomy_id=1,
            parent_id="parent-123",
        )
        
        child = ChildChunk(
            content="X" * 600,  # Exceeds 512 limit
            parent_id="parent-123",
            taxonomy_id=1,
        )
        
        validation = validate_chunks([parent], [child])
        
        assert validation["valid"] is False
        assert validation["oversized_children"] == 1


class TestParentChunk:
    """Tests for ParentChunk dataclass."""
    
    def test_auto_uuid(self):
        """Test that parent_id is auto-generated."""
        parent = ParentChunk(
            content="Test",
            taxonomy_id=1,
        )
        
        assert parent.parent_id is not None
        assert isinstance(parent.parent_id, str)
        
        # Should be valid UUID
        uuid.UUID(parent.parent_id)  # Should not raise
    
    def test_keyword_initialization(self):
        """Test that keywords list is initialized."""
        parent = ParentChunk(
            content="Test",
            taxonomy_id=1,
        )
        
        assert parent.extracted_keywords == []
        assert isinstance(parent.extracted_keywords, list)


class TestChildChunk:
    """Tests for ChildChunk dataclass."""
    
    def test_auto_uuid(self):
        """Test that child_id is auto-generated."""
        child = ChildChunk(
            content="Test",
            parent_id="parent-123",
            taxonomy_id=1,
        )
        
        assert child.child_id is not None
        assert isinstance(child.child_id, str)
        
        # Should be valid UUID
        uuid.UUID(child.child_id)  # Should not raise
    
    def test_parent_id_required(self):
        """Test that parent_id is required."""
        with pytest.raises(TypeError):
            ChildChunk(
                content="Test",
                taxonomy_id=1,
            )


# Fixtures
@pytest.fixture
def sample_elements():
    """Create sample parsed elements for testing."""
    # Long text that will split
    long_text = "This is a long paragraph. " * 100
    
    return [
        ParsedElement(
            text=long_text,
            page_number=1,
            chapter="Test Chapter",
            section="Test Section",
            taxonomy_id=1,
            grade=6,
            subject="Science",
        ),
        ParsedElement(
            text="Short text",
            page_number=1,
            chapter="Test Chapter",
            section="Test Section",
            taxonomy_id=1,
            grade=6,
            subject="Science",
        ),
    ]


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
