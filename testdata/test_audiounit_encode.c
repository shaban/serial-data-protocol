/* test_audiounit_encode.c - Encode AudioUnit plugin for cross-language testing */

#include <stdio.h>
#include "../testdata/audiounit/c/types.h"
#include "../testdata/audiounit/c/encode.h"

int main(void) {
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
    
    return 0;
}
