# Benchmark Data Workflow

This document describes how SDP uses canonical JSON test data for benchmarking and cross-language verification.

## Overview

**SDP uses JSON as the canonical input format** for benchmarks and cross-language testing. This ensures:

1. **Language independence** - JSON is universally parseable
2. **Human readability** - Easy to inspect and modify test data
3. **Version control friendly** - Text-based diffs work well
4. **Single source of truth** - One input generates tests for all languages

## Data Location

```
testdata/
├── schemas/          # Schema definitions (.sdp)
│   ├── primitives.sdp
│   ├── audiounit.sdp
│   ├── arrays.sdp
│   └── ...
│
├── data/             # CANONICAL INPUT (JSON)
│   ├── primitives.json
│   ├── plugins.json       # Used for benchmarks (115 KB, 62 plugins)
│   ├── arrays.json
│   ├── nested.json
│   └── optional.json
│
└── binaries/         # Reference wire format (.sdpb)
    ├── primitives.sdpb
    ├── nested.sdpb
    └── optional.sdpb
```

## Benchmark Workflow

### 1. Canonical Input

**Location:** `testdata/data/plugins.json`

**Content:** Real AudioUnit plugin enumeration data:
- 62 plugins with metadata
- 1,759 parameters across all plugins
- ~115 KB JSON (~58 KB SDP binary)

**Usage:** All benchmark implementations (Go, C++, Rust, Protocol Buffers, FlatBuffers) parse this same JSON file to ensure fair comparison.

### 2. Running Benchmarks

```bash
# Go benchmarks (uses testdata/data/plugins.json)
cd benchmarks
go test -bench=. -benchmem

# C++ benchmarks (future)
cd benchmarks/standalone/cpp
make bench

# Rust benchmarks (future)
cd benchmarks/standalone/rust
cargo bench
```

### 3. Benchmark Comparison

Each language implementation:
1. Parses `testdata/data/plugins.json`
2. Constructs native data structures
3. Encodes to binary format
4. Decodes back to structures
5. Verifies roundtrip correctness
6. Measures time and memory

### 4. Cross-Language Verification

The same JSON data enables cross-language wire format verification:

```bash
# Generate binary from JSON using Go
./sdp-encode -schema audiounit \
             -json testdata/data/plugins.json \
             -out /tmp/go_output.sdpb \
             -type PluginRegistry

# Compare with C++ encoder output (future)
./cpp_encoder < testdata/data/plugins.json > /tmp/cpp_output.sdpb
diff /tmp/go_output.sdpb /tmp/cpp_output.sdpb  # Should be identical

# Or use checksums
sha256sum /tmp/go_output.sdpb /tmp/cpp_output.sdpb
```

## Adding New Benchmark Data

### Step 1: Create JSON File

Add new test data to `testdata/data/`:

```bash
# Example: Create sample data for a new schema
cat > testdata/data/my_schema.json <<'EOF'
[
  {
    "field1": 42,
    "field2": "test",
    "nested": {
      "x": 1.5,
      "y": 2.5
    }
  }
]
EOF
```

### Step 2: Generate Reference Binary

Use the Go encoder to create a reference binary:

```bash
./sdp-encode -schema my_schema \
             -json testdata/data/my_schema.json \
             -out testdata/binaries/my_schema.sdpb \
             -type MyType
```

### Step 3: Verify Across Languages

Each language implementation should:
1. Parse the JSON
2. Encode to binary
3. Compare against `testdata/binaries/my_schema.sdpb`
4. Decode and verify fields match JSON

## Benchmark Metrics

### Current Measurements

From `benchmarks/RESULTS.md`:

| Operation | SDP | Protocol Buffers | Speedup |
|-----------|-----|------------------|---------|
| Encode    | 39 µs | 239 µs | **6.1×** |
| Decode    | 98 µs | 315 µs | **3.2×** |

### Data Characteristics

**plugins.json (AudioUnit benchmark):**
- JSON size: ~115 KB
- SDP binary: ~58 KB (50% compression)
- Protocol Buffers: ~62 KB (46% compression)
- FlatBuffers: N/A (zero-copy, no encoding step)

## Design Rationale

### Why JSON as Input?

1. **Universal parsing** - Every language has JSON libraries
2. **Type mapping** - JSON types map cleanly to SDP types:
   - JSON number → u32, i32, f32, f64
   - JSON string → string
   - JSON array → []T
   - JSON object → struct
   - JSON null → ?T (optional)

3. **Version control** - Human-readable diffs
4. **Tooling** - Easy to generate, validate, pretty-print

### Why Not Use Binary as Input?

- Not human-readable
- Hard to modify
- Language-specific parsers needed
- Poor diff support in git

### Why Not Use Protocol-Specific Input?

- Ties benchmarks to specific protocol
- Can't fairly compare different formats
- Harder to verify correctness

## Future Enhancements

1. **Automated verification** - Script to encode JSON with all languages and compare outputs
2. **Benchmark data generator** - Create synthetic data at various sizes
3. **Streaming benchmarks** - Test large datasets that don't fit in memory
4. **Compression benchmarks** - Compare gzip/zstd with different formats

## Related Documentation

- `benchmarks/README.md` - How to run benchmarks
- `benchmarks/RESULTS.md` - Current benchmark results
- `TESTING_STRATEGY.md` - Overall testing approach
- `tests/verify_wire_format.sh` - Cross-language verification script
