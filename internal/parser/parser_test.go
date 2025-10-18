package parser

import (
	"testing"
)

func TestParseSimpleStruct(t *testing.T) {
	input := `struct Device {
		id: u32,
		name: str,
	}`
	
	schema, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}
	
	if len(schema.Structs) != 1 {
		t.Fatalf("Expected 1 struct, got %d", len(schema.Structs))
	}
	
	s := schema.Structs[0]
	if s.Name != "Device" {
		t.Errorf("Expected struct name 'Device', got %q", s.Name)
	}
	
	if len(s.Fields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(s.Fields))
	}
	
	// Check first field
	if s.Fields[0].Name != "id" {
		t.Errorf("Expected field name 'id', got %q", s.Fields[0].Name)
	}
	if s.Fields[0].Type.String() != "u32" {
		t.Errorf("Expected field type 'u32', got %q", s.Fields[0].Type.String())
	}
	
	// Check second field
	if s.Fields[1].Name != "name" {
		t.Errorf("Expected field name 'name', got %q", s.Fields[1].Name)
	}
	if s.Fields[1].Type.String() != "str" {
		t.Errorf("Expected field type 'str', got %q", s.Fields[1].Type.String())
	}
}

func TestParseNestedTypes(t *testing.T) {
	input := `struct Plugin {
		id: u32,
		parameters: []Parameter,
	}
	
	struct Parameter {
		name: str,
		value: f64,
	}`
	
	schema, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}
	
	if len(schema.Structs) != 2 {
		t.Fatalf("Expected 2 structs, got %d", len(schema.Structs))
	}
	
	// Check Plugin struct
	plugin := schema.Structs[0]
	if plugin.Name != "Plugin" {
		t.Errorf("Expected struct name 'Plugin', got %q", plugin.Name)
	}
	
	if len(plugin.Fields) != 2 {
		t.Fatalf("Expected 2 fields in Plugin, got %d", len(plugin.Fields))
	}
	
	// Check parameters field is array of Parameter
	paramField := plugin.Fields[1]
	if paramField.Name != "parameters" {
		t.Errorf("Expected field name 'parameters', got %q", paramField.Name)
	}
	if paramField.Type.Kind != TypeKindArray {
		t.Errorf("Expected array type, got %v", paramField.Type.Kind)
	}
	if paramField.Type.Elem.Name != "Parameter" {
		t.Errorf("Expected element type 'Parameter', got %q", paramField.Type.Elem.Name)
	}
}

func TestParseArrayField(t *testing.T) {
	input := `struct List {
		items: []u32,
	}`
	
	schema, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}
	
	if len(schema.Structs) != 1 {
		t.Fatalf("Expected 1 struct, got %d", len(schema.Structs))
	}
	
	s := schema.Structs[0]
	if len(s.Fields) != 1 {
		t.Fatalf("Expected 1 field, got %d", len(s.Fields))
	}
	
	field := s.Fields[0]
	if field.Type.Kind != TypeKindArray {
		t.Errorf("Expected array type, got %v", field.Type.Kind)
	}
	if field.Type.Elem.Name != "u32" {
		t.Errorf("Expected element type 'u32', got %q", field.Type.Elem.Name)
	}
	if field.Type.String() != "[]u32" {
		t.Errorf("Expected type string '[]u32', got %q", field.Type.String())
	}
}

func TestParseDocComments(t *testing.T) {
	input := `/// A device structure.
	/// Contains device information.
	struct Device {
		/// Device identifier.
		id: u32,
		/// Device name.
		name: str,
	}`
	
	schema, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}
	
	if len(schema.Structs) != 1 {
		t.Fatalf("Expected 1 struct, got %d", len(schema.Structs))
	}
	
	s := schema.Structs[0]
	expectedStructComment := "A device structure.\nContains device information."
	if s.Comment != expectedStructComment {
		t.Errorf("Expected struct comment %q, got %q", expectedStructComment, s.Comment)
	}
	
	if len(s.Fields) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(s.Fields))
	}
	
	if s.Fields[0].Comment != "Device identifier." {
		t.Errorf("Expected field comment 'Device identifier.', got %q", s.Fields[0].Comment)
	}
	
	if s.Fields[1].Comment != "Device name." {
		t.Errorf("Expected field comment 'Device name.', got %q", s.Fields[1].Comment)
	}
}

func TestParseSyntaxError(t *testing.T) {
	testCases := []struct {
		input       string
		shouldError bool
		description string
	}{
		{
			input:       `struct Device`,
			shouldError: true,
			description: "missing brace",
		},
		{
			input:       `struct Device { id u32 }`,
			shouldError: true,
			description: "missing colon",
		},
		{
			input:       `struct Device { id: }`,
			shouldError: true,
			description: "missing type",
		},
		{
			input:       `struct { id: u32 }`,
			shouldError: true,
			description: "missing struct name",
		},
		{
			input:       `struct Device { id: u32 name: str }`,
			shouldError: true,
			description: "missing comma",
		},
		{
			input:       `Device { id: u32 }`,
			shouldError: true,
			description: "missing struct keyword",
		},
	}
	
	for _, tc := range testCases {
		_, err := ParseSchema(tc.input)
		if tc.shouldError && err == nil {
			t.Errorf("Test %q: expected error, got nil", tc.description)
		}
		if !tc.shouldError && err != nil {
			t.Errorf("Test %q: expected no error, got %v", tc.description, err)
		}
	}
}

func TestParseEmptyStruct(t *testing.T) {
	input := `struct Empty {}`
	
	schema, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}
	
	if len(schema.Structs) != 1 {
		t.Fatalf("Expected 1 struct, got %d", len(schema.Structs))
	}
	
	s := schema.Structs[0]
	if s.Name != "Empty" {
		t.Errorf("Expected struct name 'Empty', got %q", s.Name)
	}
	
	if len(s.Fields) != 0 {
		t.Errorf("Expected 0 fields, got %d", len(s.Fields))
	}
}

func TestParseMultipleStructs(t *testing.T) {
	input := `struct First {
		a: u32,
	}
	
	struct Second {
		b: str,
	}
	
	struct Third {
		c: bool,
	}`
	
	schema, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}
	
	if len(schema.Structs) != 3 {
		t.Fatalf("Expected 3 structs, got %d", len(schema.Structs))
	}
	
	names := []string{"First", "Second", "Third"}
	for i, s := range schema.Structs {
		if s.Name != names[i] {
			t.Errorf("Struct %d: expected name %q, got %q", i, names[i], s.Name)
		}
	}
}

func TestParseTrailingComma(t *testing.T) {
	testCases := []struct {
		input       string
		description string
	}{
		{
			input:       `struct Device { id: u32, }`,
			description: "trailing comma after single field",
		},
		{
			input:       `struct Device { id: u32, name: str, }`,
			description: "trailing comma after multiple fields",
		},
		{
			input:       `struct Device { id: u32 }`,
			description: "no trailing comma",
		},
	}
	
	for _, tc := range testCases {
		_, err := ParseSchema(tc.input)
		if err != nil {
			t.Errorf("Test %q: unexpected error: %v", tc.description, err)
		}
	}
}

func TestParseAllPrimitives(t *testing.T) {
	input := `struct AllTypes {
		u8_field: u8,
		u16_field: u16,
		u32_field: u32,
		u64_field: u64,
		i8_field: i8,
		i16_field: i16,
		i32_field: i32,
		i64_field: i64,
		f32_field: f32,
		f64_field: f64,
		bool_field: bool,
		str_field: str,
	}`
	
	schema, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}
	
	if len(schema.Structs) != 1 {
		t.Fatalf("Expected 1 struct, got %d", len(schema.Structs))
	}
	
	s := schema.Structs[0]
	if len(s.Fields) != 12 {
		t.Fatalf("Expected 12 fields, got %d", len(s.Fields))
	}
	
	expectedTypes := []string{
		"u8", "u16", "u32", "u64",
		"i8", "i16", "i32", "i64",
		"f32", "f64", "bool", "str",
	}
	
	for i, field := range s.Fields {
		if field.Type.String() != expectedTypes[i] {
			t.Errorf("Field %d: expected type %q, got %q", i, expectedTypes[i], field.Type.String())
		}
		if !field.Type.IsPrimitive() {
			t.Errorf("Field %d: type %q should be primitive", i, field.Type.String())
		}
	}
}

func TestParseWithComments(t *testing.T) {
	input := `// This is a regular comment
	/// Doc comment for struct
	struct Device {
		// Regular comment before field
		/// Doc comment for field
		id: u32,
	}`
	
	schema, err := ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}
	
	if len(schema.Structs) != 1 {
		t.Fatalf("Expected 1 struct, got %d", len(schema.Structs))
	}
	
	s := schema.Structs[0]
	if s.Comment != "Doc comment for struct" {
		t.Errorf("Expected struct doc comment, got %q", s.Comment)
	}
	
	if len(s.Fields) != 1 {
		t.Fatalf("Expected 1 field, got %d", len(s.Fields))
	}
	
	if s.Fields[0].Comment != "Doc comment for field" {
		t.Errorf("Expected field doc comment, got %q", s.Fields[0].Comment)
	}
}
