"""Configuration for the Local RAG Service.

Mirrors the optimized parameters from rag_config.py (grid-search results)
and the env contract of .env.local.example. APP_MODE switches between
fully-local inference and cloud-backed providers.
"""

import os
from dataclasses import dataclass, field
from pathlib import Path

SERVICE_ROOT = Path(__file__).resolve().parent.parent


def _env(name: str, default: str) -> str:
    return os.environ.get(name, default)


@dataclass(frozen=True)
class Settings:
    # Mode: "local" (everything in-process) or "production" (cloud services)
    app_mode: str = _env("APP_MODE", "local")

    # Server
    host: str = _env("HOST", "0.0.0.0")
    port: int = int(_env("PORT", "8080"))

    # Retrieval (optimized values from rag_config.py / grid search)
    top_k: int = int(_env("TOP_K", "5"))
    top_k_dense: int = int(_env("TOP_K_DENSE", "100"))
    top_k_sparse: int = int(_env("TOP_K_SPARSE", "100"))
    rrf_k: float = float(_env("RRF_K", "60.0"))
    similarity_threshold: float = float(_env("SIMILARITY_THRESHOLD", "0.36"))
    retrieval_strategy: str = _env("RETRIEVAL_STRATEGY", "hybrid")  # dense|hybrid|bm25

    # Chunking (optimal: 400 chars / 150 overlap)
    chunk_size: int = int(_env("CHILD_MAX_CHARS", "400"))
    chunk_overlap: int = int(_env("CHILD_OVERLAP", "150"))

    # Reranking
    rerank_enabled: bool = _env("RERANK_ENABLED", "true").lower() == "true"
    rerank_candidates: int = int(_env("RERANK_CANDIDATES", "20"))

    # Models (bundled, offline)
    embed_model_path: str = _env(
        "EMBED_MODEL_PATH", str(SERVICE_ROOT / "models" / "minilm-l6-v2-int8.onnx")
    )
    embed_tokenizer_path: str = _env(
        "EMBED_TOKENIZER_PATH", str(SERVICE_ROOT / "models" / "tokenizer.json")
    )
    embed_dimension: int = int(_env("EMBED_DIMENSION", "384"))
    reranker_model_path: str = _env(
        "RERANKER_MODEL_PATH",
        str(
            SERVICE_ROOT.parent
            / "reranker-service"
            / "models"
            / "ms-marco-MiniLM-L-12-v2"
            / "flashrank-MiniLM-L-12-v2_Q.onnx"
        ),
    )
    reranker_tokenizer_path: str = _env(
        "RERANKER_TOKENIZER_PATH",
        str(
            SERVICE_ROOT.parent
            / "reranker-service"
            / "models"
            / "ms-marco-MiniLM-L-12-v2"
            / "tokenizer.json"
        ),
    )

    # Storage
    db_path: str = _env("RAG_DB_PATH", str(SERVICE_ROOT / "data" / "rag_store.sqlite3"))

    # Generation provider: extractive (offline) | openai | gemini | ollama
    llm_provider: str = _env("LLM_PROVIDER", "extractive")
    openai_api_key: str = _env("OPENAI_API_KEY", "")
    openai_model: str = _env("OPENAI_MODEL", "gpt-4o-mini")
    gemini_api_key: str = _env("GEMINI_API_KEY", "")
    gemini_model: str = _env("GEMINI_MODEL", "gemini-2.0-flash")
    ollama_base_url: str = _env("OLLAMA_BASE_URL", "http://localhost:11434")
    ollama_model: str = _env("OLLAMA_LLM_MODEL", "llama3.2:3b")
    llm_temperature: float = float(_env("LLM_TEMPERATURE", "0.75"))
    max_tokens: int = int(_env("MAX_TOKENS", "512"))

    # Cache
    cache_ttl_seconds: int = int(_env("CACHE_TTL_SECONDS", "300"))
    cache_max_entries: int = int(_env("CACHE_MAX_ENTRIES", "10000"))


settings = Settings()
