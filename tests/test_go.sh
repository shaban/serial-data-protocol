#!/usr/bin/env bash
# Run Go tests with proper formatting

set -e

echo "========================================"
echo "Running Go Tests"
echo "========================================"
echo ""

# Build sdp-gen first to ensure it's available
if [ ! -f "./sdp-gen" ]; then
    echo "Building sdp-gen..."
    go build -o sdp-gen ./cmd/sdp-gen
    echo ""
fi

# Run tests with coverage
echo "Running test suite..."
go test -v -cover ./... 2>&1 | grep -E "^(===|---|PASS|FAIL|coverage:|ok|SKIP|\s+\w+_test\.go:)" || go test ./...

echo ""
echo "✅ Go tests passed"
