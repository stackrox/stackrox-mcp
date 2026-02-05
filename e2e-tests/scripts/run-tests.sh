#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
E2E_DIR="$(dirname "$SCRIPT_DIR")"
ROOT_DIR="$(dirname "$E2E_DIR")"

# Parse command-line arguments
MODE="${STACKROX_E2E_MODE:-real}"
while [[ $# -gt 0 ]]; do
    case $1 in
        --mock)
            MODE="mock"
            shift
            ;;
        --real)
            MODE="real"
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--mock|--real]"
            exit 1
            ;;
    esac
done

echo "══════════════════════════════════════════════════════════"
echo "  StackRox MCP E2E Testing with mcpchecker"
echo "  Mode: $MODE"
echo "══════════════════════════════════════════════════════════"
echo ""

# Load environment variables
if [ -f "$E2E_DIR/.env" ]; then
    echo "Loading environment variables from .env..."
    # shellcheck source=/dev/null
    set -a && source "$E2E_DIR/.env" && set +a
else
    echo "Warning: .env file not found"
fi

# Configure based on mode
if [ "$MODE" = "mock" ]; then
    echo "Configuring for mock mode (WireMock)..."

    # Check if WireMock is already running
    WIREMOCK_WAS_STARTED=false
    if ! nc -z localhost 8081 2>/dev/null; then
        echo "Starting WireMock mock service..."
        cd "$ROOT_DIR"
        make mock-start
        WIREMOCK_WAS_STARTED=true

        # Wait for WireMock to start
        echo "Waiting for WireMock to be ready..."
        # shellcheck disable=SC2034
        for _i in {1..30}; do
            if nc -z localhost 8081 2>/dev/null; then
                echo "WireMock is ready!"
                break
            fi
            sleep 1
        done

        if ! nc -z localhost 8081 2>/dev/null; then
            echo "Error: WireMock failed to start"
            exit 1
        fi
    else
        echo "WireMock already running on port 8081"
    fi

    # Set environment variables for mock mode
    export STACKROX_MCP__CENTRAL__URL="localhost:8081"
    export STACKROX_MCP__CENTRAL__API_TOKEN="test-token-admin"
    export STACKROX_MCP__CENTRAL__INSECURE_SKIP_TLS_VERIFY="true"

    # Cleanup function for WireMock (only stop if we started it)
    cleanup_wiremock() {
        if [ "$WIREMOCK_WAS_STARTED" = true ]; then
            echo "Stopping WireMock..."
            cd "$ROOT_DIR"
            make mock-stop > /dev/null 2>&1 || true
        fi
    }
    trap cleanup_wiremock EXIT

elif [ "$MODE" = "real" ]; then
    echo "Configuring for real mode (staging.demo.stackrox.com)..."

    if [ -z "$STACKROX_MCP__CENTRAL__API_TOKEN" ]; then
        echo "Error: STACKROX_MCP__CENTRAL__API_TOKEN is not set"
        echo "Please set it in .env file or export it in your environment"
        exit 1
    fi
else
    echo "Error: Invalid mode '$MODE'. Use --mock or --real"
    exit 1
fi

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
echo "  Mode: $MODE"
echo "  Central URL: ${STACKROX_MCP__CENTRAL__URL:-<from config>}"
echo "  Judge: $JUDGE_MODEL_NAME (OpenAI)"
echo "  MCP Server: stackrox-mcp (via go run)"
echo ""

# Run mcpchecker
cd "$E2E_DIR/mcpchecker"
echo "Running mcpchecker tests..."
echo ""

"$E2E_DIR/bin/mcpchecker" check eval.yaml

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
