# Changelog

All notable changes to Serial Data Protocol will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.2.0-rc1] - 2025-10-18

### Added

**Feature 1: Optional Struct Fields**
- Schema syntax: `?Type` prefix for optional fields (e.g., `metadata: ?Metadata`)
- Wire format: 1-byte presence indicator + conditional data
- Generated code uses pointers for optional fields (Go: `*Metadata`, nil if absent)
- Restrictions: Only structs can be optional (no primitives, arrays, or array elements)
- Performance: Absent fields decode 10× faster than present (3.15 ns vs 31.49 ns)

**Feature 2: Self-Describing Message Mode**
- Schema syntax: `message` keyword instead of `struct`
- Wire format: 10-byte header (8-byte type ID + 4-byte payload size) + payload
- Type ID: FNV-1a hash of message name (deterministic, collision-resistant)
- Generated dispatcher function routes by type ID
- Generated type ID constants for manual discrimination
- Performance: ~2× overhead vs regular mode (85.54 ns vs 44.25 ns roundtrip)
- Use cases: Event streams, persistent storage, protocol multiplexing

**Feature 3: Streaming I/O**
- Generated functions: `EncodeXToWriter(src, w io.Writer)` and `DecodeXFromReader(dest, r io.Reader)`
- Enables composition with stdlib: files, compression (gzip/zstd), network (net.Conn), pipes
- Zero new dependencies in generated code
- No baked-in compression - users compose with their choice of libraries
- Full integration with all features (regular structs, optional fields, messages)

**Testing & Documentation**
- 415 tests passing (unit + integration)
- Comprehensive performance benchmarks with real-world data
- Streaming I/O integration tests (file, gzip compression, network simulation)
- PERFORMANCE_ANALYSIS.md with detailed measurements
- Updated README.md with realistic comparisons and honest limitations
- Updated DESIGN_SPEC.md with RC features and performance data
- Updated QUICK_REFERENCE.md with practical examples for all features

### Performance

Measured performance characteristics:

**Large dataset (62 plugins, 1759 parameters, 115 KB):**
- Encode: 37.5 µs (1 allocation)
- Decode: 85.2 µs (4,638 allocations)
- ~10× faster than Protocol Buffers for this use case

**Small messages (primitives, ~50 bytes):**
- Regular mode: 44.25 ns roundtrip
- Message mode: 85.54 ns roundtrip (+93%)
- Optional present: 58.38 ns roundtrip
- Optional absent: 15.55 ns roundtrip (65% faster)

**Compression (with gzip):**
- Size reduction: 68% typical
- Composition via streaming I/O functions

### Changed

- Schema parser now recognizes `?` prefix for optional fields
- Schema parser now recognizes `message` keyword
- Code generator produces streaming I/O functions for all types
- Import detection includes `io` package when needed
- Wire format remains backward compatible for regular structs

### Fixed

- Linter warnings in decode_gen.go (unnecessary blank identifiers)

---

## [0.1.0] - 2025-10-15

### Added

**Core Implementation (Phases 1-4)**
- Schema parser with Rust-like syntax
- Schema validator with comprehensive checks
- Go code generator (types, encoder, decoder, errors)
- Wire format implementation (fixed-width integers, little-endian)
- Primitive types: u8, u16, u32, u64, i32, i64, f32, f64, bool, string
- Composite types: arrays (`[]T`), structs, nested structures
- Size limits: 10 MB strings, 100k array elements, 20 nesting levels
- 238 tests passing

**Documentation**
- DESIGN_SPEC.md - Complete wire format specification
- IMPLEMENTATION_PLAN.md - Development roadmap
- TESTING_STRATEGY.md - Testing approach
- CONSTITUTION.md - Project governance

**Tooling**
- Command-line code generator: `sdp-gen`
- Flags: `-schema`, `-output`, `-lang` (default: go)

### Performance

Initial benchmarks:
- Plugin enumeration (62 items): ~37 µs encode, ~85 µs decode
- Zero-copy design: single allocation for encoding
- Cross-platform tested: macOS, Linux

---

## [Unreleased]

### Planned

- C code generation (next priority)
- Rust code generation
- Swift code generation
- Python code generation
- Enum support
- Map types (as syntactic sugar for `[]KeyValue`)

### Not Planned

- Schema evolution features (breaking change detection, versioning)
- RPC framework (use gRPC with custom serialization)
- Schema registry (use existing solutions)
- Dynamic typing support
- Varint encoding (fixed-width by design)
- Built-in compression (compose with stdlib)

---

## Version History

- **0.2.0-rc1** (2025-10-18) - Release Candidate: Optional fields, message mode, streaming I/O
- **0.1.0** (2025-10-15) - Initial release: Core wire format and Go code generation

---

## Migration Guide

### From 0.1.0 to 0.2.0-rc1

**Breaking Changes:** None - all features are additive and backward compatible.

**New Features Available:**

1. **Optional Fields** - Add to existing schemas without breaking decoders:
   ```rust
   // Before (0.1.0)
   struct Plugin {
       id: u32,
       name: string,
   }
   
   // After (0.2.0-rc1) - backward compatible addition
   struct Plugin {
       id: u32,
       name: string,
       metadata: ?Metadata,  // New optional field
   }
   ```
   
   Old decoders (0.1.0) cannot read schemas with optional fields, but 0.2.0-rc1 decoders can read 0.1.0 schemas.

2. **Message Mode** - Convert regular structs to messages for type identification:
   ```rust
   // Before
   struct Event { ... }
   
   // After - use when you need type discrimination
   message Event { ... }
   ```
   
   Wire formats are incompatible - message mode adds 10-byte header.

3. **Streaming I/O** - Use new functions for composition:
   ```go
   // Before (still works)
   bytes, err := plugin.EncodePlugin(&p)
   
   // After - compose with file/compression/network
   file, _ := os.Create("data.sdp")
   err := plugin.EncodePluginToWriter(&p, file)
   ```

**Regenerate Code:**
```bash
# Regenerate all schemas with 0.2.0-rc1 generator
sdp-gen -schema your_schema.sdp -output ./generated -lang go
```

**Performance Impact:**
- Regular structs: No change
- Optional fields: +1 byte per optional field, minimal time cost
- Message mode: +10 bytes header, +93% time (use only when needed)
- Streaming I/O: Identical performance to direct encode/decode
