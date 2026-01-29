#!/bin/bash

PID_FILE="wiremock/wiremock.pid"

if [ ! -f "$PID_FILE" ]; then
    echo "WireMock is not running (no PID file found)"
    exit 0
fi

PID=$(cat "$PID_FILE")

if ps -p "$PID" > /dev/null 2>&1; then
    echo "Stopping WireMock (PID: $PID)..."
    kill "$PID"

    # Wait for process to stop
    for i in {1..10}; do
        if ! ps -p "$PID" > /dev/null 2>&1; then
            break
        fi
        sleep 1
    done

    # Force kill if still running
    if ps -p "$PID" > /dev/null 2>&1; then
        echo "Force killing WireMock..."
        kill -9 "$PID"
    fi

    echo "✓ WireMock stopped"
else
    echo "WireMock process (PID: $PID) not found"
fi

rm "$PID_FILE"
