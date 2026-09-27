# Visionary RAG Pipeline - Required Credentials & API Keys

**Complete checklist of all credentials, API keys, and configurations needed for production deployment.**

---

## 📋 Summary

| Category | Item | Required For | Priority |
|----------|------|--------------|----------|
| **GCP Infrastructure** | GCP Project + Billing | Everything | 🔴 Critical |
| **GCP Services** | 9 GCP APIs | Infrastructure | 🔴 Critical |
| **Database** | AlloyDB Connection | Storage | 🔴 Critical |
| **Cache** | Redis Connection | Sessions | 🔴 Critical |
| **ML/AI** | Vertex AI API | Embeddings + LLM | 🔴 Critical |
| **Auth** | JWT Secret | Authentication | 🔴 Critical |
| **Optional** | Cloud Trace | Observability | 🟡 Recommended |
| **Optional** | Cloud Monitoring | Alerts | 🟡 Recommended |

---

## 1. GCP Infrastructure Credentials

### 1.1 GCP Project & Billing

**Required:**
- ✅ GCP Project ID (e.g., `visionary-rag-prod`)
- ✅ Billing Account linked to project
- ✅ Service Account with appropriate permissions

**How to Get:**
```bash
# Create project
gcloud projects create visionary-rag-prod --name="Visionary RAG"

# Link billing
gcloud beta billing projects link visionary-rag-prod \
  --billing-account=YOUR-BILLING-ACCOUNT-ID

# Create service account
gcloud iam service-accounts create visionary-rag-sa \
  --display-name="Visionary RAG Service Account"
```

**Store in:** `terraform.tfvars` (gitignored) or environment variables

---

### 1.2 GCP API Keys (Enable APIs)

**Required APIs (9 total):**

```bash
gcloud services enable \
  alloydb.googleapis.com \
  redis.googleapis.com \
  run.googleapis.com \
  aiplatform.googleapis.com \
  secretmanager.googleapis.com \
  cloudtrace.googleapis.com \
  monitoring.googleapis.com \
  artifactregistry.googleapis.com \
  cloudbuild.googleapis.com
```

**Verify:**
```bash
gcloud services list --enabled | grep -E "alloydb|redis|run|aiplatform"
```

---

## 2. Database Credentials

### 2.1 AlloyDB Connection String

**Format:**
```
host=<ALLOYDB_PRIVATE_IP> port=5432 dbname=visionary user=visionary password=<PASSWORD>
```

**How to Get:**
1. Deploy Terraform (creates AlloyDB cluster)
2. Get private IP from Terraform outputs or:
   ```bash
   gcloud alloydb instances describe visionary-rag-primary \
     --region=asia-south1 --cluster=visionary-rag-cluster \
     --format='value(ipAddress)'
   ```
3. Password is set in `terraform.tfvars`

**Store in:** GCP Secret Manager as `alloydb-dsn`

```bash
echo -n "host=10.0.0.5 port=5432 dbname=visionary user=visionary password=SECRET" \
  | gcloud secrets create alloydb-dsn --data-file=-
```

---

### 2.2 PgBouncer Credentials

**File:** `orchestrator/pgbouncer/userlist.txt`

**Format:**
```
"visionary" "YOUR_SECURE_PASSWORD"
"pgbouncer" "YOUR_ADMIN_PASSWORD"
```

**Generate Password:**
```bash
openssl rand -base64 32
```

**Store in:** GCP Secret Manager as `pgbouncer-auth`

---

## 3. Redis Credentials

### 3.1 Redis Connection String

**Format:**
```
<REDIS_IP>:6379
```

**How to Get:**
1. Deploy Terraform (creates Redis instance)
2. Get IP from Terraform outputs or:
   ```bash
   gcloud redis instances describe visionary-rag-session \
     --region=asia-south1 --format='value(host)'
   ```

**Store in:** GCP Secret Manager as `redis-addr`

```bash
echo -n "10.0.0.6:6379" | gcloud secrets create redis-addr --data-file=-
```

**Password:** Not required for private VPC (Redis in same VPC)

---

## 4. Vertex AI Credentials

### 4.1 Vertex AI Service Account

**Required For:** Embeddings + LLM generation

**Create Service Account:**
```bash
# Create service account
gcloud iam service-accounts create visionary-vertex-ai \
  --display-name="Vertex AI Service Account"

# Grant Vertex AI permissions
gcloud projects add-iam-policy-binding visionary-rag-prod \
  --member="serviceAccount:visionary-vertex-ai@visionary-rag-prod.iam.gserviceaccount.com" \
  --role="roles/aiplatform.user"

# Generate key
gcloud iam service-accounts keys create service_account.json \
  --iam-account=visionary-vertex-ai@visionary-rag-prod.iam.gserviceaccount.com
```

**Store in:** GCP Secret Manager as `vertex-sa`

```bash
gcloud secrets create vertex-sa --data-file=service_account.json
rm service_account.json  # Delete local copy!
```

**Environment Variable:** `GOOGLE_APPLICATION_CREDENTIALS` (for local dev)

---

### 4.2 Vertex AI Quota Increase

**Default Quotas:**
- Embeddings: 60 requests/minute
- LLM: 60 requests/minute

**Required for Production:**
- Embeddings: 600 requests/minute (for ingestion batch)
- LLM: 300 requests/minute (for serving)

**Request Quota Increase:**
1. Go to: https://console.cloud.google.com/vertex-ai/quotas
2. Select region: `asia-south1`
3. Find `Embedding requests per minute`
4. Click "Edit" → Request increase to 600
5. Find `Generate text requests per minute`
6. Click "Edit" → Request increase to 300

**⏱️ Time:** Takes 24-48 hours for approval

---

## 5. Authentication Credentials

### 5.1 JWT Secret

**Required For:** Student authentication & session management

**Generate:**
```bash
openssl rand -base64 32 > jwt.key
```

**Store in:** GCP Secret Manager as `jwt-secret`

```bash
gcloud secrets create jwt-secret --data-file=jwt.key
rm jwt.key  # Delete local copy!
```

**JWT Claims Structure:**
```json
{
  "user_id": "student-123",
  "grade": 7,
  "subject": "Science",
  "taxonomy_id": 42,
  "session_id": "abc-xyz-123",
  "exp": 1711612800
}
```

---

### 5.2 CBSE Taxonomy Data

**Required For:** Grade/subject/chapter filtering

**Populate via:** Schema migration (already in `schema/v2_production.sql`)

```sql
-- Already included in schema
SELECT COUNT(*) FROM cbse_taxonomy;
-- Expected: 52 chapters (Grade 6-8 Science)
```

---

## 6. Optional Credentials

### 6.1 Cloud Trace (Observability)

**Required For:** Distributed tracing, TTFT monitoring

**Enable API:**
```bash
gcloud services enable cloudtrace.googleapis.com
```

**Service Account Permission:**
```bash
gcloud projects add-iam-policy-binding visionary-rag-prod \
  --member="serviceAccount:visionary-rag-sa@visionary-rag-prod.iam.gserviceaccount.com" \
  --role="roles/cloudtrace.agent"
```

---

### 6.2 Cloud Monitoring (Alerts)

**Required For:** Metrics, dashboards, PagerDuty integration

**Enable API:**
```bash
gcloud services enable monitoring.googleapis.com
```

**Create Alert Policies:**
- p99 TTFT > 500ms
- Error rate > 1%
- DLQ depth > 50
- Redis latency > 5ms

---

### 6.3 Artifact Registry (Container Images)

**Required For:** Storing Docker images

**Already created by Terraform:**
```
asia-south1-docker.pkg.dev/visionary-rag-prod/visionary-rag-images
```

**Push Images:**
```bash
# Ingestion
docker build -t asia-south1-docker.pkg.dev/visionary-rag-prod/visionary-rag-images/ingestion:latest ./ingestion-go
docker push asia-south1-docker.pkg.dev/visionary-rag-prod/visionary-rag-images/ingestion:latest

# Orchestrator
docker build -t asia-south1-docker.pkg.dev/visionary-rag-prod/visionary-rag-images/orchestrator:latest ./orchestrator
docker push asia-south1-docker.pkg.dev/visionary-rag-prod/visionary-rag-images/orchestrator:latest
```

---

## 7. Local Development Setup

### 7.1 Environment Variables (.env.local)

**File:** `.env.local` (gitignored)

```bash
# GCP Project
export GOOGLE_CLOUD_PROJECT="visionary-rag-prod"
export GOOGLE_CLOUD_REGION="asia-south1"

# Database
export ALLOYDB_DSN="host=10.0.0.5 port=5432 dbname=visionary user=visionary password=SECRET"

# Redis
export REDIS_ADDR="10.0.0.6:6379"

# Vertex AI
export VERTEX_PROJECT="visionary-rag-prod"
export VERTEX_LOCATION="asia-south1"
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/service_account.json"

# JWT
export JWT_SECRET="your-base64-secret-here"

# Logging
export LOG_LEVEL="debug"
export ENVIRONMENT="development"
```

**Load in development:**
```bash
source .env.local
```

---

### 7.2 gcloud Authentication

**For local development:**
```bash
gcloud auth application-default login
gcloud config set project visionary-rag-prod
```

---

## 8. Production Deployment Checklist

### 8.1 Pre-Deployment

- [ ] GCP project created
- [ ] Billing account linked
- [ ] All 9 APIs enabled
- [ ] Service account created with permissions
- [ ] Vertex AI quota increase requested (24-48 hours)

### 8.2 Infrastructure Deployment

```bash
# 1. Create terraform.tfvars
cat > terraform/terraform.tfvars <<EOF
gcp_project_id       = "visionary-rag-prod"
region               = "asia-south1"
alloydb_initial_password = "$(openssl rand -base64 32)"
EOF

# 2. Apply Terraform
cd terraform
terraform init
terraform apply

# 3. Populate Secret Manager
export ALLOYDB_IP=$(gcloud alloydb instances describe visionary-rag-primary \
  --region=asia-south1 --cluster=visionary-rag-cluster \
  --format='value(ipAddress)')

export REDIS_HOST=$(gcloud redis instances describe visionary-rag-session \
  --region=asia-south1 --format='value(host)')

echo -n "host=${ALLOYDB_IP} dbname=visionary user=visionary password=PASSWORD" \
  | gcloud secrets create alloydb-dsn --data-file=-

echo -n "${REDIS_HOST}:6379" \
  | gcloud secrets create redis-addr --data-file=-

gcloud secrets create vertex-sa --data-file=service_account.json
gcloud secrets create jwt-secret --data-file=<(openssl rand -base64 32)
```

### 8.3 Database Setup

```bash
# Get connection string
export ALLOYDB_DSN=$(gcloud secrets versions access latest --secret=alloydb-dsn)

# Apply schema
psql "$ALLOYDB_DSN" -f ../schema/v2_production.sql

# Verify
psql "$ALLOYDB_DSN" -c "SELECT COUNT(*) FROM cbse_taxonomy;"
# Expected: 52
```

### 8.4 Deploy Services

```bash
# Build and push images
cd ../ingestion-go
docker build -t asia-south1-docker.pkg.dev/visionary-rag-prod/visionary-rag-images/ingestion:latest .
docker push asia-south1-docker.pkg.dev/visionary-rag-prod/visionary-rag-images/ingestion:latest

cd ../orchestrator
docker build -t asia-south1-docker.pkg.dev/visionary-rag-prod/visionary-rag-images/orchestrator:latest .
docker push asia-south1-docker.pkg.dev/visionary-rag-prod/visionary-rag-images/orchestrator:latest

# Deploy orchestrator to Cloud Run
gcloud run deploy visionary-rag-orchestrator \
  --image asia-south1-docker.pkg.dev/visionary-rag-prod/visionary-rag-images/orchestrator:latest \
  --region asia-south1 \
  --min-instances 2 \
  --max-instances 100 \
  --cpu 2 \
  --memory 2Gi \
  --timeout 450ms \
  --set-env-vars GOOGLE_CLOUD_PROJECT=visionary-rag-prod,VERTEX_LOCATION=asia-south1 \
  --set-secrets ALLOYDB_DSN=alloydb-dsn:latest,REDIS_ADDR=redis-addr:latest,JWT_SECRET=jwt-secret:latest \
  --vpc-connector visionary-rag-vpc-connector \
  --vpc-egress private-ranges-only

# Deploy ingestion as Cloud Run Job
gcloud run jobs deploy visionary-rag-ingestion \
  --image asia-south1-docker.pkg.dev/visionary-rag-prod/visionary-rag-images/ingestion:latest \
  --region asia-south1 \
  --task-timeout 3600s \
  --set-env-vars GOOGLE_CLOUD_PROJECT=visionary-rag-prod,VERTEX_LOCATION=asia-south1 \
  --set-secrets ALLOYDB_DSN=alloydb-dsn:latest
```

---

## 9. Credential Storage Summary

| Credential | Storage Location | Access Method |
|------------|-----------------|---------------|
| AlloyDB DSN | Secret Manager | Env var injection |
| Redis Addr | Secret Manager | Env var injection |
| Vertex AI SA | Secret Manager | File mount |
| JWT Secret | Secret Manager | Env var injection |
| PgBouncer Auth | Secret Manager | File mount |
| Terraform State | GCS Bucket | Backend config |
| Service Account Keys | Never stored locally | Workload Identity |

---

## 10. Security Best Practices

### ✅ DO:
- Use Secret Manager for all secrets
- Use Workload Identity (not service account keys) in production
- Rotate secrets every 90 days
- Use principle of least privilege for IAM
- Enable VPC Service Controls
- Use private IPs only (no public endpoints)

### ❌ DON'T:
- Commit secrets to Git
- Store service account keys locally
- Use default service accounts
- Share credentials via chat/email
- Use same password across environments

---

## 11. Testing Credentials Locally

### Quick Test Script

```bash
#!/bin/bash
# test-credentials.sh

echo "Testing GCP credentials..."

# Test 1: Check gcloud auth
if gcloud auth list --filter=status:ACTIVE --format="value(account)" > /dev/null; then
  echo "✅ gcloud authenticated"
else
  echo "❌ gcloud not authenticated"
  exit 1
fi

# Test 2: Check project
PROJECT=$(gcloud config get-value project)
if [ -n "$PROJECT" ]; then
  echo "✅ Project set: $PROJECT"
else
  echo "❌ Project not set"
  exit 1
fi

# Test 3: Check APIs
for API in alloydb redis run aiplatform secretmanager; do
  if gcloud services list --enabled --filter="$API" --format="value(name)" | grep -q "$API"; then
    echo "✅ API enabled: $API"
  else
    echo "❌ API not enabled: $API"
  fi
done

# Test 4: Check Secret Manager access
if gcloud secrets list > /dev/null 2>&1; then
  echo "✅ Secret Manager accessible"
else
  echo "❌ Cannot access Secret Manager"
fi

echo "Credential test complete!"
```

---

## 12. Troubleshooting

### Common Issues

**1. "Permission denied" errors**
```bash
# Check service account permissions
gcloud projects get-iam-policy visionary-rag-prod \
  --flatten="bindings[].members" \
  --format="table(bindings.role)" \
  --filter="bindings.members:visionary-rag-sa"
```

**2. "API not enabled" errors**
```bash
# Enable missing API
gcloud services enable <API_NAME>
```

**3. "Quota exceeded" errors**
- Request quota increase in Vertex AI console
- Reduce batch size temporarily

**4. "Cannot connect to AlloyDB"**
- Check VPC peering is active
- Verify private IP is correct
- Check firewall rules allow port 5432

---

## 13. Cost Estimates

| Service | Monthly Cost (Estimated) |
|---------|-------------------------|
| AlloyDB (8 vCPU, 32GB) | ~$500/month |
| Redis (4GB HA) | ~$150/month |
| Cloud Run (2 min instances) | ~$50/month |
| Vertex AI Embeddings | ~$0.02 per 1K tokens |
| Vertex AI LLM | ~$0.0001 per 1K tokens |
| Secret Manager | ~$0.50/month |
| **Total (excluding usage)** | **~$700/month** |

**Usage-based costs vary based on:**
- Number of students
- Query volume
- Ingestion frequency

---

## 14. Complete Credential Checklist

Before going to production, ensure you have:

### Infrastructure
- [ ] GCP Project ID
- [ ] Billing Account ID
- [ ] Service Account created
- [ ] All 9 APIs enabled

### Database
- [ ] AlloyDB connection string
- [ ] Database password
- [ ] PgBouncer credentials
- [ ] Schema applied

### Cache
- [ ] Redis connection string
- [ ] Redis IP address

### ML/AI
- [ ] Vertex AI service account
- [ ] Service account key (JSON)
- [ ] Quota increase approved

### Authentication
- [ ] JWT secret key
- [ ] JWT expiration configured

### Deployment
- [ ] Artifact Registry created
- [ ] Docker images built
- [ ] Cloud Run deployed
- [ ] Cloud Run Job deployed

### Monitoring
- [ ] Cloud Trace enabled
- [ ] Cloud Monitoring enabled
- [ ] Alert policies created

---

**📞 Need Help?** Contact: engineering@visionary.edu

**📅 Last Updated:** March 27, 2026

**📄 Version:** 1.0
