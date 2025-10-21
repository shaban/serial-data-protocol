//
//  sdp_bridge.h
//  C wrapper for C++ SDP functions (callable from Swift)
//
//  This provides a clean C API that Swift can import directly.
//  Keeps data in C++ (opaque pointer), minimal overhead.
//

#ifndef SDP_BRIDGE_H
#define SDP_BRIDGE_H

#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

// Opaque pointer to C++ PluginRegistry
typedef struct SDPPluginRegistry SDPPluginRegistry;

// Decode: returns opaque pointer to C++ struct (or NULL on error)
SDPPluginRegistry* sdp_bridge_decode(const uint8_t* data, size_t len);

// Encode: takes opaque pointer, returns malloc'd bytes (caller must free)
uint8_t* sdp_bridge_encode(SDPPluginRegistry* reg, size_t* out_len);

// Free the registry
void sdp_bridge_free(SDPPluginRegistry* reg);

// Read-only accessors (no allocation)
uint32_t sdp_bridge_total_plugins(SDPPluginRegistry* reg);
uint32_t sdp_bridge_total_parameters(SDPPluginRegistry* reg);
size_t sdp_bridge_plugin_count(SDPPluginRegistry* reg);

// Get plugin name by index (returns pointer to internal C++ string - do not free!)
const char* sdp_bridge_plugin_name(SDPPluginRegistry* reg, size_t index);

#ifdef __cplusplus
}
#endif

#endif // SDP_BRIDGE_H
