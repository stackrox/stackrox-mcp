#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
E2E_DIR="$(dirname "$SCRIPT_DIR")"
ROOT_DIR="$(dirname "$E2E_DIR")"

echo "══════════════════════════════════════════════════════════"
echo "  E2E Tests Smoke Test (No Agent Required)"
echo "══════════════════════════════════════════════════════════"
echo ""

# Step 1: Build mcpchecker binary
echo "Step 1/4: Building mcpchecker binary..."
cd "$E2E_DIR/tools"
go build -o ../bin/mcpchecker github.com/mcpchecker/mcpchecker/cmd/mcpchecker
echo "✓ mcpchecker built successfully"
echo ""

# Step 2: Verify MCP server can compile
echo "Step 2/4: Verifying MCP server compiles..."
cd "$ROOT_DIR"
go build -o /tmp/stackrox-mcp-test ./cmd/stackrox-mcp/...
rm -f /tmp/stackrox-mcp-test
echo "✓ MCP server compiles successfully"
echo ""

# Step 3: Validate YAML files
echo "Step 3/4: Validating YAML configuration files..."
cd "$E2E_DIR"

# Check eval.yaml exists and is valid YAML
if [ ! -f "mcpchecker/eval.yaml" ]; then
    echo "✗ Error: mcpchecker/eval.yaml not found"
    exit 1
fi

# Use yq or python to validate YAML (fallback to basic check)
if command -v yq &> /dev/null; then
    yq eval '.' mcpchecker/eval.yaml > /dev/null
    echo "  ✓ eval.yaml is valid"
elif command -v python3 &> /dev/null; then
    python3 -c "import yaml; yaml.safe_load(open('mcpchecker/eval.yaml'))"
    echo "  ✓ eval.yaml is valid"
else
    echo "  ℹ Skipping YAML validation (no yq or python3 available)"
fi

# Check mcp-config.yaml
if [ ! -f "mcpchecker/mcp-config.yaml" ]; then
    echo "✗ Error: mcpchecker/mcp-config.yaml not found"
    exit 1
fi
echo "  ✓ mcp-config.yaml exists"

# Check task files
TASK_COUNT=$(find mcpchecker/tasks -name "*.yaml" -type f | wc -l)
if [ "$TASK_COUNT" -eq 0 ]; then
    echo "✗ Error: No task files found in mcpchecker/tasks/"
    exit 1
fi
echo "  ✓ Found $TASK_COUNT task files"
echo ""

# Step 4: Verify mcpchecker can show help
echo "Step 4/4: Verifying mcpchecker binary works..."
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
echo "  - MCP server: compiles successfully"
echo "  - Configuration files: valid"
echo "  - Task files: $TASK_COUNT found"
echo ""
echo "Note: This smoke test does not run actual agent evaluations."
echo "To run full e2e tests with agents, use: ./scripts/run-tests.sh"
