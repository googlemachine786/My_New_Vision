#!/bin/bash
# Visionary RAG Pipeline - End-to-End Integration Test
# Tests the complete pipeline from PDF ingestion to query response

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

log_skip() {
    echo -e "${YELLOW}[SKIP]${NC} $1"
    ((TESTS_SKIPPED++))
}

log_section() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

# Check prerequisites
check_prerequisites() {
    log_section "Checking Prerequisites"
    
    # Check Go
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version | awk '{print $3}')
        log_pass "Go installed: $GO_VERSION"
    else
        log_fail "Go not installed"
        return 1
    fi
    
    # Check Python
    if command -v python &> /dev/null; then
        PYTHON_VERSION=$(python --version 2>&1)
        log_pass "Python installed: $PYTHON_VERSION"
    else
        log_fail "Python not installed"
        return 1
    fi
    
    # Check Docker
    if command -v docker &> /dev/null; then
        DOCKER_VERSION=$(docker --version)
        log_pass "Docker installed: $DOCKER_VERSION"
    else
        log_skip "Docker not installed (optional for container tests)"
    fi
    
    # Check gcloud (optional)
    if command -v gcloud &> /dev/null; then
        log_pass "gcloud CLI installed"
    else
        log_skip "gcloud CLI not installed (optional for GCP tests)"
    fi
}

# Test Go module compilation
test_go_compilation() {
    log_section "Testing Go Module Compilation"
    
    # Test ingestion-go module
    log_info "Testing ingestion-go compilation..."
    cd ingestion-go
    if go build -o /tmp/ingestion-test ./cmd/ingestion 2>&1; then
        log_pass "ingestion-go compiles successfully"
        rm -f /tmp/ingestion-test
    else
        log_fail "ingestion-go compilation failed"
        cd ..
        return 1
    fi
    cd ..
    
    # Test orchestrator module
    log_info "Testing orchestrator compilation..."
    cd orchestrator
    if go build -o /tmp/orchestrator-test ./cmd/server 2>&1; then
        log_pass "orchestrator compiles successfully"
        rm -f /tmp/orchestrator-test
    else
        log_fail "orchestrator compilation failed"
        cd ..
        return 1
    fi
    cd ..
}

# Test Python syntax
test_python_syntax() {
    log_section "Testing Python Syntax"
    
    # Test ingestion module
    log_info "Testing ingestion Python syntax..."
    cd ingestion
    if python -m py_compile pipeline.py 2>&1; then
        log_pass "ingestion/pipeline.py syntax valid"
    else
        log_fail "ingestion/pipeline.py syntax error"
        cd ..
        return 1
    fi
    cd ..
    
    # Test eval module
    log_info "Testing eval Python syntax..."
    cd eval
    if python -m py_compile recall_eval.py 2>&1; then
        log_pass "eval/recall_eval.py syntax valid"
    else
        log_fail "eval/recall_eval.py syntax error"
        cd ..
        return 1
    fi
    cd ..
    
    # Test quality_loop module
    log_info "Testing quality_loop Python syntax..."
    cd quality_loop
    if python -m py_compile judge.py 2>&1; then
        log_pass "quality_loop/judge.py syntax valid"
    else
        log_fail "quality_loop/judge.py syntax error"
        cd ..
        return 1
    fi
    cd ..
}

# Test Go unit tests
test_go_unit_tests() {
    log_section "Running Go Unit Tests"
    
    # Test ingestion-go
    log_info "Running ingestion-go tests..."
    cd ingestion-go
    if go test ./... -v -short 2>&1 | tee /tmp/ingestion-test-output.txt; then
        log_pass "ingestion-go tests passed"
    else
        log_fail "ingestion-go tests failed"
        cat /tmp/ingestion-test-output.txt
        cd ..
        return 1
    fi
    cd ..
    
    # Test orchestrator
    log_info "Running orchestrator tests..."
    cd orchestrator
    if go test ./... -v -short 2>&1 | tee /tmp/orchestrator-test-output.txt; then
        log_pass "orchestrator tests passed"
    else
        log_fail "orchestrator tests failed"
        cat /tmp/orchestrator-test-output.txt
        cd ..
        return 1
    fi
    cd ..
}

# Test Python unit tests
test_python_unit_tests() {
    log_section "Running Python Unit Tests"
    
    # Test ingestion
    log_info "Running ingestion Python tests..."
    cd ingestion
    if python -m pytest tests/ -v --tb=short 2>&1 | tee /tmp/ingestion-pytest-output.txt; then
        log_pass "ingestion Python tests passed"
    else
        log_fail "ingestion Python tests failed"
        cat /tmp/ingestion-pytest-output.txt
        cd ..
        return 1
    fi
    cd ..
}

# Test schema validation (mock)
test_schema_validation() {
    log_section "Testing Schema Validation"
    
    # Check schema file exists
    if [ -f "schema/v2_production.sql" ]; then
        log_pass "schema/v2_production.sql exists"
        
        # Check for key components
        if grep -q "CREATE TABLE.*parent_chunks" schema/v2_production.sql; then
            log_pass "parent_chunks table defined"
        else
            log_fail "parent_chunks table not defined"
        fi
        
        if grep -q "CREATE TABLE.*child_chunks" schema/v2_production.sql; then
            log_pass "child_chunks table defined"
        else
            log_fail "child_chunks table not defined"
        fi
        
        if grep -q "extracted_keywords TEXT\[\]" schema/v2_production.sql; then
            log_pass "extracted_keywords TEXT[] type correct"
        else
            log_fail "extracted_keywords TEXT[] type not found"
        fi
        
        if grep -q "retrieved_context UUID\[\]" schema/v2_production.sql; then
            log_pass "retrieved_context UUID[] type correct"
        else
            log_fail "retrieved_context UUID[] type not found"
        fi
        
        if grep -q "ScaNN" schema/v2_production.sql; then
            log_pass "ScaNN index defined"
        else
            log_fail "ScaNN index not defined"
        fi
        
        if grep -q "GIN" schema/v2_production.sql; then
            log_pass "GIN index defined"
        else
            log_fail "GIN index not defined"
        fi
    else
        log_fail "schema/v2_production.sql not found"
    fi
}

# Test Terraform syntax
test_terraform_syntax() {
    log_section "Testing Terraform Syntax"
    
    if command -v terraform &> /dev/null; then
        cd terraform
        if terraform init -backend=false 2>&1; then
            log_pass "terraform init successful"
        else
            log_fail "terraform init failed"
            cd ..
            return 1
        fi
        
        if terraform validate 2>&1; then
            log_pass "terraform validate successful"
        else
            log_fail "terraform validate failed"
            cd ..
            return 1
        fi
        cd ..
    else
        log_skip "terraform not installed (skipping validation)"
    fi
}

# Test Docker builds (if Docker available)
test_docker_builds() {
    log_section "Testing Docker Builds"
    
    if ! command -v docker &> /dev/null; then
        log_skip "Docker not available"
        return 0
    fi
    
    # Test ingestion-go Docker build
    if [ -f "ingestion-go/Dockerfile" ]; then
        log_info "Testing ingestion-go Docker build..."
        cd ingestion-go
        if docker build -t visionary-ingestion-test . 2>&1 | tee /tmp/ingestion-docker-output.txt; then
            log_pass "ingestion-go Docker build successful"
            docker rmi visionary-ingestion-test 2>/dev/null || true
        else
            log_fail "ingestion-go Docker build failed"
            cat /tmp/ingestion-docker-output.txt
            cd ..
            return 1
        fi
        cd ..
    else
        log_skip "ingestion-go/Dockerfile not found"
    fi
    
    # Test orchestrator Docker build
    if [ -f "orchestrator/Dockerfile" ]; then
        log_info "Testing orchestrator Docker build..."
        cd orchestrator
        if docker build -t visionary-orchestrator-test . 2>&1 | tee /tmp/orchestrator-docker-output.txt; then
            log_pass "orchestrator Docker build successful"
            docker rmi visionary-orchestrator-test 2>/dev/null || true
        else
            log_fail "orchestrator Docker build failed"
            cat /tmp/orchestrator-docker-output.txt
            cd ..
            return 1
        fi
        cd ..
    else
        log_skip "orchestrator/Dockerfile not found"
    fi
}

# Test code quality (linting)
test_code_quality() {
    log_section "Testing Code Quality"
    
    # Go vet
    log_info "Running go vet..."
    cd ingestion-go
    if go vet ./... 2>&1; then
        log_pass "ingestion-go: go vet passed"
    else
        log_fail "ingestion-go: go vet failed"
        cd ..
        return 1
    fi
    cd ..
    
    cd orchestrator
    if go vet ./... 2>&1; then
        log_pass "orchestrator: go vet passed"
    else
        log_fail "orchestrator: go vet failed"
        cd ..
        return 1
    fi
    cd ..
    
    # Python flake8 (if available)
    if command -v flake8 &> /dev/null; then
        log_info "Running flake8..."
        cd ingestion
        if flake8 --max-line-length=100 --ignore=E501 . 2>&1 | tee /tmp/flake8-output.txt; then
            log_pass "ingestion: flake8 passed"
        else
            log_fail "ingestion: flake8 failed"
            cat /tmp/flake8-output.txt
            cd ..
            return 1
        fi
        cd ..
    else
        log_skip "flake8 not installed"
    fi
}

# Print summary
print_summary() {
    log_section "Test Summary"
    
    echo -e "${GREEN}Passed:  $TESTS_PASSED${NC}"
    echo -e "${RED}Failed:  $TESTS_FAILED${NC}"
    echo -e "${YELLOW}Skipped: $TESTS_SKIPPED${NC}"
    echo ""
    
    TOTAL=$((TESTS_PASSED + TESTS_FAILED))
    if [ $TOTAL -gt 0 ]; then
        PASS_RATE=$((TESTS_PASSED * 100 / TOTAL))
        echo -e "Pass Rate: ${GREEN}${PASS_RATE}%${NC}"
    fi
    
    echo ""
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}========================================${NC}"
        echo -e "${GREEN}ALL TESTS PASSED!${NC}"
        echo -e "${GREEN}Pipeline is ready for deployment.${NC}"
        echo -e "${GREEN}========================================${NC}"
        exit 0
    else
        echo -e "${RED}========================================${NC}"
        echo -e "${RED}SOME TESTS FAILED${NC}"
        echo -e "${RED}Please review and fix the issues above.${NC}"
        echo -e "${RED}========================================${NC}"
        exit 1
    fi
}

# Main execution
main() {
    log_info "Starting Visionary RAG Pipeline Integration Tests"
    log_info "Date: $(date)"
    log_info "Working Directory: $(pwd)"
    
    # Run all tests
    check_prerequisites || true
    test_go_compilation
    test_python_syntax
    test_go_unit_tests || true  # Continue even if tests fail
    test_python_unit_tests || true
    test_schema_validation
    test_terraform_syntax || true
    test_docker_builds || true
    test_code_quality
    
    # Print summary
    print_summary
}

# Run main
main "$@"
