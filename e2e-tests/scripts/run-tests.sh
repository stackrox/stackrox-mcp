#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
E2E_DIR="$(dirname "$SCRIPT_DIR")"

echo "══════════════════════════════════════════════════════════"
echo "  StackRox MCP E2E Testing with Gevals"
echo "══════════════════════════════════════════════════════════"
echo ""

# Load environment variables
if [ -f "$E2E_DIR/.env" ]; then
    echo "Loading environment variables from .env..."
    set -a && source .env && set +a
else
    echo "Warning: .env file not found"
fi

if [ -z "$STACKROX_MCP__CENTRAL__API_TOKEN" ]; then
    echo "Error: STACKROX_MCP__CENTRAL__API_TOKEN is not set"
    echo "Please set it in .env file or export it in your environment"
    exit 1
fi

# Check OpenAI API key for judge
if [ -z "$OPENAI_API_KEY" ]; then
    echo "Warning: OPENAI_API_KEY is not set (needed for LLM judge)"
    echo "Note: gevals only supports OpenAI-compatible APIs for the judge"
fi

# Build gevals if not present
if [ ! -f "$E2E_DIR/bin/gevals" ]; then
    echo "Gevals binary not found. Building..."
    "$SCRIPT_DIR/build-gevals.sh"
    echo ""
fi


# Set agent model (defaults to claude-sonnet-4-5)
export AGENT_MODEL_NAME="${AGENT_MODEL_NAME:-claude-sonnet-4-5}"

# Set judge environment variables (use OpenAI)
export JUDGE_BASE_URL="${JUDGE_BASE_URL:-https://api.openai.com/v1}"
export JUDGE_API_KEY="${JUDGE_API_KEY:-$OPENAI_API_KEY}"
export JUDGE_MODEL_NAME="${JUDGE_MODEL_NAME:-gpt-5-nano}"

echo "Configuration:"
echo "  Judge: $JUDGE_MODEL_NAME (OpenAI)"
echo "  MCP Server: stackrox-mcp (via go run)"
echo ""

# Run gevals
cd "$E2E_DIR/gevals"
echo "Running gevals tests..."
echo ""

"$E2E_DIR/bin/gevals" eval eval.yaml

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
