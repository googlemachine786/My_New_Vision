"""
Filter Translator - Python Filters to SQL

Translates QueryFilter objects into PostgreSQL WHERE clauses
for efficient metadata filtering with pgvector.
"""

from typing import Optional, Tuple, Dict, Any
from .self_query_parser import QueryFilter


def build_sql_filter(query_filter: QueryFilter) -> Tuple[str, Dict[str, Any]]:
    """
    Translate QueryFilter into PostgreSQL WHERE clause and parameters.
    
    Args:
        query_filter: Structured filter from LLM extraction
        
    Returns:
        Tuple of (WHERE clause string, parameters dict)
        
    Example:
        >>> qf = QueryFilter(grade=8, chapter_number=9, query_text="force questions")
        >>> where, params = build_sql_filter(qf)
        >>> where
        "(metadata->>'grade')::int = :grade AND (metadata->>'chapter_number')::int = :chapter_number"
        >>> params
        {'grade': 8, 'chapter_number': 9}
    """
    conditions = []
    params: Dict[str, Any] = {}
    
    # Grade filter (exact match)
    if query_filter.grade is not None:
        conditions.append("(metadata->>'grade')::int = :grade")
        params['grade'] = query_filter.grade
    
    # Chapter number filter (exact match)
    if query_filter.chapter_number is not None:
        conditions.append("(metadata->>'chapter_number')::int = :chapter_number")
        params['chapter_number'] = query_filter.chapter_number
    
    # Chapter name filter (case-insensitive partial match)
    if query_filter.chapter_name is not None:
        conditions.append("metadata->>'chapter' ILIKE :chapter_name")
        params['chapter_name'] = f"%{query_filter.chapter_name}%"
    
    # Content type filter (exact match)
    if query_filter.content_type is not None:
        # Normalize content type
        content_type = query_filter.content_type.strip().lower()
        if content_type in ['diagram', 'figure', 'image']:
            content_type = 'Figure'
        elif content_type == 'table':
            content_type = 'Table'
        elif content_type == 'formula':
            content_type = 'Formula'
        elif content_type == 'text':
            content_type = 'Text'
        
        conditions.append("metadata->>'content_type' = :content_type")
        params['content_type'] = content_type
    
    # Subject filter (exact match)
    if query_filter.subject is not None:
        conditions.append("metadata->>'subject' = :subject")
        params['subject'] = query_filter.subject
    
    # Page range filters
    if query_filter.page_min is not None:
        conditions.append("(metadata->>'page_number')::int >= :page_min")
        params['page_min'] = query_filter.page_min
    
    if query_filter.page_max is not None:
        conditions.append("(metadata->>'page_number')::int <= :page_max")
        params['page_max'] = query_filter.page_max
    
    # Combine conditions with AND
    where_clause = " AND ".join(conditions) if conditions else "TRUE"
    
    return where_clause, params


def build_sql_filter_with_operators(
    query_filter: QueryFilter,
    use_and: bool = True
) -> Tuple[str, Dict[str, Any]]:
    """
    Build SQL filter with support for OR conditions (future extension).
    
    Args:
        query_filter: Structured filter from LLM
        use_and: If True, combine with AND; if False, combine with OR
        
    Returns:
        Tuple of (WHERE clause string, parameters dict)
    """
    # For now, same as build_sql_filter
    # Future: Support multi-grade queries like "grade 8 or 9"
    return build_sql_filter(query_filter)


def count_active_filters(query_filter: QueryFilter) -> int:
    """
    Count the number of active filters in a QueryFilter.
    
    Args:
        query_filter: QueryFilter to analyze
        
    Returns:
        Number of non-null filter fields
    """
    count = 0
    if query_filter.grade is not None:
        count += 1
    if query_filter.chapter_number is not None:
        count += 1
    if query_filter.chapter_name is not None:
        count += 1
    if query_filter.content_type is not None:
        count += 1
    if query_filter.subject is not None:
        count += 1
    if query_filter.page_min is not None:
        count += 1
    if query_filter.page_max is not None:
        count += 1
    return count


def get_filter_summary(query_filter: QueryFilter) -> str:
    """
    Get a human-readable summary of active filters.
    
    Args:
        query_filter: QueryFilter to summarize
        
    Returns:
        String describing active filters
    """
    parts = []
    
    if query_filter.grade is not None:
        parts.append(f"grade={query_filter.grade}")
    if query_filter.chapter_number is not None:
        parts.append(f"chapter={query_filter.chapter_number}")
    if query_filter.chapter_name is not None:
        parts.append(f"topic='{query_filter.chapter_name}'")
    if query_filter.content_type is not None:
        parts.append(f"type={query_filter.content_type}")
    if query_filter.subject is not None:
        parts.append(f"subject={query_filter.subject}")
    if query_filter.page_min is not None and query_filter.page_max is not None:
        parts.append(f"pages={query_filter.page_min}-{query_filter.page_max}")
    elif query_filter.page_min is not None:
        parts.append(f"page>={query_filter.page_min}")
    elif query_filter.page_max is not None:
        parts.append(f"page<={query_filter.page_max}")
    
    return ", ".join(parts) if parts else "no filters"


# Example usage and testing
if __name__ == "__main__":
    from self_query_parser import QueryFilter
    
    test_filters = [
        QueryFilter(grade=8, chapter_name="photosynthesis", query_text="grade 8 photosynthesis"),
        QueryFilter(chapter_number=9, chapter_name="force", query_text="chapter 9 force"),
        QueryFilter(query_text="what is force?"),
        QueryFilter(grade=7, content_type="Figure", query_text="grade 7 diagrams"),
        QueryFilter(grade=8, chapter_number=10, page_min=90, page_max=100, query_text="grade 8 ch10 pages 90-100"),
    ]
    
    print("Testing Filter Translator\n" + "="*60)
    
    for qf in test_filters:
        where, params = build_sql_filter(qf)
        print(f"\nQueryFilter: grade={qf.grade}, chapter={qf.chapter_name}")
        print(f"  WHERE: {where}")
        print(f"  Params: {params}")
        print(f"  Active filters: {count_active_filters(qf)}")
        print(f"  Summary: {get_filter_summary(qf)}")
