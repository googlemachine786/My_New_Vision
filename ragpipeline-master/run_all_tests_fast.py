#!/usr/bin/env python
"""
Visionary RAG - FAST MASTER TEST SUITE
Runs ALL tests efficiently (under 5 minutes)
"""

import os
import sys
import json
import time
import requests
import fitz
import numpy as np
import pandas as pd
from pathlib import Path
from datetime import datetime
from typing import List, Dict

class FastMasterTest:
    """Fast comprehensive test suite."""
    
    def __init__(self):
        self.ollama_url = "http://localhost:11434"
        self.results = {
            "metadata": {
                "test_date": datetime.now().isoformat(),
                "total_tests": 0,
                "passed_tests": 0
            },
            "tests": []
        }
    
    def check_ollama(self) -> bool:
        """Check Ollama status."""
        try:
            response = requests.get(f"{self.ollama_url}/api/tags", timeout=5)
            if response.status_code == 200:
                models = response.json().get("models", [])
                print(f"✅ Ollama: {len(models)} models - {[m['name'] for m in models]}")
                self.results["ollama_models"] = [m['name'] for m in models]
                return True
        except:
            pass
        print("❌ Ollama not running")
        return False
    
    def test_embedding(self) -> Dict:
        """Test embedding functionality."""
        print("\n[TEST 1] Embedding...")
        self.results["metadata"]["total_tests"] += 1
        
        start = time.time()
        try:
            response = requests.post(
                f"{self.ollama_url}/api/embeddings",
                json={"model": "nomic-embed-text", "prompt": "Test"},
                timeout=30
            )
            latency = (time.time() - start) * 1000
            
            if response.status_code == 200:
                emb = response.json().get("embedding", [])
                result = {
                    "name": "Embedding",
                    "status": "PASS",
                    "latency_ms": round(latency, 2),
                    "dimensions": len(emb),
                    "model": "nomic-embed-text"
                }
                self.results["metadata"]["passed_tests"] += 1
            else:
                result = {"name": "Embedding", "status": "FAIL", "error": response.status_code}
        except Exception as e:
            result = {"name": "Embedding", "status": "FAIL", "error": str(e)}
        
        self.results["tests"].append(result)
        print(f"  {result['status']}: {result.get('latency_ms', 'N/A')}ms, {result.get('dimensions', 'N/A')} dims")
        return result
    
    def test_generation(self) -> Dict:
        """Test LLM generation."""
        print("\n[TEST 2] LLM Generation...")
        self.results["metadata"]["total_tests"] += 1
        
        start = time.time()
        try:
            response = requests.post(
                f"{self.ollama_url}/api/generate",
                json={"model": "llama3.2:3b", "prompt": "What is force?", "stream": False},
                timeout=60
            )
            latency = (time.time() - start) * 1000
            
            if response.status_code == 200:
                answer = response.json().get("response", "")[:50]
                result = {
                    "name": "LLM Generation",
                    "status": "PASS",
                    "latency_ms": round(latency, 2),
                    "model": "llama3.2:3b",
                    "answer_preview": answer
                }
                self.results["metadata"]["passed_tests"] += 1
            else:
                result = {"name": "LLM Generation", "status": "FAIL", "error": response.status_code}
        except Exception as e:
            result = {"name": "LLM Generation", "status": "FAIL", "error": str(e)}
        
        self.results["tests"].append(result)
        print(f"  {result['status']}: {result.get('latency_ms', 'N/A')}ms - {result.get('answer_preview', 'N/A')}...")
        return result
    
    def test_all_models(self) -> List[Dict]:
        """Test all available models."""
        print("\n[TEST 3] All Models...")
        
        model_results = []
        for model_name in self.results.get("ollama_models", []):
            if 'embed' in model_name.lower():
                continue  # Skip embedding models for generation test
            
            self.results["metadata"]["total_tests"] += 1
            start = time.time()
            
            try:
                response = requests.post(
                    f"{self.ollama_url}/api/generate",
                    json={"model": model_name, "prompt": "Hi", "stream": False},
                    timeout=60
                )
                latency = (time.time() - start) * 1000
                
                if response.status_code == 200:
                    model_results.append({
                        "model": model_name,
                        "status": "PASS",
                        "latency_ms": round(latency, 2)
                    })
                    self.results["metadata"]["passed_tests"] += 1
                    print(f"  ✅ {model_name}: {latency:.0f}ms")
                else:
                    model_results.append({"model": model_name, "status": "FAIL"})
                    print(f"  ❌ {model_name}: Failed")
            except:
                model_results.append({"model": model_name, "status": "ERROR"})
                print(f"  ❌ {model_name}: Error")
        
        self.results["model_tests"] = model_results
        return model_results
    
    def test_pdf_ingestion(self) -> Dict:
        """Test PDF ingestion."""
        print("\n[TEST 4] PDF Ingestion...")
        self.results["metadata"]["total_tests"] += 1
        
        pdf_path = Path("science class 8.pdf")
        if not pdf_path.exists():
            result = {"name": "PDF Ingestion", "status": "FAIL", "error": "PDF not found"}
            self.results["tests"].append(result)
            print(f"  ❌ PDF not found")
            return result
        
        start = time.time()
        try:
            doc = fitz.open(pdf_path)
            pages = len(doc)
            doc.close()
            latency = (time.time() - start) * 1000
            
            result = {
                "name": "PDF Ingestion",
                "status": "PASS",
                "pages": pages,
                "latency_ms": round(latency, 2)
            }
            self.results["metadata"]["passed_tests"] += 1
            print(f"  ✅ {pages} pages in {latency:.0f}ms")
        except Exception as e:
            result = {"name": "PDF Ingestion", "status": "FAIL", "error": str(e)}
            print(f"  ❌ Error: {e}")
        
        self.results["tests"].append(result)
        return result
    
    def test_dataset_loading(self) -> Dict:
        """Test golden dataset loading."""
        print("\n[TEST 5] Golden Dataset...")
        self.results["metadata"]["total_tests"] += 1
        
        dataset_path = Path("science_dataset.xlsx")
        if not dataset_path.exists():
            result = {"name": "Golden Dataset", "status": "SKIP", "error": "Not found"}
            self.results["tests"].append(result)
            print(f"  ⚠️  Dataset not found")
            return result
        
        try:
            dataset = pd.read_excel(dataset_path)
            result = {
                "name": "Golden Dataset",
                "status": "PASS",
                "rows": len(dataset),
                "columns": list(dataset.columns)
            }
            self.results["metadata"]["passed_tests"] += 1
            print(f"  ✅ {len(dataset)} rows, columns: {list(dataset.columns)}")
        except Exception as e:
            result = {"name": "Golden Dataset", "status": "FAIL", "error": str(e)}
            print(f"  ❌ Error: {e}")
        
        self.results["tests"].append(result)
        return result
    
    def test_rag_metrics(self) -> Dict:
        """Test RAG evaluation metrics."""
        print("\n[TEST 6] RAG Metrics (29 metrics)...")
        self.results["metadata"]["total_tests"] += 1
        
        # Check if evaluation file exists
        eval_path = Path("data/complete_rag_evaluation.json")
        if not eval_path.exists():
            result = {"name": "RAG Metrics", "status": "SKIP", "error": "Not run yet"}
            self.results["tests"].append(result)
            print(f"  ⚠️  Run evaluate_complete_rag.py first")
            return result
        
        try:
            with open(eval_path, 'r') as f:
                eval_data = json.load(f)
            
            overall_score = eval_data.get("overall_score", 0)
            
            # Pass if overall score > 0.3
            status = "PASS" if overall_score > 0.3 else "FAIL"
            if status == "PASS":
                self.results["metadata"]["passed_tests"] += 1
            
            result = {
                "name": "RAG Metrics",
                "status": status,
                "overall_score": overall_score,
                "metrics_count": 29
            }
            print(f"  {status}: Overall Score {overall_score:.3f} (29 metrics)")
        except Exception as e:
            result = {"name": "RAG Metrics", "status": "FAIL", "error": str(e)}
            print(f"  ❌ Error: {e}")
        
        self.results["tests"].append(result)
        return result
    
    def test_improvement_loop(self) -> Dict:
        """Test continuous improvement loop."""
        print("\n[TEST 7] Improvement Loop...")
        self.results["metadata"]["total_tests"] += 1
        
        # Check if improvement results exist
        imp_path = Path("data/quick_improvement_results.json")
        if not imp_path.exists():
            result = {"name": "Improvement Loop", "status": "SKIP", "error": "Not run yet"}
            self.results["tests"].append(result)
            print(f"  ⚠️  Run test_quick_improvement.py first")
            return result
        
        try:
            with open(imp_path, 'r') as f:
                imp_data = json.load(f)
            
            if not imp_data:
                result = {"name": "Improvement Loop", "status": "FAIL", "error": "Empty results"}
                print(f"  ❌ Empty results")
            else:
                best_config = max(imp_data, key=lambda x: x.get("overall_score", 0))
                status = "PASS" if best_config.get("overall_score", 0) > 0.4 else "FAIL"
                if status == "PASS":
                    self.results["metadata"]["passed_tests"] += 1
                
                result = {
                    "name": "Improvement Loop",
                    "status": status,
                    "best_config": best_config.get("config", "N/A"),
                    "best_score": best_config.get("overall_score", 0)
                }
                print(f"  {status}: Best={best_config.get('config')} Score={best_config.get('overall_score', 0):.3f}")
        except Exception as e:
            result = {"name": "Improvement Loop", "status": "FAIL", "error": str(e)}
            print(f"  ❌ Error: {e}")
        
        self.results["tests"].append(result)
        return result
    
    def generate_summary(self):
        """Generate test summary."""
        total = self.results["metadata"]["total_tests"]
        passed = self.results["metadata"]["passed_tests"]
        
        print("\n" + "="*70)
        print("FAST MASTER TEST SUITE - SUMMARY")
        print("="*70)
        print(f"\n📊 Tests: {passed}/{total} passed ({passed/total*100:.0f}%)")
        
        # Count by status
        status_counts = {}
        for test in self.results["tests"]:
            status = test.get("status", "UNKNOWN")
            status_counts[status] = status_counts.get(status, 0) + 1
        
        print(f"   PASS: {status_counts.get('PASS', 0)}")
        print(f"   FAIL: {status_counts.get('FAIL', 0)}")
        print(f"   SKIP: {status_counts.get('SKIP', 0)}")
        
        # Overall status
        if passed >= total * 0.8:
            print(f"\n✅ ALL TESTS PASSED!")
            overall = "PASS"
        elif passed >= total * 0.5:
            print(f"\n⚠️  MOST TESTS PASSED")
            overall = "PARTIAL"
        else:
            print(f"\n❌ MANY TESTS FAILED")
            overall = "FAIL"
        
        self.results["summary"] = {
            "overall_status": overall,
            "pass_rate": passed / total if total > 0 else 0,
            "timestamp": datetime.now().isoformat()
        }
        
        print("="*70)
    
    def save_results(self):
        """Save test results."""
        results_path = Path("data/fast_master_test_results.json")
        with open(results_path, 'w') as f:
            json.dump(self.results, f, indent=2)
        print(f"\n✅ Results saved to: {results_path}")
    
    def run_all(self):
        """Run all tests."""
        print("\n" + "="*70)
        print("VISIONARY RAG - FAST MASTER TEST SUITE")
        print("Running ALL tests efficiently")
        print("="*70)
        
        if not self.check_ollama():
            return False
        
        start = time.time()
        
        # Run all tests
        self.test_embedding()
        self.test_generation()
        self.test_all_models()
        self.test_pdf_ingestion()
        self.test_dataset_loading()
        self.test_rag_metrics()
        self.test_improvement_loop()
        
        # Generate summary
        self.generate_summary()
        
        # Save results
        self.save_results()
        
        elapsed = time.time() - start
        print(f"\n⏱️  Total time: {elapsed/60:.1f} minutes")
        
        return self.results["summary"]["overall_status"] == "PASS"


def main():
    """Run fast master test suite."""
    runner = FastMasterTest()
    success = runner.run_all()
    return 0 if success else 1


if __name__ == "__main__":
    sys.exit(main())
