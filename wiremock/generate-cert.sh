#!/bin/bash
set -e

# NOTE: This script is for MANUAL regeneration only.
# The certificate is committed to the repo with 100-year validity (until 2126).
# You should not need to run this script unless you need to regenerate the certificate.

# Get script directory and navigate to repo root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$REPO_ROOT"

CERT_DIR="wiremock/certs"
KEYSTORE_FILE="$CERT_DIR/keystore.jks"
# Note: Password is hardcoded for local development/testing only
KEYSTORE_PASS="wiremock"

# Check if keytool is available
if ! command -v keytool &> /dev/null; then
    echo "Error: keytool not found. Please install Java JDK"
    echo "  Ubuntu/Debian: sudo apt-get install openjdk-11-jdk"
    echo "  macOS: brew install openjdk@11"
    exit 1
fi

mkdir -p "$CERT_DIR"

if [ -f "$KEYSTORE_FILE" ]; then
    echo "Certificate already exists at $KEYSTORE_FILE"
    exit 0
fi

echo "Generating self-signed certificate for WireMock..."

# Generate keystore with self-signed certificate
keytool -genkeypair \
    -alias wiremock \
    -keyalg RSA \
    -keysize 2048 \
    -storetype JKS \
    -keystore "$KEYSTORE_FILE" \
    -storepass "$KEYSTORE_PASS" \
    -keypass "$KEYSTORE_PASS" \
    -validity 36500 \
    -dname "CN=localhost, OU=WireMock, O=StackRox Testing, L=Local, ST=Dev, C=US" \
    -ext "SAN=dns:localhost,ip:127.0.0.1"

echo "✓ Certificate generated at $KEYSTORE_FILE"
