/* decode_primitives_zerocopy.c - Zero-copy decoder (strings point into buffer)
 * 
 * Approach: No allocations, strings are pointers into original buffer.
 * Pros: Maximum performance, no memory management
 * Cons: Decoded struct lifetime tied to buffer, can't modify strings
 * 
 * Compile: gcc -std=c99 -O3 -march=native decode_primitives_zerocopy.c -o decode_primitives_zerocopy
 * Benchmark: time for i in {1..1000000}; do ./decode_primitives_zerocopy; done
 */

#include <stdio.h>
#include <stdint.h>
#include <string.h>
#include <time.h>

/* Endianness helpers (same as encoder) */
#if defined(__linux__) || defined(__CYGWIN__)
    #include <endian.h>
    #define LE16TOH(x) le16toh(x)
    #define LE32TOH(x) le32toh(x)
    #define LE64TOH(x) le64toh(x)
#elif defined(__APPLE__)
    #include <libkern/OSByteOrder.h>
    #define LE16TOH(x) OSSwapLittleToHostInt16(x)
    #define LE32TOH(x) OSSwapLittleToHostInt32(x)
    #define LE64TOH(x) OSSwapLittleToHostInt64(x)
#else
    /* Generic fallback */
    static inline uint16_t le16toh_fallback(uint16_t x) {
        return ((x & 0xFF) << 8) | ((x >> 8) & 0xFF);
    }
    static inline uint32_t le32toh_fallback(uint32_t x) {
        return ((x & 0xFF) << 24) | ((x & 0xFF00) << 8) |
               ((x >> 8) & 0xFF00) | ((x >> 24) & 0xFF);
    }
    #define LE16TOH(x) le16toh_fallback(x)
    #define LE32TOH(x) le32toh_fallback(x)
    #define LE64TOH(x) /* implement if needed */
#endif

/* Float conversion helpers */
static inline float le_to_f32(const uint8_t* buf) {
    uint32_t u;
    memcpy(&u, buf, 4);
    u = LE32TOH(u);
    float f;
    memcpy(&f, &u, 4);
    return f;
}

static inline double le_to_f64(const uint8_t* buf) {
    uint64_t u;
    memcpy(&u, buf, 8);
    u = LE64TOH(u);
    double d;
    memcpy(&d, &u, 8);
    return d;
}

/* Zero-copy decoded struct */
typedef struct {
    uint8_t u8_field;
    uint16_t u16_field;
    uint32_t u32_field;
    uint64_t u64_field;
    int8_t i8_field;
    int16_t i16_field;
    int32_t i32_field;
    int64_t i64_field;
    float f32_field;
    double f64_field;
    uint8_t bool_field;
    const char* str_field;  /* Points into buffer! */
    size_t str_field_len;
} AllPrimitives;

/* Decode from buffer (zero-copy) */
int decode_all_primitives(AllPrimitives* dest, const uint8_t* buf, size_t buf_len) {
    size_t offset = 0;
    
    /* Bounds check helper */
    #define CHECK_BOUNDS(n) if (offset + (n) > buf_len) return -1
    
    /* u8 */
    CHECK_BOUNDS(1);
    dest->u8_field = buf[offset];
    offset += 1;
    
    /* u16 */
    CHECK_BOUNDS(2);
    dest->u16_field = LE16TOH(*(const uint16_t*)(buf + offset));
    offset += 2;
    
    /* u32 */
    CHECK_BOUNDS(4);
    dest->u32_field = LE32TOH(*(const uint32_t*)(buf + offset));
    offset += 4;
    
    /* u64 */
    CHECK_BOUNDS(8);
    dest->u64_field = LE64TOH(*(const uint64_t*)(buf + offset));
    offset += 8;
    
    /* i8 */
    CHECK_BOUNDS(1);
    dest->i8_field = (int8_t)buf[offset];
    offset += 1;
    
    /* i16 */
    CHECK_BOUNDS(2);
    dest->i16_field = (int16_t)LE16TOH(*(const uint16_t*)(buf + offset));
    offset += 2;
    
    /* i32 */
    CHECK_BOUNDS(4);
    dest->i32_field = (int32_t)LE32TOH(*(const uint32_t*)(buf + offset));
    offset += 4;
    
    /* i64 */
    CHECK_BOUNDS(8);
    dest->i64_field = (int64_t)LE64TOH(*(const uint64_t*)(buf + offset));
    offset += 8;
    
    /* f32 */
    CHECK_BOUNDS(4);
    dest->f32_field = le_to_f32(buf + offset);
    offset += 4;
    
    /* f64 */
    CHECK_BOUNDS(8);
    dest->f64_field = le_to_f64(buf + offset);
    offset += 8;
    
    /* bool */
    CHECK_BOUNDS(1);
    dest->bool_field = buf[offset];
    offset += 1;
    
    /* string (zero-copy: pointer into buffer) */
    CHECK_BOUNDS(4);
    uint32_t str_len = LE32TOH(*(const uint32_t*)(buf + offset));
    offset += 4;
    
    CHECK_BOUNDS(str_len);
    dest->str_field = (const char*)(buf + offset);
    dest->str_field_len = str_len;
    offset += str_len;
    
    #undef CHECK_BOUNDS
    return 0;
}

/* Test data (encoded wire format) */
static const uint8_t test_data[] = {
    42,                              /* u8: 42 */
    0xe8, 0x03,                      /* u16: 1000 (little-endian) */
    0xa0, 0x86, 0x01, 0x00,          /* u32: 100000 */
    0xcb, 0x04, 0xfb, 0x71, 0x1f, 0x01, 0x00, 0x00,  /* u64: 1234567890123 */
    0xf6,                            /* i8: -10 */
    0x18, 0xfc,                      /* i16: -1000 */
    0x60, 0x79, 0xfe, 0xff,          /* i32: -100000 */
    0x16, 0xe9, 0x4f, 0xb3, 0xfd, 0xff, 0xff, 0xff,  /* i64: -9876543210 */
    0xd0, 0x0f, 0x49, 0x40,          /* f32: 3.14159 */
    0x90, 0xf7, 0xaa, 0x95, 0x09, 0xbf, 0x05, 0x40,  /* f64: 2.71828 */
    0x01,                            /* bool: true */
    0x05, 0x00, 0x00, 0x00,          /* string length: 5 */
    'h', 'e', 'l', 'l', 'o'          /* string data */
};

/* Benchmark */
int main(void) {
    AllPrimitives decoded;
    
    /* Warmup */
    for (int i = 0; i < 1000; i++) {
        if (decode_all_primitives(&decoded, test_data, sizeof(test_data)) != 0) {
            fprintf(stderr, "Decode failed\n");
            return 1;
        }
    }
    
    /* Benchmark */
    const int iterations = 10000000;
    volatile uint32_t sink = 0;  /* Prevent optimization */
    
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    
    for (int i = 0; i < iterations; i++) {
        if (decode_all_primitives(&decoded, test_data, sizeof(test_data)) != 0) {
            fprintf(stderr, "Decode failed\n");
            return 1;
        }
        sink += decoded.u32_field;  /* Use result to prevent dead code elimination */
    }
    
    clock_gettime(CLOCK_MONOTONIC, &end);
    
    uint64_t start_ns = start.tv_sec * 1000000000ULL + start.tv_nsec;
    uint64_t end_ns = end.tv_sec * 1000000000ULL + end.tv_nsec;
    uint64_t total_ns = end_ns - start_ns;
    double ns_per_op = (double)total_ns / iterations;
    
    printf("=== Zero-Copy Decode (Primitives) ===\n");
    printf("Iterations: %d\n", iterations);
    printf("Total time: %.2f ms\n", total_ns / 1e6);
    printf("Time per op: %.2f ns\n", ns_per_op);
    printf("Throughput: %.2f million ops/sec\n", 1000.0 / ns_per_op);
    
    /* Verify correctness */
    printf("\nVerification:\n");
    printf("  u8_field: %u (expected 42)\n", decoded.u8_field);
    printf("  u16_field: %u (expected 1000)\n", decoded.u16_field);
    printf("  u32_field: %u (expected 100000)\n", decoded.u32_field);
    printf("  f32_field: %.5f (expected 3.14159)\n", decoded.f32_field);
    printf("  bool_field: %u (expected 1)\n", decoded.bool_field);
    printf("  str_field: '%.*s' (expected 'hello')\n", 
           (int)decoded.str_field_len, decoded.str_field);
    
    return 0;
}
