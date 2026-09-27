"""
PPTX presentation parser using python-pptx.
Extracts slide content and speaker notes.
"""

from pathlib import Path
from typing import List, Dict, Any
from pptx import Presentation

from . import BaseParser, ParsedElement


class PPTXParser(BaseParser):
    """Parser for Microsoft PowerPoint PPTX documents."""

    def get_supported_extensions(self) -> List[str]:
        """Return supported file extensions."""
        return [".pptx", ".ppt"]

    def parse(self, file_path: str, **kwargs) -> List[ParsedElement]:
        """
        Parse a PPTX file and extract slide content.

        Args:
            file_path: Path to PPTX file
            **kwargs: Additional arguments (grade, subject, taxonomy_id)

        Returns:
            List of ParsedElement objects
        """
        grade = kwargs.get("grade", 7)
        subject = kwargs.get("subject", "Science")
        taxonomy_id = kwargs.get("taxonomy_id", 0)

        if not Path(file_path).exists():
            raise FileNotFoundError(f"PPTX file not found: {file_path}")

        prs = Presentation(file_path)
        elements = []

        current_chapter = ""
        current_section = ""

        for slide_num, slide in enumerate(prs.slides, 1):
            # Try to extract title as section/chapter
            title = ""
            if slide.shapes.title:
                title = slide.shapes.title.text.strip()
                if slide_num == 1:
                    current_chapter = title
                else:
                    current_section = title

            # Extract text from shapes
            slide_content = []
            has_table = False

            for shape in slide.shapes:
                if not shape.has_text_frame:
                    continue

                # Skip title (already processed)
                if shape == slide.shapes.title:
                    continue

                text_frame = shape.text_frame
                for paragraph in text_frame.paragraphs:
                    text = paragraph.text.strip()
                    if text:
                        # Check if this looks like a table
                        if self._is_table_shape(shape):
                            has_table = True
                        slide_content.append(text)

            # Combine slide content
            if slide_content:
                content = "\n".join(slide_content)

                # Determine content type
                content_type = "table" if has_table else "prose"

                elements.append(
                    ParsedElement(
                        text=content,
                        page_number=slide_num,  # Use slide number as page
                        chapter=current_chapter,
                        section=current_section,
                        subsection="",
                        content_type=content_type,
                        is_table=has_table,
                        grade=grade,
                        subject=subject,
                        taxonomy_id=taxonomy_id,
                        metadata={"slide_number": slide_num, "slide_title": title},
                    )
                )

            # Extract speaker notes if available
            if slide.has_notes_slide:
                notes = slide.notes_slide.notes_text_frame.text.strip()
                if notes:
                    elements.append(
                        ParsedElement(
                            text=notes,
                            page_number=slide_num,
                            chapter=current_chapter,
                            section=current_section,
                            subsection="Speaker Notes",
                            content_type="prose",
                            grade=grade,
                            subject=subject,
                            taxonomy_id=taxonomy_id,
                            metadata={
                                "slide_number": slide_num,
                                "is_speaker_notes": True,
                            },
                        )
                    )

        return elements

    def _is_table_shape(self, shape) -> bool:
        """Check if shape is a table."""
        return shape.has_table if hasattr(shape, "has_table") else False
