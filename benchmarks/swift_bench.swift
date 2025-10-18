#!/usr/bin/env swift

// Swift benchmark helper for cross-language performance testing
// Usage:
//   swift_bench encode-primitives <iterations>
//   swift_bench decode-primitives <file> <iterations>

import Foundation

// ============================================================================
// Generated AllPrimitives struct (from testdata/primitives/swift)
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
    public func encodeToData() throws -> Data {
        var data = Data()
        data.reserveCapacity(self.encodedSize())
        try self.encode(to: &data)
        return data
    }
    
    public func encode(to data: inout Data) throws {
        data.append(self.u8Field)
        withUnsafeBytes(of: self.u16Field.littleEndian) { data.append(contentsOf: $0) }
        withUnsafeBytes(of: self.u32Field.littleEndian) { data.append(contentsOf: $0) }
        withUnsafeBytes(of: self.u64Field.littleEndian) { data.append(contentsOf: $0) }
        data.append(UInt8(bitPattern: self.i8Field))
        withUnsafeBytes(of: self.i16Field.littleEndian) { data.append(contentsOf: $0) }
        withUnsafeBytes(of: self.i32Field.littleEndian) { data.append(contentsOf: $0) }
        withUnsafeBytes(of: self.i64Field.littleEndian) { data.append(contentsOf: $0) }
        let f32Bits = self.f32Field.bitPattern.littleEndian
        withUnsafeBytes(of: f32Bits) { data.append(contentsOf: $0) }
        let f64Bits = self.f64Field.bitPattern.littleEndian
        withUnsafeBytes(of: f64Bits) { data.append(contentsOf: $0) }
        data.append(self.boolField ? 1 : 0)
        let strData = self.strField.data(using: .utf8)!
        let strLen = UInt32(strData.count).littleEndian
        withUnsafeBytes(of: strLen) { data.append(contentsOf: $0) }
        data.append(strData)
    }
    
    public func encodedSize() -> Int {
        var size = 0
        size += 1  // u8Field
        size += 2  // u16Field
        size += 4  // u32Field
        size += 8  // u64Field
        size += 1  // i8Field
        size += 2  // i16Field
        size += 4  // i32Field
        size += 8  // i64Field
        size += 4  // f32Field
        size += 8  // f64Field
        size += 1  // boolField
        size += 4  // strField length
        size += self.strField.utf8.count  // strField data
        return size
    }
    
    public static func decode(from data: Data) throws -> Self {
        var offset = 0
        
        // u8Field: UInt8
        guard offset + 1 <= data.count else { throw SDPDecodeError.insufficientData }
        let u8Field = data[offset]
        offset += 1
        
        // u16Field: UInt16
        guard offset + 2 <= data.count else { throw SDPDecodeError.insufficientData }
        var u16Bytes = [UInt8](repeating: 0, count: 2)
        data.copyBytes(to: &u16Bytes, from: offset..<offset+2)
        let u16Field = UInt16(littleEndian: u16Bytes.withUnsafeBytes { $0.load(as: UInt16.self) })
        offset += 2
        
        // u32Field: UInt32
        guard offset + 4 <= data.count else { throw SDPDecodeError.insufficientData }
        var u32Bytes = [UInt8](repeating: 0, count: 4)
        data.copyBytes(to: &u32Bytes, from: offset..<offset+4)
        let u32Field = UInt32(littleEndian: u32Bytes.withUnsafeBytes { $0.load(as: UInt32.self) })
        offset += 4
        
        // u64Field: UInt64
        guard offset + 8 <= data.count else { throw SDPDecodeError.insufficientData }
        var u64Bytes = [UInt8](repeating: 0, count: 8)
        data.copyBytes(to: &u64Bytes, from: offset..<offset+8)
        let u64Field = UInt64(littleEndian: u64Bytes.withUnsafeBytes { $0.load(as: UInt64.self) })
        offset += 8
        
        // i8Field: Int8
        guard offset + 1 <= data.count else { throw SDPDecodeError.insufficientData }
        let i8Field = Int8(bitPattern: data[offset])
        offset += 1
        
        // i16Field: Int16
        guard offset + 2 <= data.count else { throw SDPDecodeError.insufficientData }
        var i16Bytes = [UInt8](repeating: 0, count: 2)
        data.copyBytes(to: &i16Bytes, from: offset..<offset+2)
        let i16Field = Int16(bitPattern: UInt16(littleEndian: i16Bytes.withUnsafeBytes { $0.load(as: UInt16.self) }))
        offset += 2
        
        // i32Field: Int32
        guard offset + 4 <= data.count else { throw SDPDecodeError.insufficientData }
        var i32Bytes = [UInt8](repeating: 0, count: 4)
        data.copyBytes(to: &i32Bytes, from: offset..<offset+4)
        let i32Field = Int32(bitPattern: UInt32(littleEndian: i32Bytes.withUnsafeBytes { $0.load(as: UInt32.self) }))
        offset += 4
        
        // i64Field: Int64
        guard offset + 8 <= data.count else { throw SDPDecodeError.insufficientData }
        var i64Bytes = [UInt8](repeating: 0, count: 8)
        data.copyBytes(to: &i64Bytes, from: offset..<offset+8)
        let i64Field = Int64(bitPattern: UInt64(littleEndian: i64Bytes.withUnsafeBytes { $0.load(as: UInt64.self) }))
        offset += 8
        
        // f32Field: Float
        guard offset + 4 <= data.count else { throw SDPDecodeError.insufficientData }
        var f32BitsBytes = [UInt8](repeating: 0, count: 4)
        data.copyBytes(to: &f32BitsBytes, from: offset..<offset+4)
        let f32Bits = UInt32(littleEndian: f32BitsBytes.withUnsafeBytes { $0.load(as: UInt32.self) })
        let f32Field = Float(bitPattern: f32Bits)
        offset += 4
        
        // f64Field: Double
        guard offset + 8 <= data.count else { throw SDPDecodeError.insufficientData }
        var f64BitsBytes = [UInt8](repeating: 0, count: 8)
        data.copyBytes(to: &f64BitsBytes, from: offset..<offset+8)
        let f64Bits = UInt64(littleEndian: f64BitsBytes.withUnsafeBytes { $0.load(as: UInt64.self) })
        let f64Field = Double(bitPattern: f64Bits)
        offset += 8
        
        // boolField: Bool
        guard offset + 1 <= data.count else { throw SDPDecodeError.insufficientData }
        let boolByte = data[offset]
        guard boolByte == 0 || boolByte == 1 else { throw SDPDecodeError.invalidBoolValue }
        let boolField = boolByte != 0
        offset += 1
        
        // strField: String
        guard offset + 4 <= data.count else { throw SDPDecodeError.insufficientData }
        var strLenBytes = [UInt8](repeating: 0, count: 4)
        data.copyBytes(to: &strLenBytes, from: offset..<offset+4)
        let strLen = Int(UInt32(littleEndian: strLenBytes.withUnsafeBytes { $0.load(as: UInt32.self) }))
        offset += 4
        guard offset + strLen <= data.count else { throw SDPDecodeError.insufficientData }
        let strData = data.subdata(in: offset..<offset+strLen)
        guard let strField = String(data: strData, encoding: .utf8) else { throw SDPDecodeError.invalidUTF8 }
        offset += strLen
        
        return Self(u8Field: u8Field, u16Field: u16Field, u32Field: u32Field, u64Field: u64Field, i8Field: i8Field, i16Field: i16Field, i32Field: i32Field, i64Field: i64Field, f32Field: f32Field, f64Field: f64Field, boolField: boolField, strField: strField)
    }
}

// ============================================================================
// Benchmark commands
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
        f64Field: 2.718281828459045,
        boolField: true,
        strField: "Hello from Swift!"
    )
    
    // Warmup
    for _ in 0..<100 {
        _ = try! data.encodeToData()
    }
    
    // Benchmark
    let start = DispatchTime.now()
    for _ in 0..<iterations {
        _ = try! data.encodeToData()
    }
    let end = DispatchTime.now()
    
    let nanos = end.uptimeNanoseconds - start.uptimeNanoseconds
    let nsPerOp = nanos / UInt64(iterations)
    
    print(nsPerOp)
}

func benchmarkDecodePrimitives(filename: String, iterations: Int) {
    let fileData = try! Data(contentsOf: URL(fileURLWithPath: filename))
    
    // Warmup
    for _ in 0..<100 {
        _ = try! AllPrimitives.decode(from: fileData)
    }
    
    // Benchmark
    let start = DispatchTime.now()
    for _ in 0..<iterations {
        _ = try! AllPrimitives.decode(from: fileData)
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
