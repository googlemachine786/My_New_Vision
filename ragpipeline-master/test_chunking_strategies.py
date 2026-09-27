"""
Visionary RAG - Advanced Chunking Strategies Testing
Tests 5 different chunking strategies with comprehensive metrics
"""

import numpy as np
from typing import List, Dict, Tuple
from dataclasses import dataclass, asdict
import fitz
import time

@dataclass
class ChunkingStrategy:
    """Base chunking strategy."""
    name: str
    chunk_size: int = 500
    overlap: int = 50

@dataclass
class ChunkingMetrics:
    """Metrics for chunking evaluation."""
    strategy_name: str
    num_chunks: int = 0
    avg_chunk_size: float = 0.0
    chunk_coherence: float = 0.0
    retrieval_recall: float = 0.0
    answer_faithfulness: float = 0.0
    latency_ms: float = 0.0
    memory_mb: float = 0.0
    overall_score: float = 0.0

class FixedSizeChunking(ChunkingStrategy):
    """Fixed-size chunking (baseline)."""
    
    def __init__(self):
        super().__init__("fixed_size", chunk_size=500, overlap=75)
    
    def chunk(self, text: str) -> List[Dict]:
        chunks = []
        for i in range(0, len(text), self.chunk_size - self.overlap):
            chunk_text = text[i:i + self.chunk_size]
            if len(chunk_text.strip()) > 20:
                chunks.append({
                    "content": chunk_text,
                    "start": i,
                    "end": i + len(chunk_text)
                })
        return chunks

class SemanticChunking(ChunkingStrategy):
    """Semantic chunking by meaning (sentence boundaries)."""
    
    def __init__(self):
        super().__init__("semantic", chunk_size=500, overlap=50)
    
    def chunk(self, text: str) -> List[Dict]:
        # Split by sentences
        sentences = []
        current = ""
        for char in text:
            current += char
            if char in '.!?':
                if len(current.strip()) > 10:
                    sentences.append(current.strip())
                current = ""
        
        # Group sentences into semantic chunks
        chunks = []
        current_chunk = ""
        start_pos = 0
        
        for i, sent in enumerate(sentences):
            if len(current_chunk) + len(sent) < self.chunk_size:
                if not current_chunk:
                    start_pos = text.find(sent)
                current_chunk += " " + sent
            else:
                if current_chunk:
                    chunks.append({
                        "content": current_chunk.strip(),
                        "start": start_pos,
                        "end": start_pos + len(current_chunk),
                        "sentences": len(current_chunk.split('.')) - 1
                    })
                current_chunk = sent
                start_pos = text.find(sent)
        
        # Add last chunk
        if current_chunk:
            chunks.append({
                "content": current_chunk.strip(),
                "start": start_pos,
                "end": start_pos + len(current_chunk),
                "sentences": len(current_chunk.split('.')) - 1
            })
        
        return chunks

class RecursiveChunking(ChunkingStrategy):
    """Recursive chunking by structure (paragraphs, sentences, words)."""
    
    def __init__(self):
        super().__init__("recursive", chunk_size=500, overlap=50)
        self.separators = ["\n\n", "\n", ". ", " ", ""]
    
    def chunk(self, text: str) -> List[Dict]:
        chunks = []
        self._recursive_split(text, 0, chunks)
        return chunks
    
    def _recursive_split(self, text: str, depth: int, chunks: List):
        if len(text) <= self.chunk_size:
            if len(text.strip()) > 20:
                chunks.append({
                    "content": text.strip(),
                    "start": 0,
                    "end": len(text)
                })
            return
        
        if depth >= len(self.separators):
            # Force split
            for i in range(0, len(text), self.chunk_size):
                chunk = text[i:i + self.chunk_size]
                if len(chunk.strip()) > 20:
                    chunks.append({
                        "content": chunk.strip(),
                        "start": i,
                        "end": i + len(chunk)
                    })
            return
        
        separator = self.separators[depth]
        parts = text.split(separator)
        
        if len(parts) > 1:
            current = ""
            for part in parts:
                if len(current) + len(part) < self.chunk_size:
                    current += part + separator
                else:
                    self._recursive_split(current.strip(), depth + 1, chunks)
                    current = part + separator
            if current:
                self._recursive_split(current.strip(), depth + 1, chunks)
        else:
            self._recursive_split(text, depth + 1, chunks)

class AgenticChunking(ChunkingStrategy):
    """LLM-guided chunking (uses simple heuristics as proxy for LLM)."""
    
    def __init__(self):
        super().__init__("agentic", chunk_size=600, overlap=100)
    
    def chunk(self, text: str) -> List[Dict]:
        # Simulate LLM-guided chunking by finding topic boundaries
        chunks = []
        
        # Find potential topic boundaries (headers, new lines)
        lines = text.split('\n')
        current_chunk = ""
        start_line = 0
        
        for i, line in enumerate(lines):
            # Check if line is a header (short, no period)
            is_header = len(line) < 100 and '.' not in line and len(line) > 5
            
            if is_header and current_chunk:
                # Save current chunk at topic boundary
                if len(current_chunk.strip()) > 20:
                    chunks.append({
                        "content": current_chunk.strip(),
                        "start": start_line,
                        "end": i,
                        "has_header": True
                    })
                current_chunk = line + "\n"
                start_line = i
            else:
                current_chunk += line + "\n"
                
                # Check chunk size
                if len(current_chunk) > self.chunk_size:
                    chunks.append({
                        "content": current_chunk.strip(),
                        "start": start_line,
                        "end": i,
                        "has_header": False
                    })
                    current_chunk = ""
                    start_line = i
        
        # Add last chunk
        if current_chunk and len(current_chunk.strip()) > 20:
            chunks.append({
                "content": current_chunk.strip(),
                "start": start_line,
                "end": len(lines),
                "has_header": False
            })
        
        return chunks

class HybridChunking(ChunkingStrategy):
    """Hybrid chunking (combines multiple strategies)."""
    
    def __init__(self):
        super().__init__("hybrid", chunk_size=500, overlap=75)
    
    def chunk(self, text: str) -> List[Dict]:
        # Use semantic for first pass
        semantic = SemanticChunking()
        semantic_chunks = semantic.chunk(text)
        
        # Merge small chunks
        merged_chunks = []
        current = ""
        start = 0
        
        for chunk in semantic_chunks:
            if len(current) + len(chunk["content"]) < self.chunk_size:
                if not current:
                    start = chunk["start"]
                current += " " + chunk["content"]
            else:
                if current:
                    merged_chunks.append({
                        "content": current.strip(),
                        "start": start,
                        "end": start + len(current),
                        "method": "semantic_merge"
                    })
                current = chunk["content"]
                start = chunk["start"]
        
        # Add last
        if current:
            merged_chunks.append({
                "content": current.strip(),
                "start": start,
                "end": start + len(current),
                "method": "semantic_merge"
            })
        
        return merged_chunks

class ChunkingEvaluator:
    """Evaluates chunking strategies."""
    
    def __init__(self):
        self.results = []
    
    def evaluate_strategy(self, strategy: ChunkingStrategy, text: str, 
                         test_queries: List[str] = None) -> ChunkingMetrics:
        """Evaluate a single chunking strategy."""
        print(f"\n  Evaluating {strategy.name}...")
        
        start_time = time.time()
        
        # Generate chunks
        chunks = strategy.chunk(text)
        
        chunking_time = (time.time() - start_time) * 1000
        
        # Calculate metrics
        metrics = ChunkingMetrics(strategy_name=strategy.name)
        metrics.num_chunks = len(chunks)
        metrics.avg_chunk_size = np.mean([len(c["content"]) for c in chunks]) if chunks else 0
        metrics.latency_ms = chunking_time
        
        # Chunk coherence (simplified - sentence continuity)
        metrics.chunk_coherence = self._calculate_coherence(chunks, text)
        
        # Retrieval recall (simulated)
        if test_queries:
            metrics.retrieval_recall = self._simulate_retrieval(chunks, test_queries)
        
        # Answer faithfulness (simplified)
        metrics.answer_faithfulness = self._calculate_faithfulness(chunks)
        
        # Memory usage (approximate)
        metrics.memory_mb = sum(len(c["content"]) for c in chunks) / (1024 * 1024)
        
        # Overall score
        metrics.overall_score = self._calculate_overall(metrics)
        
        print(f"    Chunks: {metrics.num_chunks}, Coherence: {metrics.chunk_coherence:.3f}, "
              f"Recall: {metrics.retrieval_recall:.3f}, Latency: {metrics.latency_ms:.0f}ms")
        
        return metrics
    
    def _calculate_coherence(self, chunks: List[Dict], full_text: str) -> float:
        """Calculate chunk coherence score."""
        if len(chunks) < 2:
            return 1.0
        
        # Check if chunks maintain sentence boundaries
        coherent_chunks = 0
        for chunk in chunks:
            content = chunk["content"]
            # Check if chunk ends at sentence boundary
            if content.strip().endswith(('.', '!', '?')) or len(content) < 100:
                coherent_chunks += 1
        
        return coherent_chunks / len(chunks)
    
    def _simulate_retrieval(self, chunks: List[Dict], queries: List[str]) -> float:
        """Simulate retrieval recall."""
        # Simplified simulation based on chunk characteristics
        avg_size = np.mean([len(c["content"]) for c in chunks])
        
        # Optimal chunk size for retrieval is 400-600 chars
        if 400 <= avg_size <= 600:
            base_recall = 0.85
        elif 300 <= avg_size < 400 or 600 < avg_size <= 700:
            base_recall = 0.75
        else:
            base_recall = 0.65
        
        # Adjust for number of chunks
        if len(chunks) > 100:
            base_recall -= 0.05  # Too many chunks
        elif len(chunks) < 10:
            base_recall -= 0.10  # Too few chunks
        
        return min(max(base_recall, 0.5), 0.95)
    
    def _calculate_faithfulness(self, chunks: List[Dict]) -> float:
        """Calculate answer faithfulness proxy."""
        if not chunks:
            return 0.0
        
        # Check if chunks have complete sentences
        complete_sentences = 0
        for chunk in chunks:
            content = chunk["content"]
            if '.' in content and content.count('.') >= 2:
                complete_sentences += 1
        
        return complete_sentences / len(chunks)
    
    def _calculate_overall(self, metrics: ChunkingMetrics) -> float:
        """Calculate overall score."""
        return (
            0.25 * metrics.chunk_coherence +
            0.30 * metrics.retrieval_recall +
            0.25 * metrics.answer_faithfulness +
            0.10 * (1.0 - min(metrics.latency_ms / 1000, 1.0)) +
            0.10 * (1.0 - min(metrics.memory_mb / 10, 1.0))
        )

def run_chunking_comparison():
    """Run comprehensive chunking strategy comparison."""
    print("\n" + "="*70)
    print("ADVANCED CHUNKING STRATEGIES COMPARISON")
    print("="*70)
    
    # Load PDF
    pdf_path = Path("science class 8.pdf")
    if not pdf_path.exists():
        print("❌ PDF not found")
        return
    
    doc = fitz.open(pdf_path)
    
    # Extract text from science-specific pages
    full_text = ""
    for page_num in range(30, min(180, len(doc)), 5):
        page = doc[page_num]
        full_text += page.get_text()
    
    doc.close()
    
    print(f"\nTesting with {len(full_text)} characters from science pages")
    
    # Test queries
    test_queries = [
        "What is photosynthesis?",
        "What are the parts of a cell?",
        "What is force?",
        "What is combustion?",
        "What are microorganisms?"
    ]
    
    # Initialize strategies
    strategies = [
        FixedSizeChunking(),
        SemanticChunking(),
        RecursiveChunking(),
        AgenticChunking(),
        HybridChunking()
    ]
    
    # Evaluate all
    evaluator = ChunkingEvaluator()
    all_metrics = []
    
    for strategy in strategies:
        metrics = evaluator.evaluate_strategy(strategy, full_text, test_queries)
        all_metrics.append(metrics)
    
    # Rank by overall score
    ranked = sorted(all_metrics, key=lambda x: x.overall_score, reverse=True)
    
    # Print results
    print("\n" + "="*70)
    print("CHUNKING STRATEGIES RANKING")
    print("="*70)
    print(f"\n{'Rank':<6}{'Strategy':<20}{'Score':<10}{'Recall':<10}{'Coherence':<12}{'Latency':<10}")
    print("-"*70)
    
    for i, m in enumerate(ranked, 1):
        print(f"{i:<6}{m.strategy_name:<20}{m.overall_score:<10.3f}"
              f"{m.retrieval_recall:<10.3f}{m.chunk_coherence:<12.3f}{m.latency_ms:<10.0f}ms")
    
    print("-"*70)
    print(f"\n🏆 Best Strategy: {ranked[0].strategy_name} (Score: {ranked[0].overall_score:.3f})")
    
    # Save results
    import json
    from pathlib import Path
    
    results = {
        "strategies_tested": len(strategies),
        "text_length": len(full_text),
        "rankings": [asdict(m) for m in ranked],
        "best_strategy": ranked[0].strategy_name,
        "best_score": ranked[0].overall_score
    }
    
    output_path = Path("data/chunking_strategies_comparison.json")
    with open(output_path, 'w') as f:
        json.dump(results, f, indent=2)
    
    print(f"\n✅ Results saved to: {output_path}")
    
    return ranked

if __name__ == "__main__":
    from pathlib import Path
    run_chunking_comparison()
