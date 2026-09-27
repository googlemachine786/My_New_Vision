# Visionary RAG Pipeline - Ingestion Service
# Python 3.11 with PyMuPDF, Vertex AI, psycopg3

FROM python:3.11-slim

# Set working directory
WORKDIR /app

# Install system dependencies
# - libmupdf-dev: PyMuPDF native library
# - gcc: C compiler for Python extensions
# - libpq-dev: PostgreSQL client library for psycopg
RUN apt-get update && apt-get install -y \
    libmupdf-dev \
    gcc \
    libpq-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy pyproject.toml first for better caching
COPY pyproject.toml .

# Install Python dependencies
RUN pip install --no-cache-dir .

# Copy source code
COPY . .

# Set environment variables
ENV PYTHONUNBUFFERED=1
ENV PYTHONDONTWRITEBYTECODE=1

# Default command (can be overridden)
CMD ["python", "pipeline.py"]

# Health check (checks if Python can import the pipeline)
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD python -c "import pipeline" || exit 1

# Labels
LABEL maintainer="Visionary Engineering <engineering@visionary.edu>"
LABEL version="1.0.0"
LABEL description="Visionary RAG Pipeline - PDF Ingestion Service"
