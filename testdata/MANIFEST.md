# Official Test Data Manifest

**Purpose:** Single source of truth for all SDP test data, schemas, and binary reference files.

---

## 📁 Directory Structure

```
testdata/
  schemas/          ← Official schema definitions (.sdp files)
  data/             ← Official sample data (.json files)
  binaries/         ← Official binary reference files (.sdpb files)
  generated/        ← Generated code (DO NOT EDIT - regenerate with `make generate`)
    go/
    cpp/
    rust/
    swift/
```

---

## 📝 Official Schemas (`testdata/schemas/*.sdp`)

### Core Test Schemas
- **`primitives.sdp`** - All primitive types (u8, u16, u32, u64, i8, i16, i32, i64, f32, f64, string)
- **`arrays.sdp`** - Array types (primitive arrays, struct arrays)
- **`nested.sdp`** - Nested struct composition
- **`optional.sdp`** - Optional field testing (`?StructType`)
- **`complex.sdp`** - Complex nested structures

### Real-World Schemas
- **`audiounit.sdp`** - Production example: PluginRegistry (62 plugins, 1,759 parameters, ~110KB)

### Message Mode Schemas
- **`message_test.sdp`** - Message mode testing (Point, Rectangle structs)

### Validation Schemas
- **`valid_basic.sdp`** - Basic validation tests
- **`valid_complex.sdp`** - Complex validation tests
- **`valid_crlf.sdp`** - CRLF line ending handling

---

## 📊 Official Sample Data (`testdata/data/*.json`)

Each JSON file corresponds to a schema and contains valid test data:

| JSON File | Schema | Description |
|-----------|--------|-------------|
| `primitives.json` | `primitives.sdp` | All primitive type values |
| `arrays.json` | `arrays.sdp` | Primitive and struct arrays |
| `nested.json` | `nested.sdp` | Nested struct hierarchies |
| `optional.json` | `optional.sdp` | Optional field combinations |
| `plugins.json` | `audiounit.sdp` | Real-world AudioUnit data (110KB) |

**Rules:**
- ✅ JSON files are human-readable, easy to edit
- ✅ Version-controlled for traceability
- ❌ Never duplicate - single source of truth

---

## 🔒 Official Binary Reference Files (`testdata/binaries/*.sdpb`)

Binary files used for cross-language compatibility testing:

### Byte Mode Reference Files
| File | Schema | Source | Size | Description |
|------|--------|--------|------|-------------|
| `primitives.sdpb` | primitives.sdp | primitives.json | ~200B | All primitive types |
| `arrays_primitives.sdpb` | arrays.sdp | arrays.json | ~500B | Primitive arrays |
| `arrays_structs.sdpb` | arrays.sdp | arrays.json | ~800B | Struct arrays |
| `nested.sdpb` | nested.sdp | nested.json | ~300B | Nested structs |
| `optional.sdpb` | optional.sdp | optional.json | ~150B | Optional fields |
| `audiounit.sdpb` | audiounit.sdp | plugins.json | 110KB | Real-world data |

### Message Mode Reference Files
| File | Schema | Generator | Type ID | Description |
|------|--------|-----------|---------|-------------|
| `message_point.sdpb` | message_test.sdp | Go | 1 | Point message |
| `message_rectangle.sdpb` | message_test.sdp | Go | 2 | Rectangle message |
| `message_point_cpp.sdpb` | message_test.sdp | C++ | 1 | Point (C++ encoded) |
| `message_rectangle_cpp.sdpb` | message_test.sdp | C++ | 2 | Rectangle (C++ encoded) |

**Rules:**
- ✅ Generated from official .json files using `sdp-encode`
- ✅ Version-controlled (ground truth for wire format)
- ✅ Never manually edited
- ✅ Byte-for-byte identical across all languages

---

## 🔧 Generation Commands

### Generate Binary Reference Files

```bash
# Byte mode .sdpb files
sdp-encode -schema testdata/schemas/primitives.sdp \
           -data testdata/data/primitives.json \
           -output testdata/binaries/primitives.sdpb

sdp-encode -schema testdata/schemas/audiounit.sdp \
           -data testdata/data/plugins.json \
           -output testdata/binaries/audiounit.sdpb
```

### Generate Code from Schemas

```bash
# Regenerate all code (clean slate)
make generate

# This will:
# 1. Delete testdata/generated/*
# 2. Run sdp-gen for each schema in testdata/schemas/
# 3. Generate Go, C++, and Rust code
```

---

## 🚫 What NOT to Do

### ❌ DO NOT manually edit generated code
```bash
# WRONG - editing generated files
vim testdata/generated/rust/audiounit/src/types.rs

# RIGHT - modify generator, then regenerate
vim internal/generator/rust/types_gen.go
make generate
```

### ❌ DO NOT create duplicate test data
```bash
# WRONG - creating test data in multiple places
cp testdata/data/audiounit.json benchmarks/audiounit_test.json

# RIGHT - use official data with Make variables
AUDIOUNIT_DATA=$(DATA_DIR)/plugins.json
```

### ❌ DO NOT hardcode paths
```rust
// WRONG - hardcoded relative path
let data = fs::read("../../testdata/binaries/audiounit.sdpb")?;

// RIGHT - use environment variable or Make variable
let data = fs::read(env::var("AUDIOUNIT_DATA")?)?;
```

---

## 📦 Generated Code (`testdata/generated/`)

**Status:** ⚠️ NEVER MANUALLY EDIT

Generated code directories contain ONLY output from `sdp-gen`:

```
testdata/generated/
  go/{schema}/
    types.go
    encode.go
    decode.go
    errors.go
    message_encode.go    # If schema has structs
    message_decode.go    # If schema has structs
    
  cpp/{schema}/
    types.hpp
    encode.cpp
    decode.cpp
    message_encode.cpp   # If schema has structs
    message_decode.cpp   # If schema has structs
    
  rust/{schema}/
    Cargo.toml
    src/
      types.rs
      encode.rs
      decode.rs
      errors.rs
      message_encode.rs  # If schema has structs
      message_decode.rs  # If schema has structs
```

**Verification:**
```bash
# Verify no manual edits
make verify-generated

# Output:
# ✓ Generated code is clean (no manual edits)
# OR
# ❌ ERROR: Generated code has been manually edited!
```

---

## 🎯 Usage in Tests and Benchmarks

### Tests (use generated code + official binaries)

```rust
// tests/crosslang/rust/test_audiounit.rs
use std::fs;

#[test]
fn test_decode_audiounit() {
    // Use official binary reference file
    let data = fs::read(env!("BINARIES_DIR").to_owned() + "/audiounit.sdpb")?;
    
    // Use generated decoder
    let registry = PluginRegistry::decode_from_slice(&data)?;
    
    assert_eq!(registry.plugins.len(), 62);
}
```

### Benchmarks (use generated code + official binaries)

```cpp
// benchmarks/cpp/messagemode/bench_audiounit.cpp

// Use Make variable for data path
std::string data_path = std::getenv("AUDIOUNIT_DATA");
auto data = read_file(data_path);

// Use generated decoder from testdata/generated/cpp/audiounit
PluginRegistry registry;
decode_plugin_registry(&registry, data.data(), data.size());
```

---

## 📚 Related Documentation

- **`TESTING_STRATEGY.md`** - Overall testing approach
- **`TESTING_INFRASTRUCTURE_AUDIT.md`** - Infrastructure analysis and restructuring
- **`DESIGN_SPEC.md`** - Wire format specification
- **`Makefile.vars`** - Path variables for consistent access

---

## ✅ Verification Checklist

Before committing changes:

- [ ] All .sdpb files generated from official .json files
- [ ] No duplicate test data
- [ ] No hardcoded paths (use Make variables)
- [ ] No manual edits to generated code
- [ ] `make generate` completes successfully
- [ ] `make verify-generated` passes
- [ ] `make test` passes
- [ ] `make benchmark` runs successfully

---

**Last Updated:** October 22, 2025  
**Maintained by:** SDP Contributors  
**Related Issues:** TESTING_INFRASTRUCTURE_AUDIT.md
