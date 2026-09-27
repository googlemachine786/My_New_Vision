#!/usr/bin/env python
"""
Quick Self-Query Integration Test

Tests that self-query retriever is properly integrated with ask_real_query.py
"""

import sys
import asyncio

def test_imports():
    """Test that all required modules can be imported."""
    print("Testing imports...")
    
    try:
        from retrievers import SelfQueryRetriever, QueryFilter, extract_filters
        print("  ✅ SelfQueryRetriever imported")
    except ImportError as e:
        print(f"  ❌ Failed to import SelfQueryRetriever: {e}")
        return False
    
    try:
        from retrievers.filter_translator import build_sql_filter
        print("  ✅ Filter translator imported")
    except ImportError as e:
        print(f"  ❌ Failed to import filter translator: {e}")
        return False
    
    return True


def test_filter_extraction():
    """Test LLM-based filter extraction."""
    print("\nTesting filter extraction...")
    
    from retrievers import extract_filters
    
    test_cases = [
        ("Show me grade 8 photosynthesis diagrams", 8, "photosynthesis", "Figure"),
        ("questions from chapter 9 about force", None, "force", None),
        ("what is force?", None, None, None),
    ]
    
    all_passed = True
    for query, expected_grade, expected_chapter, expected_type in test_cases:
        try:
            qf = extract_filters(query)
            
            if expected_grade and qf.grade != expected_grade:
                print(f"  ❌ Grade mismatch for '{query}': expected {expected_grade}, got {qf.grade}")
                all_passed = False
            elif expected_chapter and (not qf.chapter_name or expected_chapter not in qf.chapter_name.lower()):
                print(f"  ❌ Chapter mismatch for '{query}': expected '{expected_chapter}', got '{qf.chapter_name}'")
                all_passed = False
            else:
                print(f"  ✅ '{query[:40]}...' - OK")
                
        except Exception as e:
            print(f"  ❌ Error extracting filters for '{query}': {e}")
            all_passed = False
    
    return all_passed


def test_sql_generation():
    """Test SQL filter generation."""
    print("\nTesting SQL filter generation...")
    
    from retrievers import QueryFilter, build_sql_filter
    
    test_cases = [
        (QueryFilter(grade=8, query_text="test"), "(metadata->>'grade')::int = :grade"),
        (QueryFilter(chapter_number=9, query_text="test"), "(metadata->>'chapter_number')::int = :chapter_number"),
        (QueryFilter(query_text="test"), "TRUE"),
    ]
    
    all_passed = True
    for qf, expected_clause in test_cases:
        where, params = build_sql_filter(qf)
        
        if expected_clause not in where:
            print(f"  ❌ SQL mismatch: expected '{expected_clause}' in '{where}'")
            all_passed = False
        else:
            print(f"  ✅ SQL generation OK: {where[:60]}...")
    
    return all_passed


async def test_retriever_search():
    """Test self-query retriever search."""
    print("\nTesting retriever search...")
    
    from retrievers import SelfQueryRetriever
    
    retriever = SelfQueryRetriever()
    
    test_queries = [
        "Show me grade 8 photosynthesis diagrams",
        "what is force?",
    ]
    
    all_passed = True
    for query in test_queries:
        try:
            results = await retriever.search(query)
            
            if not isinstance(results, list):
                print(f"  ❌ Results not a list for '{query}'")
                all_passed = False
            elif len(results) == 0:
                print(f"  ⚠️  No results for '{query}' (expected with mock DB)")
            else:
                print(f"  ✅ Search returned {len(results)} results for '{query[:40]}...'")
                
                # Check metadata
                if results[0].get('metadata'):
                    print(f"     Metadata present: {list(results[0]['metadata'].keys())}")
                
        except Exception as e:
            print(f"  ❌ Search failed for '{query}': {e}")
            all_passed = False
    
    return all_passed


def test_ask_real_query_integration():
    """Test that ask_real_query.py can import self-query."""
    print("\nTesting ask_real_query.py integration...")
    
    try:
        # Check if the module has the right attributes
        with open("ask_real_query.py", encoding="utf-8") as f:
            source = f.read()
        
        if "from retrievers import SelfQueryRetriever" in source:
            print("  ✅ ask_real_query.py imports SelfQueryRetriever")
        else:
            print("  ❌ ask_real_query.py doesn't import SelfQueryRetriever")
            return False
        
        if "self_query_retriever" in source:
            print("  ✅ ask_real_query.py uses self_query_retriever")
        else:
            print("  ❌ ask_real_query.py doesn't use self_query_retriever")
            return False
        
        if "use_self_query" in source:
            print("  ✅ ask_real_query.py has use_self_query flag")
        else:
            print("  ❌ ask_real_query.py doesn't have use_self_query flag")
            return False
        
        return True
        
    except Exception as e:
        print(f"  ❌ Integration test failed: {e}")
        return False


async def run_all_tests():
    """Run all integration tests."""
    print("="*70)
    print("SELF-QUERY INTEGRATION TEST SUITE")
    print("="*70)
    
    tests = [
        ("Imports", test_imports),
        ("Filter Extraction", test_filter_extraction),
        ("SQL Generation", test_sql_generation),
        ("Retriever Search", test_retriever_search),
        ("Integration", test_ask_real_query_integration),
    ]
    
    results = []
    for name, test_func in tests:
        try:
            if asyncio.iscoroutinefunction(test_func):
                result = await test_func()
            else:
                result = test_func()
            results.append((name, result))
        except Exception as e:
            print(f"\n❌ Test '{name}' crashed: {e}")
            import traceback
            traceback.print_exc()
            results.append((name, False))
    
    # Summary
    print("\n" + "="*70)
    print("TEST SUMMARY")
    print("="*70)
    
    passed = sum(1 for _, result in results if result)
    total = len(results)
    
    for name, result in results:
        status = "✅ PASS" if result else "❌ FAIL"
        print(f"  {status}: {name}")
    
    print(f"\nTotal: {passed}/{total} tests passed ({passed/total*100:.1f}%)")
    
    if passed == total:
        print("\n🎉 All tests passed! Self-query is fully integrated.")
        print("\nNext steps:")
        print("  1. Run: python ask_real_query.py")
        print("  2. Try query: 'Show me grade 8 photosynthesis diagrams'")
        print("  3. Toggle self-query with 'toggle' command")
        print("  4. Compare results with/without self-query")
    else:
        print("\n⚠️  Some tests failed. Check the errors above.")
    
    print("="*70)
    
    return passed == total


def main():
    """Main entry point."""
    success = asyncio.run(run_all_tests())
    sys.exit(0 if success else 1)


if __name__ == "__main__":
    main()
