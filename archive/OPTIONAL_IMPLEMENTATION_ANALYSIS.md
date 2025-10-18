# Optional Fields: Runtime Representation Analysis

## The Question

**What is the ACTUAL in-memory representation of optional fields in each language?**

Not just syntax, but the real bits and bytes at runtime.

---

## Language-by-Language Runtime Analysis

### Go: Pointer to Struct

**Generated Code**:
```go
type Plugin struct {
    Name     string      // Always present: {ptr, len}
    Metadata *Metadata   // Optional: pointer (8 bytes on 64-bit)
}

type Metadata struct {
    Version uint32
}
```

**In Memory (64-bit system)**:
```
Plugin instance:
┌─────────────────────────────────────────┐
│ Name.ptr:  0x7f8a2c000100  (8 bytes)   │
│ Name.len:  6                (8 bytes)   │  
│ Metadata:  0x7f8a2c000200  (8 bytes)   │ ← Pointer to Metadata
└─────────────────────────────────────────┘
Total: 24 bytes

If Metadata is nil:
┌─────────────────────────────────────────┐
│ Name.ptr:  0x7f8a2c000100  (8 bytes)   │
│ Name.len:  6                (8 bytes)   │
│ Metadata:  0x0000000000000000 (8 bytes)│ ← NULL pointer
└─────────────────────────────────────────┘
Total: 24 bytes (same size, just NULL)

Metadata (if present, allocated separately):
┌─────────────────────────────────────────┐
│ Version: 42 (4 bytes + 4 bytes padding) │
└─────────────────────────────────────────┘
At address: 0x7f8a2c000200
```

**Key Points**:
- ✅ Pointer is **8 bytes** (or 4 bytes on 32-bit systems)
- ✅ `nil` = all zeros (`0x0000000000000000`)
- ✅ Non-nil = points to separately allocated memory
- ✅ Struct size **same** whether optional is present or not
- ⚠️ **Indirection cost**: Following pointer requires memory access
- ⚠️ **Heap allocation**: Optional struct lives on heap (GC pressure)

---

### Rust: `Option<T>` - Tagged Union

**Generated Code**:
```rust
pub struct Plugin {
    pub name: String,
    pub metadata: Option<Metadata>,
}

pub struct Metadata {
    pub version: u32,
}
```

**In Memory (64-bit system)**:
```
Plugin instance:
┌─────────────────────────────────────────────────┐
│ name.ptr:     0x7f8a2c000100  (8 bytes)        │
│ name.len:     6                (8 bytes)        │
│ name.capacity: 10              (8 bytes)        │
│ metadata.tag: 1 (Some)         (1 byte)         │ ← Discriminant
│ metadata.padding:              (3 bytes)        │ ← Alignment
│ metadata.value.version: 42     (4 bytes)        │ ← Actual data (inline!)
└─────────────────────────────────────────────────┘
Total: 32 bytes

If metadata is None:
┌─────────────────────────────────────────────────┐
│ name.ptr:     0x7f8a2c000100  (8 bytes)        │
│ name.len:     6                (8 bytes)        │
│ name.capacity: 10              (8 bytes)        │
│ metadata.tag: 0 (None)         (1 byte)         │ ← Discriminant = None
│ metadata.padding:              (3 bytes)        │
│ metadata.value: (uninitialized) (4 bytes)       │ ← Memory exists but unused
└─────────────────────────────────────────────────┘
Total: 32 bytes (same size!)
```

**Option<T> Layout** (Rust standard library):
```rust
pub enum Option<T> {
    None,        // tag = 0
    Some(T),     // tag = 1, followed by T inline
}

// Memory layout:
// [discriminant: 1 byte][padding for T alignment][T data inline]
```

**Key Points**:
- ✅ **Inline storage**: Metadata lives directly in Plugin (no separate allocation)
- ✅ **Tag-based**: 1 byte discriminant (0=None, 1=Some)
- ✅ **Fixed size**: Same size whether Some or None
- ✅ **No heap allocation**: T is stored inline
- ✅ **Cache friendly**: No pointer chasing
- ⚠️ **Space overhead**: Memory reserved even when None
- ✅ **Null pointer optimization**: For `Option<&T>` and `Option<Box<T>>`, Rust uses NULL as None (no tag byte!)

**Special case - `Option<Box<T>>`** (Null Pointer Optimization):
```rust
pub struct Plugin {
    pub metadata: Option<Box<Metadata>>,  // Uses pointer as tag
}

// Memory layout:
// [pointer: 8 bytes]
// If None: pointer = 0x0000000000000000
// If Some: pointer = 0x7f8a2c000200 (actual heap address)
// No separate tag byte needed!
```

---

### Swift: Optional - Tagged Union (Similar to Rust)

**Generated Code**:
```swift
public struct Plugin {
    public var name: String
    public var metadata: Metadata?
}

public struct Metadata {
    public var version: UInt32
}
```

**In Memory**:
```
Plugin instance:
┌─────────────────────────────────────────────────┐
│ name.ptr:     0x7f8a2c000100  (8 bytes)        │
│ metadata.tag: 1 (some)         (1 byte)         │ ← Discriminant
│ metadata.padding:              (3 bytes)        │
│ metadata.value.version: 42     (4 bytes)        │ ← Inline
└─────────────────────────────────────────────────┘
Total: 16 bytes

If metadata is nil:
┌─────────────────────────────────────────────────┐
│ name.ptr:     0x7f8a2c000100  (8 bytes)        │
│ metadata.tag: 0 (none)         (1 byte)         │
│ metadata.padding:              (3 bytes)        │
│ metadata.value: (uninitialized) (4 bytes)       │
└─────────────────────────────────────────────────┘
Total: 16 bytes (same size)
```

**Key Points**:
- ✅ Similar to Rust: inline storage with tag
- ✅ No separate heap allocation for small types
- ✅ Cache friendly
- ⚠️ Space overhead when None

---

### C: Pointer (Manual Memory Management)

**Generated Code**:
```c
typedef struct {
    char *name;           // Heap-allocated string
    Metadata *metadata;   // Pointer to Metadata (or NULL)
} Plugin;

typedef struct {
    uint32_t version;
} Metadata;
```

**In Memory**:
```
Plugin instance:
┌─────────────────────────────────────────┐
│ name:     0x7f8a2c000100  (8 bytes)    │ ← Pointer to string
│ metadata: 0x7f8a2c000200  (8 bytes)    │ ← Pointer to Metadata
└─────────────────────────────────────────┘
Total: 16 bytes

If metadata is NULL:
┌─────────────────────────────────────────┐
│ name:     0x7f8a2c000100  (8 bytes)    │
│ metadata: 0x0000000000000000 (8 bytes) │ ← NULL
└─────────────────────────────────────────┘
Total: 16 bytes (same size)

Metadata (if present, separately allocated):
┌─────────────────────────────────────────┐
│ version: 42 (4 bytes)                   │
└─────────────────────────────────────────┘
At address: 0x7f8a2c000200
```

**Key Points**:
- ✅ Pointer = 8 bytes (64-bit) or 4 bytes (32-bit)
- ✅ NULL = 0x00
- ⚠️ Manual allocation/deallocation required
- ⚠️ Pointer indirection (cache miss risk)

---

### TypeScript/JavaScript: `undefined` or object reference

**Generated Code**:
```typescript
interface Plugin {
    name: string;
    metadata?: Metadata;  // Can be undefined
}

interface Metadata {
    version: number;
}
```

**Runtime (V8 engine)**:
```javascript
// Plugin object
{
    name: "Reverb",              // String object reference
    metadata: <reference>         // Reference to Metadata object (or undefined)
}

// If metadata is present:
metadata: 0x1a2b3c4d  // Pointer to Metadata object

// If metadata is absent:
metadata: undefined   // Special undefined value
```

**Key Points**:
- ⚠️ Everything is a reference/pointer in JS
- ⚠️ `undefined` is a special value (not NULL)
- ⚠️ JIT optimizations may change layout at runtime

---

## Comparison: Memory Layout

### Small Struct (like Metadata: 4 bytes)

| Language | Representation | Size When Present | Size When Absent | Heap Allocation? |
|----------|----------------|-------------------|------------------|------------------|
| **Go** | Pointer | 8 bytes (ptr) | 8 bytes (NULL) | ✅ Yes (separate) |
| **Rust** | `Option<T>` inline | 8 bytes (1 tag + 3 pad + 4 data) | 8 bytes (1 tag + 7 unused) | ❌ No (inline) |
| **Swift** | `T?` inline | 8 bytes (1 tag + 3 pad + 4 data) | 8 bytes (1 tag + 7 unused) | ❌ No (inline) |
| **C** | Pointer | 8 bytes (ptr) | 8 bytes (NULL) | ✅ Yes (manual malloc) |

**Observation**: Go and C use pointers (indirection), Rust and Swift use inline tagged unions (no indirection).

---

### Large Struct (like AudioParams: 1024 bytes)

| Language | Representation | Size When Present | Size When Absent | Notes |
|----------|----------------|-------------------|------------------|-------|
| **Go** | Pointer | 8 bytes (ptr) | 8 bytes (NULL) | ✅ Good: no space waste |
| **Rust** | `Option<T>` inline | 1025 bytes (1 tag + 1024 data) | 1025 bytes (1 tag + 1024 wasted) | ❌ Bad: wastes 1KB when None |
| **Rust** | `Option<Box<T>>` | 8 bytes (ptr) | 8 bytes (NULL) | ✅ Good: uses null pointer optimization |
| **Swift** | `T?` inline | 1025 bytes | 1025 bytes wasted | ❌ Bad: wastes space |
| **C** | Pointer | 8 bytes (ptr) | 8 bytes (NULL) | ✅ Good: no space waste |

**Observation**: For large structs, pointers are better. Rust solves this with `Box<T>`.

---

## SDP Code Generation Strategy

### Current Approach (Go)

```go
type Plugin struct {
    Name     string
    Metadata *Metadata  // Pointer = 8 bytes, always
}
```

**Memory**:
- Present: 8 bytes pointer + heap allocation
- Absent: 8 bytes NULL pointer

### Rust Generation Strategy

**For small structs** (<= 64 bytes):
```rust
pub struct Plugin {
    pub name: String,
    pub metadata: Option<Metadata>,  // Inline
}
```

**For large structs** (> 64 bytes):
```rust
pub struct Plugin {
    pub name: String,
    pub metadata: Option<Box<Metadata>>,  // Boxed (null pointer optimization)
}
```

**Why threshold at 64 bytes?**
- Below 64 bytes: Inline is fine (fits in cache line)
- Above 64 bytes: Boxing saves memory when None

### Swift Generation Strategy

**Small structs**:
```swift
public struct Plugin {
    public var metadata: Metadata?  // Inline
}
```

**Large structs**: Swift doesn't have `Box`, so might use class:
```swift
public class MetadataRef {  // Heap-allocated
    public var version: UInt32
}

public struct Plugin {
    public var metadata: MetadataRef?  // Optional reference
}
```

### C Generation Strategy

```c
typedef struct {
    Metadata *metadata;  // Always pointer
} Plugin;
```

Simple: always pointers.

---

## Wire Format Encoding

**All languages encode the same way** (1 byte presence + data):

```
Optional field wire format:
[presence: u8][data: variable bytes] (if presence = 1)
[presence: u8]                       (if presence = 0)

Example - Plugin with metadata present:
[name: "Reverb"]
[01]              // metadata present
[02 00 00 00]     // version = 2

Example - Plugin with metadata absent:
[name: "Reverb"]
[00]              // metadata absent (no more bytes)
```

---

## Decoding Strategy

### Go Decoder

```go
func decodePlugin(dest *Plugin, data []byte, offset *int, ctx *DecodeContext) error {
    // Decode name
    // ...
    
    // Decode optional metadata
    if data[*offset] == 1 {
        *offset += 1
        // Allocate on heap
        dest.Metadata = &Metadata{}
        if err := decodeMetadata(dest.Metadata, data, offset, ctx); err != nil {
            return err
        }
    } else {
        *offset += 1
        dest.Metadata = nil  // NULL pointer
    }
    
    return nil
}
```

**Allocations**: 1 heap allocation per present optional field

### Rust Decoder

```rust
fn decode_plugin(data: &[u8], offset: &mut usize) -> Result<Plugin, Error> {
    // Decode name
    let name = decode_string(data, offset)?;
    
    // Decode optional metadata
    let metadata = if data[*offset] == 1 {
        *offset += 1;
        Some(decode_metadata(data, offset)?)  // Inline in Option
    } else {
        *offset += 1;
        None
    };
    
    Ok(Plugin { name, metadata })
}
```

**Allocations**: 0 additional allocations (metadata stored inline in Option)

**For large structs with Box**:
```rust
let metadata = if data[*offset] == 1 {
    *offset += 1;
    Some(Box::new(decode_metadata(data, offset)?))  // 1 heap allocation
} else {
    *offset += 1;
    None
};
```

---

## Performance Comparison

### Cache Performance

**Small struct (Metadata: 4 bytes)**:

| Language | Layout | Cache Lines | Pointer Chases |
|----------|--------|-------------|----------------|
| Go | Pointer | 2 (Plugin + Metadata) | 1 |
| Rust | Inline | 1 (all in Plugin) | 0 |
| Swift | Inline | 1 | 0 |
| C | Pointer | 2 | 1 |

**Winner**: Rust/Swift (inline = better cache locality)

### Memory Usage

**Large struct (AudioParams: 1KB)**:

When 50% of optionals are None:

| Language | Layout | Memory Used |
|----------|--------|-------------|
| Go | Pointer | 1,000 present × (8 + 1024) + 1,000 absent × 8 = **1,040 KB** |
| Rust | `Option<T>` | 2,000 × 1025 = **2,050 KB** ❌ |
| Rust | `Option<Box<T>>` | 1,000 × (8 + 1024) + 1,000 × 8 = **1,040 KB** ✅ |
| C | Pointer | **1,040 KB** |

**Winner**: Go/C/Rust with Box (pointers save memory for large optional structs)

---

## Recommendation for SDP

### Code Generation Strategy

**Go**: Always use pointers
```go
metadata: *Metadata
```

**Rust**: Use size-based strategy
```rust
// Small structs (<= 64 bytes)
metadata: Option<Metadata>

// Large structs (> 64 bytes)
metadata: Option<Box<Metadata>>

// Recursive structs (always)
fallback: Option<Box<Plugin>>
```

**Swift**: Use optionals (inline for value types)
```swift
metadata: Metadata?
```

**C**: Always use pointers
```c
Metadata *metadata;
```

### Why This Strategy?

1. **Go/C**: Simple, consistent (always pointers)
2. **Rust**: Optimal (inline when small, boxed when large)
3. **Swift**: Idiomatic (Swift optionals are standard)
4. **All**: Same wire format (1 byte + data)

---

## Summary: What IS an Optional Field?

### Runtime Representation by Language

| Language | Small Struct | Large Struct | Recursive |
|----------|-------------|--------------|-----------|
| **Go** | `*T` (8-byte pointer) | `*T` (8-byte pointer) | `*T` |
| **Rust** | `Option<T>` (inline, 1 tag + T) | `Option<Box<T>>` (8-byte pointer with NULL optimization) | `Option<Box<T>>` |
| **Swift** | `T?` (inline, 1 tag + T) | `T?` (inline, wastes space) or class ref | `T?` (class) |
| **C** | `T*` (8-byte pointer) | `T*` (8-byte pointer) | `T*` |

### Key Takeaways

1. **Go/C**: Optional = Pointer (8 bytes, heap allocation)
2. **Rust**: Optional = Tagged union (inline or boxed, smart optimization)
3. **Swift**: Optional = Tagged union (inline, may waste space)
4. **Wire format**: All encode the same (1 byte presence + data)

### Memory Overhead

- **Wire format**: +1 byte per optional field
- **Go runtime**: +8 bytes struct size, +sizeof(T) heap if present
- **Rust runtime**: +1 byte tag + sizeof(T) always (or +8 bytes if boxed)
- **C runtime**: +8 bytes struct size, +sizeof(T) heap if present

**Bottom line**: In Go/C, optional = pointer. In Rust/Swift, optional = tagged inline (unless boxed).
