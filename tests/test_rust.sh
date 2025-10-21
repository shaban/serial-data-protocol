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
    echo "   Skipping Rust tests"
    exit 0
fi

FAILED=0
PASSED=0

for test_dir in $RUST_TESTS; do
    test_name=$(basename "$test_dir")
    echo "Testing: $test_name"
    
    # Build and run test
    if (cd "$test_dir" && cargo test --quiet 2>&1); then
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
    echo "❌ Rust tests failed"
    exit 1
fi

echo "✅ Rust tests passed"
