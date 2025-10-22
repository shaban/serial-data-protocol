// Standalone C++ benchmark for SDP AudioUnit schema
// Reads audiounit.sdpb and benchmarks encode/decode performance
// Usage: ./bench_cpp

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
    std::cout << "BenchmarkCpp_SDP_AudioUnit_Decode" << std::endl;
    
    uint64_t total_ns = 0;
    for (int i = 0; i < iterations; i++) {
        sdp::PluginRegistry registry;
        
        uint64_t start = nanotime();
        try {
            registry = sdp::plugin_registry_decode(sdpb_data.data(), sdpb_data.size());
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
    std::cout << "BenchmarkCpp_SDP_AudioUnit_Encode" << std::endl;
    
    // Decode once to get struct for encoding
    sdp::PluginRegistry registry;
    try {
        registry = sdp::plugin_registry_decode(sdpb_data.data(), sdpb_data.size());
    } catch (const std::exception& e) {
        std::cerr << "Failed to decode test data: " << e.what() << std::endl;
        return;
    }
    
    uint64_t total_ns = 0;
    for (int i = 0; i < iterations; i++) {
        uint64_t start = nanotime();
        size_t size = sdp::plugin_registry_size(registry);
        std::vector<uint8_t> buffer(size);
        sdp::plugin_registry_encode(registry, buffer.data());
        uint64_t end = nanotime();
        
        total_ns += (end - start);
    }
    
    uint64_t avg_ns = total_ns / iterations;
    std::cout << "  " << iterations << " iterations" << std::endl;
    std::cout << "  " << avg_ns << " ns/op" << std::endl;
    std::cout << std::fixed << std::setprecision(2);
    std::cout << "  " << (avg_ns / 1000.0) << " μs/op" << std::endl;
}

// Benchmark roundtrip: struct -> binary -> struct
void bench_roundtrip(const std::vector<uint8_t>& sdpb_data, int iterations) {
    std::cout << "BenchmarkCpp_SDP_AudioUnit_Roundtrip" << std::endl;
    
    // Decode once to get struct for encoding
    sdp::PluginRegistry registry;
    try {
        registry = sdp::plugin_registry_decode(sdpb_data.data(), sdpb_data.size());
    } catch (const std::exception& e) {
        std::cerr << "Failed to decode test data: " << e.what() << std::endl;
        return;
    }
    
    uint64_t total_ns = 0;
    for (int i = 0; i < iterations; i++) {
        uint64_t start = nanotime();
        
        // Encode
        size_t size = sdp::plugin_registry_size(registry);
        std::vector<uint8_t> buffer(size);
        sdp::plugin_registry_encode(registry, buffer.data());
        
        // Decode
        sdp::PluginRegistry decoded;
        try {
            decoded = sdp::plugin_registry_decode(buffer.data(), buffer.size());
        } catch (const std::exception& e) {
            std::cerr << "Decode failed at iteration " << i << ": " << e.what() << std::endl;
            return;
        }
        
        uint64_t end = nanotime();
        total_ns += (end - start);
        
        // Verify
        if (decoded.total_plugin_count != registry.total_plugin_count) {
            std::cerr << "Roundtrip verification failed at iteration " << i << std::endl;
        }
    }
    
    uint64_t avg_ns = total_ns / iterations;
    std::cout << "  " << iterations << " iterations" << std::endl;
    std::cout << "  " << avg_ns << " ns/op" << std::endl;
    std::cout << std::fixed << std::setprecision(2);
    std::cout << "  " << (avg_ns / 1000.0) << " μs/op" << std::endl;
}

int main(int argc, char** argv) {
    int iterations = 10000;
    if (argc > 1) {
        iterations = std::atoi(argv[1]);
    }
    
    std::cout << "=== C++ SDP AudioUnit Benchmarks ===" << std::endl;
    std::cout << "Schema: audiounit.sdp (PluginRegistry)" << std::endl;
    std::cout << "Data: ../testdata/audiounit.sdpb (110KB)" << std::endl;
    std::cout << "Iterations: " << iterations << std::endl << std::endl;
    
    // Load test data
    std::vector<uint8_t> sdpb_data;
    try {
        sdpb_data = read_file("../testdata/audiounit.sdpb");
    } catch (const std::exception& e) {
        std::cerr << "Failed to load test data: " << e.what() << std::endl;
        return 1;
    }
    
    std::cout << "Loaded " << sdpb_data.size() << " bytes" << std::endl << std::endl;
    
    // Run benchmarks
    bench_encode(sdpb_data, iterations);
    std::cout << std::endl;
    bench_decode(sdpb_data, iterations);
    std::cout << std::endl;
    bench_roundtrip(sdpb_data, iterations);
    
    return 0;
}
