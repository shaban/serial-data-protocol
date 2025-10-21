#!/usr/bin/env bash
# Run C++ tests

set -e

echo "========================================"
echo "Running C++ Tests"
echo "========================================"
echo ""

# Find all C++ test directories
CPP_TESTS=$(find testdata/cpp -name Makefile -exec dirname {} \; 2>/dev/null | sort)

if [ -z "$CPP_TESTS" ]; then
    echo "⚠️  No C++ tests found (testdata/cpp/**/Makefile)"
    echo "   Skipping C++ tests"
    exit 0
fi

FAILED=0
PASSED=0

for test_dir in $CPP_TESTS; do
    test_name=$(basename "$test_dir")
    echo "Testing: $test_name"
    
    # Build and run test
    if (cd "$test_dir" && make clean > /dev/null 2>&1 && make test); then
        PASSED=$((PASSED + 1))
        echo "  ✓ $test_name passed"
    else
        FAILED=$((FAILED + 1))
        echo "  ✗ $test_name failed"
    fi
    echo ""
done

echo "Results: $PASSED passed, $FAILED failed"

if [ $FAILED -gt 0 ]; then
    echo "❌ C++ tests failed"
    exit 1
fi

echo "✅ C++ tests passed"
