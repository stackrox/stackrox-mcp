#!/bin/bash

PID_FILE="wiremock/wiremock.pid"

if [ ! -f "$PID_FILE" ]; then
    echo "WireMock is not running"
    exit 0
fi

PID=$(cat "$PID_FILE")

if ps -p "$PID" > /dev/null 2>&1; then
    kill "$PID"
    for i in {1..10}; do
        if ! ps -p "$PID" > /dev/null 2>&1; then
            break
        fi
        sleep 1
    done
    if ps -p "$PID" > /dev/null 2>&1; then
        kill -9 "$PID"
    fi
    echo "✓ WireMock stopped"
else
    echo "WireMock process not found"
fi

rm "$PID_FILE"
