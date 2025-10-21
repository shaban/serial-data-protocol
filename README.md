# Serial Data Protocol (SDP)

**A schema-based binary serialization format with code generation for cross-language data transfer.**

Version 0.2.0-rc1

---

## What is SDP?

SDP generates efficient encoders and decoders from a schema definition. It produces predictable binary output with fixed-width integers and known sizes, making it suitable for scenarios where you control both ends of the communication.

**Design priorities:**
1. Predictable performance (no dynamic allocation during encode/decode)
2. Simple wire format (fixed-width integers, no varint encoding)
3. Cross-language support (Go ✅, Rust ✅, C ✅ PRODUCTION READY)
4. Zero runtime dependencies in generated code

**Not priorities:**
- Maximum compression (use external compression libraries)
- Schema evolution (breaking changes require coordination)
- Self-describing data (use message mode for type identification)

---

## When to Use SDP

SDP works well for:

✅ **IPC between processes you control** - Both encoder and decoder use generated code from the same schema  
✅ **Known data structures** - Schema is agreed upon at compile time  
✅ **Bulk data transfer** - Moving large datasets efficiently (audio plugin lists, device info, etc.)  
✅ **FFI scenarios** - Crossing language boundaries (C ↔ Go, Swift ↔ Go)  
✅ **Performance-critical paths** - Sub-microsecond encoding/decoding for typical structures  

### Performance Characteristics

**Verified benchmarks** from real-world data (62 audio plugins, 1,759 parameters, 115 KB):

| Operation | Time | Allocations | vs Protocol Buffers | vs FlatBuffers |
|-----------|------|-------------|---------------------|----------------|
| **Encode** | 39.3 µs | 1 | **6.1× faster** | **8.6× faster** |
| **Decode** | 98.1 µs | 4,638 | **3.2× faster** | 11,000× slower* |
| **Roundtrip** | 141.0 µs | 4,639 | **3.9× faster** | **2.4× faster** |
| **Memory usage** | 313 KB peak | — | **30% less RAM** | **65% less RAM** |

\* FlatBuffers has zero-copy decode (8.8 ns) but returns accessors, not native structs

Small message overhead (primitives):
- Byte mode: 51 bytes, 44 ns roundtrip
- Message mode: 61 bytes, 85 ns roundtrip (+19% size, +93% time)
- Optional absent: 3.15 ns decode (10× faster than present)

**C++ implementation performance** (fastest):
- **Encode:** 8.6 ns primitives, 49.7 ns complex (2.6-3× faster than Go)
- **Decode (zero-copy):** 3.37 ns (7.7× faster than Go, 2.8× faster than encode)
- Wire format structs and bulk memcpy optimizations for maximum performance

See [benchmarks/](benchmarks/) for detailed cross-protocol comparison and [CPP_IMPLEMENTATION.md](CPP_IMPLEMENTATION.md) for C++-specific details.

---

## When NOT to Use SDP

SDP is **not appropriate** for:

❌ **Public APIs** - No schema evolution, breaking changes affect all clients  
❌ **Unknown schemas** - Both sides must compile with identical schema  
❌ **Long-term storage** - Schema changes make old data unreadable (use message mode + versioning)  
❌ **Untrusted data** - Limited validation, designed for trusted environments  
❌ **Variable-size integers** - Uses fixed u32/u64, not space-efficient for small numbers  
❌ **Maximum compression** - Simpler than Protocol Buffers but larger (compose with gzip for 68% reduction)

> **Safety Note:** See [BYTE_MODE_SAFETY.md](BYTE_MODE_SAFETY.md) for important guidance on when to use byte mode vs message mode.

### Comparison with Other Formats

| Format | Use Case | SDP Advantage | SDP Disadvantage |
|--------|----------|---------------|------------------|
| **JSON** | Human-readable APIs | 17.5% the size, 50× faster | No human readability |
| **Protocol Buffers** | Public APIs, evolution | **6.1× faster encode, 3.9× faster roundtrip, 30% less RAM** | No schema evolution |
| **FlatBuffers** | Zero-copy access | **2.4× faster roundtrip, simpler API** | No zero-copy (returns structs) |
| **MessagePack** | Dynamic typing | Faster, known schema | No dynamic types |
| **Cap'n Proto** | RPC frameworks | Simpler wire format | No RPC features |

**Realistic positioning**: SDP is faster than Protocol Buffers for controlled environments but lacks its schema evolution capabilities. It's simpler than FlatBuffers but doesn't offer zero-copy access. Choose based on your constraints.

---

## Quick Start

### 1. Install the Code Generator

```bash
go install github.com/shaban/serial-data-protocol/cmd/sdp-gen@latest
```

### 2. Define Your Schema

Create `audio.sdp`:

```rust
// Audio plugin metadata
struct Plugin {
    id: u32,
    name: string,
    vendor: string,
    parameters: []Parameter,
}

struct Parameter {
    id: u32,
    name: string,
    min_value: f32,
    max_value: f32,
}
```

### 3. Generate Code

```bash
# For Go
sdp-gen -schema audio.sdp -output ./audio -lang go

# For Rust
sdp-gen -schema audio.sdp -output ./audio -lang rust

# For C (planned)
sdp-gen -schema audio.sdp -output ./audio -lang c
```

This generates:
- `types.go` - Data structures
- `encode.go` - Encoding functions + streaming writers
- `decode.go` - Decoding functions + streaming readers
- `errors.go` - Error types

### 4. Use the Generated Code

**Encoding:**
```go
plugin := audio.Plugin{
    ID:   1,
    Name: "Compressor",
    Vendor: "AudioCo",
    Parameters: []audio.Parameter{
        {ID: 1, Name: "Threshold", MinValue: -60, MaxValue: 0},
        {ID: 2, Name: "Ratio", MinValue: 1, MaxValue: 20},
    },
}

// Direct encoding (single allocation)
bytes, err := audio.EncodePlugin(&plugin)

// Streaming encoding (compose with compression, files, network)
var buf bytes.Buffer
gzipWriter := gzip.NewWriter(&buf)
err := audio.EncodePluginToWriter(&plugin, gzipWriter)
gzipWriter.Close()
```

**Decoding:**
```go
var plugin audio.Plugin

// Direct decoding
err := audio.DecodePlugin(&plugin, bytes)

// Streaming decoding (compose with decompression)
gzipReader, _ := gzip.NewReader(&buf)
err := audio.DecodePluginFromReader(&plugin, gzipReader)
```

---

## Core Features

### Fixed-Width Integers (By Design)

SDP uses fixed-width integers (u8, u16, u32, u64, i32, i64, f32, f64) instead of variable-length encoding:

**Why this matters:**
- ✅ Predictable wire size: Easy to calculate buffer sizes ahead of time
- ✅ Faster encoding: No varint logic, direct byte copying
- ✅ Simpler implementation: Straightforward encode/decode
- ❌ Larger output: `u32(5)` takes 4 bytes (Protocol Buffers would use 1 byte)

**When this is acceptable:**
- You're transferring moderate amounts of data (< 1 MB)
- You can compress the entire payload afterward (gzip reduces by ~68%)
- Performance matters more than wire size

**When this is problematic:**
- Mostly small integers (< 128) - Consider Protocol Buffers
- Network bandwidth constrained without compression
- Storing millions of records with many small values

### Optional Fields

Optional fields enable partial data and backward-compatible additions:

```rust
struct Config {
    required_field: u32,
    optional_field: ?string,  // May be absent
}
```

**Performance cost:**
- Present: 31.49 ns decode (48% slower than required)
- Absent: 3.15 ns decode (10× faster, zero allocation)
- Wire overhead: 1 byte per optional field

**When to use:**
- Fields that may not apply (e.g., error_message when no error)
- Adding fields to existing schemas without breaking old decoders
- Partial updates where only changed fields are sent

### Message Mode (Type Identification)

Message mode adds type IDs to enable discrimination and routing:

```rust
message ErrorMsg { code: u32, text: string }
message DataMsg { payload: []u8 }

// Generates dispatcher that routes by type ID
```

**Performance cost:**
- Message overhead: 10 bytes header (type ID + size)
- Roundtrip time: 85.54 ns vs 44.25 ns regular mode (+93%)
- Size overhead: 19.6% for small payloads, negligible for large

**When to use:**
- Persistent storage with multiple message types
- Event streams with heterogeneous events
- Protocol implementations with different packet types

**When not to use:**
- Single message type (no disambiguation needed)
- Every nanosecond counts (stick to regular mode)

### Streaming I/O

All types generate streaming encode/decode functions:

```go
// Implements standard library composition
func EncodePluginToWriter(src *Plugin, w io.Writer) error
func DecodePluginFromReader(dest *Plugin, r io.Reader) error
```

**Composition examples:**

```go
// Write to file
file, _ := os.Create("data.sdp")
defer file.Close()
audio.EncodePluginToWriter(&plugin, file)

// Compress before writing
var buf bytes.Buffer
gzipWriter := gzip.NewWriter(&buf)
audio.EncodePluginToWriter(&plugin, gzipWriter)
gzipWriter.Close()

// Send over network
conn, _ := net.Dial("tcp", "localhost:8080")
audio.EncodePluginToWriter(&plugin, conn)
```

**Philosophy:**
SDP does NOT bake in compression, file I/O, or network protocols. It provides standard `io.Writer`/`io.Reader` interfaces so you compose with the standard library or third-party compression (gzip, zstd, brotli, etc.).

---

## Cross-Language Workflow

SDP generates both encoders and decoders for each language. Cross-language communication requires two invocations:

```bash
# Generate Go code
sdp-gen -schema plugin.sdp -output ./go/plugin -lang go

# Generate C code (example, not yet implemented)
sdp-gen -schema plugin.sdp -output ./c/plugin -lang c
```

**How it works:**
1. C program encodes `Plugin` using generated C encoder → produces bytes
2. Bytes transferred via FFI, file, network, shared memory, etc.
3. Go program decodes bytes using generated Go decoder → reconstructs `Plugin`

**Wire format is language-agnostic**: Go encoder ↔ C decoder works because the binary layout is identical.

**Limitations:**
- Both sides must use the **same schema version**
- Breaking schema changes require recompiling both sides
- No automatic versioning or negotiation (use message mode for versioning)

---

## Schema Syntax

SDP uses Rust-like syntax with explicit typing:

```rust
// Primitive types
struct Primitives {
    unsigned: u8, u16, u32, u64,
    signed: i32, i64,
    floating: f32, f64,
    text: string,
    binary: []u8,
}

// Arrays (homogeneous)
struct Arrays {
    numbers: []u32,
    nested: [][]string,  // 2D array
}

// Optional fields (nullable)
struct Optional {
    required: u32,
    optional: ?string,
    optional_array: ?[]u32,
}

// Messages (type-discriminated)
message Event {
    timestamp: u64,
    data: string,
}
```

**Size limits** (enforced at decode):
- Strings: 10 MB max
- Arrays: 100,000 elements max
- Nesting: 20 levels max

These limits prevent unbounded allocation from malicious or corrupted data.

---

## Error Handling

Generated code returns descriptive errors:

```go
var plugin audio.Plugin
err := audio.DecodePlugin(&plugin, corruptedBytes)

// Errors include context:
// - "decode Plugin.name: string too long: 15000000 bytes (max 10485760)"
// - "decode Plugin.parameters: array too large: 200000 elements (max 100000)"
// - "buffer too small: need 156 bytes, have 100"
```

**Error categories:**
- Size limit violations (strings, arrays, nesting)
- Buffer underruns (incomplete data)
- Type mismatches (message mode only)

**Philosophy:**
Errors are for programming mistakes and data corruption, not for validation logic. SDP assumes trusted data sources. Add your own validation for business rules.

---

## Limitations and Trade-offs

SDP makes explicit trade-offs for simplicity and performance:

| Limitation | Reason | Workaround |
|------------|--------|------------|
| No schema evolution | Simple wire format | Version your messages, use optional fields |
| Fixed-width integers | Predictable performance | Compress output if size matters |
| No default values | Decoder simplicity | Check for zero values or use optional |
| No field tags | Positional encoding | Breaking changes require recompile |
| No enums (yet) | Not in 0.2.0-rc1 | Use u8/u16 with constants |
| No maps | Array-based simplicity | Use `[]KeyValue` structs |
| No unions | Type safety | Use message mode for variants |

**These are intentional design decisions**, not missing features. SDP prioritizes:
1. Implementation simplicity
2. Predictable performance  
3. Zero runtime dependencies

If you need schema evolution, use Protocol Buffers. If you need dynamic typing, use MessagePack. If you need maximum compression, use Cap'n Proto or FlatBuffers.

---

## Project Status

**Version 0.2.0-rc1** (Release Candidate)

**Implemented:**
- ✅ **Unified Go-based code generation** - Single tool for all languages
- ✅ **Go code generation** - Encoder, decoder, streaming I/O (415 tests)
- ✅ **Rust code generation** - Slice-based API, matches/exceeds Go performance
- ✅ **Cross-platform wire format** - Go ↔ Rust interop verified
- ✅ Optional fields with presence tracking
- ✅ Message mode with type discrimination
- ✅ Streaming I/O via stdlib interfaces
- ✅ Comprehensive validation and error reporting

**Performance (Apple M1):**
- Go encode: 26ns (primitives), 124ns (AudioUnit)
- Rust encode: **33ns (primitives), 119ns (AudioUnit)** - slice API is 2.8-4.4x faster than trait API
- Wire format: Identical, verified byte-for-byte

**Code Generation Architecture:**
```
Go Parser → Go Templates → {Go, Rust, C, Python, ...}
```

All code generation happens in Go using templates. This makes it easy to add new language backends.

**Planned:**
- C code generation (embedded systems, maximum portability)
- Python code generation (scripting, data analysis)
- Swift code generation (iOS, macOS apps)

**Not planned** (use other tools):
- RPC framework (use gRPC)
- Schema registry (use Confluent Schema Registry with custom serialization)
- Versioning system (implement yourself with message mode)

---

## Documentation

- **[DESIGN_SPEC.md](DESIGN_SPEC.md)** - Wire format specification and technical details
- **[QUICK_REFERENCE.md](QUICK_REFERENCE.md)** - API reference and examples
- **[PERFORMANCE_ANALYSIS.md](PERFORMANCE_ANALYSIS.md)** - Detailed performance measurements
- **[BYTE_MODE_SAFETY.md](BYTE_MODE_SAFETY.md)** - Safety guide (byte mode vs message mode)
- **[TESTING_STRATEGY.md](TESTING_STRATEGY.md)** - Testing approach

---

## License

MIT License - See LICENSE file

---

## Contributing

This project is in active development (Release Candidate 0.2.0-rc1). Before contributing:

1. Read `DESIGN_SPEC.md` for wire format details
2. Check `CHANGELOG.md` for recent changes and planned features
3. Run tests: `go test ./...` (should see 415 tests pass)
4. Follow existing code generation patterns

**Focus areas:**
- Additional language implementations (C is next priority)
- Real-world usage feedback
- Performance optimization (without breaking wire format)

**Out of scope:**
- Features that compromise simplicity
- Schema evolution features
- Dynamic typing support
