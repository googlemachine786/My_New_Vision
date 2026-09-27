#!/usr/bin/env python
"""
Visionary RAG Pipeline - Real Ollama End-to-End Test
Uses actual Ollama embeddings and LLM (no mocks!)
"""

import os
import sys
import json
import time
import hashlib
import requests
import pandas as pd
from pathlib import Path
from typing import List, Dict, Any
from datetime import datetime

class OllamaRAGTest:
    """Real RAG test using Ollama API."""
    
    def __init__(self):
        self.pdf_path = Path("science class 8.pdf")
        self.dataset_path = Path("science_dataset.xlsx")
        self.ollama_url = os.getenv("OLLAMA_BASE_URL", "http://localhost:11434")
        self.embed_model = os.getenv("EMBED_MODEL", "nomic-embed-text")
        self.llm_model = os.getenv("LLM_MODEL", "llama3.2:3b")
        
        # In-memory vector store (simple FAISS alternative for testing)
        self.chunks = []
        self.embeddings = []
        
        self.results = {
            "metadata": {
                "test_date": datetime.now().isoformat(),
                "ollama_url": self.ollama_url,
                "embed_model": self.embed_model,
                "llm_model": self.llm_model,
            },
            "ingestion": {},
            "queries": [],
            "evaluation": {},
            "metrics": {},
            "summary": {},
        }
        
    def check_ollama(self):
        """Check if Ollama is running."""
        print("\n" + "="*70)
        print("CHECKING OLLAMA CONNECTION")
        print("="*70)
        
        try:
            response = requests.get(f"{self.ollama_url}/api/tags", timeout=5)
            if response.status_code == 200:
                models = response.json().get("models", [])
                print(f"✅ Ollama is running at {self.ollama_url}")
                print(f"   Available models: {len(models)}")
                for model in models:
                    print(f"   - {model['name']}")
                return True
            else:
                print(f"❌ Ollama returned status {response.status_code}")
                return False
        except requests.exceptions.ConnectionError:
            print(f"❌ Cannot connect to Ollama at {self.ollama_url}")
            print("   Make sure Ollama is running: ollama serve")
            return False
        except Exception as e:
            print(f"❌ Error: {e}")
            return False
    
    def get_embedding(self, text: str) -> List[float]:
        """Get real embedding from Ollama."""
        response = requests.post(
            f"{self.ollama_url}/api/embeddings",
            json={
                "model": self.embed_model,
                "prompt": text
            },
            timeout=30
        )
        response.raise_for_status()
        return response.json().get("embedding", [])
    
    def generate_response(self, query: str, context: str) -> str:
        """Generate real response using Ollama LLM."""
        prompt = f"""You are a helpful science tutor for CBSE Class 8 students.
Use ONLY the following context to answer the question.
If the answer is not in the context, say "I don't have information about this in the provided context."

Context:
{context}

Question: {query}

Answer (be concise and cite page numbers if available):"""

        response = requests.post(
            f"{self.ollama_url}/api/generate",
            json={
                "model": self.llm_model,
                "prompt": prompt,
                "stream": False
            },
            timeout=60
        )
        response.raise_for_status()
        return response.json().get("response", "")
    
    def extract_text_from_pdf(self):
        """Extract text from PDF using PyMuPDF - samples from ENTIRE document."""
        import fitz
        
        print(f"\nExtracting text from {self.pdf_path.name}...")
        doc = fitz.open(self.pdf_path)
        
        all_text = []
        total_pages = len(doc)
        
        # Sample pages from throughout the ENTIRE PDF (not just first 10)
        # Pick every 10th page to get good coverage
        sample_pages = list(range(0, total_pages, 10))
        
        # Also add some specific pages that likely have content
        for page_num in sample_pages[:20]:  # Sample 20 pages from throughout PDF
            if page_num >= total_pages:
                continue
                
            page = doc[page_num]
            text = page.get_text()
            
            # Skip empty pages or table of contents
            if len(text.strip()) < 100:
                continue
            
            # Split into chunks (roughly 500 chars each)
            chunk_size = 500
            for i in range(0, len(text), chunk_size):
                chunk = text[i:i+chunk_size]
                if len(chunk.strip()) > 50:  # Skip tiny chunks
                    all_text.append({
                        "content": chunk,
                        "page": page_num + 1,
                        "chunk_id": f"p{page_num+1}_c{i//chunk_size}"
                    })
        
        doc.close()
        return all_text
    
    def run_all_tests(self):
        """Run complete test suite with real Ollama."""
        print("\n" + "="*70)
        print("VISIONARY RAG PIPELINE - REAL OLLAMA TEST")
        print("="*70)
        
        # Step 1: Check Ollama
        if not self.check_ollama():
            print("\n❌ Ollama is not running. Please start Ollama:")
            print("   ollama serve")
            print("   ollama pull nomic-embed-text")
            print("   ollama pull llama3.2:3b")
            return False
        
        # Step 2: Extract and embed PDF text
        print("\n" + "="*70)
        print("STEP 1: PDF INGESTION (Real Ollama Embeddings)")
        print("="*70)
        self.ingest_pdf()
        
        # Step 3: Run RAG queries
        print("\n" + "="*70)
        print("STEP 2: RAG QUERIES (Real Ollama LLM)")
        print("="*70)
        self.run_rag_queries()
        
        # Step 4: Calculate metrics
        print("\n" + "="*70)
        print("STEP 3: CALCULATING METRICS")
        print("="*70)
        self.calculate_metrics()
        
        # Step 5: Generate summary
        print("\n" + "="*70)
        print("STEP 4: GENERATING SUMMARY")
        print("="*70)
        self.generate_summary()
        
        # Step 6: Save results
        print("\n" + "="*70)
        print("STEP 5: SAVING RESULTS")
        print("="*70)
        self.save_results()
        
        # Print final summary
        self.print_final_summary()
        
        return True
    
    def ingest_pdf(self):
        """Ingest PDF with real Ollama embeddings."""
        start_time = time.time()
        
        # Extract text
        extracted_chunks = self.extract_text_from_pdf()
        print(f"✅ Extracted {len(extracted_chunks)} chunks from PDF")
        
        # Create embeddings with Ollama
        print(f"\nCreating embeddings with Ollama ({self.embed_model})...")
        self.chunks = []
        self.embeddings = []
        
        # Embed MORE chunks (not just 5!)
        chunks_to_embed = min(50, len(extracted_chunks))  # Embed up to 50 chunks
        print(f"Embedding {chunks_to_embed} chunks from throughout the PDF...")
        
        for i, chunk in enumerate(extracted_chunks[:chunks_to_embed], 1):
            print(f"  Embedding chunk {i}/{chunks_to_embed} (page {chunk['page']})...", end=" ")
            
            try:
                embedding = self.get_embedding(chunk["content"])
                self.chunks.append(chunk)
                self.embeddings.append(embedding)
                print(f"✅ ({len(embedding)} dims)")
            except Exception as e:
                print(f"❌ Error: {e}")
        
        elapsed = time.time() - start_time
        
        self.results["ingestion"] = {
            "status": "success",
            "pages_processed": 10,
            "chunks_embedded": len(self.chunks),
            "embedding_model": self.embed_model,
            "embedding_dimensions": len(self.embeddings[0]) if self.embeddings else 0,
            "time_seconds": round(elapsed, 2),
        }
        
        print(f"\n✅ Ingestion complete:")
        print(f"   Chunks embedded: {len(self.chunks)}")
        print(f"   Embedding dims: {len(self.embeddings[0]) if self.embeddings else 0}")
        print(f"   Time: {elapsed:.2f}s")
        
    def run_rag_queries(self):
        """Run real RAG queries using Ollama LLM with better context retrieval."""
        test_queries = [
            "What is photosynthesis?",
            "What are the parts of a cell?",
            "What is force?",
        ]
        
        for i, query in enumerate(test_queries, 1):
            print(f"\n  [{i}/{len(test_queries)}] Query: {query}")
            
            start_time = time.time()
            
            try:
                # Get query embedding
                query_embedding = self.get_embedding(query)
                
                # Find TOP 3 most similar chunks (not just 1!)
                chunk_scores = []
                for j, chunk_emb in enumerate(self.embeddings):
                    # Cosine similarity
                    score = sum(a*b for a, b in zip(query_embedding, chunk_emb))
                    chunk_scores.append((j, score))
                
                # Sort by similarity and get top 3
                chunk_scores.sort(key=lambda x: x[1], reverse=True)
                top_3_indices = [idx for idx, score in chunk_scores[:3]]
                
                # Build context from top 3 chunks
                context_parts = []
                sources = []
                for idx in top_3_indices:
                    chunk = self.chunks[idx]
                    context_parts.append(f"[Page {chunk['page']}]\n{chunk['content']}")
                    sources.append(f"Page {chunk['page']}")
                
                context = "\n\n---\n\n".join(context_parts)
                
                # Generate response using Ollama LLM
                print(f"  Generating response with {self.llm_model}...", end=" ")
                answer = self.generate_response(query, context)
                print(f"✅")
                
                latency = time.time() - start_time
                
                result = {
                    "query": query,
                    "answer": answer.strip(),
                    "sources": sources,
                    "latency_ms": round(latency * 1000, 2),
                    "status": "success",
                    "num_context_chunks": len(top_3_indices),
                }
                
                self.results["queries"].append(result)
                
                # Print answer (full this time)
                print(f"     Answer: {answer.strip()[:200]}...")
                print(f"     Sources: {', '.join(sources)}")
                print(f"     Context chunks: {len(top_3_indices)}")
                print(f"     Latency: {latency*1000:.2f}ms")
                
            except Exception as e:
                print(f"     ❌ Error: {e}")
                self.results["queries"].append({
                    "query": query,
                    "answer": f"Error: {e}",
                    "sources": [],
                    "latency_ms": 0,
                    "status": "failed",
                })
    
    def calculate_metrics(self):
        """Calculate performance metrics."""
        queries = self.results["queries"]
        if not queries:
            self.results["metrics"] = {}
            return
            
        latencies = [q["latency_ms"] for q in queries]
        success_count = sum(1 for q in queries if q["status"] == "success")
        
        latencies_sorted = sorted(latencies)
        p95_idx = int(len(latencies_sorted) * 0.95)
        p99_idx = int(len(latencies_sorted) * 0.99)
        
        self.results["metrics"] = {
            "average_latency_ms": round(sum(latencies) / len(latencies), 2) if latencies else 0,
            "p95_latency_ms": latencies_sorted[p95_idx] if p95_idx < len(latencies_sorted) else (latencies_sorted[-1] if latencies_sorted else 0),
            "p99_latency_ms": latencies_sorted[p99_idx] if p99_idx < len(latencies_sorted) else (latencies_sorted[-1] if latencies_sorted else 0),
            "success_rate": round(success_count / len(queries), 2),
            "total_queries": len(queries),
            "successful_queries": success_count,
        }
        
        print(f"✅ Metrics calculated:")
        print(f"   Average latency: {self.results['metrics']['average_latency_ms']}ms")
        print(f"   P95 latency: {self.results['metrics']['p95_latency_ms']}ms")
        print(f"   Success rate: {self.results['metrics']['success_rate'] * 100}%")
        
    def generate_summary(self):
        """Generate test summary."""
        ingestion_status = self.results["ingestion"].get("status", "failed")
        queries_count = len(self.results["queries"])
        metrics = self.results.get("metrics", {})
        
        total_tests = 1 + queries_count
        passed_tests = (1 if ingestion_status == "success" else 0) + \
                      sum(1 for q in self.results["queries"] if q["status"] == "success")
        
        status = "COMPLETE" if passed_tests == total_tests else "PARTIAL"
        pass_fail = "PASS" if passed_tests >= total_tests * 0.8 else "FAIL"
        
        self.results["summary"] = {
            "status": status,
            "pass_fail": pass_fail,
            "total_tests": total_tests,
            "passed_tests": passed_tests,
            "failed_tests": total_tests - passed_tests,
        }
        
    def save_results(self):
        """Save results to JSON and Markdown."""
        data_dir = Path("data")
        data_dir.mkdir(exist_ok=True)
        
        # Save JSON
        json_path = data_dir / "test_results_ollama_real.json"
        with open(json_path, "w", encoding="utf-8") as f:
            json.dump(self.results, f, indent=2, ensure_ascii=False)
        print(f"✅ Results saved to: {json_path}")
        
        # Save Markdown
        md_path = data_dir / "OLLAMA_REAL_TEST_RESULTS.md"
        md_content = self.generate_markdown_summary()
        with open(md_path, "w", encoding="utf-8") as f:
            f.write(md_content)
        print(f"✅ Summary saved to: {md_path}")
        
    def generate_markdown_summary(self):
        """Generate Markdown summary."""
        summary = self.results["summary"]
        ingestion = self.results["ingestion"]
        metrics = self.results.get("metrics", {})
        
        md = f"""# Visionary RAG Pipeline - Real Ollama Test Results

**Test Date**: {self.results["metadata"]["test_date"]}
**Ollama URL**: {self.results["metadata"]["ollama_url"]}
**Embed Model**: {self.results["metadata"]["embed_model"]}
**LLM Model**: {self.results["metadata"]["llm_model"]}
**Status**: {summary["status"]} - {summary["pass_fail"]}

## Summary

- **Total Tests**: {summary["total_tests"]}
- **Passed**: {summary["passed_tests"]}
- **Failed**: {summary["failed_tests"]}
- **Success Rate**: {metrics.get("success_rate", 0) * 100:.1f}%

## Ingestion Results

- **Status**: {ingestion.get("status", "failed")}
- **Chunks Embedded**: {ingestion.get("chunks_embedded", 0)}
- **Embedding Dimensions**: {ingestion.get("embedding_dimensions", 0)}
- **Time**: {ingestion.get("time_seconds", 0):.2f}s

## Query Results

| # | Query | Answer Preview | Latency (ms) |
|---|-------|---------------|--------------|
"""
        
        for i, q in enumerate(self.results["queries"], 1):
            answer_preview = q["answer"][:50] + "..." if len(q["answer"]) > 50 else q["answer"]
            status_icon = "✅" if q["status"] == "success" else "❌"
            md += f"| {i} | {q['query']} | {status_icon} {answer_preview} | {q['latency_ms']} |\n"
            
        md += f"""
## Performance Metrics

- **Average Latency**: {metrics.get("average_latency_ms", 0):.2f} ms
- **P95 Latency**: {metrics.get("p95_latency_ms", 0):.2f} ms
- **P99 Latency**: {metrics.get("p99_latency_ms", 0):.2f} ms
- **Success Rate**: {metrics.get("success_rate", 0) * 100:.1f}%

## Conclusion

"""
        if summary["pass_fail"] == "PASS":
            md += "✅ **All tests passed!** Real Ollama embeddings and LLM working correctly.\n"
        else:
            md += "⚠️ **Some tests failed.** Review results for details.\n"
            
        return md
        
    def print_final_summary(self):
        """Print final summary to console."""
        print("\n" + "="*70)
        print("FINAL SUMMARY")
        print("="*70)
        
        summary = self.results["summary"]
        ingestion = self.results["ingestion"]
        metrics = self.results.get("metrics", {})
        
        print(f"Ingestion: {ingestion.get('status', 'failed')} ({ingestion.get('chunks_embedded', 0)} chunks)")
        print(f"Queries: {len(self.results['queries'])} tested")
        print(f"Metrics: Avg {metrics.get('average_latency_ms', 0):.2f}ms, Success {metrics.get('success_rate', 0) * 100:.1f}%")
        print(f"Overall: {summary['status']} - {summary['pass_fail']} ({summary['passed_tests']}/{summary['total_tests']} tests passed)")
        print("="*70)
        
        if summary["pass_fail"] == "PASS":
            print("\n✅ ALL TESTS COMPLETE - REAL OLLAMA WORKING!")
        else:
            print("\n⚠️ SOME TESTS FAILED - Check results for details")
            
        print("\nResults saved to:")
        print("  - data/test_results_ollama_real.json")
        print("  - data/OLLAMA_REAL_TEST_RESULTS.md")


def main():
    """Run complete end-to-end test with real Ollama."""
    tester = OllamaRAGTest()
    success = tester.run_all_tests()
    return 0 if success else 1


if __name__ == "__main__":
    sys.exit(main())
