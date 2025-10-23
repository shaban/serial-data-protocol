# FlatBuffers/Protocol Buffers Benchmark Issues

## 🚨 Problems Found

### 1. **Duplicate Schemas** ❌
```
benchmarks/audiounit.proto    ← DUPLICATE (identical to testdata/schemas/)
benchmarks/audiounit.fbs      ← DUPLICATE (slightly different namespace)
testdata/schemas/audiounit.proto  ← Official source
testdata/schemas/audiounit.fbs    ← Official source
```

**Violation:** Single source of truth principle

### 2. **Not Make-Orchestrated** ❌
- `testdata/flatbuffers/generate.sh` - Manual shell script
- `testdata/protobuf/generate.sh` - Manual shell script
- NOT integrated into `make generate`
- NOT integrated into `make benchmark`

**Violation:** Manual workflow, not reproducible

### 3. **Wrong Directory Structure** ❌
```
Current:
testdata/flatbuffers/
  generate.sh  ← Manual script
  go/          ← Generated code mixed with testdata

Should be:
testdata/
  schemas/
    audiounit.proto  ← Official schema
    audiounit.fbs    ← Official schema
  generated/
    protobuf/go/     ← Generated protobuf code
    flatbuffers/go/  ← Generated flatbuffers code
benchmarks/go/       ← Benchmarks only (no schemas)
```

**Violation:** Generated code not in testdata/generated/

### 4. **Broken During Restructure** ❌
- `benchmarks/comparison_test.go` accidentally created in root during `sed` operation
- Causes module ambiguity error
- Benchmarks don't compile

### 5. **No JSON Sample Data** ⚠️
- Protobuf/FlatBuffers benchmarks load data from `audiounit.sdpb` (SDP format)
- Then convert to Protobuf/FlatBuffers format in Go code
- Should use official JSON → generate .pb/.fb binaries

**Violation:** Not using official sample data workflow

---

## ✅ What They Do Right

1. ✅ Use the same AudioUnit schema (62 plugins, 1,759 parameters)
2. ✅ Fair comparison methodology (same data source)
3. ✅ Located in benchmarks/ directory
4. ✅ Use shell scripts for generation (just not integrated)

---

## 🔧 Required Fixes

### Priority 1: Remove Duplicates & Fix Build
```bash
# Remove duplicate schemas
rm benchmarks/audiounit.proto
rm benchmarks/audiounit.fbs

# Remove accidentally created file
rm benchmarks/comparison_test.go

# Update references to use testdata/schemas/
```

### Priority 2: Move Generated Code
```bash
# Move to proper location
mv testdata/flatbuffers/go testdata/generated/flatbuffers/go
mv testdata/protobuf/go testdata/generated/protobuf/go

# Update import paths in benchmarks
```

### Priority 3: Integrate into Make
```makefile
# Add to root Makefile
generate-protobuf:
	@cd testdata/protobuf && ./generate.sh

generate-flatbuffers:
	@cd testdata/flatbuffers && ./generate.sh

generate: generate-sdp generate-protobuf generate-flatbuffers
```

### Priority 4: Create JSON→Binary Workflow
```bash
# Create testdata/data/audiounit_protobuf.json
# Create testdata/data/audiounit_flatbuffers.json
# Generate .pb and .fb binaries from JSON
# Use in benchmarks instead of runtime conversion
```

---

## 📋 Implementation Plan

### Phase 1: Quick Fix (Get Benchmarks Working)
1. Remove benchmarks/comparison_test.go ✓
2. Remove duplicate schemas in benchmarks/
3. Fix generate.sh path references ✓
4. Test benchmarks compile

### Phase 2: Move Generated Code
5. Move flatbuffers/go to testdata/generated/
6. Move protobuf/go to testdata/generated/
7. Update benchmark imports
8. Test benchmarks still work

### Phase 3: Make Integration
9. Add generate-protobuf target
10. Add generate-flatbuffers target
11. Include in main generate target
12. Test `make generate` works

### Phase 4: JSON Workflow (Optional)
13. Create JSON sample data for protobuf/flatbuffers
14. Generate .pb/.fb binaries from JSON
15. Update benchmarks to use pre-generated binaries
16. Document in testdata/MANIFEST.md

---

## 🎯 Why This Matters

**Current state violates our principles:**
- ❌ Duplicate schemas (no single source of truth)
- ❌ Manual workflows (not reproducible)
- ❌ Generated code in wrong location (mixed with testdata)
- ❌ Not integrated into make generate
- ❌ Not using official sample data workflow

**After fixes:**
- ✅ Single source: testdata/schemas/audiounit.{proto,fbs}
- ✅ Make-orchestrated: `make generate` handles everything
- ✅ Proper structure: testdata/generated/{protobuf,flatbuffers}/
- ✅ Reproducible: Delete and regenerate anytime
- ✅ Consistent: Same workflow as Go/C++/Rust/Swift

---

## 📊 Current Status

**FlatBuffers:**
- Schema: benchmarks/audiounit.fbs (duplicate) + testdata/schemas/audiounit.fbs (official)
- Generated: testdata/flatbuffers/go/ (wrong location)
- Generation: Manual shell script (not in Makefile)
- Benchmarks: benchmarks/go/cross_protocol_test.go (broken - import issues)

**Protocol Buffers:**
- Schema: benchmarks/audiounit.proto (duplicate) + testdata/schemas/audiounit.proto (official)
- Generated: testdata/protobuf/go/ (wrong location)
- Generation: Manual shell script (not in Makefile)
- Benchmarks: benchmarks/go/cross_protocol_test.go (broken - import issues)

**Action Required:** Follow implementation plan to bring into compliance.
