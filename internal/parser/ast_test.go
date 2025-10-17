package parser

import "testing"

func TestTypeExprIsPrimitive(t *testing.T) {
	primitives := []string{
		"u8", "u16", "u32", "u64",
		"i8", "i16", "i32", "i64",
		"f32", "f64",
		"bool", "string",
	}

	for _, name := range primitives {
		typ := &TypeExpr{Kind: TypeKindPrimitive, Name: name}
		if !typ.IsPrimitive() {
			t.Errorf("Type %s should be primitive", name)
		}
	}

	// Non-primitives
	notPrimitive := []TypeExpr{
		{Kind: TypeKindNamed, Name: "MyStruct"},
		{Kind: TypeKindArray, Elem: &TypeExpr{Kind: TypeKindPrimitive, Name: "u32"}},
	}

	for _, typ := range notPrimitive {
		if typ.IsPrimitive() {
			t.Errorf("Type %v should not be primitive", typ)
		}
	}
}

func TestTypeExprString(t *testing.T) {
	testCases := []struct {
		typ      TypeExpr
		expected string
	}{
		{
			typ:      TypeExpr{Kind: TypeKindPrimitive, Name: "u32"},
			expected: "u32",
		},
		{
			typ:      TypeExpr{Kind: TypeKindNamed, Name: "MyStruct"},
			expected: "MyStruct",
		},
		{
			typ: TypeExpr{
				Kind: TypeKindArray,
				Elem: &TypeExpr{Kind: TypeKindPrimitive, Name: "u32"},
			},
			expected: "[]u32",
		},
		{
			typ: TypeExpr{
				Kind: TypeKindArray,
				Elem: &TypeExpr{
					Kind: TypeKindArray,
					Elem: &TypeExpr{Kind: TypeKindPrimitive, Name: "string"},
				},
			},
			expected: "[][]string",
		},
	}

	for _, tc := range testCases {
		got := tc.typ.String()
		if got != tc.expected {
			t.Errorf("Type.String(): expected %q, got %q", tc.expected, got)
		}
	}
}

func TestSchemaStructure(t *testing.T) {
	// Test that we can build a schema programmatically
	schema := &Schema{
		Structs: []Struct{
			{
				Name:    "Device",
				Comment: "Represents a hardware device",
				Fields: []Field{
					{
						Name:    "id",
						Type:    TypeExpr{Kind: TypeKindPrimitive, Name: "u32"},
						Comment: "Unique device identifier",
					},
					{
						Name:    "name",
						Type:    TypeExpr{Kind: TypeKindPrimitive, Name: "string"},
						Comment: "Device name",
					},
				},
			},
		},
	}

	if len(schema.Structs) != 1 {
		t.Errorf("Expected 1 struct, got %d", len(schema.Structs))
	}

	s := schema.Structs[0]
	if s.Name != "Device" {
		t.Errorf("Expected struct name 'Device', got %q", s.Name)
	}

	if len(s.Fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(s.Fields))
	}

	if s.Fields[0].Type.String() != "u32" {
		t.Errorf("Expected field type 'u32', got %q", s.Fields[0].Type.String())
	}
}
