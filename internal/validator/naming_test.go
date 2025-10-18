package validator

import (
	"strings"
	"testing"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

func TestValidNames(t *testing.T) {
	input := `
	struct Device {
		id: u32,
		device_name: str,
		isEnabled: bool,
		_internal: u32,
	}
	
	struct Config_v2 {
		value: f64,
	}
	`

	schema, err := parser.ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	errors := ValidateNaming(schema)
	if len(errors) != 0 {
		t.Errorf("Expected no errors for valid names, got: %v", errors)
	}
}

func TestReservedKeyword(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "struct name is Go keyword",
			input: `
			struct type {
				id: u32,
			}
			`,
			expectedError: "type",
		},
		{
			name: "struct name is Rust keyword",
			input: `
			struct impl {
				value: u32,
			}
			`,
			expectedError: "impl",
		},
		{
			name: "field name is reserved",
			input: `
			struct Config {
				async: bool,
			}
			`,
			expectedError: "async",
		},
		{
			name: "field name is Go built-in",
			input: `
			struct Data {
				len: u32,
			}
			`,
			expectedError: "len",
		},
		{
			name: "field name is Swift keyword",
			input: `
			struct Settings {
				guard: bool,
			}
			`,
			expectedError: "guard",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := parser.ParseSchema(tc.input)
			if err != nil {
				t.Fatalf("ParseSchema failed: %v", err)
			}

			errors := ValidateNaming(schema)
			if len(errors) == 0 {
				t.Fatalf("Expected error for reserved keyword %q, got none", tc.expectedError)
			}

			errMsg := errors[0].Error()
			if !strings.Contains(errMsg, tc.expectedError) {
				t.Errorf("Expected error to mention %q, got: %s", tc.expectedError, errMsg)
			}
			if !strings.Contains(errMsg, "reserved") {
				t.Errorf("Expected error to mention 'reserved', got: %s", errMsg)
			}
		})
	}
}

func TestInvalidCharacters(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expectedError string
	}{
		{
			name: "struct name starts with digit",
			input: `
			struct 2Device {
				id: u32,
			}
			`,
			expectedError: "must start with a letter or underscore",
		},
		{
			name: "struct name with hyphen",
			input: `
			struct Device-Config {
				id: u32,
			}
			`,
			expectedError: "invalid character",
		},
		{
			name: "field name with space",
			input: `
			struct Device {
				device name: str,
			}
			`,
			expectedError: "invalid character",
		},
		{
			name: "field name with special char",
			input: `
			struct Config {
				value$: u32,
			}
			`,
			expectedError: "invalid character",
		},
		{
			name: "field name starts with digit",
			input: `
			struct Data {
				1value: u32,
			}
			`,
			expectedError: "must start with a letter or underscore",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := parser.ParseSchema(tc.input)
			if err != nil {
				// Parser might catch some errors first (like space in identifier)
				// That's fine - we just need to ensure invalid names are caught somewhere
				if strings.Contains(err.Error(), "expected") {
					return // Parser caught it
				}
				t.Fatalf("ParseSchema failed unexpectedly: %v", err)
			}

			errors := ValidateNaming(schema)
			if len(errors) == 0 {
				t.Fatalf("Expected error containing %q, got none", tc.expectedError)
			}

			errMsg := errors[0].Error()
			if !strings.Contains(errMsg, tc.expectedError) {
				t.Errorf("Expected error containing %q, got: %s", tc.expectedError, errMsg)
			}
		})
	}
}

func TestDuplicateFields(t *testing.T) {
	input := `
	struct Device {
		id: u32,
		name: str,
		id: u64,
	}
	`

	schema, err := parser.ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	errors := ValidateNaming(schema)
	if len(errors) == 0 {
		t.Fatal("Expected error for duplicate field, got none")
	}

	errMsg := errors[0].Error()
	if !strings.Contains(errMsg, "duplicate field") {
		t.Errorf("Expected error to mention 'duplicate field', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "id") {
		t.Errorf("Expected error to mention field 'id', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "Device") {
		t.Errorf("Expected error to mention struct 'Device', got: %s", errMsg)
	}
}

func TestDuplicateStructs(t *testing.T) {
	input := `
	struct Device {
		id: u32,
	}
	
	struct Config {
		value: u32,
	}
	
	struct Device {
		name: str,
	}
	`

	schema, err := parser.ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	errors := ValidateNaming(schema)
	if len(errors) == 0 {
		t.Fatal("Expected error for duplicate struct, got none")
	}

	errMsg := errors[0].Error()
	if !strings.Contains(errMsg, "duplicate struct") {
		t.Errorf("Expected error to mention 'duplicate struct', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "Device") {
		t.Errorf("Expected error to mention struct 'Device', got: %s", errMsg)
	}
}

func TestMultipleErrors(t *testing.T) {
	input := `
	struct type {
		id: u32,
		len: u32,
		name: str,
		name: str,
	}
	
	struct Device {
		type: u32,
	}
	
	struct Device {
		value: u32,
	}
	`

	schema, err := parser.ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	errors := ValidateNaming(schema)

	// Should have multiple errors:
	// - struct "type" is reserved
	// - field "len" is reserved
	// - duplicate field "name"
	// - field "type" is reserved
	// - duplicate struct "Device"
	if len(errors) < 4 {
		t.Errorf("Expected at least 4 errors, got %d: %v", len(errors), errors)
	}

	allErrors := ""
	for _, e := range errors {
		allErrors += e.Error() + " "
	}

	// Check key issues are mentioned
	if !strings.Contains(allErrors, "duplicate") {
		t.Errorf("Expected errors to mention 'duplicate', got: %s", allErrors)
	}
	if !strings.Contains(allErrors, "reserved") {
		t.Errorf("Expected errors to mention 'reserved', got: %s", allErrors)
	}
}

func TestCaseSensitiveNames(t *testing.T) {
	// Different case should be valid (not duplicates)
	input := `
	struct Device {
		id: u32,
	}
	
	struct device {
		id: u32,
	}
	`

	schema, err := parser.ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	errors := ValidateNaming(schema)

	// Should NOT have duplicate struct error (case-sensitive)
	// But WILL have reserved keyword errors since "device" might be common
	// Let's just check that if there are errors, they're not about duplicates
	for _, e := range errors {
		if strings.Contains(e.Error(), "duplicate struct") {
			t.Errorf("Struct names should be case-sensitive, got duplicate error: %s", e.Error())
		}
	}
}

func TestNamingEmptySchema(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{},
	}

	errors := ValidateNaming(schema)
	if len(errors) != 0 {
		t.Errorf("Expected no errors for empty schema, got: %v", errors)
	}
}

func TestUnicodeIdentifiers(t *testing.T) {
	// Unicode identifiers are NOT supported (ASCII-only, Rust-style)
	// The parser should reject these
	input := `
	struct Données {
		valeur: u32,
	}
	`

	_, err := parser.ParseSchema(input)
	if err == nil {
		t.Fatal("Expected parser to reject unicode identifiers, but it succeeded")
	}

	// Parser correctly rejects unicode - this is expected behavior
	if !strings.Contains(err.Error(), "unexpected") {
		t.Errorf("Expected 'unexpected' in parser error for unicode, got: %v", err)
	}
}

func TestUnderscorePrefix(t *testing.T) {
	// Leading underscore is valid
	input := `
	struct _Internal {
		_private: u32,
		_hidden: str,
	}
	`

	schema, err := parser.ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	errors := ValidateNaming(schema)

	// Filter out reserved errors (underscore itself might be reserved in Swift)
	// Check there are no "invalid character" or "must start with" errors
	for _, e := range errors {
		errMsg := e.Error()
		if strings.Contains(errMsg, "invalid character") || strings.Contains(errMsg, "must start with") {
			t.Errorf("Underscore prefix should be valid, got error: %s", errMsg)
		}
	}
}

func TestMixedValidAndInvalid(t *testing.T) {
	input := `
	struct ValidName {
		valid_field: u32,
		type: u32,
		another_valid: str,
	}
	
	struct AnotherValid {
		x: u32,
		y: u32,
	}
	`

	schema, err := parser.ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	errors := ValidateNaming(schema)

	// Should have exactly 1 error (field "type" is reserved)
	if len(errors) != 1 {
		t.Errorf("Expected 1 error for reserved 'type', got %d: %v", len(errors), errors)
	}

	if len(errors) > 0 {
		errMsg := errors[0].Error()
		if !strings.Contains(errMsg, "type") {
			t.Errorf("Expected error about 'type', got: %s", errMsg)
		}
	}
}
