# SDP - Serial Data Protocol - AI Agent Instructions

## Project Overview

**SDP is a high-performance binary serialization format with code generation for cross-language data transfer.** Think "Protocol Buffers without schema evolution" - optimized for controlled environments where both encoder and decoder use the same schema version.

**Key differentiators:**
- Fixed-width integers (predictable, no varint)
- Zero runtime dependencies in generated code
- 6.1× faster encoding than Protocol Buffers
- Multi-language support: Go (reference), C++, Rust, Swift

**Version:** 0.2.0-rc1 (Release Candidate)

**⚠️ Current Status:** Testing and benchmarking infrastructure needs modernization (see `MODERNIZATION_SUMMARY.md` for details)

---

## Architecture & Core Concepts

### 1. Single Source of Truth: Documentation Structure

**CRITICAL:** Each piece of information exists in exactly ONE authoritative location (per `CONSTITUTION.md` v2.0):

| Information | Authoritative Source |
|-------------|---------------------|
| Wire format specification | `DESIGN_SPEC.md` Section 6 |
| Schema syntax | `DESIGN_SPEC.md` Section 2 |
| API usage examples | `QUICK_REFERENCE.md` |
| Performance claims | `PERFORMANCE_ANALYSIS.md` + `benchmarks/` |
| Testing approach | `TESTING_STRATEGY.md` |
| Project governance | `CONSTITUTION.md` |

**Do NOT duplicate information.** Link to authoritative sources instead. When updating specs, modify the single source only.

### 2. Code Generation Pipeline

```
.sdp schema files → Parser → Validator → Generator → {Go, C++, Rust} code
                     (AST)     (checks)   (templates)
```

**Generator location:** `internal/generator/{go,cpp,rust}/`  
**Parser:** `internal/parser/` (parses Rust-like syntax)  
**Validator:** `internal/validator/` (type resolution, circular refs, reserved keywords)

**Critical:** Generator NEVER calls `os.Exit()` directly - always return errors. CLI layer (`cmd/sdp-gen/main.go`) handles exit codes.

### 3. Multi-Language Support

Each language has its own implementation:
- **Go:** Reference implementation, `internal/generator/go/`
- **C++:** Fastest implementation, `internal/generator/cpp/`
- **Rust:** In progress, `internal/generator/rust/`
- **Swift:** Wrapper around C++

**Wire format is language-agnostic** - C++ encoder → Go decoder works seamlessly (byte-for-byte identical).

---

## Critical Design Principles

### Simplicity Bias
When in doubt, choose simplicity over features. Examples:
- Fixed-width integers instead of varint (simpler, faster)
- No built-in compression (compose with gzip instead)
- Zero dependencies in generated code

**Before adding features:** Can users solve this by composing existing features?

### Performance First
All performance claims MUST be backed by verified benchmarks in `benchmarks/`. See `benchmarks/RESULTS.md` for methodology.

**Never say "we'll optimize later."** Performance is a core feature.

### Zero Dependencies
Generated code uses ONLY standard libraries:
- Go: stdlib only
- C++: C++17 stdlib only
- No serialization frameworks, compression libs, or network code

**Rationale:** Users compose SDP with their choice of compression/transport via `io.Writer`/`io.Reader` interfaces.

---

## Development workflows, generation, and benchmarking (modernized)

This project uses a Makefile-driven orchestration layer and small shell helpers to keep generation, tests, and benchmarks reproducible across languages. The single entry-point for developer workflows is the project root Makefile and the `benchmarks/Makefile` for performance runs.

High-level commands you will use:

```bash
# Rebuild the generator and regenerate all target languages (Go/C++/Rust/Swift)
make generate

# Run the full benchmark suite (Go + C++ + Rust)
cd benchmarks && make bench

# Run a single language benchmark (examples):
cd benchmarks && make bench-go         # Go
cd benchmarks && make bench-cpp        # C++
cd benchmarks && make bench-rust       # Rust (message + byte mode)
```

Why this structure:
- `make generate` ensures `sdp-gen` is rebuilt first (FORCE pattern) and then regenerates all packages. This prevents stale generated code and avoids brittle local edits.
- Benchmarks are driven by shell scripts / Make targets under `benchmarks/` so they build the generated libraries and run the platform-appropriate harnesses (Go `go test -bench`, C++ compiled runner, Rust `cargo run` wrappers).

Where test data comes from (important for comparable benchmarks):
- Encoding (encode) benchmarks use canonical JSON fixtures under `/testdata/data` or the project-specific JSON samples in `/testdata/*/*.json`. These are decoded (once) using the generated Go code to produce in-memory structs used for encode-only benchmarks. Example: `benchmarks/go/comparison_test.go` uses `testdata/binaries/*.sdpb` and the generated Go decoder to create `testData`.
- Decoding (decode) benchmarks use pre-encoded canonical binary files stored in `/testdata/binaries/*.sdpb`. These files are the authoritative decode fixtures to ensure every language decodes the exact same bytes.

Important conventions (do not diverge):
- "Schemas" that are used for correctness tests live in `/testdata/schemas/` (these include small, focused schemas used in unit tests). These are NOT automatically the data used for performance benches unless explicitly referenced by the bench harness.
- Performance benches are driven by two sources:
    1. The canonical JSON samples (for encode benchmarks). Path: `/testdata/data` or per-schema JSON files under `/testdata/*/*.json`.
    2. The canonical binary fixtures for decode benches. Path: `/testdata/binaries/*.sdpb`.
- The `benchmarks/` harnesses will call the generated libraries (Go, C++, Rust) to encode/decode these fixtures so that the byte-for-byte encoding is verified and the timings are comparable across languages.

Generator & runtime notes:
- The generator `cmd/sdp-gen` must be rebuilt before regenerating packages to ensure runtime changes (e.g., added helper `wire_slice::check_bounds`) are propagated. `make generate` does this automatically.
- Generated Rust/C++ code may include small runtime helpers embedded in `src/wire*.rs` or `*.cpp` — keep those in sync via generation rather than hand-editing the generated outputs.

Benchmark methodology (short):
- Warm-up: run a small number of iterations to warm caches and JIT/OS caches.
- Repeatable iteration counts: benchmarks use fixed iteration counts or Go's `b.N` style to get stable measurements.
- Memory reporting: Go benches use `b.ReportAllocs()`; C++/Rust harnesses should report memory and size where appropriate.
- Compare encode vs decode separately (different bottlenecks). Use the same sample data for both encode and decode when possible (JSON→encode, pre-encoded binary→decode).

Cross-language test and bench checklist (to reproduce reliably):
1. Run `make generate` to rebuild `sdp-gen` and regenerate all outputs.
2. Build the generated libraries for each language (Go: none, C++: compiled in `benchmarks` Makefile, Rust: `cargo build`) — the `benchmarks/Makefile` builds them automatically.
3. Ensure the canonical JSON and binary fixtures exist in `/testdata/data` and `/testdata/binaries`.
4. Run `cd benchmarks && make bench` to execute the full harness. This will:
     - Use Go to pre-decode JSON fixtures into in-memory structs for encode benches
     - Use pre-generated `.sdpb` binary fixtures for decode benches
     - Build and run language-specific runners (C++ binary, Rust `cargo run`, Go `go test -bench`)
5. Collect numbers and store them in `benchmarks/RESULTS.md` with the machine specs, OS, and command used.

If you must add a new benchmark, follow this pattern:
1. Add a language-appropriate harness under `benchmarks/` (Go: *_test.go; C++: cpp/.../bench_*.cpp; Rust: small binary under `benchmarks/rust/*`)
2. Point the harness at the canonical fixtures under `/testdata` (JSON for encode, .sdpb for decode)
3. Wire the build into `benchmarks/Makefile` so `make bench` runs it
4. Run locally and add results to `benchmarks/RESULTS.md`

This keeps benchmarks reproducible, comparable, and easy to run in CI or on other machines.

---

## Project-Specific Conventions

### Schema Files (.sdp)

**Syntax:** Rust-like with SDP extensions
```rust
// Regular struct (byte mode)
struct Plugin {
    id: u32,
    name: string,
    metadata: ?Metadata,  // Optional field
}

// Self-describing message (message mode)
message PluginEvent {
    timestamp: u64,
    plugin_id: u32,
}
```

**Validation rules:**
- No circular references (direct or indirect)
- No reserved keywords (Go/Rust/C/Swift combined list)
- Self-contained schemas (no cross-file references)
- Optional fields: `?StructType` only (no `?u32`, `?string`, `?[]T`)

### Naming Conventions

**Schema:** snake_case or PascalCase (user choice)  
**Generated Go:** PascalCase types, camelCase fields  
**Generated C++:** PascalCase types, camelCase methods

Examples:
- Schema: `audio_device` → Go: `AudioDevice` → C++: `AudioDevice`
- Method: `AudioDevice::encode()`, `AudioDevice::decode()`

### Wire Format Rules

**All multi-byte values are little-endian.** No alignment padding.

**Primitives:** Direct binary encoding  
**Strings:** `[u32 length][UTF-8 bytes]`  
**Arrays:** `[u32 count][element_0]...[element_n]`  
**Structs:** Fields in schema definition order  
**Optional:** `[u8 presence][data if present]`  
**Messages:** `[u64 type_id][u32 size][payload]`

### Size Limits (Built-in Validation)

```go
MaxSerializedSize = 128 MB       // Total data
MaxStringSize     = 10 MB        // Per string
MaxArraySize      = 100,000      // Per array
MaxNestingDepth   = 20 levels    // Struct nesting
```

---

## Common Pitfalls & Solutions

### 1. Don't Duplicate Documentation

❌ **Wrong:** Copy performance numbers into README.md  
✅ **Right:** Link to `PERFORMANCE_ANALYSIS.md`

❌ **Wrong:** Repeat wire format in QUICK_REFERENCE.md  
✅ **Right:** Reference `DESIGN_SPEC.md` Section 6

### 2. No CGO in Test Files

❌ **Wrong:** `import "C"` in `_test.go` files  
✅ **Right:** Use subprocess communication or wire format fixtures

**Reason:** CGO makes tests non-portable and breaks cross-compilation.

### 3. Schema Evolution

SDP does NOT support schema evolution. Breaking changes:
- Reordering fields ❌
- Changing field types ❌
- Removing fields ❌

**Workarounds:**
- Use message mode for versioning (`message PluginV1`, `message PluginV2`)
- Use optional fields for backward-compatible additions
- Coordinate schema updates across encoder/decoder

### 4. Performance Claims

Always benchmark before claiming performance improvements. Use:
```bash
cd benchmarks && make bench
```

Compare against baseline and reference implementation (Protocol Buffers).

---

## Key Files to Reference

**When working on:**
- Wire format → `DESIGN_SPEC.md` Section 6
- Parser → `internal/parser/parser.go`, `DESIGN_SPEC.md` Section 2
- Validator → `internal/validator/validator.go`
- Go generator → `internal/generator/go/*.go`
- C++ generator → `internal/generator/cpp/*.go`, `CPP_IMPLEMENTATION.md`
- Testing → `TESTING_STRATEGY.md`, `integration_test.go`
- Benchmarks → `benchmarks/RESULTS.md`, `benchmarks/MEMORY_ANALYSIS.md`
- Documentation style → `DOCUMENTATION_GUIDELINES.md`

**Cross-language compatibility:**
- `crossplatform_test.go` - Cross-language wire format verification

---

## Generator Template Patterns

### Go Generator (internal/generator/go/)

**Type mapping:**
```go
schemaType → goType
"u32"     → "uint32"
"string"  → "string"
"[]T"     → "[]T"
"?T"      → "*T"  // Optional
```

**Generated functions per struct:**
- `EncodeSTRUCT(src *STRUCT) ([]byte, error)`
- `DecodeSTRUCT(dest *STRUCT, data []byte) error`
- `EncodeSTRUCTToWriter(src *STRUCT, w io.Writer) error`
- `DecodeSTRUCTFromReader(dest *STRUCT, r io.Reader) error`

### C++ Generator (internal/generator/cpp/)

**Optimizations applied:**
1. **Wire format structs** for fixed-layout types (10-30× speedup)
2. **Bulk memcpy** for primitive arrays (2× speedup)
3. **Inline encoding** for nested structs/arrays (5× speedup)
4. **Pre-computed string lengths** (9× faster than strlen)

See `CPP_IMPLEMENTATION.md` for detailed optimization strategy.

---

## Performance Expectations

**Go implementation (M1 Mac, baseline):**
- Primitives: ~26ns encode, ~25ns decode
- AudioUnit (1,759 params): ~39µs encode, ~98µs decode
- 6.1× faster than Protocol Buffers

**C++ implementation (fastest):**
- Primitives: 8.6ns encode, 3.37ns decode (zero-copy)
- AudioUnit: 49.7ns encode
- 3× faster than Go, 18× faster than Protocol Buffers

**Before claiming performance improvements, run benchmarks and update `benchmarks/RESULTS.md`.**

---

## Git Workflow

**Commit message format:**
```
<type>: <summary>

<body>

<references>
```

**Types:** `spec:`, `test:`, `docs:`, `gen:`, `bench:`, `fix:`, `refactor:`, `archive:`

**Example:**
```
gen: Add optional field support to C++ generator

Implement presence byte encoding for ?T fields.
Updated encoder to write 0x01 + data for present, 0x00 for absent.

Refs: CPP_IMPLEMENTATION.md Section 3.1
```

---

## What Makes This Project Successful

From `CONSTITUTION.md` Section 10:

✅ **Focus over features** - Shipped optional fields, message mode, streaming instead of endlessly planning  
✅ **Simplicity bias** - Fixed-width integers, zero dependencies kept us fast  
✅ **Performance first** - Fair benchmarks proved claims  
✅ **Single source of truth** - DESIGN_SPEC.md is authoritative  
✅ **Test-driven** - 415 tests give confidence to iterate  
✅ **Honest trade-offs** - "When NOT to Use" builds trust

**Maintain these principles when contributing.**

---

## Questions or Unclear Sections?

1. **Schema syntax unclear?** → Check `DESIGN_SPEC.md` Section 2 and `testdata/*.sdp` examples
2. **Wire format confusion?** → `DESIGN_SPEC.md` Section 6 with hex examples
3. **Performance targets?** → `benchmarks/RESULTS.md` and `PERFORMANCE_ANALYSIS.md`
4. **Testing approach?** → `TESTING_STRATEGY.md` and `integration_test.go`

**For AI agents:** Always check authoritative sources first. Don't assume - grep search the codebase or read relevant docs.
