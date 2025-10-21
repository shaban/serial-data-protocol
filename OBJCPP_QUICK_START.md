# Objective-C++ Quick Start Guide

**Using Serial Data Protocol from Objective-C++ on macOS/iOS**

---

## Overview

SDP provides a **high-performance C++17 implementation** that works seamlessly with Objective-C++. No wrappers needed - just include the C++ headers in your `.mm` files and use directly.

**Benefits:**
- ✅ Zero overhead (direct C++ usage)
- ✅ Maximum performance (~30μs for complex structures)
- ✅ Simple: Just rename `.m` → `.mm` and include headers
- ✅ Full feature parity with other languages

**When to use this:**
- Native macOS/iOS apps using Objective-C
- Performance-critical serialization
- When you need direct control over data encoding

**Alternatives:**
- **Swift:** Use Swift packages (see SWIFT_CPP_ARCHITECTURE.md)
- **Pure Objective-C:** Use wrapper generator (coming soon)

---

## Installation

### Option 1: Manual (Recommended for Getting Started)

1. **Install sdp-gen:**
   ```bash
   go install github.com/shaban/serial-data-protocol/cmd/sdp-gen@latest
   ```

2. **Generate code for your schema:**
   ```bash
   sdp-gen --schema my_schema.sdp --output ./Generated --lang cpp
   ```

3. **Add to Xcode:**
   - Add `Generated/` folder to your Xcode project
   - Set C++ Language Dialect to "C++17" in Build Settings
   - Ensure files use `.mm` extension (Objective-C++)

### Option 2: CocoaPods (Coming Soon)

```ruby
# Podfile
pod 'SerialDataProtocol'

# Then:
pod install
pod exec sdp-gen --schema my.sdp --output Generated --lang cpp
```

### Option 3: Swift Package Manager (Coming Soon)

```swift
dependencies: [
    .package(url: "https://github.com/shaban/serial-data-protocol", from: "0.2.0")
]
```

---

## Quick Example

### 1. Define Your Schema

```rust
// person.sdp
struct Person {
    id: u32,
    name: string,
    email: string,
    age: u8,
}
```

### 2. Generate C++ Code

```bash
sdp-gen --schema person.sdp --output Generated --lang cpp
```

**Generated files:**
- `Generated/types.hpp` - Struct definitions
- `Generated/encode.hpp` - Encoding functions
- `Generated/decode.hpp` - Decoding functions
- `Generated/endian.hpp` - Endian utilities (internal)

### 3. Use in Objective-C++

**Rename your file:** `MyClass.m` → `MyClass.mm`

```objectivec++
// MyClass.mm
#import "MyClass.h"

// Include generated SDP headers
#include "Generated/types.hpp"
#include "Generated/encode.hpp"
#include "Generated/decode.hpp"

// Standard library
#include <vector>

@implementation MyClass

- (NSData *)serializePerson:(NSString *)name email:(NSString *)email age:(NSUInteger)age {
    // Create C++ struct
    sdp::Person person{
        .id = 12345,
        .name = std::string([name UTF8String]),
        .email = std::string([email UTF8String]),
        .age = (uint8_t)age
    };
    
    // Encode to bytes
    std::vector<uint8_t> bytes = sdp::EncodePerson(person);
    
    // Convert to NSData
    return [NSData dataWithBytes:bytes.data() length:bytes.size()];
}

- (void)deserializePerson:(NSData *)data {
    // Convert NSData to vector
    std::vector<uint8_t> bytes((uint8_t*)data.bytes, 
                                (uint8_t*)data.bytes + data.length);
    
    // Decode
    sdp::Person person;
    if (!sdp::DecodePerson(person, bytes)) {
        NSLog(@"Failed to decode person");
        return;
    }
    
    // Use the data
    NSLog(@"ID: %u", person.id);
    NSLog(@"Name: %s", person.name.c_str());
    NSLog(@"Email: %s", person.email.c_str());
    NSLog(@"Age: %u", person.age);
}

@end
```

---

## Type Conversion Reference

### C++ to Objective-C

| C++ Type | Objective-C Type | Conversion Code |
|----------|------------------|-----------------|
| `uint8_t`, `uint16_t`, `uint32_t` | `NSUInteger` | Direct assignment |
| `int8_t`, `int16_t`, `int32_t` | `NSInteger` | Direct assignment |
| `uint64_t` | `unsigned long long` | Direct assignment |
| `int64_t` | `long long` | Direct assignment |
| `float` | `CGFloat` | Direct assignment |
| `double` | `double` | Direct assignment |
| `bool` | `BOOL` | Direct assignment (or `? YES : NO`) |
| `std::string` | `NSString*` | `[NSString stringWithUTF8String:str.c_str()]` |
| `std::vector<uint8_t>` | `NSData*` | `[NSData dataWithBytes:vec.data() length:vec.size()]` |
| `std::vector<T>` | `NSArray*` | Loop and convert each element |

### Objective-C to C++

| Objective-C Type | C++ Type | Conversion Code |
|------------------|----------|-----------------|
| `NSUInteger` | `uint32_t` | Direct cast: `(uint32_t)value` |
| `NSInteger` | `int32_t` | Direct cast: `(int32_t)value` |
| `BOOL` | `bool` | Direct assignment |
| `CGFloat` | `float` | Direct cast: `(float)value` |
| `NSString*` | `std::string` | `std::string([str UTF8String])` |
| `NSData*` | `std::vector<uint8_t>` | `std::vector((uint8_t*)data.bytes, (uint8_t*)data.bytes + data.length)` |
| `NSArray*` | `std::vector<T>` | Loop and convert each element |

---

## Common Patterns

### Pattern 1: Encoding Objective-C Objects

```objectivec++
@interface MyDataManager : NSObject
- (NSData *)encodeUser:(User *)user;
@end

@implementation MyDataManager

- (NSData *)encodeUser:(User *)user {
    // Convert Objective-C object to C++ struct
    sdp::UserData cppUser{
        .id = (uint32_t)user.userId,
        .username = std::string([user.username UTF8String]),
        .email = std::string([user.email UTF8String]),
        .active = (bool)user.isActive
    };
    
    // Encode
    std::vector<uint8_t> bytes = sdp::EncodeUserData(cppUser);
    
    // Return as NSData
    return [NSData dataWithBytes:bytes.data() length:bytes.size()];
}

@end
```

### Pattern 2: Decoding to Objective-C Objects

```objectivec++
- (User *)decodeUser:(NSData *)data {
    // Convert NSData to vector
    std::vector<uint8_t> bytes((uint8_t*)data.bytes, 
                                (uint8_t*)data.bytes + data.length);
    
    // Decode C++ struct
    sdp::UserData cppUser;
    if (!sdp::DecodeUserData(cppUser, bytes)) {
        return nil; // Decoding failed
    }
    
    // Convert to Objective-C object
    User *user = [[User alloc] init];
    user.userId = cppUser.id;
    user.username = [NSString stringWithUTF8String:cppUser.username.c_str()];
    user.email = [NSString stringWithUTF8String:cppUser.email.c_str()];
    user.isActive = cppUser.active;
    
    return user;
}
```

### Pattern 3: Handling Arrays

```objectivec++
// Schema: struct Users { users: []User }

- (NSData *)encodeUsers:(NSArray<User *> *)users {
    // Convert NSArray to std::vector
    std::vector<sdp::UserData> cppUsers;
    cppUsers.reserve(users.count);
    
    for (User *user in users) {
        cppUsers.push_back({
            .id = (uint32_t)user.userId,
            .username = std::string([user.username UTF8String]),
            .email = std::string([user.email UTF8String]),
            .active = (bool)user.isActive
        });
    }
    
    // Create wrapper struct
    sdp::Users usersData{
        .users = cppUsers
    };
    
    // Encode
    std::vector<uint8_t> bytes = sdp::EncodeUsers(usersData);
    return [NSData dataWithBytes:bytes.data() length:bytes.size()];
}

- (NSArray<User *> *)decodeUsers:(NSData *)data {
    std::vector<uint8_t> bytes((uint8_t*)data.bytes, 
                                (uint8_t*)data.bytes + data.length);
    
    sdp::Users usersData;
    if (!sdp::DecodeUsers(usersData, bytes)) {
        return nil;
    }
    
    // Convert std::vector to NSArray
    NSMutableArray<User *> *users = [NSMutableArray arrayWithCapacity:usersData.users.size()];
    
    for (const auto& cppUser : usersData.users) {
        User *user = [[User alloc] init];
        user.userId = cppUser.id;
        user.username = [NSString stringWithUTF8String:cppUser.username.c_str()];
        user.email = [NSString stringWithUTF8String:cppUser.email.c_str()];
        user.isActive = cppUser.active;
        [users addObject:user];
    }
    
    return users;
}
```

### Pattern 4: Using Helper Functions

For cleaner code, create helper functions (see `examples/objcpp-helpers.hpp`):

```objectivec++
#include "objcpp-helpers.hpp"

- (NSData *)encodeUser:(User *)user {
    sdp::UserData cppUser{
        .id = (uint32_t)user.userId,
        .username = sdp::toString(user.username),  // Helper
        .email = sdp::toString(user.email),        // Helper
        .active = (bool)user.isActive
    };
    
    std::vector<uint8_t> bytes = sdp::EncodeUserData(cppUser);
    return sdp::toNSData(bytes);  // Helper
}
```

---

## Xcode Configuration

### Build Settings

1. **C++ Language Dialect:** C++17 (`-std=c++17`)
2. **C++ Standard Library:** libc++ (LLVM)
3. **Enable Objective-C++:** Automatic (when using `.mm` files)

### Header Search Paths

Add the path to your generated headers:
```
$(PROJECT_DIR)/Generated
```

### Common Issues

**Issue:** "Unknown type name 'string'"
- **Fix:** Add `#include <string>` at top of file

**Issue:** "Use of undeclared identifier 'vector'"
- **Fix:** Add `#include <vector>` at top of file

**Issue:** "C++ headers not found"
- **Fix:** Check Header Search Paths in Build Settings

**Issue:** "Duplicate symbols" when linking
- **Fix:** Ensure `.cpp` files are only in one target

---

## Performance Tips

### 1. Pre-allocate Buffers

```objectivec++
// If you know the size in advance
size_t expectedSize = sdp::EncodedSize(myData);
std::vector<uint8_t> buffer;
buffer.reserve(expectedSize);
```

### 2. Reuse Vectors

```objectivec++
@implementation MySerializer {
    std::vector<uint8_t> _reusableBuffer;
}

- (NSData *)encode:(MyData *)data {
    _reusableBuffer.clear();
    // Use _reusableBuffer for encoding
    // ...
}
@end
```

### 3. Avoid Unnecessary Copies

```objectivec++
// Bad - copies data
NSData *data = [NSData dataWithBytes:vec.data() length:vec.size()];

// Better - use bytesNoCopy for large data (be careful with lifetime!)
NSData *data = [NSData dataWithBytesNoCopy:vec.data() 
                                    length:vec.size() 
                              freeWhenDone:NO];
// But ensure vec outlives data!
```

### 4. Batch Operations

```objectivec++
// Encode multiple items at once instead of one-by-one
std::vector<sdp::Item> items;
items.reserve(array.count);

for (id obj in array) {
    items.push_back(convertToItem(obj));
}

// Single encode call
std::vector<uint8_t> bytes = sdp::EncodeItems(items);
```

---

## Error Handling

### Decoding Errors

```objectivec++
- (nullable User *)decodeUser:(NSData *)data error:(NSError **)error {
    std::vector<uint8_t> bytes((uint8_t*)data.bytes, 
                                (uint8_t*)data.bytes + data.length);
    
    sdp::UserData cppUser;
    if (!sdp::DecodeUserData(cppUser, bytes)) {
        if (error) {
            *error = [NSError errorWithDomain:@"SDPErrorDomain"
                                         code:1
                                     userInfo:@{NSLocalizedDescriptionKey: @"Failed to decode user data"}];
        }
        return nil;
    }
    
    // Convert to Objective-C...
}
```

### Size Validation

```objectivec++
- (BOOL)validateData:(NSData *)data {
    // SDP has maximum size limits
    if (data.length > 128 * 1024 * 1024) { // 128MB
        NSLog(@"Data too large: %lu bytes", (unsigned long)data.length);
        return NO;
    }
    
    if (data.length < 4) { // Minimum valid size
        NSLog(@"Data too small: %lu bytes", (unsigned long)data.length);
        return NO;
    }
    
    return YES;
}
```

---

## Testing

### Unit Testing with XCTest

```objectivec++
// MySerializerTests.mm
#import <XCTest/XCTest.h>
#import "MySerializer.h"

@interface MySerializerTests : XCTestCase
@property (nonatomic, strong) MySerializer *serializer;
@end

@implementation MySerializerTests

- (void)setUp {
    self.serializer = [[MySerializer alloc] init];
}

- (void)testEncodeDecode {
    // Create test data
    User *user = [[User alloc] init];
    user.userId = 123;
    user.username = @"test_user";
    user.email = @"test@example.com";
    user.isActive = YES;
    
    // Encode
    NSData *encoded = [self.serializer encodeUser:user];
    XCTAssertNotNil(encoded);
    XCTAssertGreaterThan(encoded.length, 0);
    
    // Decode
    User *decoded = [self.serializer decodeUser:encoded];
    XCTAssertNotNil(decoded);
    XCTAssertEqual(decoded.userId, user.userId);
    XCTAssertEqualObjects(decoded.username, user.username);
    XCTAssertEqualObjects(decoded.email, user.email);
    XCTAssertEqual(decoded.isActive, user.isActive);
}

- (void)testInvalidData {
    NSData *invalidData = [@"invalid" dataUsingEncoding:NSUTF8StringEncoding];
    User *user = [self.serializer decodeUser:invalidData];
    XCTAssertNil(user);
}

@end
```

---

## Example Project

See `examples/objcpp-example/` for a complete working Xcode project showing:
- Project structure
- Build settings
- Serializer class implementation
- Unit tests
- Performance benchmarks

---

## Next Steps

1. **Read the Design Spec:** See `DESIGN_SPEC.md` for wire format details
2. **Check Examples:** See `macos_testing/objcpp_test/` for real-world usage
3. **Performance Analysis:** See `PERFORMANCE_ANALYSIS.md` for benchmarks
4. **Swift Alternative:** See `SWIFT_CPP_ARCHITECTURE.md` for Swift interop

---

## Troubleshooting

### Common Errors

**"Cannot initialize a variable of type 'NSString *' with an rvalue of type 'const char *'"**
- Use `[NSString stringWithUTF8String:...]` not direct assignment

**"No matching function for call to 'DecodeXXX'"**
- Check you're passing a reference: `DecodeXXX(myStruct, bytes)` not `DecodeXXX(&myStruct, bytes)`

**"Undefined symbols for architecture arm64"**
- Ensure `.cpp` implementation files are added to your target
- Check C++ Standard Library is set to libc++

**"Linker command failed with exit code 1"**
- Make sure all generated `.cpp` files are included in your target
- Verify no duplicate symbol definitions

### Getting Help

- **Documentation:** Check `DESIGN_SPEC.md`, `TESTING_STRATEGY.md`
- **Examples:** See `macos_testing/objcpp_test/` and `examples/objcpp-example/`
- **Issues:** https://github.com/shaban/serial-data-protocol/issues

---

## Summary

**Using SDP from Objective-C++ is simple:**
1. Generate C++ code with `sdp-gen`
2. Rename `.m` files to `.mm`
3. Include generated headers
4. Use C++ structs directly with simple conversions

**Benefits:**
- Zero overhead
- Maximum performance
- Full language feature parity
- Simple, predictable behavior

**When you need more:**
- Consider Swift packages for new projects
- Use wrapper generator for pure Objective-C (coming soon)
