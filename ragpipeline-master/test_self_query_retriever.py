"""
Self-Query Retriever Test Suite

Comprehensive tests for metadata-enhanced RAG retrieval.
"""

import pytest
import asyncio
import json
from pathlib import Path

# Import retriever components
from retrievers.self_query_parser import QueryFilter, extract_filters
from retrievers.filter_translator import (
    build_sql_filter,
    count_active_filters,
    get_filter_summary
)
from retrievers.self_query_retriever import SelfQueryRetriever, SelfQueryRetrieverSync


# ============================================================================
# QueryFilter Tests
# ============================================================================

class TestQueryFilter:
    """Test QueryFilter model validation."""
    
    def test_create_with_all_fields(self):
        """Test creating QueryFilter with all fields."""
        qf = QueryFilter(
            grade=8,
            chapter_number=9,
            chapter_name="Force and Pressure",
            content_type="Figure",
            subject="Science",
            page_min=90,
            page_max=100,
            query_text="Show me grade 8 force diagrams"
        )
        
        assert qf.grade == 8
        assert qf.chapter_number == 9
        assert qf.chapter_name == "Force and Pressure"
        assert qf.content_type == "Figure"
        assert qf.subject == "Science"
        assert qf.page_min == 90
        assert qf.page_max == 100
        assert qf.query_text == "Show me grade 8 force diagrams"
    
    def test_create_with_minimal_fields(self):
        """Test creating QueryFilter with only required field."""
        qf = QueryFilter(query_text="what is force?")
        
        assert qf.grade is None
        assert qf.chapter_number is None
        assert qf.query_text == "what is force?"
    
    def test_grade_validation(self):
        """Test that grade must be 6, 7, or 8."""
        # Valid grades
        for grade in [6, 7, 8]:
            qf = QueryFilter(grade=grade, query_text="test")
            assert qf.grade == grade
    
    def test_optional_fields_default_to_none(self):
        """Test that optional fields default to None."""
        qf = QueryFilter(query_text="test")
        
        assert qf.grade is None
        assert qf.chapter_number is None
        assert qf.chapter_name is None
        assert qf.content_type is None
        assert qf.subject is None
        assert qf.page_min is None
        assert qf.page_max is None


# ============================================================================
# Filter Extraction Tests
# ============================================================================

class TestFilterExtraction:
    """Test LLM-based filter extraction."""
    
    def test_extract_grade_filter(self):
        """Test extracting grade from query."""
        query = "Show me grade 8 science questions"
        qf = extract_filters(query)
        
        assert qf.grade == 8
        assert qf.query_text == query
    
    def test_extract_chapter_number_filter(self):
        """Test extracting chapter number from query."""
        query = "questions from chapter 9 about force"
        qf = extract_filters(query)
        
        assert qf.chapter_number == 9
    
    def test_extract_chapter_name_filter(self):
        """Test extracting chapter name from query."""
        query = "explain photosynthesis"
        qf = extract_filters(query)
        
        assert qf.chapter_name is not None
        assert "photosynthesis" in qf.chapter_name.lower()
    
    def test_extract_content_type_figure(self):
        """Test extracting content type for diagrams/figures."""
        query = "show me diagrams about cells"
        qf = extract_filters(query)
        
        assert qf.content_type is not None
        assert qf.content_type.lower() in ['figure', 'diagram', 'image']
    
    def test_no_filters_for_general_query(self):
        """Test that general queries have no filters."""
        query = "what is force?"
        qf = extract_filters(query)
        
        assert qf.grade is None
        assert qf.chapter_number is None
        assert qf.chapter_name is None
        assert qf.query_text == query
    
    def test_fallback_on_timeout(self):
        """Test fallback behavior when Ollama is unavailable."""
        query = "test query"
        qf = extract_filters(query, ollama_url="http://invalid-url:9999")
        
        # Should fallback to query_text only
        assert qf.query_text == query
        assert qf.grade is None


# ============================================================================
# Filter Translator Tests
# ============================================================================

class TestFilterTranslator:
    """Test SQL filter generation."""
    
    def test_build_grade_filter(self):
        """Test building grade filter SQL."""
        qf = QueryFilter(grade=8, query_text="test")
        where, params = build_sql_filter(qf)
        
        assert "(metadata->>'grade')::int = :grade" in where
        assert params['grade'] == 8
    
    def test_build_chapter_number_filter(self):
        """Test building chapter number filter SQL."""
        qf = QueryFilter(chapter_number=9, query_text="test")
        where, params = build_sql_filter(qf)
        
        assert "(metadata->>'chapter_number')::int = :chapter_number" in where
        assert params['chapter_number'] == 9
    
    def test_build_chapter_name_filter(self):
        """Test building chapter name filter SQL."""
        qf = QueryFilter(chapter_name="photosynthesis", query_text="test")
        where, params = build_sql_filter(qf)
        
        assert "metadata->>'chapter' ILIKE :chapter_name" in where
        assert params['chapter_name'] == "%photosynthesis%"
    
    def test_build_content_type_filter(self):
        """Test building content type filter SQL."""
        qf = QueryFilter(content_type="Figure", query_text="test")
        where, params = build_sql_filter(qf)
        
        assert "metadata->>'content_type' = :content_type" in where
        assert params['content_type'] == 'Figure'
    
    def test_build_multiple_filters(self):
        """Test building combined filters."""
        qf = QueryFilter(
            grade=8,
            chapter_number=9,
            content_type="Figure",
            query_text="test"
        )
        where, params = build_sql_filter(qf)
        
        assert " AND " in where
        assert params['grade'] == 8
        assert params['chapter_number'] == 9
        assert params['content_type'] == 'Figure'
    
    def test_build_no_filters(self):
        """Test building SQL with no filters."""
        qf = QueryFilter(query_text="test")
        where, params = build_sql_filter(qf)
        
        assert where == "TRUE"
        assert len(params) == 0
    
    def test_count_active_filters(self):
        """Test counting active filters."""
        qf = QueryFilter(grade=8, chapter_number=9, query_text="test")
        assert count_active_filters(qf) == 2
        
        qf2 = QueryFilter(query_text="test")
        assert count_active_filters(qf2) == 0
    
    def test_get_filter_summary(self):
        """Test filter summary generation."""
        qf = QueryFilter(grade=8, chapter_number=9, query_text="test")
        summary = get_filter_summary(qf)
        
        assert "grade=8" in summary
        assert "chapter=9" in summary


# ============================================================================
# SelfQueryRetriever Tests
# ============================================================================

class TestSelfQueryRetriever:
    """Test self-query retriever integration."""
    
    @pytest.mark.asyncio
    async def test_search_with_filters(self):
        """Test search with metadata filters."""
        retriever = SelfQueryRetriever()
        
        query = "Show me grade 8 photosynthesis diagrams"
        results = await retriever.search(query)
        
        # Should return results (mock or real)
        assert isinstance(results, list)
        assert len(results) > 0
    
    @pytest.mark.asyncio
    async def test_search_with_stats(self):
        """Test search with performance statistics."""
        retriever = SelfQueryRetriever()
        
        query = "grade 8 chapter 9 force questions"
        stats = await retriever.search_with_stats(query)
        
        assert 'results' in stats
        assert 'latency_ms' in stats
        assert 'filters_applied' in stats
        assert 'filter_summary' in stats
        assert 'active_filters' in stats
    
    @pytest.mark.asyncio
    async def test_grade_filtering_accuracy(self):
        """Test that grade filters are correctly applied."""
        retriever = SelfQueryRetriever()
        
        query = "Show me grade 8 science questions"
        results = await retriever.search(query)
        
        # All results should be grade 8 (or mock data with grade 8)
        for result in results:
            assert result['metadata'].get('grade') == 8
    
    @pytest.mark.asyncio
    async def test_chapter_filtering_accuracy(self):
        """Test that chapter filters are correctly applied."""
        retriever = SelfQueryRetriever()
        
        query = "questions from chapter 9"
        results = await retriever.search(query)
        
        # All results should be from chapter 9 (or mock data)
        for result in results:
            assert result['metadata'].get('chapter_number') == 9
    
    @pytest.mark.asyncio
    async def test_no_filter_fallback(self):
        """Test search without filters."""
        retriever = SelfQueryRetriever()
        
        query = "what is force?"
        results = await retriever.search(query)
        
        # Should return results even without filters
        assert len(results) > 0
    
    @pytest.mark.asyncio
    async def test_compare_search(self):
        """Test comparison between self-query and dense-only."""
        retriever = SelfQueryRetriever()
        
        query = "grade 8 chapter 9 force"
        comparison = await retriever.compare_search(query)
        
        assert 'self_query' in comparison
        assert 'dense_only' in comparison
        assert 'speedup' in comparison or 'filter_reduction' in comparison


# ============================================================================
# Integration Tests
# ============================================================================

class TestIntegration:
    """Integration tests for full pipeline."""
    
    @pytest.mark.asyncio
    async def test_end_to_end_retrieval(self):
        """Test complete retrieval pipeline."""
        retriever = SelfQueryRetriever()
        
        # Complex query with multiple filters
        query = "Show me grade 8 photosynthesis diagrams from chapter 5"
        stats = await retriever.search_with_stats(query)
        
        # Verify extraction
        assert stats['active_filters'] >= 2  # grade + chapter/type
        
        # Verify results
        assert len(stats['results']) > 0
        
        # Verify latency is reasonable (< 5 seconds for mock, < 1s for real)
        assert stats['latency_ms'] < 5000
    
    def test_synchronous_wrapper(self):
        """Test synchronous wrapper class."""
        retriever = SelfQueryRetrieverSync()
        
        query = "grade 8 science questions"
        results = retriever.search(query)
        
        assert isinstance(results, list)
        assert len(results) > 0


# ============================================================================
# Performance Tests
# ============================================================================

class TestPerformance:
    """Performance benchmark tests."""
    
    @pytest.mark.asyncio
    async def test_filter_extraction_latency(self):
        """Test filter extraction latency."""
        import time
        
        query = "Show me grade 8 photosynthesis diagrams"
        
        start = time.time()
        qf = extract_filters(query)
        latency_ms = (time.time() - start) * 1000
        
        # Should complete in < 5 seconds (includes LLM call)
        assert latency_ms < 5000
        print(f"\nFilter extraction latency: {latency_ms:.2f}ms")
    
    @pytest.mark.asyncio
    async def test_full_retrieval_latency(self):
        """Test end-to-end retrieval latency."""
        import time
        
        retriever = SelfQueryRetriever()
        query = "grade 8 chapter 9 force questions"
        
        start = time.time()
        results = await retriever.search(query)
        latency_ms = (time.time() - start) * 1000
        
        # Should complete in < 5 seconds (mock) or < 1 second (real DB)
        assert latency_ms < 5000
        print(f"\nFull retrieval latency: {latency_ms:.2f}ms")
    
    @pytest.mark.asyncio
    async def test_batch_filter_extraction(self):
        """Test batch filter extraction performance."""
        import time
        
        queries = [
            "grade 8 photosynthesis",
            "chapter 9 force",
            "cell diagrams",
            "what is combustion?",
            "grade 7 microorganisms"
        ]
        
        start = time.time()
        from retrievers.self_query_parser import extract_filters_batch
        results = extract_filters_batch(queries)
        latency_ms = (time.time() - start) * 1000
        
        assert len(results) == 5
        print(f"\nBatch extraction latency (5 queries): {latency_ms:.2f}ms")


# ============================================================================
# Run Tests
# ============================================================================

if __name__ == "__main__":
    # Run tests with pytest
    pytest.main([
        __file__,
        "-v",
        "--tb=short",
        "-s"  # Show print statements
    ])
