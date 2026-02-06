#!/bin/bash
set -e

WIREMOCK_DIR="wiremock"
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
    ./scripts/download-wiremock.sh
fi

if [ ! -f "$WIREMOCK_DIR/proto/descriptors/stackrox.pb" ]; then
    ./scripts/generate-proto-descriptors.sh
fi

echo "Starting WireMock with TLS..."

# Generate certificate if not exists
if [ ! -f "$WIREMOCK_DIR/certs/keystore.jks" ]; then
    ./wiremock/generate-cert.sh
fi

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

sleep 2

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
