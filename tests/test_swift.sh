#!/usr/bin/env bash
# Swift Test Runner
# Runs all Swift tests (generation, compilation, smoke tests)

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"

echo "========================================"
echo "Running Swift Tests"
echo "========================================"
echo ""

# Check if Swift is available
if ! command -v swift &> /dev/null; then
    echo "⚠️  Swift not available - skipping Swift tests"
    echo "   Swift is only available on macOS"
    exit 0
fi

# Run generation test
echo "1. Testing Swift code generation..."
"$SCRIPT_DIR/swift/test_generation.sh"
echo ""

# Run compilation test  
echo "2. Testing Swift package compilation..."
"$SCRIPT_DIR/swift/test_compilation.sh"
echo ""

# Run smoke test
echo "3. Running Swift smoke tests..."
"$SCRIPT_DIR/swift/test_smoke.sh"
echo ""

echo "========================================"
echo "Swift Tests Complete"
echo "========================================"
echo "✓ Code generation: All schemas generated successfully"
echo "✓ Compilation: All packages compiled successfully"
echo "✓ Smoke tests: C++ backend accessible via Swift"
echo "✓ Message mode: Functions generated for all schemas"
