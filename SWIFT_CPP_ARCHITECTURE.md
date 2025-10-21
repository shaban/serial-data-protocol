# Swift C++ Architecture

## Overview

As of this update, **Swift code generation has been completely redesigned**. Instead of generating pure Swift implementations, `sdp-gen -lang=swift` now generates:

1. **C++ implementation files** (types.hpp, encode.hpp/cpp, decode.hpp/cpp, endian.hpp)
2. **Swift Package Manager wrappers** (Package.swift + module.modulemap)

This approach leverages Swift's native C++ interoperability (Swift 5.9+) to provide optimal performance with minimal overhead.

## Why C++ Backend?

### Performance Results

During macOS interoperability testing, we evaluated 4 wrapper approaches:

| Approach | Encode Overhead | Decode Overhead | Status |
|----------|----------------|-----------------|--------|
| Pure Swift (original) | N/A | N/A | Removed (slow) |
| ObjC++ Object-Based | +1498% | +1608% | ❌ Rejected |
| ObjC++ Zero-Copy | +31% | +76% | ❌ Rejected |
| Swift C Bridge | +38% | +101% | ❌ Rejected |
| **Swift Barebone (C++)** | **-11%** | **+55%** | ✅ **Best** |

Pure C++ baseline: 29.3μs encode, 30.7μs decode

**Key Finding**: Direct C++ calls from Swift via module.modulemap had the best performance of all wrapper approaches, with encode actually faster than pure C++ due to compiler optimizations.

### Architecture Decision

For the **Go-orchestrated architecture** (the project's primary use case):
- Native code (Swift/ObjC++) queries platform APIs (AudioUnit, etc.)
- Builds C++ structs directly: `plugin.name = [nsString UTF8String]`
- Encodes once with C++ encoder: ~29μs
- Returns raw bytes to Go
- Go decodes in pure Go: 90μs (no CGo overhead)
- **Total overhead: <1%** (CGo called only once)

Since the C++ encoder is always used in production, there's no benefit to maintaining a separate pure Swift implementation.

## Generated Structure

When you run `sdp-gen -lang=swift`, it generates:

```
your_package/
├── Package.swift                          # SPM manifest with C++ interop
├── Sources/
│   └── your_package/
│       ├── encode.cpp                     # C++ encoder implementation
│       ├── decode.cpp                     # C++ decoder implementation
│       └── include/
│           ├── types.hpp                  # C++ type definitions
│           ├── encode.hpp                 # Encoder header
│           ├── decode.hpp                 # Decoder header
│           ├── endian.hpp                 # Endian utilities
│           └── module.modulemap           # Exposes C++ to Swift
```

## How It Works

### 1. Module Map (module.modulemap)

Declares the C++ headers as a Swift module:

```cpp
module your_package {
    header "types.hpp"
    header "encode.hpp"
    header "decode.hpp"
    requires cplusplus
    export *
}
```

### 2. Package.swift

Configured for C++ interoperability:

```swift
let package = Package(
    name: "your_package",
    targets: [
        .target(
            name: "your_package",
            cxxSettings: [
                .unsafeFlags(["-std=c++17"]),
            ],
            swiftSettings: [
                .interoperabilityMode(.Cxx),
                .unsafeFlags(["-O", "-whole-module-optimization"]),
            ]),
    ]
)
```

### 3. Swift Usage

Swift code can directly call C++ functions via namespace:

```swift
import your_package

// Decode from bytes
let registry = sdpbData.withUnsafeBytes { bufferPtr in
    sdp.plugin_registry_decode(
        bufferPtr.baseAddress!.assumingMemoryBound(to: UInt8.self),
        bufferPtr.count
    )
}

// Access C++ struct members
print("Plugins: \(registry.plugins.size())")

// Encode back to bytes
var outputBuf = [UInt8](repeating: 0, count: 200000)
let encodedSize = outputBuf.withUnsafeMutableBytes { bufPtr in
    sdp.plugin_registry_encode(
        registry,
        bufPtr.baseAddress!.assumingMemoryBound(to: UInt8.self)
    )
}
```

## Benefits

1. **Performance**: Direct C++ calls with minimal wrapper overhead (55% decode, -11% encode vs pure C++)
2. **Maintenance**: No separate Swift encoder/decoder to maintain
3. **Consistency**: Identical encoding/decoding logic across all languages
4. **Type Safety**: Swift's strong typing + C++'s type system
5. **Zero Runtime Dependencies**: Uses C++ stdlib only
6. **Cross-Platform**: Works on macOS, iOS, Linux (Swift 5.9+)

## Requirements

- Swift 5.9 or later (for C++ interoperability)
- C++17 compiler (clang on macOS/iOS, gcc on Linux)

## Migration from Pure Swift

If you were using the old pure Swift code generator:

### Before (Pure Swift)
```bash
sdp-gen -schema device.sdp -output swift/ -lang swift
```

Generated: `Types.swift`, `Encode.swift`, `Decode.swift`

### After (C++ Backend)
```bash
sdp-gen -schema device.sdp -output swift/ -lang swift
```

Generated: C++ files + Swift package wrappers

The API from Swift's perspective remains the same - you call the same functions, they just happen to be implemented in C++ under the hood.

## For macOS/iOS Native Apps

If you're building a native macOS or iOS app that needs to collect system data and encode it:

1. **Use ObjC++ bridge** (recommended, <1% overhead):
   - Create `bridge.mm` file
   - Query Apple APIs (AudioUnit, etc.)
   - Build C++ structs directly: `plugin.name = [nsString UTF8String]`
   - Encode with C++ encoder
   - Return bytes to Go (if Go orchestrator) or use directly

2. **OR use Swift package** (if you prefer pure Swift app):
   - Include the generated Swift package
   - Call C++ functions via module.modulemap
   - 55% decode overhead, but simpler integration for Swift-only apps

## See Also

- `macos_testing/COMPLETE_GUIDE.md` - Full ObjC++ bridge implementation
- `macos_testing/TEST_RESULTS.md` - Performance benchmark results
- `LANGUAGE_IMPLEMENTATION_GUIDE.md` - Cross-language architecture
