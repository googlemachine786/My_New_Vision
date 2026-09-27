"""Answer generation with switchable providers.

- extractive (default, fully offline): composes a grounded answer from the
  top retrieved parent sections — no hallucination possible by construction.
- openai / gemini / ollama: LLM generation grounded in retrieved context,
  enabled by setting LLM_PROVIDER + credentials (cloud-ready path).
"""

from __future__ import annotations

import re

import httpx

from .config import settings
from .retrieval import RetrievedChunk
from .store import VectorStore

_PROMPT = """You are a helpful CBSE Science tutor for grades 6-8.
Answer the student's question using ONLY the context below.
Cite facts from the context; if the context is insufficient, say so.

Context:
{context}

Question: {question}

Answer (simple language appropriate for the student's grade):"""


def _build_context(chunks: list[RetrievedChunk], store: VectorStore) -> str:
    seen: set[str] = set()
    parts: list[str] = []
    for ch in chunks:
        if ch.parent_id in seen:
            continue
        seen.add(ch.parent_id)
        parent = store.parent(ch.parent_id)
        text = parent["content"] if parent else ch.content
        title = parent["title"] if parent else ch.metadata.get("section", "")
        chapter = ch.metadata.get("chapter", "")
        parts.append(f"[{chapter} — {title}]\n{text}")
    return "\n\n".join(parts)


class Answerer:
    def __init__(self, store: VectorStore):
        self.store = store
        self.provider = settings.llm_provider

    def answer(self, question: str, chunks: list[RetrievedChunk]) -> tuple[str, int]:
        """Return (answer, total_tokens)."""
        if not chunks:
            return (
                "I couldn't find relevant material in the textbook for that "
                "question. Could you rephrase it, or ask about a topic from "
                "your Science syllabus?",
                0,
            )
        if self.provider == "openai":
            return self._openai(question, chunks)
        if self.provider == "gemini":
            return self._gemini(question, chunks)
        if self.provider == "ollama":
            return self._ollama(question, chunks)
        return self._extractive(question, chunks)

    # ------------------------------------------------------------ extractive

    def _extractive(
        self, question: str, chunks: list[RetrievedChunk]
    ) -> tuple[str, int]:
        """Grounded extractive answer: lead with the most relevant sentences,
        then add supporting section context."""
        top = chunks[0]
        parent = self.store.parent(top.parent_id)
        section = parent["title"] if parent else top.metadata.get("section", "")
        chapter = top.metadata.get("chapter", "")

        # Pick the sentences from the top chunks most lexically tied to the query
        q_terms = set(re.findall(r"[a-z0-9]+", question.lower())) - {
            "what", "is", "the", "a", "an", "of", "in", "and", "how", "do",
            "does", "why", "are", "to", "between", "difference",
        }
        best: list[tuple[float, str]] = []
        for ch in chunks[:3]:
            for sent in re.split(r"(?<=[.!?])\s+", ch.content):
                terms = set(re.findall(r"[a-z0-9]+", sent.lower()))
                overlap = len(q_terms & terms)
                if overlap:
                    best.append((overlap + 0.01 * len(sent) / 100, sent.strip()))
        best.sort(key=lambda t: -t[0])
        lead_sents: list[str] = []
        for _, s in best:
            # Skip fragments (chunk-overlap artifacts) and near-duplicates
            if s and not s[0].isupper():
                continue
            if any(s in prev or prev in s for prev in lead_sents):
                continue
            lead_sents.append(s)
            if len(lead_sents) >= 3:
                break
        lead = " ".join(lead_sents) if lead_sents else chunks[0].content

        answer = lead
        if chapter or section:
            answer += f"\n\n(From: {chapter} — {section})"
        return answer, 0

    # --------------------------------------------------------------- openai

    def _openai(self, question: str, chunks: list[RetrievedChunk]) -> tuple[str, int]:
        prompt = _PROMPT.format(
            context=_build_context(chunks, self.store), question=question
        )
        resp = httpx.post(
            "https://api.openai.com/v1/chat/completions",
            headers={"Authorization": f"Bearer {settings.openai_api_key}"},
            json={
                "model": settings.openai_model,
                "messages": [{"role": "user", "content": prompt}],
                "temperature": settings.llm_temperature,
                "max_tokens": settings.max_tokens,
            },
            timeout=60,
        )
        resp.raise_for_status()
        data = resp.json()
        return (
            data["choices"][0]["message"]["content"],
            data.get("usage", {}).get("total_tokens", 0),
        )

    # --------------------------------------------------------------- gemini

    def _gemini(self, question: str, chunks: list[RetrievedChunk]) -> tuple[str, int]:
        prompt = _PROMPT.format(
            context=_build_context(chunks, self.store), question=question
        )
        resp = httpx.post(
            f"https://generativelanguage.googleapis.com/v1beta/models/"
            f"{settings.gemini_model}:generateContent",
            params={"key": settings.gemini_api_key},
            json={
                "contents": [{"parts": [{"text": prompt}]}],
                "generationConfig": {
                    "temperature": settings.llm_temperature,
                    "maxOutputTokens": settings.max_tokens,
                },
            },
            timeout=60,
        )
        resp.raise_for_status()
        data = resp.json()
        text = data["candidates"][0]["content"]["parts"][0]["text"]
        tokens = data.get("usageMetadata", {}).get("totalTokenCount", 0)
        return text, tokens

    # --------------------------------------------------------------- ollama

    def _ollama(self, question: str, chunks: list[RetrievedChunk]) -> tuple[str, int]:
        prompt = _PROMPT.format(
            context=_build_context(chunks, self.store), question=question
        )
        resp = httpx.post(
            f"{settings.ollama_base_url}/api/generate",
            json={
                "model": settings.ollama_model,
                "prompt": prompt,
                "stream": False,
                "options": {
                    "temperature": settings.llm_temperature,
                    "num_predict": settings.max_tokens,
                },
            },
            timeout=120,
        )
        resp.raise_for_status()
        data = resp.json()
        return data.get("response", ""), data.get("eval_count", 0)
