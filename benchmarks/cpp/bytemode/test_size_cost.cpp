// Quick test: measure the cost of arrays_of_primitives_size()
#include <iostream>
#include <fstream>
#include <chrono>
#include <vector>
#include "types.hpp"
#include "decode.hpp"
#include "encode.hpp"

uint64_t nanotime() {
    auto now = std::chrono::high_resolution_clock::now();
    return std::chrono::duration_cast<std::chrono::nanoseconds>(now.time_since_epoch()).count();
}

std::vector<uint8_t> read_file(const std::string& path) {
    std::ifstream file(path, std::ios::binary | std::ios::ate);
    std::streamsize size = file.tellg();
    file.seekg(0, std::ios::beg);
    std::vector<uint8_t> buffer(size);
    file.read(reinterpret_cast<char*>(buffer.data()), size);
    return buffer;
}

int main() {
    auto sdpb_data = read_file("testdata/binaries/arrays_primitives.sdpb");
    auto arrays = sdp::arrays_of_primitives_decode(sdpb_data.data(), sdpb_data.size());
    
    // Benchmark size calculation
    int iterations = 100000;
    uint64_t total_ns = 0;
    for (int i = 0; i < iterations; i++) {
        uint64_t start = nanotime();
        size_t size = sdp::arrays_of_primitives_size(arrays);
        uint64_t end = nanotime();
        total_ns += (end - start);
        
        // Prevent optimization
        if (size == 0) std::cout << "impossible";
    }
    
    std::cout << "_size() call: " << (total_ns / iterations) << " ns/op" << std::endl;
    std::cout << "encode call:  39 ns/op (from benchmark)" << std::endl;
    std::cout << "Combined:     " << ((total_ns / iterations) + 39) << " ns/op" << std::endl;
    
    return 0;
}
