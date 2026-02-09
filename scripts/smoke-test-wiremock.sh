#!/bin/bash
set -e

echo "=== WireMock Smoke Test ==="

GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

TESTS_PASSED=0
TESTS_FAILED=0

cleanup() {
    echo ""
    echo "Cleaning up..."
    ./scripts/stop-mock-central.sh 2>/dev/null || true
    rm -f /tmp/mcp-smoke-test-*.log /tmp/mcp-smoke-test-*.json
}

trap cleanup EXIT

run_test() {
    local test_name="$1"
    local test_command="$2"

    echo -n "Testing: $test_name... "
    if eval "$test_command" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
    else
        echo -e "${RED}✗${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        return 1
    fi
}

if [ ! -f "go.mod" ]; then
    echo "Error: Run from project root"
    exit 1
fi

echo ""
echo "Setup..."

if [ ! -f wiremock/lib/wiremock-standalone.jar ]; then
    ./scripts/download-wiremock.sh
fi

if [ ! -f wiremock/proto/descriptors/stackrox.pb ]; then
    ./scripts/generate-proto-descriptors.sh
fi

if [ ! -L wiremock/__files ]; then
    ln -s fixtures wiremock/__files
fi

echo ""
echo "Starting WireMock..."
./scripts/start-mock-central.sh
sleep 3

echo ""
run_test "WireMock is running" "make mock-status | grep -q 'running'" || true
run_test "Admin API responds" "curl -skf https://localhost:8081/__admin/mappings > /dev/null" || true
run_test "Rejects missing auth" "curl -sk -X POST -H 'Content-Type: application/json' -d '{}' https://localhost:8081/v1.DeploymentService/ListDeployments | grep -q '\"code\":16'" || true
run_test "Returns CVE-2021-44228 data" "curl -skf -X POST -H 'Content-Type: application/json' -H 'Authorization: Bearer test-token-admin' -d '{\"query\":{\"query\":\"CVE:\\\"CVE-2021-44228\\\"\"}}' https://localhost:8081/v1.DeploymentService/ListDeployments | grep -q 'dep-004'" || true
run_test "Returns empty for unknown CVE" "curl -skf -X POST -H 'Content-Type: application/json' -H 'Authorization: Bearer test-token-admin' -d '{}' https://localhost:8081/v1.DeploymentService/ListDeployments | grep -q '\"deployments\": \[\]'" || true

echo ""
echo "Testing MCP integration..."

if [ ! -f ./stackrox-mcp ]; then
    make build > /dev/null 2>&1
fi

export STACKROX_MCP__SERVER__TYPE=stdio
export STACKROX_MCP__CENTRAL__URL=localhost:8081
export STACKROX_MCP__CENTRAL__AUTH_TYPE=static
export STACKROX_MCP__CENTRAL__API_TOKEN=test-token-admin
export STACKROX_MCP__CENTRAL__INSECURE_SKIP_TLS_VERIFY=true
export STACKROX_MCP__TOOLS__VULNERABILITY__ENABLED=true

echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' > /tmp/mcp-smoke-test-input.json

timeout 3 ./stackrox-mcp < /tmp/mcp-smoke-test-input.json > /tmp/mcp-smoke-test-stdout.log 2>/tmp/mcp-smoke-test-stderr.log || true

run_test "MCP starts with WireMock" "grep -q 'Starting StackRox MCP server' /tmp/mcp-smoke-test-stderr.log" || true
run_test "MCP registers tools" "grep -q 'get_deployments_for_cve' /tmp/mcp-smoke-test-stderr.log" || true

echo ""
echo "=== Results ==="
echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
echo -e "Failed: ${RED}$TESTS_FAILED${NC}"

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed${NC}"
    exit 0
else
    echo -e "${RED}✗ Tests failed${NC}"
    exit 1
fi
