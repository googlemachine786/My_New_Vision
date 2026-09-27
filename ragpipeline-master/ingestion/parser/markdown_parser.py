"""
Markdown parser using mistune.
Extracts text with heading hierarchy and code blocks.
"""

from pathlib import Path
from typing import List, Dict, Any
import re

from . import BaseParser, ParsedElement


class MarkdownParser(BaseParser):
    """Parser for Markdown documents."""

    def get_supported_extensions(self) -> List[str]:
        """Return supported file extensions."""
        return [".md", ".markdown", ".mdown"]

    def parse(self, file_path: str, **kwargs) -> List[ParsedElement]:
        """
        Parse a Markdown file and extract structured content.

        Args:
            file_path: Path to Markdown file
            **kwargs: Additional arguments (grade, subject, taxonomy_id)

        Returns:
            List of ParsedElement objects
        """
        grade = kwargs.get("grade", 7)
        subject = kwargs.get("subject", "Science")
        taxonomy_id = kwargs.get("taxonomy_id", 0)

        if not Path(file_path).exists():
            raise FileNotFoundError(f"Markdown file not found: {file_path}")

        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read()

        elements = []
        current_chapter = ""
        current_section = ""
        current_subsection = ""
        page_number = 1

        # Split by lines
        lines = content.split("\n")
        current_block = []
        block_type = "prose"

        for line in lines:
            # Check for headings
            heading_match = re.match(r"^(#{1,6})\s+(.+)$", line)
            if heading_match:
                # Save previous block
                if current_block:
                    elements.extend(
                        self._create_block_elements(
                            current_block,
                            page_number,
                            current_chapter,
                            current_section,
                            current_subsection,
                            block_type,
                            grade,
                            subject,
                            taxonomy_id,
                        )
                    )
                    current_block = []

                # Process heading
                level = len(heading_match.group(1))
                text = heading_match.group(2).strip()

                if level == 1:
                    current_chapter = text
                    current_section = ""
                    current_subsection = ""
                elif level == 2:
                    current_section = text
                    current_subsection = ""
                elif level == 3:
                    current_subsection = text

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
                        metadata={"heading_level": level, "markdown_heading": True},
                    )
                )
                block_type = "prose"
                continue

            # Check for code blocks
            if line.strip().startswith("```"):
                if block_type == "code":
                    # End of code block
                    elements.extend(
                        self._create_block_elements(
                            current_block,
                            page_number,
                            current_chapter,
                            current_section,
                            current_subsection,
                            "code",
                            grade,
                            subject,
                            taxonomy_id,
                        )
                    )
                    current_block = []
                    block_type = "prose"
                else:
                    # Start of code block
                    if current_block:
                        elements.extend(
                            self._create_block_elements(
                                current_block,
                                page_number,
                                current_chapter,
                                current_section,
                                current_subsection,
                                "prose",
                                grade,
                                subject,
                                taxonomy_id,
                            )
                        )
                    current_block = []
                    block_type = "code"
                continue

            # Check for tables (markdown tables)
            if re.match(r"^\|.*\|", line):
                if block_type != "table":
                    if current_block and block_type != "prose":
                        elements.extend(
                            self._create_block_elements(
                                current_block,
                                page_number,
                                current_chapter,
                                current_section,
                                current_subsection,
                                block_type,
                                grade,
                                subject,
                                taxonomy_id,
                            )
                        )
                    current_block = []
                    block_type = "table"
                current_block.append(line)
                continue

            # Regular content
            if line.strip():
                current_block.append(line)
            else:
                # Empty line - save block
                if current_block:
                    elements.extend(
                        self._create_block_elements(
                            current_block,
                            page_number,
                            current_chapter,
                            current_section,
                            current_subsection,
                            block_type,
                            grade,
                            subject,
                            taxonomy_id,
                        )
                    )
                    current_block = []
                    block_type = "prose"

        # Don't forget last block
        if current_block:
            elements.extend(
                self._create_block_elements(
                    current_block,
                    page_number,
                    current_chapter,
                    current_section,
                    current_subsection,
                    block_type,
                    grade,
                    subject,
                    taxonomy_id,
                )
            )

        return elements

    def _create_block_elements(
        self,
        block: List[str],
        page_number: int,
        chapter: str,
        section: str,
        subsection: str,
        block_type: str,
        grade: int,
        subject: str,
        taxonomy_id: int,
    ) -> List[ParsedElement]:
        """Create ParsedElement from a block of text."""
        text = "\n".join(block)

        if not text.strip():
            return []

        is_table = block_type == "table"
        content_type = "table" if is_table else "code" if block_type == "code" else "prose"

        return [
            ParsedElement(
                text=text,
                page_number=page_number,
                chapter=chapter,
                section=section,
                subsection=subsection,
                content_type=content_type,
                is_table=is_table,
                grade=grade,
                subject=subject,
                taxonomy_id=taxonomy_id,
                metadata={"block_type": block_type},
            )
        ]
