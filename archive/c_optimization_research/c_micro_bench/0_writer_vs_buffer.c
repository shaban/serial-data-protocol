/* Test 0: Writer (dynamic buffer) vs Fixed Buffer (Go's []byte equivalent)
 *
 * Comparison:
 * - Go uses: buf := make([]byte, size) - pre-allocated, no reallocation
 * - C Writer: Dynamic buffer with capacity checks and realloc
 * - C Fixed: Pre-allocated buffer like Go
 *
 * Question: Is our dynamic writer approach causing overhead?
 */

#include <stdio.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

#define ITERATIONS 1000000
#define EXPECTED_SIZE 1024

/* Dynamic writer (current approach) */
typedef struct {
    uint8_t* data;
    size_t capacity;
    size_t len;
} Writer;

Writer* writer_new(size_t initial_capacity) {
    Writer* w = malloc(sizeof(Writer));
    w->data = malloc(initial_capacity);
    w->capacity = initial_capacity;
    w->len = 0;
    return w;
}

void writer_ensure_capacity(Writer* w, size_t needed) {
    if (w->len + needed > w->capacity) {
        while (w->len + needed > w->capacity) {
            w->capacity *= 2;
        }
        w->data = realloc(w->data, w->capacity);
    }
}

void writer_write_u32(Writer* w, uint32_t value) {
    writer_ensure_capacity(w, 4);
    memcpy(w->data + w->len, &value, 4);
    w->len += 4;
}

void writer_write_string(Writer* w, const char* str, size_t len) {
    writer_ensure_capacity(w, 4 + len);
    uint32_t str_len = (uint32_t)len;
    memcpy(w->data + w->len, &str_len, 4);
    w->len += 4;
    memcpy(w->data + w->len, str, len);
    w->len += len;
}

void writer_reset(Writer* w) {
    w->len = 0;
}

void writer_free(Writer* w) {
    free(w->data);
    free(w);
}

/* Fixed buffer (Go-style []byte) */
typedef struct {
    uint8_t* data;
    size_t capacity;
    size_t offset;
} FixedBuffer;

FixedBuffer* fixed_new(size_t capacity) {
    FixedBuffer* b = malloc(sizeof(FixedBuffer));
    b->data = malloc(capacity);
    b->capacity = capacity;
    b->offset = 0;
    return b;
}

void fixed_write_u32(FixedBuffer* b, uint32_t value) {
    memcpy(b->data + b->offset, &value, 4);
    b->offset += 4;
}

void fixed_write_string(FixedBuffer* b, const char* str, size_t len) {
    uint32_t str_len = (uint32_t)len;
    memcpy(b->data + b->offset, &str_len, 4);
    b->offset += 4;
    memcpy(b->data + b->offset, str, len);
    b->offset += len;
}

void fixed_reset(FixedBuffer* b) {
    b->offset = 0;
}

void fixed_free(FixedBuffer* b) {
    free(b->data);
    free(b);
}

/* Benchmark data */
static void encode_sample_writer(Writer* w) {
    writer_write_u32(w, 12345);
    writer_write_string(w, "Hello", 5);
    writer_write_u32(w, 67890);
    writer_write_string(w, "World", 5);
    writer_write_u32(w, 11111);
}

static void encode_sample_fixed(FixedBuffer* b) {
    fixed_write_u32(b, 12345);
    fixed_write_string(b, "Hello", 5);
    fixed_write_u32(b, 67890);
    fixed_write_string(b, "World", 5);
    fixed_write_u32(b, 11111);
}

static uint64_t get_nanos(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
}

int main(void) {
    printf("Test 0: Dynamic Writer vs Fixed Buffer\n");
    printf("========================================\n\n");
    
    // Test 1: Dynamic writer (with capacity checks)
    printf("1. Dynamic Writer (with realloc safety):\n");
    Writer* w = writer_new(64);  // Small initial size to trigger realloc
    
    uint64_t start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        writer_reset(w);
        encode_sample_writer(w);
    }
    uint64_t end = get_nanos();
    
    double ns_per_op = (double)(end - start) / ITERATIONS;
    printf("   Time: %.2f ns/op\n", ns_per_op);
    printf("   Final capacity: %zu bytes\n\n", w->capacity);
    writer_free(w);
    
    // Test 2: Dynamic writer (pre-sized to avoid realloc)
    printf("2. Dynamic Writer (pre-sized, no realloc):\n");
    w = writer_new(1024);  // Large enough to avoid realloc
    
    start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        writer_reset(w);
        encode_sample_writer(w);
    }
    end = get_nanos();
    
    ns_per_op = (double)(end - start) / ITERATIONS;
    printf("   Time: %.2f ns/op\n", ns_per_op);
    printf("   Final capacity: %zu bytes\n\n", w->capacity);
    writer_free(w);
    
    // Test 3: Fixed buffer (Go-style []byte)
    printf("3. Fixed Buffer (Go-style []byte):\n");
    FixedBuffer* b = fixed_new(1024);
    
    start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        fixed_reset(b);
        encode_sample_fixed(b);
    }
    end = get_nanos();
    
    ns_per_op = (double)(end - start) / ITERATIONS;
    printf("   Time: %.2f ns/op\n\n", ns_per_op);
    fixed_free(b);
    
    printf("Conclusion:\n");
    printf("-----------\n");
    printf("If dynamic writer (pre-sized) ≈ fixed buffer:\n");
    printf("  → Capacity checks are NOT the bottleneck\n");
    printf("If dynamic writer (pre-sized) > fixed buffer:\n");
    printf("  → Function call overhead is the issue\n");
    
    return 0;
}
