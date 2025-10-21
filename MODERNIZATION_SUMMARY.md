# SDP Modernization Summary

**Date:** October 21, 2025  
**Status:** Recommendations Ready for Implementation

---

## Current Reality (CORRECTED)

### Working Implementations
1. **✅ Go:** Reference implementation, 415 tests, well-tested
2. **✅ C++:** Fastest implementation, production-ready
3. **⚠️ Rust:** Exists but needs work
4. **✅ Swift:** Wrapper around C++ (packaging only)
5. **❌ C:** Deprecated, superseded by C++, **should be removed**

### Testing State
- ✅ **Go:** Excellent (415 tests, TestMain, cross-language)
- ❌ **C++:** Fragmented (individual Makefiles)
- ❌ **Rust:** Unclear status
- ❌ **Swift:** No automated tests
- ❌ **Unified runner:** Doesn't exist

### Benchmarking State
- ❌ **No canonical test data** - Need consistent JSON input across implementations
- ❌ **Inconsistent execution** - Each language runs differently
- ❌ **Hard to compare** - No standard output format

---

## What Needs Fixing

1. **Remove C implementation** - Superseded by C++, causes confusion
2. **Create unified test runner** - Make/CMake orchestrating shell scripts
3. **Canonical benchmark data** - JSON files as input for all implementations
4. **Standardize benchmark output** - Comparable results across languages
5. **CI/CD pipeline** - Automate the above

---

## FlatBuffers Comparison: What We Learned

### They Do Better
1. **CMake orchestration** - Single `cmake . && make && ctest` runs everything
2. **Per-language test scripts** - `tests/JavaTest.sh`, `tests/RustTest.sh`, etc.
3. **Canonical wire format fixtures** - C++ generates `monsterdata_test.mon`, all languages verify
4. **Standardized benchmarks** - Same test cases, comparable output format
5. **Mature CI/CD** - Parallel language jobs, cross-language verification

### Our Advantages
1. **Simpler wire format** - Fixed-width integers (FlatBuffers has varint)
2. **Faster for our use case** - 6.1× faster than Protocol Buffers
3. **Two-tier C API** - Zero-copy expert + arena easy (FlatBuffers only has one API)
4. **TestMain auto-regeneration** - Go tests always use fresh generated code

---

## Make vs CMake Decision

### Option A: Simple Make (Recommended for Your Case)
```makefile
# Makefile (root)
.PHONY: test test-go test-cpp test-rust benchmark clean

# Run all tests
test: test-go test-cpp test-rust
	@echo "✓ All tests passed"

test-go:
	@./tests/test_go.sh

test-cpp:
	@./tests/test_cpp.sh

test-rust:
	@./tests/test_rust.sh

# Run all benchmarks with canonical data
benchmark:
	@./benchmarks/run_all.sh

clean:
	@find . -name "*.o" -delete
	@find testdata/generated -type f -delete
```

**Why Make is sufficient:**
- ✅ You just need orchestration (call scripts)
- ✅ Simpler, everyone knows it
- ✅ No learning curve
- ✅ Portable

**When you'd need CMake:**
- You need parallel test execution (`ctest -j4`)
- You need test filtering (`ctest -R cpp`)
- You need XML output for CI dashboards
- You're building C++ libraries with complex dependencies

**Verdict:** Start with Make. Add CMake later only if you need its features.

---

## Recommended Action Plan (REVISED)

### Phase 1: Cleanup & Foundation (Week 1) - **START HERE**

```bash
# 1. Remove deprecated C implementation
rm -rf testdata/*_c/
rm -rf internal/generator/c/
git rm ...  # Commit the removal

# 2. Create test orchestration structure
mkdir tests
mkdir benchmarks/data  # Canonical JSON inputs

# 3. Create Makefile
cat > Makefile << 'EOF'
.PHONY: test test-go test-cpp benchmark

test: test-go test-cpp
	@echo "✓ All tests passed"

test-go:
	@./tests/test_go.sh

test-cpp:
	@./tests/test_cpp.sh

benchmark:
	@./benchmarks/run_all.sh

clean:
	@./tests/clean_all.sh
EOF

# 4. Create Go test wrapper (already works)
cat > tests/test_go.sh << 'EOF'
#!/bin/bash
set -e
echo "=== Testing Go Implementation ==="
go test -v -cover ./...
echo "✓ Go tests passed"
EOF
chmod +x tests/test_go.sh
```

### Phase 2: C++ Integration (Week 2)

```bash
# 1. Create canonical benchmark data
cat > benchmarks/data/audiounit.json << 'EOF'
{
  "plugins": [
    {
      "name": "Reverb",
      "manufacturer_id": "ACME",
      "parameters": [...]
    }
  ]
}
EOF

# 2. Create C++ test script
cat > tests/test_cpp.sh << 'EOF'
#!/bin/bash
set -e
echo "=== Testing C++ Implementation ==="

# Build and test each schema
for schema in primitives audiounit arrays; do
    echo "Testing ${schema}..."
    cd testdata/${schema}_cpp
    make clean && make test
    cd ../..
done

# Wire format verification (Go-generated fixture)
./testdata/verify_cpp_compat

echo "✓ C++ tests passed"
EOF
chmod +x tests/test_cpp.sh

# 3. Create C++ benchmark runner
cat > benchmarks/run_cpp.sh << 'EOF'
#!/bin/bash
# Reads benchmarks/data/*.json
# Encodes using C++
# Reports: ns/op, MB/s
./benchmarks/cpp_bench benchmarks/data/audiounit.json
EOF
```

### Phase 3: Unified Benchmarking (Week 3)

```bash
# Master benchmark runner
cat > benchmarks/run_all.sh << 'EOF'
#!/bin/bash
set -e

echo "=== Running Benchmarks Against Canonical Data ==="
echo "Input: benchmarks/data/audiounit.json (115KB)"
echo ""

# Go
echo "Go Implementation:"
go test -bench=BenchmarkAudioUnit -benchtime=1s ./benchmarks/go/

# C++
echo ""
echo "C++ Implementation:"
./benchmarks/run_cpp.sh

# Rust (if working)
if [ -f benchmarks/run_rust.sh ]; then
    echo ""
    echo "Rust Implementation:"
    ./benchmarks/run_rust.sh
fi

echo ""
echo "✓ Benchmarks complete"
echo "Compare results above ^^^"
EOF
chmod +x benchmarks/run_all.sh
```

### Phase 4: CI/CD (Week 4)

```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]

jobs:
  test-go:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
      - run: make test-go
  
  test-cpp:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: sudo apt-get install -y g++ cmake
      - run: make test-cpp
  
  benchmark:
    if: github.ref == 'refs/heads/main'
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: make benchmark
```

---

## Proposed Directory Structure (REVISED)

```
serial-data-protocol/
├── .github/workflows/
│   └── test.yml                    # CI/CD
│
├── cmd/sdp-gen/                    # Generator
│
├── internal/
│   ├── parser/                     # Lexer, parser, AST
│   ├── validator/                  # Type checking
│   ├── generator/
│   │   ├── golang/                 # Go code gen ✓
│   │   ├── cpp/                    # C++ code gen ✓
│   │   ├── rust/                   # Rust code gen ⚠️
│   │   └── swift/                  # Swift wrapper gen ✓
│   └── wire/                       # Wire format helpers
│
├── tests/                          # Test orchestration scripts
│   ├── test_go.sh                  # Go: go test ./...
│   ├── test_cpp.sh                 # C++: build + run tests
│   ├── test_rust.sh                # Rust: cargo test
│   └── clean_all.sh                # Cleanup
│
├── testdata/
│   ├── schemas/                    # Test schemas (.sdp files)
│   │   ├── primitives.sdp
│   │   ├── audiounit.sdp
│   │   └── arrays.sdp
│   └── generated/                  # Generated code (gitignored)
│       ├── primitives/
│       │   ├── go/
│       │   ├── cpp/
│       │   └── rust/
│       └── audiounit/
│           ├── go/
│           └── cpp/
│
├── benchmarks/
│   ├── data/                       # Canonical input data (JSON)
│   │   ├── audiounit.json          # 62 plugins, 115KB
│   │   ├── primitives.json
│   │   └── arrays.json
│   ├── go/
│   │   └── bench_test.go           # Reads data/*.json
│   ├── cpp/
│   │   └── bench.cpp               # Reads data/*.json
│   ├── run_go.sh                   # Run Go benchmarks
│   ├── run_cpp.sh                  # Run C++ benchmarks
│   └── run_all.sh                  # Master runner + comparison
│
├── Makefile                        # Orchestration (not build)
├── go.mod
└── README.md
```

**Key changes from before:**
1. ✅ **Removed all `*_c/` directories** - C implementation gone
2. ✅ **`benchmarks/data/`** - JSON inputs (not JSON outputs)
3. ✅ **`tests/` with shell scripts** - Make just orchestrates
4. ✅ **`testdata/schemas/`** - Schemas separate from generated code

---

## Decision Points (REVISED)

### 1. Make vs CMake?
**Recommendation: Start with Make**
- ✅ Simpler for orchestration
- ✅ No learning curve
- ✅ Does what you need (call scripts)
- 🤔 Add CMake later if you need parallel execution or test filtering

### 2. Remove C Implementation?
**Recommendation: YES, remove immediately**
- ❌ Superseded by C++
- ❌ Causes confusion
- ❌ Maintenance burden
- ✅ C++ is faster anyway

### 3. Benchmark Data Format?
**Clarified: JSON is INPUT, not output**
- ✅ `benchmarks/data/audiounit.json` - canonical test data
- ✅ All implementations read same JSON
- ✅ All implementations encode to binary
- ✅ Compare encode/decode times

### 4. Rust Priority?
**Your call - it "might need work"**
- Option A: Fix Rust now (add to Phase 2)
- Option B: Focus on Go/C++ first, Rust later

---

## Immediate Next Steps (This Week)

### Day 1: Cleanup
```bash
# Remove C implementation
git rm -r testdata/*_c/
git rm -r internal/generator/c/
git commit -m "Remove deprecated C implementation (superseded by C++)"
```

### Day 2: Test Orchestration
```bash
# Create structure
mkdir tests
mkdir benchmarks/data

# Create Makefile
cat > Makefile << 'EOF'
.PHONY: test test-go test-cpp benchmark

test: test-go test-cpp
	@echo "✓ All tests passed"

test-go:
	@./tests/test_go.sh

test-cpp:
	@./tests/test_cpp.sh

benchmark:
	@./benchmarks/run_all.sh
EOF

# Create tests/test_go.sh
echo '#!/bin/bash' > tests/test_go.sh
echo 'set -e' >> tests/test_go.sh
echo 'go test -v -cover ./...' >> tests/test_go.sh
chmod +x tests/test_go.sh

# Verify: make test-go
```

### Day 3: Canonical Benchmark Data
```bash
# Create benchmarks/data/audiounit.json
# (Export from your existing test data)

# Create benchmarks/run_go.sh
# (Reads audiounit.json, runs Go benchmarks)

# Create benchmarks/run_cpp.sh
# (Reads audiounit.json, runs C++ benchmarks)

# Create benchmarks/run_all.sh
# (Calls both, displays comparison)
```

### Day 4: C++ Test Integration
```bash
# Create tests/test_cpp.sh
# (Build C++ tests, run against same data as Go)

# Verify: make test-cpp
```

### Day 5: CI/CD
```bash
# Create .github/workflows/test.yml
# (Run make test on every push)

# Push and verify GitHub Actions work
```

---

## Success Criteria (REVISED)

**When modernization is complete:**

✅ Single `make test` runs all working implementations (Go, C++)  
✅ Single `make benchmark` compares all implementations against same data  
✅ CI/CD passes on every commit  
✅ No C implementation confusion  
✅ Canonical JSON benchmark data in `benchmarks/data/`  
✅ New contributors can run tests without reading docs  

---

## Final Recommendation

**Approach:** Make-orchestrated shell scripts (not CMake)

**Timeline:** 1 week
- Day 1: Remove C
- Day 2: Create Makefile + test scripts
- Day 3: Canonical benchmark data
- Day 4: C++ integration
- Day 5: CI/CD

**Priority Order:**
1. Remove C (cleanup)
2. Make + test scripts (orchestration)
3. Canonical benchmark data (JSON inputs)
4. C++ test integration
5. CI/CD (automation)

**Defer to later:**
- Rust fixes (address when you have time)
- CMake (only if you need its features)
- Swift testing (wrapper works, low priority)

**Result:** Clean, simple, testable codebase with Go (reference) and C++ (fast) as primary implementations.
