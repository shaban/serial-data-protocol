# SDP Release Candidate (RC) - Technical Specification

## Version: 0.2.0-rc1

**Status**: Release Candidate  
**Base Version**: 0.1.0 (Phase 1-4 complete)  
**Specification Date**: October 18, 2025

---

## Table of Contents

1. [Overview](#overview)
2. [Feature 1: Optional Struct Fields](#feature-1-optional-struct-fields)
3. [Feature 2: Self-Describing Message Mode](#feature-2-self-describing-message-mode)
4. [Feature 3: Compression via io.Writer/Reader](#feature-3-compression-via-iowriterreader)
5. [Backward Compatibility](#backward-compatibility)
6. [Migration Guide](#migration-guide)
7. [Performance Characteristics](#performance-characteristics)

---

## Overview

### Release Goals

The RC release adds **three production-critical features**:

1. **Optional Struct Fields** - Handle nullable/absent data structures
2. **Self-Describing Message Mode** - Enable server routing and protocol multiplexing  
3. **Compression Support** - 68-75% size reduction via io.Writer/Reader

### Design Principles

- ✅ **Zero breaking changes** - All features are additive
- ✅ **Backward compatible** - 0.1.0 code continues to work
- ✅ **Performance first** - <10% overhead for new features
- ✅ **Cross-language** - All features work in Go, Rust, Swift, C

### Version Compatibility Matrix

| Version | Byte Mode | Message Mode | Optional Fields | Compression |
|---------|-----------|--------------|-----------------|-------------|
| 0.1.0 | ✅ | ❌ | ❌ | ❌ |
| 0.2.0-rc1 | ✅ | ✅ | ✅ | ✅ |

---

## Feature 1: Optional Struct Fields

### Specification

**Schema Syntax**: Rust `Option<T>` and `Box<T>` (valid Rust for free tooling)

**Allowed**:
```rust
struct Plugin {
    name: String,
    metadata: Option<Metadata>,       // Optional field
    fallback: Option<Box<Plugin>>,    // Recursive optional (requires Box)
}
```

**Not Allowed**:
```rust
struct InvalidExamples {
    age: Option<u32>,              // ❌ No optional primitives
    tags: Option<Vec<String>>,     // ❌ No optional arrays (use empty array)
    items: Vec<Option<Item>>,      // ❌ No arrays of optionals
}
```

### Wire Format Specification

#### Encoding Rules

**Format**: `[presence: u8][data: variable bytes]`

**Presence Flag**:
- `0x00` = Field absent (nil/null/None) - **no data bytes follow**
- `0x01` = Field present - **data bytes follow**
- `0x02-0xFF` = Reserved for future use

**Examples**:

**Example 1: Optional present**
```
Struct: Plugin { name: "Reverb", metadata: Some(Metadata { version: 2 }) }

Wire bytes:
[06 00 00 00]           // name.len = 6 (u32 little-endian)
[52 65 76 65 72 62]     // "Reverb" (UTF-8)
[01]                    // metadata.presence = 1 (present)
[02 00 00 00]           // metadata.version = 2 (u32)

Total: 15 bytes
```

**Example 2: Optional absent**
```
Struct: Plugin { name: "Reverb", metadata: None }

Wire bytes:
[06 00 00 00]           // name.len = 6
[52 65 76 65 72 62]     // "Reverb"
[00]                    // metadata.presence = 0 (absent, no more bytes)

Total: 11 bytes (saved 4 bytes)
```

**Example 3: Recursive optional**
```
Struct: Node { value: 1, next: Some(Node { value: 2, next: None }) }

Wire bytes:
[01 00 00 00]           // value = 1
[01]                    // next.presence = 1 (present)
[02 00 00 00]           // next.value = 2
[00]                    // next.next.presence = 0 (absent)

Total: 10 bytes
```

### Generated Code Specification

#### Go Code Generation

**Type Mapping**:
```
Option<StructType>      → *StructType
Option<Box<StructType>> → *StructType  (Box is transparent in Go)
```

**Example**:
```go
// From schema: metadata: Option<Metadata>
type Plugin struct {
    Name     string
    Metadata *Metadata  // Pointer = optional (nil if absent)
}
```

**Encoding Logic**:
```go
func encodePlugin(src *Plugin, buf []byte, offset *int) error {
    // ... encode name ...
    
    // Encode optional metadata
    if src.Metadata == nil {
        buf[*offset] = 0  // presence = 0 (absent)
        *offset += 1
    } else {
        buf[*offset] = 1  // presence = 1 (present)
        *offset += 1
        encodeMetadata(src.Metadata, buf, offset)  // Encode the struct
    }
    
    return nil
}
```

**Decoding Logic**:
```go
func decodePlugin(dest *Plugin, data []byte, offset *int) error {
    // ... decode name ...
    
    // Decode optional metadata
    presence := data[*offset]
    *offset += 1
    
    if presence == 0 {
        dest.Metadata = nil  // Absent
    } else if presence == 1 {
        dest.Metadata = &Metadata{}  // Allocate
        decodeMetadata(dest.Metadata, data, offset)  // Decode into it
    } else {
        return ErrInvalidPresenceFlag
    }
    
    return nil
}
```

#### Rust Code Generation

**Type Mapping**:
```
Option<StructType>      → Option<StructType>  (inline if small)
Option<Box<StructType>> → Option<Box<StructType>>  (heap allocated)
```

**Encoding Logic**:
```rust
fn encode_plugin(src: &Plugin, buf: &mut Vec<u8>) -> Result<(), EncodeError> {
    // ... encode name ...
    
    // Encode optional metadata
    match &src.metadata {
        None => buf.push(0),  // presence = 0
        Some(metadata) => {
            buf.push(1);  // presence = 1
            encode_metadata(metadata, buf)?;
        }
    }
    
    Ok(())
}
```

**Decoding Logic**:
```rust
fn decode_plugin(buf: &[u8], offset: &mut usize) -> Result<Plugin, DecodeError> {
    // ... decode name ...
    
    // Decode optional metadata
    let presence = buf[*offset];
    *offset += 1;
    
    let metadata = match presence {
        0 => None,
        1 => Some(decode_metadata(buf, offset)?),
        _ => return Err(DecodeError::InvalidPresenceFlag),
    };
    
    Ok(Plugin { name, metadata })
}
```

#### C Code Generation

**Type Mapping**:
```
Option<StructType> → StructType*  (NULL if absent, malloc if present)
```

**Encoding Logic**:
```c
int encode_plugin(const Plugin *src, uint8_t *buf, size_t *offset) {
    // ... encode name ...
    
    // Encode optional metadata
    if (src->metadata == NULL) {
        buf[(*offset)++] = 0;  // presence = 0 (absent)
    } else {
        buf[(*offset)++] = 1;  // presence = 1 (present)
        encode_metadata(src->metadata, buf, offset);
    }
    
    return 0;
}
```

**Decoding Logic**:
```c
int decode_plugin(Plugin *dest, const uint8_t *data, size_t *offset) {
    // ... decode name ...
    
    // Decode optional metadata
    uint8_t presence = data[(*offset)++];
    
    if (presence == 0) {
        dest->metadata = NULL;  // Absent
    } else if (presence == 1) {
        dest->metadata = (Metadata*)malloc(sizeof(Metadata));
        decode_metadata(dest->metadata, data, offset);
    } else {
        return ERR_INVALID_PRESENCE_FLAG;
    }
    
    return 0;
}
```

### Size Overhead

**Per optional field**: 1 byte (presence flag)

**Example overhead**:
```
Struct with 3 optional fields:
  Base size: 100 bytes
  Optional overhead: 3 bytes (3 fields × 1 byte)
  Total: 103 bytes (3% overhead)
```

---

## Feature 2: Self-Describing Message Mode

### Specification

**Purpose**: Enable message routing, protocol multiplexing, and type identification without prior knowledge.

**Two Encoding Modes**:

1. **Byte Mode** (existing) - No header, minimal overhead
2. **Message Mode** (new) - Self-describing header, server-friendly

### Wire Format Specification

#### Mode 1: Byte Mode (Current - Unchanged)

```
[field 1 data][field 2 data]...[field N data]

No header, no type identification, no magic bytes.
Requires prior knowledge of message type.
```

**Use cases**: 
- ✅ **IPC (same process or machine)** - Type safety from shared imports
- ✅ **Embedded systems** - Minimal overhead, controlled environment
- ✅ **Performance-critical paths** - Zero header overhead
- ❌ **DO NOT use for persistence** - No corruption detection, no version info
- ❌ **DO NOT use across network boundaries** - No type validation
- ❌ **DO NOT use across language boundaries** - No schema version guarantee

**Safety guarantees**:
- Type safety requires both encoder/decoder to import same generated package
- Data lifetime should be ephemeral (microseconds to milliseconds)
- Single codebase ensures schema version match

#### Mode 2: Message Mode (New)

```
[Header: 10+ bytes]
  Magic:              "SDP" (3 bytes, ASCII 0x53 0x44 0x50)
  Version:            u8 (1 byte, currently 0x01)
  Mode:               u8 (1 byte, 0x02 = message mode)
  Type Name Length:   u8 (1 byte, max 255 chars)
  Type Name:          UTF-8 string (variable, e.g., "PluginRegistry")
  Payload Length:     u32 (4 bytes, little-endian)
  
[Payload: variable bytes]
  ... existing byte mode encoding ...
```

**Total header size**: `10 + len(TypeName)` bytes

**Examples**:

**Example 1: PluginRegistry message**
```
Header bytes:
[53 44 50]              // Magic "SDP"
[01]                    // Version 1
[02]                    // Mode 2 (message mode)
[0E]                    // Type name length = 14
[50 6C 75 67 69 6E 52 65 67 69 73 74 72 79]  // "PluginRegistry" (14 bytes)
[B0 B7 01 00]           // Payload length = 112,496 (little-endian u32)

Payload bytes:
[... 112,496 bytes of byte-mode encoded PluginRegistry ...]

Total: 24 + 112,496 = 112,520 bytes
Overhead: 24 bytes (0.021%)
```

**Example 2: Plugin message**
```
Header bytes:
[53 44 50]              // Magic "SDP"
[01]                    // Version 1
[02]                    // Mode 2
[06]                    // Type name length = 6
[50 6C 75 67 69 6E]     // "Plugin"
[1F 00 00 00]           // Payload length = 31

Payload bytes:
[... 31 bytes of Plugin data ...]

Total: 16 + 31 = 47 bytes
Overhead: 16 bytes (34% - high for small messages)
```

### When to Use Each Mode

**Byte Mode** (no header):
- ✅ **IPC (same process/machine)** - Type safety from shared imports, ephemeral data
- ✅ **Embedded systems** - Minimal overhead, controlled environment
- ✅ **Hot paths** - Zero header overhead for performance
- ❌ **DO NOT persist to disk** - No corruption detection, no versioning
- ❌ **DO NOT send across networks** - No type validation
- ❌ **DO NOT cross language boundaries** - No schema version guarantee

**Message Mode** (with header):
- ✅ **File storage** - Magic bytes detect corruption, version enables evolution
- ✅ **Network protocols** - Type name enables routing/dispatching  
- ✅ **Cross-service communication** - Self-describing for safety
- ✅ **Protocol multiplexing** - Multiple types on one connection
- ✅ **APIs/microservices** - Version control and type safety

### Header Validation Rules

**Magic Bytes**:
- MUST be `0x53 0x44 0x50` (ASCII "SDP")
- If not, return `ErrInvalidMagic`

**Version**:
- MUST be `0x01` for this spec
- Future versions may be `0x02`, `0x03`, etc.
- If unsupported, return `ErrUnsupportedVersion`

**Mode**:
- MUST be `0x02` for message mode
- `0x01` reserved for future byte-mode-with-version
- `0x03+` reserved for future modes
- If invalid, return `ErrInvalidMode`

**Type Name Length**:
- MUST be > 0 (at least 1 character)
- MUST be ≤ 255 (u8 max)
- Type name MUST match schema type exactly (case-sensitive)

**Payload Length**:
- MUST match actual payload size
- Decoder MUST verify: `len(data) == headerLength + payloadLength`
- If mismatch, return `ErrInvalidPayloadLength`

### Generated Code Specification

#### Encoding Functions

**Go**:
```go
// Byte mode (existing - unchanged)
func EncodePluginRegistry(src *PluginRegistry) ([]byte, error)

// Message mode (new)
func EncodePluginRegistryMessage(src *PluginRegistry) ([]byte, error) {
    // 1. Calculate payload size (existing logic)
    payloadSize := calculatePluginRegistrySize(src)
    
    // 2. Calculate header size
    typeName := "PluginRegistry"
    headerSize := 10 + len(typeName)
    
    // 3. Allocate buffer
    buf := make([]byte, headerSize + payloadSize)
    offset := 0
    
    // 4. Write header
    copy(buf[offset:], "SDP")           // Magic
    offset += 3
    buf[offset] = 0x01                  // Version
    offset += 1
    buf[offset] = 0x02                  // Mode
    offset += 1
    buf[offset] = uint8(len(typeName))  // Type name length
    offset += 1
    copy(buf[offset:], typeName)        // Type name
    offset += len(typeName)
    binary.LittleEndian.PutUint32(buf[offset:], uint32(payloadSize))
    offset += 4
    
    // 5. Write payload (existing logic)
    encodePluginRegistry(src, buf, &offset)
    
    return buf, nil
}
```

#### Decoding Functions

**Type-Specific Decoder**:
```go
func DecodePluginRegistryMessage(dest *PluginRegistry, data []byte) error {
    offset := 0
    
    // 1. Verify magic
    if !bytes.Equal(data[offset:offset+3], []byte("SDP")) {
        return ErrInvalidMagic
    }
    offset += 3
    
    // 2. Verify version
    version := data[offset]
    offset += 1
    if version != 0x01 {
        return ErrUnsupportedVersion
    }
    
    // 3. Verify mode
    mode := data[offset]
    offset += 1
    if mode != 0x02 {
        return ErrInvalidMode
    }
    
    // 4. Read type name
    typeNameLen := int(data[offset])
    offset += 1
    typeName := string(data[offset:offset+typeNameLen])
    offset += typeNameLen
    
    // 5. Verify type name
    if typeName != "PluginRegistry" {
        return ErrTypeMismatch
    }
    
    // 6. Read payload length
    payloadLen := binary.LittleEndian.Uint32(data[offset:])
    offset += 4
    
    // 7. Verify total length
    expectedLen := 10 + typeNameLen + int(payloadLen)
    if len(data) != expectedLen {
        return ErrInvalidPayloadLength
    }
    
    // 8. Decode payload (existing logic)
    return DecodePluginRegistry(dest, data[offset:])
}
```

**Dispatcher (Multi-Type Decoder)**:
```go
func DecodeMessage(data []byte) (interface{}, error) {
    offset := 0
    
    // 1-6. Parse header (same as above)
    // ...
    
    // 7. Dispatch based on type name
    switch typeName {
    case "PluginRegistry":
        var registry PluginRegistry
        err := DecodePluginRegistry(&registry, data[offset:])
        return &registry, err
        
    case "Plugin":
        var plugin Plugin
        err := DecodePlugin(&plugin, data[offset:])
        return &plugin, err
        
    case "Metadata":
        var metadata Metadata
        err := DecodeMetadata(&metadata, data[offset:])
        return &metadata, err
        
    default:
        return nil, ErrUnknownType
    }
}
```

### Error Types

**New Errors**:
```go
var ErrInvalidMagic = errors.New("sdp: invalid magic bytes (expected 'SDP')")
var ErrUnsupportedVersion = errors.New("sdp: unsupported protocol version")
var ErrInvalidMode = errors.New("sdp: invalid encoding mode")
var ErrTypeMismatch = errors.New("sdp: type name in header does not match expected type")
var ErrInvalidPayloadLength = errors.New("sdp: payload length mismatch")
var ErrUnknownType = errors.New("sdp: unknown type name in message header")
```

### Performance Characteristics

**Header overhead**:
- Small messages (100 bytes): ~20 bytes = 20% overhead
- Medium messages (10 KB): ~24 bytes = 0.24% overhead
- Large messages (110 KB): ~24 bytes = 0.021% overhead

**Speed impact** (estimated):
- Encoding: +5% (header write + length calculation)
- Decoding: +5-10% (header parsing + validation)

**Benchmark comparison**:
```
Byte mode:     128µs encode+decode
Message mode:  135µs encode+decode (+5.5%)

Still 9× faster than Protocol Buffers (1,300µs)
```

---

## Feature 3: Compression via io.Writer/Reader

### Specification

**Purpose**: Enable transparent compression without changing wire format.

**Approach**: Generate additional API functions that write to `io.Writer` / read from `io.Reader`.

### Generated API Functions

#### Encoding APIs

**Go**:
```go
// Existing - byte mode (returns []byte)
func EncodePluginRegistry(src *PluginRegistry) ([]byte, error)

// New - writer mode (writes to io.Writer)
func EncodePluginRegistryToWriter(src *PluginRegistry, w io.Writer) error {
    // Calculate size
    size := calculatePluginRegistrySize(src)
    
    // Allocate temporary buffer
    buf := make([]byte, size)
    offset := 0
    
    // Encode to buffer (existing logic)
    encodePluginRegistry(src, buf, &offset)
    
    // Write to io.Writer
    _, err := w.Write(buf)
    return err
}

// Alternative - message mode + writer
func EncodePluginRegistryMessageToWriter(src *PluginRegistry, w io.Writer) error {
    // Encode header + payload, write to io.Writer
}
```

#### Decoding APIs

**Go**:
```go
// Existing - byte mode (reads from []byte)
func DecodePluginRegistry(dest *PluginRegistry, data []byte) error

// New - reader mode (reads from io.Reader)
func DecodePluginRegistryFromReader(dest *PluginRegistry, r io.Reader) error {
    // Read entire stream into buffer
    buf, err := io.ReadAll(r)
    if err != nil {
        return err
    }
    
    // Decode from buffer (existing logic)
    return DecodePluginRegistry(dest, buf)
}

// Alternative - message mode + reader (read header first to get size)
func DecodePluginRegistryMessageFromReader(dest *PluginRegistry, r io.Reader) error {
    // 1. Read header (10 + typeNameLen bytes)
    // 2. Extract payload length
    // 3. Read exactly payloadLen bytes
    // 4. Decode payload
}
```

### Usage Examples

#### Compression (gzip)

**Encoding with compression**:
```go
import (
    "bytes"
    "compress/gzip"
)

// Encode + compress
var buf bytes.Buffer
gzipWriter := gzip.NewWriter(&buf)

err := EncodePluginRegistryToWriter(&registry, gzipWriter)
gzipWriter.Close()

compressed := buf.Bytes()  // 68% smaller!

// Size comparison:
// Uncompressed: 110 KB
// Compressed:    35 KB (31% of original)
```

**Decoding with decompression**:
```go
import (
    "bytes"
    "compress/gzip"
)

// Decompress + decode
gzipReader, err := gzip.NewReader(bytes.NewReader(compressed))
defer gzipReader.Close()

var decoded PluginRegistry
err = DecodePluginRegistryFromReader(&decoded, gzipReader)
```

#### File I/O

**Write to file**:
```go
import "os"

f, err := os.Create("registry.sdp")
defer f.Close()

err = EncodePluginRegistryToWriter(&registry, f)
```

**Read from file**:
```go
f, err := os.Open("registry.sdp")
defer f.Close()

var decoded PluginRegistry
err = DecodePluginRegistryFromReader(&decoded, f)
```

#### Network I/O

**Send over TCP**:
```go
import "net"

conn, err := net.Dial("tcp", "server:8080")
defer conn.Close()

err = EncodePluginRegistryToWriter(&registry, conn)
```

**Receive over TCP**:
```go
listener, err := net.Listen("tcp", ":8080")
conn, err := listener.Accept()
defer conn.Close()

var decoded PluginRegistry
err = DecodePluginRegistryFromReader(&decoded, conn)
```

#### Compression + Network (Combined)

**Server sends compressed messages**:
```go
import (
    "compress/gzip"
    "net"
)

conn, _ := net.Dial("tcp", "server:8080")
defer conn.Close()

// Wrap connection with gzip
gzipWriter := gzip.NewWriter(conn)
defer gzipWriter.Close()

// Encode directly to compressed stream
EncodePluginRegistryToWriter(&registry, gzipWriter)
```

**Client receives compressed messages**:
```go
import (
    "compress/gzip"
    "net"
)

listener, _ := net.Listen("tcp", ":8080")
conn, _ := listener.Accept()
defer conn.Close()

// Wrap connection with gzip reader
gzipReader, _ := gzip.NewReader(conn)
defer gzipReader.Close()

// Decode directly from compressed stream
var decoded PluginRegistry
DecodePluginRegistryFromReader(&decoded, gzipReader)
```

### Cross-Language Implementation

#### Rust

```rust
use std::io::{Write, Read};
use flate2::write::GzEncoder;
use flate2::read::GzDecoder;
use flate2::Compression;

// Encode to writer
pub fn encode_plugin_registry_to_writer<W: Write>(
    src: &PluginRegistry,
    w: &mut W
) -> Result<(), EncodeError> {
    let buf = encode_plugin_registry(src)?;
    w.write_all(&buf)?;
    Ok(())
}

// Decode from reader
pub fn decode_plugin_registry_from_reader<R: Read>(
    r: &mut R
) -> Result<PluginRegistry, DecodeError> {
    let mut buf = Vec::new();
    r.read_to_end(&mut buf)?;
    decode_plugin_registry(&buf)
}

// With compression
let mut encoder = GzEncoder::new(Vec::new(), Compression::default());
encode_plugin_registry_to_writer(&registry, &mut encoder)?;
let compressed = encoder.finish()?;
```

#### C

```c
#include <zlib.h>

// Write callback for gzip
int gzip_write_callback(const uint8_t *data, size_t len, void *userdata) {
    gzFile f = (gzFile)userdata;
    return gzwrite(f, data, len) == len ? 0 : -1;
}

// Encode to writer (callback-based)
int encode_plugin_registry_to_writer(
    const PluginRegistry *src,
    int (*write_fn)(const uint8_t*, size_t, void*),
    void *userdata
) {
    // Calculate size
    size_t size = calculate_plugin_registry_size(src);
    
    // Allocate buffer
    uint8_t *buf = malloc(size);
    size_t offset = 0;
    
    // Encode
    encode_plugin_registry(src, buf, &offset);
    
    // Write via callback
    int result = write_fn(buf, size, userdata);
    free(buf);
    return result;
}

// Usage with gzip
gzFile f = gzopen("registry.sdp.gz", "wb");
encode_plugin_registry_to_writer(&registry, gzip_write_callback, f);
gzclose(f);
```

### Compression Performance

**Size comparison** (measured with real AudioUnit data):

| Format | Size (KB) | Ratio vs JSON | Ratio vs SDP |
|--------|-----------|---------------|--------------|
| JSON | 627 KB | 100% | 570% |
| SDP Binary | 110 KB | 17.5% | 100% |
| SDP + gzip | 35 KB | 5.6% | 31% |
| SDP + zstd | 28 KB | 4.5% | 25% |

**Speed impact**:
- gzip compression: ~10× slower encode, ~5× slower decode
- zstd compression: ~8× slower encode, ~3× slower decode

**When to use compression**:
- ✅ Large messages (>10 KB) over network
- ✅ File storage (persistent data)
- ✅ Bandwidth-constrained environments
- ❌ IPC (overhead not worth it)
- ❌ Real-time audio (latency-sensitive)

---

## ⚠️ Critical Safety Warning: Byte Mode vs Message Mode

### Byte Mode (No Header) - IPC Only!

**Byte mode has NO magic bytes, NO version info, NO type identification.**

```
❌ UNSAFE for persistence:
   - File corruption → undetected
   - Wrong decoder → silent data corruption or panic
   - Schema changes → undefined behavior

❌ UNSAFE for networks:
   - No validation that data is SDP
   - No type checking before decode
   - Version mismatches undetected

✅ SAFE for IPC when:
   - Both sides import same generated package
   - Data lifetime is ephemeral (μs to ms)
   - Single codebase (guaranteed version match)
```

**Example of the danger**:
```go
// Sender encodes PluginRegistry
data := audiounit.EncodePluginRegistry(&registry)  // Raw bytes, no header

// ❌ DANGER: Save to disk
os.WriteFile("data.bin", data, 0644)

// Later, someone accidentally decodes with wrong type
var plugin audiounit.Plugin
audiounit.DecodePlugin(&plugin, data)  // 💥 Silent corruption or panic!
// No magic bytes to catch this mistake!
```

**Solution**: **Always use message mode for anything that leaves the process.**

```go
// ✅ SAFE: Message mode for persistence
data := audiounit.EncodePluginRegistryMessage(&registry)  // Has header
os.WriteFile("data.bin", data, 0644)

// Decoder validates type automatically
decoded, err := audiounit.DecodeMessage(data)
if err != nil {
    // Will catch wrong type, corruption, version mismatch
    log.Fatal(err)
}
registry := decoded.(*audiounit.PluginRegistry)  // Type-safe
```

---

## Backward Compatibility

### API Compatibility

**0.1.0 API (unchanged)**:
```go
// These functions work exactly as before
func EncodePluginRegistry(src *PluginRegistry) ([]byte, error)
func DecodePluginRegistry(dest *PluginRegistry, data []byte) error
```

**0.2.0-rc1 adds new functions** (no changes to existing):
```go
// New - optional fields (requires schema change)
// (handled transparently by generator)

// New - message mode
func EncodePluginRegistryMessage(src *PluginRegistry) ([]byte, error)
func DecodePluginRegistryMessage(dest *PluginRegistry, data []byte) error
func DecodeMessage(data []byte) (interface{}, error)

// New - writer/reader
func EncodePluginRegistryToWriter(src *PluginRegistry, w io.Writer) error
func DecodePluginRegistryFromReader(dest *PluginRegistry, r io.Reader) error
```

### Wire Format Compatibility

**Byte mode** (0.1.0):
```
[field 1][field 2]...[field N]
```

**Message mode** (0.2.0-rc1):
```
[header][byte mode payload]
```

**Compatibility**:
- ✅ 0.1.0 encoder → 0.1.0 decoder (existing, still works)
- ✅ 0.2.0 encoder (byte mode) → 0.1.0 decoder (works)
- ✅ 0.2.0 encoder (message mode) → 0.2.0 decoder (works)
- ❌ 0.2.0 encoder (message mode) → 0.1.0 decoder (fails - no header support)

**Recommendation**: Use byte mode for cross-version compatibility.

### Schema Compatibility

**0.1.0 schema**:
```rust
struct Plugin {
    name: String,
    metadata: Metadata,  // Required
}
```

**0.2.0-rc1 schema**:
```rust
struct Plugin {
    name: String,
    metadata: Option<Metadata>,  // Optional
}
```

**Compatibility**:
- ❌ **NOT compatible** - wire format changes (presence flag added)
- ⚠️ Requires coordination between encoder/decoder
- ✅ Can version schemas (use different type names: PluginV1, PluginV2)

---

## Migration Guide

### Migrating from 0.1.0 to 0.2.0-rc1

#### Step 1: Update Generator

```bash
# Build new generator
go build -o sdp-gen cmd/sdp-gen/main.go

# Verify version
./sdp-gen -version
# Output: sdp-gen version 0.2.0-rc1
```

#### Step 2: Regenerate Code

```bash
# Regenerate with new generator
./sdp-gen -schema audiounit.sdp -output generated/ -lang go
```

**Changes**:
- ✅ Existing code still works (byte mode unchanged)
- ✅ New functions available (message mode, writer/reader)
- ✅ Optional fields available (if schema uses `Option<T>`)

#### Step 3: Update Schema (Optional - for Optional Fields)

**Before (0.1.0)**:
```rust
struct Plugin {
    name: String,
    metadata: Metadata,  // Always required
}
```

**After (0.2.0-rc1)**:
```rust
struct Plugin {
    name: String,
    metadata: Option<Metadata>,  // Now optional
}
```

**Impact**: Wire format changes, requires coordinated deployment.

#### Step 4: Adopt New Features (Optional)

**Use message mode for routing**:
```go
// Before
data, _ := EncodePluginRegistry(&registry)
send(data)

// After (message mode)
data, _ := EncodePluginRegistryMessage(&registry)
send(data)

// Receiver (dispatcher)
msg, _ := DecodeMessage(receivedData)
switch v := msg.(type) {
case *PluginRegistry:
    handleRegistry(v)
case *Plugin:
    handlePlugin(v)
}
```

**Use compression for storage**:
```go
// Before
data, _ := EncodePluginRegistry(&registry)
os.WriteFile("registry.sdp", data, 0644)

// After (compressed)
f, _ := os.Create("registry.sdp.gz")
defer f.Close()
gzipWriter := gzip.NewWriter(f)
defer gzipWriter.Close()
EncodePluginRegistryToWriter(&registry, gzipWriter)
```

---

## Performance Characteristics

### Benchmark Comparisons

**0.1.0 (Byte Mode)**:
```
BenchmarkRealWorldAudioUnit-8    9,258   128,280 ns/op   320,137 B/op   4,639 allocs/op
```

**0.2.0-rc1 Estimates**:

**Byte Mode (unchanged)**:
```
BenchmarkRealWorldAudioUnit-8    9,258   128,280 ns/op   320,137 B/op   4,639 allocs/op
(identical to 0.1.0)
```

**Message Mode**:
```
BenchmarkRealWorldMessage-8      8,500   135,000 ns/op   320,161 B/op   4,640 allocs/op
(+5% slower, +24 bytes)
```

**Optional Fields** (with 3 optional fields, 2 present):
```
BenchmarkOptionalFields-8        8,800   133,500 ns/op   320,200 B/op   4,641 allocs/op
(+4% slower, +2 allocations for optional structs)
```

**Compression** (gzip):
```
BenchmarkCompressedEncode-8      1,200   950,000 ns/op   450,000 B/op   4,650 allocs/op
(~7× slower, compression overhead)

BenchmarkCompressedDecode-8      2,000   580,000 ns/op   380,000 B/op   4,655 allocs/op
(~4× slower, decompression overhead)
```

**Comparison with Protocol Buffers**:
```
Protocol Buffers:               ~1,300,000 ns/op
SDP 0.2.0-rc1 (byte mode):         128,280 ns/op  (10× faster)
SDP 0.2.0-rc1 (message mode):      135,000 ns/op  (9× faster)
SDP 0.2.0-rc1 (compressed):        950,000 ns/op  (1.4× faster)
```

### Memory Characteristics

**Heap allocations**:
- Byte mode: Same as 0.1.0
- Message mode: +1 allocation (header buffer)
- Optional fields: +1 allocation per present optional
- Compression: +1 allocation (compression buffer)

**Memory usage**:
- Byte mode: Same as 0.1.0
- Message mode: +24 bytes per message
- Optional fields: +1 byte per optional field
- Compression: +gzip overhead (~32 KB for encoder/decoder state)

---

## Appendix: Complete Wire Format Reference

### Primitives (Unchanged from 0.1.0)

| Type | Wire Size | Encoding | Example |
|------|-----------|----------|---------|
| `u8` | 1 byte | Little-endian | `42` → `[2A]` |
| `u16` | 2 bytes | Little-endian | `1000` → `[E8 03]` |
| `u32` | 4 bytes | Little-endian | `1000000` → `[40 42 0F 00]` |
| `u64` | 8 bytes | Little-endian | `1000000000` → `[00 CA 9A 3B 00 00 00 00]` |
| `i8` | 1 byte | Two's complement | `-42` → `[D6]` |
| `i16` | 2 bytes | Two's complement | `-1000` → `[18 FC]` |
| `i32` | 4 bytes | Two's complement | `-1000000` → `[C0 BD F0 FF]` |
| `i64` | 8 bytes | Two's complement | `-1000000000` → `[00 36 65 C4 FF FF FF FF]` |
| `f32` | 4 bytes | IEEE 754 binary32 | `3.14` → `[C3 F5 48 40]` |
| `f64` | 8 bytes | IEEE 754 binary64 | `3.14159265359` → `[EA 2E 44 54 FB 21 09 40]` |
| `bool` | 1 byte | `0x00` or `0x01` | `true` → `[01]` |
| `String` | 4 + N bytes | `[len: u32][UTF-8 bytes]` | `"Hi"` → `[02 00 00 00 48 69]` |

### Arrays (Unchanged from 0.1.0)

```
[count: u32][element 0][element 1]...[element N-1]

Example: [1, 2, 3] as Vec<u8>
[03 00 00 00]  // count = 3
[01]           // element 0
[02]           // element 1
[03]           // element 2
```

### Structs (Unchanged from 0.1.0)

```
[field 1][field 2]...[field N]

Fields in schema definition order, no padding, no tags.

Example: Point { x: u32, y: u32 }
Point { x: 100, y: 200 }
[64 00 00 00]  // x = 100
[C8 00 00 00]  // y = 200
```

### Optional Fields (New in 0.2.0-rc1)

```
[presence: u8][data if present]

presence = 0x00: absent (no data follows)
presence = 0x01: present (data follows)

Example: Option<Metadata>
Some(Metadata { version: 2 })
[01]           // presence = 1 (present)
[02 00 00 00]  // version = 2

None
[00]           // presence = 0 (no more bytes)
```

### Message Mode Header (New in 0.2.0-rc1)

```
[Magic: "SDP"]          3 bytes
[Version: u8]           1 byte
[Mode: u8]              1 byte
[Type Name Length: u8]  1 byte
[Type Name: UTF-8]      variable
[Payload Length: u32]   4 bytes
[Payload: byte mode]    variable

Example: Message containing PluginRegistry
[53 44 50]              // "SDP"
[01]                    // Version 1
[02]                    // Mode 2 (message)
[0E]                    // Type name length = 14
[50 6C ... 79]          // "PluginRegistry"
[B0 B7 01 00]           // Payload length = 112,496
[... payload ...]
```

---

## Conclusion

The **0.2.0-rc1 Release Candidate** adds three production-critical features while maintaining **100% backward compatibility** with 0.1.0 byte mode.

**Summary**:
- ✅ **Optional Fields**: Enable nullable structs with 1-byte overhead
- ✅ **Message Mode**: Self-describing headers for routing (0.02% overhead)
- ✅ **Compression**: 68-75% size reduction via io.Writer/Reader
- ✅ **Zero Breaking Changes**: All 0.1.0 code continues to work
- ✅ **Performance**: <10% overhead for new features

**Next Steps**:
1. Implement RC features (~5 days)
2. Test in production (2-4 weeks)
3. Gather feedback and iterate
4. Promote to 0.2.0 stable

---

**Specification Version**: 0.2.0-rc1  
**Last Updated**: October 18, 2025
