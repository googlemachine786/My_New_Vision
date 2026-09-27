#!/usr/bin/env python
"""
Visionary RAG Pipeline - Continuous Improvement Loop
Tests different models, configurations, and tracks RAG evaluation metrics
"""

import os
import sys
import json
import time
import requests
import pandas as pd
from pathlib import Path
from typing import List, Dict, Any, Tuple
from datetime import datetime
from dataclasses import dataclass, asdict

@dataclass
class ModelConfig:
    """Model configuration for testing."""
    name: str
    embed_model: str
    llm_model: str
    top_k: int = 3
    chunk_size: int = 500
    num_chunks: int = 50

@dataclass
class EvaluationMetrics:
    """RAG evaluation metrics."""
    # Retrieval Metrics
    recall_at_1: float = 0.0
    recall_at_3: float = 0.0
    recall_at_5: float = 0.0
    mrr: float = 0.0  # Mean Reciprocal Rank
    ndcg_at_10: float = 0.0
    
    # Generation Metrics
    answer_relevance: float = 0.0
    faithfulness: float = 0.0
    answer_correctness: float = 0.0
    
    # Performance Metrics
    avg_latency_ms: float = 0.0
    p95_latency_ms: float = 0.0
    p99_latency_ms: float = 0.0
    success_rate: float = 0.0
    
    # Overall
    overall_score: float = 0.0

class ContinuousImprovementLoop:
    """Continuous testing and improvement of RAG pipeline."""
    
    def __init__(self):
        self.pdf_path = Path("science class 8.pdf")
        self.dataset_path = Path("science_dataset.xlsx")
        self.ollama_url = os.getenv("OLLAMA_BASE_URL", "http://localhost:11434")
        
        # Model configurations to test
        self.model_configs = [
            ModelConfig(
                name="llama3.2-fast",
                embed_model="nomic-embed-text",
                llm_model="llama3.2:3b",
                top_k=3,
                num_chunks=50
            ),
            ModelConfig(
                name="llama3.2-accurate",
                embed_model="nomic-embed-text",
                llm_model="llama3.2:latest",
                top_k=5,
                num_chunks=50
            ),
            ModelConfig(
                name="mistral-quality",
                embed_model="nomic-embed-text",
                llm_model="mistral:7b",
                top_k=5,
                num_chunks=50
            ),
        ]
        
        self.results_history = []
        self.chunks = []
        self.embeddings = []
        
    def check_ollama(self):
        """Check if Ollama is running and list available models."""
        try:
            response = requests.get(f"{self.ollama_url}/api/tags", timeout=5)
            if response.status_code == 200:
                models = response.json().get("models", [])
                print(f"✅ Ollama is running")
                print(f"   Available models: {len(models)}")
                for model in models:
                    print(f"   - {model['name']}")
                return True
            return False
        except Exception as e:
            print(f"❌ Cannot connect to Ollama: {e}")
            return False
    
    def get_embedding(self, text: str, model: str) -> List[float]:
        """Get embedding from Ollama."""
        response = requests.post(
            f"{self.ollama_url}/api/embeddings",
            json={"model": model, "prompt": text},
            timeout=60
        )
        response.raise_for_status()
        return response.json().get("embedding", [])
    
    def generate_response(self, query: str, context: str, model: str) -> str:
        """Generate response using Ollama LLM."""
        prompt = f"""You are a helpful science tutor for CBSE Class 8 students.
Use ONLY the following context to answer the question.
If the answer is not in the context, say "I don't have information about this in the provided context."
Always cite page numbers when available.

Context:
{context}

Question: {query}

Answer (be concise and educational):"""

        response = requests.post(
            f"{self.ollama_url}/api/generate",
            json={"model": model, "prompt": prompt, "stream": False},
            timeout=120
        )
        response.raise_for_status()
        return response.json().get("response", "")
    
    def extract_and_embed_chunks(self, config: ModelConfig):
        """Extract text and create embeddings."""
        import fitz
        
        print(f"\n{'='*70}")
        print(f"EXTRACTING & EMBEDDING CHUNKS")
        print(f"Configuration: {config.name}")
        print(f"{'='*70}")
        
        doc = fitz.open(self.pdf_path)
        total_pages = len(doc)
        
        all_chunks = []
        sample_pages = list(range(0, total_pages, max(1, total_pages // config.num_chunks)))
        
        for page_num in sample_pages[:config.num_chunks]:
            if page_num >= total_pages:
                continue
                
            page = doc[page_num]
            text = page.get_text()
            
            if len(text.strip()) < 100:
                continue
            
            # Split into chunks
            chunk_size = config.chunk_size
            for i in range(0, len(text), chunk_size):
                chunk = text[i:i+chunk_size]
                if len(chunk.strip()) > 50:
                    all_chunks.append({
                        "content": chunk,
                        "page": page_num + 1,
                        "chunk_id": f"p{page_num+1}_c{i//chunk_size}"
                    })
        
        doc.close()
        
        # Create embeddings
        print(f"Creating embeddings with {config.embed_model}...")
        self.chunks = []
        self.embeddings = []
        
        for i, chunk in enumerate(all_chunks[:min(50, len(all_chunks))], 1):
            try:
                embedding = self.get_embedding(chunk["content"], config.embed_model)
                self.chunks.append(chunk)
                self.embeddings.append(embedding)
                if i % 10 == 0:
                    print(f"  Embedded {i}/{min(50, len(all_chunks))} chunks...")
            except Exception as e:
                print(f"  ❌ Chunk {i} failed: {e}")
        
        print(f"✅ Embedded {len(self.chunks)} chunks ({len(self.embeddings[0]) if self.embeddings else 0} dims)")
        return len(self.chunks)
    
    def retrieve_chunks(self, query: str, top_k: int) -> List[Tuple[int, float]]:
        """Retrieve top-k most similar chunks."""
        query_embedding = self.get_embedding(query, "nomic-embed-text")
        
        chunk_scores = []
        for j, chunk_emb in enumerate(self.embeddings):
            score = sum(a*b for a, b in zip(query_embedding, chunk_emb))
            chunk_scores.append((j, score))
        
        chunk_scores.sort(key=lambda x: x[1], reverse=True)
        return chunk_scores[:top_k]
    
    def calculate_retrieval_metrics(self, query: str, relevant_chunk_pages: List[int]) -> Dict[str, float]:
        """Calculate retrieval metrics (Recall@K, MRR, NDCG)."""
        top_indices = self.retrieve_chunks(query, top_k=5)
        
        # Get retrieved pages
        retrieved_pages = [self.chunks[idx][0] for idx, _ in top_indices]
        
        # Recall@K
        def recall_at_k(k):
            retrieved_at_k = set(retrieved_pages[:k])
            relevant_set = set(relevant_chunk_pages)
            if not relevant_set:
                return 0.0
            return len(retrieved_at_k & relevant_set) / len(relevant_set)
        
        # MRR (Mean Reciprocal Rank)
        mrr = 0.0
        for i, page in enumerate(retrieved_pages):
            if page in relevant_chunk_pages:
                mrr = 1.0 / (i + 1)
                break
        
        # NDCG@10
        def dcg(relevances):
            return sum(rel / (i + 2) for i, rel in enumerate(relevances))
        
        relevances = [1.0 if page in relevant_chunk_pages else 0.0 for page in retrieved_pages]
        dcg_score = dcg(relevances)
        ideal_relevances = sorted(relevances, reverse=True)
        idcg_score = dcg(ideal_relevances)
        ndcg = dcg_score / idcg_score if idcg_score > 0 else 0.0
        
        return {
            "recall_at_1": recall_at_k(1),
            "recall_at_3": recall_at_k(3),
            "recall_at_5": recall_at_k(5),
            "mrr": mrr,
            "ndcg_at_10": ndcg
        }
    
    def calculate_generation_metrics(self, query: str, answer: str, golden_answer: str, context: str) -> Dict[str, float]:
        """Calculate generation metrics (Relevance, Faithfulness, Correctness)."""
        
        # Answer Relevance (does answer address the query?)
        answer_relevance = self.calculate_text_similarity(query, answer)
        
        # Faithfulness (is answer grounded in context?)
        faithfulness = self.calculate_text_similarity(answer, context)
        
        # Answer Correctness (similarity to golden answer)
        answer_correctness = self.calculate_text_similarity(answer, golden_answer)
        
        return {
            "answer_relevance": answer_relevance,
            "faithfulness": faithfulness,
            "answer_correctness": answer_correctness
        }
    
    def calculate_text_similarity(self, text1: str, text2: str) -> float:
        """Calculate cosine similarity between two texts using embeddings."""
        try:
            emb1 = self.get_embedding(text1[:500], "nomic-embed-text")  # Truncate for speed
            emb2 = self.get_embedding(text2[:500], "nomic-embed-text")
            
            dot_product = sum(a*b for a, b in zip(emb1, emb2))
            norm1 = sum(a*a for a in emb1) ** 0.5
            norm2 = sum(a*a for a in emb2) ** 0.5
            
            if norm1 == 0 or norm2 == 0:
                return 0.0
            
            return dot_product / (norm1 * norm2)
        except:
            return 0.0
    
    def run_test_iteration(self, config: ModelConfig) -> EvaluationMetrics:
        """Run one iteration of testing with given configuration."""
        print(f"\n{'='*70}")
        print(f"TESTING CONFIGURATION: {config.name}")
        print(f"Embed Model: {config.embed_model}")
        print(f"LLM Model: {config.llm_model}")
        print(f"Top-K: {config.top_k}")
        print(f"{'='*70}")
        
        # Extract and embed chunks
        self.extract_and_embed_chunks(config)
        
        if not self.chunks:
            print("❌ No chunks embedded, skipping...")
            return EvaluationMetrics()
        
        # Load golden dataset
        if self.dataset_path.exists():
            dataset = pd.read_excel(self.dataset_path)
            test_questions = dataset.head(10)  # Test with first 10 questions
        else:
            print("⚠️  No golden dataset, using sample questions")
            test_questions = pd.DataFrame({
                'x_data': ["What is force?", "What are the parts of a cell?"],
                'y_data': ["A push or pull", "Nucleus, cytoplasm, cell membrane"]
            })
        
        # Test each question
        metrics_list = []
        latencies = []
        success_count = 0
        
        for idx, row in test_questions.iterrows():
            query = row['x_data']
            golden_answer = row['y_data'] if 'y_data' in row else ""
            
            print(f"\n  [{idx+1}/{len(test_questions)}] Query: {query}")
            
            start_time = time.time()
            
            try:
                # Retrieve chunks
                top_chunks = self.retrieve_chunks(query, config.top_k)
                
                # Build context
                context_parts = []
                for chunk_idx, _ in top_chunks:
                    chunk = self.chunks[chunk_idx]
                    context_parts.append(f"[Page {chunk['page']}]\n{chunk['content']}")
                context = "\n\n---\n\n".join(context_parts)
                
                # Generate answer
                answer = self.generate_response(query, context, config.llm_model)
                
                latency = (time.time() - start_time) * 1000
                latencies.append(latency)
                success_count += 1
                
                # Calculate metrics
                retrieval_metrics = self.calculate_retrieval_metrics(
                    query, 
                    [self.chunks[idx][0] for idx, _ in top_chunks]
                )
                
                generation_metrics = self.calculate_generation_metrics(
                    query, answer, golden_answer, context
                )
                
                metrics_list.append({
                    "query": query,
                    "answer": answer[:200],
                    "latency_ms": latency,
                    **retrieval_metrics,
                    **generation_metrics
                })
                
                print(f"     Answer: {answer[:100]}...")
                print(f"     Latency: {latency:.2f}ms")
                
            except Exception as e:
                print(f"     ❌ Error: {e}")
                latencies.append(0)
        
        # Aggregate metrics
        if not metrics_list:
            return EvaluationMetrics()
        
        avg_metrics = {}
        for key in metrics_list[0].keys():
            if key not in ["query", "answer"]:
                values = [m[key] for m in metrics_list if isinstance(m[key], (int, float))]
                avg_metrics[key] = sum(values) / len(values) if values else 0.0
        
        # Calculate overall metrics
        metrics = EvaluationMetrics(
            recall_at_1=avg_metrics.get("recall_at_1", 0.0),
            recall_at_3=avg_metrics.get("recall_at_3", 0.0),
            recall_at_5=avg_metrics.get("recall_at_5", 0.0),
            mrr=avg_metrics.get("mrr", 0.0),
            ndcg_at_10=avg_metrics.get("ndcg_at_10", 0.0),
            answer_relevance=avg_metrics.get("answer_relevance", 0.0),
            faithfulness=avg_metrics.get("faithfulness", 0.0),
            answer_correctness=avg_metrics.get("answer_correctness", 0.0),
            avg_latency_ms=sum(latencies) / len(latencies) if latencies else 0.0,
            p95_latency_ms=sorted(latencies)[int(len(latencies) * 0.95)] if latencies else 0.0,
            p99_latency_ms=sorted(latencies)[int(len(latencies) * 0.99)] if latencies else 0.0,
            success_rate=success_count / len(test_questions)
        )
        
        # Calculate overall score (weighted average)
        metrics.overall_score = (
            0.25 * metrics.recall_at_5 +
            0.15 * metrics.mrr +
            0.20 * metrics.answer_relevance +
            0.20 * metrics.faithfulness +
            0.20 * metrics.answer_correctness
        )
        
        print(f"\n✅ Configuration {config.name} complete")
        print(f"   Overall Score: {metrics.overall_score:.3f}")
        print(f"   Success Rate: {metrics.success_rate * 100:.1f}%")
        print(f"   Avg Latency: {metrics.avg_latency_ms:.2f}ms")
        
        return metrics
    
    def run_continuous_improvement(self, num_iterations: int = 3):
        """Run continuous improvement loop."""
        print("\n" + "="*70)
        print("CONTINUOUS IMPROVEMENT LOOP")
        print("="*70)
        
        if not self.check_ollama():
            print("\n❌ Ollama is not running. Please start Ollama:")
            print("   ollama serve")
            return
        
        all_results = []
        
        for iteration in range(num_iterations):
            print(f"\n{'='*70}")
            print(f"ITERATION {iteration + 1}/{num_iterations}")
            print(f"{'='*70}")
            
            for config in self.model_configs:
                # Check if model is available
                try:
                    response = requests.get(f"{self.ollama_url}/api/tags", timeout=5)
                    available_models = [m['name'] for m in response.json().get('models', [])]
                    
                    if config.llm_model not in available_models:
                        print(f"\n⚠️  Model {config.llm_model} not available, skipping...")
                        continue
                except:
                    pass
                
                # Run test
                metrics = self.run_test_iteration(config)
                
                result = {
                    "iteration": iteration + 1,
                    "timestamp": datetime.now().isoformat(),
                    "config": asdict(config),
                    "metrics": asdict(metrics)
                }
                
                all_results.append(result)
                self.results_history.append(result)
        
        # Save results
        self.save_results(all_results)
        
        # Print summary
        self.print_improvement_summary(all_results)
    
    def save_results(self, results: List[Dict]):
        """Save results to JSON and CSV."""
        data_dir = Path("data")
        data_dir.mkdir(exist_ok=True)
        
        # Save JSON
        json_path = data_dir / "continuous_improvement_results.json"
        with open(json_path, "w", encoding="utf-8") as f:
            json.dump(results, f, indent=2, ensure_ascii=False)
        print(f"\n✅ Results saved to: {json_path}")
        
        # Save CSV for easier analysis
        csv_rows = []
        for result in results:
            row = {
                "iteration": result["iteration"],
                "timestamp": result["timestamp"],
                "config_name": result["config"]["name"],
                **result["metrics"]
            }
            csv_rows.append(row)
        
        if csv_rows:
            df = pd.DataFrame(csv_rows)
            csv_path = data_dir / "continuous_improvement_results.csv"
            df.to_csv(csv_path, index=False)
            print(f"✅ Results saved to: {csv_path}")
    
    def print_improvement_summary(self, results: List[Dict]):
        """Print improvement summary."""
        print("\n" + "="*70)
        print("IMPROVEMENT SUMMARY")
        print("="*70)
        
        if not results:
            print("No results to summarize")
            return
        
        # Group by config
        config_scores = {}
        for result in results:
            config_name = result["config"]["name"]
            overall_score = result["metrics"]["overall_score"]
            
            if config_name not in config_scores:
                config_scores[config_name] = []
            config_scores[config_name].append(overall_score)
        
        print("\nConfiguration Performance:")
        print("-" * 70)
        print(f"{'Config':<25} {'Avg Score':<12} {'Best Score':<12} {'Tests':<8}")
        print("-" * 70)
        
        for config_name, scores in sorted(config_scores.items(), key=lambda x: max(x[1]), reverse=True):
            avg_score = sum(scores) / len(scores)
            best_score = max(scores)
            print(f"{config_name:<25} {avg_score:<12.3f} {best_score:<12.3f} {len(scores):<8}")
        
        print("-" * 70)
        
        # Find best configuration
        best_config = max(config_scores.items(), key=lambda x: max(x[1]))
        print(f"\n🏆 Best Configuration: {best_config[0]}")
        print(f"   Best Score: {max(best_config[1]):.3f}")
        
        print("\n✅ Continuous improvement loop complete!")
        print("\nNext steps:")
        print("  1. Review results in data/continuous_improvement_results.csv")
        print("  2. Deploy best configuration to production")
        print("  3. Continue monitoring and testing")


def main():
    """Run continuous improvement loop."""
    loop = ContinuousImprovementLoop()
    loop.run_continuous_improvement(num_iterations=2)
    return 0


if __name__ == "__main__":
    sys.exit(main())
