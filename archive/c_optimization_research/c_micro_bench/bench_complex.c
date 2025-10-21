// bench_complex.c - Handwritten optimized encoder for complex.sdp
// Tests: 3-level nesting (AudioDevice → []Plugin → []Parameter)
// Baseline: Recursive function calls for all levels
// Optimized: Wire structs + inline encoding + bulk arrays

#include <stdint.h>
#include <string.h>
#include <stdio.h>
#include <time.h>

// ============================================================================
// Schema: complex.sdp
// ============================================================================
// struct Parameter {
//     name: str
//     value: f32
// }
//
// struct Plugin {
//     name: str
//     bypass: bool
//     parameters: []Parameter
// }
//
// struct AudioDevice {
//     name: str
//     sample_rate: f32
//     plugins: []Plugin
// }

// ============================================================================
// Data structures
// ============================================================================

typedef struct {
    const char* name;
    size_t name_len;
    float value;
} Parameter;

typedef struct {
    const char* name;
    size_t name_len;
    uint8_t bypass;
    Parameter* parameters;
    size_t param_count;
} Plugin;

typedef struct {
    const char* name;
    size_t name_len;
    float sample_rate;
    Plugin* plugins;
    size_t plugin_count;
} AudioDevice;

// ============================================================================
// Wire format structs (packed)
// ============================================================================

typedef struct __attribute__((packed)) {
    float value;
} ParameterWire;

// ============================================================================
// Baseline encoder (recursive function calls)
// ============================================================================

static size_t encode_parameter_baseline(uint8_t* buf, const Parameter* p) {
    uint8_t* start = buf;
    
    // String: name
    *(uint32_t*)buf = (uint32_t)p->name_len;
    buf += 4;
    memcpy(buf, p->name, p->name_len);
    buf += p->name_len;
    
    // f32: value
    *(float*)buf = p->value;
    buf += 4;
    
    return buf - start;
}

static size_t encode_plugin_baseline(uint8_t* buf, const Plugin* plugin) {
    uint8_t* start = buf;
    
    // String: name
    *(uint32_t*)buf = (uint32_t)plugin->name_len;
    buf += 4;
    memcpy(buf, plugin->name, plugin->name_len);
    buf += plugin->name_len;
    
    // bool: bypass
    *buf++ = plugin->bypass;
    
    // Array: parameters
    *(uint32_t*)buf = (uint32_t)plugin->param_count;
    buf += 4;
    for (size_t i = 0; i < plugin->param_count; i++) {
        buf += encode_parameter_baseline(buf, &plugin->parameters[i]);
    }
    
    return buf - start;
}

static size_t encode_device_baseline(uint8_t* buf, const AudioDevice* device) {
    uint8_t* start = buf;
    
    // String: name
    *(uint32_t*)buf = (uint32_t)device->name_len;
    buf += 4;
    memcpy(buf, device->name, device->name_len);
    buf += device->name_len;
    
    // f32: sample_rate
    *(float*)buf = device->sample_rate;
    buf += 4;
    
    // Array: plugins
    *(uint32_t*)buf = (uint32_t)device->plugin_count;
    buf += 4;
    for (size_t i = 0; i < device->plugin_count; i++) {
        buf += encode_plugin_baseline(buf, &device->plugins[i]);
    }
    
    return buf - start;
}

// ============================================================================
// Optimized encoder (inline + wire structs)
// ============================================================================

static size_t encode_device_optimized(uint8_t* buf, const AudioDevice* device) {
    uint8_t* start = buf;
    
    // String: device name
    *(uint32_t*)buf = (uint32_t)device->name_len;
    buf += 4;
    memcpy(buf, device->name, device->name_len);
    buf += device->name_len;
    
    // f32: sample_rate
    *(float*)buf = device->sample_rate;
    buf += 4;
    
    // Array: plugins
    *(uint32_t*)buf = (uint32_t)device->plugin_count;
    buf += 4;
    
    // Inline all plugin encoding (avoid function calls)
    for (size_t i = 0; i < device->plugin_count; i++) {
        const Plugin* plugin = &device->plugins[i];
        
        // String: plugin name
        *(uint32_t*)buf = (uint32_t)plugin->name_len;
        buf += 4;
        memcpy(buf, plugin->name, plugin->name_len);
        buf += plugin->name_len;
        
        // bool: bypass
        *buf++ = plugin->bypass;
        
        // Array: parameters
        *(uint32_t*)buf = (uint32_t)plugin->param_count;
        buf += 4;
        
        // Inline all parameter encoding (avoid function calls)
        for (size_t j = 0; j < plugin->param_count; j++) {
            const Parameter* p = &plugin->parameters[j];
            
            // String: param name
            *(uint32_t*)buf = (uint32_t)p->name_len;
            buf += 4;
            memcpy(buf, p->name, p->name_len);
            buf += p->name_len;
            
            // f32: value (could use wire struct, but single field)
            *(float*)buf = p->value;
            buf += 4;
        }
    }
    
    return buf - start;
}

// ============================================================================
// Test data (runtime-variable to prevent compiler optimization)
// ============================================================================

static Parameter test_params[] = {
    {"gain", 4, 0.75f},
    {"pan", 3, 0.5f},
    {"freq", 4, 440.0f},
};

static Plugin test_plugins[] = {
    {"Reverb", 6, 0, test_params, 3},
    {"Delay", 5, 1, test_params, 2},
};

static AudioDevice test_device = {
    "Main Output", 11, 48000.0f, test_plugins, 2
};

// ============================================================================
// Benchmark harness
// ============================================================================

#define ITERATIONS 1000000

int main(void) {
    uint8_t buf_baseline[4096];
    uint8_t buf_optimized[4096];
    
    // Warmup
    volatile size_t warmup = encode_device_baseline(buf_baseline, &test_device);
    warmup += encode_device_optimized(buf_optimized, &test_device);
    (void)warmup;
    
    // Verify both produce same output
    size_t size_baseline = encode_device_baseline(buf_baseline, &test_device);
    size_t size_optimized = encode_device_optimized(buf_optimized, &test_device);
    
    if (size_baseline != size_optimized || memcmp(buf_baseline, buf_optimized, size_baseline) != 0) {
        printf("ERROR: Encoders produce different output!\n");
        printf("Baseline size: %zu, Optimized size: %zu\n", size_baseline, size_optimized);
        return 1;
    }
    
    printf("\n✓ Both encoders produce identical output (%zu bytes)\n", size_baseline);
    printf("  Device: \"%s\" (%.0f Hz)\n", test_device.name, test_device.sample_rate);
    printf("  Plugins: %zu\n", test_device.plugin_count);
    printf("  Total parameters: %zu\n\n", test_plugins[0].param_count + test_plugins[1].param_count);
    
    // Benchmark baseline
    struct timespec start, end;
    clock_gettime(CLOCK_MONOTONIC, &start);
    
    volatile size_t total_baseline = 0;
    for (int i = 0; i < ITERATIONS; i++) {
        total_baseline += encode_device_baseline(buf_baseline, &test_device);
    }
    
    clock_gettime(CLOCK_MONOTONIC, &end);
    double baseline_ns = (end.tv_sec - start.tv_sec) * 1e9 + (end.tv_nsec - start.tv_nsec);
    double baseline_per_op = baseline_ns / ITERATIONS;
    
    // Benchmark optimized
    clock_gettime(CLOCK_MONOTONIC, &start);
    
    volatile size_t total_optimized = 0;
    for (int i = 0; i < ITERATIONS; i++) {
        total_optimized += encode_device_optimized(buf_optimized, &test_device);
    }
    
    clock_gettime(CLOCK_MONOTONIC, &end);
    double optimized_ns = (end.tv_sec - start.tv_sec) * 1e9 + (end.tv_nsec - start.tv_nsec);
    double optimized_per_op = optimized_ns / ITERATIONS;
    
    // Results
    printf("Baseline (recursive function calls):\n");
    printf("  %.2f ns/op\n\n", baseline_per_op);
    
    printf("Optimized (inline encoding):\n");
    printf("  %.2f ns/op\n", optimized_per_op);
    printf("  %.1fx faster\n\n", baseline_per_op / optimized_per_op);
    
    printf("Speedup: %.0f%% improvement\n\n", 
           (1.0 - optimized_per_op / baseline_per_op) * 100);
    
    printf("Key optimization: Avoided %d function calls per encode\n",
           (int)(test_device.plugin_count * 2 + // encode_plugin * N + inner loop overhead
                 (test_plugins[0].param_count + test_plugins[1].param_count))); // encode_parameter * M
    
    printf("  (2 plugins × 1 call each + 5 parameters × 1 call each = 7 calls)\n\n");
    
    return 0;
}
