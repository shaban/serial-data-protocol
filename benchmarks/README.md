# Cross-Protocol Benchmarks

Fair comparison of SDP against FlatBuffers and Protocol Buffers using identical real-world data.

## Benchmark Data

Uses actual AudioUnit plugin enumeration data (62 plugins, 1,759 parameters, ~115 KB):
- 62 plugins with metadata (name, manufacturer, type, subtype)
- 1,759 total parameters across all plugins
- Each parameter: 11 fields (address, name, identifier, unit, 4 values, 3 flags)

## Protocols Compared

1. **SDP (Serial Data Protocol)** - This project
2. **Protocol Buffers** - Google's data interchange format
3. **FlatBuffers** - Google's zero-copy serialization library

## Running Benchmarks

```bash
# Install dependencies
go get google.golang.org/protobuf/proto
go get github.com/google/flatbuffers/go

# Generate schemas
make generate

# Run benchmarks
go test -bench=. -benchmem -benchtime=10s

# Compare results
go test -bench=. -benchmem -count=10 | tee results.txt
benchstat results.txt
```

## Schema Equivalence

All three implementations use identical schemas with the same field types and structure:

```
PluginRegistry
├── plugins: []Plugin
├── total_plugin_count: u32
└── total_parameter_count: u32

Plugin
├── name: string
├── manufacturer_id: string
├── component_type: string
├── component_subtype: string
└── parameters: []Parameter

Parameter
├── address: u64
├── display_name: string
├── identifier: string
├── unit: string
├── min_value: f32
├── max_value: f32
├── default_value: f32
├── current_value: f32
├── raw_flags: u32
├── is_writable: bool
└── can_ramp: bool
```

## Metrics

Each benchmark measures:
- **Encode time** - Convert Go structs → wire format
- **Decode time** - Convert wire format → Go structs
- **Roundtrip time** - Full encode + decode cycle
- **Allocations** - Number of heap allocations
- **Bytes allocated** - Total memory allocated

## Fair Comparison Rules

1. ✅ **Same data source** - All use `testdata/plugins.json`
2. ✅ **Same struct construction** - Parse JSON once, reuse for all benchmarks
3. ✅ **No optimization tricks** - Standard API usage for each protocol
4. ✅ **Verification included** - Each benchmark verifies correctness
5. ✅ **Exclude JSON parsing** - b.ResetTimer() after data loading
6. ✅ **Same Go version** - All compiled with same toolchain
7. ✅ **Latest versions** - Use current stable releases of all protocols

## Expected Results

This benchmark provides honest, reproducible numbers for realistic use cases.
No cherry-picking, no micro-optimizations - just real-world performance.
