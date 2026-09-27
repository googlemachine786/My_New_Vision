# Setup Script for Visionary RAG Pipeline
# Run this to set up your local development environment

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Visionary RAG Pipeline - Setup" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check Python version
Write-Host "[1/5] Checking Python version..." -ForegroundColor Yellow
$pythonVersion = python --version 2>&1
Write-Host "  Found: $pythonVersion" -ForegroundColor Green

if (-not (python --version 2>&1 | Select-String "Python 3.1[1-9]")) {
    Write-Host "  WARNING: Python 3.11+ recommended" -ForegroundColor Red
}

# Create virtual environment
Write-Host ""
Write-Host "[2/5] Creating virtual environment..." -ForegroundColor Yellow
if (Test-Path "venv") {
    Write-Host "  Virtual environment already exists" -ForegroundColor Green
} else {
    python -m venv venv
    Write-Host "  Virtual environment created" -ForegroundColor Green
}

# Activate virtual environment
Write-Host ""
Write-Host "[3/5] Activating virtual environment..." -ForegroundColor Yellow
& ".\venv\Scripts\Activate.ps1"
Write-Host "  Activated" -ForegroundColor Green

# Install dependencies
Write-Host ""
Write-Host "[4/5] Installing dependencies..." -ForegroundColor Yellow
pip install --upgrade pip
pip install -r requirements.txt
Write-Host "  Dependencies installed" -ForegroundColor Green

# Create .env.local if it doesn't exist
Write-Host ""
Write-Host "[5/5] Setting up environment file..." -ForegroundColor Yellow
if (Test-Path ".env.local") {
    Write-Host "  .env.local already exists" -ForegroundColor Green
} else {
    Copy-Item ".env.local.example" ".env.local"
    Write-Host "  .env.local created from .env.local.example" -ForegroundColor Green
    Write-Host "  IMPORTANT: Edit .env.local and fill in your Supabase credentials!" -ForegroundColor Red
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Setup Complete!" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host "1. Edit .env.local and add your Supabase credentials" -ForegroundColor White
Write-Host "2. Run: supabase db push (to apply migrations)" -ForegroundColor White
Write-Host "3. Run: python test_e2e.py (to run tests)" -ForegroundColor White
Write-Host ""
Write-Host "To activate virtual environment manually:" -ForegroundColor Yellow
Write-Host "  .\venv\Scripts\Activate.ps1" -ForegroundColor White
Write-Host ""
