#!/bin/bash
set -e

E2E_DIR="$(cd "$(dirname "$0")/.." && pwd)"

echo "Building mcpchecker from tool dependencies..."
cd "$E2E_DIR/tools"
go build -o ../bin/mcpchecker github.com/mcpchecker/mcpchecker/cmd/mcpchecker

echo "mcpchecker built successfully"
cd "$E2E_DIR"
./bin/mcpchecker help
