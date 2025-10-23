#!/bin/bash
# Generate FlatBuffers Go code from audiounit.fbs
# This is part of the benchmark verification infrastructure

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
SCHEMA_FILE="$PROJECT_ROOT/testdata/schemas/audiounit.fbs"
OUTPUT_DIR="$PROJECT_ROOT/testdata/generated/flatbuffers/go"

echo "Generating FlatBuffers code..."
echo "Schema: $SCHEMA_FILE"
echo "Output: $OUTPUT_DIR"

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Generate Go code
flatc \
    --go \
    --go-namespace fb \
    --gen-onefile \
    -o "$OUTPUT_DIR" \
    "$SCHEMA_FILE"

# Create go.mod for the generated package
cat > "$OUTPUT_DIR/go.mod" <<EOF
module github.com/shaban/serial-data-protocol/testdata/generated/flatbuffers/go

go 1.25.1

require github.com/google/flatbuffers v23.5.26+incompatible
EOF

# Run go mod tidy to create go.sum
(cd "$OUTPUT_DIR" && go mod tidy)

echo "✅ FlatBuffers code generated successfully"
echo "   Generated: $OUTPUT_DIR/audiounit/*.go"
