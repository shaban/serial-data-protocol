// Cross-platform test helper for swift
// Usage:
//   swift run -c release crossplatform_helper encode-ArraysOfPrimitives > output.bin
//   swift run -c release crossplatform_helper decode-ArraysOfPrimitives input.bin
//   swift run -c release crossplatform_helper encode-Item > output.bin
//   swift run -c release crossplatform_helper decode-Item input.bin
//   swift run -c release crossplatform_helper encode-ArraysOfStructs > output.bin
//   swift run -c release crossplatform_helper decode-ArraysOfStructs input.bin

import Foundation
import swift

func makeTestArraysOfPrimitives() -> ArraysOfPrimitives {
    return ArraysOfPrimitives(
        u8Array: ContiguousArray([255, 255, 255]),
        u32Array: ContiguousArray([4_294_967_295, 4_294_967_295, 4_294_967_295]),
        f64Array: ContiguousArray([2.718281828459045, 2.718281828459045, 2.718281828459045]),
        strArray: ContiguousArray(["Hello from Swift!", "Hello from Swift!", "Hello from Swift!"]),
        boolArray: ContiguousArray([true, true, true])
    )
}

func makeTestItem() -> Item {
    return Item(
        id: 4_294_967_295,
        name: "Hello from Swift!"
    )
}

func makeTestArraysOfStructs() -> ArraysOfStructs {
    return ArraysOfStructs(
        items: ContiguousArray([makeTestItem(), makeTestItem(), makeTestItem()]),
        count: 4_294_967_295
    )
}

func encodeArraysOfPrimitives() throws {
    let data = makeTestArraysOfPrimitives()

    let bytes = data.encodeToBytes()
    let binaryData = Data(bytes)

    FileHandle.standardOutput.write(binaryData)
}

func decodeArraysOfPrimitives(filename: String) throws {
    let fileData = try Data(contentsOf: URL(fileURLWithPath: filename))
    let bytes = [UInt8](fileData)
    let decoded = try ArraysOfPrimitives.decode(from: bytes)

    // Basic validation
    var ok = true
    ok = ok && decoded.u8Array.count > 0
    ok = ok && decoded.u32Array.count > 0
    ok = ok && decoded.f64Array.count > 0
    ok = ok && decoded.strArray.count > 0
    ok = ok && decoded.boolArray.count > 0

    if !ok {
        fputs("Validation failed\n", stderr)
        fputs("Decoded: \(decoded)\n", stderr)
        exit(1)
    }

    fputs("✓ Swift successfully decoded and validated\n", stderr)
}

func encodeItem() throws {
    let data = makeTestItem()

    let bytes = data.encodeToBytes()
    let binaryData = Data(bytes)

    FileHandle.standardOutput.write(binaryData)
}

func decodeItem(filename: String) throws {
    let fileData = try Data(contentsOf: URL(fileURLWithPath: filename))
    let bytes = [UInt8](fileData)
    let decoded = try Item.decode(from: bytes)

    // Basic validation
    var ok = true
    ok = ok && decoded.id == 4_294_967_295

    if !ok {
        fputs("Validation failed\n", stderr)
        fputs("Decoded: \(decoded)\n", stderr)
        exit(1)
    }

    fputs("✓ Swift successfully decoded and validated\n", stderr)
}

func encodeArraysOfStructs() throws {
    let data = makeTestArraysOfStructs()

    let bytes = data.encodeToBytes()
    let binaryData = Data(bytes)

    FileHandle.standardOutput.write(binaryData)
}

func decodeArraysOfStructs(filename: String) throws {
    let fileData = try Data(contentsOf: URL(fileURLWithPath: filename))
    let bytes = [UInt8](fileData)
    let decoded = try ArraysOfStructs.decode(from: bytes)

    // Basic validation
    var ok = true
    ok = ok && decoded.items.count > 0
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
    fputs("  encode-ArraysOfPrimitives - Encode ArraysOfPrimitives to stdout\n", stderr)
    fputs("  decode-ArraysOfPrimitives <file> - Decode ArraysOfPrimitives from file\n", stderr)
    fputs("  encode-Item - Encode Item to stdout\n", stderr)
    fputs("  decode-Item <file> - Decode Item from file\n", stderr)
    fputs("  encode-ArraysOfStructs - Encode ArraysOfStructs to stdout\n", stderr)
    fputs("  decode-ArraysOfStructs <file> - Decode ArraysOfStructs from file\n", stderr)
    exit(1)
}

let command = args[1]

do {
    switch command {
    case "encode-ArraysOfPrimitives":
        try encodeArraysOfPrimitives()
    case "decode-ArraysOfPrimitives":
        guard args.count >= 3 else {
            fputs("Error: decode-ArraysOfPrimitives requires filename argument\n", stderr)
            exit(1)
        }
        try decodeArraysOfPrimitives(filename: args[2])
    case "encode-Item":
        try encodeItem()
    case "decode-Item":
        guard args.count >= 3 else {
            fputs("Error: decode-Item requires filename argument\n", stderr)
            exit(1)
        }
        try decodeItem(filename: args[2])
    case "encode-ArraysOfStructs":
        try encodeArraysOfStructs()
    case "decode-ArraysOfStructs":
        guard args.count >= 3 else {
            fputs("Error: decode-ArraysOfStructs requires filename argument\n", stderr)
            exit(1)
        }
        try decodeArraysOfStructs(filename: args[2])
    default:
        fputs("Unknown command: \(command)\n", stderr)
        exit(1)
    }
} catch {
    fputs("Error: \(error)\n", stderr)
    exit(1)
}
