// encode_reference.cpp - Generate reference .sdpb files for cross-language testing
#include "message_encode.hpp"
#include <iostream>
#include <fstream>
#include <vector>
#include <cstdint>

void write_file(const std::string& path, const std::vector<uint8_t>& data) {
    std::ofstream file(path, std::ios::binary);
    if (!file) {
        std::cerr << "Error: Cannot write to " << path << std::endl;
        exit(1);
    }
    file.write(reinterpret_cast<const char*>(data.data()), data.size());
    file.close();
    std::cout << "Created " << path << " (" << data.size() << " bytes)" << std::endl;
}

int main() {
    try {
        // Encode Point message (same values as Go)
        sdp::Point point;
        point.x = 3.14;
        point.y = 2.71;
        
        std::vector<uint8_t> point_data = sdp::EncodePointMessage(point);
        write_file("../../binaries/message_point_cpp.sdpb", point_data);
        
        // Encode Rectangle message (same values as Go)
        sdp::Rectangle rect;
        rect.top_left.x = 10.0;
        rect.top_left.y = 20.0;
        rect.width = 100.0;
        rect.height = 50.0;
        
        std::vector<uint8_t> rect_data = sdp::EncodeRectangleMessage(rect);
        write_file("../../binaries/message_rectangle_cpp.sdpb", rect_data);
        
        std::cout << "\nC++ reference .sdpb files generated successfully!" << std::endl;
        return 0;
        
    } catch (const std::exception& e) {
        std::cerr << "Error: " << e.what() << std::endl;
        return 1;
    }
}
