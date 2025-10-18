# Serial Data Protocol - Design Specification

**Version:** 0.2.0-rc1  
**Date:** October 18, 2025  
**Status:** Release Candidate

## 1. Overview

### 1.1 Purpose

Serial Data Protocol (SDP) is a binary serialization format for efficient cross-language data transfer in controlled environments. It generates encoders and decoders from schema definitions, producing predictable binary output with fixed-width integers.

### 1.2 Design Principles

- **Predictable performance:** Fixed-width integers, single allocation encoding, known buffer sizes
- **Simple wire format:** No varint encoding, no padding, straightforward byte layout
- **Type safety:** Schema-driven code generation with compile-time checks
- **Zero dependencies:** Generated code uses only language standard libraries
- **Composability:** Provides stdlib interfaces (io.Writer/Reader), not implementations

### 1.3 Target Use Cases

SDP is designed for:

- **IPC (Inter-Process Communication)** - Same-machine data transfer between processes
- **FFI scenarios** - Crossing language boundaries (C ↔ Go, Swift ↔ Go, etc.)
- **Bulk data transfer** - Moving large datasets efficiently (audio plugin lists, device info, etc.)
- **Known schemas** - Both encoder and decoder compiled from identical schema
- **Trusted environments** - Data sources under your control

### 1.4 Explicit Non-Goals

SDP does NOT provide:

- **Schema evolution** - Breaking schema changes require recompilation of both sides
- **Cross-platform serialization** - Same architecture assumed (little-endian)
- **Network protocol features** - No versioning, negotiation, or endianness handling
- **Long-term storage** - Schema changes make old data unreadable (use message mode + versioning)
- **Maximum compression** - Fixed-width integers are larger than varint (compose with gzip for 68% reduction)
- **Untrusted data validation** - Assumes trusted data sources

### 1.5 Performance Characteristics

**Verified benchmarks** from real-world workloads (see [benchmarks/](benchmarks/) for methodology):

**Large dataset (62 plugins, 1,759 parameters, 115 KB wire size):**
- Encode: 39.3 µs (1 allocation)
- Decode: 98.1 µs (4,638 allocations - one per struct/string)
- Roundtrip: 141.0 µs (4,639 allocations)
- Peak RAM: 313 KB (112 KB encode, 201 KB decode)
- **6.1× faster encoding than Protocol Buffers** (39.3 µs vs 240.7 µs)
- **3.2× faster decoding than Protocol Buffers** (98.1 µs vs 313.1 µs)
- **3.9× faster roundtrip than Protocol Buffers** (141.0 µs vs 552.3 µs)
- **30% less RAM than Protocol Buffers** (313 KB vs 446 KB peak)

**Small messages (primitives, ~50 bytes):**
- Regular mode: 44.25 ns roundtrip
- Message mode: 85.54 ns roundtrip (+93% for type identification)
- Optional present: 58.38 ns roundtrip
- Optional absent: 15.55 ns roundtrip (65% faster than present)

See [benchmarks/RESULTS.md](benchmarks/RESULTS.md) and [benchmarks/MEMORY_ANALYSIS.md](benchmarks/MEMORY_ANALYSIS.md) for detailed measurements.

## 2. Type System

### 2.1 Primitive Types

| Type      | Size    | Wire Format        |
|-----------|---------|-------------------|
| `u8`      | 1 byte  | Unsigned 8-bit    |
| `u16`     | 2 bytes | Unsigned 16-bit   |
| `u32`     | 4 bytes | Unsigned 32-bit   |
| `u64`     | 8 bytes | Unsigned 64-bit   |
| `i8`      | 1 byte  | Signed 8-bit      |
| `i16`     | 2 bytes | Signed 16-bit     |
| `i32`     | 4 bytes | Signed 32-bit     |
| `i64`     | 8 bytes | Signed 64-bit     |
| `f32`     | 4 bytes | IEEE 754 float    |
| `f64`     | 8 bytes | IEEE 754 double   |
| `bool`    | 1 byte  | 0=false, 1=true   |

### 2.2 String Type

**Type:** `str`

**Wire Format:**
```
[u32: byte_length][utf8_bytes]
```

- Length-prefixed (no null terminator in wire format)
- UTF-8 encoded
- Trust encoder (no validation on encode), validate on decode if required by language

### 2.3 Array Type

**Syntax:** `[]T` where T is primitive, string, or struct

**Constraints:**
- Homogeneous: all elements same type
- One level of nesting per array (arrays contain primitives, strings, OR structs, not arrays)

**Wire Format:**
```
[u32: element_count][element_0][element_1]...[element_n]
```

### 2.4 Struct Type

**Syntax:**
```
struct Name {
    field1: type1
    field2: type2
    ...
}
```

**Constraints:**
- Fixed set of fields defined in schema
- Fields written in definition order
- Can contain primitives, strings, arrays, or nested structs

**Wire Format:**
```
[field1_data][field2_data]...
```
(No struct metadata in wire format - structure known from schema)

### 2.5 Nested Structures

**Supported:**
```rust
struct PluginList {
    plugins: []Plugin,
}

struct Plugin {
    name: str,
    parameters: []Parameter,
}

struct Parameter {
    name: str,
    values: []f64,
}
```

**Wire Format Example:**
```
[2]                          // plugin count
  [str: "Reverb"]           // plugin[0].name
  [3]                        // plugin[0].parameters count
    [str: "wet"]            // param[0].name
    [2][0.5][0.8]           // param[0].values
    [str: "dry"]            // param[1].name
    [0]                      // param[1].values (empty)
    ...
  [str: "Delay"]            // plugin[1].name
  ...
```

### 2.6 Optional Fields Pattern

**Use arrays with 0 or 1 elements for optional struct fields:**

```rust
struct Snapshot {
    devices: []Device,
    engine_config: []EngineConfig,  // 0 or 1 element
}
```

**Usage (Go):**
```go
if len(snapshot.EngineConfig) > 0 {
    cfg := snapshot.EngineConfig[0]
    // Use config
}
```

**Rationale:** Reuses existing array semantics, no additional wire format features needed.

```

---

## 3. Release Candidate Features (0.2.0-rc1)

### 3.1 Optional Struct Fields

**Syntax:** `?Type` prefix indicates optional field

**Schema:**
```rust
struct Plugin {
    id: u32,           // Required
    name: string,      // Required
    metadata: ?Metadata,  // Optional
}

struct Metadata {
    version: string,
    author: string,
}
```

**Wire Format:**
```
[presence: u8][data: variable bytes]  // if presence = 1
[presence: u8]                        // if presence = 0 (no data follows)
```

**Generated Code (Go):**
```go
type Plugin struct {
    ID       uint32
    Name     string
    Metadata *Metadata  // nil if absent
}
```

**Encoding:**
- Present: Write `0x01`, then encode the struct normally
- Absent: Write `0x00`, done

**Decoding:**
- Read presence byte
- If `0x01`: allocate struct, decode into it
- If `0x00`: set pointer to nil

**Performance cost** (measured):
- Optional present: 31.49 ns decode (+48% vs required field)
- Optional absent: 3.15 ns decode (10× faster than present, zero allocation)
- Wire overhead: 1 byte per optional field

**Restrictions:**
- Only structs can be optional (no `?u32`, `?string`, `?[]T`)
- Cannot have optional primitives or optional arrays
- Cannot have arrays of optional items (`[]?Item`)

**Use cases:**
- Fields that may not be loaded yet
- Backward-compatible schema additions
- Recursive structures (linked lists, trees)

### 3.2 Message Mode (Self-Describing)

**Syntax:** Use `message` keyword instead of `struct`

**Schema:**
```rust
message ErrorMsg {
    code: u32,
    text: string,
}

message DataMsg {
    payload: []u8,
}
```

**Wire Format:**
```
[type_id: u64][payload_size: u32][payload: variable bytes]
```

**Type ID calculation:**
- FNV-1a hash of message name (e.g., `hash("ErrorMsg")`)
- 64-bit hash ensures collision resistance
- Deterministic across compilations

**Generated Code (Go):**
```go
// Constants for message type identification
const (
    ErrorMsgTypeID uint64 = 0x... // FNV-1a hash
    DataMsgTypeID  uint64 = 0x...
)

// Encoder includes type ID
func EncodeErrorMsg(src *ErrorMsg) ([]byte, error)

// Dispatcher routes by type ID
func DispatchMessage(data []byte) (interface{}, error) {
    typeID := binary.LittleEndian.Uint64(data[0:8])
    switch typeID {
    case ErrorMsgTypeID:
        var msg ErrorMsg
        err := DecodeErrorMsg(&msg, data)
        return &msg, err
    case DataMsgTypeID:
        // ...
    }
}
```

**Performance cost** (measured):
- Message overhead: 10 bytes header (8 type ID + 4 size)
- Roundtrip: 85.54 ns vs 44.25 ns regular mode (+93%)
- Size overhead: 19.6% for small payloads, negligible for large

**Use cases:**
- Persistent storage with multiple message types
- Event streams with heterogeneous events
- Protocol implementations needing type discrimination
- RPC-style request/response pairs

**When NOT to use:**
- Single message type (no discrimination needed)
- Performance-critical inner loops (use regular structs)

### 3.3 Streaming I/O

**Generated functions for stdlib composition:**

```go
// Every struct/message generates these functions
func EncodePluginToWriter(src *Plugin, w io.Writer) error
func DecodePluginFromReader(dest *Plugin, r io.Reader) error
```

**Implementation:**
- Encoder: Calculate size → allocate buffer → encode → `w.Write(buf)`
- Decoder: `io.ReadAll(r)` → decode from buffer
- Zero new dependencies (uses existing encode/decode functions)

**Composition examples:**

**File I/O:**
```go
file, _ := os.Create("data.sdp")
defer file.Close()
EncodePluginToWriter(&plugin, file)
```

**Compression:**
```go
var buf bytes.Buffer
gzipWriter := gzip.NewWriter(&buf)
EncodePluginToWriter(&plugin, gzipWriter)
gzipWriter.Close()
```

**Network:**
```go
conn, _ := net.Dial("tcp", "localhost:8080")
EncodePluginToWriter(&plugin, conn)
```

**Design philosophy:**
- SDP provides interfaces (io.Writer/Reader)
- Users compose with their choice of libraries
- No baked-in compression, file, or network code
- Maximum flexibility via Unix-style composition

---

## 4. Schema Definition

**File extension:** `.sdp` (Serial Data Protocol)

**Syntax:** Rust-like with SDP extensions

**Supported syntax:**
- Regular structs: `struct Name { ... }`
- Messages: `message Name { ... }` (self-describing)
- Optional fields: `field: ?Type` (structs only)
- Doc comments: `///` (attached to following declaration)
- Line comments: `//` (ignored)

**Example schema:**
```rust
// Regular struct (byte mode)
struct Plugin {
    id: u32,
    name: string,
    metadata: ?Metadata,  // Optional
}

struct Metadata {
    version: string,
    author: string,
}

// Self-describing message
message PluginEvent {
    timestamp: u64,
    plugin_id: u32,
    event_type: u8,
}
```

### 4.1 Schema Format (Legacy)

**File extension:** `.sdp` (Serial Data Protocol)

**Syntax:** **Rust subset** - `.sdp` files are valid Rust struct definitions

**Language Specification:**

.sdp` files use Rust syntax for struct definitions. The generator parses a strict subset of Rust:

### 3.1 Schema Format

**File extension:** `.sdp` (Serial Data Protocol)

**Syntax:** **Rust subset** - `.sdp` files are valid Rust struct definitions

**Language Specification:**

`.sdp` files use Rust syntax for struct definitions. The generator parses a strict subset of Rust:

**Supported Rust features:**
- Struct definitions: `struct Name { ... }`
- Field declarations: `field_name: Type`
- Doc comments: `///` (attached to following declaration)
- Line comments: `//` (ignored)
- Array types: `[]Type`
- Primitive types: `u8`, `u16`, `u32`, `u64`, `i8`, `i16`, `i32`, `i64`, `f32`, `f64`, `bool`, `str`
- Named types: References to other structs

**Rust features NOT supported in v1.0:**
- Generics, lifetimes, traits
- Enums, unions, type aliases
- Visibility modifiers (`pub`, `pub(crate)`, etc.)
- Attributes (`#[derive(...)]`, etc.)
- Block comments (`/* */`)
- Expressions, statements, functions
- Any non-struct items

**Lexical Rules:**

1. **Keywords:** Only `struct` is recognized as a keyword
2. **Field separators:** Comma `,` required after each field, **optional after last field** (Rust-style)
3. **Whitespace:** Not significant (spaces, tabs, newlines treated equally)
4. **Comments:**
   - `///` = Documentation comment (attached to next struct or field)
   - `//` = Regular comment (ignored by parser)
5. **Identifiers:** Must match `[a-zA-Z_][a-zA-Z0-9_]*`
6. **String type:** Use `str` (consistent with Rust's string slice type)

**Grammar (EBNF):**
```ebnf
Schema      = { Struct } ;
Struct      = [ DocComment ] "struct" Ident "{" [ FieldList ] "}" ;
FieldList   = Field { "," Field } [ "," ] ;
Field       = [ DocComment ] Ident ":" TypeExpr ;
TypeExpr    = Ident | "[" "]" TypeExpr ;
DocComment  = "///" text "\n" { "///" text "\n" } ;
Ident       = letter { letter | digit | "_" } ;
```

**Example:**
```rust
// plugin_list.sdp

/// PluginList contains all enumerated plugins.
struct PluginList {
    /// List of discovered plugins.
    plugins: []Plugin,
}

/// Plugin represents a single audio plugin.
struct Plugin {
    /// Unique plugin identifier.
    id: u32,
    
    /// Human-readable plugin name.
    name: str,
    
    /// Vendor/manufacturer name.
    vendor: str,
    
    /// Plugin parameters.
    parameters: []Parameter,
}

/// Parameter represents a controllable plugin parameter.
struct Parameter {
    /// Parameter name.
    name: str,
    
    /// Parameter type identifier.
    param_type: u8,
    
    /// Whether parameter can be automated.
    automatable: bool,
    
    /// Parameter values.
    values: []f64,
}
```

**Note:** Field `param_type` renamed from `type` since `type` is a Rust keyword.

**IDE Setup (VSCode):**
```json
{
  "files.associations": {
    "*.sdp": "rust"
  }
}
```

This enables Rust syntax highlighting, which works because `.sdp` files ARE valid Rust code.

### 3.2 Schema Compiler

**Input:** `*.sdp` files  
**Output:** Generated code per target language

**Go Output:**
- Struct definitions with doc comments
- `Decode(dest *T, data []byte) error` function
- Metadata (computed at `init()` via `unsafe`)

**C Output:**
- Builder API (`BeginT`, `SetT_Field`, `DiscardT` functions)
- Tentative struct definitions
- Helper functions for writing primitives

**Other Languages:**
- Rust: Similar to Go (structs + decode)
- Swift: Similar to Go (structs + decode)

### 3.3 Type Mapping

| Schema | Go            | C                          | Rust         | Swift        |
|--------|---------------|----------------------------|--------------|--------------|
| `u32`  | `uint32`      | `uint32_t`                 | `u32`        | `UInt32`     |
| `f64`  | `float64`     | `double`                   | `f64`        | `Double`     |
| `str`  | `string`      | `const char*` (param)      | `String`     | `String`     |
| `[]T`  | `[]T`         | `T* items; uint32_t count` | `Vec<T>`     | `[T]`        |

### 3.4 Naming Rules

**Schema identifiers (packages, types, fields) must:**
- Start with letter or underscore
- Contain only letters, digits, underscores
- Not be reserved words in Go, Rust, Swift, or C

**Reserved word validation:** Generator maintains list of reserved keywords across all target languages and rejects schemas using them.

**Style:** Generator does NOT enforce naming conventions (camelCase, snake_case, etc.). Use your language's conventions in schema files.

### 3.5 Schema Validation

**Generator validates schemas in two phases:**

**Phase 1: Parse all schemas**
- Syntax errors
- Malformed struct definitions

**Phase 2: Semantic validation**
- Type reference validation (unknown types within same file)
- Circular reference detection
- Empty struct detection
- Duplicate field names
- Reserved keyword usage (see section 3.5.1)

**Schema Composition Model:**

Each `.sdp` schema file is **self-contained** with all required type definitions. Common types can be **duplicated** across schema files when needed.

Example - Two independent schemas sharing common types:
```rust
// devices.sdp - Self-contained schema for device enumeration
struct DeviceList {
    devices: []Device,
}

struct Device {
    id: u32,
    name: str,
    parameters: []Parameter,  // Common type defined here
}

struct Parameter {
    name: str,
    value: f64,
}
```

```rust
// plugins.sdp - Self-contained schema for plugin enumeration
struct PluginList {
    plugins: []Plugin,
}

struct Plugin {
    id: u32,
    name: str,
    parameters: []Parameter,  // Same type, duplicated definition
}

struct Parameter {  // Duplicate definition (intentional)
    name: str,
    value: f64,
}
```

**Rationale:** No cross-schema references means:
- Each schema file generates independent code
- No build-time dependencies between schemas
- Application layer composes domain objects from multiple decoded schemas
- Empty arrays (`[]T` with 0 elements) provide optional field semantics

**Validation errors are collected and reported together** - generator does not stop at first error.

**Examples of rejected schemas:**

```rust
// ❌ Unknown type
struct Plugin {
    device: AudioDevice,  // AudioDevice not defined in this file
}

// ❌ Circular reference
struct Node {
    value: u32,
    next: Node,  // Direct self-reference
}

// ❌ Empty struct
struct Empty {
    // No fields
}

// ❌ Duplicate field
struct Plugin {
    id: u32,
    name: str,
    id: u64,  // Duplicate
}

// ❌ Reserved keyword
struct Plugin {
    type: u32,  // 'type' reserved in Go/Rust
}
```

#### 3.5.1 Reserved Keywords and Problematic Identifiers

**The generator rejects struct names and field names that would cause compilation errors, warnings, or ambiguous code in target languages.**

Validation includes:
- **Language keywords** - reserved by language specification (compilation errors)
- **Future reserved words** - reserved for future language versions (compilation errors)
- **Built-in types and functions** - would shadow standard library (warnings/errors)
- **Common attributes** - would conflict with language features (ambiguous code)

**Rationale:** Generated code must:
1. Compile without errors
2. Show no warnings
3. Be unambiguous (no shadowing, no confusing names)

This requires conservative validation against all identifiers that could cause problems in any target language.

**Go - Reserved and Built-in Identifiers:**
- **Keywords (25):** break, case, chan, const, continue, default, defer, else, fallthrough, for, func, go, goto, if, import, interface, map, package, range, return, select, struct, switch, type, var
- **Built-in types:** bool, byte, complex64, complex128, error, float32, float64, int, int8, int16, int32, int64, rune, string, uint, uint8, uint16, uint32, uint64, uintptr
- **Constants:** true, false, iota, nil
- **Built-in functions:** append, cap, close, complex, copy, delete, imag, len, make, new, panic, print, println, real, recover
- **Special identifiers:** main, init

**Rust - Keywords and Common Types:**
- **Strict keywords (35):** as, break, const, continue, crate, else, enum, extern, false, fn, for, if, impl, in, let, loop, match, mod, move, mut, pub, ref, return, self, Self, static, struct, super, trait, true, type, unsafe, use, where, while
- **Reserved keywords (15):** abstract, async, await, become, box, do, final, macro, override, priv, try, typeof, unsized, virtual, yield
- **Weak keywords (3):** union, dyn, raw
- **Common types:** Option, Result, Some, None, Ok, Err, String, Vec, Box, Rc, Arc
- **Common traits:** Copy, Clone, Send, Sync, Sized

**C - Keywords and Standard Types (C11/C23):**
- **Keywords (32):** auto, break, case, char, const, continue, default, do, double, else, enum, extern, float, for, goto, if, inline, int, long, register, restrict, return, short, signed, sizeof, static, struct, switch, typedef, union, unsigned, void, volatile, while
- **C11 extensions (10):** _Alignas, _Alignof, _Atomic, _Bool, _Complex, _Generic, _Imaginary, _Noreturn, _Static_assert, _Thread_local
- **C23 additions (4):** _BitInt, _Decimal32, _Decimal64, _Decimal128
- **Standard types:** bool, true, false, NULL, size_t, ptrdiff_t, wchar_t
- **Fixed-width integers:** int8_t, int16_t, int32_t, int64_t, uint8_t, uint16_t, uint32_t, uint64_t
- **Standard identifiers:** FILE, EOF

**Swift - Keywords, Attributes, and Common Types:**
- **Declaration keywords (24):** associatedtype, class, deinit, enum, extension, fileprivate, func, import, init, inout, internal, let, open, operator, private, precedencegroup, protocol, public, rethrows, static, struct, subscript, typealias, var
- **Statement keywords (17):** break, case, catch, continue, default, defer, do, else, fallthrough, for, guard, if, in, repeat, return, switch, throw, where, while
- **Expression keywords (9):** as, false, is, nil, self, Self, super, throws, true, try
- **Context keywords (6):** async, await, didSet, get, set, willSet
- **Declaration modifiers (11):** dynamic, final, lazy, optional, required, convenience, override, mutating, nonmutating, weak, unowned
- **Special identifiers (4):** _, Any, Type, Protocol
- **Compiler directives (16):** #available, #colorLiteral, #column, #else, #elseif, #endif, #error, #file, #fileLiteral, #fileID, #filePath, #function, #if, #imageLiteral, #line, #selector, #sourceLocation, #warning
- **Attributes (without @ prefix - 18):** available, objc, nonobjc, discardableResult, dynamicCallable, dynamicMemberLookup, escaping, autoclosure, convention, IBAction, IBOutlet, IBDesignable, IBInspectable, NSCopying, NSManaged, UIApplicationMain, NSApplicationMain, testable, warn_unqualified_access, frozen, unknown
- **Common types:** Int, Int8, Int16, Int32, Int64, UInt, UInt8, UInt16, UInt32, UInt64, Float, Double, Bool, String, Character, Array, Dictionary, Set, Optional, Error, Result

**Validation Rules:**
1. Struct names must not be reserved (case-insensitive)
2. Field names must not be reserved (case-insensitive)
3. Error messages indicate which language(s) reserve the identifier
4. Multiple errors collected and reported together

**Examples:**

```rust
// ❌ Rejected - struct name is Go/Rust/C/Swift keyword
struct type {
    id: u32,
}

// ❌ Rejected - field name would shadow Go built-in
struct Config {
    len: u32,  // 'len' is a Go built-in function
}

// ❌ Rejected - field name is Rust keyword
struct Settings {
    async: bool,  // 'async' reserved in Rust
}

// ❌ Rejected - would shadow Rust common type
struct Message {
    Result: u32,  // 'Result' is a common Rust type
}

// ❌ Rejected - Swift attribute name
struct Header {
    available: bool,  // 'available' is a Swift attribute
}

// ✓ Accepted - not reserved in any target language
struct Device {
    id: u32,
    name: str,
    enabled: bool,
}
```

## 4. Serialization (Encoder)

### 4.1 Builder Architecture

**Hierarchical tentative structure with flexible field ordering:**

```c
typedef struct tentative_base {
    tentative_type_t type;
    struct tentative_base* parent;        // Link up
    struct tentative_base* first_child;   // Link down
    struct tentative_base* next_sibling;  // Link sideways
    
    uint8_t* buffer;
    size_t buffer_size;
    size_t buffer_pos;
    
    int uncommitted_children;
    error_code_t error;
} tentative_base_t;
```

**Each struct type stores fields independently for flexible write order:**
```c
typedef struct {
    tentative_base_t base;
    
    // Per-field storage
    struct { bool written; uint32_t value; } id;
    struct { bool written; char* value; } name;
    struct { bool written; bool value; } active;
} tentative_plugin_t;
```

**Benefits:**
- Scalar fields can be written in any order
- Fields reordered to match schema order on commit
- Omitted fields use type defaults

### 4.2 Generated Builder API

**Per-schema generation with flexible field setters:**

```c
// Root builder
plugin_list_builder_t* NewPluginListBuilder(void);
serial_data_t FinalizeBuilder(plugin_list_builder_t* builder);
void DestroyBuilder(plugin_list_builder_t* builder);

// Struct builders
tentative_plugin_t* BeginPlugin(plugin_list_builder_t* parent);
void DiscardPlugin(plugin_list_builder_t* parent, tentative_plugin_t* plugin);

// Scalar field setters (any order)
void SetPluginID(tentative_plugin_t* plugin, uint32_t id);
void SetPluginName(tentative_plugin_t* plugin, const char* name);
void SetPluginVendor(tentative_plugin_t* plugin, const char* vendor);

// Array field builder (must be after scalars)
tentative_parameter_t* BeginParameter(tentative_plugin_t* parent);
void SetParameterName(tentative_parameter_t* param, const char* name);
void SetParameterType(tentative_parameter_t* param, uint8_t type);
void SetParameterAutomatable(tentative_parameter_t* param, bool automatable);

// Array element helpers
void AddParameterValue(tentative_parameter_t* param, double value);
```

**Usage pattern:**
```c
plugin_list_builder_t* builder = NewPluginListBuilder();

tentative_plugin_t* p = BeginPlugin(builder);

// Scalars in any order
SetPluginVendor(p, "Acme");
SetPluginID(p, 42);
SetPluginName(p, "Reverb");  // Order doesn't matter

// Arrays (schema order)
tentative_parameter_t* pm = BeginParameter(p);
SetParameterName(pm, "wet");
SetParameterType(pm, 1);
AddParameterValue(pm, 0.5);

serial_data_t result = FinalizeBuilder(builder);
```

### 4.3 Implicit Commit Semantics

**Key principle:** Commit happens when parent commits or when root finalizes.

**User writes:**
```c
tentative_plugin_t* p = BeginPlugin(builder);
SetPluginName(p, "Reverb");

tentative_parameter_t* pm = BeginParameter(p);
SetParameterName(pm, "wet");

// No explicit commit needed - implicit on Finalize

serial_data_t result = FinalizeBuilder(builder);
```

**Discard is explicit:**
```c
tentative_parameter_t* pm = BeginParameter(p);
SetParameterName(pm, "dry");

if (!should_keep) {
    DiscardParameter(p, pm);  // Explicit rejection
}
```

### 4.4 Field Write Order and Defaults

**Scalar fields can be written in any order:**
```c
tentative_plugin_t* p = BeginPlugin(builder);

// Write in discovery order
SetPluginVendor(p, get_vendor());   // Third field
SetPluginID(p, get_id());           // First field
SetPluginName(p, get_name());       // Second field

// On commit: reordered to schema definition order
```

**Omitted fields use type defaults:**
- `u32`, `u64`, `i32`, `i64`, `u16`, `i16`, `u8`, `i8` → 0
- `f32`, `f64` → 0.0
- `bool` → false
- `str` → empty string ""
- `[]T` → empty array (count 0)

**Example with omission:**
```c
tentative_plugin_t* p = BeginPlugin(builder);
SetPluginID(p, 42);
// name and vendor omitted → defaults to "" for both
```

**Array fields must be written in schema order after scalars:**
```c
// ✅ Correct
SetPluginID(p, 42);
SetPluginName(p, "Reverb");
BeginParameter(p);  // After all scalars

// ❌ Incorrect (implementation may reject)
BeginParameter(p);
SetPluginID(p, 42);  // Scalar after array started
```

### 4.5 Commit Algorithm

**On `FinalizeBuilder()`:**

1. Traverse tree depth-first (post-order)
2. For each tentative:
   - Commit all child tentatives first
   - Write scalar fields in schema definition order
   - Write array fields (already in schema order)
   - Copy to parent's buffer
   - Free tentative
3. Bubble up until root
4. Return root's buffer as serialized data

**Commit with field reordering:**
```c
void commit_tentative_plugin(tentative_plugin_t* p, uint8_t* parent_buffer) {
    // Write fields in schema order (not write order)
    
    // Field 1: id
    if (p->id.written) {
        write_u32(parent_buffer, p->id.value);
    } else {
        write_u32(parent_buffer, 0);  // Default
    }
    
    // Field 2: name
    if (p->name.written) {
        write_string(parent_buffer, p->name.value);
        free(p->name.value);
    } else {
        write_string(parent_buffer, "");  // Default
    }
    
    // Field 3: vendor
    if (p->vendor.written) {
        write_string(parent_buffer, p->vendor.value);
        free(p->vendor.value);
    } else {
        write_string(parent_buffer, "");  // Default
    }
    
    // Field 4: parameters (array - from children)
    uint32_t param_count = count_children_of_type(p, TYPE_PARAMETER);
    write_u32(parent_buffer, param_count);
    for_each_child(p, TYPE_PARAMETER, commit_child);
}
```

### 4.6 Discard Algorithm

**On `DiscardT(parent, tentative)`:**

1. Recursively free all children (cascade)
2. Free field storage (strings, etc.)
3. Decrement parent's uncommitted_children
4. Unlink from parent's child list
5. Free tentative struct

**Implementation:**
```c
void discard_subtree(tentative_base_t* node) {
    // Discard children first (post-order)
    tentative_base_t* child = node->first_child;
    while (child) {
        tentative_base_t* next = child->next_sibling;
        discard_subtree(child);
        child = next;
    }
    
    // Free field storage
    if (node->type == TYPE_PLUGIN) {
        tentative_plugin_t* p = (tentative_plugin_t*)node;
        if (p->name.written) free(p->name.value);
        if (p->vendor.written) free(p->vendor.value);
    }
    
    // Unlink from parent
    if (node->parent) {
        node->parent->uncommitted_children--;
        unlink_from_parent(node);
    }
    
    free(node);
}
```

### 4.7 Memory Limits

**Per-tentative buffer limits:**
```c
#define MAX_TENTATIVE_BUFFER_SIZE (2 * 1024 * 1024)  // 2MB
#define INITIAL_TENTATIVE_SIZE 256                    // 256 bytes
```

**Nesting depth limit:**
```c
#define MAX_NESTING_DEPTH 32

int get_nesting_depth(tentative_base_t* t) {
    int depth = 0;
    while (t->parent) {
        depth++;
        t = t->parent;
    }
    return depth;
}
```

**Buffer growth:**
```c
void ensure_capacity(tentative_base_t* t, size_t additional) {
    size_t required = t->buffer_pos + additional;
    
    if (required > MAX_TENTATIVE_BUFFER_SIZE) {
        t->error = ERR_TENTATIVE_TOO_LARGE;
        return;
    }
    
    if (required > t->buffer_size) {
        size_t new_size = t->buffer_size * 2;
        if (new_size < required) {
            new_size = required;
        }
        if (new_size > MAX_TENTATIVE_BUFFER_SIZE) {
            new_size = MAX_TENTATIVE_BUFFER_SIZE;
        }
        t->buffer = realloc(t->buffer, new_size);
        t->buffer_size = new_size;
    }
}
```

## 5. Deserialization (Decoder)

### 5.1 Generated Decoder API

**Go:**
```go
func Decode(dest *PluginList, data []byte) error
```

**Rust:**
```rust
fn decode(data: &[u8]) -> Result<PluginList, DecodeError>
```

### 5.2 Decode Algorithm

**Sequential parsing with pre-allocation:**

1. Validate total data size (must not exceed 128MB)
2. For each field in schema definition order:
   - If primitive: read bytes directly
   - If string: read length, validate, allocate, copy bytes
   - If array: read count, validate, allocate, decode elements
   - If struct: recurse
3. Track total elements allocated

**Example (Go):**
```go
func decodePlugin(data []byte, offset *int, ctx *DecodeContext) (Plugin, error) {
    var p Plugin
    
    // Field: id (u32)
    if *offset + 4 > len(data) {
        return Plugin{}, ErrUnexpectedEOF
    }
    p.ID = binary.LittleEndian.Uint32(data[*offset:])
    *offset += 4
    
    // Field: name (str)
    if *offset + 4 > len(data) {
        return Plugin{}, ErrUnexpectedEOF
    }
    nameLen := binary.LittleEndian.Uint32(data[*offset:])
    *offset += 4
    
    if *offset + int(nameLen) > len(data) {
        return Plugin{}, ErrUnexpectedEOF
    }
    p.Name = string(data[*offset:*offset+int(nameLen)])
    *offset += int(nameLen)
    
    // Field: vendor (str)
    // ... similar
    
    // Field: parameters ([]Parameter)
    if *offset + 4 > len(data) {
        return Plugin{}, ErrUnexpectedEOF
    }
    paramCount := binary.LittleEndian.Uint32(data[*offset:])
    *offset += 4
    
    if err := ctx.checkArraySize(paramCount); err != nil {
        return Plugin{}, err
    }
    
    p.Parameters = make([]Parameter, paramCount)
    for i := 0; i < int(paramCount); i++ {
        p.Parameters[i], err = decodeParameter(data, offset, ctx)
        if err != nil {
            return Plugin{}, err
        }
    }
    
    return p, nil
}
```

### 5.3 Size/Offset Metadata

**Challenge:** Different languages have different struct layouts.

**Solution:** Self-reporting at runtime via `init()` or static analysis.

**Purpose:** Debugging and potential optimizations (not used in decode logic).

**Go example:**
```go
var Metadata = struct {
    PluginSize      uintptr
    PluginAlignment uintptr
}{}

func init() {
    type probe struct {
        id         uint32
        name       string
        vendor     string
        parameters []Parameter
    }
    
    Metadata.PluginSize = unsafe.Sizeof(probe{})
    Metadata.PluginAlignment = unsafe.Alignof(probe{})
}
```

**Note:** Wire format is densely packed (no padding). Decoder reads sequentially and constructs native structs.

### 5.4 Error Handling

**Decoder validates:**
- ✅ Sufficient bytes remaining for read
- ✅ UTF-8 validity for strings (language-dependent: required in Rust, optional in Go)
- ✅ Array counts within limits
- ✅ Total elements allocated within limits

**Does NOT validate:**
- ❌ Struct type identity (no type tags in wire format)
- ❌ Schema version matching (not supported)
- ❌ Field presence (omitted fields use defaults)

**Error types:**
```go
var (
    ErrUnexpectedEOF      = errors.New("unexpected end of data")
    ErrInvalidUTF8        = errors.New("invalid UTF-8 string")
    ErrDataTooLarge       = errors.New("data exceeds 128MB limit")
    ErrArrayTooLarge      = errors.New("array count exceeds per-array limit")
    ErrTooManyElements    = errors.New("total elements exceed limit")
)
```

### 5.5 Size Limits

**Maximum serialized data size:** 128 MB

Decoder rejects data exceeding this limit at entry point.

**Maximum elements per array:** 1,000,000

**Maximum total elements:** 10,000,000

**Rationale:**
- Protects against malicious or corrupted data
- Prevents out-of-memory conditions
- Aligned with Protocol Buffers defaults
- Sufficient for bulk enumeration use cases

**Validation:**
```go
const (
    MaxSerializedSize = 128 * 1024 * 1024
    MaxArrayElements  = 1_000_000
    MaxTotalElements  = 10_000_000
)

type DecodeContext struct {
    totalElements int
}

func (ctx *DecodeContext) checkArraySize(count uint32) error {
    if count > MaxArrayElements {
        return ErrArrayTooLarge
    }
    
    ctx.totalElements += int(count)
    if ctx.totalElements > MaxTotalElements {
        return ErrTooManyElements
    }
    
    return nil
}

func Decode(dest *PluginList, data []byte) error {
    if len(data) > MaxSerializedSize {
        return ErrDataTooLarge
    }
    
    ctx := &DecodeContext{}
    offset := 0
    return decodePluginList(dest, data, &offset, ctx)
}
```

## 6. Wire Format Specification

### 6.1 Endianness

**Little-endian** for all multi-byte values.

**Rationale:** Most common architecture (x86, ARM in little-endian mode). Same-machine constraint means no conversion needed in practice.

### 6.2 Alignment

**No alignment padding in wire format.** Fields are densely packed.

**Decoder responsibility:** Unpack to properly aligned native structs.

### 6.3 Example Wire Format

**Schema:**
```
struct Plugin {
    id: u32
    name: str
    active: bool
}
```

**Data:**
```
Plugin { id: 42, name: "Reverb", active: true }
```

**Wire format (hex):**
```
2A 00 00 00              // id: 42 (u32, little-endian)
06 00 00 00              // name length: 6 (u32)
52 65 76 65 72 62        // name: "Reverb" (UTF-8)
01                       // active: true (bool)
```

**Total:** 15 bytes

### 6.4 Array Example

**Schema:**
```
struct DeviceList {
    devices: []u32
}
```

**Data:**
```
DeviceList { devices: [1, 2, 3] }
```

**Wire format (hex):**
```
03 00 00 00              // array count: 3
01 00 00 00              // devices[0]: 1
02 00 00 00              // devices[1]: 2
03 00 00 00              // devices[2]: 3
```

**Total:** 16 bytes

### 6.5 Empty Arrays

**Empty array is valid:**
```
00 00 00 00              // count: 0
                         // (no elements)
```

## 7. Memory Management

### 7.1 Encoder Side (C)

**Ownership:** Builder owns all tentative memory.

**Lifecycle:**
```c
builder_t* b = NewBuilder();         // Allocates root
tentative_t* t = BeginStruct(b);     // Allocates tentative
serial_data_t data = Finalize(b);    // Frees tentatives, returns buffer
// data.buffer owned by builder
DestroyBuilder(b);                   // Frees data.buffer
```

**Go copies immediately:**
```go
cData := C.FinalizeBuilder(builder)
defer C.DestroyBuilder(builder)
goData := C.GoBytes(unsafe.Pointer(cData.data), C.int(cData.len))
// Now goData is Go-managed, C can free
```

### 7.2 Decoder Side (Go)

**Ownership:** Decoder allocates Go-managed memory.

```go
var list PluginList
err := Decode(&list, data)
// list contains Go strings, slices - managed by GC
```

### 7.3 Tentative Growth

**Dynamic buffer growth:**
```c
void EnsureCapacity(tentative_base_t* t, size_t additional) {
    size_t required = t->buffer_pos + additional;
    if (required > t->buffer_size) {
        size_t new_size = t->buffer_size * 2;
        if (new_size < required) {
            new_size = required;
        }
        t->buffer = realloc(t->buffer, new_size);
        t->buffer_size = new_size;
    }
}
```

**Initial sizes:** 256 bytes per tentative (tunable).

## 8. Error Handling

### 8.1 Encoder Errors

**Strategy:** Sticky error flag, operations no-op after error.

```c
typedef struct {
    tentative_base_t root;
    error_code_t error;
    char error_msg[256];
} builder_t;

void SetPluginName(tentative_plugin_t* p, const char* name) {
    if (p->base.error != ERR_NONE) {
        return;  // Already in error state, no-op
    }
    
    if (strlen(name) > MAX_STRING_LENGTH) {
        p->base.error = ERR_STRING_TOO_LONG;
        snprintf(p->base.error_msg, sizeof(p->base.error_msg),
                 "string too long: %zu bytes", strlen(name));
        return;
    }
    
    p->name.value = strdup(name);
    if (!p->name.value) {
        p->base.error = ERR_OUT_OF_MEMORY;
        return;
    }
    p->name.written = true;
}

serial_data_t FinalizeBuilder(builder_t* b) {
    if (b->root.error != ERR_NONE) {
        return (serial_data_t){
            .data = NULL,
            .len = 0,
            .error = b->root.error,
            .error_msg = b->root.error_msg
        };
    }
    
    // Normal finalize...
}
```

**Error codes:**
```c
typedef enum {
    ERR_NONE = 0,
    ERR_OUT_OF_MEMORY,
    ERR_TENTATIVE_TOO_LARGE,
    ERR_NESTING_TOO_DEEP,
    ERR_STRING_TOO_LONG,
} error_code_t;
```

### 8.2 Decoder Errors

**Strategy:** Return error immediately, no partial results.

Covered in Section 5.4 and 5.5.

## 9. Implementation Phases

### Phase 1: Core Infrastructure
- [ ] Schema parser (`.sdp` → AST)
- [ ] Type system validation
- [ ] Wire format specification finalized

### Phase 2: C Encoder
- [ ] Tentative structure implementation
- [ ] Builder API code generation
- [ ] Implicit commit/explicit discard logic
- [ ] Array count handling

### Phase 3: Go Decoder
- [ ] Struct generation
- [ ] Decode function generation
- [ ] `init()` self-reporting
- [ ] Error handling

### Phase 4: Testing
- [ ] Unit tests (see Testing Strategy)
- [ ] Integration tests (C encoder → Go decoder)
- [ ] Fuzzing (malformed data)
- [ ] Performance benchmarks

### Phase 5: Additional Languages
- [ ] Rust encoder/decoder
- [ ] Swift encoder/decoder
- [ ] Language-specific idioms

## 10. Performance Considerations

### 10.1 Encoder Performance

**Optimizations:**
- Pre-allocate tentative buffers (256B default)
- Geometric growth on realloc (2x)
- Single memory copy per commit (tentative → parent)
- No intermediate serialization step

**Expected overhead:**
- Struct allocation: ~100ns per tentative
- Buffer growth: amortized O(1)
- Commit copy: memcpy, O(n) in data size

### 10.2 Decoder Performance

**Optimizations:**
- Pre-allocate arrays with known count
- Single-pass decode (no validation pass)
- Direct memory reads (no buffering)

**Expected overhead:**
- Array allocation: `make([]T, count)` per array
- String allocation: one per string
- Struct allocation: stack or inline

### 10.3 CGO Overhead

**Single FFI call per logical unit:**
```
C.EnumeratePlugins() → returns serialized blob (one call)
```

**Not:**
```
for each plugin:
    C.GetPlugin() → many calls
```

**CGO call cost:** ~6ns on M1 (acceptable for bulk transfer).

## 11. Best Practices

### 11.1 Schema Evolution

**Field order matters.** Wire format depends on field definition order in schema.

**Breaking changes:**
- Reordering fields ❌
- Changing field types ❌
- Removing fields ❌

**Non-breaking changes:**
- Renaming fields ✅ (wire format has no field names)
- Adding doc comments ✅

**Note:** Schemas are not versioned. Any structural change breaks compatibility between encoder and decoder.

### 11.2 Thread Safety

**Builder API is NOT thread-safe.** Use one of these patterns:

**Pattern 1: One builder per thread**
```c
dispatch_apply(count, queue, ^(size_t i) {
    builder_t* thread_builder = NewBuilder();
    // ... build
    collect_result(FinalizeBuilder(thread_builder));
});
```

**Pattern 2: Parallel collect, serial serialize (recommended)**
```c
// Parallel: collect raw data
dispatch_apply(count, queue, ^(size_t i) {
    raw_data[i] = expensive_enumeration(i);
});

// Serial: build from collected data
builder_t* builder = NewBuilder();
for (int i = 0; i < count; i++) {
    build_from_raw_data(builder, raw_data[i]);
}
```

### 11.3 Extension Pattern

**Extend generated types via embedding (Go example):**

```go
// Generated (don't modify)
package audio

type Device struct {
    ID   uint32
    Name string
}

// Hand-written extension
package audio

type ExtendedDevice struct {
    Device  // Embed generated type
    
    // Additional application fields
    IsActive bool
}

func (d ExtendedDevice) DisplayName() string {
    return fmt.Sprintf("%s (%s)", d.Name, status(d.IsActive))
}
```

**Decoder output → convert to extended type:**
```go
var deviceList audio.DeviceList
audio.Decode(&deviceList, data)

extended := make([]ExtendedDevice, len(deviceList.Devices))
for i, dev := range deviceList.Devices {
    extended[i] = ExtendedDevice{
        Device:   dev,
        IsActive: checkIfActive(dev.ID),
    }
}
```

### 11.4 UTF-8 Handling

**Encoder:** Trust native code to produce valid UTF-8.

**Decoder:** 
- Go: Trust (accepts invalid UTF-8, though not recommended)
- Rust: Validates (String requires valid UTF-8)
- Swift: Trust (String handles conversion)
- C: Trust (bytes are bytes)

**If encoding non-UTF-8:** Use `[]u8` instead of `str`.

### 11.5 Limitations

**Current version limitations:**
- No schema versioning or evolution
- No optional field syntax (use arrays with 0/1 elements)
- No enums (use u8/u16 with constants in comments)
- No maps/dictionaries (use parallel arrays or structs)
- Cross-schema references not supported

**Workarounds:**

**Optional fields:** Use `[]Type` with 0 or 1 element.

**Enums:** Use primitives with constants.
```rust
/// Status represents device state.
/// 0 = Inactive, 1 = Active, 2 = Error
struct Device {
    status: u8,
}
```

**Maps:** Use struct with parallel arrays.
```rust
struct StringMap {
    keys: []str,
    values: []str,
}
```

---

**End of Design Specification**
