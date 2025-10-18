#!/usr/bin/env swift

// Swift benchmark helper for cross-language performance testing
// Usage:
//   swift_bench encode-primitives <iterations>
//   swift_bench decode-primitives <file> <iterations>

import Foundation

// ============================================================================
// Generated AllPrimitives struct (optimized [UInt8] version)
// ============================================================================

public struct AllPrimitives {
    public var u8Field: UInt8
    public var u16Field: UInt16
    public var u32Field: UInt32
    public var u64Field: UInt64
    public var i8Field: Int8
    public var i16Field: Int16
    public var i32Field: Int32
    public var i64Field: Int64
    public var f32Field: Float
    public var f64Field: Double
    public var boolField: Bool
    public var strField: String
    
    public init(u8Field: UInt8, u16Field: UInt16, u32Field: UInt32, u64Field: UInt64, i8Field: Int8, i16Field: Int16, i32Field: Int32, i64Field: Int64, f32Field: Float, f64Field: Double, boolField: Bool, strField: String) {
        self.u8Field = u8Field
        self.u16Field = u16Field
        self.u32Field = u32Field
        self.u64Field = u64Field
        self.i8Field = i8Field
        self.i16Field = i16Field
        self.i32Field = i32Field
        self.i64Field = i64Field
        self.f32Field = f32Field
        self.f64Field = f64Field
        self.boolField = boolField
        self.strField = strField
    }
}

public enum SDPDecodeError: Error {
    case insufficientData
    case invalidUTF8
    case invalidBoolValue
}

extension AllPrimitives {
    /// Encode to byte array
    public func encodeToBytes() -> [UInt8] {
        var bytes = [UInt8]()
        bytes.reserveCapacity(encodedSize())
        encode(to: &bytes)
        return bytes
    }

    /// Encode to a byte array buffer
    public func encode(to bytes: inout [UInt8]) {
        bytes.append(self.u8Field)
        withUnsafeBytes(of: self.u16Field.littleEndian) { bytes.append(contentsOf: $0) }
        withUnsafeBytes(of: self.u32Field.littleEndian) { bytes.append(contentsOf: $0) }
        withUnsafeBytes(of: self.u64Field.littleEndian) { bytes.append(contentsOf: $0) }
        bytes.append(UInt8(bitPattern: self.i8Field))
        withUnsafeBytes(of: self.i16Field.littleEndian) { bytes.append(contentsOf: $0) }
        withUnsafeBytes(of: self.i32Field.littleEndian) { bytes.append(contentsOf: $0) }
        withUnsafeBytes(of: self.i64Field.littleEndian) { bytes.append(contentsOf: $0) }
        let f32FieldBits = self.f32Field.bitPattern.littleEndian
        withUnsafeBytes(of: f32FieldBits) { bytes.append(contentsOf: $0) }
        let f64FieldBits = self.f64Field.bitPattern.littleEndian
        withUnsafeBytes(of: f64FieldBits) { bytes.append(contentsOf: $0) }
        bytes.append(self.boolField ? 1 : 0)
        let strFieldBytes = Array(self.strField.utf8)
        let strFieldLen = UInt32(strFieldBytes.count).littleEndian
        withUnsafeBytes(of: strFieldLen) { bytes.append(contentsOf: $0) }
        bytes.append(contentsOf: strFieldBytes)
    }

    /// Calculate the encoded size in bytes
    public func encodedSize() -> Int {
        var size = 0
        size += 1
        size += 2
        size += 4
        size += 8
        size += 1
        size += 2
        size += 4
        size += 8
        size += 4
        size += 8
        size += 1
        size += 4 + self.strField.utf8.count
        return size
    }
}

extension AllPrimitives {
    /// Decode from byte array
    public static func decode(from bytes: [UInt8]) throws -> Self {
        var offset = 0

        guard offset < bytes.count else { throw SDPDecodeError.insufficientData }
        let u8Field = bytes[offset]
        offset += 1
        guard offset + 2 <= bytes.count else { throw SDPDecodeError.insufficientData }
        let u16Field = bytes[offset..<offset+2].withUnsafeBytes { UInt16(littleEndian: $0.load(as: UInt16.self)) }
        offset += 2
        guard offset + 4 <= bytes.count else { throw SDPDecodeError.insufficientData }
        let u32Field = bytes[offset..<offset+4].withUnsafeBytes { UInt32(littleEndian: $0.load(as: UInt32.self)) }
        offset += 4
        guard offset + 8 <= bytes.count else { throw SDPDecodeError.insufficientData }
        let u64Field = bytes[offset..<offset+8].withUnsafeBytes { UInt64(littleEndian: $0.load(as: UInt64.self)) }
        offset += 8
        guard offset < bytes.count else { throw SDPDecodeError.insufficientData }
        let i8Field = Int8(bitPattern: bytes[offset])
        offset += 1
        guard offset + 2 <= bytes.count else { throw SDPDecodeError.insufficientData }
        let i16Field = bytes[offset..<offset+2].withUnsafeBytes { Int16(bitPattern: UInt16(littleEndian: $0.load(as: UInt16.self))) }
        offset += 2
        guard offset + 4 <= bytes.count else { throw SDPDecodeError.insufficientData }
        let i32Field = bytes[offset..<offset+4].withUnsafeBytes { Int32(bitPattern: UInt32(littleEndian: $0.load(as: UInt32.self))) }
        offset += 4
        guard offset + 8 <= bytes.count else { throw SDPDecodeError.insufficientData }
        let i64Field = bytes[offset..<offset+8].withUnsafeBytes { Int64(bitPattern: UInt64(littleEndian: $0.load(as: UInt64.self))) }
        offset += 8
        guard offset + 4 <= bytes.count else { throw SDPDecodeError.insufficientData }
        let f32FieldBits = bytes[offset..<offset+4].withUnsafeBytes { UInt32(littleEndian: $0.load(as: UInt32.self)) }
        let f32Field = Float(bitPattern: f32FieldBits)
        offset += 4
        guard offset + 8 <= bytes.count else { throw SDPDecodeError.insufficientData }
        let f64FieldBits = bytes[offset..<offset+8].withUnsafeBytes { UInt64(littleEndian: $0.load(as: UInt64.self)) }
        let f64Field = Double(bitPattern: f64FieldBits)
        offset += 8
        guard offset < bytes.count else { throw SDPDecodeError.insufficientData }
        let boolFieldByte = bytes[offset]
        guard boolFieldByte == 0 || boolFieldByte == 1 else { throw SDPDecodeError.invalidBoolValue }
        let boolField = boolFieldByte == 1
        offset += 1
        guard offset + 4 <= bytes.count else { throw SDPDecodeError.insufficientData }
        let strFieldLen = Int(bytes[offset..<offset+4].withUnsafeBytes { UInt32(littleEndian: $0.load(as: UInt32.self)) })
        offset += 4
        guard offset + strFieldLen <= bytes.count else { throw SDPDecodeError.insufficientData }
        let strFieldBytes = Array(bytes[offset..<offset+strFieldLen])
        guard let strField = String(bytes: strFieldBytes, encoding: .utf8) else {
            throw SDPDecodeError.invalidUTF8
        }
        offset += strFieldLen

        return Self(u8Field: u8Field, u16Field: u16Field, u32Field: u32Field, u64Field: u64Field, i8Field: i8Field, i16Field: i16Field, i32Field: i32Field, i64Field: i64Field, f32Field: f32Field, f64Field: f64Field, boolField: boolField, strField: strField)
    }
}

// ============================================================================
// Benchmark Commands
// ============================================================================

enum Command: String {
    case encodePrimitives = "encode-primitives"
    case decodePrimitives = "decode-primitives"
}

func benchmarkEncodePrimitives(iterations: Int) {
    let data = AllPrimitives(
        u8Field: 255,
        u16Field: 65535,
        u32Field: 4294967295,
        u64Field: 18446744073709551615,
        i8Field: -128,
        i16Field: -32768,
        i32Field: -2147483648,
        i64Field: -9223372036854775808,
        f32Field: 3.14159,
        f64Field: 2.71828,
        boolField: true,
        strField: "Hello from Swift!"
    )
    
    // Warmup
    for _ in 0..<1000 {
        _ = data.encodeToBytes()
    }
    
    let start = DispatchTime.now()
    for _ in 0..<iterations {
        _ = data.encodeToBytes()
    }
    let end = DispatchTime.now()
    
    let nanos = end.uptimeNanoseconds - start.uptimeNanoseconds
    let nsPerOp = nanos / UInt64(iterations)
    
    print(nsPerOp)
}

func benchmarkDecodePrimitives(filename: String, iterations: Int) {
    // Read file
    guard let data = try? Data(contentsOf: URL(fileURLWithPath: filename)) else {
        fputs("Error: Cannot read file \(filename)\n", stderr)
        exit(1)
    }
    let bytes = [UInt8](data)
    
    // Warmup
    for _ in 0..<1000 {
        _ = try? AllPrimitives.decode(from: bytes)
    }
    
    let start = DispatchTime.now()
    for _ in 0..<iterations {
        _ = try? AllPrimitives.decode(from: bytes)
    }
    let end = DispatchTime.now()
    
    let nanos = end.uptimeNanoseconds - start.uptimeNanoseconds
    let nsPerOp = nanos / UInt64(iterations)
    
    print(nsPerOp)
}

// ============================================================================
// Main entry point
// ============================================================================

guard CommandLine.arguments.count >= 2 else {
    fputs("Usage: swift_bench <command> [args]\n", stderr)
    fputs("Commands:\n", stderr)
    fputs("  encode-primitives <iterations>\n", stderr)
    fputs("  decode-primitives <file> <iterations>\n", stderr)
    exit(1)
}

let commandString = CommandLine.arguments[1]

guard let command = Command(rawValue: commandString) else {
    fputs("Unknown command: \(commandString)\n", stderr)
    exit(1)
}

switch command {
case .encodePrimitives:
    guard CommandLine.arguments.count >= 3 else {
        fputs("Error: encode-primitives requires iterations argument\n", stderr)
        exit(1)
    }
    guard let iterations = Int(CommandLine.arguments[2]) else {
        fputs("Error: iterations must be an integer\n", stderr)
        exit(1)
    }
    benchmarkEncodePrimitives(iterations: iterations)
    
case .decodePrimitives:
    guard CommandLine.arguments.count >= 4 else {
        fputs("Error: decode-primitives requires filename and iterations arguments\n", stderr)
        exit(1)
    }
    let filename = CommandLine.arguments[2]
    guard let iterations = Int(CommandLine.arguments[3]) else {
        fputs("Error: iterations must be an integer\n", stderr)
        exit(1)
    }
    benchmarkDecodePrimitives(filename: filename, iterations: iterations)
}
