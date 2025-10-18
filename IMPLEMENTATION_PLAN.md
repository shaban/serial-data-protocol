# Implementation Plan

**Version:** 1.0.0  
**Date:** October 18, 2025

## 1. Overview

This document breaks down the implementation of Serial Data Protocol v1.0 into small, verifiable, testable tasks.

### 1.1 Implementation Strategy

**Phase 1-4: Foundation & Go Complete (Rock Solid)**
- Build schema parser and validator
- Generate Go decoder (all field types, nested structs, arrays)
- Generate Go encoder (symmetric with decoder)
- Test Go→Go roundtrip in single language
- Validates wire format is rock solid before cross-language complexity

**Phase 5+: Add C Encoder**
- Reuse parser/validator infrastructure
- Generate C11 encoder code
- Cross-language testing (C encoder → Go decoder)

**Rationale:** Complete and test Go encoding/decoding thoroughly before introducing
cross-language complexity. This de-risks the wire format design and provides a
reference implementation for debugging C interop issues later.

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
**Status:** `[✓]`

**Work:**
1. Create `internal/validator/errors.go`:
   - Define standardized error codes for all validation failures
   - Error constructors for consistent error messages
   - Error codes are stable and can be relied upon by tools and other implementations
2. Create `internal/validator/validator.go`:
   - `Validate(schema *Schema) error`
   - Run all validators (structure, types, cycles, naming)
   - Collect all errors (doesn't stop at first error)
   - Return combined error with all issues formatted (one issue per line)
3. Update all validators to use error constructors:
   - structure.go, types.go, cycles.go, naming.go

**Error Codes Defined:**
- `EMPTY_SCHEMA` - schema defines no structs
- `EMPTY_STRUCT` - struct has no fields
- `UNKNOWN_TYPE` - type reference to undefined struct
- `CIRCULAR_REFERENCE` - circular struct reference detected
- `INVALID_IDENTIFIER` - identifier violates naming rules
- `RESERVED_KEYWORD` - identifier is reserved in target language
- `DUPLICATE_STRUCT` - multiple structs with same name
- `DUPLICATE_FIELD` - multiple fields with same name in struct

**Tests:** `internal/validator/validator_test.go` (13 test functions)
```go
func TestValidateValidSchema(t *testing.T)             // Valid schemas pass
func TestIntegrationEmptySchema(t *testing.T)          // Empty schema rejected with code
func TestValidateEmptyStruct(t *testing.T)             // Empty struct with error code
func TestIntegrationUnknownType(t *testing.T)          // Unknown type with code
func TestValidateCycle(t *testing.T)                   // Circular reference with code
func TestValidateReservedKeyword(t *testing.T)         // Reserved keyword with code
func TestValidateDuplicateStructs(t *testing.T)        // Duplicate structs with code
func TestValidateDuplicateFields(t *testing.T)         // Duplicate fields with code
func TestIntegrationMultipleErrors(t *testing.T)       // All errors reported together
func TestValidateComplexValidSchema(t *testing.T)      // Complex valid schema passes
func TestValidateInvalidIdentifier(t *testing.T)       // Invalid identifier with code
func TestValidateArrayOfUnknownType(t *testing.T)      // Array element type validated
```

**Verification:**
- ✓ All validators orchestrated by Validate()
- ✓ All errors collected and reported together
- ✓ Error codes included in all error messages (e.g., [UNKNOWN_TYPE])
- ✓ Clear, formatted error output (one issue per line)
- ✓ Valid schemas pass with nil error
- ✓ Invalid schemas report all issues at once
- ✓ All 59 validator tests passing
- ✓ Validator package coverage: 95.9%
- ✓ Error codes documented for other implementations

**Time:** 2 hours

---

## 5. Phase 4: Go Decoder Generator

**Goal:** Generate complete Go decoder from schema (structs, decode functions, errors, context).

**Status:** `[✓]` Complete - All 10 tasks finished with 94.7% test coverage

### Task 4.1: Type Name Mapping
**Status:** `[✓]`

**Work:**
1. Create `internal/generator/golang/types.go` (package renamed from 'go' to avoid keyword):
   - `MapTypeToGo(typeExpr *TypeExpr) (string, error)`
   - Map all 12 primitive types: `u32` → `uint32`, `str` → `string`, etc.
   - Handle arrays recursively: `[]T` → `[]GoType`
   - Preserve named types as-is (defer PascalCase to Task 4.2)
   - Return errors (not panic, not os.Exit) for testability

**Tests:** `internal/generator/golang/types_test.go` (10 test functions, 35 subtests)
```go
func TestMapPrimitiveTypes(t *testing.T)        // All 12 primitives with subtests
func TestMapUnknownPrimitive(t *testing.T)      // Error handling
func TestMapNamedTypes(t *testing.T)            // Struct names preserved (various cases)
func TestMapArrayOfPrimitives(t *testing.T)     // Arrays of primitives
func TestMapArrayOfStructs(t *testing.T)        // Arrays of named types
func TestMapNestedArrays(t *testing.T)          // Multi-dimensional arrays
func TestMapArrayWithoutElement(t *testing.T)   // Malformed array error
func TestMapArrayWithInvalidElement(t *testing.T) // Error propagation
func TestMapNilTypeExpr(t *testing.T)           // Nil input handling
func TestMapUnknownTypeKind(t *testing.T)       // Invalid type kind
```

**Verification:**
- ✓ All 12 primitive type mappings correct per DESIGN_SPEC.md Section 3.3
- ✓ Array types mapped recursively (including nested arrays)
- ✓ Named types preserved as-is (snake_case, camelCase, PascalCase)
- ✓ Comprehensive error handling (unknown types, nil inputs, malformed arrays)
- ✓ All 35 tests passing with 100% coverage
- ✓ Follows testing strategy: unit tests only, integration deferred

**Time:** 1 hour

---

### Task 4.2: Name Conversion
**Status:** `[✓]`

**Work:**
1. Add to `internal/generator/golang/types.go`:
   - `ToGoName(name string) string` - Unified PascalCase conversion
   - `capitalizeFirst(s string) string` - Helper for first letter capitalization
   - Works for both struct names and field names (both exported in Go)
   - Preserve existing capitals (HTTPResponse → HTTPResponse)
   - Handle snake_case (audio_device → AudioDevice)
   - Handle multiple/leading/trailing underscores
   - Unicode support (école → École)

**Tests:** `internal/generator/golang/types_test.go` (4 test functions, 42 subtests)
```go
func TestToGoName(t *testing.T)              // 24 cases: snake_case, capitals, underscores, edge cases
func TestToGoNamePreservesCase(t *testing.T) // 6 cases: HTTP, ID, camelCase preservation
func TestToGoNameWithSnakeCase(t *testing.T) // 6 real-world examples
func TestCapitalizeFirst(t *testing.T)       // 8 cases: basic, Unicode, edge cases
```

**Verification:**
- ✓ snake_case → PascalCase (audio_device → AudioDevice)
- ✓ Existing capitals preserved (HTTPResponse, deviceID)
- ✓ Single words capitalized (device → Device)
- ✓ Multiple underscores handled (audio__device → AudioDevice)
- ✓ Leading/trailing underscores removed (_device_ → Device)
- ✓ Empty string handled
- ✓ Unicode support (école → École)
- ✓ All 77 tests passing (type mapping + name conversion)
- ✓ 100% coverage maintained

**Time:** 1 hour

---

### Task 4.3: Struct Generator
**Status:** `[✓]`

**Work:**
1. Create `internal/generator/golang/struct_gen.go`:
   - `GenerateStructs(schema *Schema) (string, error)`
   - Generate Go struct definitions with PascalCase conversion
   - Include doc comments from schema
   - Format field comments
   - Handle all field types (primitives, strings, arrays, nested structs)

**Tests:** `internal/generator/golang/struct_gen_test.go` (11 tests)
```go
func TestGenerateSimpleStruct(t *testing.T)
func TestGenerateMultipleStructs(t *testing.T)
func TestGenerateWithArrays(t *testing.T)
func TestGenerateWithNestedArrays(t *testing.T)
func TestGenerateSnakeCaseConversion(t *testing.T)
func TestGenerateWithoutComments(t *testing.T)
func TestGenerateAllPrimitiveTypes(t *testing.T)
func TestGenerateNilSchema(t *testing.T)
func TestGenerateEmptySchema(t *testing.T)
func TestGenerateInvalidFieldType(t *testing.T)
func TestGenerateComplexSchema(t *testing.T)
```

**Verification:**
- ✓ Generated code compiles
- ✓ Doc comments present
- ✓ Field types correct
- ✓ 100% coverage maintained

**Time:** 2 hours (actual)

---

### Task 4.4: Decode Function Entry Points
**Status:** `[✓]`

**Work:**
1. Create `internal/generator/golang/decode_gen.go`:
   - `GenerateDecoder(schema *Schema) (string, error)`
   - Generate public `DecodeStructName(dest *StructName, data []byte) error` functions
   - Entry point validation (128MB size check)
   - Create DecodeContext
   - Call helper decode function

**Tests:** `internal/generator/golang/decode_gen_test.go` (9 tests)
```go
func TestGenerateDecoderSimple(t *testing.T)
func TestGenerateDecoderMultipleStructs(t *testing.T)
func TestGenerateDecoderSnakeCaseConversion(t *testing.T)
func TestGenerateDecoderValidation(t *testing.T)
func TestGenerateDecoderDocComments(t *testing.T)
func TestGenerateDecoderComplexSchema(t *testing.T)
func TestGenerateDecoderNilSchema(t *testing.T)
func TestGenerateDecoderEmptySchema(t *testing.T)
func TestGenerateDecoderFunctionStructure(t *testing.T)
```

**Verification:**
- ✓ Function signature correct (public Decode functions)
- ✓ 128MB size validation present
- ✓ DecodeContext creation
- ✓ Helper function calls with proper parameters
- ✓ 100% coverage

**Time:** 2 hours (actual)

---

### Task 4.5: Primitive Decode Logic
**Status:** `[✓]`

**Work:**
1. Add to `internal/generator/golang/decode_gen.go`:
   - `GenerateDecodeHelpers(schema *Schema) (string, error)` 
   - Generate helper decode functions for each struct
   - Generate primitive decode code for all 11 types:
     - u8, u16, u32, u64, i8, i16, i32, i64, f32, f64, bool
   - Include bounds checks before every read
   - Proper offset tracking (pointer to int)
   - Little-endian decoding via binary.LittleEndian

**Tests:** `internal/generator/golang/decode_gen_test.go` (12 new tests, 21 total)
```go
func TestGenerateDecodeHelpersSimple(t *testing.T)
func TestGenerateDecodeHelpersAllPrimitives(t *testing.T)
func TestGenerateDecodeHelpersMultipleFields(t *testing.T)
func TestGenerateDecodeHelpersMultipleStructs(t *testing.T)
func TestGenerateDecodeHelpersBoundsChecks(t *testing.T)
// ... 7 more tests
```

**Verification:**
- ✓ All 11 primitive types decode correctly
- ✓ Bounds checks present for every field
- ✓ Correct byte sizes (u32=4, u64=8, etc.)
- ✓ Little-endian decoding
- ✓ Offset tracking via pointer
- ✓ 100% coverage maintained

**Time:** 2 hours (actual)

---

### Task 4.6: String Decode Logic
**Status:** `[✓]`

**Work:**
1. Add string decoding to `decode_gen.go`:
   - Generate string field decode code
   - Two bounds checks: length prefix + string bytes
   - Length-prefixed format: `[u32: length][utf8_bytes]`
   - Proper offset advancement

**Tests:** `internal/generator/golang/decode_gen_test.go` (6 new tests, 27 total)
```go
func TestGenerateDecodeHelpersWithString(t *testing.T)
func TestGenerateDecodeHelpersWithMultipleStrings(t *testing.T)
func TestGenerateDecodeHelpersMixedTypes(t *testing.T)
func TestGenerateDecodeHelpersStringBoundsChecks(t *testing.T)
func TestGenerateDecodeHelpersAllTypesIncludingString(t *testing.T)
func TestGenerateDecodeHelpersStringSnakeCase(t *testing.T)
```

**Verification:**
- ✓ Length-prefixed strings work
- ✓ Empty strings work
- ✓ Two bounds checks present
- ✓ UTF-8 bytes correctly extracted
- ✓ 100% coverage maintained

**Time:** 1 hour (actual)

---

### Task 4.7: Array Decode Logic
**Status:** `[✓]`

**Work:**
1. Add array decoding to `decode_gen.go`:
   - Generate array field decode code
   - Read count (u32)
   - Call `ctx.checkArraySize(count)` for validation
   - Allocate slice
   - Loop through elements
   - Decode each element (primitives and strings)
   - Support all 12 array element types

**Tests:** `internal/generator/golang/decode_gen_test.go` (8 new tests, 35 total)
```go
func TestGenerateDecodeHelpersWithPrimitiveArray(t *testing.T)
func TestGenerateDecodeHelpersWithStringArray(t *testing.T)
func TestGenerateDecodeHelpersWithMultiplePrimitiveArrays(t *testing.T)
func TestGenerateDecodeHelpersMixedFieldsWithArrays(t *testing.T)
func TestGenerateDecodeHelpersArrayCheckCount(t *testing.T)
func TestGenerateDecodeHelpersArrayLoopStructure(t *testing.T)
func TestGenerateDecodeHelpersAllArrayPrimitiveTypes(t *testing.T)
func TestGenerateDecodeHelpersArrayElementOrder(t *testing.T)
```

**Verification:**
- ✓ Empty arrays work
- ✓ Size limits enforced via ctx.checkArraySize()
- ✓ Element decoding correct for all 12 types
- ✓ Proper loop structure
- ✓ 100% coverage maintained

**Time:** 2 hours (actual)

---

### Task 4.8: Nested Struct Decode Logic
**Status:** `[✓]`

**Work:**
1. Add nested struct decoding to `decode_gen.go`:
   - Generate code to call helper decode functions for nested structs
   - `generateNamedTypeDecode()` - handles struct fields
   - `generateArrayNamedTypeElementDecode()` - handles arrays of structs
   - Recursive helper calls: `decodeStructName(&dest.Field, data, offset, ctx)`
   - Pass offset pointer and context through recursion
   - Proper error propagation

**Tests:** `internal/generator/golang/decode_gen_test.go` (9 new tests, 44 total)
```go
func TestGenerateDecodeHelpersWithNamedType(t *testing.T)
func TestGenerateDecodeHelpersWithMultipleNamedTypes(t *testing.T)
func TestGenerateDecodeHelpersMixedWithNamedTypes(t *testing.T)
func TestGenerateDecodeHelpersNamedTypeSnakeCase(t *testing.T)
func TestGenerateDecodeHelpersWithNamedTypeArray(t *testing.T)
func TestGenerateDecodeHelpersComplexNesting(t *testing.T)
func TestGenerateDecodeHelpersNamedTypePassesContext(t *testing.T)
func TestGenerateDecodeHelpersArrayNamedTypePassesContext(t *testing.T)
```

**Verification:**
- ✓ Nested structs decode correctly
- ✓ Arrays of structs decode correctly
- ✓ Offset tracking works through recursion
- ✓ Context passed through for array validation
- ✓ Error propagation works
- ✓ 94.3% coverage (slightly lower due to complexity)

**Time:** 2 hours (actual)

---

### Task 4.9: Error Types Generator
**Status:** `[✓]`

**Work:**
1. Create `internal/generator/golang/errors_gen.go`:
   - `GenerateErrors() string`
   - Generate all 5 error variables from DESIGN_SPEC.md Section 5.4:
     ```go
     var (
         ErrUnexpectedEOF      = errors.New("unexpected end of data")
         ErrInvalidUTF8        = errors.New("invalid UTF-8 string")
         ErrDataTooLarge       = errors.New("data exceeds 128MB limit")
         ErrArrayTooLarge      = errors.New("array count exceeds per-array limit")
         ErrTooManyElements    = errors.New("total elements exceed limit")
     )
     ```

**Tests:** `internal/generator/golang/errors_gen_test.go` (8 tests)
```go
func TestGenerateErrors(t *testing.T)
func TestGenerateErrorsMessages(t *testing.T)
func TestGenerateErrorsFormat(t *testing.T)
func TestGenerateErrorsStructure(t *testing.T)
func TestGenerateErrorsAlignment(t *testing.T)
func TestGenerateErrorsOrder(t *testing.T)
func TestGenerateErrorsNoExtraContent(t *testing.T)
func TestGenerateErrorsMatchesDesignSpec(t *testing.T)
```

**Verification:**
- ✓ All 5 error types from DESIGN_SPEC.md Section 5.4
- ✓ Error messages match spec exactly
- ✓ Proper alignment and formatting
- ✓ 94.4% coverage maintained

**Time:** 30 minutes (actual)

---

### Task 4.10: DecodeContext Generator
**Status:** `[✓]`

**Work:**
1. Create `internal/generator/golang/context_gen.go`:
   - `GenerateDecodeContext() string`
   - Generate 3 constants from DESIGN_SPEC.md Section 5.5:
     - `MaxSerializedSize = 128 * 1024 * 1024`
     - `MaxArrayElements = 1_000_000`
     - `MaxTotalElements = 10_000_000`
   - Generate DecodeContext type with totalElements field
   - Generate checkArraySize method:
     - Validates count against MaxArrayElements
     - Accumulates totalElements
     - Validates against MaxTotalElements
     - Returns appropriate errors

**Tests:** `internal/generator/golang/context_gen_test.go` (13 tests)
```go
func TestGenerateDecodeContext(t *testing.T)
func TestGenerateDecodeContextConstants(t *testing.T)
func TestGenerateDecodeContextConstantValues(t *testing.T)
func TestGenerateDecodeContextTypeStructure(t *testing.T)
func TestGenerateDecodeContextMethodSignature(t *testing.T)
func TestGenerateDecodeContextMethodLogic(t *testing.T)
func TestGenerateDecodeContextComments(t *testing.T)
func TestGenerateDecodeContextConstantBlock(t *testing.T)
func TestGenerateDecodeContextMatchesDesignSpec(t *testing.T)
func TestGenerateDecodeContextOrder(t *testing.T)
func TestGenerateDecodeContextNoExtraContent(t *testing.T)
func TestGenerateDecodeContextMethodReturnPaths(t *testing.T)
func TestGenerateDecodeContextAlignment(t *testing.T)
```

**Verification:**
- ✓ Limits match DESIGN_SPEC.md Section 5.5 exactly
- ✓ Context tracks total elements correctly
- ✓ checkArraySize logic matches spec
- ✓ All 3 constants generated
- ✓ 94.7% coverage (phase 4 complete!)

**Time:** 1 hour (actual)

---

## 5.5. Phase 4 Summary

**Status:** `[✓]` Complete

**Accomplishments:**
- ✅ 10 tasks completed
- ✅ 6 generator files created (types, struct_gen, decode_gen, errors_gen, context_gen + tests)
- ✅ 60 tests passing (35 subtests in types, 11 struct, 39 decode, 8 errors, 13 context)
- ✅ 94.7% test coverage
- ✅ All generated code matches DESIGN_SPEC.md exactly
- ✅ Supports all field types: primitives, strings, arrays, nested structs
- ✅ Complete error handling and size validation

**Total Time:** ~13 hours actual vs ~16 hours estimated

**Files Created:**
- `internal/generator/golang/types.go` + `types_test.go` (140 + 670 lines)
- `internal/generator/golang/struct_gen.go` + `struct_gen_test.go` (118 + 365 lines)
- `internal/generator/golang/decode_gen.go` + `decode_gen_test.go` (635 + 1350 lines)
- `internal/generator/golang/errors_gen.go` + `errors_gen_test.go` (20 + 175 lines)
- `internal/generator/golang/context_gen.go` + `context_gen_test.go` (46 + 260 lines)

**Next:** Phase 4.5 - Go Encoder Generator

---

## 5.5. Phase 4.5: Go Encoder Generator

**Goal:** Generate complete Go encoder from schema (symmetric with decoder).

**Rationale:** Complete Go encoding before C to enable Go→Go roundtrip testing.
This validates the wire format is rock solid before cross-language complexity.

**Performance Strategy:** Pre-calculated Size + Direct Writes
- **Approach**: Calculate size → Single allocation → Direct buffer writes
- **Rationale**: Optimal for FFI hot path (2x faster than bytes.Buffer)
- **Benefits**: Single allocation, predictable performance, zero-copy return
- **Cost**: Size calculation adds ~50ns overhead (negligible vs encoding work)

### Task 4.11: Encoder Entry Points
**Status:** `[→]`

**Work:**
1. Create `internal/generator/golang/encode_gen.go`:
   - `GenerateEncoder(schema *Schema) (string, error)`
   - Generate public `EncodeStructName(src *StructName) ([]byte, error)` functions
   - Generate size calculation helpers: `calculateStructNameSize(src *StructName) int`
   - Single allocation: `buf := make([]byte, size)`
   - Call helper encode function with buffer and offset pointer
   - Return buffer (zero-copy, ownership transfer)

**Example output:**
```go
// Size calculation (fast traversal)
func calculateDeviceSize(src *Device) int {
    size := 4  // ID field (u32)
    size += 4 + len(src.Name)  // Name: length prefix + bytes
    size += 1  // Active field (bool)
    return size
}

// Public encoder (single allocation)
func EncodeDevice(src *Device) ([]byte, error) {
    size := calculateDeviceSize(src)
    buf := make([]byte, size)
    offset := 0
    if err := encodeDevice(src, buf, &offset); err != nil {
        return nil, err
    }
    return buf, nil
}
```

**Tests:** `internal/generator/golang/encode_gen_test.go`
```go
func TestGenerateEncoderSimple(t *testing.T)
func TestGenerateEncoderMultipleStructs(t *testing.T)
func TestGenerateEncoderSnakeCaseConversion(t *testing.T)
func TestGenerateEncoderDocComments(t *testing.T)
func TestGenerateEncoderSignature(t *testing.T)
func TestGenerateEncoderSizeCalculation(t *testing.T)
func TestGenerateEncoderNilSchema(t *testing.T)
func TestGenerateEncoderEmptySchema(t *testing.T)
```

**Verification:**
- ✓ Function signature correct: `func EncodeStructName(src *StructName) ([]byte, error)`
- ✓ Size calculation function generated
- ✓ Single allocation with exact size
- ✓ Helper function calls with offset pointer
- ✓ Doc comments present
- ✓ Nil/empty schema handling

**Time:** 2 hours

---

### Task 4.12: Encoder Field Logic
**Status:** `[ ]`

**Work:**
1. Add to `encode_gen.go`:
   - `GenerateEncodeHelpers(schema *Schema) (string, error)`
   - Generate helper encode functions: `func encodeStructName(src *StructName, buf []byte, offset *int) error`
   - Generate encode code for all field types using direct buffer writes:
     
     **Primitives (11 types):** Direct binary.LittleEndian writes
     ```go
     binary.LittleEndian.PutUint32(buf[*offset:], src.ID)
     *offset += 4
     ```
     
     **Strings:** Write length prefix + copy bytes
     ```go
     binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Name)))
     *offset += 4
     copy(buf[*offset:], src.Name)
     *offset += len(src.Name)
     ```
     
     **Arrays:** Write count + encode elements in loop
     ```go
     binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Items)))
     *offset += 4
     for i := range src.Items {
         // encode element
     }
     ```
     
     **Nested structs:** Call helper encode functions recursively
     ```go
     if err := encodePlugin(&src.Plugin, buf, offset); err != nil {
         return err
     }
     ```
   
   - Proper offset tracking via pointer (matches decoder pattern)
   - No error handling needed for buffer writes (pre-sized buffer)
   - Error returns only for consistency (always return nil for primitives)

**Tests:** `internal/generator/golang/encode_gen_test.go`
```go
func TestGenerateEncodeHelpersSimple(t *testing.T)
func TestGenerateEncodeHelpersAllPrimitives(t *testing.T)
func TestGenerateEncodeHelpersWithString(t *testing.T)
func TestGenerateEncodeHelpersWithArray(t *testing.T)
func TestGenerateEncodeHelpersWithNamedType(t *testing.T)
func TestGenerateEncodeHelpersArrayOfStructs(t *testing.T)
func TestGenerateEncodeHelpersMixedFields(t *testing.T)
func TestGenerateEncodeHelpersComplexNesting(t *testing.T)
func TestGenerateEncodeHelpersOffsetTracking(t *testing.T)
func TestGenerateEncodeHelpersMultipleStructs(t *testing.T)
```

**Verification:**
- ✓ All 11 primitive types encode correctly with binary.LittleEndian
- ✓ Strings use length prefix + copy
- ✓ Arrays write count then elements
- ✓ Nested structs call helpers recursively
- ✓ Offset tracking via pointer (matches decoder)
- ✓ Generated code mirrors decoder logic
- ✓ Direct buffer writes (no interface overhead)
- ✓ 100% coverage

**Time:** 3 hours

---

### Task 4.13: Encoder Tests
**Status:** `[ ]`

**Work:**
1. Complete test coverage for encoder generator:
   - Test all primitive types
   - Test strings (empty, normal, UTF-8)
   - Test arrays (empty, single element, multiple elements)
   - Test nested structs
   - Test arrays of structs
   - Test mixed field types
   - Test snake_case conversion
   - Test error cases

**Tests:** Additional tests in `encode_gen_test.go` (targeting 15-20 tests total)
```go
func TestGenerateEncodeHelpersAllArrayTypes(t *testing.T)
func TestGenerateEncodeHelpersEmptyArray(t *testing.T)
func TestGenerateEncodeHelpersNameConversion(t *testing.T)
func TestGenerateEncodeHelpersDocComments(t *testing.T)
// ... more edge cases
```

**Verification:**
- ✓ 100% line coverage
- ✓ All edge cases covered
- ✓ Error paths tested
- ✓ Complex schemas tested

**Time:** 2 hours

---

### Task 4.14: Go→Go Integration Test
**Status:** `[ ]`

**Work:**
1. Create `internal/generator/golang/integration_test.go`:
   - Full pipeline test: Schema → Generate → Compile → Encode → Decode → Compare
   - Test multiple schemas:
     - Simple struct with primitives
     - Struct with strings
     - Struct with arrays
     - Nested structs
     - Complex real-world example (Plugin with Parameters)
   - Verify roundtrip: `original == decoded`
   - Test wire format directly (inspect bytes)

**Tests:**
```go
func TestGoRoundtripSimplePrimitives(t *testing.T)
func TestGoRoundtripWithStrings(t *testing.T)
func TestGoRoundtripWithArrays(t *testing.T)
func TestGoRoundtripNestedStructs(t *testing.T)
func TestGoRoundtripComplexPlugin(t *testing.T)
func TestGoWireFormatBytes(t *testing.T)  // Inspect actual wire bytes
func TestGoEncodeDecodeErrorConditions(t *testing.T)
```

**Verification:**
- ✓ All roundtrips succeed (original == decoded)
- ✓ Wire format matches DESIGN_SPEC.md
- ✓ Generated code compiles
- ✓ No panics or crashes
- ✓ Error conditions handled properly
- ✓ Performance acceptable (benchmark optional)

**Time:** 2-3 hours

---

## 5.6. Phase 4.5 Summary

**Estimated Time:** ~9-10 hours total
- Task 4.11: 2 hours
- Task 4.12: 3 hours
- Task 4.13: 2 hours
- Task 4.14: 2-3 hours

**Deliverables:**
- Complete Go encoder generator
- ~20 encoder generator tests
- Full Go→Go integration tests
- Wire format validation
- Rock solid Go implementation before C

**Success Criteria:**
- ✅ All tests passing
- ✅ >94% coverage maintained
- ✅ Roundtrip tests prove wire format correctness
- ✅ Generated Go code is production-ready
- ✅ Ready to proceed to C encoder with confidence

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
