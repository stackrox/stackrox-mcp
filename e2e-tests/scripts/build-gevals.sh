#!/bin/bash
set -e

cd "$(dirname "$0")/.."

echo "Building gevals from tool dependencies..."
go build -o bin/gevals github.com/genmcp/gevals/cmd/gevals

echo "gevals built successfully: bin/gevals"
./bin/gevals --version
