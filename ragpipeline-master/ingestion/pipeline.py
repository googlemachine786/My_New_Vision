"""
Visionary RAG Pipeline - Main Ingestion Orchestrator

End-to-end PDF ingestion pipeline:
1. Load PDF and calibrate font thresholds (Pass 1)
2. Extract tables with bounding boxes (Pass 2)
3. Map headings to hierarchy (Pass 3)
4. Detect formulas and equations (Pass 4)
5. Enrich metadata (Pass 5)
6. Create parent-child chunks
7. Extract keywords with YAKE
8. Embed chunks with Vertex AI
9. Write to AlloyDB with transaction support

Usage:
    python pipeline.py --pdf path/to/textbook.pdf --grade 7 --subject "Science" --taxonomy_id 42
"""

import asyncio
import argparse
import sys
from pathlib import Path
from typing import List, Dict, Any, Optional
from dataclasses import dataclass
import structlog

# Import parser modules
from parser import (
    calibrate_font_thresholds,
    extract_tables,
    map_headings,
    detect_formulas,
    enrich_metadata,
    create_parsed_elements_from_blocks,
    ParsedElement,
)
from parser.font_calibrator import FontThresholds
from parser.heading_mapper import HeadingContext
from parser.table_extractor import TableElement

# Import chunker
from chunker import create_parent_child_chunks, ParentChunk, ChildChunk

# Import keywords
from keywords import extract_keywords

# Import embedder
from embedder import embed_batch

# Import writer
from writer import AlloyDBWriter, WriterConfig, insert_parent_child_batch

logger = structlog.get_logger()


@dataclass
class PipelineConfig:
    """Pipeline configuration.
    
    Attributes:
        pdf_path: Path to PDF file
        grade: Grade level (6-8)
        subject: Subject name
        taxonomy_id: Reference to cbse_taxonomy
        project_id: GCP project ID
        location: GCP region
        alloydb_dsn: Database connection string
        batch_size: Embedding batch size
        concurrency: Concurrent DB inserts
    """
    pdf_path: str
    grade: int
    subject: str
    taxonomy_id: int
    project_id: str
    location: str
    alloydb_dsn: str
    batch_size: int = 5
    concurrency: int = 5
    
    def validate(self) -> bool:
        """Validate configuration."""
        if not Path(self.pdf_path).exists():
            return False
        if self.grade not in [6, 7, 8]:
            return False
        if self.taxonomy_id <= 0:
            return False
        return True


class IngestionPipeline:
    """Main ingestion pipeline orchestrator.
    
    Attributes:
        config: PipelineConfig instance
        stats: Pipeline execution statistics
    """
    
    def __init__(self, config: PipelineConfig):
        """Initialize pipeline.
        
        Args:
            config: PipelineConfig with all settings
        """
        self.config = config
        self.stats = {
            "pages_processed": 0,
            "elements_created": 0,
            "parent_chunks": 0,
            "child_chunks": 0,
            "keywords_extracted": 0,
            "embeddings_created": 0,
            "db_inserts_success": 0,
            "db_inserts_failure": 0,
        }
        
        logger.info("Pipeline initialized",
                   pdf=config.pdf_path,
                   grade=config.grade,
                   subject=config.subject,
                   taxonomy_id=config.taxonomy_id)
    
    async def run(self) -> Dict[str, Any]:
        """Run the complete ingestion pipeline.
        
        Returns:
            Dict with pipeline statistics
        """
        logger.info("Starting ingestion pipeline",
                   pdf=self.config.pdf_path)
        
        try:
            # Step 1: Load PDF and extract text blocks
            logger.info("Step 1: Loading PDF")
            import fitz
            import pdfplumber
            
            doc = fitz.open(self.config.pdf_path)
            
            with pdfplumber.open(self.config.pdf_path) as pdf:
                # Step 2: Calibrate font thresholds (Pass 1)
                logger.info("Step 2: Calibrating font thresholds")
                thresholds = calibrate_font_thresholds(doc, max_pages=20)
                
                # Step 3: Extract tables (Pass 2)
                logger.info("Step 3: Extracting tables")
                page_tables: Dict[int, List[TableElement]] = {}
                
                for page_num in range(len(pdf)):
                    page = pdf.pages[page_num]
                    tables = extract_tables(page, page_num + 1)
                    if tables:
                        page_tables[page_num + 1] = tables
                
                # Step 4: Map headings (Pass 3)
                logger.info("Step 4: Mapping headings")
                toc = doc.get_toc()
                heading_map = map_headings(doc, thresholds, toc)
                
                # Step 5: Extract text blocks with metadata
                logger.info("Step 5: Extracting text blocks")
                text_blocks = []
                
                for page_num in range(len(doc)):
                    page = doc[page_num]
                    text_dict = page.get_text("dict")
                    
                    for block in text_dict.get("blocks", []):
                        if block.get("type") != 0:
                            continue
                        
                        for line in block.get("lines", []):
                            for span in line.get("spans", []):
                                text = span.get("text", "").strip()
                                if len(text) < 10:
                                    continue
                                
                                text_blocks.append({
                                    "page": page_num + 1,
                                    "text": text,
                                    "font_size": span.get("size", 0),
                                    "flags": span.get("flags", 0),
                                })
                
                self.stats["pages_processed"] = len(doc)
                
                # Step 6: Create parsed elements (Pass 5)
                logger.info("Step 6: Creating parsed elements")
                elements = create_parsed_elements_from_blocks(
                    text_blocks=text_blocks,
                    heading_map=heading_map,
                    taxonomy_id=self.config.taxonomy_id,
                    grade=self.config.grade,
                    subject=self.config.subject,
                    page_tables=page_tables,
                )
                
                self.stats["elements_created"] = len(elements)
                
                # Step 7: Create parent-child chunks
                logger.info("Step 7: Creating parent-child chunks")
                parents, children = create_parent_child_chunks(elements)
                
                self.stats["parent_chunks"] = len(parents)
                self.stats["child_chunks"] = len(children)
                
                # Step 8: Extract keywords for parents
                logger.info("Step 8: Extracting keywords")
                for parent in parents:
                    keywords = extract_keywords(parent.content)
                    parent.extracted_keywords = keywords
                    self.stats["keywords_extracted"] += len(keywords)
                
                # Step 9: Embed children
                logger.info("Step 9: Creating embeddings")
                child_texts = [child.content for child in children]
                
                embeddings = embed_batch(
                    child_texts,
                    project_id=self.config.project_id,
                    location=self.config.location,
                )
                
                # Attach embeddings to children
                for child, embedding in zip(children, embeddings):
                    child.metadata["embedding"] = embedding
                
                self.stats["embeddings_created"] = len(embeddings)
                
                # Step 10: Write to AlloyDB
                logger.info("Step 10: Writing to database")
                await self._write_to_db(parents, children)
                
                logger.info("Pipeline complete", stats=self.stats)
                
                return self.stats
            
        finally:
            doc.close()
    
    async def _write_to_db(
        self,
        parents: List[ParentChunk],
        children: List[ChildChunk],
    ):
        """Write chunks to database.
        
        Args:
            parents: List of parent chunks
            children: List of child chunks
        """
        writer = AlloyDBWriter(
            WriterConfig(dsn=self.config.alloydb_dsn)
        )
        
        await writer.connect()
        
        try:
            # Group children by parent
            parent_children: Dict[str, List[ChildChunk]] = {}
            for child in children:
                if child.parent_id not in parent_children:
                    parent_children[child.parent_id] = []
                parent_children[child.parent_id].append(child)
            
            # Create parent-child pairs
            pairs = []
            for parent in parents:
                child_list = parent_children.get(parent.parent_id, [])
                pairs.append((parent, child_list))
            
            # Batch insert
            result = await insert_parent_child_batch(
                writer,
                pairs,
                concurrency=self.config.concurrency,
            )
            
            self.stats["db_inserts_success"] = result["success"]
            self.stats["db_inserts_failure"] = result["failure"]
            
        finally:
            await writer.close()


async def run_pipeline(config: PipelineConfig) -> Dict[str, Any]:
    """Run ingestion pipeline.
    
    Args:
        config: PipelineConfig
    
    Returns:
        Pipeline statistics
    """
    pipeline = IngestionPipeline(config)
    return await pipeline.run()


def main():
    """Main entry point."""
    parser = argparse.ArgumentParser(
        description="Visionary RAG Ingestion Pipeline"
    )
    
    parser.add_argument(
        "--pdf",
        required=True,
        help="Path to PDF file"
    )
    parser.add_argument(
        "--grade",
        type=int,
        required=True,
        choices=[6, 7, 8],
        help="Grade level (6-8)"
    )
    parser.add_argument(
        "--subject",
        required=True,
        help="Subject name (e.g., 'Science')"
    )
    parser.add_argument(
        "--taxonomy-id",
        type=int,
        required=True,
        help="Reference to cbse_taxonomy"
    )
    parser.add_argument(
        "--project-id",
        default="${GOOGLE_CLOUD_PROJECT}",
        help="GCP project ID"
    )
    parser.add_argument(
        "--location",
        default="asia-south1",
        help="GCP region"
    )
    parser.add_argument(
        "--alloydb-dsn",
        default="${ALLOYDB_DSN}",
        help="AlloyDB connection string"
    )
    parser.add_argument(
        "--batch-size",
        type=int,
        default=5,
        help="Embedding batch size"
    )
    parser.add_argument(
        "--concurrency",
        type=int,
        default=5,
        help="Concurrent DB inserts"
    )
    parser.add_argument(
        "--verbose",
        action="store_true",
        help="Enable verbose logging"
    )
    
    args = parser.parse_args()
    
    # Configure logging
    log_level = "DEBUG" if args.verbose else "INFO"
    structlog.configure(
        wrapper_class=structlog.make_filtering_bound_logger(
            getattr(structlog, log_level)
        ),
    )
    
    # Expand environment variables
    import os
    project_id = os.getenv("GOOGLE_CLOUD_PROJECT", args.project_id)
    alloydb_dsn = os.getenv("ALLOYDB_DSN", args.alloydb_dsn)
    
    # Create config
    config = PipelineConfig(
        pdf_path=args.pdf,
        grade=args.grade,
        subject=args.subject,
        taxonomy_id=args.taxonomy_id,
        project_id=project_id,
        location=args.location,
        alloydb_dsn=alloydb_dsn,
        batch_size=args.batch_size,
        concurrency=args.concurrency,
    )
    
    # Validate
    if not config.validate():
        logger.error("Invalid configuration")
        sys.exit(1)
    
    # Run pipeline
    try:
        stats = asyncio.run(run_pipeline(config))
        
        print("\n" + "="*50)
        print("Pipeline Complete")
        print("="*50)
        print(f"Pages processed: {stats['pages_processed']}")
        print(f"Elements created: {stats['elements_created']}")
        print(f"Parent chunks: {stats['parent_chunks']}")
        print(f"Child chunks: {stats['child_chunks']}")
        print(f"Keywords extracted: {stats['keywords_extracted']}")
        print(f"Embeddings created: {stats['embeddings_created']}")
        print(f"DB inserts success: {stats['db_inserts_success']}")
        print(f"DB inserts failure: {stats['db_inserts_failure']}")
        
    except Exception as e:
        logger.error("Pipeline failed", error=str(e))
        sys.exit(1)


if __name__ == "__main__":
    main()
