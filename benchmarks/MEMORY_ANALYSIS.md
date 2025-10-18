# Memory Usage Analysis

Detailed memory profiling comparison of SDP, Protocol Buffers, and FlatBuffers using real-world AudioUnit plugin data (62 plugins, 1,759 parameters, ~115 KB wire size).

## Summary

| Protocol | Peak Heap (Encode) | Peak Heap (Decode) | Total Peak | Allocations (Encode) | Allocations (Decode) |
|----------|-------------------|-------------------|------------|---------------------|---------------------|
| **SDP** | 112 KB | 201 KB | **313 KB** | 1 | 4,638 |
| **Protocol Buffers** | 136 KB | 310 KB | 446 KB | 1 | 6,651 |
| **FlatBuffers** | 900 KB | **0 KB** | 900 KB | 51 | **0** |

### Key Findings

- **SDP uses the least total RAM**: 313 KB total peak (30% less than Protocol Buffers, 65% less than FlatBuffers)
- **FlatBuffers has zero-copy decode**: Literally 0 bytes allocated during decode (accesses buffer directly)
- **SDP has minimal encode overhead**: Only 1 allocation per encode operation
- **Protocol Buffers uses most allocations**: 6,651 allocations during decode (43% more than SDP)

## Detailed Breakdown

### Encoding Memory Usage

**Peak Heap Allocation:**
```
SDP:              112.00 KB  (baseline)
Protocol Buffers: 135.70 KB  (+21% vs SDP)
FlatBuffers:      900.48 KB  (+704% vs SDP, +564% vs PB)
```

**Allocation Count:**
```
SDP:              1 allocation
Protocol Buffers: 1 allocation  
FlatBuffers:      51 allocations  (must build entire structure upfront)
```

**Analysis:**
- SDP and Protocol Buffers both use a single allocation for the output buffer
- FlatBuffers requires 51 allocations to build the structure (offsets, vectors, strings)
- FlatBuffers allocates 8× more memory than SDP because it builds the entire serialized format in memory
- SDP's fixed-size encoding allows precise pre-allocation with minimal overhead

### Decoding Memory Usage

**Peak Heap Allocation:**
```
FlatBuffers:      0.00 KB    (zero-copy, accesses buffer directly)
SDP:              200.63 KB  (baseline for deserialization)
Protocol Buffers: 310.05 KB  (+55% vs SDP)
```

**Allocation Count:**
```
FlatBuffers:      0 allocations  (zero-copy)
SDP:              4,638 allocations
Protocol Buffers: 6,651 allocations  (+43% vs SDP)
```

**Analysis:**
- FlatBuffers true zero-copy: doesn't allocate ANY memory, accesses buffer in-place
- SDP allocates 201 KB to deserialize into native Go structs (62 plugins × 1,759 parameters)
- Protocol Buffers allocates 310 KB (55% more than SDP) due to additional pointer indirection
- SDP has 2,013 fewer allocations than Protocol Buffers (30% reduction)

### Total Memory Footprint

**Combined Encode + Decode Peak:**
```
SDP:              312.63 KB  (baseline, LOWEST)
Protocol Buffers: 445.74 KB  (+43% vs SDP)
FlatBuffers:      900.48 KB  (+188% vs SDP, +102% vs PB)
```

**Why SDP Uses Less Memory:**
1. **Efficient encoding**: Fixed-size format with precise allocation (no varint overhead)
2. **Compact decode structs**: Direct mapping to Go native types without wrapper objects
3. **Fewer allocations**: 30% fewer allocations than Protocol Buffers during decode
4. **No intermediate buffers**: Decodes directly into target structs

**Why Protocol Buffers Uses More:**
1. **Pointer overhead**: Generated code uses pointers for many fields (`*string`, `*int32`)
2. **Additional indirection**: Wrapper types for handling proto3 semantics
3. **More allocations**: 43% more allocations during decode creates fragmentation

**Why FlatBuffers Dominates Decode (but not Encode):**
1. **Zero-copy philosophy**: Designed to never deserialize (access buffer directly)
2. **Encode cost**: Must build entire structure upfront (900 KB peak)
3. **Trade-off**: Fast decode, slow encode, huge memory footprint during serialization

## Allocation Patterns

### Per-Operation Breakdown (from benchmark profiling)

**Encoding:**
```
SDP:              114,689 bytes/op, 1.015 mallocs/op
Protocol Buffers:  98,305 bytes/op, 1.007 mallocs/op
FlatBuffers:      922,104 bytes/op, 51.15 mallocs/op
```

**Decoding:**
```
SDP:              205,452 bytes/op, 4,638 mallocs/op
Protocol Buffers: 317,494 bytes/op, 6,651 mallocs/op
FlatBuffers:            0 bytes/op, 0 mallocs/op (zero-copy)
```

### Memory Efficiency Ratios

**Allocations per Parameter (1,759 total parameters):**
```
SDP Decode:              2.64 allocations/parameter
Protocol Buffers Decode: 3.78 allocations/parameter  (+43%)
FlatBuffers Encode:      0.03 allocations/parameter  (upfront cost)
```

**Bytes per Parameter:**
```
SDP Encode:              65 bytes/parameter
Protocol Buffers Encode: 56 bytes/parameter  (varint compression)
FlatBuffers Encode:      524 bytes/parameter (no compression, alignment)
```

## Garbage Collection Impact

### GC Pressure Analysis

**Encode (objects created per operation):**
- SDP: 1 allocation → minimal GC pressure
- Protocol Buffers: 1 allocation → minimal GC pressure  
- FlatBuffers: 51 allocations → moderate GC pressure

**Decode (objects created per operation):**
- SDP: 4,638 allocations → moderate GC pressure
- Protocol Buffers: 6,651 allocations → **high GC pressure** (+43% vs SDP)
- FlatBuffers: 0 allocations → **zero GC pressure**

**Real-World Implications:**
- **High-frequency IPC (1000s/sec)**: SDP's 30% fewer allocations reduces GC pauses
- **Batch processing**: FlatBuffers zero-copy decode eliminates GC completely (if you can work with accessors)
- **Memory-constrained systems**: SDP uses 65% less RAM than FlatBuffers, 30% less than Protocol Buffers

## Use Case Recommendations

### Choose SDP if:
- ✅ Memory usage is critical (30-65% less than alternatives)
- ✅ You need deserialized Go structs (not accessors)
- ✅ You want minimal GC pressure (30% fewer allocations vs PB)
- ✅ Total RAM footprint matters more than wire size
- ✅ You're building high-frequency IPC (e.g., plugin communication)

### Choose Protocol Buffers if:
- ✅ Wire size is more critical than RAM (14% smaller than SDP)
- ✅ Schema evolution is required (adding fields without breaking compatibility)
- ✅ Cross-language compatibility is essential
- ⚠️ Accept 43% higher memory usage and GC pressure vs SDP

### Choose FlatBuffers if:
- ✅ Decode performance is the ONLY metric that matters (8.8 ns vs 98 µs)
- ✅ You can work with accessor methods instead of native structs
- ✅ You have 3× more RAM available for encoding (900 KB vs 313 KB total)
- ⚠️ Encode is 8.6× slower than SDP
- ⚠️ Wire format is 8× larger than SDP/PB

## Methodology

**Test Data:**
- Real-world AudioUnit plugin data from `testdata/plugins.json`
- 62 plugins with metadata
- 1,759 parameters across all plugins
- Realistic distribution of string lengths and numeric values

**Measurement Approach:**
1. **Peak Heap**: Measured via `runtime.ReadMemStats()` before/after operations
2. **Allocation Count**: Reported by Go benchmark harness (`allocs/op`)
3. **Bytes Allocated**: Total heap allocation per operation (`B/op`)
4. **GC Impact**: Forced GC before each measurement to isolate operation impact

**Platform:**
- macOS darwin/arm64
- Apple M1
- Go 1.21

**Reproducibility:**
```bash
# Run peak heap tests
go test -run=TestPeakHeapUsage -v

# Run detailed memory benchmarks  
go test -bench=BenchmarkMemory -benchmem -benchtime=3s
```

## Conclusion

**SDP achieves the best overall memory profile:**
- **30% less total RAM** than Protocol Buffers (313 KB vs 446 KB)
- **65% less total RAM** than FlatBuffers (313 KB vs 900 KB)
- **30% fewer allocations** than Protocol Buffers during decode (4,638 vs 6,651)
- **Minimal GC pressure** with single-allocation encode

**Trade-offs are honest:**
- **Wire size**: Protocol Buffers is 14% smaller (varint encoding)
- **Decode speed**: FlatBuffers is 11,000× faster (zero-copy, but you don't get structs)
- **Schema evolution**: Protocol Buffers supports field additions, SDP doesn't

**For memory-constrained, high-frequency IPC**: SDP is the clear winner. If you need to deserialize thousands of messages per second into native Go structs while minimizing memory footprint and GC pressure, SDP uses 30-65% less RAM than alternatives.
