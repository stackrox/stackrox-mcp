#!/bin/bash
set -e

STACKROX_REPO=""

if [ -d "../stackrox" ]; then
    STACKROX_REPO="../stackrox"
elif [ -d "../../stackrox" ]; then
    STACKROX_REPO="../../stackrox"
elif [ -n "$STACKROX_REPO_PATH" ]; then
    STACKROX_REPO="$STACKROX_REPO_PATH"
fi

if [ -z "$STACKROX_REPO" ] || [ ! -d "$STACKROX_REPO" ]; then
    echo "Error: StackRox repository not found"
    echo "Set STACKROX_REPO_PATH or clone to ../stackrox"
    exit 1
fi

echo "Copying proto files from $STACKROX_REPO..."

mkdir -p wiremock/proto/stackrox wiremock/proto/googleapis

cp -r "$STACKROX_REPO/proto/"* wiremock/proto/stackrox/
cp -r "$STACKROX_REPO/third_party/googleapis/"* wiremock/proto/googleapis/

mkdir -p wiremock/proto/stackrox/scanner/api/v1
if [ -d "$STACKROX_REPO/qa-tests-backend/src/main/proto/scanner/api/v1" ]; then
    cp "$STACKROX_REPO/qa-tests-backend/src/main/proto/scanner/api/v1/"*.proto wiremock/proto/stackrox/scanner/api/v1/
fi

echo "✓ Proto files copied"
echo "Next: ./scripts/generate-proto-descriptors.sh"
