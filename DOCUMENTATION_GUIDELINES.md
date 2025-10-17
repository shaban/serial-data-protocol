# Documentation Guidelines

**Version:** 1.0.0  
**Date:** October 18, 2025

## 1. Overview

This document defines documentation standards and practices for the Serial Data Protocol project.

## 2. Documentation Principles

### 2.1 Core Principles

- **Clarity over cleverness** - Simple, direct language
- **Examples are essential** - Show, don't just tell
- **Audience-appropriate** - Different docs for different users
- **Maintainable** - Documentation lives close to code
- **Accurate** - Outdated docs are worse than no docs

### 2.2 Documentation Types

| Type | Audience | Location | Format |
|------|----------|----------|--------|
| Design Spec | Contributors | `DESIGN_SPEC.md` | Markdown |
| Testing Strategy | Contributors | `TESTING_STRATEGY.md` | Markdown |
| User Guide | End users | `docs/user-guide.md` | Markdown |
| API Reference | End users | Generated code | Doc comments |
| Examples | End users | `examples/` | Runnable code |

## 3. Schema Documentation

### 3.1 Schema Comments

**Use doc comments (`///`) for types and fields:**

```rust
/// AudioDevice represents a physical or virtual audio interface.
///
/// Devices are enumerated from the system and contain metadata
/// about available audio inputs/outputs.
struct AudioDevice {
    /// Unique device identifier assigned by the system.
    ///
    /// This ID is stable across program runs but may change
    /// if hardware configuration changes.
    id: u32,
    
    /// Human-readable device name (e.g., "Built-in Microphone").
    name: str,
    
    /// Number of audio channels supported by this device.
    ///
    /// Common values: 1 (mono), 2 (stereo), 6 (5.1), 8 (7.1)
    channels: u8,
    
    /// Sample rate in Hz (e.g., 44100, 48000, 96000).
    sample_rate: f64,
}
```

**Guidelines:**
- First line: brief summary (50-70 chars)
- Additional lines: detailed explanation if needed
- Reference related types when relevant
- Document valid ranges or common values
- Explain non-obvious semantics

### 3.2 Field Naming

**Use clear, descriptive names:**

```rust
// ✅ Good
struct Plugin {
    id: u32,
    name: str,
    vendor: str,
    is_enabled: bool,
}

// ❌ Avoid abbreviations
struct Plugin {
    id: u32,
    nm: str,       // Unclear
    vnd: str,      // Unclear
    en: bool,      // Unclear
}
```

**For booleans, use `is_` or `has_` prefix:**
```rust
struct Device {
    is_active: bool,
    has_midi: bool,
    can_record: bool,
}
```

### 3.3 Enum-Like Fields

**Document allowed values in comments:**

```rust
struct Parameter {
    /// Parameter type identifier.
    ///
    /// Valid values:
    /// - 0: Continuous (float)
    /// - 1: Discrete (integer)
    /// - 2: Boolean (on/off)
    /// - 3: String (text)
    type: u8,
}
```

## 4. Generated Code Documentation

### 4.1 Go Code Comments

**Generator emits doc comments from schema:**

**Schema:**
```rust
/// Device represents an audio device.
struct Device {
    /// Unique identifier.
    id: u32,
}
```

**Generated Go:**
```go
// Device represents an audio device.
type Device struct {
    // ID is the unique identifier.
    ID uint32
}
```

**Guidelines:**
- Convert `///` to `//`
- Capitalize first letter of field comments
- Add field name to comment ("ID is...")
- Preserve multi-line comments

### 4.2 Function Documentation

**Decoder functions include usage examples:**

```go
// Decode deserializes a PluginList from binary wire format.
//
// The data parameter must contain a complete, valid serialized
// PluginList. Partial or malformed data returns an error.
//
// Example:
//
//	data := C.EnumeratePlugins()
//	var plugins PluginList
//	if err := Decode(&plugins, data); err != nil {
//	    log.Fatal(err)
//	}
//
// Maximum data size: 128 MB
// Maximum array elements: 1,000,000 per array
func Decode(dest *PluginList, data []byte) error
```

### 4.3 Error Documentation

**Document error conditions:**

```go
var (
    // ErrUnexpectedEOF indicates data ended before expected.
    // Check wire format integrity.
    ErrUnexpectedEOF = errors.New("unexpected end of data")
    
    // ErrDataTooLarge indicates data exceeds 128MB limit.
    // Consider chunking data or using smaller messages.
    ErrDataTooLarge = errors.New("data exceeds 128MB limit")
    
    // ErrArrayTooLarge indicates array count exceeds per-array limit of 1M elements.
    ErrArrayTooLarge = errors.New("array count exceeds limit")
)
```

## 5. User Guide

### 5.1 Structure

**User guide organization:**

```
docs/user-guide.md
  1. Introduction
  2. Quick Start
  3. Writing Schemas
  4. Generating Code
  5. Encoding Data (C/Swift)
  6. Decoding Data (Go/Rust)
  7. Error Handling
  8. Best Practices
  9. Troubleshooting
  10. FAQ
```

### 5.2 Quick Start

**Provide complete, runnable examples:**

```markdown
## Quick Start

### 1. Define Schema

Create `device.sdp`:

```rust
/// Device represents an audio device.
struct Device {
    id: u32,
    name: str,
}
```

### 2. Generate Code

```bash
sdp-gen -schema device.sdp -output device -lang go
```

### 3. Encode (C)

```c
#include "device_builder.h"

device_builder_t* builder = NewDeviceBuilder();
SetDeviceID(builder, 42);
SetDeviceName(builder, "Built-in");
serial_data_t data = FinalizeBuilder(builder);

// Send data to Go via CGO
```

### 4. Decode (Go)

```go
import "yourproject/device"

cData := C.EncodeDevice()
goData := C.GoBytes(unsafe.Pointer(cData.data), C.int(cData.len))

var d device.Device
if err := device.Decode(&d, goData); err != nil {
    log.Fatal(err)
}

fmt.Printf("Device: %s (ID: %d)\n", d.Name, d.ID)
```
```

### 5.3 Code Examples

**Include working examples in `examples/` directory:**

```
examples/
  01-basic/
    schema/
      device.sdp
    c-encoder/
      main.c
      Makefile
    go-decoder/
      main.go
      go.mod
    README.md
  
  02-nested-structs/
    ...
  
  03-optional-fields/
    ...
  
  04-error-handling/
    ...
```

**Each example includes:**
- Complete schema
- Encoder implementation (C/Swift/Rust)
- Decoder implementation (Go)
- Build instructions
- Expected output

## 6. API Reference

### 6.1 Generator API

**Document CLI flags:**

```markdown
## sdp-gen

Generate code from Serial Data Protocol schemas.

### Usage

```bash
sdp-gen -schema <file> -output <dir> -lang <language>
```

### Flags

- `-schema <file>` - Input schema file (.sdp)
- `-output <dir>` - Output directory for generated code
- `-lang <language>` - Target language: go, rust, swift, c
- `-validate-only` - Validate schema without generating code
- `-verbose` - Print detailed generation info

### Examples

**Generate Go decoder:**
```bash
sdp-gen -schema plugin.sdp -output generated/plugin -lang go
```

**Validate schema:**
```bash
sdp-gen -schema plugin.sdp -validate-only
```
```

### 6.2 Builder API (C)

**Document generated builder functions:**

```markdown
## C Builder API

Generated from `plugin.sdp`.

### Types

**`plugin_list_builder_t`**

Root builder for PluginList.

**`tentative_plugin_t`**

Tentative plugin being built.

### Functions

**`plugin_list_builder_t* NewPluginListBuilder(void)`**

Creates a new builder.

**Returns:** Builder instance or NULL on allocation failure.

**`tentative_plugin_t* BeginPlugin(plugin_list_builder_t* builder)`**

Begins building a new plugin.

**Parameters:**
- `builder` - Parent builder

**Returns:** Tentative plugin instance or NULL on error.

**`void SetPluginID(tentative_plugin_t* plugin, uint32_t id)`**

Sets plugin ID field.

**Parameters:**
- `plugin` - Plugin being built
- `id` - Unique plugin identifier

**`void DiscardPlugin(plugin_list_builder_t* builder, tentative_plugin_t* plugin)`**

Discards plugin without including in output.

**Parameters:**
- `builder` - Parent builder
- `plugin` - Plugin to discard

**`serial_data_t FinalizeBuilder(plugin_list_builder_t* builder)`**

Finalizes builder and returns serialized data.

**Parameters:**
- `builder` - Builder to finalize

**Returns:** Serialized data or NULL on error. Check `error` field.

**Example:**

```c
plugin_list_builder_t* builder = NewPluginListBuilder();

tentative_plugin_t* p = BeginPlugin(builder);
SetPluginID(p, 42);
SetPluginName(p, "Reverb");

if (!should_include) {
    DiscardPlugin(builder, p);
}

serial_data_t result = FinalizeBuilder(builder);
if (result.error != ERR_NONE) {
    fprintf(stderr, "Error: %s\n", result.error_msg);
}

// Use result.data, result.len

DestroyBuilder(builder);
```
```

## 7. Changelog

### 7.1 Format

**Use Keep a Changelog format:**

```markdown
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [1.0.0] - 2025-10-18

### Added
- Initial release
- Go decoder generation
- C encoder generation
- Schema validation
- Wire format specification

### Changed
- Nothing (initial release)

### Deprecated
- Nothing

### Removed
- Nothing

### Fixed
- Nothing

### Security
- Nothing
```

## 8. Internal Documentation

### 8.1 Code Comments

**Comment non-obvious code:**

```go
// commitFieldsInOrder writes fields to buffer in schema definition order,
// regardless of the order they were written by the user.
// This ensures wire format consistency.
func commitFieldsInOrder(p *tentativePlugin, buf *buffer) {
    // Schema order: id, name, vendor, parameters
    
    if p.id.written {
        buf.WriteU32(p.id.value)
    } else {
        buf.WriteU32(0)  // Default for omitted field
    }
    
    // ... remaining fields
}
```

**Avoid obvious comments:**

```go
// ❌ Redundant
// Increment count
count++

// ✅ Useful
// Track total for limit check in finalize
totalElements++
```

### 8.2 Architecture Decisions

**Document important design decisions:**

```markdown
## Architecture Decision Record: Field Reordering

**Date:** 2025-10-18

**Status:** Accepted

**Context:**

Native code gathers data in discovery order, which may not match
schema field order. Requiring native code to track and reorder
fields is error-prone.

**Decision:**

Allow flexible field write order on encoder side. Tentative
structs store fields independently. On commit, fields are
written to wire format in schema definition order.

**Consequences:**

**Positive:**
- Encoder API more flexible
- Natural data gathering flow
- User happiness increased

**Negative:**
- Tentative structs larger (storage + flag per field)
- Slightly more complex commit logic
- ~50 bytes overhead per tentative

**Alternatives Considered:**

1. Enforce schema order - rejected (poor UX)
2. Wire format includes field IDs - rejected (bloat)
```

## 9. Maintenance

### 9.1 Documentation Reviews

**Review docs with code changes:**

- Update docs in same PR as code changes
- Mark outdated docs with `[OUTDATED]` prefix
- Remove docs for deleted features
- Update examples to match current API

### 9.2 Link Checking

**Verify internal links regularly:**

```bash
# Check markdown links
markdown-link-check DESIGN_SPEC.md
markdown-link-check TESTING_STRATEGY.md
markdown-link-check docs/**/*.md
```

### 9.3 Example Validation

**Ensure examples still work:**

```bash
# Run all examples
for dir in examples/*/; do
    cd "$dir"
    make clean
    make test
    cd -
done
```

## 10. Style Guide

### 10.1 Markdown

**Use consistent formatting:**

```markdown
# H1 - Document title only

## H2 - Major sections

### H3 - Subsections

#### H4 - Rarely needed

**Bold** for emphasis
*Italic* for terms
`code` for identifiers
```

**Code blocks with language:**

````markdown
```rust
struct Device {
    id: u32,
}
```

```go
type Device struct {
    ID uint32
}
```
````

**Lists:**

```markdown
- Use hyphens for unordered lists
- Consistent indentation
- Blank line before and after list

1. Use numbers for ordered lists
2. Generator handles numbering
3. Blank line before and after
```

### 10.2 Voice and Tone

**Use active voice:**

```markdown
✅ "The decoder reads fields sequentially."
❌ "Fields are read sequentially by the decoder."
```

**Be direct:**

```markdown
✅ "Omit optional fields to use defaults."
❌ "Optional fields may be omitted if one wishes to utilize default values."
```

**Address the reader:**

```markdown
✅ "You can set fields in any order."
❌ "The user is able to set fields in any order."
```

---

**End of Documentation Guidelines**
