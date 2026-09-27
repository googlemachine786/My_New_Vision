# Phase 1 Deployment Guide

**Status**: Files Ready · **Next**: Manual GCP Deployment Steps

This guide covers the manual steps required to deploy Phase 1 infrastructure on GCP.

---

## Prerequisites Checklist

```bash
# Verify gcloud is installed and authenticated
gcloud version
gcloud auth list

# Verify Terraform is installed
terraform version

# Verify psql client is installed
psql --version

# Verify redis-cli is installed  
redis-cli --version
```

---

## Step 1: Create GCP Project (T1.1)

```bash
# Create new project (or use existing)
gcloud projects create visionary-rag-prod --name="Visionary RAG Pipeline"

# Set as active project
gcloud config set project visionary-rag-prod

# Link billing account (required for AlloyDB)
gcloud beta billing projects link visionary-rag-prod \
  --billing-account=YOUR-BILLING-ACCOUNT-ID

# Enable required APIs (9 APIs)
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

# Verify APIs are enabled
gcloud services list --enabled | grep -E "alloydb|redis|run|aiplatform"
```

**Exit Criteria:**
- ✅ Project created and billing linked
- ✅ All 9 APIs show as enabled

---

## Step 2: Initialize Terraform (T1.2)

```bash
cd terraform/

# Initialize Terraform (downloads providers)
terraform init

# Create terraform.tfvars with your configuration
cat > terraform.tfvars <<EOF
gcp_project_id       = "visionary-rag-prod"
region               = "asia-south1"
zone                 = "asia-south1-a"
alloydb_initial_password = "$(openssl rand -base64 32)"
EOF

# IMPORTANT: Add terraform.tfvars to .gitignore (contains password)
echo "terraform.tfvars" >> ../.gitignore

# Review configuration
terraform plan -out=tfplan

# Expected output: 9 resources to create, 0 errors
```

**Exit Criteria:**
- ✅ `terraform plan` shows 9 resources, 0 errors
- ✅ `terraform.tfvars` is in `.gitignore`

---

## Step 3: Apply Terraform (T1.3)

```bash
# Staged apply for better error isolation
terraform apply -target=google_compute_network.vpc
terraform apply -target=google_alloydb_cluster.visionary
terraform apply  # Remaining resources

# Verify resources created
gcloud alloydb clusters list --region=asia-south1
gcloud redis instances list --region=asia-south1
gcloud run services list --region=asia-south1
```

**Expected Output:**
```
AlloyDB Cluster: visionary-rag-cluster (REGIONAL)
AlloyDB Instance: visionary-rag-primary (n2-standard-8)
Redis Instance: visionary-rag-session (4GB, REDIS_7_0)
Cloud Run Service: visionary-rag-orchestrator
```

**Exit Criteria:**
- ✅ AlloyDB REGIONAL cluster running
- ✅ Redis 7.x instance running (4GB, STANDARD_HA)
- ✅ Cloud Run service created (min=2 instances)

---

## Step 4: Apply Database Schema (T1.4)

```bash
# Get AlloyDB connection details
export ALLOYDB_IP=$(gcloud alloydb instances describe visionary-rag-primary \
  --region=asia-south1 --cluster=visionary-rag-cluster \
  --format='value(ipAddress)')

export ALLOYDB_DSN="host=${ALLOYDB_IP} dbname=visionary user=visionary password=YOUR_PASSWORD"

# Apply schema
psql "$ALLOYDB_DSN" -f ../schema/v2_production.sql 2>&1 | grep -E "ERROR|WARNING|CREATE"

# Expected: Zero ERROR lines
```

**Exit Criteria:**
- ✅ Zero ERROR lines in output
- ✅ All tables created
- ✅ All indexes created (ScaNN, GIN, B-tree)

---

## Step 5: Verify Schema (T1.5)

```bash
# Run verification script
chmod +x ../scripts/verify-schema.sh
../scripts/verify-schema.sh

# Expected: All 10 tests pass
```

**Manual Verification Queries:**
```sql
-- Verify TEXT[] type
SELECT data_type, udt_name FROM information_schema.columns
WHERE table_name = 'parent_chunks' AND column_name = 'extracted_keywords';
-- Expected: ARRAY, _text

-- Verify UUID[] type
SELECT data_type, udt_name FROM information_schema.columns
WHERE table_name = 'ai_feedback_loop' AND column_name = 'retrieved_context';
-- Expected: ARRAY, _uuid

-- Verify ScaNN index
SELECT indexname FROM pg_indexes
WHERE tablename = 'child_chunks' AND indexname = 'idx_child_embedding_scann';
-- Expected: 1 row returned

-- Verify GIN index
SELECT indexname FROM pg_indexes
WHERE tablename = 'parent_chunks' AND indexname = 'idx_parent_keywords_gin';
-- Expected: 1 row returned
```

**Exit Criteria:**
- ✅ `extracted_keywords` is TEXT[] (ARRAY, _text)
- ✅ `retrieved_context` is UUID[] (ARRAY, _uuid)
- ✅ ScaNN index `idx_child_embedding_scann` exists
- ✅ GIN index `idx_parent_keywords_gin` exists

---

## Step 6: Deploy PgBouncer Sidecar (T1.6)

PgBouncer runs as a sidecar container in Cloud Run. Update the Cloud Run service:

```bash
# Create userlist.txt with actual credentials
cat > orchestrator/pgbouncer/userlist.txt <<EOF
"visionary" "YOUR_SECURE_PASSWORD"
"pgbouncer" "YOUR_ADMIN_PASSWORD"
EOF

# Upload to Secret Manager
gcloud secrets create pgbouncer-auth --data-file=orchestrator/pgbouncer/userlist.txt

# Update Cloud Run service to include PgBouncer sidecar
# (Requires Cloud Run YAML configuration - see Cloud Run docs for multi-container deployment)
```

**Alternative: Direct AlloyDB Connection (Development Only)**

For development/testing, connect directly to AlloyDB without PgBouncer:
```bash
export ALLOYDB_DSN="host=${ALLOYDB_IP} port=5432 dbname=visionary user=visionary password=YOUR_PASSWORD"
```

**Exit Criteria:**
- ✅ `pgbouncer.ini` configured (default_pool_size=18, max_client_conn=10000)
- ✅ `userlist.txt` created with credentials
- ✅ Secret `pgbouncer-auth` created in Secret Manager

---

## Step 7: Verify PgBouncer Pool (T1.7)

```bash
# Connect through PgBouncer (once deployed)
psql -h localhost -p 6432 -U pgbouncer pgbouncer -c "SHOW POOLS;"
psql -h localhost -p 6432 -U pgbouncer pgbouncer -c "SHOW CONFIG;"

# Expected:
# default_pool_size = 18
# max_client_conn = 10000
```

**Exit Criteria:**
- ✅ `default_pool_size = 18` (not 10000)
- ✅ `max_client_conn = 10000`
- ✅ Pool shows 18 idle server connections

---

## Step 8: Verify Redis Latency (T1.8)

```bash
# Get Redis IP
export REDIS_HOST=$(gcloud redis instances describe visionary-rag-session \
  --region=asia-south1 --format='value(host)')

export REDIS_ADDR="${REDIS_HOST}:6379"

# Run verification script
chmod +x ../scripts/verify-redis.sh
../scripts/verify-redis.sh

# Expected: p50 < 1ms, max < 2ms
```

**Manual Verification:**
```bash
redis-cli -h $REDIS_HOST ping
# Expected: PONG

redis-cli -h $REDIS_HOST --latency-history -i 1
# Expected: min: 0ms, max: 1ms, avg: 0.3ms
```

**Exit Criteria:**
- ✅ Redis responds to PING with PONG
- ✅ p50 latency < 1ms
- ✅ max latency < 2ms

---

## Step 9: Populate Secret Manager (T1.9)

```bash
# 1. AlloyDB DSN
echo -n "host=${ALLOYDB_IP} dbname=visionary user=visionary password=YOUR_PASSWORD" \
  | gcloud secrets create alloydb-dsn --data-file=-

# 2. Redis Address
echo -n "${REDIS_HOST}:6379" \
  | gcloud secrets create redis-addr --data-file=-

# 3. Vertex AI Service Account
# Create service account with Vertex AI permissions
gcloud iam service-accounts create visionary-vertex-ai \
  --display-name="Visionary Vertex AI Service Account"

gcloud projects add-iam-policy-binding visionary-rag-prod \
  --member="serviceAccount:visionary-vertex-ai@visionary-rag-prod.iam.gserviceaccount.com" \
  --role="roles/aiplatform.user"

# Generate and download key
gcloud iam service-accounts keys create service_account.json \
  --iam-account=visionary-vertex-ai@visionary-rag-prod.iam.gserviceaccount.com

gcloud secrets create vertex-sa --data-file=service_account.json
rm service_account.json  # Delete local copy immediately

# 4. JWT Secret
openssl rand -base64 32 > jwt.key
gcloud secrets create jwt-secret --data-file=jwt.key
rm jwt.key  # Delete local copy immediately

# Verify all secrets created
gcloud secrets list
```

**Exit Criteria:**
- ✅ All 4 secrets created: `alloydb-dsn`, `redis-addr`, `vertex-sa`, `jwt-secret`
- ✅ No local copies of sensitive files remain

---

## Phase 1 Completion Checklist

```
✅ AlloyDB REGIONAL cluster running, ScaNN extension enabled
✅ extracted_keywords is TEXT[] (verified via pg_indexes + column query)
✅ retrieved_context is UUID[] (verified via column query)
✅ ScaNN index on child_chunks.embedding created successfully
✅ GIN index on parent_chunks.extracted_keywords created successfully
✅ PgBouncer default_pool_size = 18, max_client_conn = 10000
✅ Redis p50 < 1ms from VPC connector range
✅ All 4 secrets in Secret Manager
✅ Terraform state saved (in GCS bucket or local .terraform/state)
✅ VPC connector allows Cloud Run → AlloyDB/Redis communication
```

---

## Next Steps: Phase 2 (Python Ingestion)

After completing Phase 1:

1. **Create ingestion/ directory structure**
   - parser/, chunker/, keywords/, embedder/, writer/

2. **Implement 5-pass PDF parser**
   - font_calibrator.py, table_extractor.py, heading_mapper.py, etc.

3. **Build parent-child chunker**
   - 1500-char parents, 512-char children, 15% overlap

4. **Deploy as Cloud Run Job**
   - Batch mode, billed only during execution

See PLAN.md Phase 2 tasks (T2.1–T2.12) for detailed implementation.

---

## Troubleshooting

### AlloyDB Connection Fails
```bash
# Check VPC connector is attached to Cloud Run
gcloud run services describe visionary-rag-orchestrator \
  --region=asia-south1 --format='yaml(spec.template.vpcAccess)'

# Verify AlloyDB allows connections from VPC CIDR
psql "host=${ALLOYDB_IP} dbname=postgres user=postgres" \
  -c "SELECT * FROM pg_hba_file_rules;"
```

### Redis Connection Timeout
```bash
# Verify VPC peering is active
gcloud compute networks peerings list --network=visionary-rag-vpc

# Check Redis authorized network
gcloud redis instances describe visionary-rag-session \
  --region=asia-south1 --format='value(authorizedNetwork)'
```

### ScaNN Index Creation Fails
```bash
# Verify connected to AlloyDB (not Cloud SQL)
psql "$ALLOYDB_DSN" -c "SELECT version();"
# Expected: PostgreSQL 15.x with AlloyDB extensions

# Verify vector extension is installed
psql "$ALLOYDB_DSN" -c "SELECT * FROM pg_extension WHERE extname = 'vector';"
```

---

**Document Version**: 1.0
**Last Updated**: March 27, 2026
**Contact**: engineering@visionary.edu
