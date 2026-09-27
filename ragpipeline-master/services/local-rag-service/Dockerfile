# Local RAG Service — single-container, fully-offline RAG stack
# Build context must be the repository root (needs the bundled reranker model):
#   docker build -f services/local-rag-service/Dockerfile -t local-rag .
FROM python:3.11-slim

WORKDIR /app

COPY services/local-rag-service/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY services/local-rag-service/ /app/
# Bundled cross-encoder reranker (shared with reranker-service)
COPY services/reranker-service/models/ /reranker-models/
ENV RERANKER_MODEL_PATH=/reranker-models/ms-marco-MiniLM-L-12-v2/flashrank-MiniLM-L-12-v2_Q.onnx \
    RERANKER_TOKENIZER_PATH=/reranker-models/ms-marco-MiniLM-L-12-v2/tokenizer.json \
    PORT=8080

# Ingest the bundled corpus at build time so the image is ready to answer
RUN python -m scripts.ingest_corpus

EXPOSE 8080
HEALTHCHECK CMD python -c "import urllib.request;urllib.request.urlopen('http://localhost:8080/health')"
CMD ["python", "-m", "uvicorn", "app.main:app", "--host", "0.0.0.0", "--port", "8080"]
