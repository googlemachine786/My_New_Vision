#!/bin/bash
# Visionary RAG Pipeline - Schema Verification Script
# Phase 1 - Task T1.5
# Purpose: Verify schema correctness after applying v2_production.sql
#
# Usage: ./scripts/verify-schema.sh
# Prerequisites:
#   - AlloyDB cluster running
#   - Schema applied: psql $ALLOYDB_DSN -f schema/v2_production.sql
#   - psql client installed
#   - ALLOYDB_DSN environment variable set

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Counters
PASS=0
FAIL=0

# Helper functions
pass() {
    echo -e "${GREEN}✓${NC} $1"
    ((PASS++))
}

fail() {
    echo -e "${RED}✗${NC} $1"
    ((FAIL++))
}

info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

# Check if ALLOYDB_DSN is set
if [[ -z "${ALLOYDB_DSN:-}" ]]; then
    echo -e "${RED}Error: ALLOYDB_DSN environment variable not set${NC}"
    echo "Usage: export ALLOYDB_DSN='host=IP dbname=visionary user=visionary password=XXX'"
    exit 1
fi

echo "========================================"
echo "Visionary RAG - Schema Verification"
echo "========================================"
echo ""

# =============================================================================
# Test 1: Verify TEXT[] for extracted_keywords
# =============================================================================
info "Test 1: Verifying extracted_keywords is TEXT[] (not TEXT)..."

RESULT=$(psql "$ALLOYDB_DSN" -t -A -c "
SELECT data_type, udt_name
FROM information_schema.columns
WHERE table_name = 'parent_chunks' 
  AND column_name = 'extracted_keywords';
")

DATA_TYPE=$(echo "$RESULT" | cut -d'|' -f1)
UDT_NAME=$(echo "$RESULT" | cut -d'|' -f2)

if [[ "$DATA_TYPE" == "ARRAY" && "$UDT_NAME" == "_text" ]]; then
    pass "extracted_keywords is TEXT[] (data_type=$DATA_TYPE, udt_name=$UDT_NAME)"
else
    fail "extracted_keywords is NOT TEXT[] (data_type=$DATA_TYPE, udt_name=$UDT_NAME)"
fi

# =============================================================================
# Test 2: Verify UUID[] for retrieved_context
# =============================================================================
info "Test 2: Verifying retrieved_context is UUID[] (not UUID)..."

RESULT=$(psql "$ALLOYDB_DSN" -t -A -c "
SELECT data_type, udt_name
FROM information_schema.columns
WHERE table_name = 'ai_feedback_loop' 
  AND column_name = 'retrieved_context';
")

DATA_TYPE=$(echo "$RESULT" | cut -d'|' -f1)
UDT_NAME=$(echo "$RESULT" | cut -d'|' -f2)

if [[ "$DATA_TYPE" == "ARRAY" && "$UDT_NAME" == "_uuid" ]]; then
    pass "retrieved_context is UUID[] (data_type=$DATA_TYPE, udt_name=$UDT_NAME)"
else
    fail "retrieved_context is NOT UUID[] (data_type=$DATA_TYPE, udt_name=$UDT_NAME)"
fi

# =============================================================================
# Test 3: Verify ScaNN index exists
# =============================================================================
info "Test 3: Verifying ScaNN index on child_chunks..."

RESULT=$(psql "$ALLOYDB_DSN" -t -A -c "
SELECT indexname, indexdef 
FROM pg_indexes 
WHERE tablename = 'child_chunks' 
  AND indexname = 'idx_child_embedding_scann';
")

if [[ -n "$RESULT" && "$RESULT" == *"scann"* ]]; then
    pass "ScaNN index idx_child_embedding_scann exists"
else
    fail "ScaNN index idx_child_embedding_scann NOT found"
fi

# =============================================================================
# Test 4: Verify GIN index on extracted_keywords
# =============================================================================
info "Test 4: Verifying GIN index on parent_chunks.extracted_keywords..."

RESULT=$(psql "$ALLOYDB_DSN" -t -A -c "
SELECT indexname, indexdef 
FROM pg_indexes 
WHERE tablename = 'parent_chunks' 
  AND indexname = 'idx_parent_keywords_gin';
")

if [[ -n "$RESULT" && "$RESULT" == *"gin"* && "$RESULT" == *"extracted_keywords"* ]]; then
    pass "GIN index idx_parent_keywords_gin exists"
else
    fail "GIN index idx_parent_keywords_gin NOT found"
fi

# =============================================================================
# Test 5: Verify partial indexes exist
# =============================================================================
info "Test 5: Verifying partial indexes..."

RESULT=$(psql "$ALLOYDB_DSN" -t -A -c "
SELECT indexname 
FROM pg_indexes 
WHERE indexdef LIKE '%WHERE%';
")

if [[ "$RESULT" == *"idx_dlq_failed"* ]]; then
    pass "Partial index idx_dlq_failed exists"
else
    fail "Partial index idx_dlq_failed NOT found"
fi

if [[ "$RESULT" == *"idx_feedback_unprocessed"* ]]; then
    pass "Partial index idx_feedback_unprocessed exists"
else
    fail "Partial index idx_feedback_unprocessed NOT found"
fi

# =============================================================================
# Test 6: Verify extensions are installed
# =============================================================================
info "Test 6: Verifying required extensions..."

for EXT in vector btree_gin pgcrypto; do
    RESULT=$(psql "$ALLOYDB_DSN" -t -A -c "
    SELECT 1 FROM pg_extension WHERE extname = '$EXT' LIMIT 1;
    ")
    
    if [[ "$RESULT" == "1" ]]; then
        pass "Extension '$EXT' is installed"
    else
        fail "Extension '$EXT' is NOT installed"
    fi
done

# =============================================================================
# Test 7: Verify taxonomy population
# =============================================================================
info "Test 7: Verifying cbse_taxonomy population..."

RESULT=$(psql "$ALLOYDB_DSN" -t -A -c "
SELECT COUNT(*) FROM cbse_taxonomy;
")

if [[ "$RESULT" -ge 50 ]]; then
    pass "Taxonomy populated with $RESULT entries (expected: ~52)"
else
    fail "Taxonomy has only $RESULT entries (expected: ~52)"
fi

# Verify grade distribution
info "Test 8: Verifying grade distribution..."

for GRADE in 6 7 8; do
    RESULT=$(psql "$ALLOYDB_DSN" -t -A -c "
    SELECT COUNT(*) FROM cbse_taxonomy WHERE grade = $GRADE;
    ")
    
    if [[ "$RESULT" -ge 15 ]]; then
        pass "Grade $GRADE has $RESULT chapters"
    else
        fail "Grade $GRADE has only $RESULT chapters (expected: ~16-18)"
    fi
done

# =============================================================================
# Test 9: Verify vector type is registered
# =============================================================================
info "Test 9: Verifying vector type registration..."

RESULT=$(psql "$ALLOYDB_DSN" -t -A -c "
SELECT 1 FROM pg_type WHERE typname = 'vector' LIMIT 1;
")

if [[ "$RESULT" == "1" ]]; then
    pass "Vector type is registered"
else
    fail "Vector type is NOT registered"
fi

# =============================================================================
# Test 10: Verify table structure
# =============================================================================
info "Test 10: Verifying table structure..."

TABLES=("cbse_taxonomy" "parent_chunks" "child_chunks" "ingestion_dlq" "ai_feedback_loop")

for TABLE in "${TABLES[@]}"; do
    RESULT=$(psql "$ALLOYDB_DSN" -t -A -c "
    SELECT 1 FROM information_schema.tables 
    WHERE table_name = '$TABLE' LIMIT 1;
    ")
    
    if [[ "$RESULT" == "1" ]]; then
        pass "Table '$TABLE' exists"
    else
        fail "Table '$TABLE' does NOT exist"
    fi
done

# =============================================================================
# Summary
# =============================================================================
echo ""
echo "========================================"
echo "Verification Summary"
echo "========================================"
echo -e "${GREEN}Passed: $PASS${NC}"
echo -e "${RED}Failed: $FAIL${NC}"
echo ""

if [[ $FAIL -eq 0 ]]; then
    echo -e "${GREEN}✓ All tests passed! Schema is correctly applied.${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed. Review output above.${NC}"
    exit 1
fi
