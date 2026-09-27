"""
Self-Query Parser - LLM-based Filter Extraction

Extracts structured metadata filters from natural language queries
using local Ollama LLM (llama3.2:3b).
"""

from pydantic import BaseModel, Field
from typing import Optional, List
import requests
import json


class QueryFilter(BaseModel):
    """
    Structured filter extracted from natural language query.
    
    Attributes:
        grade: Grade level (6, 7, or 8)
        chapter_number: Chapter number (1-18)
        chapter_name: Chapter name (e.g., "Force and Pressure", "Photosynthesis")
        content_type: Type of content (Text, Table, Figure, Formula)
        subject: Subject name (e.g., "Science")
        page_min: Minimum page number
        page_max: Maximum page number
        query_text: The original query text for embedding search
    """
    grade: Optional[int] = Field(None, description="Grade level (6, 7, 8)")
    chapter_number: Optional[int] = Field(None, description="Chapter number (1-18)")
    chapter_name: Optional[str] = Field(None, description="Chapter name")
    content_type: Optional[str] = Field(None, description="Text, Table, Figure, or Formula")
    subject: Optional[str] = Field(None, description="Subject name")
    page_min: Optional[int] = Field(None, description="Minimum page number")
    page_max: Optional[int] = Field(None, description="Maximum page number")
    query_text: str = Field(..., description="The original query text for embedding")


# System prompt for filter extraction
SYSTEM_PROMPT = """
You are a metadata filter extractor for CBSE Science Q&A system.
Your task is to extract structured filters from natural language queries.

EXTRACTION RULES:
1. GRADE: If query mentions grade/class (e.g., "grade 8", "class 8th", "8th grade"), set grade to 6, 7, or 8
2. CHAPTER_NUMBER: If query mentions chapter number (e.g., "chapter 9", "ch 9", "chapter 10"), set chapter_number
3. CHAPTER_NAME: If query mentions chapter topic (e.g., "photosynthesis", "force and pressure", "cell"), set chapter_name
4. CONTENT_TYPE: If query asks about diagrams/figures/tables/formulas, set content_type to "Figure", "Table", or "Formula"
5. SUBJECT: Default to "Science" for CBSE science queries
6. PAGE_RANGE: If query mentions specific pages, set page_min and/or page_max
7. QUERY_TEXT: Always preserve the full original query for semantic search

IMPORTANT:
- Return ONLY valid JSON matching the QueryFilter schema
- Do not include explanations or markdown formatting
- If no filter applies, set that field to null
- Always include query_text with the original query

EXAMPLES:
Query: "Show me grade 8 photosynthesis diagrams"
→ {"grade": 8, "chapter_name": "photosynthesis", "content_type": "Figure", "query_text": "Show me grade 8 photosynthesis diagrams"}

Query: "questions from chapter 9 about force"
→ {"chapter_number": 9, "chapter_name": "force", "query_text": "questions from chapter 9 about force"}

Query: "what is force?"
→ {"query_text": "what is force?"}
"""


def extract_filters(query: str, ollama_url: str = "http://localhost:11434", model: str = "llama3.2:3b") -> QueryFilter:
    """
    Extract metadata filters from natural language query using LLM.
    
    Args:
        query: User's natural language question
        ollama_url: Ollama API endpoint (default: http://localhost:11434)
        model: Ollama model to use (default: llama3.2:3b)
        
    Returns:
        QueryFilter: Structured filters + cleaned query text
        
    Raises:
        requests.RequestException: If Ollama API call fails
        json.JSONDecodeError: If response is not valid JSON
        pydantic.ValidationError: If JSON doesn't match QueryFilter schema
    """
    try:
        response = requests.post(
            f"{ollama_url}/api/generate",
            json={
                "model": model,
                "prompt": f"Extract filters from this query: {query}",
                "system": SYSTEM_PROMPT,
                "stream": False,
                "format": "json",
                "options": {
                    "temperature": 0.1,
                    "num_predict": 256
                }
            },
            timeout=30
        )
        
        if response.status_code != 200:
            raise requests.RequestException(f"Ollama API error: {response.status_code}")
        
        response_data = response.json()
        content = response_data.get("response", "")
        
        # Parse JSON response
        # Handle potential markdown code blocks
        content = content.strip()
        if content.startswith("```json"):
            content = content[7:]
        if content.endswith("```"):
            content = content[:-3]
        content = content.strip()
        
        # Parse and validate
        filter_dict = json.loads(content)
        return QueryFilter(**filter_dict)
        
    except requests.Timeout:
        # Timeout - return fallback with no filters
        print(f"⚠️  Filter extraction timeout, using fallback")
        return QueryFilter(query_text=query)
    except requests.RequestException as e:
        # Connection error - return fallback
        print(f"⚠️  Filter extraction failed: {e}, using fallback")
        return QueryFilter(query_text=query)
    except (json.JSONDecodeError, ValueError) as e:
        # Invalid JSON - return fallback
        print(f"⚠️  Invalid JSON response: {e}, using fallback")
        return QueryFilter(query_text=query)


def extract_filters_batch(queries: List[str], ollama_url: str = "http://localhost:11434") -> List[QueryFilter]:
    """
    Extract filters from multiple queries in batch.
    
    Args:
        queries: List of natural language queries
        ollama_url: Ollama API endpoint
        
    Returns:
        List of QueryFilter objects
    """
    return [extract_filters(q, ollama_url) for q in queries]


# Example usage and testing
if __name__ == "__main__":
    test_queries = [
        "Show me grade 8 photosynthesis diagrams",
        "questions from chapter 9 about force",
        "what is force?",
        "Explain cell structure for class 7",
        "Show me tables from chapter 5",
        "What are microorganisms?",
        "grade 8 chapter 10 force and pressure questions",
    ]
    
    print("Testing Self-Query Parser\n" + "="*60)
    
    for query in test_queries:
        print(f"\nQuery: {query}")
        try:
            filters = extract_filters(query)
            print(f"  Grade: {filters.grade}")
            print(f"  Chapter #: {filters.chapter_number}")
            print(f"  Chapter Name: {filters.chapter_name}")
            print(f"  Content Type: {filters.content_type}")
            print(f"  Query Text: {filters.query_text}")
        except Exception as e:
            print(f"  Error: {e}")
