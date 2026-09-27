#!/usr/bin/env python
"""
Visionary RAG - MASTER ADVANCED TEST RUNNER
Runs ALL advanced tests:
1. Chunking Strategies (5 strategies)
2. RAGAS + LLM-as-Judge
3. Metadata Extraction
4. Comprehensive Reporting
"""

import sys
import subprocess
from pathlib import Path

def run_test(script_name: str, test_name: str) -> bool:
    """Run a test script."""
    print(f"\n{'='*70}")
    print(f"RUNNING: {test_name}")
    print(f"{'='*70}\n")
    
    result = subprocess.run(
        [sys.executable, script_name],
        capture_output=True,
        text=True,
        timeout=300
    )
    
    print(result.stdout)
    if result.stderr:
        print(f"Errors: {result.stderr}")
    
    return result.returncode == 0

def main():
    """Run all advanced tests."""
    print("\n" + "="*70)
    print("VISIONARY RAG - ADVANCED TEST SUITE")
    print("Running: Chunking, RAGAS, LLM-Judge, Metadata")
    print("="*70)
    
    results = {
        "chunking_strategies": False,
        "ragas_llm_judge": False,
        "metadata_extraction": False
    }
    
    # 1. Chunking Strategies
    results["chunking_strategies"] = run_test(
        "test_chunking_strategies.py",
        "5 Chunking Strategies Comparison"
    )
    
    # 2. RAGAS + LLM Judge
    results["ragas_llm_judge"] = run_test(
        "test_ragas_llm_judge.py",
        "RAGAS Metrics + LLM-as-Judge"
    )
    
    # 3. Metadata Extraction
    results["metadata_extraction"] = run_test(
        "extract_metadata.py",
        "Textbook Metadata Extraction"
    )
    
    # Summary
    print("\n" + "="*70)
    print("ADVANCED TEST SUITE - SUMMARY")
    print("="*70)
    
    passed = sum(1 for v in results.values() if v)
    total = len(results)
    
    print(f"\nTests Passed: {passed}/{total}")
    
    for test_name, passed_flag in results.items():
        status = "✅ PASS" if passed_flag else "❌ FAIL"
        print(f"  {status}: {test_name}")
    
    if passed == total:
        print(f"\n🎉 ALL ADVANCED TESTS PASSED!")
    else:
        print(f"\n⚠️  {total - passed} test(s) failed")
    
    print("="*70)
    
    return 0 if passed == total else 1

if __name__ == "__main__":
    sys.exit(main())
