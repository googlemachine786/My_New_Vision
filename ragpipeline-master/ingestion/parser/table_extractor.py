"""
Pass 2: Table Extractor

Extracts tables from PDF pages with bounding boxes for exclusion zone mapping.

Inputs:  pdfplumber.Page
Outputs: List[TableElement] with markdown content and bbox

Algorithm:
  1. Use pdfplumber.find_tables() to locate tables
  2. Extract each table with proper cell merging
  3. Convert to markdown using tabulate library (handles edge cases)
  4. Store bounding box for Pass 5 exclusion zones
  5. Atomic: never split a table across pages
"""

import pdfplumber
from dataclasses import dataclass, field
from typing import List, Tuple, Optional, Any

import pandas as pd
from tabulate import tabulate

import structlog

logger = structlog.get_logger()


@dataclass
class TableElement:
    """Extracted table with metadata.
    
    Attributes:
        markdown: Table content in markdown format
        bbox: Bounding box (x0, y0, x1, y1) for exclusion zones
        page: Page number where table was found
        row_count: Number of rows in the table
        col_count: Number of columns in the table
    """
    markdown: str
    bbox: Tuple[float, float, float, float]
    page: int
    row_count: int = 0
    col_count: int = 0
    metadata: Dict[str, Any] = field(default_factory=dict)


def extract_tables(
    page: pdfplumber.Page,
    page_num: int,
    min_rows: int = 2,
    min_cols: int = 2
) -> List[TableElement]:
    """Extract all tables from a PDF page.
    
    Args:
        page: pdfplumber Page object
        page_num: Page number (1-indexed)
        min_rows: Minimum rows to consider as table
        min_cols: Minimum columns to consider as table
    
    Returns:
        List of TableElement objects (may be empty if no tables found)
    """
    tables: List[TableElement] = []
    
    try:
        # Find all tables on the page
        found_tables = page.find_tables()
        
        for table_idx, table in enumerate(found_tables):
            # Extract table data
            table_data = table.extract()
            
            if not table_data:
                continue
            
            # Validate table dimensions
            row_count = len(table_data)
            col_count = max(len(row) for row in table_data) if table_data else 0
            
            if row_count < min_rows or col_count < min_cols:
                logger.debug("Table too small, skipping",
                           page=page_num,
                           rows=row_count,
                           cols=col_count)
                continue
            
            # Convert to markdown
            markdown = _table_to_markdown(table_data)
            
            # Get bounding box
            bbox = (
                table.bbox.x0,
                table.bbox.y0,
                table.bbox.x1,
                table.bbox.y1
            )
            
            table_elem = TableElement(
                markdown=markdown,
                bbox=bbox,
                page=page_num,
                row_count=row_count,
                col_count=col_count,
                metadata={
                    "table_index": table_idx,
                    "has_header": _detect_header(table_data),
                }
            )
            
            tables.append(table_elem)
            logger.info("Table extracted",
                       page=page_num,
                       rows=row_count,
                       cols=col_count,
                       bbox=bbox)
    
    except Exception as e:
        logger.error("Table extraction failed",
                    page=page_num,
                    error=str(e))
    
    return tables


def _table_to_markdown(table_data: List[List[str]]) -> str:
    """Convert table data to markdown format using tabulate library.

    Tabulate handles edge cases that hand-rolled conversion misses:
    - Multi-line cells
    - Merged cells
    - Special character escaping (pipes, backslashes, etc.)
    - Proper alignment

    Args:
        table_data: 2D list of cell values

    Returns:
        Markdown-formatted table string
    """
    if not table_data:
        return ""

    # Clean cell values
    cleaned = []
    for row in table_data:
        cleaned_row = []
        for cell in row:
            if cell is None:
                cleaned_row.append("")
            else:
                cleaned_row.append(str(cell).strip())
        cleaned.append(cleaned_row)

    # Pad rows to uniform length
    col_count = max(len(row) for row in cleaned)
    cleaned = [row + [""] * (col_count - len(row)) for row in cleaned]

    # Use tabulate for proper markdown conversion
    if cleaned:
        # First row as header
        headers = cleaned[0]
        data_rows = cleaned[1:]
        return tabulate(data_rows, headers=headers, tablefmt="github")
    return ""


def _detect_header(table_data: List[List[str]]) -> bool:
    """Detect if table has a header row.
    
    Simple heuristic: first row contains non-numeric values
    while other rows contain numeric values.
    
    Args:
        table_data: 2D list of cell values
    
    Returns:
        True if table appears to have a header row
    """
    if len(table_data) < 2:
        return False
    
    first_row = table_data[0]
    second_row = table_data[1] if len(table_data) > 1 else []
    
    # Check if first row is more "text-like" than second row
    first_row_text_ratio = sum(
        1 for cell in first_row 
        if cell and not _is_numeric(cell)
    ) / len(first_row) if first_row else 0
    
    second_row_text_ratio = sum(
        1 for cell in second_row 
        if cell and not _is_numeric(cell)
    ) / len(second_row) if second_row else 0
    
    return first_row_text_ratio > second_row_text_ratio


def _is_numeric(text: str) -> bool:
    """Check if text is primarily numeric.
    
    Args:
        text: Text to check
    
    Returns:
        True if text appears to be a number
    """
    import re
    
    # Remove common numeric patterns
    numeric_text = re.sub(r"[\d.,%\-+$]", "", text.strip())
    
    # If mostly empty after removing numbers, it's numeric
    return len(numeric_text) < len(text) * 0.3


def extract_tables_from_doc(
    pdf_path: str,
    page_range: Optional[Tuple[int, int]] = None
) -> List[TableElement]:
    """Extract all tables from a PDF document.
    
    Args:
        pdf_path: Path to PDF file
        page_range: Optional (start_page, end_page) range (1-indexed)
    
    Returns:
        List of TableElement objects
    """
    all_tables: List[TableElement] = []
    
    with pdfplumber.open(pdf_path) as pdf:
        # Determine page range
        if page_range:
            start, end = page_range
            pages = range(start - 1, min(end, len(pdf)))
        else:
            pages = range(len(pdf))
        
        for page_num in pages:
            page = pdf.pages[page_num]
            tables = extract_tables(page, page_num + 1)
            all_tables.extend(tables)
    
    logger.info("Document table extraction complete",
               total_tables=len(all_tables))
    
    return all_tables


# Example usage:
if __name__ == "__main__":
    import sys
    
    if len(sys.argv) < 2:
        print("Usage: python table_extractor.py <pdf_path> [start_page] [end_page]")
        sys.exit(1)
    
    pdf_path = sys.argv[1]
    page_range = None
    
    if len(sys.argv) >= 4:
        page_range = (int(sys.argv[2]), int(sys.argv[3]))
    
    tables = extract_tables_from_doc(pdf_path, page_range)
    
    print(f"Found {len(tables)} tables:")
    for table in tables:
        print(f"\nPage {table.page}: {table.row_count}x{table.col_count}")
        print(f"BBOX: {table.bbox}")
        print(table.markdown[:200] + "...")
