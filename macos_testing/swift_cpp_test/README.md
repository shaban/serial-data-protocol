# Swift 6 C++ Interop Test

## What This Tests

This test validates Swift 6's native C++ interoperability feature - directly calling C++ code from Swift without Objective-C bridges.

## Features Tested

- **Direct C++ Import**: `import SDPAudioUnitCpp` imports C++ headers
- **C++ Type Conversion**: `std::vector<uint8_t>` ↔ Swift `[UInt8]`
- **C++ Strings**: `std::string` ↔ Swift `String`
- **Value Semantics**: Swift structs wrapping C++ data
- **Performance**: Measuring overhead of Swift → C++ → Swift conversion

## Requirements

- Swift 6.0+
- macOS 13+
- Xcode 15+ (for Swift 6 support)

## How It Works

1. Swift 6's `.interoperabilityMode(.Cxx)` enables C++ interop
2. C++ functions are directly callable from Swift
3. Swift types convert to/from C++ types automatically (where supported)
4. Manual conversion for complex types (std::vector → Array)

## Build & Run

```bash
# Build the package
swift build -c release

# Run the test
swift run -c release test-swift-cpp
```

## Expected Results

✓ All tests pass (decode, encode, roundtrip)  
✓ Overhead < 10% vs pure C++ (conversion cost acceptable)  
✓ Clean Swift API (value types, error handling, Sendable)  
✓ No Objective-C bridge needed

## If It Works

This proves we can:
- Use C++ implementation directly from Swift 6+
- Provide idiomatic Swift API (structs, enums, errors)
- Avoid writing separate pure Swift implementation
- Accept ~5-10% overhead for Swift convenience
- Skip Objective-C bridge layer entirely

## If It Doesn't Work

Fallback options:
1. Use Objective-C++ bridge (objcpp_test/)
2. Keep pure Swift implementation (slower but native)
3. Wait for Swift 7+ improved C++ interop

## Known Limitations

Swift 6 C++ interop doesn't fully support:
- C++ templates (partial support)
- C++ exceptions (can catch, limited throw)
- C++ move semantics (copies instead)
- Some STL types

Our C++ code uses: std::vector, std::string, basic structs - should all work!
