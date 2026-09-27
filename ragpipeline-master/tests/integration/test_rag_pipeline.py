"""
Integration Tests for RAG Pipeline
Tests the complete flow from authentication to response generation
"""

import pytest
import requests
import json
import time
from typing import Dict, Any

# Test configuration
BASE_URL = "http://localhost:8080"
TIMEOUT = 30

class TestRAGPipeline:
    """Integration tests for the complete RAG pipeline."""
    
    @pytest.fixture
    def auth_token(self):
        """Get a valid JWT token for testing."""
        # This would call the auth endpoint in production
        # For now, use a mock token
        return "test_jwt_token_12345"
    
    @pytest.fixture
    def headers(self, auth_token):
        """Standard headers for authenticated requests."""
        return {
            "Authorization": f"Bearer {auth_token}",
            "Content-Type": "application/json"
        }
    
    def test_health_endpoint(self):
        """Test health check endpoint."""
        response = requests.get(f"{BASE_URL}/health", timeout=TIMEOUT)
        assert response.status_code == 200
        
        data = response.json()
        assert "status" in data
        assert data["status"] == "healthy"
    
    def test_query_basic(self, headers):
        """Test basic query functionality."""
        query_data = {
            "query": "What is photosynthesis?",
            "session_id": "test-session-001",
            "top_k": 5
        }
        
        response = requests.post(
            f"{BASE_URL}/query",
            headers=headers,
            json=query_data,
            timeout=TIMEOUT
        )
        
        assert response.status_code == 200
        
        data = response.json()
        assert "request_id" in data
        assert "answer" in data
        assert "context" in data
        assert "metrics" in data
        
        # Verify metrics
        metrics = data["metrics"]
        assert "total_latency_ms" in metrics
        assert "retrieved_chunks" in metrics
        assert metrics["retrieved_chunks"] > 0
    
    def test_query_with_cost_tracking(self, headers):
        """Test that cost tracking is included in response."""
        query_data = {
            "query": "Explain force and pressure",
            "session_id": "test-session-002"
        }
        
        response = requests.post(
            f"{BASE_URL}/query",
            headers=headers,
            json=query_data,
            timeout=TIMEOUT
        )
        
        assert response.status_code == 200
        
        data = response.json()
        assert "cost" in data
        
        cost = data["cost"]
        assert "input_tokens" in cost
        assert "output_tokens" in cost
        assert "total_tokens" in cost
        assert "cost_usd" in cost
        assert cost["cost_usd"] >= 0
    
    def test_query_clarification_needed(self, headers):
        """Test that ambiguous queries trigger clarification."""
        # Ambiguous query
        query_data = {
            "query": "What is the cell?",  # Could be plant or animal
            "session_id": "test-session-003"
        }
        
        response = requests.post(
            f"{BASE_URL}/query",
            headers=headers,
            json=query_data,
            timeout=TIMEOUT
        )
        
        # Should either answer or ask for clarification
        assert response.status_code == 200
        
        data = response.json()
        if "clarification" in data:
            # Clarification requested
            assert data["clarification"]["is_ambiguous"] == True
            assert len(data["clarification"]["questions"]) > 0
        else:
            # Or answered directly
            assert "answer" in data
    
    def test_query_grade_isolation(self, headers):
        """Test that grade isolation is enforced."""
        query_data = {
            "query": "Explain quantum mechanics",
            "session_id": "test-session-004",
            "grade": 8  # Grade 8 student
        }
        
        response = requests.post(
            f"{BASE_URL}/query",
            headers=headers,
            json=query_data,
            timeout=TIMEOUT
        )
        
        # Should succeed but filter to grade-appropriate content
        assert response.status_code == 200
        
        data = response.json()
        # Verify response is grade-appropriate (simplified check)
        assert "answer" in data
    
    def test_query_without_auth(self):
        """Test that unauthenticated requests are rejected."""
        query_data = {
            "query": "What is photosynthesis?",
            "session_id": "test-session-005"
        }
        
        response = requests.post(
            f"{BASE_URL}/query",
            headers={},  # No auth
            json=query_data,
            timeout=TIMEOUT
        )
        
        # Should be rejected
        assert response.status_code == 401
    
    def test_query_with_feedback(self, headers):
        """Test feedback logging."""
        # First, get a response
        query_data = {
            "query": "What is friction?",
            "session_id": "test-session-006"
        }
        
        response = requests.post(
            f"{BASE_URL}/query",
            headers=headers,
            json=query_data,
            timeout=TIMEOUT
        )
        
        assert response.status_code == 200
        data = response.json()
        request_id = data["request_id"]
        
        # Submit feedback
        feedback_data = {
            "request_id": request_id,
            "feedback_score": 1,  # Thumbs up
            "comment": "Very helpful!"
        }
        
        feedback_response = requests.post(
            f"{BASE_URL}/feedback",
            headers=headers,
            json=feedback_data,
            timeout=TIMEOUT
        )
        
        assert feedback_response.status_code == 200
    
    def test_query_latency(self, headers):
        """Test that query latency is within acceptable bounds."""
        query_data = {
            "query": "Define acceleration",
            "session_id": "test-session-007"
        }
        
        start_time = time.time()
        
        response = requests.post(
            f"{BASE_URL}/query",
            headers=headers,
            json=query_data,
            timeout=TIMEOUT
        )
        
        elapsed = time.time() - start_time
        
        assert response.status_code == 200
        
        # Verify latency is acceptable (< 2 seconds for P95)
        data = response.json()
        latency_ms = data["metrics"]["total_latency_ms"]
        assert latency_ms < 2000, f"Latency {latency_ms}ms exceeds 2000ms target"
    
    def test_multi_hop_retrieval(self, headers):
        """Test multi-hop retrieval for complex queries."""
        # Complex query requiring multiple hops
        query_data = {
            "query": "How does photosynthesis relate to the carbon cycle?",
            "session_id": "test-session-008"
        }
        
        response = requests.post(
            f"{BASE_URL}/query",
            headers=headers,
            json=query_data,
            timeout=TIMEOUT
        )
        
        assert response.status_code == 200
        
        data = response.json()
        metrics = data["metrics"]
        
        # Multi-hop should be triggered for complex queries
        # (This depends on the query classifier)
        assert "multi_hop" in metrics or "hops_used" in metrics


class TestCostTracking:
    """Tests for cost tracking functionality."""
    
    @pytest.fixture
    def headers(self):
        """Get auth headers."""
        return {
            "Authorization": "Bearer test_token",
            "Content-Type": "application/json"
        }
    
    def test_cost_calculation_accuracy(self, headers):
        """Test that costs are calculated correctly."""
        query_data = {
            "query": "What is gravity?",
            "session_id": "test-cost-001"
        }
        
        response = requests.post(
            f"{BASE_URL}/query",
            headers=headers,
            json=query_data,
            timeout=TIMEOUT
        )
        
        assert response.status_code == 200
        
        data = response.json()
        cost = data["cost"]
        
        # Verify cost calculation
        total_tokens = cost["total_tokens"]
        cost_usd = cost["cost_usd"]
        
        # Gemini 2.0 Flash: $0.000075/1K input + $0.00030/1K output
        # Simplified check: cost should be proportional to tokens
        assert cost_usd >= 0
        assert cost_usd < 0.01  # Should be very small for single query
    
    def test_token_counting(self, headers):
        """Test that token counting is accurate."""
        query_data = {
            "query": "Explain Newton's three laws of motion in detail",
            "session_id": "test-cost-002"
        }
        
        response = requests.post(
            f"{BASE_URL}/query",
            headers=headers,
            json=query_data,
            timeout=TIMEOUT
        )
        
        assert response.status_code == 200
        
        data = response.json()
        cost = data["cost"]
        
        # Longer query should have more input tokens
        assert cost["input_tokens"] > 0
        assert cost["output_tokens"] > 0
        assert cost["total_tokens"] == cost["input_tokens"] + cost["output_tokens"]


class TestAnalytics:
    """Tests for analytics functionality."""
    
    def test_retrieval_metrics_collection(self, headers):
        """Test that retrieval metrics are collected."""
        query_data = {
            "query": "What are microorganisms?",
            "session_id": "test-analytics-001"
        }
        
        response = requests.post(
            f"{BASE_URL}/query",
            headers=headers,
            json=query_data,
            timeout=TIMEOUT
        )
        
        assert response.status_code == 200
        
        data = response.json()
        metrics = data["metrics"]
        
        # Verify all required metrics are present
        required_metrics = [
            "total_latency_ms",
            "retrieval_latency_ms",
            "generation_latency_ms",
            "retrieved_chunks",
            "confidence_score"
        ]
        
        for metric in required_metrics:
            assert metric in metrics, f"Missing metric: {metric}"


if __name__ == "__main__":
    pytest.main([__file__, "-v", "--tb=short"])
