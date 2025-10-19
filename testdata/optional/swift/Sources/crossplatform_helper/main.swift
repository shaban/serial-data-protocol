// Cross-platform test helper for swift
// Usage:
//   swift run -c release crossplatform_helper encode-Request > output.bin
//   swift run -c release crossplatform_helper decode-Request input.bin
//   swift run -c release crossplatform_helper encode-Metadata > output.bin
//   swift run -c release crossplatform_helper decode-Metadata input.bin
//   swift run -c release crossplatform_helper encode-Config > output.bin
//   swift run -c release crossplatform_helper decode-Config input.bin
//   swift run -c release crossplatform_helper encode-DatabaseConfig > output.bin
//   swift run -c release crossplatform_helper decode-DatabaseConfig input.bin
//   swift run -c release crossplatform_helper encode-CacheConfig > output.bin
//   swift run -c release crossplatform_helper decode-CacheConfig input.bin
//   swift run -c release crossplatform_helper encode-Document > output.bin
//   swift run -c release crossplatform_helper decode-Document input.bin
//   swift run -c release crossplatform_helper encode-TagList > output.bin
//   swift run -c release crossplatform_helper decode-TagList input.bin

import Foundation
import swift

func makeTestRequest() -> Request {
    return Request(
        id: 4_294_967_295,
        metadata: .some(makeTestMetadata())
    )
}

func makeTestMetadata() -> Metadata {
    return Metadata(
        userId: 18_446_744_073_709_551_615,
        username: "Hello from Swift!"
    )
}

func makeTestConfig() -> Config {
    return Config(
        name: "Hello from Swift!",
        database: .some(makeTestDatabaseConfig()),
        cache: .some(makeTestCacheConfig())
    )
}

func makeTestDatabaseConfig() -> DatabaseConfig {
    return DatabaseConfig(
        host: "Hello from Swift!",
        port: 65535
    )
}

func makeTestCacheConfig() -> CacheConfig {
    return CacheConfig(
        sizeMb: 4_294_967_295,
        ttlSeconds: 4_294_967_295
    )
}

func makeTestDocument() -> Document {
    return Document(
        id: 4_294_967_295,
        tags: .some(makeTestTagList())
    )
}

func makeTestTagList() -> TagList {
    return TagList(
        items: ContiguousArray(["Hello from Swift!", "Hello from Swift!", "Hello from Swift!"])
    )
}

func encodeRequest() throws {
    let data = makeTestRequest()

    let bytes = data.encodeToBytes()
    let binaryData = Data(bytes)

    FileHandle.standardOutput.write(binaryData)
}

func decodeRequest(filename: String) throws {
    let fileData = try Data(contentsOf: URL(fileURLWithPath: filename))
    let bytes = [UInt8](fileData)
    let decoded = try Request.decode(from: bytes)

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

func encodeMetadata() throws {
    let data = makeTestMetadata()

    let bytes = data.encodeToBytes()
    let binaryData = Data(bytes)

    FileHandle.standardOutput.write(binaryData)
}

func decodeMetadata(filename: String) throws {
    let fileData = try Data(contentsOf: URL(fileURLWithPath: filename))
    let bytes = [UInt8](fileData)
    let decoded = try Metadata.decode(from: bytes)

    // Basic validation
    var ok = true
    ok = ok && decoded.userId == 18_446_744_073_709_551_615

    if !ok {
        fputs("Validation failed\n", stderr)
        fputs("Decoded: \(decoded)\n", stderr)
        exit(1)
    }

    fputs("✓ Swift successfully decoded and validated\n", stderr)
}

func encodeConfig() throws {
    let data = makeTestConfig()

    let bytes = data.encodeToBytes()
    let binaryData = Data(bytes)

    FileHandle.standardOutput.write(binaryData)
}

func decodeConfig(filename: String) throws {
    let fileData = try Data(contentsOf: URL(fileURLWithPath: filename))
    let bytes = [UInt8](fileData)
    let decoded = try Config.decode(from: bytes)

    // Basic validation
    var ok = true

    if !ok {
        fputs("Validation failed\n", stderr)
        fputs("Decoded: \(decoded)\n", stderr)
        exit(1)
    }

    fputs("✓ Swift successfully decoded and validated\n", stderr)
}

func encodeDatabaseConfig() throws {
    let data = makeTestDatabaseConfig()

    let bytes = data.encodeToBytes()
    let binaryData = Data(bytes)

    FileHandle.standardOutput.write(binaryData)
}

func decodeDatabaseConfig(filename: String) throws {
    let fileData = try Data(contentsOf: URL(fileURLWithPath: filename))
    let bytes = [UInt8](fileData)
    let decoded = try DatabaseConfig.decode(from: bytes)

    // Basic validation
    var ok = true
    ok = ok && decoded.port == 65535

    if !ok {
        fputs("Validation failed\n", stderr)
        fputs("Decoded: \(decoded)\n", stderr)
        exit(1)
    }

    fputs("✓ Swift successfully decoded and validated\n", stderr)
}

func encodeCacheConfig() throws {
    let data = makeTestCacheConfig()

    let bytes = data.encodeToBytes()
    let binaryData = Data(bytes)

    FileHandle.standardOutput.write(binaryData)
}

func decodeCacheConfig(filename: String) throws {
    let fileData = try Data(contentsOf: URL(fileURLWithPath: filename))
    let bytes = [UInt8](fileData)
    let decoded = try CacheConfig.decode(from: bytes)

    // Basic validation
    var ok = true
    ok = ok && decoded.sizeMb == 4_294_967_295
    ok = ok && decoded.ttlSeconds == 4_294_967_295

    if !ok {
        fputs("Validation failed\n", stderr)
        fputs("Decoded: \(decoded)\n", stderr)
        exit(1)
    }

    fputs("✓ Swift successfully decoded and validated\n", stderr)
}

func encodeDocument() throws {
    let data = makeTestDocument()

    let bytes = data.encodeToBytes()
    let binaryData = Data(bytes)

    FileHandle.standardOutput.write(binaryData)
}

func decodeDocument(filename: String) throws {
    let fileData = try Data(contentsOf: URL(fileURLWithPath: filename))
    let bytes = [UInt8](fileData)
    let decoded = try Document.decode(from: bytes)

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

func encodeTagList() throws {
    let data = makeTestTagList()

    let bytes = data.encodeToBytes()
    let binaryData = Data(bytes)

    FileHandle.standardOutput.write(binaryData)
}

func decodeTagList(filename: String) throws {
    let fileData = try Data(contentsOf: URL(fileURLWithPath: filename))
    let bytes = [UInt8](fileData)
    let decoded = try TagList.decode(from: bytes)

    // Basic validation
    var ok = true
    ok = ok && decoded.items.count > 0

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
    fputs("  encode-Request - Encode Request to stdout\n", stderr)
    fputs("  decode-Request <file> - Decode Request from file\n", stderr)
    fputs("  encode-Metadata - Encode Metadata to stdout\n", stderr)
    fputs("  decode-Metadata <file> - Decode Metadata from file\n", stderr)
    fputs("  encode-Config - Encode Config to stdout\n", stderr)
    fputs("  decode-Config <file> - Decode Config from file\n", stderr)
    fputs("  encode-DatabaseConfig - Encode DatabaseConfig to stdout\n", stderr)
    fputs("  decode-DatabaseConfig <file> - Decode DatabaseConfig from file\n", stderr)
    fputs("  encode-CacheConfig - Encode CacheConfig to stdout\n", stderr)
    fputs("  decode-CacheConfig <file> - Decode CacheConfig from file\n", stderr)
    fputs("  encode-Document - Encode Document to stdout\n", stderr)
    fputs("  decode-Document <file> - Decode Document from file\n", stderr)
    fputs("  encode-TagList - Encode TagList to stdout\n", stderr)
    fputs("  decode-TagList <file> - Decode TagList from file\n", stderr)
    exit(1)
}

let command = args[1]

do {
    switch command {
    case "encode-Request":
        try encodeRequest()
    case "decode-Request":
        guard args.count >= 3 else {
            fputs("Error: decode-Request requires filename argument\n", stderr)
            exit(1)
        }
        try decodeRequest(filename: args[2])
    case "encode-Metadata":
        try encodeMetadata()
    case "decode-Metadata":
        guard args.count >= 3 else {
            fputs("Error: decode-Metadata requires filename argument\n", stderr)
            exit(1)
        }
        try decodeMetadata(filename: args[2])
    case "encode-Config":
        try encodeConfig()
    case "decode-Config":
        guard args.count >= 3 else {
            fputs("Error: decode-Config requires filename argument\n", stderr)
            exit(1)
        }
        try decodeConfig(filename: args[2])
    case "encode-DatabaseConfig":
        try encodeDatabaseConfig()
    case "decode-DatabaseConfig":
        guard args.count >= 3 else {
            fputs("Error: decode-DatabaseConfig requires filename argument\n", stderr)
            exit(1)
        }
        try decodeDatabaseConfig(filename: args[2])
    case "encode-CacheConfig":
        try encodeCacheConfig()
    case "decode-CacheConfig":
        guard args.count >= 3 else {
            fputs("Error: decode-CacheConfig requires filename argument\n", stderr)
            exit(1)
        }
        try decodeCacheConfig(filename: args[2])
    case "encode-Document":
        try encodeDocument()
    case "decode-Document":
        guard args.count >= 3 else {
            fputs("Error: decode-Document requires filename argument\n", stderr)
            exit(1)
        }
        try decodeDocument(filename: args[2])
    case "encode-TagList":
        try encodeTagList()
    case "decode-TagList":
        guard args.count >= 3 else {
            fputs("Error: decode-TagList requires filename argument\n", stderr)
            exit(1)
        }
        try decodeTagList(filename: args[2])
    default:
        fputs("Unknown command: \(command)\n", stderr)
        exit(1)
    }
} catch {
    fputs("Error: \(error)\n", stderr)
    exit(1)
}
