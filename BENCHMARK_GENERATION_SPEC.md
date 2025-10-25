# Benchmark Generation Specification

**Version:** 0.1.0  
**Status:** Design Phase  
**Last Updated:** October 23, 2025

## Overview

The `sdp-gen` tool can automatically generate language-specific benchmark harnesses from schema + sample JSON data. This enables:

1. **Reproducible performance testing** - Same input → comparable results
2. **User adoption** - "Try SDP with your data in 30 seconds"
3. **Cross-language verification** - Ensure wire format compatibility
4. **Canonical test fixtures** - Automated `.sdpb` generation for decode benchmarks

## Command-Line Interface

### Basic Usage

```bash
./sdp-gen -schema <schema.sdp> -output <path> -lang <language> -bench <sample.json>
```

**Example:**
```bash
./sdp-gen -schema testdata/audiounit.sdp \
          -output testdata/generated/go/audiounit \
          -lang go \
          -bench testdata/audiounit.json
```

### Flags

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `-schema` | string | Yes | Path to `.sdp` schema file |
| `-output` | string | Yes | Output directory for generated code |
| `-lang` | string | Yes | Target language: `go`, `cpp`, `rust`, `rustexp` |
| `-bench` | string | No | Path to sample JSON file for benchmark generation |

**When `-bench` is omitted:**
- No benchmarks generated (default behavior, backward compatible)
- Only library code produced

**When `-bench` is provided:**
- Validates JSON against schema (strict type checking)
- Encodes JSON → canonical `.sdpb` binary
- Generates language-specific benchmark harness
- Optionally runs cross-language verification

## Generated Output

### File Structure (Go Example)

```
testdata/generated/go/audiounit/
├── audiounit.go              # Generated library (always)
├── audiounit_bench_test.go   # Generated benchmarks (if -bench used)
└── audiounit_sample.sdpb     # Canonical binary (if -bench used)
```

### Benchmark File Naming

| Language | Pattern | Example |
|----------|---------|---------|
| Go | `{schema}_bench_test.go` | `audiounit_bench_test.go` |
| C++ | `{schema}_benchmark.cpp` | `audiounit_benchmark.cpp` |
| Rust | `benches/{schema}_bench.rs` | `benches/audiounit_bench.rs` |

### Canonical Binary Naming

Pattern: `{schema}_sample.sdpb`

Example: `audiounit_sample.sdpb`

**Purpose:**
- Decode benchmarks use this pre-encoded binary (fast, no JSON parsing overhead)
- Cross-language verification compares this file across all languages
- Can be committed to git as test fixture

## JSON → Schema Type Mapping

### Validation Strategy: **Strict Matching**

JSON values must match schema types exactly. No implicit coercion.

### Type Rules

| Schema Type | JSON Type | Valid Examples | Invalid Examples |
|-------------|-----------|----------------|------------------|
| `u8` | number | `0`, `255` | `256`, `-1`, `"42"`, `null` |
| `u16` | number | `0`, `65535` | `65536`, `-1` |
| `u32` | number | `0`, `4294967295` | `-1`, `4.5` |
| `u64` | number | `0`, `2^53-1`* | Larger than safe integer |
| `i8` | number | `-128`, `127` | `128`, `"0"` |
| `i16` | number | `-32768`, `32767` | `32768` |
| `i32` | number | `-2147483648`, `2147483647` | Out of range |
| `i64` | number | `-(2^53-1)`, `2^53-1`* | Larger than safe integer |
| `f32` | number | `3.14`, `-0.5`, `1e10` | `"3.14"`, `null` |
| `f64` | number | `3.14159265359`, `1e308` | `"3.14"`, `null` |
| `bool` | boolean | `true`, `false` | `1`, `0`, `"true"`, `null` |
| `string` | string | `"hello"`, `""`, `"🎵"` | `42`, `null`, `["a"]` |
| `[]T` | array | `[1, 2, 3]`, `[]` | `null`, `42`, `"[1,2,3]"` |
| `StructType` | object | `{"field": value}` | `null`, `42`, `"string"` |
| `?T` | null or T | `null`, `{"field": value}` | `undefined`, omitted field** |

\* JSON numbers are limited to 53-bit integers (JavaScript safe integer range)  
\** Optional fields can be omitted from JSON (treated as `null`)

### Error Messages

**Type Mismatch:**
```
Error: Type mismatch at 'plugins[0].id'
  Expected: u32
  Got: string "invalid"
  JSON location: line 15, column 12
```

**Range Overflow:**
```
Error: Value out of range at 'parameters[5].value'
  Expected: u8 (0-255)
  Got: 300
  JSON location: line 47, column 23
```

**Missing Required Field:**
```
Error: Missing required field 'name' in struct Plugin
  At: 'plugins[1]'
  JSON location: line 28, column 3
```

**Array Length Mismatch:**
```
Error: Array exceeds maximum size
  Field: 'parameters'
  Max allowed: 100000
  Actual: 150000
  JSON location: line 52, column 5
```

## Generated Benchmark Structure

### Go Benchmarks (`_bench_test.go`)

```go
package audiounit

import (
    "testing"
)

var (
    // Parsed once at package init, not in benchmark loop
    sampleData AudioUnit
    sampleBinary []byte
)

func init() {
    // Load and decode sample JSON (one-time setup)
    sampleData = loadSampleFromJSON("audiounit.json")
    
    // Pre-encode for decode benchmarks (one-time setup)
    var err error
    sampleBinary, err = EncodeAudioUnit(&sampleData)
    if err != nil {
        panic(err)
    }
}

func BenchmarkAudioUnit_Encode(b *testing.B) {
    b.ReportAllocs()
    b.SetBytes(int64(len(sampleBinary))) // Report throughput
    
    for i := 0; i < b.N; i++ {
        _, err := EncodeAudioUnit(&sampleData)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkAudioUnit_Decode(b *testing.B) {
    b.ReportAllocs()
    b.SetBytes(int64(len(sampleBinary)))
    
    var dest AudioUnit
    for i := 0; i < b.N; i++ {
        err := DecodeAudioUnit(&dest, sampleBinary)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkAudioUnit_Roundtrip(b *testing.B) {
    b.ReportAllocs()
    b.SetBytes(int64(len(sampleBinary)))
    
    var dest AudioUnit
    for i := 0; i < b.N; i++ {
        encoded, err := EncodeAudioUnit(&sampleData)
        if err != nil {
            b.Fatal(err)
        }
        
        err = DecodeAudioUnit(&dest, encoded)
        if err != nil {
            b.Fatal(err)
        }
    }
}

// Also generate message mode benchmarks if schema has message types
func BenchmarkAudioUnitMessage_Encode(b *testing.B) { /* ... */ }
func BenchmarkAudioUnitMessage_Decode(b *testing.B) { /* ... */ }
```

### C++ Benchmarks (`_benchmark.cpp`)

```cpp
#include "audiounit.hpp"
#include <chrono>
#include <iostream>
#include <fstream>
#include <vector>

using namespace std::chrono;

// Load pre-encoded binary
std::vector<uint8_t> loadBinary(const char* path) {
    std::ifstream file(path, std::ios::binary);
    return std::vector<uint8_t>(
        std::istreambuf_iterator<char>(file),
        std::istreambuf_iterator<char>()
    );
}

// Warmup runs to prime caches
template<typename Fn>
void warmup(Fn&& fn, int iterations = 10) {
    for (int i = 0; i < iterations; ++i) {
        fn();
    }
}

// Measure median of N runs (outlier resistant)
template<typename Fn>
double benchmark(Fn&& fn, int iterations = 1000) {
    std::vector<double> times;
    times.reserve(iterations);
    
    for (int i = 0; i < iterations; ++i) {
        auto start = high_resolution_clock::now();
        fn();
        auto end = high_resolution_clock::now();
        
        times.push_back(duration_cast<nanoseconds>(end - start).count());
    }
    
    std::sort(times.begin(), times.end());
    return times[times.size() / 2]; // Median
}

int main() {
    // Load sample data
    auto binary = loadBinary("audiounit_sample.sdpb");
    
    // Decode sample for encode benchmarks
    sdp::AudioUnit sample;
    sdp::DecodeAudioUnit(&sample, binary.data(), binary.size());
    
    // Encode benchmark
    warmup([&]() {
        std::vector<uint8_t> output;
        sdp::EncodeAudioUnit(&sample, output);
    });
    
    double encode_ns = benchmark([&]() {
        std::vector<uint8_t> output;
        sdp::EncodeAudioUnit(&sample, output);
    });
    
    std::cout << "AudioUnit Encode: " << encode_ns << " ns/iter\n";
    std::cout << "  Throughput: " << (binary.size() / encode_ns * 1e9 / 1024 / 1024) << " MB/s\n";
    
    // Decode benchmark
    warmup([&]() {
        sdp::AudioUnit dest;
        sdp::DecodeAudioUnit(&dest, binary.data(), binary.size());
    });
    
    double decode_ns = benchmark([&]() {
        sdp::AudioUnit dest;
        sdp::DecodeAudioUnit(&dest, binary.data(), binary.size());
    });
    
    std::cout << "AudioUnit Decode: " << decode_ns << " ns/iter\n";
    std::cout << "  Throughput: " << (binary.size() / decode_ns * 1e9 / 1024 / 1024) << " MB/s\n";
    
    return 0;
}
```

### Rust Benchmarks (`benches/{schema}_bench.rs`)

```rust
use criterion::{criterion_group, criterion_main, Criterion, BenchmarkId, Throughput};
use sdp_audiounit::{AudioUnit, decode_audio_unit, encode_audio_unit};
use std::fs;

fn load_binary(path: &str) -> Vec<u8> {
    fs::read(path).expect("Failed to read binary")
}

fn benchmark_audiounit(c: &mut Criterion) {
    // Load pre-encoded binary
    let binary = load_binary("audiounit_sample.sdpb");
    
    // Decode for encode benchmarks
    let sample = decode_audio_unit(&binary).expect("Failed to decode sample");
    
    // Setup throughput measurement
    let mut group = c.benchmark_group("AudioUnit");
    group.throughput(Throughput::Bytes(binary.len() as u64));
    
    // Encode benchmark
    group.bench_function("encode", |b| {
        b.iter(|| {
            encode_audio_unit(&sample).expect("Encode failed")
        })
    });
    
    // Decode benchmark
    group.bench_function("decode", |b| {
        b.iter(|| {
            decode_audio_unit(&binary).expect("Decode failed")
        })
    });
    
    // Roundtrip benchmark
    group.bench_function("roundtrip", |b| {
        b.iter(|| {
            let encoded = encode_audio_unit(&sample).expect("Encode failed");
            decode_audio_unit(&encoded).expect("Decode failed")
        })
    });
    
    group.finish();
}

criterion_group!(benches, benchmark_audiounit);
criterion_main!(benches);
```

## Cross-Language Verification

**Goal:** Ensure wire format compatibility across all language implementations.

### Verification Process

When `-bench` is used with multiple languages:

```bash
# Generate benchmarks for all languages
./sdp-gen -schema audiounit.sdp -bench audiounit.json -lang go -output testdata/generated/go/audiounit
./sdp-gen -schema audiounit.sdp -bench audiounit.json -lang cpp -output testdata/generated/cpp/audiounit
./sdp-gen -schema audiounit.sdp -bench audiounit.json -lang rust -output testdata/generated/rust/audiounit
./sdp-gen -schema audiounit.sdp -bench audiounit.json -lang rustexp -output testdata/generated/rustexp/audiounit
```

**Automated verification:**
1. Each language encodes the JSON → `{lang}_audiounit_sample.sdpb`
2. Compare all binaries byte-for-byte
3. Report differences or success

### Verification Output

**Success:**
```
✓ Cross-language verification passed
  All encoders produced identical output (110,245 bytes)
  Languages: go, cpp, rust, rustexp
  SHA256: a3f5c8e9d2b1f4a6e8c7d5a3b2f1e9c8d7a6b5f4e3d2c1b0a9f8e7d6c5b4a3f2
```

**Failure:**
```
✗ Cross-language verification FAILED
  go:      110,245 bytes  SHA256: a3f5c8e9...
  cpp:     110,245 bytes  SHA256: a3f5c8e9...  ✓ matches go
  rust:    110,247 bytes  SHA256: b2e4d7a1...  ✗ DIFFERS at offset 1024
  rustexp: 110,245 bytes  SHA256: a3f5c8e9...  ✓ matches go

Byte-level diff (rust vs go):
  Offset 1024: rust=0x05, go=0x03 (field: plugins[12].parameters[5].value)
  Offset 1025: rust=0x00, go=0x00
```

## Makefile Integration

### New Targets

```makefile
# Generate all benchmarks from canonical samples
.PHONY: generate-benchmarks
generate-benchmarks: build
	@echo "=== Generating Benchmarks ==="
	./sdp-gen -schema testdata/audiounit.sdp -bench testdata/audiounit.json -lang go -output testdata/generated/go/audiounit
	./sdp-gen -schema testdata/audiounit.sdp -bench testdata/audiounit.json -lang cpp -output testdata/generated/cpp/audiounit
	./sdp-gen -schema testdata/audiounit.sdp -bench testdata/audiounit.json -lang rust -output testdata/generated/rust/audiounit
	./sdp-gen -schema testdata/audiounit.sdp -bench testdata/audiounit.json -lang rustexp -output testdata/generated/rustexp/audiounit
	./sdp-gen -schema testdata/arrays.sdp -bench testdata/arrays.json -lang go -output testdata/generated/go/arrays
	./sdp-gen -schema testdata/arrays.sdp -bench testdata/arrays.json -lang cpp -output testdata/generated/cpp/arrays
	./sdp-gen -schema testdata/arrays.sdp -bench testdata/arrays.json -lang rust -output testdata/generated/rust/arrays
	./sdp-gen -schema testdata/arrays.sdp -bench testdata/arrays.json -lang rustexp -output testdata/generated/rustexp/arrays

# Verify cross-language wire format
.PHONY: verify-wire-format
verify-wire-format: generate-benchmarks
	@echo "=== Verifying Wire Format Compatibility ==="
	diff testdata/generated/go/audiounit/audiounit_sample.sdpb testdata/generated/cpp/audiounit/audiounit_sample.sdpb
	diff testdata/generated/go/audiounit/audiounit_sample.sdpb testdata/generated/rust/audiounit/audiounit_sample.sdpb
	diff testdata/generated/go/audiounit/audiounit_sample.sdpb testdata/generated/rustexp/audiounit/audiounit_sample.sdpb
	@echo "✓ All encoders produce identical output"

# Run all generated benchmarks
.PHONY: bench-all-generated
bench-all-generated: generate-benchmarks
	@echo "=== Running Generated Benchmarks ==="
	cd testdata/generated/go/audiounit && go test -bench=. -benchmem
	cd testdata/generated/cpp/audiounit && make bench
	cd testdata/generated/rust/audiounit && cargo bench
	cd testdata/generated/rustexp/audiounit && cargo bench
```

## Performance Metrics

### What to Measure

Each benchmark should report:

1. **Time per operation** - ns/iter (median or mean with stddev)
2. **Throughput** - MB/s (bytes / time)
3. **Memory allocations** - Count and total bytes (if language supports)
4. **Binary size** - Bytes (for size comparison)

### Example Output (Go)

```
BenchmarkAudioUnit_Encode-10           31847 ns/op    3.38 MB/s    15 allocs/op    110245 B/op
BenchmarkAudioUnit_Decode-10          255851 ns/op    0.43 MB/s   1759 allocs/op    450123 B/op
BenchmarkAudioUnit_Roundtrip-10       297029 ns/op    0.37 MB/s   1774 allocs/op    560368 B/op
```

### Example Output (C++)

```
AudioUnit Encode:    49.7 ns/iter    2127.5 MB/s    0 allocs
AudioUnit Decode:   112.3 ns/iter     981.2 MB/s    0 allocs
AudioUnit Roundtrip: 165.8 ns/iter     664.7 MB/s    0 allocs
```

### Example Output (Rust)

```
AudioUnit/encode     time: [28.47 µs 28.48 µs 28.49 µs]   thrpt: [3.69 MB/s]
AudioUnit/decode     time: [230.0 µs 230.2 µs 230.5 µs]   thrpt: [456 KB/s]
AudioUnit/roundtrip  time: [262.0 µs 262.2 µs 262.5 µs]   thrpt: [401 KB/s]
```

## Implementation Notes

### JSON Parser Selection

**Go:** `encoding/json` (stdlib, fast enough)

**C++:** Generate embedded JSON as C++ initialization (avoid runtime parsing):
```cpp
// Generated at build time from JSON
AudioUnit createSample() {
    AudioUnit au;
    au.plugins.push_back({1, "Reverb", /*...*/ });
    // ...
    return au;
}
```

**Rust:** `serde_json` (de facto standard, fast)

### Benchmark Stability

**Recommendations:**
- Run on idle machine (close browsers, heavy apps)
- Disable CPU frequency scaling if possible
- Report median (outlier resistant) or mean with stddev
- Warmup iterations before timed runs (prime caches)
- Use Criterion-style statistical analysis where available

### CI/CD Integration

Store baseline results and compare on each commit:

```yaml
# .github/workflows/benchmarks.yml
- name: Generate and run benchmarks
  run: |
    make generate-benchmarks
    make bench-all-generated > bench_results.txt
    
- name: Compare against baseline
  run: |
    python scripts/compare_bench.py baseline.txt bench_results.txt
    # Fail if performance regressed >5%
```

## User Documentation

### Quick Start Example

Add to `README.md`:

```markdown
## Benchmarking SDP with Your Data

Want to test SDP's performance with your own schema and data?

1. **Create a schema** (e.g., `mydata.sdp`):
   ```rust
   struct MyData {
       id: u64,
       name: string,
       values: []f64,
   }
   ```

2. **Create sample JSON** (`mydata.json`):
   ```json
   {
       "id": 12345,
       "name": "Sample",
       "values": [1.0, 2.0, 3.0, 4.0, 5.0]
   }
   ```

3. **Generate benchmarks**:
   ```bash
   ./sdp-gen -schema mydata.sdp -bench mydata.json -lang go -output go/mydata
   ```

4. **Run benchmarks**:
   ```bash
   cd go/mydata && go test -bench=.
   ```

Results show encode/decode speed and memory usage for YOUR data.
```

## Open Questions

1. **Should `-bench` support multiple JSON files?**
   ```bash
   -bench "testdata/samples/*.json"  # Glob pattern
   ```
   Pro: Test variety of data distributions  
   Con: More complex implementation, harder to interpret results

2. **Should we generate comparison reports?**
   Auto-generate markdown tables comparing Go vs C++ vs Rust?

3. **Should `-bench` also generate correctness tests?**
   Not just performance, but also round-trip equality checks?

4. **How to handle large JSON files (>100MB)?**
   Stream parsing? Size limits? Performance warnings?

5. **Should we support YAML/TOML as input?**
   Some users might prefer YAML for readability.

## Next Steps

1. **Phase 1: Core Implementation**
   - [ ] Add `-bench` flag parsing to `cmd/sdp-gen/main.go`
   - [ ] Implement JSON → native struct decoder (`internal/benchgen/`)
   - [ ] Generate `.sdpb` canonical binary
   - [ ] Test with `audiounit.json`, `arrays.json`

2. **Phase 2: Go Benchmarks**
   - [ ] Create Go benchmark template
   - [ ] Generate `_bench_test.go` files
   - [ ] Validate with existing hand-written benchmarks

3. **Phase 3: C++ Benchmarks**
   - [ ] Create C++ benchmark template
   - [ ] Generate CMakeLists.txt / Makefile
   - [ ] Test compilation and execution

4. **Phase 4: Rust Benchmarks**
   - [ ] Create Criterion template
   - [ ] Generate `benches/*.rs` files
   - [ ] Test with both `rust` and `rustexp`

5. **Phase 5: Cross-Language Verification**
   - [ ] Implement byte-for-byte comparison
   - [ ] Generate diff reports on mismatch
   - [ ] Add to CI/CD pipeline

6. **Phase 6: Documentation**
   - [ ] Update `QUICK_REFERENCE.md`
   - [ ] Add examples to `README.md`
   - [ ] Create `benchmarks/README.md`

## Conclusion

This specification provides a comprehensive plan for automated benchmark generation that:

- ✅ Solves canonical test data creation
- ✅ Enables user adoption ("test with your data")
- ✅ Verifies cross-language wire format
- ✅ Generates idiomatic language-specific benchmarks
- ✅ Integrates with existing Makefile structure
- ✅ Provides clear error messages
- ✅ Follows SDP's "simplicity first" principle

**Recommendation:** Proceed with Phase 1 implementation, validate with existing JSON samples, then expand to other languages.
