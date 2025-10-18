# Cross-Language Benchmarks

This directory contains benchmarks comparing SDP performance across Go and Rust implementations.

## Running Benchmarks

### Rust Benchmarks (using Criterion)

```bash
# From rust/sdp directory
cd rust/sdp

# Run all benchmarks
cargo bench

# Run specific benchmark
cargo bench wire_bench
cargo bench generated_bench
```

### Go Benchmarks

```bash
# From benchmarks directory
cd benchmarks

# Run all SDP benchmarks (Go only)
go test -bench=BenchmarkSDP -benchmem

# Run cross-language comparison (requires Rust)
go test -bench=. -benchmem

# Compare Go vs Rust on specific schema
go test -bench='Primitives' -benchmem
go test -bench='AudioUnit' -benchmem
```

### Manual Comparison

To get detailed comparison output:

```bash
# Go benchmarks
go test -bench='Go_' -benchmem > go_results.txt

# Rust benchmarks (via Go test)
go test -bench='Rust_' -benchmem > rust_results.txt

# Compare
diff -y go_results.txt rust_results.txt
```

## Benchmark Structure

### Rust Benchmarks

1. **`wire_bench.rs`** - Low-level wire format primitives
   - `encode_u32`, `decode_u32`
   - `encode_string`, `decode_string`
   - `encode_array`, `decode_array`

2. **`generated_bench.rs`** - Generated code performance
   - `primitives_encode`, `primitives_decode`, `primitives_roundtrip`
   - `audiounit_encode`, `audiounit_decode`, `audiounit_roundtrip`

3. **`rust-bench` binary** - Cross-language helper
   - Called by Go benchmarks to measure Rust performance
   - Outputs timing in nanoseconds per operation

### Go Benchmarks

1. **`comparison_test.go`** - SDP vs Protobuf vs FlatBuffers
   - Existing benchmarks comparing different serialization libraries

2. **`crosslang_bench_test.go`** - Go vs Rust comparison
   - `BenchmarkGo_Primitives_Encode/Decode`
   - `BenchmarkRust_Primitives_Encode/Decode`
   - `BenchmarkGo_AudioUnit_Encode/Decode`
   - `BenchmarkRust_AudioUnit_Encode/Decode`

## What Gets Measured

### Schema: Primitives
- 12 fields: u8, u16, u32, u64, i8, i16, i32, i64, f32, f64, bool, string
- Wire size: ~61 bytes
- Tests basic type encoding/decoding

### Schema: AudioUnit
- Nested structs: PluginRegistry → Plugin → Parameter
- Arrays of structs
- 2 plugins, 3 total parameters
- Wire size: ~300 bytes
- Tests real-world complex data structures

## Performance Metrics

Each benchmark reports:
- **ns/op**: Nanoseconds per operation (lower is better)
- **ops/sec**: Operations per second (higher is better)
- **B/op**: Bytes allocated per operation (Go benchmarks only)
- **allocs/op**: Allocations per operation (Go benchmarks only)

## Expected Results

Based on language characteristics:

**Rust advantages:**
- Zero-cost abstractions
- No garbage collection
- LLVM optimizations
- Likely faster for CPU-bound encoding/decoding

**Go advantages:**
- Simpler memory model
- Faster compilation
- Better for concurrent workloads

**Note:** Both implementations use the same wire format, so encoded output is byte-for-byte identical. Performance differences come from language implementation details.

## Interpreting Results

Example output:

```
BenchmarkGo_Primitives_Encode-10       2000000       650 ns/op      64 B/op       1 allocs/op
BenchmarkRust_Primitives_Encode-10     2000000       420 ns/op
```

This means:
- Go: 650 ns/op (~1.5 million ops/sec)
- Rust: 420 ns/op (~2.4 million ops/sec)
- Rust is ~1.5x faster for this operation
- Go allocates 64 bytes per operation

## Cross-Language Testing

The cross-language benchmarks validate that:
1. Rust can decode Go-encoded data
2. Go can decode Rust-encoded data
3. Performance can be compared fairly

This is achieved by:
- Using identical test data in both languages
- Measuring actual encode/decode operations (not data transfer)
- Running Rust benchmarks as a subprocess from Go tests
