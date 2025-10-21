/* Test 4: Struct Copy Optimizations
 *
 * Nested structs are common. Test approaches:
 * 1. Field-by-field recursive encoding (current)
 * 2. Flattened encoding (inline struct fields)
 * 3. Bulk copy for fixed-size structs
 * 4. Layout-aware encoding
 */

#include <stdio.h>
#include <stdint.h>
#include <string.h>
#include <time.h>

#define ITERATIONS 1000000

static uint64_t get_nanos(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return (uint64_t)ts.tv_sec * 1000000000ULL + (uint64_t)ts.tv_nsec;
}

/* Example: nested.sdp - Point, Rectangle, Scene */
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

/* Wire format for Point (8 bytes) */
typedef struct __attribute__((packed)) {
    uint32_t x_wire;  // f32 as u32
    uint32_t y_wire;
} PointWire;

/* Wire format for Rectangle (20 bytes) */
typedef struct __attribute__((packed)) {
    PointWire top_left;
    PointWire bottom_right;
    uint32_t color;
} RectangleWire;

static inline uint32_t f32_to_wire(float f) {
    uint32_t u;
    memcpy(&u, &f, 4);
    return u;
}

/* Approach 1: Recursive encoding (current approach) */
static void encode_point_recursive(uint8_t* buf, size_t* offset, const Point* p) {
    uint32_t x = f32_to_wire(p->x);
    uint32_t y = f32_to_wire(p->y);
    memcpy(buf + *offset, &x, 4); *offset += 4;
    memcpy(buf + *offset, &y, 4); *offset += 4;
}

static void encode_rectangle_recursive(uint8_t* buf, size_t* offset, const Rectangle* r) {
    encode_point_recursive(buf, offset, &r->top_left);
    encode_point_recursive(buf, offset, &r->bottom_right);
    memcpy(buf + *offset, &r->color, 4); *offset += 4;
}

static void encode_scene_recursive(uint8_t* buf, size_t* offset, const Scene* s) {
    // String
    uint32_t name_len = (uint32_t)s->name_len;
    memcpy(buf + *offset, &name_len, 4); *offset += 4;
    memcpy(buf + *offset, s->name, s->name_len); *offset += s->name_len;
    
    // Rectangle (nested)
    encode_rectangle_recursive(buf, offset, &s->main_rect);
    
    // Count
    memcpy(buf + *offset, &s->count, 4); *offset += 4;
}

/* Approach 2: Flattened (inline all fields) */
static void encode_scene_flattened(uint8_t* buf, size_t* offset, const Scene* s) {
    // String
    uint32_t name_len = (uint32_t)s->name_len;
    memcpy(buf + *offset, &name_len, 4); *offset += 4;
    memcpy(buf + *offset, s->name, s->name_len); *offset += s->name_len;
    
    // Rectangle fields (flattened)
    uint32_t x1 = f32_to_wire(s->main_rect.top_left.x);
    uint32_t y1 = f32_to_wire(s->main_rect.top_left.y);
    uint32_t x2 = f32_to_wire(s->main_rect.bottom_right.x);
    uint32_t y2 = f32_to_wire(s->main_rect.bottom_right.y);
    
    memcpy(buf + *offset, &x1, 4); *offset += 4;
    memcpy(buf + *offset, &y1, 4); *offset += 4;
    memcpy(buf + *offset, &x2, 4); *offset += 4;
    memcpy(buf + *offset, &y2, 4); *offset += 4;
    memcpy(buf + *offset, &s->main_rect.color, 4); *offset += 4;
    
    // Count
    memcpy(buf + *offset, &s->count, 4); *offset += 4;
}

/* Approach 3: Bulk copy for fixed struct */
static void encode_scene_bulk(uint8_t* buf, size_t* offset, const Scene* s) {
    // String (variable)
    uint32_t name_len = (uint32_t)s->name_len;
    memcpy(buf + *offset, &name_len, 4); *offset += 4;
    memcpy(buf + *offset, s->name, s->name_len); *offset += s->name_len;
    
    // Rectangle (bulk copy as wire struct)
    RectangleWire rect_wire;
    rect_wire.top_left.x_wire = f32_to_wire(s->main_rect.top_left.x);
    rect_wire.top_left.y_wire = f32_to_wire(s->main_rect.top_left.y);
    rect_wire.bottom_right.x_wire = f32_to_wire(s->main_rect.bottom_right.x);
    rect_wire.bottom_right.y_wire = f32_to_wire(s->main_rect.bottom_right.y);
    rect_wire.color = s->main_rect.color;
    
    memcpy(buf + *offset, &rect_wire, sizeof(RectangleWire));
    *offset += sizeof(RectangleWire);
    
    // Count
    memcpy(buf + *offset, &s->count, 4); *offset += 4;
}

/* Approach 4: Direct pointer writes (fastest) */
static void encode_scene_direct(uint8_t* buf, size_t* offset, const Scene* s) {
    // String
    *(uint32_t*)(buf + *offset) = (uint32_t)s->name_len;
    *offset += 4;
    memcpy(buf + *offset, s->name, s->name_len);
    *offset += s->name_len;
    
    // Rectangle (direct pointer writes)
    RectangleWire* rect = (RectangleWire*)(buf + *offset);
    rect->top_left.x_wire = f32_to_wire(s->main_rect.top_left.x);
    rect->top_left.y_wire = f32_to_wire(s->main_rect.top_left.y);
    rect->bottom_right.x_wire = f32_to_wire(s->main_rect.bottom_right.x);
    rect->bottom_right.y_wire = f32_to_wire(s->main_rect.bottom_right.y);
    rect->color = s->main_rect.color;
    *offset += sizeof(RectangleWire);
    
    // Count
    *(uint32_t*)(buf + *offset) = s->count;
    *offset += 4;
}

int main(void) {
    printf("Test 4: Struct Copy Optimizations\n");
    printf("==================================\n\n");
    
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
    volatile size_t offset;
    
    // Verify all produce same output
    uint8_t buf1[256], buf2[256], buf3[256], buf4[256];
    size_t off1 = 0, off2 = 0, off3 = 0, off4 = 0;
    
    encode_scene_recursive(buf1, &off1, &scene);
    encode_scene_flattened(buf2, &off2, &scene);
    encode_scene_bulk(buf3, &off3, &scene);
    encode_scene_direct(buf4, &off4, &scene);
    
    if (off1 != off2 || off1 != off3 || off1 != off4) {
        printf("ERROR: Size mismatch! %zu vs %zu vs %zu vs %zu\n", off1, off2, off3, off4);
        return 1;
    }
    
    if (memcmp(buf1, buf2, off1) != 0 || memcmp(buf1, buf3, off1) != 0 || memcmp(buf1, buf4, off1) != 0) {
        printf("ERROR: Output mismatch!\n");
        return 1;
    }
    
    printf("✓ All methods produce identical output (%zu bytes)\n\n", off1);
    
    // Benchmark 1: Recursive
    printf("1. Recursive encoding (function calls):\n");
    uint64_t start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        offset = 0;
        encode_scene_recursive(buf, (size_t*)&offset, &scene);
    }
    uint64_t end = get_nanos();
    double recursive = (double)(end - start) / ITERATIONS;
    printf("   %.2f ns/op\n\n", recursive);
    
    // Benchmark 2: Flattened
    printf("2. Flattened encoding (inline fields):\n");
    start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        offset = 0;
        encode_scene_flattened(buf, (size_t*)&offset, &scene);
    }
    end = get_nanos();
    double flattened = (double)(end - start) / ITERATIONS;
    printf("   %.2f ns/op\n", flattened);
    printf("   %.1fx faster than recursive\n\n", recursive / flattened);
    
    // Benchmark 3: Bulk copy
    printf("3. Bulk copy with wire struct:\n");
    start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        offset = 0;
        encode_scene_bulk(buf, (size_t*)&offset, &scene);
    }
    end = get_nanos();
    double bulk = (double)(end - start) / ITERATIONS;
    printf("   %.2f ns/op\n", bulk);
    printf("   %.1fx faster than recursive\n\n", recursive / bulk);
    
    // Benchmark 4: Direct pointer
    printf("4. Direct pointer writes:\n");
    start = get_nanos();
    for (int i = 0; i < ITERATIONS; i++) {
        offset = 0;
        encode_scene_direct(buf, (size_t*)&offset, &scene);
    }
    end = get_nanos();
    double direct = (double)(end - start) / ITERATIONS;
    printf("   %.2f ns/op\n", direct);
    printf("   %.1fx faster than recursive\n\n", recursive / direct);
    
    printf("Summary:\n");
    printf("--------\n");
    printf("Recursive:  %.2f ns (baseline)\n", recursive);
    printf("Flattened:  %.2f ns (%.0f%% speedup)\n", flattened, ((recursive - flattened) / recursive) * 100);
    printf("Bulk copy:  %.2f ns (%.0f%% speedup)\n", bulk, ((recursive - bulk) / recursive) * 100);
    printf("Direct:     %.2f ns (%.0f%% speedup)\n\n", direct, ((recursive - direct) / recursive) * 100);
    
    printf("Recommendation:\n");
    printf("  - Avoid recursive function calls for nested structs\n");
    printf("  - Inline struct field encoding when possible\n");
    printf("  - Use wire format structs for bulk copy\n");
    printf("  - Direct pointer writes are fastest\n");
    
    return 0;
}
