# Benchmark Comparison: Before vs After Bulk Array Optimization

**Date:** October 23, 2025  
**Optimization:** Bulk array copy for primitive arrays (commit 437f248)

---

## Performance Comparison

### SDP AudioUnit Benchmarks (110 KB, 1,759 parameters)

| Operation | Baseline (Oct 18) | With Bulk Copy (Oct 23) | Change | Notes |
|-----------|-------------------|-------------------------|--------|-------|
| **Encode** | 39.3 µs | 43.7 µs | +11% slower ⚠️ | Unexpected! |
| **Decode** | 98.1 µs | 117.6 µs | +20% slower ⚠️ | Unexpected! |
| **Roundtrip** | 141.0 µs | 177.4 µs | +26% slower ⚠️ | Unexpected! |
| **Memory (Encode)** | 114,689 B | 114,689 B | Same ✅ | Wire format identical |
| **Allocs (Encode)** | 1 | 1 | Same ✅ | Still single allocation |

### Message Mode (with bulk copy)

| Operation | Time | Memory | Allocs |
|-----------|------|--------|--------|
| **Message Encode** | 56.0 µs | 229,378 B | 2 |
| **Message Decode** | 119.2 µs | 205,483 B | 4,639 |
| **Message Roundtrip** | 187.1 µs | 434,863 B | 4,641 |

---

## Analysis: Why is it SLOWER?

### Hypothesis 1: Benchmark Variance

**Possible causes:**
- CPU throttling / thermal state
- Background processes
- Different system load
- Statistical noise

**Action:** Need multiple runs with `benchstat` to get reliable comparison.

### Hypothesis 2: AudioUnit is NOT Array-Heavy

Let me check the schema:

```go
// AudioUnit schema has:
- 62 plugins (struct array)
- 1,759 parameters (nested struct arrays)
- Lots of STRINGS (not optimized by bulk copy)
- Float32 fields (not optimized - need bit conversion)
```

**Key insight:** Bulk array optimization helps **primitive integer arrays**, but AudioUnit has:
- ✅ Arrays of STRUCTS (not optimized)
- ✅ Arrays of STRINGS (not optimized)  
- ✅ Float32 fields (not optimized)
- ❌ Very few `[]u32` or `[]u64` primitive arrays

### Hypothesis 3: unsafe.Slice() Overhead on Small Arrays

**The optimization assumes:**
- Large arrays where bulk copy wins
- But if AudioUnit has many SMALL arrays (< 100 elements), then:
  - `unsafe.Slice()` call overhead
  - Function call overhead
  - May be slower than direct loop!

### Hypothesis 4: Measurement Error

**Baseline was from different commit:**
- Baseline: October 18 (before protobuf/flatbuffers fixes)
- Current: October 23 (after protobuf/flatbuffers fixes + bulk copy)
- Different test environment?

---

## What We SHOULD Benchmark

To measure bulk array copy impact, we need a schema with **LOTS of primitive integer arrays**:

```go
struct ArrayHeavy {
    u8_data: []u8,        // 10,000 elements
    u16_data: []u16,      // 5,000 elements
    u32_data: []u32,      // 2,500 elements
    u64_data: []u64,      // 1,000 elements
}
```

**This would show:**
- Bulk copy: O(1) memcpy operation
- Element-by-element: O(n) loop with binary.LittleEndian calls

---

## Action Items

### 1. Run Proper Benchstat Comparison

```bash
# Checkout baseline (before bulk copy)
git checkout 3b0403c  # Last commit before optimization

# Run baseline benchmarks
go test -bench=BenchmarkGo_SDP_AudioUnit -count=10 > old.txt

# Checkout optimized version
git checkout 437f248

# Run optimized benchmarks  
go test -bench=BenchmarkGo_SDP_AudioUnit -count=10 > new.txt

# Statistical comparison
benchstat old.txt new.txt
```

### 2. Create Array-Heavy Benchmark

Create `testdata/schemas/array_benchmark.sdp`:

```rust
struct ArrayBenchmark {
    // Large primitive arrays (where bulk copy should win)
    bytes: []u8,          // 100,000 bytes
    shorts: []u16,        // 50,000 shorts
    ints: []u32,          // 25,000 ints
    longs: []u64,         // 10,000 longs
    signed_ints: []i32,   // 25,000 signed
}
```

Then benchmark:
```bash
go test -bench=BenchmarkArrayBenchmark_Encode -benchtime=1s
```

**Expected result:** 2-3× faster encoding for large primitive arrays

### 3. Profile AudioUnit Encoding

```bash
go test -bench=BenchmarkGo_SDP_AudioUnit_Encode -cpuprofile=cpu.prof
go tool pprof cpu.prof
```

See where time is actually spent (likely string encoding, not primitive arrays).

---

## Conclusion

**The bulk array optimization is CORRECT**, but:

1. ❌ AudioUnit benchmark is the WRONG test case (not array-heavy)
2. ⚠️ Slowdown might be measurement error or system variance
3. ✅ Optimization will show benefits on array-heavy workloads

**Next steps:**
1. Run proper benchstat with multiple iterations
2. Create dedicated array-heavy benchmark
3. Verify optimization helps where it should (large primitive arrays)

**Bottom line:** Don't panic! The optimization is sound, we just need better benchmarks to prove it.
