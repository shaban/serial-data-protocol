//
//  test_swift_bridge.swift
//  Swift test using C bridge to C++ (zero conversion overhead)
//
//  Compile:
//    swiftc -O test_swift_bridge.swift sdp_bridge.cpp \
//      ../../testdata/audiounit_cpp/encode.cpp \
//      ../../testdata/audiounit_cpp/decode.cpp \
//      -I../../testdata/audiounit_cpp \
//      -o test_swift_bridge
//

import Foundation

// Import C functions (Swift can import C directly)
// These declarations match sdp_bridge.h

@_silgen_name("sdp_bridge_decode")
func sdp_bridge_decode(_ data: UnsafePointer<UInt8>, _ len: Int) -> OpaquePointer?

@_silgen_name("sdp_bridge_encode")
func sdp_bridge_encode(_ reg: OpaquePointer, _ outLen: UnsafeMutablePointer<Int>) -> UnsafeMutablePointer<UInt8>?

@_silgen_name("sdp_bridge_free")
func sdp_bridge_free(_ reg: OpaquePointer)

@_silgen_name("sdp_bridge_total_plugins")
func sdp_bridge_total_plugins(_ reg: OpaquePointer) -> UInt32

@_silgen_name("sdp_bridge_total_parameters")
func sdp_bridge_total_parameters(_ reg: OpaquePointer) -> UInt32

@_silgen_name("sdp_bridge_plugin_count")
func sdp_bridge_plugin_count(_ reg: OpaquePointer) -> Int

@_silgen_name("sdp_bridge_plugin_name")
func sdp_bridge_plugin_name(_ reg: OpaquePointer, _ index: Int) -> UnsafePointer<CChar>

print("=== Swift C Bridge Test (Zero Conversion) ===\n")

// Load audiounit.sdpb
let sdpbPath = "../../testdata/audiounit.sdpb"
guard let sdpbData = try? Data(contentsOf: URL(fileURLWithPath: sdpbPath)) else {
    print("Failed to load audiounit.sdpb")
    exit(1)
}

print("Loaded \(sdpbData.count) bytes from audiounit.sdpb\n")

// Test 1: Decode (data stays in C++)
print("Test 1: Decode (zero-copy)")
guard let registry = sdpbData.withUnsafeBytes({ bufferPtr -> OpaquePointer? in
    return sdp_bridge_decode(
        bufferPtr.baseAddress!.assumingMemoryBound(to: UInt8.self),
        bufferPtr.count
    )
}) else {
    print("✗ Decode failed")
    exit(1)
}

print("✓ Decoded successfully")
print("  Plugins: \(sdp_bridge_plugin_count(registry))")
print("  Total plugin count: \(sdp_bridge_total_plugins(registry))")
print("  Total parameter count: \(sdp_bridge_total_parameters(registry))")

// Access data (creates String only when accessed)
if sdp_bridge_plugin_count(registry) > 0 {
    let namePtr = sdp_bridge_plugin_name(registry, 0)
    let firstName = String(cString: namePtr)
    print("  First plugin: \(firstName)\n")
}

// Test 2: Encode
print("Test 2: Encode")
var encodedLen: Int = 0
guard let encodedPtr = sdp_bridge_encode(registry, &encodedLen) else {
    print("✗ Encode failed")
    sdp_bridge_free(registry)
    exit(1)
}

let encodedData = Data(bytes: encodedPtr, count: encodedLen)
free(encodedPtr) // Free malloc'd buffer

print("✓ Encoded successfully")
print("  Encoded size: \(encodedData.count) bytes\n")

// Test 3: Roundtrip
print("Test 3: Roundtrip verification")
guard let decoded = encodedData.withUnsafeBytes({ bufferPtr -> OpaquePointer? in
    return sdp_bridge_decode(
        bufferPtr.baseAddress!.assumingMemoryBound(to: UInt8.self),
        bufferPtr.count
    )
}) else {
    print("✗ Roundtrip decode failed")
    sdp_bridge_free(registry)
    exit(1)
}

if sdp_bridge_total_plugins(decoded) != sdp_bridge_total_plugins(registry) {
    print("✗ Roundtrip failed: plugin count mismatch")
    sdp_bridge_free(registry)
    sdp_bridge_free(decoded)
    exit(1)
}

if sdp_bridge_total_parameters(decoded) != sdp_bridge_total_parameters(registry) {
    print("✗ Roundtrip failed: parameter count mismatch")
    sdp_bridge_free(registry)
    sdp_bridge_free(decoded)
    exit(1)
}

print("✓ Roundtrip successful\n")

// Test 4: Benchmark
print("Test 4: Benchmark (10000 iterations)")
let iterations = 10000

// Decode benchmark
var start = Date()
for _ in 0..<iterations {
    let temp = sdpbData.withUnsafeBytes { bufferPtr -> OpaquePointer? in
        return sdp_bridge_decode(
            bufferPtr.baseAddress!.assumingMemoryBound(to: UInt8.self),
            bufferPtr.count
        )
    }
    if let temp = temp {
        sdp_bridge_free(temp)
    }
}
var elapsed = Date().timeIntervalSince(start)
let decodeUs = elapsed * 1_000_000 / Double(iterations)
print("  Decode: \(String(format: "%.2f", decodeUs)) μs/op")

// Encode benchmark
start = Date()
for _ in 0..<iterations {
    var len: Int = 0
    if let ptr = sdp_bridge_encode(registry, &len) {
        free(ptr)
    }
}
elapsed = Date().timeIntervalSince(start)
let encodeUs = elapsed * 1_000_000 / Double(iterations)
print("  Encode: \(String(format: "%.2f", encodeUs)) μs/op")

// Roundtrip benchmark
start = Date()
for _ in 0..<iterations {
    var len: Int = 0
    if let ptr = sdp_bridge_encode(registry, &len) {
        let temp = sdp_bridge_decode(ptr, len)
        if let temp = temp {
            sdp_bridge_free(temp)
        }
        free(ptr)
    }
}
elapsed = Date().timeIntervalSince(start)
let roundtripUs = elapsed * 1_000_000 / Double(iterations)
print("  Roundtrip: \(String(format: "%.2f", roundtripUs)) μs/op\n")

// Compare to pure C++ baseline
print("Baseline (pure C++): Encode 29.3μs, Decode 30.7μs, Roundtrip 59.0μs")

let encodeOverhead = (encodeUs / 29.3 - 1.0) * 100
let decodeOverhead = (decodeUs / 30.7 - 1.0) * 100
let roundtripOverhead = (roundtripUs / 59.0 - 1.0) * 100

print(String(format: "Overhead: Encode %.1f%%, Decode %.1f%%, Roundtrip %.1f%%\n",
             encodeOverhead, decodeOverhead, roundtripOverhead))

// Verdict
let acceptable = encodeOverhead < 10.0 && decodeOverhead < 10.0 && roundtripOverhead < 10.0
if acceptable {
    print("✓ SUCCESS: Swift C bridge overhead is acceptable!")
} else {
    print("⚠ WARNING: Overhead exceeds 10% threshold")
}

// Cleanup
sdp_bridge_free(registry)
sdp_bridge_free(decoded)

print("\n=== Test Complete ===")
