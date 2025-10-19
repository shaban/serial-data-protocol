import Foundation

// MARK: - Benchmark Utilities

struct BenchmarkResult {
    let name: String
    let iterations: Int
    let totalTime: TimeInterval
    let avgTime: TimeInterval
    let opsPerSecond: Double
    
    func print() {
        Swift.print(String(format: "%-60s %10d iters  %8.2f ns/op  %12.0f ops/sec",
                          name, iterations, avgTime * 1_000_000_000, opsPerSecond))
    }
}

func benchmark(name: String, iterations: Int = 1_000_000, _ block: () -> Void) -> BenchmarkResult {
    // Warmup
    for _ in 0..<min(iterations / 10, 1000) {
        block()
    }
    
    let start = CFAbsoluteTimeGetCurrent()
    for _ in 0..<iterations {
        block()
    }
    let end = CFAbsoluteTimeGetCurrent()
    
    let totalTime = end - start
    let avgTime = totalTime / Double(iterations)
    let opsPerSecond = Double(iterations) / totalTime
    
    return BenchmarkResult(
        name: name,
        iterations: iterations,
        totalTime: totalTime,
        avgTime: avgTime,
        opsPerSecond: opsPerSecond
    )
}

// MARK: - Test Message

struct Message {
    var id: UInt32
    var items: [UInt32]
}

// MARK: - Encoding Approaches

extension Message {
    // Approach 1: Current implementation (element-by-element with withUnsafeBytes)
    func encodeNaive() -> [UInt8] {
        var bytes = [UInt8]()
        bytes.reserveCapacity(4 + 4 + items.count * 4)
        
        // Encode id
        withUnsafeBytes(of: id.littleEndian) { bytes.append(contentsOf: $0) }
        
        // Encode array length
        let len = UInt32(items.count).littleEndian
        withUnsafeBytes(of: len) { bytes.append(contentsOf: $0) }
        
        // Encode array items (element by element)
        for item in items {
            withUnsafeBytes(of: item.littleEndian) { bytes.append(contentsOf: $0) }
        }
        
        return bytes
    }
    
    // Approach 2: Bulk copy with withUnsafeBufferPointer
    func encodeBulkCopy() -> [UInt8] {
        var bytes = [UInt8]()
        bytes.reserveCapacity(4 + 4 + items.count * 4)
        
        // Encode id
        withUnsafeBytes(of: id.littleEndian) { bytes.append(contentsOf: $0) }
        
        // Encode array length
        let len = UInt32(items.count).littleEndian
        withUnsafeBytes(of: len) { bytes.append(contentsOf: $0) }
        
        // Encode array items in bulk
        items.withUnsafeBufferPointer { buffer in
            for item in buffer {
                withUnsafeBytes(of: item.littleEndian) { bytes.append(contentsOf: $0) }
            }
        }
        
        return bytes
    }
    
    // Approach 3: Direct byte manipulation with pre-allocated buffer
    func encodePreallocated() -> [UInt8] {
        let capacity = 4 + 4 + items.count * 4
        var bytes = [UInt8](repeating: 0, count: capacity)
        
        var offset = 0
        
        // Encode id
        withUnsafeBytes(of: id.littleEndian) { src in
            bytes.withUnsafeMutableBytes { dst in
                dst.baseAddress!.advanced(by: offset).copyMemory(from: src.baseAddress!, byteCount: 4)
            }
        }
        offset += 4
        
        // Encode array length
        let len = UInt32(items.count).littleEndian
        withUnsafeBytes(of: len) { src in
            bytes.withUnsafeMutableBytes { dst in
                dst.baseAddress!.advanced(by: offset).copyMemory(from: src.baseAddress!, byteCount: 4)
            }
        }
        offset += 4
        
        // Encode array items
        items.withUnsafeBufferPointer { buffer in
            for item in buffer {
                withUnsafeBytes(of: item.littleEndian) { src in
                    bytes.withUnsafeMutableBytes { dst in
                        dst.baseAddress!.advanced(by: offset).copyMemory(from: src.baseAddress!, byteCount: 4)
                    }
                }
                offset += 4
            }
        }
        
        return bytes
    }
    
    // Approach 4: Maximum unsafe optimization
    func encodeUnsafeOptimized() -> [UInt8] {
        let capacity = 4 + 4 + items.count * 4
        return [UInt8](unsafeUninitializedCapacity: capacity) { buffer, count in
            var offset = 0
            
            // Encode id
            let idLE = id.littleEndian
            withUnsafeBytes(of: idLE) { src in
                let dst = UnsafeMutableRawPointer(buffer.baseAddress!).advanced(by: offset)
                dst.copyMemory(from: src.baseAddress!, byteCount: 4)
            }
            offset += 4
            
            // Encode array length
            let len = UInt32(items.count).littleEndian
            withUnsafeBytes(of: len) { src in
                let dst = UnsafeMutableRawPointer(buffer.baseAddress!).advanced(by: offset)
                dst.copyMemory(from: src.baseAddress!, byteCount: 4)
            }
            offset += 4
            
            // Encode array items
            items.withUnsafeBufferPointer { itemBuffer in
                for item in itemBuffer {
                    let itemLE = item.littleEndian
                    withUnsafeBytes(of: itemLE) { src in
                        let dst = UnsafeMutableRawPointer(buffer.baseAddress!).advanced(by: offset)
                        dst.copyMemory(from: src.baseAddress!, byteCount: 4)
                    }
                    offset += 4
                }
            }
            
            count = capacity
        }
    }
    
    // Approach 5: Bulk memory copy (if items already little-endian)
    func encodeMemcpy() -> [UInt8] {
        let capacity = 4 + 4 + items.count * 4
        return [UInt8](unsafeUninitializedCapacity: capacity) { buffer, count in
            var offset = 0
            
            // Encode id
            let idLE = id.littleEndian
            withUnsafeBytes(of: idLE) { src in
                let dst = UnsafeMutableRawPointer(buffer.baseAddress!).advanced(by: offset)
                dst.copyMemory(from: src.baseAddress!, byteCount: 4)
            }
            offset += 4
            
            // Encode array length
            let len = UInt32(items.count).littleEndian
            withUnsafeBytes(of: len) { src in
                let dst = UnsafeMutableRawPointer(buffer.baseAddress!).advanced(by: offset)
                dst.copyMemory(from: src.baseAddress!, byteCount: 4)
            }
            offset += 4
            
            // BULK COPY: Copy entire array at once (requires converting to little-endian first)
            let leItems = items.map { $0.littleEndian }
            leItems.withUnsafeBytes { src in
                let dst = UnsafeMutableRawPointer(buffer.baseAddress!).advanced(by: offset)
                dst.copyMemory(from: src.baseAddress!, byteCount: items.count * 4)
            }
            
            count = capacity
        }
    }
}

// MARK: - Decoding Approaches

extension Message {
    // Approach 1: Current implementation (slice-by-slice with withUnsafeBytes)
    static func decodeNaive(from bytes: [UInt8]) -> Message? {
        guard bytes.count >= 8 else { return nil }
        
        var offset = 0
        
        // Decode id
        let id = bytes[offset..<offset+4].withUnsafeBytes {
            UInt32(littleEndian: $0.load(as: UInt32.self))
        }
        offset += 4
        
        // Decode array length
        let len = Int(bytes[offset..<offset+4].withUnsafeBytes {
            UInt32(littleEndian: $0.load(as: UInt32.self))
        })
        offset += 4
        
        guard bytes.count >= offset + len * 4 else { return nil }
        
        // Decode array items
        var items = [UInt32]()
        items.reserveCapacity(len)
        for _ in 0..<len {
            let item = bytes[offset..<offset+4].withUnsafeBytes {
                UInt32(littleEndian: $0.load(as: UInt32.self))
            }
            items.append(item)
            offset += 4
        }
        
        return Message(id: id, items: items)
    }
    
    // Approach 2: Use withUnsafeBufferPointer on byte array
    static func decodeBufferPointer(from bytes: [UInt8]) -> Message? {
        guard bytes.count >= 8 else { return nil }
        
        return bytes.withUnsafeBufferPointer { buffer in
            var offset = 0
            
            // Decode id
            let idPtr = UnsafeRawPointer(buffer.baseAddress!).advanced(by: offset)
            let id = idPtr.load(as: UInt32.self).littleEndian
            offset += 4
            
            // Decode array length
            let lenPtr = UnsafeRawPointer(buffer.baseAddress!).advanced(by: offset)
            let len = Int(lenPtr.load(as: UInt32.self).littleEndian)
            offset += 4
            
            guard bytes.count >= offset + len * 4 else { return nil }
            
            // Decode array items
            var items = [UInt32]()
            items.reserveCapacity(len)
            for _ in 0..<len {
                let itemPtr = UnsafeRawPointer(buffer.baseAddress!).advanced(by: offset)
                let item = itemPtr.load(as: UInt32.self).littleEndian
                items.append(item)
                offset += 4
            }
            
            return Message(id: id, items: items)
        }
    }
    
    // Approach 3: Bulk copy for array items
    static func decodeBulkCopy(from bytes: [UInt8]) -> Message? {
        guard bytes.count >= 8 else { return nil }
        
        return bytes.withUnsafeBytes { buffer in
            var offset = 0
            
            // Decode id
            let id = UInt32(littleEndian: buffer.load(fromByteOffset: offset, as: UInt32.self))
            offset += 4
            
            // Decode array length
            let len = Int(UInt32(littleEndian: buffer.load(fromByteOffset: offset, as: UInt32.self)))
            offset += 4
            
            guard bytes.count >= offset + len * 4 else { return nil }
            
            // Bulk copy array items
            let items = [UInt32](unsafeUninitializedCapacity: len) { itemBuffer, count in
                let src = buffer.baseAddress!.advanced(by: offset).assumingMemoryBound(to: UInt32.self)
                for i in 0..<len {
                    itemBuffer[i] = UInt32(littleEndian: src[i])
                }
                count = len
            }
            
            return Message(id: id, items: items)
        }
    }
}

// MARK: - Main Entry Point

print("=== Swift Unsafe Optimization Benchmarks ===\n")

print("Configuration:")
print("- Compiler: Swift 5.9+ with -O -whole-module-optimization")
print("- Platform: macOS 13+")
print("- Array size: 100 UInt32 elements")
print("- Testing approaches: Naive, BulkCopy, Preallocated, UnsafeOptimized, Memcpy")
print()

let testData: [UInt32] = (0..<100).map { UInt32($0) }
let testMessage = Message(id: 42, items: testData)

print("--- Encoding Performance ---")

benchmark(name: "Approach 1: Naive (current - element by element)", iterations: 10_000) {
    let _ = testMessage.encodeNaive()
}.print()

benchmark(name: "Approach 2: BulkCopy (withUnsafeBufferPointer)", iterations: 10_000) {
    let _ = testMessage.encodeBulkCopy()
}.print()

benchmark(name: "Approach 3: Preallocated (copyMemory)", iterations: 10_000) {
    let _ = testMessage.encodePreallocated()
}.print()

benchmark(name: "Approach 4: UnsafeOptimized (unsafeUninitializedCapacity)", iterations: 10_000) {
    let _ = testMessage.encodeUnsafeOptimized()
}.print()

benchmark(name: "Approach 5: Memcpy (bulk array copy)", iterations: 10_000) {
    let _ = testMessage.encodeMemcpy()
}.print()

print()
print("--- Decoding Performance ---")

let naiveBytes = testMessage.encodeNaive()
let bufferBytes = testMessage.encodeBulkCopy()
let unsafeBytes = testMessage.encodeUnsafeOptimized()

benchmark(name: "Approach 1: Naive (current - slice by slice)", iterations: 10_000) {
    let _ = Message.decodeNaive(from: naiveBytes)
}.print()

benchmark(name: "Approach 2: BufferPointer (direct pointer access)", iterations: 10_000) {
    let _ = Message.decodeBufferPointer(from: bufferBytes)
}.print()

benchmark(name: "Approach 3: BulkCopy (bulk array decode)", iterations: 10_000) {
    let _ = Message.decodeBulkCopy(from: unsafeBytes)
}.print()

print()
print("--- Roundtrip Performance ---")

benchmark(name: "Naive -> Naive", iterations: 10_000) {
    let bytes = testMessage.encodeNaive()
    let _ = Message.decodeNaive(from: bytes)
}.print()

benchmark(name: "UnsafeOptimized -> BulkCopy", iterations: 10_000) {
    let bytes = testMessage.encodeUnsafeOptimized()
    let _ = Message.decodeBulkCopy(from: bytes)
}.print()

print()
print("--- Verification ---")

let bytes1 = testMessage.encodeNaive()
let bytes2 = testMessage.encodeBulkCopy()
let bytes3 = testMessage.encodePreallocated()
let bytes4 = testMessage.encodeUnsafeOptimized()
let bytes5 = testMessage.encodeMemcpy()

print("All encode methods produce identical output: \(bytes1 == bytes2 && bytes2 == bytes3 && bytes3 == bytes4 && bytes4 == bytes5)")

let msg1 = Message.decodeNaive(from: bytes1)
let msg2 = Message.decodeBufferPointer(from: bytes2)
let msg3 = Message.decodeBulkCopy(from: bytes3)

print("All decode methods produce identical output: \(msg1?.id == msg2?.id && msg2?.id == msg3?.id && msg1?.items == msg2?.items && msg2?.items == msg3?.items)")

print()
print("--- Performance Summary ---")
print("Expected improvements:")
print("- UnsafeOptimized encoding: ~2-3x faster (no append reallocation)")
print("- BulkCopy decoding: ~1.5-2x faster (direct memory access)")
print("- Memcpy: Depends on array size (better for large arrays)")
