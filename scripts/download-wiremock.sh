#!/bin/bash
set -e

WIREMOCK_VERSION="3.9.1"
GRPC_EXTENSION_VERSION="0.8.0"

WIREMOCK_DIR="wiremock/lib"
mkdir -p "$WIREMOCK_DIR"

echo "Downloading WireMock standalone JAR..."
curl -L "https://repo1.maven.org/maven2/org/wiremock/wiremock-standalone/${WIREMOCK_VERSION}/wiremock-standalone-${WIREMOCK_VERSION}.jar" \
  -o "$WIREMOCK_DIR/wiremock-standalone.jar"

echo "Downloading WireMock gRPC extension..."
curl -L "https://repo1.maven.org/maven2/org/wiremock/wiremock-grpc-extension-standalone/${GRPC_EXTENSION_VERSION}/wiremock-grpc-extension-standalone-${GRPC_EXTENSION_VERSION}.jar" \
  -o "$WIREMOCK_DIR/wiremock-grpc-extension.jar"

echo "✓ WireMock JARs downloaded to $WIREMOCK_DIR"
