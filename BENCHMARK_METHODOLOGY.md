# Cross-Language Benchmark Methodology

## Current Results (macOS ARM64)

### Go (In-Process)
- **Encode**: 33.7M ops/sec (30 ns/op)
- **Decode**: 14.8M ops/sec (68 ns/op)

### Swift (Single Process Spawn)
- **Encode**: 4.7ms per invocation
- **Decode**: 7.1ms per invocation

## Understanding the Overhead

### Why Swift appears "slower"
The Swift measurements include **~4-7ms of process spawn overhead**:
- Process creation (fork/exec): ~100-500µs
- Binary loading: ~1-2ms (cached after first run)
- Swift runtime initialization: ~2-4ms
- Pipe setup (stdin/stdout): ~50-100µs

The actual encoding/decoding is likely **comparable to Go** (~30-100ns/op), but we're measuring the wrong thing.

## Three Benchmark Approaches

### 1. In-Process Benchmarks (Most Accurate)
**What**: Measure encode/decode within each language's native test harness
**How**: 
- Go: `go test -bench`
- Rust: `cargo bench`
- Swift: XCTest performance tests

**Pros**: 
- No exec overhead
- Direct algorithmic comparison
- Highest precision

**Cons**: 
- Doesn't test CLI usage
- Requires separate tooling per language

**Use when**: Comparing library/SDK performance

### 2. Amortized Batch Benchmarks (Recommended)
**What**: Helper executes N operations per invocation, amortizing spawn cost

**Example**:
```bash
# Helper does 100k encodes internally, writes all bytes to stdout
$ crossplatform_helper encode --count 100000
# Go measures total time and divides by 100k
```

**Calculation**:
- Total time: 50ms
- Spawn overhead: ~5ms (one-time)
- Encoding time: 45ms / 100,000 = **450ns/op**

**Pros**:
- Amortizes spawn cost over many operations
- Represents batch/pipeline workloads
- Still uses real CLI
- Fair cross-language comparison

**Cons**:
- Requires updating helpers to support batch mode
- Memory usage for large batches

**Use when**: Comparing batch processing performance

### 3. Single-Op CLI Benchmarks (Current Approach)
**What**: Spawn helper once per measurement, include all overhead

**Pros**:
- Tests real CLI usage pattern
- No helper changes needed
- Honest about total cost

**Cons**:
- Dominated by spawn overhead
- Can't compare algorithmic performance
- Misleading for library users

**Use when**: Measuring actual CLI invocation cost

## Recommendations

### For This Project
Given that SDP is a library first (with CLI helpers for testing), I recommend:

**Short term** (Current):
- Document that CLI benchmarks include spawn overhead
- Note: Go is in-process, others are single-exec
- Be honest that comparison isn't apples-to-apples

**Medium term** (Best):
- Add `--count N` flag to Rust/Swift helpers
- Helpers execute N encode/decode ops internally
- Report: "Swift encode: 100k ops in 47ms = 470ns/op (amortized)"
- Now comparable to Go's 30ns/op (with context that Swift has overhead)

**Long term** (Gold standard):
- Create native benchmark suites in each language:
  - `benchmarks/go_bench_test.go` (existing)
  - `benchmarks/rust/benches/encode.rs` (cargo bench)
  - `benchmarks/swift/Tests/BenchmarkTests.swift` (XCTest)
- Compare pure algorithmic performance
- Keep CLI benchmarks separate as "integration benchmarks"

## Answering Your Question

> Is os.Exec getting warmed up too on subsequent calls?

**Short answer**: Yes, but only certain parts.

**What gets warmed up:**
- ✅ OS page cache: Binary pages stay in RAM (2-3ms → ~200µs)
- ✅ Disk cache: Executable reads are cached
- ✅ Branch predictor: CPU learns patterns

**What doesn't warm up:**
- ❌ Process creation: Every spawn is a full fork/exec (~100-500µs)
- ❌ Runtime init: Swift stdlib initializes every time (~2-4ms)
- ❌ Memory allocation: Fresh heap per process
- ❌ File descriptors: New stdin/stdout pipes each time

**With enough iterations**: The binary load time vanishes (cached), but runtime init remains. For Swift, even after 1000 runs, you'll still pay ~2-4ms per invocation for runtime startup.

**Best solution**: Batch mode. Run `encode --count 10000` once instead of `encode` 10,000 times.

## Example: Batch Mode Implementation

### Swift Helper Update
```swift
if command.hasPrefix("encode-") {
    let count = CommandLine.arguments.contains("--count") 
        ? Int(CommandLine.arguments[...])! 
        : 1
    
    for _ in 0..<count {
        let bytes = makeTestAllPrimitives().encodeToBytes()
        FileHandle.standardOutput.write(Data(bytes))
    }
}
```

### Go Benchmark
```go
const batchSize = 100000
cmd := exec.Command(helper, "encode-AllPrimitives", "--count", strconv.Itoa(batchSize))
start := time.Now()
output, _ := cmd.Output()
elapsed := time.Since(start)

nsPerOp := float64(elapsed.Nanoseconds()) / float64(batchSize)
// Now comparable to Go's 30ns/op
```

## Conclusion

Your intuition is correct: **warmup helps, but not enough**. Process spawn overhead is fundamental and doesn't amortize with repeated calls. The best approach is **batch mode** where helpers perform N operations per invocation, amortizing the spawn cost over many operations.

For now, the benchmarks are honest about including overhead, and the documentation explains the tradeoffs. The path forward is clear: add batch support to helpers for fair comparison.
