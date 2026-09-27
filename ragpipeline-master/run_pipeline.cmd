@echo off
setlocal
echo =======================================================
echo Visionary RAG Pipeline - Terminal Bootstrapper 🚀
echo =======================================================

echo.
echo [1/3] Starting dependencies (Postgres, Redis, Ollama)...
call docker-compose up -d

echo.
echo [2/3] Checking if Go Orchestrator is already running...
netstat -ano | findstr :8080 >nul
if %ERRORLEVEL% equ 0 (
    echo ✔️ Go Orchestrator is already running on port 8080.
) else (
    echo 🚀 Starting Go Orchestrator in the background...
    start "Visionary Go Orchestrator" cmd /c "go run orchestrator/cmd/server/main.go"
    echo ⏳ Waiting for Orchestrator to become ready...
    timeout /t 5 >nul
)

echo.
echo [3/3] Launching Terminal Client...
echo -------------------------------------------------------
python terminal_client.py

echo.
echo 🛑 Shutting down backend? Use: docker-compose down
endlocal
