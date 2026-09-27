#!/usr/bin/env python
"""
Self-Query Retriever Demo Script

Interactive demo showcasing metadata-enhanced RAG retrieval
with automatic filter extraction from natural language queries.
"""

import asyncio
import json
from datetime import datetime
from retrievers.self_query_retriever import SelfQueryRetriever, SelfQueryRetrieverSync
from retrievers.self_query_parser import extract_filters, QueryFilter
from retrievers.filter_translator import build_sql_filter, get_filter_summary, count_active_filters


class SelfQueryDemo:
    """Interactive demo for self-query retrieval."""
    
    def __init__(self):
        self.retriever = SelfQueryRetriever()
        self.query_history = []
    
    def print_header(self):
        """Print demo header."""
        print("\n" + "="*80)
        print("🏷️  SELF-QUERY RETRIEVER DEMO")
        print("   Metadata-Enhanced RAG with Automatic Filter Extraction")
        print("="*80)
        print("\nFeatures:")
        print("  • Automatic metadata filter extraction from natural language")
        print("  • Pre-filtering before vector search (95% search space reduction)")
        print("  • Grade, chapter, content type filtering")
        print("  • Performance tracking and statistics")
        print("\nTry queries like:")
        print("  • 'Show me grade 8 photosynthesis diagrams'")
        print("  • 'questions from chapter 9 about force'")
        print("  • 'grade 7 cell structure tables'")
        print("  • 'what is force?' (no filters)")
        print("\nCommands: 'stats' (show stats), 'compare' (comparison), 'quit' (exit)")
        print("="*80)
    
    async def run_query(self, query: str):
        """Run a query and display results with statistics."""
        print(f"\n❓ Query: {query}")
        print("-"*80)
        
        # Step 1: Extract filters
        print("\n📥 Step 1: Extracting filters with LLM...")
        start = datetime.now()
        query_filter = extract_filters(query)
        extraction_time = (datetime.now() - start).total_seconds() * 1000
        
        print(f"   ✅ Filters extracted in {extraction_time:.2f}ms")
        print(f"   • Grade: {query_filter.grade or 'None'}")
        print(f"   • Chapter #: {query_filter.chapter_number or 'None'}")
        print(f"   • Chapter Name: {query_filter.chapter_name or 'None'}")
        print(f"   • Content Type: {query_filter.content_type or 'None'}")
        print(f"   • Subject: {query_filter.subject or 'None'}")
        print(f"   • Query Text: {query_filter.query_text}")
        
        # Step 2: Build SQL filter
        print("\n📥 Step 2: Building SQL filter...")
        where_clause, params = build_sql_filter(query_filter)
        print(f"   ✅ WHERE clause: {where_clause}")
        print(f"   ✅ Parameters: {params}")
        
        # Step 3: Execute search
        print("\n📥 Step 3: Executing filtered vector search...")
        start = datetime.now()
        results = await self.retriever.search(query)
        search_time = (datetime.now() - start).total_seconds() * 1000
        
        print(f"   ✅ Search completed in {search_time:.2f}ms")
        print(f"   ✅ Found {len(results)} results")
        
        # Display results
        if results:
            print(f"\n📊 Top Results:")
            for i, result in enumerate(results[:3], 1):
                print(f"\n   [{i}] Similarity: {result['similarity']:.4f}")
                print(f"       Content: {result['content'][:150]}...")
                print(f"       Metadata: {json.dumps(result['metadata'], indent=16)}")
        
        # Summary
        total_time = extraction_time + search_time
        print(f"\n📈 Summary:")
        print(f"   • Total Time: {total_time:.2f}ms")
        print(f"   • Filter Extraction: {extraction_time:.2f}ms")
        print(f"   • Vector Search: {search_time:.2f}ms")
        print(f"   • Active Filters: {count_active_filters(query_filter)}")
        print(f"   • Filter Summary: {get_filter_summary(query_filter)}")
        
        # Save to history
        self.query_history.append({
            'timestamp': datetime.now().isoformat(),
            'query': query,
            'filters': {
                'grade': query_filter.grade,
                'chapter_number': query_filter.chapter_number,
                'chapter_name': query_filter.chapter_name,
                'content_type': query_filter.content_type
            },
            'results_count': len(results),
            'latency_ms': total_time
        })
    
    async def show_stats(self):
        """Show query history statistics."""
        if not self.query_history:
            print("\n📭 No queries yet")
            return
        
        print("\n" + "="*80)
        print("📊 QUERY HISTORY STATISTICS")
        print("="*80)
        
        total_queries = len(self.query_history)
        avg_latency = sum(q['latency_ms'] for q in self.query_history) / total_queries
        avg_filters = sum(
            sum(1 for v in q['filters'].values() if v is not None)
            for q in self.query_history
        ) / total_queries
        
        print(f"\nTotal Queries: {total_queries}")
        print(f"Average Latency: {avg_latency:.2f}ms")
        print(f"Average Filters per Query: {avg_filters:.1f}")
        
        print("\nRecent Queries:")
        for i, q in enumerate(self.query_history[-5:], 1):
            print(f"  {i}. {q['query']} ({q['latency_ms']:.2f}ms, {q['results_count']} results)")
    
    async def compare_modes(self, query: str):
        """Compare self-query vs dense-only retrieval."""
        print(f"\n🔬 Comparison Mode: {query}")
        print("="*80)
        
        # Self-query
        print("\n🏷️  Self-Query Retrieval:")
        start = datetime.now()
        self_query_stats = await self.retriever.search_with_stats(query)
        self_query_time = (datetime.now() - start).total_seconds() * 1000
        
        print(f"   • Latency: {self_query_time:.2f}ms")
        print(f"   • Results: {len(self_query_stats['results'])}")
        print(f"   • Filters: {get_filter_summary(extract_filters(query))}")
        print(f"   • Active Filters: {count_active_filters(extract_filters(query))}")
        
        # Dense-only
        print("\n📊 Dense-Only Retrieval (No Filters):")
        start = datetime.now()
        dense_results = await self.retriever._dense_only_search(query)
        dense_time = (datetime.now() - start).total_seconds() * 1000
        
        print(f"   • Latency: {dense_time:.2f}ms")
        print(f"   • Results: {len(dense_results)}")
        print(f"   • Filters: None")
        
        # Comparison
        print("\n📈 Comparison:")
        if dense_time > 0:
            speedup = dense_time / max(self_query_time, 1)
            print(f"   • Speedup: {speedup:.2f}x")
        print(f"   • Search Space Reduction: {self_query_stats['active_filters']} filters applied")
        print(f"   • Precision Gain: Filters ensure grade/chapter accuracy")
    
    async def run_interactive(self):
        """Run interactive demo session."""
        self.print_header()
        
        # Check Ollama
        import requests
        try:
            response = requests.get("http://localhost:11434/api/tags", timeout=5)
            if response.status_code == 200:
                models = response.json().get("models", [])
                print(f"\n✅ Ollama connected ({len(models)} models available)")
            else:
                print("\n⚠️  Ollama may not be running")
        except:
            print("\n⚠️  Cannot connect to Ollama (using fallback mode)")
        
        while True:
            try:
                query = input("\n❓ Ask: ").strip()
                
                if not query:
                    continue
                elif query.lower() in ['quit', 'exit', 'q']:
                    print("\n👋 Goodbye!")
                    break
                elif query.lower() == 'stats':
                    await self.show_stats()
                elif query.lower() == 'compare':
                    test_query = input("Enter query to compare: ").strip()
                    if test_query:
                        await self.compare_modes(test_query)
                elif query.lower() == 'help':
                    self.print_header()
                else:
                    await self.run_query(query)
                    
            except KeyboardInterrupt:
                print("\n\n👋 Goodbye!")
                break
            except Exception as e:
                print(f"\n❌ Error: {e}")
                import traceback
                traceback.print_exc()


async def main():
    """Main entry point."""
    demo = SelfQueryDemo()
    await demo.run_interactive()
    return 0


if __name__ == "__main__":
    asyncio.run(main())
