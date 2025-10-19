// Cross-platform test helper for swift
// Usage:
//   swift run -c release crossplatform_helper encode-Parameter > output.bin
//   swift run -c release crossplatform_helper decode-Parameter input.bin
//   swift run -c release crossplatform_helper encode-Plugin > output.bin
//   swift run -c release crossplatform_helper decode-Plugin input.bin
//   swift run -c release crossplatform_helper encode-PluginRegistry > output.bin
//   swift run -c release crossplatform_helper decode-PluginRegistry input.bin

import Foundation
import swift

func makeTestParameter() -> Parameter {
    return Parameter(
        address: 18_446_744_073_709_551_615,
        displayName: "Hello from Swift!",
        identifier: "Hello from Swift!",
        unit: "Hello from Swift!",
        minValue: 3.14159,
        maxValue: 3.14159,
        defaultValue: 3.14159,
        currentValue: 3.14159,
        rawFlags: 4_294_967_295,
        isWritable: true,
        canRamp: true
    )
}

func makeTestPlugin() -> Plugin {
    return Plugin(
        name: "Hello from Swift!",
        manufacturerId: "Hello from Swift!",
        componentType: "Hello from Swift!",
        componentSubtype: "Hello from Swift!",
        parameters: ContiguousArray([makeTestParameter(), makeTestParameter(), makeTestParameter()])
    )
}

func makeTestPluginRegistry() -> PluginRegistry {
    return PluginRegistry(
        plugins: ContiguousArray([makeTestPlugin(), makeTestPlugin(), makeTestPlugin()]),
        totalPluginCount: 4_294_967_295,
        totalParameterCount: 4_294_967_295
    )
}

func encodeParameter() throws {
    let data = makeTestParameter()

    let bytes = data.encodeToBytes()
    let binaryData = Data(bytes)

    FileHandle.standardOutput.write(binaryData)
}

func decodeParameter(filename: String) throws {
    let fileData = try Data(contentsOf: URL(fileURLWithPath: filename))
    let bytes = [UInt8](fileData)
    let decoded = try Parameter.decode(from: bytes)

    // Basic validation
    var ok = true
    ok = ok && decoded.address == 18_446_744_073_709_551_615
    ok = ok && abs(decoded.minValue - 3.14159) < 0.0001
    ok = ok && abs(decoded.maxValue - 3.14159) < 0.0001
    ok = ok && abs(decoded.defaultValue - 3.14159) < 0.0001
    ok = ok && abs(decoded.currentValue - 3.14159) < 0.0001
    ok = ok && decoded.rawFlags == 4_294_967_295
    ok = ok && decoded.isWritable == true
    ok = ok && decoded.canRamp == true

    if !ok {
        fputs("Validation failed\n", stderr)
        fputs("Decoded: \(decoded)\n", stderr)
        exit(1)
    }

    fputs("✓ Swift successfully decoded and validated\n", stderr)
}

func encodePlugin() throws {
    let data = makeTestPlugin()

    let bytes = data.encodeToBytes()
    let binaryData = Data(bytes)

    FileHandle.standardOutput.write(binaryData)
}

func decodePlugin(filename: String) throws {
    let fileData = try Data(contentsOf: URL(fileURLWithPath: filename))
    let bytes = [UInt8](fileData)
    let decoded = try Plugin.decode(from: bytes)

    // Basic validation
    var ok = true
    ok = ok && decoded.parameters.count > 0

    if !ok {
        fputs("Validation failed\n", stderr)
        fputs("Decoded: \(decoded)\n", stderr)
        exit(1)
    }

    fputs("✓ Swift successfully decoded and validated\n", stderr)
}

func encodePluginRegistry() throws {
    let data = makeTestPluginRegistry()

    let bytes = data.encodeToBytes()
    let binaryData = Data(bytes)

    FileHandle.standardOutput.write(binaryData)
}

func decodePluginRegistry(filename: String) throws {
    let fileData = try Data(contentsOf: URL(fileURLWithPath: filename))
    let bytes = [UInt8](fileData)
    let decoded = try PluginRegistry.decode(from: bytes)

    // Basic validation
    var ok = true
    ok = ok && decoded.plugins.count > 0
    ok = ok && decoded.totalPluginCount == 4_294_967_295
    ok = ok && decoded.totalParameterCount == 4_294_967_295

    if !ok {
        fputs("Validation failed\n", stderr)
        fputs("Decoded: \(decoded)\n", stderr)
        exit(1)
    }

    fputs("✓ Swift successfully decoded and validated\n", stderr)
}

// Main entry point
let args = CommandLine.arguments

if args.count < 2 {
    fputs("Usage: \(args[0]) <command> [args]\n", stderr)
    fputs("Commands:\n", stderr)
    fputs("  encode-Parameter - Encode Parameter to stdout\n", stderr)
    fputs("  decode-Parameter <file> - Decode Parameter from file\n", stderr)
    fputs("  encode-Plugin - Encode Plugin to stdout\n", stderr)
    fputs("  decode-Plugin <file> - Decode Plugin from file\n", stderr)
    fputs("  encode-PluginRegistry - Encode PluginRegistry to stdout\n", stderr)
    fputs("  decode-PluginRegistry <file> - Decode PluginRegistry from file\n", stderr)
    exit(1)
}

let command = args[1]

do {
    switch command {
    case "encode-Parameter":
        try encodeParameter()
    case "decode-Parameter":
        guard args.count >= 3 else {
            fputs("Error: decode-Parameter requires filename argument\n", stderr)
            exit(1)
        }
        try decodeParameter(filename: args[2])
    case "encode-Plugin":
        try encodePlugin()
    case "decode-Plugin":
        guard args.count >= 3 else {
            fputs("Error: decode-Plugin requires filename argument\n", stderr)
            exit(1)
        }
        try decodePlugin(filename: args[2])
    case "encode-PluginRegistry":
        try encodePluginRegistry()
    case "decode-PluginRegistry":
        guard args.count >= 3 else {
            fputs("Error: decode-PluginRegistry requires filename argument\n", stderr)
            exit(1)
        }
        try decodePluginRegistry(filename: args[2])
    default:
        fputs("Unknown command: \(command)\n", stderr)
        exit(1)
    }
} catch {
    fputs("Error: \(error)\n", stderr)
    exit(1)
}
