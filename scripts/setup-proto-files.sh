#!/bin/bash
set -e

# Setup script to copy proto files from stackrox repo to wiremock directory
# This is needed for generating proto descriptors for WireMock gRPC support

echo "=== Setting up proto files for WireMock ==="
echo ""

# Detect stackrox repo location
STACKROX_REPO=""

# Check common locations
if [ -d "../stackrox" ]; then
    STACKROX_REPO="../stackrox"
elif [ -d "../../stackrox" ]; then
    STACKROX_REPO="../../stackrox"
elif [ -n "$STACKROX_REPO_PATH" ]; then
    STACKROX_REPO="$STACKROX_REPO_PATH"
fi

# Prompt if not found
if [ -z "$STACKROX_REPO" ] || [ ! -d "$STACKROX_REPO" ]; then
    echo "Error: StackRox repository not found."
    echo ""
    echo "Please specify the path to the stackrox repository:"
    echo "  export STACKROX_REPO_PATH=/path/to/stackrox"
    echo "  ./scripts/setup-proto-files.sh"
    echo ""
    echo "Or clone it to a sibling directory:"
    echo "  cd .."
    echo "  git clone https://github.com/stackrox/stackrox"
    echo "  cd stackrox-mcp"
    echo "  ./scripts/setup-proto-files.sh"
    exit 1
fi

echo "Using StackRox repository: $STACKROX_REPO"
echo ""

# Create destination directories
echo "Creating destination directories..."
mkdir -p wiremock/proto/stackrox
mkdir -p wiremock/proto/googleapis

# Copy proto files
echo "Copying proto files..."
cp -r "$STACKROX_REPO/proto/"* wiremock/proto/stackrox/
echo "✓ Copied stackrox proto files"

cp -r "$STACKROX_REPO/third_party/googleapis/"* wiremock/proto/googleapis/
echo "✓ Copied googleapis proto files"

# Copy scanner proto files
mkdir -p wiremock/proto/stackrox/scanner/api/v1
if [ -d "$STACKROX_REPO/qa-tests-backend/src/main/proto/scanner/api/v1" ]; then
    cp "$STACKROX_REPO/qa-tests-backend/src/main/proto/scanner/api/v1/"*.proto wiremock/proto/stackrox/scanner/api/v1/
    echo "✓ Copied scanner proto files"
else
    echo "⚠ Warning: Scanner proto files not found (optional)"
fi

echo ""
echo "=== Proto files setup complete ==="
echo ""
echo "Next steps:"
echo "  1. Run: ./scripts/generate-proto-descriptors.sh"
echo "  2. Run: make mock-start"
