# Objective-C++ Bridge Test

## What This Tests

This test validates that we can use the C++ SDP implementation from Objective-C/Objective-C++ code on macOS/iOS.

## Files

- `SDPAudioUnit.h` - Objective-C interface (can be imported in Swift pre-6.0)
- `SDPAudioUnit.mm` - Objective-C++ implementation (bridges to C++)
- `test_objcpp.mm` - Test program with benchmarks

## How It Works

1. `.mm` files can mix Objective-C and C++ code
2. C++ `std::vector<uint8_t>` ↔ `NSData`
3. C++ `std::string` ↔ `NSString`
4. C++ structs ↔ Objective-C objects

## Build & Run

```bash
# Compile the test
clang++ -std=c++17 -ObjC++ -O3 -framework Foundation \
  test_objcpp.mm SDPAudioUnit.mm \
  ../../testdata/audiounit_cpp/encode.cpp \
  ../../testdata/audiounit_cpp/decode.cpp \
  -I../../testdata/audiounit_cpp \
  -o test_objcpp

# Run it
./test_objcpp
```

## Expected Results

✓ All tests pass (decode, encode, roundtrip)  
✓ Overhead < 10% vs pure C++ (conversion cost acceptable)  
✓ Clean Objective-C API for macOS/iOS apps

## If It Works

This proves we can:
- Use C++ implementation from Cocoa/UIKit apps
- Wrap it in clean Objective-C interfaces
- Accept ~5-10% overhead for convenience
- Eliminate need for separate C implementation
