# Visionary RAG - Local Setup Script (Windows PowerShell)
# Run: .\scripts\setup-local-windows.ps1

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Visionary RAG - Local Setup" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$ErrorActionPreference = "Stop"

# Check if Docker is running
Write-Host "[1/5] Checking Docker..." -ForegroundColor Yellow
try {
    $dockerStatus = docker ps
    Write-Host "✓ Docker is running" -ForegroundColor Green
} catch {
    Write-Host "✗ Docker is not running. Please start Docker Desktop." -ForegroundColor Red
    exit 1
}

# Check if docker-compose is available
Write-Host "[2/5] Checking docker-compose..." -ForegroundColor Yellow
if (Get-Command docker-compose -ErrorAction SilentlyContinue) {
    Write-Host "✓ docker-compose is available" -ForegroundColor Green
} else {
    Write-Host "✗ docker-compose not found. Please install Docker Desktop." -ForegroundColor Red
    exit 1
}

# Start services
Write-Host "[3/5] Starting local services..." -ForegroundColor Yellow
Set-Location $PSScriptRoot\..
docker-compose up -d

if ($LASTEXITCODE -ne 0) {
    Write-Host "✗ Failed to start services" -ForegroundColor Red
    exit 1
}

Write-Host "✓ Services started" -ForegroundColor Green

# Wait for services to be healthy
Write-Host "[4/5] Waiting for services to be ready..." -ForegroundColor Yellow
Write-Host "  This may take 2-3 minutes..." -ForegroundColor Gray

$maxAttempts = 30
$attempt = 0

do {
    Start-Sleep -Seconds 5
    $attempt++
    
    $postgresReady = docker exec visionary-postgres pg_isready -U visionary 2>&1
    $redisReady = docker exec visionary-redis redis-cli ping 2>&1
    $ollamaReady = docker exec visionary-ollama curl -s http://localhost:11434/api/tags 2>&1
    
    Write-Host "  Attempt $attempt/$maxAttempts - PostgreSQL: $($postgresReady -like '*accepting*'), Redis: $($redisReady -eq 'PONG'), Ollama: $($ollamaReady -notlike '*error*')" -ForegroundColor Gray
    
    if ($attempt -ge $maxAttempts) {
        Write-Host "✗ Services took too long to start" -ForegroundColor Red
        exit 1
    }
} while (
    ($postgresReady -notlike '*accepting*') -or 
    ($redisReady -ne 'PONG') -or 
    ($ollamaReady -notlike '*models*')
)

Write-Host "✓ All services are ready" -ForegroundColor Green

# Pull Ollama models
Write-Host "[5/5] Pulling Ollama models..." -ForegroundColor Yellow
Write-Host "  - nomic-embed-text (embeddings)" -ForegroundColor Gray
Write-Host "  - llama3.2:3b (LLM)" -ForegroundColor Gray

docker exec visionary-ollama ollama pull nomic-embed-text
if ($LASTEXITCODE -ne 0) {
    Write-Host "✗ Failed to pull nomic-embed-text" -ForegroundColor Red
} else {
    Write-Host "✓ nomic-embed-text pulled" -ForegroundColor Green
}

docker exec visionary-ollama ollama pull llama3.2:3b
if ($LASTEXITCODE -ne 0) {
    Write-Host "✗ Failed to pull llama3.2:3b" -ForegroundColor Red
} else {
    Write-Host "✓ llama3.2:3b pulled" -ForegroundColor Green
}

# Initialize database
Write-Host ""
Write-Host "Initializing database..." -ForegroundColor Yellow
$env:PGPASSWORD = "localdev123"
psql -h localhost -p 5432 -U visionary -d visionary -f schema\v2_production.sql

if ($LASTEXITCODE -eq 0) {
    Write-Host "✓ Database initialized" -ForegroundColor Green
} else {
    Write-Host "⚠ Database may already be initialized" -ForegroundColor Yellow
}

# Create .env.local if it doesn't exist
if (-not (Test-Path ".env.local")) {
    Write-Host ""
    Write-Host "Creating .env.local..." -ForegroundColor Yellow
    Copy-Item ".env.local.example" ".env.local"
    Write-Host "✓ .env.local created" -ForegroundColor Green
    Write-Host "  Please review and update the configuration" -ForegroundColor Gray
}

# Summary
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Setup Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Services running:" -ForegroundColor White
Write-Host "  - PostgreSQL:  localhost:5432" -ForegroundColor Cyan
Write-Host "  - Redis:       localhost:6379" -ForegroundColor Cyan
Write-Host "  - Ollama:      localhost:11434" -ForegroundColor Cyan
Write-Host "  - PGAdmin:     localhost:5050 (optional, email: admin@visionary.local, pass: admin123)" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor White
Write-Host "  1. Review .env.local configuration" -ForegroundColor Gray
Write-Host "  2. Run ingestion: cd ingestion-go && go run cmd/ingestion/main.go --pdf <path>" -ForegroundColor Gray
Write-Host "  3. Run orchestrator: cd orchestrator && go run cmd/server/main.go" -ForegroundColor Gray
Write-Host "  4. Test query: curl -X POST http://localhost:8080/query -d '{`"query`":`"test`"}'" -ForegroundColor Gray
Write-Host ""
Write-Host "To stop services: docker-compose down" -ForegroundColor Yellow
Write-Host ""
