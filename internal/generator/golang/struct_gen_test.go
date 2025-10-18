package golang

import (
	"strings"
	"testing"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// TestGenerateSimpleStruct verifies basic struct generation
func TestGenerateSimpleStruct(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name:    "Device",
				Comment: "represents an audio device.",
				Fields: []parser.Field{
					{
						Name:    "id",
						Type:    parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"},
						Comment: "is the unique identifier.",
					},
					{
						Name:    "name",
						Type:    parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"},
						Comment: "is the device name.",
					},
				},
			},
		},
	}

	result, err := GenerateStructs(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check struct declaration
	if !strings.Contains(result, "type Device struct {") {
		t.Errorf("missing struct declaration, got:\n%s", result)
	}

	// Check struct doc comment
	if !strings.Contains(result, "// Device represents an audio device.") {
		t.Errorf("missing or incorrect struct doc comment, got:\n%s", result)
	}

	// Check field declarations
	if !strings.Contains(result, "Id uint32") {
		t.Errorf("missing Id field, got:\n%s", result)
	}
	if !strings.Contains(result, "Name string") {
		t.Errorf("missing Name field, got:\n%s", result)
	}

	// Check field doc comments
	if !strings.Contains(result, "// Id is the unique identifier.") {
		t.Errorf("missing Id doc comment, got:\n%s", result)
	}
	if !strings.Contains(result, "// Name is the device name.") {
		t.Errorf("missing Name doc comment, got:\n%s", result)
	}
}

// TestGenerateMultipleStructs verifies generation of multiple structs with spacing
func TestGenerateMultipleStructs(t *testing.T) {
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

	result, err := GenerateStructs(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check both structs present
	if !strings.Contains(result, "type Point struct {") {
		t.Errorf("missing Point struct, got:\n%s", result)
	}
	if !strings.Contains(result, "type Line struct {") {
		t.Errorf("missing Line struct, got:\n%s", result)
	}

	// Check fields with name conversion
	if !strings.Contains(result, "X float32") {
		t.Errorf("missing or incorrect Point.X field, got:\n%s", result)
	}
	if !strings.Contains(result, "Start Point") {
		t.Errorf("missing or incorrect Line.Start field, got:\n%s", result)
	}
	if !strings.Contains(result, "End Point") {
		t.Errorf("missing or incorrect Line.End field, got:\n%s", result)
	}

	// Verify blank line between structs
	lines := strings.Split(result, "\n")
	foundBlankLine := false
	inSecondStruct := false
	for _, line := range lines {
		if strings.Contains(line, "type Line struct") {
			inSecondStruct = true
		}
		if !inSecondStruct && strings.Contains(line, "}") {
			// After first struct closes, next should be blank before second struct
			foundBlankLine = true
		}
	}
	if !foundBlankLine {
		t.Errorf("expected blank line between structs")
	}
}

// TestGenerateWithArrays verifies array type generation
func TestGenerateWithArrays(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Container",
				Fields: []parser.Field{
					{
						Name: "data",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"},
						},
					},
					{
						Name: "items",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Item"},
						},
					},
				},
			},
		},
	}

	result, err := GenerateStructs(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check array of primitive
	if !strings.Contains(result, "Data []uint8") {
		t.Errorf("missing or incorrect Data field, got:\n%s", result)
	}

	// Check array of named type (with name conversion)
	if !strings.Contains(result, "Items []Item") {
		t.Errorf("missing or incorrect Items field, got:\n%s", result)
	}
}

// TestGenerateWithNestedArrays verifies nested array generation
func TestGenerateWithNestedArrays(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Matrix",
				Fields: []parser.Field{
					{
						Name: "values",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{
								Kind: parser.TypeKindArray,
								Elem: &parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f64"},
							},
						},
					},
				},
			},
		},
	}

	result, err := GenerateStructs(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, "Values [][]float64") {
		t.Errorf("missing or incorrect nested array field, got:\n%s", result)
	}
}

// TestGenerateSnakeCaseConversion verifies snake_case to PascalCase conversion
func TestGenerateSnakeCaseConversion(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "audio_device",
				Fields: []parser.Field{
					{Name: "device_id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "sample_rate", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{
						Name: "plugin_list",
						Type: parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "plugin_info"},
					},
				},
			},
		},
	}

	result, err := GenerateStructs(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check struct name conversion
	if !strings.Contains(result, "type AudioDevice struct {") {
		t.Errorf("missing or incorrect struct name conversion, got:\n%s", result)
	}

	// Check field name conversions
	if !strings.Contains(result, "DeviceId uint32") {
		t.Errorf("missing or incorrect DeviceId field, got:\n%s", result)
	}
	if !strings.Contains(result, "SampleRate uint32") {
		t.Errorf("missing or incorrect SampleRate field, got:\n%s", result)
	}

	// Check named type conversion
	if !strings.Contains(result, "PluginList PluginInfo") {
		t.Errorf("missing or incorrect named type conversion, got:\n%s", result)
	}
}

// TestGenerateWithoutComments verifies structs without doc comments
func TestGenerateWithoutComments(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name:    "Simple",
				Comment: "", // No comment
				Fields: []parser.Field{
					{
						Name:    "value",
						Type:    parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"},
						Comment: "", // No comment
					},
				},
			},
		},
	}

	result, err := GenerateStructs(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have struct and field, but no doc comments
	if !strings.Contains(result, "type Simple struct {") {
		t.Errorf("missing struct declaration, got:\n%s", result)
	}
	if !strings.Contains(result, "Value uint32") {
		t.Errorf("missing field declaration, got:\n%s", result)
	}

	// Should not have comment lines (except for the field itself)
	lines := strings.Split(result, "\n")
	commentCount := 0
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "//") {
			commentCount++
		}
	}
	if commentCount > 0 {
		t.Errorf("expected no doc comments, found %d comment lines", commentCount)
	}
}

// TestGenerateAllPrimitiveTypes verifies all primitive type mappings in struct context
func TestGenerateAllPrimitiveTypes(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "AllTypes",
				Fields: []parser.Field{
					{Name: "u8_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"}},
					{Name: "u16_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u16"}},
					{Name: "u32_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
					{Name: "u64_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u64"}},
					{Name: "i8_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i8"}},
					{Name: "i16_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i16"}},
					{Name: "i32_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i32"}},
					{Name: "i64_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i64"}},
					{Name: "f32_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f32"}},
					{Name: "f64_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f64"}},
					{Name: "bool_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "bool"}},
					{Name: "str_field", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}},
				},
			},
		},
	}

	result, err := GenerateStructs(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedFields := []string{
		"U8Field uint8",
		"U16Field uint16",
		"U32Field uint32",
		"U64Field uint64",
		"I8Field int8",
		"I16Field int16",
		"I32Field int32",
		"I64Field int64",
		"F32Field float32",
		"F64Field float64",
		"BoolField bool",
		"StrField string",
	}

	for _, expected := range expectedFields {
		if !strings.Contains(result, expected) {
			t.Errorf("missing field %q, got:\n%s", expected, result)
		}
	}
}

// TestGenerateNilSchema verifies error handling for nil schema
func TestGenerateNilSchema(t *testing.T) {
	result, err := GenerateStructs(nil)

	if err == nil {
		t.Errorf("expected error for nil schema, got result: %s", result)
	}
	if !strings.Contains(err.Error(), "schema is nil") {
		t.Errorf("expected 'schema is nil' in error, got: %v", err)
	}
}

// TestGenerateEmptySchema verifies error handling for empty schema
func TestGenerateEmptySchema(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{},
	}

	result, err := GenerateStructs(schema)

	if err == nil {
		t.Errorf("expected error for empty schema, got result: %s", result)
	}
	if !strings.Contains(err.Error(), "no structs") {
		t.Errorf("expected 'no structs' in error, got: %v", err)
	}
}

// TestGenerateInvalidFieldType verifies error handling for invalid field types
func TestGenerateInvalidFieldType(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Bad",
				Fields: []parser.Field{
					{
						Name: "bad_field",
						Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "invalid"},
					},
				},
			},
		},
	}

	result, err := GenerateStructs(schema)

	if err == nil {
		t.Errorf("expected error for invalid field type, got result: %s", result)
	}
	if !strings.Contains(err.Error(), "Bad") {
		t.Errorf("expected struct name in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "bad_field") {
		t.Errorf("expected field name in error, got: %v", err)
	}
}

// TestGenerateComplexSchema verifies realistic complex schema
func TestGenerateComplexSchema(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name:    "Parameter",
				Comment: "represents a plugin parameter.",
				Fields: []parser.Field{
					{Name: "name", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}, Comment: "is the parameter name."},
					{Name: "value", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "f64"}, Comment: "is the parameter value."},
				},
			},
			{
				Name:    "Plugin",
				Comment: "represents an audio plugin.",
				Fields: []parser.Field{
					{Name: "id", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}, Comment: "is the plugin ID."},
					{Name: "name", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "str"}, Comment: "is the plugin name."},
					{
						Name: "parameters",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Parameter"},
						},
						Comment: "are the plugin parameters.",
					},
				},
			},
			{
				Name:    "PluginList",
				Comment: "contains all enumerated plugins.",
				Fields: []parser.Field{
					{
						Name: "plugins",
						Type: parser.TypeExpr{
							Kind: parser.TypeKindArray,
							Elem: &parser.TypeExpr{Kind: parser.TypeKindNamed, Name: "Plugin"},
						},
						Comment: "are the discovered plugins.",
					},
				},
			},
		},
	}

	result, err := GenerateStructs(schema)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify all three structs
	if !strings.Contains(result, "type Parameter struct {") {
		t.Errorf("missing Parameter struct")
	}
	if !strings.Contains(result, "type Plugin struct {") {
		t.Errorf("missing Plugin struct")
	}
	if !strings.Contains(result, "type PluginList struct {") {
		t.Errorf("missing PluginList struct")
	}

	// Verify nested types
	if !strings.Contains(result, "Parameters []Parameter") {
		t.Errorf("missing or incorrect Plugin.Parameters field")
	}
	if !strings.Contains(result, "Plugins []Plugin") {
		t.Errorf("missing or incorrect PluginList.Plugins field")
	}

	// Verify doc comments preserved
	if !strings.Contains(result, "// Parameter represents a plugin parameter.") {
		t.Errorf("missing Parameter doc comment")
	}
	if !strings.Contains(result, "// Parameters are the plugin parameters.") {
		t.Errorf("missing Parameters field doc comment")
	}
}
