// Test C++ message mode implementation
#include "message_encode.hpp"
#include "message_decode.hpp"
#include <iostream>
#include <cassert>
#include <cmath>

using namespace sdp;

void testPointMessage() {
    std::cout << "Testing Point message mode..." << std::endl;
    
    // Create a Point
    Point p;
    p.x = 3.14;
    p.y = 2.71;
    
    // Encode to message
    std::vector<uint8_t> encoded = EncodePointMessage(p);
    
    // Verify header size
    assert(encoded.size() >= MESSAGE_HEADER_SIZE);
    std::cout << "  Encoded size: " << encoded.size() << " bytes" << std::endl;
    
    // Verify magic bytes
    assert(encoded[0] == 'S' && encoded[1] == 'D' && encoded[2] == 'P');
    std::cout << "  Magic bytes: OK" << std::endl;
    
    // Verify version
    assert(encoded[3] == MESSAGE_VERSION);
    std::cout << "  Version: " << (int)encoded[3] << std::endl;
    
    // Decode using specific decoder
    Point decoded = DecodePointMessage(encoded);
    
    // Verify values
    assert(std::abs(decoded.x - 3.14) < 0.0001);
    assert(std::abs(decoded.y - 2.71) < 0.0001);
    std::cout << "  Decoded: x=" << decoded.x << ", y=" << decoded.y << std::endl;
    
    // Decode using dispatcher
    MessageVariant variant = DecodeMessage(encoded);
    Point* dispatchedPoint = std::get_if<Point>(&variant);
    assert(dispatchedPoint != nullptr);
    assert(std::abs(dispatchedPoint->x - 3.14) < 0.0001);
    assert(std::abs(dispatchedPoint->y - 2.71) < 0.0001);
    std::cout << "  Dispatcher: OK" << std::endl;
    
    std::cout << "✓ Point message mode test passed" << std::endl << std::endl;
}

void testRectangleMessage() {
    std::cout << "Testing Rectangle message mode..." << std::endl;
    
    // Create a Rectangle
    Rectangle r;
    r.top_left.x = 10.0;
    r.top_left.y = 20.0;
    r.width = 100.0;
    r.height = 50.0;
    
    // Encode to message
    std::vector<uint8_t> encoded = EncodeRectangleMessage(r);
    
    // Verify header size
    assert(encoded.size() >= MESSAGE_HEADER_SIZE);
    std::cout << "  Encoded size: " << encoded.size() << " bytes" << std::endl;
    
    // Decode using specific decoder
    Rectangle decoded = DecodeRectangleMessage(encoded);
    
    // Verify values
    assert(std::abs(decoded.top_left.x - 10.0) < 0.0001);
    assert(std::abs(decoded.top_left.y - 20.0) < 0.0001);
    assert(std::abs(decoded.width - 100.0) < 0.0001);
    assert(std::abs(decoded.height - 50.0) < 0.0001);
    std::cout << "  Decoded: top_left=(" << decoded.top_left.x << "," << decoded.top_left.y << "), "
              << "size=" << decoded.width << "x" << decoded.height << std::endl;
    
    // Decode using dispatcher
    MessageVariant variant = DecodeMessage(encoded);
    Rectangle* dispatchedRect = std::get_if<Rectangle>(&variant);
    assert(dispatchedRect != nullptr);
    assert(std::abs(dispatchedRect->width - 100.0) < 0.0001);
    std::cout << "  Dispatcher: OK" << std::endl;
    
    std::cout << "✓ Rectangle message mode test passed" << std::endl << std::endl;
}

void testWrongTypeID() {
    std::cout << "Testing wrong type ID error..." << std::endl;
    
    // Create a Point message
    Point p;
    p.x = 1.0;
    p.y = 2.0;
    std::vector<uint8_t> encoded = EncodePointMessage(p);
    
    // Try to decode as Rectangle - should throw
    try {
        [[maybe_unused]] Rectangle r = DecodeRectangleMessage(encoded);
        assert(false && "Should have thrown MessageDecodeError");
    } catch (const MessageDecodeError& e) {
        std::cout << "  Caught expected error: " << e.what() << std::endl;
    }
    
    std::cout << "✓ Wrong type ID test passed" << std::endl << std::endl;
}

int main() {
    std::cout << "=== C++ Message Mode Tests ===" << std::endl << std::endl;
    
    try {
        testPointMessage();
        testRectangleMessage();
        testWrongTypeID();
        
        std::cout << "=== All tests passed! ===" << std::endl;
        return 0;
    } catch (const std::exception& e) {
        std::cerr << "ERROR: " << e.what() << std::endl;
        return 1;
    }
}
