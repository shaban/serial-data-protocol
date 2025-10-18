# Streaming Mode Feasibility Analysis

## Question

Would a **streaming encoder** make sense if we prefix messages with a **self-describing header**?

**Modes**:
1. **Byte mode** (current): For IPC, no header, minimal overhead
2. **Streaming mode** (proposed): For server use, self-describing header, allows decode-while-encoding

**Key concern**: Would there be too many changes to the decoding side?

---

## Current Architecture

### Encoding Strategy (Two-Pass)

```go
// Pass 1: Calculate size
func calculatePluginRegistrySize(src *PluginRegistry) int {
    size := 0
    size += 4  // plugins array count
    for _, plugin := range src.Plugins {
        size += calculatePluginSize(&plugin)
    }
    size += 4  // TotalPluginCount
    size += 4  // TotalParameterCount
    return size
}

// Pass 2: Allocate once and write
func EncodePluginRegistry(src *PluginRegistry) ([]byte, error) {
    size := calculatePluginRegistrySize(src)
    buf := make([]byte, size)  // Single allocation
    offset := 0
    encodePluginRegistry(src, buf, &offset)
    return buf, nil
}
```

**Characteristics**:
- ✅ Single allocation (optimal for IPC)
- ✅ Predictable memory usage
- ✅ Fast (~37µs for 1,759 parameters)
- ❌ Must traverse data twice
- ❌ Cannot start sending before fully encoded

### Decoding Strategy (Sequential)

```go
func DecodePluginRegistry(dest *PluginRegistry, data []byte) error {
    ctx := &DecodeContext{}
    offset := 0
    
    // Read array count
    if offset + 4 > len(data) { return ErrTruncated }
    count := binary.LittleEndian.Uint32(data[offset:])
    offset += 4
    
    // Decode each element
    dest.Plugins = make([]Plugin, count)
    for i := 0; i < int(count); i++ {
        if err := decodePlugin(&dest.Plugins[i], data, &offset, ctx); err != nil {
            return err
        }
    }
    
    // Continue reading fields...
    return nil
}
```

**Characteristics**:
- ✅ Requires entire message upfront
- ✅ Single pass through data
- ✅ Fast (~90µs for 1,759 parameters)
- ❌ Cannot start decoding partial data
- ❌ Must buffer entire message

---

## Proposed: Streaming Mode with Self-Describing Header

### Option A: Type-Length-Value (TLV) Format

**Wire format**:
```
[Header]
  Magic: "SDP\x01" (4 bytes)
  Mode: 0x02 (streaming mode) (1 byte)
  Type ID: u32 (4 bytes) - hash of struct name
  Total Length: u32 (4 bytes) - total message size
  
[Body]
  [Field 1 Tag: u8][Field 1 Length: u32][Field 1 Data: variable]
  [Field 2 Tag: u8][Field 2 Length: u32][Field 2 Data: variable]
  ...
```

**Encoder changes**:
```go
func EncodePluginRegistryStreaming(src *PluginRegistry, w io.Writer) error {
    // Write header
    header := make([]byte, 13)
    copy(header[0:4], "SDP\x01")
    header[4] = 0x02  // streaming mode
    binary.LittleEndian.PutUint32(header[5:9], typeIDForPluginRegistry)
    // Length calculated on the fly (set to 0 if unknown)
    binary.LittleEndian.PutUint32(header[9:13], 0)
    w.Write(header)
    
    // Write each field with tag + length + data
    // Field 1: Plugins array
    w.Write([]byte{0x01})  // tag for field 1
    arrayData := encodePluginsArray(src.Plugins)
    binary.Write(w, binary.LittleEndian, uint32(len(arrayData)))
    w.Write(arrayData)
    
    // Field 2: TotalPluginCount
    w.Write([]byte{0x02})  // tag for field 2
    binary.Write(w, binary.LittleEndian, uint32(4))  // length
    binary.Write(w, binary.LittleEndian, src.TotalPluginCount)
    
    // Continue for remaining fields...
    return nil
}
```

**Decoder changes**:
```go
func DecodePluginRegistryStreaming(dest *PluginRegistry, r io.Reader) error {
    // Read header
    header := make([]byte, 13)
    if _, err := io.ReadFull(r, header); err != nil {
        return err
    }
    
    // Verify magic + mode
    if string(header[0:4]) != "SDP\x01" || header[4] != 0x02 {
        return ErrInvalidHeader
    }
    
    typeID := binary.LittleEndian.Uint32(header[5:9])
    totalLen := binary.LittleEndian.Uint32(header[9:13])
    
    // Read fields as they arrive
    for {
        // Read tag
        tagBuf := make([]byte, 1)
        if _, err := io.ReadFull(r, tagBuf); err != nil {
            if err == io.EOF { break }
            return err
        }
        
        // Read length
        lenBuf := make([]byte, 4)
        io.ReadFull(r, lenBuf)
        fieldLen := binary.LittleEndian.Uint32(lenBuf)
        
        // Read field data
        fieldData := make([]byte, fieldLen)
        io.ReadFull(r, fieldData)
        
        // Dispatch to field decoder based on tag
        switch tagBuf[0] {
        case 0x01:  // Plugins
            decodePluginsArrayFromBytes(dest, fieldData)
        case 0x02:  // TotalPluginCount
            dest.TotalPluginCount = binary.LittleEndian.Uint32(fieldData)
        // ...
        }
    }
    
    return nil
}
```

### Option B: Chunked Encoding (Simpler)

**Wire format**:
```
[Header]
  Magic: "SDP\x01" (4 bytes)
  Mode: 0x02 (streaming mode) (1 byte)
  Type Name Length: u8 (1 byte)
  Type Name: string (variable)
  Total Length: u32 (4 bytes) - or 0xFFFFFFFF if unknown
  
[Body - Same as byte mode, no changes!]
  ... existing wire format ...
```

**Encoder changes**:
```go
func EncodePluginRegistryStreamingSimple(src *PluginRegistry, w io.Writer) error {
    // Write self-describing header
    typeName := "PluginRegistry"
    header := make([]byte, 10 + len(typeName))
    copy(header[0:4], "SDP\x01")
    header[4] = 0x02  // streaming mode
    header[5] = uint8(len(typeName))
    copy(header[6:6+len(typeName)], typeName)
    
    offset := 6 + len(typeName)
    
    // Option 1: Calculate size (same as byte mode)
    size := calculatePluginRegistrySize(src)
    binary.LittleEndian.PutUint32(header[offset:offset+4], uint32(size))
    w.Write(header)
    
    // Encode body exactly as byte mode
    buf := make([]byte, size)
    bodyOffset := 0
    encodePluginRegistry(src, buf, &bodyOffset)
    w.Write(buf)
    
    return nil
}
```

**Decoder changes**:
```go
func DecodeStreamingAuto(r io.Reader) (interface{}, error) {
    // Read header
    headerBuf := make([]byte, 6)
    io.ReadFull(r, headerBuf)
    
    // Verify magic + mode
    if string(headerBuf[0:4]) != "SDP\x01" || headerBuf[4] != 0x02 {
        return nil, ErrInvalidHeader
    }
    
    // Read type name
    typeNameLen := int(headerBuf[5])
    typeName := make([]byte, typeNameLen)
    io.ReadFull(r, typeName)
    
    // Read total length
    lenBuf := make([]byte, 4)
    io.ReadFull(r, lenBuf)
    totalLen := binary.LittleEndian.Uint32(lenBuf)
    
    // Read entire body
    body := make([]byte, totalLen)
    io.ReadFull(r, body)
    
    // Dispatch to appropriate decoder based on type name
    switch string(typeName) {
    case "PluginRegistry":
        var dest PluginRegistry
        if err := DecodePluginRegistry(&dest, body); err != nil {
            return nil, err
        }
        return &dest, nil
    // ... other types
    }
    
    return nil, ErrUnknownType
}

// Or use existing decoder if type is known:
func DecodePluginRegistryStreaming(dest *PluginRegistry, r io.Reader) error {
    // Read and verify header
    // ... header validation ...
    
    // Read body into buffer
    body := make([]byte, totalLen)
    io.ReadFull(r, body)
    
    // Use existing byte-mode decoder!
    return DecodePluginRegistry(dest, body)
}
```

---

## Comparison: Option A (TLV) vs Option B (Chunked)

| Aspect | Option A: TLV Format | Option B: Chunked/Header-Only |
|--------|---------------------|-------------------------------|
| **Encoder Complexity** | HIGH - Must add tags/length per field | LOW - Just prepend header |
| **Decoder Complexity** | HIGH - Tag dispatch, field reordering | LOW - Reuse existing decoder |
| **Wire Format Changes** | MAJOR - Every field tagged | MINOR - Just header |
| **Code Generation** | EXTENSIVE - New templates for TLV | MINIMAL - Add header generation |
| **True Streaming** | ✅ YES - Can decode partial messages | ❌ NO - Still needs full buffer |
| **Self-Describing** | ✅ YES - Type ID in header | ✅ YES - Type name in header |
| **Size Overhead** | HIGH - 5 bytes per field (tag + len) | LOW - 10-20 bytes per message |
| **Performance Impact** | -30% to -50% (tag processing) | -5% to -10% (header overhead) |
| **Migration Path** | BREAKING - Incompatible wire format | NON-BREAKING - Mode flag distinguishes |
| **Implementation Time** | 2-3 weeks (rewrite generators) | 2-3 days (add header logic) |

---

## Analysis: Do We Need True Streaming?

### What "Streaming" Actually Enables

**True streaming** (Option A) allows:
1. ✅ Start decoding before entire message arrives
2. ✅ Decode fields as they arrive (out of order possible)
3. ✅ Lower latency for large messages (>10MB)
4. ✅ Memory-bounded decoding (don't buffer entire message)

**Header-only streaming** (Option B) allows:
1. ✅ Self-describing messages (type name/ID in header)
2. ✅ Router/dispatcher can read type without full decode
3. ✅ Protocol versioning (version in header)
4. ❌ Still requires buffering entire message
5. ❌ Cannot decode incrementally

### AudioUnit Plugin Use Case Analysis

**Typical message sizes**:
- Single parameter update: ~100 bytes
- Plugin state: ~2-5 KB
- Full registry (62 plugins): ~110 KB

**Latency requirements**:
- Real-time audio: <10ms end-to-end
- Current encode+decode: 128µs (0.128ms)
- **Plenty of headroom** - not latency bound

**Network characteristics**:
- IPC (same machine): Gigabytes/sec, microsecond latency
- Local network: 1 Gbps = 125 MB/s
- 110 KB message: 0.88ms transfer time at 1 Gbps

**Conclusion**: 
- ❌ **True streaming NOT needed** - Messages arrive faster than decode time
- ✅ **Self-describing header IS useful** - Type dispatch for routing

### Server Use Case Analysis

**Typical scenarios**:
1. **Message routing**: Receive message → read type → route to handler
2. **Protocol multiplexing**: Multiple message types on one connection
3. **Load balancing**: Peek at message type without full decode
4. **Logging/monitoring**: Log message type without decode

**What's needed**:
- ✅ Self-describing header (type name/ID)
- ✅ Message length (for framing)
- ❌ NOT field-level tags (complexity not justified)

---

## Recommendation

### Implement Option B: Self-Describing Header Mode

**Wire format**:
```
[Mode 1: Byte Mode - No Header (Current)]
  ... existing wire format ...
  
[Mode 2: Message Mode - Self-Describing Header (Proposed)]
  Magic: "SDP" (3 bytes)
  Version: 0x01 (1 byte)
  Flags: 0x02 (1 byte) - bit 1 = self-describing mode
  Type Name Length: u8 (1 byte)
  Type Name: string (variable, 1-255 bytes)
  Payload Length: u32 (4 bytes)
  Payload: ... existing byte mode format ... (variable)
```

**Total header overhead**: 10 + len(typeName) bytes
- Example: "PluginRegistry" = 10 + 14 = **24 bytes** (0.02% overhead for 110KB message)

### Implementation Plan (2-3 days)

#### Day 1: Add Header Generation (4 hours)

```go
// internal/generator/golang/encode_gen.go

func GenerateEncoderWithMode(schema *parser.Schema) (string, error) {
    var buf strings.Builder
    
    for _, s := range schema.Structs {
        structName := ToGoName(s.Name)
        
        // Generate existing byte-mode encoder (unchanged)
        generateEncoderByteMode(&buf, &s, structName)
        
        // Generate new message-mode encoder
        buf.WriteString("\n")
        buf.WriteString("// Encode")
        buf.WriteString(structName)
        buf.WriteString("Message encodes ")
        buf.WriteString(structName)
        buf.WriteString(" with a self-describing header.\n")
        buf.WriteString("// Use this for network protocols, message queues, or when type ")
        buf.WriteString("information is needed.\n")
        buf.WriteString("func Encode")
        buf.WriteString(structName)
        buf.WriteString("Message(src *")
        buf.WriteString(structName)
        buf.WriteString(") ([]byte, error) {\n")
        
        // Calculate sizes
        buf.WriteString("\ttypeName := \"")
        buf.WriteString(s.Name)  // Use original schema name
        buf.WriteString("\"\n")
        buf.WriteString("\tpayloadSize := calculate")
        buf.WriteString(structName)
        buf.WriteString("Size(src)\n")
        buf.WriteString("\theaderSize := 10 + len(typeName)\n")
        buf.WriteString("\ttotalSize := headerSize + payloadSize\n\n")
        
        // Allocate buffer
        buf.WriteString("\tbuf := make([]byte, totalSize)\n\n")
        
        // Write header
        buf.WriteString("\t// Magic + Version\n")
        buf.WriteString("\tcopy(buf[0:3], \"SDP\")\n")
        buf.WriteString("\tbuf[3] = 0x01  // version\n")
        buf.WriteString("\tbuf[4] = 0x02  // flags: self-describing mode\n\n")
        
        buf.WriteString("\t// Type name\n")
        buf.WriteString("\tbuf[5] = uint8(len(typeName))\n")
        buf.WriteString("\tcopy(buf[6:6+len(typeName)], typeName)\n\n")
        
        buf.WriteString("\t// Payload length\n")
        buf.WriteString("\toffset := 6 + len(typeName)\n")
        buf.WriteString("\tbinary.LittleEndian.PutUint32(buf[offset:], uint32(payloadSize))\n")
        buf.WriteString("\toffset += 4\n\n")
        
        // Write payload using existing encoder
        buf.WriteString("\t// Payload (byte mode format)\n")
        buf.WriteString("\tif err := encode")
        buf.WriteString(structName)
        buf.WriteString("(src, buf, &offset); err != nil {\n")
        buf.WriteString("\t\treturn nil, err\n")
        buf.WriteString("\t}\n\n")
        
        buf.WriteString("\treturn buf, nil\n")
        buf.WriteString("}\n")
    }
    
    return buf.String(), nil
}
```

#### Day 1: Add Decoder Dispatcher (2 hours)

```go
// internal/generator/golang/decode_gen.go

func GenerateMessageDecoder(schema *parser.Schema) (string, error) {
    var buf strings.Builder
    
    // Generate header parsing
    buf.WriteString("// DecodeMessage decodes a self-describing SDP message.\n")
    buf.WriteString("// Returns (typeName, decodedValue, error).\n")
    buf.WriteString("func DecodeMessage(data []byte) (string, interface{}, error) {\n")
    buf.WriteString("\tif len(data) < 10 {\n")
    buf.WriteString("\t\treturn \"\", nil, ErrTruncated\n")
    buf.WriteString("\t}\n\n")
    
    buf.WriteString("\t// Verify header\n")
    buf.WriteString("\tif string(data[0:3]) != \"SDP\" {\n")
    buf.WriteString("\t\treturn \"\", nil, ErrInvalidMagic\n")
    buf.WriteString("\t}\n")
    buf.WriteString("\tif data[3] != 0x01 {\n")
    buf.WriteString("\t\treturn \"\", nil, ErrUnsupportedVersion\n")
    buf.WriteString("\t}\n")
    buf.WriteString("\tif data[4] != 0x02 {\n")
    buf.WriteString("\t\treturn \"\", nil, ErrInvalidMode\n")
    buf.WriteString("\t}\n\n")
    
    buf.WriteString("\t// Parse type name\n")
    buf.WriteString("\ttypeNameLen := int(data[5])\n")
    buf.WriteString("\tif len(data) < 6 + typeNameLen + 4 {\n")
    buf.WriteString("\t\treturn \"\", nil, ErrTruncated\n")
    buf.WriteString("\t}\n")
    buf.WriteString("\ttypeName := string(data[6:6+typeNameLen])\n\n")
    
    buf.WriteString("\t// Parse payload length\n")
    buf.WriteString("\toffset := 6 + typeNameLen\n")
    buf.WriteString("\tpayloadLen := binary.LittleEndian.Uint32(data[offset:])\n")
    buf.WriteString("\toffset += 4\n\n")
    
    buf.WriteString("\t// Extract payload\n")
    buf.WriteString("\tif len(data) < offset + int(payloadLen) {\n")
    buf.WriteString("\t\treturn typeName, nil, ErrTruncated\n")
    buf.WriteString("\t}\n")
    buf.WriteString("\tpayload := data[offset:offset+int(payloadLen)]\n\n")
    
    // Generate type dispatcher
    buf.WriteString("\t// Dispatch to type-specific decoder\n")
    buf.WriteString("\tswitch typeName {\n")
    for _, s := range schema.Structs {
        structName := ToGoName(s.Name)
        buf.WriteString("\tcase \"")
        buf.WriteString(s.Name)
        buf.WriteString("\":\n")
        buf.WriteString("\t\tvar dest ")
        buf.WriteString(structName)
        buf.WriteString("\n")
        buf.WriteString("\t\tif err := Decode")
        buf.WriteString(structName)
        buf.WriteString("(&dest, payload); err != nil {\n")
        buf.WriteString("\t\t\treturn typeName, nil, err\n")
        buf.WriteString("\t\t}\n")
        buf.WriteString("\t\treturn typeName, &dest, nil\n")
    }
    buf.WriteString("\tdefault:\n")
    buf.WriteString("\t\treturn typeName, nil, ErrUnknownType\n")
    buf.WriteString("\t}\n")
    buf.WriteString("}\n")
    
    return buf.String(), nil
}
```

#### Day 2: Update Error Types (1 hour)

```go
// internal/generator/golang/errors_gen.go

// Add new error types
var ErrInvalidMagic = errors.New("sdp: invalid magic bytes")
var ErrUnsupportedVersion = errors.New("sdp: unsupported protocol version")
var ErrInvalidMode = errors.New("sdp: invalid encoding mode")
var ErrUnknownType = errors.New("sdp: unknown type name in message header")
```

#### Day 2: Add Tests (4 hours)

```go
// Test message mode encoding/decoding
func TestMessageModeRoundtrip(t *testing.T) {
    src := &audiounit.PluginRegistry{
        TotalPluginCount: 62,
        TotalParameterCount: 1759,
        // ...
    }
    
    // Encode with message mode
    data, err := audiounit.EncodePluginRegistryMessage(src)
    if err != nil {
        t.Fatalf("Encode failed: %v", err)
    }
    
    // Verify header
    if string(data[0:3]) != "SDP" {
        t.Error("Invalid magic")
    }
    if data[3] != 0x01 {
        t.Error("Invalid version")
    }
    if data[4] != 0x02 {
        t.Error("Invalid mode")
    }
    
    // Decode with dispatcher
    typeName, decoded, err := audiounit.DecodeMessage(data)
    if err != nil {
        t.Fatalf("Decode failed: %v", err)
    }
    
    if typeName != "PluginRegistry" {
        t.Errorf("Wrong type: got %q, want %q", typeName, "PluginRegistry")
    }
    
    // Verify decoded data
    dst := decoded.(*audiounit.PluginRegistry)
    if dst.TotalPluginCount != src.TotalPluginCount {
        t.Error("Data mismatch")
    }
}

// Test header overhead
func TestMessageModeOverhead(t *testing.T) {
    // ... measure size difference ...
    // Should be ~24 bytes overhead
}

// Test dispatcher with unknown type
func TestMessageModeUnknownType(t *testing.T) {
    // ... craft message with unknown type ...
    // Should return ErrUnknownType
}
```

#### Day 3: Documentation (2 hours)

Update docs with:
- When to use byte mode vs message mode
- Wire format specification for message mode
- Performance implications
- Migration guide

---

## Performance Impact Analysis

### Message Mode Overhead

**For 110 KB message (PluginRegistry)**:
- Header: 24 bytes (0.02%)
- Encoding: Same two-pass algorithm + header write
- Decoding: Same decoder + header parsing
- **Expected slowdown**: 5-10% (header parsing overhead)

**Benchmark estimate**:
```
Current:
  BenchmarkRealWorldAudioUnit-8    9,258   128,280 ns/op   320,137 B/op   4,639 allocs/op

With message mode:
  BenchmarkRealWorldMessage-8      8,500   135,000 ns/op   320,161 B/op   4,640 allocs/op
                                           ^^^^^^^^         ^^^^^^^        ^^^^^^
                                           +5% slower       +24 bytes      +1 alloc (header)
```

**Still 9× faster than Protocol Buffers** (135µs vs 1,300µs).

### Memory Impact

- **Byte mode**: 0 overhead (current)
- **Message mode**: +24 bytes per message
- **Negligible** for typical use cases

---

## Conclusion

### ✅ YES - Message Mode Makes Sense

**Implement Option B**: Self-describing header mode

**Benefits**:
1. ✅ **Self-describing**: Type name in header enables routing/dispatching
2. ✅ **Minimal changes**: Reuse 100% of existing encoder/decoder logic
3. ✅ **Low overhead**: 24 bytes per message (<0.02% for typical sizes)
4. ✅ **Fast implementation**: 2-3 days vs 2-3 weeks for TLV
5. ✅ **Non-breaking**: Mode flag distinguishes byte vs message mode
6. ✅ **Server-friendly**: Perfect for message queues, routers, dispatchers

**Trade-offs**:
1. ❌ **Not true streaming**: Still requires full message buffer
2. ❌ **Cannot decode incrementally**: Must wait for entire message
3. ⚠️ **5-10% slower**: Header parsing overhead

**When to use each mode**:

| Use Case | Mode | Why |
|----------|------|-----|
| **AudioUnit IPC** | Byte | Minimal overhead, type known |
| **Shared memory** | Byte | Type known, zero copy |
| **Internal RPC** | Byte | Type in RPC envelope |
| **Message queue** | Message | Type routing needed |
| **Protocol multiplexing** | Message | Multiple types on one connection |
| **Network protocol** | Message | Self-describing for debugging |
| **File storage** | Message | Type identification |

### ❌ NO - True Streaming (TLV) Does NOT Make Sense

**Reasons**:
1. ❌ **Not needed**: Messages < 1MB arrive faster than decode time
2. ❌ **Too complex**: Complete rewrite of generators (2-3 weeks)
3. ❌ **Performance hit**: 30-50% slower due to tag processing
4. ❌ **Breaking change**: Incompatible wire format
5. ❌ **Use case unclear**: Real-time audio doesn't need >10MB messages

**When you WOULD need true streaming**:
- Multi-GB messages (video files, ML models)
- Satellite/cellular links (high latency, need incremental decode)
- Memory-constrained devices (cannot buffer full message)
- **None of these apply to AudioUnit plugin communication**

---

## Implementation Recommendation

**Phase 1** (Now): Implement message mode with self-describing header
- Estimated time: 2-3 days
- Risk: Low (reuses existing code)
- Value: High (enables server use cases)

**Phase 2** (Future, if needed): Consider true streaming
- Only if use case emerges (>10MB messages, high-latency links)
- Estimated time: 2-3 weeks
- Risk: High (major refactor)
- Value: Low (no current use case)

**Start with Phase 1** - you can always add true streaming later if needed, but the self-describing header gives you 80% of the benefit with 10% of the complexity.
