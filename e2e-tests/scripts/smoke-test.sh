#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
E2E_DIR="$(dirname "$SCRIPT_DIR")"

echo "══════════════════════════════════════════════════════════"
echo "  E2E Tests Smoke Test"
echo "══════════════════════════════════════════════════════════"
echo ""

# Step 1: Build mcpchecker binary
echo "Step 1/2: Building mcpchecker binary..."
cd "$E2E_DIR/tools"
go build -o ../bin/mcpchecker github.com/mcpchecker/mcpchecker/cmd/mcpchecker
echo "✓ mcpchecker built successfully"
echo ""

# Step 2: Verify mcpchecker binary works
echo "Step 2/2: Verifying mcpchecker binary works..."
"$E2E_DIR/bin/mcpchecker" --version > /dev/null 2>&1 || true
"$E2E_DIR/bin/mcpchecker" help > /dev/null
echo "✓ mcpchecker binary works correctly"
echo ""

echo "══════════════════════════════════════════════════════════"
echo "  ✓ All Smoke Tests Passed!"
echo "══════════════════════════════════════════════════════════"
echo ""
echo "Summary:"
echo "  - mcpchecker binary: built and working"
echo ""
echo "Note: This smoke test does not run actual agent evaluations."
echo "To run full e2e tests with agents, use: ./scripts/run-tests.sh"
