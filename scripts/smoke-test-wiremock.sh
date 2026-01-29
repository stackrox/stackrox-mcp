#!/bin/bash
set -e

# Smoke test for WireMock mock Central service
# Tests that WireMock starts, MCP connects, and can execute vulnerability queries

echo "=== WireMock Smoke Test ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Cleanup function
cleanup() {
    echo ""
    echo "Cleaning up..."
    ./scripts/stop-mock-central.sh 2>/dev/null || true
    rm -f /tmp/mcp-smoke-test-*.log
}

# Set trap to cleanup on exit
trap cleanup EXIT

# Helper function to run a test
run_test() {
    local test_name="$1"
    local test_command="$2"

    echo -n "Testing: $test_name... "
    if eval "$test_command" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

# Ensure we're in project root
if [ ! -f "go.mod" ]; then
    echo -e "${RED}Error: Must run from project root${NC}"
    exit 1
fi

echo "1. Setting up WireMock..."
echo ""

# Download WireMock JARs if not present
if [ ! -f wiremock/lib/wiremock-standalone.jar ]; then
    echo "Downloading WireMock JARs..."
    ./scripts/download-wiremock.sh
fi

# Generate proto descriptors if not present
if [ ! -f wiremock/proto/descriptors/stackrox.pb ]; then
    echo "Generating proto descriptors..."
    ./scripts/generate-proto-descriptors.sh
fi

# Create __files symlink if not present
if [ ! -L wiremock/__files ]; then
    ln -s fixtures wiremock/__files
fi

echo "2. Starting WireMock service..."
echo ""

./scripts/start-mock-central.sh

# Wait for WireMock to be ready
sleep 3

run_test "WireMock is running" "make mock-status | grep -q 'running'" || true

echo ""
echo "3. Testing WireMock endpoints..."
echo ""

# Test admin API
run_test "Admin API responds" "curl -sf http://localhost:8081/__admin/mappings > /dev/null" || true

# Test authentication rejection (should return code 16 = Unauthenticated)
run_test "Rejects missing auth token" "curl -s -X POST -H 'Content-Type: application/json' -d '{}' http://localhost:8081/v1.DeploymentService/ListDeployments | grep -q '\"code\":16'" || true

# Test CVE query with valid token
run_test "Returns deployments for CVE-2021-44228" "curl -sf -X POST -H 'Content-Type: application/json' -H 'Authorization: Bearer test-token-admin' -d '{\"query\":{\"query\":\"CVE:\\\"CVE-2021-44228\\\"\"}}' http://localhost:8081/v1.DeploymentService/ListDeployments | grep -q 'dep-123-log4j'" || true

# Test empty query returns empty results
run_test "Returns empty for unknown CVE" "curl -sf -X POST -H 'Content-Type: application/json' -H 'Authorization: Bearer test-token-admin' -d '{}' http://localhost:8081/v1.DeploymentService/ListDeployments | grep -q '\"deployments\": \[\]'" || true

echo ""
echo "4. Verifying MCP can connect to WireMock..."
echo ""

# Build MCP server if binary doesn't exist
if [ ! -f ./stackrox-mcp ]; then
    echo "Building MCP server..."
    make build
fi

# Quick test: just verify MCP can start with WireMock config
export STACKROX_MCP__SERVER__TYPE=stdio
export STACKROX_MCP__CENTRAL__URL=localhost:8081
export STACKROX_MCP__CENTRAL__AUTH_TYPE=static
export STACKROX_MCP__CENTRAL__API_TOKEN=test-token-admin
export STACKROX_MCP__CENTRAL__INSECURE_SKIP_TLS_VERIFY=true
export STACKROX_MCP__TOOLS__VULNERABILITY__ENABLED=true

# Create simple initialize request
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' > /tmp/mcp-smoke-test-input.json

# Run MCP server briefly to verify it starts
timeout 3 ./stackrox-mcp < /tmp/mcp-smoke-test-input.json > /tmp/mcp-smoke-test-stdout.log 2>/tmp/mcp-smoke-test-stderr.log || true

# Verify MCP started successfully
run_test "MCP server starts with WireMock config" "grep -q 'Starting StackRox MCP server' /tmp/mcp-smoke-test-stderr.log" || true

# Verify tools are registered
run_test "MCP registers vulnerability tools" "grep -q 'get_deployments_for_cve' /tmp/mcp-smoke-test-stderr.log" || true

echo ""
echo "=== Test Summary ==="
echo -e "Tests passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Tests failed: ${RED}$TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All smoke tests passed!${NC}"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    echo ""
    echo "Check logs for details:"
    echo "  - WireMock: wiremock/wiremock.log"
    echo "  - MCP stdout: /tmp/mcp-smoke-test-stdout.log"
    echo "  - MCP stderr: /tmp/mcp-smoke-test-stderr.log"
    exit 1
fi
