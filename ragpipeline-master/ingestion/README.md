# Visionary RAG Pipeline - Ingestion Module

Python-based PDF ingestion pipeline for CBSE Science textbooks.

## Overview

This module processes PDF textbooks and converts them into searchable chunks with embeddings for the RAG pipeline.

## Architecture

### 5-Pass PDF Parser

1. **Pass 1: Font Calibrator** - Detects font size thresholds for heading detection
2. **Pass 2: Table Extractor** - Extracts tables with bounding boxes
3. **Pass 3: Heading Mapper** - Maps headings to chapter/section/subsection hierarchy
4. **Pass 4: Formula Detector** - Identifies chemical formulas and equations
5. **Pass 5: Metadata Enricher** - Combines all passes into structured elements

### Chunking Strategy

- **Parent chunks**: 1500 characters max, 100 char overlap
- **Child chunks**: 512 characters max, 77 char overlap (15%)
- **Tables**: Atomic (parent_id == child_id, no splitting)

### Keyword Extraction

- YAKE with CBSE-optimized parameters:
  - `n=2` (bigrams only)
  - `dedupLim=0.7` (aggressive deduplication)
  - `top=8` (top 8 keywords)

### Embeddings

- Vertex AI text-embedding-005
- Task type: `RETRIEVAL_DOCUMENT`
- Output: 768 dimensions
- Batch size: ≤5 texts per API call

## Installation

```bash
# Install with development dependencies
pip install -e ".[dev]"

# Or just production dependencies
pip install -e .
```

## Usage

### Command Line

```bash
python pipeline.py \
  --pdf path/to/textbook.pdf \
  --grade 7 \
  --subject "Science" \
  --taxonomy-id 42 \
  --project-id my-gcp-project \
  --location asia-south1 \
  --alloydb-dsn "host=IP dbname=visionary user=visionary password=XXX"
```

### Environment Variables

```bash
export GOOGLE_CLOUD_PROJECT=my-gcp-project
export ALLOYDB_DSN="host=IP dbname=visionary user=visionary password=XXX"
```

### Python API

```python
from ingestion.pipeline import IngestionPipeline, PipelineConfig
import asyncio

config = PipelineConfig(
    pdf_path="textbook.pdf",
    grade=7,
    subject="Science",
    taxonomy_id=42,
    project_id="my-gcp-project",
    location="asia-south1",
    alloydb_dsn="host=IP dbname=visionary user=visionary password=XXX",
)

pipeline = IngestionPipeline(config)
stats = asyncio.run(pipeline.run())

print(f"Processed {stats['pages_processed']} pages")
print(f"Created {stats['child_chunks']} child chunks")
```

## Module Structure

```
ingestion/
├── parser/              # 5-pass PDF parser
│   ├── __init__.py
│   ├── font_calibrator.py    # Pass 1
│   ├── table_extractor.py    # Pass 2
│   ├── heading_mapper.py     # Pass 3
│   ├── formula_detector.py   # Pass 4
│   └── metadata_enricher.py  # Pass 5
├── chunker/             # Parent-child chunking
│   ├── __init__.py
│   └── parent_child.py
├── keywords/            # YAKE keyword extraction
│   ├── __init__.py
│   └── yake_extractor.py
├── embedder/            # Vertex AI embeddings
│   ├── __init__.py
│   └── vertex_batch.py
├── writer/              # AlloyDB writer
│   ├── __init__.py
│   └── alloydb_writer.py
├── pipeline.py          # Main orchestrator
├── pyproject.toml
├── Dockerfile
└── tests/
    ├── test_parser.py
    ├── test_chunker.py
    └── ...
```

## Testing

```bash
# Run all tests
pytest

# Run with coverage
pytest --cov=ingestion --cov-report=term-missing

# Run specific test file
pytest tests/test_parser.py -v

# Run specific test
pytest tests/test_parser.py::TestFormulaDetector::test_formula_detection_CO2 -v
```

## Docker

```bash
# Build image
docker build -t visionary-ingestion .

# Run pipeline
docker run -e GOOGLE_CLOUD_PROJECT=my-project \
           -e ALLOYDB_DSN="host=IP dbname=visionary ..." \
           -v ./pdfs:/app/pdfs \
           visionary-ingestion \
           python pipeline.py --pdf /app/pdfs/textbook.pdf --grade 7 --subject Science --taxonomy-id 42
```

## Deploy to Cloud Run Jobs

```bash
# Build and push
docker build -t asia-south1-docker.pkg.dev/my-project/visionary-rag-images/ingestion:latest .
docker push asia-south1-docker.pkg.dev/my-project/visionary-rag-images/ingestion:latest

# Deploy as Cloud Run Job
gcloud run jobs deploy visionary-rag-ingestion \
  --image asia-south1-docker.pkg.dev/my-project/visionary-rag-images/ingestion:latest \
  --region asia-south1 \
  --task-timeout 3600s \
  --set-env-vars GOOGLE_CLOUD_PROJECT=my-project,ALLOYDB_DSN=host=IP...
```

## Performance

| Metric | Target | Notes |
|--------|--------|-------|
| Pages/minute | ~50 | Depends on PDF complexity |
| Embedding batch latency | ~500ms | 5 texts per batch |
| DB insert latency | ~50ms | Per parent-child pair |

## Error Handling

- **Transaction rollback**: If child insert fails, parent is rolled back
- **DLQ routing**: Failed inserts go to `ingestion_dlq` table
- **Retry logic**: Vertex AI API calls retry with exponential backoff (max 5 attempts)

## Output Schema

### parent_chunks

| Column | Type | Description |
|--------|------|-------------|
| parent_id | UUID | Primary key |
| taxonomy_id | INT | FK to cbse_taxonomy |
| content | TEXT | Chunk content (≤1500 chars) |
| extracted_keywords | TEXT[] | Keywords from YAKE |
| page_number | INT | Source page |
| chapter | VARCHAR | Chapter title |
| section | VARCHAR | Section title |
| content_type | VARCHAR | prose/table/formula/equation |

### child_chunks

| Column | Type | Description |
|--------|------|-------------|
| child_id | UUID | Primary key |
| parent_id | UUID | FK to parent_chunks |
| taxonomy_id | INT | FK to cbse_taxonomy |
| content | TEXT | Chunk content (≤512 chars) |
| embedding | VECTOR(768) | Vertex AI embedding |
| page_number | INT | Source page |
| content_type | VARCHAR | prose/table/formula/equation |

## License

Proprietary - Visionary Education Technologies
