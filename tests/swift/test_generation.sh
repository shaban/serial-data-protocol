#!/bin/bash
# Test Swift code generation
# Verifies sdp-gen can generate valid Swift packages for all schemas

set -e

PROJECT_ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
SCHEMAS_DIR="$PROJECT_ROOT/testdata/schemas"
GENERATED_DIR="$PROJECT_ROOT/testdata/generated/swift"
SDP_GEN="$PROJECT_ROOT/sdp-gen"

echo "=== Swift Code Generation Test ==="
echo ""

# Build sdp-gen if not exists
if [ ! -f "$SDP_GEN" ]; then
    echo "Building sdp-gen..."
    cd "$PROJECT_ROOT"
    go build -o sdp-gen ./cmd/sdp-gen
fi

passed=0
failed=0

# Test each schema
for schema in "$SCHEMAS_DIR"/*.sdp; do
    name=$(basename "$schema" .sdp)
    output_dir="$GENERATED_DIR/$name"
    
    echo "Testing: $name.sdp"
    
    # Generate Swift package
    if "$SDP_GEN" -schema "$schema" -output "$output_dir" -lang swift > /dev/null 2>&1; then
        # Check required files exist
        if [ -f "$output_dir/Package.swift" ] && \
           [ -f "$output_dir/Sources/$name/include/module.modulemap" ] && \
           [ -f "$output_dir/Sources/$name/include/types.hpp" ] && \
           [ -f "$output_dir/Sources/$name/encode.cpp" ] && \
           [ -f "$output_dir/Sources/$name/decode.cpp" ]; then
            
            # Check if message mode files exist (for schemas with structs)
            has_message_files=false
            if [ -f "$output_dir/Sources/$name/message_encode.cpp" ] && \
               [ -f "$output_dir/Sources/$name/message_decode.cpp" ]; then
                has_message_files=true
            fi
            
            if [ "$has_message_files" = true ]; then
                echo "  ✓ Generated (byte + message mode)"
            else
                echo "  ✓ Generated (byte mode only)"
            fi
            ((passed++))
        else
            echo "  ✗ Missing required files"
            ((failed++))
        fi
    else
        echo "  ✗ Generation failed"
        ((failed++))
    fi
done

echo ""
echo "=== Results ==="
echo "Passed: $passed"
echo "Failed: $failed"

if [ $failed -eq 0 ]; then
    echo "✓ All Swift packages generated successfully"
    exit 0
else
    echo "✗ Some Swift packages failed to generate"
    exit 1
fi
