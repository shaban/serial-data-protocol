/* test_c_encoder.c - C encoder for cross-language compatibility tests
 * 
 * Encodes test data and writes to stdout for verification
 * Usage: ./test_c_encoder <schema> > output.bin
 */

#include <stdio.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

int main(int argc, char** argv) {
    if (argc != 2) {
        fprintf(stderr, "Usage: %s <schema>\n", argv[0]);
        fprintf(stderr, "Schemas: primitives|audiounit|optional\n");
        return 1;
    }
    
    const char* schema = argv[1];
    
    if (strcmp(schema, "primitives") == 0) {
        #include "../testdata/primitives/c/types.h"
        #include "../testdata/primitives/c/encode.h"
        
        SDPAllPrimitives prim = {
            .u8_field = 42,
            .u16_field = 1000,
            .u32_field = 100000,
            .u64_field = 1234567890123ULL,
            .i8_field = -10,
            .i16_field = -1000,
            .i32_field = -100000,
            .i64_field = -9876543210LL,
            .f32_field = 3.14159f,
            .f64_field = 2.71828,
            .bool_field = 1,
            .str_field = "hello",
            .str_field_len = 5
        };
        
        uint8_t buf[256];
        size_t size = sdp_all_primitives_encode(&prim, buf);
        fwrite(buf, 1, size, stdout);
        
    } else if (strcmp(schema, "audiounit") == 0) {
        #include "../testdata/audiounit/c/types.h"
        #include "../testdata/audiounit/c/encode.h"
        
        SDPParameter params[2] = {
            {
                .address = 0x1000,
                .display_name = "Volume",
                .display_name_len = 6,
                .identifier = "vol",
                .identifier_len = 3,
                .unit = "dB",
                .unit_len = 2,
                .min_value = -96.0f,
                .max_value = 6.0f,
                .default_value = 0.0f,
                .current_value = -3.0f,
                .raw_flags = 0x01,
                .is_writable = 1,
                .can_ramp = 1
            },
            {
                .address = 0x2000,
                .display_name = "Pan",
                .display_name_len = 3,
                .identifier = "pan",
                .identifier_len = 3,
                .unit = "%",
                .unit_len = 1,
                .min_value = -100.0f,
                .max_value = 100.0f,
                .default_value = 0.0f,
                .current_value = 0.0f,
                .raw_flags = 0x02,
                .is_writable = 1,
                .can_ramp = 1
            }
        };
        
        SDPPlugin plugin = {
            .name = "TestPlugin",
            .name_len = 10,
            .manufacturer_id = "ACME",
            .manufacturer_id_len = 4,
            .component_type = "aufx",
            .component_type_len = 4,
            .component_subtype = "test",
            .component_subtype_len = 4,
            .parameters = params,
            .parameters_len = 2
        };
        
        uint8_t buf[1024];
        size_t size = sdp_plugin_encode(&plugin, buf);
        fwrite(buf, 1, size, stdout);
        
    } else if (strcmp(schema, "optional") == 0) {
        #include "../testdata/optional/c/types.h"
        #include "../testdata/optional/c/encode.h"
        
        SDPMetadata meta = {
            .user_id = 12345,
            .username = "testuser",
            .username_len = 8
        };
        
        SDPRequest req = {
            .request_id = 99,
            .body = "test body",
            .body_len = 9,
            .metadata = &meta  /* Present */
        };
        
        uint8_t buf[256];
        size_t size = sdp_request_encode(&req, buf);
        fwrite(buf, 1, size, stdout);
        
    } else {
        fprintf(stderr, "Unknown schema: %s\n", schema);
        return 1;
    }
    
    return 0;
}
