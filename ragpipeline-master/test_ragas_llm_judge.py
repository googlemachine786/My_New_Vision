"""
Visionary RAG - RAGAS & LLM-as-Judge Evaluation
Implements RAGAS metrics and LLM judge for comprehensive evaluation
"""

import requests
import json
from pathlib import Path
from typing import List, Dict
from dataclasses import dataclass, asdict
import numpy as np
from datetime import datetime

@dataclass
class RAGASEvaluation:
    """RAGAS evaluation metrics."""
    faithfulness: float = 0.0
    answer_relevance: float = 0.0
    context_recall: float = 0.0
    context_precision: float = 0.0
    answer_correctness: float = 0.0
    overall_ragas_score: float = 0.0

@dataclass
class LLMJudgeResult:
    """LLM Judge evaluation result."""
    query: str
    answer: str
    context: str
    judge_score: float = 0.0
    judge_feedback: str = ""
    hallucination_detected: bool = False
    completeness_score: float = 0.0
    clarity_score: float = 0.0

class RAGASEvaluator:
    """Implements RAGAS metrics using LLM."""
    
    def __init__(self, ollama_url: str = "http://localhost:11434", model: str = "llama3.2:latest"):
        self.ollama_url = ollama_url
        self.model = model
    
    def evaluate(self, query: str, answer: str, contexts: List[str], 
                ground_truth: str = None) -> RAGASEvaluation:
        """Evaluate using RAGAS metrics."""
        
        eval_result = RAGASEvaluation()
        
        # 1. Faithfulness
        eval_result.faithfulness = self._calculate_faithfulness(answer, contexts)
        
        # 2. Answer Relevance
        eval_result.answer_relevance = self._calculate_relevance(query, answer)
        
        # 3. Context Recall
        eval_result.context_recall = self._calculate_context_recall(answer, ground_truth) if ground_truth else 0.5
        
        # 4. Context Precision
        eval_result.context_precision = self._calculate_context_precision(contexts, ground_truth) if ground_truth else 0.5
        
        # 5. Answer Correctness
        eval_result.answer_correctness = self._calculate_correctness(answer, ground_truth) if ground_truth else 0.5
        
        # Overall RAGAS score
        eval_result.overall_ragas_score = (
            0.25 * eval_result.faithfulness +
            0.20 * eval_result.answer_relevance +
            0.20 * eval_result.context_recall +
            0.15 * eval_result.context_precision +
            0.20 * eval_result.answer_correctness
        )
        
        return eval_result
    
    def _calculate_faithfulness(self, answer: str, contexts: List[str]) -> float:
        """Calculate faithfulness using LLM."""
        context_text = "\n\n".join(contexts)
        
        prompt = f"""You are a RAG evaluation expert. Evaluate if the answer is faithful to the context.

Context:
{context_text}

Answer:
{answer}

Rate faithfulness from 0.0 to 1.0:
- 1.0: All claims in answer can be inferred from context
- 0.5: Some claims cannot be inferred
- 0.0: Major claims contradict context or cannot be inferred

Respond with ONLY a number between 0.0 and 1.0:"""

        try:
            response = requests.post(
                f"{self.ollama_url}/api/generate",
                json={"model": self.model, "prompt": prompt, "stream": False},
                timeout=30
            )
            
            if response.status_code == 200:
                score_text = response.json().get("response", "0.5").strip()
                # Extract number
                for word in score_text.split():
                    try:
                        score = float(word.replace(',', '.'))
                        if 0.0 <= score <= 1.0:
                            return score
                    except:
                        pass
        except:
            pass
        
        return 0.5  # Default
    
    def _calculate_relevance(self, query: str, answer: str) -> float:
        """Calculate answer relevance using LLM."""
        
        prompt = f"""Rate how relevant the answer is to the query.

Query: {query}
Answer: {answer}

Rate from 0.0 to 1.0:
- 1.0: Directly answers the query completely
- 0.5: Partially answers or includes irrelevant information
- 0.0: Does not answer the query

Respond with ONLY a number between 0.0 and 1.0:"""

        try:
            response = requests.post(
                f"{self.ollama_url}/api/generate",
                json={"model": self.model, "prompt": prompt, "stream": False},
                timeout=30
            )
            
            if response.status_code == 200:
                score_text = response.json().get("response", "0.5").strip()
                for word in score_text.split():
                    try:
                        score = float(word.replace(',', '.'))
                        if 0.0 <= score <= 1.0:
                            return score
                    except:
                        pass
        except:
            pass
        
        return 0.5
    
    def _calculate_context_recall(self, answer: str, ground_truth: str) -> float:
        """Calculate context recall."""
        # Simplified: check overlap between answer and ground truth
        answer_words = set(answer.lower().split())
        truth_words = set(ground_truth.lower().split())
        
        if not truth_words:
            return 0.0
        
        overlap = len(answer_words & truth_words)
        return min(overlap / len(truth_words), 1.0)
    
    def _calculate_context_precision(self, contexts: List[str], ground_truth: str) -> float:
        """Calculate context precision."""
        if not contexts or not ground_truth:
            return 0.0
        
        # Check if ground truth appears in early contexts
        truth_words = set(ground_truth.lower().split())
        
        for i, context in enumerate(contexts[:3]):  # Check first 3 contexts
            context_words = set(context.lower().split())
            overlap = len(truth_words & context_words)
            if overlap > len(truth_words) * 0.3:
                return 1.0 / (i + 1)  # Higher score for earlier appearance
        
        return 0.3
    
    def _calculate_correctness(self, answer: str, ground_truth: str) -> float:
        """Calculate answer correctness."""
        if not answer or not ground_truth:
            return 0.0
        
        # Semantic similarity proxy
        answer_words = set(answer.lower().split())
        truth_words = set(ground_truth.lower().split())
        
        if not answer_words or not truth_words:
            return 0.0
        
        # F1 score
        overlap = len(answer_words & truth_words)
        precision = overlap / len(answer_words)
        recall = overlap / len(truth_words)
        
        if precision + recall == 0:
            return 0.0
        
        return 2 * (precision * recall) / (precision + recall)

class LLMJudge:
    """LLM-as-Judge for comprehensive evaluation."""
    
    def __init__(self, ollama_url: str = "http://localhost:11434", model: str = "llama3.2:latest"):
        self.ollama_url = ollama_url
        self.model = model
    
    def judge(self, query: str, answer: str, context: str) -> LLMJudgeResult:
        """Judge answer quality."""
        
        result = LLMJudgeResult(query=query, answer=answer, context=context)
        
        # Comprehensive judging prompt
        prompt = f"""You are an expert judge evaluating RAG system outputs.

Query: {query}

Context Provided:
{context[:1000]}...

Answer Generated:
{answer}

Evaluate the answer on these criteria:

1. HALLUCINATION CHECK: Does the answer contain information NOT in the context?
   Answer YES or NO

2. COMPLETENESS: Does the answer fully address the query?
   Rate 0.0 to 1.0

3. CLARITY: Is the answer clear and well-structured?
   Rate 0.0 to 1.0

4. OVERALL QUALITY: Considering all factors
   Rate 0.0 to 1.0

Respond in this exact format:
HALLUCINATION: YES or NO
COMPLETENESS: [0.0-1.0]
CLARITY: [0.0-1.0]
OVERALL: [0.0-1.0]
FEEDBACK: [Brief explanation]"""

        try:
            response = requests.post(
                f"{self.ollama_url}/api/generate",
                json={"model": self.model, "prompt": prompt, "stream": False},
                timeout=60
            )
            
            if response.status_code == 200:
                judgment = response.json().get("response", "")
                
                # Parse judgment
                for line in judgment.split('\n'):
                    line = line.strip()
                    if line.startswith("HALLUCINATION:"):
                        result.hallucination_detected = "YES" in line.upper()
                    elif line.startswith("COMPLETENESS:"):
                        try:
                            result.completeness_score = float(line.split(':')[1].strip())
                        except:
                            pass
                    elif line.startswith("CLARITY:"):
                        try:
                            result.clarity_score = float(line.split(':')[1].strip())
                        except:
                            pass
                    elif line.startswith("OVERALL:"):
                        try:
                            result.judge_score = float(line.split(':')[1].strip())
                        except:
                            pass
                    elif line.startswith("FEEDBACK:"):
                        result.judge_feedback = line.split(':', 1)[1].strip()
                
        except Exception as e:
            result.judge_feedback = f"Judge error: {str(e)}"
        
        return result

def run_ragas_and_judge_evaluation():
    """Run comprehensive RAGAS + LLM Judge evaluation."""
    print("\n" + "="*70)
    print("RAGAS & LLM-AS-JUDGE EVALUATION")
    print("="*70)
    
    # Load test results
    results_path = Path("data/quick_improvement_results.json")
    if not results_path.exists():
        print("⚠️  No test results found, using sample data")
        
        # Sample data
        test_cases = [
            {
                "query": "What is photosynthesis?",
                "answer": "Photosynthesis is the process by which plants use sunlight, water, and carbon dioxide to produce glucose and oxygen.",
                "context": "Photosynthesis occurs in chloroplasts. Plants use sunlight to convert CO2 and H2O into glucose. Oxygen is released as a byproduct.",
                "ground_truth": "Photosynthesis is the process by which green plants and some other organisms use sunlight to synthesize foods with the help of chlorophyll."
            },
            {
                "query": "What is force?",
                "answer": "Force is a push or pull acting upon an object.",
                "context": "In physics, a force is any interaction that, when unopposed, will change the motion of an object. Force can be described as a push or pull.",
                "ground_truth": "A force is a push or pull upon an object resulting from the object's interaction with another object."
            }
        ]
    else:
        with open(results_path, 'r') as f:
            all_results = json.load(f)
        
        # Convert to test cases
        test_cases = []
        if isinstance(all_results, list):
            for r in all_results[:5]:  # Sample 5
                test_cases.append({
                    "query": r.get("query", ""),
                    "answer": r.get("answer", ""),
                    "context": r.get("context", "Sample context"),
                    "ground_truth": ""
                })
    
    print(f"\nEvaluating {len(test_cases)} test cases...")
    
    # Initialize evaluators
    ragas_eval = RAGASEvaluator()
    llm_judge = LLMJudge()
    
    all_ragas_results = []
    all_judge_results = []
    
    for i, case in enumerate(test_cases, 1):
        print(f"\n[{i}/{len(test_cases)}] Evaluating: {case['query'][:50]}...")
        
        # RAGAS evaluation
        ragas_result = ragas_eval.evaluate(
            query=case["query"],
            answer=case["answer"],
            contexts=[case["context"]],
            ground_truth=case.get("ground_truth")
        )
        all_ragas_results.append(asdict(ragas_result))
        
        print(f"  RAGAS Score: {ragas_result.overall_ragas_score:.3f}")
        print(f"    Faithfulness: {ragas_result.faithfulness:.3f}")
        print(f"    Relevance: {ragas_result.answer_relevance:.3f}")
        
        # LLM Judge evaluation
        judge_result = llm_judge.judge(
            query=case["query"],
            answer=case["answer"],
            context=case["context"]
        )
        all_judge_results.append(asdict(judge_result))
        
        print(f"  Judge Score: {judge_result.judge_score:.3f}")
        print(f"    Hallucination: {judge_result.hallucination_detected}")
        print(f"    Completeness: {judge_result.completeness_score:.3f}")
        print(f"    Feedback: {judge_result.judge_feedback[:80]}...")
    
    # Calculate aggregate metrics
    if all_ragas_results:
        avg_ragas = np.mean([r["overall_ragas_score"] for r in all_ragas_results])
        avg_faithfulness = np.mean([r["faithfulness"] for r in all_ragas_results])
        avg_relevance = np.mean([r["answer_relevance"] for r in all_ragas_results])
        
        print("\n" + "="*70)
        print("RAGAS AGGREGATE METRICS")
        print("="*70)
        print(f"  Average RAGAS Score: {avg_ragas:.3f}")
        print(f"  Average Faithfulness: {avg_faithfulness:.3f}")
        print(f"  Average Relevance: {avg_relevance:.3f}")
    
    if all_judge_results:
        avg_judge = np.mean([r["judge_score"] for r in all_judge_results])
        hallucination_rate = sum(1 for r in all_judge_results if r["hallucination_detected"]) / len(all_judge_results)
        
        print("\n" + "="*70)
        print("LLM JUDGE AGGREGATE METRICS")
        print("="*70)
        print(f"  Average Judge Score: {avg_judge:.3f}")
        print(f"  Hallucination Rate: {hallucination_rate:.1%}")
        print(f"  Average Completeness: {np.mean([r['completeness_score'] for r in all_judge_results]):.3f}")
    
    # Save results
    results = {
        "evaluation_date": datetime.now().isoformat(),
        "test_cases": len(test_cases),
        "ragas_metrics": all_ragas_results,
        "llm_judge_metrics": all_judge_results,
        "aggregate": {
            "avg_ragas_score": avg_ragas if all_ragas_results else 0,
            "avg_judge_score": avg_judge if all_judge_results else 0,
            "hallucination_rate": hallucination_rate if all_judge_results else 0
        }
    }
    
    output_path = Path("data/ragas_judge_evaluation.json")
    with open(output_path, 'w') as f:
        json.dump(results, f, indent=2)
    
    print(f"\n✅ Results saved to: {output_path}")
    
    return results

if __name__ == "__main__":
    run_ragas_and_judge_evaluation()
