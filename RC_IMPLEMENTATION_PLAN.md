# SDP Release Candidate (RC) - Implementation Plan

## Version: 0.2.0-rc1

**Target Release Date**: TBD  
**Current Version**: 0.1.0 (Phase 1-4 complete, 238 tests passing)

---

## Executive Summary

The Release Candidate adds **three production-critical features** to SDP:

1. ✅ **Optional Struct Fields** - Handle nullable/absent data structures
2. ✅ **Self-Describing Message Mode** - Enable server routing and protocol multiplexing
3. ✅ **Compression Support** - 68-75% size reduction via io.Writer/Reader

**Total Implementation Effort**: ~4-5 days  
**Breaking Changes**: None (all features are additive)  
**Performance Impact**: <10% slowdown for message mode, 0% for byte mode

---

## Feature 1: Optional Struct Fields

### Problem Statement

Currently, all struct fields are **required**. Users cannot express:
- "Plugin might not have metadata loaded"
- "Effect might not have fallback chain"
- "Recursive references" (e.g., linked lists, trees)

### Solution: Rust `Option<T>` Syntax

**Schema (.sdp file)** - Valid Rust:
```rust
struct Plugin {
    name: String,
    metadata: Option<Metadata>,       // Optional field
    fallback: Option<Box<Plugin>>,    // Recursive optional
}

struct Metadata {
    version: u32,
    author: String,
}
```

**Generated Go Code** - Idiomatic:
```go
type Plugin struct {
    Name     string
    Metadata *Metadata  // Pointer = optional (nil if absent)
    Fallback *Plugin
}

type Metadata struct {
    Version uint32
    Author  string
}
```

**Generated Rust Code** - Native:
```rust
pub struct Plugin {
    pub name: String,
    pub metadata: Option<Metadata>,      // Inline or Box
    pub fallback: Option<Box<Plugin>>,
}
```

### Wire Format

**Encoding**:
```
[presence: u8][data: variable] (if presence = 1)
[presence: u8]                 (if presence = 0, no data)
```

**Example - Plugin with metadata**:
```
[06 00 00 00]        // name.len = 6
[52 65 76 65 72 62]  // "Reverb"
[01]                 // metadata present
[02 00 00 00]        // metadata.version = 2
[06 00 00 00]        // metadata.author.len = 6
[41 75 74 68 6F 72]  // "Author"
[00]                 // fallback absent

Total: 27 bytes
```

**Example - Plugin without metadata**:
```
[06 00 00 00]        // name.len = 6
[52 65 76 65 72 62]  // "Reverb"
[00]                 // metadata absent
[00]                 // fallback absent

Total: 12 bytes (saved 15 bytes)
```

**Overhead**: 1 byte per optional field

### Implementation Tasks

| Task | Component | Lines | Time |
|------|-----------|-------|------|
| **1.1** | Parser: Support `Option<T>` and `Box<T>` generics | ~100 | 1h |
| **1.2** | AST: Add `Optional bool` and `Boxed bool` to Type | ~10 | 15m |
| **1.3** | Validator: Verify Option only wraps struct types | ~20 | 30m |
| **1.4** | Go Generator: Emit `*T` for optional fields | ~30 | 30m |
| **1.5** | Encode/Decode: Add presence flag logic | ~50 | 1h |
| **1.6** | Tests: Add optional.sdp + roundtrip tests | ~60 | 1h |
| **Total** | | **~270 lines** | **~4h** |

### Validation Rules

```
✅ Allowed:
  field: Option<StructType>
  field: Option<Box<RecursiveType>>
  field: String  // Non-optional as before

❌ Not allowed:
  field: Option<u32>          // No optional primitives
  field: Option<Vec<Type>>    // No optional arrays (use empty array)
  field: Vec<Option<Type>>    // No arrays of optionals
```

### Performance Impact

- **Size**: +1 byte per optional field
- **Speed**: +1 branch per optional field (~5% slower decode for many optionals)
- **Memory**: +1 allocation per present optional (Go), inline (Rust)

---

## Feature 2: Self-Describing Message Mode

### Problem Statement

Current **byte mode** is optimal for IPC but lacks:
- ✅ Type identification (router must know type beforehand)
- ✅ Protocol versioning
- ✅ Message framing for streams

**Use cases needing message mode**:
1. Server routing (dispatch based on message type)
2. Protocol multiplexing (multiple types on one connection)
3. Load balancing (peek at type without full decode)
4. Logging/monitoring (log type without decode)

### Solution: Self-Describing Header

**Two modes** - backward compatible:

#### Mode 1: Byte Mode (Current - No Header)
```
... existing wire format ...
```
- ✅ Minimal overhead (0 bytes)
- ✅ Optimal for IPC (same process/machine)
- ✅ Fast (~128µs encode+decode)
- ⚠️ **NOT for persistence** - no corruption detection, no versioning
- ⚠️ **NOT for networks** - no type validation, no safety checks

#### Mode 2: Message Mode (New - With Header)
```
[Magic: "SDP"]          3 bytes
[Version: u8]           1 byte   (e.g., 0x01)
[Mode: u8]              1 byte   (0x02 = message mode)
[Type Name Length: u8]  1 byte
[Type Name: UTF-8]      variable (e.g., "PluginRegistry" = 14 bytes)
[Payload Length: u32]   4 bytes
[Payload: byte mode]    variable (existing format)
```

**Example - PluginRegistry message**:
```
[53 44 50]              // Magic "SDP"
[01]                    // Version 1
[02]                    // Mode 2 (message)
[0E]                    // Type name length = 14
[50 6C 75 67 69 6E ... 79]  // "PluginRegistry"
[B0 B7 01 00]           // Payload length = 112,496 bytes
[... existing byte mode payload ...]

Header overhead: 3 + 1 + 1 + 1 + 14 + 4 = 24 bytes
Percentage: 24 / 112,520 = 0.021%
```

### Generated API

**Encoding**:
```go
// Byte mode (current)
func EncodePluginRegistry(src *PluginRegistry) ([]byte, error)

// Message mode (new)
func EncodePluginRegistryMessage(src *PluginRegistry) ([]byte, error) {
    // Prepend header, then call byte mode encoder
}

// Writer mode (new - see Feature 3)
func EncodePluginRegistryToWriter(src *PluginRegistry, w io.Writer) error
```

**Decoding**:
```go
// Byte mode (current)
func DecodePluginRegistry(dest *PluginRegistry, data []byte) error

// Message mode - Type-specific
func DecodePluginRegistryMessage(dest *PluginRegistry, data []byte) error {
    // Verify header, extract payload, call byte mode decoder
}

// Message mode - Dispatcher
func DecodeMessage(data []byte) (interface{}, error) {
    header := parseHeader(data)
    switch header.TypeName {
    case "PluginRegistry":
        var registry PluginRegistry
        DecodePluginRegistryMessage(&registry, data)
        return &registry, nil
    case "Plugin":
        var plugin Plugin
        DecodePluginMessage(&plugin, data)
        return &plugin, nil
    default:
        return nil, ErrUnknownType
    }
}
```

### Implementation Tasks

| Task | Component | Lines | Time |
|------|-----------|-------|------|
| **2.1** | Wire format: Define header structure | ~0 | 30m (spec) |
| **2.2** | Encoder: Add `EncodeXMessage()` functions | ~100 | 2h |
| **2.3** | Decoder: Add header parsing logic | ~80 | 2h |
| **2.4** | Decoder: Generate `DecodeMessage()` dispatcher | ~120 | 3h |
| **2.5** | Errors: Add header-related error types | ~30 | 30m |
| **2.6** | Tests: Header validation, roundtrip, unknown type | ~100 | 3h |
| **Total** | | **~430 lines** | **~11h** |

### Performance Impact

**Benchmark estimates**:
```
Byte mode (current):
  BenchmarkRealWorldAudioUnit-8    9,258   128,280 ns/op

Message mode (estimated):
  BenchmarkRealWorldMessage-8      8,500   135,000 ns/op   (+5%)
```

**Still 9× faster than Protocol Buffers** (135µs vs 1,300µs)

### Migration Guide

**No breaking changes** - byte mode remains default:

```go
// Old code - still works
data, _ := EncodePluginRegistry(&registry)
DecodePluginRegistry(&decoded, data)

// New code - opt-in to message mode
data, _ := EncodePluginRegistryMessage(&registry)
DecodePluginRegistryMessage(&decoded, data)

// New code - dispatcher
msg, _ := DecodeMessage(data)
registry := msg.(*PluginRegistry)
```

---

## Feature 3: Compression Support via io.Writer/Reader

### Problem Statement

Binary format is 17.5% of JSON, but **compression** can reduce further:
- gzip: 68% reduction (110 KB → 35 KB = 31% of uncompressed)
- zstd: 75% reduction (110 KB → 28 KB = 25% of uncompressed)

Current API only supports `[]byte` - users must manually compress.

### Solution: io.Writer/Reader APIs

**Generated Encoding APIs**:
```go
// Existing - byte mode
func EncodePluginRegistry(src *PluginRegistry) ([]byte, error)

// New - writer mode
func EncodePluginRegistryToWriter(src *PluginRegistry, w io.Writer) error {
    // Write directly to io.Writer (no intermediate buffer)
}
```

**Generated Decoding APIs**:
```go
// Existing - byte mode
func DecodePluginRegistry(dest *PluginRegistry, data []byte) error

// New - reader mode
func DecodePluginRegistryFromReader(dest *PluginRegistry, r io.Reader) error {
    // Read directly from io.Reader
}
```

### Usage Examples

**Compression (Go)**:
```go
import "compress/gzip"

// Compress while encoding
var buf bytes.Buffer
gzipWriter := gzip.NewWriter(&buf)
EncodePluginRegistryToWriter(&registry, gzipWriter)
gzipWriter.Close()

compressed := buf.Bytes()  // 68% smaller!

// Decompress while decoding
gzipReader, _ := gzip.NewReader(bytes.NewReader(compressed))
var decoded PluginRegistry
DecodePluginRegistryFromReader(&decoded, gzipReader)
```

**File I/O (Go)**:
```go
// Write to file
f, _ := os.Create("registry.sdp")
defer f.Close()
EncodePluginRegistryToWriter(&registry, f)

// Read from file
f, _ := os.Open("registry.sdp")
defer f.Close()
var decoded PluginRegistry
DecodePluginRegistryFromReader(&decoded, f)
```

**Network (Go)**:
```go
// Send over TCP
conn, _ := net.Dial("tcp", "server:8080")
EncodePluginRegistryToWriter(&registry, conn)

// Receive over TCP
conn, _ := listener.Accept()
var decoded PluginRegistry
DecodePluginRegistryFromReader(&decoded, conn)
```

### Cross-Language Support

**Rust**:
```rust
use std::io::{Write, Read};
use flate2::write::GzEncoder;

// Encode + compress
let mut encoder = GzEncoder::new(Vec::new(), Compression::default());
encode_plugin_registry_to_writer(&registry, &mut encoder)?;
let compressed = encoder.finish()?;

// Decompress + decode
let mut decoder = GzDecoder::new(&compressed[..]);
let decoded = decode_plugin_registry_from_reader(&mut decoder)?;
```

**Swift**:
```swift
import Compression

// Encode + compress
let outputStream = OutputStream(toMemory: ())
outputStream.open()
let compressStream = outputStream.compressed(using: .lzfse)
encodePluginRegistryToStream(&registry, stream: compressStream)
```

**C**:
```c
#include <zlib.h>

// Encode + compress
gzFile f = gzopen("registry.sdp.gz", "wb");
encode_plugin_registry_to_writer(&registry, gzip_write_callback, f);
gzclose(f);

// Decompress + decode
gzFile f = gzopen("registry.sdp.gz", "rb");
PluginRegistry decoded;
decode_plugin_registry_from_reader(&decoded, gzip_read_callback, f);
gzclose(f);
```

### Implementation Tasks

| Task | Component | Lines | Time |
|------|-----------|-------|------|
| **3.1** | Encoder: Generate `EncodeXToWriter()` functions | ~120 | 3h |
| **3.2** | Decoder: Generate `DecodeXFromReader()` functions | ~150 | 4h |
| **3.3** | Size calculation: Support streaming mode | ~50 | 1h |
| **3.4** | Tests: File I/O, compression, network simulation | ~100 | 3h |
| **3.5** | Docs: Usage examples for gzip/zstd | ~30 | 1h |
| **Total** | | **~450 lines** | **~12h** |

### Performance Characteristics

**Compression ratios** (measured):
```
Original JSON:     627 KB
SDP Binary:        110 KB  (17.5% of JSON)
SDP + gzip:         35 KB  (5.6% of JSON, 31% of SDP)
SDP + zstd:         28 KB  (4.5% of JSON, 25% of SDP)
```

**Speed impact**:
- Compression: ~10× slower (gzip overhead)
- Decompression: ~5× slower (gzip overhead)
- **Still faster than Protocol Buffers** for typical use cases

### Design Decisions

**Why io.Writer/Reader instead of []byte?**
- ✅ Composable (gzip.Writer, file, network, etc.)
- ✅ Streaming-friendly (no intermediate allocation)
- ✅ Standard library pattern (encoding/json, encoding/binary use it)
- ✅ Works with all Go libraries (net, os, compress/*)

**Backward compatibility**:
- ✅ Keep existing `[]byte` API (90% of users)
- ✅ Add `io.Writer/Reader` for advanced cases (compression, streaming)

---

## RC Release Checklist

### Phase 1: Optional Fields (4 hours)
- [ ] Parser: Support `Option<T>` and `Box<T>` parsing
- [ ] AST: Add optional field tracking
- [ ] Validator: Enforce optional-only-for-structs rule
- [ ] Go Generator: Emit `*T` for optionals
- [ ] Encode/Decode: Add presence flag logic
- [ ] Tests: Add optional.sdp schema + roundtrip tests
- [ ] Docs: Update DESIGN_SPEC.md with optional syntax

### Phase 2: Message Mode (11 hours)
- [ ] Spec: Define wire format for message header
- [ ] Encoder: Generate `EncodeXMessage()` functions
- [ ] Decoder: Add header parsing and validation
- [ ] Decoder: Generate `DecodeMessage()` dispatcher
- [ ] Errors: Add `ErrInvalidMagic`, `ErrUnsupportedVersion`, `ErrUnknownType`
- [ ] Tests: Header roundtrip, unknown type handling
- [ ] Docs: Update wire format spec with message mode

### Phase 3: Compression/Writer APIs (12 hours)
- [ ] Encoder: Generate `EncodeXToWriter()` functions
- [ ] Decoder: Generate `DecodeXFromReader()` functions
- [ ] Size calculation: Support streaming size estimation
- [ ] Tests: File I/O, gzip compression, network simulation
- [ ] Docs: Add compression examples to README
- [ ] Docs: Add io.Writer/Reader patterns to LANGUAGE_IMPLEMENTATION_GUIDE.md

### Phase 4: Integration & Documentation (8 hours)
- [ ] Integration tests: All three features working together
- [ ] Performance benchmarks: Measure RC vs 0.1.0
- [ ] Update README.md with RC features
- [ ] Update QUICK_REFERENCE.md
- [ ] Create RC_SPEC.md (wire format + API changes)
- [ ] Migration guide for 0.1.0 → 0.2.0-rc1

### Phase 5: Release (2 hours)
- [ ] Tag release: `v0.2.0-rc1`
- [ ] Update CHANGELOG.md
- [ ] Build binaries for Mac/Linux/Windows
- [ ] Create GitHub release with binaries
- [ ] Announce RC for testing

---

## Total Effort Estimate

| Phase | Hours | Days (8h/day) |
|-------|-------|---------------|
| Optional Fields | 4h | 0.5 days |
| Message Mode | 11h | 1.5 days |
| Compression/Writer APIs | 12h | 1.5 days |
| Integration & Docs | 8h | 1 day |
| Release | 2h | 0.25 days |
| **Total** | **37h** | **~5 days** |

---

## Success Criteria

**Before releasing RC**:
- ✅ All 238 existing tests pass
- ✅ 50+ new tests for RC features pass
- ✅ No breaking changes to 0.1.0 API
- ✅ Performance <10% regression for byte mode
- ✅ Compression achieves >60% reduction with gzip
- ✅ Documentation complete for all features
- ✅ Cross-language examples (Go, Rust, C, Swift)

**RC Testing Period**:
- Use RC in production for 2-4 weeks
- Gather feedback from early adopters
- Fix bugs, refine APIs
- Promote to 0.2.0 stable if no major issues

---

## Post-RC Roadmap

### Version 0.3.0 (Future)
- C code generator (Phase 5)
- Swift code generator
- Rust code generator
- Python code generator

### Version 0.4.0 (Future)
- Union types (tagged unions)
- Default values for primitives
- Schema evolution (versioned wire format)
- Nested optionals (if use case arises)

### Version 1.0.0 (Stable)
- Production-tested for 6+ months
- At least 3 target languages (Go, C, Rust/Swift)
- Comprehensive benchmarks vs Protocol Buffers
- Full documentation + tutorials
