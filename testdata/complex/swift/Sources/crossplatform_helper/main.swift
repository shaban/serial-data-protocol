// Cross-platform test helper for swift
// Usage:
//   swift run -c release crossplatform_helper encode-Parameter > output.bin
//   swift run -c release crossplatform_helper decode-Parameter input.bin
//   swift run -c release crossplatform_helper encode-Plugin > output.bin
//   swift run -c release crossplatform_helper decode-Plugin input.bin
//   swift run -c release crossplatform_helper encode-AudioDevice > output.bin
//   swift run -c release crossplatform_helper decode-AudioDevice input.bin

import Foundation
import swift

func makeTestParameter() -> Parameter {
    return Parameter(
        id: 4_294_967_295,
        name: "Hello from Swift!",
        value: 3.14159,
        min: 3.14159,
        max: 3.14159
    )
}

func makeTestPlugin() -> Plugin {
    return Plugin(
        id: 4_294_967_295,
        name: "Hello from Swift!",
        manufacturer: "Hello from Swift!",
        version: 4_294_967_295,
        enabled: true,
        parameters: ContiguousArray([makeTestParameter(), makeTestParameter(), makeTestParameter()])
    )
}

func makeTestAudioDevice() -> AudioDevice {
    return AudioDevice(
        deviceId: 4_294_967_295,
        deviceName: "Hello from Swift!",
        sampleRate: 4_294_967_295,
        bufferSize: 4_294_967_295,
        inputChannels: 65535,
        outputChannels: 65535,
        isDefault: true,
        activePlugins: ContiguousArray([makeTestPlugin(), makeTestPlugin(), makeTestPlugin()])
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
    ok = ok && decoded.id == 4_294_967_295
    ok = ok && abs(decoded.value - 3.14159) < 0.0001
    ok = ok && abs(decoded.min - 3.14159) < 0.0001
    ok = ok && abs(decoded.max - 3.14159) < 0.0001

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
    ok = ok && decoded.id == 4_294_967_295
    ok = ok && decoded.version == 4_294_967_295
    ok = ok && decoded.enabled == true
    ok = ok && decoded.parameters.count > 0

    if !ok {
        fputs("Validation failed\n", stderr)
        fputs("Decoded: \(decoded)\n", stderr)
        exit(1)
    }

    fputs("✓ Swift successfully decoded and validated\n", stderr)
}

func encodeAudioDevice() throws {
    let data = makeTestAudioDevice()

    let bytes = data.encodeToBytes()
    let binaryData = Data(bytes)

    FileHandle.standardOutput.write(binaryData)
}

func decodeAudioDevice(filename: String) throws {
    let fileData = try Data(contentsOf: URL(fileURLWithPath: filename))
    let bytes = [UInt8](fileData)
    let decoded = try AudioDevice.decode(from: bytes)

    // Basic validation
    var ok = true
    ok = ok && decoded.deviceId == 4_294_967_295
    ok = ok && decoded.sampleRate == 4_294_967_295
    ok = ok && decoded.bufferSize == 4_294_967_295
    ok = ok && decoded.inputChannels == 65535
    ok = ok && decoded.outputChannels == 65535
    ok = ok && decoded.isDefault == true
    ok = ok && decoded.activePlugins.count > 0

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
    fputs("  encode-AudioDevice - Encode AudioDevice to stdout\n", stderr)
    fputs("  decode-AudioDevice <file> - Decode AudioDevice from file\n", stderr)
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
    case "encode-AudioDevice":
        try encodeAudioDevice()
    case "decode-AudioDevice":
        guard args.count >= 3 else {
            fputs("Error: decode-AudioDevice requires filename argument\n", stderr)
            exit(1)
        }
        try decodeAudioDevice(filename: args[2])
    default:
        fputs("Unknown command: \(command)\n", stderr)
        exit(1)
    }
} catch {
    fputs("Error: \(error)\n", stderr)
    exit(1)
}
