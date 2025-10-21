//
//  sdp_bridge.cpp
//  C wrapper implementation
//

#include "sdp_bridge.h"
#include "../../testdata/audiounit_cpp/types.hpp"
#include "../../testdata/audiounit_cpp/decode.hpp"
#include "../../testdata/audiounit_cpp/encode.hpp"

using namespace sdp;

extern "C" {

SDPPluginRegistry* sdp_bridge_decode(const uint8_t* data, size_t len) {
    try {
        // Decode and move into heap allocation
        auto reg = new PluginRegistry(plugin_registry_decode(data, len));
        return reinterpret_cast<SDPPluginRegistry*>(reg);
    } catch (...) {
        return nullptr;
    }
}

uint8_t* sdp_bridge_encode(SDPPluginRegistry* reg, size_t* out_len) {
    if (!reg) {
        *out_len = 0;
        return nullptr;
    }
    
    try {
        auto* cpp_reg = reinterpret_cast<PluginRegistry*>(reg);
        size_t size = plugin_registry_size(*cpp_reg);
        uint8_t* buf = static_cast<uint8_t*>(malloc(size));
        if (!buf) {
            *out_len = 0;
            return nullptr;
        }
        *out_len = plugin_registry_encode(*cpp_reg, buf);
        return buf;
    } catch (...) {
        *out_len = 0;
        return nullptr;
    }
}

void sdp_bridge_free(SDPPluginRegistry* reg) {
    if (reg) {
        delete reinterpret_cast<PluginRegistry*>(reg);
    }
}

uint32_t sdp_bridge_total_plugins(SDPPluginRegistry* reg) {
    if (!reg) return 0;
    return reinterpret_cast<PluginRegistry*>(reg)->total_plugin_count;
}

uint32_t sdp_bridge_total_parameters(SDPPluginRegistry* reg) {
    if (!reg) return 0;
    return reinterpret_cast<PluginRegistry*>(reg)->total_parameter_count;
}

size_t sdp_bridge_plugin_count(SDPPluginRegistry* reg) {
    if (!reg) return 0;
    return reinterpret_cast<PluginRegistry*>(reg)->plugins.size();
}

const char* sdp_bridge_plugin_name(SDPPluginRegistry* reg, size_t index) {
    if (!reg) return "";
    auto* cpp_reg = reinterpret_cast<PluginRegistry*>(reg);
    if (index >= cpp_reg->plugins.size()) return "";
    return cpp_reg->plugins[index].name.c_str();
}

} // extern "C"
