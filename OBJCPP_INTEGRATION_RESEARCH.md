# Objective-C++ Integration Research

**Date:** October 21, 2025  
**Status:** Research Phase  
**Goal:** Streamline SDP usage in Objective-C++ projects without unnecessary wrappers

---

## Current State Analysis

### ✅ What Works Today

1. **C++ Implementation** (`testdata/cpp/*/`)
   - Complete encoder/decoder in C++17
   - Header-only design: `types.hpp`, `encode.hpp`, `decode.hpp`, `endian.hpp`
   - Zero dependencies (stdlib only)
   - Performance: ~30μs for AudioUnit (1,759 params)

2. **Objective-C++ Bridge** (`macos_testing/objcpp_test/`)
   - **Proof-of-concept:** Working Objective-C wrapper
   - Files: `SDPAudioUnit.h/.mm` (only for AudioUnit schema)
   - Conversion: C++ ↔ NSString, std::vector ↔ NSData
   - Overhead: ~5-10% (acceptable)
   - **Problem:** Manual wrapper per schema, not generated

3. **Swift Integration** (Production-ready)
   - Swift Package Manager packages
   - Auto-generated via `sdp-gen --lang swift`
   - C++ interop via module.modulemap
   - Performance: Similar to pure C++

### ❌ What's Missing for Objective-C++

1. **No Package Manager Integration**
   - CocoaPods: No `.podspec`
   - Carthage: No `Cartfile` support
   - SPM: No Objective-C++ specific package

2. **No Generated Wrappers**
   - Manual wrapper creation required
   - Each schema needs hand-written .h/.mm files
   - Error-prone type conversions

3. **No Distribution Method**
   - Can't `pod install sdp`
   - Can't add as Xcode framework
   - Must copy files manually

4. **No Documentation**
   - No "Using SDP from Objective-C++" guide
   - No best practices documented
   - No example Xcode project

---

## Design Philosophy: Direct C++ Usage vs Wrapper

### Option 1: Direct C++ Usage (RECOMMENDED)

**Approach:** Use C++ implementation directly in `.mm` files, no wrapper layer

```objectivec++
// MyViewController.mm
#import "MyViewController.h"
#include "sdp/primitives/encode.hpp"
#include "sdp/primitives/decode.hpp"
#include "sdp/primitives/types.hpp"

@implementation MyViewController

- (NSData *)serializeData:(MyData *)data {
    // Direct C++ struct usage
    sdp::Primitives cppData{
        .field1 = data.field1,
        .field2 = [data.field2 UTF8String],
        // ...
    };
    
    // Encode to vector
    std::vector<uint8_t> bytes = sdp::EncodePrimitives(cppData);
    
    // Convert to NSData
    return [NSData dataWithBytes:bytes.data() length:bytes.size()];
}

@end
```

**Pros:**
- ✅ Zero overhead (no wrapper layer)
- ✅ No code generation needed for wrappers
- ✅ Direct access to C++ performance
- ✅ Simpler: Just include C++ headers

**Cons:**
- ❌ Must write conversions manually
- ❌ Requires Objective-C++ knowledge
- ❌ Can't use in pure Objective-C files

**Verdict:** **Use this for performance-critical code or when you're comfortable with C++**

---

### Option 2: Generated Objective-C Wrappers (ALTERNATIVE)

**Approach:** Generate `.h/.mm` wrappers that hide C++ internals

```objectivec
// MyViewController.m (pure Objective-C)
#import "MyViewController.h"
#import "SDPPrimitives.h"  // Generated wrapper

@implementation MyViewController

- (NSData *)serializeData:(MyData *)data {
    SDPPrimitives *sdpData = [[SDPPrimitives alloc] init];
    sdpData.field1 = data.field1;
    sdpData.field2 = data.field2;
    
    return [SDPPrimitivesCodec encode:sdpData error:nil];
}

@end
```

**Pros:**
- ✅ Pure Objective-C usage (no .mm files needed)
- ✅ Familiar NSObject-based API
- ✅ Automatic NSString/NSData conversions
- ✅ Can be used from Swift (pre-6.0)

**Cons:**
- ❌ 5-10% overhead from conversions
- ❌ Requires code generation
- ❌ More files to maintain
- ❌ Memory allocations for wrapper objects

**Verdict:** **Use this for ease-of-use or when you need pure Objective-C compatibility**

---

## Recommended Approach: Hybrid

### Core Principle: **No wrappers by default, provide packaging for direct usage**

**Why:**
1. Most macOS/iOS developers comfortable with Objective-C++ (just rename .m → .mm)
2. Zero overhead is a core SDP principle
3. Wrapper generation adds complexity without clear benefit
4. Swift is the future anyway (use Swift packages for new code)

**What we provide:**
1. ✅ **CocoaPods/SPM integration** for easy distribution
2. ✅ **Header organization** for clean imports
3. ✅ **Documentation** with conversion patterns
4. ✅ **Example project** showing best practices
5. ⚠️ **Optional:** Wrapper generator for those who want it

---

## Proposed Solution: CocoaPods + SPM Distribution

### 1. CocoaPods Podspec

Create `SerialDataProtocol.podspec`:

```ruby
Pod::Spec.new do |s|
  s.name             = 'SerialDataProtocol'
  s.version          = '0.2.0'
  s.summary          = 'High-performance binary serialization for macOS/iOS'
  s.homepage         = 'https://github.com/shaban/serial-data-protocol'
  s.license          = { :type => 'MIT', :file => 'LICENSE' }
  s.author           = { 'shaban' => 'your@email.com' }
  s.source           = { :git => 'https://github.com/shaban/serial-data-protocol.git', :tag => s.version.to_s }

  s.ios.deployment_target = '13.0'
  s.osx.deployment_target = '10.15'
  
  # Compiler settings for C++17
  s.pod_target_xcconfig = {
    'CLANG_CXX_LANGUAGE_STANDARD' => 'c++17',
    'CLANG_CXX_LIBRARY' => 'libc++'
  }
  
  # No source files - header only library
  # Users generate their own schemas with sdp-gen
  # This pod just provides the generator binary
  
  s.preserve_paths = 'cmd/sdp-gen/**/*'
  s.prepare_command = <<-CMD
    go build -o sdp-gen ./cmd/sdp-gen
  CMD
end
```

**Usage:**
```ruby
# Podfile
pod 'SerialDataProtocol'

# Generate code
pod exec sdp-gen --schema my_schema.sdp --output ./Generated --lang cpp

# Use in .mm files
#include "Generated/types.hpp"
#include "Generated/encode.hpp"
```

### 2. Swift Package Manager (for C++ headers)

Create `Package.swift` at root:

```swift
// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "SerialDataProtocol",
    platforms: [
        .macOS(.v10_15),
        .iOS(.v13)
    ],
    products: [
        // Generator binary
        .executable(
            name: "sdp-gen",
            targets: ["sdp-gen"]
        ),
        // Example C++ library (for reference)
        .library(
            name: "SDPExample",
            targets: ["SDPExample"]
        )
    ],
    targets: [
        // Generator (Go binary)
        .binaryTarget(
            name: "sdp-gen",
            path: "bin/sdp-gen"  // Pre-built binary
        ),
        
        // Example C++ target showing structure
        .target(
            name: "SDPExample",
            dependencies: [],
            path: "testdata/cpp/primitives",
            sources: ["encode.cpp", "decode.cpp"],
            publicHeadersPath: ".",
            cxxSettings: [
                .headerSearchPath("."),
                .define("SDP_CPP17")
            ]
        )
    ],
    cxxLanguageStandard: .cxx17
)
```

**Usage:**
```swift
// Package.swift in your project
dependencies: [
    .package(url: "https://github.com/shaban/serial-data-protocol", from: "0.2.0")
],
targets: [
    .target(
        name: "MyApp",
        dependencies: [
            .product(name: "sdp-gen", package: "serial-data-protocol")
        ]
    )
]
```

### 3. Xcode Framework (Manual)

For developers who don't use package managers:

```
SerialDataProtocol.xcframework/
├── ios-arm64/
│   └── SerialDataProtocol.framework/
│       ├── Headers/
│       │   ├── encode.hpp (template)
│       │   ├── decode.hpp (template)
│       │   └── types.hpp (template)
│       └── Modules/
│           └── module.modulemap
├── ios-arm64-simulator/
└── macos-arm64/
```

**Problem:** Framework contains templates, not actual schemas  
**Solution:** Users still need to run `sdp-gen` locally

---

## Documentation Needs

### 1. Quick Start Guide

```markdown
# Using SDP from Objective-C++

## Installation

### CocoaPods
1. Add to Podfile: `pod 'SerialDataProtocol'`
2. Run: `pod install`
3. Generate code: `pod exec sdp-gen --schema my.sdp --output Generated --lang cpp`

### Manual
1. Install sdp-gen: `go install github.com/shaban/serial-data-protocol/cmd/sdp-gen@latest`
2. Generate code: `sdp-gen --schema my.sdp --output Generated --lang cpp`
3. Add generated files to Xcode project

## Usage

### 1. Rename files .m → .mm
Objective-C++ requires `.mm` extension to mix C++ and Objective-C.

### 2. Include generated headers
```objectivec++
#include "Generated/types.hpp"
#include "Generated/encode.hpp"
#include "Generated/decode.hpp"
```

### 3. Use directly
```objectivec++
- (NSData *)encode:(MyData *)data {
    // Create C++ struct
    sdp::MySchema cppData{
        .field1 = data.field1,
        .field2 = [data.field2 UTF8String]
    };
    
    // Encode
    std::vector<uint8_t> bytes = sdp::EncodeMySchema(cppData);
    
    // Convert to NSData
    return [NSData dataWithBytes:bytes.data() length:bytes.size()];
}

- (MyData *)decode:(NSData *)data {
    // Convert from NSData
    std::vector<uint8_t> bytes((uint8_t*)data.bytes, 
                                (uint8_t*)data.bytes + data.length);
    
    // Decode
    sdp::MySchema cppData;
    if (!sdp::DecodeMySchema(cppData, bytes)) {
        return nil;
    }
    
    // Convert to Objective-C
    MyData *result = [[MyData alloc] init];
    result.field1 = cppData.field1;
    result.field2 = [NSString stringWithUTF8String:cppData.field2.c_str()];
    return result;
}
```
```

### 2. Conversion Patterns Reference

```markdown
# Type Conversion Patterns

## C++ → Objective-C

| C++ Type | Objective-C Type | Conversion |
|----------|------------------|------------|
| `uint32_t` | `NSUInteger` | Direct assignment |
| `std::string` | `NSString*` | `[NSString stringWithUTF8String:s.c_str()]` |
| `std::vector<T>` | `NSArray*` | Loop + convert each element |
| `std::vector<uint8_t>` | `NSData*` | `[NSData dataWithBytes:v.data() length:v.size()]` |
| `bool` | `BOOL` | Direct assignment |
| `float/double` | `CGFloat` | Direct assignment |

## Objective-C → C++

| Objective-C Type | C++ Type | Conversion |
|------------------|----------|------------|
| `NSString*` | `std::string` | `[str UTF8String]` |
| `NSData*` | `std::vector<uint8_t>` | `std::vector((uint8_t*)data.bytes, ...)` |
| `NSArray*` | `std::vector<T>` | Loop + convert each element |
| `BOOL` | `bool` | Direct assignment |

## Helper Functions

```objectivec++
// NSData ↔ std::vector<uint8_t>
std::vector<uint8_t> toVector(NSData *data) {
    return std::vector<uint8_t>((uint8_t*)data.bytes, 
                                 (uint8_t*)data.bytes + data.length);
}

NSData* toNSData(const std::vector<uint8_t> &vec) {
    return [NSData dataWithBytes:vec.data() length:vec.size()];
}

// NSString ↔ std::string
std::string toString(NSString *str) {
    return [str UTF8String];
}

NSString* toNSString(const std::string &str) {
    return [NSString stringWithUTF8String:str.c_str()];
}
```
```

### 3. Example Xcode Project

Create `examples/objcpp-example/`:
```
objcpp-example/
├── objcpp-example.xcodeproj
├── schema.sdp
├── Generated/          # Output from sdp-gen
│   ├── types.hpp
│   ├── encode.hpp
│   └── decode.hpp
├── AppDelegate.h
├── AppDelegate.mm      # Uses SDP
└── ViewController.mm   # Uses SDP
```

---

## Action Items

### Immediate (High Priority)

- [ ] **Create OBJCPP_QUICK_START.md**
  - Installation instructions (manual, CocoaPods, SPM)
  - Basic usage with .mm files
  - Type conversion reference
  - Common patterns

- [ ] **Add CocoaPods support**
  - Create `SerialDataProtocol.podspec`
  - Test `pod install` workflow
  - Publish to CocoaPods trunk

- [ ] **Create conversion helpers header**
  - `objcpp_helpers.hpp` with common conversions
  - NSData ↔ std::vector helpers
  - NSString ↔ std::string helpers
  - Place in `examples/objcpp-helpers/`

### Future (Optional)

- [ ] **Generate Objective-C wrappers (optional)**
  - Add `--lang objc` to sdp-gen
  - Generate .h/.mm files with NSObject wrappers
  - For users who want pure Objective-C

- [ ] **Create XCFramework distribution**
  - Bundle sdp-gen binary
  - Include example schemas
  - Provide build script

- [ ] **Add to Carthage**
  - Create Cartfile
  - Test Carthage build

---

## Decision: What to Build

### ✅ Build This (No Wrappers)

1. **Documentation** (`OBJCPP_QUICK_START.md`)
   - How to use C++ directly from .mm files
   - Type conversion patterns
   - Best practices
   - Example code

2. **Helper Header** (`examples/objcpp-helpers.hpp`)
   - Common type conversions
   - Utility functions
   - Can be copied into projects

3. **Example Project** (`examples/objcpp-example/`)
   - Working Xcode project
   - Shows real-world usage
   - Can be used as template

4. **Package Manager Support**
   - CocoaPods: Distribute sdp-gen binary
   - SPM: Distribute sdp-gen binary
   - Both: Users generate their own code

### ❌ Don't Build This (Yet)

1. **Objective-C Wrapper Generator**
   - Adds complexity
   - Performance overhead
   - Most users fine with .mm files
   - Can add later if demand exists

2. **XCFramework**
   - Can't pre-generate schemas
   - Users need sdp-gen anyway
   - More complexity than value

---

## Summary

**Philosophy:** Provide excellent tooling for direct C++ usage, not abstraction layers

**What users get:**
- ✅ Clear documentation on .mm file usage
- ✅ Helper functions for type conversions
- ✅ Example Xcode project to copy
- ✅ Easy installation via CocoaPods/SPM
- ✅ Zero performance overhead

**What users don't get (intentionally):**
- ❌ Auto-generated Objective-C wrappers (can add if needed)
- ❌ Pre-compiled schemas (defeats purpose)
- ❌ Heavy framework overhead

**Result:** Fast, simple, and idiomatic Objective-C++ integration that respects SDP's "no dependencies, maximum performance" philosophy.
