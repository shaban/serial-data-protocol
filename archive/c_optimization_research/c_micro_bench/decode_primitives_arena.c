/* decode_primitives_arena.c - Arena-based decoder (strings are copied)
 * 
 * Approach: Arena allocator owns all memory, single free at end.
 * Pros: Simple API, strings are mutable, no lifetime issues
 * Cons: Requires allocation, slightly slower than zero-copy
 * 
 * Compile: gcc -std=c99 -O3 -march=native decode_primitives_arena.c -o decode_primitives_arena
 */

#include <stdio.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

/* Endianness helpers */
#if defined(__APPLE__)
    #include <libkern/OSByteOrder.h>
    #define LE16TOH(x) OSSwapLittleToHostInt16(x)
    #define LE32TOH(x) OSSwapLittleToHostInt32(x)
    #define LE64TOH(x) OSSwapLittleToHostInt64(x)
#elif defined(__linux__)
    #include <endian.h>
    #define LE16TOH(x) le16toh(x)
    #define LE32TOH(x) le32toh(x)
    #define LE64TOH(x) le64toh(x)
#endif

/* Float conversion */
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

/* Simple bump allocator arena */
typedef struct {
    uint8_t* memory;
    size_t capacity;
    size_t offset;
} Arena;

static Arena* arena_create(size_t capacity) {
    Arena* arena = malloc(sizeof(Arena));
    arena->memory = malloc(capacity);
    arena->capacity = capacity;
    arena->offset = 0;
    return arena;
}

static void arena_destroy(Arena* arena) {
    free(arena->memory);
    free(arena);
}

static void arena_reset(Arena* arena) {
    arena->offset = 0;
}

static void* arena_alloc(Arena* arena, size_t size) {
    if (arena->offset + size > arena->capacity) {
        return NULL;  /* Out of memory */
    }
    void* ptr = arena->memory + arena->offset;
    arena->offset += size;
    return ptr;
}

/* Arena-allocated decoded struct */
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
    char* str_field;  /* Owned by arena */
    size_t str_field_len;
} AllPrimitives;

/* Decode with arena allocation */
AllPrimitives* decode_all_primitives(const uint8_t* buf, size_t buf_len, Arena* arena) {
    size_t offset = 0;
    
    #define CHECK_BOUNDS(n) if (offset + (n) > buf_len) return NULL
    
    /* Allocate struct from arena */
    AllPrimitives* dest = arena_alloc(arena, sizeof(AllPrimitives));
    if (!dest) return NULL;
    
    /* Decode primitives (same as zero-copy) */
    CHECK_BOUNDS(1);
    dest->u8_field = buf[offset];
    offset += 1;
    
    CHECK_BOUNDS(2);
    dest->u16_field = LE16TOH(*(const uint16_t*)(buf + offset));
    offset += 2;
    
    CHECK_BOUNDS(4);
    dest->u32_field = LE32TOH(*(const uint32_t*)(buf + offset));
    offset += 4;
    
    CHECK_BOUNDS(8);
    dest->u64_field = LE64TOH(*(const uint64_t*)(buf + offset));
    offset += 8;
    
    CHECK_BOUNDS(1);
    dest->i8_field = (int8_t)buf[offset];
    offset += 1;
    
    CHECK_BOUNDS(2);
    dest->i16_field = (int16_t)LE16TOH(*(const uint16_t*)(buf + offset));
    offset += 2;
    
    CHECK_BOUNDS(4);
    dest->i32_field = (int32_t)LE32TOH(*(const uint32_t*)(buf + offset));
    offset += 4;
    
    CHECK_BOUNDS(8);
    dest->i64_field = (int64_t)LE64TOH(*(const uint64_t*)(buf + offset));
    offset += 8;
    
    CHECK_BOUNDS(4);
    dest->f32_field = le_to_f32(buf + offset);
    offset += 4;
    
    CHECK_BOUNDS(8);
    dest->f64_field = le_to_f64(buf + offset);
    offset += 8;
    
    CHECK_BOUNDS(1);
    dest->bool_field = buf[offset];
    offset += 1;
    
    /* String (allocate and copy from arena) */
    CHECK_BOUNDS(4);
    uint32_t str_len = LE32TOH(*(const uint32_t*)(buf + offset));
    offset += 4;
    
    CHECK_BOUNDS(str_len);
    dest->str_field = arena_alloc(arena, str_len + 1);  /* +1 for null terminator */
    if (!dest->str_field) return NULL;
    
    memcpy(dest->str_field, buf + offset, str_len);
    dest->str_field[str_len] = '\0';  /* Null-terminate */
    dest->str_field_len = str_len;
    offset += str_len;
    
    #undef CHECK_BOUNDS
    return dest;
}

/* Test data */
static const uint8_t test_data[] = {
    42, 0xe8, 0x03, 0xa0, 0x86, 0x01, 0x00,
    0xcb, 0x04, 0xfb, 0x71, 0x1f, 0x01, 0x00, 0x00,
    0xf6, 0x18, 0xfc, 0x60, 0x79, 0xfe, 0xff,
    0x16, 0xe9, 0x4f, 0xb3, 0xfd, 0xff, 0xff, 0xff,
    0xd0, 0x0f, 0x49, 0x40,
    0x90, 0xf7, 0xaa, 0x95, 0x09, 0xbf, 0x05, 0x40,
    0x01, 0x05, 0x00, 0x00, 0x00, 'h', 'e', 'l', 'l', 'o'
};

int main(void) {
    Arena* arena = arena_create(1024 * 1024);  /* 1MB arena */
    
    /* Warmup */
    for (int i = 0; i < 1000; i++) {
        arena_reset(arena);
        AllPrimitives* decoded = decode_all_primitives(test_data, sizeof(test_data), arena);
        if (!decoded) {
            fprintf(stderr, "Decode failed\n");
            return 1;
        }
    }
    
    /* Benchmark */
    const int iterations = 10000000;
    volatile uint32_t sink = 0;
    
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    
    for (int i = 0; i < iterations; i++) {
        arena_reset(arena);  /* Reset arena for each iteration */
        AllPrimitives* decoded = decode_all_primitives(test_data, sizeof(test_data), arena);
        if (!decoded) {
            fprintf(stderr, "Decode failed\n");
            return 1;
        }
        sink += decoded->u32_field;
    }
    
    clock_gettime(CLOCK_MONOTONIC, &end);
    
    uint64_t start_ns = start.tv_sec * 1000000000ULL + start.tv_nsec;
    uint64_t end_ns = end.tv_sec * 1000000000ULL + end.tv_nsec;
    uint64_t total_ns = end_ns - start_ns;
    double ns_per_op = (double)total_ns / iterations;
    
    printf("=== Arena-Based Decode (Primitives) ===\n");
    printf("Iterations: %d\n", iterations);
    printf("Total time: %.2f ms\n", total_ns / 1e6);
    printf("Time per op: %.2f ns\n", ns_per_op);
    printf("Throughput: %.2f million ops/sec\n", 1000.0 / ns_per_op);
    
    /* Verify last decode */
    arena_reset(arena);
    AllPrimitives* final = decode_all_primitives(test_data, sizeof(test_data), arena);
    printf("\nVerification:\n");
    printf("  u8_field: %u (expected 42)\n", final->u8_field);
    printf("  u16_field: %u (expected 1000)\n", final->u16_field);
    printf("  u32_field: %u (expected 100000)\n", final->u32_field);
    printf("  f32_field: %.5f (expected 3.14159)\n", final->f32_field);
    printf("  str_field: '%s' (expected 'hello')\n", final->str_field);
    
    arena_destroy(arena);
    return 0;
}
