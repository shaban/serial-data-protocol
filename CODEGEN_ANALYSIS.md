# Code Generation: Go Templates vs Rust Introspection

## The Question

Should the Rust code generator be written in:
1. **Go** (using text/template, existing parser)
2. **Rust** (using AST introspection, proc macros)

## Current Approach (Hybrid)

We're using **Go → JSON → Rust**:

```
┌──────────┐    JSON AST     ┌────────────┐
│ Go Parser│ ────────────> │ Rust Codegen│
│ (sdp-gen)│  --ast-json   │   (binary)  │
└──────────┘                └────────────┘
```

**Pros:**
- ✅ Reuses existing Go parser (proven correct)
- ✅ Single source of truth (Go AST)
- ✅ Language-agnostic (JSON intermediate)
- ✅ Can generate any language (Python, C, etc.)

**Cons:**
- ❌ Two languages in toolchain
- ❌ Extra JSON serialization step
- ❌ No Rust type safety for AST

## Option 1: Pure Go (with Templates)

Write Rust generator in Go using `text/template`:

```go
// internal/generator/rust/generator.go
package rust

import "text/template"

const rustStructTemplate = `
#[derive(Debug, Clone, PartialEq)]
pub struct {{.Name}} {
{{- range .Fields}}
    pub {{.Name}}: {{.RustType}},
{{- end}}
}

impl {{.Name}} {
    pub fn encode_to_slice(&self, buf: &mut [u8]) -> Result<usize> {
        use sdp::wire_slice;
        let mut offset = 0;
        {{range .Fields}}
        {{template "encodeField" .}}
        {{end}}
        Ok(offset)
    }
}
`

func Generate(schema *ast.Schema, outputDir string) error {
    tmpl := template.Must(template.New("rust").Parse(rustStructTemplate))
    // ... render templates
}
```

### Pros ✅

1. **Single Language Toolchain**
   - Only need Go installed
   - Simpler CI/CD
   - Easier for contributors

2. **Existing Infrastructure**
   - Parser already in Go
   - Validator already in Go
   - Can reuse helper functions

3. **Go Template Power**
   - Built-in template engine
   - Loops, conditionals, helpers
   - Well-documented

4. **Type Safety**
   - Go compiler checks template expansion
   - AST is strongly typed

5. **Easier Debugging**
   - Can inspect AST directly
   - Printf debugging works
   - No JSON marshaling issues

### Cons ❌

1. **String-Based Generation**
   - Rust code as strings
   - No syntax checking until compile
   - Indentation is manual

2. **No Rust Type System**
   - Can't validate Rust types at generation time
   - Might generate invalid Rust

3. **Template Escaping**
   - Rust has `{{ }}` syntax too (macros)
   - Need careful escaping

4. **Learning Curve**
   - Go templates have quirks
   - Different from Rust macros

### Example (How it would look)

```go
// internal/generator/rust/encode.go
func generateEncode(w io.Writer, s *ast.Struct) error {
    return encodeTemplate.Execute(w, struct {
        Name   string
        Fields []FieldInfo
    }{
        Name: s.Name,
        Fields: convertFields(s.Fields),
    })
}

const encodeTemplate = `
pub fn encode_to_slice(&self, buf: &mut [u8]) -> Result<usize> {
    use sdp::wire_slice;
    let mut offset = 0;
    {{range .Fields -}}
    {{if .IsPrimitive -}}
    wire_slice::encode_{{.WireType}}(buf, offset, self.{{.Name}})?;
    offset += {{.Size}};
    {{else if .IsString -}}
    let written = wire_slice::encode_string(buf, offset, &self.{{.Name}})?;
    offset += written;
    {{end -}}
    {{end -}}
    Ok(offset)
}
`
```

## Option 2: Pure Rust (with Proc Macros)

Use Rust procedural macros to generate code:

```rust
// Hypothetical approach (would require Rust parser)
use sdp_macros::sdp_schema;

sdp_schema! {
    struct AllPrimitives {
        u32_field: u32,
        str_field: string,
    }
}

// Expands to:
#[derive(Debug, Clone, PartialEq)]
pub struct AllPrimitives {
    pub u32_field: u32,
    pub str_field: String,
}
// ... impl blocks
```

### Pros ✅

1. **Rust Type Safety**
   - Generated code is checked by Rust compiler
   - Can't generate invalid syntax
   - Better IDE support

2. **Introspection**
   - Can use Rust's type system
   - Derive macros give free metadata
   - Syn crate for AST manipulation

3. **No String Templates**
   - Use `quote!` macro for code generation
   - Hygiene is automatic
   - Proper syntax highlighting

4. **Single Ecosystem**
   - Pure Rust toolchain
   - cargo build handles everything
   - No external dependencies

### Cons ❌

1. **Need Rust Parser**
   - Would have to rewrite parser in Rust
   - Or use syn to parse SDP schema (hacky)
   - Or keep Go parser (current hybrid approach)

2. **Proc Macro Complexity**
   - Harder to debug than templates
   - Compile-time only errors
   - Steeper learning curve

3. **Build Times**
   - Proc macros slow down compilation
   - Every schema change = recompile

4. **No Multi-Language**
   - Rust generator only works for Rust
   - Can't reuse for Python, C, etc.

## Recommendation: **Stick with Go Templates**

Here's why:

### For Your Use Case

1. **Already Have Go Parser** ✅
   - 415 tests passing
   - Don't throw away working code
   - Parser is the hard part

2. **Multi-Language Future** ✅
   - Might want Python generator later
   - Might want C generator
   - Go → JSON → X is flexible

3. **Go Template Advantages** ✅
   - Simpler than proc macros
   - Easier to iterate on
   - Better for prototyping

4. **You Asked "Easier for You"** ✅
   - You (as an AI) can generate Go templates easily
   - Templates are declarative
   - Less ceremony than Rust proc macros

### Hybrid Approach Evolution

Current:
```
Go Parser → JSON → Rust Binary → Rust Code
```

Proposed:
```
Go Parser → Go Template → Rust Code
           (internal/generator/rust/)
```

Same `sdp-gen` binary, just with Rust backend using templates!

### Example Implementation

```go
// cmd/sdp-gen/main.go
switch *lang {
case "go":
    err = golang.Generate(schema, *outputDir, *verbose)
case "rust":
    err = rust.Generate(schema, *outputDir, *verbose)
default:
    return fmt.Errorf("unsupported language: %s", *lang)
}
```

```go
// internal/generator/rust/generator.go
package rust

func Generate(schema *ast.Schema, outputDir string, verbose bool) error {
    // Generate lib.rs
    if err := generateLib(schema, outputDir); err != nil {
        return err
    }
    
    // Generate types.rs
    if err := generateTypes(schema, outputDir); err != nil {
        return err
    }
    
    // Generate encode.rs (slice API)
    if err := generateEncode(schema, outputDir); err != nil {
        return err
    }
    
    // Generate decode.rs (slice API)
    if err := generateDecode(schema, outputDir); err != nil {
        return err
    }
    
    return nil
}
```

## Concrete Answer

**Writing Rust generator in Go with templates would be:**

1. ✅ **Easier** - Templates simpler than proc macros
2. ✅ **Faster** - Don't need new Rust parser
3. ✅ **More Flexible** - Can add Python, C later
4. ✅ **Better for You** - Templates are declarative
5. ✅ **Proven Pattern** - Many code generators use this

**The current approach (Go binary calling Rust binary) is acceptable but:**
- ❌ More complex build process
- ❌ Two languages in toolchain
- ✅ But validates we can be language-agnostic

**Best path forward:**
1. Keep Go parser (it works!)
2. Add `internal/generator/rust/` package
3. Use Go templates (like golang generator)
4. Single `sdp-gen` binary outputs any language
5. Remove separate Rust generator binary

This matches the pattern of tools like:
- `protoc` (C++ core, multi-language plugins)
- `flatc` (C++ core, multi-language backends)
- Our own Go generator!

Would you like me to start implementing the Go template-based Rust generator? 🚀
