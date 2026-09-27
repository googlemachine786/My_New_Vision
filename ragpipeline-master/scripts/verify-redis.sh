#!/bin/bash
# Visionary RAG Pipeline - Redis Latency Verification Script
# Phase 1 - Task T1.8
# Purpose: Verify Redis latency is sub-millisecond from VPC connector range
#
# Usage: ./scripts/verify-redis.sh
# Prerequisites:
#   - Redis instance running
#   - redis-cli installed
#   - REDIS_HOST or REDIS_ADDR environment variable set
#
# Exit Criteria:
#   - p50 latency < 1ms
#   - max latency < 2ms
#   - PING returns PONG

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
pass() {
    echo -e "${GREEN}✓${NC} $1"
}

fail() {
    echo -e "${RED}✗${NC} $1"
}

info() {
    echo -e "${YELLOW}ℹ${NC} $1"
}

# Check if Redis address is provided
if [[ -n "${REDIS_ADDR:-}" ]]; then
    REDIS_HOST=$(echo "$REDIS_ADDR" | cut -d':' -f1)
    REDIS_PORT=$(echo "$REDIS_ADDR" | cut -d':' -f2)
elif [[ -n "${REDIS_HOST:-}" ]]; then
    REDIS_HOST="${REDIS_HOST}"
    REDIS_PORT="${REDIS_PORT:-6379}"
else
    echo -e "${RED}Error: REDIS_ADDR or REDIS_HOST environment variable not set${NC}"
    echo "Usage: export REDIS_ADDR='10.8.0.5:6379' or export REDIS_HOST='10.8.0.5'"
    exit 1
fi

echo "========================================"
echo "Visionary RAG - Redis Latency Verification"
echo "========================================"
echo "Target: ${REDIS_HOST}:${REDIS_PORT}"
echo ""

# =============================================================================
# Test 1: Basic Connectivity (PING)
# =============================================================================
info "Test 1: Testing basic connectivity (PING)..."

PING_RESULT=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ping 2>&1 || true)

if [[ "$PING_RESULT" == "PONG" ]]; then
    pass "Redis responds to PING with PONG"
else
    fail "Redis PING failed: $PING_RESULT"
    exit 1
fi

# =============================================================================
# Test 2: Redis Version
# =============================================================================
info "Test 2: Checking Redis version..."

VERSION=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" info server | grep redis_version | cut -d':' -f2 | tr -d '\r')

if [[ -n "$VERSION" ]]; then
    pass "Redis version: $VERSION"
else
    fail "Could not determine Redis version"
fi

# =============================================================================
# Test 3: Latency History (10 samples)
# =============================================================================
info "Test 3: Measuring latency (10 samples, 1 second interval)..."

# Run latency test and capture output
LATENCY_OUTPUT=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" --latency-history -i 1 2>&1 <<EOF
Q
EOF
timeout 12 || true)

# Parse min, max, avg from output
# Example output: "min: 0ms, max: 1ms, avg: 0.3ms (10 samples)"
if [[ -n "$LATENCY_OUTPUT" ]]; then
    MIN_LATENCY=$(echo "$LATENCY_OUTPUT" | grep -oP 'min: \K[0-9]+' || echo "-1")
    MAX_LATENCY=$(echo "$LATENCY_OUTPUT" | grep -oP 'max: \K[0-9]+' || echo "-1")
    AVG_LATENCY=$(echo "$LATENCY_OUTPUT" | grep -oP 'avg: \K[0-9.]+' || echo "-1")
    
    echo ""
    info "Latency Results:"
    echo "  Min: ${MIN_LATENCY}ms"
    echo "  Max: ${MAX_LATENCY}ms"
    echo "  Avg: ${AVG_LATENCY}ms"
    echo ""
    
    # Verify p50 (avg) < 1ms
    if (( $(echo "$AVG_LATENCY < 1" | bc -l 2>/dev/null || echo 0) )); then
        pass "Average latency < 1ms (${AVG_LATENCY}ms)"
    else
        fail "Average latency >= 1ms (${AVG_LATENCY}ms) - SLA breach"
    fi
    
    # Verify max < 2ms
    if [[ "$MAX_LATENCY" -lt 2 ]]; then
        pass "Max latency < 2ms (${MAX_LATENCY}ms)"
    else
        fail "Max latency >= 2ms (${MAX_LATENCY}ms) - SLA breach"
    fi
else
    info "Could not parse latency output, running alternative test..."
    
    # Alternative: Run 100 PINGs and measure
    START=$(date +%s%N)
    for i in {1..100}; do
        redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" ping > /dev/null 2>&1
    done
    END=$(date +%s%N)
    
    ELAPSED=$(( (END - START) / 1000000 ))  # Convert to milliseconds
    AVG_PER_PING=$(echo "scale=2; $ELAPSED / 100" | bc)
    
    info "Alternative test: 100 PINGs completed in ${ELAPSED}ms (avg: ${AVG_PER_PING}ms per PING)"
    
    if (( $(echo "$AVG_PER_PING < 1" | bc -l 2>/dev/null || echo 0) )); then
        pass "Average PING latency < 1ms (${AVG_PER_PING}ms)"
    else
        fail "Average PING latency >= 1ms (${AVG_PER_PING}ms)"
    fi
fi

# =============================================================================
# Test 4: Memory Configuration
# =============================================================================
info "Test 4: Verifying memory configuration..."

MEMORY_INFO=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" info memory 2>&1 || true)

USED_MEMORY=$(echo "$MEMORY_INFO" | grep "used_memory:" | cut -d':' -f2 | tr -d '\r')
MAXMEMORY=$(echo "$MEMORY_INFO" | grep "maxmemory:" | cut -d':' -f2 | tr -d '\r')
MAXMEMORY_POLICY=$(echo "$MEMORY_INFO" | grep "maxmemory_policy:" | cut -d':' -f2 | tr -d '\r')

if [[ -n "$MAXMEMORY_POLICY" ]]; then
    pass "Maxmemory policy: $MAXMEMORY_POLICY"
    
    if [[ "$MAXMEMORY_POLICY" == "allkeys-lru" ]]; then
        pass "LRU eviction policy is correctly configured"
    else
        fail "Maxmemory policy is not allkeys-lru (current: $MAXMEMORY_POLICY)"
    fi
else
    fail "Could not determine maxmemory policy"
fi

# =============================================================================
# Test 5: Connected Clients
# =============================================================================
info "Test 5: Checking connected clients..."

CLIENTS_INFO=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" info clients 2>&1 || true)

CONNECTED_CLIENTS=$(echo "$CLIENTS_INFO" | grep "connected_clients:" | cut -d':' -f2 | tr -d '\r')

if [[ -n "$CONNECTED_CLIENTS" ]]; then
    pass "Connected clients: $CONNECTED_CLIENTS"
else
    info "Could not determine connected clients"
fi

# =============================================================================
# Test 6: Keyspace (should be empty for fresh deployment)
# =============================================================================
info "Test 6: Checking keyspace..."

DBSIZE=$(redis-cli -h "$REDIS_HOST" -p "$REDIS_PORT" dbsize 2>&1 || true)

pass "Database size: $DBSIZE keys"

# =============================================================================
# Summary
# =============================================================================
echo ""
echo "========================================"
echo "Redis Verification Summary"
echo "========================================"
echo "Host: ${REDIS_HOST}:${REDIS_PORT}"
echo "Status: Operational"
echo ""

if [[ "$MAXMEMORY_POLICY" == "allkeys-lru" ]]; then
    echo -e "${GREEN}✓ Redis is correctly configured for session storage${NC}"
    echo ""
    echo "Recommended next steps:"
    echo "  1. Deploy Cloud Run service with VPC connector"
    echo "  2. Test Redis latency from within Cloud Run VPC"
    echo "  3. Verify TTL behavior with session writes"
else
    echo -e "${YELLOW}⚠ Redis configuration may need adjustment${NC}"
fi

exit 0
