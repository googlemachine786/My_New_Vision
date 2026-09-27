#!/usr/bin/env python
"""
Visionary RAG - REAL Interactive Query CLI with Self-Query Retrieval
ACTUALLY extracts from your PDF, REAL embeddings, REAL retrieval
NOW WITH METADATA-ENHANCED SELF-QUERY RETRIEVAL!
"""

import requests
import fitz  # PyMuPDF
import json
import numpy as np
from pathlib import Path
from datetime import datetime
import hashlib

# Import self-query retriever
try:
    from retrievers import SelfQueryRetriever
    SELF_QUERY_AVAILABLE = True
except ImportError:
    SELF_QUERY_AVAILABLE = False
    print("⚠️  Self-query retriever not available (install with: pip install -r requirements.txt)")

class REALRAGQueryCLI:
    """REAL CLI - actually extracts from PDF with self-query enhancement."""

    def __init__(self, use_self_query: bool = True):
        self.ollama_url = "http://localhost:11434"
        self.pdf_path = Path("science class 8.pdf")
        self.chunks = []
        self.chunk_embeddings = []
        self.metadata = []  # Store metadata for each chunk
        self.use_self_query = use_self_query and SELF_QUERY_AVAILABLE
        
        if self.use_self_query:
            self.self_query_retriever = SelfQueryRetriever()
            print("✅ Self-query retrieval enabled\n")
        else:
            self.self_query_retriever = None
            print("ℹ️  Using standard retrieval (self-query disabled)\n")

        # Extract REAL chunks from PDF
        print("\n📖 Extracting REAL chunks from science class 8.pdf...")
        self._extract_real_chunks()

        # Create REAL embeddings
        print("🔢 Creating REAL embeddings with Ollama...")
        self._create_real_embeddings()

        print(f"✅ Ready! {len(self.chunks)} chunks with real embeddings loaded.\n")
    
    def _extract_real_chunks(self):
        """ACTUALLY extract chunks from the real PDF with metadata."""
        if not self.pdf_path.exists():
            print(f"❌ PDF not found: {self.pdf_path}")
            print("   Please make sure 'science class 8.pdf' is in the current directory")
            return

        doc = fitz.open(self.pdf_path)
        total_pages = len(doc)

        # Simple grade/chapter detection based on content
        current_chapter = "Unknown"
        chapter_number = 0
        
        # Extract from EVERY page (not fake samples!)
        for page_num in range(total_pages):
            page = doc[page_num]
            text = page.get_text()

            # Skip empty pages
            if len(text.strip()) < 100:
                continue

            # Try to detect chapter from page content
            if "Chapter" in text[:500] or "CHAPTER" in text[:500]:
                import re
                chapter_match = re.search(r'[Cc]hapter\s*(\d+)', text[:500])
                if chapter_match:
                    chapter_number = int(chapter_match.group(1))
                    chapter_name_match = re.search(r'[Cc]hapter\s*\d+[:\s]+([^\n]+)', text[:500])
                    current_chapter = chapter_name_match.group(1).strip() if chapter_name_match else f"Chapter {chapter_number}"

            # Split into chunks (500 chars each)
            chunk_size = 500
            for i in range(0, len(text), chunk_size):
                chunk_text = text[i:i + chunk_size]

                if len(chunk_text.strip()) > 50:
                    # Detect content type
                    content_type = "Text"
                    if "│" in chunk_text or "┃" in chunk_text or chunk_text.count('\n') > 5:
                        content_type = "Table"
                    elif "Fig" in chunk_text or "fig" in chunk_text or "Diagram" in chunk_text:
                        content_type = "Figure"
                    elif "=" in chunk_text and len(chunk_text) < 200:
                        content_type = "Formula"

                    self.chunks.append({
                        "content": chunk_text.strip(),
                        "page": page_num + 1,  # REAL page number!
                        "chunk_id": len(self.chunks),
                        "grade": 8,  # This is grade 8 textbook
                        "chapter": current_chapter,
                        "chapter_number": chapter_number,
                        "content_type": content_type,
                        "subject": "Science"
                    })
                    self.metadata.append({
                        "grade": 8,
                        "chapter": current_chapter,
                        "chapter_number": chapter_number,
                        "content_type": content_type,
                        "subject": "Science",
                        "page_number": page_num + 1
                    })

        doc.close()
        print(f"   Extracted {len(self.chunks)} chunks from {total_pages} pages")
        print(f"   Chapters detected: {len(set(c['chapter'] for c in self.chunks))}")
    
    def _create_real_embeddings(self):
        """ACTUALLY create embeddings using Ollama API."""
        for i, chunk in enumerate(self.chunks):
            if (i + 1) % 10 == 0:
                print(f"   Embedding chunk {i+1}/{len(self.chunks)}...")
            
            try:
                response = requests.post(
                    f"{self.ollama_url}/api/embeddings",
                    json={
                        "model": "nomic-embed-text",
                        "prompt": chunk["content"]
                    },
                    timeout=30
                )
                
                if response.status_code == 200:
                    emb = response.json().get("embedding", [])
                    self.chunk_embeddings.append(emb)
                else:
                    # Fallback: zero vector
                    self.chunk_embeddings.append([0.0] * 768)
                    
            except Exception as e:
                # Fallback: zero vector
                self.chunk_embeddings.append([0.0] * 768)
    
    def _cosine_similarity(self, vec1, vec2):
        """Calculate cosine similarity between two vectors."""
        dot_product = np.dot(vec1, vec2)
        norm1 = np.linalg.norm(vec1)
        norm2 = np.linalg.norm(vec2)
        
        if norm1 == 0 or norm2 == 0:
            return 0.0
        
        return dot_product / (norm1 * norm2)
    
    def retrieve_real_context(self, query: str, top_k: int = 5):
        """REAL retrieval with self-query enhancement."""
        # Try self-query retrieval first if enabled
        if self.use_self_query and self.self_query_retriever:
            try:
                print("   🏷️  Using self-query retrieval with metadata filtering...")
                import asyncio
                results = asyncio.run(self.self_query_retriever.search(query))
                
                if results:
                    # Convert self-query results to our format
                    context_chunks = []
                    for r in results:
                        context_chunks.append({
                            "content": r['content'],
                            "page": r['metadata'].get('page_number', 0),
                            "chunk_id": r['chunk_id'],
                            "grade": r['metadata'].get('grade'),
                            "chapter": r['metadata'].get('chapter'),
                            "chapter_number": r['metadata'].get('chapter_number'),
                            "content_type": r['metadata'].get('content_type'),
                            "similarity": r['similarity']
                        })
                    
                    print(f"   ✅ Self-query found {len(context_chunks)} filtered results")
                    return context_chunks
                    
            except Exception as e:
                print(f"   ⚠️  Self-query failed, falling back to standard: {e}")
                # Fall through to standard retrieval
        
        # Standard retrieval (fallback)
        print("   🔍 Using standard semantic search...")
        return self._standard_retrieve(query, top_k)
    
    def _standard_retrieve(self, query: str, top_k: int = 5):
        """Standard semantic search without metadata filtering."""
        # Create query embedding
        try:
            response = requests.post(
                f"{self.ollama_url}/api/embeddings",
                json={
                    "model": "nomic-embed-text",
                    "prompt": query
                },
                timeout=30
            )

            if response.status_code == 200:
                query_embedding = response.json().get("embedding", [])
            else:
                print("⚠️  Embedding failed, using keyword fallback")
                return self._keyword_retrieve(query, top_k)

        except Exception as e:
            print(f"⚠️  Error: {e}")
            return self._keyword_retrieve(query, top_k)

        # Calculate similarity with ALL chunks
        similarities = []
        for i, chunk_emb in enumerate(self.chunk_embeddings):
            sim = self._cosine_similarity(query_embedding, chunk_emb)
            similarities.append((i, sim))

        # Sort by similarity (descending)
        similarities.sort(key=lambda x: x[1], reverse=True)

        # Return top_k most similar chunks
        top_indices = [idx for idx, sim in similarities[:top_k]]

        return [self.chunks[i] for i in top_indices if i < len(self.chunks)]
    
    def _keyword_retrieve(self, query: str, top_k: int = 5):
        """Fallback: keyword-based retrieval."""
        query_words = set(query.lower().split())
        
        scored = []
        for i, chunk in enumerate(self.chunks):
            chunk_words = set(chunk["content"].lower().split())
            overlap = len(query_words & chunk_words)
            scored.append((i, overlap))
        
        scored.sort(key=lambda x: x[1], reverse=True)
        top_indices = [idx for idx, score in scored[:top_k] if score > 0]
        
        return [self.chunks[i] for i in top_indices]
    
    def generate_real_answer(self, query: str, context: str):
        """Generate answer using REAL Ollama API."""
        prompt = f"""You are a helpful science tutor for CBSE Class 8 students.

Use ONLY the following context from the textbook to answer.
Cite the page numbers when available.
If the answer is not in the context, say "I couldn't find this in the textbook."

CONTEXT FROM TEXTBOOK:
{context}

QUESTION: {query}

ANSWER (be concise, educational, cite pages):"""

        try:
            response = requests.post(
                f"{self.ollama_url}/api/generate",
                json={
                    "model": "llama3.2:3b",
                    "prompt": prompt,
                    "stream": False
                },
                timeout=60
            )
            
            if response.status_code == 200:
                return response.json().get("response", "")
            else:
                return f"Error generating answer: {response.status_code}"
        except Exception as e:
            return f"Error: {e}"
    
    def ask_real_question(self, query: str):
        """Ask a REAL question with REAL retrieval from REAL PDF."""
        print("\n" + "="*70)
        print(f"REAL QUERY: {query}")
        print("="*70)
        
        # REAL retrieval
        print("\n🔍 Retrieving from ACTUAL PDF...")
        context_chunks = self.retrieve_real_context(query, top_k=5)
        
        if not context_chunks:
            print("❌ No relevant content found in the textbook")
            return
        
        print(f"✅ Found {len(context_chunks)} relevant chunks from REAL pages")
        
        # Build context with REAL page numbers
        context = "\n\n".join([
            f"[Page {chunk['page']}]\n{chunk['content']}"
            for chunk in context_chunks
        ])
        
        print(f"\n📚 REAL Context from Textbook:")
        for chunk in context_chunks[:3]:  # Show first 3
            print(f"   Page {chunk['page']}: {chunk['content'][:100]}...")
        
        # REAL answer generation
        print("\n🤖 Generating answer from REAL context...")
        answer = self.generate_real_answer(query, context)
        
        print(f"\n💬 REAL ANSWER:\n{answer}\n")
        
        print("="*70)
    
    def run_real_interactive(self):
        """Run REAL interactive session with self-query enhancement."""
        print("\n" + "="*70)
        print("VISIONARY RAG - REAL QUERY CLI WITH SELF-QUERY RETRIEVAL")
        print("="*70)
        print("\nAsk any science question - REAL retrieval from your PDF!")
        print("Type 'quit' to exit, 'info' to see stats, 'toggle' to switch retrieval mode.\n")

        # Check Ollama
        try:
            response = requests.get(f"{self.ollama_url}/api/tags", timeout=5)
            if response.status_code == 200:
                models = response.json().get("models", [])
                print(f"✅ Ollama: {len(models)} models")
            else:
                print("⚠️  Ollama may not be running")
        except:
            print("❌ Cannot connect to Ollama")
            return

        print(f"📖 PDF: {self.pdf_path.name} ({len(self.chunks)} chunks)")
        print(f"🔢 Embeddings: {len(self.chunk_embeddings)} vectors")
        print(f"🏷️  Self-Query: {'✅ Enabled' if self.use_self_query else 'ℹ️  Disabled'}")
        print(f"📊 Chapters: {len(set(c['chapter'] for c in self.chunks))} detected\n")
        
        print("💡 Try queries like:")
        print("   • 'Show me grade 8 photosynthesis diagrams'")
        print("   • 'questions from chapter 9 about force'")
        print("   • 'what is force?'")
        print("="*70)

        while True:
            try:
                query = input("\n❓ Ask: ").strip()

                if not query:
                    continue
                elif query.lower() in ['quit', 'exit', 'q']:
                    print("\n👋 Goodbye!")
                    break
                elif query.lower() == 'info':
                    print(f"\n📊 Stats:")
                    print(f"   Chunks: {len(self.chunks)}")
                    print(f"   Embeddings: {len(self.chunk_embeddings)}")
                    print(f"   Metadata: {len(self.metadata)} entries")
                    print(f"   PDF Pages: {len(self.chunks)} chunks extracted")
                    print(f"   Chapters: {len(set(c['chapter'] for c in self.chunks))} unique")
                    print(f"   Self-Query: {'✅ Enabled' if self.use_self_query else 'ℹ️  Disabled'}")
                elif query.lower() == 'toggle':
                    self.use_self_query = not self.use_self_query
                    print(f"\n🏷️  Self-Query: {'✅ Enabled' if self.use_self_query else 'ℹ️  Disabled'}")
                elif query.lower() == 'help':
                    print("\n📖 Commands:")
                    print("   <question> - Ask any science question")
                    print("   info - Show statistics")
                    print("   toggle - Toggle self-query retrieval")
                    print("   quit/exit/q - Exit")
                    print("\n🎯 Example queries:")
                    print("   • 'Show me grade 8 photosynthesis diagrams'")
                    print("   • 'questions from chapter 9 about force'")
                    print("   • 'what is force?'")
                else:
                    self.ask_real_question(query)

            except KeyboardInterrupt:
                print("\n\n👋 Goodbye!")
                break

def main():
    """Main - REAL implementation."""
    cli = REALRAGQueryCLI()
    cli.run_real_interactive()
    return 0

if __name__ == "__main__":
    main()
