# Optional Syntax: Maintaining Valid Rust Format

## Current Requirement (from DESIGN_SPEC.md)

> **Syntax:** Rust subset - `.sdp` files are valid Rust struct definitions
> This enables Rust syntax highlighting, which works because `.sdp` files ARE valid Rust code.

**Goal**: Leverage existing Rust tooling (syntax highlighting, rust-analyzer, rustfmt) for free.

---

## The Problem: Adding Optionals While Staying Valid Rust

We need optional struct fields that:
1. ✅ Parse as valid Rust
2. ✅ Get syntax highlighting for free
3. ✅ Can be formatted with `rustfmt`
4. ✅ Can be checked with `rust-analyzer`
5. ✅ Map cleanly to Go (`*T`), Swift (`T?`), C (`T*`)

---

## Option 1: Use Rust's `Option<T>` (Stay 100% Rust-Compatible)

### Schema Syntax

```rust
struct Plugin {
    name: String,
    metadata: Option<Metadata>,
    fallback: Option<Box<Plugin>>,  // Recursive requires Box
}

struct Metadata {
    version: u32,
}
```

### Pros
- ✅ **100% valid Rust** - can copy-paste into Rust project
- ✅ **Free tooling** - rust-analyzer, rustfmt, syntax highlighting all work perfectly
- ✅ **Familiar to Rust developers**
- ✅ **Clear semantics** - `Option` is universally understood

### Cons
- ❌ **Verbose** - `Option<Metadata>` vs `?Metadata`
- ❌ **Type mapping complexity**:
  - Need to parse `Option<T>` as generic
  - Need to handle `Box<T>` for recursion
  - Parser becomes more complex
- ❌ **String type inconsistency**:
  - Currently use `str` in SDP
  - Rust would require `String` (owned) for valid syntax
  - `str` is only valid as `&str` (lifetime required)

### Parser Impact

Need to support:
```rust
// Parser must handle:
Type       = Ident
           | "Option" "<" Type ">"
           | "Box" "<" Type ">"
           | "[" "]" Type
           ;
```

**Complexity**: Medium (generic-like parsing)

---

## Option 2: Use Type Aliases (Valid Rust via Prelude)

### Approach

Provide a standard "prelude" that makes `.sdp` files valid Rust:

**File: `sdp_prelude.rs`** (shipped with SDP)
```rust
// SDP Prelude - makes .sdp files valid Rust
#![allow(dead_code)]

pub type str = String;  // Map str to String
pub type Opt<T> = Option<T>;  // Shorter optional
```

**Schema: `plugin.sdp`**
```rust
use sdp_prelude::*;

struct Plugin {
    name: str,          // Maps to String
    metadata: Opt<Metadata>,  // Maps to Option<Metadata>
    fallback: Opt<Box<Plugin>>,
}

struct Metadata {
    version: u32,
}
```

### Validation

```bash
# .sdp file is valid Rust!
$ rustc --crate-type lib plugin.sdp
# Works!

$ rust-analyzer check plugin.sdp
# Works!

$ rustfmt plugin.sdp
# Works!
```

### Pros
- ✅ **Valid Rust** with prelude import
- ✅ **Free tooling** - all Rust tools work
- ✅ **Shorter syntax** - `Opt<T>` vs `Option<T>`
- ✅ **Consistent with current SDP** - keeps `str` type

### Cons
- ⚠️ **Requires prelude file** - users need `sdp_prelude.rs`
- ⚠️ **Import line needed** - `use sdp_prelude::*;`
- ❌ **Still verbose** - `Opt<Metadata>` vs `?Metadata`

### Parser Impact

Same as Option 1 - need generic parsing.

---

## Option 3: Use Comments + Macros (Hack, but Works)

### Approach

Use Rust macros to make syntax valid:

**File: `sdp_macros.rs`**
```rust
macro_rules! opt {
    ($t:ty) => { Option<$t> };
}

macro_rules! str {
    () => { String };
}
```

**Schema: `plugin.sdp`**
```rust
#[macro_use]
extern crate sdp_macros;

struct Plugin {
    name: str!(),
    metadata: opt!(Metadata),
}
```

### Pros
- ✅ Valid Rust with macro expansion

### Cons
- ❌ **Ugly syntax** - `opt!(Metadata)`, `str!()`
- ❌ **Macro expansion required** - doesn't work out of box
- ❌ **Confusing** - macros hide what's happening

**Verdict**: ❌ Too hacky

---

## Option 4: Attribute-Based Optionals (Valid Rust Pattern)

### Approach

Use Rust attributes to mark optional fields:

**Schema: `plugin.sdp`**
```rust
struct Plugin {
    name: String,
    
    #[sdp(optional)]
    metadata: Metadata,
    
    #[sdp(optional)]
    fallback: Plugin,
}
```

### SDP Parser

Parses attributes and marks field as optional in AST:
```go
type Field struct {
    Name     string
    Type     Type
    Optional bool  // Set to true if #[sdp(optional)] present
    DocComment string
}
```

### Generated Code

**Go**:
```go
type Plugin struct {
    Name     string
    Metadata *Metadata  // Pointer because optional
    Fallback *Plugin
}
```

**Rust**:
```rust
pub struct Plugin {
    pub name: String,
    pub metadata: Option<Metadata>,
    pub fallback: Option<Box<Plugin>>,
}
```

### Pros
- ✅ **Valid Rust** - attributes are standard Rust
- ✅ **Free tooling** - rust-analyzer, rustfmt work
- ✅ **Clear intent** - `#[sdp(optional)]` is explicit
- ✅ **Extensible** - can add more attributes later
- ✅ **Type is the actual type** - `metadata: Metadata` (not wrapped)

### Cons
- ⚠️ **Slightly verbose** - attribute on separate line
- ⚠️ **Parser complexity** - need to parse attributes
- ⚠️ **Rust compilation requires proc macro** - users need `sdp_derive` crate

### Parser Impact

```rust
// Parser must handle:
Field = { Attribute } Ident ":" Type ;
Attribute = "#" "[" Ident [ "(" Args ")" ] "]" ;
```

**Complexity**: Low-Medium (attributes are simple to parse)

### Rust Validation Setup

**File: `Cargo.toml`** (for Rust validation only)
```toml
[dependencies]
sdp_derive = "0.1"  # Provides #[sdp(...)] attributes
```

**The proc macro does nothing** - it's just for Rust validation:
```rust
// sdp_derive/src/lib.rs
use proc_macro::TokenStream;

#[proc_macro_attribute]
pub fn sdp(_attr: TokenStream, item: TokenStream) -> TokenStream {
    // Pass through unchanged - only for syntax validation
    item
}
```

---

## Option 5: Switch to Zig Format (Breaking Change)

### Justification

Zig has:
- ✅ Simpler syntax than Rust
- ✅ Built-in optional syntax: `?Type`
- ✅ No generics complexity
- ✅ Structs are simpler (no visibility, traits, etc.)

### Schema Syntax (Zig)

```zig
const Plugin = struct {
    name: []const u8,
    metadata: ?Metadata,
    fallback: ?*Plugin,  // Recursive
};

const Metadata = struct {
    version: u32,
};
```

### Pros
- ✅ **Prefix `?` syntax** - exactly what we wanted
- ✅ **Free tooling** - Zig LSP, syntax highlighting
- ✅ **Simpler than Rust** - no generics, lifetimes, etc.
- ✅ **Clean optional syntax** - `?Type` is built-in

### Cons
- ❌ **Breaking change** from spec
- ❌ **Less popular than Rust** - fewer users familiar
- ⚠️ **String type** - Zig uses `[]const u8` (byte slices), not `str`

### Parser Impact

**Much simpler** than Rust:
```zig
Type = [ "?" ] TypeExpr ;
TypeExpr = Ident | "[" "]" TypeExpr | "*" TypeExpr ;
```

No generics, no lifetimes, no attributes.

---

## Option 6: Custom Format with Language Server (Nuclear Option)

### Approach

Create `.sdp` syntax that's NOT valid Rust, but provide:
1. **Language Server Protocol (LSP)** for VSCode/editors
2. **Tree-sitter grammar** for syntax highlighting
3. **Formatter** (like `sdp fmt`)

### Schema Syntax

```
struct Plugin {
    name: str,
    metadata: ?Metadata,
    fallback: ?Plugin,
}
```

### Tooling Provided

**VSCode Extension**:
```json
{
  "name": "sdp-language-support",
  "contributes": {
    "languages": [{
      "id": "sdp",
      "extensions": [".sdp"],
      "configuration": "./language-configuration.json"
    }],
    "grammars": [{
      "language": "sdp",
      "scopeName": "source.sdp",
      "path": "./syntaxes/sdp.tmLanguage.json"
    }]
  }
}
```

**Tree-sitter grammar** for syntax highlighting  
**LSP server** for errors, autocomplete, go-to-definition

### Pros
- ✅ **Perfect syntax** - design exactly what we want
- ✅ **Professional tooling** - full IDE support
- ✅ **No compromises** - not limited by Rust syntax

### Cons
- ❌ **Massive effort** - LSP + tree-sitter + VSCode extension
- ❌ **Maintenance burden** - must maintain tooling
- ❌ **User setup** - users must install extension
- ❌ **Breaking from spec**

**Verdict**: ❌ Overkill for this project

---

## Comparison Matrix

| Option | Valid Rust? | Free Tooling? | Optional Syntax | Parser Complexity | Breaking Change? |
|--------|-------------|---------------|----------------|-------------------|------------------|
| **1. `Option<T>`** | ✅ Yes | ✅ Yes | `Option<Metadata>` | Medium (generics) | ⚠️ Must use `String` not `str` |
| **2. Type Aliases** | ✅ With prelude | ✅ Yes | `Opt<Metadata>` | Medium (generics) | ⚠️ Requires prelude |
| **3. Macros** | ✅ With crate | ⚠️ Partial | `opt!(Metadata)` | Medium | ❌ Ugly |
| **4. Attributes** | ✅ With proc macro | ✅ Yes | `#[sdp(optional)]<br>metadata: Metadata` | Low-Medium | ❌ No |
| **5. Switch to Zig** | ✅ Yes (Zig) | ✅ Yes (Zig) | `?Metadata` | Low | ✅ Yes (major) |
| **6. Custom LSP** | ❌ No | ⚠️ Custom | `?Metadata` | Low | ✅ Yes (major) |

---

## Recommended Solution

### ✅ **Option 4: Attribute-Based Optionals**

**Schema: `plugin.sdp`**
```rust
/// Plugin metadata information
struct Metadata {
    version: u32,
    author: String,
}

/// Audio plugin descriptor
struct Plugin {
    /// Plugin name
    name: String,
    
    /// Optional metadata (might not be available)
    #[sdp(optional)]
    metadata: Metadata,
    
    /// Optional fallback plugin
    #[sdp(optional)]
    fallback: Plugin,
}
```

### Why This is Best

1. **✅ Valid Rust** - attributes are standard Rust feature
2. **✅ Free tooling** - rust-analyzer, rustfmt, clippy all work
3. **✅ No breaking change** - still Rust syntax
4. **✅ Clear semantics** - `#[sdp(optional)]` is self-documenting
5. **✅ Extensible** - can add more attributes:
   ```rust
   #[sdp(optional)]
   #[sdp(deprecated)]
   #[sdp(default = "1.0")]
   value: f32,
   ```

### Generated Code Examples

**Go**:
```go
type Plugin struct {
    Name     string
    Metadata *Metadata  // Pointer = optional
    Fallback *Plugin
}
```

**Rust**:
```go
pub struct Plugin {
    pub name: String,
    pub metadata: Option<Metadata>,
    pub fallback: Option<Box<Plugin>>,
}
```

**Swift**:
```swift
public struct Plugin {
    public var name: String
    public var metadata: Metadata?
    public var fallback: Plugin?
}
```

**C**:
```c
typedef struct {
    char *name;
    Metadata *metadata;  // NULL if absent
    Plugin *fallback;
} Plugin;
```

### Wire Format (Same as Before)

```
[presence: u8][data if present]
```

### Parser Changes (~100 lines)

```go
// internal/parser/lexer.go
const (
    TokenHash         // '#'
    TokenLeftBracket  // '['
    TokenRightBracket // ']'
    // ... existing tokens
)

// internal/parser/ast.go
type Field struct {
    Name       string
    Type       Type
    Attributes []Attribute  // NEW
    DocComment string
}

type Attribute struct {
    Name string      // "sdp"
    Args []string    // ["optional"]
}

// internal/parser/parser.go
func (p *Parser) parseAttributes() ([]Attribute, error) {
    var attrs []Attribute
    for p.current.Type == TokenHash {
        // Parse #[name(args)]
        attr := p.parseAttribute()
        attrs = append(attrs, attr)
    }
    return attrs, nil
}

func (p *Parser) parseField() (*Field, error) {
    // Parse attributes first
    attrs, _ := p.parseAttributes()
    
    // Parse field name and type
    name := p.current.Value
    p.advance()
    p.expect(TokenColon)
    typ := p.parseType()
    
    // Check for #[sdp(optional)]
    optional := false
    for _, attr := range attrs {
        if attr.Name == "sdp" && contains(attr.Args, "optional") {
            optional = true
            break
        }
    }
    
    return &Field{
        Name: name,
        Type: Type{..., Optional: optional},
        Attributes: attrs,
    }, nil
}
```

### Validation Support (Optional)

Users can optionally validate `.sdp` files with Rust:

**File: `Cargo.toml`**
```toml
[dependencies]
# Dummy crate that accepts any #[sdp(...)] attribute
sdp-attributes = "0.1"
```

**File: `lib.rs`** (for validation only)
```rust
use sdp_attributes::*;

include!("schema.sdp");  // Include .sdp file
```

Then run: `cargo check` to validate syntax.

---

## Alternative: If You Want Simpler Syntax

### Consider Switching to Zig Format

If the attribute syntax feels too verbose, **Zig format** is the cleanest option:

```zig
const Plugin = struct {
    name: []const u8,
    metadata: ?Metadata,    // Built-in optional
    fallback: ?*Plugin,     // Recursive optional
};
```

**Pros**:
- ✅ Built-in `?Type` syntax
- ✅ Simpler than Rust (no generics/lifetimes)
- ✅ Free Zig tooling (LSP, formatter)
- ✅ Clean, minimal syntax

**Cons**:
- ❌ Breaking change from spec
- ⚠️ Less familiar than Rust to most developers

**If you're willing to break from Rust**, Zig is the best choice for clean optional syntax.

---

## Final Recommendation

### Stay with Rust, Use Attributes

**Syntax**:
```rust
struct Plugin {
    name: String,
    
    #[sdp(optional)]
    metadata: Metadata,
}
```

**Implementation**: ~150 lines of code
- Parser: +100 lines (attribute parsing)
- Validator: +20 lines (check attribute validity)
- Generator: +30 lines (check for optional attribute)

**Benefits**:
- ✅ Stays true to spec (Rust format)
- ✅ Free tooling (rust-analyzer, rustfmt)
- ✅ Extensible (more attributes later)
- ✅ No breaking changes

**Alternative** (if willing to break):
- Switch to **Zig format** for cleaner `?Type` syntax
- More work (update all docs/tests) but cleaner long-term

Which direction do you prefer?
