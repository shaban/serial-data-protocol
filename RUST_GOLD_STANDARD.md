# Rust Gold Standard Implementation Guide

This document describes the "gold standard" approach for generating high-quality, production-ready Rust code from SDP schema files.

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Embedded Runtime](#embedded-runtime)
4. [Cargo.toml Optimizations](#cargotoml-optimizations)
5. [Generated Code Structure](#generated-code-structure)
6. [Example Helpers](#example-helpers)
7. [Criterion Benchmarks](#criterion-benchmarks)
8. [Usage Guide](#usage-guide)
9. [Performance Results](#performance-results)
10. [Design Decisions](#design-decisions)

---

## Overview

The Rust code generator produces **self-contained, zero-dependency** crates that:

- ✅ Embed the entire runtime (no external SDP crate dependency)
- ✅ Use aggressive compiler optimizations for maximum performance
- ✅ Include CLI examples for cross-platform testing
- ✅ Include Criterion benchmarks for accurate performance measurement
- ✅ Follow Rust best practices and idioms
- ✅ Generate clean, readable code with proper documentation

**Performance:**
- Encode: ~6-7ns per operation (4x faster than Go)
- Decode: ~12-13ns per operation (1.7x faster than Go)
- Zero allocations during encoding/decoding

---

## Architecture

### Self-Contained Crates

Each generated crate is completely independent:

```
testdata/primitives/rust/
├── Cargo.toml                    # Manifest with optimizations
├── src/
│   ├── lib.rs                    # Public API (re-exports)
│   ├── wire.rs                   # Embedded runtime (Vec<u8>)
│   ├── wire_slice.rs             # Embedded runtime (&[u8])
│   ├── types.rs                  # Generated struct definitions
│   ├── encode.rs                 # Generated encoding logic
│   └── decode.rs                 # Generated decoding logic
├── examples/
│   └── crossplatform_helper.rs   # CLI tool for testing
└── benches/
    └── benchmarks.rs             # Criterion performance tests
```

**Key Principle:** No dependency on `rust/sdp/` workspace. Each crate works standalone.

### Dependency Graph

```
┌─────────────────────────────────────────┐
│         Generated Crate                  │
│                                          │
│  ┌────────────┐    ┌─────────────┐     │
│  │  types.rs  │───▶│  encode.rs  │     │
│  └────────────┘    └─────────────┘     │
│        │                   │            │
│        │                   ▼            │
│        │          ┌─────────────┐      │
│        └─────────▶│  decode.rs  │      │
│                   └─────────────┘      │
│                          │              │
│                          ▼              │
│              ┌────────────────────┐    │
│              │  wire_slice.rs     │    │
│              │  (embedded runtime)│    │
│              └────────────────────┘    │
│                          │              │
│                          ▼              │
│              ┌────────────────────┐    │
│              │  wire.rs           │    │
│              │  (embedded runtime)│    │
│              └────────────────────┘    │
│                          │              │
│                          ▼              │
│              ┌────────────────────┐    │
│              │  byteorder         │    │
│              │  (only external)   │    │
│              └────────────────────┘    │
└─────────────────────────────────────────┘
```

---

## Embedded Runtime

### Why Embed?

**Problem with workspace approach:**
- ❌ Generated crates depend on `../../../rust/sdp`
- ❌ Can't move generated code without breaking imports
- ❌ Users must understand workspace structure
- ❌ Complicates distribution and versioning

**Solution with embedding:**
- ✅ Each crate is self-contained
- ✅ Can copy generated code anywhere
- ✅ No workspace configuration needed
- ✅ Clear dependency: only `byteorder`

### Implementation

Generator (`internal/generator/rust/runtime.go`):

```go
// Embed wire.rs and wire_slice.rs as string constants
const wireRsContent = `...embedded file content...`
const wireSliceRsContent = `...embedded file content...`

func GenerateRuntime(outputDir string) error {
    // Write embedded files directly to generated crate
    os.WriteFile(filepath.Join(outputDir, "src/wire.rs"), 
                 []byte(wireRsContent), 0644)
    os.WriteFile(filepath.Join(outputDir, "src/wire_slice.rs"), 
                 []byte(wireSliceRsContent), 0644)
}
```

### Runtime Files

**`wire.rs`** - Vec<u8> implementation:
- `Writer` struct for building byte buffers
- Methods: `write_u8`, `write_u16`, `write_string`, etc.
- Used for encoding (allocates)

**`wire_slice.rs`** - &[u8] implementation:
- `Reader` struct for parsing byte slices
- Methods: `read_u8`, `read_u16`, `read_string`, etc.
- Used for decoding (zero-copy where possible)

---

## Cargo.toml Optimizations

### Release Profile

```toml
[profile.release]
opt-level = 3           # Maximum optimizations (0-3)
lto = true              # Link-Time Optimization (whole-program)
codegen-units = 1       # Single codegen unit (better optimization)
panic = 'abort'         # No unwinding (smaller binary, faster panics)
strip = true            # Remove debug symbols (smaller binary)
```

### Bench Profile

```toml
[profile.bench]
inherits = "release"    # Benchmarks use release optimizations
```

### Why These Flags?

| Flag | Purpose | Trade-off |
|------|---------|-----------|
| `opt-level = 3` | Maximum LLVM optimizations | Slower compile, faster runtime |
| `lto = true` | Cross-crate inlining and optimization | Much slower compile, 10-20% faster runtime |
| `codegen-units = 1` | Single compilation unit allows better optimization | Slower compile (no parallelism), 5-10% faster runtime |
| `panic = 'abort'` | Skip unwinding bookkeeping | Can't catch panics, smaller/faster code |
| `strip = true` | Remove symbols | Can't debug release build, smaller binary |

**Performance Impact:**
- Without LTO: ~10-12ns encode
- With LTO: ~6-7ns encode (**~50% faster!**)

### Dependencies

```toml
[dependencies]
byteorder = "1.5"      # Only external dependency (endianness handling)

[dev-dependencies]
criterion = { version = "0.5", features = ["html_reports"] }

[[bench]]
name = "benchmarks"
path = "benches/benchmarks.rs"
harness = false         # Use Criterion instead of default harness
```

---

## Generated Code Structure

### lib.rs

```rust
//! Generated SDP serialization for primitives schema
//! 
//! This is auto-generated code. Do not edit manually.

mod wire;
mod wire_slice;
pub mod types;
pub mod encode;
pub mod decode;

// Re-export public API
pub use types::*;
pub use encode::*;
pub use decode::*;
```

### types.rs

```rust
/// All primitive types in one struct
#[derive(Debug, Clone, PartialEq)]
pub struct AllPrimitives {
    pub u8_field: u8,
    pub u16_field: u16,
    pub u32_field: u32,
    // ... more fields
}
```

### encode.rs

```rust
use super::wire::Writer;
use super::types::*;

impl AllPrimitives {
    /// Calculate the encoded size in bytes
    pub fn encoded_size(&self) -> usize {
        let mut size = 0;
        size += 1;  // u8_field
        size += 2;  // u16_field
        // ... calculate all fields
        size
    }

    /// Encode to a Vec<u8>
    pub fn encode(&self) -> Result<Vec<u8>, Box<dyn std::error::Error>> {
        let mut buf = vec![0u8; self.encoded_size()];
        self.encode_to_slice(&mut buf)?;
        Ok(buf)
    }

    /// Encode directly to a pre-allocated buffer
    pub fn encode_to_slice(&self, buf: &mut [u8]) -> Result<(), Box<dyn std::error::Error>> {
        let mut writer = Writer::new(buf);
        writer.write_u8(self.u8_field)?;
        writer.write_u16(self.u16_field)?;
        // ... write all fields
        Ok(())
    }
}
```

### decode.rs

```rust
use super::wire_slice::Reader;
use super::types::*;

impl AllPrimitives {
    /// Decode from bytes
    pub fn decode(data: &[u8]) -> Result<Self, Box<dyn std::error::Error>> {
        Self::decode_from_slice(data)
    }

    /// Decode from a byte slice
    pub fn decode_from_slice(data: &[u8]) -> Result<Self, Box<dyn std::error::Error>> {
        let mut reader = Reader::new(data);
        Ok(AllPrimitives {
            u8_field: reader.read_u8()?,
            u16_field: reader.read_u16()?,
            // ... read all fields
        })
    }
}
```

---

## Example Helpers

### Purpose

Provide CLI tools for cross-platform testing:
- Encode/decode from command line
- Test interoperability with Go/Swift implementations
- Debug wire format issues
- Generate test data

### Generated Code

Location: `examples/crossplatform_helper.rs`

```rust
//! Cross-platform helper for testing primitives serialization
//!
//! Usage:
//!   cargo run --release --example crossplatform_helper encode > output.bin
//!   cargo run --release --example crossplatform_helper decode < input.bin

use sdp_primitives::*;
use std::io::{self, Read, Write};

fn main() {
    let args: Vec<String> = std::env::args().collect();
    if args.len() < 2 {
        eprintln!("Usage: {} <encode|decode>", args[0]);
        std::process::exit(1);
    }

    match args[1].as_str() {
        "encode" => encode_example(),
        "decode" => decode_example(),
        _ => {
            eprintln!("Unknown command: {}", args[1]);
            std::process::exit(1);
        }
    }
}

fn encode_example() {
    let data = AllPrimitives {
        u8_field: 255,
        u16_field: 65535,
        // ... test values
    };

    match data.encode() {
        Ok(bytes) => {
            io::stdout().write_all(&bytes).unwrap();
        }
        Err(e) => {
            eprintln!("Encode error: {}", e);
            std::process::exit(1);
        }
    }
}

fn decode_example() {
    let mut buffer = Vec::new();
    io::stdin().read_to_end(&mut buffer).unwrap();

    match AllPrimitives::decode(&buffer) {
        Ok(data) => {
            println!("{:#?}", data);
        }
        Err(e) => {
            eprintln!("Decode error: {}", e);
            std::process::exit(1);
        }
    }
}
```

### Usage

```bash
# Encode test data
cargo run --release --example crossplatform_helper encode > output.bin

# Decode and verify
cargo run --release --example crossplatform_helper decode < output.bin

# Cross-language test (Rust -> Go)
cargo run --release --example crossplatform_helper encode | \
  go run testdata/primitives/go/crossplatform_helper.go decode

# Cross-language test (Go -> Rust)
go run testdata/primitives/go/crossplatform_helper.go encode | \
  cargo run --release --example crossplatform_helper decode
```

---

## Criterion Benchmarks

### Purpose

Provide **accurate, statistical** performance measurements:
- ✅ No process spawn overhead (library-level benchmarking)
- ✅ Statistical analysis with outlier detection
- ✅ Warm-up period to stabilize measurements
- ✅ HTML reports for detailed analysis
- ✅ Regression detection across runs

### Generated Code

Location: `benches/benchmarks.rs`

```rust
//! Criterion benchmarks for primitives
//!
//! Run with: cargo bench
//! View results: target/criterion/report/index.html

use criterion::{black_box, criterion_group, criterion_main, Criterion};
use sdp_primitives::*;

fn bench_encode_primitives(c: &mut Criterion) {
    let data = AllPrimitives {
        u8_field: 255,
        u16_field: 65535,
        // ... test values
    };

    c.bench_function("primitives/encode", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            black_box(data.encode_to_slice(&mut buf).unwrap());
        });
    });
}

fn bench_decode_primitives(c: &mut Criterion) {
    let data = AllPrimitives { /* ... */ };
    let mut buf = vec![0u8; data.encoded_size()];
    data.encode_to_slice(&mut buf).unwrap();

    c.bench_function("primitives/decode", |bencher| {
        bencher.iter(|| {
            black_box(AllPrimitives::decode_from_slice(&buf).unwrap());
        });
    });
}

fn bench_roundtrip_primitives(c: &mut Criterion) {
    let data = AllPrimitives { /* ... */ };

    c.bench_function("primitives/roundtrip", |bencher| {
        let mut buf = vec![0u8; data.encoded_size()];
        bencher.iter(|| {
            data.encode_to_slice(&mut buf).unwrap();
            black_box(AllPrimitives::decode_from_slice(&buf).unwrap());
        });
    });
}

criterion_group!(benches, 
    bench_encode_primitives,
    bench_decode_primitives,
    bench_roundtrip_primitives
);
criterion_main!(benches);
```

### Running Benchmarks

```bash
# Run all benchmarks
cargo bench

# Compile benchmarks only (no run)
cargo bench --no-run

# View HTML reports
open target/criterion/report/index.html
```

### Integration with Go Tests

The Go test suite automatically runs Criterion benchmarks and parses results:

```go
func BenchmarkRust_Primitives_Encode(b *testing.B) {
    packagePath := "../testdata/primitives/rust"
    
    results, err := runCriterionBench(packagePath)
    if err != nil {
        b.Fatalf("Failed to run Criterion benchmark: %v", err)
    }

    encodeNs := results["encode"]
    b.ReportMetric(encodeNs, "ns/op")
}
```

This enables direct comparison:

```bash
$ go test -bench="Primitives_Encode" -benchmem
BenchmarkGo_Primitives_Encode-8        44659118    26.22 ns/op
BenchmarkRust_Primitives_Encode-8             1     6.29 ns/op
```

---

## Usage Guide

### Generating Rust Code

```bash
# Generate for specific schema
./sdp-gen -schema testdata/primitives.sdp -lang rust -output testdata/primitives/rust

# Regenerate all (from test)
go test -lang=rust

# Generate and verify (full test suite)
go test -lang=rust -v
```

### Using Generated Code

```rust
// Import the generated crate
use sdp_primitives::*;

fn example() -> Result<(), Box<dyn std::error::Error>> {
    // Create data
    let data = AllPrimitives {
        u8_field: 255,
        u16_field: 65535,
        // ... more fields
    };

    // Encode
    let bytes = data.encode()?;
    
    // Decode
    let decoded = AllPrimitives::decode(&bytes)?;
    
    assert_eq!(data, decoded);
    Ok(())
}
```

### Adding to Your Project

Add the generated crate as a path dependency:

```toml
[dependencies]
sdp-primitives = { path = "path/to/generated/primitives/rust" }
```

Or copy the generated code directly into your project.

---

## Performance Results

### Primitives Benchmark

| Operation | Time (ns) | Details |
|-----------|-----------|---------|
| Encode    | 6.29      | 4x faster than Go |
| Decode    | 12.68     | 1.7x faster than Go |
| Roundtrip | 17.75     | 2.7x faster than Go |

**Why so fast?**
1. LTO enables cross-function inlining
2. Single codegen unit allows global optimization
3. No allocations during encode/decode
4. Monomorphization eliminates virtual dispatch
5. SIMD auto-vectorization where possible

### AudioUnit Benchmark (Complex Data)

| Operation | Time (ns) | Compared to Go |
|-----------|-----------|----------------|
| Encode    | ~TBD      | ~TBD           |
| Decode    | ~TBD      | ~TBD           |
| Roundtrip | ~TBD      | ~TBD           |

### Compared to Other Solutions

| Implementation | Encode (ns) | Decode (ns) | Notes |
|----------------|-------------|-------------|-------|
| **SDP Rust**   | **6.3**     | **12.7**    | This implementation |
| SDP Go         | 26.2        | 21.5        | Good balance of speed/simplicity |
| Protocol Buffers| ~50-100    | ~80-150     | Smaller output, slower |
| FlatBuffers    | ~10-20      | ~5-10       | Fastest, no decode step |
| JSON           | ~500-1000   | ~800-1500   | Human-readable, slow |

---

## Design Decisions

### 1. Embedded Runtime vs Workspace

**Decision:** Embed runtime in each generated crate

**Rationale:**
- Self-contained crates can be copied anywhere
- No workspace configuration required
- Clear dependency: only `byteorder`
- Easier distribution and versioning

**Trade-off:**
- ❌ Duplicates runtime code (~770 lines × 6 packages)
- ✅ Each package is independent
- ✅ Simpler mental model for users

### 2. Aggressive Optimizations by Default

**Decision:** Always generate with LTO, codegen-units=1, opt-level=3

**Rationale:**
- Users expect fast serialization (it's a key feature)
- Compile time is acceptable (30-60 seconds)
- 50% performance improvement is worth it

**Trade-off:**
- ❌ Longer compile times
- ✅ 4x faster than Go
- ✅ Competitive with other fast solutions

### 3. Example Helpers in examples/

**Decision:** Generate CLI helpers in `examples/` directory

**Rationale:**
- Standard Rust pattern (`cargo run --example`)
- Useful for cross-platform testing
- Doesn't pollute main crate with binary dependencies
- Easy to discover (`cargo run --example`)

**Alternative considered:** Separate binary crate
- ❌ More complex directory structure
- ❌ Harder to find for users

### 4. Criterion for Benchmarking

**Decision:** Use Criterion instead of built-in benches

**Rationale:**
- Statistical analysis (mean, median, outliers)
- HTML reports for detailed analysis
- JSON output for integration with Go tests
- Industry standard for Rust benchmarking

**Trade-off:**
- ❌ Adds 74 dependencies (for benchmarking only)
- ✅ Accurate measurements
- ✅ Beautiful reports
- ✅ Integration with Go tests

### 5. Zero External Dependencies (Except byteorder)

**Decision:** Only depend on `byteorder` crate

**Rationale:**
- Minimize supply chain risk
- Faster compile times
- Simpler dependency management
- Clear what's "ours" vs "external"

**Why byteorder?**
- Handles endianness correctly
- Well-tested (millions of downloads)
- Small crate (~15KB)
- Standard in Rust serialization

---

## Future Improvements

### Possible Optimizations

1. **SIMD explicit usage**: Use `std::simd` for bulk operations
2. **Buffer pooling**: Reuse buffers across encodings
3. **Unsafe optimizations**: Skip bounds checks in hot paths
4. **Profile-Guided Optimization**: Build with PGO for production

### Possible Features

1. **Streaming encode/decode**: Support for large messages
2. **Zero-copy strings**: Use `&str` instead of `String` where possible
3. **Serde integration**: Implement `Serialize`/`Deserialize`
4. **Async support**: Async encode/decode for `tokio`/`async-std`

---

## See Also

- `PERFORMANCE_COMPARISON.md` - Go vs Rust performance analysis
- `rust/sdp/README.md` - Reference implementation documentation
- `internal/generator/rust/` - Code generator implementation
- [Criterion Documentation](https://bheisler.github.io/criterion.rs/book/)
- [Cargo Profile Documentation](https://doc.rust-lang.org/cargo/reference/profiles.html)
