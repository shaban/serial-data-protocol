/* Test 2: String Copy Methods
 *
 * Compare different approaches to string copying:
 * 1. snprintf (what we avoided in bench)
 * 2. strlen + memcpy (common pattern)
 * 3. Pre-computed length + memcpy (our current approach)
 * 4. Direct pointer copy (if string is const)
 */

#include <stdio.h>
#include <stdint.h>
#include <string.h>
#include <time.h>

#define ITERATIONS 10000000

static uint64_t get_nanos(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
}

/* Test data: NON-CONST to prevent compile-time optimization */
static char test_string_buffer[64] = "Input 1 Gain";  // 12 bytes
static volatile char* TEST_STRING = test_string_buffer;  // Force runtime behavior

/* Approach 1: snprintf (for comparison - what we DON'T want) */
static void copy_snprintf(uint8_t* buf, size_t* offset, volatile char* str) {
    char temp[64];
    int len = snprintf(temp, sizeof(temp), "%s", (char*)str);
    uint32_t str_len = (uint32_t)len;
    memcpy(buf + *offset, &str_len, 4);
    *offset += 4;
    memcpy(buf + *offset, temp, len);
    *offset += len;
}

/* Approach 2: strlen + memcpy (dynamic length) */
static void copy_strlen(uint8_t* buf, size_t* offset, volatile char* str) {
    size_t len = strlen((char*)str);
    uint32_t str_len = (uint32_t)len;
    memcpy(buf + *offset, &str_len, 4);
    *offset += 4;
    memcpy(buf + *offset, (char*)str, len);
    *offset += len;
}

/* Approach 3: Pre-computed length + memcpy (current approach) */
static void copy_precomputed(uint8_t* buf, size_t* offset, volatile char* str, size_t len) {
    uint32_t str_len = (uint32_t)len;
    memcpy(buf + *offset, &str_len, 4);
    *offset += 4;
    memcpy(buf + *offset, (char*)str, len);
    *offset += len;
}

/* Approach 4: Compile-time length (macro) */
#define COPY_LITERAL(buf, offset, literal) do { \
    const size_t len = sizeof(literal) - 1; \
    uint32_t str_len = (uint32_t)len; \
    memcpy((buf) + *(offset), &str_len, 4); \
    *(offset) += 4; \
    memcpy((buf) + *(offset), literal, len); \
    *(offset) += len; \
} while(0)

int main(void) {
    printf("Test 2: String Copy Methods\n");
    printf("============================\n\n");
    
    uint8_t buf[1024];
    volatile size_t offset;  // Prevent optimization
    
    printf("Test string: \"%s\" (%zu bytes)\n\n", (char*)TEST_STRING, strlen((char*)TEST_STRING));
    
    // Benchmark 1: snprintf (baseline - slow)
    printf("1. snprintf (format string):\n");
    uint64_t start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        offset = 0;
        copy_snprintf(buf, (size_t*)&offset, TEST_STRING);
    }
    uint64_t end = get_nanos();
    double snprintf_time = (double)(end - start) / ITERATIONS;
    printf("   %.2f ns/op\n\n", snprintf_time);
    
    // Benchmark 2: strlen + memcpy
    printf("2. strlen + memcpy (runtime length):\n");
    start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        offset = 0;
        copy_strlen(buf, (size_t*)&offset, TEST_STRING);
    }
    end = get_nanos();
    double strlen_time = (double)(end - start) / ITERATIONS;
    printf("   %.2f ns/op\n", strlen_time);
    printf("   %.1fx faster than snprintf\n\n", snprintf_time / strlen_time);
    
    // Benchmark 3: Pre-computed length
    printf("3. Pre-computed length + memcpy:\n");
    const size_t precomp_len = 12;
    start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        offset = 0;
        copy_precomputed(buf, (size_t*)&offset, TEST_STRING, precomp_len);
    }
    end = get_nanos();
    double precomp_time = (double)(end - start) / ITERATIONS;
    printf("   %.2f ns/op\n", precomp_time);
    printf("   %.1fx faster than strlen\n\n", strlen_time / precomp_time);
    
    // Benchmark 4: Compile-time literal
    printf("4. Compile-time literal (macro):\n");
    start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        offset = 0;
        COPY_LITERAL(buf, &offset, "Input 1 Gain");
    }
    end = get_nanos();
    double literal_time = (double)(end - start) / ITERATIONS;
    printf("   %.2f ns/op\n", literal_time);
    printf("   %.1fx faster than strlen\n\n", strlen_time / literal_time);
    
    printf("Summary:\n");
    printf("--------\n");
    printf("snprintf:        %.2f ns (baseline)\n", snprintf_time);
    printf("strlen:          %.2f ns (%.0f%% of snprintf)\n", strlen_time, (strlen_time / snprintf_time) * 100);
    printf("pre-computed:    %.2f ns (%.0f%% of strlen)\n", precomp_time, (precomp_time / strlen_time) * 100);
    printf("literal macro:   %.2f ns (%.0f%% of strlen)\n\n", literal_time, (literal_time / strlen_time) * 100);
    
    printf("Recommendation:\n");
    printf("  - Always require caller to provide string length\n");
    printf("  - Never use strlen() in generated encode functions\n");
    printf("  - For const strings, codegen should emit lengths as constants\n");
    
    return 0;
}
