#!/bin/bash
# Display the version of stackrox modules being used for protos

set -e

ROX_VERSION=$(go list -f '{{.Version}}' -m github.com/stackrox/rox)
ROX_DIR=$(go list -f '{{.Dir}}' -m github.com/stackrox/rox)
echo "StackRox proto files from github.com/stackrox/rox@$ROX_VERSION"
echo "  Location: $ROX_DIR"

SCANNER_VERSION=$(go list -f '{{.Version}}' -m github.com/stackrox/scanner)
SCANNER_DIR=$(go list -f '{{.Dir}}' -m github.com/stackrox/scanner)
echo ""
echo "Scanner proto files from github.com/stackrox/scanner@$SCANNER_VERSION"
echo "  Location: $SCANNER_DIR"
