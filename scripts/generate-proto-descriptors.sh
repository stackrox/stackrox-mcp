#!/bin/bash
set -e

# Output directory for descriptors
DESCRIPTOR_DIR="wiremock/proto/descriptors"
mkdir -p "$DESCRIPTOR_DIR"

# Use local proto files (copied from stackrox repo)
ROX_PROTO_PATH="wiremock/proto/stackrox"
GOOGLEAPIS_PATH="wiremock/proto/googleapis"

# Check if local proto files exist
if [ ! -d "$ROX_PROTO_PATH" ]; then
    echo "Error: Local proto files not found at $ROX_PROTO_PATH"
    echo ""
    echo "To set up proto files, run from the stackrox-mcp directory:"
    echo "  cp -r ../stackrox/proto/* wiremock/proto/stackrox/"
    echo "  cp -r ../stackrox/third_party/googleapis/* wiremock/proto/googleapis/"
    exit 1
fi

echo "Using local proto files from: $ROX_PROTO_PATH"

# Check if protoc is installed
if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed."
    echo "Install protoc from: https://grpc.io/docs/protoc-installation/"
    exit 1
fi

echo "Generating proto descriptors..."

# Generate descriptor set for StackRox services
protoc \
  --descriptor_set_out="$DESCRIPTOR_DIR/stackrox.pb" \
  --include_imports \
  --proto_path="$ROX_PROTO_PATH" \
  --proto_path="$GOOGLEAPIS_PATH" \
  api/v1/deployment_service.proto \
  api/v1/image_service.proto \
  api/v1/node_service.proto \
  api/v1/cluster_service.proto \
  2>&1 || {
    echo "Error: Failed to generate proto descriptors."
    echo "Make sure the proto files exist at: $ROX_PROTO_PATH"
    exit 1
  }

echo "✓ Proto descriptors generated at $DESCRIPTOR_DIR/stackrox.pb"

# Copy to grpc directory for WireMock
GRPC_DIR="wiremock/grpc"
mkdir -p "$GRPC_DIR"
cp "$DESCRIPTOR_DIR/stackrox.pb" "$GRPC_DIR/"
echo "✓ Proto descriptors copied to $GRPC_DIR/"
