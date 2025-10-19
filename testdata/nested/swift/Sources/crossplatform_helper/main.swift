// Cross-platform test helper for swift
// Usage:
//   swift run -c release crossplatform_helper encode-Point > output.bin
//   swift run -c release crossplatform_helper decode-Point input.bin
//   swift run -c release crossplatform_helper encode-Rectangle > output.bin
//   swift run -c release crossplatform_helper decode-Rectangle input.bin
//   swift run -c release crossplatform_helper encode-Scene > output.bin
//   swift run -c release crossplatform_helper decode-Scene input.bin

import Foundation
import swift

func makeTestPoint() -> Point {
    return Point(
        x: 3.14159,
        y: 3.14159
    )
}

func makeTestRectangle() -> Rectangle {
    return Rectangle(
        topLeft: makeTestPoint(),
        bottomRight: makeTestPoint(),
        color: 4_294_967_295
    )
}

func makeTestScene() -> Scene {
    return Scene(
        name: "Hello from Swift!",
        mainRect: makeTestRectangle(),
        count: 4_294_967_295
    )
}

func encodePoint() throws {
    let data = makeTestPoint()

    let bytes = data.encodeToBytes()
    let binaryData = Data(bytes)

    FileHandle.standardOutput.write(binaryData)
}

func decodePoint(filename: String) throws {
    let fileData = try Data(contentsOf: URL(fileURLWithPath: filename))
    let bytes = [UInt8](fileData)
    let decoded = try Point.decode(from: bytes)

    // Basic validation
    var ok = true
    ok = ok && abs(decoded.x - 3.14159) < 0.0001
    ok = ok && abs(decoded.y - 3.14159) < 0.0001

    if !ok {
        fputs("Validation failed\n", stderr)
        fputs("Decoded: \(decoded)\n", stderr)
        exit(1)
    }

    fputs("✓ Swift successfully decoded and validated\n", stderr)
}

func encodeRectangle() throws {
    let data = makeTestRectangle()

    let bytes = data.encodeToBytes()
    let binaryData = Data(bytes)

    FileHandle.standardOutput.write(binaryData)
}

func decodeRectangle(filename: String) throws {
    let fileData = try Data(contentsOf: URL(fileURLWithPath: filename))
    let bytes = [UInt8](fileData)
    let decoded = try Rectangle.decode(from: bytes)

    // Basic validation
    var ok = true
    ok = ok && decoded.color == 4_294_967_295

    if !ok {
        fputs("Validation failed\n", stderr)
        fputs("Decoded: \(decoded)\n", stderr)
        exit(1)
    }

    fputs("✓ Swift successfully decoded and validated\n", stderr)
}

func encodeScene() throws {
    let data = makeTestScene()

    let bytes = data.encodeToBytes()
    let binaryData = Data(bytes)

    FileHandle.standardOutput.write(binaryData)
}

func decodeScene(filename: String) throws {
    let fileData = try Data(contentsOf: URL(fileURLWithPath: filename))
    let bytes = [UInt8](fileData)
    let decoded = try Scene.decode(from: bytes)

    // Basic validation
    var ok = true
    ok = ok && decoded.count == 4_294_967_295

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
    fputs("  encode-Point - Encode Point to stdout\n", stderr)
    fputs("  decode-Point <file> - Decode Point from file\n", stderr)
    fputs("  encode-Rectangle - Encode Rectangle to stdout\n", stderr)
    fputs("  decode-Rectangle <file> - Decode Rectangle from file\n", stderr)
    fputs("  encode-Scene - Encode Scene to stdout\n", stderr)
    fputs("  decode-Scene <file> - Decode Scene from file\n", stderr)
    exit(1)
}

let command = args[1]

do {
    switch command {
    case "encode-Point":
        try encodePoint()
    case "decode-Point":
        guard args.count >= 3 else {
            fputs("Error: decode-Point requires filename argument\n", stderr)
            exit(1)
        }
        try decodePoint(filename: args[2])
    case "encode-Rectangle":
        try encodeRectangle()
    case "decode-Rectangle":
        guard args.count >= 3 else {
            fputs("Error: decode-Rectangle requires filename argument\n", stderr)
            exit(1)
        }
        try decodeRectangle(filename: args[2])
    case "encode-Scene":
        try encodeScene()
    case "decode-Scene":
        guard args.count >= 3 else {
            fputs("Error: decode-Scene requires filename argument\n", stderr)
            exit(1)
        }
        try decodeScene(filename: args[2])
    default:
        fputs("Unknown command: \(command)\n", stderr)
        exit(1)
    }
} catch {
    fputs("Error: \(error)\n", stderr)
    exit(1)
}
