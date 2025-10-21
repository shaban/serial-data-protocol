# Current C API Problems & Solutions

## Current API Issues

### Problem 1: Decode Returns Pointers Into Buffer (Zero-Copy)
**Current Decode API:**
```c
SDPPlugin decoded;
int result = sdp_plugin_decode(&decoded, buffer, buffer_len);

// ❌ PROBLEM: Arrays are unusable!
// decoded.parameters is a pointer to encoded data, not decoded structs
for (size_t i = 0; i < decoded.parameters_len; i++) {
    // ❌ Can't access decoded.parameters[i] - it's pointing to wire format!
}

// ❌ PROBLEM: Strings not null-terminated
// decoded.name is NOT a C string - must use decoded.name_len
printf("%s\n", decoded.name);  // ❌ May read past end!
```

**Why This Happened:**
- Optimized for zero-copy (fastest possible decode: 2.94ns)
- Strings point into buffer (no malloc, no copy)
- Arrays stored as pointers to encoded data (can't decode all at once)

### Problem 2: Manual String Length Management
**Current Encode API:**
```c
SDPParameter param;
param.display_name = "Gain";
param.display_name_len = 4;  // ❌ User must manually compute!
param.identifier = "gain";
param.identifier_len = 4;     // ❌ Error-prone!
```

### Problem 3: Arrays Require Manual Iteration
```c
// ❌ Can't decode array of structs in one call
// Must manually iterate and decode each element
for (uint32_t i = 0; i < plugin.parameters_len; i++) {
    SDPParameter temp;
    int result = sdp_parameter_decode(&temp, buf + offset, buf_len - offset);
    // ... now what? Where to store temp?
}
```

## Solution: Two-Tier API

### Tier 1: High-Level API (Convenience, ~10-20ns overhead)
```c
/* Builder API - computes string lengths automatically */
SDPParameterBuilder* builder = sdp_parameter_builder_new();
sdp_parameter_builder_set_display_name(builder, "Gain");  // Auto strlen
sdp_parameter_builder_set_identifier(builder, "gain");
sdp_parameter_builder_set_min_value(builder, -96.0f);
// ... set other fields

uint8_t* buffer = malloc(4096);
size_t encoded_size = sdp_parameter_builder_encode(builder, buffer, 4096);
sdp_parameter_builder_free(builder);

/* Arena Decode API - allocates everything, null-terminates strings */
SDPArena* arena = sdp_arena_new(4096);
SDPPlugin* plugin = sdp_plugin_decode_arena(buffer, buffer_len, arena);

// ✅ Easy to use - everything is ready
printf("Plugin: %s\n", plugin->name);  // ✅ Null-terminated!
for (size_t i = 0; i < plugin->parameters_len; i++) {
    SDPParameter* p = &plugin->parameters[i];
    printf("  Param: %s (%.1f to %.1f)\n", 
           p->display_name, p->min_value, p->max_value);
}

sdp_arena_free(arena);  // ✅ Single free for everything
```

**Performance:** ~10ns for primitives, ~25ns for AudioUnit (2-3x slower than zero-copy, still 2x faster than Go!)

### Tier 2: Low-Level API (Maximum Speed, Current API)
```c
/* Expert API - manual control for hot paths */
SDPParameter param = {
    .display_name = "Gain",
    .display_name_len = 4,  // Manual but explicit
    .identifier = "gain",
    .identifier_len = 4,
    // ... other fields
};

size_t size = sdp_parameter_size(&param);
uint8_t* buffer = malloc(size);
size_t encoded = sdp_parameter_encode(&param, buffer);

/* Zero-copy decode - 2.94ns, but requires careful handling */
SDPParameter decoded;
int result = sdp_parameter_decode(&decoded, buffer, size);
// Strings point into buffer - must keep buffer alive
// No null terminators - must use _len fields
```

**Performance:** 2.94ns primitives, 5.56ns Parameter (fastest possible)

## Recommended Implementation Plan

### Phase 1: Add Helper Macros (Quick Win - 30 min)
```c
/* In types.h - make string assignment easier */
#define SDP_STR(dest, str) do { \
    (dest) = (str); \
    (dest##_len) = strlen(str); \
} while(0)

// Usage:
SDPParameter param;
SDP_STR(param.display_name, "Gain");  // Auto-computes length
```

### Phase 2: Add Arena Allocator (2-3 hours)
```c
/* arena.h - Bump allocator for zero-free decode */
typedef struct SDPArena SDPArena;

SDPArena* sdp_arena_new(size_t initial_capacity);
void* sdp_arena_alloc(SDPArena* arena, size_t size);
void sdp_arena_free(SDPArena* arena);  // Frees everything
void sdp_arena_reset(SDPArena* arena); // Reuse without realloc

/* In decode.h - arena-based decode variants */
SDPParameter* sdp_parameter_decode_arena(
    const uint8_t* buf, size_t buf_len, SDPArena* arena);

SDPPlugin* sdp_plugin_decode_arena(
    const uint8_t* buf, size_t buf_len, SDPArena* arena);
```

**Implementation:**
- Single malloc for arena
- Bump pointer allocation (no fragmentation)
- Null-terminate strings automatically
- Decode arrays into contiguous memory
- ~8x slower than zero-copy (2.71ns → 23.5ns), still fast!

### Phase 3: Add Builder API (Optional - 3-4 hours)
```c
/* builder.h - Type-safe builder with auto strlen */
typedef struct SDPParameterBuilder SDPParameterBuilder;

SDPParameterBuilder* sdp_parameter_builder_new(void);
void sdp_parameter_builder_set_address(SDPParameterBuilder* b, uint64_t addr);
void sdp_parameter_builder_set_display_name(SDPParameterBuilder* b, const char* str);
// ... setters for each field

size_t sdp_parameter_builder_encode(
    SDPParameterBuilder* b, uint8_t* buf, size_t buf_len);
void sdp_parameter_builder_free(SDPParameterBuilder* b);
```

## Proposed Documentation Update

### Quick Start (Arena API)
```c
#include "types.h"
#include "encode.h"
#include "decode.h"
#include "arena.h"

// Encode
SDPParameter param;
SDP_STR(param.display_name, "Gain");  // Helper macro
param.min_value = -96.0f;
// ...

uint8_t buffer[256];
size_t size = sdp_parameter_encode(&param, buffer);

// Decode (easy mode)
SDPArena* arena = sdp_arena_new(256);
SDPParameter* decoded = sdp_parameter_decode_arena(buffer, size, arena);
printf("Name: %s\n", decoded->display_name);  // Null-terminated!
sdp_arena_free(arena);
```

### Performance Guide
- **Arena API:** Easiest to use, ~10-25ns, single free
- **Zero-Copy API:** Expert mode, 2-6ns, manual lifetime management
- **Builder API:** Most convenient, auto strlen, ~15-30ns

## Files to Create

1. **internal/generator/c/arena_gen.go** - Generate arena.h and arena.c
2. **internal/generator/c/decode_arena_gen.go** - Generate arena decode variants
3. **internal/generator/c/builder_gen.go** (optional) - Generate builder API
4. **Update decode_gen.go** - Keep zero-copy as default, add arena variants

## Benchmark Comparison (Estimated)

| API Style | Primitives | AudioUnit | Ease of Use |
|-----------|------------|-----------|-------------|
| **Zero-Copy** (current) | 2.94 ns | 5.56 ns | ⭐ Expert |
| **Arena** (proposed) | ~10 ns | ~25 ns | ⭐⭐⭐⭐⭐ Easy |
| **Builder** (optional) | ~15 ns | ~30 ns | ⭐⭐⭐⭐⭐ Easiest |
| **Go Baseline** | 25.9 ns | 130.7 ns | ⭐⭐⭐⭐ Good |

**Winner:** Arena API gives 2-5x speedup over Go with much better ergonomics than zero-copy!
