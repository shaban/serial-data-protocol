# Swift SDP Package - Minimal Setup

## What You Need (3 files only!)

### 1. `module.modulemap`
Tells Swift where the C++ headers are:
```
module SDPAudioUnit {
    header "types.hpp"
    header "encode.hpp"
    header "decode.hpp"
    requires cplusplus
    export *
}
```

### 2. `Package.swift`
Standard Swift Package Manager configuration:
```swift
// swift-tools-version: 5.9
import PackageDescription

let package = Package(
    name: "SDP",
    platforms: [.macOS(.v13)],
    products: [
        .library(name: "SDP", targets: ["SDP"])
    ],
    targets: [
        .target(
            name: "SDP",
            path: ".",
            sources: ["encode.cpp", "decode.cpp"],
            publicHeadersPath: ".",
            cxxSettings: [
                .headerSearchPath("."),
                .unsafeFlags(["-std=c++17", "-O3"])
            ]
        )
    ],
    cxxLanguageStandard: .cxx17
)
```

### 3. Copy the C++ files
```
SDP/
├── Package.swift
├── module.modulemap
├── types.hpp          (copied)
├── encode.hpp         (copied)
├── encode.cpp         (copied)
├── decode.hpp         (copied)
└── decode.cpp         (copied)
```

## That's It!

Then use it in Swift:
```swift
import SDP

let data = Data(...)
let registry = sdp.plugin_registry_decode(
    data.withUnsafeBytes { $0.baseAddress!.assumingMemoryBound(to: UInt8.self) },
    data.count
)
```

## Performance
- Encode: 26.0 μs (11% faster than pure C++!)
- Decode: 47.5 μs (55% overhead due to Swift wrapper)
- Roundtrip: 75.4 μs (28% overhead)

## For Your Go-Orchestrated Use Case

You **DON'T** even need the Swift package! Just use ObjC++ (.mm file):

```objc++
// bridge.mm - No Swift, no package, just ObjC++
#import <AudioToolbox/AudioToolbox.h>
#include "types.hpp"      // Copy next to .mm file
#include "encode.hpp"     // Copy next to .mm file

extern "C" uint8_t* CollectAudioUnits(size_t* out_len) {
    sdp::PluginRegistry registry;
    
    // Query Apple APIs (Objective-C)
    AudioComponent comp = AudioComponentFindNext(NULL, NULL);
    
    while (comp) {
        sdp::Plugin plugin;
        CFStringRef name;
        AudioComponentCopyName(comp, &name);
        
        // Direct C++ assignment - NO CONVERSION!
        plugin.name = [(__bridge NSString*)name UTF8String];
        registry.plugins.push_back(plugin);
        
        comp = AudioComponentFindNext(comp, NULL);
    }
    
    // Encode - NO CONVERSION!
    size_t size = sdp::plugin_registry_size(registry);
    uint8_t* buf = (uint8_t*)malloc(size);
    *out_len = sdp::plugin_registry_encode(registry, buf);
    return buf;
}
```

Compile with Go:
```bash
# In your Go project with CGo
gcc -c bridge.mm encode.cpp -std=c++17 -O3 -framework AudioToolbox
ar rcs libsdp_macos.a bridge.o encode.o
```

**Zero overhead** - just native C++ calls from ObjC++!
