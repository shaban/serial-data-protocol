# Message Mode Testing Refactoring Summary

**Date:** October 22, 2025  
**Status:** ✅ Complete

## Problem

Message mode cross-language tests were using **hardcoded hex strings** embedded in test files:

```cpp
// Old approach - PROBLEMATIC
const uint8_t go_point_bytes[] = {
    0x53, 0x44, 0x50, 0x32, 0x01, 0x00, 0x10, 0x00, ...
};
```

**Issues:**
- ❌ Not reusable across languages
- ❌ Hard to maintain (update hex in every file on format change)
- ❌ Limited coverage (only tested simple Point/Rectangle)
- ❌ Inconsistent with project's byte mode testing pattern
- ❌ No version control for test data evolution

## Solution

Refactored to use **`.sdpb` binary reference files** stored in `testdata/binaries/`:

```cpp
// New approach - CLEAN
std::vector<uint8_t> data = read_file("../../binaries/message_point.sdpb");
sdp::Point point = sdp::DecodePointMessage(data);
```

**Benefits:**
- ✅ Single source of truth in `testdata/binaries/`
- ✅ Reusable across Go/C++/Rust tests
- ✅ Easy to regenerate with encoder tools
- ✅ Version controlled for regression testing
- ✅ Matches established byte mode testing pattern (`primitives.sdpb`, `arrays.sdpb`, etc.)
- ✅ Can expand to complex schemas (arrays, nested, optional, audiounit)

## Implementation

### 1. Created Reference File Generators

**Go encoder tool:**
```bash
testdata/tools/encode_messagemode_reference.go
```
- Standalone Go program (no package conflicts)
- Generates `message_point.sdpb` and `message_rectangle.sdpb`
- Uses correct 3-byte magic `SDP` (not `SDP:`)

**C++ encoder tool:**
```bash
testdata/cpp/messagemode/encode_reference.cpp
```
- Uses generated C++ message mode code
- Generates `message_point_cpp.sdpb` and `message_rectangle_cpp.sdpb`
- Verifies C++ encoder produces identical output to Go

### 2. Generated Reference Files

All files stored in `testdata/binaries/`:

| File | Size | Description |
|------|------|-------------|
| `message_point.sdpb` | 26 bytes | Point{x: 3.14, y: 2.71} from Go |
| `message_point_cpp.sdpb` | 26 bytes | Same Point from C++ |
| `message_rectangle.sdpb` | 42 bytes | Rectangle{...} from Go |
| `message_rectangle_cpp.sdpb` | 42 bytes | Same Rectangle from C++ |

**Wire format verification:**
```bash
$ diff <(xxd message_point.sdpb) <(xxd message_point_cpp.sdpb)
✓ Identical (byte-for-byte)
```

### 3. Updated C++ Tests

**File:** `testdata/cpp/messagemode/test_crosslang.cpp`

**Changes:**
- ❌ Removed: 80 lines of hardcoded hex arrays
- ✅ Added: File I/O to load `.sdpb` files
- ✅ Added: 5 comprehensive tests:
  1. Decode Go Point message
  2. Decode Go Rectangle message  
  3. Verify C++ encoding matches Go reference
  4. Decode C++-generated message
  5. Message dispatcher with Go files

**Test output:**
```
=== C++/Go Cross-Language Message Mode Test ===
Using .sdpb reference files from testdata/binaries/

Test 1: Decode Go Point message
  Loaded 26 bytes
  Hex: 534450320100100000001f85eb51b81e0940ae47e17a14ae0540
  Decoded: Point(3.14, 2.71)
  ✓ Point decode OK

...

=== All cross-language tests passed! ===
```

### 4. Documentation

Created `testdata/binaries/MESSAGE_MODE_README.md` documenting:
- Wire format specification
- How to regenerate reference files
- Usage examples for C++/Go/Rust
- Cross-language compatibility guarantees

## Wire Format Details

**Message header (10 bytes):**
```
[SDP:3][version:1][type_id:2][length:4]
```

**Example - Point message (26 bytes total):**
```
Offset | Hex                              | Description
-------|----------------------------------|------------------
0-2    | 53 44 50                         | Magic "SDP"
3      | 32                               | Version '2' (ASCII)
4-5    | 01 00                            | Type ID 1 (Point, LE)
6-9    | 10 00 00 00                      | Payload len 16 (LE)
10-25  | 1f 85 eb 51 b8 1e 09 40 ...     | Payload (x, y: f64)
```

## Critical Bug Fix

During refactoring, discovered **Go standalone tool was using 4-byte magic** `SDP:` instead of 3-byte `SDP`:

```go
// ❌ Wrong (original)
const MessageMagic = "SDP:"  // 4 bytes
copy(buf[0:4], MessageMagic)

// ✅ Correct (fixed)
const MessageMagic = "SDP"   // 3 bytes
copy(buf[0:3], MessageMagic)
```

This would have caused **complete cross-language incompatibility**. Fixed before it became a problem.

## Next Steps

### For Rust Implementation

1. **Generate `.sdpb` files from Rust encoder** (add to testdata/binaries/)
2. **Rust tests load existing Go/C++ `.sdpb` files** (verify decode)
3. **Compare Rust-encoded output** to Go/C++ references (byte-for-byte)

Example:
```rust
let data = std::fs::read("../binaries/message_point.sdpb")?;
let point = decode_point_message(&data)?;
assert_eq!(point.x, 3.14);
```

### For Complex Schemas

Expand `.sdpb` reference files to cover:
- **arrays.sdp** → `message_devicelist.sdpb`
- **nested.sdp** → `message_pluginlist.sdpb`
- **optional.sdp** → `message_config.sdpb`
- **audiounit.sdp** → `message_audiounit.sdpb`

This provides **comprehensive cross-language verification** across all schema features.

## Testing Pattern Consistency

**Before (inconsistent):**
- Byte mode: ✅ Uses `.sdpb` files (primitives.sdpb, arrays.sdpb, etc.)
- Message mode: ❌ Uses hardcoded hex in test files

**After (consistent):**
- Byte mode: ✅ Uses `.sdpb` files
- Message mode: ✅ Uses `.sdpb` files

**Both modes now follow the same pattern:**
1. Schema → Generate code
2. Encode with known data → `.sdpb` files
3. Tests load `.sdpb` files → Verify decode
4. Compare encoded output → Byte-for-byte identical

## Files Changed

### Added
- `testdata/tools/encode_messagemode_reference.go` - Go encoder tool
- `testdata/cpp/messagemode/encode_reference.cpp` - C++ encoder tool
- `testdata/binaries/message_point.sdpb` - Go Point reference
- `testdata/binaries/message_rectangle.sdpb` - Go Rectangle reference
- `testdata/binaries/message_point_cpp.sdpb` - C++ Point reference
- `testdata/binaries/message_rectangle_cpp.sdpb` - C++ Rectangle reference
- `testdata/binaries/MESSAGE_MODE_README.md` - Documentation

### Modified
- `testdata/cpp/messagemode/test_crosslang.cpp` - Refactored to use `.sdpb` files

### Verified
- ✅ All tests pass (basic + cross-language)
- ✅ C++ and Go produce identical wire format
- ✅ Benchmarks still run correctly
- ✅ No regressions introduced

## Impact

**Code quality:** Cleaner, more maintainable tests  
**Reusability:** Same `.sdpb` files work across all languages  
**Consistency:** Matches project's established testing patterns  
**Future-proof:** Easy to expand to more complex schemas  
**Reliability:** Version-controlled test data prevents regressions  

**This refactoring sets the foundation for robust cross-language message mode testing across Go, C++, and Rust.**
