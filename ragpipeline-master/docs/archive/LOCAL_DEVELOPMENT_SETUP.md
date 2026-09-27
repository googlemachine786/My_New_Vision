# Local Development Setup - Visionary RAG Pipeline

**Complete guide for running the entire pipeline locally without GCP credentials.**

---

## 🎯 Overview

This guide shows you how to run Visionary RAG locally using:

| Production Service | Local Alternative | Purpose |
|-------------------|-------------------|---------|
| Vertex AI Embeddings | **Ollama** (nomic-embed-text) | Text embeddings |
| Vertex AI LLM | **Ollama** (llama3.2, mistral) | Text generation |
| AlloyDB | **PostgreSQL + pgvector** | Vector database |
| Memorystore Redis | **Docker Redis** | Session cache |
| Cloud Run | **Local Go binary** | HTTP server |
| Secret Manager | **.env file** | Configuration |

---

## 📦 Prerequisites

### Install These Tools

1. **Docker Desktop** (Windows/Mac/Linux)
   - Download: https://www.docker.com/products/docker-desktop/
   - Required for: PostgreSQL, Redis, Ollama

2. **Go 1.25+** (Already installed ✅)
   - For running Go services locally

3. **Python 3.11+** (Already installed ✅)
   - For running Python scripts

4. **PostgreSQL 15+** (via Docker)
   - With pgvector extension

---

## 🚀 Quick Start (5 Minutes)

### Step 1: Start Local Services

```bash
# Create docker-compose.yml (see below)
docker-compose up -d

# Verify services are running
docker-compose ps
```

**Expected output:**
```
NAME                STATUS
postgres            Up (healthy)
redis               Up (healthy)
ollama              Up (healthy)
```

### Step 2: Configure Environment

```bash
# Copy example env file
cp .env.local.example .env.local

# Load environment
source .env.local  # Linux/Mac
# OR
.\.env.local       # Windows PowerShell
```

### Step 3: Initialize Database

```bash
cd scripts
./init-local-db.sh  # Creates database + schema
```

### Step 4: Test Locally

```bash
# Test embedding
python test_local_embedding.py "What is photosynthesis?"

# Test query endpoint
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -d '{"query":"What is photosynthesis?","session_id":"test"}'
```

---

## 📝 Complete Setup Guide

### 1. Docker Compose Configuration

**File:** `docker-compose.yml`

```yaml
version: '3.8'

services:
  # PostgreSQL with pgvector
  postgres:
    image: pgvector/pgvector:pg15
    container_name: visionary-postgres
    environment:
      POSTGRES_USER: visionary
      POSTGRES_PASSWORD: localdev123
      POSTGRES_DB: visionary
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./schema:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U visionary"]
      interval: 5s
      timeout: 5s
      retries: 5

  # Redis for sessions
  redis:
    image: redis:7-alpine
    container_name: visionary-redis
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    command: redis-server --appendonly yes
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

  # Ollama for embeddings + LLM
  ollama:
    image: ollama/ollama:latest
    container_name: visionary-ollama
    ports:
      - "11434:11434"
    volumes:
      - ollama_data:/root/.ollama
    environment:
      - OLLAMA_HOST=0.0.0.0
    deploy:
      resources:
        reservations:
          devices:
            - driver: nvidia
              count: all
              capabilities: [gpu]  # Optional: GPU acceleration

volumes:
  postgres_data:
  redis_data:
  ollama_data:
```

---

### 2. Environment Configuration

**File:** `.env.local.example`

```bash
# ===========================================
# Local Development Configuration
# ===========================================

# Mode: "local" or "production"
APP_MODE=local

# ===========================================
# PostgreSQL (Local AlloyDB alternative)
# ===========================================
ALLOYDB_DSN=host=localhost port=5432 dbname=visionary user=visionary password=localdev123
PGBOUNCER_HOST=localhost
PGBOUNCER_PORT=5432
PG_MAX_CONNS=10
PG_MIN_CONNS=2

# ===========================================
# Redis (Local Memorystore alternative)
# ===========================================
REDIS_ADDR=localhost:6379
REDIS_POOL_SIZE=50
SESSION_TTL=2100s

# ===========================================
# Ollama (Local Vertex AI alternative)
# ===========================================
OLLAMA_BASE_URL=http://localhost:11434
EMBED_MODEL=nomic-embed-text
LLM_MODEL=llama3.2:3b
EMBED_DIMENSION=768
EMBED_TASK_TYPE=RETRIEVAL_QUERY

# ===========================================
# JWT (Same for local/prod)
# ===========================================
JWT_SECRET=local-dev-secret-change-in-production
JWT_EXPIRATION=24h

# ===========================================
# Application Settings
# ===========================================
PORT=8080
LOG_LEVEL=debug
ENVIRONMENT=development
REQUEST_TIMEOUT=450ms
TOP_K=5
RRF_K=60.0

# ===========================================
# Feature Flags
# ===========================================
USE_LOCAL_EMBEDDINGS=true
USE_LOCAL_LLM=true
ENABLE_CLOUD_TRACE=false
ENABLE_CLOUD_MONITORING=false
```

**Usage:**
```bash
# Copy example
cp .env.local.example .env.local

# Load environment
# Linux/Mac:
source .env.local

# Windows PowerShell:
Get-Content .env.local | ForEach-Object {
  if ($_ -match '^\s*([^#][^=]+)=(.+)\s*$') {
    [Environment]::SetEnvironmentVariable($Matches[1].Trim(), $Matches[2].Trim())
  }
}
```

---

### 3. Database Initialization Script

**File:** `scripts/init-local-db.sh`

```bash
#!/bin/bash
# Initialize local PostgreSQL database

set -euo pipefail

echo "🚀 Initializing local database..."

# Wait for PostgreSQL to be ready
echo "⏳ Waiting for PostgreSQL..."
until docker exec visionary-postgres pg_isready -U visionary > /dev/null 2>&1; do
  sleep 1
done
echo "✅ PostgreSQL is ready"

# Get connection string
DSN="host=localhost port=5432 dbname=visionary user=visionary password=localdev123"

# Apply schema
echo "📝 Applying schema..."
psql "$DSN" -f schema/v2_production.sql

# Verify
echo "✅ Verifying installation..."
psql "$DSN" -c "SELECT COUNT(*) as taxonomy_count FROM cbse_taxonomy;"
psql "$DSN" -c "SELECT indexname FROM pg_indexes WHERE tablename = 'child_chunks';"

echo "🎉 Database initialized successfully!"
```

**Windows PowerShell version:** `scripts/init-local-db.ps1`

```powershell
# Initialize local database
Write-Host "🚀 Initializing local database..." -ForegroundColor Green

# Wait for PostgreSQL
Write-Host "⏳ Waiting for PostgreSQL..."
do {
  Start-Sleep -Seconds 2
  $result = docker exec visionary-postgres pg_isready -U visionary 2>&1
} while ($result -notlike "*accepting connections*")

Write-Host "✅ PostgreSQL is ready" -ForegroundColor Green

# Apply schema
Write-Host "📝 Applying schema..." -ForegroundColor Green
$env:PGPASSWORD = "localdev123"
psql -h localhost -p 5432 -U visionary -d visionary -f ..\schema\v2_production.sql

Write-Host "🎉 Database initialized!" -ForegroundColor Green
```

---

### 4. Ollama Model Setup

**File:** `scripts/setup-ollama.sh`

```bash
#!/bin/bash
# Pull required Ollama models

echo "🤖 Setting up Ollama models..."

# Embedding model
echo "⏳ Pulling nomic-embed-text..."
docker exec visionary-ollama ollama pull nomic-embed-text

# LLM model (choose one)
echo "⏳ Pulling llama3.2:3b (fast, good for testing)..."
docker exec visionary-ollama ollama pull llama3.2:3b

# Alternative LLM (optional)
# echo "⏳ Pulling mistral:7b (better quality, slower)..."
# docker exec visionary-ollama ollama pull mistral:7b

echo "✅ Ollama models ready!"
echo ""
echo "Available models:"
docker exec visionary-ollama ollama list
```

**Test embedding:**
```bash
docker exec visionary-ollama ollama run nomic-embed-text "Hello world"
```

---

### 5. Local Embedding Client

**File:** `ingestion-go/embedder/ollama_client.go`

```go
// Package embedder provides Ollama embedding for local development.
package embedder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OllamaEmbedClient wraps Ollama embedding API.
type OllamaEmbedClient struct {
	baseURL    string
	model      string
	httpClient *http.Client
}

// NewOllamaEmbedClient creates Ollama client.
func NewOllamaEmbedClient(baseURL, model string) *OllamaEmbedClient {
	return &OllamaEmbedClient{
		baseURL: baseURL,
		model:   model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// EmbedRequest is Ollama embedding request.
type EmbedRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

// EmbedResponse is Ollama embedding response.
type EmbedResponse struct {
	Embedding []float32 `json:"embedding"`
}

// EmbedDocuments creates embeddings using Ollama.
func (c *OllamaEmbedClient) EmbedDocuments(ctx context.Context, chunks []ChildChunk) ([][]float32, error) {
	embeddings := make([][]float32, len(chunks))
	
	for i, chunk := range chunks {
		emb, err := c.embedOne(ctx, chunk.Content)
		if err != nil {
			return nil, err
		}
		embeddings[i] = emb
	}
	
	return embeddings, nil
}

func (c *OllamaEmbedClient) embedOne(ctx context.Context, text string) ([]float32, error) {
	req := EmbedRequest{
		Model:  c.model,
		Prompt: text,
	}
	
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	
	resp, err := c.httpClient.Post(
		c.baseURL+"/api/embeddings",
		"application/json",
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var result EmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	
	return result.Embedding, nil
}

// EmbedQuery creates embedding for a query.
func (c *OllamaEmbedClient) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return c.embedOne(ctx, text)
}
```

---

### 6. Configuration Switcher

**File:** `orchestrator/config/config_local.go`

```go
package config

// LocalConfig provides local development configuration.
type LocalConfig struct {
	UseLocalEmbeddings bool
	UseLocalLLM        bool
	OllamaBaseURL      string
	EmbedModel         string
	LLMModel           string
}

// GetLocalConfig returns configuration based on APP_MODE.
func GetLocalConfig() (*Config, *LocalConfig) {
	mode := getEnv("APP_MODE", "local")
	
	localCfg := &LocalConfig{
		UseLocalEmbeddings: mode == "local",
		UseLocalLLM:        mode == "local",
		OllamaBaseURL:      getEnv("OLLAMA_BASE_URL", "http://localhost:11434"),
		EmbedModel:         getEnv("EMBED_MODEL", "nomic-embed-text"),
		LLMModel:           getEnv("LLM_MODEL", "llama3.2:3b"),
	}
	
	// Base config works for both local and prod
	cfg := &Config{
		// ... existing config fields
	}
	
	if mode == "local" {
		// Override for local
		cfg.AlloyDBDSN = getEnv("ALLOYDB_DSN", "host=localhost port=5432 dbname=visionary user=visionary password=localdev123")
		cfg.RedisAddr = getEnv("REDIS_ADDR", "localhost:6379")
		cfg.VertexProject = "local"
		cfg.VertexLocation = "local"
		cfg.EnableCloudTrace = false
		cfg.EnableCloudMonitoring = false
	}
	
	return cfg, localCfg
}
```

---

### 7. Usage Examples

#### Run Locally

```bash
# 1. Start services
docker-compose up -d

# 2. Setup database
cd scripts && ./init-local-db.sh

# 3. Setup Ollama
./setup-ollama.sh

# 4. Run ingestion (local mode)
cd ingestion-go
APP_MODE=local go run cmd/ingestion/main.go \
  --pdf ../sample.pdf \
  --grade 7 \
  --subject Science \
  --taxonomy-id 42

# 5. Run orchestrator (local mode)
cd orchestrator
APP_MODE=local go run cmd/server/main.go
```

#### Switch to Production

```bash
# Just change APP_MODE!
export APP_MODE=production
export ALLOYDB_DSN="host=... port=5432 ..."  # Production DSN
export REDIS_ADDR="production-redis:6379"
export GOOGLE_CLOUD_PROJECT="production-project"

# Run same binaries
./ingestion-go/ingestion.exe --pdf ...
./orchestrator/orchestrator.exe
```

---

### 8. Testing Local Setup

**File:** `test_local_embedding.py`

```python
#!/usr/bin/env python
"""Test local Ollama embedding."""

import requests
import json

def test_ollama_embedding(text: str):
    """Test Ollama embedding endpoint."""
    url = "http://localhost:11434/api/embeddings"
    
    payload = {
        "model": "nomic-embed-text",
        "prompt": text
    }
    
    response = requests.post(url, json=payload)
    response.raise_for_status()
    
    result = response.json()
    embedding = result.get("embedding", [])
    
    print(f"✅ Embedding generated!")
    print(f"   Dimensions: {len(embedding)}")
    print(f"   First 10 values: {embedding[:10]}")
    
    return embedding

if __name__ == "__main__":
    import sys
    text = sys.argv[1] if len(sys.argv) > 1 else "What is photosynthesis?"
    test_ollama_embedding(text)
```

**Run test:**
```bash
python test_local_embedding.py "Hello world"
```

---

### 9. Local Development Workflow

```bash
# Terminal 1: Start services
docker-compose up -d
docker-compose logs -f  # Follow logs

# Terminal 2: Run ingestion
cd ingestion-go
APP_MODE=local go run cmd/ingestion/main.go --pdf textbook.pdf --grade 7

# Terminal 3: Run orchestrator
cd orchestrator
APP_MODE=local go run cmd/server/main.go

# Terminal 4: Test queries
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -d '{"query":"What is photosynthesis?","session_id":"test123"}'
```

---

### 10. Comparison: Local vs Production

| Feature | Local | Production |
|---------|-------|------------|
| **Embeddings** | Ollama (nomic-embed-text) | Vertex AI (text-embedding-005) |
| **LLM** | Ollama (llama3.2:3b) | Vertex AI (Gemini 1.5 Flash) |
| **Database** | PostgreSQL + pgvector | AlloyDB + ScaNN |
| **Redis** | Docker Redis | Memorystore Redis HA |
| **Serving** | Go binary | Cloud Run |
| **Cost** | Free | ~$700/month |
| **Performance** | Good for testing | Production-grade |
| **Switching** | Change `APP_MODE` env var | Change `APP_MODE` env var |

---

### 11. Troubleshooting

**Ollama not responding:**
```bash
docker logs visionary-ollama
docker restart visionary-ollama
```

**PostgreSQL connection refused:**
```bash
docker-compose ps  # Check if running
docker-compose logs postgres  # Check logs
docker-compose restart postgres
```

**Redis connection failed:**
```bash
docker exec -it visionary-redis redis-cli ping
# Should return: PONG
```

**Embedding dimensions mismatch:**
- nomic-embed-text: 768 dimensions ✅
- Verify: `docker exec visionary-ollama ollama show nomic-embed-text`

---

## 🎉 Summary

You now have **complete flexibility**:

1. **Develop locally** with free, open-source tools
2. **Test thoroughly** before deploying to GCP
3. **Switch to production** by changing environment variables
4. **Keep all code** - no deletions, just configuration switches!

**No GCP credentials needed for local development!** 🚀
