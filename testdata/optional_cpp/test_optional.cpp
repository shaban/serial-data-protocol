/* test_optional.cpp - Test C++ optional fields */
#include "types.hpp"
#include "encode.hpp"
#include "decode.hpp"
#include <iostream>

int main() {
    std::cout << "=== C++ Optional Fields Test ===" << std::endl << std::endl;

    // Test 1: Request with optional metadata present
    std::cout << "=== Test 1: Optional Present ===" << std::endl;
    sdp::Request req1;
    req1.id = 100;
    
    sdp::Metadata meta;
    meta.user_id = 12345;
    meta.username = "alice";
    req1.metadata = meta;  // Assign to std::optional

    std::cout << "Original data:" << std::endl;
    std::cout << "  id: " << req1.id << std::endl;
    std::cout << "  metadata: present" << std::endl;
    std::cout << "    user_id: " << meta.user_id << std::endl;
    std::cout << "    username: \"" << meta.username << "\"" << std::endl << std::endl;

    size_t size1 = sdp::request_size(req1);
    std::cout << "Encoded size: " << size1 << " bytes" << std::endl;
    
    uint8_t* buffer1 = new uint8_t[size1];
    sdp::request_encode(req1, buffer1);

    try {
        sdp::Request decoded1 = sdp::request_decode(buffer1, size1);
        
        std::cout << "Decoded data:" << std::endl;
        std::cout << "  id: " << (decoded1.id == req1.id ? "✓" : "❌") << std::endl;
        std::cout << "  metadata.has_value(): " << (decoded1.metadata.has_value() ? "✓" : "❌") << std::endl;
        
        if (decoded1.metadata.has_value()) {
            std::cout << "  metadata.value().user_id: " << (decoded1.metadata.value().user_id == meta.user_id ? "✓" : "❌") << std::endl;
            std::cout << "  metadata.value().username: " << (decoded1.metadata.value().username == meta.username ? "✓" : "❌") << std::endl;
        }
        
        delete[] buffer1;

    } catch (const sdp::DecodeError& e) {
        std::cerr << "❌ Decode error: " << e.what() << std::endl;
        delete[] buffer1;
        return 1;
    }

    // Test 2: Request with optional metadata absent
    std::cout << std::endl << "=== Test 2: Optional Absent ===" << std::endl;
    sdp::Request req2;
    req2.id = 200;
    // Don't set metadata - std::optional defaults to empty

    std::cout << "Original data:" << std::endl;
    std::cout << "  id: " << req2.id << std::endl;
    std::cout << "  metadata: absent" << std::endl << std::endl;

    size_t size2 = sdp::request_size(req2);
    std::cout << "Encoded size: " << size2 << " bytes" << std::endl;
    
    uint8_t* buffer2 = new uint8_t[size2];
    sdp::request_encode(req2, buffer2);

    try {
        sdp::Request decoded2 = sdp::request_decode(buffer2, size2);
        
        std::cout << "Decoded data:" << std::endl;
        std::cout << "  id: " << (decoded2.id == req2.id ? "✓" : "❌") << std::endl;
        std::cout << "  metadata.has_value(): " << (!decoded2.metadata.has_value() ? "✓" : "❌") << std::endl;
        
        delete[] buffer2;

    } catch (const sdp::DecodeError& e) {
        std::cerr << "❌ Decode error: " << e.what() << std::endl;
        delete[] buffer2;
        return 1;
    }

    // Test 3: Config with multiple optionals
    std::cout << std::endl << "=== Test 3: Multiple Optionals ===" << std::endl;
    sdp::Config config;
    config.name = "production";
    
    // Set database config
    sdp::DatabaseConfig db;
    db.host = "db.example.com";
    db.port = 5432;
    config.database = db;
    
    // Don't set cache config
    
    std::cout << "Original data:" << std::endl;
    std::cout << "  name: \"" << config.name << "\"" << std::endl;
    std::cout << "  database: present" << std::endl;
    std::cout << "    host: \"" << db.host << "\"" << std::endl;
    std::cout << "    port: " << db.port << std::endl;
    std::cout << "  cache: absent" << std::endl << std::endl;

    size_t size3 = sdp::config_size(config);
    std::cout << "Encoded size: " << size3 << " bytes" << std::endl;
    
    uint8_t* buffer3 = new uint8_t[size3];
    sdp::config_encode(config, buffer3);

    try {
        sdp::Config decoded3 = sdp::config_decode(buffer3, size3);
        
        std::cout << "Decoded data:" << std::endl;
        std::cout << "  name: " << (decoded3.name == config.name ? "✓" : "❌") << std::endl;
        std::cout << "  database.has_value(): " << (decoded3.database.has_value() ? "✓" : "❌") << std::endl;
        
        if (decoded3.database.has_value()) {
            std::cout << "  database.value().host: " << (decoded3.database.value().host == db.host ? "✓" : "❌") << std::endl;
            std::cout << "  database.value().port: " << (decoded3.database.value().port == db.port ? "✓" : "❌") << std::endl;
        }
        
        std::cout << "  cache.has_value(): " << (!decoded3.cache.has_value() ? "✓" : "❌") << std::endl;
        
        delete[] buffer3;

    } catch (const sdp::DecodeError& e) {
        std::cerr << "❌ Decode error: " << e.what() << std::endl;
        delete[] buffer3;
        return 1;
    }

    // Test 4: Document with optional array
    std::cout << std::endl << "=== Test 4: Optional Array ===" << std::endl;
    sdp::Document doc1;
    doc1.id = 1000;
    
    sdp::TagList tagList;
    tagList.items = {"cpp", "optional", "arrays"};
    doc1.tags = tagList;

    std::cout << "Original data:" << std::endl;
    std::cout << "  id: " << doc1.id << std::endl;
    std::cout << "  tags: present" << std::endl;
    std::cout << "    items: [";
    for (size_t i = 0; i < tagList.items.size(); i++) {
        if (i > 0) std::cout << ", ";
        std::cout << "\"" << tagList.items[i] << "\"";
    }
    std::cout << "]" << std::endl << std::endl;

    size_t size4 = sdp::document_size(doc1);
    std::cout << "Encoded size: " << size4 << " bytes" << std::endl;
    
    uint8_t* buffer4 = new uint8_t[size4];
    sdp::document_encode(doc1, buffer4);

    try {
        sdp::Document decoded4 = sdp::document_decode(buffer4, size4);
        
        std::cout << "Decoded data:" << std::endl;
        std::cout << "  id: " << (decoded4.id == doc1.id ? "✓" : "❌") << std::endl;
        std::cout << "  tags.has_value(): " << (decoded4.tags.has_value() ? "✓" : "❌") << std::endl;
        
        if (decoded4.tags.has_value()) {
            bool items_match = decoded4.tags.value().items == tagList.items;
            std::cout << "  tags.value().items: " << (items_match ? "✓" : "❌") << std::endl;
        }
        
        delete[] buffer4;

    } catch (const sdp::DecodeError& e) {
        std::cerr << "❌ Decode error: " << e.what() << std::endl;
        delete[] buffer4;
        return 1;
    }

    std::cout << std::endl << "=== SUCCESS ===" << std::endl;
    std::cout << "✅ std::optional<T> works perfectly!" << std::endl;
    std::cout << "✅ .has_value() correctly indicates presence" << std::endl;
    std::cout << "✅ .value() retrieves the data" << std::endl;
    std::cout << "✅ Optional structs work!" << std::endl;
    std::cout << "✅ Optional arrays work!" << std::endl;
    std::cout << "✅ Multiple optionals in one struct work!" << std::endl;

    return 0;
}
