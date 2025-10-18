package validator

import (
	"strings"
	"testing"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

func TestValidateKnownTypes(t *testing.T) {
	input := `
	struct Device {
		id: u32,
		name: str,
		config: Config,
		tags: []str,
	}
	
	struct Config {
		enabled: bool,
		priority: i32,
	}
	`

	schema, err := parser.ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	errors := ValidateTypeReferences(schema)
	if len(errors) != 0 {
		t.Errorf("Expected no errors for valid types, got: %v", errors)
	}
}

func TestValidatePrimitiveTypes(t *testing.T) {
	input := `
	struct AllPrimitives {
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
	}
	`

	schema, err := parser.ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	errors := ValidateTypeReferences(schema)
	if len(errors) != 0 {
		t.Errorf("Expected no errors for primitive types, got: %v", errors)
	}
}

func TestValidateUnknownType(t *testing.T) {
	input := `
	struct Device {
		id: u32,
		config: UnknownConfig,
	}
	`

	schema, err := parser.ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	errors := ValidateTypeReferences(schema)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d: %v", len(errors), errors)
	}

	errMsg := errors[0].Error()
	if !strings.Contains(errMsg, "UnknownConfig") {
		t.Errorf("Error should mention 'UnknownConfig', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "Device") {
		t.Errorf("Error should mention struct name 'Device', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "config") {
		t.Errorf("Error should mention field name 'config', got: %s", errMsg)
	}
}

func TestValidateArrayType(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		shouldError bool
		errorText   string
	}{
		{
			name: "array of primitive",
			input: `
			struct List {
				items: []u32,
			}
			`,
			shouldError: false,
		},
		{
			name: "array of defined struct",
			input: `
			struct List {
				devices: []Device,
			}
			struct Device {
				id: u32,
			}
			`,
			shouldError: false,
		},
		{
			name: "array of unknown type",
			input: `
			struct List {
				items: []UnknownType,
			}
			`,
			shouldError: true,
			errorText:   "UnknownType",
		},
		{
			name: "nested array",
			input: `
			struct Matrix {
				rows: [][]u32,
			}
			`,
			shouldError: false,
		},
		{
			name: "nested array of struct",
			input: `
			struct Groups {
				data: [][]Device,
			}
			struct Device {
				id: u32,
			}
			`,
			shouldError: false,
		},
		{
			name: "nested array with unknown type",
			input: `
			struct Groups {
				data: [][]UnknownType,
			}
			`,
			shouldError: true,
			errorText:   "UnknownType",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := parser.ParseSchema(tc.input)
			if err != nil {
				t.Fatalf("ParseSchema failed: %v", err)
			}

			errors := ValidateTypeReferences(schema)

			if tc.shouldError {
				if len(errors) == 0 {
					t.Errorf("Expected error containing %q, got no errors", tc.errorText)
				} else if !strings.Contains(errors[0].Error(), tc.errorText) {
					t.Errorf("Expected error containing %q, got: %s", tc.errorText, errors[0].Error())
				}
			} else {
				if len(errors) != 0 {
					t.Errorf("Expected no errors, got: %v", errors)
				}
			}
		})
	}
}

func TestValidateMultipleErrors(t *testing.T) {
	input := `
	struct Device {
		id: u32,
		config: UnknownConfig,
		settings: UnknownSettings,
		items: []UnknownItem,
	}
	
	struct Plugin {
		data: AnotherUnknown,
	}
	`

	schema, err := parser.ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	errors := ValidateTypeReferences(schema)
	
	// Should have 4 errors (one for each unknown type)
	if len(errors) != 4 {
		t.Errorf("Expected 4 errors, got %d: %v", len(errors), errors)
	}

	// Check that all unknown types are mentioned
	allErrors := ""
	for _, e := range errors {
		allErrors += e.Error() + " "
	}

	expectedTypes := []string{"UnknownConfig", "UnknownSettings", "UnknownItem", "AnotherUnknown"}
	for _, typeName := range expectedTypes {
		if !strings.Contains(allErrors, typeName) {
			t.Errorf("Expected errors to mention %q, got: %s", typeName, allErrors)
		}
	}
}

func TestValidateForwardReference(t *testing.T) {
	// Type can reference a struct defined later in the file
	input := `
	struct Device {
		config: Config,
	}
	
	struct Config {
		enabled: bool,
	}
	`

	schema, err := parser.ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	errors := ValidateTypeReferences(schema)
	if len(errors) != 0 {
		t.Errorf("Expected no errors for forward reference, got: %v", errors)
	}
}

func TestValidateEmptySchema(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{},
	}

	errors := ValidateTypeReferences(schema)
	if len(errors) != 0 {
		t.Errorf("Expected no errors for empty schema, got: %v", errors)
	}
}

func TestValidateComplexNesting(t *testing.T) {
	input := `
	struct Root {
		items: []Item,
	}
	
	struct Item {
		children: []Child,
	}
	
	struct Child {
		metadata: Metadata,
	}
	
	struct Metadata {
		tags: []str,
	}
	`

	schema, err := parser.ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	errors := ValidateTypeReferences(schema)
	if len(errors) != 0 {
		t.Errorf("Expected no errors for complex nesting, got: %v", errors)
	}
}
