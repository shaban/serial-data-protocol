# Benchmarks Directory Structure

**Reorganized:** October 22, 2025

## Directory Layout

```
benchmarks/
├── README.md                   # Overview and methodology
├── RESULTS.md                  # Performance analysis and results
├── MEMORY_ANALYSIS.md          # Memory profiling details
├── STRUCTURE.md                # This file
├── Makefile                    # Unified benchmark orchestration
│
├── go/                         # Go benchmarks
│   ├── go.mod                  # Go dependencies
│   ├── go.sum
│   ├── comparison_test.go      # SDP byte mode benchmarks
│   ├── message_mode_test.go    # SDP message mode benchmarks
│   └── cross_protocol_test.go  # SDP vs ProtoBuf vs FlatBuffers
│
├── cpp/
│   ├── bytemode/              # C++ byte mode benchmarks
│   │   ├── bench_c.c          # C implementation benchmark
│   │   └── bench_cpp.cpp      # C++ implementation benchmark
│   └── messagemode/           # C++ message mode benchmarks
│       └── bench_audiounit.cpp # AudioUnit message mode benchmark
│
└── rust/                      # Rust benchmarks (future)
    ├── bytemode/
    └── messagemode/
```

## Pattern Consistency

This structure **mirrors** `testdata/` organization:

```
testdata/                       benchmarks/
├── cpp/                        ├── cpp/
│   ├── audiounit/             │   ├── bytemode/
│   ├── arrays/                │   └── messagemode/
│   └── messagemode/           │
├── go/                         ├── go/
│   ├── audiounit/             │   ├── comparison_test.go
│   └── messagemode/           │   └── message_mode_test.go
└── rust/                       └── rust/
```

**Benefits:**
- ✅ Easy to navigate (same pattern everywhere)
- ✅ Language-first organization
- ✅ Clear separation: byte mode vs message mode
- ✅ Ready for Rust implementation

## File Purposes

### Go Benchmarks

| File | Purpose | Compares |
|------|---------|----------|
| `comparison_test.go` | SDP byte mode only | - |
| `message_mode_test.go` | SDP message mode only | Byte vs Message overhead |
| `cross_protocol_test.go` | Cross-protocol | SDP vs ProtoBuf vs FlatBuffers |

### C++ Benchmarks

| File | Purpose | Language |
|------|---------|----------|
| `cpp/bytemode/bench_c.c` | Byte mode | C99 |
| `cpp/bytemode/bench_cpp.cpp` | Byte mode | C++17 |
| `cpp/messagemode/bench_audiounit.cpp` | Message mode | C++17 |

## Running Benchmarks

```bash
# All benchmarks (Go + C++)
make bench

# Language-specific
make bench-go
make bench-cpp

# Mode-specific
make bench-go-byte        # Go byte mode
make bench-go-message     # Go message mode
make bench-cpp-byte       # C++ byte mode
make bench-cpp-message    # C++ message mode

# Cross-protocol comparison
make bench-go-cross       # SDP vs PB vs FB
```

## Benchmark Data

All benchmarks use **AudioUnit** schema with real-world data:
- **Source:** `testdata/binaries/audiounit.sdpb` (110 KB)
- **Content:** 62 plugins, 1,759 parameters
- **Why:** Realistic workload (not micro-benchmarks)

## Schema Files

Schema definitions are in `testdata/schemas/`:
- `audiounit.sdp` - SDP schema
- `audiounit.proto` - Protocol Buffers schema
- `audiounit.fbs` - FlatBuffers schema

**Previously:** These were in `benchmarks/` (moved for consistency)

## Migration Notes

**Before (October 21, 2025):**
```
benchmarks/
├── comparison_test.go         ❌ Root level
├── message_mode_test.go       ❌ Root level
├── cross_protocol_test.go     ❌ Root level
├── standalone/
│   ├── bench_c.c              ❌ Unclear location
│   └── bench_cpp.cpp          ❌ Unclear location
├── audiounit.proto            ❌ Should be in testdata/schemas/
└── audiounit.fbs              ❌ Should be in testdata/schemas/
```

**After (October 22, 2025):**
```
benchmarks/
├── go/                        ✅ Language-first
│   ├── comparison_test.go     ✅ Organized
│   └── ...
├── cpp/                       ✅ Language-first
│   ├── bytemode/             ✅ Mode separation
│   └── messagemode/          ✅ Mode separation
└── rust/                      ✅ Ready for future
```

## Adding New Benchmarks

**For byte mode:**
```bash
# Go
benchmarks/go/new_bytemode_test.go

# C++
benchmarks/cpp/bytemode/bench_newfeature.cpp
```

**For message mode:**
```bash
# Go
benchmarks/go/new_messagemode_test.go

# C++
benchmarks/cpp/messagemode/bench_newfeature.cpp
```

**For Rust (future):**
```bash
benchmarks/rust/bytemode/bench_audiounit.rs
benchmarks/rust/messagemode/bench_audiounit.rs
```

## See Also

- [README.md](README.md) - Benchmark methodology and fairness rules
- [RESULTS.md](RESULTS.md) - Performance results and analysis
- [MEMORY_ANALYSIS.md](MEMORY_ANALYSIS.md) - Memory profiling details
- [Makefile](Makefile) - Build and run commands
