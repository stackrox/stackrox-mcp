#!/bin/bash
set -e

DESCRIPTOR_DIR="wiremock/proto/descriptors"
ROX_PROTO_PATH="wiremock/proto/stackrox"
GOOGLEAPIS_PATH="wiremock/proto/googleapis"
GRPC_DIR="wiremock/grpc"

mkdir -p "$DESCRIPTOR_DIR" "$GRPC_DIR"

# Ensure proto files are present
if [ ! -d "$ROX_PROTO_PATH" ]; then
    echo "Proto files not found. Running setup..."
    ./scripts/setup-proto-files.sh
fi

# Use PROTOC_BIN if set (from Makefile), otherwise use system protoc
PROTOC_CMD="${PROTOC_BIN:-protoc}"

if ! command -v "$PROTOC_CMD" &> /dev/null; then
    echo "Error: protoc is not installed"
    echo "Install with: make proto-install"
    echo "Or install manually from: https://grpc.io/docs/protoc-installation/"
    exit 1
fi

echo "Generating proto descriptors with $PROTOC_CMD..."

"$PROTOC_CMD" \
  --descriptor_set_out="$DESCRIPTOR_DIR/stackrox.dsc" \
  --include_imports \
  --proto_path="$ROX_PROTO_PATH" \
  --proto_path="$GOOGLEAPIS_PATH" \
  api/v1/deployment_service.proto \
  api/v1/image_service.proto \
  api/v1/node_service.proto \
  api/v1/cluster_service.proto

cp "$DESCRIPTOR_DIR/stackrox.dsc" "$GRPC_DIR/"

echo "✓ Proto descriptors generated at $DESCRIPTOR_DIR/stackrox.dsc"
