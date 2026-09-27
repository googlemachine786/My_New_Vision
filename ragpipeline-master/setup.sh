#!/bin/bash
# Setup Script for Visionary RAG Pipeline (Linux/Mac)
# Run this to set up your local development environment

echo "========================================"
echo "Visionary RAG Pipeline - Setup"
echo "========================================"
echo ""

# Check Python version
echo "[1/5] Checking Python version..."
python --version
if ! python --version 2>&1 | grep -q "Python 3.1[1-9]"; then
    echo "  WARNING: Python 3.11+ recommended"
fi

# Create virtual environment
echo ""
echo "[2/5] Creating virtual environment..."
if [ -d "venv" ]; then
    echo "  Virtual environment already exists"
else
    python -m venv venv
    echo "  Virtual environment created"
fi

# Activate virtual environment
echo ""
echo "[3/5] Activating virtual environment..."
source venv/bin/activate
echo "  Activated"

# Install dependencies
echo ""
echo "[4/5] Installing dependencies..."
pip install --upgrade pip
pip install -r requirements.txt
echo "  Dependencies installed"

# Create .env.local if it doesn't exist
echo ""
echo "[5/5] Setting up environment file..."
if [ -f ".env.local" ]; then
    echo "  .env.local already exists"
else
    cp .env.local.example .env.local
    echo "  .env.local created from .env.local.example"
    echo "  IMPORTANT: Edit .env.local and fill in your Supabase credentials!"
fi

echo ""
echo "========================================"
echo "Setup Complete!"
echo "========================================"
echo ""
echo "Next Steps:"
echo "1. Edit .env.local and add your Supabase credentials"
echo "2. Run: supabase db push (to apply migrations)"
echo "3. Run: python test_e2e.py (to run tests)"
echo ""
echo "To activate virtual environment manually:"
echo "  source venv/bin/activate"
echo ""
