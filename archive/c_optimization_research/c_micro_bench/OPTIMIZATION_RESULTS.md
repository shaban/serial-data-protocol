# C Encoder Optimization Research Results

## Executive Summary

Research conducted to close the **1.75x performance gap** between C encoder (65.34 µs) and Go encoder (37.4 µs) on the AudioUnit benchmark with real plugins.json data.

**Key Finding:** Function call overhead and field-by-field encoding are the primary bottlenecks. Wire format structs with bulk memcpy and inline encoding can deliver **1.2x to 10.9x speedups** depending on schema characteristics.

---

## Performance Baseline

### Real-World Benchmark (plugins.json)
- **Go:** 37.4 µs/op (baseline)
- **C:** 65.34 µs/op (current implementation)
- **Gap:** 1.75x slower

**Root Causes Identified:**
1. Recursive function calls for nested structs (2-9 calls per encode)
2. Field-by-field memcpy instead of bulk operations
3. No compile-time layout optimization
4. String operations (strlen) in hot path

---

## Schema Benchmark Results

### 1. Primitives (primitives.sdp)
**Schema:** 12 fields (11 primitives + 1 string), 60 bytes total

| Approach | ns/op | Speedup |
|----------|-------|---------|
| Baseline (field-by-field) | 9.17 | 1.0x |
| **Optimized (wire struct)** | **0.84** | **10.9x** |

**Optimization:** `AllPrimitivesWire` struct (43 bytes packed) prepared on stack, bulk memcpy for fixed portion.

**Impact:** 91% improvement - **highest gain**

---

### 2. Arrays (arrays.sdp)
**Schema:** 5 arrays (u8[], u32[], f64[], str[], bool[]), 50 elements each, 1170 bytes total

| Approach | ns/op | Speedup |
|----------|-------|---------|
| Baseline (loop per element) | 606.37 | 1.0x |
| **Optimized (bulk memcpy)** | **292.35** | **2.1x** |

**Optimization:** Single memcpy for primitive arrays (u8, u32, f64, bool). String array still requires loop (variable-length).

**Impact:** 52% improvement - **limited by string arrays**

---

### 3. Nested (nested.sdp)
**Schema:** Scene → Rectangle(Point, Point, u32) → Point(f32, f32), 37 bytes total

| Approach | ns/op | Speedup |
|----------|-------|---------|
| Baseline (recursive calls) | 2.51 | 1.0x |
| **Optimized (inline + wire)** | **0.48** | **5.3x** |

**Optimization:** `RectangleWire` struct (20 bytes packed), inline all nested encoding. **Avoids 3 function calls** per encode.

**Impact:** 81% improvement - **function call elimination is critical**

---

### 4. Complex (complex.sdp)
**Schema:** AudioDevice → []Plugin → []Parameter (3-level nesting), 110 bytes total

| Approach | ns/op | Speedup |
|----------|-------|---------|
| Baseline (recursive calls) | 28.00 | 1.0x |
| **Optimized (inline)** | **24.23** | **1.2x** |

**Optimization:** Inline all plugin and parameter encoding. **Avoids 7 function calls** per encode.

**Impact:** 13% improvement - **limited by string-heavy schema**

---

## Micro-Benchmark Results (Isolated Techniques)

### String Operations
| Method | ns/op | vs snprintf | vs strlen |
|--------|-------|-------------|-----------|
| snprintf | 39.20 | 1.0x | 6.0x |
| strlen + memcpy | 6.49 | 6.0x | 1.0x |
| **Pre-computed length** | **0.71** | **55x** | **9.1x** |

**Conclusion:** Caller MUST provide pre-computed string lengths. Never use strlen() in generated code.

---

### Array Encoding
| Method | ns/op (100 elements) | Speedup |
|--------|---------------------|---------|
| Loop (f32) | 106.36 | 1.0x |
| **Bulk memcpy (f32)** | **21.59** | **4.9x** |
| Loop (u32) | 78.81 | 1.0x |
| **Bulk memcpy (u32)** | **14.69** | **5.4x** |

**Conclusion:** Primitive arrays should use bulk memcpy. String arrays still need loops.

---

### Struct Layout
| Method | ns/op | Speedup |
|--------|-------|---------|
| Field-by-field | 6.29 | 1.0x |
| **Bulk copy (wire struct)** | **0.80** | **7.9x** |
| **Direct offsets** | **0.49** | **12.8x** |

**Conclusion:** Wire format structs enable massive speedups for fixed-layout types.

---

### Nested Struct Encoding
| Method | ns/op | Speedup |
|--------|-------|---------|
| Recursive function calls | 17.39 | 1.0x |
| Flattened (1 function) | 12.51 | 1.4x |
| **Bulk copy (wire struct)** | **1.34** | **13.0x** |

**Conclusion:** Function call overhead is significant. Inline nested encoding when possible.

---

## Optimization Patterns (Proven)

### 1. Wire Format Structs
```c
typedef struct __attribute__((packed)) {
    uint8_t field1;
    uint16_t field2;
    uint32_t field3;
    float field4;
} MyStructWire;

// Usage:
MyStructWire wire = { data.field1, data.field2, data.field3, data.field4 };
memcpy(buf, &wire, sizeof(wire));
buf += sizeof(wire);
```

**When to Use:** Structs with only fixed-size fields (no strings/arrays)  
**Speedup:** 7.9x - 13.0x  
**Tradeoff:** More complex codegen, but worth it

---

### 2. Bulk Array Copy
```c
// Primitive arrays (u8, u16, u32, u64, i8, i16, i32, i64, f32, f64, bool)
*(uint32_t*)buf = count;
buf += 4;
memcpy(buf, array, count * sizeof(element_type));
buf += count * sizeof(element_type);

// String arrays (still need loop)
*(uint32_t*)buf = count;
buf += 4;
for (size_t i = 0; i < count; i++) {
    *(uint32_t*)buf = string_lengths[i];
    buf += 4;
    memcpy(buf, strings[i], string_lengths[i]);
    buf += string_lengths[i];
}
```

**When to Use:** Arrays of primitive types  
**Speedup:** 4.9x - 5.4x  
**Tradeoff:** Simple, always beneficial

---

### 3. Inline Nested Encoding
```c
// Instead of:
encode_point(&data.top_left);
encode_point(&data.bottom_right);

// Do:
// top_left
*(float*)buf = data.top_left.x;
buf += 4;
*(float*)buf = data.top_left.y;
buf += 4;
// bottom_right
*(float*)buf = data.bottom_right.x;
buf += 4;
*(float*)buf = data.bottom_right.y;
buf += 4;
```

**When to Use:** Nested structs with small number of fields  
**Speedup:** 1.4x - 5.3x  
**Tradeoff:** More code, but avoids function call overhead

---

### 4. Pre-Computed String Lengths
```c
// API must require:
encode_my_struct(buf, data, name_len, description_len);

// Never:
size_t len = strlen(data->name);  // 6.49 ns
memcpy(buf, data->name, len);

// Always:
memcpy(buf, data->name, name_len);  // 0.71 ns
```

**When to Use:** All string fields  
**Speedup:** 9.1x vs strlen, 55x vs snprintf  
**Tradeoff:** Caller responsibility, but massive win

---

## Schema Knowledge Boundaries

### ✅ Valid Compile-Time Optimizations
- **Struct field offsets** in wire format (fixed at schema design)
- **Fixed-size portions** of structs (before variable fields)
- **Array element types** (enables type-specific bulk copy)
- **Nested struct layouts** (enables wire format structs)
- **Field count and order** (enables unrolled loops)

### ❌ Invalid Compile-Time Assumptions
- **String lengths** (runtime-variable per instance)
- **Array lengths** (runtime-variable per instance)
- **Total encoded size** (depends on variable fields)
- **String literal optimization** (only works for const test data)

**Key Insight:** Optimizations must work with volatile, runtime-variable data. Benchmarks using const test data will be misleading (compiler optimizes away).

---

## Recommendations for Official C Generator

### High Priority (Implement)
1. **Wire format structs** for fixed-layout types
   - Generate `__attribute__((packed))` structs
   - Use bulk memcpy for fixed portion
   - Expected: 8-13x speedup

2. **Bulk array copy** for primitive arrays
   - Single memcpy for u8/u16/u32/u64/i8/i16/i32/i64/f32/f64/bool
   - Loop only for string arrays
   - Expected: 5x speedup

3. **Pre-computed string lengths** in API
   - All `encode_*` functions take `*_len` parameters
   - Document in C_API_SPECIFICATION.md
   - Expected: 9x speedup vs strlen

4. **Inline small nested structs** (< 5 fields)
   - Avoid function call overhead
   - Generate flat encoding code
   - Expected: 1.4x - 5x speedup

### Medium Priority (Consider)
5. **Unrolled loops** for small fixed-size arrays
   - If array size is schema-defined constant
   - Avoid loop overhead
   - Expected: 1.5x - 2x speedup

6. **Direct offset calculation** for known layouts
   - Pre-compute fixed offsets in schema
   - Avoid incremental pointer arithmetic
   - Expected: 1.2x - 1.5x speedup

### Low Priority (Skip for now)
7. **SIMD operations** for large arrays
   - Complex, platform-specific
   - Requires alignment guarantees
   - Expected: 2x - 4x speedup (only for large arrays)

---

## Expected Impact on AudioUnit Benchmark

### Current Performance
- **C:** 65.34 µs/op
- **Go:** 37.4 µs/op
- **Gap:** 1.75x slower

### Estimated Improvement (Conservative)
Assuming AudioUnit schema characteristics:
- 30% primitives/fixed structs → **8x speedup** → save ~18 µs
- 20% arrays → **2x speedup** → save ~7 µs
- 50% strings (limited) → **1.2x speedup** → save ~5 µs

**Estimated optimized performance:** 65.34 - 30 = **~35 µs/op**

**New gap:** 35 / 37.4 = **0.94x** (faster than Go!)

**Conclusion:** With full optimization, C encoder should match or exceed Go performance.

---

## Benchmark Methodology Lessons

### ❌ Invalid Approach (Our Initial Mistake)
```c
// Benchmark measures BOTH data creation AND encoding
for (int i = 0; i < iterations; i++) {
    // Create test data
    Plugin plugin = { generate_name(), ... };  // WRONG!
    
    // Encode
    encode_plugin(&plugin);
}
```

### ✅ Correct Approach (Matching Go)
```c
// Create test data ONCE (outside measurement)
Plugin plugin = { "Reverb", 6, ... };

// Benchmark measures ONLY encoding
clock_gettime(&start);
for (int i = 0; i < iterations; i++) {
    encode_plugin(&plugin);  // ONLY THIS
}
clock_gettime(&end);
```

**Key Insight:** Separate data preparation (once) from encoding (measured loop). This is how Go benchmarks work (data in structs before `b.ResetTimer()`).

---

## Compiler Optimization Traps

### Problem: Const Data Gets Optimized Away
```c
// Compiler sees const string, optimizes strlen away
static const char* TEST_STRING = "Input 1 Gain";
size_t len = strlen(TEST_STRING);  // Optimized to constant!
```

### Solution: Use Volatile
```c
// Forces runtime string handling
static char test_buffer[64] = "Input 1 Gain";
static volatile char* TEST_STRING = test_buffer;
size_t len = strlen((char*)TEST_STRING);  // Real strlen call
```

**Lesson:** Benchmarks must use volatile data to prevent compiler from optimizing based on const test values. Real-world data is runtime-variable.

---

## Next Steps

1. ✅ **Research complete** - All findings documented
2. ⏳ **Update C_API_SPECIFICATION.md** with optimization patterns
3. ⏳ **Implement optimized C generator** with wire structs + bulk copy + inline nested
4. ⏳ **Validate with AudioUnit benchmark** (expect ~35 µs, 0.94x vs Go)
5. ⏳ **Remove c_micro_bench/** after extracting all learnings
6. ⏳ **Document in README.md** - C encoder performance characteristics

---

## Files Generated

### Micro-Benchmarks (Isolated Techniques)
- `0_writer_vs_buffer.c` - Dynamic writer vs pre-sized buffer
- `1_layout_optimization.c` - Field-by-field vs bulk vs direct offsets
- `2_string_methods.c` - snprintf vs strlen vs pre-computed
- `3_array_optimization.c` - Loop vs bulk memcpy for arrays
- `4_struct_optimization.c` - Recursive vs flattened vs wire struct

### Schema Benchmarks (Full Implementations)
- `bench_primitives.c` - AllPrimitives (12 fields) → 10.9x speedup
- `bench_arrays.c` - ArraysOfPrimitives (5 arrays) → 2.1x speedup
- `bench_nested.c` - Scene → Rectangle → Point → 5.3x speedup
- `bench_complex.c` - AudioDevice → Plugin → Parameter → 1.2x speedup

### Real-World Benchmark
- `testdata/audiounit/c/generate_test_data.go` - Convert plugins.json to C structs
- `testdata/audiounit/c/bench_real_data.c` - Proper apples-to-apples benchmark (65.34 µs)

---

## Conclusion

The **1.75x performance gap** can be closed with proven optimization techniques:

1. **Wire format structs** (8-13x speedup for fixed layouts)
2. **Bulk array copy** (5x speedup for primitive arrays)
3. **Pre-computed string lengths** (9x speedup vs strlen)
4. **Inline nested encoding** (1.4-5x speedup, avoid function calls)

**Expected result:** C encoder should **match or exceed Go performance** (~35 µs vs 37.4 µs) with full optimization.

**Key insight:** Schema knowledge enables compile-time optimizations even with runtime-variable data. The generator can produce specialized code for each schema's unique characteristics.

**Trade-off:** More complex codegen (wire structs, inline expansion) vs significant performance gains. Worth it.
