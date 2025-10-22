#!/bin/bash
# Generate Protocol Buffers Go code from audiounit.proto
# This is part of the benchmark verification infrastructure

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCHEMA_DIR="$SCRIPT_DIR/../schemas"
OUTPUT_DIR="$SCRIPT_DIR/go"

echo "Generating Protocol Buffers code..."
echo "Schema: $SCHEMA_DIR/audiounit.proto"
echo "Output: $OUTPUT_DIR"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Generate Go code
protoc \
    --proto_path="$SCHEMA_DIR" \
    --go_out="$OUTPUT_DIR" \
    --go_opt=paths=source_relative \
    "$SCHEMA_DIR/audiounit.proto"

echo "✅ Protocol Buffers code generated successfully"
echo "   Generated: $OUTPUT_DIR/audiounit.pb.go"
