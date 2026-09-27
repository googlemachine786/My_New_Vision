@echo off
REM Start Embedding Service with GCP API Key
set "PATH=C:\Program Files\Go\bin;%PATH%"
set "PORT=8081"
set "VERTEX_PROJECT=visionary-rag-test"
set "VERTEX_LOCATION=us-central1"
set "VERTEX_MODEL=text-embedding-005"
set "VERTEX_DIMENSION=768"
set "GOOGLE_API_KEY=AIzaSyBcxwsSdb2mNnqYGcaYbMH1furWLae-W1k"
set "ENVIRONMENT=development"
set "LOG_LEVEL=debug"
set "REDIS_URL=redis://localhost:6379/1"
set "EMBEDDING_CACHE_ENABLED=true"

cd /d C:\Users\kommi\ragpipeline
echo Starting Embedding Service on port 8081...
.\bin\embedding-service.exe
