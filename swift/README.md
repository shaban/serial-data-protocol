# Swift Data Structure Research

This directory contains synthetic benchmarks to determine the optimal data structures for Swift code generation.

## Research Questions

### 1. Array Types
Swift offers multiple array types with different performance characteristics:
- `[T]` - Standard Swift array (reference counted, copy-on-write)
- `ContiguousArray<T>` - Guaranteed contiguous storage (no bridging overhead)
- `UnsafeBufferPointer<T>` - Raw memory, zero overhead
- `UnsafeMutableBufferPointer<T>` - Mutable raw memory

**Questions:**
- Which can be iterated without needing an external count variable?
- What's the performance difference for encode/decode operations?
- Can we use unsafe types while maintaining safety guarantees?

### 2. Struct vs Class for Data Containers
**Questions:**
- Can Swift arrays exist inside C-compatible structs?
- Does Swift have "self-aware" arrays (with embedded length) suitable for C interop?
- What's the overhead of struct vs class for our use case?

### 3. Mutability and Optimization
**Questions:**
- What's the performance difference between mutable and immutable arrays?
- Can we make something mutable immutable for optimization?
- What's the conversion overhead?
- Does `let` vs `var` affect performance at this level?

### 4. Memory Layout
**Questions:**
- Can we control memory layout for better cache performance?
- What's the alignment and padding overhead?
- Can we use `@frozen` or other attributes?

## Benchmark Structure

Each test measures:
1. **Allocation time**: Creating the data structure
2. **Write time**: Populating with data
3. **Read time**: Iterating and accessing data
4. **Encode time**: Converting to bytes (simulated wire format)
5. **Decode time**: Converting from bytes back to structure
6. **Memory overhead**: Actual memory used vs theoretical minimum

## Running Benchmarks

```bash
cd swift
swift build -c release
swift run -c release SwiftResearch
```

## Expected Outcomes

This research will inform decisions for the Swift gold standard:
- Which array type to use for variable-length fields
- Whether to use structs or classes for messages
- Optimal mutability patterns
- Whether unsafe types are worth the complexity
