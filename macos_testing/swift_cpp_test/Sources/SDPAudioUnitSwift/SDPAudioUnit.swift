//
//  SDPAudioUnit.swift
//  Swift 6 wrapper for C++ SDP implementation
//
//  Uses Swift 6's native C++ interop to directly call C++ functions.
//  Provides idiomatic Swift API with value types.
//

import Foundation
import SDPAudioUnitCpp  // Import C++ module

/// Swift-native parameter type
public struct SDPParameter: Sendable {
    public let address: UInt64
    public let displayName: String
    public let identifier: String
    public let unit: String
    public let minValue: Float
    public let maxValue: Float
    public let defaultValue: Float
    public let currentValue: Float
    public let rawFlags: UInt32
    public let isWritable: Bool
    public let canRamp: Bool
    
    public init(
        address: UInt64,
        displayName: String,
        identifier: String,
        unit: String,
        minValue: Float,
        maxValue: Float,
        defaultValue: Float,
        currentValue: Float,
        rawFlags: UInt32,
        isWritable: Bool,
        canRamp: Bool
    ) {
        self.address = address
        self.displayName = displayName
        self.identifier = identifier
        self.unit = unit
        self.minValue = minValue
        self.maxValue = maxValue
        self.defaultValue = defaultValue
        self.currentValue = currentValue
        self.rawFlags = rawFlags
        self.isWritable = isWritable
        self.canRamp = canRamp
    }
    
    // Convert from C++ type
    init(cppParameter: sdp.Parameter) {
        self.address = cppParameter.address
        self.displayName = String(cppParameter.display_name)
        self.identifier = String(cppParameter.identifier)
        self.unit = String(cppParameter.unit)
        self.minValue = cppParameter.min_value
        self.maxValue = cppParameter.max_value
        self.defaultValue = cppParameter.default_value
        self.currentValue = cppParameter.current_value
        self.rawFlags = cppParameter.raw_flags
        self.isWritable = cppParameter.is_writable
        self.canRamp = cppParameter.can_ramp
    }
    
    // Convert to C++ type
    func toCpp() -> sdp.Parameter {
        var cpp = sdp.Parameter()
        cpp.address = address
        cpp.display_name = std.string(displayName)
        cpp.identifier = std.string(identifier)
        cpp.unit = std.string(unit)
        cpp.min_value = minValue
        cpp.max_value = maxValue
        cpp.default_value = defaultValue
        cpp.current_value = currentValue
        cpp.raw_flags = rawFlags
        cpp.is_writable = isWritable
        cpp.can_ramp = canRamp
        return cpp
    }
}

/// Swift-native plugin type
public struct SDPPlugin: Sendable {
    public let name: String
    public let manufacturerID: String
    public let componentType: String
    public let componentSubtype: String
    public let parameters: [SDPParameter]
    
    public init(
        name: String,
        manufacturerID: String,
        componentType: String,
        componentSubtype: String,
        parameters: [SDPParameter]
    ) {
        self.name = name
        self.manufacturerID = manufacturerID
        self.componentType = componentType
        self.componentSubtype = componentSubtype
        self.parameters = parameters
    }
    
    // Convert from C++ type
    init(cppPlugin: sdp.Plugin) {
        self.name = String(cppPlugin.name)
        self.manufacturerID = String(cppPlugin.manufacturer_id)
        self.componentType = String(cppPlugin.component_type)
        self.componentSubtype = String(cppPlugin.component_subtype)
        self.parameters = cppPlugin.parameters.map { SDPParameter(cppParameter: $0) }
    }
    
    // Convert to C++ type
    func toCpp() -> sdp.Plugin {
        var cpp = sdp.Plugin()
        cpp.name = std.string(name)
        cpp.manufacturer_id = std.string(manufacturerID)
        cpp.component_type = std.string(componentType)
        cpp.component_subtype = std.string(componentSubtype)
        cpp.parameters = std.vector<sdp.Parameter>(parameters.map { $0.toCpp() })
        return cpp
    }
}

/// Swift-native plugin registry type
public struct SDPPluginRegistry: Sendable {
    public let plugins: [SDPPlugin]
    public let totalPluginCount: UInt32
    public let totalParameterCount: UInt32
    
    public init(
        plugins: [SDPPlugin],
        totalPluginCount: UInt32,
        totalParameterCount: UInt32
    ) {
        self.plugins = plugins
        self.totalPluginCount = totalPluginCount
        self.totalParameterCount = totalParameterCount
    }
    
    // Convert from C++ type
    init(cppRegistry: sdp.PluginRegistry) {
        self.plugins = cppRegistry.plugins.map { SDPPlugin(cppPlugin: $0) }
        self.totalPluginCount = cppRegistry.total_plugin_count
        self.totalParameterCount = cppRegistry.total_parameter_count
    }
    
    // Convert to C++ type
    func toCpp() -> sdp.PluginRegistry {
        var cpp = sdp.PluginRegistry()
        cpp.plugins = std.vector<sdp.Plugin>(plugins.map { $0.toCpp() })
        cpp.total_plugin_count = totalPluginCount
        cpp.total_parameter_count = totalParameterCount
        return cpp
    }
}

/// Swift codec for SDP AudioUnit data
public enum SDPAudioUnitCodec {
    
    public enum CodecError: Error {
        case decodeFailed(String)
        case encodeFailed(String)
    }
    
    /// Decode binary SDP data to Swift structs
    public static func decode(_ data: Data) throws -> SDPPluginRegistry {
        try data.withUnsafeBytes { buffer in
            guard let baseAddress = buffer.baseAddress else {
                throw CodecError.decodeFailed("Invalid data buffer")
            }
            
            let bytes = baseAddress.assumingMemoryBound(to: UInt8.self)
            
            // Call C++ decoder
            let cppRegistry = sdp.plugin_registry_decode(bytes, buffer.count)
            
            // Convert to Swift
            return SDPPluginRegistry(cppRegistry: cppRegistry)
        }
    }
    
    /// Encode Swift structs to binary SDP data
    public static func encode(_ registry: SDPPluginRegistry) throws -> Data {
        // Convert to C++
        let cppRegistry = registry.toCpp()
        
        // Get size
        let size = sdp.plugin_registry_size(cppRegistry)
        
        // Allocate buffer
        var buffer = [UInt8](repeating: 0, count: size)
        
        // Call C++ encoder
        let encoded = buffer.withUnsafeMutableBytes { bufferPtr in
            sdp.plugin_registry_encode(
                cppRegistry,
                bufferPtr.baseAddress!.assumingMemoryBound(to: UInt8.self)
            )
        }
        
        // Return as Data
        return Data(buffer.prefix(encoded))
    }
    
    /// Get encoded size for a registry
    public static func size(_ registry: SDPPluginRegistry) -> Int {
        let cppRegistry = registry.toCpp()
        return sdp.plugin_registry_size(cppRegistry)
    }
}
