"""
Visionary RAG - Metadata Extraction from Textbooks
Extracts chapter, section, page numbers, and other metadata for filtering
"""

import fitz
import json
import re
from pathlib import Path
from typing import List, Dict
from dataclasses import dataclass, asdict
from datetime import datetime

@dataclass
class TextbookMetadata:
    """Extracted textbook metadata."""
    chapter: str = ""
    section: str = ""
    subsection: str = ""
    page_number: int = 0
    grade: int = 0
    subject: str = ""
    topic: str = ""
    content_type: str = "text"  # text, table, figure, formula
    taxonomy_id: int = 0

class MetadataExtractor:
    """Extracts metadata from textbook PDFs."""
    
    def __init__(self):
        self.toc = []  # Table of contents
        self.headings = []
    
    def extract_all(self, pdf_path: str) -> List[Dict]:
        """Extract all metadata from PDF."""
        print(f"\nExtracting metadata from {pdf_path}...")
        
        doc = fitz.open(pdf_path)
        
        # 1. Extract Table of Contents
        self.toc = doc.get_toc()
        print(f"  Found {len(self.toc)} TOC entries")
        
        # 2. Extract headings from each page
        all_metadata = []
        
        for page_num in range(len(doc)):
            page = doc[page_num]
            
            # Extract text with position info
            blocks = page.get_text("dict")["blocks"]
            
            for block in blocks:
                if block["type"] != 0:  # Not text block
                    continue
                
                # Extract metadata from block
                metadata = self._extract_block_metadata(block, page_num + 1, doc)
                if metadata:
                    all_metadata.append(asdict(metadata))
        
        doc.close()
        
        print(f"  Extracted metadata for {len(all_metadata)} blocks")
        
        return all_metadata
    
    def _extract_block_metadata(self, block: Dict, page_num: int, doc) -> TextbookMetadata:
        """Extract metadata from a text block."""
        metadata = TextbookMetadata()
        metadata.page_number = page_num
        
        # Get text
        text = ""
        font_size = 0
        is_bold = False
        
        for line in block.get("lines", []):
            for span in line.get("spans", []):
                text += span.get("text", "")
                font_size = max(font_size, span.get("size", 0))
                flags = span.get("flags", 0)
                is_bold = is_bold or (flags & 2**4)  # Bold flag
        
        text = text.strip()
        if len(text) < 20:
            return None
        
        # Determine if heading based on font size
        if font_size > 14:
            metadata.chapter = text
            metadata.content_type = "heading"
        elif font_size > 12 and is_bold:
            metadata.section = text
            metadata.content_type = "subheading"
        else:
            metadata.content_type = "text"
        
        # Match with TOC
        if self.toc:
            for toc_entry in self.toc:
                level, title, toc_page = toc_entry[:3]
                if abs(page_num - toc_page) <= 2:
                    if level == 1:
                        metadata.chapter = title
                    elif level == 2:
                        metadata.section = title
                    elif level == 3:
                        metadata.subsection = title
        
        # Detect grade and subject
        text_lower = text.lower()
        if "class" in text_lower or "grade" in text_lower:
            # Extract grade number
            match = re.search(r'(class|grade)\s*(\d+)', text_lower)
            if match:
                metadata.grade = int(match.group(2))
        
        if "science" in text_lower:
            metadata.subject = "Science"
        elif "math" in text_lower or "mathematics" in text_lower:
            metadata.subject = "Mathematics"
        
        # Detect content type
        if "|" in text and "-" in text:
            metadata.content_type = "table"
        elif "fig" in text_lower or "figure" in text_lower:
            metadata.content_type = "figure"
        elif re.search(r'\d+\s*[+\-×÷=]', text):
            metadata.content_type = "formula"
        
        return metadata
    
    def create_taxonomy_mapping(self, metadata_list: List[Dict]) -> Dict:
        """Create taxonomy ID mapping."""
        taxonomy = {}
        current_id = 1
        
        for meta in metadata_list:
            key = f"{meta['grade']}_{meta['subject']}_{meta['chapter']}"
            if key and key not in taxonomy:
                taxonomy[key] = {
                    "taxonomy_id": current_id,
                    "grade": meta['grade'],
                    "subject": meta['subject'],
                    "chapter": meta['chapter'],
                    "section": meta['section']
                }
                current_id += 1
        
        return taxonomy

def run_metadata_extraction():
    """Run metadata extraction."""
    print("\n" + "="*70)
    print("TEXTBOOK METADATA EXTRACTION")
    print("="*70)
    
    pdf_path = Path("science class 8.pdf")
    if not pdf_path.exists():
        print("❌ PDF not found")
        return
    
    extractor = MetadataExtractor()
    metadata_list = extractor.extract_all(str(pdf_path))
    
    # Create taxonomy
    taxonomy = extractor.create_taxonomy_mapping(metadata_list)
    
    # Save results
    results = {
        "extraction_date": datetime.now().isoformat(),
        "pdf_file": str(pdf_path),
        "total_metadata_entries": len(metadata_list),
        "taxonomy_entries": len(taxonomy),
        "metadata_samples": metadata_list[:20],  # Save first 20
        "taxonomy": taxonomy
    }
    
    output_path = Path("data/textbook_metadata.json")
    with open(output_path, 'w') as f:
        json.dump(results, f, indent=2)
    
    print(f"\n✅ Metadata saved to: {output_path}")
    
    # Print summary
    print("\n" + "="*70)
    print("METADATA SUMMARY")
    print("="*70)
    
    # Count by content type
    content_types = {}
    for meta in metadata_list:
        ct = meta['content_type']
        content_types[ct] = content_types.get(ct, 0) + 1
    
    print(f"\nContent Types:")
    for ct, count in sorted(content_types.items()):
        print(f"  {ct}: {count}")
    
    # Count by grade
    grades = {}
    for meta in metadata_list:
        grade = meta['grade']
        if grade > 0:
            grades[grade] = grades.get(grade, 0) + 1
    
    if grades:
        print(f"\nBy Grade:")
        for grade, count in sorted(grades.items()):
            print(f"  Grade {grade}: {count}")
    
    # Count by subject
    subjects = {}
    for meta in metadata_list:
        subject = meta['subject']
        if subject:
            subjects[subject] = subjects.get(subject, 0) + 1
    
    if subjects:
        print(f"\nBy Subject:")
        for subject, count in sorted(subjects.items()):
            print(f"  {subject}: {count}")
    
    print(f"\nTaxonomy IDs created: {len(taxonomy)}")
    
    return results

if __name__ == "__main__":
    run_metadata_extraction()
