#!/bin/bash
set -e

WIREMOCK_DIR="wiremock"
PID_FILE="$WIREMOCK_DIR/wiremock.pid"
LOG_FILE="$WIREMOCK_DIR/wiremock.log"

# Check if already running
if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE")
    if ps -p "$PID" > /dev/null 2>&1; then
        echo "WireMock is already running (PID: $PID)"
        exit 0
    else
        rm "$PID_FILE"
    fi
fi

# Download WireMock JARs if needed
if [ ! -f "$WIREMOCK_DIR/lib/wiremock-standalone.jar" ]; then
    echo "WireMock JARs not found. Downloading..."
    ./scripts/download-wiremock.sh
fi

# Generate proto descriptors if needed
if [ ! -f "$WIREMOCK_DIR/proto/descriptors/stackrox.pb" ]; then
    echo "Proto descriptors not found. Generating..."
    ./scripts/generate-proto-descriptors.sh
fi

echo "Starting WireMock Mock Central..."

# Start WireMock with gRPC support
cd "$WIREMOCK_DIR"
java -cp "lib/wiremock-standalone.jar:lib/wiremock-grpc-extension.jar" \
  wiremock.Run \
  --port 8081 \
  --global-response-templating \
  --verbose \
  --root-dir . \
  > wiremock.log 2>&1 &

WIREMOCK_PID=$!
echo $WIREMOCK_PID > wiremock.pid
cd ..

# Wait for WireMock to start
sleep 2

if ps -p "$WIREMOCK_PID" > /dev/null 2>&1; then
    echo "✓ WireMock Mock Central started (PID: $WIREMOCK_PID)"
    echo "✓ HTTP/gRPC endpoint: localhost:8081"
    echo "✓ Admin API: http://localhost:8081/__admin"
    echo "✓ Logs: $LOG_FILE"
    echo ""
    echo "To connect MCP server to mock:"
    echo "  export STACKROX_MCP__SERVER__TYPE=stdio"
    echo "  export STACKROX_MCP__CENTRAL__URL=localhost:8081"
    echo "  export STACKROX_MCP__CENTRAL__AUTH_TYPE=static"
    echo "  export STACKROX_MCP__CENTRAL__API_TOKEN=test-token-admin"
    echo "  export STACKROX_MCP__CENTRAL__INSECURE_SKIP_TLS_VERIFY=true"
    echo "  export STACKROX_MCP__TOOLS__VULNERABILITY__ENABLED=true"
else
    echo "✗ Failed to start WireMock. Check $LOG_FILE for details."
    rm "$WIREMOCK_DIR/wiremock.pid"
    exit 1
fi
