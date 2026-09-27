# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added
- 5-layer Cache Augmented Generation (CAG) system for 72% cost reduction
- Shared `pkg/` packages: `types`, `httperrors`, `contextkeys`, `circuitbreaker`
- Functional options pattern for all constructors with 3+ parameters
- Config struct pattern for handlers with multiple dependencies
- Comprehensive test coverage for new packages

### Changed
- Replaced monolithic `orchestrator/` with 4 independent Go microservices
- Rate limiter now uses `golang.org/x/time/rate` (no goroutine leaks)
- Async cache writes use detached background contexts
- Circuit breaker `Stats()` is pure read with no side effects
- All cache constructors use shared `CacheOption` pattern

### Removed
- Legacy `ingestion-go/` stub (use Python `ingestion/`)
- Duplicate circuit breaker implementation
- Unnecessary mutex in synchronous retry logic
- 45+ redundant documentation files archived

### Fixed
- Goroutine leak in rate limiter token replenishment
- Async goroutines using cancelled request contexts
- `wg.Wait()` blocking indefinitely on cancelled contexts
- Race condition potential in local rate limiter map access
- `interface{}` usage replaced with typed `SearchCache` interface

## [1.0.0] - 2026-04-06

### Added
- 4 Go microservices with independent deployment
- Multi-layer CAG caching (exact, semantic, embedding, search, template)
- JWT authentication middleware
- Distributed rate limiting via Redis
- Server-Sent Events (SSE) streaming for LLM responses
- Circuit breakers for all downstream service calls
- OpenTelemetry distributed tracing
- Docker Compose for local development
- GitHub Actions CI/CD pipeline

[Unreleased]: https://github.com/ruthvik-visionary/ragpipeline/compare/main...reorganize-microservices
[1.0.0]: https://github.com/ruthvik-visionary/ragpipeline/releases/tag/v1.0.0
