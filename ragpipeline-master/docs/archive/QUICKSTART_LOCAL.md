# Quick Start - Local Development (5 Minutes)

**Get the Visionary RAG Pipeline running locally WITHOUT any GCP credentials!**

---

## 🚀 One-Command Setup (Windows)

```powershell
# Run the automated setup
.\scripts\setup-local-windows.ps1
```

This will:
1. ✅ Start PostgreSQL + pgvector
2. ✅ Start Redis
3. ✅ Start Ollama (for embeddings + LLM)
4. ✅ Pull required models (nomic-embed-text, llama3.2:3b)
5. ✅ Initialize database with schema
6. ✅ Create `.env.local` configuration

---

## 📦 What You Get

| Service | Local Alternative | Port |
|---------|------------------|------|
| **Database** | PostgreSQL 15 + pgvector | 5432 |
| **Cache** | Redis 7 | 6379 |
| **Embeddings** | Ollama (nomic-embed-text) | 11434 |
| **LLM** | Ollama (llama3.2:3b) | 11434 |
| **PG Admin** | Database GUI | 5050 |
| **Redis Commander** | Redis GUI | 8081 |

**All FREE, all running locally!**

---

## 🎯 Usage

### 1. Start Services

```bash
docker-compose up -d
```

### 2. Set Environment

```bash
# Copy example config
cp .env.local.example .env.local

# Load environment (PowerShell)
Get-Content .env.local | ForEach-Object {
  if ($_ -match '^\s*([^#][^=]+)=(.+)\s*$') {
    [Environment]::SetEnvironmentVariable($Matches[1].Trim(), $Matches[2].Trim())
  }
}
```

### 3. Run Ingestion

```bash
cd ingestion-go
go run cmd/ingestion/main.go --pdf path/to/textbook.pdf --grade 7 --subject Science --taxonomy-id 17
```

### 4. Run Orchestrator

```bash
cd orchestrator
go run cmd/server/main.go
```

### 5. Test Query

```bash
curl -X POST http://localhost:8080/query \
  -H "Content-Type: application/json" \
  -d "{\"query\":\"What is photosynthesis?\",\"session_id\":\"test\"}"
```

---

## 🔄 Switch to Production

Just change environment variables:

```bash
# Set to production mode
$env:APP_MODE="production"

# Set production credentials
$env:ALLOYDB_DSN="host=production-ip port=5432 ..."
$env:REDIS_ADDR="production-redis:6379"
$env:GOOGLE_CLOUD_PROJECT="production-project"
$env:USE_LOCAL_EMBEDDINGS="false"
$env:USE_LOCAL_LLM="false"

# Run same binaries
./ingestion-go/ingestion.exe
./orchestrator/orchestrator.exe
```

**No code changes needed!** ✅

---

## 📊 Comparison

| Feature | Local | Production |
|---------|-------|------------|
| **Cost** | FREE | ~$700/month |
| **Setup Time** | 5 minutes | 1-2 days |
| **Credentials** | None | GCP required |
| **Performance** | Good for dev | Production-grade |
| **Data Persistence** | Local volumes | Managed services |

---

## 🛠️ Troubleshooting

**Services not starting?**
```bash
docker-compose down -v
docker-compose up -d
docker-compose logs
```

**Ollama models not downloading?**
```bash
docker exec visionary-ollama ollama pull nomic-embed-text
docker exec visionary-ollama ollama pull llama3.2:3b
```

**Database connection failed?**
```bash
docker exec visionary-postgres pg_isready -U visionary
# Should return: accepting connections
```

---

## 📖 Full Documentation

- **Complete Setup Guide**: `LOCAL_DEVELOPMENT_SETUP.md`
- **Credentials Guide**: `CREDENTIALS_REQUIRED.md`
- **Main README**: `README.md`

---

## ✅ What's Installed

After setup, you'll have:

```bash
# Check running containers
docker-compose ps

# Expected output:
# visionary-postgres    Up (healthy)
# visionary-redis       Up (healthy)
# visionary-ollama      Up (healthy)
```

**Ready to develop!** 🎉

---

**Next Step**: Run `.\scripts\setup-local-windows.ps1` to get started!
