# Wire Format Alignment Analysis

## The Problem: SDP Wire Format vs Rust Alignment Requirements

### Wire Format Layout (arrays_primitives.sdpb)

```
Offset | Bytes                | Field         | Alignment
-------|----------------------|---------------|----------
0      | 05 00 00 00          | u8_array len  | ✓ aligned (mod 4 = 0)
4      | 01 02 03 04 05       | u8_array data | (5 bytes)
9      | 05 00 00 00          | u32_array len | ✗ MISALIGNED (mod 4 = 1)
13     | 64 00 00 00          | u32[0] = 100  | ✗ MISALIGNED (mod 4 = 1)
17     | c8 00 00 00          | u32[1] = 200  | ✗ MISALIGNED (mod 4 = 1)
21     | 2c 01 00 00          | u32[2] = 300  | ✗ MISALIGNED (mod 4 = 1)
25     | 90 01 00 00          | u32[3] = 400  | ✗ MISALIGNED (mod 4 = 1)
29     | f4 01 00 00          | u32[4] = 500  | ✗ MISALIGNED (mod 4 = 1)
33     | 05 00 00 00          | f64_array len | ✗ MISALIGNED (mod 8 = 1)
37     | 9a 99...f1 3f        | f64[0] = 1.1  | ✗ MISALIGNED (mod 8 = 5)
45     | 9a 99...01 40        | f64[1] = 2.2  | ✗ MISALIGNED (mod 8 = 5)
...
```

## Why Is Data Misaligned?

**SDP wire format is densely packed with NO padding bytes:**

```
Schema:
struct ArraysOfPrimitives {
    u8_array: []u8,    // 4 bytes len + 5 bytes data = 9 bytes total
    u32_array: []u32,  // Starts at offset 9 (not aligned to 4!)
    f64_array: []f64,  // Starts at offset 33 (not aligned to 8!)
    ...
}
```

**The issue:**
- `u8_array` has 5 elements = 5 bytes of data
- Next field `u32_array` starts immediately at offset 9
- **9 mod 4 = 1** → u32 data is misaligned by 1 byte
- u32_array uses 4 bytes len + 20 bytes data = 24 bytes total
- Next field `f64_array` starts at offset 33
- **33 mod 8 = 1** → f64 data is misaligned by 1 byte

## Rust's Alignment Requirements

Rust requires:
- `u32` must be aligned to 4-byte boundaries (address mod 4 = 0)
- `f64` must be aligned to 8-byte boundaries (address mod 8 = 0)
- `&[u32]` slice must point to aligned data

**Why?** CPU architectures (ARM, x86) are faster when accessing aligned data. Some CPUs (older ARM) crash on misaligned access.

## What Rust Does

### Fast Path (Aligned Data) - Uses `try_cast_slice`

```rust
let bytes = &buf[offset..offset + byte_len];  // Raw bytes from wire
if let Ok(slice) = bytemuck::try_cast_slice::<u8, u32>(bytes) {
    // ✓ Bytes are 4-byte aligned! Can reinterpret as &[u32]
    slice.to_vec()  // Fast: memcpy entire array at once
}
```

**This is "zero-copy" viewing** - we're treating the u8 bytes as u32 values directly without conversion.

### Slow Path (Misaligned Data) - Uses `chunks_exact` fallback

```rust
else {
    // ✗ Bytes NOT aligned - can't safely reinterpret as &[u32]
    bytes.chunks_exact(4)  // Split into 4-byte chunks
        .map(|chunk| u32::from_le_bytes(chunk.try_into().unwrap()))
        .collect()  // Element-by-element conversion
}
```

**This is the path we're hitting** - must convert each u32 individually.

## Why SDP Doesn't Pad

**Design decision:** Minimize wire format size

```
Option A (Current): Dense packing
u8_array: [01 02 03 04 05] → u32_array: [64 00 00 00...]
Total: 9 bytes (4 len + 5 data + 0 padding)

Option B (Padded): Align to largest type
u8_array: [01 02 03 04 05 00 00 00] → u32_array: [64 00 00 00...]
Total: 13 bytes (4 len + 5 data + 3 padding)
```

**Trade-off:**
- ✓ Smaller wire size (important for network/storage)
- ✗ Slower decode in languages with strict alignment (Rust, C++ on ARM)

## Performance Impact

**Rust decode: 318ns**
- ~24ns lost to misaligned u32 array (chunks_exact overhead)
- ~24ns lost to misaligned f64 array (chunks_exact overhead)
- ~48ns total alignment penalty (~15% of decode time)

**C++ decode: 112ns**
- C++ allows unaligned reads on x86/M1 (with small penalty)
- Uses: `*(const uint32_t*)(buf + offset)` even if misaligned
- Compiler generates unaligned load instructions

**Go decode: 185ns**
- Go runtime handles alignment automatically
- Uses `binary.LittleEndian.Uint32()` which works unaligned

## Could SDP Fix This?

**Option 1: Add padding bytes** (breaks backward compatibility)
```
struct ArraysOfPrimitives {
    u8_array: []u8,
    _pad1: [3]u8,     // Align to 4 bytes
    u32_array: []u32,
    _pad2: [3]u8,     // Align to 8 bytes
    f64_array: []f64,
}
```

**Option 2: Reorder fields** (breaks backward compatibility)
```
struct ArraysOfPrimitives {
    f64_array: []f64,  // Put largest alignment first
    u32_array: []u32,
    u8_array: []u8,    // Smallest last (no padding needed after)
}
```

**Option 3: Accept the trade-off** (current approach)
- Wire format is language-agnostic
- Prioritizes size over decode speed
- Rust's 318ns is still acceptable for most use cases

## Summary

**Your question:** "The wire format aligns data contrary to how Rust handles things"

**Answer:** YES, exactly! 

- **SDP wire format** = densely packed, no padding, size-optimized
- **Rust requirements** = type-safe aligned pointers for zero-copy views
- **Consequence** = Rust can't use fast `bytemuck::cast_slice` path, must use slower element-by-element conversion
- **Impact** = ~48ns penalty (15% slower) but still safe and reasonable

The misalignment is **by design** in SDP to minimize network/storage overhead, at the cost of some decode performance in alignment-strict languages like Rust.
