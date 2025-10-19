// Cross-platform test helper for swift
// Usage:
//   swift run -c release crossplatform_helper encode-AllPrimitives > output.bin
//   swift run -c release crossplatform_helper decode-AllPrimitives input.bin

import Foundation
import swift

func makeTestAllPrimitives() -> AllPrimitives {
    return AllPrimitives(
        u8Field: 255,
        u16Field: 65535,
        u32Field: 4_294_967_295,
        u64Field: 18_446_744_073_709_551_615,
        i8Field: -128,
        i16Field: -32768,
        i32Field: -2_147_483_648,
        i64Field: -9_223_372_036_854_775_808,
        f32Field: 3.14159,
        f64Field: 2.718281828459045,
        boolField: true,
        strField: "Hello from Swift!"
    )
}

func encodeAllPrimitives() throws {
    let data = makeTestAllPrimitives()

    let bytes = data.encodeToBytes()
    let binaryData = Data(bytes)

    FileHandle.standardOutput.write(binaryData)
}

func decodeAllPrimitives(filename: String) throws {
    let fileData = try Data(contentsOf: URL(fileURLWithPath: filename))
    let bytes = [UInt8](fileData)
    let decoded = try AllPrimitives.decode(from: bytes)

    // Basic validation
    var ok = true
    ok = ok && decoded.u8Field == 255
    ok = ok && decoded.u16Field == 65535
    ok = ok && decoded.u32Field == 4_294_967_295
    ok = ok && decoded.u64Field == 18_446_744_073_709_551_615
    ok = ok && decoded.i8Field == -128
    ok = ok && decoded.i16Field == -32768
    ok = ok && decoded.i32Field == -2_147_483_648
    ok = ok && decoded.i64Field == -9_223_372_036_854_775_808
    ok = ok && abs(decoded.f32Field - 3.14159) < 0.0001
    ok = ok && abs(decoded.f64Field - 2.718281828459045) < 0.0000001
    ok = ok && decoded.boolField == true

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
    fputs("  encode-AllPrimitives - Encode AllPrimitives to stdout\n", stderr)
    fputs("  decode-AllPrimitives <file> - Decode AllPrimitives from file\n", stderr)
    exit(1)
}

let command = args[1]

do {
    switch command {
    case "encode-AllPrimitives":
        try encodeAllPrimitives()
    case "decode-AllPrimitives":
        guard args.count >= 3 else {
            fputs("Error: decode-AllPrimitives requires filename argument\n", stderr)
            exit(1)
        }
        try decodeAllPrimitives(filename: args[2])
    default:
        fputs("Unknown command: \(command)\n", stderr)
        exit(1)
    }
} catch {
    fputs("Error: \(error)\n", stderr)
    exit(1)
}
