# Implementation Plan

**Version:** 1.0.0  
**Date:** October 18, 2025

## 1. Overview

This document breaks down the implementation of Serial Data Protocol v1.0 into small, verifiable, testable tasks.

### 1.1 Implementation Strategy

**Phase 1-2: Go-to-Go (Proof of Concept)**
- Build schema parser and validator
- Generate Go encoder with tentative structures
- Generate Go decoder
- Test roundtrip in single language
- Validates wire format and tentative structure design

**Phase 3+: Add C Encoder**
- Reuse parser/validator infrastructure
- Generate C11 encoder code
- Cross-language testing (C encoder → Go decoder)

This approach de-risks the design before dealing with FFI complexity.

### 1.2 Principles

- **Small steps** - Each task is 1-4 hours of work
- **Verifiable** - Each task has clear completion criteria
- **Testable** - Each task includes tests
- **Sequential** - Tasks build on previous tasks
- **No skipping** - Complete each phase before moving to next

### 1.3 Technical Decisions

- **Go Module**: `github.com/shaban/serial-data-protocol`
- **License**: MIT
- **C Standard**: C11 (for `_Generic`, `static_assert`, anonymous structs)
- **Initial Target**: Go encoder + Go decoder (same-language proof of concept)
- **Repository**: Local git until working, then push to GitHub

### 1.4 Task Status

Each task has a status:
- `[ ]` Not started
- `[→]` In progress
- `[✓]` Complete
- `[✗]` Blocked

## 2. Phase 1: Foundation (Bootstrap)

**Goal:** Set up project structure and wire format basics.

### Task 1.1: Project Initialization
**Status:** `[✓]`

**Work:**
1. Initialize git repository
2. Create Go module: `go mod init github.com/shaban/serial-data-protocol`
3. Create directory structure:
   ```
   cmd/sdp-gen/
   internal/wire/
   internal/parser/
   internal/validator/
   internal/generator/go/
   internal/generator/c/     (for later)
   testdata/
   ```
4. Create `.gitignore` per TESTING_STRATEGY.md
5. Add MIT LICENSE file

**Tests:** None (infrastructure)

**Verification:**
- ✓ `git status` shows clean repository
- ✓ `go mod verify` succeeds
- ✓ Directory structure matches spec
- ✓ Git ignores generated files

**Time:** 30 minutes

---

### Task 1.2: Wire Format Primitives
**Status:** `[✓]`

**Work:**
1. Create `internal/wire/encode.go`:
   - `EncodeU8(buf []byte, offset int, val uint8)`
   - `EncodeU16(buf []byte, offset int, val uint16)`
   - `EncodeU32(buf []byte, offset int, val uint32)`
   - `EncodeU64(buf []byte, offset int, val uint64)`
   - `EncodeI8, EncodeI16, EncodeI32, EncodeI64`
   - `EncodeF32, EncodeF64`
   - `EncodeBool`
   - All use little-endian

2. Create `internal/wire/decode.go`:
   - Corresponding `DecodeXXX` functions

**Tests:** `internal/wire/wire_test.go`
```go
func TestEncodeDecodeU32(t *testing.T)
func TestEncodeDecodeF64(t *testing.T)
func TestEncodeBool(t *testing.T)
func TestLittleEndian(t *testing.T)
```

**Verification:**
- ✓ All primitive types encode/decode correctly
- ✓ Little-endian byte order verified
- ✓ 100% test coverage for wire package

**Time:** 2 hours

---

### Task 1.3: String Wire Format
**Status:** `[✓]`

**Work:**
1. Add to `internal/wire/encode.go`:
   - `EncodeString(w io.Writer, s string) (int, error)`
   - Writes: `[u32: length][utf8_bytes]`

2. Add to `internal/wire/decode.go`:
   - `DecodeString(r io.Reader) (string, error)`
   - Validates sufficient bytes
   - Returns string

3. Add array header helpers:
   - `EncodeArrayHeader(w io.Writer, count uint32) (int, error)`
   - `DecodeArrayHeader(r io.Reader) (uint32, error)`

**Tests:** `internal/wire/wire_test.go`
```go
func TestEncodeDecodeString(t *testing.T)
func TestEncodeDecodeStringFormat(t *testing.T)
func TestDecodeStringErrors(t *testing.T)
func TestEncodeDecodeArrayHeader(t *testing.T)
func TestDecodeArrayHeaderErrors(t *testing.T)
```

**Verification:**
- ✓ Empty strings work
- ✓ UTF-8 strings work
- ✓ EOF detection works
- ✓ Array count encoding works
- ✓ 98.1% test coverage

**Time:** 1.5 hours (combined Task 1.3 and 1.4)

---

## 3. Phase 2: Schema Parser

**Goal:** Parse `.sdp` schema files into AST.

### Task 2.1: Schema AST Types
**Status:** `[✓]`

**Work:**
1. Create `internal/parser/ast.go`:
   - Schema, Struct, Field types
   - TypeExpr for representing type expressions
   - TypeKind for distinguishing primitive/named/array types
   - Helper methods: IsPrimitive(), String()

**Tests:** `internal/parser/ast_test.go`
```go
func TestTypeExprIsPrimitive(t *testing.T)
func TestTypeExprString(t *testing.T)
func TestSchemaStructure(t *testing.T)
```

**Verification:**
- ✓ Types compile
- ✓ Fields match DESIGN_SPEC.md Section 3
- ✓ 80% test coverage

**Time:** 30 minutes

---

### Task 2.2: Lexer
**Status:** `[✓]`

**Work:**
1. Create `internal/parser/lexer.go`:
   - Tokenize schema file
   - Handle `///` doc comments
   - Handle `//` regular comments
   - Recognize keywords: `struct`
   - Recognize operators: `{`, `}`, `,`, `:`
   - Recognize identifiers
   - Handle `[]` for arrays
   - Track line and column numbers

**Tests:** `internal/parser/lexer_test.go`
```go
func TestLexStruct(t *testing.T)
func TestLexDocComment(t *testing.T)
func TestLexArray(t *testing.T)
func TestLexIdentifiers(t *testing.T)
func TestLexKeyword(t *testing.T)
func TestLexComments(t *testing.T)
func TestLexLineNumbers(t *testing.T)
func TestLexError(t *testing.T)
func TestLexEmpty(t *testing.T)
func TestLexWhitespace(t *testing.T)
```

**Verification:**
- ✓ All token types recognized
- ✓ Comments extracted correctly
- ✓ Line numbers tracked
- ✓ All tests pass

**Time:** 3 hours

---

### Task 2.3: Parser
**Status:** `[✓]`

**Work:**
1. Create `internal/parser/parser.go`:
   - `ParseSchema(input string) (*Schema, error)`
   - Parse struct definitions
   - Parse field lists
   - Associate doc comments with declarations
   - Handle commas (optional after last field)
   - Distinguish regular comments from doc comments

**Tests:** `internal/parser/parser_test.go`
```go
func TestParseSimpleStruct(t *testing.T)
func TestParseNestedTypes(t *testing.T)
func TestParseArrayField(t *testing.T)
func TestParseDocComments(t *testing.T)
func TestParseSyntaxError(t *testing.T)
func TestParseEmptyStruct(t *testing.T)
func TestParseMultipleStructs(t *testing.T)
func TestParseTrailingComma(t *testing.T)
func TestParseAllPrimitives(t *testing.T)
func TestParseWithComments(t *testing.T)
```

**Verification:**
- ✓ Valid schemas parse correctly
- ✓ Invalid schemas return clear errors
- ✓ Doc comments preserved
- ✓ Error messages include line numbers
- ✓ All tests pass

**Time:** 4 hours

---

### Task 2.4: Schema File Loader
**Status:** `[✓]`

**Work:**
1. Create `internal/parser/loader.go`:
   - `LoadSchemaFile(path string) (*Schema, error)`
   - Read file
   - Normalize line endings (CRLF → LF)
   - Parse schema
   - Return errors with filename

**Tests:** `internal/parser/loader_test.go`
```go
func TestLoadSchemaFile_Valid(t *testing.T)           // basic & complex schemas
func TestLoadSchemaFile_MissingFile(t *testing.T)
func TestLoadSchemaFile_InvalidSyntax(t *testing.T)
func TestLoadSchemaFile_CRLF(t *testing.T)
func TestLoadSchemaFile_PreservesDocComments(t *testing.T)
```

**Verification:**
- ✓ Files load correctly
- ✓ Line ending normalization works
- ✓ Errors include filename
- ✓ Doc comments preserved
- ✓ All tests pass (89.2% parser coverage)

**Time:** 1 hour

---

## 4. Phase 3: Schema Validator

**Goal:** Validate schemas per DESIGN_SPEC.md Section 3.5.

### Task 3.1: Reserved Keywords List
**Status:** `[ ]`

**Work:**
1. Create `internal/validator/reserved.go`:
   - Lists of reserved words for Go, Rust, Swift, C
   - `IsReserved(word string) bool`
   - `GetReservedLanguage(word string) string` (returns which language)

**Tests:** `internal/validator/reserved_test.go`
```go
func TestGoReservedWords(t *testing.T)
func TestRustReservedWords(t *testing.T)
func TestIsReservedCaseInsensitive(t *testing.T)
```

**Verification:**
- All reserved words from spec included
- Case-insensitive matching works

**Time:** 1 hour

---

### Task 3.2: Type Reference Validator
**Status:** `[✓]`

**Work:**
1. Create `internal/validator/types.go`:
   - `ValidateTypeReferences(schema *Schema) []error`
   - Check all field types resolve to:
     - Primitive (u8, u16, u32, u64, i8, i16, i32, i64, f32, f64, bool, str)
     - Struct defined in schema
     - Array of valid type `[]T`
   - Return all errors (don't stop at first)

**Tests:** `internal/validator/types_test.go`
```go
func TestValidateKnownTypes(t *testing.T)
func TestValidatePrimitiveTypes(t *testing.T)
func TestValidateUnknownType(t *testing.T)
func TestValidateArrayType(t *testing.T)         // 6 subtests
func TestValidateMultipleErrors(t *testing.T)
func TestValidateForwardReference(t *testing.T)
func TestValidateEmptySchema(t *testing.T)
func TestValidateComplexNesting(t *testing.T)
```

**Verification:**
- ✓ Known types pass (primitives and struct references)
- ✓ Unknown types rejected with clear error messages
- ✓ Array types validated recursively (including nested)
- ✓ All errors reported together
- ✓ Forward references work (struct can reference later-defined struct)
- ✓ All 8 tests passing (94.9% validator coverage)

**Time:** 1 hour

---

### Task 3.3: Circular Reference Detector
**Status:** `[✓]`

**Work:**
1. Create `internal/validator/cycles.go`:
   - `DetectCycles(schema *Schema) []error`
   - Use DFS with visited set and recursion stack
   - Detect direct self-reference: `struct Node { next: Node }`
   - Detect indirect cycles: `A → B → C → A`
   - Detect cycles through arrays: `struct Node { children: []Node }`
   - Return all cycles found

**Tests:** `internal/validator/cycles_test.go`
```go
func TestNoCycle(t *testing.T)              // 5 subtests
func TestDirectSelfReference(t *testing.T)
func TestIndirectCycle(t *testing.T)        // 3 subtests
func TestMultipleCycles(t *testing.T)
func TestCycleViaArray(t *testing.T)
func TestComplexCycle(t *testing.T)
func TestNoCycleWithSharedDependency(t *testing.T)
func TestEmptySchema(t *testing.T)
func TestCycleWithMixedFields(t *testing.T)
```

**Verification:**
- ✓ Acyclic graphs pass (linear chains, trees, diamonds)
- ✓ All cycles detected (direct, indirect, via arrays)
- ✓ Clear error messages with cycle path (A → B → C → A)
- ✓ Multiple independent cycles found
- ✓ All 11 tests passing (97.5% validator coverage)

**Time:** 1.5 hours

---

### Task 3.4: Naming Validator
**Status:** `[✓]`

**Work:**
1. Create `internal/validator/naming.go`:
   - `ValidateNaming(schema *Schema) []error`
   - Check struct names valid (start with letter/underscore, alphanumeric + underscore)
   - Check field names valid (same rules)
   - Check no reserved words used (via IsReserved())
   - Check no duplicate field names in struct
   - Check no duplicate struct names

**Tests:** `internal/validator/naming_test.go`
```go
func TestValidNames(t *testing.T)
func TestReservedKeyword(t *testing.T)        // 5 subtests
func TestInvalidCharacters(t *testing.T)      // 5 subtests
func TestDuplicateFields(t *testing.T)
func TestDuplicateStructs(t *testing.T)
func TestMultipleErrors(t *testing.T)
func TestCaseSensitiveNames(t *testing.T)
func TestNamingEmptySchema(t *testing.T)
func TestUnicodeIdentifiers(t *testing.T)     // Confirms ASCII-only
func TestUnderscorePrefix(t *testing.T)
func TestMixedValidAndInvalid(t *testing.T)
```

**Verification:**
- ✓ Valid names pass (snake_case, camelCase, underscores)
- ✓ Reserved keywords rejected (all target languages)
- ✓ Invalid characters rejected (spaces, hyphens, special chars)
- ✓ Duplicate names detected (structs and fields)
- ✓ Clear error messages with context
- ✓ All 12 tests passing (93.9% validator coverage)

**Time:** 1.5 hours

**Tests:** `internal/validator/naming_test.go`
```go
func TestValidNames(t *testing.T)
func TestReservedKeyword(t *testing.T)
func TestInvalidCharacters(t *testing.T)
func TestDuplicateFields(t *testing.T)
func TestDuplicateStructs(t *testing.T)
```

**Verification:**
- Valid names pass
- All invalid cases caught
- Clear error messages

**Time:** 2 hours

---

### Task 3.5: Structure Validator
**Status:** `[✓]`

**Work:**
1. Create `internal/validator/structure.go`:
   - `ValidateStructure(schema *Schema) []error`
   - Check no empty structs (at least 1 field per struct)
   - Check at least one struct defined in schema
   - Return all errors found (doesn't stop at first error)

**Tests:** `internal/validator/structure_test.go` (8 test functions)
```go
func TestStructureEmptySchema(t *testing.T)     // Schema with no structs rejected
func TestEmptyStruct(t *testing.T)              // Struct with no fields rejected
func TestMultipleEmptyStructs(t *testing.T)     // All empty structs reported
func TestValidSingleStruct(t *testing.T)        // Single valid struct passes
func TestValidMultipleStructs(t *testing.T)     // Multiple valid structs pass
func TestSingleFieldStruct(t *testing.T)        // Boundary: exactly 1 field is valid
func TestComplexValidSchema(t *testing.T)       // Arrays, nesting, references
```

**Verification:**
- ✓ Empty schemas rejected (no structs defined)
- ✓ Empty structs rejected (no fields)
- ✓ Multiple empty structs all reported
- ✓ Valid structures pass (single, multiple, complex)
- ✓ Clear error messages with struct names
- ✓ All 8 tests passing
- ✓ 100% coverage on structure.go (37 lines)
- ✓ Validator package: 94.3% coverage overall

**Time:** 1 hour

---

### Task 3.6: Validator Integration
**Status:** `[ ]`

**Work:**
1. Create `internal/validator/validator.go`:
   - `Validate(schema *Schema) error`
   - Run all validators
   - Collect all errors
   - Return combined error with all issues
   - Format error nicely (one issue per line)

**Tests:** `internal/validator/validator_test.go`
```go
func TestValidateAllChecks(t *testing.T)
func TestMultipleErrors(t *testing.T)
func TestValidSchema(t *testing.T)
```

**Verification:**
- All validators run
- All errors collected
- Nice error formatting

**Time:** 1 hour

---

## 5. Phase 4: Go Code Generator

**Goal:** Generate Go decoder from schema.

### Task 4.1: Type Name Mapping
**Status:** `[ ]`

**Work:**
1. Create `internal/generator/go/types.go`:
   - `MapSchemaTypeToGo(schemaType string) string`
   - Map: `u32` → `uint32`, `str` → `string`, etc.
   - Handle arrays: `[]T` → `[]GoType`

**Tests:** `internal/generator/go/types_test.go`
```go
func TestMapPrimitives(t *testing.T)
func TestMapArrays(t *testing.T)
func TestMapStructs(t *testing.T)
```

**Verification:**
- All type mappings correct per DESIGN_SPEC.md Section 3.3

**Time:** 1 hour

---

### Task 4.2: Name Conversion
**Status:** `[ ]`

**Work:**
1. Add to `internal/generator/go/types.go`:
   - `ToGoStructName(name string) string` - PascalCase
   - `ToGoFieldName(name string) string` - PascalCase
   - Preserve existing capitals
   - Handle underscores

**Tests:** `internal/generator/go/types_test.go`
```go
func TestToGoStructName(t *testing.T)
func TestToGoFieldName(t *testing.T)
func TestPreserveCase(t *testing.T)
```

**Verification:**
- snake_case → PascalCase
- Existing capitals preserved
- Single words capitalized

**Time:** 1 hour

---

### Task 4.3: Struct Generator
**Status:** `[ ]`

**Work:**
1. Create `internal/generator/go/struct_gen.go`:
   - `GenerateStructs(schema *Schema) (string, error)`
   - Generate Go struct definitions
   - Include doc comments from schema
   - Format field comments as "FieldName is..."
   - Use `gofmt` or template

**Tests:** `internal/generator/go/struct_gen_test.go`
```go
func TestGenerateSimpleStruct(t *testing.T)
func TestGenerateNestedStruct(t *testing.T)
func TestGenerateDocComments(t *testing.T)
```

**Verification:**
- Generated code compiles
- Doc comments present
- Field types correct

**Time:** 3 hours

---

### Task 4.4: Decode Function Template
**Status:** `[ ]`

**Work:**
1. Create `internal/generator/go/decode_gen.go`:
   - `GenerateDecoder(schema *Schema) (string, error)`
   - Generate `Decode(dest *T, data []byte) error`
   - Entry point validation (size check)
   - Create DecodeContext
   - Call struct decoder

**Tests:** `internal/generator/go/decode_gen_test.go`
```go
func TestGenerateDecoder(t *testing.T)
func TestDecoderSignature(t *testing.T)
```

**Verification:**
- Function signature correct
- Imports included

**Time:** 2 hours

---

### Task 4.5: Primitive Decode Logic
**Status:** `[ ]`

**Work:**
1. Add to `internal/generator/go/decode_gen.go`:
   - Generate primitive decode code:
     ```go
     if offset + 4 > len(data) {
         return ErrUnexpectedEOF
     }
     dest.Field = binary.LittleEndian.Uint32(data[offset:])
     offset += 4
     ```
   - For all primitive types
   - Include bounds checks

**Tests:** `internal/generator/go/decode_gen_test.go`
```go
func TestGeneratePrimitiveDecode(t *testing.T)
func TestGenerateBoundsCheck(t *testing.T)
```

**Verification:**
- Bounds checks present
- Correct byte sizes
- Little-endian

**Time:** 2 hours

---

### Task 4.6: String Decode Logic
**Status:** `[ ]`

**Work:**
1. Add string decoding to decode generator:
   ```go
   // Read length
   if offset + 4 > len(data) {
       return ErrUnexpectedEOF
   }
   strLen := binary.LittleEndian.Uint32(data[offset:])
   offset += 4
   
   // Read string
   if offset + int(strLen) > len(data) {
       return ErrUnexpectedEOF
   }
   dest.Field = string(data[offset:offset+int(strLen)])
   offset += int(strLen)
   ```

**Tests:** Integration test (generated code test)

**Verification:**
- Length-prefixed strings work
- Empty strings work
- Bounds checks present

**Time:** 1 hour

---

### Task 4.7: Array Decode Logic
**Status:** `[ ]`

**Work:**
1. Add array decoding to decode generator:
   ```go
   // Read count
   count := binary.LittleEndian.Uint32(data[offset:])
   offset += 4
   
   // Validate count
   if err := ctx.checkArraySize(count); err != nil {
       return err
   }
   
   // Allocate
   dest.Field = make([]Type, count)
   
   // Decode elements
   for i := 0; i < int(count); i++ {
       // decode element
   }
   ```

**Tests:** Integration test

**Verification:**
- Empty arrays work
- Size limits enforced
- Element decoding correct

**Time:** 2 hours

---

### Task 4.8: Nested Struct Decode Logic
**Status:** `[ ]`

**Work:**
1. Add nested struct decoding:
   - Generate helper function `decodeStructName(data []byte, offset *int, ctx *DecodeContext) (StructName, error)`
   - Call from parent struct decoder
   - Pass offset and context through

**Tests:** Integration test with nested schemas

**Verification:**
- Nested structs decode correctly
- Offset tracking works
- Context passed through

**Time:** 2 hours

---

### Task 4.9: Error Types Generator
**Status:** `[ ]`

**Work:**
1. Create `internal/generator/go/errors_gen.go`:
   - Generate error variables:
     ```go
     var (
         ErrUnexpectedEOF = errors.New("...")
         ErrDataTooLarge = errors.New("...")
         // etc.
     )
     ```

**Tests:** Check errors present in generated code

**Verification:**
- All error types from DESIGN_SPEC.md Section 5.4
- Error messages clear

**Time:** 30 minutes

---

### Task 4.10: DecodeContext Generator
**Status:** `[ ]`

**Work:**
1. Add to decode generator:
   - Generate DecodeContext type
   - Generate checkArraySize method
   - Constants for limits

**Tests:** Integration test

**Verification:**
- Limits match DESIGN_SPEC.md Section 5.5
- Context tracks total elements

**Time:** 1 hour

---

### Task 4.11: Package File Generator
**Status:** `[ ]`

**Work:**
1. Create `internal/generator/go/package.go`:
   - `GeneratePackage(schema *Schema, packageName string) error`
   - Generate `types.go` (structs)
   - Generate `decode.go` (decoder)
   - Generate `metadata.go` (init with unsafe)
   - Add package comment
   - Add "Code generated. DO NOT EDIT." comment

**Tests:** End-to-end generator test

**Verification:**
- All files generated
- Package compiles
- No imports missing

**Time:** 2 hours

---

## 6. Phase 5: C Code Generator

**Goal:** Generate C encoder builder API.

### Task 5.1: C Type Mapping
**Status:** `[ ]`

**Work:**
1. Create `internal/generator/c/types.go`:
   - `MapSchemaTypeToC(schemaType string) string`
   - Map: `u32` → `uint32_t`, `str` → `const char*`, etc.

**Tests:** `internal/generator/c/types_test.go`

**Verification:**
- Type mappings match DESIGN_SPEC.md Section 3.3

**Time:** 1 hour

---

### Task 5.2: Tentative Struct Generator
**Status:** `[ ]`

**Work:**
1. Create `internal/generator/c/tentative_gen.go`:
   - Generate tentative struct definitions
   - Per-field storage with `written` flags
   - Include tentative_base_t
   - Generate for each struct in schema

**Tests:** Check generated C compiles

**Verification:**
- Structs have all fields with flags
- Base included
- Compiles with gcc

**Time:** 3 hours

---

### Task 5.3: Builder Struct Generator
**Status:** `[ ]`

**Work:**
1. Add to C generator:
   - Generate root builder struct
   - Include error fields
   - Include buffer fields

**Tests:** Compiles

**Verification:**
- Matches DESIGN_SPEC.md Section 4.1

**Time:** 1 hour

---

### Task 5.4: New/Destroy Functions
**Status:** `[ ]`

**Work:**
1. Generate:
   - `NewXXXBuilder()` - Allocate builder, initialize buffer
   - `DestroyBuilder()` - Free all memory
   - Initial buffer size: 256 bytes

**Tests:** C compilation + memory leak check

**Verification:**
- No memory leaks (valgrind)
- Proper initialization

**Time:** 2 hours

---

### Task 5.5: Scalar Field Setters
**Status:** `[ ]`

**Work:**
1. Generate setter functions:
   - `SetStructField(tentative_t* t, type value)`
   - Store value
   - Set `written` flag
   - Check for errors first

**Tests:** C compilation + basic test

**Verification:**
- All field types covered
- Error checking present

**Time:** 3 hours

---

### Task 5.6: BeginStruct Functions
**Status:** `[ ]`

**Work:**
1. Generate:
   - `BeginStruct(parent)` - Allocate tentative
   - Link to parent (add to child list)
   - Check nesting depth
   - Check for errors

**Tests:** C compilation + nesting test

**Verification:**
- Nesting tracked
- Depth limit enforced (32)

**Time:** 2 hours

---

### Task 5.7: Discard Functions
**Status:** `[ ]`

**Work:**
1. Generate:
   - `DiscardStruct(parent, tentative)` - Free recursively
   - Unlink from parent
   - Free field storage (strings)
   - Free tentative

**Tests:** Memory leak check

**Verification:**
- No leaks
- Cascading discard works

**Time:** 2 hours

---

### Task 5.8: Commit Logic Generator
**Status:** `[ ]`

**Work:**
1. Generate commit functions:
   - Write fields in schema order
   - Write defaults for omitted fields
   - Write to parent buffer
   - Handle buffer growth

**Tests:** Integration test (encode → decode)

**Verification:**
- Field order correct
- Defaults work
- Buffer growth works

**Time:** 4 hours

---

### Task 5.9: Finalize Function
**Status:** `[ ]`

**Work:**
1. Generate:
   - `FinalizeBuilder()` - Commit all tentatives
   - Return serial_data_t
   - Check for errors

**Tests:** End-to-end encode test

**Verification:**
- Complete data returned
- Error handling works

**Time:** 2 hours

---

### Task 5.10: C Header File Generator
**Status:** `[ ]`

**Work:**
1. Generate header file:
   - All type definitions
   - All function declarations
   - Include guards
   - Doc comments

**Tests:** Header compiles standalone

**Verification:**
- No missing declarations
- Include guards work

**Time:** 2 hours

---

## 7. Phase 6: CLI Tool

**Goal:** Create `sdp-gen` command-line tool.

### Task 6.1: CLI Flags
**Status:** `[ ]`

**Work:**
1. Create `cmd/sdp-gen/main.go`:
   - Use `flag` package
   - Flags: `-schema`, `-output`, `-lang`, `-validate-only`, `-verbose`
   - Help text

**Tests:** Manual CLI test

**Verification:**
- Flags parse correctly
- Help shows usage

**Time:** 1 hour

---

### Task 6.2: CLI Workflow
**Status:** `[ ]`

**Work:**
1. Implement main workflow:
   - Load schema
   - Validate schema
   - Generate code (if not validate-only)
   - Write files
   - Report errors nicely

**Tests:** CLI integration test

**Verification:**
- End-to-end generation works
- Errors reported clearly

**Time:** 2 hours

---

### Task 6.3: Output Directory Creation
**Status:** `[ ]`

**Work:**
1. Add:
   - Create output directory if doesn't exist
   - Check write permissions
   - Handle errors

**Tests:** File I/O test

**Verification:**
- Directories created
- Permissions checked

**Time:** 1 hour

---

## 8. Phase 7: Integration Testing

**Goal:** Test complete encode → decode flow.

### Task 7.1: Test Schema Files
**Status:** `[ ]`

**Work:**
1. Create test schemas:
   - `testdata/primitives.sdp` - All primitive types
   - `testdata/strings.sdp` - String handling
   - `testdata/arrays.sdp` - Array types
   - `testdata/nested.sdp` - Nested structs
   - `testdata/edge_cases.sdp` - Empty arrays, optional fields

**Tests:** Schema validation

**Verification:**
- All schemas valid
- Cover all features

**Time:** 2 hours

---

### Task 7.2: TestMain Setup
**Status:** `[ ]`

**Work:**
1. Create `integration_test.go`:
   - Implement TestMain per TESTING_STRATEGY.md Section 3.1
   - Clean testdata
   - Build generator
   - Generate test packages

**Tests:** TestMain runs

**Verification:**
- Generator builds
- Packages generated
- Tests can import generated packages

**Time:** 2 hours

---

### Task 7.3: Hand-Crafted Wire Format Tests
**Status:** `[ ]`

**Work:**
1. Create tests with hand-crafted wire format:
   - Primitives test
   - Strings test
   - Empty arrays test
   - Verify decoder produces expected values

**Tests:** `TestDecodePrimitives`, `TestDecodeStrings`, etc.

**Verification:**
- Decoder handles hand-crafted data
- All primitive types work
- Strings work

**Time:** 3 hours

---

### Task 7.4: Error Handling Tests
**Status:** `[ ]`

**Work:**
1. Test error cases:
   - Unexpected EOF
   - Data too large
   - Array count too large
   - Total elements exceeded

**Tests:** `TestDecodeErrors`

**Verification:**
- All error conditions caught
- Clear error messages

**Time:** 2 hours

---

### Task 7.5: C Encoder Fixture
**Status:** `[ ]`

**Work:**
1. Create `testdata/encoders/c_encoder.c`:
   - Uses generated C API
   - Reads JSON from stdin
   - Writes binary to stdout
   - Logs to stderr

**Tests:** Cross-language test

**Verification:**
- C encoder compiles
- Produces valid wire format
- Go decoder reads it

**Time:** 3 hours

---

### Task 7.6: Roundtrip Tests
**Status:** `[ ]`

**Work:**
1. Test C encode → Go decode:
   - Build C encoder
   - Run encoder (subprocess)
   - Decode output in Go
   - Verify values

**Tests:** `TestCEncoderRoundtrip`

**Verification:**
- Full roundtrip works
- All types preserved
- Nested structs work

**Time:** 2 hours

---

## 9. Phase 8: Examples

**Goal:** Create working examples for users.

### Task 8.1: Basic Example
**Status:** `[ ]`

**Work:**
1. Create `examples/01-basic/`:
   - Simple device enumeration schema
   - C encoder
   - Go decoder
   - Makefile
   - README with instructions

**Tests:** Example builds and runs

**Verification:**
- Complete working example
- Clear instructions
- Output shown

**Time:** 3 hours

---

### Task 8.2: Nested Structs Example
**Status:** `[ ]`

**Work:**
1. Create `examples/02-nested-structs/`:
   - Plugin enumeration with parameters
   - Shows nested structures
   - Shows arrays

**Tests:** Example works

**Verification:**
- Demonstrates nesting
- Clear and useful

**Time:** 2 hours

---

### Task 8.3: Optional Fields Example
**Status:** `[ ]`

**Work:**
1. Create `examples/03-optional-fields/`:
   - Shows array with 0/1 elements pattern
   - Documents workaround clearly

**Tests:** Example works

**Verification:**
- Optional pattern clear
- Workaround explained

**Time:** 2 hours

---

## 10. Phase 9: Documentation

**Goal:** User-facing documentation.

### Task 9.1: User Guide
**Status:** `[ ]`

**Work:**
1. Create `docs/user-guide.md`:
   - All sections from DOCUMENTATION_GUIDELINES.md Section 5.1
   - Complete with examples
   - Link to DESIGN_SPEC.md for details

**Tests:** None (documentation)

**Verification:**
- All sections complete
- Examples work
- Links valid

**Time:** 4 hours

---

### Task 9.2: API Reference
**Status:** `[ ]`

**Work:**
1. Document generated APIs:
   - Go decoder API
   - C builder API
   - CLI flags
   - Error types

**Tests:** None

**Verification:**
- All APIs documented
- Examples included

**Time:** 3 hours

---

## 11. Phase 10: Polish

**Goal:** Final touches for v1.0 release.

### Task 10.1: Error Message Audit
**Status:** `[ ]`

**Work:**
1. Review all error messages:
   - Parser errors include line numbers
   - Validator errors are actionable
   - Decoder errors are clear
   - Generator errors helpful

**Tests:** Error cases

**Verification:**
- All errors have good messages
- Line numbers present where applicable

**Time:** 2 hours

---

### Task 10.2: Performance Benchmarks
**Status:** `[ ]`

**Work:**
1. Create benchmarks:
   - Decode primitives
   - Decode nested structs
   - Large arrays
   - Document results

**Tests:** `go test -bench=.`

**Verification:**
- Benchmarks run
- Results reasonable
- No obvious performance issues

**Time:** 2 hours

---

### Task 10.3: Fuzzing
**Status:** `[ ]`

**Work:**
1. Create fuzz tests:
   - Fuzz decoder with random data
   - Should never panic
   - Run for 1 hour minimum

**Tests:** `go test -fuzz=.`

**Verification:**
- No panics found
- Errors handled gracefully

**Time:** 3 hours

---

### Task 10.4: Documentation Review
**Status:** `[ ]`

**Work:**
1. Review all canonical documents:
   - Fix typos
   - Verify consistency
   - Check all internal links
   - Update examples if needed

**Tests:** Link checker

**Verification:**
- No broken links
- Consistent terminology
- Clear and accurate

**Time:** 2 hours

---

### Task 10.5: Final Testing Pass
**Status:** `[ ]`

**Work:**
1. Run all tests:
   - Unit tests
   - Integration tests
   - Cross-language tests
   - Examples
   - Verify coverage thresholds

**Tests:** `go test -race -cover ./...`

**Verification:**
- All tests pass
- Coverage > 85%
- No race conditions

**Time:** 1 hour

---

## 12. Release Checklist

**Before tagging v1.0.0:**

- [ ] All tasks marked `[✓]` Complete
- [ ] All tests passing
- [ ] Coverage thresholds met (per CONSTITUTION.md Section 8.1)
- [ ] Examples work
- [ ] Documentation complete
- [ ] No TODO/FIXME in code
- [ ] CONSTITUTION.md followed
- [ ] README.md updated
- [ ] LICENSE file present
- [ ] Git tagged: `v1.0.0`

---

## 13. Time Estimates

| Phase | Tasks | Estimated Time |
|-------|-------|----------------|
| 1. Foundation | 4 | 4 hours |
| 2. Parser | 4 | 9.5 hours |
| 3. Validator | 6 | 10 hours |
| 4. Go Generator | 11 | 19.5 hours |
| 5. C Generator | 10 | 23 hours |
| 6. CLI Tool | 3 | 4 hours |
| 7. Integration Testing | 6 | 14 hours |
| 8. Examples | 3 | 7 hours |
| 9. Documentation | 2 | 7 hours |
| 10. Polish | 5 | 10 hours |
| **Total** | **54 tasks** | **~108 hours** |

**Estimated calendar time:** 3-4 weeks for one developer working full-time.

---

**End of Implementation Plan**
