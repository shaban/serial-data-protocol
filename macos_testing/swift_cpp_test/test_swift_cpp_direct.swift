//
//  test_swift_cpp_direct.swift
//  Direct C++ calls from Swift - ACTUAL BENCHMARK
//
//  This directly imports and calls C++ functions without any wrapper.
//

import Foundation
import SDPAudioUnit  // Our C++ module

print("=== Swift Direct C++ Call (Barebone) - BENCHMARK ===\n")

// Load audiounit.sdpb
let sdpbPath = "../../testdata/audiounit.sdpb"
guard let sdpbData = try? Data(contentsOf: URL(fileURLWithPath: sdpbPath)) else {
    print("Failed to load audiounit.sdpb")
    exit(1)
}

print("Loaded \(sdpbData.count) bytes from audiounit.sdpb\n")

// Test 1: Decode (direct C++ call!)
print("Test 1: Decode (direct C++ call)")
do {
    let registry = try sdpbData.withUnsafeBytes { bufferPtr -> sdp.PluginRegistry in
        let ptr = bufferPtr.baseAddress!.assumingMemoryBound(to: UInt8.self)
        // Direct C++ function call!
        return sdp.plugin_registry_decode(ptr, bufferPtr.count)
    }
    
    print("✓ Decoded successfully")
    print("  Plugins: \(registry.plugins.size())")
    print("  Total plugin count: \(registry.total_plugin_count)")
    print("  Total parameter count: \(registry.total_parameter_count)")
    
    // Access C++ string
    if registry.plugins.size() > 0 {
        let firstPlugin = registry.plugins[0]
        let name = String(firstPlugin.name)  // std::string -> Swift String
        print("  First plugin: \(name)\n")
    }
    
    // Test 2: Encode (direct C++ call!)
    print("Test 2: Encode (direct C++ call)")
    let size = sdp.plugin_registry_size(registry)
    var buffer = [UInt8](repeating: 0, count: size)
    
    let encoded = buffer.withUnsafeMutableBytes { bufPtr -> Int in
        let ptr = bufPtr.baseAddress!.assumingMemoryBound(to: UInt8.self)
        // Direct C++ function call!
        return sdp.plugin_registry_encode(registry, ptr)
    }
    
    print("✓ Encoded successfully")
    print("  Encoded size: \(encoded) bytes\n")
    
    // Test 3: Roundtrip
    print("Test 3: Roundtrip verification")
    let decoded = try buffer.prefix(encoded).withUnsafeBytes { bufPtr -> sdp.PluginRegistry in
        let ptr = bufPtr.baseAddress!.assumingMemoryBound(to: UInt8.self)
        return sdp.plugin_registry_decode(ptr, bufPtr.count)
    }
    
    if decoded.total_plugin_count != registry.total_plugin_count {
        print("✗ Roundtrip failed: plugin count mismatch")
        exit(1)
    }
    
    if decoded.total_parameter_count != registry.total_parameter_count {
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
        _ = try? sdpbData.withUnsafeBytes { bufPtr -> sdp.PluginRegistry in
            let ptr = bufPtr.baseAddress!.assumingMemoryBound(to: UInt8.self)
            return sdp.plugin_registry_decode(ptr, bufPtr.count)
        }
    }
    var elapsed = Date().timeIntervalSince(start)
    let decodeUs = elapsed * 1_000_000 / Double(iterations)
    print("  Decode: \(String(format: "%.2f", decodeUs)) μs/op")
    
    // Encode benchmark
    start = Date()
    for _ in 0..<iterations {
        let size = sdp.plugin_registry_size(registry)
        var buf = [UInt8](repeating: 0, count: size)
        _ = buf.withUnsafeMutableBytes { bufPtr -> Int in
            let ptr = bufPtr.baseAddress!.assumingMemoryBound(to: UInt8.self)
            return sdp.plugin_registry_encode(registry, ptr)
        }
    }
    elapsed = Date().timeIntervalSince(start)
    let encodeUs = elapsed * 1_000_000 / Double(iterations)
    print("  Encode: \(String(format: "%.2f", encodeUs)) μs/op")
    
    // Roundtrip benchmark
    start = Date()
    for _ in 0..<iterations {
        let size = sdp.plugin_registry_size(registry)
        var buf = [UInt8](repeating: 0, count: size)
        _ = buf.withUnsafeMutableBytes { bufPtr -> Int in
            let ptr = bufPtr.baseAddress!.assumingMemoryBound(to: UInt8.self)
            return sdp.plugin_registry_encode(registry, ptr)
        }
        _ = try? buf.withUnsafeBytes { bufPtr -> sdp.PluginRegistry in
            let ptr = bufPtr.baseAddress!.assumingMemoryBound(to: UInt8.self)
            return sdp.plugin_registry_decode(ptr, bufPtr.count)
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
        print("✓ SUCCESS: Direct C++ call overhead is acceptable!")
    } else {
        print("⚠ WARNING: Overhead exceeds 10% threshold")
    }
    
    print("\nNote: This is calling C++ directly with NO wrapper layer!")
    
} catch {
    print("✗ Error: \(error)")
    exit(1)
}

print("\n=== Test Complete ===")
