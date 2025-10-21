/* Handwritten Optimized Encoder: primitives.sdp
 *
 * Schema:
 * struct AllPrimitives {
 *     u8_field: u8,
 *     u16_field: u16,
 *     u32_field: u32,
 *     u64_field: u64,
 *     i8_field: i8,
 *     i16_field: i16,
 *     i32_field: i32,
 *     i64_field: i64,
 *     f32_field: f32,
 *     f64_field: f64,
 *     bool_field: bool,
 *     str_field: str
 * }
 *
 * Optimizations applied:
 * 1. Wire format struct for fixed primitives (bulk copy)
 * 2. Pre-computed string length (caller provides)
 * 3. Direct pointer writes
 * 4. Single capacity check
 */

#include <stdio.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

#define ITERATIONS 10000000

/* User-facing struct */
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
    const char* str_field;
    size_t str_field_len;
} AllPrimitives;

/* Wire format (packed, little-endian) */
typedef struct __attribute__((packed)) {
    uint8_t u8_field;
    uint16_t u16_field;
    uint32_t u32_field;
    uint64_t u64_field;
    int8_t i8_field;
    int16_t i16_field;
    int32_t i32_field;
    int64_t i64_field;
    uint32_t f32_wire;
    uint64_t f64_wire;
    uint8_t bool_field;
} AllPrimitivesWire;

#define FIXED_SIZE 43  // Size without string

static inline uint32_t f32_to_wire(float f) {
    uint32_t u;
    memcpy(&u, &f, 4);
    return u;
}

static inline uint64_t f64_to_wire(double d) {
    uint64_t u;
    memcpy(&u, &d, 8);
    return u;
}

/* BASELINE: Field-by-field (current approach) */
static size_t encode_baseline(const AllPrimitives* src, uint8_t* buf) {
    size_t offset = 0;
    
    buf[offset++] = src->u8_field;
    memcpy(buf + offset, &src->u16_field, 2); offset += 2;
    memcpy(buf + offset, &src->u32_field, 4); offset += 4;
    memcpy(buf + offset, &src->u64_field, 8); offset += 8;
    buf[offset++] = (uint8_t)src->i8_field;
    memcpy(buf + offset, &src->i16_field, 2); offset += 2;
    memcpy(buf + offset, &src->i32_field, 4); offset += 4;
    memcpy(buf + offset, &src->i64_field, 8); offset += 8;
    
    uint32_t f32w = f32_to_wire(src->f32_field);
    memcpy(buf + offset, &f32w, 4); offset += 4;
    
    uint64_t f64w = f64_to_wire(src->f64_field);
    memcpy(buf + offset, &f64w, 8); offset += 8;
    
    buf[offset++] = src->bool_field;
    
    uint32_t str_len = (uint32_t)src->str_field_len;
    memcpy(buf + offset, &str_len, 4); offset += 4;
    memcpy(buf + offset, src->str_field, src->str_field_len);
    offset += src->str_field_len;
    
    return offset;
}

/* OPTIMIZED: Bulk copy with wire struct */
static size_t encode_optimized(const AllPrimitives* src, uint8_t* buf) {
    // Prepare wire struct on stack
    AllPrimitivesWire wire = {
        .u8_field = src->u8_field,
        .u16_field = src->u16_field,
        .u32_field = src->u32_field,
        .u64_field = src->u64_field,
        .i8_field = src->i8_field,
        .i16_field = src->i16_field,
        .i32_field = src->i32_field,
        .i64_field = src->i64_field,
        .f32_wire = f32_to_wire(src->f32_field),
        .f64_wire = f64_to_wire(src->f64_field),
        .bool_field = src->bool_field,
    };
    
    // Bulk copy fixed portion (43 bytes)
    memcpy(buf, &wire, FIXED_SIZE);
    
    // String (variable-length)
    *(uint32_t*)(buf + FIXED_SIZE) = (uint32_t)src->str_field_len;
    memcpy(buf + FIXED_SIZE + 4, src->str_field, src->str_field_len);
    
    return FIXED_SIZE + 4 + src->str_field_len;
}

static uint64_t get_nanos(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
}

int main(void) {
    printf("Primitives Schema Benchmark\n");
    printf("============================\n\n");
    
    AllPrimitives data = {
        .u8_field = 255,
        .u16_field = 65535,
        .u32_field = 4294967295,
        .u64_field = 18446744073709551615ULL,
        .i8_field = -128,
        .i16_field = -32768,
        .i32_field = -2147483648,
        .i64_field = -9223372036854775807LL - 1,
        .f32_field = 3.14159f,
        .f64_field = 2.718281828459045,
        .bool_field = 1,
        .str_field = "Hello, World!",
        .str_field_len = 13
    };
    
    uint8_t buf[256];
    volatile size_t size;
    
    // Verify both produce same output
    uint8_t buf1[256], buf2[256];
    size_t s1 = encode_baseline(&data, buf1);
    size_t s2 = encode_optimized(&data, buf2);
    
    if (s1 != s2 || memcmp(buf1, buf2, s1) != 0) {
        printf("ERROR: Output mismatch!\n");
        return 1;
    }
    
    printf("✓ Both encoders produce identical output (%zu bytes)\n", s1);
    printf("  Fixed size: %d bytes\n", FIXED_SIZE);
    printf("  String: %zu bytes (4 + %zu)\n\n", 4 + data.str_field_len, data.str_field_len);
    
    // Benchmark baseline
    printf("Baseline (field-by-field):\n");
    uint64_t start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        size = encode_baseline(&data, buf);
    }
    uint64_t end = get_nanos();
    double baseline_time = (double)(end - start) / ITERATIONS;
    printf("  %.2f ns/op\n\n", baseline_time);
    
    // Benchmark optimized
    printf("Optimized (wire struct + bulk copy):\n");
    start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        size = encode_optimized(&data, buf);
    }
    end = get_nanos();
    double optimized_time = (double)(end - start) / ITERATIONS;
    printf("  %.2f ns/op\n", optimized_time);
    printf("  %.1fx faster\n\n", baseline_time / optimized_time);
    
    printf("Speedup: %.0f%% improvement\n", ((baseline_time - optimized_time) / baseline_time) * 100);
    
    return 0;
}
