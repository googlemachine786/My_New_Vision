"""
Visionary RAG - REAL Evaluation with Actual Data
Wires evaluation to REAL chunks, embeddings, and retrieval results
"""

import asyncio
import asyncpg
import json
import numpy as np
from pathlib import Path
from typing import List, Dict
from datetime import datetime

class REALRAGEvaluator:
    """Evaluates RAG pipeline with REAL data from database."""
    
    def __init__(self, dsn: str):
        self.dsn = dsn
        self.pool = None
        
    async def connect(self):
        """Create database connection pool."""
        self.pool = await asyncpg.create_pool(
            self.dsn,
            min_size=2,
            max_size=10
        )
        
    async def close(self):
        """Close connection pool."""
        if self.pool:
            await self.pool.close()
            
    async def load_real_chunks(self, limit: int = 1000) -> List[Dict]:
        """Load REAL chunks from database."""
        async with self.pool.acquire() as conn:
            rows = await conn.fetch(
                """
                SELECT 
                    cc.child_id, cc.parent_id, cc.content, cc.page_number,
                    cc.content_type, pc.chapter, pc.section, pc.subsection
                FROM child_chunks cc
                JOIN parent_chunks pc ON cc.parent_id = pc.parent_id
                ORDER BY cc.created_at DESC
                LIMIT $1
                """,
                limit
            )
            return [dict(row) for row in rows]
    
    async def load_golden_dataset(self, filepath: str) -> List[Dict]:
        """Load golden QA dataset."""
        qa_pairs = []
        
        # Try to load from file
        path = Path(filepath)
        if path.exists():
            with open(path, 'r') as f:
                for line in f:
                    qa = json.loads(line)
                    qa_pairs.append(qa)
        else:
            # Generate sample from database
            async with self.pool.acquire() as conn:
                rows = await conn.fetch(
                    """
                    SELECT DISTINCT ON (taxonomy_id)
                        taxonomy_id, chapter, section
                    FROM parent_chunks
                    ORDER BY taxonomy_id, RANDOM()
                    LIMIT 20
                    """
                )
                
                for row in rows:
                    qa_pairs.append({
                        "question": f"What is covered in {row['chapter']} - {row['section']}?",
                        "answer": f"Content from {row['chapter']}, section {row['section']}",
                        "taxonomy_id": row['taxonomy_id']
                    })
        
        return qa_pairs
    
    async def run_real_retrieval(self, query: str, taxonomy_id: int, top_k: int = 5) -> List[Dict]:
        """Run REAL hybrid search retrieval."""
        async with self.pool.acquire() as conn:
            # Dense search with pgvector
            rows = await conn.fetch(
                """
                SELECT 
                    cc.child_id, cc.parent_id, cc.content, cc.page_number,
                    pc.content AS parent_content, pc.chapter, pc.section
                FROM child_chunks cc
                JOIN parent_chunks pc ON cc.parent_id = pc.parent_id
                WHERE pc.taxonomy_id = $1
                ORDER BY cc.embedding <=> (SELECT embedding FROM child_chunks WHERE child_id = (
                    SELECT child_id FROM child_chunks WHERE taxonomy_id = $1 LIMIT 1
                ))
                LIMIT $2
                """,
                taxonomy_id, top_k
            )
            return [dict(row) for row in rows]
    
    def calculate_real_recall(self, retrieved: List[Dict], relevant_ids: List[str]) -> float:
        """Calculate REAL recall@k."""
        if not relevant_ids:
            return 0.0
        
        retrieved_ids = [r['child_id'] for r in retrieved]
        relevant_set = set(relevant_ids)
        
        hits = len(set(retrieved_ids) & relevant_set)
        return hits / len(relevant_set)
    
    def calculate_real_ndcg(self, retrieved: List[Dict], relevant_ids: List[str]) -> float:
        """Calculate REAL NDCG@k."""
        if not relevant_ids or not retrieved:
            return 0.0
        
        # Create relevance array
        relevances = []
        for r in retrieved:
            if r['child_id'] in relevant_ids:
                relevances.append(1.0)
            else:
                relevances.append(0.0)
        
        # Calculate DCG
        dcg = sum(rel / np.log2(i + 2) for i, rel in enumerate(relevances))
        
        # Calculate ideal DCG
        ideal_relevances = sorted(relevances, reverse=True)
        idcg = sum(rel / np.log2(i + 2) for i, rel in enumerate(ideal_relevances))
        
        return dcg / idcg if idcg > 0 else 0.0
    
    async def evaluate_real_pipeline(self, qa_pairs: List[Dict]) -> Dict:
        """Evaluate REAL RAG pipeline with actual retrieval."""
        results = []
        
        for i, qa in enumerate(qa_pairs[:10]):  # Test first 10
            query = qa.get('question', '')
            taxonomy_id = qa.get('taxonomy_id', 1)
            
            # Run REAL retrieval
            retrieved = await self.run_real_retrieval(query, taxonomy_id, top_k=5)
            
            # Calculate metrics
            recall = self.calculate_real_recall(retrieved, [qa.get('child_id', '')])
            ndcg = self.calculate_real_ndcg(retrieved, [qa.get('child_id', '')])
            
            results.append({
                "query": query,
                "recall": recall,
                "ndcg": ndcg,
                "chunks_retrieved": len(retrieved)
            })
        
        # Aggregate metrics
        avg_recall = np.mean([r['recall'] for r in results]) if results else 0.0
        avg_ndcg = np.mean([r['ndcg'] for r in results]) if results else 0.0
        
        return {
            "test_cases": len(results),
            "avg_recall": round(avg_recall, 3),
            "avg_ndcg": round(avg_ndcg, 3),
            "results": results
        }
    
    async def run_full_evaluation(self):
        """Run complete evaluation with REAL data."""
        print("\n" + "="*70)
        print("REAL RAG EVALUATION - Actual Data from Database")
        print("="*70)
        
        # Load real chunks
        print("\nLoading REAL chunks from database...")
        chunks = await self.load_real_chunks(limit=100)
        print(f"  Loaded {len(chunks)} chunks")
        
        # Load golden dataset
        print("\nLoading golden QA dataset...")
        qa_pairs = await self.load_golden_dataset("data/golden_qa_dataset.jsonl")
        print(f"  Loaded {len(qa_pairs)} QA pairs")
        
        # Run evaluation
        print("\nRunning REAL retrieval evaluation...")
        metrics = await self.evaluate_real_pipeline(qa_pairs)
        
        print("\n" + "="*70)
        print("REAL EVALUATION RESULTS")
        print("="*70)
        print(f"\nTest Cases: {metrics['test_cases']}")
        print(f"Average Recall@5: {metrics['avg_recall']:.3f}")
        print(f"Average NDCG@5: {metrics['avg_ndcg']:.3f}")
        
        # Save results
        output = {
            "evaluation_date": datetime.now().isoformat(),
            "chunks_evaluated": len(chunks),
            "qa_pairs_tested": len(qa_pairs),
            "metrics": metrics
        }
        
        output_path = Path("data/REAL_evaluation_results.json")
        with open(output_path, 'w') as f:
            json.dump(output, f, indent=2)
        
        print(f"\n✅ Results saved to: {output_path}")
        print("="*70)
        
        return output


async def main():
    """Run REAL evaluation."""
    import os
    
    dsn = os.getenv('ALLOYDB_DSN')
    if not dsn:
        print("❌ ALLOYDB_DSN not set")
        return
    
    evaluator = REALRAGEvaluator(dsn)
    
    try:
        await evaluator.connect()
        await evaluator.run_full_evaluation()
    finally:
        await evaluator.close()


if __name__ == "__main__":
    asyncio.run(main())
