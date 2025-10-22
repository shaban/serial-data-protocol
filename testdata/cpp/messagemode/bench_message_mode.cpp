// Benchmark C++ message mode performance
#include "message_encode.hpp"
#include "message_decode.hpp"
#include <iostream>
#include <chrono>
#include <vector>

using namespace sdp;
using namespace std::chrono;

template<typename Func>
double benchmarkNs(const char* name, Func func, int iterations = 100000) {
    // Warmup
    for (int i = 0; i < 1000; ++i) {
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
    
    std::cout << name << ": " << avgNs << " ns/op (" << iterations << " iterations)" << std::endl;
    return avgNs;
}

void benchmarkPoint() {
    std::cout << "\n=== Point Benchmarks ===" << std::endl;
    
    Point p;
    p.x = 3.14;
    p.y = 2.71;
    
    // Encode benchmark
    benchmarkNs("EncodePointMessage", [&]() {
        volatile auto encoded = EncodePointMessage(p);
    });
    
    // Decode benchmark
    auto encoded = EncodePointMessage(p);
    benchmarkNs("DecodePointMessage", [&]() {
        volatile auto decoded = DecodePointMessage(encoded);
    });
    
    // Roundtrip benchmark
    benchmarkNs("Point Roundtrip", [&]() {
        auto enc = EncodePointMessage(p);
        volatile auto dec = DecodePointMessage(enc);
    });
}

void benchmarkRectangle() {
    std::cout << "\n=== Rectangle Benchmarks ===" << std::endl;
    
    Rectangle r;
    r.top_left.x = 10.0;
    r.top_left.y = 20.0;
    r.width = 100.0;
    r.height = 50.0;
    
    // Encode benchmark
    benchmarkNs("EncodeRectangleMessage", [&]() {
        volatile auto encoded = EncodeRectangleMessage(r);
    });
    
    // Decode benchmark
    auto encoded = EncodeRectangleMessage(r);
    benchmarkNs("DecodeRectangleMessage", [&]() {
        volatile auto decoded = DecodeRectangleMessage(encoded);
    });
    
    // Roundtrip benchmark
    benchmarkNs("Rectangle Roundtrip", [&]() {
        auto enc = EncodeRectangleMessage(r);
        volatile auto dec = DecodeRectangleMessage(enc);
    });
}

void benchmarkDispatcher() {
    std::cout << "\n=== Dispatcher Benchmarks ===" << std::endl;
    
    Point p;
    p.x = 3.14;
    p.y = 2.71;
    auto pointMsg = EncodePointMessage(p);
    
    Rectangle r;
    r.top_left.x = 10.0;
    r.top_left.y = 20.0;
    r.width = 100.0;
    r.height = 50.0;
    auto rectMsg = EncodeRectangleMessage(r);
    
    // Dispatcher overhead - Point
    benchmarkNs("DecodeMessage (Point)", [&]() {
        volatile auto variant = DecodeMessage(pointMsg);
    });
    
    // Dispatcher overhead - Rectangle
    benchmarkNs("DecodeMessage (Rectangle)", [&]() {
        volatile auto variant = DecodeMessage(rectMsg);
    });
}

int main() {
    std::cout << "=== C++ Message Mode Benchmarks ===" << std::endl;
    std::cout << "Platform: " << sizeof(void*) * 8 << "-bit" << std::endl;
    
    benchmarkPoint();
    benchmarkRectangle();
    benchmarkDispatcher();
    
    std::cout << "\n=== Benchmarks complete ===" << std::endl;
    return 0;
}
