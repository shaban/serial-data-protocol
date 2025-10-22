#!/bin/bash
# Test Swift package compilation
# Verifies swift build can compile generated packages

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
GENERATED_DIR="$PROJECT_ROOT/testdata/generated/swift"

echo "=== Swift Package Compilation Test ==="
echo ""

# Check if Swift is available
if ! command -v swift &> /dev/null; then
    echo "✗ Swift compiler not found"
    echo "  Swift is only available on macOS"
    exit 1
fi

# Check Swift version (need 5.9+ for C++ interop)
swift_version=$(swift --version | head -1 | grep -oE '[0-9]+\.[0-9]+' | head -1)
echo "Swift version: $swift_version"

if [ ! -d "$GENERATED_DIR" ] || [ -z "$(ls -A "$GENERATED_DIR" 2>/dev/null)" ]; then
    echo "✗ No generated Swift packages found"
    echo "  Run: make generate"
    exit 1
fi

passed=0
failed=0

# Test compilation of each package
for package_dir in "$GENERATED_DIR"/*; do
    if [ -d "$package_dir" ]; then
        name=$(basename "$package_dir")
        
        echo "Compiling: $name"
        
        cd "$package_dir"
        
        # Try to build
        if swift build -c release > /dev/null 2>&1; then
            # Check if message mode symbols exist (for packages that should have them)
            if [ -f "Sources/$name/message_encode.cpp" ]; then
                echo "  ✓ Compiled (byte + message mode)"
            else
                echo "  ✓ Compiled (byte mode only)"
            fi
            ((passed++))
        else
            echo "  ✗ Compilation failed"
            # Show error for debugging
            swift build -c release 2>&1 | tail -5
            ((failed++))
        fi
        
        cd "$PROJECT_ROOT"
    fi
done

echo ""
echo "=== Results ==="
echo "Passed: $passed"
echo "Failed: $failed"

if [ $failed -eq 0 ]; then
    echo "✓ All Swift packages compiled successfully"
    exit 0
else
    echo "✗ Some Swift packages failed to compile"
    exit 1
fi
