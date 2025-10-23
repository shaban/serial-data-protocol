# Arrays Bulk Optimization: 3-Way Benchmark Comparison

**Test Date:** October 23, 2025  
**Machine:** Apple M1 Pro, macOS  
**Schema:** `arrays.sdp` (ArraysOfPrimitives)  
**Data:** `testdata/binaries/arrays_primitives.sdpb` (141 bytes, 5-element arrays)  
**Optimization:** Bulk copy for primitive integer arrays (u8, u32)

## Benchmark Results

### Encode Performance (struct → binary)

| Language | Time (ns/op) | Relative | Allocations | Notes |
|----------|--------------|----------|-------------|-------|
| **Rust** | **41 ns**    | 1.0×     | 0 (pre-allocated) | Buffer-based API |
| **C++**  | **39 ns**    | 0.95×    | 0 (pre-allocated) | Encode only (size pre-calculated) |
| **C++*** | **71 ns**    | 1.73×    | 0 (pre-allocated) | **Realistic: size() + encode()** |
| **Go**   | **72 ns**    | 1.76×    | 144 B (1 alloc) | Allocates internally |

**Winner:** Rust (41 ns) - Fastest with pre-allocated buffer

**Important C++ Note:** The C++ `encode()` function requires a buffer of the correct size. In practice, you must call `_size()` first (~32 ns), making the realistic total **71 ns** - essentially tied with Go.

### Decode Performance (binary → struct)

| Language | Time (ns/op) | Relative | Allocations |
|----------|--------------|----------|-------------|
| **C++**  | **152 ns**   | 1.0×     | ~5 vectors |
| **Go**   | **182 ns**   | 1.20×    | 200 B (10 allocs) |
| **Rust** | **329 ns**   | 2.16×    | ~5 vectors + alignment check |

**Winner:** C++ (152 ns) - 17% faster than Go, 54% faster than Rust

### Roundtrip Performance (encode + decode)

| Language | Time (ns/op) | Relative |
|----------|--------------|----------|
| **C++**  | **156 ns**   | 1.0×     |
| **Go**   | **263 ns**   | 1.69×    |
| **Rust** | **370 ns**   | 2.37×    |

**Winner:** C++ (156 ns) - 41% faster than Go, 58% faster than Rust

## Analysis

### C++ Two-Step API Consideration

**C++ requires a two-step process:**
1. `size_t size = arrays_of_primitives_size(msg)` - Calculate encoded size (~32 ns)
2. `arrays_of_primitives_encode(msg, buffer)` - Encode to pre-allocated buffer (~39 ns)

**Total realistic encode time: 71 ns**

Our initial benchmark measured only step 2 (39 ns) because the size was pre-calculated outside the timing loop. This is valid for **repeated encodes of the same schema** (where size is constant), but for **general use**, you need both steps.

**Comparison:**
- **C++ (realistic):** 71 ns = size() + encode()
- **Go:** 72 ns = single call that allocates
- **Rust:** 41 ns = single call with pre-allocated buffer

### Rust Performance Advantages (Encode)

**Rust is fastest for realistic single-message encoding:**
- **Rust:** 41 ns/op (pre-allocated buffer, single-step API)
- **C++ (realistic):** 71 ns/op (two-step: 32 ns size + 39 ns encode)
- **Go:** 72 ns/op (single-step with allocation)

**When C++ encode-only (39 ns) matters:**
- Repeated encodes of same schema (size calculated once, reused)
- Tight loop scenarios where size is constant
- Amortized cost over many operations

For general-purpose encoding (one-off messages), Rust's single-step API is both fastest and most ergonomic.

### C++ Performance Advantages (Decode)

**C++ Wins Decoding:**
- **C++:** 152 ns/op (fastest, zero-copy pointer arithmetic)
- **Go:** 182 ns/op (20% slower, allocates intermediate slices)
- **Rust:** 329 ns/op (2.2× slower, alignment checks + fallback)

**Why C++ is fastest:**
1. **Zero-copy decode:** Direct pointer arithmetic, no intermediate buffers
2. **Pre-allocated buffers:** Size calculated upfront, single allocation
3. **Compiler optimizations:** LLVM/GCC aggressive inlining and vectorization
4. **No bounds checking:** Assumes valid input (validated once at top level)

### Rust Performance Notes
1. **Alignment checks:** `try_cast_slice` adds overhead for misaligned data
2. **Fallback path:** When misaligned, uses `chunks_exact` + `from_le_bytes` (slower)
3. **Safety first:** Bounds checking and alignment validation on every operation
4. **Still competitive:** Only 5% slower than C++ on encode, reasonable decode

### Go Performance Notes
1. **Allocation overhead:** Returns `[]byte` slice, requires allocation
2. **Interface boundaries:** `io.Writer`/`io.Reader` add indirection
3. **Simpler codegen:** Less aggressive bulk optimizations than C++
4. **Good enough:** 72 ns encode is fast for most use cases

## Bulk Optimization Impact

All three languages benefit from bulk copy optimization:

**Before optimization (element-by-element loops):**
- Go: ~180 ns encode, ~425 ns decode (estimated)
- Rust: ~120 ns encode, ~750 ns decode (estimated)
- C++: ~85 ns encode (encode-only, not including size), ~350 ns decode (estimated)

**After optimization (bulk copy):**
- **Go:** 2.5× faster encode, 2.3× faster decode
- **Rust:** 2.9× faster encode, 2.3× faster decode
- **C++:** 2.2× faster encode (encode-only), 2.3× faster decode

**Note:** C++ estimates exclude size calculation overhead for consistency with encode-only measurement.

## When to Use Each Language

**Use Rust** when:
- Single-message encoding performance is critical (fastest at 41 ns)
- Safety and correctness are paramount
- You need memory safety guarantees
- Cross-platform deployment with strong typing
- Balance of performance and ergonomics matters

**Use C++** when:
- Decode performance is critical (fastest at 152 ns)
- You're encoding same schema repeatedly (39 ns encode-only amortizes well)
- Zero-copy semantics are essential
- Binary size matters (no runtime)
- You control both encoder and decoder

**Use Go** when:
- Development speed matters more than raw performance
- You need simple, readable code
- 72 ns encode and 182 ns decode are "fast enough"
- You want easy integration with Go services
- Rapid iteration and testing are priorities

## Reproducibility

All benchmarks use the same canonical test data:
- **Encode:** Uses struct decoded from `arrays_primitives.sdpb`
- **Decode:** Uses `arrays_primitives.sdpb` binary directly

Run benchmarks:
```bash
cd benchmarks
make bench-go-arrays     # Go
make bench-cpp-arrays    # C++
make bench-rust-byte     # Rust
```

## Methodology

- **Iterations:** 10,000 (encode/decode), 5,000 (roundtrip)
- **Warm-up:** 100 iterations before timing
- **Data:** 5-element arrays (u8[], u32[], f64[], str[], bool[])
- **Measurement:** High-resolution timers (nanosecond precision)
- **Allocation tracking:** Go only (via `b.ReportAllocs()`)

## Conclusion

**Rust is fastest for single-message encoding, C++ wins decode, Go is simplest:**

**Encode Performance (realistic, one-off messages):**
- **Rust:** 41 ns (baseline - fastest)
- **C++ (realistic):** 71 ns (+73%) - includes required size calculation
- **C++ (encode-only):** 39 ns - amortizable in tight loops
- **Go:** 72 ns (+76%)

**Decode Performance:**
- **C++:** 152 ns (baseline - fastest)
- **Go:** 182 ns (+20%)
- **Rust:** 329 ns (+116%)

**Key Insights:**
1. **Rust** has the fastest single-step encode API (41 ns)
2. **C++** requires two-step process for general use (size + encode = 71 ns)
3. **C++** wins decode convincingly (152 ns vs 182 ns Go, 329 ns Rust)
4. **Go** is essentially tied with C++ for encoding (72 ns vs 71 ns)
5. All three benefit from 2-3× bulk array optimization

**Recommendation:** 
- Use **Rust** for general-purpose encoding with safety
- Use **C++** for repeated encoding (amortize size calc) or decode-heavy workloads
- Use **Go** for rapid development with good-enough performance

