package validator

import (
	"strings"
	"testing"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// TestStructureEmptySchema verifies that schemas with no structs are rejected
func TestStructureEmptySchema(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{},
	}

	errors := ValidateStructure(schema)

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	errMsg := errors[0].Error()
	if !strings.Contains(errMsg, "schema must define at least one struct") {
		t.Errorf("expected error about empty schema, got: %s", errMsg)
	}
}

// TestEmptyStruct verifies that structs with no fields are rejected
func TestEmptyStruct(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name:   "Empty",
				Fields: []parser.Field{},
			},
		},
	}

	errors := ValidateStructure(schema)

	if len(errors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(errors))
	}

	errMsg := errors[0].Error()
	if !strings.Contains(errMsg, "Empty") {
		t.Errorf("expected error to mention struct name 'Empty', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "cannot be empty") {
		t.Errorf("expected error about empty struct, got: %s", errMsg)
	}
}

// TestMultipleEmptyStructs verifies all empty structs are reported
func TestMultipleEmptyStructs(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name:   "First",
				Fields: []parser.Field{},
			},
			{
				Name:   "Second",
				Fields: []parser.Field{},
			},
			{
				Name: "Third",
				Fields: []parser.Field{
					{Name: "x", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
				},
			},
			{
				Name:   "Fourth",
				Fields: []parser.Field{},
			},
		},
	}

	errors := ValidateStructure(schema)

	// Should report First, Second, and Fourth (Third is valid)
	if len(errors) != 3 {
		t.Fatalf("expected 3 errors, got %d", len(errors))
	}

	// Check all three empty structs are mentioned
	errMsgs := make([]string, len(errors))
	for i, err := range errors {
		errMsgs[i] = err.Error()
	}

	for _, structName := range []string{"First", "Second", "Fourth"} {
		found := false
		for _, msg := range errMsgs {
			if strings.Contains(msg, structName) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected error about empty struct %q, errors: %v", structName, errMsgs)
		}
	}

	// Third should NOT be mentioned (it's valid)
	for _, msg := range errMsgs {
		if strings.Contains(msg, "Third") {
			t.Errorf("valid struct 'Third' should not be reported as error: %s", msg)
		}
	}
}

// TestValidSingleStruct verifies a schema with one valid struct passes
func TestValidSingleStruct(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Point",
				Fields: []parser.Field{
					{Name: "x", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
					{Name: "y", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
				},
			},
		},
	}

	errors := ValidateStructure(schema)

	if len(errors) != 0 {
		t.Errorf("expected no errors for valid schema, got: %v", errors)
	}
}

// TestValidMultipleStructs verifies a schema with multiple valid structs passes
func TestValidMultipleStructs(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Point",
				Fields: []parser.Field{
					{Name: "x", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
				},
			},
			{
				Name: "Line",
				Fields: []parser.Field{
					{Name: "start", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Point"}},
					{Name: "end", Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Point"}},
				},
			},
			{
				Name: "Shape",
				Fields: []parser.Field{
					{
						Name: "edges",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Line"},
						},
					},
				},
			},
		},
	}

	errors := ValidateStructure(schema)

	if len(errors) != 0 {
		t.Errorf("expected no errors for valid schema, got: %v", errors)
	}
}

// TestSingleFieldStruct verifies that a struct with exactly one field is valid
func TestSingleFieldStruct(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Counter",
				Fields: []parser.Field{
					{Name: "count", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u64"}},
				},
			},
		},
	}

	errors := ValidateStructure(schema)

	if len(errors) != 0 {
		t.Errorf("expected no errors for single-field struct, got: %v", errors)
	}
}

// TestComplexValidSchema verifies complex schemas with arrays and nested types
func TestComplexValidSchema(t *testing.T) {
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

	errors := ValidateStructure(schema)

	if len(errors) != 0 {
		t.Errorf("expected no errors for complex valid schema, got: %v", errors)
	}
}
