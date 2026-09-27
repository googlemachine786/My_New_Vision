# Contributing to RAG Pipeline

Thank you for your interest in contributing to the Visionary RAG Pipeline! This document provides guidelines for contributing to this project.

## Code of Conduct

- Be respectful and inclusive
- Provide constructive feedback
- Focus on what is best for the community

## Getting Started

### Prerequisites

- Go 1.21+
- Python 3.11+ (for ingestion pipeline)
- Docker & Docker Compose
- Redis (for caching)
- PostgreSQL with pgvector or Supabase

### Development Setup

1. **Fork and clone the repository**
```bash
git clone https://github.com/YOUR_USERNAME/ragpipeline.git
cd ragpipeline
```

2. **Start all services with Docker Compose**
```bash
docker-compose up -d
```

3. **Run services locally (without Docker)**
```bash
# Start Redis
docker run -d -p 6379:6379 redis:7-alpine

# Start each service in separate terminals
cd services/embedding-service && go run cmd/server/main.go
cd services/vector-search-service && go run cmd/server/main.go
cd services/query-understanding-service && go run cmd/server/main.go
cd services/api-gateway && go run cmd/server/main.go
```

## Architecture

This project uses a **microservices architecture** with 4 Go services:

| Service | Port | Purpose |
|---------|------|---------|
| API Gateway | 8080 | Request orchestration, auth, caching |
| Embedding Service | 8081 | Vertex AI embedding generation |
| Vector Search Service | 8082 | Hybrid search (dense + sparse + RRF) |
| Query Understanding | 8083 | Query rewriting, self-query parsing |

**Key Features:**
- Multi-layer caching (CAG) for 80% cost reduction
- Circuit breakers for fault tolerance
- SSE streaming for real-time responses
- Redis-backed caching and session management

## Development Workflow

### 1. Create a Feature Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-description
```

### 2. Write Code

**Go Code Standards:**
- Follow [Effective Go](https://go.dev/doc/effective_go)
- Use `gofmt` for formatting
- Run `go vet` before committing
- Write tests for new functionality
- Use structured logging with zerolog
- Include request ID tracing
- Implement proper error handling

**Python Code Standards:**
- Follow PEP 8
- Use type hints
- Write docstrings
- Add tests with pytest

### 3. Run Tests

```bash
# Run all tests
make test

# Run tests for specific service
cd services/api-gateway && go test ./...

# Run with coverage
make test-coverage
```

### 4. Commit Your Changes

We use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add semantic cache for query understanding
fix: resolve circuit breaker race condition
docs: update API gateway documentation
test: add integration tests for vector search
refactor: extract common middleware to pkg/
perf: optimize RRF fusion algorithm
```

### 5. Push and Create Pull Request

```bash
git push origin feature/your-feature-name
```

Open a PR on GitHub with:
- Clear description of changes
- Link to related issues
- Screenshots/logs if applicable
- Test results

## Pull Request Process

1. **Review Requirements**
   - All tests must pass
   - Code must be formatted (`gofmt`)
   - No linting errors
   - Documentation updated if needed

2. **Code Review**
   - At least one maintainer must approve
   - Address all review comments
   - Keep discussions respectful

3. **Merge**
   - Squash and merge preferred
   - Delete feature branch after merge

## Project Structure

```
ragpipeline/
├── services/                 # Go microservices
│   ├── api-gateway/         # Main HTTP interface
│   ├── embedding-service/   # Vertex AI embeddings
│   ├── vector-search-service/  # Hybrid search
│   └── query-understanding-service/  # LLM query understanding
├── ingestion/               # Python ingestion pipeline
├── frontend/                # React web interface
├── docs/                    # Documentation
│   ├── architecture/        # System design
│   ├── caching/            # CAG implementation
│   ├── guides/             # Technical guides
│   └── audits/             # Code audits
├── terraform/              # GCP infrastructure
├── scripts/                # Utility scripts
└── data/                   # Test data
```

## Testing

### Unit Tests
```bash
make test-unit
```

### Integration Tests
```bash
make test-integration
```

### Load Tests
```bash
make test-load
```

## Documentation

- Update `README.md` for user-facing changes
- Update `docs/` for technical details
- Add godoc comments for public APIs
- Include examples in documentation

## Reporting Bugs

Use the GitHub issue tracker with:
- **Description**: Clear, concise description
- **Steps to Reproduce**: Exact steps
- **Expected Behavior**: What should happen
- **Actual Behavior**: What actually happens
- **Environment**: OS, Go version, Docker version
- **Logs**: Relevant error messages

## Suggesting Enhancements

- **Title**: Clear, specific title
- **Description**: Detailed explanation
- **Use Case**: Why this is valuable
- **Implementation**: How it could work
- **Alternatives**: Other approaches considered

## License

By contributing, you agree that your contributions will be licensed under the project's license.

## Questions?

- **Architecture**: See `docs/architecture/`
- **Setup**: See `README.md` Quick Start
- **API**: See service `README.md` files
- **Caching**: See `docs/caching/`

Thank you for contributing! 🚀
