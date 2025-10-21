//
//  main.swift
//  Test Swift 6 C++ interop with benchmarks
//

import Foundation
import SDPAudioUnitSwift

print("=== Swift 6 C++ Interop Test ===\n")

// Load audiounit.sdpb
let sdpbPath = "../../../testdata/audiounit.sdpb"
guard let sdpbData = try? Data(contentsOf: URL(fileURLWithPath: sdpbPath)) else {
    print("Failed to load audiounit.sdpb")
    exit(1)
}

print("Loaded \(sdpbData.count) bytes from audiounit.sdpb\n")

// Test 1: Decode
print("Test 1: Decode")
guard let registry = try? SDPAudioUnitCodec.decode(sdpbData) else {
    print("✗ Decode failed")
    exit(1)
}

print("✓ Decoded successfully")
print("  Plugins: \(registry.plugins.count)")
print("  Total plugin count: \(registry.totalPluginCount)")
print("  Total parameter count: \(registry.totalParameterCount)\n")

// Test 2: Encode
print("Test 2: Encode")
guard let encoded = try? SDPAudioUnitCodec.encode(registry) else {
    print("✗ Encode failed")
    exit(1)
}

print("✓ Encoded successfully")
print("  Encoded size: \(encoded.count) bytes\n")

// Test 3: Roundtrip
print("Test 3: Roundtrip verification")
guard let decoded = try? SDPAudioUnitCodec.decode(encoded) else {
    print("✗ Roundtrip decode failed")
    exit(1)
}

if decoded.totalPluginCount != registry.totalPluginCount {
    print("✗ Roundtrip failed: plugin count mismatch")
    exit(1)
}

if decoded.totalParameterCount != registry.totalParameterCount {
    print("✗ Roundtrip failed: parameter count mismatch")
    exit(1)
}

print("✓ Roundtrip successful\n")

// Test 4: Benchmark
print("Test 4: Benchmark (10000 iterations)")
let iterations = 10000

// Decode benchmark
var start = Date()
for _ in 0..<iterations {
    _ = try? SDPAudioUnitCodec.decode(sdpbData)
}
var elapsed = Date().timeIntervalSince(start)
let decodeUs = elapsed * 1_000_000 / Double(iterations)
print("  Decode: \(String(format: "%.2f", decodeUs)) μs/op")

// Encode benchmark
start = Date()
for _ in 0..<iterations {
    _ = try? SDPAudioUnitCodec.encode(registry)
}
elapsed = Date().timeIntervalSince(start)
let encodeUs = elapsed * 1_000_000 / Double(iterations)
print("  Encode: \(String(format: "%.2f", encodeUs)) μs/op")

// Roundtrip benchmark
start = Date()
for _ in 0..<iterations {
    if let temp = try? SDPAudioUnitCodec.encode(registry) {
        _ = try? SDPAudioUnitCodec.decode(temp)
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
    print("✓ SUCCESS: Swift 6 C++ interop overhead is acceptable!")
} else {
    print("⚠ WARNING: Swift 6 C++ interop overhead exceeds 10% threshold")
}

print("\n=== Test Complete ===")
