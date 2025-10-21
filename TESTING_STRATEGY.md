# Testing Strategy

**Version:** 2.0.0  
**Date:** October 21, 2025

## 1. Overview

This document defines the comprehensive testing strategy for the Serial Data Protocol implementation, with emphasis on avoiding CGO in test files and unified test orchestration via Make.

## 2. Quick Start

### 2.1 Running Tests

**Use the unified Make-based test orchestration:**

```bash
# Run all tests (Go, C++, Rust)
make test

# Run specific language tests
make test-go      # Go tests only (fast, 415+ tests)
make test-cpp     # C++ tests only
make test-rust    # Rust tests only
make test-swift   # Points to macos_testing/

# Run benchmarks
make benchmark

# Clean all generated code
make clean

# Show all available targets
make help
```

### 2.2 Test Organization

```
tests/                   → Shell scripts for orchestration
  test_go.sh            → Wraps 'go test ./...' with formatting
  test_cpp.sh           → Finds and runs C++ tests in testdata/cpp/
  test_rust.sh          → Finds and runs Rust tests in testdata/rust/
  test_swift.sh         → Points to macos_testing/ directory

Makefile                → Root orchestration (calls test scripts)

testdata/
  schemas/*.sdp         → Canonical schema files
  data/*.json           → Benchmark input data
  binaries/*.sdpb       → Reference wire format files
  go/*/                 → Generated Go test packages
  cpp/*/                → Generated C++ test packages
  rust/*/               → Generated Rust test packages
  swift/*/              → Generated Swift packages
```

**Key principle:** Go tests (`go test ./...`) test **only Go scope** (parser, validator, Go code generation, wire format). Cross-language testing is handled by shell scripts, not Go test files.

## 3. Testing Principles

### 3.1 Core Principles

- **No CGO in `_test.go` files** - Use wire format testing, subprocess communication
- **Wire format is source of truth** - Test at protocol level, not implementation
- **Clean slate testing** - Regenerate test packages on every test run
- **Comprehensive validation** - Test happy paths, edge cases, and error conditions
- **Language scope separation** - Go tests don't test C++/Rust/Swift directly

### 3.2 Test Levels

```
Level 0: Make Orchestration (root Makefile + tests/*.sh)
  ↓ Unified entry point for all languages
  ↓ Consistent output formatting
  ↓ Parallelizable by language

Level 1: Unit Tests (internal/)
  ↓ Parser, templates, wire format helpers
  ↓ No generated code needed
  ↓ Run via: go test ./internal/...
  
Level 2: Generator Tests
  ↓ Run generator on test schemas
  ↓ Validate generated code structure
  ↓ Run via: go test ./internal/generator/...
  
Level 3: Integration Tests (testdata/)
  ↓ Use generated packages
  ↓ Wire format fixtures
  ↓ Run via: go test . (integration_test.go)
  
Level 4: Cross-Language Tests
  ↓ C++/Rust/Swift tests in testdata/{cpp,rust,swift}/
  ↓ Run via: tests/test_{cpp,rust,swift}.sh
  ↓ Each language has its own build system (Makefile, Cargo.toml, Package.swift)
```

## 4. Test Infrastructure

### 3.1 TestMain Setup

**Every integration test package uses TestMain to ensure fresh generated code:**

```go
// integration_test.go
package serialdataprotocol_test

import (
    "os"
    "os/exec"
    "testing"
)

func TestMain(m *testing.M) {
    // Step 1: Clean old generated files
    cleanTestdata()
    
    // Step 2: Build generator (ensures latest version)
    buildGenerator()
    
    // Step 3: Generate test packages
    generateTestPackages()
    
    // Step 4: Run tests
    code := m.Run()
    
    // Step 5: Exit (leave generated files for inspection)
    os.Exit(code)
}

func cleanTestdata() {
    dirs := []string{
        "testdata/primitives",
        "testdata/nested",
        "testdata/arrays",
    }
    
    for _, dir := range dirs {
        os.RemoveAll(dir)
    }
}

func buildGenerator() {
    cmd := exec.Command("go", "build", "-o", "testdata/sdp-gen", "./cmd/sdp-gen")
    cmd.Stdout = os.Stdout
    cmd.Stderr = os.Stderr
    
    if err := cmd.Run(); err != nil {
        panic("Failed to build generator: " + err.Error())
    }
}

func generateTestPackages() {
    schemas := []struct {
        schema string
        output string
    }{
        {"testdata/primitives.sdp", "testdata/primitives"},
        {"testdata/nested.sdp", "testdata/nested"},
        {"testdata/arrays.sdp", "testdata/arrays"},
    }
    
    for _, s := range schemas {
        cmd := exec.Command(
            "testdata/sdp-gen",
            "-schema", s.schema,
            "-output", s.output,
            "-lang", "go",
        )
        cmd.Stdout = os.Stdout
        cmd.Stderr = os.Stderr
        
        if err := cmd.Run(); err != nil {
            panic("Failed to generate " + s.schema + ": " + err.Error())
        }
    }
}
```

### 3.2 Test Data Organization

```
testdata/
  # Schema files (checked into git)
  primitives.sdp
  nested.sdp
  arrays.sdp
  edge_cases.sdp
  
  # Generated packages (gitignored, regenerated on test)
  primitives/
    types.go
    decode.go
    metadata.go
  nested/
    types.go
    decode.go
    metadata.go
  
  # Wire format fixtures (checked into git)
  fixtures/
    primitives_valid.bin
    primitives_invalid.bin
    nested_deep.bin
    arrays_empty.bin
    arrays_large.bin
  
  # Cross-language encoders (checked into git)
  encoders/
    c_encoder.c
    rust_encoder.rs
    swift_encoder.swift
  
  # Generator binary (gitignored)
  sdp-gen
```

**.gitignore:**
```gitignore
# Generated test packages
testdata/*/
!testdata/*.sdp
!testdata/fixtures/
!testdata/encoders/

# Generated binaries
testdata/sdp-gen
testdata/encoders/c_encoder
testdata/encoders/rust_encoder
testdata/encoders/swift_encoder
```

## 11. Unit Tests (Level 1)

### 5.1 Schema Parser Tests

**Test:** `internal/parser/parser_test.go`

```go
func TestParseSimpleStruct(t *testing.T) {
    input := `
    struct Device {
        id: u32,
        name: str,
    }
    `
    
    schema, err := ParseSchema(input)
    require.NoError(t, err)
    require.Len(t, schema.Structs, 1)
    
    s := schema.Structs[0]
    assert.Equal(t, "Device", s.Name)
    assert.Len(t, s.Fields, 2)
    assert.Equal(t, "id", s.Fields[0].Name)
    assert.Equal(t, "u32", s.Fields[0].Type)
}

func TestParseDocComments(t *testing.T) {
    input := `
    /// Device represents an audio device.
    struct Device {
        /// Unique identifier.
        id: u32,
    }
    `
    
    schema, err := ParseSchema(input)
    require.NoError(t, err)
    
    assert.Equal(t, "Device represents an audio device.", schema.Structs[0].Comment)
    assert.Equal(t, "Unique identifier.", schema.Structs[0].Fields[0].Comment)
}

func TestRejectInvalidSyntax(t *testing.T) {
    tests := []struct {
        name  string
        input string
    }{
        {"missing comma", "struct A { id: u32 name: str }"},
        {"invalid type", "struct A { id: unknown }"},
        {"empty struct", "struct A { }"},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := ParseSchema(tt.input)
            assert.Error(t, err)
        })
    }
}
```

### 5.2 Validator Tests

**Test:** `internal/validator/validator_test.go`

```go
func TestValidateCircularReference(t *testing.T) {
    schema := &Schema{
        Structs: []Struct{
            {
                Name: "Node",
                Fields: []Field{
                    {Name: "value", Type: "u32"},
                    {Name: "next", Type: "Node"},
                },
            },
        },
    }
    
    err := ValidateSchema(schema)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "circular reference")
}

func TestValidateReservedKeywords(t *testing.T) {
    schema := &Schema{
        Structs: []Struct{
            {
                Name: "Config",
                Fields: []Field{
                    {Name: "type", Type: "u32"},  // 'type' reserved in Go/Rust
                },
            },
        },
    }
    
    err := ValidateSchema(schema)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "reserved")
}

func TestValidateUnknownType(t *testing.T) {
    schema := &Schema{
        Structs: []Struct{
            {
                Name: "Plugin",
                Fields: []Field{
                    {Name: "device", Type: "AudioDevice"},  // Not defined
                },
            },
        },
    }
    
    err := ValidateSchema(schema)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "unknown type")
}
```

### 5.3 Wire Format Tests

**Test:** `internal/wire/wire_test.go`

```go
func TestEncodeDecodeU32(t *testing.T) {
    buf := make([]byte, 4)
    EncodeU32(buf, 0, 0x12345678)
    
    assert.Equal(t, []byte{0x78, 0x56, 0x34, 0x12}, buf) // Little-endian
    
    val := DecodeU32(buf, 0)
    assert.Equal(t, uint32(0x12345678), val)
}

func TestEncodeDecodeString(t *testing.T) {
    buf := new(bytes.Buffer)
    EncodeString(buf, "Hello")
    
    // [length: 5][bytes: Hello]
    expected := []byte{0x05, 0x00, 0x00, 0x00, 'H', 'e', 'l', 'l', 'o'}
    assert.Equal(t, expected, buf.Bytes())
    
    r := bytes.NewReader(buf.Bytes())
    str, err := DecodeString(r)
    require.NoError(t, err)
    assert.Equal(t, "Hello", str)
}

func TestEncodeEmptyArray(t *testing.T) {
    buf := make([]byte, 4)
    EncodeU32(buf, 0, 0)  // count = 0
    
    assert.Equal(t, []byte{0x00, 0x00, 0x00, 0x00}, buf)
}
```

## 11. Generator Tests (Level 2)

**CRITICAL TESTING COMMITMENT:**

Generator components are tested in two phases:

**Phase 1: Unit Tests (during development)**
- Test individual generator functions in isolation
- Focus on correctness of type mapping, name conversion, template logic
- Fast feedback loop during implementation
- Example: `TestMapTypeToGo()`, `TestToGoFieldName()`, `TestGenerateStructDefinition()`

**Phase 2: Integration Tests (after generator complete)**
- Run full generator on test schemas
- Compile generated code and verify it works end-to-end
- Test against wire format fixtures
- Deferred until generator pipeline is complete

**Error Handling Convention:**
- Generator functions return `(result, error)` - never call `os.Exit()` directly
- CLI layer (`cmd/sdp-gen/main.go`) handles error printing and exit codes:
  ```go
  if err != nil {
      fmt.Fprintf(os.Stderr, "Error: %v\n", err)
      os.Exit(1)  // Non-zero for errors (0 = success)
  }
  ```
- This keeps generator package testable and reusable

### 5.1 Template Rendering Tests

**Test:** `internal/generator/go_test.go`

```go
func TestGenerateGoStruct(t *testing.T) {
    schema := &Schema{
        Structs: []Struct{
            {
                Name:    "Device",
                Comment: "Device represents an audio device.",
                Fields: []Field{
                    {Name: "id", Type: "u32", Comment: "Unique identifier."},
                    {Name: "name", Type: "str", Comment: "Device name."},
                },
            },
        },
    }
    
    output := GenerateGoTypes(schema)
    
    assert.Contains(t, output, "// Device represents an audio device.")
    assert.Contains(t, output, "type Device struct {")
    assert.Contains(t, output, "// ID is the unique identifier.")
    assert.Contains(t, output, "ID uint32")
    assert.Contains(t, output, "// Name is the device name.")
    assert.Contains(t, output, "Name string")
}

func TestGenerateGoDecoder(t *testing.T) {
    schema := &Schema{
        Structs: []Struct{
            {
                Name: "Simple",
                Fields: []Field{
                    {Name: "value", Type: "u32"},
                },
            },
        },
    }
    
    output := GenerateGoDecoder(schema)
    
    assert.Contains(t, output, "func Decode(dest *Simple, data []byte) error")
    assert.Contains(t, output, "binary.LittleEndian.Uint32")
    assert.Contains(t, output, "MaxSerializedSize")
}
```

### 5.2 Generated Code Validation

**Test:** `internal/generator/validate_test.go`

```go
func TestGeneratedCodeCompiles(t *testing.T) {
    schema := parseSchemaFile(t, "testdata/primitives.sdp")
    
    // Generate Go code
    typesCode := GenerateGoTypes(schema)
    decodeCode := GenerateGoDecoder(schema)
    
    // Write to temp directory
    tmpDir := t.TempDir()
    writeFile(t, filepath.Join(tmpDir, "types.go"), typesCode)
    writeFile(t, filepath.Join(tmpDir, "decode.go"), decodeCode)
    
    // Try to compile
    cmd := exec.Command("go", "build", tmpDir)
    output, err := cmd.CombinedOutput()
    
    if err != nil {
        t.Fatalf("Generated code does not compile:\n%s", output)
    }
}

func TestNoRuntimeDependency(t *testing.T) {
    schema := parseSchemaFile(t, "testdata/primitives.sdp")
    code := GenerateGoDecoder(schema)
    
    // Parse imports
    fset := token.NewFileSet()
    f, err := parser.ParseFile(fset, "", code, parser.ImportsOnly)
    require.NoError(t, err)
    
    for _, imp := range f.Imports {
        path := strings.Trim(imp.Path.Value, `"`)
        assert.NotContains(t, path, "serial-data-protocol/runtime")
    }
}
```

## 11. Integration Tests (Level 3)

### 6.1 Decode Tests

**Test:** `integration_test.go` (uses generated testdata packages)

```go
func TestDecodePrimitives(t *testing.T) {
    // Hand-crafted wire format
    wire := []byte{
        // u32: 42
        0x2A, 0x00, 0x00, 0x00,
        // u64: 1000
        0xE8, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
        // f64: 3.14159
        0x6E, 0x86, 0x1B, 0xF0, 0xF9, 0x21, 0x09, 0x40,
        // bool: true
        0x01,
        // str: "test"
        0x04, 0x00, 0x00, 0x00, 't', 'e', 's', 't',
    }
    
    var data primitives.AllTypes
    err := primitives.Decode(&data, wire)
    
    require.NoError(t, err)
    assert.Equal(t, uint32(42), data.U32Field)
    assert.Equal(t, uint64(1000), data.U64Field)
    assert.InDelta(t, 3.14159, data.F64Field, 0.00001)
    assert.True(t, data.BoolField)
    assert.Equal(t, "test", data.StrField)
}

func TestDecodeNestedStructs(t *testing.T) {
    // Load pre-encoded fixture
    wire, err := os.ReadFile("testdata/fixtures/nested_valid.bin")
    require.NoError(t, err)
    
    var data nested.PluginList
    err = nested.Decode(&data, wire)
    
    require.NoError(t, err)
    assert.Len(t, data.Plugins, 2)
    assert.Equal(t, "Reverb", data.Plugins[0].Name)
    assert.Len(t, data.Plugins[0].Parameters, 3)
}

func TestDecodeEmptyArrays(t *testing.T) {
    wire := []byte{
        0x00, 0x00, 0x00, 0x00,  // plugins count: 0
    }
    
    var data nested.PluginList
    err := nested.Decode(&data, wire)
    
    require.NoError(t, err)
    assert.Len(t, data.Plugins, 0)
}
```

### 6.2 Error Handling Tests

```go
func TestDecodeUnexpectedEOF(t *testing.T) {
    wire := []byte{0x01, 0x00}  // Truncated
    
    var data primitives.AllTypes
    err := primitives.Decode(&data, wire)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "unexpected end")
}

func TestDecodeExceedsMaxSize(t *testing.T) {
    // Create 129MB of data
    wire := make([]byte, 129*1024*1024)
    
    var data primitives.AllTypes
    err := primitives.Decode(&data, wire)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "exceeds")
}

func TestDecodeArrayTooLarge(t *testing.T) {
    wire := []byte{
        0xFF, 0xFF, 0xFF, 0xFF,  // count: 4294967295 (max uint32)
    }
    
    var data arrays.DeviceList
    err := arrays.Decode(&data, wire)
    
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "array count exceeds")
}
```

### 6.3 Roundtrip Tests

```go
func TestRoundtripViaFile(t *testing.T) {
    // Write wire format to file
    wire := []byte{ /* valid data */ }
    tmpFile := filepath.Join(t.TempDir(), "data.bin")
    err := os.WriteFile(tmpFile, wire, 0644)
    require.NoError(t, err)
    
    // Read and decode
    data, err := os.ReadFile(tmpFile)
    require.NoError(t, err)
    
    var decoded primitives.AllTypes
    err = primitives.Decode(&decoded, data)
    require.NoError(t, err)
}
```

## 11. Cross-Language Tests (Level 4)

### 7.1 C Encoder Tests

**C Encoder:** `testdata/encoders/c_encoder.c`

```c
#include <stdio.h>
#include <stdint.h>
#include <string.h>

int main() {
    // Read input from stdin (JSON or simple format)
    // Encode to stdout (binary)
    
    uint32_t id = 42;
    const char* name = "TestDevice";
    
    // Write to stdout
    fwrite(&id, sizeof(id), 1, stdout);
    
    uint32_t name_len = strlen(name);
    fwrite(&name_len, sizeof(name_len), 1, stdout);
    fwrite(name, name_len, 1, stdout);
    
    // Logs to stderr
    fprintf(stderr, "Encoded device id=%u name=%s\n", id, name);
    
    return 0;
}
```

**Go Test:**

```go
func TestCEncoder(t *testing.T) {
    // Build C encoder
    buildCmd := exec.Command("gcc", 
        "testdata/encoders/c_encoder.c", 
        "-o", "testdata/encoders/c_encoder")
    buildCmd.Stderr = os.Stderr
    require.NoError(t, buildCmd.Run())
    
    // Run encoder
    encodeCmd := exec.Command("testdata/encoders/c_encoder")
    
    var stdout, stderr bytes.Buffer
    encodeCmd.Stdout = &stdout
    encodeCmd.Stderr = &stderr
    
    err := encodeCmd.Run()
    if err != nil {
        t.Logf("C encoder stderr:\n%s", stderr.String())
        t.Fatalf("C encoder failed: %v", err)
    }
    
    // Decode in Go
    var device primitives.Device
    err = primitives.Decode(&device, stdout.Bytes())
    require.NoError(t, err)
    
    assert.Equal(t, uint32(42), device.ID)
    assert.Equal(t, "TestDevice", device.Name)
    
    t.Logf("C encoder output:\n%s", stderr.String())
}
```

### 7.2 Rust Encoder Tests

**Rust Encoder:** `testdata/encoders/rust_encoder.rs`

```rust
use std::io::{self, Write};

fn main() {
    let id: u32 = 42;
    let name = "TestDevice";
    
    let mut stdout = io::stdout();
    
    // Write binary to stdout
    stdout.write_all(&id.to_le_bytes()).unwrap();
    stdout.write_all(&(name.len() as u32).to_le_bytes()).unwrap();
    stdout.write_all(name.as_bytes()).unwrap();
    
    // Logs to stderr
    eprintln!("Encoded device id={} name={}", id, name);
}
```

**Go Test:** Similar to C encoder test.

## 11. Fuzzing

### 8.1 Fuzz Decoder

```go
func FuzzDecode(f *testing.F) {
    // Seed corpus
    f.Add([]byte{0x00, 0x00, 0x00, 0x00})  // Empty array
    f.Add([]byte{0x01, 0x00, 0x00, 0x00, 0x2A, 0x00, 0x00, 0x00})  // One element
    
    f.Fuzz(func(t *testing.T, data []byte) {
        var decoded primitives.AllTypes
        
        // Should never panic
        _ = primitives.Decode(&decoded, data)
    })
}
```

## 11. Benchmarks

### 9.1 Decode Performance

```go
func BenchmarkDecodePrimitives(b *testing.B) {
    wire := createValidWireFormat()
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var data primitives.AllTypes
        _ = primitives.Decode(&data, wire)
    }
}

func BenchmarkDecodeNested(b *testing.B) {
    wire, _ := os.ReadFile("testdata/fixtures/nested_valid.bin")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var data nested.PluginList
        _ = nested.Decode(&data, wire)
    }
}
```

## 11. Test Execution

### 10.1 Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific test
go test -run TestDecodePrimitives

# Run benchmarks
go test -bench=.

# Run with race detector
go test -race ./...

# Run with coverage
go test -cover ./...
```

### 10.2 CI Configuration

```yaml
# .github/workflows/test.yml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Install C compiler
        run: sudo apt-get install -y gcc
      
      - name: Run tests
        run: go test -v -race -cover ./...
      
      - name: Run benchmarks
        run: go test -bench=. -benchtime=1s ./...
```

---

**End of Testing Strategy**
