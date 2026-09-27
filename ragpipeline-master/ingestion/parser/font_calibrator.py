"""
Pass 1: Font Calibrator

Detects font size thresholds for distinguishing chapter headings, section headings,
and body text across different PDF publishers.

Inputs:  fitz.Document (PyMuPDF)
Outputs: FontThresholds dataclass with body, section, chapter font sizes

Algorithm:
  1. Sample first 20 pages via page.get_text("dict")
  2. Collect all span sizes where len(text.strip()) > 3
  3. body = statistics.mode(sizes) - most common font size
  4. section = 90th percentile of sizes
  5. chapter = 97th percentile of sizes
  6. Apply safety clamps to ensure separation
"""

import fitz  # PyMuPDF
import statistics
from dataclasses import dataclass
from typing import Dict, List, Optional
import structlog

logger = structlog.get_logger()


@dataclass
class FontThresholds:
    """Font size thresholds for heading detection.
    
    Attributes:
        body: Body text font size (mode of all font sizes)
        section: Section heading font size (90th percentile)
        chapter: Chapter heading font size (97th percentile)
        sample_count: Number of font samples analyzed
    """
    body: float
    section: float
    chapter: float
    sample_count: int
    
    def validate(self) -> bool:
        """Validate thresholds are properly separated.
        
        Returns:
            True if thresholds are valid (chapter > section > body)
        """
        return (
            self.chapter > self.section > self.body and
            self.section >= self.body + 1.0 and
            self.chapter >= self.section + 1.0
        )


def calibrate_font_thresholds(
    doc: fitz.Document,
    max_pages: int = 20,
    min_text_length: int = 3
) -> FontThresholds:
    """Calibrate font size thresholds from a PDF document.
    
    Args:
        doc: PyMuPDF document object
        max_pages: Maximum number of pages to sample (default: 20)
        min_text_length: Minimum text length to consider (filter noise)
    
    Returns:
        FontThresholds dataclass with body, section, chapter sizes
    
    Raises:
        ValueError: If document has no valid font sizes
    """
    logger.info("Starting font calibration", 
                page_count=len(doc), 
                max_pages=max_pages)
    
    # Collect all font sizes from sampled pages
    all_sizes: List[float] = []
    
    sample_pages = min(len(doc), max_pages)
    
    for page_num in range(sample_pages):
        page = doc[page_num]
        
        try:
            # Extract text with font information
            text_dict = page.get_text("dict")
            
            for block in text_dict.get("blocks", []):
                if block.get("type") != 0:  # Skip non-text blocks
                    continue
                
                for line in block.get("lines", []):
                    for span in line.get("spans", []):
                        text = span.get("text", "").strip()
                        
                        # Filter out noise (page numbers, headers, etc.)
                        if len(text) < min_text_length:
                            continue
                        
                        # Skip if text looks like noise
                        if _is_noise(text):
                            continue
                        
                        size = span.get("size", 0)
                        if size > 0:
                            all_sizes.append(round(size, 2))
                            
        except Exception as e:
            logger.warning("Page extraction failed", 
                          page=page_num, 
                          error=str(e))
            continue
    
    if not all_sizes:
        raise ValueError("No valid font sizes found in document")
    
    # Calculate thresholds
    all_sizes.sort()
    n = len(all_sizes)
    
    # Body = mode (most common font size)
    try:
        body = statistics.mode(all_sizes)
    except statistics.StatisticsError:
        # Fallback to median if no unique mode
        body = statistics.median(all_sizes)
    
    # Section = 90th percentile
    section_idx = int(n * 0.90)
    section = all_sizes[section_idx] if section_idx < n else all_sizes[-1]
    
    # Chapter = 97th percentile
    chapter_idx = int(n * 0.97)
    chapter = all_sizes[chapter_idx] if chapter_idx < n else all_sizes[-1]
    
    # Apply safety clamps
    section = max(section, body + 1.0)
    chapter = max(chapter, section + 1.0)
    
    thresholds = FontThresholds(
        body=round(body, 2),
        section=round(section, 2),
        chapter=round(chapter, 2),
        sample_count=n
    )
    
    logger.info("Font calibration complete",
                body=thresholds.body,
                section=thresholds.section,
                chapter=thresholds.chapter,
                samples=thresholds.sample_count)
    
    if not thresholds.validate():
        logger.warning("Font thresholds may be unreliable",
                      validation=thresholds.validate())
    
    return thresholds


def _is_noise(text: str) -> bool:
    """Check if text is likely noise (page numbers, URLs, etc.).
    
    Args:
        text: Text string to check
    
    Returns:
        True if text appears to be noise
    """
    import re
    
    # Page numbers
    if re.match(r"^\d{1,3}$", text):
        return True
    
    # Page X of Y patterns
    if re.match(r"^\d+\s*/\s*\d+$", text):
        return True
    
    # URLs
    if text.startswith("www.") or text.startswith("http"):
        return True
    
    # Single characters (except letters)
    if len(text) == 1 and not text.isalpha():
        return True
    
    return False


def calibrate_from_file(
    pdf_path: str,
    max_pages: int = 20
) -> FontThresholds:
    """Convenience function to calibrate from a file path.
    
    Args:
        pdf_path: Path to PDF file
        max_pages: Maximum pages to sample
    
    Returns:
        FontThresholds dataclass
    """
    doc = fitz.open(pdf_path)
    try:
        return calibrate_font_thresholds(doc, max_pages)
    finally:
        doc.close()


# Example usage:
if __name__ == "__main__":
    import sys
    
    if len(sys.argv) < 2:
        print("Usage: python font_calibrator.py <pdf_path>")
        sys.exit(1)
    
    thresholds = calibrate_from_file(sys.argv[1])
    print(f"Body: {thresholds.body}pt")
    print(f"Section: {thresholds.section}pt")
    print(f"Chapter: {thresholds.chapter}pt")
    print(f"Samples: {thresholds.sample_count}")
