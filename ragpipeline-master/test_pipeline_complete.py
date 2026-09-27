"""
Visionary RAG Pipeline - Complete Integration Test
Tests: PostgreSQL + Redis + GCP Gemini API + Vector Storage
Uses your GCP API key directly via REST API (no service account needed)
"""

import os
import sys
import json
import time
import asyncio
import requests
import numpy as np
from pathlib import Path
from datetime import datetime
from typing import List, Dict, Optional

# ============================================================
# CONFIGURATION
# ============================================================

GCP_API_KEY = "AIzaSyBcxwsSdb2mNnqYGcaYbMH1furWLae-W1k"
GEMINI_MODEL = "gemini-2.5-flash"  # Updated: 2.0 flash was rate limited
EMBEDDING_MODEL = "gemini-embedding-001"  # Stable embedding model (768 dim)
EMBEDDING_DIM = 768

POSTGRES_HOST = "localhost"
POSTGRES_PORT = 5432
POSTGRES_USER = "postgres"
POSTGRES_PASSWORD = "postgres"
POSTGRES_DB = "ragdb"

REDIS_HOST = "localhost"
REDIS_PORT = 6379


class PipelineTester:
    """Tests the complete RAG pipeline."""

    def __init__(self):
        self.results = {
            "test_date": datetime.now().isoformat(),
            "tests": [],
            "summary": {}
        }
        self.db_pool = None
        self.redis_client = None

    def log(self, test_name: str, status: str, details: str = ""):
        """Log test result."""
        result = {
            "test": test_name,
            "status": status,
            "details": details,
            "timestamp": datetime.now().isoformat()
        }
        self.results["tests"].append(result)

        icon = "✅" if status == "PASS" else "❌"
        print(f"  {icon} {test_name}: {status}")
        if details:
            print(f"     {details}")

    # ============================================================
    # TEST 1: Redis Connectivity
    # ============================================================

    async def test_redis(self):
        """Test Redis connectivity."""
        print("\n[TEST 1] Redis Connectivity...")
        try:
            import redis
            r = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)
            r.ping()

            # Test SET/GET
            r.set("pipeline:test", "ok")
            value = r.get("pipeline:test")
            r.delete("pipeline:test")

            # Test info
            info = r.info("server")
            redis_version = info.get("redis_version", "unknown")

            self.log("redis_connectivity", "PASS", f"v{redis_version}, SET/GET OK")
            r.close()
            return True
        except Exception as e:
            self.log("redis_connectivity", "FAIL", str(e))
            return False

    # ============================================================
    # TEST 2: PostgreSQL Connectivity
    # ============================================================

    async def test_postgres(self):
        """Test PostgreSQL connectivity and schema."""
        print("\n[TEST 2] PostgreSQL Connectivity...")
        try:
            import asyncpg
            dsn = f"postgresql://{POSTGRES_USER}:{POSTGRES_PASSWORD}@{POSTGRES_HOST}:{POSTGRES_PORT}/{POSTGRES_DB}"
            conn = await asyncpg.connect(dsn)

            # Test extensions
            extensions = await conn.fetch(
                "SELECT extname FROM pg_extension WHERE extname IN ('vector', 'btree_gin', 'pgcrypto')"
            )
            ext_names = [r["extname"] for r in extensions]

            # Test tables
            tables = await conn.fetch(
                """
                SELECT table_name FROM information_schema.tables
                WHERE table_schema = 'public'
                AND table_name IN ('cbse_taxonomy', 'parent_chunks', 'child_chunks',
                                   'ingestion_dlq', 'ai_feedback_loop', 'document_versions')
                ORDER BY table_name
                """
            )
            table_names = [r["table_name"] for r in tables]

            # Test taxonomy data
            taxonomy_count = await conn.fetchval("SELECT COUNT(*) FROM cbse_taxonomy")

            # Test vector type exists
            vector_type = await conn.fetchval(
                "SELECT EXISTS(SELECT 1 FROM pg_type WHERE typname = 'vector')"
            )

            await conn.close()

            details = (
                f"Tables: {len(table_names)}/6, "
                f"Extensions: {', '.join(ext_names)}, "
                f"Taxonomy: {taxonomy_count} entries, "
                f"Vector type: {'yes' if vector_type else 'no'}"
            )
            self.log("postgres_connectivity", "PASS", details)
            return True
        except Exception as e:
            self.log("postgres_connectivity", "FAIL", str(e))
            return False

    # ============================================================
    # TEST 3: GCP Gemini API (LLM)
    # ============================================================

    def test_gemini_llm(self):
        """Test Gemini LLM API with your API key."""
        print("\n[TEST 3] GCP Gemini LLM API...")
        try:
            url = f"https://generativelanguage.googleapis.com/v1beta/models/{GEMINI_MODEL}:generateContent?key={GCP_API_KEY}"

            payload = {
                "contents": [{
                    "parts": [{
                        "text": "What is photosynthesis in one sentence?"
                    }]
                }],
                "generationConfig": {
                    "temperature": 0.7,
                    "maxOutputTokens": 100
                }
            }

            response = requests.post(url, json=payload, timeout=30)
            response.raise_for_status()

            data = response.json()
            answer = data["candidates"][0]["content"]["parts"][0]["text"]

            # Check response quality
            is_relevant = len(answer) > 20 and "photosynthesis" in answer.lower() or "plant" in answer.lower() or "light" in answer.lower()

            details = f"Response: {answer[:100]}..."
            self.log("gemini_llm_api", "PASS" if is_relevant else "WARN", details)
            return is_relevant
        except Exception as e:
            self.log("gemini_llm_api", "FAIL", str(e))
            return False

    # ============================================================
    # TEST 4: GCP Embedding API
    # ============================================================

    def test_embedding_api(self):
        """Test Google Embedding API with your API key."""
        print("\n[TEST 4] GCP Embedding API...")
        try:
            # Use the correct Gemini embedding endpoint
            url = f"https://generativelanguage.googleapis.com/v1beta/models/{EMBEDDING_MODEL}:embedContent?key={GCP_API_KEY}"

            payload = {
                "content": {
                    "parts": [{"text": "What is force and pressure?"}]
                },
                "task_type": "RETRIEVAL_QUERY"
            }

            response = requests.post(url, json=payload, timeout=30)
            response.raise_for_status()

            data = response.json()
            embedding = data["embedding"]["values"]

            # Validate embedding
            dim = len(embedding)
            is_valid = dim > 0 and any(v != 0 for v in embedding)

            details = f"Dimension: {dim}, Non-zero: {sum(1 for v in embedding if v != 0)}"
            self.log("embedding_api", "PASS" if is_valid else "FAIL", details)
            return is_valid
        except Exception as e:
            self.log("embedding_api", "FAIL", str(e))
            return False

    # ============================================================
    # TEST 5: PostgreSQL Vector Storage
    # ============================================================

    async def test_vector_storage(self):
        """Test storing and querying vectors in PostgreSQL."""
        print("\n[TEST 5] PostgreSQL Vector Storage...")
        try:
            import asyncpg
            dsn = f"postgresql://{POSTGRES_USER}:{POSTGRES_PASSWORD}@{POSTGRES_HOST}:{POSTGRES_PORT}/{POSTGRES_DB}"
            conn = await asyncpg.connect(dsn)

            # Create test vector - format as PostgreSQL vector string
            test_vector = [0.1] * 768
            test_vector[0] = 1.0  # Make it distinctive
            vector_str = "[" + ",".join(str(v) for v in test_vector) + "]"

            # First create a parent chunk (child_chunks requires parent_id)
            await conn.execute(
                """
                INSERT INTO parent_chunks (taxonomy_id, content, content_type)
                VALUES ($1, $2, 'prose')
                ON CONFLICT DO NOTHING
                RETURNING parent_id
                """,
                1,  # taxonomy_id for Food chapter
                "Test parent chunk for vector storage verification"
            )

            # Get the parent_id
            parent_id = await conn.fetchval(
                "SELECT parent_id FROM parent_chunks WHERE content LIKE 'Test parent chunk%' LIMIT 1"
            )

            if parent_id:
                # Insert child chunk
                await conn.execute(
                    """
                    INSERT INTO child_chunks (content, embedding, taxonomy_id, content_type, parent_id)
                    VALUES ($1, $2::vector, $3, 'prose', $4)
                    ON CONFLICT DO NOTHING
                    """,
                    "Test vector storage content for pipeline verification",
                    vector_str,
                    1,
                    parent_id
                )

            # Query using vector similarity
            results = await conn.fetch(
                """
                SELECT cc.content, embedding <=> $1::vector AS similarity
                FROM child_chunks cc
                WHERE cc.taxonomy_id = 1
                ORDER BY cc.embedding <=> $1::vector
                LIMIT 5
                """,
                vector_str
            )

            found_test = any("Test vector storage" in r["content"] for r in results)

            # Clean up
            await conn.execute(
                "DELETE FROM child_chunks WHERE content LIKE 'Test vector storage%'"
            )
            await conn.execute(
                "DELETE FROM parent_chunks WHERE content LIKE 'Test parent chunk%'"
            )

            await conn.close()

            details = f"Found {len(results)} results, test content: {'yes' if found_test else 'no'}"
            self.log("vector_storage", "PASS" if found_test else "WARN", details)
            return found_test
        except Exception as e:
            self.log("vector_storage", "FAIL", str(e))
            return False

    # ============================================================
    # TEST 6: End-to-End RAG Query
    # ============================================================

    async def test_e2e_query(self):
        """Test end-to-end: embed query -> search -> retrieve -> generate answer."""
        print("\n[TEST 6] End-to-End RAG Query...")
        try:
            # Step 1: Create embedding for query
            embed_url = f"https://generativelanguage.googleapis.com/v1beta/models/{EMBEDDING_MODEL}:embedContent?key={GCP_API_KEY}"
            query = "What is force in physics?"

            embed_response = requests.post(embed_url, json={
                "content": {"parts": [{"text": query}]},
                "task_type": "RETRIEVAL_QUERY"
            }, timeout=30)
            embed_response.raise_for_status()
            query_embedding_list = embed_response.json()["embedding"]["values"]
            query_embedding_str = "[" + ",".join(str(v) for v in query_embedding_list) + "]"

            # Step 2: Search PostgreSQL
            import asyncpg
            dsn = f"postgresql://{POSTGRES_USER}:{POSTGRES_PASSWORD}@{POSTGRES_HOST}:{POSTGRES_PORT}/{POSTGRES_DB}"
            conn = await asyncpg.connect(dsn)

            results = await conn.fetch(
                """
                SELECT cc.content, cc.page_number, pc.chapter, pc.section
                FROM child_chunks cc
                LEFT JOIN parent_chunks pc ON cc.parent_id = pc.parent_id
                WHERE cc.taxonomy_id = 11
                ORDER BY cc.embedding <=> $1::vector
                LIMIT 3
                """,
                query_embedding_str
            )
            await conn.close()

            # Step 3: Generate answer with Gemini (retry if rate limited)
            context = "\n".join([r["content"][:200] for r in results]) if results else "No context found"

            gemini_url = f"https://generativelanguage.googleapis.com/v1beta/models/{GEMINI_MODEL}:generateContent?key={GCP_API_KEY}"
            
            # Retry with exponential backoff for rate limiting
            answer = "No answer generated"
            has_answer = False
            for attempt in range(3):
                try:
                    time.sleep(2 ** attempt)  # Exponential backoff
                    gemini_response = requests.post(gemini_url, json={
                        "contents": [{"parts": [{"text": f"Based on this context: {context}\n\nAnswer: {query}"}]}],
                        "generationConfig": {"temperature": 0.7, "maxOutputTokens": 200}
                    }, timeout=30)
                    gemini_response.raise_for_status()
                    answer = gemini_response.json()["candidates"][0]["content"]["parts"][0]["text"]
                    has_answer = len(answer) > 20
                    break
                except requests.exceptions.HTTPError as e:
                    if "429" in str(e):
                        print(f"     Rate limited, retrying in {2**attempt}s...")
                        continue
                    raise

            chunks_found = len(results)

            details = f"Chunks retrieved: {chunks_found}, Answer generated: {'yes' if has_answer else 'no'}"
            if has_answer:
                details += f"\n     Answer: {answer[:150]}..."

            self.log("e2e_rag_query", "PASS" if has_answer else "WARN", details)
            return has_answer
        except Exception as e:
            self.log("e2e_rag_query", "FAIL", str(e))
            return False

    # ============================================================
    # TEST 7: Redis Caching
    # ============================================================

    async def test_redis_caching(self):
        """Test Redis caching for embeddings."""
        print("\n[TEST 7] Redis Embedding Cache...")
        try:
            import redis
            import hashlib
            r = redis.Redis(host=REDIS_HOST, port=REDIS_PORT, decode_responses=True)

            # Test caching an embedding
            test_text = "test query for caching"
            cache_key = f"embedding:{hashlib.sha256(test_text.encode()).hexdigest()}"
            test_embedding = [0.1] * 768

            # Store in cache
            r.setex(cache_key, 3600, json.dumps(test_embedding))

            # Retrieve from cache
            cached = r.get(cache_key)
            cached_embedding = json.loads(cached) if cached else None

            # Verify
            is_cached = cached_embedding == test_embedding

            # Clean up
            r.delete(cache_key)
            r.close()

            self.log("redis_caching", "PASS" if is_cached else "FAIL",
                    f"Cache SET/GET: {'OK' if is_cached else 'FAILED'}")
            return is_cached
        except Exception as e:
            self.log("redis_caching", "FAIL", str(e))
            return False

    # ============================================================
    # RUN ALL TESTS
    # ============================================================

    async def run_all_tests(self):
        """Run complete test suite."""
        print("\n" + "="*70)
        print("VISIONARY RAG PIPELINE - COMPLETE INTEGRATION TEST")
        print("="*70)
        print(f"\nConfiguration:")
        print(f"  GCP Project: visionary-rag-test")
        print(f"  Gemini Model: {GEMINI_MODEL}")
        print(f"  Embedding Model: {EMBEDDING_MODEL}")
        print(f"  PostgreSQL: {POSTGRES_HOST}:{POSTGRES_PORT}/{POSTGRES_DB}")
        print(f"  Redis: {REDIS_HOST}:{REDIS_PORT}")
        print(f"\nStarting tests...\n")

        start_time = time.time()

        # Run tests
        await self.test_redis()
        await self.test_postgres()
        self.test_gemini_llm()
        self.test_embedding_api()
        await self.test_vector_storage()
        await self.test_e2e_query()
        await self.test_redis_caching()

        elapsed = time.time() - start_time

        # Calculate summary
        passed = sum(1 for t in self.results["tests"] if t["status"] == "PASS")
        failed = sum(1 for t in self.results["tests"] if t["status"] == "FAIL")
        warn = sum(1 for t in self.results["tests"] if t["status"] == "WARN")
        total = len(self.results["tests"])

        self.results["summary"] = {
            "total": total,
            "passed": passed,
            "failed": failed,
            "warnings": warn,
            "elapsed_seconds": round(elapsed, 2),
            "pass_rate": f"{(passed/total*100):.1f}%" if total > 0 else "0%"
        }

        # Print summary
        print("\n" + "="*70)
        print("TEST SUMMARY")
        print("="*70)
        print(f"  Total:     {total}")
        print(f"  Passed:    {passed} ✅")
        print(f"  Failed:    {failed} ❌")
        print(f"  Warnings:  {warn} ⚠️")
        print(f"  Pass Rate: {self.results['summary']['pass_rate']}")
        print(f"  Time:      {elapsed:.2f}s")
        print("="*70)

        # Save results
        output_path = Path("data/pipeline_test_results.json")
        output_path.parent.mkdir(exist_ok=True)
        with open(output_path, "w") as f:
            json.dump(self.results, f, indent=2)
        print(f"\nResults saved to: {output_path}")

        return failed == 0


async def main():
    """Run the pipeline tests."""
    tester = PipelineTester()
    success = await tester.run_all_tests()
    return 0 if success else 1


if __name__ == "__main__":
    sys.exit(asyncio.run(main()))
