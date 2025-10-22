// C++ AudioUnit Message Mode Benchmark
// Measures encoding/decoding performance with real-world AudioUnit data
// Schema: PluginRegistry (62 plugins, 1,759 parameters, ~110KB)

#include "message_encode.hpp"
#include "message_decode.hpp"
#include <iostream>
#include <fstream>
#include <chrono>
#include <vector>
#include <iomanip>

using namespace sdp;
using namespace std::chrono;

// Load binary file
std::vector<uint8_t> loadFile(const std::string& path) {
    std::ifstream file(path, std::ios::binary | std::ios::ate);
    if (!file) {
        throw std::runtime_error("Cannot open file: " + path);
    }
    
    std::streamsize size = file.tellg();
    file.seekg(0, std::ios::beg);
    
    std::vector<uint8_t> buffer(size);
    if (!file.read(reinterpret_cast<char*>(buffer.data()), size)) {
        throw std::runtime_error("Cannot read file: " + path);
    }
    
    return buffer;
}

// Benchmark helper
template<typename Func>
double benchmarkNs(const char* name, Func func, int iterations) {
    // Warmup
    for (int i = 0; i < std::min(1000, iterations / 10); ++i) {
        func();
    }
    
    // Benchmark
    auto start = high_resolution_clock::now();
    for (int i = 0; i < iterations; ++i) {
        func();
    }
    auto end = high_resolution_clock::now();
    
    double totalNs = duration_cast<nanoseconds>(end - start).count();
    double avgNs = totalNs / iterations;
    
    std::cout << std::setw(40) << std::left << name 
              << std::setw(12) << std::right << std::fixed << std::setprecision(2) << avgNs << " ns/op"
              << std::setw(12) << iterations << " iters" << std::endl;
    return avgNs;
}

int main(int argc, char** argv) {
    int iterations = 10000;
    if (argc > 1) {
        iterations = std::atoi(argv[1]);
    }
    
    try {
        std::cout << "=== C++ AudioUnit Message Mode Benchmarks ===" << std::endl;
        std::cout << "Iterations: " << iterations << std::endl;
        std::cout << "Data: 62 plugins, 1,759 parameters, ~110KB" << std::endl << std::endl;
        
        // Load byte mode .sdpb file (run from project root)
        std::cout << "Loading testdata..." << std::endl;
        std::vector<uint8_t> byteModeBinary = loadFile("testdata/binaries/audiounit.sdpb");
        std::cout << "Loaded " << byteModeBinary.size() << " bytes" << std::endl << std::endl;
        
        // Decode to get struct (byte mode)
        PluginRegistry registry = plugin_registry_decode(byteModeBinary.data(), byteModeBinary.size());
        std::cout << "Decoded: " << registry.total_plugin_count << " plugins, " 
                  << registry.total_parameter_count << " parameters" << std::endl << std::endl;
        
        // === Encode Benchmarks ===
        std::cout << "=== Encode Benchmarks ===" << std::endl;
        
        double encodeByteNs = benchmarkNs("Byte Mode: EncodePluginRegistry", [&]() {
            size_t size = plugin_registry_size(registry);
            std::vector<uint8_t> buf(size);
            plugin_registry_encode(registry, buf.data());
        }, iterations);
        
        double encodeMsgNs = benchmarkNs("Message Mode: EncodePluginRegistryMessage", [&]() {
            volatile auto encoded = EncodePluginRegistryMessage(registry);
        }, iterations);
        
        double encodeOverhead = ((encodeMsgNs - encodeByteNs) / encodeByteNs) * 100;
        std::cout << "  → Message mode overhead: " << std::fixed << std::setprecision(1) 
                  << encodeOverhead << "% (" << (encodeMsgNs - encodeByteNs) << " ns)" << std::endl << std::endl;
        
        // === Decode Benchmarks ===
        std::cout << "=== Decode Benchmarks ===" << std::endl;
        
        // Prepare message mode binary
        std::vector<uint8_t> messageModeBinary = EncodePluginRegistryMessage(registry);
        std::cout << "Message mode size: " << messageModeBinary.size() << " bytes "
                  << "(header: 10 bytes, payload: " << (messageModeBinary.size() - 10) << " bytes)" << std::endl;
        
        double decodeByteNs = benchmarkNs("Byte Mode: DecodePluginRegistry", [&]() {
            volatile auto decoded = plugin_registry_decode(byteModeBinary.data(), byteModeBinary.size());
        }, iterations);
        
        double decodeMsgNs = benchmarkNs("Message Mode: DecodePluginRegistryMessage", [&]() {
            volatile auto decoded = DecodePluginRegistryMessage(messageModeBinary);
        }, iterations);
        
        double decodeOverhead = ((decodeMsgNs - decodeByteNs) / decodeByteNs) * 100;
        std::cout << "  → Message mode overhead: " << std::fixed << std::setprecision(1) 
                  << decodeOverhead << "% (" << (decodeMsgNs - decodeByteNs) << " ns)" << std::endl << std::endl;
        
        // === Roundtrip Benchmarks ===
        std::cout << "=== Roundtrip Benchmarks ===" << std::endl;
        
        double roundtripByteNs = benchmarkNs("Byte Mode: Encode + Decode", [&]() {
            size_t size = plugin_registry_size(registry);
            std::vector<uint8_t> buf(size);
            plugin_registry_encode(registry, buf.data());
            volatile auto decoded = plugin_registry_decode(buf.data(), buf.size());
        }, iterations);
        
        double roundtripMsgNs = benchmarkNs("Message Mode: Encode + Decode", [&]() {
            auto encoded = EncodePluginRegistryMessage(registry);
            volatile auto decoded = DecodePluginRegistryMessage(encoded);
        }, iterations);
        
        double roundtripOverhead = ((roundtripMsgNs - roundtripByteNs) / roundtripByteNs) * 100;
        std::cout << "  → Message mode overhead: " << std::fixed << std::setprecision(1) 
                  << roundtripOverhead << "% (" << (roundtripMsgNs - roundtripByteNs) << " ns)" << std::endl << std::endl;
        
        // === Dispatcher Benchmark ===
        std::cout << "=== Dispatcher Benchmark ===" << std::endl;
        
        double dispatcherNs = benchmarkNs("DecodeMessage (with variant)", [&]() {
            volatile auto decoded = DecodeMessage(messageModeBinary);
        }, iterations);
        
        double dispatcherOverhead = dispatcherNs - decodeMsgNs;
        std::cout << "  → Dispatcher overhead: " << std::fixed << std::setprecision(2) 
                  << dispatcherOverhead << " ns (negligible)" << std::endl << std::endl;
        
        // === Summary ===
        std::cout << "=== Summary ===" << std::endl;
        std::cout << "Data size: " << byteModeBinary.size() << " bytes (payload)" << std::endl;
        std::cout << "Message size: " << messageModeBinary.size() << " bytes (10-byte header + payload)" << std::endl;
        std::cout << "Header overhead: 10 bytes (0.009%)" << std::endl << std::endl;
        
        std::cout << "Performance (110KB AudioUnit data):" << std::endl;
        std::cout << "  Byte mode:    " << std::fixed << std::setprecision(0) 
                  << encodeByteNs << " ns encode, " << decodeByteNs << " ns decode" << std::endl;
        std::cout << "  Message mode: " << std::fixed << std::setprecision(0) 
                  << encodeMsgNs << " ns encode, " << decodeMsgNs << " ns decode" << std::endl;
        std::cout << "  Overhead:     " << std::fixed << std::setprecision(1) 
                  << encodeOverhead << "% encode, " << decodeOverhead << "% decode" << std::endl << std::endl;
        
        std::cout << "✓ All benchmarks complete" << std::endl;
        return 0;
        
    } catch (const std::exception& e) {
        std::cerr << "ERROR: " << e.what() << std::endl;
        return 1;
    }
}
