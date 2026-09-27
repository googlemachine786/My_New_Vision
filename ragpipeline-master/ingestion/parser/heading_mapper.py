"""
Pass 3: Heading Mapper

Maps PDF headings to hierarchy levels (chapter, section, subsection) using font sizes
and structural cues from the document's table of contents.

Inputs:  fitz.Document, FontThresholds, PyMuPDF ToC
Outputs: Dict[int, HeadingContext] - page → heading hierarchy mapping

Detection signals (any one sufficient):
  1. font_size >= chapter_threshold → chapter
  2. font_size >= section_threshold → section
  3. bold flag (flags & 2**4) AND size > body+0.5 → section
  4. numbered_section_regex match AND size >= body-0.5 → subsection

Skip noise: page numbers, headers, footers, URLs
"""

import fitz  # PyMuPDF
import re
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Tuple
from .font_calibrator import FontThresholds
import structlog

logger = structlog.get_logger()


@dataclass
class HeadingContext:
    """Heading hierarchy context for a page.
    
    Attributes:
        chapter: Chapter title (or None if not on chapter start)
        section: Current section title
        subsection: Current subsection title
        page: Page number
        heading_level: 1=chapter, 2=section, 3=subsection
        heading_text: Actual heading text detected
    """
    chapter: Optional[str] = None
    section: Optional[str] = None
    subsection: Optional[str] = None
    page: int = 0
    heading_level: int = 0
    heading_text: str = ""
    
    def copy_with_updates(
        self,
        section: Optional[str] = None,
        subsection: Optional[str] = None,
        **kwargs
    ) -> "HeadingContext":
        """Create a copy with updated fields."""
        return HeadingContext(
            chapter=self.chapter if "chapter" not in kwargs else kwargs.get("chapter"),
            section=section if section is not None else self.section,
            subsection=subsection if subsection is not None else self.subsection,
            page=kwargs.get("page", self.page),
            heading_level=kwargs.get("heading_level", self.heading_level),
            heading_text=kwargs.get("heading_text", self.heading_text),
        )


# Numbered section patterns (CBSE textbook styles)
NUMBERED_SECTION_REGEX = re.compile(
    r"^(?:Chapter\s+\d+|\d+\.\d+(?:\.\d+)?|[A-Z][A-Z ]{4,})\b"
)

# Noise patterns to skip
NOISE_REGEX = re.compile(r"^(\d{1,3}|page\s*\d+|\d+\s*/\s*\d+|www\.|http)$", re.IGNORECASE)


def map_headings(
    doc: fitz.Document,
    thresholds: FontThresholds,
    toc: Optional[List] = None
) -> Dict[int, HeadingContext]:
    """Map headings to hierarchy levels for all pages.
    
    Args:
        doc: PyMuPDF document
        thresholds: Font size thresholds from Pass 1
        toc: Optional table of contents from PyMuPDF
    
    Returns:
        Dict mapping page numbers to HeadingContext
    """
    logger.info("Starting heading mapping",
                page_count=len(doc),
                thresholds=thresholds)
    
    heading_map: Dict[int, HeadingContext] = {}
    current_context = HeadingContext()
    
    # Build chapter map from ToC if available
    chapter_map = _build_chapter_map_from_toc(toc) if toc else {}
    
    for page_num in range(len(doc)):
        page = doc[page_num]
        page_heading = page_num + 1  # 1-indexed
        
        # Check if this page starts a chapter (from ToC)
        if page_heading in chapter_map:
            current_context = HeadingContext(
                chapter=chapter_map[page_heading],
                page=page_heading,
                heading_level=1,
                heading_text=chapter_map[page_heading],
            )
            heading_map[page_heading] = current_context
            logger.debug("Chapter start detected from ToC",
                        page=page_heading,
                        chapter=current_context.chapter)
            continue
        
        # Analyze page for headings
        detected_heading = _detect_heading_on_page(
            page, page_heading, thresholds
        )
        
        if detected_heading:
            # Update context based on heading level
            if detected_heading.heading_level == 1:
                current_context = HeadingContext(
                    chapter=detected_heading.heading_text,
                    page=page_heading,
                    heading_level=1,
                    heading_text=detected_heading.heading_text,
                )
            elif detected_heading.heading_level == 2:
                current_context = current_context.copyWith_updates(
                    section=detected_heading.heading_text,
                    page=page_heading,
                    heading_level=2,
                    heading_text=detected_heading.heading_text,
                )
            elif detected_heading.heading_level == 3:
                current_context = current_context.copyWith_updates(
                    subsection=detected_heading.heading_text,
                    page=page_heading,
                    heading_level=3,
                    heading_text=detected_heading.heading_text,
                )
        
        # Store context for this page
        heading_map[page_heading] = current_context
    
    logger.info("Heading mapping complete",
                pages_mapped=len(heading_map),
                chapters=sum(1 for h in heading_map.values() if h.heading_level == 1))
    
    return heading_map


def _build_chapter_map_from_toc(
    toc: List
) -> Dict[int, str]:
    """Build page → chapter title map from ToC.
    
    Args:
        toc: PyMuPDF ToC (list of [level, title, page] lists)
    
    Returns:
        Dict mapping page numbers to chapter titles
    """
    chapter_map = {}
    
    for item in toc:
        if len(item) >= 3:
            level, title, page = item[0], item[1], item[2]
            # Only map level 1 (chapters)
            if level == 1:
                chapter_map[page] = title
    
    return chapter_map


def _detect_heading_on_page(
    page: fitz.Page,
    page_num: int,
    thresholds: FontThresholds
) -> Optional[HeadingContext]:
    """Detect the primary heading on a page.
    
    Args:
        page: PyMuPDF page
        page_num: Page number (1-indexed)
        thresholds: Font size thresholds
    
    Returns:
        HeadingContext if heading found, None otherwise
    """
    try:
        text_dict = page.get_text("dict")
        candidates: List[Tuple[float, str, int]] = []  # (size, text, flags)
        
        for block in text_dict.get("blocks", []):
            if block.get("type") != 0:
                continue
            
            for line in block.get("lines", []):
                for span in line.get("spans", []):
                    text = span.get("text", "").strip()
                    size = span.get("size", 0)
                    flags = span.get("flags", 0)
                    
                    # Skip noise
                    if len(text) < 3 or _is_noise(text):
                        continue
                    
                    # Skip if too short to be a heading
                    if len(text) < 10 and size < thresholds.section:
                        continue
                    
                    candidates.append((size, text, flags))
        
        if not candidates:
            return None
        
        # Find the largest text (potential heading)
        candidates.sort(key=lambda x: x[0], reverse=True)
        best_size, best_text, best_flags = candidates[0]
        
        # Determine heading level
        heading_level = _classify_heading_level(
            best_size, best_text, best_flags, thresholds
        )
        
        if heading_level > 0:
            return HeadingContext(
                page=page_num,
                heading_level=heading_level,
                heading_text=best_text,
            )
        
        return None
        
    except Exception as e:
        logger.warning("Heading detection failed",
                      page=page_num,
                      error=str(e))
        return None


def _classify_heading_level(
    font_size: float,
    text: str,
    flags: int,
    thresholds: FontThresholds
) -> int:
    """Classify heading level based on font size and text patterns.
    
    Args:
        font_size: Font size in points
        text: Heading text
        flags: Font flags (bold, italic, etc.)
        thresholds: Font size thresholds
    
    Returns:
        Heading level: 0=none, 1=chapter, 2=section, 3=subsection
    """
    # Chapter: font_size >= chapter_threshold
    if font_size >= thresholds.chapter:
        return 1
    
    # Section: font_size >= section_threshold
    if font_size >= thresholds.section:
        return 2
    
    # Section: bold flag AND size > body+0.5
    is_bold = bool(flags & (1 << 4))  # Bit 4 indicates bold
    if is_bold and font_size > thresholds.body + 0.5:
        return 2
    
    # Subsection: numbered pattern AND size >= body-0.5
    if NUMBERED_SECTION_REGEX.match(text) and font_size >= thresholds.body - 0.5:
        return 3
    
    return 0


def _is_noise(text: str) -> bool:
    """Check if text is likely noise.
    
    Args:
        text: Text to check
    
    Returns:
        True if text is noise
    """
    return bool(NOISE_REGEX.match(text))


def map_headings_from_file(
    pdf_path: str,
    max_pages: int = 20
) -> Tuple[FontThresholds, Dict[int, HeadingContext]]:
    """Convenience function to calibrate and map from a file path.
    
    Args:
        pdf_path: Path to PDF file
        max_pages: Maximum pages for font calibration
    
    Returns:
        Tuple of (FontThresholds, heading_map)
    """
    from .font_calibrator import calibrate_font_thresholds
    
    doc = fitz.open(pdf_path)
    try:
        # Get ToC
        toc = doc.get_toc()
        
        # Calibrate fonts
        thresholds = calibrate_font_thresholds(doc, max_pages)
        
        # Map headings
        heading_map = map_headings(doc, thresholds, toc)
        
        return thresholds, heading_map
    finally:
        doc.close()


# Example usage:
if __name__ == "__main__":
    import sys
    
    if len(sys.argv) < 2:
        print("Usage: python heading_mapper.py <pdf_path>")
        sys.exit(1)
    
    thresholds, heading_map = map_headings_from_file(sys.argv[1])
    
    print(f"Font thresholds: body={thresholds.body}, section={thresholds.section}, chapter={thresholds.chapter}")
    print(f"\nHeading map:")
    
    for page, ctx in sorted(heading_map.items()):
        if ctx.heading_level > 0:
            indent = "  " * (ctx.heading_level - 1)
            print(f"  Page {page}: {indent}{ctx.heading_text} (L{ctx.heading_level})")
