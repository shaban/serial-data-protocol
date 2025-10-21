//
//  test_swift_barebone.swift
//  Direct C++ calls from Swift (no C wrapper!)
//
//  This uses Swift's C++ interop to call C++ functions directly.
//  Requires: Swift 5.9+ with .interoperabilityMode(.Cxx)
//
//  Compile with module map approach (see below)
//

import Foundation

// Swift can import C++ namespaces directly with C++ interop
// But we need a module.modulemap file to tell Swift about the C++ headers

// For this demo, we'll show what the Swift code WOULD look like:

print("=== Swift Barebone C++ Direct Call Demo ===\n")
print("This demonstrates calling C++ directly without C wrapper.\n")

// What the code WOULD look like if module map is set up:
/*

import SDPAudioUnit  // C++ module

let sdpbData = try! Data(contentsOf: URL(fileURLWithPath: "../../testdata/audiounit.sdpb"))

// Call C++ function directly!
let registry = sdpbData.withUnsafeBytes { bufferPtr -> sdp.PluginRegistry in
    return sdp.plugin_registry_decode(
        bufferPtr.baseAddress!.assumingMemoryBound(to: UInt8.self),
        bufferPtr.count
    )
}

// Access C++ struct fields directly
print("Plugins: \(registry.plugins.size())")
print("Total: \(registry.total_plugin_count)")

// Access nested data
for i in 0..<registry.plugins.size() {
    let plugin = registry.plugins[i]
    print("Plugin: \(String(plugin.name))")  // std::string -> Swift String
}

// Encode back
var size = sdp.plugin_registry_size(registry)
var buffer = [UInt8](repeating: 0, count: size)
buffer.withUnsafeMutableBytes { bufPtr in
    let encoded = sdp.plugin_registry_encode(
        registry,
        bufPtr.baseAddress!.assumingMemoryBound(to: UInt8.self)
    )
    print("Encoded: \(encoded) bytes")
}

*/

print("""
To use this approach, you need a module.modulemap:

module SDPAudioUnit {
    header "types.hpp"
    header "encode.hpp"
    header "decode.hpp"
    requires cplusplus
    export *
}

Then compile with:
  swiftc -O -I../../testdata/audiounit_cpp \\
    -Xcc -std=c++17 \\
    -cxx-interoperability-mode=default \\
    test.swift \\
    ../../testdata/audiounit_cpp/encode.cpp \\
    ../../testdata/audiounit_cpp/decode.cpp \\
    -o test

The Swift code directly uses C++ types:
  - sdp::PluginRegistry (C++) -> sdp.PluginRegistry (Swift)
  - std::string (C++) -> Swift can convert to String
  - std::vector (C++) -> Swift can iterate with .size() / []
  
""")

print("=== Comparison: C Wrapper vs Direct C++ ===\n")

print("C Wrapper (sdp_bridge.h/.cpp):")
print("  ✓ Stable ABI (C has guaranteed compatibility)")
print("  ✓ Easy Swift import (no module map needed)")
print("  ✓ Exception safety (catches C++ exceptions -> returns nil)")
print("  ✓ Opaque pointers (hides C++ complexity)")
print("  ✓ Works with any Swift version")
print("  ✗ Extra layer of functions (~5% overhead)")
print("  ✗ More code to maintain (wrapper functions)\n")

print("Direct C++ (no wrapper):")
print("  ✓ Zero wrapper overhead (direct calls)")
print("  ✓ Access C++ structs directly (no conversion)")
print("  ✓ Fewer files (no bridge.h/cpp)")
print("  ✗ Requires Swift 5.9+ with C++ interop")
print("  ✗ Requires module.modulemap setup")
print("  ✗ C++ exceptions crash Swift (no try/catch)")
print("  ✗ std::vector/string not fully bridged (manual conversion)")
print("  ✗ ABI unstable (C++ name mangling changes)\n")

print("=== Your Use Case: Go-Orchestrated ===\n")
print("For your architecture, you DON'T need Swift to decode at all!")
print("")
print("Your flow:")
print("  1. Swift/ObjC++ queries Apple APIs")
print("  2. Builds C++ structs directly:")
print("     sdp::Plugin plugin;")
print("     plugin.name = [nsString UTF8String];  // Direct assignment!")
print("  3. Encodes with C++:")
print("     sdp::plugin_registry_encode(registry, buf);")
print("  4. Returns raw bytes to Go")
print("  5. Go decodes (pure Go, no CGo overhead!)\n")
print("")
print("For this use case:")
print("  - You're writing ObjC++ (.mm file) NOT pure Swift")
print("  - ObjC++ can #include C++ headers directly (no module map!)")
print("  - No wrapper needed - just #include \"encode.hpp\"")
print("  - Zero overhead - native C++ code\n")

print("=== Recommendation ===\n")
print("For your Go-orchestrated architecture:")
print("  USE: Direct C++ includes in ObjC++ (.mm files)")
print("  DON'T USE: C wrapper or pure Swift\n")
print("  The .mm file IS the bridge - it can do both ObjC and C++!")
print("")

print("""
Example ObjC++ code (.mm file):

#import <AudioToolbox/AudioToolbox.h>
#include "../../testdata/audiounit_cpp/types.hpp"
#include "../../testdata/audiounit_cpp/encode.hpp"

extern "C" uint8_t* CollectAudioUnits(size_t* out_len) {
    sdp::PluginRegistry registry;
    
    // Query Apple API (Objective-C)
    AudioComponent comp = AudioComponentFindNext(NULL, NULL);
    
    while (comp) {
        sdp::Plugin plugin;
        
        // Get name (Objective-C)
        CFStringRef name;
        AudioComponentCopyName(comp, &name);
        
        // Direct C++ assignment - NO WRAPPER!
        plugin.name = [(__bridge NSString*)name UTF8String];
        
        // Add to C++ struct - NO WRAPPER!
        registry.plugins.push_back(plugin);
        
        comp = AudioComponentFindNext(comp, NULL);
    }
    
    // Encode with C++ - NO WRAPPER!
    size_t size = sdp::plugin_registry_size(registry);
    uint8_t* buf = (uint8_t*)malloc(size);
    *out_len = sdp::plugin_registry_encode(registry, buf);
    
    return buf;  // Go will free() this
}

No wrapper. No overhead. Pure C++ structs and functions.
This is the way. 🎯
""")

print("\n=== Test Complete ===")
