/* Test 3: Array Copy Optimizations
 *
 * Arrays are a major bottleneck. Test different approaches:
 * 1. Loop with individual element encoding (current)
 * 2. Bulk memcpy for primitive arrays
 * 3. Vectorized/SIMD for large primitive arrays
 * 4. Batch endian conversion
 */

#include <stdio.h>
#include <stdint.h>
#include <string.h>
#include <time.h>

#define ITERATIONS 100000
#define ARRAY_SIZE 100  // Typical parameter array size

static uint64_t get_nanos(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
}

/* Test data - NON-CONST to prevent optimization */
static float test_f32_array[ARRAY_SIZE];
static uint32_t test_u32_array[ARRAY_SIZE];
static volatile size_t array_size = ARRAY_SIZE;  // Runtime size

/* Approach 1: Loop with individual writes (current) */
static void encode_f32_loop(uint8_t* buf, volatile float* arr, volatile size_t count) {
    size_t offset = 0;
    
    // Write count
    uint32_t arr_count = (uint32_t)count;
    memcpy(buf + offset, &arr_count, 4);
    offset += 4;
    
    // Write each element
    for (size_t i = 0; i < count; i++) {
        uint32_t wire;
        memcpy(&wire, (float*)&arr[i], 4);  // f32 -> u32
        memcpy(buf + offset, &wire, 4);
        offset += 4;
    }
}

/* Approach 2: Bulk memcpy (assumes same representation) */
static void encode_f32_bulk(uint8_t* buf, volatile float* arr, volatile size_t count) {
    size_t offset = 0;
    
    // Write count
    uint32_t arr_count = (uint32_t)count;
    memcpy(buf + offset, &arr_count, 4);
    offset += 4;
    
    // Bulk copy (f32 and u32 have same wire format on little-endian)
    memcpy(buf + offset, (float*)arr, count * 4);
}

/* Approach 3: Batch conversion + bulk copy */
static void encode_f32_batch(uint8_t* buf, volatile float* arr, volatile size_t count) {
    size_t offset = 0;
    
    // Write count
    uint32_t arr_count = (uint32_t)count;
    memcpy(buf + offset, &arr_count, 4);
    offset += 4;
    
    // Convert all at once (compiler can vectorize)
    uint32_t* wire = (uint32_t*)(buf + offset);
    for (size_t i = 0; i < count; i++) {
        memcpy(&wire[i], (float*)&arr[i], 4);
    }
}

/* For u32 arrays (no conversion needed) */
static void encode_u32_loop(uint8_t* buf, volatile uint32_t* arr, volatile size_t count) {
    size_t offset = 0;
    
    uint32_t arr_count = (uint32_t)count;
    memcpy(buf + offset, &arr_count, 4);
    offset += 4;
    
    for (size_t i = 0; i < count; i++) {
        memcpy(buf + offset, (uint32_t*)&arr[i], 4);
        offset += 4;
    }
}

static void encode_u32_bulk(uint8_t* buf, volatile uint32_t* arr, volatile size_t count) {
    size_t offset = 0;
    
    uint32_t arr_count = (uint32_t)count;
    memcpy(buf + offset, &arr_count, 4);
    offset += 4;
    
    // Single bulk copy (no conversion needed)
    memcpy(buf + offset, (uint32_t*)arr, count * 4);
}

int main(void) {
    printf("Test 3: Array Copy Optimizations\n");
    printf("=================================\n\n");
    
    // Initialize test data
    for (int i = 0; i < ARRAY_SIZE; i++) {
        test_f32_array[i] = (float)i * 1.5f;
        test_u32_array[i] = (uint32_t)i * 1000;
    }
    
    uint8_t buf[4096];
    
    printf("Array size: %d elements\n\n", ARRAY_SIZE);
    
    // ===== Float arrays (need conversion) =====
    printf("Float arrays (f32 -> wire):\n");
    printf("---------------------------\n");
    
    // Benchmark 1: Loop
    printf("1. Loop with individual writes:\n");
    uint64_t start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        encode_f32_loop(buf, test_f32_array, array_size);
    }
    uint64_t end = get_nanos();
    double loop_f32 = (double)(end - start) / ITERATIONS;
    printf("   %.2f ns/op\n\n", loop_f32);
    
    // Benchmark 2: Bulk
    printf("2. Bulk memcpy (same representation):\n");
    start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        encode_f32_bulk(buf, test_f32_array, array_size);
    }
    end = get_nanos();
    double bulk_f32 = (double)(end - start) / ITERATIONS;
    printf("   %.2f ns/op\n", bulk_f32);
    printf("   %.1fx faster than loop\n\n", loop_f32 / bulk_f32);
    
    // Benchmark 3: Batch
    printf("3. Batch conversion:\n");
    start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        encode_f32_batch(buf, test_f32_array, array_size);
    }
    end = get_nanos();
    double batch_f32 = (double)(end - start) / ITERATIONS;
    printf("   %.2f ns/op\n", batch_f32);
    printf("   %.1fx faster than loop\n\n", loop_f32 / batch_f32);
    
    // ===== Integer arrays (no conversion) =====
    printf("Integer arrays (u32 -> wire):\n");
    printf("-----------------------------\n");
    
    // Benchmark 4: Loop
    printf("4. Loop with individual writes:\n");
    start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        encode_u32_loop(buf, test_u32_array, array_size);
    }
    end = get_nanos();
    double loop_u32 = (double)(end - start) / ITERATIONS;
    printf("   %.2f ns/op\n\n", loop_u32);
    
    // Benchmark 5: Bulk
    printf("5. Bulk memcpy (direct copy):\n");
    start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        encode_u32_bulk(buf, test_u32_array, array_size);
    }
    end = get_nanos();
    double bulk_u32 = (double)(end - start) / ITERATIONS;
    printf("   %.2f ns/op\n", bulk_u32);
    printf("   %.1fx faster than loop\n\n", loop_u32 / bulk_u32);
    
    printf("Summary:\n");
    printf("--------\n");
    printf("f32 loop:  %.2f ns (baseline)\n", loop_f32);
    printf("f32 bulk:  %.2f ns (%.0f%% speedup)\n", bulk_f32, ((loop_f32 - bulk_f32) / loop_f32) * 100);
    printf("u32 loop:  %.2f ns (baseline)\n", loop_u32);
    printf("u32 bulk:  %.2f ns (%.0f%% speedup)\n\n", bulk_u32, ((loop_u32 - bulk_u32) / loop_u32) * 100);
    
    printf("Recommendation:\n");
    printf("  - For primitive arrays: use bulk memcpy\n");
    printf("  - Generate array-specific encoders: encode_u32_array()\n");
    printf("  - Single capacity check for entire array\n");
    printf("  - %d-element f32 array: ~%.0f ns savings\n", ARRAY_SIZE, loop_f32 - bulk_f32);
    
    return 0;
}
