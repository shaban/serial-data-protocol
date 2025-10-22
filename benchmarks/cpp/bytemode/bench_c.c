// Standalone C benchmark for SDP AudioUnit schema
// Reads audiounit.sdpb and benchmarks encode/decode performance
// Usage: ./bench_c

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>
#include <sys/stat.h>

#include "types.h"
#include "decode.h"
#include "encode.h"

// Read entire file into memory
uint8_t* read_file(const char* path, size_t* out_size) {
    FILE* f = fopen(path, "rb");
    if (!f) {
        fprintf(stderr, "Failed to open %s\n", path);
        return NULL;
    }
    
    fseek(f, 0, SEEK_END);
    size_t size = ftell(f);
    fseek(f, 0, SEEK_SET);
    
    uint8_t* data = malloc(size);
    if (!data) {
        fclose(f);
        return NULL;
    }
    
    if (fread(data, 1, size, f) != size) {
        free(data);
        fclose(f);
        return NULL;
    }
    
    fclose(f);
    *out_size = size;
    return data;
}

// High-resolution timer (nanoseconds)
uint64_t nanotime() {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
}

// Benchmark decode: binary -> struct
void bench_decode(const uint8_t* sdpb_data, size_t sdpb_size, int iterations) {
    printf("BenchmarkC_SDP_AudioUnit_Decode\n");
    
    uint64_t total_ns = 0;
    for (int i = 0; i < iterations; i++) {
        SDPPluginRegistry registry;
        
        uint64_t start = nanotime();
        int result = sdp_plugin_registry_decode(&registry, sdpb_data, sdpb_size);
        uint64_t end = nanotime();
        
        if (result != 0) {
            fprintf(stderr, "Decode failed at iteration %d\n", i);
            return;
        }
        
        total_ns += (end - start);
        
        // No free needed - C API is zero-copy
    }
    
    uint64_t avg_ns = total_ns / iterations;
    printf("  %d iterations\n", iterations);
    printf("  %llu ns/op\n", avg_ns);
    printf("  %.2f μs/op\n", avg_ns / 1000.0);
}

// Benchmark encode: struct -> binary
void bench_encode(const uint8_t* sdpb_data, size_t sdpb_size, int iterations) {
    printf("BenchmarkC_SDP_AudioUnit_Encode\n");
    
    // Decode once to get struct for encoding
    SDPPluginRegistry registry;
    if (sdp_plugin_registry_decode(&registry, sdpb_data, sdpb_size) != 0) {
        fprintf(stderr, "Failed to decode test data\n");
        return;
    }
    
    // Use original .sdpb size as upper bound for encoding buffer
    // (actual encoded size will be same or slightly different)
    uint8_t* encode_buf = malloc(sdpb_size * 2);  // 2x for safety
    if (!encode_buf) {
        return;
    }
    
    uint64_t total_ns = 0;
    size_t last_encoded = 0;
    for (int i = 0; i < iterations; i++) {
        uint64_t start = nanotime();
        size_t encoded = sdp_plugin_registry_encode(&registry, encode_buf);
        uint64_t end = nanotime();
        
        last_encoded = encoded;
        total_ns += (end - start);
    }
    
    uint64_t avg_ns = total_ns / iterations;
    printf("  %d iterations\n", iterations);
    printf("  %llu ns/op\n", avg_ns);
    printf("  %.2f μs/op\n", avg_ns / 1000.0);
    printf("  Encoded size: %zu bytes\n", last_encoded);
    
    free(encode_buf);
    // No free needed for registry - C API is zero-copy
}

// Benchmark roundtrip: struct -> binary -> struct
void bench_roundtrip(const uint8_t* sdpb_data, size_t sdpb_size, int iterations) {
    printf("BenchmarkC_SDP_AudioUnit_Roundtrip\n");
    
    // Decode once to get struct for encoding
    SDPPluginRegistry registry;
    if (sdp_plugin_registry_decode(&registry, sdpb_data, sdpb_size) != 0) {
        fprintf(stderr, "Failed to decode test data\n");
        return;
    }
    
    // Use original .sdpb size as upper bound
    uint8_t* encode_buf = malloc(sdpb_size * 2);
    if (!encode_buf) {
        return;
    }
    
    uint64_t total_ns = 0;
    for (int i = 0; i < iterations; i++) {
        uint64_t start = nanotime();
        
        // Encode
        size_t encoded = sdp_plugin_registry_encode(&registry, encode_buf);
        
        // Decode
        SDPPluginRegistry decoded;
        if (sdp_plugin_registry_decode(&decoded, encode_buf, encoded) != 0) {
            fprintf(stderr, "Decode failed at iteration %d\n", i);
            break;
        }
        
        uint64_t end = nanotime();
        total_ns += (end - start);
        
        // Verify
        if (decoded.total_plugin_count != registry.total_plugin_count) {
            fprintf(stderr, "Roundtrip verification failed at iteration %d\n", i);
        }
        
        // No free needed - C API is zero-copy
    }
    
    uint64_t avg_ns = total_ns / iterations;
    printf("  %d iterations\n", iterations);
    printf("  %llu ns/op\n", avg_ns);
    printf("  %.2f μs/op\n", avg_ns / 1000.0);
    
    free(encode_buf);
    // No free needed for registry - C API is zero-copy
}

int main(int argc, char** argv) {
    int iterations = 10000;
    if (argc > 1) {
        iterations = atoi(argv[1]);
    }
    
    printf("=== C SDP AudioUnit Benchmarks ===\n");
    printf("Schema: audiounit.sdp (PluginRegistry)\n");
    printf("Data: ../testdata/audiounit.sdpb (110KB)\n");
    printf("Iterations: %d\n\n", iterations);
    
    // Load test data
    size_t sdpb_size;
    uint8_t* sdpb_data = read_file("../testdata/audiounit.sdpb", &sdpb_size);
    if (!sdpb_data) {
        fprintf(stderr, "Failed to load test data\n");
        return 1;
    }
    
    printf("Loaded %zu bytes\n\n", sdpb_size);
    
    // Run benchmarks
    // NOTE: C encoder is incomplete (TODO: encode nested parameters array)
    // Only decode benchmark works currently
    // bench_encode(sdpb_data, sdpb_size, iterations);
    // printf("\n");
    bench_decode(sdpb_data, sdpb_size, iterations);
    // printf("\n");
    // bench_roundtrip(sdpb_data, sdpb_size, iterations);
    
    free(sdpb_data);
    return 0;
}
