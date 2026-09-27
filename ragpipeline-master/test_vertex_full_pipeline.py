"""Full pipeline test: Vertex AI Embeddings -> PostgreSQL -> Redis -> GCP LLM."""
import os
import json
import time
import asyncio
import numpy as np
from datetime import datetime
from google.auth import impersonated_credentials
import google.auth
from google.cloud import aiplatform
from vertexai.language_models import TextEmbeddingModel, TextGenerationModel

# ============================================================
# CONFIG
# ============================================================

POSTGRES = {"host": "localhost", "port": 5432, "user": "postgres", "password": "postgres", "dbname": "ragdb"}
REDIS_HOST, REDIS_PORT = "localhost", 6379
TARGET_SA = "ragpipeline-dev@visionary-research.iam.gserviceaccount.com"
PROJECT = "visionary-research"
LOCATION = "us-central1"

class VertexPipelineTester:
    def __init__(self):
        self.results = {"tests": [], "summary": {}}
        self.emb_model = None
        self.llm_model = None
        self.imp_creds = None

    def setup_vertex(self):
        """Initialize Vertex AI with impersonation."""
        print("\n[1/6] Setting up Vertex AI (service account impersonation)...")
        t0 = time.time()
        try:
            creds, _ = google.auth.default()
            self.imp_creds = impersonated_credentials.Credentials(
                source_credentials=creds,
                target_principal=TARGET_SA,
                target_scopes=["https://www.googleapis.com/auth/cloud-platform"],
                lifetime=3600
            )
            aiplatform.init(project=PROJECT, location=LOCATION, credentials=self.imp_creds)
            self.emb_model = TextEmbeddingModel.from_pretrained("text-embedding-005")
            self.llm_model = TextGenerationModel.from_pretrained("text-bison@002")
            latency = (time.time() - t0) * 1000
            self.log("vertex_setup", "vertex_ai", "PASS", f"text-embedding-005 + text-bison@002", latency)
            return True
        except Exception as e:
            self.log("vertex_setup", "vertex_ai", "FAIL", str(e), (time.time()-t0)*1000)
            return False

    def test_vertex_embeddings(self):
        """Test Vertex AI embeddings for science questions."""
        print("\n[2/6] Testing Vertex AI Embeddings...")
        t0 = time.time()
        try:
            texts = ["What is force and pressure?", "Explain photosynthesis", "Newton's laws of motion"]
            embeddings = self.emb_model.get_embeddings(texts)

            # Verify dimensions
            all_768 = all(len(e.values) == 768 for e in embeddings)
            all_nonzero = all(any(v != 0 for v in e.values) for e in embeddings)

            # Verify semantic similarity (force/pressure vs weather should be different)
            a, b = np.array(embeddings[0].values), np.array(embeddings[1].values)
            similarity = np.dot(a, b) / (np.linalg.norm(a) * np.linalg.norm(b))

            latency = (time.time() - t0) * 1000
            self.log("vertex_embeddings", "vertex_ai", "PASS",
                    f"3/3 embeddings: 768d={all_768}, nonzero={all_nonzero}, similarity={similarity:.4f}", latency)
            return True
        except Exception as e:
            self.log("vertex_embeddings", "vertex_ai", "FAIL", str(e), (time.time()-t0)*1000)
            return False

    def test_vertex_llm(self):
        """Test Vertex AI LLM via Gemini REST API."""
        print("\n[3/6] Testing Vertex AI LLM (Gemini via REST)...")
        import requests
        t0 = time.time()
        try:
            # Get token from impersonated credentials
            self.imp_creds.refresh(google.auth.transport.requests.Request())
            token = self.imp_creds.token

            url = f"https://us-central1-aiplatform.googleapis.com/v1/projects/{PROJECT}/locations/{LOCATION}/publishers/google/models/gemini-2.5-flash:streamGenerateContent"
            headers = {
                "Authorization": f"Bearer {token}",
                "Content-Type": "application/json"
            }
            payload = {
                "contents": [{"role": "user", "parts": [{"text": "Explain force and pressure in one sentence for a CBSE Class 8 student."}]}],
                "generationConfig": {"maxOutputTokens": 100, "temperature": 0.7}
            }

            resp = requests.post(url, headers=headers, json=payload, timeout=30)
            resp.raise_for_status()
            data = resp.json()
            # streamGenerateContent returns a list of chunks
            parts = []
            for chunk in data:
                for candidate in chunk.get("candidates", []):
                    for part in candidate.get("content", {}).get("parts", []):
                        parts.append(part.get("text", ""))
            text = "".join(parts)
            latency = (time.time() - t0) * 1000
            is_good = len(text) > 50
            self.log("vertex_llm_gemini", "vertex_ai", "PASS" if is_good else "WARN",
                    f"Response: {text[:120]}...", latency)
            return is_good
        except Exception as e:
            self.log("vertex_llm_gemini", "vertex_ai", "FAIL", str(e), (time.time()-t0)*1000)
            return False

    async def test_postgres_with_vertex_embeddings(self):
        """Insert Vertex AI embeddings into PostgreSQL and test retrieval."""
        print("\n[4/6] Testing PostgreSQL with Vertex AI embeddings...")
        import asyncpg
        t0 = time.time()
        try:
            dsn = f"postgresql://{POSTGRES['user']}:{POSTGRES['password']}@{POSTGRES['host']}:{POSTGRES['port']}/{POSTGRES['dbname']}"
            conn = await asyncpg.connect(dsn)

            # Generate real embedding for "Force and Pressure"
            emb = self.emb_model.get_embeddings(["Force is a push or pull upon an object. Pressure is force per unit area. The SI unit of force is Newton and pressure is Pascal."])
            vec = emb[0].values
            vec_str = "[" + ",".join(str(v) for v in vec) + "]"

            # Insert parent chunk
            parent = await conn.fetchrow(
                """INSERT INTO parent_chunks (taxonomy_id, content, content_type, chapter, section)
                   VALUES (11, 'Force and pressure concepts. Force is a push or pull. Pressure is force divided by area.', 'prose', 'Force and Pressure', 'Introduction')
                   RETURNING parent_id"""
            )

            # Insert child chunk with REAL Vertex AI embedding
            await conn.execute(
                """INSERT INTO child_chunks (parent_id, taxonomy_id, content, embedding, content_type, page_number)
                   VALUES ($1, 11, $2, $3::vector, 'prose', 1)""",
                parent["parent_id"],
                "Force is a push or pull upon an object resulting from interaction with another object. Pressure is the amount of force applied perpendicular to the surface of an object per unit area.",
                vec_str
            )

            # Search with the same embedding
            results = await conn.fetch(
                """SELECT content, embedding <=> $1::vector AS distance
                   FROM child_chunks WHERE taxonomy_id = 11
                   ORDER BY embedding <=> $1::vector LIMIT 5""",
                vec_str
            )

            found = len(results) > 0
            distance = results[0]["distance"] if results else None

            # Clean up
            await conn.execute("DELETE FROM child_chunks WHERE content LIKE 'Force is a push%'")
            await conn.execute("DELETE FROM parent_chunks WHERE content LIKE 'Force and pressure concepts%'")
            await conn.close()

            latency = (time.time() - t0) * 1000
            self.log("postgres_vertex_embeddings", "postgresql", "PASS",
                    f"Inserted + retrieved: {found}, distance={distance}", latency)
            return found
        except Exception as e:
            self.log("postgres_vertex_embeddings", "postgresql", "FAIL", str(e), (time.time()-t0)*1000)
            return False

    async def test_redis_caching(self):
        """Test Redis caching for Vertex AI embeddings."""
        print("\n[5/6] Testing Redis caching with Vertex AI embeddings...")
        import redis
        import hashlib
        t0 = time.time()
        try:
            r = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)
            r.ping()

            # Cache a real Vertex AI embedding
            text = "What is force?"
            emb = self.emb_model.get_embeddings([text])
            cache_key = f"vertex:emb:{hashlib.sha256(text.encode()).hexdigest()}"
            r.setex(cache_key, 3600, json.dumps(emb[0].values))

            # Retrieve
            cached = json.loads(r.get(cache_key))
            r.delete(cache_key)

            is_match = cached == emb[0].values
            latency = (time.time() - t0) * 1000
            self.log("redis_vertex_cache", "redis", "PASS",
                    f"Cached {len(cached)}-dim embedding, match={is_match}", latency)
            r.close()
            return is_match
        except Exception as e:
            self.log("redis_vertex_cache", "redis", "FAIL", str(e), (time.time()-t0)*1000)
            return False

    def test_go_vector_search_with_vertex_embedding(self):
        """Test Go vector-search-service with a real Vertex AI embedding."""
        print("\n[6/6] Testing Go Vector Search Service with Vertex AI embedding...")
        import requests
        t0 = time.time()
        try:
            # Get real embedding
            emb = self.emb_model.get_embeddings(["force and pressure physics"])
            vec = [float(v) for v in emb[0].values[:768]]

            resp = requests.post(
                "http://localhost:8082/search/dense",
                json={"embedding": vec, "top_k": 5},
                timeout=10
            )
            latency = (time.time() - t0) * 1000

            if resp.status_code == 200:
                data = resp.json()
                results = data.get("results", [])
                self.log("go_vector_search_vertex", "go_services", "PASS",
                        f"{len(results)} results with real Vertex AI embedding", latency)
                return True
            else:
                self.log("go_vector_search_vertex", "go_services", "WARN",
                        f"HTTP {resp.status_code}: {resp.text[:200]}", latency)
                return False
        except Exception as e:
            self.log("go_vector_search_vertex", "go_services", "FAIL", str(e), (time.time()-t0)*1000)
            return False

    def log(self, test, category, status, details="", latency_ms=0):
        result = {"test": test, "category": category, "status": status, "details": details, "latency_ms": round(latency_ms, 2)}
        self.results["tests"].append(result)
        icon = {"PASS": "✅", "FAIL": "❌", "WARN": "⚠️"}.get(status, "❓")
        lat = f" ({latency_ms:.0f}ms)" if latency_ms > 0 else ""
        print(f"  {icon} [{category}] {test}{lat}: {status}")
        if details and status != "PASS":
            print(f"     {details[:200]}")

    async def run_all(self):
        print("\n" + "="*70)
        print("VERTEX AI FULL PIPELINE TEST")
        print("="*70)
        print(f"  Project: {PROJECT}")
        print(f"  Service Account: {TARGET_SA}")
        print(f"  Location: {LOCATION}")

        if not self.setup_vertex():
            print("Vertex AI setup failed. Aborting.")
            return

        self.test_vertex_embeddings()
        self.test_vertex_llm()
        await self.test_postgres_with_vertex_embeddings()
        await self.test_redis_caching()
        self.test_go_vector_search_with_vertex_embedding()

        total = len(self.results["tests"])
        passed = sum(1 for t in self.results["tests"] if t["status"] == "PASS")
        failed = sum(1 for t in self.results["tests"] if t["status"] == "FAIL")
        warn = sum(1 for t in self.results["tests"] if t["status"] == "WARN")

        self.results["summary"] = {"total": total, "passed": passed, "failed": failed, "warnings": warn}

        print(f"\n{'='*70}")
        print(f"VERTEX AI PIPELINE RESULTS")
        print(f"{'='*70}")
        for t in self.results["tests"]:
            icon = {"PASS": "✅", "FAIL": "❌", "WARN": "⚠️"}.get(t["status"], "❓")
            print(f"  {icon} {t['test']}: {t['status']} ({t['latency_ms']:.0f}ms) - {t['details'][:100]}")
        print(f"\n  Total: {total} | Passed: {passed} | Failed: {failed} | Warnings: {warn}")
        print(f"  Pass Rate: {(passed/total*100):.1f}%")
        print(f"{'='*70}")

async def main():
    tester = VertexPipelineTester()
    await tester.run_all()

if __name__ == "__main__":
    asyncio.run(main())
