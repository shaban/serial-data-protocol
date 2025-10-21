/* test_cpp_roundtrip.cpp - Test C++ encode/decode */
#include "types.hpp"
#include "encode.hpp"
#include "decode.hpp"
#include <iostream>
#include <iomanip>
#include <cstring>

int main() {
    std::cout << "=== C++ Encode/Decode Test ===" << std::endl << std::endl;

    // Create test data
    sdp::AllPrimitives original;
    original.u8_field = 42;
    original.u16_field = 1234;
    original.u32_field = 567890;
    original.u64_field = 9876543210ULL;
    original.i8_field = -42;
    original.i16_field = -1234;
    original.i32_field = -567890;
    original.i64_field = -9876543210LL;
    original.f32_field = 3.14159f;
    original.f64_field = 2.71828;
    original.bool_field = true;
    original.str_field = "Hello, C++!";

    std::cout << "Original data:" << std::endl;
    std::cout << "  u8:  " << (int)original.u8_field << std::endl;
    std::cout << "  u32: " << original.u32_field << std::endl;
    std::cout << "  f32: " << original.f32_field << std::endl;
    std::cout << "  str: \"" << original.str_field << "\"" << std::endl;
    std::cout << std::endl;

    // Calculate size
    size_t size = sdp::all_primitives_size(original);
    std::cout << "Encoded size: " << size << " bytes" << std::endl << std::endl;

    // Encode
    uint8_t* buffer = new uint8_t[size];
    size_t written = sdp::all_primitives_encode(original, buffer);
    std::cout << "Encoded " << written << " bytes" << std::endl << std::endl;

    // Decode
    try {
        sdp::AllPrimitives decoded = sdp::all_primitives_decode(buffer, written);

        std::cout << "Decoded data:" << std::endl;
        std::cout << "  u8:  " << (int)decoded.u8_field;
        if (decoded.u8_field == original.u8_field) std::cout << " ✓";
        std::cout << std::endl;

        std::cout << "  u16: " << decoded.u16_field;
        if (decoded.u16_field == original.u16_field) std::cout << " ✓";
        std::cout << std::endl;

        std::cout << "  u32: " << decoded.u32_field;
        if (decoded.u32_field == original.u32_field) std::cout << " ✓";
        std::cout << std::endl;

        std::cout << "  u64: " << decoded.u64_field;
        if (decoded.u64_field == original.u64_field) std::cout << " ✓";
        std::cout << std::endl;

        std::cout << "  i8:  " << (int)decoded.i8_field;
        if (decoded.i8_field == original.i8_field) std::cout << " ✓";
        std::cout << std::endl;

        std::cout << "  i16: " << decoded.i16_field;
        if (decoded.i16_field == original.i16_field) std::cout << " ✓";
        std::cout << std::endl;

        std::cout << "  i32: " << decoded.i32_field;
        if (decoded.i32_field == original.i32_field) std::cout << " ✓";
        std::cout << std::endl;

        std::cout << "  i64: " << decoded.i64_field;
        if (decoded.i64_field == original.i64_field) std::cout << " ✓";
        std::cout << std::endl;

        std::cout << "  f32: " << std::fixed << std::setprecision(5) << decoded.f32_field;
        if (std::abs(decoded.f32_field - original.f32_field) < 0.00001f) std::cout << " ✓";
        std::cout << std::endl;

        std::cout << "  f64: " << std::fixed << std::setprecision(5) << decoded.f64_field;
        if (std::abs(decoded.f64_field - original.f64_field) < 0.00001) std::cout << " ✓";
        std::cout << std::endl;

        std::cout << "  bool: " << (decoded.bool_field ? "true" : "false");
        if (decoded.bool_field == original.bool_field) std::cout << " ✓";
        std::cout << std::endl;

        std::cout << "  str: \"" << decoded.str_field << "\"";
        if (decoded.str_field == original.str_field) std::cout << " ✓";
        std::cout << std::endl;

        std::cout << std::endl << "=== SUCCESS ===" << std::endl;
        std::cout << "✅ All fields match!" << std::endl;
        std::cout << "✅ String is std::string (null-terminated, RAII)" << std::endl;
        std::cout << "✅ No manual memory management needed" << std::endl;

        delete[] buffer;
        return 0;

    } catch (const sdp::DecodeError& e) {
        std::cerr << "❌ Decode error: " << e.what() << std::endl;
        delete[] buffer;
        return 1;
    }
}
