#!/bin/bash
set -e

DESCRIPTOR_DIR="wiremock/proto/descriptors"
ROX_PROTO_PATH="wiremock/proto/stackrox"
GOOGLEAPIS_PATH="wiremock/proto/googleapis"
GRPC_DIR="wiremock/grpc"

mkdir -p "$DESCRIPTOR_DIR" "$GRPC_DIR"

if [ ! -d "$ROX_PROTO_PATH" ]; then
    echo "Error: Proto files not found at $ROX_PROTO_PATH"
    echo "Run: ./scripts/setup-proto-files.sh"
    exit 1
fi

if ! command -v protoc &> /dev/null; then
    echo "Error: protoc is not installed"
    echo "Install from: https://grpc.io/docs/protoc-installation/"
    exit 1
fi

echo "Generating proto descriptors..."

protoc \
  --descriptor_set_out="$DESCRIPTOR_DIR/stackrox.pb" \
  --include_imports \
  --proto_path="$ROX_PROTO_PATH" \
  --proto_path="$GOOGLEAPIS_PATH" \
  api/v1/deployment_service.proto \
  api/v1/image_service.proto \
  api/v1/node_service.proto \
  api/v1/cluster_service.proto

cp "$DESCRIPTOR_DIR/stackrox.pb" "$GRPC_DIR/"

echo "✓ Proto descriptors generated at $DESCRIPTOR_DIR/stackrox.pb"
