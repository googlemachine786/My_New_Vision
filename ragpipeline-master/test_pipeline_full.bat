@echo off
REM Visionary RAG Pipeline - Complete Pipeline Test Script
REM Tests: PostgreSQL + Redis + Go Services + GCP Vertex AI

setlocal enabledelayedexpansion

set "GOPATH=C:\Program Files\Go\bin"
set "PATH=%GOPATH%;%PATH%"
set "PROJECT_ROOT=C:\Users\kommi\ragpipeline"

REM Your GCP API key
set "GCP_API_KEY=AIzaSyBcxwsSdb2mNnqYGcaYbMH1furWLae-W1k"

REM Environment variables for all services
set "VERTEX_PROJECT=visionary-rag-test"
set "VERTEX_LOCATION=us-central1"
set "VERTEX_MODEL=text-embedding-005"
set "VERTEX_DIMENSION=768"
set "GOOGLE_API_KEY=%GCP_API_KEY%"
set "ENVIRONMENT=development"
set "LOG_LEVEL=debug"

REM Database connection
set "DATABASE_URL=postgresql://postgres:postgres@localhost:5432/ragdb"
set "DB_HOST=localhost"
set "DB_PORT=5432"
set "DB_USER=postgres"
set "DB_PASSWORD=postgres"
set "DB_NAME=ragdb"

REM Redis
set "REDIS_URL=redis://localhost:6379/0"
set "REDIS_ADDR=localhost:6379"

REM Service ports
set "EMBEDDING_PORT=8081"
set "VECTOR_SEARCH_PORT=8082"
set "QUERY_UNDERSTANDING_PORT=8083"
set "API_GATEWAY_PORT=8080"
set "RERANKER_PORT=8085"

REM Cache settings
set "EMBEDDING_CACHE_ENABLED=true"
set "EMBEDDING_CACHE_TTL=86400s"
set "EMBEDDING_CACHE_MAX_SIZE=10000"

cd /d "%PROJECT_ROOT%"

echo ================================================================
echo Visionary RAG Pipeline - Complete Integration Test
echo ================================================================
echo.

REM Step 1: Check Docker services
echo [1/7] Checking Docker services...
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"
echo.

REM Step 2: Test PostgreSQL connection
echo [2/7] Testing PostgreSQL...
docker exec ragpipeline-postgres-1 pg_isready -U postgres
if %ERRORLEVEL% NEQ 0 (
    echo ERROR: PostgreSQL is not ready
    exit /b 1
)
echo.

REM Step 3: Test Redis connection
echo [3/7] Testing Redis...
for /f "delims=" %%i in ('docker exec ragpipeline-redis-1 redis-cli ping') do set REDIS_PING=%%i
if "!REDIS_PING!" NEQ "PONG" (
    echo ERROR: Redis is not responding
    exit /b 1
)
echo Redis: OK
echo.

REM Step 4: Verify database schema
echo [4/7] Verifying database schema...
docker exec ragpipeline-postgres-1 psql -U postgres -d ragdb -c "SELECT COUNT(*) as table_count FROM information_schema.tables WHERE table_schema = 'public' AND table_name IN ('cbse_taxonomy', 'parent_chunks', 'child_chunks', 'ingestion_dlq', 'ai_feedback_loop', 'document_versions');"
docker exec ragpipeline-postgres-1 psql -U postgres -d ragdb -c "SELECT COUNT(*) as taxonomy_count FROM cbse_taxonomy;"
docker exec ragpipeline-postgres-1 psql -U postgres -d ragdb -c "SELECT extname FROM pg_extension WHERE extname IN ('vector', 'btree_gin', 'pgcrypto');"
echo.

REM Step 5: Verify binaries exist
echo [5/7] Checking service binaries...
if not exist "bin\embedding-service.exe" (
    echo ERROR: embedding-service.exe not found
    exit /b 1
)
if not exist "bin\vector-search-service.exe" (
    echo ERROR: vector-search-service.exe not found
    exit /b 1
)
if not exist "bin\query-understanding-service.exe" (
    echo ERROR: query-understanding-service.exe not found
    exit /b 1
)
if not exist "bin\api-gateway.exe" (
    echo ERROR: api-gateway.exe not found
    exit /b 1
)
echo All binaries present
echo.

REM Step 6: Start services in background
echo [6/7] Starting services...
echo.

REM Start Embedding Service (Port 8081)
echo   - Starting Embedding Service on port 8081...
set "PORT=%EMBEDDING_PORT%"
start "embedding-service" /MIN cmd /c "cd /d %PROJECT_ROOT% && set PORT=%EMBEDDING_PORT% && embedding-service.exe"
timeout /t 3 /nobreak >nul

REM Start Vector Search Service (Port 8082)
echo   - Starting Vector Search Service on port 8082...
start "vector-search-service" /MIN cmd /c "cd /d %PROJECT_ROOT% && set PORT=%VECTOR_SEARCH_PORT% && vector-search-service.exe"
timeout /t 3 /nobreak >nul

REM Start Query Understanding Service (Port 8083)
echo   - Starting Query Understanding Service on port 8083...
start "query-understanding-service" /MIN cmd /c "cd /d %PROJECT_ROOT% && set PORT=%QUERY_UNDERSTANDING_PORT% && query-understanding-service.exe"
timeout /t 3 /nobreak >nul

REM Start API Gateway (Port 8080)
echo   - Starting API Gateway on port 8080...
start "api-gateway" /MIN cmd /c "cd /d %PROJECT_ROOT% && set PORT=%API_GATEWAY_PORT% && api-gateway.exe"
timeout /t 5 /nobreak >nul

echo.
echo   Services starting... waiting for health checks...
timeout /t 10 /nobreak >nul

REM Step 7: Health checks
echo [7/7] Running health checks...
echo.

echo   Checking Embedding Service...
curl -s http://localhost:8081/health
echo.

echo   Checking Vector Search Service...
curl -s http://localhost:8082/health
echo.

echo   Checking Query Understanding Service...
curl -s http://localhost:8083/health
echo.

echo   Checking API Gateway...
curl -s http://localhost:8080/health
echo.

echo.
echo ================================================================
echo Pipeline Test Complete!
echo ================================================================
echo.
echo Services running:
echo   - PostgreSQL:  localhost:5432 (ragdb)
echo   - Redis:       localhost:6379
echo   - Embedding Service:       http://localhost:8081
echo   - Vector Search Service:   http://localhost:8082
echo   - Query Understanding:     http://localhost:8083
echo   - API Gateway:             http://localhost:8080
echo.
echo GCP Project: %VERTEX_PROJECT%
echo GCP Location: %VERTEX_LOCATION%
echo GCP Model: %VERTEX_MODEL%
echo.
echo To stop services: taskkill /F /IM *.exe /FI "WINDOWTITLE eq embedding*" ^& taskkill /F /IM *.exe /FI "WINDOWTITLE eq vector*" ^& taskkill /F /IM *.exe /FI "WINDOWTITLE eq query*" ^& taskkill /F /IM *.exe /FI "WINDOWTITLE eq api*"
echo.
pause
