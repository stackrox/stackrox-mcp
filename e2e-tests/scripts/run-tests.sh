#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
E2E_DIR="$(dirname "$SCRIPT_DIR")"
ROOT_DIR="$(dirname "$E2E_DIR")"

# Cleanup function
WIREMOCK_WAS_STARTED=false
cleanup() {
    if [ "$WIREMOCK_WAS_STARTED" = true ]; then
        echo "Stopping WireMock..."
        cd "$ROOT_DIR"
        make mock-stop > /dev/null 2>&1 || true
    fi
}
trap cleanup EXIT

echo "══════════════════════════════════════════════════════════"
echo "  StackRox MCP E2E Testing with mcpchecker"
echo "  Mode: mock (WireMock)"
echo "══════════════════════════════════════════════════════════"
echo ""

# Load environment variables
if [ -f "$E2E_DIR/.env" ]; then
    echo "Loading environment variables from .env..."
    # shellcheck source=/dev/null
    set -a && source "$E2E_DIR/.env" && set +a
fi

# Check if WireMock is already running
if ! curl -skf https://localhost:8081/__admin/mappings > /dev/null 2>&1; then
    echo "Starting WireMock mock service..."
    cd "$ROOT_DIR"
    make mock-start
    WIREMOCK_WAS_STARTED=true
else
    echo "WireMock already running on port 8081"
fi

# Set environment variables for mock mode
export STACKROX_MCP__CENTRAL__URL="localhost:8081"
export STACKROX_MCP__CENTRAL__API_TOKEN="test-token-admin"
export STACKROX_MCP__CENTRAL__INSECURE_SKIP_TLS_VERIFY="true"

# Check OpenAI API key for judge
if [ -z "$OPENAI_API_KEY" ]; then
    echo "Warning: OPENAI_API_KEY is not set (needed for LLM judge)"
    echo "Note: mcpchecker only supports OpenAI-compatible APIs for the judge"
fi

# Build mcpchecker if not present
if [ ! -f "$E2E_DIR/bin/mcpchecker" ]; then
    echo "mcpchecker binary not found. Building..."
    "$SCRIPT_DIR/build-mcpchecker.sh"
    echo ""
fi


# Set agent model (defaults to claude-sonnet-4-5)
export AGENT_MODEL_NAME="${AGENT_MODEL_NAME:-claude-sonnet-4-5}"

# Set judge environment variables (use OpenAI)
export JUDGE_BASE_URL="${JUDGE_BASE_URL:-https://api.openai.com/v1}"
export JUDGE_API_KEY="${JUDGE_API_KEY:-$OPENAI_API_KEY}"
export JUDGE_MODEL_NAME="${JUDGE_MODEL_NAME:-gpt-5-nano}"

echo "Configuration:"
echo "  Central URL: $STACKROX_MCP__CENTRAL__URL (WireMock)"
echo "  Judge: $JUDGE_MODEL_NAME (OpenAI)"
echo "  MCP Server: stackrox-mcp (via go run)"
echo ""

# Run mcpchecker
cd "$E2E_DIR/mcpchecker"
echo "Running mcpchecker tests..."
echo ""

EVAL_FILE="eval.yaml"
echo "Using eval file: $EVAL_FILE"
"$E2E_DIR/bin/mcpchecker" check "$EVAL_FILE"

EXIT_CODE=$?

echo ""
if [ $EXIT_CODE -eq 0 ]; then
    echo "══════════════════════════════════════════════════════════"
    echo "  Tests Completed Successfully!"
    echo "══════════════════════════════════════════════════════════"
else
    echo "══════════════════════════════════════════════════════════"
    echo "  Tests Failed"
    echo "══════════════════════════════════════════════════════════"
    exit $EXIT_CODE
fi
