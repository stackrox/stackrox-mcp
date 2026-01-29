#!/bin/bash
set -e

echo "Setting up proto files from go modules..."

# Ensure go modules are downloaded
go mod download

# Discover rox module location using go list
ROX_DIR=$(go list -f '{{.Dir}}' -m github.com/stackrox/rox)

if [ -z "$ROX_DIR" ]; then
    echo "Error: github.com/stackrox/rox module not found"
    echo "Run: go mod download"
    exit 1
fi

echo "Using proto files from: $ROX_DIR"

# Create target directories
mkdir -p wiremock/proto/stackrox wiremock/proto/googleapis

# Copy proto files from rox module
# Note: Files from go mod cache are read-only, so we copy and chmod
cp -r "$ROX_DIR/proto/"* wiremock/proto/stackrox/
cp -r "$ROX_DIR/third_party/googleapis/"* wiremock/proto/googleapis/

# Copy scanner protos from scanner module (following stackrox pattern)
SCANNER_DIR=$(go list -f '{{.Dir}}' -m github.com/stackrox/scanner)
if [ -n "$SCANNER_DIR" ] && [ -d "$SCANNER_DIR/proto/scanner" ]; then
    echo "Using scanner proto files from: $SCANNER_DIR"
    cp -r "$SCANNER_DIR/proto/scanner" wiremock/proto/stackrox/
fi

# Make files writable (go mod cache files are read-only)
chmod -R u+w wiremock/proto/

echo "✓ Proto files copied from go mod cache"
echo "Next: ./scripts/generate-proto-descriptors.sh"
