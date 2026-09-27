FROM python:3.11-slim AS base

WORKDIR /app

# Install dependencies first for better layer caching
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Download model at build time so the container starts fast
RUN python -c "from flashrank import Ranker; Ranker(model_name='ms-marco-MiniLM-L-6-v2', cache_dir='/app/models')"

COPY main.py .

EXPOSE 8000

CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
