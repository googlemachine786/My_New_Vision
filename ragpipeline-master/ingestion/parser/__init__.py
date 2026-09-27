"""
Multi-format document parser factory and base classes.
Supports: PDF, DOCX, PPTX, Markdown, HTML
"""

from abc import ABC, abstractmethod
from pathlib import Path
from typing import List, Dict, Any
import json


class ParsedElement:
    """Represents a parsed document element."""

    def __init__(
        self,
        text: str,
        page_number: int,
        chapter: str = "",
        section: str = "",
        subsection: str = "",
        content_type: str = "prose",
        is_table: bool = False,
        grade: int = 7,
        subject: str = "Science",
        taxonomy_id: int = 0,
        metadata: Dict[str, Any] = None,
    ):
        self.text = text
        self.page_number = page_number
        self.chapter = chapter
        self.section = section
        self.subsection = subsection
        self.content_type = content_type
        self.is_table = is_table
        self.grade = grade
        self.subject = subject
        self.taxonomy_id = taxonomy_id
        self.metadata = metadata or {}

    def to_dict(self) -> Dict[str, Any]:
        """Convert to dictionary for JSON serialization."""
        return {
            "page_number": self.page_number,
            "text": self.text,
            "chapter": self.chapter,
            "section": self.section,
            "subsection": self.subsection,
            "content_type": self.content_type,
            "is_table": self.is_table,
            "grade": self.grade,
            "subject": self.subject,
            "taxonomy_id": self.taxonomy_id,
            "metadata": self.metadata,
        }


class BaseParser(ABC):
    """Abstract base class for document parsers."""

    @abstractmethod
    def parse(self, file_path: str, **kwargs) -> List[ParsedElement]:
        """Parse a document and return list of elements."""
        pass

    @abstractmethod
    def get_supported_extensions(self) -> List[str]:
        """Return list of supported file extensions."""
        pass


class ParserFactory:
    """Factory for creating document parsers."""

    _parsers = {}

    @classmethod
    def register_parser(cls, parser_class: BaseParser):
        """Register a parser class."""
        instance = parser_class()
        for ext in instance.get_supported_extensions():
            cls._parsers[ext.lower()] = instance

    @classmethod
    def get_parser(cls, file_path: str) -> BaseParser:
        """Get appropriate parser for file extension."""
        ext = Path(file_path).suffix.lower()
        if ext not in cls._parsers:
            raise ValueError(f"Unsupported file format: {ext}")
        return cls._parsers[ext]

    @classmethod
    def parse_file(
        cls, file_path: str, grade: int = 7, subject: str = "Science", taxonomy_id: int = 0
    ) -> List[ParsedElement]:
        """Parse a file using the appropriate parser."""
        parser = cls.get_parser(file_path)
        return parser.parse(file_path, grade=grade, subject=subject, taxonomy_id=taxonomy_id)


# Import and register parsers
from .pdf_parser import PDFParser
from .docx_parser import DOCXParser
from .pptx_parser import PPTXParser
from .markdown_parser import MarkdownParser
from .html_parser import HTMLParser

ParserFactory.register_parser(PDFParser)
ParserFactory.register_parser(DOCXParser)
ParserFactory.register_parser(PPTXParser)
ParserFactory.register_parser(MarkdownParser)
ParserFactory.register_parser(HTMLParser)


def parse_document(
    file_path: str,
    grade: int = 7,
    subject: str = "Science",
    taxonomy_id: int = 0,
    output_format: str = "json",
) -> str:
    """
    Parse a document and return JSON output.

    Args:
        file_path: Path to the document
        grade: Grade level (6-8)
        subject: Subject name
        taxonomy_id: CBSE taxonomy ID
        output_format: Output format (json or dict)

    Returns:
        JSON string or dictionary of parsed elements
    """
    elements = ParserFactory.parse_file(file_path, grade, subject, taxonomy_id)

    result = [elem.to_dict() for elem in elements]

    if output_format == "json":
        return json.dumps(result, indent=2)
    return result
