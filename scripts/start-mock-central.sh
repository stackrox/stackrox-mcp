#!/bin/bash
set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

WIREMOCK_DIR="$PROJECT_ROOT/wiremock"
PID_FILE="$WIREMOCK_DIR/wiremock.pid"
LOG_FILE="$WIREMOCK_DIR/wiremock.log"

if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE")
    if ps -p "$PID" > /dev/null 2>&1; then
        echo "WireMock is already running (PID: $PID)"
        exit 0
    fi
    rm "$PID_FILE"
fi

if [ ! -f "$WIREMOCK_DIR/lib/wiremock-standalone.jar" ]; then
    "$PROJECT_ROOT/scripts/download-wiremock.sh"
fi

if [ ! -f "$WIREMOCK_DIR/proto/descriptors/stackrox.dsc" ]; then
    "$PROJECT_ROOT/scripts/generate-proto-descriptors.sh"
fi

# Create __files symlink if needed (WireMock expects this)
if [ ! -L "$WIREMOCK_DIR/__files" ]; then
    cd "$WIREMOCK_DIR"
    ln -s fixtures __files
    cd "$PROJECT_ROOT"
fi

echo "Starting WireMock with TLS..."

# Use subshell to avoid having to cd back
(
cd "$WIREMOCK_DIR"
java -cp "lib/wiremock-standalone.jar:lib/wiremock-grpc-extension.jar" \
  wiremock.Run \
  --port 8080 \
  --https-port 8081 \
  --https-keystore certs/keystore.jks \
  --keystore-password wiremock \
  --key-manager-password wiremock \
  --keystore-type JKS \
  --global-response-templating \
  --verbose \
  --root-dir . \
  > wiremock.log 2>&1 &

WIREMOCK_PID=$!
echo $WIREMOCK_PID > wiremock.pid
)

# Wait for WireMock to be ready
echo "Waiting for WireMock to be ready..."
MAX_WAIT=30
WAITED=0
while [ $WAITED -lt $MAX_WAIT ]; do
    if curl -skf https://localhost:8081/__admin/mappings > /dev/null 2>&1; then
        break
    fi
    sleep 1
    WAITED=$((WAITED + 1))
done

if [ $WAITED -eq $MAX_WAIT ]; then
    echo "✗ WireMock failed to start within ${MAX_WAIT}s. Check $LOG_FILE"
    exit 1
fi

# Read PID from file (written inside subshell)
if [ -f "$PID_FILE" ]; then
    WIREMOCK_PID=$(cat "$PID_FILE")
    if ps -p "$WIREMOCK_PID" > /dev/null 2>&1; then
        echo "✓ WireMock started (PID: $WIREMOCK_PID) on https://localhost:8081"
    else
        echo "✗ Failed to start WireMock. Check $LOG_FILE"
        rm "$PID_FILE"
        exit 1
    fi
else
    echo "✗ Failed to start WireMock. PID file not created. Check $LOG_FILE"
    exit 1
fi
