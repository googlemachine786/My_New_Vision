@echo off
REM ================================================================
REM Visionary RAG Pipeline - Complete 12-Aspect Test Suite
REM Tests ALL missing coverage aspects from the audit
REM ================================================================
setlocal enabledelayedexpansion

set "GOPATH=C:\Program Files\Go\bin"
set "PATH=%GOPATH%;%PATH%"
set "PROJECT_ROOT=C:\Users\kommi\ragpipeline"

REM GCP Configuration
set "GCP_API_KEY=AIzaSyBcxwsSdb2mNnqYGcaYbMH1furWLae-W1k"
set "VERTEX_PROJECT=visionary-rag-test"
set "VERTEX_LOCATION=us-central1"
set "VERTEX_MODEL=text-embedding-005"
set "VERTEX_DIMENSION=768"
set "GEMINI_MODEL=gemini-2.5-flash"

REM Database
set "DATABASE_URL=postgresql://postgres:postgres@localhost:5432/ragdb"
set "DB_HOST=localhost"
set "DB_PORT=5432"
set "DB_USER=postgres"
set "DB_PASSWORD=postgres"
set "DB_NAME=ragdb"

REM Redis
set "REDIS_URL=redis://localhost:6379/0"
set "REDIS_ADDR=localhost:6379"

REM Feature flags
set "ENVIRONMENT=development"
set "LOG_LEVEL=debug"
set "EMBEDDING_CACHE_ENABLED=false"
set "SEARCH_CACHE_ENABLED=false"

cd /d "%PROJECT_ROOT%"

echo.
echo ================================================================
echo Visionary RAG Pipeline - 12-Aspect Complete Test Suite
echo ================================================================
echo.
echo Infrastructure:
echo   PostgreSQL: localhost:5432/ragdb
echo   Redis:      localhost:6379
echo   GCP API:    gemini-embedding-001 + gemini-2.5-flash
echo   Go:         %GOPATH%
echo.

REM ================================================================
REM STEP 1: Verify infrastructure is running
REM ================================================================
echo [1/12] Verifying Infrastructure...
echo.

docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>nul
if !ERRORLEVEL! NEQ 0 (
    echo   Starting Docker services...
    docker-compose -f docker-compose-local.yml up -d
    timeout /t 10 /nobreak >nul
)

REM Test Redis
for /f "delims=" %%i in ('docker exec ragpipeline-redis-1 redis-cli ping 2^>nul') do set REDIS_PING=%%i
if "!REDIS_PING!"=="PONG" (
    echo   [PASS] Redis: healthy
) else (
    echo   [FAIL] Redis: not responding
    exit /b 1
)

REM Test PostgreSQL
for /f "delims=" %%i in ('docker exec ragpipeline-postgres-1 pg_isready -U postgres 2^>nul ^| findstr "accepting"') do set PG_STATUS=%%i
if not "!PG_STATUS!"=="" (
    echo   [PASS] PostgreSQL: healthy
) else (
    echo   [FAIL] PostgreSQL: not ready
    exit /b 1
)

REM Verify database schema
for /f %%i in ('docker exec ragpipeline-postgres-1 psql -U postgres -d ragdb -t -c "SELECT COUNT(*) FROM cbse_taxonomy;" 2^>nul') do set TAX_COUNT=%%i
echo   [INFO] Taxonomy entries: !TAX_COUNT!

echo.

REM ================================================================
REM STEP 2: Kill any existing service processes
REM ================================================================
echo [2/12] Cleaning up existing processes...
echo.
taskkill /F /IM embedding-service.exe 2>nul
taskkill /F /IM vector-search-service.exe 2>nul
taskkill /F /IM query-understanding-service.exe 2>nul
taskkill /F /IM api-gateway.exe 2>nul
timeout /t 2 /nobreak >nul
echo   Done.

echo.

REM ================================================================
REM STEP 3: Start Vector Search Service (depends on PostgreSQL + Redis)
REM ================================================================
echo [3/12] Starting Vector Search Service on port 8082...
echo.
set "PORT=8082"
set "REDIS_URL=redis://localhost:6379/2"
start "vector-search-service" /MIN cmd /c "cd /d %PROJECT_ROOT% && set PORT=8082 && set DATABASE_URL=%DATABASE_URL% && set REDIS_URL=redis://localhost:6379/2 && set SEARCH_CACHE_ENABLED=false && set ENVIRONMENT=development && set LOG_LEVEL=debug && .\bin\vector-search-service.exe"
timeout /t 5 /nobreak >nul

echo   Checking health...
curl -s http://localhost:8082/health >nul 2>&1
if !ERRORLEVEL! EQU 0 (
    echo   [PASS] Vector Search Service: healthy
) else (
    echo   [WARN] Vector Search Service: may not be ready yet, will retry later
)

echo.

REM ================================================================
REM STEP 4: Start Query Understanding Service (depends on GCP)
REM ================================================================
echo [4/12] Starting Query Understanding Service on port 8083...
echo.
start "query-understanding-service" /MIN cmd /c "cd /d %PROJECT_ROOT% && set PORT=8083 && set GEMINI_PROJECT_ID=%VERTEX_PROJECT% && set GEMINI_LOCATION=%VERTEX_LOCATION% && set GEMINI_MODEL=%GEMINI_MODEL% && set GOOGLE_API_KEY=%GCP_API_KEY% && set ENVIRONMENT=development && set LOG_LEVEL=debug && .\bin\query-understanding-service.exe"
timeout /t 5 /nobreak >nul

echo   Checking health...
curl -s http://localhost:8083/health >nul 2>&1
if !ERRORLEVEL! EQU 0 (
    echo   [PASS] Query Understanding Service: healthy
) else (
    echo   [WARN] Query Understanding Service: may not be ready yet, will retry later
)

echo.

REM ================================================================
REM STEP 5: Start API Gateway (depends on all other services)
REM ================================================================
echo [5/12] Starting API Gateway on port 8080...
echo.
start "api-gateway" /MIN cmd /c "cd /d %PROJECT_ROOT% && set PORT=8080 && set EMBEDDING_SERVICE_URL=http://localhost:8081 && set VECTOR_SEARCH_SERVICE_URL=http://localhost:8082 && set QUERY_UNDERSTANDING_SERVICE_URL=http://localhost:8083 && set REDIS_URL=redis://localhost:6379/0 && set JWT_SECRET=local-test-secret && set ENVIRONMENT=development && set LOG_LEVEL=debug && .\bin\api-gateway.exe"
timeout /t 5 /nobreak >nul

echo   Checking health...
curl -s http://localhost:8080/health >nul 2>&1
if !ERRORLEVEL! EQU 0 (
    echo   [PASS] API Gateway: healthy
) else (
    echo   [WARN] API Gateway: may not be ready yet, will retry later
)

echo.

REM ================================================================
REM STEP 6: Wait for all services to stabilize
REM ================================================================
echo [6/12] Waiting for services to stabilize...
timeout /t 10 /nobreak >nul
echo.

echo   Service health checks:
echo   - Vector Search:
curl -s http://localhost:8082/health 2>nul || echo     UNREACHABLE
echo.
echo   - Query Understanding:
curl -s http://localhost:8083/health 2>nul || echo     UNREACHABLE
echo.
echo   - API Gateway:
curl -s http://localhost:8080/health 2>nul || echo     UNREACHABLE
echo.

echo.

REM ================================================================
REM STEP 7: Test HTTP Endpoints
REM ================================================================
echo [7/12] Testing HTTP Endpoints...
echo.

echo   Testing /health endpoints...
echo     Vector Search /health:
curl -s http://localhost:8082/health 2>nul | python -m json.tool 2>nul || echo       FAILED
echo.

echo     Query Understanding /health:
curl -s http://localhost:8083/health 2>nul | python -m json.tool 2>nul || echo       FAILED
echo.

echo     API Gateway /health:
curl -s http://localhost:8080/health 2>nul | python -m json.tool 2>nul || echo       FAILED
echo.

echo   Testing /version endpoints...
echo     Vector Search /version:
curl -s http://localhost:8082/version 2>nul | python -m json.tool 2>nul || echo       FAILED
echo.

echo     API Gateway /version:
curl -s http://localhost:8080/version 2>nul | python -m json.tool 2>nul || echo       FAILED
echo.

echo.

REM ================================================================
REM STEP 8: Run Python-based deep tests
REM ================================================================
echo [8/12] Running Python-based pipeline tests...
echo.
python test_pipeline_complete.py
echo.

REM ================================================================
REM STEP 9: Test GCP Gemini API directly
REM ================================================================
echo [9/12] Testing GCP Gemini API directly...
echo.

echo   Testing Gemini LLM (gemini-2.5-flash)...
curl -s -X POST "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=%GCP_API_KEY%" -H "Content-Type: application/json" -d "{\"contents\":[{\"parts\":[{\"text\":\"What is force and pressure in one sentence?\"}]}]}" >nul 2>&1
if !ERRORLEVEL! EQU 0 (
    echo   [PASS] Gemini LLM: responding
) else (
    echo   [FAIL] Gemini LLM: failed
)

echo.
echo   Testing Gemini Embedding (gemini-embedding-001)...
curl -s -X POST "https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-001:embedContent?key=%GCP_API_KEY%" -H "Content-Type: application/json" -d "{\"content\":{\"parts\":[{\"text\":\"What is photosynthesis?\"}]}}" >nul 2>&1
if !ERRORLEVEL! EQU 0 (
    echo   [PASS] Embedding API: responding
) else (
    echo   [FAIL] Embedding API: failed
)

echo.

REM ================================================================
REM STEP 10: Verify Go binaries
REM ================================================================
echo [10/12] Verifying Go service binaries...
echo.

for %%f in (bin\embedding-service.exe bin\vector-search-service.exe bin\query-understanding-service.exe bin\api-gateway.exe) do (
    if exist "%%f" (
        for %%I in ("%%f") do (
            set "SIZE=%%~zI"
            set /a "SIZE_MB=!SIZE!/1048576"
            echo   [PASS] %%f (!SIZE_MB! MB)
        )
    ) else (
        echo   [FAIL] %%f not found
    )
)

echo.

REM ================================================================
REM STEP 11: Run Go tests
REM ================================================================
echo [11/12] Running Go service tests...
echo.

echo   Testing embedding-service module...
cd services\embedding-service
go test ./... -v -short -count=1 2>&1 | findstr /C:"PASS" /C:"FAIL" /C:"ok" /C:"FAIL"
cd ..\..

echo.
echo   Testing vector-search-service module...
cd services\vector-search-service
go test ./... -v -short -count=1 2>&1 | findstr /C:"PASS" /C:"FAIL" /C:"ok" /C:"FAIL"
cd ..\..

echo.
echo   Testing query-understanding-service module...
cd services\query-understanding-service
go test ./... -v -short -count=1 2>&1 | findstr /C:"PASS" /C:"FAIL" /C:"ok" /C:"FAIL"
cd ..\..

echo.
echo   Testing api-gateway module...
cd services\api-gateway
go test ./... -v -short -count=1 2>&1 | findstr /C:"PASS" /C:"FAIL" /C:"ok" /C:"FAIL"
cd ..\..

echo.

REM ================================================================
REM STEP 12: Final Summary
REM ================================================================
echo.
echo ================================================================
echo FINAL SUMMARY
echo ================================================================
echo.

echo Infrastructure:
docker ps --format "  {{.Names}}: {{.Status}}" 2>nul
echo.

echo Go Services (running):
tasklist /FI "IMAGENAME eq embedding-service.exe" 2>nul | findstr "exe" >nul && echo   embedding-service:         RUNNING || echo   embedding-service:         NOT RUNNING
tasklist /FI "IMAGENAME eq vector-search-service.exe" 2>nul | findstr "exe" >nul && echo   vector-search-service:     RUNNING || echo   vector-search-service:     NOT RUNNING
tasklist /FI "IMAGENAME eq query-understanding-service.exe" 2>nul | findstr "exe" >nul && echo   query-understanding:       RUNNING || echo   query-understanding:       NOT RUNNING
tasklist /FI "IMAGENAME eq api-gateway.exe" 2>nul | findstr "exe" >nul && echo   api-gateway:               RUNNING || echo   api-gateway:               NOT RUNNING
echo.

echo Service Endpoints:
echo   API Gateway:             http://localhost:8080
echo   Embedding Service:       http://localhost:8081
echo   Vector Search Service:   http://localhost:8082
echo   Query Understanding:     http://localhost:8083
echo.

echo Database:
for /f %%i in ('docker exec ragpipeline-postgres-1 psql -U postgres -d ragdb -t -c "SELECT COUNT(*) FROM cbse_taxonomy;" 2^>nul') do echo   Taxonomy entries: %%i
for /f %%i in ('docker exec ragpipeline-postgres-1 psql -U postgres -d ragdb -t -c "SELECT COUNT(*) FROM parent_chunks;" 2^>nul') do echo   Parent chunks: %%i
for /f %%i in ('docker exec ragpipeline-postgres-1 psql -U postgres -d ragdb -t -c "SELECT COUNT(*) FROM child_chunks;" 2^>nul') do echo   Child chunks: %%i
echo.

echo GCP APIs:
echo   Gemini LLM:              gemini-2.5-flash
echo   Embedding:               gemini-embedding-001 (3072 dim)
echo   API Key:                 AIzaSyB...W1k
echo.

echo ================================================================
echo Test Suite Complete!
echo ================================================================
echo.
pause
