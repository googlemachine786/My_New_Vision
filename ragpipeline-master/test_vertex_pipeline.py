"""Test the full RAG pipeline with Vertex AI - no mocks, no Ollama."""
import vertexai
from vertexai.language_models import TextEmbeddingModel, TextGenerationModel
from vertexai.generative_models import GenerativeModel
import sys

PROJECT = "visionary-research"
LOCATION = "us-central1"

def test_embedding():
    """Test text-embedding-005 model."""
    print("=" * 60)
    print("TEST 1: Vertex AI Embeddings (text-embedding-005)")
    print("=" * 60)
    print("  ⏭️  SKIP — Need roles/aiplatform.user on visionary-research")
    print("  Ask org admin to run:")
    print("    gcloud projects add-iam-policy-binding visionary-research \\")
    print("      --member='user:ruthvik@visionary.org.in' \\")
    print("      --role='roles/aiplatform.user'")
    return None

def test_gemini_llm():
    """Test Gemini 2.0 Flash for text generation."""
    print("\n" + "=" * 60)
    print("TEST 2: Vertex AI Gemini 2.0 Flash (LLM Generation)")
    print("=" * 60)
    try:
        model = GenerativeModel("gemini-2.0-flash")
        response = model.generate_content(
            "In 2 sentences, what is the nature of viruses?",
            generation_config={"temperature": 0.3, "max_output_tokens": 100}
        )
        print(f"  ✅ Response: {response.text.strip()[:200]}")
        return True
    except Exception as e:
        print(f"  ❌ FAILED: {e}")
        return False

def test_gemini_embed():
    """Test Gemini embeddings as alternative."""
    print("\n" + "=" * 60)
    print("TEST 3: Vertex AI Gemini Embeddings (gemini-embedding-001)")
    print("=" * 60)
    try:
        model = TextEmbeddingModel.from_pretrained("gemini-embedding-001")
        emb = model.get_embeddings(["Test query"])[0]
        print(f"  ✅ Dimension: {len(emb.values)}")
        return True
    except Exception as e:
        print(f"  ❌ FAILED: {e}")
        return False

def test_gcp_database():
    """Test connection to GCP PostgreSQL on the VM."""
    print("\n" + "=" * 60)
    print("TEST 4: GCP PostgreSQL + pgvector (VM: 35.226.237.147)")
    print("=" * 60)
    try:
        import psycopg2
        conn = psycopg2.connect(
            host="35.226.237.147",
            port=5432,
            user="postgres",
            password="postgres",
            database="ragdb",
            connect_timeout=10
        )
        cur = conn.cursor()
        cur.execute("SELECT version();")
        version = cur.fetchone()[0]
        print(f"  ✅ PostgreSQL: {version[:60]}...")
        
        # Check pgvector extension
        cur.execute("SELECT extname FROM pg_extension WHERE extname = 'vector';")
        ext = cur.fetchone()
        if ext:
            print(f"  ✅ pgvector extension: INSTALLED")
        else:
            print(f"  ⚠️  pgvector extension: NOT INSTALLED (run: CREATE EXTENSION vector;)")
        
        # Create test table
        cur.execute("""
            CREATE TABLE IF NOT EXISTS test_chunks (
                id SERIAL PRIMARY KEY,
                text TEXT,
                embedding vector(768)
            );
        """)
        conn.commit()
        print(f"  ✅ Test table created: test_chunks")
        
        cur.close()
        conn.close()
        return True
    except ImportError:
        print(f"  ⚠️  psycopg2 not installed. Run: pip install psycopg2-binary")
        return None
    except Exception as e:
        print(f"  ❌ FAILED: {e}")
        return False

def test_gcp_redis():
    """Test connection to GCP Redis on the VM."""
    print("\n" + "=" * 60)
    print("TEST 5: GCP Redis (VM: 35.226.237.147:6379)")
    print("=" * 60)
    try:
        import redis
        r = redis.Redis(host="35.226.237.147", port=6379, socket_timeout=10)
        r.ping()
        r.set("test_key", "pipeline_test_ok")
        val = r.get("test_key")
        print(f"  ✅ Redis PING: OK, test_key={val.decode()}")
        r.delete("test_key")
        return True
    except ImportError:
        print(f"  ⚠️  redis not installed. Run: pip install redis")
        return None
    except Exception as e:
        print(f"  ❌ FAILED: {e}")
        return False

def main():
    print("\n" + "=" * 60)
    print("VISIONARY RAG - FULL PIPELINE TEST (GCP Vertex AI)")
    print(f"Project: {PROJECT}, Location: {LOCATION}")
    print("=" * 60)
    
    # Initialize Vertex AI
    print("\nInitializing Vertex AI...")
    try:
        vertexai.init(project=PROJECT, location=LOCATION)
        print("✅ Vertex AI initialized")
    except Exception as e:
        print(f"❌ Vertex AI init failed: {e}")
        print("\nFix: Ensure GOOGLE_APPLICATION_CREDENTIALS points to a valid service account JSON")
        print("Or run: gcloud auth application-default login")
        sys.exit(1)
    
    results = {}
    results["Embeddings"] = test_embedding()
    results["Gemini LLM"] = test_gemini_llm()
    results["Gemini Embed"] = test_gemini_embed()
    results["PostgreSQL"] = test_gcp_database()
    results["Redis"] = test_gcp_redis()
    
    # Summary
    print("\n" + "=" * 60)
    print("PIPELINE TEST SUMMARY")
    print("=" * 60)
    passed = 0
    failed = 0
    skipped = 0
    for name, result in results.items():
        if result is True:
            print(f"  ✅ {name}: PASS")
            passed += 1
        elif result is False:
            print(f"  ❌ {name}: FAIL")
            failed += 1
        else:
            print(f"  ⚠️  {name}: SKIP (needs aiplatform.user role)")
            skipped += 1
    
    total = passed + failed + skipped
    print(f"\n  Passed: {passed}/{total}, Failed: {failed}/{total}, Skipped: {skipped}/{total}")
    
    if failed == 0:
        print("\n🎉 All tested services passed!")
    else:
        print(f"\n⚠️  {failed} service(s) failed. Fix and retry.")
    
    return 0 if failed == 0 else 1

if __name__ == "__main__":
    sys.exit(main())
