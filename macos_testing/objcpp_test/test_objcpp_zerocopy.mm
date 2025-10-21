//
//  test_objcpp_zerocopy.mm
//  Test zero-copy Objective-C++ bridge
//
//  Compile: clang++ -std=c++17 -ObjC++ -O3 -framework Foundation \
//           test_objcpp_zerocopy.mm SDPAudioUnit_ZeroCopy.mm \
//           ../../testdata/audiounit_cpp/encode.cpp \
//           ../../testdata/audiounit_cpp/decode.cpp \
//           -I../../testdata/audiounit_cpp \
//           -o test_objcpp_zerocopy
//

#import <Foundation/Foundation.h>
#import "SDPAudioUnit_ZeroCopy.h"
#include <chrono>
#include <iostream>

int main(int argc, const char * argv[]) {
    @autoreleasepool {
        NSLog(@"=== Objective-C++ ZERO-COPY Bridge Test ===\n");
        
        // Load audiounit.sdpb
        NSString *sdpbPath = @"../../testdata/audiounit.sdpb";
        NSError *error = nil;
        NSData *sdpbData = [NSData dataWithContentsOfFile:sdpbPath
                                                   options:0
                                                     error:&error];
        
        if (!sdpbData) {
            NSLog(@"Failed to load audiounit.sdpb: %@", error);
            return 1;
        }
        
        NSLog(@"Loaded %lu bytes from audiounit.sdpb\n", (unsigned long)sdpbData.length);
        
        // Test 1: Decode (NO object allocation!)
        NSLog(@"Test 1: Decode (zero-copy)");
        SDPPluginRegistry *registry = [SDPPluginRegistry decodeFromData:sdpbData
                                                                  error:&error];
        if (!registry) {
            NSLog(@"Decode failed: %@", error);
            return 1;
        }
        
        NSLog(@"✓ Decoded successfully");
        NSLog(@"  Plugins: %ld", (long)registry.pluginCount);
        NSLog(@"  Total plugin count: %u", registry.totalPluginCount);
        NSLog(@"  Total parameter count: %u", registry.totalParameterCount);
        
        // Sample access (creates strings on demand)
        if (registry.pluginCount > 0) {
            NSString *firstPluginName = [registry pluginNameAtIndex:0];
            NSLog(@"  First plugin: %@\n", firstPluginName);
        }
        
        // Test 2: Encode
        NSLog(@"Test 2: Encode");
        NSData *encoded = [registry encodeWithError:&error];
        if (!encoded) {
            NSLog(@"Encode failed: %@", error);
            return 1;
        }
        
        NSLog(@"✓ Encoded successfully");
        NSLog(@"  Encoded size: %lu bytes\n", (unsigned long)encoded.length);
        
        // Test 3: Roundtrip
        NSLog(@"Test 3: Roundtrip verification");
        SDPPluginRegistry *decoded = [SDPPluginRegistry decodeFromData:encoded
                                                                 error:&error];
        if (!decoded) {
            NSLog(@"Roundtrip decode failed: %@", error);
            return 1;
        }
        
        if (decoded.totalPluginCount != registry.totalPluginCount) {
            NSLog(@"✗ Roundtrip failed: plugin count mismatch");
            return 1;
        }
        
        if (decoded.totalParameterCount != registry.totalParameterCount) {
            NSLog(@"✗ Roundtrip failed: parameter count mismatch");
            return 1;
        }
        
        NSLog(@"✓ Roundtrip successful\n");
        
        // Test 4: Benchmark
        NSLog(@"Test 4: Benchmark (10000 iterations)");
        int iterations = 10000;
        
        // Decode benchmark (zero-copy!)
        auto start = std::chrono::high_resolution_clock::now();
        for (int i = 0; i < iterations; i++) {
            SDPPluginRegistry *temp = [SDPPluginRegistry decodeFromData:sdpbData
                                                                  error:nil];
            (void)temp; // Silence unused warning
        }
        auto end = std::chrono::high_resolution_clock::now();
        auto decode_ns = std::chrono::duration_cast<std::chrono::nanoseconds>(end - start).count();
        double decode_us = decode_ns / 1000.0 / iterations;
        
        NSLog(@"  Decode: %.2f μs/op", decode_us);
        
        // Encode benchmark
        start = std::chrono::high_resolution_clock::now();
        for (int i = 0; i < iterations; i++) {
            NSData *temp = [registry encodeWithError:nil];
            (void)temp;
        }
        end = std::chrono::high_resolution_clock::now();
        auto encode_ns = std::chrono::duration_cast<std::chrono::nanoseconds>(end - start).count();
        double encode_us = encode_ns / 1000.0 / iterations;
        
        NSLog(@"  Encode: %.2f μs/op", encode_us);
        
        // Roundtrip benchmark
        start = std::chrono::high_resolution_clock::now();
        for (int i = 0; i < iterations; i++) {
            NSData *temp = [registry encodeWithError:nil];
            SDPPluginRegistry *temp2 = [SDPPluginRegistry decodeFromData:temp error:nil];
            (void)temp2;
        }
        end = std::chrono::high_resolution_clock::now();
        auto roundtrip_ns = std::chrono::duration_cast<std::chrono::nanoseconds>(end - start).count();
        double roundtrip_us = roundtrip_ns / 1000.0 / iterations;
        
        NSLog(@"  Roundtrip: %.2f μs/op\n", roundtrip_us);
        
        // Compare to pure C++ baseline
        NSLog(@"Baseline (pure C++): Encode 29.3μs, Decode 30.7μs, Roundtrip 59.0μs");
        
        double encode_overhead = (encode_us / 29.3 - 1.0) * 100;
        double decode_overhead = (decode_us / 30.7 - 1.0) * 100;
        double roundtrip_overhead = (roundtrip_us / 59.0 - 1.0) * 100;
        
        NSLog(@"Overhead: Encode %.1f%%, Decode %.1f%%, Roundtrip %.1f%%\n",
              encode_overhead, decode_overhead, roundtrip_overhead);
        
        // Verdict
        BOOL acceptable = (encode_overhead < 10.0 && decode_overhead < 10.0 && roundtrip_overhead < 10.0);
        if (acceptable) {
            NSLog(@"✓ SUCCESS: Zero-copy Objective-C++ bridge overhead is acceptable!");
        } else {
            NSLog(@"⚠ WARNING: Overhead exceeds 10%% threshold");
            NSLog(@"  But this is MUCH better than 1500%% overhead of the object-based version!");
        }
        
        NSLog(@"\n=== Test Complete ===");
    }
    return 0;
}
