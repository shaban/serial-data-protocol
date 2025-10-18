package validator

import (
	"strings"
	"testing"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// TestValidateValidSchema verifies that a completely valid schema passes all validators
func TestValidateValidSchema(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Point",
				Fields: []parser.Field{
					{Name: "x", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
					{Name: "y", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
				},
			},
			{
				Name: "Line",
				Fields: []parser.Field{
					{Name: "start", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Point"}},
					{Name: "end", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Point"}},
				},
			},
		},
	}

	err := Validate(schema)
	if err != nil {
		t.Errorf("expected valid schema to pass, got error: %v", err)
	}
}

// TestIntegrationEmptySchema verifies that empty schemas are rejected
func TestIntegrationEmptySchema(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{},
	}

	err := Validate(schema)
	if err == nil {
		t.Fatal("expected error for empty schema, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "schema validation failed") {
		t.Errorf("expected validation failure message, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "EMPTY_SCHEMA") {
		t.Errorf("expected EMPTY_SCHEMA error code, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "at least one struct") {
		t.Errorf("expected error about empty schema, got: %s", errMsg)
	}
}

// TestValidateEmptyStruct verifies that empty structs are rejected
func TestValidateEmptyStruct(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name:   "Empty",
				Fields: []parser.Field{},
			},
		},
	}

	err := Validate(schema)
	if err == nil {
		t.Fatal("expected error for empty struct, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "Empty") {
		t.Errorf("expected error to mention struct 'Empty', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "EMPTY_STRUCT") {
		t.Errorf("expected EMPTY_STRUCT error code, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "cannot be empty") {
		t.Errorf("expected error about empty struct, got: %s", errMsg)
	}
}

// TestIntegrationUnknownType verifies that unknown types are caught
func TestIntegrationUnknownType(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Test",
				Fields: []parser.Field{
					{Name: "field", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "UnknownType"}},
				},
			},
		},
	}

	err := Validate(schema)
	if err == nil {
		t.Fatal("expected error for unknown type, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "UnknownType") {
		t.Errorf("expected error to mention 'UnknownType', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "UNKNOWN_TYPE") {
		t.Errorf("expected UNKNOWN_TYPE error code, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "unknown type") {
		t.Errorf("expected 'unknown type' in error, got: %s", errMsg)
	}
}

// TestValidateCycle verifies that circular references are detected
func TestValidateCycle(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "A",
				Fields: []parser.Field{
					{Name: "b", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "B"}},
				},
			},
			{
				Name: "B",
				Fields: []parser.Field{
					{Name: "a", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "A"}},
				},
			},
		},
	}

	err := Validate(schema)
	if err == nil {
		t.Fatal("expected error for circular reference, got nil")
	}

	errMsg := err.Error()
	// Check for either "cycle" or "circular reference"
	if !strings.Contains(errMsg, "cycle") && !strings.Contains(errMsg, "circular") {
		t.Errorf("expected error about cycle or circular reference, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "CIRCULAR_REFERENCE") {
		t.Errorf("expected CIRCULAR_REFERENCE error code, got: %s", errMsg)
	}
	// Check for cycle path (should mention both A and B)
	if !strings.Contains(errMsg, "A") || !strings.Contains(errMsg, "B") {
		t.Errorf("expected cycle path to mention both A and B, got: %s", errMsg)
	}
}

// TestValidateReservedKeyword verifies that reserved keywords are caught
func TestValidateReservedKeyword(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "func", // Reserved in Go
				Fields: []parser.Field{
					{Name: "value", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
				},
			},
		},
	}

	err := Validate(schema)
	if err == nil {
		t.Fatal("expected error for reserved keyword, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "func") {
		t.Errorf("expected error to mention 'func', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "RESERVED_KEYWORD") {
		t.Errorf("expected RESERVED_KEYWORD error code, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "reserved") {
		t.Errorf("expected 'reserved' in error, got: %s", errMsg)
	}
}

// TestValidateDuplicateStructs verifies that duplicate struct names are caught
func TestValidateDuplicateStructs(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Point",
				Fields: []parser.Field{
					{Name: "x", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
				},
			},
			{
				Name: "Point", // Duplicate
				Fields: []parser.Field{
					{Name: "y", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
				},
			},
		},
	}

	err := Validate(schema)
	if err == nil {
		t.Fatal("expected error for duplicate struct names, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "Point") {
		t.Errorf("expected error to mention 'Point', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "DUPLICATE_STRUCT") {
		t.Errorf("expected DUPLICATE_STRUCT error code, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "duplicate") {
		t.Errorf("expected 'duplicate' in error, got: %s", errMsg)
	}
}

// TestValidateDuplicateFields verifies that duplicate field names are caught
func TestValidateDuplicateFields(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Point",
				Fields: []parser.Field{
					{Name: "x", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
					{Name: "x", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}}, // Duplicate
				},
			},
		},
	}

	err := Validate(schema)
	if err == nil {
		t.Fatal("expected error for duplicate field names, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "x") {
		t.Errorf("expected error to mention field 'x', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "DUPLICATE_FIELD") {
		t.Errorf("expected DUPLICATE_FIELD error code, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "duplicate") {
		t.Errorf("expected 'duplicate' in error, got: %s", errMsg)
	}
}

// TestIntegrationMultipleErrors verifies that all errors are reported together
func TestIntegrationMultipleErrors(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name:   "Empty", // Empty struct error
				Fields: []parser.Field{},
			},
			{
				Name: "func", // Reserved keyword error
				Fields: []parser.Field{
					{Name: "field", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Unknown"}}, // Unknown type error
				},
			},
		},
	}

	err := Validate(schema)
	if err == nil {
		t.Fatal("expected errors, got nil")
	}

	errMsg := err.Error()

	// Should report all errors with error codes
	expectedErrors := []string{
		"EMPTY_STRUCT",    // Empty struct error code
		"RESERVED_KEYWORD", // Reserved keyword error code
		"UNKNOWN_TYPE",    // Unknown type error code
		"Empty",           // Empty struct name
		"func",            // Reserved keyword
		"Unknown",         // Unknown type
	}

	for _, expected := range expectedErrors {
		if !strings.Contains(errMsg, expected) {
			t.Errorf("expected error message to contain %q, got: %s", expected, errMsg)
		}
	}

	// Verify multiple errors are reported
	lines := strings.Split(errMsg, "\n")
	if len(lines) < 3 { // At least header + 2 errors
		t.Errorf("expected multiple error lines, got: %s", errMsg)
	}
}

// TestValidateComplexValidSchema verifies a complex but valid schema passes
func TestValidateComplexValidSchema(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Metadata",
				Fields: []parser.Field{
					{Name: "version", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "timestamp", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u64"}},
				},
			},
			{
				Name: "Record",
				Fields: []parser.Field{
					{Name: "meta", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Metadata"}},
					{
						Name: "data",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"},
						},
					},
				},
			},
			{
				Name: "Database",
				Fields: []parser.Field{
					{
						Name: "records",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Record"},
						},
					},
				},
			},
		},
	}

	err := Validate(schema)
	if err != nil {
		t.Errorf("expected complex valid schema to pass, got error: %v", err)
	}
}

// TestValidateInvalidIdentifier verifies that invalid identifier formats are caught
func TestValidateInvalidIdentifier(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Valid",
				Fields: []parser.Field{
					{Name: "123invalid", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}}, // Starts with digit
				},
			},
		},
	}

	err := Validate(schema)
	if err == nil {
		t.Fatal("expected error for invalid identifier, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "123invalid") {
		t.Errorf("expected error to mention '123invalid', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "INVALID_IDENTIFIER") {
		t.Errorf("expected INVALID_IDENTIFIER error code, got: %s", errMsg)
	}
}

// TestValidateArrayOfUnknownType verifies arrays with unknown element types are caught
func TestValidateArrayOfUnknownType(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Container",
				Fields: []parser.Field{
					{
						Name: "items",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "UnknownType"},
						},
					},
				},
			},
		},
	}

	err := Validate(schema)
	if err == nil {
		t.Fatal("expected error for array of unknown type, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "UnknownType") {
		t.Errorf("expected error to mention 'UnknownType', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "UNKNOWN_TYPE") {
		t.Errorf("expected UNKNOWN_TYPE error code, got: %s", errMsg)
	}
}
