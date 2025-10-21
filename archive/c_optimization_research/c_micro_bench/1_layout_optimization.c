/* Test 1: Layout Data - Sizes, Offsets, Padding
 *
 * Goal: Extract compile-time layout information for optimization
 * 
 * For a struct like AllPrimitives (12 fields):
 * - What's the fixed-size portion?
 * - What are the field offsets in wire format?
 * - Can we pre-compute the size?
 * - Can we use bulk memcpy?
 */

#include <stdio.h>
#include <stdint.h>
#include <stddef.h>
#include <string.h>
#include <time.h>

#define ITERATIONS 10000000

/* Example: AllPrimitives from primitives.sdp */
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

/* Wire format layout (little-endian, no padding) */
typedef struct __attribute__((packed)) {
    uint8_t u8_field;       // offset 0
    uint16_t u16_field;     // offset 1
    uint32_t u32_field;     // offset 3
    uint64_t u64_field;     // offset 7
    int8_t i8_field;        // offset 15
    int16_t i16_field;      // offset 16
    int32_t i32_field;      // offset 18
    int64_t i64_field;      // offset 22
    uint32_t f32_wire;      // offset 30 (float as uint32)
    uint64_t f64_wire;      // offset 34 (double as uint64)
    uint8_t bool_field;     // offset 42
    // String: offset 43, length prefix (4 bytes) + data (variable)
} AllPrimitivesWire;

/* Compile-time constants */
#define WIRE_FIXED_SIZE 43  // All fields except string
#define WIRE_U8_OFFSET 0
#define WIRE_U16_OFFSET 1
#define WIRE_U32_OFFSET 3
#define WIRE_U64_OFFSET 7
#define WIRE_I8_OFFSET 15
#define WIRE_I16_OFFSET 16
#define WIRE_I32_OFFSET 18
#define WIRE_I64_OFFSET 22
#define WIRE_F32_OFFSET 30
#define WIRE_F64_OFFSET 34
#define WIRE_BOOL_OFFSET 42
#define WIRE_STR_OFFSET 43

/* Helper: float to wire format */
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

/* Approach 1: Field-by-field encoding (current approach) */
static size_t encode_field_by_field(const AllPrimitives* src, uint8_t* buf) {
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

/* Approach 2: Bulk copy with pre-computed layout */
static size_t encode_bulk_copy(const AllPrimitives* src, uint8_t* buf) {
    // Prepare wire format struct on stack
    AllPrimitivesWire wire;
    wire.u8_field = src->u8_field;
    wire.u16_field = src->u16_field;
    wire.u32_field = src->u32_field;
    wire.u64_field = src->u64_field;
    wire.i8_field = src->i8_field;
    wire.i16_field = src->i16_field;
    wire.i32_field = src->i32_field;
    wire.i64_field = src->i64_field;
    wire.f32_wire = f32_to_wire(src->f32_field);
    wire.f64_wire = f64_to_wire(src->f64_field);
    wire.bool_field = src->bool_field;
    
    // Bulk copy fixed portion
    memcpy(buf, &wire, WIRE_FIXED_SIZE);
    
    // String (variable-length)
    uint32_t str_len = (uint32_t)src->str_field_len;
    memcpy(buf + WIRE_FIXED_SIZE, &str_len, 4);
    memcpy(buf + WIRE_FIXED_SIZE + 4, src->str_field, src->str_field_len);
    
    return WIRE_FIXED_SIZE + 4 + src->str_field_len;
}

/* Approach 3: Direct write with offsets */
static size_t encode_direct_offsets(const AllPrimitives* src, uint8_t* buf) {
    // Write directly to pre-computed offsets
    buf[WIRE_U8_OFFSET] = src->u8_field;
    *(uint16_t*)(buf + WIRE_U16_OFFSET) = src->u16_field;
    *(uint32_t*)(buf + WIRE_U32_OFFSET) = src->u32_field;
    *(uint64_t*)(buf + WIRE_U64_OFFSET) = src->u64_field;
    buf[WIRE_I8_OFFSET] = (uint8_t)src->i8_field;
    *(int16_t*)(buf + WIRE_I16_OFFSET) = src->i16_field;
    *(int32_t*)(buf + WIRE_I32_OFFSET) = src->i32_field;
    *(int64_t*)(buf + WIRE_I64_OFFSET) = src->i64_field;
    *(uint32_t*)(buf + WIRE_F32_OFFSET) = f32_to_wire(src->f32_field);
    *(uint64_t*)(buf + WIRE_F64_OFFSET) = f64_to_wire(src->f64_field);
    buf[WIRE_BOOL_OFFSET] = src->bool_field;
    
    // String
    uint32_t str_len = (uint32_t)src->str_field_len;
    *(uint32_t*)(buf + WIRE_STR_OFFSET) = str_len;
    memcpy(buf + WIRE_STR_OFFSET + 4, src->str_field, src->str_field_len);
    
    return WIRE_FIXED_SIZE + 4 + src->str_field_len;
}

static uint64_t get_nanos(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
}

int main(void) {
    printf("Test 1: Layout-based Optimization\n");
    printf("==================================\n\n");
    
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
    volatile size_t size;  // Prevent optimization
    
    printf("Layout Analysis:\n");
    printf("  C struct size:     %zu bytes\n", sizeof(AllPrimitives));
    printf("  Wire fixed size:   %d bytes\n", WIRE_FIXED_SIZE);
    printf("  Wire total size:   %d bytes (with 13-byte string)\n\n", WIRE_FIXED_SIZE + 4 + 13);
    
    // Verify all methods produce same output
    uint8_t buf1[256], buf2[256], buf3[256];
    size_t s1 = encode_field_by_field(&data, buf1);
    size_t s2 = encode_bulk_copy(&data, buf2);
    size_t s3 = encode_direct_offsets(&data, buf3);
    
    if (s1 != s2 || s1 != s3) {
        printf("ERROR: Size mismatch! %zu vs %zu vs %zu\n", s1, s2, s3);
        return 1;
    }
    
    if (memcmp(buf1, buf2, s1) != 0 || memcmp(buf1, buf3, s1) != 0) {
        printf("ERROR: Output mismatch!\n");
        return 1;
    }
    
    printf("✓ All methods produce identical output (%zu bytes)\n\n", s1);
    
    // Benchmark 1: Field-by-field
    printf("1. Field-by-field (current approach):\n");
    uint64_t start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        size = encode_field_by_field(&data, buf);
    }
    uint64_t end = get_nanos();
    printf("   %.2f ns/op\n\n", (double)(end - start) / ITERATIONS);
    
    // Benchmark 2: Bulk copy
    printf("2. Bulk copy with wire struct:\n");
    start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        size = encode_bulk_copy(&data, buf);
    }
    end = get_nanos();
    printf("   %.2f ns/op\n\n", (double)(end - start) / ITERATIONS);
    
    // Benchmark 3: Direct offsets
    printf("3. Direct write to offsets:\n");
    start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        size = encode_direct_offsets(&data, buf);
    }
    end = get_nanos();
    printf("   %.2f ns/op\n\n", (double)(end - start) / ITERATIONS);
    
    printf("Conclusion:\n");
    printf("-----------\n");
    printf("Best approach shows potential optimization from layout knowledge\n");
    
    return 0;
}
