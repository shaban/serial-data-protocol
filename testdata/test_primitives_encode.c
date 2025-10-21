/* test_primitives_encode.c - Encode primitives for cross-language testing */

#include <stdio.h>
#include "../testdata/primitives/c/types.h"
#include "../testdata/primitives/c/encode.h"

int main(void) {
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
    
    return 0;
}
