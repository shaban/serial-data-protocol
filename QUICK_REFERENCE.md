# SDP Quick Reference

**Version:** 0.2.0-rc1  
**For:** Developers using SDP-generated code

---

## Table of Contents

1. [Schema Syntax](#schema-syntax)
2. [Generated Code (Go)](#generated-code-go)
3. [Regular Structs](#regular-structs)
4. [Optional Fields](#optional-fields)
5. [Message Mode](#message-mode)
6. [Streaming I/O](#streaming-io)
7. [Error Handling](#error-handling)
8. [Common Patterns](#common-patterns)

---

## Schema Syntax

### Basic Struct

```rust
struct Plugin {
    id: u32,
    name: string,
    vendor: string,
}
```

### With Optional Fields

```rust
struct Plugin {
    id: u32,
    name: string,
    metadata: ?Metadata,  // Optional - may be absent
}

struct Metadata {
    version: string,
    author: string,
}
```

### Message (Self-Describing)

```rust
message PluginEvent {
    timestamp: u64,
    plugin_id: u32,
    event_type: u8,
}
```

### Primitive Types

| Schema Type | Go Type | Size |
|------------|---------|------|
| `u8` | `uint8` | 1 byte |
| `u16` | `uint16` | 2 bytes |
| `u32` | `uint32` | 4 bytes |
| `u64` | `uint64` | 8 bytes |
| `i32` | `int32` | 4 bytes |
| `i64` | `int64` | 8 bytes |
| `f32` | `float32` | 4 bytes |
| `f64` | `float64` | 8 bytes |
| `string` | `string` | 4 + len |
| `[]T` | `[]T` | 4 + elements |

---

## Generated Code (Go)

For schema file `plugin.sdp`, generate with:

```bash
sdp-gen -schema plugin.sdp -output ./plugin -lang go
```

Generated files:
- `types.go` - Struct definitions
- `encode.go` - Encoding functions + streaming writers
- `decode.go` - Decoding functions + streaming readers
- `errors.go` - Error types and size limits

---

## Regular Structs

### Encoding

```go
import "yourproject/plugin"

p := plugin.Plugin{
    ID:     1,
    Name:   "Compressor",
    Vendor: "AudioCo",
}

// Direct encoding (single allocation)
bytes, err := plugin.EncodePlugin(&p)
if err != nil {
    log.Fatal(err)
}

// Use bytes for FFI, file, network, etc.
```

### Decoding

```go
var p plugin.Plugin

err := plugin.DecodePlugin(&p, bytes)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Plugin: %s by %s\n", p.Name, p.Vendor)
```

### Arrays

```rust
struct PluginList {
    plugins: []Plugin,
}
```

```go
list := plugin.PluginList{
    Plugins: []plugin.Plugin{
        {ID: 1, Name: "Reverb", Vendor: "AudioCo"},
        {ID: 2, Name: "Delay", Vendor: "FXCorp"},
    },
}

bytes, err := plugin.EncodePluginList(&list)
```

---

## Optional Fields

### Schema

```rust
struct Config {
    port: u32,             // Required
    tls_config: ?TLSConfig,  // Optional
}

struct TLSConfig {
    cert_path: string,
    key_path: string,
}
```

### Encoding with Optional Present

```go
cfg := plugin.Config{
    Port: 8080,
    TLSConfig: &plugin.TLSConfig{  // Provide pointer
        CertPath: "/etc/certs/server.crt",
        KeyPath:  "/etc/certs/server.key",
    },
}

bytes, err := plugin.EncodeConfig(&cfg)
```

### Encoding with Optional Absent

```go
cfg := plugin.Config{
    Port:      8080,
    TLSConfig: nil,  // Explicitly nil (or omit)
}

bytes, err := plugin.EncodeConfig(&cfg)
// Wire format: 1 byte smaller (just presence=0)
```

### Decoding and Checking

```go
var cfg plugin.Config
err := plugin.DecodeConfig(&cfg, bytes)

if cfg.TLSConfig != nil {
    fmt.Printf("TLS enabled: %s\n", cfg.TLSConfig.CertPath)
} else {
    fmt.Println("TLS disabled")
}
```

### Performance

- **Absent:** 3.15 ns decode (10× faster, zero allocation)
- **Present:** 31.49 ns decode (+48% vs required field)
- Use absent optionals for maximum performance

---

## Message Mode

### Schema

```rust
message ErrorMsg {
    code: u32,
    text: string,
}

message DataMsg {
    payload: []u8,
}
```

### Encoding Messages

```go
// Encode error message
err := plugin.ErrorMsg{
    Code: 404,
    Text: "Plugin not found",
}

bytes, err := plugin.EncodeErrorMsg(&err)
// Wire format includes 10-byte header (type ID + size)
```

### Decoding with Dispatcher

```go
// Received bytes from file, network, etc.
msg, err := plugin.DispatchMessage(bytes)
if err != nil {
    log.Fatal(err)
}

// Type assertion to get concrete type
switch m := msg.(type) {
case *plugin.ErrorMsg:
    fmt.Printf("Error %d: %s\n", m.Code, m.Text)
case *plugin.DataMsg:
    fmt.Printf("Data: %d bytes\n", len(m.Payload))
default:
    fmt.Println("Unknown message type")
}
```

### Type ID Constants

```go
// Generated automatically
const (
    ErrorMsgTypeID uint64 = 0x... // FNV-1a hash of "ErrorMsg"
    DataMsgTypeID  uint64 = 0x...
)

// Check type ID directly
typeID := binary.LittleEndian.Uint64(bytes[0:8])
if typeID == plugin.ErrorMsgTypeID {
    // It's an error message
}
```

### Use Cases

- **Event streams:** Mix different event types in single stream
- **Storage:** Save different message types to same file
- **RPC:** Request/response with different message types
- **Protocols:** Multiple packet types needing discrimination

---

## Streaming I/O

All structs and messages generate streaming functions:

```go
func EncodePluginToWriter(src *Plugin, w io.Writer) error
func DecodePluginFromReader(dest *Plugin, r io.Reader) error
```

### File I/O

```go
// Write to file
file, err := os.Create("plugins.sdp")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

err = plugin.EncodePluginToWriter(&p, file)
if err != nil {
    log.Fatal(err)
}
```

```go
// Read from file
file, err := os.Open("plugins.sdp")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

var p plugin.Plugin
err = plugin.DecodePluginFromReader(&p, file)
```

### Compression

```go
import "compress/gzip"

// Compress before writing
var buf bytes.Buffer
gzipWriter := gzip.NewWriter(&buf)

err := plugin.EncodePluginToWriter(&p, gzipWriter)
if err != nil {
    log.Fatal(err)
}

gzipWriter.Close()

// buf now contains compressed data (typically 68% smaller)
```

```go
// Decompress while reading
gzipReader, err := gzip.NewReader(&buf)
if err != nil {
    log.Fatal(err)
}
defer gzipReader.Close()

var p plugin.Plugin
err = plugin.DecodePluginFromReader(&p, gzipReader)
```

### Network I/O

```go
import "net"

// Send over TCP
conn, err := net.Dial("tcp", "localhost:8080")
if err != nil {
    log.Fatal(err)
}
defer conn.Close()

err = plugin.EncodePluginToWriter(&p, conn)
```

```go
// Receive over TCP
listener, _ := net.Listen("tcp", ":8080")
conn, _ := listener.Accept()
defer conn.Close()

var p plugin.Plugin
err := plugin.DecodePluginFromReader(&p, conn)
```

### Pipe (In-Memory)

```go
// Simulate streaming between goroutines
r, w := io.Pipe()

go func() {
    defer w.Close()
    plugin.EncodePluginToWriter(&p, w)
}()

var received plugin.Plugin
plugin.DecodePluginFromReader(&received, r)
```

---

## Error Handling

### Common Errors

```go
var p plugin.Plugin
err := plugin.DecodePlugin(&p, bytes)

// Error examples:
// - "buffer too small: need 156 bytes, have 100"
// - "decode Plugin.name: string too long: 15000000 bytes (max 10485760)"
// - "decode Plugin.parameters: array too large: 200000 elements (max 100000)"
// - "unknown message type: 0x123456789abcdef0"
```

### Size Limits (Built-in)

| Limit | Value | Constant |
|-------|-------|----------|
| Max string size | 10 MB | `MaxStringSize` |
| Max array elements | 100,000 | `MaxArraySize` |
| Max nesting depth | 20 levels | `MaxNestingDepth` |

### Error Types

```go
import "errors"

err := plugin.DecodePlugin(&p, bytes)

if errors.Is(err, plugin.ErrUnexpectedEOF) {
    // Buffer too small
}

if errors.Is(err, plugin.ErrStringSizeExceeded) {
    // String exceeds 10 MB
}

if errors.Is(err, plugin.ErrArraySizeExceeded) {
    // Array exceeds 100,000 elements
}
```

### Validation Pattern

```go
var p plugin.Plugin
err := plugin.DecodePlugin(&p, bytes)
if err != nil {
    return fmt.Errorf("decode failed: %w", err)
}

// SDP validates size limits, you validate business rules
if p.Name == "" {
    return errors.New("plugin name cannot be empty")
}

if p.ID == 0 {
    return errors.New("plugin ID must be non-zero")
}
```

---

## Common Patterns

### FFI (C to Go)

**C side:**
```c
// Encode plugin in C
plugin_t p = {.id = 1, .name = "Reverb"};
uint8_t* bytes;
size_t len;
encode_plugin(&p, &bytes, &len);

// Transfer to Go
pass_to_go(bytes, len);
```

**Go side:**
```go
//export ReceivePlugin
func ReceivePlugin(bytes *C.uint8_t, length C.size_t) {
    data := C.GoBytes(unsafe.Pointer(bytes), C.int(length))
    
    var p plugin.Plugin
    err := plugin.DecodePlugin(&p, data)
    if err != nil {
        log.Printf("FFI decode error: %v", err)
        return
    }
    
    // Use plugin in Go
    processPlugin(&p)
}
```

### Batching Multiple Items

```go
var buf bytes.Buffer

// Write multiple plugins to buffer
for _, p := range plugins {
    err := plugin.EncodePluginToWriter(&p, &buf)
    if err != nil {
        log.Fatal(err)
    }
}

// Later: read them back
reader := bytes.NewReader(buf.Bytes())
for {
    var p plugin.Plugin
    err := plugin.DecodePluginFromReader(&p, reader)
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }
    
    processPlugin(&p)
}
```

### Versioning with Message Mode

```rust
message PluginV1 {
    id: u32,
    name: string,
}

message PluginV2 {
    id: u32,
    name: string,
    vendor: string,  // New field
}
```

```go
// Decoder handles both versions
msg, err := plugin.DispatchMessage(bytes)

switch m := msg.(type) {
case *plugin.PluginV1:
    // Handle old version
    fmt.Printf("V1 Plugin: %s\n", m.Name)
    
case *plugin.PluginV2:
    // Handle new version
    fmt.Printf("V2 Plugin: %s by %s\n", m.Name, m.Vendor)
}
```

### Compression Decision Tree

```go
func encodeWithOptionalCompression(p *plugin.Plugin, size int) []byte {
    if size < 1024 {
        // Small data: direct encoding (compression overhead not worth it)
        bytes, _ := plugin.EncodePlugin(p)
        return bytes
    }
    
    // Large data: compress (68% reduction typical)
    var buf bytes.Buffer
    gzipWriter := gzip.NewWriter(&buf)
    plugin.EncodePluginToWriter(p, gzipWriter)
    gzipWriter.Close()
    return buf.Bytes()
}
```

---

## Performance Tips

1. **Use optional absent fields** - 10× faster than present (3.15 ns vs 31.49 ns)
2. **Prefer regular structs over messages** - 2× faster (44 ns vs 85 ns roundtrip)
3. **Compress large payloads** - 68% size reduction with gzip
4. **Batch small messages** - Amortize overhead across multiple items
5. **Reuse buffers** - Pool `[]byte` buffers for encoding

---

## Cross-Language Workflow

### Generate code for both languages

```bash
# Generate Go code
sdp-gen -schema plugin.sdp -output ./go/plugin -lang go

# Generate C code (when available)
sdp-gen -schema plugin.sdp -output ./c/plugin -lang c
```

### Wire format is language-agnostic

- C encoder → bytes → Go decoder ✅
- Go encoder → bytes → C decoder ✅
- Both sides must use **identical schema version**

---

## Size Calculation

```go
// SDP generates size calculation functions
size := plugin.CalculatePluginSize(&p)
fmt.Printf("Wire size: %d bytes\n", size)

// Useful for:
// - Pre-allocating buffers
// - Estimating storage requirements
// - Deciding whether to compress
```
| `str` | Var | `[u32 len][UTF-8]` | Read len, read bytes |
| `[]T` | Var | `[u32 count][elems]` | Read count, decode each |

---

## 🔧 Encoder Pattern (Copy-Paste Template)

```
Public API:
    1. Calculate size
    2. Allocate buffer
    3. Call helper encoder
    4. Return buffer

Helper Encoder:
    For each field:
        1. Write to buffer at offset
        2. Increment offset by field size
```

---

## 🔍 Decoder Pattern (Copy-Paste Template)

```
Public API:
    1. Create DecodeContext
    2. Call helper decoder
    3. Return error or success

Helper Decoder:
    For each field:
        1. Check bounds (offset + size <= len)
        2. Read from buffer at offset
        3. Increment offset by field size
        4. Check array limits if needed
```

---

## ⚠️ Must-Have Error Checks

```
Decode every field:
    if (offset + SIZE > len) return ErrUnexpectedEOF;

Decode array:
    if (count > MaxArrayElements) return ErrArrayTooLarge;
    ctx.total += count;
    if (ctx.total > MaxTotalElements) return ErrTooManyElements;

Decode string:
    if (offset + 4 > len) return ErrUnexpectedEOF;
    len = read_u32_le(data + offset);
    if (offset + 4 + len > data_len) return ErrUnexpectedEOF;
```

---

## 🎯 Performance Targets

| Test | Target (M1 Base) |
|------|------------------|
| Simple struct (2 fields) | < 30 ns encode, < 25 ns decode |
| Nested struct (3 levels) | < 25 ns encode, < 20 ns decode |
| Small array (5 elements) | < 60 ns encode, < 140 ns decode |
| Real-world (1,759 params) | < 150 µs roundtrip |

**Must beat:** Protocol Buffers (1,300 µs) by at least **8×**

---

## ✅ Test Checklist

Wire Format Tests (11):
- [ ] All 12 primitive types with hand-crafted binary
- [ ] Truncated data detection
- [ ] Invalid string length detection
- [ ] 3-level nested structs
- [ ] Primitive arrays
- [ ] Empty arrays
- [ ] Oversized array rejection
- [ ] Struct arrays

Roundtrip Tests (8):
- [ ] All primitives with max/min values + Unicode
- [ ] Empty string edge case
- [ ] Nested structs with negative floats
- [ ] All array types with Unicode
- [ ] Empty arrays
- [ ] Arrays of structs
- [ ] Complex nested + arrays
- [ ] Large data (1000 elements)

---

## 🚀 C-Specific Quick Notes

**Memory:**
```c
uint8_t* encode_X(X* src, uint32_t* out_size) {
    uint32_t size = calculate_X_size(src);
    uint8_t* buf = malloc(size);
    // ... encode ...
    *out_size = size;
    return buf;  // Caller must free()
}

int decode_X(X* dst, uint8_t* data, uint32_t len) {
    // ... decode ...
    return 0;  // 0 = success, error code otherwise
}
```

**Little-Endian Helpers:**
```c
void write_u32_le(uint8_t* buf, uint32_t v) {
    buf[0]=v; buf[1]=v>>8; buf[2]=v>>16; buf[3]=v>>24;
}
uint32_t read_u32_le(uint8_t* buf) {
    return buf[0]|(buf[1]<<8)|(buf[2]<<16)|(buf[3]<<24);
}
```

---

## 🚀 Swift-Specific Quick Notes

**Memory:**
```swift
func encodeX(_ src: X) throws -> Data {
    let size = calculateXSize(src)
    var data = Data(capacity: size)
    // ... encode ...
    return data
}

func decodeX(_ dst: inout X, from data: Data) throws {
    var offset = 0
    // ... decode ...
}
```

**Little-Endian Helpers:**
```swift
extension Data {
    mutating func appendLE(_ value: UInt32) {
        var v = value.littleEndian
        withUnsafeBytes(of: &v) { append(contentsOf: $0) }
    }
    func readU32LE(at offset: Int) -> UInt32 {
        withUnsafeBytes { 
            $0.load(fromByteOffset: offset, as: UInt32.self)
        }.littleEndian
    }
}
```

---

## 📊 Final Validation

Before calling it done:
- [ ] All 238+ tests pass
- [ ] Benchmarks show 8-10× faster than Protocol Buffers
- [ ] plugins.json roundtrip works perfectly
- [ ] No memory leaks (C: valgrind, Swift: Instruments)
- [ ] Cross-language compatibility (Go ↔ C ↔ Swift)

---

**Time Budget:**
- C: 4 hours (careful with memory)
- Swift: 3 hours (easier than C)
- Total: 7 hours → Both done in one day! 🎉

