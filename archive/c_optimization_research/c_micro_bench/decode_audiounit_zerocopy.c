/* decode_audiounit_zerocopy.c - Zero-copy AudioUnit decoder
 * 
 * Tests decode performance on complex real-world schema with:
 * - Nested struct arrays
 * - Multiple strings
 * - Mixed primitive types
 */

#include <stdio.h>
#include <stdint.h>
#include <string.h>
#include <time.h>

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

static inline float le_to_f32(const uint8_t* buf) {
    uint32_t u;
    memcpy(&u, buf, 4);
    u = LE32TOH(u);
    float f;
    memcpy(&f, &u, 4);
    return f;
}

/* Zero-copy structs */
typedef struct {
    uint64_t address;
    const char* display_name;
    size_t display_name_len;
    const char* identifier;
    size_t identifier_len;
    const char* unit;
    size_t unit_len;
    float min_value;
    float max_value;
    float default_value;
    float current_value;
    uint32_t raw_flags;
    uint8_t is_writable;
    uint8_t can_ramp;
} Parameter;

typedef struct {
    const char* name;
    size_t name_len;
    const char* manufacturer_id;
    size_t manufacturer_id_len;
    const char* component_type;
    size_t component_type_len;
    const char* component_subtype;
    size_t component_subtype_len;
    Parameter* parameters;  /* Array (zero-copy pointers) */
    size_t parameters_len;
} Plugin;

/* Decode parameter array */
static int decode_parameters(Parameter* params, const uint8_t* buf, size_t buf_len, size_t* offset, size_t count) {
    for (size_t i = 0; i < count; i++) {
        Parameter* p = &params[i];
        
        #define CHECK(n) if (*offset + (n) > buf_len) return -1
        
        /* u64 address */
        CHECK(8);
        p->address = LE64TOH(*(const uint64_t*)(buf + *offset));
        *offset += 8;
        
        /* display_name */
        CHECK(4);
        uint32_t len = LE32TOH(*(const uint32_t*)(buf + *offset));
        *offset += 4;
        CHECK(len);
        p->display_name = (const char*)(buf + *offset);
        p->display_name_len = len;
        *offset += len;
        
        /* identifier */
        CHECK(4);
        len = LE32TOH(*(const uint32_t*)(buf + *offset));
        *offset += 4;
        CHECK(len);
        p->identifier = (const char*)(buf + *offset);
        p->identifier_len = len;
        *offset += len;
        
        /* unit */
        CHECK(4);
        len = LE32TOH(*(const uint32_t*)(buf + *offset));
        *offset += 4;
        CHECK(len);
        p->unit = (const char*)(buf + *offset);
        p->unit_len = len;
        *offset += len;
        
        /* floats */
        CHECK(16);
        p->min_value = le_to_f32(buf + *offset);
        *offset += 4;
        p->max_value = le_to_f32(buf + *offset);
        *offset += 4;
        p->default_value = le_to_f32(buf + *offset);
        *offset += 4;
        p->current_value = le_to_f32(buf + *offset);
        *offset += 4;
        
        /* u32 raw_flags */
        CHECK(4);
        p->raw_flags = LE32TOH(*(const uint32_t*)(buf + *offset));
        *offset += 4;
        
        /* bools */
        CHECK(2);
        p->is_writable = buf[*offset];
        *offset += 1;
        p->can_ramp = buf[*offset];
        *offset += 1;
        
        #undef CHECK
    }
    return 0;
}

/* Decode plugin (requires temp buffer for parameter array) */
static int decode_plugin(Plugin* dest, const uint8_t* buf, size_t buf_len, Parameter* param_storage) {
    size_t offset = 0;
    
    #define CHECK(n) if (offset + (n) > buf_len) return -1
    
    /* name */
    CHECK(4);
    uint32_t len = LE32TOH(*(const uint32_t*)(buf + offset));
    offset += 4;
    CHECK(len);
    dest->name = (const char*)(buf + offset);
    dest->name_len = len;
    offset += len;
    
    /* manufacturer_id */
    CHECK(4);
    len = LE32TOH(*(const uint32_t*)(buf + offset));
    offset += 4;
    CHECK(len);
    dest->manufacturer_id = (const char*)(buf + offset);
    dest->manufacturer_id_len = len;
    offset += len;
    
    /* component_type */
    CHECK(4);
    len = LE32TOH(*(const uint32_t*)(buf + offset));
    offset += 4;
    CHECK(len);
    dest->component_type = (const char*)(buf + offset);
    dest->component_type_len = len;
    offset += len;
    
    /* component_subtype */
    CHECK(4);
    len = LE32TOH(*(const uint32_t*)(buf + offset));
    offset += 4;
    CHECK(len);
    dest->component_subtype = (const char*)(buf + offset);
    dest->component_subtype_len = len;
    offset += len;
    
    /* parameters array */
    CHECK(4);
    uint32_t param_count = LE32TOH(*(const uint32_t*)(buf + offset));
    offset += 4;
    
    dest->parameters = param_storage;
    dest->parameters_len = param_count;
    
    if (decode_parameters(dest->parameters, buf, buf_len, &offset, param_count) != 0) {
        return -1;
    }
    
    #undef CHECK
    return 0;
}

/* Test data (encoded Plugin with 2 parameters) */
static const uint8_t test_data[] = {
    /* name: "TestPlugin" (10 bytes) */
    0x0a, 0x00, 0x00, 0x00, 'T', 'e', 's', 't', 'P', 'l', 'u', 'g', 'i', 'n',
    /* manufacturer_id: "ACME" (4 bytes) */
    0x04, 0x00, 0x00, 0x00, 'A', 'C', 'M', 'E',
    /* component_type: "aufx" (4 bytes) */
    0x04, 0x00, 0x00, 0x00, 'a', 'u', 'f', 'x',
    /* component_subtype: "test" (4 bytes) */
    0x04, 0x00, 0x00, 0x00, 't', 'e', 's', 't',
    /* parameters count: 2 */
    0x02, 0x00, 0x00, 0x00,
    /* Parameter 0 */
    0x00, 0x10, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,  /* address: 0x1000 */
    0x06, 0x00, 0x00, 0x00, 'V', 'o', 'l', 'u', 'm', 'e',  /* display_name */
    0x03, 0x00, 0x00, 0x00, 'v', 'o', 'l',  /* identifier */
    0x02, 0x00, 0x00, 0x00, 'd', 'B',  /* unit */
    0x00, 0x00, 0xc0, 0xc2,  /* min: -96.0 */
    0x00, 0x00, 0xc0, 0x40,  /* max: 6.0 */
    0x00, 0x00, 0x00, 0x00,  /* default: 0.0 */
    0x00, 0x00, 0x40, 0xc0,  /* current: -3.0 */
    0x01, 0x00, 0x00, 0x00,  /* raw_flags: 1 */
    0x01, 0x01,  /* is_writable, can_ramp */
    /* Parameter 1 */
    0x00, 0x20, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,  /* address: 0x2000 */
    0x03, 0x00, 0x00, 0x00, 'P', 'a', 'n',  /* display_name */
    0x03, 0x00, 0x00, 0x00, 'p', 'a', 'n',  /* identifier */
    0x01, 0x00, 0x00, 0x00, '%',  /* unit */
    0x00, 0x00, 0xc8, 0xc2,  /* min: -100.0 */
    0x00, 0x00, 0xc8, 0x42,  /* max: 100.0 */
    0x00, 0x00, 0x00, 0x00,  /* default: 0.0 */
    0x00, 0x00, 0x00, 0x00,  /* current: 0.0 */
    0x02, 0x00, 0x00, 0x00,  /* raw_flags: 2 */
    0x01, 0x01   /* is_writable, can_ramp */
};

int main(void) {
    Plugin decoded;
    Parameter param_storage[10];  /* Temp storage for parameters */
    
    /* Warmup */
    for (int i = 0; i < 1000; i++) {
        if (decode_plugin(&decoded, test_data, sizeof(test_data), param_storage) != 0) {
            fprintf(stderr, "Decode failed\n");
            return 1;
        }
    }
    
    /* Benchmark */
    const int iterations = 10000000;
    volatile uint64_t sink = 0;
    
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    
    for (int i = 0; i < iterations; i++) {
        if (decode_plugin(&decoded, test_data, sizeof(test_data), param_storage) != 0) {
            fprintf(stderr, "Decode failed\n");
            return 1;
        }
        sink += decoded.parameters[0].address;
    }
    
    clock_gettime(CLOCK_MONOTONIC, &end);
    
    uint64_t start_ns = start.tv_sec * 1000000000ULL + start.tv_nsec;
    uint64_t end_ns = end.tv_sec * 1000000000ULL + end.tv_nsec;
    uint64_t total_ns = end_ns - start_ns;
    double ns_per_op = (double)total_ns / iterations;
    
    printf("=== Zero-Copy Decode (AudioUnit Plugin) ===\n");
    printf("Iterations: %d\n", iterations);
    printf("Total time: %.2f ms\n", total_ns / 1e6);
    printf("Time per op: %.2f ns\n", ns_per_op);
    printf("Throughput: %.2f million ops/sec\n", 1000.0 / ns_per_op);
    
    /* Verify */
    decode_plugin(&decoded, test_data, sizeof(test_data), param_storage);
    printf("\nVerification:\n");
    printf("  name: '%.*s'\n", (int)decoded.name_len, decoded.name);
    printf("  parameters_len: %zu\n", decoded.parameters_len);
    printf("  param[0].address: 0x%llx\n", (unsigned long long)decoded.parameters[0].address);
    printf("  param[0].display_name: '%.*s'\n", 
           (int)decoded.parameters[0].display_name_len, decoded.parameters[0].display_name);
    printf("  param[0].min_value: %.1f\n", decoded.parameters[0].min_value);
    
    return 0;
}
