"""
Pass 5: Metadata Enricher

Combines outputs from passes 1-4 into structured ParsedElement objects
ready for chunking.

Inputs:  Raw text, HeadingContext, page_number, content_type, taxonomy_id
Outputs: ParsedElement with all metadata fields populated

This is the final pass that integrates:
- Font calibration (Pass 1) - for heading detection
- Table extraction (Pass 2) - for table content and bbox
- Heading mapping (Pass 3) - for chapter/section/subsection
- Formula detection (Pass 4) - for content_type classification
"""

from dataclasses import dataclass, field
from typing import Optional, Dict, Any, List
from .heading_mapper import HeadingContext
from .table_extractor import TableElement
from .formula_detector import FormulaResult, detect_formulas
import structlog

logger = structlog.get_logger()


@dataclass
class ParsedElement:
    """Structured parsed element ready for chunking.
    
    Attributes:
        text: Raw text content
        page_number: Page number in PDF (1-indexed)
        chapter: Chapter title (or None)
        section: Section title (or None)
        subsection: Subsection title (or None)
        grade: Grade level (6, 7, or 8)
        subject: Subject name (e.g., "Science")
        content_type: "prose", "formula", "table", or "equation"
        taxonomy_id: Reference to cbse_taxonomy table
        is_table: True if this is a table element
        table_bbox: Bounding box if table (for exclusion zones)
        formula_annotations: List of detected formulas
        metadata: Additional metadata (font_size, flags, etc.)
    """
    text: str
    page_number: int = 0
    chapter: Optional[str] = None
    section: Optional[str] = None
    subsection: Optional[str] = None
    grade: int = 0
    subject: str = ""
    content_type: str = "prose"
    taxonomy_id: int = 0
    is_table: bool = False
    table_bbox: Optional[tuple] = None
    formula_annotations: List[str] = field(default_factory=list)
    metadata: Dict[str, Any] = field(default_factory=dict)
    
    def validate(self) -> bool:
        """Validate required fields are populated.
        
        Returns:
            True if element is valid
        """
        return (
            len(self.text.strip()) > 0 and
            self.page_number > 0 and
            self.taxonomy_id > 0
        )
    
    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for serialization."""
        return {
            "text": self.text,
            "page_number": self.page_number,
            "chapter": self.chapter,
            "section": self.section,
            "subsection": self.subsection,
            "grade": self.grade,
            "subject": self.subject,
            "content_type": self.content_type,
            "taxonomy_id": self.taxonomy_id,
            "is_table": self.is_table,
            "table_bbox": self.table_bbox,
            "formula_annotations": self.formula_annotations,
            "metadata": self.metadata,
        }


def enrich_metadata(
    text: str,
    heading_context: HeadingContext,
    page_number: int,
    taxonomy_id: int,
    grade: int,
    subject: str,
    content_type: str = "prose",
    is_table: bool = False,
    table_element: Optional[TableElement] = None,
    font_size: float = 0.0,
    span_flags: int = 0,
) -> ParsedElement:
    """Enrich raw text with metadata from all passes.
    
    Args:
        text: Raw text content
        heading_context: Heading hierarchy from Pass 3
        page_number: Page number (1-indexed)
        taxonomy_id: Reference to cbse_taxonomy
        grade: Grade level (6-8)
        subject: Subject name
        content_type: Initial content type
        is_table: True if this is a table
        table_element: TableElement if table (from Pass 2)
        font_size: Font size (for formula detection)
        span_flags: PyMuPDF span flags (for formula detection)
    
    Returns:
        ParsedElement with all metadata populated
    """
    # Detect formulas (Pass 4)
    formula_result = detect_formulas(text, span_flags)
    
    # Override content type if formula detected with high confidence
    final_content_type = content_type
    if formula_result.confidence > 0.7:
        final_content_type = formula_result.content_type
    
    # Tables override content type
    if is_table:
        final_content_type = "table"
    
    # Build parsed element
    element = ParsedElement(
        text=text,
        page_number=page_number,
        chapter=heading_context.chapter,
        section=heading_context.section,
        subsection=heading_context.subsection,
        grade=grade,
        subject=subject,
        content_type=final_content_type,
        taxonomy_id=taxonomy_id,
        is_table=is_table,
        table_bbox=table_element.bbox if table_element else None,
        formula_annotations=formula_result.formula_annotations,
        metadata={
            "heading_level": heading_context.heading_level,
            "font_size": font_size,
            "span_flags": span_flags,
            "formula_confidence": formula_result.confidence,
            "has_subscripts": formula_result.has_subscripts,
            "has_measurements": formula_result.has_measurements,
        },
    )
    
    # Validate
    if not element.validate():
        logger.warning("ParsedElement validation failed",
                      page=page_number,
                      text_length=len(text),
                      taxonomy_id=taxonomy_id)
    
    return element


def enrich_table_element(
    table: TableElement,
    heading_context: HeadingContext,
    taxonomy_id: int,
    grade: int,
    subject: str,
) -> ParsedElement:
    """Create ParsedElement from TableElement.
    
    Args:
        table: TableElement from Pass 2
        heading_context: Heading context for the page
        taxonomy_id: Reference to cbse_taxonomy
        grade: Grade level
        subject: Subject name
    
    Returns:
        ParsedElement for the table
    """
    return ParsedElement(
        text=table.markdown,
        page_number=table.page,
        chapter=heading_context.chapter,
        section=heading_context.section,
        subsection=heading_context.subsection,
        grade=grade,
        subject=subject,
        content_type="table",
        taxonomy_id=taxonomy_id,
        is_table=True,
        table_bbox=table.bbox,
        metadata={
            "row_count": table.row_count,
            "col_count": table.col_count,
            "has_header": table.metadata.get("has_header", False),
            "table_index": table.metadata.get("table_index", 0),
        },
    )


def create_parsed_elements_from_blocks(
    text_blocks: List[Dict[str, Any]],
    heading_map: Dict[int, HeadingContext],
    taxonomy_id: int,
    grade: int,
    subject: str,
    page_tables: Dict[int, List[TableElement]] = None,
) -> List[ParsedElement]:
    """Create ParsedElements from PyMuPDF text blocks.
    
    Args:
        text_blocks: List of text blocks from PyMuPDF
        heading_map: Heading mapping from Pass 3
        taxonomy_id: Reference to cbse_taxonomy
        grade: Grade level
        subject: Subject name
        page_tables: Tables per page from Pass 2
    
    Returns:
        List of ParsedElement objects
    """
    elements: List[ParsedElement] = []
    page_tables = page_tables or {}
    
    for block in text_blocks:
        page_num = block.get("page", 1)
        text = block.get("text", "").strip()
        
        # Skip empty blocks
        if not text:
            continue
        
        # Get heading context for this page
        heading_context = heading_map.get(page_num, HeadingContext())
        
        # Check if this block is part of a table
        tables_on_page = page_tables.get(page_num, [])
        is_table_block = False
        matching_table = None
        
        for table in tables_on_page:
            bbox = table.bbox
            block_bbox = block.get("bbox", (0, 0, 0, 0))
            
            # Check if block bbox overlaps with table bbox
            if _bbox_overlap(bbox, block_bbox):
                is_table_block = True
                matching_table = table
                break
        
        # Skip if this block is part of a table (handled separately)
        if is_table_block and matching_table:
            continue
        
        # Get font info
        font_size = block.get("font_size", 0.0)
        span_flags = block.get("flags", 0)
        
        # Create enriched element
        element = enrich_metadata(
            text=text,
            heading_context=heading_context,
            page_number=page_num,
            taxonomy_id=taxonomy_id,
            grade=grade,
            subject=subject,
            font_size=font_size,
            span_flags=span_flags,
        )
        
        elements.append(element)
    
    # Add table elements
    for page_num, tables in page_tables.items():
        heading_context = heading_map.get(page_num, HeadingContext())
        
        for table in tables:
            table_element = enrich_table_element(
                table=table,
                heading_context=heading_context,
                taxonomy_id=taxonomy_id,
                grade=grade,
                subject=subject,
            )
            elements.append(table_element)
    
    logger.info("Metadata enrichment complete",
               elements_created=len(elements),
               tables_added=sum(len(t) for t in page_tables.values()))
    
    return elements


def _bbox_overlap(
    bbox1: tuple,
    bbox2: tuple,
    threshold: float = 0.5
) -> bool:
    """Check if two bounding boxes overlap significantly.
    
    Args:
        bbox1: First bbox (x0, y0, x1, y1)
        bbox2: Second bbox
        threshold: Overlap threshold (0.0-1.0)
    
    Returns:
        True if boxes overlap
    """
    x0_1, y0_1, x1_1, y1_1 = bbox1
    x0_2, y0_2, x1_2, y1_2 = bbox2
    
    # Calculate overlap
    x_overlap = max(0, min(x1_1, x1_2) - max(x0_1, x0_2))
    y_overlap = max(0, min(y1_1, y1_2) - max(y0_1, y0_2))
    
    overlap_area = x_overlap * y_overlap
    
    if overlap_area == 0:
        return False
    
    # Calculate smaller box area
    area1 = (x1_1 - x0_1) * (y1_1 - y0_1)
    area2 = (x1_2 - x0_2) * (y1_2 - y0_2)
    smaller_area = min(area1, area2)
    
    return (overlap_area / smaller_area) > threshold if smaller_area > 0 else False


# Example usage:
if __name__ == "__main__":
    # Test enrichment
    from .heading_mapper import HeadingContext
    
    heading = HeadingContext(
        chapter="Cell Structure",
        section="The Cell Membrane",
        subsection=None,
        page=42,
        heading_level=2,
    )
    
    element = enrich_metadata(
        text="The cell membrane (H₂O) controls what enters and exits.",
        heading_context=heading,
        page_number=42,
        taxonomy_id=15,
        grade=8,
        subject="Science",
    )
    
    print(f"Text: {element.text}")
    print(f"Chapter: {element.chapter}")
    print(f"Section: {element.section}")
    print(f"Content Type: {element.content_type}")
    print(f"Formulas: {element.formula_annotations}")
