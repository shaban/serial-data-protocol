#!/usr/bin/env bash
# Run Swift tests

set -e

echo "========================================"
echo "Running Swift Tests"
echo "========================================"
echo ""

# Find all Swift test directories (those with Package.swift)
SWIFT_TESTS=$(find testdata/swift -name Package.swift -exec dirname {} \; 2>/dev/null | sort)

if [ -z "$SWIFT_TESTS" ]; then
    echo "⚠️  No Swift tests found (testdata/swift/**/Package.swift)"
    echo "   Skipping Swift tests"
    exit 0
fi

FAILED=0
PASSED=0

for test_dir in $SWIFT_TESTS; do
    test_name=$(basename "$test_dir")
    echo "Testing: $test_name"
    
    # Swift packages in testdata are just wrappers - they don't have tests
    # The actual testing is done via the macos_testing directory
    echo "  ℹ️  $test_name is a Swift package wrapper (no unit tests)"
    echo "     See macos_testing/ for Swift integration tests"
    echo ""
done

echo "ℹ️  Swift unit tests should be run via macos_testing/"
echo "   Run: cd macos_testing && make test"
