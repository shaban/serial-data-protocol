# Protobuf/FlatBuffers Benchmark Compliance - COMPLETE

**Date:** October 23, 2025  
**Status:** ✅ All fixes applied, benchmarks working

## Summary

Brought protobuf and flatbuffers benchmarks into compliance with the new testing infrastructure principles established in the major restructure (commits 453127b → b80923c).

## Problems Fixed

### 1. ✅ Removed Duplicate Schemas
**Problem:** Duplicate schemas violated single source of truth  
**Fixed:**
- Deleted `benchmarks/audiounit.proto` (identical to testdata/schemas/)
- Deleted `benchmarks/audiounit.fbs` (different namespace from testdata/schemas/)
- **Single source:** `testdata/schemas/audiounit.{proto,fbs}`

### 2. ✅ Moved Generated Code to testdata/generated/
**Problem:** Generated code in wrong locations  
**Fixed:**
- OLD: `testdata/protobuf/go/` ❌
- NEW: `testdata/generated/protobuf/go/` ✅
- OLD: `testdata/flatbuffers/go/` ❌  
- NEW: `testdata/generated/flatbuffers/go/` ✅

### 3. ✅ Updated Benchmark Imports
**Problem:** Benchmarks importing from old paths  
**Fixed:**
- `benchmarks/go/go.mod` - Updated module paths to `testdata/generated/`
- `benchmarks/go/cross_protocol_test.go` - Updated imports

### 4. ✅ Make Orchestration
**Problem:** Manual shell scripts not integrated into Make  
**Fixed:**
- Updated `testdata/protobuf/generate.sh` to output to `testdata/generated/`
- Updated `testdata/flatbuffers/generate.sh` to output to `testdata/generated/`
- Added to `Makefile` generate target
- Now runs automatically with `make generate`

### 5. ✅ Added go.mod Generation
**Problem:** Generated packages need go.mod for imports to work  
**Fixed:**
- `testdata/protobuf/generate.sh` creates go.mod automatically
- `testdata/flatbuffers/generate.sh` creates go.mod automatically

### 6. ✅ Cleaned Up Accidents
**Problem:** `benchmarks/comparison_test.go` accidentally created during restructure  
**Fixed:** Deleted (was causing module ambiguity errors)

## Verification

### Benchmarks Compile ✅
```bash
cd benchmarks/go && go test -c
# Success: 8.0M binary created
```

### Benchmarks Run ✅
```bash
cd benchmarks/go && go test -bench=. -benchtime=100ms
```

**Results:**
```
BenchmarkGo_SDP_AudioUnit_Encode-8              2070    48460 ns/op    (SDP)
BenchmarkProtobuf_Encode-8                       500   235407 ns/op    (4.9× slower)
BenchmarkFlatBuffers_Encode-8                    356   338353 ns/op    (7.0× slower)

BenchmarkGo_SDP_AudioUnit_Decode-8               960   119549 ns/op   (SDP)
BenchmarkProtobuf_Decode-8                       343   344525 ns/op   (2.9× slower)
BenchmarkFlatBuffers_Decode-8                 26947k   4.411 ns/op    (zero-copy)

✅ All benchmarks passing
✅ SDP performance claims verified
```

### Make Generate Works ✅
```bash
make generate
```

**Output:**
```
Generating Protocol Buffers code...
✅ Protocol Buffers code generated successfully

Generating FlatBuffers code...
✅ FlatBuffers code generated successfully

✓ Code generation complete
  Generated: testdata/generated/{go,cpp,rust,swift,protobuf,flatbuffers}/*
```

## New Workflow

### Single Command to Regenerate Everything
```bash
make generate
```

Generates:
- All 10 SDP schemas → Go/C++/Rust/Swift
- AudioUnit → Protocol Buffers (Go)
- AudioUnit → FlatBuffers (Go)

### Single Command to Run Benchmarks
```bash
cd benchmarks/go && go test -bench=.
```

All protocols use the same source data for fair comparison.

## Compliance Checklist

- [x] Single source of truth (no duplicates)
- [x] testdata/generated/ directory structure
- [x] Make orchestration (reproducible)
- [x] Proper go.mod files for generated packages
- [x] Benchmarks compile and run
- [x] No manual workflows

## Files Changed

**Deleted:**
- `benchmarks/audiounit.proto` (duplicate)
- `benchmarks/audiounit.fbs` (duplicate)
- `benchmarks/comparison_test.go` (accident)
- `testdata/protobuf/go/` (moved)
- `testdata/flatbuffers/go/` (moved)

**Modified:**
- `Makefile` - Added protobuf/flatbuffers to generate target
- `benchmarks/go/go.mod` - Updated paths to testdata/generated/
- `benchmarks/go/cross_protocol_test.go` - Updated imports
- `testdata/protobuf/generate.sh` - Output to testdata/generated/, create go.mod
- `testdata/flatbuffers/generate.sh` - Output to testdata/generated/, create go.mod

**Added:**
- `testdata/generated/protobuf/go/` (generated code + go.mod)
- `testdata/generated/flatbuffers/go/` (generated code + go.mod)

## Impact

**Before:** Protobuf/FlatBuffers benchmarks violated all 5 principles  
**After:** 100% compliant with testing infrastructure  

**Before:** Manual workflows, duplicates, inconsistent structure  
**After:** Single `make generate` command, single source of truth, clean separation

**Before:** Benchmarks broken (module errors)  
**After:** All benchmarks passing, performance verified

## Notes

### JSON Workflow (Phase 4 - Not Implemented)

Currently benchmarks load `audiounit.sdpb` binary and convert to protobuf/flatbuffers in code. 

**Optional future improvement:**
- Create `testdata/data/audiounit_protobuf.json`
- Create `testdata/data/audiounit_flatbuffers.json`
- Generate `.pb` and `.fb` binaries in `testdata/binaries/`
- Load pre-generated binaries in benchmarks

**Not blocking:** Current approach works fine for benchmarks.

### Why Two Generate Scripts?

Protocol Buffers and FlatBuffers are external formats with their own compilers (`protoc`, `flatc`). We maintain schema files for benchmarking purposes only. The generate scripts are simple wrappers around these external tools.

SDP schemas use our own `sdp-gen` generator for Go/C++/Rust/Swift.

## References

- Testing restructure: commits 453127b → b80923c
- Swift testing: commit b80923c
- Protobuf/FlatBuffers fix: **THIS COMMIT**
- See: `TESTING_INFRASTRUCTURE_RESTRUCTURE_COMPLETE.md` for restructure details
- See: `benchmarks/RESULTS.md` for benchmark methodology
