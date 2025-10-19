import Foundation

// MARK: - Benchmark Utilities

struct BenchmarkResult {
    let name: String
    let iterations: Int
    let totalTime: TimeInterval
    let avgTime: TimeInterval
    let opsPerSecond: Double
    
    func print() {
        Swift.print(String(format: "%-50s %10d iters  %8.2f ns/op  %12.0f ops/sec",
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

// MARK: - Test Data Structures

// Standard Swift Array
struct MessageWithArray {
    var id: UInt32
    var items: [UInt32]
    
    init(id: UInt32, items: [UInt32]) {
        self.id = id
        self.items = items
    }
}

// ContiguousArray (guaranteed contiguous storage)
struct MessageWithContiguousArray {
    var id: UInt32
    var items: ContiguousArray<UInt32>
    
    init(id: UInt32, items: ContiguousArray<UInt32>) {
        self.id = id
        self.items = items
    }
}

// Manual memory management with UnsafeBufferPointer
struct MessageWithUnsafeBuffer {
    var id: UInt32
    private var storage: UnsafeMutableBufferPointer<UInt32>
    
    var items: UnsafeBufferPointer<UInt32> {
        UnsafeBufferPointer(storage)
    }
    
    init(id: UInt32, capacity: Int) {
        self.id = id
        self.storage = UnsafeMutableBufferPointer<UInt32>.allocate(capacity: capacity)
    }
    
    mutating func append(_ value: UInt32) {
        // Simplified - in real code would need to track count and resize
        storage[0] = value
    }
    
    func deallocate() {
        storage.deallocate()
    }
}

// Immutable version with Array
struct ImmutableMessageWithArray {
    let id: UInt32
    let items: [UInt32]
}

// Class-based (reference semantics)
class MessageClass {
    var id: UInt32
    var items: [UInt32]
    
    init(id: UInt32, items: [UInt32]) {
        self.id = id
        self.items = items
    }
}

// MARK: - Encoding/Decoding Simulation

extension MessageWithArray {
    func encodeToBytes() -> [UInt8] {
        var bytes = [UInt8]()
        bytes.reserveCapacity(4 + 4 + items.count * 4)
        
        // Encode id
        withUnsafeBytes(of: id.littleEndian) { bytes.append(contentsOf: $0) }
        
        // Encode array length
        let len = UInt32(items.count).littleEndian
        withUnsafeBytes(of: len) { bytes.append(contentsOf: $0) }
        
        // Encode array items
        for item in items {
            withUnsafeBytes(of: item.littleEndian) { bytes.append(contentsOf: $0) }
        }
        
        return bytes
    }
    
    static func decode(from bytes: [UInt8]) -> MessageWithArray? {
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
        
        return MessageWithArray(id: id, items: items)
    }
}

extension MessageWithContiguousArray {
    func encodeToBytes() -> [UInt8] {
        var bytes = [UInt8]()
        bytes.reserveCapacity(4 + 4 + items.count * 4)
        
        withUnsafeBytes(of: id.littleEndian) { bytes.append(contentsOf: $0) }
        
        let len = UInt32(items.count).littleEndian
        withUnsafeBytes(of: len) { bytes.append(contentsOf: $0) }
        
        for item in items {
            withUnsafeBytes(of: item.littleEndian) { bytes.append(contentsOf: $0) }
        }
        
        return bytes
    }
    
    static func decode(from bytes: [UInt8]) -> MessageWithContiguousArray? {
        guard bytes.count >= 8 else { return nil }
        
        var offset = 0
        
        let id = bytes[offset..<offset+4].withUnsafeBytes {
            UInt32(littleEndian: $0.load(as: UInt32.self))
        }
        offset += 4
        
        let len = Int(bytes[offset..<offset+4].withUnsafeBytes {
            UInt32(littleEndian: $0.load(as: UInt32.self))
        })
        offset += 4
        
        guard bytes.count >= offset + len * 4 else { return nil }
        
        var items = ContiguousArray<UInt32>()
        items.reserveCapacity(len)
        for _ in 0..<len {
            let item = bytes[offset..<offset+4].withUnsafeBytes {
                UInt32(littleEndian: $0.load(as: UInt32.self))
            }
            items.append(item)
            offset += 4
        }
        
        return MessageWithContiguousArray(id: id, items: items)
    }
}

// MARK: - Main Benchmark Suite

func runArrayTypeBenchmarks() {
    print("--- Array Type Comparison ---")
    
    let testData: [UInt32] = (0..<100).map { UInt32($0) }
    
    // Standard Array allocation
    benchmark(name: "Array - Allocation", iterations: 100_000) {
        let _ = [UInt32]()
    }.print()
    
    benchmark(name: "Array - Allocation with capacity", iterations: 100_000) {
        var arr = [UInt32]()
        arr.reserveCapacity(100)
    }.print()
    
    // ContiguousArray allocation
    benchmark(name: "ContiguousArray - Allocation", iterations: 100_000) {
        let _ = ContiguousArray<UInt32>()
    }.print()
    
    benchmark(name: "ContiguousArray - Allocation with capacity", iterations: 100_000) {
        var arr = ContiguousArray<UInt32>()
        arr.reserveCapacity(100)
    }.print()
    
    // Population benchmarks
    benchmark(name: "Array - Populate 100 items", iterations: 10_000) {
        var arr = [UInt32]()
        arr.reserveCapacity(100)
        for i in 0..<100 {
            arr.append(UInt32(i))
        }
    }.print()
    
    benchmark(name: "ContiguousArray - Populate 100 items", iterations: 10_000) {
        var arr = ContiguousArray<UInt32>()
        arr.reserveCapacity(100)
        for i in 0..<100 {
            arr.append(UInt32(i))
        }
    }.print()
    
    // Iteration benchmarks
    var sum: UInt32 = 0
    benchmark(name: "Array - Iterate and sum", iterations: 100_000) {
        for item in testData {
            sum = sum &+ item
        }
    }.print()
    
    let contiguousData = ContiguousArray(testData)
    benchmark(name: "ContiguousArray - Iterate and sum", iterations: 100_000) {
        for item in contiguousData {
            sum = sum &+ item
        }
    }.print()
}

func runStructVsClassBenchmarks() {
    print("--- Struct vs Class Comparison ---")
    
    let testData: [UInt32] = (0..<100).map { UInt32($0) }
    
    // Struct allocation
    benchmark(name: "Struct with Array - Allocation", iterations: 100_000) {
        let _ = MessageWithArray(id: 42, items: [])
    }.print()
    
    benchmark(name: "Struct with Array - Full init", iterations: 10_000) {
        let _ = MessageWithArray(id: 42, items: testData)
    }.print()
    
    // Class allocation
    benchmark(name: "Class with Array - Allocation", iterations: 100_000) {
        let _ = MessageClass(id: 42, items: [])
    }.print()
    
    benchmark(name: "Class with Array - Full init", iterations: 10_000) {
        let _ = MessageClass(id: 42, items: testData)
    }.print()
    
    // Copy semantics
    let structMsg = MessageWithArray(id: 42, items: testData)
    benchmark(name: "Struct - Copy", iterations: 100_000) {
        let _ = structMsg
    }.print()
    
    let classMsg = MessageClass(id: 42, items: testData)
    benchmark(name: "Class - Reference copy", iterations: 100_000) {
        let _ = classMsg
    }.print()
}

func runMutabilityBenchmarks() {
    print("--- Mutability Comparison ---")
    
    let testData: [UInt32] = (0..<100).map { UInt32($0) }
    
    // Mutable
    benchmark(name: "Mutable var - Create and modify", iterations: 100_000) {
        var msg = MessageWithArray(id: 42, items: testData)
        msg.id = 43
    }.print()
    
    // Immutable
    benchmark(name: "Immutable let - Create", iterations: 100_000) {
        let msg = ImmutableMessageWithArray(id: 42, items: testData)
        let _ = msg.id
    }.print()
    
    // Conversion overhead
    benchmark(name: "Convert mutable to immutable", iterations: 100_000) {
        let mutable = MessageWithArray(id: 42, items: testData)
        let immutable = ImmutableMessageWithArray(id: mutable.id, items: mutable.items)
        let _ = immutable
    }.print()
}

func runEncodingBenchmarks() {
    print("--- Encoding/Decoding Performance ---")
    
    let testData: [UInt32] = (0..<100).map { UInt32($0) }
    
    // Array encoding
    let arrayMsg = MessageWithArray(id: 42, items: testData)
    benchmark(name: "Array - Encode to bytes", iterations: 10_000) {
        let _ = arrayMsg.encodeToBytes()
    }.print()
    
    let arrayBytes = arrayMsg.encodeToBytes()
    benchmark(name: "Array - Decode from bytes", iterations: 10_000) {
        let _ = MessageWithArray.decode(from: arrayBytes)
    }.print()
    
    benchmark(name: "Array - Roundtrip (encode + decode)", iterations: 10_000) {
        let bytes = arrayMsg.encodeToBytes()
        let _ = MessageWithArray.decode(from: bytes)
    }.print()
    
    // ContiguousArray encoding
    let contiguousMsg = MessageWithContiguousArray(id: 42, items: ContiguousArray(testData))
    benchmark(name: "ContiguousArray - Encode to bytes", iterations: 10_000) {
        let _ = contiguousMsg.encodeToBytes()
    }.print()
    
    let contiguousBytes = contiguousMsg.encodeToBytes()
    benchmark(name: "ContiguousArray - Decode from bytes", iterations: 10_000) {
        let _ = MessageWithContiguousArray.decode(from: contiguousBytes)
    }.print()
    
    benchmark(name: "ContiguousArray - Roundtrip (encode + decode)", iterations: 10_000) {
        let bytes = contiguousMsg.encodeToBytes()
        let _ = MessageWithContiguousArray.decode(from: bytes)
    }.print()
}

func runMemoryLayoutTests() {
    print("--- Memory Layout Analysis ---")
    
    print(String(format: "MemoryLayout<MessageWithArray>: size=%d, stride=%d, alignment=%d",
                MemoryLayout<MessageWithArray>.size,
                MemoryLayout<MessageWithArray>.stride,
                MemoryLayout<MessageWithArray>.alignment))
    
    print(String(format: "MemoryLayout<MessageWithContiguousArray>: size=%d, stride=%d, alignment=%d",
                MemoryLayout<MessageWithContiguousArray>.size,
                MemoryLayout<MessageWithContiguousArray>.stride,
                MemoryLayout<MessageWithContiguousArray>.alignment))
    
    print(String(format: "MemoryLayout<ImmutableMessageWithArray>: size=%d, stride=%d, alignment=%d",
                MemoryLayout<ImmutableMessageWithArray>.size,
                MemoryLayout<ImmutableMessageWithArray>.stride,
                MemoryLayout<ImmutableMessageWithArray>.alignment))
    
    print(String(format: "MemoryLayout<[UInt32]>: size=%d, stride=%d, alignment=%d",
                MemoryLayout<[UInt32]>.size,
                MemoryLayout<[UInt32]>.stride,
                MemoryLayout<[UInt32]>.alignment))
    
    print(String(format: "MemoryLayout<ContiguousArray<UInt32>>: size=%d, stride=%d, alignment=%d",
                MemoryLayout<ContiguousArray<UInt32>>.size,
                MemoryLayout<ContiguousArray<UInt32>>.stride,
                MemoryLayout<ContiguousArray<UInt32>>.alignment))
}

// MARK: - Main Entry Point

print("=== Swift Data Structure Benchmarks ===\n")

print("Configuration:")
print("- Compiler: Swift 5.9+ with -O -whole-module-optimization")
print("- Platform: macOS 13+")
print("- Array size: 100 UInt32 elements")
print()

runArrayTypeBenchmarks()
print()
runStructVsClassBenchmarks()
print()
runMutabilityBenchmarks()
print()
runEncodingBenchmarks()
print()
runMemoryLayoutTests()
