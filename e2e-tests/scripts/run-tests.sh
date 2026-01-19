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
    export $(grep -v '^#' "$E2E_DIR/.env" | grep -v '^$' | xargs)
else
    echo "Warning: .env file not found"
fi

# Check required environment variables
if [ -z "$OPENAI_API_KEY" ]; then
    echo "Error: OPENAI_API_KEY is not set"
    echo "Please set it in .env file or export it in your environment"
    exit 1
fi

if [ -z "$STACKROX_MCP__CENTRAL__API_TOKEN" ]; then
    echo "Error: STACKROX_MCP__CENTRAL__API_TOKEN is not set"
    echo "Please set it in .env file or export it in your environment"
    exit 1
fi

# Build gevals if not present
if [ ! -f "$E2E_DIR/bin/gevals" ]; then
    echo "Gevals binary not found. Building..."
    "$SCRIPT_DIR/build-gevals.sh"
    echo ""
fi

# Set judge environment variables (use same OpenAI key)
export JUDGE_BASE_URL="${JUDGE_BASE_URL:-https://api.openai.com/v1}"
export JUDGE_API_KEY="${JUDGE_API_KEY:-$OPENAI_API_KEY}"
export JUDGE_MODEL_NAME="${JUDGE_MODEL_NAME:-gpt-4o}"

# Set agent environment variables
export MODEL_BASE_URL="${MODEL_BASE_URL:-https://api.openai.com/v1}"
export MODEL_KEY="${MODEL_KEY:-$OPENAI_API_KEY}"

echo "Configuration:"
echo "  Agent Model: gpt-4o"
echo "  Judge Model: $JUDGE_MODEL_NAME"
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
