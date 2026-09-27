# Query Understanding Service

A production-ready Go microservice for the RAG pipeline that handles query rewriting, self-query parsing, and intent classification using Google Gemini.

## Overview

This service extracts query understanding logic from the monolith. It provides three core capabilities:

1. **Query Rewriting** - Rewrites queries with conversation history context for standalone retrieval
2. **Self-Query Parsing** - Extracts metadata filters (grade, subject, chapter, topic, content type) from natural language
3. **Intent Classification** - Classifies query intent into categories (information-seeking, problem-solving, content-retrieval, etc.)

## Quick Start

### Prerequisites

- Go 1.23+
- Google Cloud project with Vertex AI API enabled
- Service account with `roles/aiplatform.user` role

### Setup

```bash
cd services/query-understanding-service

# Copy and configure environment
cp .env.example .env
# Edit .env with your GCP project ID and credentials

# Install dependencies
go mod tidy

# Run the service
go run .
```

The server starts on port **8083** by default.

## API Endpoints

### POST /rewrite

Rewrite a query using conversation history context.

**Request:**
```json
{
  "query": "explain this further",
  "history": [
    {"role": "user", "content": "What is quantum physics?"},
    {"role": "assistant", "content": "Quantum physics is the study of matter and energy at the smallest scales."}
  ]
}
```

**Response:**
```json
{
  "data": {
    "rewritten_query": "Can you explain quantum physics concepts in more detail?",
    "confidence": 0.92,
    "changes_made": ["resolved 'this' to 'quantum physics'", "converted to detailed explanation request"]
  }
}
```

### POST /parse-filters

Extract metadata filters from a query (self-query).

**Request:**
```json
{
  "query": "Show me class 10 physics chapter 3 exercises about optics"
}
```

**Response:**
```json
{
  "data": {
    "filters": {
      "grade": ["10"],
      "subject": ["physics"],
      "chapter": ["3"],
      "topic": ["optics"],
      "content_type": ["exercises"],
      "board": [],
      "language": []
    },
    "confidence": 0.88,
    "explanation": "Extracted grade 10, physics subject, chapter 3, optics topic, and exercises content type"
  }
}
```

### POST /classify-intent

Classify the intent of a user query.

**Request:**
```json
{
  "query": "Solve the equation x^2 + 5x + 6 = 0"
}
```

**Response:**
```json
{
  "data": {
    "intent": "problem_solving",
    "confidence": 0.95,
    "secondary_intent": null,
    "reasoning": "Query asks to solve a mathematical equation, which is a problem-solving task"
  }
}
```

### GET /health

Health check endpoint.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2026-04-03T12:00:00Z",
  "version": "v0.1.0"
}
```

### GET /ready

Readiness check for Kubernetes deployments.

**Response:**
```json
{
  "ready": true,
  "services": "all"
}
```

## Architecture

```
query-understanding-service/
├── main.go                          # Entry point, server setup
├── config/
│   └── config.go                    # Environment-based configuration
├── llm/
│   └── gemini_client.go             # Google Gemini API wrapper
├── prompts/
│   └── templates.go                 # Versioned prompt templates
├── rewriter/
│   └── query_rewriter.go            # Query rewriting with history
├── parser/
│   └── self_query.go                # Metadata filter extraction
├── classifier/
│   └── intent.go                    # Intent classification
└── handler/
    └── query_handler.go             # HTTP request handlers
```

## Configuration

All configuration is via environment variables. See `.env.example` for all options.

| Variable | Default | Description |
|----------|---------|-------------|
| `GEMINI_PROJECT_ID` | (required) | GCP project ID |
| `GEMINI_LOCATION` | `us-central1` | Vertex AI region |
| `GEMINI_MODEL` | `gemini-2.0-flash` | Model name |
| `GEMINI_TEMPERATURE` | `0.1` | Generation temperature |
| `PORT` | `8083` | HTTP server port |
| `MAX_HISTORY_TURNS` | `10` | Max conversation history |
| `PROMPT_REWRITE_VERSION` | `v1` | Prompt version for rewriting |

## Fallback Behavior

When the Gemini API is unavailable or returns an error, each service falls back gracefully:

- **Rewrite**: Returns the original query unchanged with confidence 0.5
- **Parse Filters**: Returns empty filters with confidence 0.3
- **Classify Intent**: Defaults to `information_seeking` with confidence 0.3

## Prompt Versioning

Prompt templates support versioning for A/B testing. Currently supported versions:

- `v1` - Standard prompts
- `v2` - Optimized prompts with domain-specific enhancements

Set the version via `PROMPT_*_VERSION` environment variables.

## Intent Categories

| Intent | Description |
|--------|-------------|
| `information_seeking` | Factual questions and explanations |
| `problem_solving` | Math problems, exercises, equations |
| `content_retrieval` | Finding specific content by chapter/grade/type |
| `clarification` | Follow-ups, elaboration requests |
| `comparison` | Comparing concepts, differences |
| `definition` | Definitions, meanings |
| `procedural` | How-to, step-by-step instructions |
| `conversational` | Greetings, acknowledgments |
| `exam_prep` | Exam questions, preparation |

## Filter Schema

| Filter | Description | Normalization |
|--------|-------------|---------------|
| `grade` | Grade/class numbers | Extracted as numeric string "10" |
| `subject` | Academic subjects | Lowercase |
| `chapter` | Chapter numbers or names | As extracted |
| `topic` | Specific topics/concepts | As extracted |
| `content_type` | Theory, exercises, etc. | Lowercase |
| `board` | Education board (CBSE, ICSE) | Uppercase standard boards |
| `language` | Content language | Lowercase |

---

## See Also

### Related Services
- 📦 [API Gateway](../api-gateway/) — Main HTTP gateway with CAG orchestration
- 📦 [Embedding Service](../embedding-service/) — Vertex AI embeddings with caching
- 📦 [Vector Search Service](../vector-search-service/) — Hybrid search (dense + sparse + RRF)

### Architecture & Design
- 📄 [Complete Architecture](../../docs/architecture/MICROSERVICES_COMPLETE.md) — Full microservices architecture
- 📄 [CAG Architecture](../../docs/caching/CAG_ARCHITECTURE.md) — 5-layer caching design
- 📄 [Main README](../../README.md) — Project overview
- 📄 [Complete Documentation](../../docs/COMPLETE_DOCUMENTATION.md) — Full technical reference (~1500 lines)

### Technical Guides
- 📖 [Query Expansion Guide](../../docs/guides/guide_query_expansion.md) — Query expansion strategies
- 📖 [Retrieval Guide](../../docs/guides/guide_retrieval.md) — Retrieval strategies
- 📖 [Reranking Guide](../../docs/guides/guide_reranking.md) — Reranking strategies

### Audits & Performance
- 📊 [Query Expansion Audit](../../docs/audits/audit_langchaingo_query_expansion.md) — Query expansion analysis
- 📊 [Advanced Retrieval Audit](../../docs/audits/audit_langchaingo_advanced_retrieval.md) — Retrieval analysis
- 📊 [Latency Audit](../../docs/audits/audit_langchaingo_latency.md) — Pipeline latency

### Quality & Security
- 🛡️ [Defensive Fixes](../../DEFENSIVE_FIXES.md) — Defensive programming improvements
- 🛡️ [API Design Fixes](../../GO_API_DESIGN_FIXES.md) — API design improvements
- 🛡️ [Code Quality](../../docs/quality/CODE_QUALITY_IMPROVEMENTS.md) — Code quality improvements
- 🛡️ [Sanitizer Tests](sanitizer/sanitizer_test.go) — Input sanitization tests

### Deployment
- 🚀 [Docker Compose](../../docker-compose.yml) — Local deployment
- 🚀 [Supabase Guide](../../supabase/README.md) — Cloud deployment
- 🚀 [Terraform/GCP](../../terraform/) — Enterprise deployment
