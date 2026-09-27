"""
DOCX document parser using python-docx.
Extracts text with heading hierarchy detection.
"""

from pathlib import Path
from typing import List, Dict, Any
from docx import Document
from docx.oxml.ns import qn

from . import BaseParser, ParsedElement


class DOCXParser(BaseParser):
    """Parser for Microsoft Word DOCX documents."""

    def get_supported_extensions(self) -> List[str]:
        """Return supported file extensions."""
        return [".docx", ".doc"]

    def parse(self, file_path: str, **kwargs) -> List[ParsedElement]:
        """
        Parse a DOCX file and extract text with structure.

        Args:
            file_path: Path to DOCX file
            **kwargs: Additional arguments (grade, subject, taxonomy_id)

        Returns:
            List of ParsedElement objects
        """
        grade = kwargs.get("grade", 7)
        subject = kwargs.get("subject", "Science")
        taxonomy_id = kwargs.get("taxonomy_id", 0)

        if not Path(file_path).exists():
            raise FileNotFoundError(f"DOCX file not found: {file_path}")

        doc = Document(file_path)
        elements = []

        current_chapter = ""
        current_section = ""
        current_subsection = ""
        page_number = 1

        for para in doc.paragraphs:
            text = para.text.strip()
            if not text:
                continue

            # Detect heading level
            style_name = para.style.name if para.style else ""
            is_heading = False
            heading_level = 0

            if "Heading" in style_name:
                is_heading = True
                try:
                    heading_level = int(style_name.replace("Heading", "").strip())
                except ValueError:
                    heading_level = 1

            # Also check for outline level
            if para._element.xpath("./w:pPr/w:outlineLvl"):
                is_heading = True
                outline_val = para._element.xpath("./w:pPr/w:outlineLvl/@w:val")
                if outline_val:
                    try:
                        heading_level = int(outline_val[0]) + 1
                    except ValueError:
                        heading_level = 1

            # Update hierarchy based on heading level
            if is_heading:
                if heading_level == 1:
                    current_chapter = text
                    current_section = ""
                    current_subsection = ""
                elif heading_level == 2:
                    current_section = text
                    current_subsection = ""
                elif heading_level == 3:
                    current_subsection = text

                # Add heading as element
                elements.append(
                    ParsedElement(
                        text=text,
                        page_number=page_number,
                        chapter=current_chapter,
                        section=current_section,
                        subsection=current_subsection,
                        content_type="heading",
                        grade=grade,
                        subject=subject,
                        taxonomy_id=taxonomy_id,
                        metadata={"heading_level": heading_level},
                    )
                )
            else:
                # Regular paragraph
                # Check if it's a table
                is_table = self._is_table_paragraph(para)

                if is_table:
                    elements.append(
                        ParsedElement(
                            text=text,
                            page_number=page_number,
                            chapter=current_chapter,
                            section=current_section,
                            subsection=current_subsection,
                            content_type="table",
                            is_table=True,
                            grade=grade,
                            subject=subject,
                            taxonomy_id=taxonomy_id,
                        )
                    )
                else:
                    elements.append(
                        ParsedElement(
                            text=text,
                            page_number=page_number,
                            chapter=current_chapter,
                            section=current_section,
                            subsection=current_subsection,
                            content_type="prose",
                            grade=grade,
                            subject=subject,
                            taxonomy_id=taxonomy_id,
                        )
                    )

        # Also extract tables from document
        table_num = 0
        for table in doc.tables:
            table_num += 1
            table_text = self._extract_table_text(table)
            if table_text:
                elements.append(
                    ParsedElement(
                        text=table_text,
                        page_number=page_number,
                        chapter=current_chapter,
                        section=current_section,
                        subsection=current_subsection,
                        content_type="table",
                        is_table=True,
                        grade=grade,
                        subject=subject,
                        taxonomy_id=taxonomy_id,
                        metadata={"table_number": table_num},
                    )
                )

        return elements

    def _is_table_paragraph(self, para) -> bool:
        """Check if paragraph is part of a table."""
        # Check if paragraph is inside a table
        tbl = para._element.getparent()
        while tbl is not None:
            if tbl.tag.endswith("tbl"):
                return True
            tbl = tbl.getparent()
        return False

    def _extract_table_text(self, table) -> str:
        """Extract text from a DOCX table in markdown-like format."""
        rows = []
        for row in table.rows:
            cells = [cell.text.strip() for cell in row.cells]
            rows.append(" | ".join(cells))

        return "\n".join(rows)

    def _estimate_page_number(self, char_count: int, current_page: int) -> int:
        """Estimate page number based on character count."""
        # Rough estimate: ~2000 characters per page
        chars_per_page = 2000
        estimated_page = (char_count // chars_per_page) + 1
        return max(current_page, estimated_page)
