#!/usr/bin/env python
"""
Visionary RAG Pipeline - Quick Improvement Loop
Fast testing with Science Class 8 specific questions
"""

import os
import sys
import json
import time
import requests
import fitz
from pathlib import Path
from typing import List, Dict
from datetime import datetime

class QuickImprovementLoop:
    """Fast continuous improvement testing."""
    
    def __init__(self):
        self.pdf_path = Path("science class 8.pdf")
        self.ollama_url = "http://localhost:11434"
        
        # Science Class 8 test questions with expected topics
        self.test_questions = [
            {
                "query": "What is photosynthesis?",
                "expected_topics": ["photosynthesis", "chloroplast", "glucose", "oxygen"],
            },
            {
                "query": "What are the parts of a cell?",
                "expected_topics": ["nucleus", "cytoplasm", "cell membrane", "mitochondria"],
            },
            {
                "query": "What is force?",
                "expected_topics": ["push", "pull", "object", "newton"],
            },
            {
                "query": "What is combustion?",
                "expected_topics": ["combustion", "oxygen", "heat", "light", "fire"],
            },
            {
                "query": "What are microorganisms?",
                "expected_topics": ["microorganism", "microscopic", "bacteria", "virus"],
            },
        ]
        
        # Configurations to test
        self.configs = [
            {"name": "llama3.2-fast", "llm": "llama3.2:3b", "top_k": 3},
            {"name": "llama3.2-accurate", "llm": "llama3.2:latest", "top_k": 5},
        ]
        
        self.chunks = []
        self.embeddings = []
        
    def check_ollama(self):
        """Check Ollama status."""
        try:
            response = requests.get(f"{self.ollama_url}/api/tags", timeout=5)
            if response.status_code == 200:
                models = response.json().get("models", [])
                print(f"✅ Ollama running - {len(models)} models available")
                return True
        except:
            pass
        print("❌ Ollama not running")
        return False
    
    def get_embedding(self, text: str) -> List[float]:
        """Get embedding."""
        response = requests.post(
            f"{self.ollama_url}/api/embeddings",
            json={"model": "nomic-embed-text", "prompt": text},
            timeout=60
        )
        return response.json().get("embedding", [])
    
    def generate_response(self, query: str, context: str, model: str) -> str:
        """Generate response."""
        prompt = f"""You are a helpful science tutor for CBSE Class 8.
Use ONLY this context to answer. If not in context, say "Not found in context".

Context:
{context}

Question: {query}

Answer (be concise):"""

        response = requests.post(
            f"{self.ollama_url}/api/generate",
            json={"model": model, "prompt": prompt, "stream": False},
            timeout=120
        )
        return response.json().get("response", "")
    
    def extract_and_embed(self):
        """Extract and embed chunks from PDF."""
        print(f"\n{'='*60}")
        print("EXTRACTING & EMBEDDING")
        print(f"{'='*60}")
        
        doc = fitz.open(self.pdf_path)
        total_pages = len(doc)
        
        all_chunks = []
        # Sample every 10th page
        for page_num in range(0, total_pages, 10):
            if page_num >= total_pages:
                continue
            page = doc[page_num]
            text = page.get_text()
            if len(text.strip()) < 100:
                continue
            
            # Split into chunks
            for i in range(0, len(text), 500):
                chunk = text[i:i+500]
                if len(chunk.strip()) > 50:
                    all_chunks.append({
                        "content": chunk,
                        "page": page_num + 1
                    })
        
        doc.close()
        
        # Embed chunks
        print(f"Embedding {min(30, len(all_chunks))} chunks...")
        self.chunks = []
        self.embeddings = []
        
        for i, chunk in enumerate(all_chunks[:30]):
            try:
                emb = self.get_embedding(chunk["content"])
                self.chunks.append(chunk)
                self.embeddings.append(emb)
                if (i + 1) % 10 == 0:
                    print(f"  {i+1}/{min(30, len(all_chunks))} embedded")
            except Exception as e:
                print(f"  Error: {e}")
        
        print(f"✅ Embedded {len(self.chunks)} chunks")
        return len(self.chunks)
    
    def retrieve(self, query: str, top_k: int):
        """Retrieve top-k chunks."""
        query_emb = self.get_embedding(query)
        
        scores = []
        for j, chunk_emb in enumerate(self.embeddings):
            score = sum(a*b for a, b in zip(query_emb, chunk_emb))
            scores.append((j, score))
        
        scores.sort(key=lambda x: x[1], reverse=True)
        return scores[:top_k]
    
    def test_config(self, config: Dict) -> Dict:
        """Test one configuration."""
        print(f"\n{'='*60}")
        print(f"TESTING: {config['name']}")
        print(f"LLM: {config['llm']}, Top-K: {config['top_k']}")
        print(f"{'='*60}")
        
        # Embed chunks
        if not self.chunks:
            self.extract_and_embed()
        
        # Test each question
        results = []
        latencies = []
        topic_matches = []
        
        for i, q in enumerate(self.test_questions, 1):
            print(f"\n  [{i}/{len(self.test_questions)}] {q['query']}")
            
            start = time.time()
            
            try:
                # Retrieve
                top_chunks = self.retrieve(q['query'], config['top_k'])
                
                # Build context
                context_parts = []
                for idx, _ in top_chunks:
                    chunk = self.chunks[idx]
                    context_parts.append(f"[Page {chunk['page']}] {chunk['content']}")
                context = "\n\n".join(context_parts)
                
                # Generate
                answer = self.generate_response(q['query'], context, config['llm'])
                
                latency = (time.time() - start) * 1000
                latencies.append(latency)
                
                # Check topic match (simple keyword matching)
                answer_lower = answer.lower()
                matches = sum(1 for topic in q['expected_topics'] if topic.lower() in answer_lower)
                topic_score = matches / len(q['expected_topics'])
                topic_matches.append(topic_score)
                
                results.append({
                    "query": q['query'],
                    "answer": answer[:150],
                    "latency_ms": latency,
                    "topic_score": topic_score,
                    "status": "success"
                })
                
                print(f"     Answer: {answer[:100]}...")
                print(f"     Topic match: {topic_score:.2f}")
                print(f"     Latency: {latency:.0f}ms")
                
            except Exception as e:
                print(f"     ❌ Error: {e}")
                latencies.append(0)
                topic_matches.append(0)
                results.append({
                    "query": q['query'],
                    "answer": f"Error: {e}",
                    "latency_ms": 0,
                    "topic_score": 0,
                    "status": "failed"
                })
        
        # Calculate metrics
        avg_latency = sum(latencies) / len(latencies) if latencies else 0
        avg_topic_score = sum(topic_matches) / len(topic_matches) if topic_matches else 0
        success_rate = sum(1 for r in results if r['status'] == 'success') / len(results)
        
        # Overall score
        overall = (avg_topic_score * 0.6) + (success_rate * 0.4)
        
        metrics = {
            "config": config['name'],
            "llm": config['llm'],
            "top_k": config['top_k'],
            "avg_latency_ms": round(avg_latency, 2),
            "avg_topic_score": round(avg_topic_score, 3),
            "success_rate": round(success_rate, 3),
            "overall_score": round(overall, 3),
            "timestamp": datetime.now().isoformat()
        }
        
        print(f"\n✅ {config['name']} complete")
        print(f"   Overall Score: {overall:.3f}")
        print(f"   Topic Match: {avg_topic_score:.2f}")
        print(f"   Success Rate: {success_rate:.1%}")
        
        return metrics
    
    def run_loop(self):
        """Run improvement loop."""
        print("\n" + "="*60)
        print("QUICK IMPROVEMENT LOOP")
        print("="*60)
        
        if not self.check_ollama():
            return
        
        all_metrics = []
        
        for config in self.configs:
            metrics = self.test_config(config)
            all_metrics.append(metrics)
        
        # Save results
        self.save_results(all_metrics)
        
        # Print summary
        self.print_summary(all_metrics)
    
    def save_results(self, metrics_list: List[Dict]):
        """Save results."""
        data_dir = Path("data")
        data_dir.mkdir(exist_ok=True)
        
        # JSON
        json_path = data_dir / "quick_improvement_results.json"
        with open(json_path, "w") as f:
            json.dump(metrics_list, f, indent=2)
        print(f"\n✅ Saved: {json_path}")
        
        # CSV
        import pandas as pd
        df = pd.DataFrame(metrics_list)
        csv_path = data_dir / "quick_improvement_results.csv"
        df.to_csv(csv_path, index=False)
        print(f"✅ Saved: {csv_path}")
    
    def print_summary(self, metrics_list: List[Dict]):
        """Print summary."""
        print("\n" + "="*60)
        print("SUMMARY")
        print("="*60)
        
        print(f"\n{'Config':<25} {'Score':<10} {'Topic':<10} {'Latency':<10}")
        print("-"*60)
        
        for m in sorted(metrics_list, key=lambda x: x['overall_score'], reverse=True):
            print(f"{m['config']:<25} {m['overall_score']:<10.3f} {m['avg_topic_score']:<10.3f} {m['avg_latency_ms']:<10.0f}ms")
        
        print("-"*60)
        
        # Best config
        best = max(metrics_list, key=lambda x: x['overall_score'])
        print(f"\n🏆 Best: {best['config']} (Score: {best['overall_score']:.3f})")
        print("\n✅ Loop complete!")


def main():
    loop = QuickImprovementLoop()
    loop.run_loop()
    return 0


if __name__ == "__main__":
    sys.exit(main())
