//
//  SDPAudioUnit.mm
//  Objective-C++ implementation bridging to C++ SDP library
//
//  This .mm file can mix Objective-C and C++ code.
//  It wraps the C++ implementation in Objective-C objects.
//

#import "SDPAudioUnit.h"

// Import C++ headers
#include "../../testdata/audiounit_cpp/types.hpp"
#include "../../testdata/audiounit_cpp/decode.hpp"
#include "../../testdata/audiounit_cpp/encode.hpp"

#include <vector>
#include <string>

using namespace sdp;

// ============================================================================
// Objective-C Wrapper Implementations
// ============================================================================

@implementation SDPParameter
@end

@implementation SDPPlugin
@end

@implementation SDPPluginRegistry
@end

@implementation SDPAudioUnitCodec

// Convert C++ Parameter to Objective-C SDPParameter
+ (SDPParameter *)parameterFromCpp:(const Parameter&)cppParam {
    SDPParameter *param = [[SDPParameter alloc] init];
    param.address = cppParam.address;
    param.displayName = [NSString stringWithUTF8String:cppParam.display_name.c_str()];
    param.identifier = [NSString stringWithUTF8String:cppParam.identifier.c_str()];
    param.unit = [NSString stringWithUTF8String:cppParam.unit.c_str()];
    param.minValue = cppParam.min_value;
    param.maxValue = cppParam.max_value;
    param.defaultValue = cppParam.default_value;
    param.currentValue = cppParam.current_value;
    param.rawFlags = cppParam.raw_flags;
    param.isWritable = cppParam.is_writable;
    param.canRamp = cppParam.can_ramp;
    return param;
}

// Convert C++ Plugin to Objective-C SDPPlugin
+ (SDPPlugin *)pluginFromCpp:(const Plugin&)cppPlugin {
    SDPPlugin *plugin = [[SDPPlugin alloc] init];
    plugin.name = [NSString stringWithUTF8String:cppPlugin.name.c_str()];
    plugin.manufacturerID = [NSString stringWithUTF8String:cppPlugin.manufacturer_id.c_str()];
    plugin.componentType = [NSString stringWithUTF8String:cppPlugin.component_type.c_str()];
    plugin.componentSubtype = [NSString stringWithUTF8String:cppPlugin.component_subtype.c_str()];
    
    NSMutableArray<SDPParameter *> *params = [NSMutableArray arrayWithCapacity:cppPlugin.parameters.size()];
    for (const auto& cppParam : cppPlugin.parameters) {
        [params addObject:[self parameterFromCpp:cppParam]];
    }
    plugin.parameters = params;
    
    return plugin;
}

// Convert C++ PluginRegistry to Objective-C SDPPluginRegistry
+ (SDPPluginRegistry *)registryFromCpp:(const PluginRegistry&)cppRegistry {
    SDPPluginRegistry *registry = [[SDPPluginRegistry alloc] init];
    
    NSMutableArray<SDPPlugin *> *plugins = [NSMutableArray arrayWithCapacity:cppRegistry.plugins.size()];
    for (const auto& cppPlugin : cppRegistry.plugins) {
        [plugins addObject:[self pluginFromCpp:cppPlugin]];
    }
    registry.plugins = plugins;
    registry.totalPluginCount = cppRegistry.total_plugin_count;
    registry.totalParameterCount = cppRegistry.total_parameter_count;
    
    return registry;
}

// Convert Objective-C SDPParameter to C++ Parameter
+ (Parameter)cppFromParameter:(SDPParameter *)param {
    Parameter cppParam;
    cppParam.address = param.address;
    cppParam.display_name = [param.displayName UTF8String];
    cppParam.identifier = [param.identifier UTF8String];
    cppParam.unit = [param.unit UTF8String];
    cppParam.min_value = param.minValue;
    cppParam.max_value = param.maxValue;
    cppParam.default_value = param.defaultValue;
    cppParam.current_value = param.currentValue;
    cppParam.raw_flags = param.rawFlags;
    cppParam.is_writable = param.isWritable;
    cppParam.can_ramp = param.canRamp;
    return cppParam;
}

// Convert Objective-C SDPPlugin to C++ Plugin
+ (Plugin)cppFromPlugin:(SDPPlugin *)plugin {
    Plugin cppPlugin;
    cppPlugin.name = [plugin.name UTF8String];
    cppPlugin.manufacturer_id = [plugin.manufacturerID UTF8String];
    cppPlugin.component_type = [plugin.componentType UTF8String];
    cppPlugin.component_subtype = [plugin.componentSubtype UTF8String];
    
    for (SDPParameter *param in plugin.parameters) {
        cppPlugin.parameters.push_back([self cppFromParameter:param]);
    }
    
    return cppPlugin;
}

// Convert Objective-C SDPPluginRegistry to C++ PluginRegistry
+ (PluginRegistry)cppFromRegistry:(SDPPluginRegistry *)registry {
    PluginRegistry cppRegistry;
    
    for (SDPPlugin *plugin in registry.plugins) {
        cppRegistry.plugins.push_back([self cppFromPlugin:plugin]);
    }
    cppRegistry.total_plugin_count = registry.totalPluginCount;
    cppRegistry.total_parameter_count = registry.totalParameterCount;
    
    return cppRegistry;
}

// Decode binary SDP data to Objective-C objects
+ (nullable SDPPluginRegistry *)decodePluginRegistry:(NSData *)data
                                               error:(NSError **)error {
    @try {
        const uint8_t *bytes = (const uint8_t *)data.bytes;
        size_t length = data.length;
        
        // Call C++ decoder
        PluginRegistry cppRegistry = plugin_registry_decode(bytes, length);
        
        // Convert to Objective-C
        return [self registryFromCpp:cppRegistry];
    }
    @catch (NSException *exception) {
        if (error) {
            *error = [NSError errorWithDomain:@"SDPAudioUnitCodec"
                                        code:-1
                                    userInfo:@{NSLocalizedDescriptionKey: exception.reason}];
        }
        return nil;
    }
}

// Encode Objective-C objects to binary SDP data
+ (nullable NSData *)encodePluginRegistry:(SDPPluginRegistry *)registry
                                    error:(NSError **)error {
    @try {
        // Convert to C++
        PluginRegistry cppRegistry = [self cppFromRegistry:registry];
        
        // Get size and allocate buffer
        size_t size = plugin_registry_size(cppRegistry);
        std::vector<uint8_t> buffer(size);
        
        // Call C++ encoder
        size_t encoded = plugin_registry_encode(cppRegistry, buffer.data());
        
        // Create NSData
        return [NSData dataWithBytes:buffer.data() length:encoded];
    }
    @catch (NSException *exception) {
        if (error) {
            *error = [NSError errorWithDomain:@"SDPAudioUnitCodec"
                                        code:-1
                                    userInfo:@{NSLocalizedDescriptionKey: exception.reason}];
        }
        return nil;
    }
}

// Get encoded size
+ (size_t)pluginRegistrySize:(SDPPluginRegistry *)registry {
    PluginRegistry cppRegistry = [self cppFromRegistry:registry];
    return plugin_registry_size(cppRegistry);
}

@end
