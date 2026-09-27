.PHONY: help build test run clean lint fmt vet docker-up docker-down db-migrate db-seed

# Default target
help: ## Show this help message
	@echo "Visionary RAG Pipeline - Makefile"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# Build all services
build: ## Build all Go services
	@echo "Building all services..."
	@cd services/api-gateway && go build -o ../../bin/api-gateway ./cmd/server/main.go
	@cd services/embedding-service && go build -o ../../bin/embedding-service ./cmd/server/main.go
	@cd services/vector-search-service && go build -o ../../bin/vector-search-service ./cmd/server/main.go
	@cd services/query-understanding-service && go build -o ../../bin/query-understanding-service ./cmd/server/main.go
	@echo "Build complete! Binaries in bin/"

# Run all services
run: ## Run all services locally
	@echo "Starting all services..."
	@docker-compose up -d redis
	@cd services/embedding-service && go run cmd/server/main.go &
	@cd services/vector-search-service && go run cmd/server/main.go &
	@cd services/query-understanding-service && go run cmd/server/main.go &
	@cd services/api-gateway && go run cmd/server/main.go &
	@echo "Services started. Press Ctrl+C to stop."

# Run tests
test: ## Run all tests
	@echo "Running tests..."
	@cd services/api-gateway && go test -v ./...
	@cd services/embedding-service && go test -v ./...
	@cd services/vector-search-service && go test -v ./...
	@cd services/query-understanding-service && go test -v ./...
	@cd orchestrator && go test -v ./...
	@cd pkg && go test -v ./...
	@cd shared/go && go test -v ./...
	@echo "Tests complete!"

# P2-20: Run tests with race detector
test-race: ## Run tests with race detector (P2-20)
	@echo "Running tests with race detector..."
	@cd services/api-gateway && go test -race -v ./...
	@cd services/embedding-service && go test -race -v ./...
	@cd services/vector-search-service && go test -race -v ./...
	@cd services/query-understanding-service && go test -race -v ./...
	@cd orchestrator && go test -race -v ./...
	@cd pkg && go test -race -v ./...
	@cd shared/go && go test -race -v ./...
	@echo "Race detection complete!"

# Run tests with coverage
test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	@mkdir -p coverage
	@cd services/api-gateway && go test -v -coverprofile=../../coverage/api-gateway.out ./...
	@cd services/embedding-service && go test -v -coverprofile=../../coverage/embedding.out ./...
	@cd services/vector-search-service && go test -v -coverprofile=../../coverage/vector-search.out ./...
	@cd services/query-understanding-service && go test -v -coverprofile=../../coverage/query-understanding.out ./...
	@cd orchestrator && go test -v -coverprofile=../../coverage/orchestrator.out ./...
	@cd pkg && go test -v -coverprofile=../../coverage/pkg.out ./...
	@cd shared/go && go test -v -coverprofile=../../coverage/shared.out ./...
	@go tool cover -html=coverage/api-gateway.out -o coverage/api-gateway.html
	@echo "Coverage reports generated in coverage/"

# P2-19: Run benchmarks for hot paths
bench: ## Run benchmarks for hot paths (P2-19)
	@echo "Running benchmarks..."
	@cd orchestrator/retrieval && go test -bench=. -benchmem -count=3 ./...
	@cd orchestrator/observability && go test -bench=. -benchmem -count=3 ./...
	@cd orchestrator/reranker && go test -bench=. -benchmem -count=3 ./...
	@cd pkg/circuitbreaker && go test -bench=. -benchmem -count=3 ./...
	@cd shared/go/utils && go test -bench=. -benchmem -count=3 ./...
	@cd shared/go/analytics && go test -bench=. -benchmem -count=3 ./...
	@cd services/vector-search-service/search && go test -bench=. -benchmem -count=3 ./...
	@echo "Benchmarks complete!"

# Lint all services
lint: ## Run linters
	@echo "Running linters..."
	@cd services/api-gateway && golangci-lint run
	@cd services/embedding-service && golangci-lint run
	@cd services/vector-search-service && golangci-lint run
	@cd services/query-understanding-service && golangci-lint run
	@echo "Linting complete!"

# Format code
fmt: ## Format Go code
	@echo "Formatting code..."
	@find services -name "*.go" -exec gofmt -w {} \;
	@echo "Formatting complete!"

# Run go vet
vet: ## Run go vet
	@echo "Running go vet..."
	@cd services/api-gateway && go vet ./...
	@cd services/embedding-service && go vet ./...
	@cd services/vector-search-service && go vet ./...
	@cd services/query-understanding-service && go vet ./...
	@echo "Vet complete!"

# Docker commands
docker-up: ## Start all services with Docker Compose
	@echo "Starting services with Docker Compose..."
	@docker-compose up -d
	@echo "Services started!"
	@docker-compose ps

docker-down: ## Stop all Docker Compose services
	@echo "Stopping services..."
	@docker-compose down
	@echo "Services stopped!"

docker-logs: ## Show Docker Compose logs
	@docker-compose logs -f

docker-build: ## Build Docker images
	@docker-compose build

# Database commands
db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	@cd schema && psql $$DATABASE_URL -f 001_initial_schema.sql
	@echo "Migrations complete!"

db-seed: ## Seed database with test data
	@echo "Seeding database..."
	@cd scripts && python seed_test_data.py
	@echo "Seeding complete!"

# Clean build artifacts
clean: ## Clean build artifacts
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -rf coverage/
	@rm -rf services/*/bin/
	@find . -name "*.test" -delete
	@find . -name "*.out" -delete
	@echo "Clean complete!"

# Initialize development environment
setup: ## Setup development environment
	@echo "Setting up development environment..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@docker-compose up -d redis postgres
	@echo "Setup complete!"

# Check service health
health: ## Check health of all services
	@echo "Checking service health..."
	@curl -s http://localhost:8080/health | jq .
	@curl -s http://localhost:8081/health | jq .
	@curl -s http://localhost:8082/health | jq .
	@curl -s http://localhost:8083/health | jq .

# Show cache statistics
cache-stats: ## Show CAG cache statistics
	@echo "Cache statistics:"
	@curl -s http://localhost:8080/cag/stats | jq .
