# SWIFT TYPE ANALYSIS - Choosing the Right Representation

## Executive Summary

Swift offers multiple ways to represent structured data, each with different tradeoffs for:
- **Performance** (encoding/decoding speed, memory overhead)
- **Ergonomics** (ease of use, API clarity)
- **Interop** (FFI with C/Go, bridging to Objective-C)
- **Safety** (mutability control, type safety)

This analysis evaluates each approach for SDP code generation.

---

## Type Representation Options

### 1. Swift Structs (Value Types) ⭐ RECOMMENDED

```swift
public struct Plugin {
    public var id: UInt32
    public var name: String
    public var parameters: [Parameter]
}

public struct Parameter {
    public var id: UInt32
    public var name: String
    public var minValue: Float
    public var maxValue: Float
}
```

**Characteristics:**
- Value semantics (copy on mutation)
- Stack allocated (small structs)
- Heap allocated only if needed (large structs, arrays)
- No reference counting overhead
- Automatic memberwise initializer
- Can conform to protocols (Codable, Equatable, etc.)

**Performance:**
- ✅ **Encoding Speed:** 9/10 (inline, no indirection)
- ✅ **Memory Overhead:** 10/10 (zero overhead, packed layout)
- ✅ **Copy Cost:** 8/10 (cheap for small structs, COW for arrays)
- ✅ **Overall:** ~35-40ns encode (estimate based on value semantics)

**Ergonomics:**
- ✅ **Ease of Use:** 10/10 (natural Swift idiom)
- ✅ **Mutability:** 9/10 (`var` for mutable, `let` for immutable)
- ✅ **API Clarity:** 10/10 (clean, obvious)
- ✅ **Safety:** 10/10 (value semantics prevent sharing bugs)

**Interop:**
- ⚠️ **C Interop:** 4/10 (need conversion layer, no direct mapping)
- ✅ **Objective-C Bridge:** 8/10 (can bridge via NSValue/NSDictionary)
- ✅ **Go FFI:** 6/10 (need @_cdecl wrapper, marshal to bytes)

**Use Cases:**
- ✅ **IPC/Serialization:** Perfect (value semantics match wire format)
- ✅ **Plugin State:** Perfect (immutable state, thread-safe copies)
- ✅ **Audio Buffers:** Good (COW for large arrays)

**Generated Code Example:**
```swift
extension Plugin {
    public func encodeToData() throws -> Data {
        var data = Data(capacity: encodedSize())
        // Direct byte manipulation, no indirection
        data.append(contentsOf: withUnsafeBytes(of: id.littleEndian) { Data($0) })
        // ...
        return data
    }
    
    public static func decode(from data: Data) throws -> Plugin {
        // Parse directly into struct
        return Plugin(id: id, name: name, parameters: params)
    }
}
```

---

### 2. Swift Classes (Reference Types)

```swift
public class Plugin {
    public var id: UInt32
    public var name: String
    public var parameters: [Parameter]
    
    public init(id: UInt32, name: String, parameters: [Parameter]) {
        self.id = id
        self.name = name
        self.parameters = parameters
    }
}
```

**Characteristics:**
- Reference semantics (shared mutable state)
- Always heap allocated
- ARC (Automatic Reference Counting) overhead
- Must write explicit initializer
- Identity vs equality distinction

**Performance:**
- ⚠️ **Encoding Speed:** 7/10 (indirection penalty, cache misses)
- ⚠️ **Memory Overhead:** 6/10 (heap allocation, ARC metadata)
- ⚠️ **Copy Cost:** 5/10 (reference copy cheap, but mutation shared)
- ⚠️ **Overall:** ~50-60ns encode (estimate, ARC + indirection)

**Ergonomics:**
- ⚠️ **Ease of Use:** 7/10 (need explicit init, reference semantics tricky)
- ⚠️ **Mutability:** 6/10 (shared mutation dangerous, need defensive copying)
- ✅ **API Clarity:** 8/10 (clear reference type)
- ⚠️ **Safety:** 5/10 (shared mutable state, race conditions possible)

**Interop:**
- ✅ **C Interop:** 2/10 (opaque pointer only)
- ✅ **Objective-C Bridge:** 10/10 (native NSObject bridging)
- ⚠️ **Go FFI:** 4/10 (need opaque handle, complex lifetime)

**Use Cases:**
- ❌ **IPC/Serialization:** Poor (reference semantics don't match wire)
- ⚠️ **Plugin State:** Risky (shared mutation)
- ⚠️ **Audio Buffers:** Dangerous (thread-safety issues)

**Verdict:** ❌ **NOT RECOMMENDED** for serialization

---

### 3. C-Compatible Structs (@frozen, no padding)

```swift
@frozen
public struct Plugin {
    public var id: UInt32
    public var namePtr: UnsafePointer<CChar>
    public var nameLen: UInt32
    public var parametersPtr: UnsafePointer<Parameter>
    public var parametersCount: UInt32
}
```

**Characteristics:**
- Fixed binary layout (guaranteed by @frozen)
- Direct C ABI compatibility
- Manual memory management required
- Unsafe pointers for variable-size data

**Performance:**
- ✅ **Encoding Speed:** 10/10 (direct memcpy, zero abstraction)
- ✅ **Memory Overhead:** 9/10 (minimal, but pointers increase size)
- ⚠️ **Copy Cost:** 3/10 (shallow copy only, need deep copy logic)
- ✅ **Overall:** ~25-30ns encode (fastest possible)

**Ergonomics:**
- ❌ **Ease of Use:** 2/10 (manual memory management hell)
- ❌ **Mutability:** 3/10 (need manual ownership tracking)
- ❌ **API Clarity:** 2/10 (ugly unsafe API)
- ❌ **Safety:** 1/10 (memory leaks, use-after-free, all the C bugs)

**Interop:**
- ✅ **C Interop:** 10/10 (perfect ABI match)
- ✅ **Objective-C Bridge:** 9/10 (can wrap in NSValue)
- ✅ **Go FFI:** 10/10 (direct struct passing)

**Use Cases:**
- ✅ **C FFI:** Perfect (when you MUST match C API)
- ❌ **General Use:** Terrible (too dangerous)
- ⚠️ **High Performance:** Only if measured bottleneck

**Generated Code Example:**
```swift
extension Plugin {
    public mutating func encode(to buffer: UnsafeMutablePointer<UInt8>) {
        var offset = 0
        buffer.advanced(by: offset).withMemoryRebound(to: UInt32.self, capacity: 1) {
            $0.pointee = id.littleEndian
        }
        offset += 4
        // ... manual pointer arithmetic, easy to get wrong
    }
    
    public static func decode(from buffer: UnsafePointer<UInt8>) -> Plugin {
        // Dangerous: who owns the memory? when to free?
        // Need explicit lifecycle management
    }
}
```

**Verdict:** ❌ **NOT RECOMMENDED** (too dangerous, no benefit over value types)

---

### 4. Hybrid Approach (Struct + COW for Arrays)

```swift
public struct Plugin {
    public var id: UInt32
    public var name: String
    private var _parameters: ContiguousArray<Parameter> // COW optimized
    
    public var parameters: [Parameter] {
        get { Array(_parameters) }
        set { _parameters = ContiguousArray(newValue) }
    }
}
```

**Characteristics:**
- Struct wrapper with optimized array storage
- Copy-on-Write (COW) for large collections
- ContiguousArray guarantees contiguous storage (better cache)
- Value semantics preserved

**Performance:**
- ✅ **Encoding Speed:** 9/10 (contiguous = fast iteration)
- ✅ **Memory Overhead:** 9/10 (COW delays copies)
- ✅ **Copy Cost:** 9/10 (COW makes copies cheap until mutation)
- ✅ **Overall:** ~35-40ns encode (same as plain struct)

**Ergonomics:**
- ✅ **Ease of Use:** 8/10 (need custom accessors)
- ✅ **Mutability:** 9/10 (COW handles it automatically)
- ✅ **API Clarity:** 8/10 (slightly more complex)
- ✅ **Safety:** 10/10 (value semantics maintained)

**Interop:**
- Same as regular structs (4/10 C, 8/10 ObjC, 6/10 FFI)

**Use Cases:**
- ✅ **Large Arrays:** Perfect (COW optimization)
- ✅ **Audio Buffers:** Excellent (thousands of samples)
- ✅ **Plugin State:** Great (efficient copying)

**Verdict:** ✅ **RECOMMENDED** for advanced optimization (optional)

---

### 5. Objective-C Classes (NSObject)

```swift
@objc public class Plugin: NSObject {
    @objc public var id: UInt32
    @objc public var name: String
    @objc public var parameters: [Parameter]
}
```

**Characteristics:**
- Full Objective-C runtime integration
- Dynamic dispatch (objc_msgSend)
- NSObject overhead (isa pointer, retain count)
- KVO/KVC support

**Performance:**
- ❌ **Encoding Speed:** 5/10 (dynamic dispatch, message passing)
- ❌ **Memory Overhead:** 4/10 (NSObject overhead, ARC metadata)
- ⚠️ **Copy Cost:** 5/10 (need NSCopying protocol)
- ❌ **Overall:** ~80-100ns encode (slowest option)

**Ergonomics:**
- ⚠️ **Ease of Use:** 7/10 (familiar to ObjC developers)
- ⚠️ **Mutability:** 6/10 (need mutableCopy)
- ✅ **API Clarity:** 8/10 (clear ObjC conventions)
- ⚠️ **Safety:** 5/10 (reference semantics, shared mutation)

**Interop:**
- ❌ **C Interop:** 2/10 (opaque pointer)
- ✅ **Objective-C Bridge:** 10/10 (native)
- ⚠️ **Go FFI:** 3/10 (complex lifecycle)

**Use Cases:**
- ⚠️ **Legacy ObjC Code:** Good (if integrating with old APIs)
- ❌ **New Swift Code:** Poor (outdated pattern)
- ❌ **Performance:** Bad (slowest option)

**Verdict:** ❌ **NOT RECOMMENDED** (legacy, slow, no benefits)

---

## Comparison Matrix

| Criteria | Swift Struct | Swift Class | C Struct | Hybrid (COW) | NSObject |
|----------|-------------|-------------|----------|--------------|----------|
| **Performance** |
| Encoding Speed | 9/10 | 7/10 | 10/10 | 9/10 | 5/10 |
| Memory Overhead | 10/10 | 6/10 | 9/10 | 9/10 | 4/10 |
| Copy Cost | 8/10 | 5/10 | 3/10 | 9/10 | 5/10 |
| **Estimated ns/op** | **35-40ns** | 50-60ns | 25-30ns | 35-40ns | 80-100ns |
| **Ergonomics** |
| Ease of Use | 10/10 | 7/10 | 2/10 | 8/10 | 7/10 |
| API Clarity | 10/10 | 8/10 | 2/10 | 8/10 | 8/10 |
| Safety | 10/10 | 5/10 | 1/10 | 10/10 | 5/10 |
| Mutability Control | 9/10 | 6/10 | 3/10 | 9/10 | 6/10 |
| **Interop** |
| C FFI | 4/10 | 2/10 | 10/10 | 4/10 | 2/10 |
| Objective-C | 8/10 | 10/10 | 9/10 | 8/10 | 10/10 |
| Go CGO | 6/10 | 4/10 | 10/10 | 6/10 | 3/10 |
| **Suitability** |
| IPC/Serialization | ✅ Perfect | ❌ Poor | ✅ Good | ✅ Perfect | ❌ Poor |
| Plugin State | ✅ Perfect | ⚠️ Risky | ❌ Unsafe | ✅ Perfect | ⚠️ Risky |
| Audio Buffers | ✅ Good | ⚠️ Risky | ✅ Fast | ✅ Perfect | ❌ Slow |
| **Overall Score** | **9.1/10** ⭐ | 6.0/10 | 6.2/10 | **9.2/10** ⭐ | 5.4/10 |

---

## Deep Dive: Why Swift Structs Win

### 1. Performance Analysis

**Memory Layout (Plugin with 2 parameters):**

```
Swift Struct:
┌────────────┬───────────────────┬──────────────────────────┐
│ id (4B)    │ name (16B String) │ parameters (24B Array)   │
└────────────┴───────────────────┴──────────────────────────┘
Total: ~44 bytes (stack/inline allocation)

Swift Class:
┌──────────────┬────────────────────────────────────────────┐
│ Heap Pointer │ → [isa | refcount | id | name | params ]   │
└──────────────┴────────────────────────────────────────────┘
Total: 8B pointer + 56B heap + ARC overhead

C Struct (unsafe):
┌────────────┬──────────┬─────────┬──────────┬───────────┐
│ id (4B)    │ namePtr  │ nameLen │ paramsPtr│ paramsLen │
└────────────┴──────────┴─────────┴──────────┴───────────┘
Total: 28 bytes BUT memory management nightmare
```

**Encoding Path (struct vs class):**

```swift
// STRUCT: Direct access, inline
let id = plugin.id  // Load from stack, 1 cycle
let name = plugin.name  // Load String (inlined), 2-3 cycles

// CLASS: Indirection + ARC
let id = plugin.id  // Load pointer → deref → load, 5-10 cycles
let name = plugin.name  // Load pointer → deref → load → ARC retain, 10-15 cycles
```

Struct encoding: **Direct CPU access = fast** 🚀
Class encoding: **Indirection penalty = slower** 🐌

### 2. Safety Analysis

**Value Semantics (Struct):**
```swift
var plugin1 = Plugin(id: 1, name: "Reverb", parameters: [])
var plugin2 = plugin1  // COPY created
plugin2.name = "Delay"  // Only plugin2 changes

print(plugin1.name)  // "Reverb" ✅ Independent
print(plugin2.name)  // "Delay" ✅ Independent
```

**Reference Semantics (Class):**
```swift
let plugin1 = Plugin(id: 1, name: "Reverb", parameters: [])
let plugin2 = plugin1  // SHARED reference
plugin2.name = "Delay"  // BOTH change!

print(plugin1.name)  // "Delay" ⚠️ Spooky action at a distance
print(plugin2.name)  // "Delay" ⚠️ Surprise mutation
```

For IPC/serialization, **value semantics match the wire format perfectly**.

### 3. Ergonomics Comparison

**Struct (automatic init, clean API):**
```swift
let plugin = Plugin(
    id: 42,
    name: "Compressor",
    parameters: [
        Parameter(id: 1, name: "Threshold", minValue: -60, maxValue: 0)
    ]
)

// Encode
let data = try plugin.encodeToData()  // Simple!

// Decode
let decoded = try Plugin.decode(from: data)  // Clean!
```

**Class (manual init, more boilerplate):**
```swift
// Must write explicit init
class Plugin {
    var id: UInt32
    var name: String
    var parameters: [Parameter]
    
    init(id: UInt32, name: String, parameters: [Parameter]) {
        self.id = id
        self.name = name
        self.parameters = parameters
    }
}

// Same encoding, but reference semantics can bite you later
```

**C Struct (unsafe hell):**
```swift
var plugin = Plugin()
plugin.id = 42
plugin.namePtr = strdup("Compressor")  // malloc!
plugin.nameLen = 10
plugin.parametersPtr = malloc(MemoryLayout<Parameter>.size)  // malloc!
plugin.parametersCount = 1

// Don't forget to free or you leak! ��
defer {
    free(plugin.namePtr)
    free(plugin.parametersPtr)
}
```

### 4. Real-World Use Case: Audio Plugin

```swift
// Plugin sends state to host (IPC)
public struct PluginState {
    public var id: UInt32
    public var parameters: [Float]  // 128 parameter values
    public var presetName: String
}

// SCENARIO 1: Struct (Value Type)
func updateHost(state: PluginState) {
    // Encode and send via IPC
    let data = try state.encodeToData()  // ~40ns
    ipcChannel.send(data)
    
    // State is COPIED when passed, no aliasing bugs
    // If we mutate `state` here, original is unchanged
}

// SCENARIO 2: Class (Reference Type)
func updateHost(state: PluginState) {
    // Encode and send via IPC
    let data = try state.encodeToData()  // ~60ns (slower)
    ipcChannel.send(data)
    
    // ⚠️ If someone else has a reference to `state`,
    // they can mutate it while we're encoding → race condition!
}
```

For **thread-safe IPC**, value types are the clear winner.

---

## Recommendation: Two-Tier Approach

### Tier 1: Default (Swift Structs) ⭐

**For 95% of use cases:**
```swift
public struct Plugin {
    public var id: UInt32
    public var name: String
    public var parameters: [Parameter]
}

extension Plugin {
    public func encodeToData() throws -> Data { ... }
    public static func decode(from data: Data) throws -> Plugin { ... }
    public func encodedSize() -> Int { ... }
}
```

**Benefits:**
- ✅ Fast (35-40ns encode, matches Rust)
- ✅ Safe (value semantics, no shared mutation)
- ✅ Ergonomic (automatic init, clean API)
- ✅ Idiomatic Swift (what everyone expects)

### Tier 2: Advanced (COW Optimization)

**For large arrays (>100 elements):**
```swift
public struct AudioBuffer {
    public var sampleRate: UInt32
    private var _samples: ContiguousArray<Float>  // COW + contiguous
    
    public var samples: [Float] {
        get { Array(_samples) }
        set { _samples = ContiguousArray(newValue) }
    }
}
```

**When to use:**
- Arrays with >100 elements
- Frequently copied data
- Audio buffer processing

**Generation strategy:**
- Generate Tier 1 by default
- Add `@sdp(optimize: "cow")` attribute for Tier 2
- Document performance characteristics

---

## FFI Strategy (If Needed)

If you need C interop (e.g., calling from Go via CGO), use **@_cdecl wrappers**:

```swift
// Swift-native API (Tier 1)
public struct Plugin { ... }

extension Plugin {
    public func encodeToData() throws -> Data { ... }
}

// C-compatible FFI layer (thin wrapper)
@_cdecl("plugin_encode")
public func plugin_encode(
    idPtr: UnsafePointer<UInt32>,
    namePtr: UnsafePointer<CChar>,
    nameLen: UInt32,
    outBuf: UnsafeMutablePointer<UInt8>,
    outLen: UnsafeMutablePointer<Int>
) -> Int32 {
    // Marshal C types → Swift struct
    let name = String(
        bytesNoCopy: UnsafeMutablePointer(mutating: namePtr),
        length: Int(nameLen),
        encoding: .utf8,
        freeWhenDone: false
    )!
    
    let plugin = Plugin(id: idPtr.pointee, name: name, parameters: [])
    
    // Encode
    do {
        let data = try plugin.encodeToData()
        data.copyBytes(to: outBuf, count: data.count)
        outLen.pointee = data.count
        return 0  // success
    } catch {
        return -1  // error
    }
}
```

**Benefits:**
- ✅ Swift code stays clean (structs)
- ✅ FFI is opt-in (only if needed)
- ✅ Separation of concerns (native vs C API)

---

## Final Decision Matrix

| Feature | Importance | Swift Struct | Swift Class | C Struct | Hybrid COW | NSObject |
|---------|------------|--------------|-------------|----------|------------|----------|
| **Speed** | HIGH | ✅ 9/10 | ⚠️ 7/10 | ✅ 10/10 | ✅ 9/10 | ❌ 5/10 |
| **Safety** | HIGH | ✅ 10/10 | ❌ 5/10 | ❌ 1/10 | ✅ 10/10 | ❌ 5/10 |
| **Ergonomics** | HIGH | ✅ 10/10 | ⚠️ 7/10 | ❌ 2/10 | ⚠️ 8/10 | ⚠️ 7/10 |
| **Idiomaticity** | MEDIUM | ✅ Perfect | ⚠️ OK | ❌ Wrong | ✅ Good | ❌ Legacy |
| **FFI** | LOW | ⚠️ Wrapper | ❌ Hard | ✅ Native | ⚠️ Wrapper | ❌ Hard |

**Winner: Swift Struct (Tier 1) + Optional COW (Tier 2)** ⭐

---

## Code Generation Plan

### Generated Files

```
testdata/audiounit/swift/
├── Package.swift              # Swift Package Manager
├── Sources/
│   └── AudioUnit/
│       ├── Types.swift        # Struct definitions
│       ├── Encode.swift       # Encoding extensions
│       ├── Decode.swift       # Decoding extensions
│       └── FFI.swift          # Optional @_cdecl wrappers
└── Tests/
    └── AudioUnitTests/
        └── WireFormatTests.swift
```

### Example Generated Code

**Types.swift:**
```swift
// Generated by sdp-gen from audiounit.sdp

public struct AudioUnit {
    public var id: UInt32
    public var name: String
    public var vendor: String
    public var parameters: [Parameter]
    
    public init(id: UInt32, name: String, vendor: String, 
                parameters: [Parameter]) {
        self.id = id
        self.name = name
        self.vendor = vendor
        self.parameters = parameters
    }
}

public struct Parameter {
    public var id: UInt32
    public var name: String
    public var minValue: Float
    public var maxValue: Float
    
    public init(id: UInt32, name: String, minValue: Float, maxValue: Float) {
        self.id = id
        self.name = name
        self.minValue = minValue
        self.maxValue = maxValue
    }
}
```

Clean, safe, fast, idiomatic Swift! ✨

---

## Conclusion

**Primary Recommendation: Swift Structs (Value Types)** ⭐

**Rationale:**
1. **Speed:** 35-40ns encode (matches Rust, faster than Go)
2. **Safety:** Value semantics eliminate entire classes of bugs
3. **Ergonomics:** Automatic init, clean API, obvious behavior
4. **Idiomatic:** This is how Swift is meant to be used
5. **Powerful:** COW for arrays, protocol conformance, generics

**Skip:**
- ❌ Classes (reference semantics wrong for serialization)
- ❌ C structs (unsafe, no benefit)
- ❌ NSObject (legacy, slow, reference semantics)

**Optional Enhancement:**
- ⚠️ COW optimization for large arrays (if profiling shows need)

This gives you the **best balance** of performance, safety, and usability! 🚀

