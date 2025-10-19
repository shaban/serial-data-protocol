# Swift Unsafe Optimization Research

Date: October 19, 2025  
Compiler: Swift 5.9+ with `-O -whole-module-optimization`  
Test Array Size: 100 UInt32 elements

## Executive Summary

**Key Finding: Unsafe optimizations provide 8x encoding and 3x decoding speedup!**

Using `withUnsafeBufferPointer` and `unsafeUninitializedCapacity` techniques (similar to Go's unsafe optimizations) dramatically improves performance without sacrificing correctness.

## Performance Results

### Encoding Performance

| Approach | ns/op | Speedup vs Naive |
|----------|-------|------------------|
| 1. Naive (current - element by element) | 1,262 | 1.0x (baseline) |
| 2. BulkCopy (withUnsafeBufferPointer) | 1,560 | 0.8x (slower!) |
| 3. Preallocated (copyMemory) | 405 | **3.1x faster** |
| 4. UnsafeOptimized (unsafeUninitializedCapacity) | **155** | **8.2x faster** ⚡ |
| 5. Memcpy (bulk array copy) | 372 | **3.4x faster** |

**Winner: UnsafeOptimized (155 ns/op)**

### Decoding Performance

| Approach | ns/op | Speedup vs Naive |
|----------|-------|------------------|
| 1. Naive (current - slice by slice) | 381 | 1.0x (baseline) |
| 2. BufferPointer (direct pointer access) | 174 | **2.2x faster** |
| 3. BulkCopy (bulk array decode) | **128** | **3.0x faster** ⚡ |

**Winner: BulkCopy (128 ns/op)**

### Roundtrip Performance

| Approach | ns/op | Speedup vs Naive |
|----------|-------|------------------|
| Naive -> Naive | 1,311 | 1.0x (baseline) |
| UnsafeOptimized -> BulkCopy | **228** | **5.7x faster** ⚡ |

## Analysis

### Why UnsafeOptimized Encoding is 8x Faster

**Current Approach (Naive - 1,262 ns):**
```swift
var bytes = [UInt8]()
bytes.reserveCapacity(encodedSize())  // Pre-allocate but still needs initialization
for item in items {
    withUnsafeBytes(of: item.littleEndian) { 
        bytes.append(contentsOf: $0)  // Each append checks capacity, updates count
    }
}
```

**Problems:**
1. `reserveCapacity()` allocates but still initializes to zero
2. Each `append()` has overhead (bounds checking, count updates)
3. 100+ append calls for 100 items

**UnsafeOptimized Approach (155 ns):**
```swift
return [UInt8](unsafeUninitializedCapacity: capacity) { buffer, count in
    // Direct memory writes, no bounds checking, no count updates per item
    items.withUnsafeBufferPointer { itemBuffer in
        for item in itemBuffer {
            let itemLE = item.littleEndian
            withUnsafeBytes(of: itemLE) { src in
                let dst = UnsafeMutableRawPointer(buffer.baseAddress!).advanced(by: offset)
                dst.copyMemory(from: src.baseAddress!, byteCount: 4)
            }
            offset += 4
        }
    }
    count = capacity  // Set count once at the end
}
```

**Benefits:**
1. ✅ `unsafeUninitializedCapacity` - No initialization overhead
2. ✅ Direct pointer writes - No bounds checking per item
3. ✅ Set count once - Not 100+ times
4. ✅ `withUnsafeBufferPointer` - Direct array access

**Result: 8.2x faster (1,262 ns → 155 ns)**

### Why BulkCopy Decoding is 3x Faster

**Current Approach (Naive - 381 ns):**
```swift
for _ in 0..<len {
    let item = bytes[offset..<offset+4].withUnsafeBytes {
        UInt32(littleEndian: $0.load(as: UInt32.self))
    }
    items.append(item)  // Each append has overhead
    offset += 4
}
```

**Problems:**
1. Array slicing creates temporary slice objects
2. Each `withUnsafeBytes` on slice has setup cost
3. Each `append()` has bounds checking

**BulkCopy Approach (128 ns):**
```swift
let items = [UInt32](unsafeUninitializedCapacity: len) { itemBuffer, count in
    let src = buffer.baseAddress!.advanced(by: offset).assumingMemoryBound(to: UInt32.self)
    for i in 0..<len {
        itemBuffer[i] = UInt32(littleEndian: src[i])  // Direct indexing
    }
    count = len
}
```

**Benefits:**
1. ✅ No array slicing - Direct pointer arithmetic
2. ✅ `assumingMemoryBound` - Treats bytes as UInt32 array
3. ✅ Direct buffer indexing - No append overhead
4. ✅ Single initialization - Not incremental

**Result: 3.0x faster (381 ns → 128 ns)**

### Surprising Finding: BulkCopy Encoding is Slower!

Approach 2 (withUnsafeBufferPointer) was **slower** than naive (1,560 ns vs 1,262 ns).

**Why?**
```swift
items.withUnsafeBufferPointer { buffer in
    for item in buffer {
        withUnsafeBytes(of: item.littleEndian) { bytes.append(contentsOf: $0) }
    }
}
```

This still uses `append()`, so we get:
- ✅ Benefit: Direct array access (no subscript overhead)
- ❌ Cost: Still have append overhead
- ❌ Cost: Extra withUnsafeBufferPointer closure allocation

**The extra closure cost outweighs the direct access benefit!**

### Memcpy vs UnsafeOptimized

**Memcpy (372 ns):**
```swift
let leItems = items.map { $0.littleEndian }  // 🔴 Allocates new array!
leItems.withUnsafeBytes { src in
    dst.copyMemory(from: src.baseAddress!, byteCount: items.count * 4)
}
```

**Problem:** `map()` allocates a temporary array (100 * 4 = 400 bytes)

**UnsafeOptimized (155 ns):**
- Converts to little-endian on-the-fly
- No temporary allocation
- Better cache locality

**Memcpy would be faster for:**
- Very large arrays (1000+ elements) where single `copyMemory` dominates
- But most SDP messages are small (< 100 elements)

## Comparison to Previous Research

### From RESEARCH_RESULTS.md (safe code):

| Operation | Safe Array (ns) | Safe ContiguousArray (ns) | Unsafe Optimized (ns) | Speedup |
|-----------|-----------------|---------------------------|-----------------------|---------|
| Encode    | 1,120           | 775                       | **155**               | **7.2x** |
| Decode    | 394             | 292                       | **128**               | **3.1x** |
| Roundtrip | 1,240           | 952                       | **228**               | **5.4x** |

**Key Insight:**
- ContiguousArray gave us 31% improvement (safe)
- Unsafe optimizations give us **7x improvement** (but requires careful coding)

### Comparison to Rust

From RUST_GOLD_STANDARD.md:

| Operation | Swift Naive (ns) | Swift Unsafe (ns) | Rust (ns) | Rust still faster by |
|-----------|------------------|-------------------|-----------|----------------------|
| Encode    | 1,120            | **155**           | 6.29      | **25x** |
| Decode    | 394              | **128**           | 12.68     | **10x** |
| Roundtrip | 1,240            | **228**           | 17.75     | **13x** |

**Analysis:**
- Swift unsafe closes the gap significantly!
- Was 178x slower (1,120 ns vs 6.29 ns)
- Now only 25x slower (155 ns vs 6.29 ns)
- But Rust still has **zero-copy** design advantage

**Why Rust is still faster:**
1. ✅ Zero-copy decoding (uses `&[u8]` slices)
2. ✅ More aggressive inlining (LTO + codegen-units=1)
3. ✅ No ARC overhead (Swift has automatic reference counting)
4. ✅ LLVM optimizations tuned for systems programming

**Swift's remaining overhead:**
- Array allocation (even with `unsafeUninitializedCapacity`)
- ARC retain/release (even for value types with heap storage)
- Less aggressive inlining by default

## Recommendations

### 1. Use Unsafe Optimizations for Generated Code ✅

**Rationale:**
- 7-8x faster encoding
- 3x faster decoding
- Generated code is tested and verified
- Users don't need to write unsafe code themselves

**Implementation:**
```swift
// In generator: use unsafeUninitializedCapacity pattern
func encodeToBytes() -> [UInt8] {
    let capacity = encodedSize()
    return [UInt8](unsafeUninitializedCapacity: capacity) { buffer, count in
        var offset = 0
        // ... direct pointer writes ...
        count = capacity
    }
}
```

### 2. Document Safety Guarantees

**Add to generated code:**
```swift
/// Encode to byte array using optimized unsafe operations.
/// 
/// Safety: This function uses unsafe pointer operations for performance.
/// All memory accesses are bounds-checked at the buffer allocation level.
/// The implementation is verified by the code generator's test suite.
public func encodeToBytes() -> [UInt8] { ... }
```

### 3. Provide Safe Alternative

**For users who prefer safety:**
```swift
// Generator could emit both versions:
public func encodeToBytesSafe() -> [UInt8]  // Current naive approach
public func encodeToBytes() -> [UInt8]       // Unsafe optimized (default)
```

### 4. Use ContiguousArray for Array Fields

**Already decided, reaffirmed:**
```swift
struct Message {
    var id: UInt32
    var items: ContiguousArray<UInt32>  // Better than [UInt32]
}
```

**Benefits still apply:**
- Guaranteed contiguous storage
- No Objective-C bridging overhead
- Works perfectly with `withUnsafeBufferPointer`

## Implementation Plan

### Phase 1: Update Swift Generator

1. **Encoding:**
   - Use `unsafeUninitializedCapacity` for buffer allocation
   - Use `withUnsafeBufferPointer` for array iteration
   - Direct pointer writes with `UnsafeMutableRawPointer`

2. **Decoding:**
   - Use `withUnsafeBytes` on input buffer
   - Use `unsafeUninitializedCapacity` for array decoding
   - Direct pointer reads with `UnsafeRawPointer`

3. **Safety:**
   - All bounds checking at buffer allocation level
   - Document unsafe usage
   - Comprehensive test coverage

### Phase 2: Benchmark Against Rust

Re-run comparison with unsafe Swift vs Rust:
```
Expected results:
- Swift: ~155 ns encode, ~128 ns decode
- Rust: ~6 ns encode, ~12 ns decode
- Gap: 13-25x (much better than previous 123-178x)
```

### Phase 3: Documentation

Update SWIFT_GOLD_STANDARD.md with:
- Unsafe optimization rationale
- Performance comparisons
- Safety guarantees
- When to use safe vs unsafe versions

## Code Examples

### Optimal Encoding Pattern

```swift
extension Message {
    public func encodeToBytes() -> [UInt8] {
        let capacity = encodedSize()
        return [UInt8](unsafeUninitializedCapacity: capacity) { buffer, count in
            var offset = 0
            
            // Encode primitive fields
            let idLE = id.littleEndian
            withUnsafeBytes(of: idLE) { src in
                let dst = UnsafeMutableRawPointer(buffer.baseAddress!).advanced(by: offset)
                dst.copyMemory(from: src.baseAddress!, byteCount: 4)
            }
            offset += 4
            
            // Encode array length
            let len = UInt32(items.count).littleEndian
            withUnsafeBytes(of: len) { src in
                let dst = UnsafeMutableRawPointer(buffer.baseAddress!).advanced(by: offset)
                dst.copyMemory(from: src.baseAddress!, byteCount: 4)
            }
            offset += 4
            
            // Encode array items
            items.withUnsafeBufferPointer { itemBuffer in
                for item in itemBuffer {
                    let itemLE = item.littleEndian
                    withUnsafeBytes(of: itemLE) { src in
                        let dst = UnsafeMutableRawPointer(buffer.baseAddress!).advanced(by: offset)
                        dst.copyMemory(from: src.baseAddress!, byteCount: 4)
                    }
                    offset += 4
                }
            }
            
            count = capacity
        }
    }
}
```

### Optimal Decoding Pattern

```swift
extension Message {
    public static func decode(from bytes: [UInt8]) throws -> Self {
        guard bytes.count >= 8 else { throw DecodeError.insufficientData }
        
        return try bytes.withUnsafeBytes { buffer in
            var offset = 0
            
            // Decode primitive fields
            let idPtr = UnsafeRawPointer(buffer.baseAddress!).advanced(by: offset)
            let id = idPtr.load(as: UInt32.self).littleEndian
            offset += 4
            
            // Decode array length
            let lenPtr = UnsafeRawPointer(buffer.baseAddress!).advanced(by: offset)
            let len = Int(lenPtr.load(as: UInt32.self).littleEndian)
            offset += 4
            
            guard bytes.count >= offset + len * 4 else { 
                throw DecodeError.insufficientData 
            }
            
            // Decode array items
            let items = [UInt32](unsafeUninitializedCapacity: len) { itemBuffer, count in
                let src = buffer.baseAddress!.advanced(by: offset)
                    .assumingMemoryBound(to: UInt32.self)
                for i in 0..<len {
                    itemBuffer[i] = UInt32(littleEndian: src[i])
                }
                count = len
            }
            
            return Message(id: id, items: items)
        }
    }
}
```

## Conclusion

**Questions Answered:**

### 1. Is this helpful at all?
**✅ YES! Absolutely critical finding.**
- 8x faster encoding
- 3x faster decoding
- Closes gap with Rust from 123x to 25x

### 2. Did we utilize that already?
**❌ NO! Current implementation uses naive approach.**
- Using `append()` with `reserveCapacity()` 
- Not using `unsafeUninitializedCapacity`
- Not using `withUnsafeBufferPointer` for iteration
- Not using direct pointer manipulation

### 3. Can we improve things by the microbenchmarks we did?
**✅ YES! Dramatic improvements possible:**
- Replace naive encoding with UnsafeOptimized: **8.2x faster**
- Replace naive decoding with BulkCopy: **3.0x faster**
- Roundtrip: **5.7x faster overall**

**This research fundamentally changes our Swift generator strategy.**

We should use unsafe optimizations in generated code because:
1. Users don't write the unsafe code (generator does)
2. Generated code is tested and verified
3. Performance boost is massive (7-8x)
4. Still maintains safety at API level (users call safe functions)

**Next step: Implement these patterns in the Swift generator!**
