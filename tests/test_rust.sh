#!/usr/bin/env bash
# Run Rust tests

set -e

echo "========================================"
echo "Running Rust Tests"
echo "========================================"
echo ""

# Find all Rust test directories (those with Cargo.toml)
RUST_TESTS=$(find testdata/rust -name Cargo.toml -exec dirname {} \; 2>/dev/null | sort)

if [ -z "$RUST_TESTS" ]; then
    echo "⚠️  No Rust tests found (testdata/rust/**/Cargo.toml)"
    echo "   Run sdp-gen to generate Rust test packages first"
    exit 0
fi

FAILED=0
PASSED=0
TOTAL=0

for test_dir in $RUST_TESTS; do
    TOTAL=$((TOTAL + 1))
    test_name=$(basename "$test_dir")
    echo "Testing: $test_name"
    
    # Build library only (skip examples/benches which may not exist)
    # Generated packages are libraries without test functions, so we just verify compilation
    if (cd "$test_dir" && cargo build --lib --quiet 2>&1 | grep -v "^warning:" > /dev/null); then
        PASSED=$((PASSED + 1))
        echo "  ✓ builds successfully"
    else
        FAILED=$((FAILED + 1))
        echo "  ✗ FAILED to build"
        # Show error output on failure
        (cd "$test_dir" && cargo build --lib 2>&1) | grep -E "^error" | sed 's/^/    /'
    fi
    echo ""
done

echo "========================================"
echo "Rust Test Results"
echo "========================================"
echo "Total:  $TOTAL"
echo "Passed: $PASSED"
echo "Failed: $FAILED"

if [ $FAILED -gt 0 ]; then
    echo ""
    echo "❌ Rust tests failed"
    exit 1
fi

echo ""
echo "✅ All Rust tests passed"
