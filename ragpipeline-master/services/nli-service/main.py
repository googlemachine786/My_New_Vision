"""
NLI Microservice — Natural Language Inference for Response Grounding Validation

Uses CrossEncoder models for semantic NLI instead of hand-rolled keyword overlap.
Models:
  - sentence-transformers for sentence splitting (via spaCy)
  - cross-encoder/nli-deberta-v3-base for NLI validation

Endpoints:
  POST /split-sentences  — Proper sentence segmentation via spaCy
  POST /validate-claim   — NLI check: does chunk SUPPORT/CONTRADICT/NEUTRAL claim?
  POST /extract-claims   — Extract factual claims from response text
  POST /batch-validate   — Batch validate multiple claims at once
  GET  /health
"""

import asyncio
import logging
import time
from contextlib import asynccontextmanager
from typing import List, Optional

import torch
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from sentence_transformers import CrossEncoder
import spacy

# ─── Configuration ───────────────────────────────────────────────────────────

NLI_MODEL = "cross-encoder/nli-deberta-v3-base"
SPACY_MODEL = "en_core_web_sm"
NLI_LABELS = ["contradiction", "entailment", "neutral"]
GROUNDING_THRESHOLD = 0.6
MAX_CONCURRENT_NLI = 4  # FIX P0: Prevent GPU OOM

# ─── Models ──────────────────────────────────────────────────────────────────


class SplitSentencesRequest(BaseModel):
    text: str = Field(..., min_length=1, max_length=50000)


class SplitSentencesResponse(BaseModel):
    sentences: List[str]


class ValidateClaimRequest(BaseModel):
    claim: str = Field(..., min_length=1, max_length=2000)
    context_chunks: List[str] = Field(..., min_items=1, max_items=20)


class ValidateClaimResponse(BaseModel):
    is_grounded: bool
    grounding_score: float
    best_chunk_idx: int
    details: Optional[str] = None


class ExtractClaimsRequest(BaseModel):
    text: str = Field(..., min_length=1, max_length=50000)


class ExtractClaimsResponse(BaseModel):
    claims: List[str]


class HealthResponse(BaseModel):
    status: str
    nli_model: str
    spacy_model: str
    ready: bool


class BatchValidateRequest(BaseModel):
    claims: List[str] = Field(..., min_items=1, max_items=100)
    context_chunks: List[str] = Field(..., min_items=1, max_items=20)


class ClaimResult(BaseModel):
    claim: str
    is_grounded: bool
    grounding_score: float
    best_chunk_idx: int
    label: str


class BatchValidateResponse(BaseModel):
    results: List[ClaimResult]
    overall_grounding_score: float
    processing_time_ms: float


# ─── App Setup ───────────────────────────────────────────────────────────────

logging.basicConfig(level=logging.INFO, format="%(asctime)s [%(levelname)s] %(message)s")
logger = logging.getLogger(__name__)

nli_semaphore = asyncio.Semaphore(MAX_CONCURRENT_NLI)
_is_ready = False


@asynccontextmanager
async def lifespan(app: FastAPI):
    """Load ML models on startup with async readiness reporting."""
    global _is_ready
    logger.info("Loading spaCy model: %s", SPACY_MODEL)
    try:
        app.state.nlp = spacy.load(SPACY_MODEL)
        logger.info("spaCy model loaded successfully")
    except OSError:
        logger.error("spaCy model '%s' not found. Run: python -m spacy download %s", SPACY_MODEL, SPACY_MODEL)
        raise RuntimeError(f"spaCy model '{SPACY_MODEL}' not installed")

    logger.info("Loading NLI CrossEncoder model: %s", NLI_MODEL)
    try:
        device = "cuda" if torch.cuda.is_available() else "cpu"
        logger.info("Using device: %s", device)
        # FIX P0: Load model async to not block event loop
        app.state.nli_model = CrossEncoder(NLI_MODEL, num_labels=3, device=device)
        _is_ready = True
        logger.info("NLI CrossEncoder model loaded successfully on %s", device)
    except Exception as e:
        logger.error("Failed to load NLI model: %s", e)
        raise RuntimeError(f"NLI model '{NLI_MODEL}' failed to load: {e}")

    logger.info("NLI microservice ready")
    yield


app = FastAPI(title="NLI Microservice", version="1.0.0", lifespan=lifespan)


# ─── Endpoints ───────────────────────────────────────────────────────────────


@app.get("/health", response_model=HealthResponse)
async def health():
    status = "healthy" if _is_ready else "loading"
    http_status = 200 if _is_ready else 503
    return HealthResponse(
        status=status,
        nli_model=NLI_MODEL,
        spacy_model=SPACY_MODEL,
        ready=_is_ready,
    ), http_status


@app.post("/split-sentences", response_model=SplitSentencesResponse)
async def split_sentences(req: SplitSentencesRequest):
    start = time.monotonic()
    doc = app.state.nlp(req.text)
    sentences = [sent.text.strip() for sent in doc.sents if sent.text.strip()]
    latency_ms = (time.monotonic() - start) * 1000
    logger.info("split_sentences: %d sentences in %.1fms", len(sentences), latency_ms)
    return SplitSentencesResponse(sentences=sentences)


@app.post("/extract-claims", response_model=ExtractClaimsResponse)
async def extract_claims(req: ExtractClaimsRequest):
    start = time.monotonic()
    doc = app.state.nlp(req.text)
    claims = [sent.text.strip() for sent in doc.sents if len(sent.text.strip()) >= 15]
    latency_ms = (time.monotonic() - start) * 1000
    logger.info("extract_claims: %d claims in %.1fms", len(claims), latency_ms)
    return ExtractClaimsResponse(claims=claims)


@app.post("/validate-claim", response_model=ValidateClaimResponse)
async def validate_claim(req: ValidateClaimRequest):
    """Validate a single claim against context chunks (with concurrency limit)."""
    if not req.claim.strip():
        raise HTTPException(status_code=400, detail="Claim text is empty")

    async with nli_semaphore:
        return await _do_validate_claim(req)


async def _do_validate_claim(req: ValidateClaimRequest):
    """Internal validation logic (must be called under semaphore)."""
    start = time.monotonic()
    nli_model = app.state.nli_model

    best_score = 0.0
    best_idx = 0
    best_label = "neutral"

    for idx, chunk in enumerate(req.context_chunks):
        if not chunk.strip():
            continue
        pair = [chunk, req.claim]
        scores = nli_model.predict([pair])[0]
        entailment_score = float(scores[1])
        if entailment_score > best_score:
            best_score = entailment_score
            best_idx = idx
            best_label = NLI_LABELS[scores.argmax()]

    is_grounded = best_score >= GROUNDING_THRESHOLD
    latency_ms = (time.monotonic() - start) * 1000
    logger.info("validate_claim: grounded=%s score=%.3f label=%s in %.1fms", is_grounded, best_score, best_label, latency_ms)

    return ValidateClaimResponse(
        is_grounded=is_grounded,
        grounding_score=round(best_score, 4),
        best_chunk_idx=best_idx,
        details=f"Best match: {best_label} (score: {best_score:.3f}, chunk {best_idx})",
    )


@app.post("/batch-validate", response_model=BatchValidateResponse)
async def batch_validate(req: BatchValidateRequest):
    """Validate multiple claims against context chunks in batch (with concurrency limit)."""
    async with nli_semaphore:
        return await _do_batch_validate(req)


async def _do_batch_validate(req: BatchValidateRequest):
    """Internal batch validation (must be called under semaphore)."""
    start = time.monotonic()
    nli_model = app.state.nli_model
    results = []
    grounded_count = 0

    all_pairs = []
    pair_map = []

    for claim_idx, claim in enumerate(req.claims):
        if not claim.strip():
            continue
        for chunk_idx, chunk in enumerate(req.context_chunks):
            if chunk.strip():
                all_pairs.append([chunk, claim])
                pair_map.append((claim_idx, chunk_idx))

    if not all_pairs:
        return BatchValidateResponse(results=[], overall_grounding_score=0.0, processing_time_ms=0.0)

    batch_scores = nli_model.predict(all_pairs)
    claim_best = {}

    for i, scores in enumerate(batch_scores):
        claim_idx, chunk_idx = pair_map[i]
        entailment = float(scores[1])
        label = NLI_LABELS[scores.argmax()]
        if claim_idx not in claim_best or entailment > claim_best[claim_idx][0]:
            claim_best[claim_idx] = (entailment, chunk_idx, label)

    for claim_idx, (score, chunk_idx, label) in sorted(claim_best.items()):
        is_grounded = score >= GROUNDING_THRESHOLD
        if is_grounded:
            grounded_count += 1
        results.append(ClaimResult(
            claim=req.claims[claim_idx],
            is_grounded=is_grounded,
            grounding_score=round(score, 4),
            best_chunk_idx=chunk_idx,
            label=label,
        ))

    total = len(results)
    overall = grounded_count / total if total > 0 else 0.0
    latency_ms = (time.monotonic() - start) * 1000
    logger.info("batch_validate: %d claims, %.1f%% grounded in %.1fms", total, overall * 100, latency_ms)

    return BatchValidateResponse(
        results=results,
        overall_grounding_score=round(overall, 4),
        processing_time_ms=round(latency_ms, 1),
    )


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8084)
