# Optional Structs Analysis

## The Problem

Currently, all struct fields are **required**. You can't have:
```
struct AudioEffect {
    name: str,
    parameters: ParameterSet,  // Always present - what if no parameters?
}
```

You want:
```
struct AudioEffect {
    name: str,
    parameters: ?ParameterSet,  // Optional - might be nil
}
```

---

## Requirements Analysis

**What you need**:
- ✅ Optional single structs (not in arrays)
- ✅ Pointer to struct (can be nil)
- ❌ NOT pointers to primitives (complexity not worth it)
- ❌ NOT arrays of pointers (complexity explosion)

**Use cases**:
```
struct Plugin {
    name: str,
    metadata: ?Metadata,      // Optional - might not have metadata
    fallback: ?Plugin,        // Optional - recursive reference
    current_preset: ?Preset,  // Optional - might not be loaded
}
```

---

## Option 1: Explicit Optional Syntax (Recommended)

### Schema Syntax

```
struct AudioEffect {
    name: str,
    parameters: ?ParameterSet,  // ? means optional
    fallback_effect: ?AudioEffect,  // Allows recursive references
}

struct ParameterSet {
    count: u32,
    values: []f32,
}
```

### Wire Format

**Encoding**: 1-byte presence flag + data (if present)
```
Optional field wire format:
  [presence: u8][data: variable] (if presence = 1)
  [presence: u8]                 (if presence = 0, no data follows)

Example - AudioEffect with parameters:
  [name length: 4][name: "Echo"]  // name field
  [presence: 1]                    // parameters present
  [count: 3][values count: 3][1.0, 2.0, 3.0]  // ParameterSet data

Example - AudioEffect without parameters:
  [name length: 4][name: "Mute"]  // name field
  [presence: 0]                    // parameters absent (no more bytes)
```

**Overhead**: 1 byte per optional field

### Generated Go Code

```go
type AudioEffect struct {
    Name       string
    Parameters *ParameterSet  // Pointer indicates optional
}

type ParameterSet struct {
    Count  uint32
    Values []float32
}
```

### Size Calculation

```go
func calculateAudioEffectSize(src *AudioEffect) int {
    size := 0
    
    // Name field (required)
    size += 4 + len(src.Name)
    
    // Parameters field (optional)
    size += 1  // presence flag
    if src.Parameters != nil {
        size += calculateParameterSetSize(src.Parameters)
    }
    
    return size
}
```

### Encoding

```go
func encodeAudioEffect(src *AudioEffect, buf []byte, offset *int) error {
    // Encode name
    binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Name)))
    *offset += 4
    copy(buf[*offset:], src.Name)
    *offset += len(src.Name)
    
    // Encode optional parameters
    if src.Parameters != nil {
        buf[*offset] = 1  // present
        *offset += 1
        if err := encodeParameterSet(src.Parameters, buf, offset); err != nil {
            return err
        }
    } else {
        buf[*offset] = 0  // absent
        *offset += 1
    }
    
    return nil
}
```

### Decoding

```go
func decodeAudioEffect(dest *AudioEffect, data []byte, offset *int, ctx *DecodeContext) error {
    // Decode name
    if *offset + 4 > len(data) {
        return ErrTruncated
    }
    nameLen := binary.LittleEndian.Uint32(data[*offset:])
    *offset += 4
    
    if *offset + int(nameLen) > len(data) {
        return ErrTruncated
    }
    dest.Name = string(data[*offset:*offset + int(nameLen)])
    *offset += int(nameLen)
    
    // Decode optional parameters
    if *offset + 1 > len(data) {
        return ErrTruncated
    }
    if data[*offset] == 1 {
        *offset += 1
        // Present - allocate and decode
        dest.Parameters = &ParameterSet{}
        if err := decodeParameterSet(dest.Parameters, data, offset, ctx); err != nil {
            return err
        }
    } else {
        *offset += 1
        // Absent - leave as nil
        dest.Parameters = nil
    }
    
    return nil
}
```

### Parser Changes (Minimal)

```go
// internal/parser/lexer.go
const (
    // ... existing tokens ...
    TokenQuestion  // '?'
)

// internal/parser/ast.go
type Type struct {
    Kind     TypeKind
    Name     string
    Optional bool      // NEW: true if prefixed with ?
    Element  *Type     // For arrays
}

// internal/parser/parser.go
func (p *Parser) parseType() (*Type, error) {
    // Check for optional marker
    optional := false
    if p.current.Type == TokenQuestion {
        optional = true
        p.advance()
    }
    
    // Parse base type
    typ := &Type{Optional: optional}
    
    if p.current.Type == TokenLeftBracket {
        // Array type: []T
        p.advance()
        p.advance() // skip ]
        elem, err := p.parseType()
        if err != nil {
            return nil, err
        }
        typ.Kind = TypeKindArray
        typ.Element = elem
        
        // Optional arrays not allowed (use empty array instead)
        if optional {
            return nil, fmt.Errorf("optional arrays not supported, use empty array instead")
        }
    } else if p.current.Type == TokenIdentifier {
        // Named type (struct or primitive)
        typ.Name = p.current.Value
        
        // Determine if primitive or named
        if isPrimitive(typ.Name) {
            typ.Kind = TypeKindPrimitive
            // Optional primitives not allowed
            if optional {
                return nil, fmt.Errorf("optional primitives not supported: ?%s", typ.Name)
            }
        } else {
            typ.Kind = TypeKindNamed
            // Optional structs are allowed
        }
        
        p.advance()
    }
    
    return typ, nil
}
```

### Validator Changes (Minimal)

```go
// internal/validator/validator.go

func (v *Validator) validateType(typ *parser.Type) error {
    if typ.Optional {
        // Only allow optional on named types (structs)
        if typ.Kind != parser.TypeKindNamed {
            return fmt.Errorf("optional marker (?) only allowed on struct types, not %v", typ.Kind)
        }
        
        // Check that the named type is actually a struct
        if _, exists := v.structNames[typ.Name]; !exists {
            return fmt.Errorf("optional type ?%s is not a known struct", typ.Name)
        }
    }
    
    // ... existing validation ...
    
    return nil
}
```

### Generator Changes (Moderate)

```go
// internal/generator/golang/types.go

func (g *Generator) fieldType(field *parser.Field) string {
    if field.Type.Optional {
        // Pointer to struct
        return "*" + ToGoName(field.Type.Name)
    }
    
    // ... existing logic ...
}

// internal/generator/golang/encode_gen.go

func generateFieldSizeCalculation(buf *strings.Builder, field *parser.Field) error {
    if field.Type.Optional {
        // Optional struct
        buf.WriteString("\tsize += 1  // presence flag\n")
        buf.WriteString("\tif src.")
        buf.WriteString(ToGoName(field.Name))
        buf.WriteString(" != nil {\n")
        buf.WriteString("\t\tsize += calculate")
        buf.WriteString(ToGoName(field.Type.Name))
        buf.WriteString("Size(src.")
        buf.WriteString(ToGoName(field.Name))
        buf.WriteString(")\n")
        buf.WriteString("\t}\n")
        return nil
    }
    
    // ... existing logic ...
}

func generateFieldEncoding(buf *strings.Builder, field *parser.Field) error {
    if field.Type.Optional {
        fieldName := ToGoName(field.Name)
        
        // Generate if/else for presence
        buf.WriteString("\tif src.")
        buf.WriteString(fieldName)
        buf.WriteString(" != nil {\n")
        buf.WriteString("\t\tbuf[*offset] = 1\n")
        buf.WriteString("\t\t*offset += 1\n")
        buf.WriteString("\t\tif err := encode")
        buf.WriteString(ToGoName(field.Type.Name))
        buf.WriteString("(src.")
        buf.WriteString(fieldName)
        buf.WriteString(", buf, offset); err != nil {\n")
        buf.WriteString("\t\t\treturn err\n")
        buf.WriteString("\t\t}\n")
        buf.WriteString("\t} else {\n")
        buf.WriteString("\t\tbuf[*offset] = 0\n")
        buf.WriteString("\t\t*offset += 1\n")
        buf.WriteString("\t}\n")
        return nil
    }
    
    // ... existing logic ...
}

// internal/generator/golang/decode_gen.go

func generateFieldDecoding(buf *strings.Builder, field *parser.Field) error {
    if field.Type.Optional {
        fieldName := ToGoName(field.Name)
        typeName := ToGoName(field.Type.Name)
        
        // Check presence flag
        buf.WriteString("\tif *offset + 1 > len(data) {\n")
        buf.WriteString("\t\treturn ErrTruncated\n")
        buf.WriteString("\t}\n")
        buf.WriteString("\tif data[*offset] == 1 {\n")
        buf.WriteString("\t\t*offset += 1\n")
        buf.WriteString("\t\tdest.")
        buf.WriteString(fieldName)
        buf.WriteString(" = &")
        buf.WriteString(typeName)
        buf.WriteString("{}\n")
        buf.WriteString("\t\tif err := decode")
        buf.WriteString(typeName)
        buf.WriteString("(dest.")
        buf.WriteString(fieldName)
        buf.WriteString(", data, offset, ctx); err != nil {\n")
        buf.WriteString("\t\t\treturn err\n")
        buf.WriteString("\t\t}\n")
        buf.WriteString("\t} else {\n")
        buf.WriteString("\t\t*offset += 1\n")
        buf.WriteString("\t\tdest.")
        buf.WriteString(fieldName)
        buf.WriteString(" = nil\n")
        buf.WriteString("\t}\n")
        return nil
    }
    
    // ... existing logic ...
}
```

---

## Option 2: Always-Present Wrapper Struct

### Schema Syntax (No Language Changes)

```
struct AudioEffect {
    name: str,
    has_parameters: bool,
    parameters: ParameterSet,  // Ignored if has_parameters = false
}
```

**Pros**:
- ❌ No language changes needed
- ❌ Simpler parser

**Cons**:
- ❌ Wastes space (always encodes struct even if not used)
- ❌ Manual validation (must check bool before using struct)
- ❌ Error-prone (forget to check bool)
- ❌ Not idiomatic in Go (pointers are standard for optional)

**Verdict**: ❌ **Not recommended** - wastes space and error-prone

---

## Option 3: Union Types (Future Enhancement)

### Schema Syntax

```
union OptionalParams {
    none: {},
    some: ParameterSet,
}

struct AudioEffect {
    name: str,
    parameters: OptionalParams,
}
```

**Pros**:
- ✅ More general (supports multiple alternatives)
- ✅ Type-safe

**Cons**:
- ❌ Much more complex to implement
- ❌ Overkill for simple optional case
- ❌ Can be simulated with optional syntax

**Verdict**: ⚠️ **Future feature** - implement optional first, unions later if needed

---

## Option 4: Default Values (Different Problem)

### Schema Syntax

```
struct AudioEffect {
    name: str,
    gain: f32 = 1.0,  // Default if not present
}
```

**This solves a different problem**: Optional *primitives* with defaults, not optional structs.

**Verdict**: ⚠️ **Different feature** - orthogonal to optional structs

---

## Recommended Implementation: Option 1 (Optional Syntax)

### Summary of Changes

| Component | Lines Changed | Complexity |
|-----------|--------------|------------|
| **Lexer** | +5 | Trivial (add `?` token) |
| **Parser** | +20 | Low (parse optional marker) |
| **AST** | +1 | Trivial (add `Optional bool` field) |
| **Validator** | +15 | Low (check only structs are optional) |
| **Type Gen** | +5 | Trivial (emit `*Type`) |
| **Encode Gen** | +30 | Medium (if/else for presence) |
| **Decode Gen** | +35 | Medium (allocate if present) |
| **Tests** | +50 | Medium (test optional fields) |
| **Total** | **~161 lines** | **2-3 hours** |

### Wire Format Examples

**Example 1: Optional present**
```
struct Plugin {
    name: str,
    metadata: ?Metadata,
}

struct Metadata {
    version: u32,
}

// Encoding: Plugin{name: "Reverb", metadata: &Metadata{version: 2}}
[04 00 00 00]  // name length = 4
[52 65 76 65 72 62]  // "Reverb"
[01]  // metadata present
[02 00 00 00]  // version = 2

Total: 15 bytes
```

**Example 2: Optional absent**
```
// Encoding: Plugin{name: "Reverb", metadata: nil}
[04 00 00 00]  // name length = 4
[52 65 76 65 72 62]  // "Reverb"
[00]  // metadata absent

Total: 11 bytes (saved 4 bytes)
```

**Example 3: Recursive optional**
```
struct Node {
    value: u32,
    next: ?Node,
}

// Encoding: Node{value: 1, next: &Node{value: 2, next: nil}}
[01 00 00 00]  // value = 1
[01]  // next present
[02 00 00 00]  // next.value = 2
[00]  // next.next absent

Total: 10 bytes (1 + 4 + 1 + 4 + 1 - 1)
```

### Performance Impact

**Size overhead**: 1 byte per optional field
- AudioEffect with 3 optional fields: +3 bytes
- Negligible compared to typical struct sizes

**Speed overhead**: 1 branch per optional field
- Modern CPUs predict branches well
- Expected impact: <5% for structs with many optionals

**Memory allocations**:
- Decode: +1 allocation per present optional field
- Encode: 0 additional allocations

### Comparison with Protocol Buffers

**Protocol Buffers**:
- All fields are optional by default (proto3: presence tracking)
- Uses field tags (1-2 bytes per field)
- Varint encoding for presence
- More complex wire format

**SDP with optional**:
- Explicit opt-in (`?` marker)
- 1 byte presence flag (simpler than varint)
- No field tags (fixed order)
- More predictable performance

---

## Alternative: Separate Optional Types (More Explicit)

### Schema Syntax

```
struct AudioEffect {
    name: str,
    parameters: Optional<ParameterSet>,
}
```

### Generated Code

```go
// Generic optional type
type Optional[T any] struct {
    Value   *T
    Present bool
}

// Usage
type AudioEffect struct {
    Name       string
    Parameters Optional[ParameterSet]
}

// Check if present
if effect.Parameters.Present {
    params := effect.Parameters.Value
    // ... use params ...
}
```

**Pros**:
- ✅ Very explicit
- ✅ Type-safe

**Cons**:
- ❌ Verbose (`.Value`, `.Present` everywhere)
- ❌ Not idiomatic Go (pointers are standard)
- ❌ Requires Go 1.18+ (generics)

**Verdict**: ❌ **Not recommended** - Go pointers are better

---

## Implementation Checklist

### Phase 1: Parser & Validator (1 hour)
- [ ] Add `?` token to lexer
- [ ] Parse optional marker in type parsing
- [ ] Add `Optional bool` to AST Type
- [ ] Validate only structs can be optional
- [ ] Add tests for parser

### Phase 2: Code Generation (1.5 hours)
- [ ] Update type generator (emit `*Type`)
- [ ] Update size calculator (1 byte + conditional size)
- [ ] Update encoder (if/else for presence)
- [ ] Update decoder (allocate if present)
- [ ] Generate tests for optional fields

### Phase 3: Integration Tests (0.5 hours)
- [ ] Test optional present
- [ ] Test optional absent
- [ ] Test recursive optional
- [ ] Test multiple optionals
- [ ] Test roundtrip

**Total: 3 hours**

---

## Example Test Schema

```
// testdata/optional.sdp

struct Metadata {
    version: u32,
    author: str,
}

struct Config {
    setting_a: bool,
    setting_b: u32,
}

struct Plugin {
    name: str,
    metadata: ?Metadata,
    config: ?Config,
    fallback: ?Plugin,
}

struct PluginRegistry {
    plugins: []Plugin,
    default_config: ?Config,
}
```

### Generated Usage

```go
// Create plugin with metadata
plugin := &Plugin{
    Name: "Reverb",
    Metadata: &Metadata{
        Version: 2,
        Author: "AudioCo",
    },
    Config: nil,  // No custom config
    Fallback: nil,  // No fallback
}

// Create plugin without metadata
simplePlugin := &Plugin{
    Name: "Mute",
    Metadata: nil,
    Config: nil,
    Fallback: nil,
}

// Recursive reference
linkedPlugin := &Plugin{
    Name: "Primary",
    Metadata: nil,
    Config: nil,
    Fallback: &Plugin{
        Name: "Backup",
        Metadata: nil,
        Config: nil,
        Fallback: nil,
    },
}
```

---

## Conclusion

### ✅ Recommended: Implement Optional Syntax (`?`)

**Reasons**:
1. ✅ **Simple**: 1 byte overhead, 1 branch per field
2. ✅ **Non-invasive**: ~161 lines of code
3. ✅ **Idiomatic**: Uses Go pointers (standard pattern)
4. ✅ **Flexible**: Works with recursive references
5. ✅ **Fast to implement**: 2-3 hours
6. ✅ **Clear semantics**: `?` is universal "maybe" symbol

**Restrictions**:
- ✅ Only struct types can be optional (not primitives)
- ✅ Not allowed on array types (use empty arrays)
- ✅ Not allowed on array elements (no `[]?Struct`)

**Wire format**:
- 1 byte presence flag
- Data follows if present
- Clean and simple

**This solves your "single ugliness" with minimal invasiveness!**

### Implementation Priority

**Now** (3 hours):
- Implement optional struct syntax
- Update all generators
- Add comprehensive tests

**Later** (if needed):
- Union types (for more complex alternatives)
- Default values (for optional primitives)

**Never**:
- Optional primitives (use default values instead)
- Arrays of pointers (complexity not worth it)
