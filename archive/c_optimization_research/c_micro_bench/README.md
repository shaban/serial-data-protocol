# C Encoder Micro-Benchmarks

Experimental directory for isolating and testing C encoder optimization techniques.

## Purpose

Before implementing the final C encoder, we need to understand:
- Which approaches are fastest
- What layout information helps
- How to close the 1.75x performance gap with Go

## Tests

### Test 0: Writer vs Fixed Buffer
**Question:** Is our dynamic writer causing overhead compared to Go's pre-allocated `[]byte`?

- Dynamic writer with realloc safety
- Dynamic writer pre-sized (no realloc)
- Fixed buffer (Go-style)

**Expected:** Pre-sized writer ≈ fixed buffer (capacity checks are cheap)

### Test 1: Layout Optimization
**Question:** Can compile-time layout knowledge speed up encoding?

Approaches:
- Field-by-field (current approach)
- Bulk copy with wire format struct
- Direct pointer writes to computed offsets

**Data:** `AllPrimitives` (12 fields: 11 primitives + 1 string)

### Test 2: String Methods
**Question:** What's the fastest way to copy strings?

Comparison:
- `snprintf` (what we avoided in bench)
- `strlen + memcpy` (dynamic)
- Pre-computed length + memcpy (current)
- Compile-time literal macro

**Expected:** Pre-computed >> strlen >> snprintf

### Test 3: Array Optimization
**Question:** Can we speed up array encoding?

Current: Loop with individual element writes
Optimized: Bulk memcpy for primitive arrays

Tests:
- f32 arrays (with conversion)
- u32 arrays (direct copy)

**Expected:** Bulk copy ~10x faster for 100-element arrays

### Test 4: Struct Optimization
**Question:** How to handle nested structs efficiently?

Approaches:
- Recursive function calls (current)
- Flattened inline encoding
- Bulk copy with wire struct
- Direct pointer writes

**Data:** `Scene` with nested `Rectangle` and `Point`

## Running

```bash
# Build and run all tests
make run

# Run individual test
make 0_writer_vs_buffer
./0_writer_vs_buffer

# Clean
make clean
```

## Next Steps

After gathering results:

A. Handwrite isolated implementations of basic schemas (primitives, arrays, nested, complex)
B. Get results comparing approaches
C. Discuss results and finalize approach
D. Update C_API_SPECIFICATION.md with findings
E. Remove c_micro_bench directory
F. Implement official C encoder with proven optimizations

## Expected Outcomes

1. **Baseline:** Understand current approach performance
2. **Opportunities:** Identify biggest optimization wins
3. **Trade-offs:** Balance code complexity vs performance
4. **Strategy:** Design codegen to emit optimal patterns

## Current Hypothesis

The 1.75x gap with Go (65µs vs 37µs) comes from:
- Function call overhead (writer API)
- Individual field encoding (not bulk)
- Array element loops (not bulk copy)
- Missing layout optimizations

Goal: Close gap to <1.2x of Go performance.
