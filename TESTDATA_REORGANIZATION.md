# Testdata Reorganization Plan

## Problem
Current `testdata/` is messy:
- `.sdp` schemas mixed with generated code directories
- `.json` canonical data scattered at root
- `.sdpb` reference binaries at root
- Per-schema subdirs (`primitives/`, `arrays/`) contain generated code per language
- Root-level `*_cpp/` directories for C++ generated code
- Hard to tell what's canonical vs generated
- Cross-language tests in Go test suite (wrong layer)
- `bench` mode in sdp-gen generates TCP servers (wrong abstraction)

## Proposed Structure

```
testdata/
├── schemas/              # Source of truth: .sdp schema files
│   ├── primitives.sdp
│   ├── arrays.sdp
│   ├── nested.sdp
│   ├── optional.sdp
│   ├── audiounit.sdp
│   ├── complex.sdp
│   └── valid_*.sdp       # Validation test schemas
│
├── data/                 # Canonical JSON input for benchmarks
│   ├── primitives.json
│   ├── arrays.json
│   ├── nested.json
│   ├── optional.json
│   └── plugins.json      # AudioUnit data
│
├── binaries/            # Reference wire format (.sdpb)
│   ├── primitives.sdpb
│   ├── arrays_primitives.sdpb
│   ├── arrays_structs.sdpb
│   ├── nested.sdpb
│   ├── optional.sdpb
│   └── audiounit.sdpb
│
├── go/                  # Generated Go code for all schemas
│   ├── primitives/
│   ├── arrays/
│   ├── nested/
│   ├── optional/
│   ├── audiounit/
│   └── complex/
│
├── cpp/                 # Generated C++ code for all schemas
│   ├── primitives/
│   ├── arrays/
│   ├── optional/
│   └── audiounit/
│
├── rust/                # Generated Rust code for all schemas
│   ├── primitives/
│   ├── arrays/
│   ├── nested/
│   ├── optional/
│   ├── audiounit/
│   └── complex/
│
└── README.md            # Explains structure and data flow
```

## What Gets Removed

### From testdata/
- ❌ `primitives/`, `arrays/`, `nested/`, `optional/`, `audiounit/`, `complex/` - flatten into `go/primitives/`, etc.
- ❌ `primitives_cpp/`, `arrays_cpp/`, `optional_cpp/`, `audiounit_cpp/` - move to `cpp/primitives/`, etc.
- ❌ `test_*.c` files - obsolete C test stubs
- ❌ `sdp-gen` binary - wrong location, belongs in `cmd/` build

### From Go tests
- ❌ `crossplatform_test.go` - Cross-language wire format testing (should be shell scripts)
- ❌ `crossplatform_bench_test.go` - Cross-language benchmarking (should be Make targets)
- ❌ `benchmarks/cpp_vs_go_test.go` - Comparison benchmarks (should be external scripts)

### From sdp-gen
- ❌ `-mode bench` flag - Don't generate TCP servers in codegen
- ❌ `-ast-json` flag - Not used, over-engineered
- ❌ Benchmark server generation in Rust/Swift generators

## Go Testing Scope (What Stays)

`integration_test.go` should ONLY test:
1. **Schema parsing** - Does parser correctly read `.sdp` files?
2. **Validation** - Does validator catch errors (circular refs, reserved keywords)?
3. **Code generation** - Does Go generator produce compilable code?
4. **Wire format** - Does encoded data match canonical `.sdpb` files?
5. **Roundtrip** - Encode → bytes → decode → same struct?

`size_test.go` can stay (tests size limits).

NO cross-language tests. That's the job of shell scripts + Make orchestration.

## Data Flow

```
Schema (.sdp) → Parser → Validator → Generator → Code (Go/C++/Rust)
                                                     ↓
JSON data → Code.Encode() → Binary (.sdpb) → Code.Decode() → Verify
```

**Canonical data:**
- `.sdp` files = schemas (source code)
- `.json` files = test/benchmark inputs (readable)
- `.sdpb` files = reference wire format (byte-for-byte identical across languages)

**Generated code:**
- `testdata/go/primitives/` = Go code for primitives schema
- `testdata/cpp/primitives/` = C++ code for primitives schema

## Migration Steps

1. Create new directory structure
2. Move `.sdp` files to `schemas/`
3. Move `.json` files to `data/`
4. Move `.sdpb` files to `binaries/`
5. Flatten generated code directories
6. Update `integration_test.go` imports
7. Delete cross-language test files
8. Remove `bench` mode from sdp-gen
9. Add `testdata/README.md`

## Why This Structure?

✅ **Clear separation** - Canonical vs generated  
✅ **Language-agnostic** - Per-language dirs at top level  
✅ **Discoverable** - `data/` is obviously input data  
✅ **Composable** - Shell scripts can iterate `schemas/*.sdp`  
✅ **Maintainable** - No nested per-schema directories  
✅ **Testable** - Go tests only test Go, scripts test cross-lang

## Example Commands After Migration

```bash
# Generate all Go code
for schema in testdata/schemas/*.sdp; do
    name=$(basename $schema .sdp)
    sdp-gen -schema $schema -output testdata/go/$name -lang go
done

# Generate all C++ code
for schema in testdata/schemas/*.sdp; do
    name=$(basename $schema .sdp)
    sdp-gen -schema $schema -output testdata/cpp/$name -lang cpp
done

# Run cross-language wire format test
./tests/verify_wire_format.sh primitives

# Benchmark Go vs C++
./benchmarks/compare.sh primitives
```

Clean, simple, composable.
