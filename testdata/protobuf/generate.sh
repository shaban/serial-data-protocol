#!/bin/bash
# Generate Protocol Buffers Go code from audiounit.proto
# This is part of the benchmark verification infrastructure

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
SCHEMA_FILE="$PROJECT_ROOT/testdata/schemas/audiounit.proto"
SCHEMA_DIR="$(dirname "$SCHEMA_FILE")"
OUTPUT_DIR="$PROJECT_ROOT/testdata/generated/protobuf/go"

echo "Generating Protocol Buffers code..."
echo "Schema: $SCHEMA_FILE"
echo "Output: $OUTPUT_DIR"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Generate Go code
protoc \
    --proto_path="$SCHEMA_DIR" \
    --go_out="$OUTPUT_DIR" \
    --go_opt=paths=source_relative \
    "$SCHEMA_FILE"

# Create go.mod for the generated package
cat > "$OUTPUT_DIR/go.mod" <<EOF
module github.com/shaban/serial-data-protocol/testdata/generated/protobuf/go

go 1.25.1

require google.golang.org/protobuf v1.31.0
EOF

echo "✅ Protocol Buffers code generated successfully"
echo "   Generated: $OUTPUT_DIR/audiounit.pb.go"
