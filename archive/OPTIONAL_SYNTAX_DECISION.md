# Optional Syntax Analysis: Language Conventions

## The Question

Should we use `?Type` or `*Type` for optionals, considering our `.sdp` files should be valid/familiar across languages?

---

## Language-by-Language Analysis

### Rust: `Option<T>`

**Canonical Rust**:
```rust
struct Plugin {
    name: String,
    metadata: Option<Metadata>,  // Canonical Rust
}
```

**Convention**: `Option<T>` wraps the type, `?` is for error handling (not optionals)

### Swift: `T?`

**Canonical Swift**:
```swift
struct Plugin {
    var name: String
    var metadata: Metadata?  // Canonical Swift - postfix ?
}
```

**Convention**: Postfix `?` for optionals (e.g., `String?`, `Int?`)

### TypeScript: `T | null` or `T?`

**Canonical TypeScript**:
```typescript
interface Plugin {
    name: string;
    metadata?: Metadata;  // Canonical TypeScript - postfix ?
    // OR
    metadata: Metadata | null;
}
```

**Convention**: Either `?:` (optional property) or `| null` (nullable type)

### Kotlin: `T?`

**Canonical Kotlin**:
```kotlin
data class Plugin(
    val name: String,
    val metadata: Metadata?  // Canonical Kotlin - postfix ?
)
```

**Convention**: Postfix `?` for nullable types

### C#: `T?` (value types) or nullable reference types

**Canonical C#**:
```csharp
struct Plugin {
    public string Name;
    public Metadata? Metadata;  // Canonical C# - postfix ?
}
```

**Convention**: Postfix `?` for nullable (C# 8.0+)

### Go: `*T`

**Canonical Go**:
```go
type Plugin struct {
    Name     string
    Metadata *Metadata  // Canonical Go - pointer
}
```

**Convention**: Pointers indicate nullable (nil)

### C/C++: `T*`

**Canonical C/C++**:
```c
struct Plugin {
    char *name;
    Metadata *metadata;  // Canonical C - pointer
};
```

**Convention**: Pointers indicate nullable (NULL)

### Python: `Optional[T]`

**Canonical Python (with type hints)**:
```python
from typing import Optional

class Plugin:
    name: str
    metadata: Optional[Metadata]  # Canonical Python
```

**Convention**: `Optional[T]` from typing module

### Zig: `?T`

**Canonical Zig**:
```zig
const Plugin = struct {
    name: []const u8,
    metadata: ?Metadata,  // Canonical Zig - prefix ?
};
```

**Convention**: Prefix `?` for optional types

---

## Frequency Count Across Languages

| Syntax | Languages | Count |
|--------|-----------|-------|
| **Postfix `?`** | Swift, TypeScript, Kotlin, C# | **4** |
| **Prefix `?`** | Zig | **1** |
| **Pointer `*`** | Go, C, C++ | **3** |
| **Wrapper type** | Rust: `Option<T>`, Python: `Optional[T]` | **2** |

**Most common**: Postfix `?` (4 languages)

---

## Analysis for SDP Schema Syntax

### Option 1: Postfix `?` (Most Common)

```
struct Plugin {
    name: str,
    metadata: Metadata?,  // <-- Postfix question mark
}
```

**Pros**:
- ✅ Familiar to Swift, TypeScript, Kotlin, C# developers (most popular modern languages)
- ✅ Reads naturally: "Metadata, optional"
- ✅ Similar to array syntax: `[]Type` (bracket after type)

**Cons**:
- ❌ Different from Rust (would be `Option<Metadata>`)
- ❌ Unfamiliar to Go developers (would be `*Metadata`)
- ❌ Parser slightly more complex (postfix operator)

### Option 2: Prefix `?` (Zig-style)

```
struct Plugin {
    name: str,
    metadata: ?Metadata,  // <-- Prefix question mark
}
```

**Pros**:
- ✅ Matches Zig
- ✅ Symmetric with error handling operators (`!`)
- ✅ Easier to parse (prefix operator)
- ✅ Consistent position with array: `[]Type` vs `?Type` (both prefix)

**Cons**:
- ❌ Unfamiliar to most developers
- ❌ Only Zig uses this

### Option 3: Pointer `*` (C/Go-style)

```
struct Plugin {
    name: str,
    metadata: *Metadata,  // <-- Pointer asterisk
}
```

**Pros**:
- ✅ Familiar to C, C++, Go developers
- ✅ Direct mapping to Go output (`*Metadata`)
- ✅ Clear semantics: "pointer = nullable"

**Cons**:
- ❌ Confusing: "pointer" in schema doesn't mean pointer in all target languages
- ❌ In Rust, this would generate `Option<Box<Metadata>>` (not `&Metadata`)
- ❌ Implies memory model, but SDP is transport format
- ❌ Ambiguous: what about `*u32`? Is that allowed?

### Option 4: Wrapper Type (Rust-style)

```
struct Plugin {
    name: str,
    metadata: Optional<Metadata>,  // <-- Wrapper type
}
```

**Pros**:
- ✅ Familiar to Rust, Python developers
- ✅ Very explicit
- ✅ Extensible (could add `Result<T, E>` later)

**Cons**:
- ❌ Verbose
- ❌ Looks like generics, but SDP doesn't have generics
- ❌ Parser more complex (generic-like syntax)

---

## Rust Validity Concern

### You said: "Our .sdp files are to be valid Rust"

**Question**: Do you mean:
1. **.sdp syntax should *resemble* Rust** (feel familiar to Rust devs)
2. **.sdp files should *literally* parse as Rust** (copy-paste into Rust)

### If Literally Valid Rust

Then we MUST use Rust syntax exactly:
```rust
// This would need to be valid Rust
struct Plugin {
    name: String,
    metadata: Option<Metadata>,
}
```

**Problem**: This forces very Rust-specific choices:
- `String` instead of `str`
- `Option<T>` instead of `?T`
- `Vec<T>` instead of `[]T`
- Can't have `u8, u16, u32, u64` directly (Rust has these, but inconsistent with `str`)

**This would make it LESS familiar to Swift/TypeScript/Kotlin/C# developers**

### If Just Rust-Inspired

Then we can use common patterns:
```
struct Plugin {
    name: str,      // Not valid Rust (would be String)
    metadata: ?Metadata,  // Not valid Rust (would be Option<Metadata>)
}
```

This is **inspired** by Rust but optimized for clarity and cross-language familiarity.

---

## Recommendation

### ✅ Use **Prefix `?`** (Option 2)

**Syntax**:
```
struct Plugin {
    name: str,
    metadata: ?Metadata,
    fallback: ?Plugin,
}
```

**Reasoning**:

1. **Consistency with arrays**: Both are prefix
   ```
   array_field: []Type,
   optional_field: ?Type,
   ```

2. **Parser simplicity**: Prefix operators are easier to parse
   ```
   type = '?' named_type
        | '[' ']' named_type
        | named_type
   ```

3. **Semantic clarity**: `?` universally means "maybe/optional"
   - Used in regex, SQL, and many languages
   - Familiar to developers even if exact syntax differs

4. **Language mappings are clear**:
   ```
   SDP: ?Metadata
   
   → Go:         *Metadata
   → Rust:       Option<Metadata>
   → Swift:      Metadata?
   → TypeScript: Metadata | null
   → C:          Metadata*
   → Zig:        ?Metadata
   ```

5. **Not trying to be valid Rust**: SDP is a cross-language format
   - Similar to how Protocol Buffers has its own syntax
   - Similar to how Thrift has its own syntax
   - Optimized for clarity, not Rust validity

### Validation Rules

```
✅ Allowed:
  field: ?StructType
  field: ?NestedStruct

❌ Not allowed:
  field: ?u32          // No optional primitives
  field: ?str          // No optional strings
  field: []?Type       // No arrays of optionals
  field: ?[]Type       // Use empty array instead
```

---

## Wire Format (Same Regardless of Syntax)

**All options encode the same way**:
```
[presence: u8][data: variable]

presence = 0: field is absent (nil/null/None)
presence = 1: field is present, data follows
```

**Examples**:
```
// Plugin{name: "Reverb", metadata: Some(Metadata{version: 2})}
[name: "Reverb"]
[01]              // metadata present
[02 00 00 00]     // version = 2

// Plugin{name: "Mute", metadata: None}
[name: "Mute"]
[00]              // metadata absent (no more bytes)
```

---

## Language Implementation Guide Updates

### Rust
```rust
// SDP schema
struct Plugin {
    name: str,
    metadata: ?Metadata,
}

// Generated Rust
pub struct Plugin {
    pub name: String,
    pub metadata: Option<Metadata>,
}
```

### Swift
```swift
// SDP schema
struct Plugin {
    name: str,
    metadata: ?Metadata,
}

// Generated Swift
public struct Plugin {
    public var name: String
    public var metadata: Metadata?
}
```

### Go
```go
// SDP schema
struct Plugin {
    name: str,
    metadata: ?Metadata,
}

// Generated Go
type Plugin struct {
    Name     string
    Metadata *Metadata
}
```

### C
```c
// SDP schema
struct Plugin {
    name: str,
    metadata: ?Metadata,
}

// Generated C
typedef struct {
    char *name;      // Heap-allocated
    Metadata *metadata;  // NULL if absent
} Plugin;
```

---

## Alternative: Support Multiple Syntaxes (Not Recommended)

We *could* support multiple syntaxes:
```
struct Plugin {
    metadata1: ?Metadata,      // Prefix ? (Zig-style)
    metadata2: Metadata?,      // Postfix ? (Swift-style)
    metadata3: *Metadata,      // Pointer (Go-style)
    metadata4: Option<Metadata>,  // Wrapper (Rust-style)
}
```

**All parse to the same AST**: `Type{Optional: true, Name: "Metadata"}`

**Pros**:
- ✅ Maximum flexibility
- ✅ Each developer can use familiar syntax

**Cons**:
- ❌ Inconsistent codebases (team uses different styles)
- ❌ Parser complexity
- ❌ Confusing for newcomers
- ❌ More documentation needed

**Verdict**: ❌ **Not recommended** - pick one canonical syntax

---

## Final Recommendation

### ✅ **Prefix `?` Syntax**

```
struct Plugin {
    name: str,
    metadata: ?Metadata,
    config: ?Config,
    fallback: ?Plugin,
}
```

**Reasons**:
1. ✅ Consistent with array syntax (`[]Type`)
2. ✅ Easier to parse (prefix operator)
3. ✅ Universally understood (`?` = "maybe")
4. ✅ Clean mapping to all target languages
5. ✅ Not tied to any single language's syntax

**Implementation**: 3 hours (as analyzed previously)

**Documentation**: Update LANGUAGE_IMPLEMENTATION_GUIDE.md with optional field mappings for each target language

---

## Summary Table

| Syntax | Example | Pros | Cons | Verdict |
|--------|---------|------|------|---------|
| **Prefix `?`** | `?Metadata` | ✅ Consistent with arrays<br>✅ Easy to parse<br>✅ Universal symbol | ⚠️ Only Zig uses this exactly | ✅ **RECOMMENDED** |
| **Postfix `?`** | `Metadata?` | ✅ Most common (4 languages) | ❌ Inconsistent with `[]Type`<br>❌ Harder to parse | ⚠️ Alternative |
| **Pointer `*`** | `*Metadata` | ✅ Familiar to C/Go | ❌ Implies memory model<br>❌ Confusing in Rust/Swift | ❌ Not recommended |
| **Wrapper** | `Option<Metadata>` | ✅ Very explicit | ❌ Verbose<br>❌ Looks like generics | ❌ Not recommended |

**Winner**: **Prefix `?`** for consistency, simplicity, and universal understanding.
