# SDP Language Implementation Guide

**Version:** 1.0  
**Date:** October 18, 2025  
**Purpose:** Step-by-step guide for implementing SDP code generators for any programming language

---

## Table of Contents

1. [Overview](#overview)
2. [Wire Format Specification](#wire-format-specification)
3. [Language-Independent Components](#language-independent-components)
4. [Language-Dependent Components](#language-dependent-components)
5. [Type Mapping Requirements](#type-mapping-requirements)
6. [Code Generation Steps](#code-generation-steps)
7. [Required APIs](#required-apis)
8. [Error Handling](#error-handling)
9. [Testing Strategy](#testing-strategy)
10. [Performance Targets](#performance-targets)

---

## Overview

### What You're Building

A code generator that transforms SDP schemas into native code for your target language, producing:
- **Type definitions** (structs/classes)
- **Encoder functions** (struct → bytes)
- **Decoder functions** (bytes → struct)
- **Error types** (for decode failures)
- **Size calculation helpers** (for pre-allocation)

### Design Goals

1. **Zero-copy where possible** - Minimize allocations
2. **Single allocation encoding** - Calculate size, allocate once, write directly
3. **Type safety** - Compile-time checks for all operations
4. **Predictable performance** - No reflection, no dynamic dispatch
5. **Clear error messages** - Make debugging easy

### Time Estimate

- **C implementation:** 3-4 hours (manual memory management)
- **Swift implementation:** 2-3 hours (similar to Go)
- **Rust implementation:** 3-4 hours (borrow checker considerations)

---

## Wire Format Specification

### Fundamental Rules

**Byte Order:** Little-endian for ALL multi-byte values  
**Alignment:** None - packed binary format  
**Padding:** None - all fields are tightly packed

### Primitive Types Encoding

| SDP Type | Wire Size | Encoding | Notes |
|----------|-----------|----------|-------|
| `u8` | 1 byte | Direct byte value | 0-255 |
| `u16` | 2 bytes | Little-endian uint16 | |
| `u32` | 4 bytes | Little-endian uint32 | |
| `u64` | 8 bytes | Little-endian uint64 | |
| `i8` | 1 byte | Two's complement | -128 to 127 |
| `i16` | 2 bytes | Little-endian, two's complement | |
| `i32` | 4 bytes | Little-endian, two's complement | |
| `i64` | 8 bytes | Little-endian, two's complement | |
| `f32` | 4 bytes | IEEE 754 single precision, little-endian | |
| `f64` | 8 bytes | IEEE 754 double precision, little-endian | |
| `bool` | 1 byte | 0 = false, 1 = true | |
| `str` | Variable | [4-byte length][UTF-8 bytes] | Length is byte count |

### Array Encoding

```
[4-byte count (u32)][element 0][element 1]...[element N-1]
```

- Count is number of elements (not bytes)
- Elements are encoded back-to-back
- Empty array: count = 0, no element data

### Struct Encoding

```
[field 0][field 1]...[field N-1]
```

- Fields encoded in schema definition order
- No field tags or metadata
- Nested structs encoded inline

### Example Wire Format

**Schema:**
```
struct Point {
    x: f32,
    y: f32
}
```

**Encoding of `Point{x: 1.5, y: 2.5}`:**
```
Bytes: [0x00, 0x00, 0xC0, 0x3F, 0x00, 0x00, 0x20, 0x40]
        \_____f32(1.5)_____/  \_____f32(2.5)_____/
```

---

## Language-Independent Components

### Already Implemented (Reusable)

These components are **language-agnostic** and work for ALL target languages:

1. **Lexer** (`internal/parser/lexer.go`)
   - Tokenizes `.sdp` files
   - No changes needed

2. **Parser** (`internal/parser/parser.go`)
   - Builds AST from tokens
   - No changes needed

3. **Validator** (`internal/validator/validator.go`)
   - Checks for cycles, reserved keywords, type errors
   - No changes needed

4. **Wire Format** (`internal/wire/`)
   - Encode/decode utilities for testing
   - No changes needed

### What Gets Reused

```
┌─────────────┐
│  .sdp file  │
└─────┬───────┘
      │
      ▼
┌─────────────┐
│   Lexer     │ ◄── REUSE (language-independent)
└─────┬───────┘
      │
      ▼
┌─────────────┐
│   Parser    │ ◄── REUSE (language-independent)
└─────┬───────┘
      │
      ▼
┌─────────────┐
│  Validator  │ ◄── REUSE (language-independent)
└─────┬───────┘
      │
      ▼         ┌──────────────┐
│     AST     │─┤ Go Generator │ ◄── IMPLEMENT per language
└─────┬───────┘ └──────────────┘
      │         ┌──────────────┐
      ├─────────┤ C Generator  │ ◄── IMPLEMENT per language
      │         └──────────────┘
      │         ┌──────────────┐
      └─────────┤Swift Generator│ ◄── IMPLEMENT per language
                └──────────────┘
```

---

## Language-Dependent Components

### What You Must Implement

For each target language, you need to create a **generator package**:

**Location:** `internal/generator/{language}/`

**Required Files:**

```
internal/generator/{language}/
├── types.go           # Generate type definitions
├── encode_gen.go      # Generate encoder functions
├── decode_gen.go      # Generate decoder functions
├── errors_gen.go      # Generate error types
├── context_gen.go     # Generate decode context (optional)
├── types_test.go      # Unit tests for type generation
├── encode_gen_test.go # Unit tests for encode generation
└── decode_gen_test.go # Unit tests for decode generation
```

### Generator Interface Pattern

Each generator should implement:

```go
type Generator interface {
    // Generate all code files for the schema
    Generate(schema *parser.Schema, packageName string) (map[string]string, error)
}
```

Returns: `map[filename]file_content`

Example:
```go
files := map[string]string{
    "types.h":   "struct Point { float x; float y; };",
    "encode.c":  "uint8_t* encode_point(Point* p) { ... }",
    "decode.c":  "int decode_point(Point* p, uint8_t* data) { ... }",
}
```

---

## Type Mapping Requirements

### Primitive Type Mapping

You must map SDP types to native types in your target language:

| SDP Type | Go | C | Swift | Rust | Notes |
|----------|-----|---|-------|------|-------|
| `u8` | `uint8` | `uint8_t` | `UInt8` | `u8` | Unsigned 8-bit |
| `u16` | `uint16` | `uint16_t` | `UInt16` | `u16` | Unsigned 16-bit |
| `u32` | `uint32` | `uint32_t` | `UInt32` | `u32` | Unsigned 32-bit |
| `u64` | `uint64` | `uint64_t` | `UInt64` | `u64` | Unsigned 64-bit |
| `i8` | `int8` | `int8_t` | `Int8` | `i8` | Signed 8-bit |
| `i16` | `int16` | `int16_t` | `Int16` | `i16` | Signed 16-bit |
| `i32` | `int32` | `int32_t` | `Int32` | `i32` | Signed 32-bit |
| `i64` | `int64` | `int64_t` | `Int64` | `i64` | Signed 64-bit |
| `f32` | `float32` | `float` | `Float` | `f32` | IEEE 754 single |
| `f64` | `float64` | `double` | `Double` | `f64` | IEEE 754 double |
| `bool` | `bool` | `bool` (C99) | `Bool` | `bool` | Boolean |
| `str` | `string` | `char*` | `String` | `String` | UTF-8 encoded |

### Array Type Mapping

| SDP Type | Go | C | Swift | Rust |
|----------|-----|---|-------|------|
| `[]T` | `[]T` | `T* items; uint32_t count;` | `[T]` | `Vec<T>` |

### Struct Type Mapping

SDP structs map to:
- **Go:** `struct`
- **C:** `struct` (with typedef)
- **Swift:** `struct`
- **Rust:** `struct`

---

## Code Generation Steps

### Step 1: Setup Generator Package

Create directory structure:
```bash
mkdir -p internal/generator/{language}
touch internal/generator/{language}/{types,encode_gen,decode_gen,errors_gen}.go
```

### Step 2: Implement Type Generator

**Goal:** Transform AST structs into native type definitions

**Input:** `*parser.Schema` with list of struct definitions  
**Output:** String containing type definitions

**Example for Go:**

```go
// Input AST:
// struct Point { x: f32, y: f32 }

// Output code:
type Point struct {
    X float32
    Y float32
}
```

**Key Considerations:**

1. **Field name transformation:**
   - SDP: `snake_case` → Target language convention
   - Go: `PascalCase` (exported)
   - C: `snake_case`
   - Swift: `camelCase`

2. **Reserved keyword handling:**
   - Check if field name is reserved in target language
   - Append suffix if needed (e.g., `type_` in Go)

3. **Array fields:**
   - Go: Use slices `[]T`
   - C: Separate count + pointer fields
   - Swift: Use arrays `[T]`

### Step 3: Implement Size Calculator Generator

**Goal:** Generate functions that calculate wire format size BEFORE encoding

**Purpose:** Enables single-allocation encoding strategy

**Example for Go:**

```go
// Input: struct Point { x: f32, y: f32 }

// Output:
func calculatePointSize(src *Point) int {
    size := 0
    size += 4 // x: f32
    size += 4 // y: f32
    return size
}
```

**For strings:**
```go
size += 4 + len(src.Name) // 4-byte length prefix + string bytes
```

**For arrays:**
```go
size += 4 // array count
size += len(src.Items) * sizeOfElement
```

**For nested structs:**
```go
size += calculateNestedStructSize(&src.NestedField)
```

### Step 4: Implement Encoder Generator

**Goal:** Generate functions that encode structs to wire format

**Strategy:** Two-pass encoding
1. Calculate total size
2. Allocate buffer
3. Write fields sequentially

**Example for Go:**

```go
// Public API function
func EncodePoint(src *Point) ([]byte, error) {
    size := calculatePointSize(src)
    buf := make([]byte, size)
    offset := 0
    if err := encodePoint(src, buf, &offset); err != nil {
        return nil, err
    }
    return buf, nil
}

// Helper function
func encodePoint(src *Point, buf []byte, offset *int) error {
    // Encode x (f32)
    binary.LittleEndian.PutUint32(buf[*offset:], math.Float32bits(src.X))
    *offset += 4
    
    // Encode y (f32)
    binary.LittleEndian.PutUint32(buf[*offset:], math.Float32bits(src.Y))
    *offset += 4
    
    return nil
}
```

**Encoding Patterns by Type:**

**Primitives (u8-u64, i8-i64):**
```go
// u8, i8
buf[*offset] = src.Field
*offset += 1

// u16, i16
binary.LittleEndian.PutUint16(buf[*offset:], uint16(src.Field))
*offset += 2

// u32, i32
binary.LittleEndian.PutUint32(buf[*offset:], uint32(src.Field))
*offset += 4

// u64, i64
binary.LittleEndian.PutUint64(buf[*offset:], uint64(src.Field))
*offset += 8
```

**Floats:**
```go
// f32
binary.LittleEndian.PutUint32(buf[*offset:], math.Float32bits(src.Field))
*offset += 4

// f64
binary.LittleEndian.PutUint64(buf[*offset:], math.Float64bits(src.Field))
*offset += 8
```

**Boolean:**
```go
if src.Field {
    buf[*offset] = 1
} else {
    buf[*offset] = 0
}
*offset += 1
```

**String:**
```go
binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Field)))
*offset += 4
copy(buf[*offset:], src.Field)
*offset += len(src.Field)
```

**Array:**
```go
binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.ArrayField)))
*offset += 4

for i := range src.ArrayField {
    // Encode element (primitive or nested struct)
    encodeElement(&src.ArrayField[i], buf, offset)
}
```

**Nested Struct:**
```go
if err := encodeNestedStruct(&src.NestedField, buf, offset); err != nil {
    return err
}
```

### Step 5: Implement Decoder Generator

**Goal:** Generate functions that decode wire format to structs

**Strategy:** Sequential reads with bounds checking

**Example for Go:**

```go
// Public API function
func DecodePoint(dst *Point, data []byte) error {
    ctx := &DecodeContext{}
    offset := 0
    return decodePoint(dst, data, &offset, ctx)
}

// Helper function
func decodePoint(dst *Point, data []byte, offset *int, ctx *DecodeContext) error {
    // Decode x (f32)
    if *offset + 4 > len(data) {
        return ErrUnexpectedEOF
    }
    bits := binary.LittleEndian.Uint32(data[*offset:])
    dst.X = math.Float32frombits(bits)
    *offset += 4
    
    // Decode y (f32)
    if *offset + 4 > len(data) {
        return ErrUnexpectedEOF
    }
    bits = binary.LittleEndian.Uint32(data[*offset:])
    dst.Y = math.Float32frombits(bits)
    *offset += 4
    
    return nil
}
```

**Decoding Patterns by Type:**

**Bounds Checking Template:**
```go
if *offset + SIZE > len(data) {
    return ErrUnexpectedEOF
}
```

**Primitives:**
```go
// u8, i8
dst.Field = data[*offset]
*offset += 1

// u16
dst.Field = binary.LittleEndian.Uint16(data[*offset:])
*offset += 2

// i16
dst.Field = int16(binary.LittleEndian.Uint16(data[*offset:]))
*offset += 2

// u32
dst.Field = binary.LittleEndian.Uint32(data[*offset:])
*offset += 4

// i32
dst.Field = int32(binary.LittleEndian.Uint32(data[*offset:]))
*offset += 4

// u64
dst.Field = binary.LittleEndian.Uint64(data[*offset:])
*offset += 8

// i64
dst.Field = int64(binary.LittleEndian.Uint64(data[*offset:]))
*offset += 8
```

**Floats:**
```go
// f32
bits := binary.LittleEndian.Uint32(data[*offset:])
dst.Field = math.Float32frombits(bits)
*offset += 4

// f64
bits := binary.LittleEndian.Uint64(data[*offset:])
dst.Field = math.Float64frombits(bits)
*offset += 8
```

**Boolean:**
```go
dst.Field = data[*offset] != 0
*offset += 1
```

**String:**
```go
// Read length
if *offset + 4 > len(data) {
    return ErrUnexpectedEOF
}
strLen := binary.LittleEndian.Uint32(data[*offset:])
*offset += 4

// Validate length
if *offset + int(strLen) > len(data) {
    return ErrUnexpectedEOF
}

// Read string bytes
dst.Field = string(data[*offset : *offset + int(strLen)])
*offset += int(strLen)
```

**Array:**
```go
// Read count
if *offset + 4 > len(data) {
    return ErrUnexpectedEOF
}
arrCount := binary.LittleEndian.Uint32(data[*offset:])
*offset += 4

// Check array size limits
if arrCount > MaxArrayElements {
    return ErrArrayTooLarge
}

ctx.TotalElementCount += uint64(arrCount)
if ctx.TotalElementCount > MaxTotalElements {
    return ErrTooManyElements
}

// Allocate array
dst.ArrayField = make([]ElementType, arrCount)

// Decode elements
for i := uint32(0); i < arrCount; i++ {
    if err := decodeElement(&dst.ArrayField[i], data, offset, ctx); err != nil {
        return err
    }
}
```

**Nested Struct:**
```go
if err := decodeNestedStruct(&dst.NestedField, data, offset, ctx); err != nil {
    return err
}
```

### Step 6: Implement Error Types Generator

**Goal:** Generate error constants/types for decode failures

**Required Errors:**

1. **ErrUnexpectedEOF**
   - When: Not enough bytes to decode field
   - Message: "unexpected end of data"

2. **ErrArrayTooLarge**
   - When: Array count exceeds MaxArrayElements
   - Message: "array size exceeds maximum allowed elements"

3. **ErrTooManyElements**
   - When: Total cumulative array elements exceed MaxTotalElements
   - Message: "total array elements across all arrays exceeds limit"

**Constants:**
```go
const (
    MaxArrayElements  = 1_000_000   // Per-array limit
    MaxTotalElements  = 10_000_000  // Cumulative limit
    MaxSerializedSize = 128 * 1024 * 1024 // 128 MB
)
```

**Example for Go:**
```go
package mypackage

import "errors"

var (
    ErrUnexpectedEOF    = errors.New("unexpected end of data")
    ErrArrayTooLarge    = errors.New("array size exceeds maximum allowed elements")
    ErrTooManyElements  = errors.New("total array elements across all arrays exceeds limit")
)

const (
    MaxArrayElements  = 1_000_000
    MaxTotalElements  = 10_000_000
    MaxSerializedSize = 128 * 1024 * 1024
)
```

### Step 7: Implement Decode Context Generator (Optional)

**Goal:** Track cumulative state during decoding to enforce security limits

**Purpose:**
- Prevent DoS attacks via malicious array counts
- Track total elements across nested arrays

**Example for Go:**
```go
type DecodeContext struct {
    TotalElementCount uint64
}
```

**Usage in decoder:**
```go
ctx.TotalElementCount += uint64(arrayCount)
if ctx.TotalElementCount > MaxTotalElements {
    return ErrTooManyElements
}
```

---

## Required APIs

### Public Encoder API

**Pattern:** `Encode{StructName}(src *{StructName}) ([]byte, error)`

**Responsibilities:**
1. Calculate wire format size
2. Allocate buffer
3. Call helper encoder
4. Return bytes or error

**Example signatures:**

**Go:**
```go
func EncodePoint(src *Point) ([]byte, error)
```

**C:**
```c
uint8_t* encode_point(Point* src, uint32_t* out_size);
// Returns malloc'd buffer, sets out_size
// Caller must free() the returned buffer
// Returns NULL on error
```

**Swift:**
```swift
func encodePoint(_ src: Point) throws -> Data
```

**Rust:**
```rust
pub fn encode_point(src: &Point) -> Result<Vec<u8>, EncodeError>
```

### Public Decoder API

**Pattern:** `Decode{StructName}(dst *{StructName}, data []byte) error`

**Responsibilities:**
1. Initialize decode context
2. Call helper decoder
3. Return error or success

**Example signatures:**

**Go:**
```go
func DecodePoint(dst *Point, data []byte) error
```

**C:**
```c
int decode_point(Point* dst, uint8_t* data, uint32_t data_len);
// Returns 0 on success, error code on failure
```

**Swift:**
```swift
func decodePoint(_ dst: inout Point, from data: Data) throws
```

**Rust:**
```rust
pub fn decode_point(dst: &mut Point, data: &[u8]) -> Result<(), DecodeError>
```

### Helper Functions

**Private/Internal helpers:**
- `calculate{StructName}Size(src *{StructName}) int`
- `encode{StructName}(src *{StructName}, buf []byte, offset *int) error`
- `decode{StructName}(dst *{StructName}, data []byte, offset *int, ctx *DecodeContext) error`

---

## Error Handling

### Error Philosophy

- **Decode errors are expected** - Invalid data, truncated messages, malicious input
- **Encode errors are rare** - Only if struct is invalid (shouldn't happen with type safety)
- **Clear error messages** - Make debugging easy

### Error Handling by Language

**Go:**
```go
// Return error values
if err := DecodePoint(&p, data); err != nil {
    if err == ErrUnexpectedEOF {
        // Handle truncated data
    }
}
```

**C:**
```c
// Return error codes
int result = decode_point(&p, data, len);
if (result != 0) {
    // Handle error
    switch (result) {
        case ERR_UNEXPECTED_EOF:
            // Handle truncated data
            break;
    }
}
```

**Swift:**
```swift
// Throw errors
do {
    try decodePoint(&p, from: data)
} catch DecodeError.unexpectedEOF {
    // Handle truncated data
}
```

**Rust:**
```rust
// Return Result
match decode_point(&mut p, &data) {
    Ok(_) => { /* Success */ },
    Err(DecodeError::UnexpectedEOF) => { /* Handle truncated data */ },
}
```

---

## Testing Strategy

### Test Pyramid

```
           ┌──────────────────┐
           │  Integration     │  ← Real-world data (plugins.json)
           │  Tests (3)       │
           └──────────────────┘
          ┌────────────────────┐
          │  Wire Format       │  ← Hand-crafted binary
          │  Tests (11)        │
          └────────────────────┘
        ┌──────────────────────┐
        │  Roundtrip           │  ← Encode → Decode
        │  Tests (8)           │
        └──────────────────────┘
      ┌────────────────────────┐
      │  Unit Tests            │  ← Code generation
      │  (238)                 │
      └────────────────────────┘
```

### Required Test Types

#### 1. Unit Tests (Per Generator)

Test that generated code is **syntactically correct**:

**Test files:**
- `types_test.go` - Test type generation
- `encode_gen_test.go` - Test encoder generation
- `decode_gen_test.go` - Test decoder generation

**Example:**
```go
func TestGenerateStructWithPrimitives(t *testing.T) {
    schema := parseSchema("struct Point { x: f32, y: f32 }")
    code := GenerateTypes(schema)
    
    // Verify code contains expected patterns
    if !strings.Contains(code, "type Point struct") {
        t.Error("Missing struct definition")
    }
    if !strings.Contains(code, "X float32") {
        t.Error("Missing field X")
    }
}
```

#### 2. Compilation Tests

Test that generated code **compiles without errors**:

**Setup:**
```go
func TestMain(m *testing.M) {
    generateCode()  // Generate from test schemas
    compileCode()   // Try to compile generated code
    code := m.Run()
    cleanup()
    os.Exit(code)
}
```

#### 3. Wire Format Tests

Test that generated code produces **spec-compliant binary**:

**Test with hand-crafted binary data:**
```go
func TestWireFormatPoint(t *testing.T) {
    // Hand-craft binary: x=1.5, y=2.5
    data := []byte{
        0x00, 0x00, 0xC0, 0x3F,  // f32(1.5) little-endian
        0x00, 0x00, 0x20, 0x40,  // f32(2.5) little-endian
    }
    
    var p Point
    err := DecodePoint(&p, data)
    if err != nil {
        t.Fatal(err)
    }
    
    if p.X != 1.5 || p.Y != 2.5 {
        t.Errorf("Got (%f, %f), want (1.5, 2.5)", p.X, p.Y)
    }
}
```

#### 4. Roundtrip Tests

Test that **encode → decode produces identical data**:

```go
func TestRoundtripPoint(t *testing.T) {
    original := Point{X: 1.5, Y: 2.5}
    
    encoded, err := EncodePoint(&original)
    if err != nil {
        t.Fatal(err)
    }
    
    var decoded Point
    err = DecodePoint(&decoded, encoded)
    if err != nil {
        t.Fatal(err)
    }
    
    if decoded.X != original.X || decoded.Y != original.Y {
        t.Errorf("Roundtrip failed: got %+v, want %+v", decoded, original)
    }
}
```

**Test cases:**
- All primitive types (max/min values)
- Empty strings/arrays
- Nested structs (3+ levels deep)
- Large arrays (1000+ elements)
- Unicode strings
- Mixed types

#### 5. Error Handling Tests

Test that decoders **reject invalid input**:

```go
func TestDecodeTruncatedData(t *testing.T) {
    data := []byte{0x00}  // Only 1 byte, Point needs 8
    
    var p Point
    err := DecodePoint(&p, data)
    
    if err != ErrUnexpectedEOF {
        t.Errorf("Expected ErrUnexpectedEOF, got %v", err)
    }
}

func TestDecodeOversizedArray(t *testing.T) {
    // Array count: 10,000,000 (exceeds MaxArrayElements)
    data := []byte{0x80, 0x96, 0x98, 0x00}
    
    var arr ArrayStruct
    err := DecodeArrayStruct(&arr, data)
    
    if err != ErrArrayTooLarge {
        t.Errorf("Expected ErrArrayTooLarge, got %v", err)
    }
}
```

#### 6. Performance Benchmarks

Test that implementation meets **performance targets**:

```go
func BenchmarkEncodePoint(b *testing.B) {
    p := Point{X: 1.5, Y: 2.5}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := EncodePoint(&p)
        if err != nil {
            b.Fatal(err)
        }
    }
}

func BenchmarkDecodePoint(b *testing.B) {
    p := Point{X: 1.5, Y: 2.5}
    data, _ := EncodePoint(&p)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var result Point
        err := DecodePoint(&result, data)
        if err != nil {
            b.Fatal(err)
        }
    }
}
```

### Cross-Language Compatibility Tests

**Test that different language implementations can interoperate:**

```go
func TestGoCanDecodeSwiftEncoded(t *testing.T) {
    // Load data encoded by Swift implementation
    swiftEncoded := loadFile("testdata/point_swift.bin")
    
    var p Point
    err := DecodePoint(&p, swiftEncoded)
    if err != nil {
        t.Fatal(err)
    }
    
    // Verify expected values
    if p.X != 1.5 || p.Y != 2.5 {
        t.Errorf("Cross-language decode failed")
    }
}
```

---

## Performance Targets

### Baseline Targets (M1 Base CPU)

Based on Go reference implementation benchmarks:

| Operation | Target | Notes |
|-----------|--------|-------|
| Simple struct encode | < 30 ns | 2-3 primitive fields |
| Simple struct decode | < 25 ns | 2-3 primitive fields |
| Nested struct (3 levels) | < 25 ns encode, < 20 ns decode | Point → Rectangle → Scene |
| Array encode (small) | < 60 ns | 4-5 elements |
| Array decode (small) | < 140 ns | 4-5 elements (allocations) |
| Complex struct | < 80 ns encode, < 190 ns decode | 2 plugins, 3 params each |
| Large array (1000 elem) | < 1100 ns encode, < 1300 ns decode | Linear scaling |

### Real-World Target

**62 plugins × 1,759 parameters:**
- **Roundtrip:** < 150 µs (SDP Go: 128.5 µs)
- **Encode only:** < 40 µs (SDP Go: 37.4 µs)
- **Decode only:** < 90 µs (SDP Go: 85.1 µs)

### Comparison Baseline

Must beat:
- **Protocol Buffers:** 1,300 µs roundtrip
- **FlatBuffers:** 1,000 µs roundtrip

**Target:** At least **8-10× faster** than competitors

### Language-Specific Considerations

**C:**
- Should match or beat Go performance (no GC overhead)
- Memory management is manual (malloc/free)
- Target: 10-15% faster than Go

**Swift:**
- May be slightly slower due to ARC overhead
- Copy-on-write for arrays/strings
- Target: Within 20% of Go performance

**Rust:**
- Should match or beat Go (zero-cost abstractions)
- Borrowing instead of copying
- Target: 5-10% faster than Go

---

## Implementation Checklist

Use this checklist when implementing a new language generator:

### Phase 1: Setup
- [ ] Create `internal/generator/{language}/` directory
- [ ] Create `types.go`, `encode_gen.go`, `decode_gen.go`, `errors_gen.go`
- [ ] Set up test files: `*_test.go`
- [ ] Document type mappings for your language

### Phase 2: Type Generator
- [ ] Implement struct generation
- [ ] Handle field name transformation (snake_case → language convention)
- [ ] Handle reserved keywords
- [ ] Handle array types properly
- [ ] Add unit tests for type generation
- [ ] Verify generated code compiles

### Phase 3: Encoder Generator
- [ ] Implement size calculator generation
- [ ] Implement encoder for primitives (u8-u64, i8-i64, f32, f64, bool)
- [ ] Implement string encoder (length prefix + UTF-8)
- [ ] Implement array encoder (count + elements)
- [ ] Implement nested struct encoder
- [ ] Add unit tests for encoder generation
- [ ] Verify generated code compiles

### Phase 4: Decoder Generator
- [ ] Implement decoder for primitives with bounds checking
- [ ] Implement string decoder with validation
- [ ] Implement array decoder with size limits
- [ ] Implement nested struct decoder
- [ ] Implement DecodeContext for security
- [ ] Add unit tests for decoder generation
- [ ] Verify generated code compiles

### Phase 5: Error Types
- [ ] Generate ErrUnexpectedEOF
- [ ] Generate ErrArrayTooLarge
- [ ] Generate ErrTooManyElements
- [ ] Define MaxArrayElements, MaxTotalElements, MaxSerializedSize
- [ ] Add unit tests for error generation

### Phase 6: Integration Testing
- [ ] Create test schemas (primitives, nested, arrays, complex)
- [ ] Set up TestMain for automatic code generation
- [ ] Add wire format tests (hand-crafted binary data)
- [ ] Add roundtrip tests (encode → decode)
- [ ] Add error handling tests (truncated data, oversized arrays)
- [ ] Add cross-language compatibility tests
- [ ] Verify all tests pass

### Phase 7: Performance Testing
- [ ] Add benchmarks for simple structs
- [ ] Add benchmarks for nested structs
- [ ] Add benchmarks for arrays
- [ ] Add benchmarks for complex structs
- [ ] Add real-world benchmark (plugins.json equivalent)
- [ ] Verify performance meets targets
- [ ] Compare against Protocol Buffers/FlatBuffers

### Phase 8: Documentation
- [ ] Document public API
- [ ] Add usage examples
- [ ] Document error handling patterns
- [ ] Add language-specific best practices
- [ ] Update main README with language support

---

## Language-Specific Guides

### C Implementation Notes

**Memory Management:**
- Encoder: Return malloc'd buffer, caller must free
- Decoder: Caller provides pre-allocated struct
- Strings: malloc'd, must be freed separately
- Arrays: malloc'd, must be freed separately

**Error Handling:**
- Return 0 for success, error code for failure
- Define error codes as enum or #define

**Header Structure:**
```c
// types.h
typedef struct {
    float x;
    float y;
} Point;

// encode.h
uint8_t* encode_point(Point* src, uint32_t* out_size);

// decode.h
int decode_point(Point* dst, uint8_t* data, uint32_t data_len);

// errors.h
#define ERR_UNEXPECTED_EOF 1
#define ERR_ARRAY_TOO_LARGE 2
#define ERR_TOO_MANY_ELEMENTS 3
```

**Endianness Handling:**
```c
#include <stdint.h>

// Write little-endian uint32
static inline void write_u32_le(uint8_t* buf, uint32_t val) {
    buf[0] = (val >> 0) & 0xFF;
    buf[1] = (val >> 8) & 0xFF;
    buf[2] = (val >> 16) & 0xFF;
    buf[3] = (val >> 24) & 0xFF;
}

// Read little-endian uint32
static inline uint32_t read_u32_le(const uint8_t* buf) {
    return (uint32_t)buf[0] 
         | ((uint32_t)buf[1] << 8)
         | ((uint32_t)buf[2] << 16)
         | ((uint32_t)buf[3] << 24);
}
```

**Float Encoding:**
```c
#include <string.h>

// Encode f32
uint32_t f32_bits;
memcpy(&f32_bits, &src->field, 4);
write_u32_le(buf + offset, f32_bits);

// Decode f32
uint32_t f32_bits = read_u32_le(data + offset);
memcpy(&dst->field, &f32_bits, 4);
```

### Swift Implementation Notes

**Memory Management:**
- Use `Data` for byte buffers (automatic memory management)
- Strings are value types (automatic)
- Arrays are value types (copy-on-write)

**Error Handling:**
```swift
enum DecodeError: Error {
    case unexpectedEOF
    case arrayTooLarge
    case tooManyElements
}
```

**Endianness Handling:**
```swift
extension Data {
    mutating func appendLittleEndian(_ value: UInt32) {
        var v = value.littleEndian
        withUnsafeBytes(of: &v) { self.append(contentsOf: $0) }
    }
    
    func readUInt32LittleEndian(at offset: Int) -> UInt32 {
        let value = self.withUnsafeBytes { $0.load(fromByteOffset: offset, as: UInt32.self) }
        return UInt32(littleEndian: value)
    }
}
```

**Float Encoding:**
```swift
// Encode f32
let bits = value.bitPattern
data.appendLittleEndian(bits)

// Decode f32
let bits = data.readUInt32LittleEndian(at: offset)
value = Float(bitPattern: bits)
```

### Rust Implementation Notes

**Memory Management:**
- Use `Vec<u8>` for buffers
- Strings are `String` or `&str`
- Arrays are `Vec<T>`

**Error Handling:**
```rust
#[derive(Debug)]
pub enum DecodeError {
    UnexpectedEOF,
    ArrayTooLarge,
    TooManyElements,
}

impl std::fmt::Display for DecodeError {
    fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
        match self {
            DecodeError::UnexpectedEOF => write!(f, "unexpected end of data"),
            DecodeError::ArrayTooLarge => write!(f, "array size exceeds maximum"),
            DecodeError::TooManyElements => write!(f, "too many total elements"),
        }
    }
}

impl std::error::Error for DecodeError {}
```

**Endianness Handling:**
```rust
use std::io::Cursor;
use byteorder::{LittleEndian, ReadBytesExt, WriteBytesExt};

// Encode u32
buf.write_u32::<LittleEndian>(value)?;

// Decode u32
let value = cursor.read_u32::<LittleEndian>()?;
```

**Float Encoding:**
```rust
// Encode f32
buf.write_f32::<LittleEndian>(value)?;

// Decode f32
let value = cursor.read_f32::<LittleEndian>()?;
```

---

## Summary

### What's Language-Independent (Reuse)
✅ Lexer  
✅ Parser  
✅ Validator  
✅ Wire format specification  
✅ Test schemas  
✅ Integration test data (plugins.json)

### What's Language-Dependent (Implement)
🔨 Type generator  
🔨 Encoder generator  
🔨 Decoder generator  
🔨 Error types generator  
🔨 Language-specific tests  
🔨 Benchmarks

### Expected Timeline
- **C:** 3-4 hours (manual memory management)
- **Swift:** 2-3 hours (similar to Go)
- **Both in one day:** Achievable with this guide! 🚀

### Success Criteria
✅ All tests pass  
✅ Performance beats Protocol Buffers by 8-10×  
✅ Cross-language compatibility verified  
✅ Real-world data (plugins.json) works

---

**Next Steps:**
1. Review this guide
2. Choose target language (C or Swift)
3. Follow Phase 1-8 checklist
4. Celebrate when benchmarks beat Protocol Buffers! 🎉

