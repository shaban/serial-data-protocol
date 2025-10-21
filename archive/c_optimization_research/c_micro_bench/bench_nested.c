/* Handwritten Optimized Encoder: nested.sdp
 *
 * Schema:
 * struct Point {
 *     x: f32,
 *     y: f32
 * }
 *
 * struct Rectangle {
 *     top_left: Point,
 *     bottom_right: Point,
 *     color: u32
 * }
 *
 * struct Scene {
 *     name: str,
 *     main_rect: Rectangle,
 *     count: u32
 * }
 *
 * Optimizations applied:
 * 1. Inline nested struct encoding (no function calls)
 * 2. Wire format struct for Rectangle (bulk copy)
 * 3. Pre-computed string length
 * 4. Direct pointer writes
 */

#include <stdio.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <time.h>

#define ITERATIONS 1000000

/* User-facing structs */
typedef struct {
    float x;
    float y;
} Point;

typedef struct {
    Point top_left;
    Point bottom_right;
    uint32_t color;
} Rectangle;

typedef struct {
    const char* name;
    size_t name_len;
    Rectangle main_rect;
    uint32_t count;
} Scene;

/* Wire format for Rectangle (20 bytes, fixed size) */
typedef struct __attribute__((packed)) {
    uint32_t tl_x_wire;
    uint32_t tl_y_wire;
    uint32_t br_x_wire;
    uint32_t br_y_wire;
    uint32_t color;
} RectangleWire;

static inline uint32_t f32_to_wire(float f) {
    uint32_t u;
    memcpy(&u, &f, 4);
    return u;
}

/* BASELINE: Recursive with function calls */
static void encode_point_baseline(uint8_t* buf, size_t* offset, const Point* p) {
    uint32_t x = f32_to_wire(p->x);
    uint32_t y = f32_to_wire(p->y);
    memcpy(buf + *offset, &x, 4); *offset += 4;
    memcpy(buf + *offset, &y, 4); *offset += 4;
}

static void encode_rectangle_baseline(uint8_t* buf, size_t* offset, const Rectangle* r) {
    encode_point_baseline(buf, offset, &r->top_left);
    encode_point_baseline(buf, offset, &r->bottom_right);
    memcpy(buf + *offset, &r->color, 4); *offset += 4;
}

static size_t encode_scene_baseline(const Scene* src, uint8_t* buf) {
    size_t offset = 0;
    
    // String
    uint32_t name_len = (uint32_t)src->name_len;
    memcpy(buf + offset, &name_len, 4); offset += 4;
    memcpy(buf + offset, src->name, src->name_len); offset += src->name_len;
    
    // Rectangle (nested, via function calls)
    encode_rectangle_baseline(buf, &offset, &src->main_rect);
    
    // Count
    memcpy(buf + offset, &src->count, 4); offset += 4;
    
    return offset;
}

/* OPTIMIZED: Inline + bulk copy with wire struct */
static size_t encode_scene_optimized(const Scene* src, uint8_t* buf) {
    size_t offset = 0;
    
    // String
    *(uint32_t*)(buf + offset) = (uint32_t)src->name_len;
    offset += 4;
    memcpy(buf + offset, src->name, src->name_len);
    offset += src->name_len;
    
    // Rectangle - prepare wire struct and bulk copy (20 bytes)
    RectangleWire rect_wire = {
        .tl_x_wire = f32_to_wire(src->main_rect.top_left.x),
        .tl_y_wire = f32_to_wire(src->main_rect.top_left.y),
        .br_x_wire = f32_to_wire(src->main_rect.bottom_right.x),
        .br_y_wire = f32_to_wire(src->main_rect.bottom_right.y),
        .color = src->main_rect.color,
    };
    memcpy(buf + offset, &rect_wire, sizeof(RectangleWire));
    offset += sizeof(RectangleWire);
    
    // Count
    *(uint32_t*)(buf + offset) = src->count;
    offset += 4;
    
    return offset;
}

static uint64_t get_nanos(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
}

int main(void) {
    printf("Nested Schema Benchmark\n");
    printf("========================\n\n");
    
    Scene scene = {
        .name = "MainScene",
        .name_len = 9,
        .main_rect = {
            .top_left = {.x = 0.0f, .y = 0.0f},
            .bottom_right = {.x = 1920.0f, .y = 1080.0f},
            .color = 0xFF0000FF
        },
        .count = 42
    };
    
    uint8_t buf[256];
    volatile size_t size;
    
    // Verify both produce same output
    uint8_t buf1[256], buf2[256];
    size_t s1 = encode_scene_baseline(&scene, buf1);
    size_t s2 = encode_scene_optimized(&scene, buf2);
    
    if (s1 != s2 || memcmp(buf1, buf2, s1) != 0) {
        printf("ERROR: Output mismatch!\n");
        return 1;
    }
    
    printf("✓ Both encoders produce identical output (%zu bytes)\n", s1);
    printf("  String: %zu bytes\n", 4 + scene.name_len);
    printf("  Rectangle: %zu bytes (nested Point + Point + u32)\n", sizeof(RectangleWire));
    printf("  Count: 4 bytes\n\n");
    
    // Benchmark baseline
    printf("Baseline (recursive function calls):\n");
    uint64_t start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        size = encode_scene_baseline(&scene, buf);
    }
    uint64_t end = get_nanos();
    double baseline_time = (double)(end - start) / ITERATIONS;
    printf("  %.2f ns/op\n\n", baseline_time);
    
    // Benchmark optimized
    printf("Optimized (inline + wire struct):\n");
    start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        size = encode_scene_optimized(&scene, buf);
    }
    end = get_nanos();
    double optimized_time = (double)(end - start) / ITERATIONS;
    printf("  %.2f ns/op\n", optimized_time);
    printf("  %.1fx faster\n\n", baseline_time / optimized_time);
    
    printf("Speedup: %.0f%% improvement\n", ((baseline_time - optimized_time) / baseline_time) * 100);
    printf("\nKey optimization: Avoided %d function calls per encode\n", 3);  // encode_rectangle + 2x encode_point
    
    return 0;
}
