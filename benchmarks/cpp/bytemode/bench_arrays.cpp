// Standalone C++ benchmark for SDP Arrays schema (bulk optimization)
// Reads arrays_primitives.sdpb and benchmarks encode/decode performance
// Usage: ./bench_arrays

#include <iostream>
#include <fstream>
#include <vector>
#include <chrono>
#include <iomanip>

#include "types.hpp"
#include "decode.hpp"
#include "encode.hpp"

// Read entire file into memory
std::vector<uint8_t> read_file(const std::string& path) {
    std::ifstream file(path, std::ios::binary | std::ios::ate);
    if (!file) {
        throw std::runtime_error("Failed to open " + path);
    }
    
    std::streamsize size = file.tellg();
    file.seekg(0, std::ios::beg);
    
    std::vector<uint8_t> buffer(size);
    if (!file.read(reinterpret_cast<char*>(buffer.data()), size)) {
        throw std::runtime_error("Failed to read " + path);
    }
    
    return buffer;
}

// High-resolution timer (nanoseconds)
uint64_t nanotime() {
    auto now = std::chrono::high_resolution_clock::now();
    return std::chrono::duration_cast<std::chrono::nanoseconds>(now.time_since_epoch()).count();
}

// Benchmark decode: binary -> struct
void bench_decode(const std::vector<uint8_t>& sdpb_data, int iterations) {
    std::cout << "BenchmarkCpp_SDP_Arrays_Decode" << std::endl;
    
    uint64_t total_ns = 0;
    for (int i = 0; i < iterations; i++) {
        sdp::ArraysOfPrimitives arrays;
        
        uint64_t start = nanotime();
        try {
            arrays = sdp::arrays_of_primitives_decode(sdpb_data.data(), sdpb_data.size());
        } catch (const std::exception& e) {
            std::cerr << "Decode failed at iteration " << i << ": " << e.what() << std::endl;
            return;
        }
        uint64_t end = nanotime();
        
        total_ns += (end - start);
    }
    
    uint64_t avg_ns = total_ns / iterations;
    std::cout << "  " << iterations << " iterations" << std::endl;
    std::cout << "  " << avg_ns << " ns/op" << std::endl;
    std::cout << std::fixed << std::setprecision(2);
    std::cout << "  " << (avg_ns / 1000.0) << " μs/op" << std::endl;
}

// Benchmark encode: struct -> binary
void bench_encode(const std::vector<uint8_t>& sdpb_data, int iterations) {
    std::cout << "BenchmarkCpp_SDP_Arrays_Encode" << std::endl;
    
    // Decode once to get struct for encoding
    sdp::ArraysOfPrimitives arrays;
    try {
        arrays = sdp::arrays_of_primitives_decode(sdpb_data.data(), sdpb_data.size());
    } catch (const std::exception& e) {
        std::cerr << "Failed to decode test data: " << e.what() << std::endl;
        return;
    }
    
    uint64_t total_ns = 0;
    size_t encoded_size = 0;
    for (int i = 0; i < iterations; i++) {
        uint64_t start = nanotime();
        // Include size calculation in timing - this is what users must do
        size_t size = sdp::arrays_of_primitives_size(arrays);
        std::vector<uint8_t> buffer(size);
        try {
            sdp::arrays_of_primitives_encode(arrays, buffer.data());
        } catch (const std::exception& e) {
            std::cerr << "Encode failed at iteration " << i << ": " << e.what() << std::endl;
            return;
        }
        uint64_t end = nanotime();
        
        total_ns += (end - start);
        encoded_size = size;
    }
    
    uint64_t avg_ns = total_ns / iterations;
    std::cout << "  " << iterations << " iterations" << std::endl;
    std::cout << "  " << avg_ns << " ns/op" << std::endl;
    std::cout << std::fixed << std::setprecision(2);
    std::cout << "  " << (avg_ns / 1000.0) << " μs/op" << std::endl;
    std::cout << "  Encoded size: " << encoded_size << " bytes" << std::endl;
}

// Benchmark roundtrip: encode + decode
void bench_roundtrip(const std::vector<uint8_t>& sdpb_data, int iterations) {
    std::cout << "BenchmarkCpp_SDP_Arrays_Roundtrip" << std::endl;
    
    // Decode once to get struct
    sdp::ArraysOfPrimitives original;
    try {
        original = sdp::arrays_of_primitives_decode(sdpb_data.data(), sdpb_data.size());
    } catch (const std::exception& e) {
        std::cerr << "Failed to decode test data: " << e.what() << std::endl;
        return;
    }
    
    uint64_t total_ns = 0;
    for (int i = 0; i < iterations; i++) {
        uint64_t start = nanotime();
        
        // Encode (with size calculation)
        size_t size = sdp::arrays_of_primitives_size(original);
        std::vector<uint8_t> encoded(size);
        try {
            sdp::arrays_of_primitives_encode(original, encoded.data());
        } catch (const std::exception& e) {
            std::cerr << "Encode failed at iteration " << i << ": " << e.what() << std::endl;
            return;
        }
        
        // Decode
        sdp::ArraysOfPrimitives decoded;
        try {
            decoded = sdp::arrays_of_primitives_decode(encoded.data(), encoded.size());
        } catch (const std::exception& e) {
            std::cerr << "Decode failed at iteration " << i << ": " << e.what() << std::endl;
            return;
        }
        
        uint64_t end = nanotime();
        total_ns += (end - start);
        
        // Verify
        if (decoded.u8_array.size() != original.u8_array.size()) {
            std::cerr << "Roundtrip verification failed at iteration " << i << std::endl;
            return;
        }
    }
    
    uint64_t avg_ns = total_ns / iterations;
    std::cout << "  " << iterations << " iterations" << std::endl;
    std::cout << "  " << avg_ns << " ns/op" << std::endl;
    std::cout << std::fixed << std::setprecision(2);
    std::cout << "  " << (avg_ns / 1000.0) << " μs/op" << std::endl;
}

int main(int argc, char** argv) {
    std::cout << "=== C++ SDP Byte Mode: Arrays Benchmark ===" << std::endl;
    std::cout << "Schema: arrays.sdp (ArraysOfPrimitives)" << std::endl;
    std::cout << "Data: testdata/binaries/arrays_primitives.sdpb (canonical)" << std::endl;
    std::cout << std::endl;
    
    // Read canonical test data
    std::vector<uint8_t> sdpb_data;
    try {
        sdpb_data = read_file("testdata/binaries/arrays_primitives.sdpb");
    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        std::cerr << "Run from project root" << std::endl;
        return 1;
    }
    
    std::cout << "Loaded " << sdpb_data.size() << " bytes from canonical fixture" << std::endl;
    
    // Decode once to show data
    try {
        sdp::ArraysOfPrimitives arrays = sdp::arrays_of_primitives_decode(sdpb_data.data(), sdpb_data.size());
        std::cout << "u8_array: " << arrays.u8_array.size() << " elements" << std::endl;
        std::cout << "u32_array: " << arrays.u32_array.size() << " elements" << std::endl;
        std::cout << "f64_array: " << arrays.f64_array.size() << " elements" << std::endl;
        std::cout << "str_array: " << arrays.str_array.size() << " elements" << std::endl;
        std::cout << "bool_array: " << arrays.bool_array.size() << " elements" << std::endl;
    } catch (const std::exception& e) {
        std::cerr << "Failed to decode: " << e.what() << std::endl;
        return 1;
    }
    
    std::cout << std::endl;
    
    int iterations = 10000;
    if (argc > 1) {
        iterations = std::atoi(argv[1]);
    }
    
    // Warm up
    for (int i = 0; i < 100; i++) {
        sdp::ArraysOfPrimitives arrays = sdp::arrays_of_primitives_decode(sdpb_data.data(), sdpb_data.size());
        size_t size = sdp::arrays_of_primitives_size(arrays);
        std::vector<uint8_t> encoded(size);
        sdp::arrays_of_primitives_encode(arrays, encoded.data());
    }
    
    // Run benchmarks
    bench_encode(sdpb_data, iterations);
    std::cout << std::endl;
    
    bench_decode(sdpb_data, iterations);
    std::cout << std::endl;
    
    bench_roundtrip(sdpb_data, iterations / 2);
    std::cout << std::endl;
    
    return 0;
}
