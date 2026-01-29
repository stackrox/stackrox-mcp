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

echo "Starting WireMock..."

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

sleep 2

if ps -p "$WIREMOCK_PID" > /dev/null 2>&1; then
    echo "✓ WireMock started (PID: $WIREMOCK_PID) on http://localhost:8081"
else
    echo "✗ Failed to start WireMock. Check $LOG_FILE"
    rm "$WIREMOCK_DIR/wiremock.pid"
    exit 1
fi
