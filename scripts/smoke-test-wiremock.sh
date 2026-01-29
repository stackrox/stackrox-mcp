#!/bin/bash
set -e

# Smoke test for WireMock mock Central service
# Verifies that WireMock is properly configured and can communicate with MCP server

echo "=== WireMock Smoke Test ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Helper function to run a test
run_test() {
    local test_name="$1"
    local test_command="$2"

    echo -n "Testing: $test_name... "
    if eval "$test_command" > /dev/null 2>&1; then
        echo -e "${GREEN}✓ PASS${NC}"
        ((TESTS_PASSED++))
        return 0
    else
        echo -e "${RED}✗ FAIL${NC}"
        ((TESTS_FAILED++))
        return 1
    fi
}

# Ensure we're in project root
if [ ! -f "go.mod" ]; then
    echo -e "${RED}Error: Must run from project root${NC}"
    exit 1
fi

echo "1. Checking required files..."
echo ""

# Check scripts exist
run_test "download-wiremock.sh exists" "[ -f scripts/download-wiremock.sh ]"
run_test "generate-proto-descriptors.sh exists" "[ -f scripts/generate-proto-descriptors.sh ]"
run_test "start-mock-central.sh exists" "[ -f scripts/start-mock-central.sh ]"
run_test "stop-mock-central.sh exists" "[ -f scripts/stop-mock-central.sh ]"

# Check directories exist
run_test "wiremock/mappings exists" "[ -d wiremock/mappings ]"
run_test "wiremock/fixtures exists" "[ -d wiremock/fixtures ]"

# Check mapping files
run_test "auth mapping exists" "[ -f wiremock/mappings/auth.json ]"
run_test "deployments mapping exists" "[ -f wiremock/mappings/deployments.json ]"
run_test "images mapping exists" "[ -f wiremock/mappings/images.json ]"
run_test "nodes mapping exists" "[ -f wiremock/mappings/nodes.json ]"
run_test "clusters mapping exists" "[ -f wiremock/mappings/clusters.json ]"

# Check fixture files
run_test "log4j_cve fixture exists" "[ -f wiremock/fixtures/deployments/log4j_cve.json ]"
run_test "empty deployment fixture exists" "[ -f wiremock/fixtures/deployments/empty.json ]"

echo ""
echo "2. Setting up WireMock..."
echo ""

# Download WireMock JARs if not present
if [ ! -f wiremock/lib/wiremock-standalone.jar ]; then
    echo "Downloading WireMock JARs..."
    ./scripts/download-wiremock.sh
fi
run_test "WireMock standalone JAR exists" "[ -f wiremock/lib/wiremock-standalone.jar ]"
run_test "WireMock gRPC extension exists" "[ -f wiremock/lib/wiremock-grpc-extension.jar ]"

# Generate proto descriptors if not present
if [ ! -f wiremock/proto/descriptors/stackrox.pb ]; then
    echo "Generating proto descriptors..."
    ./scripts/generate-proto-descriptors.sh
fi
run_test "Proto descriptors exist" "[ -f wiremock/proto/descriptors/stackrox.pb ]"

# Create __files symlink if not present
if [ ! -L wiremock/__files ]; then
    ln -s fixtures wiremock/__files
fi
run_test "__files symlink exists" "[ -L wiremock/__files ]"

echo ""
echo "3. Starting WireMock service..."
echo ""

# Start WireMock
./scripts/start-mock-central.sh

# Wait for WireMock to be ready
sleep 3

run_test "WireMock process is running" "make mock-status | grep -q 'running'"

echo ""
echo "4. Testing WireMock endpoints..."
echo ""

# Test admin API
run_test "Admin API responds" "curl -sf http://localhost:8081/__admin/mappings > /dev/null"
run_test "Has loaded mappings" "curl -sf http://localhost:8081/__admin/mappings | grep -q '\"mappings\"'"

# Test authentication rejection
run_test "Rejects missing auth token" "curl -sf -X POST -H 'Content-Type: application/json' -d '{}' http://localhost:8081/v1.DeploymentService/ListDeployments | grep -q '\"code\":16'"

# Test authentication acceptance and CVE query
run_test "Accepts valid auth token" "curl -sf -X POST -H 'Content-Type: application/json' -H 'Authorization: Bearer test-token-admin' -d '{\"query\":{\"query\":\"CVE:\\\"CVE-2021-44228\\\"\"}}' http://localhost:8081/v1.DeploymentService/ListDeployments | grep -q 'deployments'"

# Test specific CVE returns correct data
DEPLOYMENT_COUNT=$(curl -sf -X POST -H 'Content-Type: application/json' -H 'Authorization: Bearer test-token-admin' -d '{"query":{"query":"CVE:\"CVE-2021-44228\""}}' http://localhost:8081/v1.DeploymentService/ListDeployments | grep -o '"id"' | wc -l)
run_test "CVE-2021-44228 returns 3 deployments" "[ '$DEPLOYMENT_COUNT' -eq 3 ]"

# Test empty query returns empty
EMPTY_COUNT=$(curl -sf -X POST -H 'Content-Type: application/json' -H 'Authorization: Bearer test-token-admin' -d '{}' http://localhost:8081/v1.DeploymentService/ListDeployments | grep -o '"id"' | wc -l)
run_test "Empty query returns no deployments" "[ '$EMPTY_COUNT' -eq 0 ]"

echo ""
echo "5. Testing MCP server integration (optional)..."
echo ""

# Build MCP server if binary doesn't exist
if [ ! -f ./stackrox-mcp ]; then
    echo "Building MCP server..."
    make build
fi
run_test "MCP binary exists" "[ -f ./stackrox-mcp ]"

# Test MCP can connect to WireMock (basic smoke test)
# Create a simple test that just verifies the server can start
export STACKROX_MCP__SERVER__TYPE=stdio
export STACKROX_MCP__CENTRAL__URL=localhost:8081
export STACKROX_MCP__CENTRAL__AUTH_TYPE=static
export STACKROX_MCP__CENTRAL__API_TOKEN=test-token-admin
export STACKROX_MCP__CENTRAL__INSECURE_SKIP_TLS_VERIFY=true
export STACKROX_MCP__TOOLS__VULNERABILITY__ENABLED=true

# Test that MCP server can start and respond to initialize
MCP_TEST_INPUT='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}'
echo "$MCP_TEST_INPUT" | timeout 5 ./stackrox-mcp 2>/tmp/mcp-smoke-test-stderr.log > /tmp/mcp-smoke-test-stdout.log || true

# Check if MCP started successfully (logs should contain "Starting")
run_test "MCP server starts successfully" "grep -q 'Starting StackRox MCP server' /tmp/mcp-smoke-test-stderr.log"
run_test "MCP registers vulnerability tools" "grep -q 'get_deployments_for_cve' /tmp/mcp-smoke-test-stderr.log"

# Cleanup temp files
rm -f /tmp/mcp-smoke-test-*.log

echo ""
echo "6. Cleaning up..."
echo ""

# Stop WireMock
./scripts/stop-mock-central.sh
run_test "WireMock stopped successfully" "! make mock-status 2>/dev/null | grep -q 'running'"

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
    exit 1
fi
