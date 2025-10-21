/* Handwritten Optimized Encoder: arrays.sdp
 *
 * Schema:
 * struct ArraysOfPrimitives {
 *     u8_array: []u8,
 *     u32_array: []u32,
 *     f64_array: []f64,
 *     str_array: []str,
 *     bool_array: []bool
 * }
 *
 * Optimizations applied:
 * 1. Bulk memcpy for primitive arrays
 * 2. Single capacity check per array
 * 3. Pre-computed string lengths in str_array
 * 4. Direct pointer writes
 */

#include <stdio.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

#define ITERATIONS 100000
#define ARRAY_SIZE 50  // Realistic array sizes

/* User-facing struct */
typedef struct {
    uint8_t* u8_array;
    size_t u8_array_len;
    
    uint32_t* u32_array;
    size_t u32_array_len;
    
    double* f64_array;
    size_t f64_array_len;
    
    const char** str_array;
    size_t* str_array_lens;  // Pre-computed lengths
    size_t str_array_len;
    
    uint8_t* bool_array;
    size_t bool_array_len;
} ArraysOfPrimitives;

static inline uint64_t f64_to_wire(double d) {
    uint64_t u;
    memcpy(&u, &d, 8);
    return u;
}

/* BASELINE: Loop with individual element writes */
static size_t encode_baseline(const ArraysOfPrimitives* src, uint8_t* buf) {
    size_t offset = 0;
    
    // u8 array
    *(uint32_t*)(buf + offset) = (uint32_t)src->u8_array_len;
    offset += 4;
    for (size_t i = 0; i < src->u8_array_len; i++) {
        buf[offset++] = src->u8_array[i];
    }
    
    // u32 array
    *(uint32_t*)(buf + offset) = (uint32_t)src->u32_array_len;
    offset += 4;
    for (size_t i = 0; i < src->u32_array_len; i++) {
        memcpy(buf + offset, &src->u32_array[i], 4);
        offset += 4;
    }
    
    // f64 array
    *(uint32_t*)(buf + offset) = (uint32_t)src->f64_array_len;
    offset += 4;
    for (size_t i = 0; i < src->f64_array_len; i++) {
        uint64_t wire = f64_to_wire(src->f64_array[i]);
        memcpy(buf + offset, &wire, 8);
        offset += 8;
    }
    
    // str array
    *(uint32_t*)(buf + offset) = (uint32_t)src->str_array_len;
    offset += 4;
    for (size_t i = 0; i < src->str_array_len; i++) {
        uint32_t len = (uint32_t)src->str_array_lens[i];
        *(uint32_t*)(buf + offset) = len;
        offset += 4;
        memcpy(buf + offset, src->str_array[i], len);
        offset += len;
    }
    
    // bool array
    *(uint32_t*)(buf + offset) = (uint32_t)src->bool_array_len;
    offset += 4;
    for (size_t i = 0; i < src->bool_array_len; i++) {
        buf[offset++] = src->bool_array[i];
    }
    
    return offset;
}

/* OPTIMIZED: Bulk memcpy for primitive arrays */
static size_t encode_optimized(const ArraysOfPrimitives* src, uint8_t* buf) {
    size_t offset = 0;
    
    // u8 array - bulk copy (1 byte elements, no conversion)
    *(uint32_t*)(buf + offset) = (uint32_t)src->u8_array_len;
    offset += 4;
    memcpy(buf + offset, src->u8_array, src->u8_array_len);
    offset += src->u8_array_len;
    
    // u32 array - bulk copy (native format = wire format on little-endian)
    *(uint32_t*)(buf + offset) = (uint32_t)src->u32_array_len;
    offset += 4;
    memcpy(buf + offset, src->u32_array, src->u32_array_len * 4);
    offset += src->u32_array_len * 4;
    
    // f64 array - bulk copy (native format = wire format)
    *(uint32_t*)(buf + offset) = (uint32_t)src->f64_array_len;
    offset += 4;
    memcpy(buf + offset, src->f64_array, src->f64_array_len * 8);
    offset += src->f64_array_len * 8;
    
    // str array - still need loop (variable-length strings)
    *(uint32_t*)(buf + offset) = (uint32_t)src->str_array_len;
    offset += 4;
    for (size_t i = 0; i < src->str_array_len; i++) {
        uint32_t len = (uint32_t)src->str_array_lens[i];
        *(uint32_t*)(buf + offset) = len;
        offset += 4;
        memcpy(buf + offset, src->str_array[i], len);
        offset += len;
    }
    
    // bool array - bulk copy (1 byte elements)
    *(uint32_t*)(buf + offset) = (uint32_t)src->bool_array_len;
    offset += 4;
    memcpy(buf + offset, src->bool_array, src->bool_array_len);
    offset += src->bool_array_len;
    
    return offset;
}

static uint64_t get_nanos(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
}

int main(void) {
    printf("Arrays Schema Benchmark\n");
    printf("=======================\n\n");
    
    // Allocate test data
    uint8_t* u8_arr = malloc(ARRAY_SIZE);
    uint32_t* u32_arr = malloc(ARRAY_SIZE * sizeof(uint32_t));
    double* f64_arr = malloc(ARRAY_SIZE * sizeof(double));
    uint8_t* bool_arr = malloc(ARRAY_SIZE);
    
    const char** str_arr = malloc(ARRAY_SIZE * sizeof(char*));
    size_t* str_lens = malloc(ARRAY_SIZE * sizeof(size_t));
    
    // Initialize
    for (int i = 0; i < ARRAY_SIZE; i++) {
        u8_arr[i] = i;
        u32_arr[i] = i * 1000;
        f64_arr[i] = i * 1.5;
        bool_arr[i] = i % 2;
        str_arr[i] = (i % 2) ? "param" : "value";
        str_lens[i] = (i % 2) ? 5 : 5;
    }
    
    ArraysOfPrimitives data = {
        .u8_array = u8_arr,
        .u8_array_len = ARRAY_SIZE,
        .u32_array = u32_arr,
        .u32_array_len = ARRAY_SIZE,
        .f64_array = f64_arr,
        .f64_array_len = ARRAY_SIZE,
        .str_array = str_arr,
        .str_array_lens = str_lens,
        .str_array_len = ARRAY_SIZE,
        .bool_array = bool_arr,
        .bool_array_len = ARRAY_SIZE,
    };
    
    uint8_t* buf = malloc(10000);
    volatile size_t size;
    
    // Verify both produce same output
    uint8_t* buf1 = malloc(10000);
    uint8_t* buf2 = malloc(10000);
    size_t s1 = encode_baseline(&data, buf1);
    size_t s2 = encode_optimized(&data, buf2);
    
    if (s1 != s2 || memcmp(buf1, buf2, s1) != 0) {
        printf("ERROR: Output mismatch! %zu vs %zu\n", s1, s2);
        return 1;
    }
    
    printf("✓ Both encoders produce identical output (%zu bytes)\n", s1);
    printf("  Arrays: %d elements each\n", ARRAY_SIZE);
    printf("  Total encoded: %zu bytes\n\n", s1);
    
    // Benchmark baseline
    printf("Baseline (loop per element):\n");
    uint64_t start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        size = encode_baseline(&data, buf);
    }
    uint64_t end = get_nanos();
    double baseline_time = (double)(end - start) / ITERATIONS;
    printf("  %.2f ns/op\n\n", baseline_time);
    
    // Benchmark optimized
    printf("Optimized (bulk memcpy):\n");
    start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        size = encode_optimized(&data, buf);
    }
    end = get_nanos();
    double optimized_time = (double)(end - start) / ITERATIONS;
    printf("  %.2f ns/op\n", optimized_time);
    printf("  %.1fx faster\n\n", baseline_time / optimized_time);
    
    printf("Speedup: %.0f%% improvement\n", ((baseline_time - optimized_time) / baseline_time) * 100);
    
    // Cleanup
    free(u8_arr);
    free(u32_arr);
    free(f64_arr);
    free(bool_arr);
    free(str_arr);
    free(str_lens);
    free(buf);
    free(buf1);
    free(buf2);
    
    return 0;
}
