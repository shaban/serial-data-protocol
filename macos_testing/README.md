# macOS C++ Interop Testing

This directory contains experiments to validate C++ interoperability with macOS native languages:

## Objective

Determine if we can use the C++ implementation (`testdata/audiounit_cpp/`) from:
1. **Objective-C++** (.mm files) - for macOS/iOS native apps
2. **Swift 6** (native C++ interop) - for modern Swift codebases

## Success Criteria

1. **Compilation**: Code compiles without errors
2. **Functionality**: Can encode/decode audiounit.sdpb successfully
3. **Performance**: Swift wrapper overhead should be < 10% vs pure C++
4. **API Quality**: Resulting API feels native to the language

## Tests

### 1. Objective-C++ Bridge (`objcpp_test/`)
- `.mm` file that imports C++ headers
- Calls C++ encode/decode functions
- Wraps in Objective-C interface for Cocoa apps
- Measures performance vs pure C++

### 2. Swift 6 C++ Interop (`swift_cpp_test/`)
- Swift Package that imports C++ headers directly
- Uses Swift 6's `@_expose(Cxx)` / C++ interop
- Adds Swift-native API sugar
- Benchmarks against pure C++ and pure Swift

## Decision Tree

**If both work well:**
- ✅ Keep C++ as the native implementation
- ✅ Provide Objective-C++ bridge for legacy code
- ✅ Provide Swift wrapper with < 10% overhead
- ❌ Archive C implementation (incomplete encoder)
- ❌ Archive pure Swift implementation (too slow)

**If Swift 6 interop has issues:**
- ✅ Keep C++ for Windows/Linux
- ✅ Keep pure Swift for macOS/iOS (accept slower performance)
- ✅ Provide Objective-C++ bridge as alternative

**If both fail:**
- ✅ Keep C++ for Windows/Linux
- ✅ Keep pure Swift for macOS/iOS
- 🔧 Fix and complete C implementation for maximum performance

## Benchmark Baseline

Pure C++ performance (from `benchmarks/standalone/bench_cpp`):
- Encode: 29.3 μs/op
- Decode: 30.7 μs/op
- Roundtrip: 59.0 μs/op

Target for Swift wrapper: < 32 μs encode, < 34 μs decode (< 10% overhead)
