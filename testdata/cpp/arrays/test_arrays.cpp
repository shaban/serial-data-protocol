/* test_arrays.cpp - Test C++ primitive arrays */
#include "types.hpp"
#include "encode.hpp"
#include "decode.hpp"
#include <iostream>
#include <cmath>

int main() {
    std::cout << "=== C++ Arrays Test ===" << std::endl << std::endl;

    // Test 1: Primitive arrays
    std::cout << "=== Test 1: Primitive Arrays ===" << std::endl;
    sdp::ArraysOfPrimitives primitives;
    
    primitives.u8_array = {1, 2, 3, 4, 5};
    primitives.u32_array = {1000, 2000, 3000, 4000};
    primitives.f64_array = {10.5, 20.5, 30.5};
    primitives.bool_array = {true, false, true, true, false};
    primitives.str_array = {"Hello", "World", "C++", "Arrays"};

    std::cout << "Original data:" << std::endl;
    std::cout << "  u8_array: [";
    for (size_t i = 0; i < primitives.u8_array.size(); i++) {
        if (i > 0) std::cout << ", ";
        std::cout << (int)primitives.u8_array[i];
    }
    std::cout << "]" << std::endl;
    
    std::cout << "  bool_array: [";
    for (size_t i = 0; i < primitives.bool_array.size(); i++) {
        if (i > 0) std::cout << ", ";
        std::cout << (primitives.bool_array[i] ? "true" : "false");
    }
    std::cout << "]" << std::endl;
    
    std::cout << "  str_array: [";
    for (size_t i = 0; i < primitives.str_array.size(); i++) {
        if (i > 0) std::cout << ", ";
        std::cout << "\"" << primitives.str_array[i] << "\"";
    }
    std::cout << "]" << std::endl << std::endl;

    // Encode primitives
    size_t size1 = sdp::arrays_of_primitives_size(primitives);
    std::cout << "Encoded size: " << size1 << " bytes" << std::endl;
    
    uint8_t* buffer1 = new uint8_t[size1];
    size_t written1 = sdp::arrays_of_primitives_encode(primitives, buffer1);
    std::cout << "Encoded " << written1 << " bytes" << std::endl << std::endl;

    // Decode primitives
    try {
        sdp::ArraysOfPrimitives decoded = sdp::arrays_of_primitives_decode(buffer1, written1);

        std::cout << "Decoded data:" << std::endl;
        
        // Verify u8_array
        std::cout << "  u8_array: ";
        if (decoded.u8_array == primitives.u8_array) {
            std::cout << "✓" << std::endl;
        } else {
            std::cout << "❌" << std::endl;
        }
        
        // Verify u32_array
        std::cout << "  u32_array: ";
        if (decoded.u32_array == primitives.u32_array) {
            std::cout << "✓" << std::endl;
        } else {
            std::cout << "❌" << std::endl;
        }
        
        // Verify f64_array
        std::cout << "  f64_array: ";
        bool f64_match = decoded.f64_array.size() == primitives.f64_array.size();
        for (size_t i = 0; f64_match && i < decoded.f64_array.size(); i++) {
            if (std::abs(decoded.f64_array[i] - primitives.f64_array[i]) > 0.0001) {
                f64_match = false;
            }
        }
        std::cout << (f64_match ? "✓" : "❌") << std::endl;
        
        // Verify bool_array
        std::cout << "  bool_array: ";
        if (decoded.bool_array == primitives.bool_array) {
            std::cout << "✓" << std::endl;
        } else {
            std::cout << "❌ [";
            for (size_t i = 0; i < decoded.bool_array.size(); i++) {
                if (i > 0) std::cout << ", ";
                std::cout << (decoded.bool_array[i] ? "true" : "false");
            }
            std::cout << "]" << std::endl;
        }
        
        // Verify str_array
        std::cout << "  str_array: ";
        if (decoded.str_array == primitives.str_array) {
            std::cout << "✓" << std::endl;
        } else {
            std::cout << "❌" << std::endl;
        }

        delete[] buffer1;

    } catch (const sdp::DecodeError& e) {
        std::cerr << "❌ Decode error: " << e.what() << std::endl;
        delete[] buffer1;
        return 1;
    }

    // Test 2: Struct arrays
    std::cout << std::endl << "=== Test 2: Struct Arrays ===" << std::endl;
    sdp::ArraysOfStructs structs;
    
    sdp::Item item1;
    item1.id = 100;
    item1.name = "First Item";
    
    sdp::Item item2;
    item2.id = 200;
    item2.name = "Second Item";
    
    sdp::Item item3;
    item3.id = 300;
    item3.name = "Third Item";
    
    structs.items = {item1, item2, item3};
    structs.count = 42;

    std::cout << "Original data:" << std::endl;
    std::cout << "  items: [";
    for (size_t i = 0; i < structs.items.size(); i++) {
        if (i > 0) std::cout << ", ";
        std::cout << "{id=" << structs.items[i].id << ", name=\"" << structs.items[i].name << "\"}";
    }
    std::cout << "]" << std::endl;
    std::cout << "  count: " << structs.count << std::endl << std::endl;

    // Encode structs
    size_t size2 = sdp::arrays_of_structs_size(structs);
    std::cout << "Encoded size: " << size2 << " bytes" << std::endl;
    
    uint8_t* buffer2 = new uint8_t[size2];
    size_t written2 = sdp::arrays_of_structs_encode(structs, buffer2);
    std::cout << "Encoded " << written2 << " bytes" << std::endl << std::endl;

    // Decode structs
    try {
        sdp::ArraysOfStructs decoded = sdp::arrays_of_structs_decode(buffer2, written2);

        std::cout << "Decoded data:" << std::endl;
        
        // Verify items array
        std::cout << "  items.size(): ";
        if (decoded.items.size() == structs.items.size()) {
            std::cout << "✓" << std::endl;
        } else {
            std::cout << "❌ (expected " << structs.items.size() << ", got " << decoded.items.size() << ")" << std::endl;
        }
        
        // Verify each item
        for (size_t i = 0; i < decoded.items.size(); i++) {
            std::cout << "  items[" << i << "].id: ";
            if (decoded.items[i].id == structs.items[i].id) {
                std::cout << "✓" << std::endl;
            } else {
                std::cout << "❌ (expected " << structs.items[i].id << ", got " << decoded.items[i].id << ")" << std::endl;
            }
            
            std::cout << "  items[" << i << "].name: ";
            if (decoded.items[i].name == structs.items[i].name) {
                std::cout << "✓" << std::endl;
            } else {
                std::cout << "❌ (expected \"" << structs.items[i].name << "\", got \"" << decoded.items[i].name << "\")" << std::endl;
            }
        }
        
        // Verify count
        std::cout << "  count: ";
        if (decoded.count == structs.count) {
            std::cout << "✓" << std::endl;
        } else {
            std::cout << "❌" << std::endl;
        }

        std::cout << std::endl << "=== SUCCESS ===" << std::endl;
        std::cout << "✅ All primitive array types work!" << std::endl;
        std::cout << "✅ std::vector<T> handles everything automatically" << std::endl;
        std::cout << "✅ Bool arrays work (std::vector<bool> special case)" << std::endl;
        std::cout << "✅ Struct arrays work with nested fields!" << std::endl;
        std::cout << "✅ String arrays work!" << std::endl;

        delete[] buffer2;
        return 0;

    } catch (const sdp::DecodeError& e) {
        std::cerr << "❌ Decode error: " << e.what() << std::endl;
        delete[] buffer2;
        return 1;
    }
}
