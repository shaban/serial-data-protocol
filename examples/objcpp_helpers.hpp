// objcpp_helpers.hpp
// Helper functions for Objective-C++ / C++ type conversions
// Part of Serial Data Protocol (SDP)
//
// Usage:
//   #include "objcpp_helpers.hpp"
//
//   NSString *str = @"Hello";
//   std::string cppStr = sdp::toString(str);
//
//   std::vector<uint8_t> vec = {1, 2, 3};
//   NSData *data = sdp::toNSData(vec);

#pragma once

#import <Foundation/Foundation.h>
#include <string>
#include <vector>

namespace sdp {

// ==============================================================================
// NSString ↔ std::string
// ==============================================================================

/// Convert NSString to std::string (UTF-8)
inline std::string toString(NSString *nsString) {
    if (!nsString) {
        return std::string();
    }
    return std::string([nsString UTF8String]);
}

/// Convert std::string to NSString (UTF-8)
inline NSString *toNSString(const std::string &str) {
    return [NSString stringWithUTF8String:str.c_str()];
}

// ==============================================================================
// NSData ↔ std::vector<uint8_t>
// ==============================================================================

/// Convert NSData to std::vector<uint8_t>
inline std::vector<uint8_t> toVector(NSData *data) {
    if (!data) {
        return std::vector<uint8_t>();
    }
    const uint8_t *bytes = static_cast<const uint8_t *>(data.bytes);
    return std::vector<uint8_t>(bytes, bytes + data.length);
}

/// Convert std::vector<uint8_t> to NSData (copies data)
inline NSData *toNSData(const std::vector<uint8_t> &vec) {
    return [NSData dataWithBytes:vec.data() length:vec.size()];
}

/// Convert std::vector<uint8_t> to NSData (no copy, caller manages lifetime)
/// WARNING: The vector must outlive the returned NSData!
inline NSData *toNSDataNoCopy(std::vector<uint8_t> &vec) {
    return [NSData dataWithBytesNoCopy:vec.data() 
                                length:vec.size() 
                          freeWhenDone:NO];
}

// ==============================================================================
// NSArray ↔ std::vector (for strings)
// ==============================================================================

/// Convert NSArray<NSString *> to std::vector<std::string>
inline std::vector<std::string> toStringVector(NSArray<NSString *> *array) {
    std::vector<std::string> vec;
    if (!array) {
        return vec;
    }
    
    vec.reserve(array.count);
    for (NSString *str in array) {
        vec.push_back(toString(str));
    }
    return vec;
}

/// Convert std::vector<std::string> to NSArray<NSString *>
inline NSArray<NSString *> *toNSArray(const std::vector<std::string> &vec) {
    NSMutableArray<NSString *> *array = [NSMutableArray arrayWithCapacity:vec.size()];
    for (const auto &str : vec) {
        [array addObject:toNSString(str)];
    }
    return array;
}

// ==============================================================================
// Numeric type helpers
// ==============================================================================

/// Safe conversion from NSUInteger to uint32_t with range check
inline uint32_t toUInt32(NSUInteger value) {
    if (value > UINT32_MAX) {
        NSLog(@"Warning: NSUInteger %lu exceeds uint32_t range, truncating", (unsigned long)value);
        return UINT32_MAX;
    }
    return static_cast<uint32_t>(value);
}

/// Safe conversion from NSInteger to int32_t with range check
inline int32_t toInt32(NSInteger value) {
    if (value > INT32_MAX) {
        NSLog(@"Warning: NSInteger %ld exceeds int32_t max, truncating", (long)value);
        return INT32_MAX;
    }
    if (value < INT32_MIN) {
        NSLog(@"Warning: NSInteger %ld exceeds int32_t min, truncating", (long)value);
        return INT32_MIN;
    }
    return static_cast<int32_t>(value);
}

/// Convert BOOL to bool
inline bool toBool(BOOL value) {
    return value ? true : false;
}

/// Convert bool to BOOL
inline BOOL toBOOL(bool value) {
    return value ? YES : NO;
}

// ==============================================================================
// Optional field helpers
// ==============================================================================

/// Check if optional field is present (non-null pointer)
template<typename T>
inline bool isPresent(const T *ptr) {
    return ptr != nullptr;
}

/// Get value from optional field or default
template<typename T>
inline T valueOr(const T *ptr, const T &defaultValue) {
    return ptr ? *ptr : defaultValue;
}

// ==============================================================================
// Error handling helpers
// ==============================================================================

/// Create NSError for SDP decoding failures
inline NSError *makeDecodeError(NSString *description) {
    return [NSError errorWithDomain:@"SDPErrorDomain"
                               code:1
                           userInfo:@{NSLocalizedDescriptionKey: description}];
}

/// Create NSError for SDP encoding failures
inline NSError *makeEncodeError(NSString *description) {
    return [NSError errorWithDomain:@"SDPErrorDomain"
                               code:2
                           userInfo:@{NSLocalizedDescriptionKey: description}];
}

/// Create NSError for validation failures
inline NSError *makeValidationError(NSString *description) {
    return [NSError errorWithDomain:@"SDPErrorDomain"
                               code:3
                           userInfo:@{NSLocalizedDescriptionKey: description}];
}

// ==============================================================================
// Validation helpers
// ==============================================================================

/// Validate data size is within SDP limits (128 MB)
inline bool isValidDataSize(NSData *data) {
    static const size_t kMaxSerializedSize = 128 * 1024 * 1024; // 128 MB
    return data && data.length <= kMaxSerializedSize;
}

/// Validate string length is within SDP limits (10 MB)
inline bool isValidStringLength(NSString *str) {
    if (!str) return true; // nil is valid
    static const size_t kMaxStringSize = 10 * 1024 * 1024; // 10 MB
    return str.length <= kMaxStringSize;
}

/// Validate array count is within SDP limits (100,000)
inline bool isValidArrayCount(NSArray *array) {
    if (!array) return true; // nil is valid
    static const size_t kMaxArraySize = 100000;
    return array.count <= kMaxArraySize;
}

// ==============================================================================
// Performance helpers
// ==============================================================================

/// Pre-allocate vector with capacity
inline std::vector<uint8_t> makeBuffer(size_t capacity) {
    std::vector<uint8_t> buffer;
    buffer.reserve(capacity);
    return buffer;
}

/// Reusable buffer class for reducing allocations
class Buffer {
public:
    Buffer() = default;
    
    /// Get writable vector (cleared and ready for use)
    std::vector<uint8_t> &get() {
        buffer_.clear();
        return buffer_;
    }
    
    /// Reserve capacity
    void reserve(size_t capacity) {
        buffer_.reserve(capacity);
    }
    
    /// Get current capacity
    size_t capacity() const {
        return buffer_.capacity();
    }
    
private:
    std::vector<uint8_t> buffer_;
};

} // namespace sdp

// ==============================================================================
// Example usage patterns
// ==============================================================================

/*

// Basic conversions
NSString *nsStr = @"Hello";
std::string cppStr = sdp::toString(nsStr);
NSString *backToNS = sdp::toNSString(cppStr);

// Data conversions
NSData *data = ...; 
std::vector<uint8_t> vec = sdp::toVector(data);
NSData *backToData = sdp::toNSData(vec);

// Array conversions
NSArray<NSString *> *strings = @[@"a", @"b", @"c"];
std::vector<std::string> cppStrings = sdp::toStringVector(strings);

// Safe numeric conversions
NSUInteger bigNumber = 12345;
uint32_t u32 = sdp::toUInt32(bigNumber);

// Validation
if (sdp::isValidDataSize(data)) {
    // Process data
}

// Error handling
NSError *error = sdp::makeDecodeError(@"Invalid data format");

// Reusable buffer for performance
sdp::Buffer buffer;
buffer.reserve(1024);

std::vector<uint8_t> &vec1 = buffer.get();
// Use vec1...

std::vector<uint8_t> &vec2 = buffer.get();
// Use vec2 (vec1 is now cleared)

*/
