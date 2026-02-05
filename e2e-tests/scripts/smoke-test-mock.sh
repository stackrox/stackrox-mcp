#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
E2E_DIR="$(dirname "$SCRIPT_DIR")"
ROOT_DIR="$(dirname "$E2E_DIR")"

echo "══════════════════════════════════════════════════════════"
echo "  WireMock Integration Smoke Test"
echo "══════════════════════════════════════════════════════════"
echo ""

# Start WireMock
echo "1. Starting WireMock..."
cd "$ROOT_DIR"
make mock-stop > /dev/null 2>&1 || true
make mock-start

# Wait for WireMock to be ready
echo ""
echo "2. Waiting for WireMock to be ready..."
for i in {1..10}; do
    if nc -z localhost 8081 2>/dev/null; then
        echo "✓ WireMock is ready"
        break
    fi
    sleep 1
done

# Test MCP server can connect
echo ""
echo "3. Testing MCP server connection..."
cd "$ROOT_DIR"

# Run MCP server and test a simple tool call
timeout 10 bash -c '
export STACKROX_MCP__CENTRAL__URL=localhost:8081
export STACKROX_MCP__CENTRAL__API_TOKEN=test-token-admin
export STACKROX_MCP__CENTRAL__INSECURE_SKIP_TLS_VERIFY=true

# Start MCP server in background
go run ./cmd/stackrox-mcp --config e2e-tests/stackrox-mcp-e2e-config.yaml 2>&1 | grep -m1 "Tools registration complete" && echo "✓ MCP server started successfully"
' || echo "✗ MCP server failed to start"

# Test with grpcurl
echo ""
echo "4. Testing WireMock responses..."

# Test auth (should accept test-token-admin)
AUTH_RESULT=$(grpcurl -plaintext -H "Authorization: Bearer test-token-admin" \
  -d '{}' localhost:8081 v1.ClustersService/GetClusters 2>&1 || true)

if echo "$AUTH_RESULT" | grep -q "clusters"; then
    echo "✓ Authentication works"
else
    echo "✗ Authentication failed"
    echo "$AUTH_RESULT"
fi

# Test CVE query
CVE_RESULT=$(grpcurl -plaintext -H "Authorization: Bearer test-token-admin" \
  -d '{"query": {"query": "CVE-2021-44228"}}' \
  localhost:8081 v1.DeploymentService/ListDeployments 2>&1 || true)

if echo "$CVE_RESULT" | grep -q "deployments"; then
    echo "✓ CVE query returns data"
else
    echo "✗ CVE query failed"
    echo "$CVE_RESULT"
fi

# Cleanup
echo ""
echo "5. Cleaning up..."
cd "$ROOT_DIR"
make mock-stop

echo ""
echo "══════════════════════════════════════════════════════════"
echo "  Smoke Test Complete!"
echo "══════════════════════════════════════════════════════════"
