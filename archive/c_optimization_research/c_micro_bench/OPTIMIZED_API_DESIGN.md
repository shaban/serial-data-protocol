# Optimized C API Design

## Current API (Writer-based, 65.34 µs on AudioUnit)

### Current Approach
```c
// Dynamic writer with realloc
SDPWriter* w = sdp_writer_new(1024);

// Encode parameter (field-by-field)
sdp_encode_parameter(
    w,
    0x12345678,
    "Input Gain", 10,  // pre-computed length
    "gain", 4,
    "dB", 2,
    -96.0f, 12.0f, 0.0f, -6.0f,
    0x00000001,
    true, true
);

// Get result
size_t len;
const uint8_t* bytes = sdp_writer_bytes(w, &len);
sdp_writer_free(w);
```

**Problems:**
1. Dynamic allocation (writer manages buffer)
2. Field-by-field encoding (no bulk operations)
3. No compile-time layout knowledge used
4. Function call overhead for nested structs

---

## Optimized API Design

### Option 1: Pre-Sized Buffer (Like Go)

**User provides buffer, encoder fills it:**

```c
// Calculate size first (generated function with pre-computed fixed portion)
size_t size = sdp_calculate_parameter_size(
    display_name_len,
    identifier_len,
    unit_len
);

// User allocates (stack or heap)
uint8_t buf[512];  // or malloc

// Encode directly into buffer
size_t written = sdp_encode_parameter(
    buf,
    0x12345678,
    "Input Gain", 10,
    "gain", 4,
    "dB", 2,
    -96.0f, 12.0f, 0.0f, -6.0f,
    0x00000001,
    true, true
);
```

**Benefits:**
- No allocations (user controls)
- Single memcpy for fixed portion (wire struct optimization)
- Can use stack buffer for small messages
- Clear ownership model

**API:**
```c
// Calculate size (cheap - uses pre-computed constants)
size_t sdp_calculate_parameter_size(
    size_t display_name_len,
    size_t identifier_len,
    size_t unit_len
);

// Encode to buffer (returns bytes written)
size_t sdp_encode_parameter(
    uint8_t* buf,
    // ... all fields
);
```

---

### Option 2: Struct-Based (Most Ergonomic)

**User creates struct, encoder handles it:**

```c
typedef struct {
    uint64_t address;
    const char* display_name;
    size_t display_name_len;
    const char* identifier;
    size_t identifier_len;
    const char* unit;
    size_t unit_len;
    float min_value;
    float max_value;
    float default_value;
    float current_value;
    uint32_t raw_flags;
    bool is_writable;
    bool can_ramp;
} SDPParameter;

// Usage
SDPParameter param = {
    .address = 0x12345678,
    .display_name = "Input Gain",
    .display_name_len = 10,
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

// Calculate + encode
size_t size = sdp_parameter_size(&param);
uint8_t* buf = malloc(size);
sdp_parameter_encode(&param, buf);
```

**Benefits:**
- Clean separation of data and encoding
- Easier to construct incrementally
- Matches our benchmark test data pattern
- Can optimize struct layout separately

**Generated Code Uses Wire Structs:**
```c
// Internal wire format (generated, not exposed)
typedef struct __attribute__((packed)) {
    uint64_t address;
    // strings come after
    float min_value;
    float max_value;
    float default_value;
    float current_value;
    uint32_t raw_flags;
    uint8_t is_writable;
    uint8_t can_ramp;
} SDPParameterWire;  // 34 bytes fixed

size_t sdp_parameter_encode(const SDPParameter* src, uint8_t* buf) {
    size_t offset = 0;
    
    // Fixed portion: bulk copy via wire struct
    SDPParameterWire wire = {
        .address = src->address,
        .min_value = src->min_value,
        .max_value = src->max_value,
        .default_value = src->default_value,
        .current_value = src->current_value,
        .raw_flags = src->raw_flags,
        .is_writable = src->is_writable ? 1 : 0,
        .can_ramp = src->can_ramp ? 1 : 0
    };
    
    *(uint64_t*)(buf + offset) = wire.address;
    offset += 8;
    
    // String 1: display_name
    *(uint32_t*)(buf + offset) = src->display_name_len;
    offset += 4;
    memcpy(buf + offset, src->display_name, src->display_name_len);
    offset += src->display_name_len;
    
    // String 2: identifier
    *(uint32_t*)(buf + offset) = src->identifier_len;
    offset += 4;
    memcpy(buf + offset, src->identifier, src->identifier_len);
    offset += src->identifier_len;
    
    // String 3: unit
    *(uint32_t*)(buf + offset) = src->unit_len;
    offset += 4;
    memcpy(buf + offset, src->unit, src->unit_len);
    offset += src->unit_len;
    
    // Remaining fixed fields (could also be in wire struct)
    *(float*)(buf + offset) = wire.min_value;
    offset += 4;
    *(float*)(buf + offset) = wire.max_value;
    offset += 4;
    *(float*)(buf + offset) = wire.default_value;
    offset += 4;
    *(float*)(buf + offset) = wire.current_value;
    offset += 4;
    *(uint32_t*)(buf + offset) = wire.raw_flags;
    offset += 4;
    buf[offset++] = wire.is_writable;
    buf[offset++] = wire.can_ramp;
    
    return offset;
}
```

---

## Array Encoding

### Current (Field-by-field)
```c
sdp_writer_array_begin(w, param_count);
for (size_t i = 0; i < param_count; i++) {
    sdp_encode_parameter(w, params[i]...);
}
sdp_writer_array_end(w);
```

### Optimized (Inline + Bulk)
```c
// For arrays of structs
size_t sdp_parameter_array_encode(
    const SDPParameter* params,
    size_t count,
    uint8_t* buf
) {
    size_t offset = 0;
    
    // Count
    *(uint32_t*)(buf + offset) = count;
    offset += 4;
    
    // Inline all encoding (avoid function calls)
    for (size_t i = 0; i < count; i++) {
        const SDPParameter* p = &params[i];
        
        // Inline the encoding (no function call!)
        *(uint64_t*)(buf + offset) = p->address;
        offset += 8;
        
        // String 1
        *(uint32_t*)(buf + offset) = p->display_name_len;
        offset += 4;
        memcpy(buf + offset, p->display_name, p->display_name_len);
        offset += p->display_name_len;
        
        // ... rest inlined
    }
    
    return offset;
}
```

**For primitive arrays:**
```c
// Bulk copy (no loop!)
size_t sdp_u32_array_encode(const uint32_t* arr, size_t count, uint8_t* buf) {
    *(uint32_t*)buf = count;
    memcpy(buf + 4, arr, count * 4);
    return 4 + count * 4;
}

// String arrays still need loop
size_t sdp_string_array_encode(
    const char** strings,
    const size_t* lengths,
    size_t count,
    uint8_t* buf
) {
    size_t offset = 0;
    *(uint32_t*)(buf + offset) = count;
    offset += 4;
    
    for (size_t i = 0; i < count; i++) {
        *(uint32_t*)(buf + offset) = lengths[i];
        offset += 4;
        memcpy(buf + offset, strings[i], lengths[i]);
        offset += lengths[i];
    }
    
    return offset;
}
```

---

## Layout Macros (Compile-Time Knowledge)

### What Can Be Pre-Computed?

**Schema: Parameter**
```
struct Parameter {
    address: u64           // 8 bytes (offset 0)
    display_name: str      // variable
    identifier: str        // variable
    unit: str              // variable
    min_value: f32         // 4 bytes
    max_value: f32         // 4 bytes
    default_value: f32     // 4 bytes
    current_value: f32     // 4 bytes
    raw_flags: u32         // 4 bytes
    is_writable: bool      // 1 byte
    can_ramp: bool         // 1 byte
}
```

**Generated Macros:**
```c
// Fixed portion size (all non-variable fields)
#define SDP_PARAMETER_FIXED_SIZE 34

// Field offsets in wire format (if fields were contiguous)
#define SDP_PARAMETER_ADDRESS_OFFSET 0
#define SDP_PARAMETER_MIN_VALUE_OFFSET 8  // after variable strings
// ... but strings move these, so offsets are dynamic

// What we CAN pre-compute:
#define SDP_PARAMETER_FLOAT_FIELDS_SIZE 16  // 4 floats
#define SDP_PARAMETER_PRIMITIVE_SIZE 34     // all fixed fields
```

**Usage in size calculation:**
```c
size_t sdp_parameter_size(const SDPParameter* p) {
    return SDP_PARAMETER_PRIMITIVE_SIZE  // 34 bytes (known at compile-time)
           + 4 + p->display_name_len     // string 1
           + 4 + p->identifier_len        // string 2
           + 4 + p->unit_len;             // string 3
}
```

### Wire Struct Layout (Generated)

```c
// For fields BEFORE variable-length data
typedef struct __attribute__((packed)) {
    uint64_t address;
} SDPParameterWirePrefix;

// For fields AFTER variable-length data
typedef struct __attribute__((packed)) {
    float min_value;
    float max_value;
    float default_value;
    float current_value;
    uint32_t raw_flags;
    uint8_t is_writable;
    uint8_t can_ramp;
} SDPParameterWireSuffix;

// Then bulk copy:
SDPParameterWirePrefix prefix = { .address = src->address };
memcpy(buf, &prefix, sizeof(prefix));

// ... strings ...

SDPParameterWireSuffix suffix = {
    .min_value = src->min_value,
    // ...
};
memcpy(buf + offset, &suffix, sizeof(suffix));
```

---

## Nested Encoding (Inline Everything)

### Current (Recursive Calls)
```c
void sdp_encode_plugin(SDPWriter* w, Plugin* plugin) {
    sdp_writer_string(w, plugin->name, plugin->name_len);
    // ... other fields
    
    sdp_writer_array_begin(w, plugin->param_count);
    for (size_t i = 0; i < plugin->param_count; i++) {
        sdp_encode_parameter(w, plugin->params[i]...);  // FUNCTION CALL
    }
    sdp_writer_array_end(w);
}
```

### Optimized (Inline)
```c
size_t sdp_plugin_encode(const SDPPlugin* plugin, uint8_t* buf) {
    size_t offset = 0;
    
    // Plugin name
    *(uint32_t*)(buf + offset) = plugin->name_len;
    offset += 4;
    memcpy(buf + offset, plugin->name, plugin->name_len);
    offset += plugin->name_len;
    
    // ... other plugin fields
    
    // Parameters array - INLINE encoding (no function calls!)
    *(uint32_t*)(buf + offset) = plugin->param_count;
    offset += 4;
    
    for (size_t i = 0; i < plugin->param_count; i++) {
        const SDPParameter* p = &plugin->params[i];
        
        // INLINE all parameter encoding here
        *(uint64_t*)(buf + offset) = p->address;
        offset += 8;
        
        *(uint32_t*)(buf + offset) = p->display_name_len;
        offset += 4;
        memcpy(buf + offset, p->display_name, p->display_name_len);
        offset += p->display_name_len;
        
        // ... all other parameter fields inlined
    }
    
    return offset;
}
```

**Speedup:** 3-5x (eliminates function call overhead, 1.4x from our complex benchmark)

---

## Full AudioUnit Example

### API Design

```c
// Data structures (user-facing)
typedef struct {
    uint64_t address;
    const char* display_name;
    size_t display_name_len;
    // ... all parameter fields
} SDPParameter;

typedef struct {
    const char* name;
    size_t name_len;
    const char* manufacturer_id;
    size_t manufacturer_id_len;
    // ... other plugin fields
    SDPParameter* parameters;
    size_t parameter_count;
} SDPPlugin;

typedef struct {
    SDPPlugin* plugins;
    size_t plugin_count;
    uint32_t total_parameter_count;
} SDPPluginRegistry;

// API
size_t sdp_plugin_registry_size(const SDPPluginRegistry* reg);
size_t sdp_plugin_registry_encode(const SDPPluginRegistry* reg, uint8_t* buf);
```

### Usage

```c
// Construct data
SDPParameter params[] = {
    { .address = 0x1000, .display_name = "Gain", .display_name_len = 4, ... },
    { .address = 0x2000, .display_name = "Pan", .display_name_len = 3, ... },
};

SDPPlugin plugins[] = {
    {
        .name = "Reverb",
        .name_len = 6,
        .parameters = params,
        .parameter_count = 2,
        ...
    },
};

SDPPluginRegistry registry = {
    .plugins = plugins,
    .plugin_count = 1,
    .total_parameter_count = 2,
};

// Encode
size_t size = sdp_plugin_registry_size(&registry);
uint8_t* buf = malloc(size);
size_t written = sdp_plugin_registry_encode(&registry, buf);

// Use buf...
free(buf);
```

### Generated Encoder (Optimized)

```c
size_t sdp_plugin_registry_encode(const SDPPluginRegistry* reg, uint8_t* buf) {
    size_t offset = 0;
    
    // Registry counts
    *(uint32_t*)(buf + offset) = reg->plugin_count;
    offset += 4;
    *(uint32_t*)(buf + offset) = reg->total_parameter_count;
    offset += 4;
    
    // Plugins array - INLINE EVERYTHING
    for (size_t i = 0; i < reg->plugin_count; i++) {
        const SDPPlugin* plugin = &reg->plugins[i];
        
        // Plugin name
        *(uint32_t*)(buf + offset) = plugin->name_len;
        offset += 4;
        memcpy(buf + offset, plugin->name, plugin->name_len);
        offset += plugin->name_len;
        
        // Plugin manufacturer_id
        *(uint32_t*)(buf + offset) = plugin->manufacturer_id_len;
        offset += 4;
        memcpy(buf + offset, plugin->manufacturer_id, plugin->manufacturer_id_len);
        offset += plugin->manufacturer_id_len;
        
        // ... other plugin fields
        
        // Parameters array - DOUBLE INLINE (no function calls at all!)
        *(uint32_t*)(buf + offset) = plugin->parameter_count;
        offset += 4;
        
        for (size_t j = 0; j < plugin->parameter_count; j++) {
            const SDPParameter* p = &plugin->parameters[j];
            
            // ALL parameter encoding inlined here
            *(uint64_t*)(buf + offset) = p->address;
            offset += 8;
            
            *(uint32_t*)(buf + offset) = p->display_name_len;
            offset += 4;
            memcpy(buf + offset, p->display_name, p->display_name_len);
            offset += p->display_name_len;
            
            *(uint32_t*)(buf + offset) = p->identifier_len;
            offset += 4;
            memcpy(buf + offset, p->identifier, p->identifier_len);
            offset += p->identifier_len;
            
            *(uint32_t*)(buf + offset) = p->unit_len;
            offset += 4;
            memcpy(buf + offset, p->unit, p->unit_len);
            offset += p->unit_len;
            
            *(float*)(buf + offset) = p->min_value;
            offset += 4;
            *(float*)(buf + offset) = p->max_value;
            offset += 4;
            *(float*)(buf + offset) = p->default_value;
            offset += 4;
            *(float*)(buf + offset) = p->current_value;
            offset += 4;
            
            *(uint32_t*)(buf + offset) = p->raw_flags;
            offset += 4;
            
            buf[offset++] = p->is_writable ? 1 : 0;
            buf[offset++] = p->can_ramp ? 1 : 0;
        }
    }
    
    return offset;
}
```

**Key Optimizations:**
1. ✅ No dynamic allocations (user-provided buffer)
2. ✅ No function calls (everything inlined)
3. ✅ Direct offset arithmetic (no pointer updates)
4. ✅ Bulk memcpy for strings (pre-computed lengths)
5. ✅ Could add wire structs for fixed portions (further optimization)

---

## Optimization Decision Tree

### When to Use Each Technique

**1. Wire Structs** (10-30x speedup)
- Use when: Struct has only fixed-size fields OR fixed prefix/suffix
- Skip when: Struct is mostly strings/arrays
- Example: Primitives (30.7x faster), Nested (46x faster)

**2. Bulk Array Copy** (2-5x speedup)
- Use when: Array elements are primitive types (u8, u32, f64, etc.)
- Skip when: Array size < 10 (loop is faster)
- Skip when: Array of strings (need loop anyway)
- Example: Large arrays (2.1x faster), Small arrays (0.9x - worse!)

**3. Inline Nested Encoding** (1.4-5x speedup)
- Use when: Nested struct has < 20 fields
- Use when: Called in hot path (loops)
- Skip when: Struct is huge (code bloat)
- Example: Complex (1.2x faster - 7 function calls eliminated)

**4. Pre-Computed String Lengths** (9x speedup)
- Use: ALWAYS! Make caller provide lengths
- Never use: strlen() in generated code
- API: All string parameters take `const char* str, size_t len`

**5. Adaptive Array Threshold**
- Arrays < 10 elements: Use loop (lower overhead)
- Arrays >= 10 elements: Use bulk memcpy
- Generator can emit: `if (count < 10) { loop } else { memcpy }`
- Or: Just use loop for simplicity (1.9x still faster than Go)

---

## Recommended API (Final Design)

### Struct-Based with Pre-Sized Buffers

```c
// 1. User defines data in structs
SDPParameter param = {
    .address = 0x1000,
    .display_name = "Gain",
    .display_name_len = 4,  // PRE-COMPUTED (required)
    // ...
};

// 2. Calculate size (cheap, uses macros for fixed portion)
size_t size = sdp_parameter_size(&param);

// 3. Allocate buffer (user controls: stack, heap, or static)
uint8_t buf[512];  // stack
// or: uint8_t* buf = malloc(size);  // heap

// 4. Encode (optimized: inline, bulk, wire structs)
size_t written = sdp_parameter_encode(&param, buf);

// 5. Use buffer
send_to_network(buf, written);
```

**Benefits:**
- Clean separation (data vs encoding)
- No hidden allocations
- Easy to test (just struct equality)
- Full optimization potential
- Matches our benchmark pattern (pre-populated structs)

**Generated API:**
```c
// Types
typedef struct { /* fields */ } SDPParameter;
typedef struct { /* fields */ } SDPPlugin;

// Size calculation (uses macros, cheap)
size_t sdp_parameter_size(const SDPParameter* p);
size_t sdp_plugin_size(const SDPPlugin* p);

// Encoding (optimized, no allocations)
size_t sdp_parameter_encode(const SDPParameter* p, uint8_t* buf);
size_t sdp_plugin_encode(const SDPPlugin* p, uint8_t* buf);

// Macros (compile-time constants)
#define SDP_PARAMETER_FIXED_SIZE 34
#define SDP_PLUGIN_FIXED_SIZE 12
```

---

## Migration Path

### Phase 1: Change API (buffer-based)
- Remove `SDPWriter*` 
- Add `uint8_t* buf` parameter
- Generate size calculation functions
- **Expected:** Match Go's allocation pattern

### Phase 2: Add Wire Structs
- Generate `typedef struct __attribute__((packed))` for fixed layouts
- Use bulk memcpy for fixed portions
- **Expected:** 10-30x speedup on primitives/nested

### Phase 3: Inline Arrays
- Generate inline encoding for arrays of structs
- Use adaptive threshold for primitive arrays (< 10: loop, >= 10: bulk)
- **Expected:** 1.5-2x speedup on complex schemas

### Phase 4: Macro Optimization
- Add compile-time constants for fixed sizes
- Pre-compute offsets where possible
- **Expected:** 1.2-1.5x additional speedup

**Final Expected Performance:**
- Primitives: ~0.8 ns (30x faster than Go's 25.77 ns) ✅ Already proven
- Arrays: ~30 ns (1.8x faster than Go's 56.02 ns) ✅ Already proven
- Nested: ~0.5 ns (44x faster than Go's 22.11 ns) ✅ Already proven
- Complex: ~24 ns (3.1x faster than Go's 75.93 ns) ✅ Already proven
- **AudioUnit: ~20-25 µs (1.5-1.9x faster than Go's 37.4 µs)** 🎯 TARGET

---

## Summary

**Intended Usage:**
```c
// 1. Define struct
SDPPluginRegistry registry = { /* data */ };

// 2. Calculate size
size_t size = sdp_plugin_registry_size(&registry);

// 3. Allocate
uint8_t* buf = malloc(size);

// 4. Encode
sdp_plugin_registry_encode(&registry, buf);

// 5. Use
send(buf, size);
free(buf);
```

**Optimizations Requiring Layout Info:**
1. ✅ **Wire structs** - Fixed-size portion known at schema time
2. ✅ **Size macros** - `#define SDP_PARAMETER_FIXED_SIZE 34`
3. ✅ **Field offsets** - Only for all-fixed structs (rare)
4. ❌ **String lengths** - Runtime only
5. ❌ **Array lengths** - Runtime only

**Code Generation Strategy:**
- Emit struct definitions with pre-computed length fields
- Emit size functions using fixed-size macros
- Emit encode functions with inline nested encoding
- Use wire structs for fixed portions (prefix/suffix pattern)
- Adaptive array threshold (< 10: loop, >= 10: bulk)

**Result:** Should close the 1.75x gap and potentially beat Go!
