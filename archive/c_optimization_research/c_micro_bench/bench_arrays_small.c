// bench_arrays_small.c - Fair comparison with Go's small dataset
// Go uses 4-5 elements, not 50!

#include <stdio.h>
#include <stdint.h>
#include <string.h>
#include <time.h>

typedef struct {
    uint8_t* u8_array;
    size_t u8_array_len;
    uint32_t* u32_array;
    size_t u32_array_len;
    double* f64_array;
    size_t f64_array_len;
    char** str_array;
    size_t* str_array_lens;
    size_t str_array_len;
    uint8_t* bool_array;
    size_t bool_array_len;
} ArraysOfPrimitives;

/* BASELINE: Loop per element */
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
        *(uint32_t*)(buf + offset) = src->u32_array[i];
        offset += 4;
    }
    
    // f64 array
    *(uint32_t*)(buf + offset) = (uint32_t)src->f64_array_len;
    offset += 4;
    for (size_t i = 0; i < src->f64_array_len; i++) {
        *(double*)(buf + offset) = src->f64_array[i];
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

/* OPTIMIZED: Bulk memcpy */
static size_t encode_optimized(const ArraysOfPrimitives* src, uint8_t* buf) {
    size_t offset = 0;
    
    // u8 array
    *(uint32_t*)(buf + offset) = (uint32_t)src->u8_array_len;
    offset += 4;
    memcpy(buf + offset, src->u8_array, src->u8_array_len);
    offset += src->u8_array_len;
    
    // u32 array
    *(uint32_t*)(buf + offset) = (uint32_t)src->u32_array_len;
    offset += 4;
    memcpy(buf + offset, src->u32_array, src->u32_array_len * 4);
    offset += src->u32_array_len * 4;
    
    // f64 array
    *(uint32_t*)(buf + offset) = (uint32_t)src->f64_array_len;
    offset += 4;
    memcpy(buf + offset, src->f64_array, src->f64_array_len * 8);
    offset += src->f64_array_len * 8;
    
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
    memcpy(buf + offset, src->bool_array, src->bool_array_len);
    offset += src->bool_array_len;
    
    return offset;
}

static uint64_t get_nanos(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000000000 + ts.tv_nsec;
}

#define ITERATIONS 10000000

int main(void) {
    // Match Go benchmark data EXACTLY:
    // U8Array:   []uint8{1, 2, 3, 255},
    // U32Array:  []uint32{100, 200, 300, 4294967295},
    // F64Array:  []float64{1.1, 2.2, 3.3, math.Pi, math.E},
    // StrArray:  []string{"hello", "world", "", "test 🚀"},
    // BoolArray: []bool{true, false, true, false, true},
    
    static uint8_t u8_data[] = {1, 2, 3, 255};
    static uint32_t u32_data[] = {100, 200, 300, 4294967295};
    static double f64_data[] = {1.1, 2.2, 3.3, 3.141592653589793, 2.718281828459045};
    static char* str_data[] = {"hello", "world", "", "test 🚀"};
    static size_t str_lens[] = {5, 5, 0, 11};  // "test 🚀" is 11 bytes UTF-8
    static uint8_t bool_data[] = {1, 0, 1, 0, 1};
    
    ArraysOfPrimitives test_data = {
        u8_data, 4,
        u32_data, 4,
        f64_data, 5,
        str_data, str_lens, 4,
        bool_data, 5
    };
    
    uint8_t buf_baseline[1024];
    uint8_t buf_optimized[1024];
    
    // Verify both produce same output
    size_t size_baseline = encode_baseline(&test_data, buf_baseline);
    size_t size_optimized = encode_optimized(&test_data, buf_optimized);
    
    if (size_baseline != size_optimized || memcmp(buf_baseline, buf_optimized, size_baseline) != 0) {
        printf("ERROR: Encoders produce different output!\n");
        return 1;
    }
    
    printf("\n✓ Both encoders produce identical output (%zu bytes)\n", size_baseline);
    printf("  Arrays: u8[4], u32[4], f64[5], str[4], bool[5]\n");
    printf("  (Matching Go benchmark dataset)\n\n");
    
    // Warmup
    volatile size_t warmup = encode_baseline(&test_data, buf_baseline);
    warmup += encode_optimized(&test_data, buf_optimized);
    (void)warmup;
    
    // Benchmark baseline
    uint64_t start = get_nanos();
    volatile size_t total_baseline = 0;
    for (int i = 0; i < ITERATIONS; i++) {
        total_baseline += encode_baseline(&test_data, buf_baseline);
    }
    uint64_t end = get_nanos();
    double baseline_ns = (double)(end - start) / ITERATIONS;
    
    // Benchmark optimized
    start = get_nanos();
    volatile size_t total_optimized = 0;
    for (int i = 0; i < ITERATIONS; i++) {
        total_optimized += encode_optimized(&test_data, buf_optimized);
    }
    end = get_nanos();
    double optimized_ns = (double)(end - start) / ITERATIONS;
    
    // Results
    printf("Baseline (loop per element):\n");
    printf("  %.2f ns/op\n\n", baseline_ns);
    
    printf("Optimized (bulk memcpy):\n");
    printf("  %.2f ns/op\n", optimized_ns);
    printf("  %.1fx faster\n\n", baseline_ns / optimized_ns);
    
    printf("Speedup: %.0f%% improvement\n\n", 
           (1.0 - optimized_ns / baseline_ns) * 100);
    
    printf("====================================\n");
    printf("Comparison with Go:\n");
    printf("  Go:             56.02 ns/op\n");
    printf("  C (baseline):   %.2f ns/op (%.1fx vs Go)\n", baseline_ns, 56.02 / baseline_ns);
    printf("  C (optimized):  %.2f ns/op (%.1fx vs Go)\n", optimized_ns, 56.02 / optimized_ns);
    printf("====================================\n\n");
    
    return 0;
}
