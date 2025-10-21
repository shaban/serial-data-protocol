# C API Specification for Serial Data Protocol

**Version:** 0.2.0  
**Date:** October 20, 2025  
**Status:** Optimized Design

## Overview

The C API for SDP provides a **high-performance, zero-allocation** interface for encoding and decoding structured binary data. Based on extensive benchmarking (see `c_micro_bench/OPTIMIZATION_RESULTS.md`), the optimized C encoder is **1.7x to 46x faster than Go** on equivalent schemas.

**Key Features:**
1. **Zero allocations** - User provides buffer, no dynamic memory
2. **Wire format structs** - 10-30x speedup via bulk memcpy
3. **Inline encoding** - Eliminates function call overhead (3-5x speedup)
4. **Pre-computed lengths** - Caller provides string/array lengths (9x faster than strlen)
5. **FFI-friendly** - Works seamlessly with Swift, Objective-C, C++, Rust
6. **Portable** - Standard C99, works on all platforms

**Performance vs Go (M1, -O3):**
- Primitives: 0.84 ns vs 25.77 ns → **30.7x faster**
- Arrays (small): 29.96 ns vs 56.02 ns → **1.9x faster**
- Nested structs: 0.48 ns vs 22.11 ns → **46x faster**
- Complex (3-level): 24.23 ns vs 75.93 ns → **3.1x faster**

## Core Principles

### 1. Buffer-Based API (No Dynamic Allocation)

**User provides the buffer:**
```c
// Calculate size (cheap - uses compile-time constants)
size_t size = sdp_parameter_size(&param);

// Allocate (stack, heap, or static - user choice)
uint8_t buf[512];  // or malloc(size)

// Encode (no hidden allocations)
size_t written = sdp_parameter_encode(&param, buf);
```

**Benefits:**
- Predictable performance (no malloc/free overhead)
- Embedded-friendly (static buffers, no heap)
- Cache-friendly (user controls alignment/placement)

### 2. Struct-Based Data Model

**Separation of data and encoding:**
```c
// User constructs data in typed structs
SDPParameter param = {
    .address = 0x1000,
    .display_name = "Gain",
    .display_name_len = 4,  // Caller MUST provide lengths
    .min_value = -96.0f,
    // ...
};

// Generator provides optimized encoding
sdp_parameter_encode(&param, buf);
```

**Benefits:**
- Easy to construct/test (just struct initialization)
- Matches real-world usage patterns
- Clear ownership (user owns data, encoder writes bytes)

### 3. Pre-Computed String Lengths

**ALL string parameters require length:**
```c
void sdp_parameter_encode(
    const SDPParameter* p,
    uint8_t* buf
);

// Struct requires pre-computed lengths
typedef struct {
    const char* name;
    size_t name_len;  // Caller MUST provide
} SDPParameter;
```

**Why:** `strlen()` is 9x slower than using pre-computed lengths (55x slower than snprintf!). Generated code NEVER calls `strlen()`.

### 4. Wire Format Optimization

**Generated wire format structs for fixed portions:**
```c
// Generated (internal use, not exposed)
typedef struct __attribute__((packed)) {
    uint64_t address;
    float min_value;
    float max_value;
    float default_value;
    float current_value;
    uint32_t raw_flags;
    uint8_t is_writable;
    uint8_t can_ramp;
} SDPParameterWire;  // 34 bytes fixed

// Encoding uses bulk memcpy
*(uint64_t*)(buf + 0) = param->address;
// ... strings ...
memcpy(buf + offset, &wire_suffix, sizeof(wire_suffix));
```

**Speedup:** 10-30x faster than field-by-field encoding

### 5. Inline Nested Encoding

**No function calls for nested structs/arrays:**
```c
// Generated code inlines everything
for (size_t i = 0; i < plugin->param_count; i++) {
    // Inline all parameter encoding HERE (no function call)
    *(uint64_t*)(buf + offset) = params[i].address;
    offset += 8;
    // ... all fields inlined ...
}
```

**Speedup:** 1.4-5x faster (eliminates function call overhead)

### 6. Endianness

All wire format data is **little-endian**. API uses direct writes on little-endian platforms (M1/x64), byte swapping on big-endian.

## API Design

### Generated Types

For each schema struct, the generator produces:

```c
// Example: Parameter from audiounit.sdp

// 1. Public struct (user-facing)
typedef struct {
    uint64_t address;
    const char* display_name;
    size_t display_name_len;  // Pre-computed
    const char* identifier;
    size_t identifier_len;    // Pre-computed
    const char* unit;
    size_t unit_len;          // Pre-computed
    float min_value;
    float max_value;
    float default_value;
    float current_value;
    uint32_t raw_flags;
    bool is_writable;
    bool can_ramp;
} SDPParameter;

// 2. Size calculation (uses compile-time constants)
size_t sdp_parameter_size(const SDPParameter* p);

// 3. Encoding (optimized, no allocations)
size_t sdp_parameter_encode(const SDPParameter* p, uint8_t* buf);

// 4. Size macro (compile-time constant for fixed portion)
#define SDP_PARAMETER_FIXED_SIZE 34
```

### Encoding API

#### Struct Encoding Pattern

```c
// 1. Construct data
SDPParameter param = {
    .address = 0x1000,
    .display_name = "Input Gain",
    .display_name_len = 10,  // Caller provides
    .identifier = "gain",
    .identifier_len = 4,
    .unit = "dB",
    .unit_len = 2,
    .min_value = -96.0f,
    .max_value = 12.0f,
    .default_value = 0.0f,
    .current_value = -6.0f,
    .raw_flags = 0x00000001,
    .is_writable = true,
    .can_ramp = true
};

// 2. Calculate size
size_t size = sdp_parameter_size(&param);

// 3. Allocate buffer (stack, heap, or static)
uint8_t buf[512];  // Stack
// or: uint8_t* buf = malloc(size);  // Heap
// or: static uint8_t buf[1024];     // Static

// 4. Encode
size_t written = sdp_parameter_encode(&param, buf);

// 5. Use buffer
send_to_network(buf, written);
```

#### Array Encoding Pattern

**Primitive arrays (bulk memcpy):**
```c
// For arrays of u8, u16, u32, u64, i8, i16, i32, i64, f32, f64, bool
typedef struct {
    const uint32_t* values;
    size_t count;
} SDPu32Array;

size_t sdp_u32_array_encode(const SDPu32Array* arr, uint8_t* buf) {
    *(uint32_t*)buf = arr->count;
    memcpy(buf + 4, arr->values, arr->count * sizeof(uint32_t));
    return 4 + arr->count * sizeof(uint32_t);
}
```

**String arrays (loop required):**
```c
typedef struct {
    const char** strings;
    const size_t* lengths;  // Pre-computed
    size_t count;
} SDPStringArray;

size_t sdp_string_array_encode(const SDPStringArray* arr, uint8_t* buf) {
    size_t offset = 0;
    *(uint32_t*)(buf + offset) = arr->count;
    offset += 4;
    
    for (size_t i = 0; i < arr->count; i++) {
        *(uint32_t*)(buf + offset) = arr->lengths[i];
        offset += 4;
        memcpy(buf + offset, arr->strings[i], arr->lengths[i]);
        offset += arr->lengths[i];
    }
    
    return offset;
}
```

**Struct arrays (inline encoding):**
```c
typedef struct {
    SDPParameter* items;
    size_t count;
} SDPParameterArray;

size_t sdp_parameter_array_encode(const SDPParameterArray* arr, uint8_t* buf) {
    size_t offset = 0;
    *(uint32_t*)(buf + offset) = arr->count;
    offset += 4;
    
    // Inline all parameter encoding (no function calls)
    for (size_t i = 0; i < arr->count; i++) {
        const SDPParameter* p = &arr->items[i];
        
        // ALL fields inlined here
        *(uint64_t*)(buf + offset) = p->address;
        offset += 8;
        
        *(uint32_t*)(buf + offset) = p->display_name_len;
        offset += 4;
        memcpy(buf + offset, p->display_name, p->display_name_len);
        offset += p->display_name_len;
        
        // ... all other fields inlined ...
    }
    
    return offset;
}
```

#### Error Handling

**No error returns** - all errors are programming errors (buffer too small, NULL pointers):
```c
// Returns bytes written (for validation)
size_t sdp_parameter_encode(const SDPParameter* p, uint8_t* buf);

// Caller validates:
assert(written == size);
```

**Alternative**: Add bounds checking in debug builds:
```c
#ifdef SDP_DEBUG
    if (buf == NULL || p == NULL) {
        abort();  // Programming error
    }
#endif
```

**Option A: Void functions, check at end**
```c
void sdp_writer_u32(SDPWriter* w, uint32_t value);
## Code Generation Strategy

### Optimization Decision Tree

The generator applies optimizations based on schema characteristics:

#### 1. Wire Format Structs (10-30x speedup)

**When to generate:**
- Struct has only fixed-size fields (no strings/arrays)
- OR struct has fixed prefix/suffix (strings/arrays in middle)

**Generated code:**
```c
// For struct with only primitives
typedef struct __attribute__((packed)) {
    uint64_t field1;
    uint32_t field2;
    float field3;
    uint8_t field4;
} MyStructWire;

size_t sdp_my_struct_encode(const SDPMyStruct* src, uint8_t* buf) {
    MyStructWire wire = {
        .field1 = src->field1,
        .field2 = src->field2,
        .field3 = src->field3,
        .field4 = src->field4
    };
    memcpy(buf, &wire, sizeof(wire));
    return sizeof(wire);
}
```

**When to skip:**
- Struct is mostly strings/arrays (wire struct wouldn't help much)

#### 2. Bulk Array Copy (2-5x speedup for large arrays)

**When to generate:**
- Array of primitive types: u8, u16, u32, u64, i8, i16, i32, i64, f32, f64, bool

**Generated code:**
```c
size_t sdp_u32_array_encode(const uint32_t* arr, size_t count, uint8_t* buf) {
    *(uint32_t*)buf = count;
    memcpy(buf + 4, arr, count * sizeof(uint32_t));
    return 4 + count * sizeof(uint32_t);
}
```

**When to use loop instead:**
- Array of strings (variable-length, can't bulk copy)
- Array of structs with strings (inline encoding instead)
- Array size typically < 10 elements (loop overhead lower)

#### 3. Inline Nested Encoding (1.4-5x speedup)

**When to inline:**
- Nested struct called inside array loop
- Nested struct has < 20 fields (avoid code bloat)
- Hot path (called frequently)

**Generated code:**
```c
// Instead of calling sdp_parameter_encode()
for (size_t i = 0; i < plugin->param_count; i++) {
    const SDPParameter* p = &plugin->params[i];
    
    // INLINE all parameter encoding
    *(uint64_t*)(buf + offset) = p->address;
    offset += 8;
    
    *(uint32_t*)(buf + offset) = p->display_name_len;
    offset += 4;
    memcpy(buf + offset, p->display_name, p->display_name_len);
    offset += p->display_name_len;
    
    // ... all fields inlined
}
```

**When to keep function call:**
- Top-level encoding (not in a loop)
- Struct is huge (> 20 fields, avoid code bloat)
- Rarely called (not worth inlining)

#### 4. Pre-Computed Size Macros

**Always generate for structs:**
```c
// Fixed portion size (all non-variable fields)
#define SDP_PARAMETER_FIXED_SIZE 34

// Use in size calculation
size_t sdp_parameter_size(const SDPParameter* p) {
    return SDP_PARAMETER_FIXED_SIZE
           + 4 + p->display_name_len
           + 4 + p->identifier_len
           + 4 + p->unit_len;
}
```

### Generated File Structure

For each `.sdp` schema, generate:

```
output_dir/
├── types.h              // Struct definitions and constants
├── encode.h             // Encoding API declarations
├── encode.c             // Encoding implementation
├── decode.h             // Decoding API (future)
└── decode.c             // Decoding implementation (future)
```

**Example for `primitives.sdp`:**
```
testdata/primitives/c/
├── types.h              // SDPAllPrimitives struct
├── encode.h             // sdp_all_primitives_encode(), sdp_all_primitives_size()
├── encode.c             // Implementation with wire struct optimization
└── Makefile
```

### Generated Code Example

**Input schema (primitives.sdp):**
```
struct AllPrimitives {
    u8_field: u8
    u16_field: u16
    u32_field: u32
    u64_field: u64
    i8_field: i8
    i16_field: i16
    i32_field: i32
    i64_field: i64
    f32_field: f32
    f64_field: f64
    bool_field: bool
    str_field: str
}
```

**Generated types.h:**
```c
#ifndef SDP_PRIMITIVES_TYPES_H
#define SDP_PRIMITIVES_TYPES_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

typedef struct {
    uint8_t u8_field;
    uint16_t u16_field;
    uint32_t u32_field;
    uint64_t u64_field;
    int8_t i8_field;
    int16_t i16_field;
    int32_t i32_field;
    int64_t i64_field;
    float f32_field;
    double f64_field;
    bool bool_field;
    const char* str_field;
    size_t str_field_len;  // Pre-computed
} SDPAllPrimitives;

// Fixed portion size (11 primitives = 43 bytes)
#define SDP_ALL_PRIMITIVES_FIXED_SIZE 43

#endif
```

**Generated encode.h:**
```c
#ifndef SDP_PRIMITIVES_ENCODE_H
#define SDP_PRIMITIVES_ENCODE_H

#include "types.h"

// Calculate wire format size
size_t sdp_all_primitives_size(const SDPAllPrimitives* p);

// Encode to buffer (returns bytes written)
size_t sdp_all_primitives_encode(const SDPAllPrimitives* p, uint8_t* buf);

#endif
```

**Generated encode.c (optimized with wire struct):**
```c
#include "encode.h"
#include <string.h>

// Internal wire format struct (packed, fixed portion only)
typedef struct __attribute__((packed)) {
    uint8_t u8_field;
    uint16_t u16_field;
    uint32_t u32_field;
    uint64_t u64_field;
    int8_t i8_field;
    int16_t i16_field;
    int32_t i32_field;
    int64_t i64_field;
    float f32_field;
    double f64_field;
    uint8_t bool_field;
} SDPAllPrimitivesWire;

size_t sdp_all_primitives_size(const SDPAllPrimitives* p) {
    return SDP_ALL_PRIMITIVES_FIXED_SIZE + 4 + p->str_field_len;
}

size_t sdp_all_primitives_encode(const SDPAllPrimitives* p, uint8_t* buf) {
    size_t offset = 0;
    
    // Fixed portion: bulk copy via wire struct
    SDPAllPrimitivesWire wire = {
        .u8_field = p->u8_field,
        .u16_field = p->u16_field,
        .u32_field = p->u32_field,
        .u64_field = p->u64_field,
        .i8_field = p->i8_field,
        .i16_field = p->i16_field,
        .i32_field = p->i32_field,
        .i64_field = p->i64_field,
        .f32_field = p->f32_field,
        .f64_field = p->f64_field,
        .bool_field = p->bool_field ? 1 : 0
    };
    
    memcpy(buf + offset, &wire, sizeof(wire));
    offset += sizeof(wire);
    
    // String: variable-length
    *(uint32_t*)(buf + offset) = (uint32_t)p->str_field_len;
    offset += 4;
    memcpy(buf + offset, p->str_field, p->str_field_len);
    offset += p->str_field_len;
    
    return offset;
}
```

**Performance:** 0.84 ns (30.7x faster than Go's 25.77 ns)

### Complex Example (Nested with Arrays)

**Input schema (audiounit.sdp):**
```
struct Parameter {
    address: u64
    display_name: str
    // ... more fields
}

struct Plugin {
    name: str
    parameters: []Parameter
}

struct PluginRegistry {
    plugins: []Plugin
}
```

**Generated encode.c (inline optimization):**
```c
size_t sdp_plugin_registry_encode(const SDPPluginRegistry* reg, uint8_t* buf) {
    size_t offset = 0;
    
    // Plugins array count
    *(uint32_t*)(buf + offset) = reg->plugin_count;
    offset += 4;
    
    // Inline all plugin encoding (no function calls)
    for (size_t i = 0; i < reg->plugin_count; i++) {
        const SDPPlugin* plugin = &reg->plugins[i];
        
        // Plugin name
        *(uint32_t*)(buf + offset) = plugin->name_len;
        offset += 4;
        memcpy(buf + offset, plugin->name, plugin->name_len);
        offset += plugin->name_len;
        
        // Parameters array count
        *(uint32_t*)(buf + offset) = plugin->param_count;
        offset += 4;
        
        // DOUBLE INLINE: parameter encoding inside plugin loop
        for (size_t j = 0; j < plugin->param_count; j++) {
            const SDPParameter* p = &plugin->params[j];
            
            // ALL parameter fields inlined (no function call!)
            *(uint64_t*)(buf + offset) = p->address;
            offset += 8;
            
            *(uint32_t*)(buf + offset) = p->display_name_len;
            offset += 4;
            memcpy(buf + offset, p->display_name, p->display_name_len);
            offset += p->display_name_len;
            
            // ... all other parameter fields inlined
        }
    }
    
    return offset;
}
```

**Performance:** Expected ~20-25 µs on AudioUnit (vs current 65.34 µs, Go's 37.4 µs)

For `audiounit.sdp`:

```c
// sdp_types.h
#ifndef SDP_AUDIOUNIT_TYPES_H
#define SDP_AUDIOUNIT_TYPES_H

#include <stdint.h>
#include <stddef.h>
#include <stdbool.h>

// Writer type (opaque)
typedef struct SDPWriter SDPWriter;

// Result codes
typedef enum {
    SDP_OK = 0,
    SDP_ERR_OUT_OF_MEMORY = 1,
    SDP_ERR_INVALID_UTF8 = 2,
} SDPResult;

#endif
```

```c
// sdp_encode.h
#ifndef SDP_AUDIOUNIT_ENCODE_H
#define SDP_AUDIOUNIT_ENCODE_H

#include "sdp_types.h"

// Writer lifecycle
SDPWriter* sdp_writer_new(size_t capacity_hint);
const uint8_t* sdp_writer_bytes(const SDPWriter* w, size_t* out_len);
uint8_t* sdp_writer_take(SDPWriter* w, size_t* out_len);
void sdp_writer_free(SDPWriter* w);

// Primitive writers
void sdp_writer_u8(SDPWriter* w, uint8_t value);
void sdp_writer_u16(SDPWriter* w, uint16_t value);
void sdp_writer_u32(SDPWriter* w, uint32_t value);
void sdp_writer_u64(SDPWriter* w, uint64_t value);
void sdp_writer_i8(SDPWriter* w, int8_t value);
void sdp_writer_i16(SDPWriter* w, int16_t value);
void sdp_writer_i32(SDPWriter* w, int32_t value);
void sdp_writer_i64(SDPWriter* w, int64_t value);
void sdp_writer_f32(SDPWriter* w, float value);
void sdp_writer_f64(SDPWriter* w, double value);
void sdp_writer_bool(SDPWriter* w, bool value);
void sdp_writer_string(SDPWriter* w, const char* str, size_t len);
void sdp_writer_cstring(SDPWriter* w, const char* str);

// Array helpers
void sdp_writer_array_begin(SDPWriter* w, uint32_t count);
void sdp_writer_array_end(SDPWriter* w);

// High-level struct encoders (generated)
void sdp_encode_parameter(SDPWriter* w, uint32_t id, const char* name, size_t name_len, double value);
void sdp_encode_plugin(SDPWriter* w, const char* name, size_t name_len, 
                       const char* manufacturer_id, size_t manufacturer_id_len,
                       const char* component_type, size_t component_type_len,
                       const char* component_subtype, size_t component_subtype_len,
                       /* parameter array passed separately */);

#endif
```

## Usage Examples

### Example 1: Encoding a Plugin

```c
#include "sdp_encode.h"

void encode_my_plugin(void) {
    SDPWriter* w = sdp_writer_new(512);
    
    // Plugin fields
    sdp_writer_cstring(w, "My Reverb");
    sdp_writer_cstring(w, "ACME");
    sdp_writer_cstring(w, "aufx");
    sdp_writer_cstring(w, "rvb1");
    
    // Parameters array
    sdp_writer_array_begin(w, 2);
    
    // Parameter 1
    sdp_writer_u32(w, 1);
    sdp_writer_cstring(w, "Mix");
    sdp_writer_f64(w, 0.5);
    
    // Parameter 2
    sdp_writer_u32(w, 2);
    sdp_writer_cstring(w, "Size");
    sdp_writer_f64(w, 0.75);
    
    sdp_writer_array_end(w);
    
    // Write to file
    size_t len;
    const uint8_t* bytes = sdp_writer_bytes(w, &len);
    FILE* f = fopen("plugin.bin", "wb");
    fwrite(bytes, 1, len, f);
    fclose(f);
    
    sdp_writer_free(w);
}
```

### Example 2: Encoding with Structs (Alternative High-Level API)

The generator could also produce struct-based encoding:

```c
// Generated struct types
typedef struct {
    uint32_t id;
    const char* name;
    size_t name_len;
    double value;
} SDPParameter;

typedef struct {
    const char* name;
    size_t name_len;
    const char* manufacturer_id;
    size_t manufacturer_id_len;
    const char* component_type;
    size_t component_type_len;
    const char* component_subtype;
    size_t component_subtype_len;
    const SDPParameter* parameters;
    uint32_t parameter_count;
} SDPPlugin;

// High-level encoder
void sdp_encode_plugin_struct(SDPWriter* w, const SDPPlugin* plugin) {
    sdp_writer_string(w, plugin->name, plugin->name_len);
    sdp_writer_string(w, plugin->manufacturer_id, plugin->manufacturer_id_len);
    sdp_writer_string(w, plugin->component_type, plugin->component_type_len);
    sdp_writer_string(w, plugin->component_subtype, plugin->component_subtype_len);
    
    sdp_writer_array_begin(w, plugin->parameter_count);
    for (uint32_t i = 0; i < plugin->parameter_count; i++) {
        sdp_writer_u32(w, plugin->parameters[i].id);
        sdp_writer_string(w, plugin->parameters[i].name, plugin->parameters[i].name_len);
        sdp_writer_f64(w, plugin->parameters[i].value);
    }
    sdp_writer_array_end(w);
}

// Usage
void encode_with_struct(void) {
    SDPParameter params[] = {
        { .id = 1, .name = "Mix", .name_len = 3, .value = 0.5 },
        { .id = 2, .name = "Size", .name_len = 4, .value = 0.75 },
    };
    
    SDPPlugin plugin = {
        .name = "My Reverb",
        .name_len = 9,
        .manufacturer_id = "ACME",
        .manufacturer_id_len = 4,
        .component_type = "aufx",
        .component_type_len = 4,
        .component_subtype = "rvb1",
        .component_subtype_len = 4,
        .parameters = params,
        .parameter_count = 2,
    };
    
    SDPWriter* w = sdp_writer_new(512);
    sdp_encode_plugin_struct(w, &plugin);
    
    size_t len;
    const uint8_t* bytes = sdp_writer_bytes(w, &len);
    // ... use bytes ...
    
    sdp_writer_free(w);
}
```

## Implementation Details

### SDPWriter Structure

```c
typedef struct {
    uint8_t* data;        // Buffer
    size_t capacity;      // Allocated size
    size_t len;           // Current write position
    int error;            // Error code (0 = no error)
    size_t* array_stack;  // Stack of array positions
    size_t array_depth;   // Current nesting level
} SDPWriter;
```

### Buffer Growth Strategy

When buffer is full:
1. Double capacity (growth factor = 2.0)
2. Minimum growth = 64 bytes
3. Use `realloc()` to resize

### Endianness Handling

```c
#include <endian.h>  // For Linux
// Or use platform-specific headers

static inline uint32_t encode_u32(uint32_t value) {
    return htole32(value);
}

static inline uint32_t decode_u32(uint32_t value) {
    return le32toh(value);
}
```

### Array Tracking

```c
void sdp_writer_array_begin(SDPWriter* w, uint32_t count) {
    // Push current position onto stack
    array_stack[array_depth++] = w->len;
    
    // Write count (will be validated at end)
    sdp_writer_u32(w, count);
}

void sdp_writer_array_end(SDPWriter* w) {
    // Pop position from stack
    size_t count_pos = array_stack[--array_depth];
    
    // Could validate: read count, verify we wrote that many items
    // (Requires tracking items written, adds complexity)
}
```

## Platform Support

The C API targets:

- **C Standard:** C99 or later
- **Platforms:** Linux, macOS, Windows, iOS, Android
- **Compilers:** GCC, Clang, MSVC
- **Endianness:** Little-endian (Intel/ARM), Big-endian (handled via `htole*()`)

### Platform-Specific Considerations

**Endianness headers:**
- Linux: `<endian.h>`
- macOS/iOS: `<libkern/OSByteOrder.h>` (use `OSSwapHostToLittleInt32()`)
- Windows: Manual implementation or `<winsock2.h>`

**Solution:** Provide portable `sdp_endian.h` header with abstraction.

## Testing Strategy

### Unit Tests

Test each primitive type encoder:
```c
void test_encode_u32(void) {
    SDPWriter* w = sdp_writer_new(16);
    sdp_writer_u32(w, 0x12345678);
    
    size_t len;
    const uint8_t* bytes = sdp_writer_bytes(w, &len);
    
    assert(len == 4);
    assert(bytes[0] == 0x78);  // Little-endian
    assert(bytes[1] == 0x56);
    assert(bytes[2] == 0x34);
    assert(bytes[3] == 0x12);
    
    sdp_writer_free(w);
}
```

### Cross-Language Tests

Encode data with C, decode with Go/Rust/Swift to verify compatibility:

```c
// C encoder
SDPWriter* w = sdp_writer_new(256);
sdp_encode_plugin_struct(w, &test_plugin);
const uint8_t* bytes = sdp_writer_bytes(w, &len);
write_file("test.bin", bytes, len);

// Then in Go/Rust/Swift tests:
// Read test.bin and verify it decodes correctly
```

## Performance Goals

Target performance (encoding):

- **Primitives:** < 5 CPU cycles per field
- **Strings:** ~10 ns + strlen time
- **Arrays:** ~5 ns overhead + item encoding time
- **Structs:** Sum of field times + ~20 ns overhead

Expected to be **comparable to Rust** (~7 ns for primitives) since:
- Direct memory writes
- Minimal branching
- Inline functions
- No allocations in hot path (buffer pre-grown)

## Performance Summary

Based on extensive micro-benchmarking (see `c_micro_bench/OPTIMIZATION_RESULTS.md`):

### Benchmark Results (Apple M1, -O3 -march=native)

| Schema | C Baseline | C Optimized | Go | C Speedup |
|--------|------------|-------------|----|-----------| 
| **Primitives** (12 fields) | 9.17 ns | **0.84 ns** | 25.77 ns | **30.7x** |
| **Arrays (small)** (4-5 elements) | 29.96 ns | 32.95 ns | 56.02 ns | **1.9x** |
| **Arrays (large)** (50 elements) | 606.37 ns | **292.35 ns** | - | **2.1x** |
| **Nested** (3 levels) | 2.51 ns | **0.48 ns** | 22.11 ns | **46x** |
| **Complex** (nested + arrays) | 28.00 ns | **24.23 ns** | 75.93 ns | **3.1x** |

### Key Findings

1. **Wire format structs** provide 10-30x speedup for primitive-heavy schemas
2. **Bulk memcpy** helps for arrays > 10 elements (2x speedup)
3. **Inline encoding** eliminates function call overhead (1.4-5x speedup)
4. **Pre-computed string lengths** are 9x faster than strlen, 55x faster than snprintf
5. **Small arrays** (< 10 elements) are faster with loops than bulk memcpy

### Expected AudioUnit Performance

Current C implementation: 65.34 µs (baseline, field-by-field)  
Go implementation: 37.4 µs (1.75x faster)  
**Optimized C target: ~20-25 µs (1.5-1.9x faster than Go)**

## Implementation Checklist

### Phase 1: Core API ✅
- [x] Specification complete
- [x] Optimization research complete
- [x] Handwritten reference implementations (c_micro_bench/)

### Phase 2: Code Generation (Next)
- [ ] Update C generator to emit struct-based API
- [ ] Generate wire format structs for fixed-layout types
- [ ] Generate inline encoding for nested/arrays
- [ ] Generate size calculation functions with macros
- [ ] Add compilation flags (-O3 -march=native recommended)

### Phase 3: Validation
- [ ] Regenerate all test schemas (primitives, arrays, nested, complex, audiounit)
- [ ] Run micro-benchmarks to verify speedups
- [ ] Compare generated code against handwritten benchmarks
- [ ] Cross-language compatibility tests

### Phase 4: Documentation
- [ ] Update QUICK_REFERENCE.md with C examples
- [ ] Add C performance guide
- [ ] Document when to use each optimization
- [ ] Migration guide from old API

## API Migration Notes

### Old API (Dynamic Writer)
```c
SDPWriter* w = sdp_writer_new(1024);
sdp_writer_u64(w, value);
sdp_writer_string(w, str, len);
const uint8_t* bytes = sdp_writer_bytes(w, &out_len);
sdp_writer_free(w);
```

### New API (Struct-Based, Zero-Allocation)
```c
SDPParameter param = {
    .address = value,
    .display_name = str,
    .display_name_len = len,
    // ...
};
size_t size = sdp_parameter_size(&param);
uint8_t buf[512];  // Stack or malloc
sdp_parameter_encode(&param, buf);
```

**Benefits of new API:**
- No allocations (user controls buffer)
- Easier to test (just struct comparison)
- Enables all optimizations (wire structs, inline, bulk)
- Matches Go's approach (pre-sized buffer)

## Future Work

### Decoding API (Not Yet Implemented)
```c
// Zero-copy decoding
SDPParameter param;
sdp_parameter_decode(&param, bytes, len);
// param.display_name points directly into bytes
```

### Optional Fields (Future Extension)
```c
// Nullable types
typedef struct {
    bool has_value;
    uint32_t value;
} SDPOptionalU32;
```

### Compression Integration
```c
// User can compress the output buffer
uint8_t buf[1024];
size_t written = sdp_parameter_encode(&param, buf);
compress(buf, written);  // User's choice of algorithm
```

---

**End of Specification**
