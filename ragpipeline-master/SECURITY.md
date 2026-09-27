# Security Policy

## Supported Versions

| Version | Supported |
|---------|-----------|
| main (reorganize-microservices) | ✅ |
| earlier commits | ❌ |

## Reporting a Vulnerability

Report security vulnerabilities by opening a GitHub Security Advisory at:
https://github.com/ruthvik-visionary/ragpipeline/security/advisories/new

**Do NOT** report vulnerabilities in public issues or discussions.

You will receive a response within 48 hours. If the issue is confirmed, we will release a patch as soon as possible.

## Security Measures

### Authentication
- JWT Bearer token validation on all protected endpoints
- HMAC-SHA256 signature verification
- Token expiry validation
- User ID extraction from claims for per-user rate limiting

### Input Validation
- Query length limits (1000 chars max)
- Conversation history limits (20 entries, 2000 chars each)
- Prompt injection pattern detection (12+ patterns blocked)
- Control character and null byte stripping
- Role validation (no user-injected system messages)

### Rate Limiting
- Per-user rate limiting (100 req/min default)
- Redis-backed distributed rate limiting
- In-memory fallback when Redis is unavailable
- Burst size configurable (20 requests)

### Infrastructure
- GCP Secret Manager for credentials (never in code or logs)
- Service Account with minimal permissions
- VPC-only service-to-service communication
- HTTPS for all external traffic
- Redis with password protection

### Data Protection
- PII stripped from error messages and logs
- No sensitive data in request/response bodies
- Secure token generation using `crypto/rand`
- No `math/rand` usage for security-sensitive operations

### Concurrency Safety
- All shared state protected by mutexes
- Atomic operations for simple flags/counters
- No goroutine leaks (verified with context cancellation)
- Proper `defer cancel()` after derived contexts
- `defer Unlock()` for all mutex acquisitions

### Dependencies
- All dependencies pinned via `go.sum`
- Regular `go mod tidy` to remove unused deps
- No known vulnerable dependencies (verified via `govulncheck`)

## Known Limitations

- Prompt injection detection relies on pattern matching, not ML classification
- Rate limiting uses fixed windows (not sliding) for simplicity
- JWT secret must be rotated manually (no automated key rotation yet)
- No Web Application Firewall (WAF) in front of the API Gateway
