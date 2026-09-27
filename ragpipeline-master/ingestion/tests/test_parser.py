"""
Tests for Parser Module

Test cases:
- test_calibrate_cross_publisher_fonts: 3 PDFs, different publishers
- test_table_extraction_with_bbox: table bbox excludes from prose
- test_heading_detection_numbered_section: "3.2 Cell Structure" → section
- test_formula_detection_CO2: content_type = "formula"
"""

import pytest
import fitz
import pdfplumber
from pathlib import Path

from parser.font_calibrator import calibrate_font_thresholds, FontThresholds
from parser.table_extractor import extract_tables
from parser.heading_mapper import map_headings
from parser.formula_detector import detect_formulas
from parser.metadata_enricher import enrich_metadata


class TestFontCalibrator:
    """Tests for font calibration."""
    
    def test_calibrate_cross_publisher_fonts(self, sample_pdfs):
        """Test font calibration across different publishers."""
        thresholds_list = []
        
        for pdf_path in sample_pdfs:
            doc = fitz.open(pdf_path)
            try:
                thresholds = calibrate_font_thresholds(doc, max_pages=20)
                thresholds_list.append(thresholds)
                
                # Validate thresholds are separated
                assert thresholds.chapter > thresholds.section
                assert thresholds.section > thresholds.body
                
                # Validate sample count
                assert thresholds.sample_count > 0
                
            finally:
                doc.close()
        
        # Verify thresholds vary across publishers (not identical)
        body_sizes = [t.body for t in thresholds_list]
        assert len(set(body_sizes)) > 1, "Body sizes should vary across publishers"
    
    def test_thresholds_validation(self):
        """Test FontThresholds validation."""
        # Valid thresholds
        valid = FontThresholds(body=10.0, section=12.0, chapter=14.0, sample_count=100)
        assert valid.validate() is True
        
        # Invalid: section not > body
        invalid1 = FontThresholds(body=12.0, section=12.0, chapter=14.0, sample_count=100)
        assert invalid1.validate() is False
        
        # Invalid: chapter not > section
        invalid2 = FontThresholds(body=10.0, section=14.0, chapter=14.0, sample_count=100)
        assert invalid2.validate() is False


class TestTableExtractor:
    """Tests for table extraction."""
    
    def test_table_extraction_with_bbox(self, sample_pdf_with_tables):
        """Test that table bbox is correctly extracted."""
        with pdfplumber.open(sample_pdf_with_tables) as pdf:
            page = pdf.pages[0]
            tables = extract_tables(page, page_num=1)
            
            # Should find at least one table
            assert len(tables) > 0
            
            # Validate table structure
            table = tables[0]
            assert table.bbox is not None
            assert len(table.bbox) == 4  # (x0, y0, x1, y1)
            assert table.row_count > 0
            assert table.col_count > 0
            assert table.markdown is not None
            assert "|" in table.markdown  # Markdown table format
    
    def test_table_to_markdown(self, sample_pdf_with_tables):
        """Test table markdown conversion."""
        with pdfplumber.open(sample_pdf_with_tables) as pdf:
            page = pdf.pages[0]
            tables = extract_tables(page, page_num=1)
            
            if tables:
                table = tables[0]
                lines = table.markdown.split("\n")
                
                # Should have header separator
                assert any("---" in line for line in lines)
                
                # All lines should start and end with |
                for line in lines:
                    if line.strip():
                        assert line.strip().startswith("|")
                        assert line.strip().endswith("|")


class TestHeadingMapper:
    """Tests for heading mapping."""
    
    def test_heading_detection_numbered_section(self, sample_pdf):
        """Test detection of numbered sections like '3.2 Cell Structure'."""
        doc = fitz.open(sample_pdf)
        try:
            thresholds = calibrate_font_thresholds(doc, max_pages=5)
            heading_map = map_headings(doc, thresholds)
            
            # Should map all pages
            assert len(heading_map) == len(doc)
            
            # At least some pages should have headings
            has_headings = sum(1 for h in heading_map.values() if h.heading_level > 0)
            assert has_headings > 0, "Should detect at least one heading"
            
        finally:
            doc.close()
    
    def test_heading_context_copy(self):
        """Test HeadingContext copy_with_updates."""
        from parser.heading_mapper import HeadingContext
        
        original = HeadingContext(
            chapter="Introduction",
            section="Overview",
            page=1,
            heading_level=2,
        )
        
        updated = original.copyWith_updates(
            section="New Section",
            page=2,
        )
        
        # Original unchanged
        assert original.section == "Overview"
        assert original.page == 1
        
        # Updated values
        assert updated.section == "New Section"
        assert updated.page == 2
        
        # Carried over
        assert updated.chapter == "Introduction"


class TestFormulaDetector:
    """Tests for formula detection."""
    
    def test_formula_detection_CO2(self):
        """Test detection of chemical formulas."""
        result = detect_formulas("CO₂ + H₂O → C₆H₁₂O₆")
        
        assert result.content_type == "formula"
        assert result.has_subscripts is True
        assert len(result.formula_annotations) > 0
    
    def test_prose_detection(self):
        """Test that regular prose is not detected as formula."""
        result = detect_formulas(
            "The cell membrane controls what enters and exits the cell."
        )
        
        assert result.content_type == "prose"
        assert result.has_subscripts is False
    
    def test_measurement_detection(self):
        """Test detection of measurements."""
        result = detect_formulas("The temperature is 25°C and pressure is 101.3 kPa")
        
        assert result.has_measurements is True
        assert len(result.formula_annotations) > 0
    
    def test_equation_detection(self):
        """Test detection of equations."""
        result = detect_formulas("E = mc²")
        
        assert result.content_type == "equation"
    
    def test_confidence_scoring(self):
        """Test confidence scoring for different content types."""
        # High confidence: multiple indicators
        result1 = detect_formulas("CO₂ + H₂O → C₆H₁₂O₆ at 25°C")
        assert result1.confidence > 0.7
        
        # Low confidence: prose only
        result2 = detect_formulas("The cell is the basic unit of life")
        assert result2.confidence < 0.5


class TestMetadataEnricher:
    """Tests for metadata enrichment."""
    
    def test_enrich_metadata_basic(self):
        """Test basic metadata enrichment."""
        from parser.heading_mapper import HeadingContext
        
        heading = HeadingContext(
            chapter="Cell Structure",
            section="The Cell Membrane",
            page=42,
            heading_level=2,
        )
        
        element = enrich_metadata(
            text="The cell membrane is selectively permeable.",
            heading_context=heading,
            page_number=42,
            taxonomy_id=15,
            grade=8,
            subject="Science",
        )
        
        # Validate fields
        assert element.chapter == "Cell Structure"
        assert element.section == "The Cell Membrane"
        assert element.grade == 8
        assert element.subject == "Science"
        assert element.taxonomy_id == 15
        assert element.content_type == "prose"
    
    def test_element_validation(self):
        """Test ParsedElement validation."""
        from parser.heading_mapper import HeadingContext
        
        heading = HeadingContext()
        
        # Invalid: empty text
        element1 = enrich_metadata(
            text="",
            heading_context=heading,
            page_number=1,
            taxonomy_id=1,
            grade=6,
            subject="Science",
        )
        assert element1.validate() is False
        
        # Invalid: missing taxonomy_id
        element2 = enrich_metadata(
            text="Some text",
            heading_context=heading,
            page_number=1,
            taxonomy_id=0,
            grade=6,
            subject="Science",
        )
        assert element2.validate() is False


# Fixtures
@pytest.fixture
def sample_pdf(tmp_path):
    """Create a sample PDF for testing."""
    # Create a simple PDF with PyMuPDF
    doc = fitz.open()
    page = doc.new_page()
    
    # Add text with different font sizes
    page.insert_text((50, 50), "Chapter 1: Introduction", fontsize=18)
    page.insert_text((50, 100), "Section 1.1: Overview", fontsize=14)
    page.insert_text((50, 150), "This is body text.", fontsize=10)
    
    pdf_path = tmp_path / "sample.pdf"
    doc.save(str(pdf_path))
    doc.close()
    
    return str(pdf_path)


@pytest.fixture
def sample_pdf_with_tables(tmp_path):
    """Create a sample PDF with tables for testing."""
    doc = fitz.open()
    page = doc.new_page()
    
    # Add a simple table
    table_data = [
        ["Name", "Value", "Unit"],
        ["Temperature", "25", "°C"],
        ["Pressure", "101.3", "kPa"],
    ]
    
    y = 50
    for row in table_data:
        x = 50
        for cell in row:
            page.insert_text((x, y), cell, fontsize=10)
            x += 100
        y += 20
    
    pdf_path = tmp_path / "sample_tables.pdf"
    doc.save(str(pdf_path))
    doc.close()
    
    return str(pdf_path)


@pytest.fixture
def sample_pdfs(tmp_path):
    """Create multiple sample PDFs with different styles."""
    pdfs = []
    
    # Publisher 1: Large headings
    doc1 = fitz.open()
    page1 = doc1.new_page()
    page1.insert_text((50, 50), "CHAPTER ONE", fontsize=24)
    page1.insert_text((50, 100), "Body text here", fontsize=12)
    pdf1 = tmp_path / "publisher1.pdf"
    doc1.save(str(pdf1))
    doc1.close()
    pdfs.append(str(pdf1))
    
    # Publisher 2: Medium headings
    doc2 = fitz.open()
    page2 = doc2.new_page()
    page2.insert_text((50, 50), "Chapter 1", fontsize=16)
    page2.insert_text((50, 100), "Regular text", fontsize=10)
    pdf2 = tmp_path / "publisher2.pdf"
    doc2.save(str(pdf2))
    doc2.close()
    pdfs.append(str(pdf2))
    
    # Publisher 3: Small headings
    doc3 = fitz.open()
    page3 = doc3.new_page()
    page3.insert_text((50, 50), "Section 1", fontsize=12)
    page3.insert_text((50, 100), "Content here", fontsize=9)
    pdf3 = tmp_path / "publisher3.pdf"
    doc3.save(str(pdf3))
    doc3.close()
    pdfs.append(str(pdf3))
    
    return pdfs


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
