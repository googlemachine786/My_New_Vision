"""
HTML parser using BeautifulSoup.
Extracts text content while removing boilerplate.
"""

from pathlib import Path
from typing import List, Dict, Any
from bs4 import BeautifulSoup, NavigableString, Tag

from . import BaseParser, ParsedElement


class HTMLParser(BaseParser):
    """Parser for HTML documents."""

    def get_supported_extensions(self) -> List[str]:
        """Return supported file extensions."""
        return [".html", ".htm"]

    def parse(self, file_path: str, **kwargs) -> List[ParsedElement]:
        """
        Parse an HTML file and extract content.

        Args:
            file_path: Path to HTML file
            **kwargs: Additional arguments (grade, subject, taxonomy_id)

        Returns:
            List of ParsedElement objects
        """
        grade = kwargs.get("grade", 7)
        subject = kwargs.get("subject", "Science")
        taxonomy_id = kwargs.get("taxonomy_id", 0)

        if not Path(file_path).exists():
            raise FileNotFoundError(f"HTML file not found: {file_path}")

        with open(file_path, "r", encoding="utf-8") as f:
            content = f.read()

        soup = BeautifulSoup(content, "html.parser")

        # Remove script and style elements
        for tag in soup(["script", "style", "noscript", "iframe", "svg"]):
            tag.decompose()

        elements = []
        page_number = 1
        current_chapter = ""
        current_section = ""

        # Try to extract title
        title_tag = soup.find("title")
        if title_tag and title_tag.string:
            current_chapter = title_tag.string.strip()
            elements.append(
                ParsedElement(
                    text=current_chapter,
                    page_number=page_number,
                    chapter=current_chapter,
                    section="",
                    subsection="",
                    content_type="heading",
                    grade=grade,
                    subject=subject,
                    taxonomy_id=taxonomy_id,
                    metadata={"source": "title"},
                )
            )

        # Extract headings
        for heading_level in range(1, 7):
            heading_tags = soup.find_all(f"h{heading_level}")
            for tag in heading_tags:
                text = tag.get_text().strip()
                if not text:
                    continue

                if heading_level == 1:
                    current_chapter = text
                    current_section = ""
                elif heading_level == 2:
                    current_section = text

                elements.append(
                    ParsedElement(
                        text=text,
                        page_number=page_number,
                        chapter=current_chapter,
                        section=current_section,
                        subsection="",
                        content_type="heading",
                        grade=grade,
                        subject=subject,
                        taxonomy_id=taxonomy_id,
                        metadata={"heading_level": heading_level},
                    )
                )

        # Extract main content sections
        # Look for semantic HTML5 elements
        main_content = soup.find("main") or soup.find("article") or soup.find("div", class_="content")

        if not main_content:
            main_content = soup.find("body") or soup

        # Extract paragraphs and other content
        for tag in main_content.find_all(["p", "div", "li", "td", "th"], recursive=True):
            text = tag.get_text().strip()
            if not text or len(text) < 10:  # Skip very short text
                continue

            # Skip if it's a heading (already processed)
            if tag.name.startswith("h") and tag.name[1:].isdigit():
                continue

            # Determine content type
            is_table = tag.name in ["td", "th"] or (
                tag.parent and tag.parent.name == "table"
            )

            # Skip boilerplate
            if self._is_boilerplate(tag):
                continue

            elements.append(
                ParsedElement(
                    text=text,
                    page_number=page_number,
                    chapter=current_chapter,
                    section=current_section,
                    subsection="",
                    content_type="table" if is_table else "prose",
                    is_table=is_table,
                    grade=grade,
                    subject=subject,
                    taxonomy_id=taxonomy_id,
                    metadata={"html_tag": tag.name},
                )
            )

        # Extract tables
        tables = soup.find_all("table")
        for table_num, table in enumerate(tables, 1):
            table_text = self._extract_table_text(table)
            if table_text:
                elements.append(
                    ParsedElement(
                        text=table_text,
                        page_number=page_number,
                        chapter=current_chapter,
                        section=current_section,
                        subsection="",
                        content_type="table",
                        is_table=True,
                        grade=grade,
                        subject=subject,
                        taxonomy_id=taxonomy_id,
                        metadata={"table_number": table_num, "source": "html_table"},
                    )
                )

        return elements

    def _is_boilerplate(self, tag: Tag) -> bool:
        """Check if tag is likely boilerplate content."""
        # Check for common boilerplate classes/IDs
        boilerplate_indicators = [
            "footer",
            "header",
            "nav",
            "sidebar",
            "advertisement",
            "ad-",
            "cookie",
            "newsletter",
            "subscribe",
        ]

        # Check class and id attributes
        for attr in ["class", "id"]:
            if tag.has_attr(attr):
                attr_value = tag[attr]
                if isinstance(attr_value, list):
                    attr_value = " ".join(attr_value)
                attr_value = attr_value.lower()

                for indicator in boilerplate_indicators:
                    if indicator in attr_value:
                        return True

        return False

    def _extract_table_text(self, table: Tag) -> str:
        """Extract text from HTML table."""
        rows = []
        for row in table.find_all("tr"):
            cells = []
            for cell in row.find_all(["td", "th"]):
                cells.append(cell.get_text().strip())
            if cells:
                rows.append(" | ".join(cells))
        return "\n".join(rows)
