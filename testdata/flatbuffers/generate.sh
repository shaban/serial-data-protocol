#!/bin/bash
# Generate FlatBuffers Go code from audiounit.fbs
# This is part of the benchmark verification infrastructure

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCHEMA_DIR="$SCRIPT_DIR/../schemas"
OUTPUT_DIR="$SCRIPT_DIR/go"

echo "Generating FlatBuffers code..."
echo "Schema: $SCHEMA_DIR/audiounit.fbs"
echo "Output: $OUTPUT_DIR"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Generate Go code
flatc \
    --go \
    --go-namespace fb \
    --gen-onefile \
    -o "$OUTPUT_DIR" \
    "$SCHEMA_DIR/audiounit.fbs"

echo "✅ FlatBuffers code generated successfully"
echo "   Generated: $OUTPUT_DIR/audiounit/*.go"
