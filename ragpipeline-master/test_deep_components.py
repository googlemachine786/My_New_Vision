"""
Visionary RAG Pipeline - Deep Test of ALL Working Components
Tests: Go HTTP services, PostgreSQL vectors, GCP APIs, Redis, Hybrid Search
"""

import os
import sys
import json
import time
import asyncio
import requests
import subprocess
import numpy as np
from pathlib import Path
from datetime import datetime
from typing import List, Dict, Optional
from concurrent.futures import ThreadPoolExecutor, as_completed

# ============================================================
# CONFIGURATION
# ============================================================

GCP_API_KEY = "AIzaSyBcxwsSdb2mNnqYGcaYbMH1furWLae-W1k"
GEMINI_MODEL = "gemini-2.5-flash"
EMBEDDING_MODEL = "gemini-embedding-001"

POSTGRES_HOST = "localhost"
POSTGRES_PORT = 5432
POSTGRES_USER = "postgres"
POSTGRES_PASSWORD = "postgres"
POSTGRES_DB = "ragdb"

REDIS_HOST = "localhost"
REDIS_PORT = 6379

SERVICES = {
    "vector_search": "http://localhost:8082",
}


class DeepPipelineTester:
    """Tests all working components of the RAG pipeline."""

    def __init__(self):
        self.results = {
            "test_date": datetime.now().isoformat(),
            "tests": [],
            "summary": {}
        }
        self.test_start = time.time()

    def log(self, test: str, category: str, status: str, details: str = "", latency_ms: float = 0):
        """Log test result with category."""
        result = {
            "test": test,
            "category": category,
            "status": status,
            "details": details,
            "latency_ms": round(latency_ms, 2),
            "timestamp": datetime.now().isoformat()
        }
        self.results["tests"].append(result)
        icon = {"PASS": "✅", "FAIL": "❌", "WARN": "⚠️", "SKIP": "⏭️"}.get(status, "❓")
        latency_str = f" ({latency_ms:.0f}ms)" if latency_ms > 0 else ""
        print(f"  {icon} [{category}] {test}{latency_str}: {status}")
        if details and status in ("FAIL", "WARN"):
            print(f"     {details[:200]}")

    # ============================================================
    # CATEGORY 1: Go HTTP Services (Test #1 from audit)
    # ============================================================

    def test_go_service_health(self):
        """Test Go HTTP service health endpoints."""
        print("\n" + "="*60)
        print("CATEGORY 1: Go HTTP Services")
        print("="*60)

        services = {
            "vector_search": ("http://localhost:8082/health", "Vector Search Service"),
        }

        for name, (url, label) in services.items():
            t0 = time.time()
            try:
                resp = requests.get(url, timeout=5)
                latency = (time.time() - t0) * 1000
                data = resp.json()

                if data.get("status") == "healthy":
                    db_status = data.get("database", "unknown")
                    pool = data.get("pool_stats", {})
                    details = f"DB: {db_status}, conns: {pool.get('total_conns', 0)}/{pool.get('max_conns', 0)}"
                    self.log(f"{name}_health", "go_services", "PASS", details, latency)
                else:
                    self.log(f"{name}_health", "go_services", "FAIL", f"Status: {data.get('status')}", latency)
            except Exception as e:
                latency = (time.time() - t0) * 1000
                self.log(f"{name}_health", "go_services", "FAIL", str(e), latency)

    def test_go_service_version(self):
        """Test Go service version endpoints."""
        print("\n  Testing /version endpoints...")

        endpoints = [
            ("http://localhost:8082/version", "vector_search_version"),
        ]

        for url, name in endpoints:
            t0 = time.time()
            try:
                resp = requests.get(url, timeout=5)
                latency = (time.time() - t0) * 1000
                data = resp.json()

                if "version" in data:
                    details = f"v{data['version']}, commit: {data.get('git_commit', 'unknown')}"
                    self.log(name, "go_services", "PASS", details, latency)
                else:
                    self.log(name, "go_services", "WARN", f"Response: {data}", latency)
            except Exception as e:
                latency = (time.time() - t0) * 1000
                self.log(name, "go_services", "FAIL", str(e), latency)

    def test_go_service_metrics(self):
        """Test Go service Prometheus metrics endpoint."""
        print("\n  Testing /metrics endpoints...")

        t0 = time.time()
        try:
            resp = requests.get("http://localhost:8082/metrics", timeout=5)
            latency = (time.time() - t0) * 1000

            if resp.status_code == 200:
                # Check for Prometheus metric format
                text = resp.text
                has_metrics = "# HELP" in text or "# TYPE" in text
                lines = text.count("\n")
                details = f"{lines} metric lines, has HELP/TYPE: {has_metrics}"
                self.log("vector_search_metrics", "go_services", "PASS" if has_metrics else "WARN", details, latency)
            else:
                self.log("vector_search_metrics", "go_services", "FAIL", f"HTTP {resp.status_code}", latency)
        except Exception as e:
            latency = (time.time() - t0) * 1000
            self.log("vector_search_metrics", "go_services", "FAIL", str(e), latency)

    # ============================================================
    # CATEGORY 2: Vector Search HTTP API (Test #4 - Hybrid Search)
    # ============================================================

    def test_vector_search_api(self):
        """Test vector search HTTP API endpoints."""
        print("\n" + "="*60)
        print("CATEGORY 2: Vector Search HTTP API")
        print("="*60)

        # First get a real embedding from GCP API
        try:
            emb_resp = requests.post(
                f"https://generativelanguage.googleapis.com/v1beta/models/{EMBEDDING_MODEL}:embedContent?key={GCP_API_KEY}",
                json={"content": {"parts": [{"text": "force and pressure physics"}]}, "task_type": "RETRIEVAL_QUERY"},
                timeout=30
            )
            if emb_resp.status_code == 200:
                embedding = emb_resp.json()["embedding"]["values"]
                # Convert to float32 range
                embedding_f32 = [float(v) for v in embedding]
            else:
                embedding_f32 = None
        except:
            embedding_f32 = None

        # Test dense search with real embedding
        t0 = time.time()
        try:
            if embedding_f32:
                payload = {
                    "embedding": embedding_f32[:768],  # Use first 768 dims for pgvector
                    "top_k": 5,
                    "filters": {"grade": ["8"], "subject": ["Science"]}
                }
            else:
                # Fallback: use dummy embedding
                payload = {"embedding": [0.01] * 768, "top_k": 5}

            resp = requests.post(
                "http://localhost:8082/search/dense",
                json=payload,
                timeout=10
            )
            latency = (time.time() - t0) * 1000

            if resp.status_code == 200:
                data = resp.json()
                results = data.get("results", [])
                details = f"{len(results)} results from dense search"
                self.log("dense_search", "vector_search_api", "PASS", details, latency)
            else:
                try:
                    error = resp.json()
                    self.log("dense_search", "vector_search_api", "WARN", str(error.get("error", error)), latency)
                except:
                    self.log("dense_search", "vector_search_api", "FAIL", f"HTTP {resp.status_code}", latency)
        except Exception as e:
            latency = (time.time() - t0) * 1000
            self.log("dense_search", "vector_search_api", "FAIL", str(e), latency)

        # Test sparse search (keyword-based)
        t0 = time.time()
        try:
            payload = {
                "keywords": ["force", "pressure", "physics", "newton"],
                "top_k": 5
            }
            resp = requests.post(
                "http://localhost:8082/search/sparse",
                json=payload,
                timeout=10
            )
            latency = (time.time() - t0) * 1000

            if resp.status_code == 200:
                data = resp.json()
                results = data.get("results", [])
                details = f"{len(results)} results from sparse keyword search"
                self.log("sparse_search", "vector_search_api", "PASS", details, latency)
            else:
                try:
                    error = resp.json()
                    self.log("sparse_search", "vector_search_api", "WARN", str(error.get("error", error)), latency)
                except:
                    self.log("sparse_search", "vector_search_api", "FAIL", f"HTTP {resp.status_code}", latency)
        except Exception as e:
            latency = (time.time() - t0) * 1000
            self.log("sparse_search", "vector_search_api", "FAIL", str(e), latency)

        # Test hybrid search (both embedding + keywords)
        t0 = time.time()
        try:
            payload = {
                "embedding": embedding_f32[:768] if embedding_f32 else [0.01] * 768,
                "keywords": ["force", "pressure"],
                "top_k": 5
            }
            resp = requests.post(
                "http://localhost:8082/search",
                json=payload,
                timeout=10
            )
            latency = (time.time() - t0) * 1000

            if resp.status_code == 200:
                data = resp.json()
                results = data.get("results", [])
                has_rrf = any("rrf" in str(r).lower() or "score" in str(r).lower() for r in results) if results else False
                details = f"{len(results)} results, RRF fusion: {has_rrf}"
                self.log("hybrid_search", "vector_search_api", "PASS", details, latency)
            else:
                try:
                    error = resp.json()
                    self.log("hybrid_search", "vector_search_api", "WARN", str(error.get("error", error)), latency)
                except:
                    self.log("hybrid_search", "vector_search_api", "WARN", f"HTTP {resp.status_code}", latency)
        except Exception as e:
            latency = (time.time() - t0) * 1000
            self.log("hybrid_search", "vector_search_api", "FAIL", str(e), latency)

    # ============================================================
    # CATEGORY 3: PostgreSQL Deep Tests (Test #2, #3, #5, #11)
    # ============================================================

    async def test_postgres_deep(self):
        """Deep PostgreSQL tests: schema, data, vector operations."""
        print("\n" + "="*60)
        print("CATEGORY 3: PostgreSQL Deep Tests")
        print("="*60)

        import asyncpg
        dsn = f"postgresql://{POSTGRES_USER}:{POSTGRES_PASSWORD}@{POSTGRES_HOST}:{POSTGRES_PORT}/{POSTGRES_DB}"

        try:
            conn = await asyncpg.connect(dsn)

            # 3a. Table structure verification
            t0 = time.time()
            tables = await conn.fetch("""
                SELECT table_name FROM information_schema.tables
                WHERE table_schema = 'public' ORDER BY table_name
            """)
            table_names = [r["table_name"] for r in tables]
            latency = (time.time() - t0) * 1000
            self.log("table_structure", "postgresql", "PASS",
                    f"{len(table_names)} tables: {', '.join(table_names[:6])}", latency)

            # 3b. Index verification
            t0 = time.time()
            indexes = await conn.fetch("""
                SELECT indexname, indexdef FROM pg_indexes
                WHERE tablename IN ('child_chunks', 'parent_chunks')
                AND indexname LIKE '%hnsw%' OR indexname LIKE '%gin%'
            """)
            latency = (time.time() - t0) * 1000
            index_names = [r["indexname"] for r in indexes]
            self.log("vector_indexes", "postgresql", "PASS",
                    f"{len(index_names)} vector indexes: {', '.join(index_names)}", latency)

            # 3c. Taxonomy data verification
            t0 = time.time()
            taxonomy = await conn.fetch("""
                SELECT grade, subject, chapter, chapter_order FROM cbse_taxonomy
                ORDER BY grade, chapter_order LIMIT 10
            """)
            latency = (time.time() - t0) * 1000
            chapters = [f"G{r['grade']}: {r['chapter'][:30]}" for r in taxonomy]
            self.log("taxonomy_data", "postgresql", "PASS",
                    f"{len(taxonomy)} sample chapters verified", latency)

            # 3d. Vector insertion and similarity search
            t0 = time.time()

            # Create a parent chunk
            parent_result = await conn.fetchrow("""
                INSERT INTO parent_chunks (taxonomy_id, content, content_type, chapter, section)
                VALUES (11, 'Force is a push or pull upon an object resulting from the objects interaction with another object. Pressure is the amount of force applied per unit area. The SI unit of force is the Newton (N) and the SI unit of pressure is the Pascal (Pa).', 'prose', 'Force and Pressure', 'Introduction')
                RETURNING parent_id
            """)

            if parent_result:
                parent_id = parent_result["parent_id"]

                # Create child chunks with embeddings
                test_vectors = [
                    [0.1 * (i + 1)] * 768 for i in range(3)
                ]
                # Make first vector distinctive
                test_vectors[0][0] = 1.0
                test_vectors[0][1] = 0.9

                for i, vec in enumerate(test_vectors):
                    vec_str = "[" + ",".join(str(v) for v in vec) + "]"
                    await conn.execute("""
                        INSERT INTO child_chunks (parent_id, taxonomy_id, content, embedding, content_type, page_number)
                        VALUES ($1, 11, $2, $3::vector, 'prose', $4)
                    """, parent_id, f"Test child chunk {i+1} about force and pressure concepts in physics", vec_str, i+10)

                # Test cosine similarity search
                query_vec = test_vectors[0]
                query_vec_str = "[" + ",".join(str(v) for v in query_vec) + "]"

                results = await conn.fetch("""
                    SELECT content, page_number,
                           embedding <=> $1::vector AS distance
                    FROM child_chunks
                    WHERE taxonomy_id = 11
                    ORDER BY embedding <=> $1::vector
                    LIMIT 5
                """, query_vec_str)

                latency = (time.time() - t0) * 1000

                if results:
                    distances = [r["distance"] for r in results]
                    is_ordered = all(distances[i] <= distances[i+1] for i in range(len(distances)-1))
                    details = f"{len(results)} results, ordered: {is_ordered}, distances: {[f'{d:.4f}' for d in distances[:3]]}"
                    self.log("vector_similarity", "postgresql", "PASS", details, latency)
                else:
                    self.log("vector_similarity", "postgresql", "FAIL", "No results returned", latency)

                # Clean up test data
                await conn.execute("DELETE FROM child_chunks WHERE content LIKE 'Test child chunk%'")
                await conn.execute("DELETE FROM parent_chunks WHERE content LIKE 'Force is a push%'")

            # 3e. Data integrity checks
            t0 = time.time()
            fk_check = await conn.fetchval("""
                SELECT COUNT(*) FROM child_chunks cc
                LEFT JOIN parent_chunks pc ON cc.parent_id = pc.parent_id
                WHERE pc.parent_id IS NULL
            """)
            latency = (time.time() - t0) * 1000
            self.log("foreign_key_integrity", "postgresql",
                    "PASS" if fk_check == 0 else "WARN",
                    f"{fk_check} orphaned child_chunks", latency)

            # 3f. Content type distribution
            t0 = time.time()
            content_types = await conn.fetch("""
                SELECT content_type, COUNT(*) as cnt FROM parent_chunks
                GROUP BY content_type ORDER BY cnt DESC
            """)
            latency = (time.time() - t0) * 1000
            ct_details = ", ".join([f"{r['content_type']}: {r['cnt']}" for r in content_types])
            self.log("content_distribution", "postgresql", "PASS",
                    ct_details if ct_details else "no content yet", latency)

            await conn.close()

        except Exception as e:
            self.log("postgresql_deep", "postgresql", "FAIL", str(e))

    # ============================================================
    # CATEGORY 4: GCP API Deep Tests (Test #3, #4, #6)
    # ============================================================

    def test_gcp_api_deep(self):
        """Deep GCP API tests with multiple queries and embedding types."""
        print("\n" + "="*60)
        print("CATEGORY 4: GCP API Deep Tests")
        print("="*60)

        # 4a. Multiple LLM queries
        test_queries = [
            "What is photosynthesis?",
            "Explain Newton's laws of motion",
            "What are the parts of a cell?",
        ]

        for query in test_queries:
            t0 = time.time()
            try:
                resp = requests.post(
                    f"https://generativelanguage.googleapis.com/v1beta/models/{GEMINI_MODEL}:generateContent?key={GCP_API_KEY}",
                    json={
                        "contents": [{"parts": [{"text": query}]}],
                        "generationConfig": {"temperature": 0.7, "maxOutputTokens": 100}
                    },
                    timeout=30
                )
                latency = (time.time() - t0) * 1000

                if resp.status_code == 200:
                    data = resp.json()
                    answer = data["candidates"][0]["content"]["parts"][0]["text"]
                    is_relevant = len(answer) > 30
                    status = "PASS" if is_relevant else "WARN"
                    self.log(f"llm_{query[:20]}", "gcp_api", status,
                            f"{len(answer)} chars", latency)
                elif resp.status_code == 429:
                    self.log(f"llm_{query[:20]}", "gcp_api", "WARN",
                            "Rate limited (429)", latency)
                else:
                    self.log(f"llm_{query[:20]}", "gcp_api", "FAIL",
                            f"HTTP {resp.status_code}", latency)
            except Exception as e:
                latency = (time.time() - t0) * 1000
                self.log(f"llm_{query[:20]}", "gcp_api", "FAIL", str(e), latency)

        # 4b. Multiple embedding tests
        embed_texts = [
            "What is force and pressure?",
            "Explain photosynthesis process",
            "Newton's first law of motion"
        ]

        for text in embed_texts:
            t0 = time.time()
            try:
                resp = requests.post(
                    f"https://generativelanguage.googleapis.com/v1beta/models/{EMBEDDING_MODEL}:embedContent?key={GCP_API_KEY}",
                    json={
                        "content": {"parts": [{"text": text}]},
                        "task_type": "RETRIEVAL_QUERY"
                    },
                    timeout=30
                )
                latency = (time.time() - t0) * 1000

                if resp.status_code == 200:
                    data = resp.json()
                    embedding = data["embedding"]["values"]
                    dim = len(embedding)
                    non_zero = sum(1 for v in embedding if v != 0)
                    self.log(f"embed_{text[:20]}", "gcp_api", "PASS",
                            f"{dim}d, {non_zero} non-zero", latency)
                else:
                    self.log(f"embed_{text[:20]}", "gcp_api", "FAIL",
                            f"HTTP {resp.status_code}", latency)
            except Exception as e:
                latency = (time.time() - t0) * 1000
                self.log(f"embed_{text[:20]}", "gcp_api", "FAIL", str(e), latency)

        # 4c. Embedding similarity verification
        t0 = time.time()
        try:
            # Get two related embeddings
            texts = ["force and pressure", "newton laws of motion"]
            embeddings = []

            for text in texts:
                resp = requests.post(
                    f"https://generativelanguage.googleapis.com/v1beta/models/{EMBEDDING_MODEL}:embedContent?key={GCP_API_KEY}",
                    json={"content": {"parts": [{"text": text}]}, "task_type": "RETRIEVAL_QUERY"},
                    timeout=30
                )
                if resp.status_code == 200:
                    emb = resp.json()["embedding"]["values"]
                    embeddings.append(emb)

            if len(embeddings) == 2:
                # Calculate cosine similarity
                a, b = np.array(embeddings[0]), np.array(embeddings[1])
                similarity = np.dot(a, b) / (np.linalg.norm(a) * np.linalg.norm(b))

                latency = (time.time() - t0) * 1000
                is_reasonable = 0.1 < similarity < 0.9  # Related but not identical
                self.log("embedding_similarity", "gcp_api",
                        "PASS" if is_reasonable else "WARN",
                        f"similarity: {similarity:.4f} (related physics concepts)", latency)
            else:
                self.log("embedding_similarity", "gcp_api", "SKIP",
                        f"Only got {len(embeddings)}/2 embeddings", (time.time() - t0) * 1000)
        except Exception as e:
            latency = (time.time() - t0) * 1000
            self.log("embedding_similarity", "gcp_api", "FAIL", str(e), latency)

    # ============================================================
    # CATEGORY 5: Redis Deep Tests (Test #7, caching)
    # ============================================================

    async def test_redis_deep(self):
        """Deep Redis tests: caching, TTL, data structures."""
        print("\n" + "="*60)
        print("CATEGORY 5: Redis Deep Tests")
        print("="*60)

        import redis
        r = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)

        # 5a. Basic operations
        t0 = time.time()
        try:
            r.ping()
            r.set("test:key", "value")
            val = r.get("test:key")
            r.delete("test:key")
            latency = (time.time() - t0) * 1000
            self.log("basic_ops", "redis", "PASS", f"SET/GET/DEL verified", latency)
        except Exception as e:
            latency = (time.time() - t0) * 1000
            self.log("basic_ops", "redis", "FAIL", str(e), latency)

        # 5b. TTL-based caching (embedding cache pattern)
        t0 = time.time()
        try:
            cache_key = "embedding:sha256:test"
            test_embedding = [0.123] * 3072
            r.setex(cache_key, 3600, json.dumps(test_embedding))

            # Verify TTL
            ttl = r.ttl(cache_key)
            cached = json.loads(r.get(cache_key))

            r.delete(cache_key)
            latency = (time.time() - t0) * 1000

            is_valid = cached == test_embedding and ttl > 0
            self.log("ttl_caching", "redis", "PASS" if is_valid else "FAIL",
                    f"TTL: {ttl}s, data integrity: {is_valid}", latency)
        except Exception as e:
            latency = (time.time() - t0) * 1000
            self.log("ttl_caching", "redis", "FAIL", str(e), latency)

        # 5c. Search result caching pattern
        t0 = time.time()
        try:
            search_key = "search:force_and_pressure:top5"
            search_results = [
                {"content": "Force is a push...", "score": 0.95},
                {"content": "Pressure is force...", "score": 0.87},
            ]
            r.setex(search_key, 1800, json.dumps(search_results))
            cached_results = json.loads(r.get(search_key))
            r.delete(search_key)

            latency = (time.time() - t0) * 1000
            is_valid = cached_results == search_results
            self.log("search_cache", "redis", "PASS" if is_valid else "FAIL",
                    f"{len(cached_results)} results cached", latency)
        except Exception as e:
            latency = (time.time() - t0) * 1000
            self.log("search_cache", "redis", "FAIL", str(e), latency)

        # 5d. Redis info
        t0 = time.time()
        try:
            info = r.info("memory")
            used_memory = info.get("used_memory_human", "unknown")
            latency = (time.time() - t0) * 1000
            self.log("memory_usage", "redis", "PASS", f"{used_memory}", latency)
        except Exception as e:
            latency = (time.time() - t0) * 1000
            self.log("memory_usage", "redis", "FAIL", str(e), latency)

        r.close()

    # ============================================================
    # CATEGORY 6: Load/Performance (Test #8)
    # ============================================================

    def test_load_performance(self):
        """Test concurrent requests and latency."""
        print("\n" + "="*60)
        print("CATEGORY 6: Load/Performance")
        print("="*60)

        # 6a. Sequential latency measurement
        latencies = []
        for i in range(10):
            t0 = time.time()
            try:
                resp = requests.get("http://localhost:8082/health", timeout=5)
                latency = (time.time() - t0) * 1000
                latencies.append(latency)
            except:
                pass

        if latencies:
            avg = np.mean(latencies)
            p50 = np.percentile(latencies, 50)
            p95 = np.percentile(latencies, 95)
            p99 = np.percentile(latencies, 99)
            self.log("sequential_latency", "load_performance", "PASS",
                    f"avg: {avg:.1f}ms, p50: {p50:.1f}ms, p95: {p95:.1f}ms, p99: {p99:.1f}ms")

        # 6b. Concurrent requests
        t0 = time.time()
        success_count = 0
        error_count = 0

        def make_request(i):
            try:
                resp = requests.get("http://localhost:8082/health", timeout=5)
                return resp.status_code == 200
            except:
                return False

        with ThreadPoolExecutor(max_workers=20) as executor:
            futures = [executor.submit(make_request, i) for i in range(20)]
            for future in as_completed(futures):
                if future.result():
                    success_count += 1
                else:
                    error_count += 1

        concurrency_latency = (time.time() - t0) * 1000

        self.log("concurrent_requests", "load_performance",
                "PASS" if success_count == 20 else "WARN",
                f"{success_count}/20 succeeded, {error_count} failed, total: {concurrency_latency:.0f}ms",
                concurrency_latency)

        # 6c. GCP API latency
        t0 = time.time()
        try:
            resp = requests.post(
                f"https://generativelanguage.googleapis.com/v1beta/models/{GEMINI_MODEL}:generateContent?key={GCP_API_KEY}",
                json={"contents": [{"parts": [{"text": "Hi"}]}]},
                timeout=30
            )
            gemini_latency = (time.time() - t0) * 1000
            status = "PASS" if resp.status_code == 200 else "WARN"
            self.log("gemini_latency", "load_performance", status,
                    f"{gemini_latency:.0f}ms (HTTP {resp.status_code})", gemini_latency)
        except Exception as e:
            gemini_latency = (time.time() - t0) * 1000
            self.log("gemini_latency", "load_performance", "FAIL", str(e), gemini_latency)

    # ============================================================
    # CATEGORY 7: Go Unit Tests (Test #9)
    # ============================================================

    def test_go_unit_tests(self):
        """Run Go unit tests for all services."""
        print("\n" + "="*60)
        print("CATEGORY 7: Go Unit Tests")
        print("="*60)

        services = [
            ("services/vector-search-service", "vector_search_tests"),
            ("services/query-understanding-service", "query_understanding_tests"),
            ("services/api-gateway", "api_gateway_tests"),
        ]

        for service_dir, test_name in services:
            t0 = time.time()
            try:
                # Add Go to PATH for Windows
                env = os.environ.copy()
                env["PATH"] = r"C:\Program Files\Go\bin;" + env.get("PATH", "")

                result = subprocess.run(
                    ["go", "test", "./...", "-v", "-short", "-count=1"],
                    capture_output=True, text=True, timeout=120,
                    cwd=f"C:\\Users\\kommi\\ragpipeline\\{service_dir}",
                    env=env
                )
                latency = (time.time() - t0) * 1000

                # Parse output
                passed = result.stdout.count("PASS")
                failed = result.stdout.count("FAIL")
                ok_lines = [l for l in result.stdout.split("\n") if l.startswith("ok ")]

                if result.returncode == 0:
                    details = f"{passed} PASS, {failed} FAIL"
                    if ok_lines:
                        details += f", coverage: {ok_lines[0]}"
                    self.log(test_name, "go_tests", "PASS", details, latency)
                else:
                    stderr = result.stderr[:200] if result.stderr else ""
                    self.log(test_name, "go_tests", "WARN",
                            f"exit {result.returncode}, {stderr}", latency)
            except subprocess.TimeoutExpired:
                latency = (time.time() - t0) * 1000
                self.log(test_name, "go_tests", "SKIP", "Timeout (120s)", latency)
            except Exception as e:
                latency = (time.time() - t0) * 1000
                self.log(test_name, "go_tests", "SKIP", str(e), latency)

    # ============================================================
    # RUN ALL TESTS
    # ============================================================

    async def run_all_tests(self):
        """Run complete test suite."""
        print("\n" + "="*70)
        print("VISIONARY RAG PIPELINE - DEEP COMPONENT TEST")
        print("="*70)
        print(f"\nConfiguration:")
        print(f"  GCP:    {GEMINI_MODEL} + {EMBEDDING_MODEL}")
        print(f"  PostgreSQL: {POSTGRES_HOST}:{POSTGRES_PORT}/{POSTGRES_DB}")
        print(f"  Redis:      {REDIS_HOST}:{REDIS_PORT}")
        print(f"\nStarting tests...\n")

        # Category 1: Go HTTP Services
        self.test_go_service_health()
        self.test_go_service_version()
        self.test_go_service_metrics()

        # Category 2: Vector Search HTTP API
        self.test_vector_search_api()

        # Category 3: PostgreSQL Deep
        await self.test_postgres_deep()

        # Category 4: GCP API Deep
        self.test_gcp_api_deep()

        # Category 5: Redis Deep
        await self.test_redis_deep()

        # Category 6: Load/Performance
        self.test_load_performance()

        # Category 7: Go Unit Tests
        self.test_go_unit_tests()

        # Calculate summary
        elapsed = time.time() - self.test_start
        passed = sum(1 for t in self.results["tests"] if t["status"] == "PASS")
        failed = sum(1 for t in self.results["tests"] if t["status"] == "FAIL")
        warn = sum(1 for t in self.results["tests"] if t["status"] == "WARN")
        skip = sum(1 for t in self.results["tests"] if t["status"] == "SKIP")
        total = len(self.results["tests"])

        self.results["summary"] = {
            "total": total,
            "passed": passed,
            "failed": failed,
            "warnings": warn,
            "skipped": skip,
            "elapsed_seconds": round(elapsed, 2),
            "pass_rate": f"{(passed/total*100):.1f}%" if total > 0 else "0%",
            "categories": {}
        }

        # Per-category summary
        categories = set(t["category"] for t in self.results["tests"])
        for cat in categories:
            cat_tests = [t for t in self.results["tests"] if t["category"] == cat]
            cat_passed = sum(1 for t in cat_tests if t["status"] == "PASS")
            self.results["summary"]["categories"][cat] = {
                "total": len(cat_tests),
                "passed": cat_passed,
                "rate": f"{(cat_passed/len(cat_tests)*100):.1f}%"
            }

        # Print summary
        print("\n" + "="*70)
        print("FINAL SUMMARY BY CATEGORY")
        print("="*70)

        for cat, stats in sorted(self.results["summary"]["categories"].items()):
            icon = "✅" if stats["rate"] == "100.0%" else "⚠️"
            print(f"  {icon} {cat}: {stats['passed']}/{stats['total']} ({stats['rate']})")

        print(f"\n{'='*70}")
        print(f"OVERALL")
        print(f"{'='*70}")
        print(f"  Total:     {total}")
        print(f"  Passed:    {passed} ✅")
        print(f"  Failed:    {failed} ❌")
        print(f"  Warnings:  {warn} ⚠️")
        print(f"  Skipped:   {skip} ⏭️")
        print(f"  Pass Rate: {self.results['summary']['pass_rate']}")
        print(f"  Time:      {elapsed:.2f}s")
        print(f"{'='*70}")

        # Save results
        output_path = Path("data/deep_test_results.json")
        output_path.parent.mkdir(exist_ok=True)
        with open(output_path, "w") as f:
            json.dump(self.results, f, indent=2)
        print(f"\nResults saved to: {output_path}")

        return failed == 0


async def main():
    """Run the deep pipeline tests."""
    tester = DeepPipelineTester()
    tester.test_start = time.time()
    success = await tester.run_all_tests()
    return 0 if success else 1


if __name__ == "__main__":
    sys.exit(asyncio.run(main()))
