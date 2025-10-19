# SDP Rust Reference Implementation

This directory contains the **reference implementation** of the SDP wire protocol runtime in Rust.

## Purpose

This is **NOT** the generated code that users interact with. Instead, it serves as:

1. **Source of truth** for `wire.rs` and `wire_slice.rs` implementations
2. **Reference** for understanding the protocol at a low level
3. **Testing ground** for wire format changes before codegen updates

## Generated Code Location

**Users should use the generated Rust crates**, not this reference implementation:

```bash
# Generated crates are self-contained and have no dependencies on this directory
testdata/primitives/rust/     # Self-contained primitives package
testdata/audiounit/rust/      # Self-contained audiounit package
testdata/arrays/rust/         # Self-contained arrays package
testdata/optional/rust/       # Self-contained optional package
testdata/nested/rust/         # Self-contained nested package
testdata/complex/rust/        # Self-contained complex package
```

Each generated crate includes:
- ✅ Embedded runtime (`src/wire.rs`, `src/wire_slice.rs`)
- ✅ Generated types (`src/types.rs`)
- ✅ Generated encode logic (`src/encode.rs`)
- ✅ Generated decode logic (`src/decode.rs`)
- ✅ Example CLI helper (`examples/crossplatform_helper.rs`)
- ✅ Criterion benchmarks (`benches/benchmarks.rs`)
- ✅ Only external dependency: `byteorder`

## Why This Exists

The generator (`internal/generator/rust/runtime.go`) embeds these files directly into each generated crate. This directory is the **source** of those embedded files.

## Making Changes

If you need to modify the wire protocol:

1. **Edit the source files here** (`src/wire.rs`, `src/wire_slice.rs`)
2. **Update the generator** (`internal/generator/rust/runtime.go`) with the new content
3. **Regenerate all packages**: `go test -lang=rust`
4. **Verify tests pass**: All 51 tests should pass

## Structure

```
rust/sdp/
├── Cargo.toml          # Reference implementation manifest
├── README.md           # This file
└── src/
    ├── lib.rs          # Library root (re-exports wire modules)
    ├── wire.rs         # Core wire protocol (Vec<u8> implementation)
    └── wire_slice.rs   # Zero-copy wire protocol (&[u8] implementation)
```

## Generated Code is Independent

**Important:** Generated crates do **NOT** depend on this directory. They are completely self-contained. This directory is only used during code generation.

```
┌─────────────────┐         ┌──────────────────┐
│  rust/sdp/      │  ──────▶│  Generator       │
│  (source)       │  embed  │  (runtime.go)    │
└─────────────────┘         └──────────────────┘
                                     │
                                     │ generates
                                     ▼
                            ┌──────────────────┐
                            │  Generated Crate │
                            │  (self-contained)│
                            │                  │
                            │  src/wire.rs     │ ◀─ embedded copy
                            │  src/wire_slice.rs│ ◀─ embedded copy
                            │  src/types.rs    │ ◀─ generated
                            │  src/encode.rs   │ ◀─ generated
                            │  src/decode.rs   │ ◀─ generated
                            └──────────────────┘
```

## Testing

This reference implementation has no tests of its own. Testing happens in the generated crates:

```bash
# Test all generated Rust crates
go test -lang=rust

# Test specific package
cd testdata/primitives/rust && cargo test

# Benchmark specific package
cd testdata/primitives/rust && cargo bench
```

## See Also

- `RUST_GOLD_STANDARD.md` - Complete documentation of Rust code generation
- `PERFORMANCE_COMPARISON.md` - Go vs Rust performance analysis
- `internal/generator/rust/` - Rust code generator implementation
