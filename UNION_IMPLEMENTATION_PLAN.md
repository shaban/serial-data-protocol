# Union Types Implementation Plan

**Date:** October 22, 2025  
**Goal:** Add union/variant types to SDP with minimal overhead (~4ns)  
**Availability:** Both byte mode and message mode

---

## Design Overview

### Schema Syntax (Rust-like)

```rust
// Basic union with unit variants
union Event {
    PluginLoaded,
    PluginUnloaded,
    AudioStarted,
}

// Union with data-carrying variants
union AudioEvent {
    PluginLoaded { plugin_id: u32, name: string },
    PluginUnloaded { plugin_id: u32 },
    ParameterChanged { plugin_id: u32, param_id: u32, value: f32 },
    AudioStarted { sample_rate: u32, buffer_size: u32 },
}

// Nested unions (unions in structs)
struct Message {
    timestamp: u64,
    event: AudioEvent,
}

// Optional unions
struct LogEntry {
    level: u8,
    message: string,
    context: ?AudioEvent,  // Optional union
}
```

### Wire Format

**Tag byte approach (minimal overhead):**

```
[u8 tag][variant data]

Examples:
- PluginLoaded (unit):         [0x00]
- PluginUnloaded (unit):        [0x01]
- ParameterChanged (data):      [0x02][u32 plugin_id][u32 param_id][f32 value]
```

**Overhead: 1 byte per union + variant data**

---

## Implementation Phases

### Phase 1: Parser & Validator (Day 1)

**Files to modify:**
- `internal/parser/parser.go` - Parse union syntax
- `internal/parser/ast.go` - Add UnionType and UnionVariant AST nodes
- `internal/validator/validator.go` - Validate union definitions

**Parser changes:**

```go
// ast.go
type UnionType struct {
    Name     string
    Variants []UnionVariant
}

type UnionVariant struct {
    Name   string
    Fields []Field  // Empty for unit variants
}

// parser.go
func (p *Parser) parseUnion() (*UnionType, error) {
    // Parse: union EventType { ... }
    // Parse variants with optional fields
}
```

**Validation rules:**
1. Union must have at least 2 variants
2. Maximum 256 variants (u8 tag limit)
3. Variant names must be unique within union
4. No recursive unions (union containing itself)
5. Variants can contain any valid SDP types (including other unions)

### Phase 2: Go Code Generation (Day 2-3)

**Files to modify:**
- `internal/generator/go/types.go` - Generate Go types
- `internal/generator/go/encode.go` - Generate encoding functions
- `internal/generator/go/decode.go` - Generate decoding functions

**Generated Go code:**

```go
// For: union AudioEvent { PluginLoaded { plugin_id: u32 }, PluginUnloaded { plugin_id: u32 } }

// Type definitions (interface + concrete types)
type AudioEvent interface {
    isAudioEvent()
    _tag() uint8
}

type AudioEventPluginLoaded struct {
    PluginId uint32
}
func (AudioEventPluginLoaded) isAudioEvent() {}
func (AudioEventPluginLoaded) _tag() uint8 { return 0 }

type AudioEventPluginUnloaded struct {
    PluginId uint32
}
func (AudioEventPluginUnloaded) isAudioEvent() {}
func (AudioEventPluginUnloaded) _tag() uint8 { return 1 }

// Encoding function
func EncodeAudioEvent(event AudioEvent, w io.Writer) error {
    // Write tag byte
    tag := event._tag()
    if err := binary.Write(w, binary.LittleEndian, tag); err != nil {
        return err
    }
    
    // Encode variant data
    switch v := event.(type) {
    case AudioEventPluginLoaded:
        if err := binary.Write(w, binary.LittleEndian, v.PluginId); err != nil {
            return err
        }
    case AudioEventPluginUnloaded:
        if err := binary.Write(w, binary.LittleEndian, v.PluginId); err != nil {
            return err
        }
    default:
        return fmt.Errorf("unknown AudioEvent variant")
    }
    
    return nil
}

// Decoding function
func DecodeAudioEvent(r io.Reader) (AudioEvent, error) {
    // Read tag byte
    var tag uint8
    if err := binary.Read(r, binary.LittleEndian, &tag); err != nil {
        return nil, err
    }
    
    // Decode variant based on tag
    switch tag {
    case 0: // PluginLoaded
        var pluginId uint32
        if err := binary.Read(r, binary.LittleEndian, &pluginId); err != nil {
            return nil, err
        }
        return AudioEventPluginLoaded{PluginId: pluginId}, nil
        
    case 1: // PluginUnloaded
        var pluginId uint32
        if err := binary.Read(r, binary.LittleEndian, &pluginId); err != nil {
            return nil, err
        }
        return AudioEventPluginUnloaded{PluginId: pluginId}, nil
        
    default:
        return nil, fmt.Errorf("invalid AudioEvent tag: %d", tag)
    }
}

// Byte mode variants (for buffers)
func EncodeAudioEventToBytes(event AudioEvent) ([]byte, error) {
    var buf bytes.Buffer
    if err := EncodeAudioEvent(event, &buf); err != nil {
        return nil, err
    }
    return buf.Bytes(), nil
}

func DecodeAudioEventFromBytes(data []byte) (AudioEvent, error) {
    return DecodeAudioEvent(bytes.NewReader(data))
}

// Message mode variants (if struct is marked as message)
func EncodeAudioEventMessage(event AudioEvent) ([]byte, error) {
    // Compute type ID from union name
    typeID := computeTypeID("AudioEvent")
    
    // Encode variant data
    variantData, err := EncodeAudioEventToBytes(event)
    if err != nil {
        return nil, err
    }
    
    // Wrap in message header
    return encodeMessageHeader(typeID, variantData), nil
}
```

### Phase 3: C++ Code Generation (Day 4-5)

**Files to modify:**
- `internal/generator/cpp/types.go` - Generate C++ types
- `internal/generator/cpp/encode.go` - Generate encoding functions
- `internal/generator/cpp/decode.go` - Generate decoding functions

**Generated C++ code:**

```cpp
// For: union AudioEvent { PluginLoaded { plugin_id: u32 }, PluginUnloaded { plugin_id: u32 } }

// Type definitions (std::variant wrapper)
struct AudioEventPluginLoaded {
    uint32_t plugin_id;
};

struct AudioEventPluginUnloaded {
    uint32_t plugin_id;
};

using AudioEvent = std::variant<
    AudioEventPluginLoaded,
    AudioEventPluginUnloaded
>;

// Helper to get tag
inline uint8_t get_tag(const AudioEvent& event) {
    return static_cast<uint8_t>(event.index());
}

// Encoding function
inline std::vector<uint8_t> encode_audio_event(const AudioEvent& event) {
    std::vector<uint8_t> buffer;
    
    // Write tag
    buffer.push_back(get_tag(event));
    
    // Encode variant
    std::visit([&buffer](auto&& variant) {
        using T = std::decay_t<decltype(variant)>;
        if constexpr (std::is_same_v<T, AudioEventPluginLoaded>) {
            encode_u32(buffer, variant.plugin_id);
        } else if constexpr (std::is_same_v<T, AudioEventPluginUnloaded>) {
            encode_u32(buffer, variant.plugin_id);
        }
    }, event);
    
    return buffer;
}

// Decoding function
inline AudioEvent decode_audio_event(const uint8_t* data, size_t size) {
    if (size < 1) {
        throw std::runtime_error("Buffer too small for AudioEvent tag");
    }
    
    uint8_t tag = data[0];
    const uint8_t* pos = data + 1;
    
    switch (tag) {
    case 0: { // PluginLoaded
        uint32_t plugin_id = decode_u32(pos);
        return AudioEventPluginLoaded{plugin_id};
    }
    case 1: { // PluginUnloaded
        uint32_t plugin_id = decode_u32(pos);
        return AudioEventPluginUnloaded{plugin_id};
    }
    default:
        throw std::runtime_error("Invalid AudioEvent tag: " + std::to_string(tag));
    }
}
```

### Phase 4: Rust Code Generation (Day 6-7)

**Files to modify:**
- `internal/generator/rust/types.go` - Generate Rust types
- `internal/generator/rust/encode.go` - Generate encoding functions
- `internal/generator/rust/decode.go` - Generate decoding functions

**Generated Rust code:**

```rust
// For: union AudioEvent { PluginLoaded { plugin_id: u32 }, PluginUnloaded { plugin_id: u32 } }

// Type definitions (native Rust enum)
#[derive(Debug, Clone, PartialEq)]
pub enum AudioEvent {
    PluginLoaded { plugin_id: u32 },
    PluginUnloaded { plugin_id: u32 },
}

impl AudioEvent {
    fn tag(&self) -> u8 {
        match self {
            AudioEvent::PluginLoaded { .. } => 0,
            AudioEvent::PluginUnloaded { .. } => 1,
        }
    }
}

// Encoding
pub fn encode_audio_event(event: &AudioEvent) -> Vec<u8> {
    let mut buffer = Vec::new();
    
    // Write tag
    buffer.push(event.tag());
    
    // Encode variant
    match event {
        AudioEvent::PluginLoaded { plugin_id } => {
            buffer.extend_from_slice(&plugin_id.to_le_bytes());
        }
        AudioEvent::PluginUnloaded { plugin_id } => {
            buffer.extend_from_slice(&plugin_id.to_le_bytes());
        }
    }
    
    buffer
}

// Decoding
pub fn decode_audio_event(data: &[u8]) -> Result<AudioEvent, DecodeError> {
    if data.is_empty() {
        return Err(DecodeError::BufferTooSmall);
    }
    
    let tag = data[0];
    let pos = &data[1..];
    
    match tag {
        0 => { // PluginLoaded
            let plugin_id = u32::from_le_bytes(pos[0..4].try_into()?);
            Ok(AudioEvent::PluginLoaded { plugin_id })
        }
        1 => { // PluginUnloaded
            let plugin_id = u32::from_le_bytes(pos[0..4].try_into()?);
            Ok(AudioEvent::PluginUnloaded { plugin_id })
        }
        _ => Err(DecodeError::InvalidTag(tag)),
    }
}
```

### Phase 5: Testing & Documentation (Day 8)

**New test files:**
- `testdata/schemas/unions.sdp` - Union test schemas
- `testdata/data/unions.json` - Test data
- `integration_test.go` - Union roundtrip tests

**Test cases:**
1. Unit variants (no data)
2. Data-carrying variants
3. Nested unions in structs
4. Optional unions
5. Cross-language wire format compatibility
6. Message mode with unions

---

## Detailed Schema Syntax

### Grammar Extension

```
union_decl ::= "union" IDENT "{" union_variant ("," union_variant)* ","? "}"

union_variant ::= IDENT                          // Unit variant
                | IDENT "{" field_list "}"       // Data variant

field_list ::= field ("," field)* ","?
field ::= IDENT ":" type_expr
```

### Examples

**Example 1: Simple event system**

```rust
union AudioEvent {
    Started,
    Stopped,
    ParameterChanged { param_id: u32, value: f32 },
}

struct Message {
    timestamp: u64,
    event: AudioEvent,
}
```

**Wire format:**
```
Message with AudioEvent::Started:
[u64 timestamp][u8 tag=0x00]

Message with AudioEvent::ParameterChanged:
[u64 timestamp][u8 tag=0x02][u32 param_id][f32 value]
```

**Example 2: Plugin state**

```rust
union PluginState {
    Uninitialized,
    Initialized { config: PluginConfig },
    Running { buffer_size: u32, sample_rate: f32 },
    Error { code: u32, message: string },
}

struct PluginConfig {
    name: string,
    version: u32,
}
```

**Example 3: Nested unions**

```rust
union Value {
    Int { value: i32 },
    Float { value: f32 },
    String { value: string },
    Array { values: []Value },  // Recursive!
}

struct Parameter {
    name: string,
    current: Value,
    default: Value,
}
```

**Validation:** Must check for circular references:
- ✅ `union Value { Array { values: []Value } }` - OK (via indirection)
- ❌ `union Value { Recursive { inner: Value } }` - ERROR (direct recursion)

---

## Wire Format Specification

### Tag Byte Encoding

```
Tag: uint8 (0-255)
- 0x00 = First variant
- 0x01 = Second variant
- ...
- 0xFF = 256th variant (maximum)
```

### Variant Data Encoding

**Unit variants:**
```
[tag]  // No data, just tag byte
```

**Data variants:**
```
[tag][field_0][field_1]...[field_n]
// Fields encoded in declaration order
// Same encoding rules as struct fields
```

### Size Calculation

```go
func calculateUnionSize(variant UnionVariant) int {
    size := 1  // Tag byte
    for _, field := range variant.Fields {
        size += calculateFieldSize(field)
    }
    return size
}

// Maximum size (for buffer allocation)
func calculateUnionMaxSize(union UnionType) int {
    maxSize := 1  // Tag byte
    for _, variant := range union.Variants {
        variantSize := 0
        for _, field := range variant.Fields {
            variantSize += calculateFieldSize(field)
        }
        if variantSize > maxSize-1 {
            maxSize = variantSize + 1
        }
    }
    return maxSize
}
```

---

## Integration with Existing Features

### 1. Unions in Structs (Byte Mode)

```rust
union Status { Ok, Error { code: u32 } }

struct Response {
    id: u64,
    status: Status,
}
```

**Generated Go:**

```go
type Response struct {
    Id     uint64
    Status Status
}

func EncodeResponse(src *Response) ([]byte, error) {
    var buf bytes.Buffer
    
    // Encode id
    if err := binary.Write(&buf, binary.LittleEndian, src.Id); err != nil {
        return nil, err
    }
    
    // Encode status (union)
    if err := EncodeStatus(src.Status, &buf); err != nil {
        return nil, err
    }
    
    return buf.Bytes(), nil
}
```

### 2. Unions in Message Mode

```rust
message AudioEvent {
    PluginLoaded { plugin_id: u32 },
    PluginUnloaded { plugin_id: u32 },
}
```

**Generated Go:**

```go
// Message mode wrapper
func EncodeAudioEventMessage(event AudioEvent) ([]byte, error) {
    // Compute type ID from union name
    typeID := computeTypeID("AudioEvent")
    
    // Encode variant data
    variantData, err := EncodeAudioEventToBytes(event)
    if err != nil {
        return nil, err
    }
    
    // Compute size
    size := uint32(len(variantData))
    
    // Build message: [u64 type_id][u32 size][variant_data]
    var buf bytes.Buffer
    binary.Write(&buf, binary.LittleEndian, typeID)
    binary.Write(&buf, binary.LittleEndian, size)
    buf.Write(variantData)
    
    return buf.Bytes(), nil
}

// Decoder (for dispatcher)
func DecodeAudioEventMessage(data []byte) (AudioEvent, error) {
    // Skip message header (already validated by dispatcher)
    variantData := data[12:]  // Skip 8-byte type_id + 4-byte size
    return DecodeAudioEventFromBytes(variantData)
}
```

### 3. Optional Unions

```rust
struct Config {
    name: string,
    error: ?Status,  // Optional union
}
```

**Wire format:**
```
[string name][u8 presence][union data if present]

If error is None:
[string name][0x00]

If error is Some(Status::Error{code: 42}):
[string name][0x01][u8 tag=0x01][u32 code=42]
```

### 4. Arrays of Unions

```rust
struct EventLog {
    events: []AudioEvent,
}
```

**Wire format:**
```
[u32 count][union_0][union_1]...[union_n]

Each union:
[u8 tag][variant data]
```

---

## Performance Characteristics

### Overhead Analysis

**Tag byte overhead:**
- Cost: 1 byte + 1 branch per union
- Encoding: ~2 ns (write byte + match)
- Decoding: ~2 ns (read byte + switch)
- Total: ~4 ns per union

**On AudioEvent example:**
```rust
struct Message {
    timestamp: u64,      // 8 bytes
    event: AudioEvent,   // 1 byte (tag) + variant data
}
```

**Encoding:**
- timestamp: ~2 ns
- union tag: ~2 ns
- variant data: same as regular struct
- Total overhead: ~2 ns (0.001%)

**Memory:**
- Go: interface{} overhead (~16 bytes per union)
- C++: std::variant overhead (~8 bytes discriminant)
- Rust: enum overhead (~1 byte discriminant, optimized)

### Comparison to Protocol Buffers `oneof`

```
Protocol Buffers oneof:
- Wire format: [varint field_num][varint which][data]
- Overhead: ~2 bytes + varint parsing (~20 ns)

SDP union:
- Wire format: [u8 tag][data]
- Overhead: 1 byte + direct read (~2 ns)

SDP is 10× faster for unions!
```

---

## Migration Path

### Existing Schemas

**No breaking changes:**
- Existing schemas continue to work
- Unions are opt-in feature
- Wire format unchanged for non-union types

### Adding Unions to Existing Schemas

**Before:**
```rust
struct Event {
    event_type: u8,  // 0=Started, 1=Stopped, 2=ParamChanged
    param_id: u32,   // Only used if event_type=2
    value: f32,      // Only used if event_type=2
}
```

**After:**
```rust
union EventData {
    Started,
    Stopped,
    ParamChanged { param_id: u32, value: f32 },
}

struct Event {
    data: EventData,
}
```

**Benefits:**
- Type safety at compile time
- Smaller wire size (no unused fields)
- Clearer API

---

## Implementation Checklist

### Day 1: Parser & AST
- [ ] Add UnionType and UnionVariant to AST
- [ ] Implement parseUnion() function
- [ ] Add union syntax tests
- [ ] Update validator for unions
- [ ] Check for recursive unions
- [ ] Enforce 256 variant limit

### Day 2-3: Go Generator
- [ ] Generate union interface + concrete types
- [ ] Generate EncodeUnion() functions
- [ ] Generate DecodeUnion() functions
- [ ] Generate byte mode helpers
- [ ] Generate message mode helpers (if applicable)
- [ ] Add union tests to integration_test.go

### Day 4-5: C++ Generator
- [ ] Generate std::variant definitions
- [ ] Generate encode functions with std::visit
- [ ] Generate decode functions with switch
- [ ] Add C++ union tests
- [ ] Verify cross-language compatibility

### Day 6-7: Rust Generator
- [ ] Generate native enum definitions
- [ ] Generate match-based encoding
- [ ] Generate match-based decoding
- [ ] Add Rust union tests
- [ ] Verify cross-language compatibility

### Day 8: Testing & Documentation
- [ ] Create testdata/schemas/unions.sdp
- [ ] Add union examples to QUICK_REFERENCE.md
- [ ] Update DESIGN_SPEC.md with union syntax
- [ ] Add cross-language wire format tests
- [ ] Benchmark union overhead
- [ ] Update CHANGELOG.md

---

## Timeline

**Total: 8 days (1.6 weeks)**

- Day 1: Parser & Validator
- Day 2-3: Go code generation
- Day 4-5: C++ code generation
- Day 6-7: Rust code generation
- Day 8: Testing & documentation

**Can be parallelized:**
- C++ and Rust generators can be done in parallel after Go is complete
- Total time: ~5-6 days with parallel work

---

## Success Criteria

✅ Unions work in byte mode  
✅ Unions work in message mode  
✅ Cross-language wire format compatibility  
✅ Overhead < 5 ns per union (measured)  
✅ Type-safe APIs in all languages  
✅ Comprehensive test coverage  
✅ Documentation updated  

---

## Next Steps

**Ready to start?** I can begin with Phase 1 (Parser & AST) and create:
1. `internal/parser/ast.go` - Add UnionType/UnionVariant
2. `internal/parser/parser.go` - Parse union syntax
3. `internal/validator/validator.go` - Validate unions
4. `testdata/schemas/unions.sdp` - Test schemas

**Should we proceed with implementation?**
