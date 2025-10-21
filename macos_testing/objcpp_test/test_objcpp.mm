//
//  test_objcpp.mm
//  Test program to verify Objective-C++ bridge works
//
//  Compile: clang++ -std=c++17 -ObjC++ -framework Foundation \
//           test_objcpp.mm SDPAudioUnit.mm \
//           ../../testdata/audiounit_cpp/encode.cpp \
//           ../../testdata/audiounit_cpp/decode.cpp \
//           -I../../testdata/audiounit_cpp \
//           -o test_objcpp
//
//  Run: ./test_objcpp
//

#import <Foundation/Foundation.h>
#import "SDPAudioUnit.h"
#include <chrono>
#include <iostream>

int main(int argc, const char * argv[]) {
    @autoreleasepool {
        NSLog(@"=== Objective-C++ Bridge Test ===\n");
        
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
        
        // Test 1: Decode
        NSLog(@"Test 1: Decode");
        SDPPluginRegistry *registry = [SDPAudioUnitCodec decodePluginRegistry:sdpbData
                                                                        error:&error];
        if (!registry) {
            NSLog(@"Decode failed: %@", error);
            return 1;
        }
        
        NSLog(@"✓ Decoded successfully");
        NSLog(@"  Plugins: %lu", (unsigned long)registry.plugins.count);
        NSLog(@"  Total plugin count: %u", registry.totalPluginCount);
        NSLog(@"  Total parameter count: %u\n", registry.totalParameterCount);
        
        // Test 2: Encode
        NSLog(@"Test 2: Encode");
        NSData *encoded = [SDPAudioUnitCodec encodePluginRegistry:registry
                                                            error:&error];
        if (!encoded) {
            NSLog(@"Encode failed: %@", error);
            return 1;
        }
        
        NSLog(@"✓ Encoded successfully");
        NSLog(@"  Encoded size: %lu bytes\n", (unsigned long)encoded.length);
        
        // Test 3: Roundtrip
        NSLog(@"Test 3: Roundtrip verification");
        SDPPluginRegistry *decoded = [SDPAudioUnitCodec decodePluginRegistry:encoded
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
        
        // Decode benchmark
        auto start = std::chrono::high_resolution_clock::now();
        for (int i = 0; i < iterations; i++) {
            SDPPluginRegistry *temp = [SDPAudioUnitCodec decodePluginRegistry:sdpbData
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
            NSData *temp = [SDPAudioUnitCodec encodePluginRegistry:registry
                                                             error:nil];
            (void)temp;
        }
        end = std::chrono::high_resolution_clock::now();
        auto encode_ns = std::chrono::duration_cast<std::chrono::nanoseconds>(end - start).count();
        double encode_us = encode_ns / 1000.0 / iterations;
        
        NSLog(@"  Encode: %.2f μs/op", encode_us);
        
        // Roundtrip benchmark
        start = std::chrono::high_resolution_clock::now();
        for (int i = 0; i < iterations; i++) {
            NSData *temp = [SDPAudioUnitCodec encodePluginRegistry:registry error:nil];
            SDPPluginRegistry *temp2 = [SDPAudioUnitCodec decodePluginRegistry:temp error:nil];
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
            NSLog(@"✓ SUCCESS: Objective-C++ bridge overhead is acceptable!");
        } else {
            NSLog(@"⚠ WARNING: Objective-C++ bridge overhead exceeds 10%% threshold");
        }
        
        NSLog(@"\n=== Test Complete ===");
    }
    return 0;
}
