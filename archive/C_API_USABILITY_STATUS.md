# C API Usability Improvements - Status

## Problem Statement

The initial zero-copy C decoder API is **extremely fast** (2.94-5.56ns) but **hard to use**:

### Current Pain Points
1. **Manual string length management** - User must call `strlen()` and set `_len` fields
2. **Non-null-terminated strings** - Decoded strings point into buffer, no `\0`
3. **Unusable arrays** - Arrays of structs point to encoded data, can't access elements
4. **Manual iteration** - Must manually iterate and decode array elements with careful offset tracking

### Example of Current Problems
```c
// ❌ Manual strlen for every string
param.display_name = "Gain";
param.display_name_len = 4;  // Easy to get wrong!

// ❌ Decoded strings not null-terminated
SDPParameter decoded;
sdp_parameter_decode(&decoded, buffer, size);
printf("%s\n", decoded.display_name);  // ❌ May read past end!
printf("%.*s\n", (int)decoded.display_name_len, decoded.display_name);  // ✅ Must use this

// ❌ Arrays completely unusable
SDPPlugin plugin;
sdp_plugin_decode(&plugin, buffer, size);
plugin.parameters[0].display_name;  // ❌ Segfault! Points to wire format!
```

## Solutions Implemented

### ✅ Phase 1: SDP_STR Helper Macro (DONE)

**File:** `internal/generator/c/types_gen.go`

Automatic `strlen()` computation for encode:

```c
SDPParameter param;
SDP_STR(param.display_name, "Gain");  // ✅ Auto-computes length!
SDP_STR(param.identifier, "gain");
SDP_STR(param.unit, "dB");
```

**Benefits:**
- No manual `strlen()` calls
- No risk of wrong length
- Cleaner, less error-prone code

**Performance:** Same as manual (0ns overhead, just a macro)

### ✅ Phase 2: Arena Allocator (DONE)

**Files:** 
- `internal/generator/c/arena_gen.go` 
- Generated: `arena.h`, `arena.c`

Bump allocator for efficient decode memory management:

```c
SDPArena* arena = sdp_arena_new(1024);

// Allocate memory (8-byte aligned, grows automatically)
char* str = sdp_arena_alloc(arena, 32);
int* nums = sdp_arena_alloc(arena, sizeof(int) * 100);

// Reuse arena (reset without freeing)
sdp_arena_reset(arena);

// Free everything at once
sdp_arena_free(arena);
```

**Features:**
- Single `malloc()` for arena
- Bump pointer allocation (no fragmentation)
- Auto-grows with `realloc()` if needed
- 8-byte alignment for performance
- Single `free()` for all allocations
- Can reset and reuse

**Performance:** ~8x slower than zero-copy (2.71ns vs 0.34ns from manual benchmarks)

### ⏳ Phase 3: Arena Decode API (TODO)

**File:** `internal/generator/c/decode_arena_gen.go` (needs to be created)

Generate arena-based decode functions:

```c
// Desired API
SDPArena* arena = sdp_arena_new(1024);
SDPParameter* param = sdp_parameter_decode_arena(buffer, size, arena);

// ✅ Strings are null-terminated!
printf("Name: %s\n", param->display_name);  

// ✅ Arrays work naturally!
SDPPlugin* plugin = sdp_plugin_decode_arena(buffer, size, arena);
for (size_t i = 0; i < plugin->parameters_len; i++) {
    printf("Param[%zu]: %s\n", i, plugin->parameters[i].display_name);
}

sdp_arena_free(arena);  // ✅ Single free!
```

**Implementation Plan:**

1. **Create `decode_arena_gen.go`:**
   - `GenerateDecodeArena(schema, packageName)` for declarations in decode.h
   - `GenerateDecodeArenaImpl(schema, packageName)` for implementations in decode.c
   - For each struct, generate `sdp_TYPE_decode_arena()`

2. **Arena decode logic:**
   ```c
   SDPParameter* sdp_parameter_decode_arena(const uint8_t* buf, size_t buf_len, SDPArena* arena) {
       // Allocate struct
       SDPParameter* param = sdp_arena_alloc(arena, sizeof(SDPParameter));
       
       // Decode primitives (same as zero-copy)
       param->address = read_u64(buf, &offset);
       
       // Decode strings (copy + null-terminate)
       uint32_t len = read_u32(buf, &offset);
       char* str = sdp_arena_alloc(arena, len + 1);  // +1 for null terminator
       memcpy(str, buf + offset, len);
       str[len] = '\0';  // ✅ Null-terminate!
       param->display_name = str;
       param->display_name_len = len;
       
       // Decode arrays (allocate + decode each element)
       uint32_t count = read_u32(buf, &offset);
       SDPSubStruct* array = sdp_arena_alloc(arena, sizeof(SDPSubStruct) * count);
       for (uint32_t i = 0; i < count; i++) {
           sdp_substruct_decode_arena_inplace(&array[i], buf, &offset, arena);
       }
       param->array = array;
       param->array_len = count;
       
       return param;
   }
   ```

3. **Update generator.go:**
   - Call `GenerateDecodeArena()` and `GenerateDecodeArenaImpl()`
   - Functions added to decode.h and decode.c

**Estimated Performance:** ~10-25ns (2-3x slower than zero-copy, still 2-5x faster than Go!)

## API Tiers Summary

| API | Speed | Ease of Use | Use Case |
|-----|-------|-------------|----------|
| **Zero-Copy** (current) | 2.94-5.56ns | ⭐ Expert | Hot paths, max speed |
| **Zero-Copy + SDP_STR** | 2.94-5.56ns | ⭐⭐ Better | When encode speed matters |
| **Arena** (coming) | ~10-25ns | ⭐⭐⭐⭐⭐ Easy | Default for most users |

### Comparison with Go

| Operation | Zero-Copy C | Arena C | Go |
|-----------|-------------|---------|-----|
| Primitives Encode | 8.28ns | 8.28ns | 25.9ns |
| Primitives Decode | 2.94ns | ~10ns | ~25ns* |
| AudioUnit Encode | 49.7ns | 49.7ns | 130.7ns |
| AudioUnit Decode | 5.56ns | ~25ns | ~130ns* |

\* Go decode speed estimated from encode benchmarks

**Key Insight:** Arena C API is still **2-5x faster than Go** while being much easier to use than zero-copy!

## Testing

### Current Tests
- ✅ Zero-copy decode works (primitives, nested structs)
- ✅ Arena allocator works (tested manually)
- ✅ SDP_STR macro works
- ✅ C encode → C decode roundtrip passes

### Tests Needed for Arena Decode
- Test arena decode with primitives
- Test arena decode with nested structs
- Test arena decode with string arrays
- Test arena decode with struct arrays
- Test arena decode with optionals
- Benchmark arena decode vs zero-copy vs Go

## Files

### Completed
- `internal/generator/c/arena_gen.go` - Arena allocator generation
- `internal/generator/c/types_gen.go` - Added SDP_STR macro
- `internal/generator/c/generator.go` - Added arena.h/arena.c generation
- `testdata/primitives_c/example_improved_api.c` - Demo of current improvements

### Needed
- `internal/generator/c/decode_arena_gen.go` - Arena decode functions
- Update `internal/generator/c/generator.go` - Call arena decode generation
- More test programs demonstrating arena decode
- Benchmarks comparing all three APIs

## Next Steps

1. **Create `decode_arena_gen.go`** (2-3 hours)
   - Generate arena decode function declarations
   - Generate arena decode implementations
   - Handle strings (copy + null-terminate)
   - Handle arrays (allocate + decode each element)
   - Handle nested structs (recursive arena decode)

2. **Regenerate schemas** (5 min)
   - All 6 testdata schemas get arena decode functions

3. **Test arena decode** (1 hour)
   - Write test programs
   - Verify null-terminated strings
   - Verify working arrays
   - Test nested structs

4. **Benchmark arena decode** (30 min)
   - Compare vs zero-copy (expect ~2-3x slower)
   - Compare vs Go (expect ~2-5x faster)

5. **Update documentation** (1 hour)
   - README examples showing both APIs
   - API comparison guide
   - Performance recommendations

## Recommendation

**Default API should be Arena** for best usability while maintaining good performance:
- ✅ 2-5x faster than Go
- ✅ Null-terminated strings (works with all C string functions)
- ✅ Working arrays (natural iteration)
- ✅ Single free (no memory leaks)
- ✅ Can reuse arena (zero allocations for repeated decodes)

**Keep Zero-Copy for experts** who need maximum speed:
- ✅ 24x faster than encode (0.34-5.56ns)
- ⚠️ Requires careful lifetime management
- ⚠️ Strings not null-terminated
- ⚠️ Arrays require manual iteration

## Current State

✅ **Phase 1 Complete:** SDP_STR macro makes encode easier  
✅ **Phase 2 Complete:** Arena allocator works and is tested  
⏳ **Phase 3 In Progress:** Need to generate arena decode functions

**Estimated time to complete:** ~4-5 hours of implementation + testing + docs
