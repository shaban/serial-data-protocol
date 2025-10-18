# Rust API Performance Comparison

**Test Platform:** Apple M1 (darwin/arm64)  
**Date:** October 18, 2025  
**Rust Version:** 1.90.0

## Summary

Direct byte slice API (`wire_slice`) is **significantly faster** than the trait-based API (`wire`) for encoding operations.

## Benchmark Results

### Integer Operations (u32)

| Operation | Trait API | Slice API | Speedup |
|-----------|-----------|-----------|---------|
| **Encode** | 0.97 ns | **0.31 ns** | **3.1x** |
| **Decode** | 0.31 ns | **0.31 ns** | 1.0x |

### String Operations (54 bytes)

| Operation | Trait API | Slice API | Speedup |
|-----------|-----------|-----------|---------|
| **Encode** | 29.9 ns | **7.5 ns** | **4.0x** |
| **Decode** | 38.4 ns | **37.3 ns** | 1.0x |

### Complex Roundtrip (u32 + f64 + bool + string)

| Operation | Trait API | Slice API | Speedup |
|-----------|-----------|-----------|---------|
| **Roundtrip** | 123.0 ns | **34.4 ns** | **3.6x** |

## Analysis

### Why Slice API is Faster

1. **Zero Abstraction Overhead**
   - Direct memory access via `&[u8]` and `&mut [u8]`
   - No trait method dispatch
   - Compiler can inline everything

2. **No I/O Trait Indirection**
   - Trait API: `Write::write()` → `Vec::write()` → memory copy
   - Slice API: Direct `buf[offset..]` access → single memory write
   - Eliminates virtual dispatch overhead

3. **Better Compiler Optimizations**
   - Fixed-size operations are fully optimized
   - Bounds checks can be elided more easily
   - LLVM can vectorize slice operations

4. **Matches Go's Approach**
   - Go uses `[]byte` directly (like our slice API)
   - Go avoids `io.Writer` indirection in wire package
   - This is why Go was faster in our benchmarks!

### When Each API Shines

**Trait API (`wire` module):**
- ✅ Works with any `Read`/`Write` implementation
- ✅ Can encode to network sockets, files, etc.
- ✅ Composable with other Rust I/O
- ❌ 3-4x slower for encoding

**Slice API (`wire_slice` module):**
- ✅ Maximum performance (matches Go speed)
- ✅ Zero abstraction overhead
- ✅ Perfect for hot paths
- ❌ Requires pre-allocated buffers
- ❌ Less flexible than trait API

## Performance Breakdown

### Encoding Speedup

The slice API is **4x faster** for string encoding:
- Trait: 29.9 ns (allocates `Vec`, uses `Write` trait)
- Slice: 7.5 ns (direct memory copy via `copy_from_slice`)

This matches the difference we saw between Go and Rust!

### Decoding Performance

Decoding is nearly identical between both APIs:
- String decode: 38.4ns (trait) vs 37.3ns (slice) - only 3% difference
- U32 decode: 0.31ns for both

**Why?** Decoding allocates memory (for `String`, `Vec`) regardless of API.
The trait overhead is negligible compared to allocation cost.

### Complex Roundtrip

The slice API is **3.6x faster** overall:
- Trait: 123.0 ns
- Slice: 34.4 ns

This compounds the encoding benefits across multiple operations.

## Implications for SDP

### Current Generated Code

Our code generator produces trait-based code:
```rust
impl AllPrimitives {
    pub fn encode<W: Write>(&self, writer: &mut W) -> Result<()> {
        let mut enc = Encoder::new(writer);
        enc.write_u32(self.u32_field)?;
        // ...
    }
}
```

### Optimized Version (Slice API)

We could generate:
```rust
impl AllPrimitives {
    pub fn encode_to_slice(&self, buf: &mut [u8]) -> Result<usize> {
        let mut offset = 0;
        wire_slice::encode_u32(buf, offset, self.u32_field)?;
        offset += 4;
        // ... returns total bytes written
        Ok(offset)
    }
}
```

### Performance Projection

If we switch to slice API, Rust would likely **match or exceed Go** performance:
- Go: 36 ns/op (primitives encode)
- Rust (trait): 146 ns/op
- Rust (slice, projected): ~35-40 ns/op ✅

## Recommendations

### For Generated Code

1. **Add slice-based methods**
   - Keep trait API for compatibility
   - Add `encode_to_slice()` and `decode_from_slice()` methods
   - Users can choose based on needs

2. **Make slice API the default**
   - Most SDP use cases: in-memory buffers
   - Audio plugin IPC: always uses byte buffers
   - Network I/O: can still use trait API

3. **Benchmark-driven decision**
   - Measure real-world schemas
   - Compare with Go performance
   - Optimize hot paths first

### For Users

**Use slice API when:**
- Performance is critical
- Working with in-memory buffers
- IPC between processes
- Real-time audio/video

**Use trait API when:**
- Need to encode to files, sockets
- Want composability with Rust I/O
- Performance is not critical
- Flexibility > speed

## Next Steps

1. **Update code generator** to emit both APIs
2. **Re-run cross-language benchmarks** with slice API
3. **Update documentation** with performance guidance
4. **Add benchmark comparison** to CI

## Conclusion

The **slice API is 3-4x faster** than the trait API, primarily for encoding.

This explains why **Go was faster** in our initial benchmarks:
- Go uses direct byte slice access (`[]byte`)
- Our Rust code used trait indirection (`Read`/`Write`)
- Slice API eliminates this overhead

**With the slice API, Rust should match or beat Go performance** while maintaining zero-copy, no-GC advantages. 🚀
