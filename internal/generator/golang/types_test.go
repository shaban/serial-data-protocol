package golang

import (
	"strings"
	"testing"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// TestMapPrimitiveTypes verifies all primitive type mappings
func TestMapPrimitiveTypes(t *testing.T) {
	tests := []struct {
		name       string
		schemaType string
		goType     string
	}{
		// Unsigned integers
		{"u8", "u8", "uint8"},
		{"u16", "u16", "uint16"},
		{"u32", "u32", "uint32"},
		{"u64", "u64", "uint64"},

		// Signed integers
		{"i8", "i8", "int8"},
		{"i16", "i16", "int16"},
		{"i32", "i32", "int32"},
		{"i64", "i64", "int64"},

		// Floating point
		{"f32", "f32", "float32"},
		{"f64", "f64", "float64"},

		// Other primitives
		{"bool", "bool", "bool"},
		{"str", "str", "string"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeExpr := &parser.TypeExpr{
				Kind: parser.TypeKindPrimitive,
				Name: tt.schemaType,
			}

			result, err := MapTypeToGo(typeExpr)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.goType {
				t.Errorf("MapTypeToGo(%q) = %q, want %q", tt.schemaType, result, tt.goType)
			}
		})
	}
}

// TestMapUnknownPrimitive verifies error handling for unknown primitive types
func TestMapUnknownPrimitive(t *testing.T) {
	typeExpr := &parser.TypeExpr{
		Kind: parser.TypeKindPrimitive,
		Name: "unknown",
	}

	result, err := MapTypeToGo(typeExpr)

	if err == nil {
		t.Errorf("expected error for unknown primitive, got result: %q", result)
	}
	if !strings.Contains(err.Error(), "unknown primitive type") {
		t.Errorf("expected 'unknown primitive type' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "unknown") {
		t.Errorf("expected type name in error, got: %v", err)
	}
}

// TestMapNamedTypes verifies that user-defined struct names are preserved
func TestMapNamedTypes(t *testing.T) {
	tests := []struct {
		name       string
		structName string
	}{
		{"simple", "Device"},
		{"snake_case", "audio_device"},
		{"camelCase", "audioDevice"},
		{"PascalCase", "AudioDevice"},
		{"with_underscores", "my_struct_name"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeExpr := &parser.TypeExpr{
				Kind: parser.TypeKindNamed,
				Name: tt.structName,
			}

			result, err := MapTypeToGo(typeExpr)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			// Named types should be preserved as-is (no conversion yet)
			if result != tt.structName {
				t.Errorf("MapTypeToGo(%q) = %q, want %q", tt.structName, result, tt.structName)
			}
		})
	}
}

// TestMapArrayOfPrimitives verifies array type mappings with primitive elements
func TestMapArrayOfPrimitives(t *testing.T) {
	tests := []struct {
		name       string
		schemaType string
		goType     string
	}{
		{"array of u8", "u8", "[]uint8"},
		{"array of u32", "u32", "[]uint32"},
		{"array of f64", "f64", "[]float64"},
		{"array of str", "str", "[]string"},
		{"array of bool", "bool", "[]bool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeExpr := &parser.TypeExpr{
				Kind: parser.TypeKindArray,
				Elem: &parser.TypeExpr{
					Kind: parser.TypeKindPrimitive,
					Name: tt.schemaType,
				},
			}

			result, err := MapTypeToGo(typeExpr)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.goType {
				t.Errorf("MapTypeToGo([]%s) = %q, want %q", tt.schemaType, result, tt.goType)
			}
		})
	}
}

// TestMapArrayOfStructs verifies array type mappings with named struct elements
func TestMapArrayOfStructs(t *testing.T) {
	tests := []struct {
		name       string
		structName string
		expected   string
	}{
		{"array of Device", "Device", "[]Device"},
		{"array of snake_case", "audio_device", "[]audio_device"},
		{"array of Parameter", "Parameter", "[]Parameter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			typeExpr := &parser.TypeExpr{
				Kind: parser.TypeKindArray,
				Elem: &parser.TypeExpr{
					Kind: parser.TypeKindNamed,
					Name: tt.structName,
				},
			}

			result, err := MapTypeToGo(typeExpr)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("MapTypeToGo([]%s) = %q, want %q", tt.structName, result, tt.expected)
			}
		})
	}
}

// TestMapNestedArrays verifies nested array type mappings
func TestMapNestedArrays(t *testing.T) {
	tests := []struct {
		name     string
		typeExpr *parser.TypeExpr
		expected string
	}{
		{
			name: "array of array of u32",
			typeExpr: &parser.TypeExpr{
				Kind: parser.TypeKindArray,
				Elem: &parser.TypeExpr{
					Kind: parser.TypeKindArray,
					Elem: &parser.TypeExpr{
						Kind: parser.TypeKindPrimitive,
						Name: "u32",
					},
				},
			},
			expected: "[][]uint32",
		},
		{
			name: "array of array of string",
			typeExpr: &parser.TypeExpr{
				Kind: parser.TypeKindArray,
				Elem: &parser.TypeExpr{
					Kind: parser.TypeKindArray,
					Elem: &parser.TypeExpr{
						Kind: parser.TypeKindPrimitive,
						Name: "str",
					},
				},
			},
			expected: "[][]string",
		},
		{
			name: "array of array of Device",
			typeExpr: &parser.TypeExpr{
				Kind: parser.TypeKindArray,
				Elem: &parser.TypeExpr{
					Kind: parser.TypeKindArray,
					Elem: &parser.TypeExpr{
						Kind: parser.TypeKindNamed,
						Name: "Device",
					},
				},
			},
			expected: "[][]Device",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := MapTypeToGo(tt.typeExpr)

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if result != tt.expected {
				t.Errorf("MapTypeToGo() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// TestMapArrayWithoutElement verifies error handling for malformed array types
func TestMapArrayWithoutElement(t *testing.T) {
	typeExpr := &parser.TypeExpr{
		Kind: parser.TypeKindArray,
		Elem: nil, // Missing element type
	}

	result, err := MapTypeToGo(typeExpr)

	if err == nil {
		t.Errorf("expected error for array without element type, got result: %q", result)
	}
	if !strings.Contains(err.Error(), "no element type") {
		t.Errorf("expected 'no element type' in error, got: %v", err)
	}
}

// TestMapArrayWithInvalidElement verifies error propagation from nested types
func TestMapArrayWithInvalidElement(t *testing.T) {
	typeExpr := &parser.TypeExpr{
		Kind: parser.TypeKindArray,
		Elem: &parser.TypeExpr{
			Kind: parser.TypeKindPrimitive,
			Name: "invalid",
		},
	}

	result, err := MapTypeToGo(typeExpr)

	if err == nil {
		t.Errorf("expected error for array with invalid element type, got result: %q", result)
	}
	if !strings.Contains(err.Error(), "array element type error") {
		t.Errorf("expected 'array element type error' in error, got: %v", err)
	}
}

// TestMapNilTypeExpr verifies error handling for nil input
func TestMapNilTypeExpr(t *testing.T) {
	result, err := MapTypeToGo(nil)

	if err == nil {
		t.Errorf("expected error for nil type expression, got result: %q", result)
	}
	if !strings.Contains(err.Error(), "nil") {
		t.Errorf("expected 'nil' in error message, got: %v", err)
	}
}

// TestMapUnknownTypeKind verifies error handling for invalid type kinds
func TestMapUnknownTypeKind(t *testing.T) {
	typeExpr := &parser.TypeExpr{
		Kind: parser.TypeKind(999), // Invalid type kind
		Name: "something",
	}

	result, err := MapTypeToGo(typeExpr)

	if err == nil {
		t.Errorf("expected error for unknown type kind, got result: %q", result)
	}
	if !strings.Contains(err.Error(), "unknown type kind") {
		t.Errorf("expected 'unknown type kind' in error, got: %v", err)
	}
}
