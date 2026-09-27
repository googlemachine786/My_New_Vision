# Auto-Detection Setup - Local vs GCP

**The pipeline now automatically detects whether you're running locally or on GCP!**

---

## 🎯 What Changed

### Before
```bash
# Had to manually set everything
export APP_MODE=local
export USE_FAISS=true
export USE_OLLAMA=true
export DATABASE_URL=...
export OLLAMA_BASE_URL=...
```

### After
```bash
# Just run it!
go run cmd/ingestion/main.go --pdf textbook.pdf

# Auto-detects:
# - Running locally? → Uses Ollama + FAISS
# - On Cloud Run? → Uses Vertex AI + AlloyDB
```

---

## 🧠 How Auto-Detection Works

### 5-Level Detection

1. **APP_MODE env var** (explicit override)
   ```bash
   export APP_MODE=local      # Force local
   export APP_MODE=production # Force production
   ```

2. **Cloud Run detection**
   - Checks `K_SERVICE` or `CLOUD_RUN_JOB` env vars
   - If present → **GCP mode**

3. **GCP Project check**
   - Checks `GOOGLE_CLOUD_PROJECT`
   - If real project + credentials → **GCP mode**

4. **Database DSN check**
   - Checks if DSN contains `.alloydb.` or `.googleapis.com`
   - If yes → **GCP mode**

5. **Default**
   - If none above → **Local mode** ✅

---

## 🔄 What Switches Automatically

| Component | Local (Auto) | GCP (Auto) |
|-----------|--------------|------------|
| **Vector DB** | FAISS + PostgreSQL | AlloyDB + ScaNN |
| **Embeddings** | Ollama API | Vertex AI API |
| **LLM** | Ollama API | Vertex AI API |
| **Redis** | Docker (no auth) | Memorystore (auth) |
| **Secrets** | .env file | Secret Manager |
| **Trace** | Disabled | Cloud Trace |

---

## 🚀 Quick Start

### Local Development

```bash
# 1. Start local services
docker-compose up -d

# 2. Run ingestion (auto-detects local)
cd ingestion-go
go run cmd/ingestion/main.go --pdf textbook.pdf --grade 7

# Output:
# ========================================
# Visionary RAG Configuration
# ========================================
# Mode:            local
# Database:        FAISS (local)
# Embeddings:      Ollama (nomic-embed-text)
# LLM:             Ollama (llama3.2:3b)
# ========================================
```

### Production Deployment

```bash
# Deploy to Cloud Run (auto-detects GCP)
gcloud run deploy visionary-rag-orchestrator \
  --image asia-south1-docker.pkg.dev/project/images/orchestrator:latest \
  --set-env-vars GOOGLE_CLOUD_PROJECT=visionary-rag-prod

# Service auto-detects GCP mode:
# Mode:            production
# Database:        AlloyDB + ScaNN (GCP)
# Embeddings:      Vertex AI (text-embedding-005)
# LLM:             Vertex AI (gemini-1.5-flash)
```

---

## 📊 Configuration Output

When you start any service, it prints:

```
========================================
Visionary RAG Configuration
========================================
Mode:            local
Port:            8080
Database:        FAISS (local)
Embeddings:      Ollama (nomic-embed-text)
LLM:             Ollama (llama3.2:3b)
Redis:           Redis Docker (no auth)
Log Level:       debug
========================================
```

---

## 🔧 Manual Override

You can always override auto-detection:

### Force Local Mode
```bash
export APP_MODE=local
export USE_FAISS=true
export USE_OLLAMA=true
export DATABASE_URL="host=localhost port=5432 ..."
```

### Force GCP Mode
```bash
export APP_MODE=production
export GOOGLE_CLOUD_PROJECT=visionary-rag-prod
export ALLOYDB_DSN="host=10.0.0.5 port=5432 ..."
export REDIS_ADDR="production-redis:6379"
export GOOGLE_APPLICATION_CREDENTIALS=~/service-account.json
```

---

## 🎯 Benefits

✅ **Zero Configuration** - Works out of the box  
✅ **No Code Changes** - Same binary everywhere  
✅ **Safe Defaults** - Local mode by default (no accidental GCP charges)  
✅ **Environment Aware** - Detects Cloud Run automatically  
✅ **Easy Override** - Can explicitly set mode if needed  
✅ **All Code Preserved** - Nothing deleted, just smarter detection  

---

## 📝 Example Scenarios

### Scenario 1: Local Development
```bash
# No environment variables
# → Auto-detects as LOCAL
# → Uses: Ollama, FAISS, Docker Redis
go run cmd/server/main.go
```

### Scenario 2: Cloud Run
```yaml
# Cloud Run sets K_SERVICE automatically
# → Auto-detects as GCP
# → Uses: Vertex AI, AlloyDB, Memorystore
gcloud run deploy ...
```

### Scenario 3: Hybrid Testing
```bash
# Test with production database
export APP_MODE=production
export ALLOYDB_DSN="host=production-ip ..."
# → Uses GCP services but runs locally
go run cmd/ingestion/main.go --pdf textbook.pdf
```

---

## 🎉 Summary

**One binary, two modes, zero hassle!**

- Developing locally? → **Just run it**
- Deployed to GCP? → **Auto-detects**
- Need to override? → **Set APP_MODE**

**No configuration headaches!** 🚀

---

## 📖 Related Documentation

- `LOCAL_DEVELOPMENT_SETUP.md` - Complete local setup guide
- `CREDENTIALS_REQUIRED.md` - GCP credentials guide
- `ingestion-go/config/README.md` - Configuration details
