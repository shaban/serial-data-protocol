#!/usr/bin/env bash
# Run C++ tests

set -e

echo "========================================"
echo "Running C++ Tests"
echo "========================================"
echo ""

# Find all C++ test executables
CPP_TESTS=$(find testdata/cpp -type f -perm +111 -name "test_*" 2>/dev/null | sort)

if [ -z "$CPP_TESTS" ]; then
    echo "⚠️  No C++ tests found (testdata/cpp/*/test_*)"
    echo "   Run sdp-gen to generate C++ test packages first"
    exit 0
fi

FAILED=0
PASSED=0
TOTAL=0

for test_exe in $CPP_TESTS; do
    TOTAL=$((TOTAL + 1))
    test_name=$(basename $(dirname "$test_exe"))/$(basename "$test_exe")
    echo "Running: $test_name"
    
    # Run test executable
    if "$test_exe" > /dev/null 2>&1; then
        PASSED=$((PASSED + 1))
        echo "  ✓ passed"
    else
        FAILED=$((FAILED + 1))
        echo "  ✗ FAILED"
        # Show error output
        "$test_exe" 2>&1 | sed 's/^/    /'
    fi
    echo ""
done

echo "========================================"
echo "C++ Test Results"
echo "========================================"
echo "Total:  $TOTAL"
echo "Passed: $PASSED"
echo "Failed: $FAILED"

if [ $FAILED -gt 0 ]; then
    echo ""
    echo "❌ C++ tests failed"
    exit 1
fi

echo ""
echo "✅ All C++ tests passed"
