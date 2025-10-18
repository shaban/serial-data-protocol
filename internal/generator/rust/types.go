package rust

import (
	"fmt"
	"strings"
	"unicode"
)

// ToRustName converts an SDP identifier to idiomatic Rust (snake_case).
// Examples:
//   - "MyStruct" → "MyStruct" (types stay PascalCase)
//   - "fieldName" → "field_name"
//   - "u32Field" → "u32_field"
func ToRustName(name string) string {
	// For type names, keep PascalCase
	if unicode.IsUpper(rune(name[0])) {
		return name
	}
	
	// For field names, convert to snake_case
	return toSnakeCase(name)
}

// toSnakeCase converts camelCase or PascalCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	
	for i, r := range s {
		if unicode.IsUpper(r) && i > 0 {
			// Add underscore before uppercase letter (unless start)
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	
	return result.String()
}

// MapTypeToRust converts SDP type to Rust type.
// Primitives:
//   - u8, u16, u32, u64, i8, i16, i32, i64 → same
//   - f32, f64 → same
//   - bool → bool
//   - str → String (owned)
//   - bytes → Vec<u8>
//
// Arrays: []T → Vec<T>
// Optional: ?T → Option<T>
// Named types: CustomType → CustomType (PascalCase)
func MapTypeToRust(typeName string, isArray bool, isOptional bool) (string, error) {
	// Base type mapping
	var rustType string
	
	switch typeName {
	case "u8", "u16", "u32", "u64":
		rustType = typeName
	case "i8", "i16", "i32", "i64":
		rustType = typeName
	case "f32", "f64":
		rustType = typeName
	case "bool":
		rustType = "bool"
	case "str":
		rustType = "String"
	case "bytes":
		rustType = "Vec<u8>"
	default:
		// Named type (struct reference)
		rustType = typeName
	}
	
	// Wrap in Vec if array
	if isArray {
		rustType = fmt.Sprintf("Vec<%s>", rustType)
	}
	
	// Wrap in Option if optional
	if isOptional {
		rustType = fmt.Sprintf("Option<%s>", rustType)
	}
	
	return rustType, nil
}

// WireTypeToRust converts SDP primitive type to wire_slice function suffix
// Used for generating encode/decode calls
func WireTypeToRust(typeName string) string {
	switch typeName {
	case "u8":
		return "u8"
	case "u16":
		return "u16"
	case "u32":
		return "u32"
	case "u64":
		return "u64"
	case "i8":
		return "i8"
	case "i16":
		return "i16"
	case "i32":
		return "i32"
	case "i64":
		return "i64"
	case "f32":
		return "f32"
	case "f64":
		return "f64"
	case "bool":
		return "bool"
	case "str":
		return "string"
	case "bytes":
		return "bytes"
	default:
		return ""
	}
}

// IsPrimitive returns true if the type is a primitive (not a named struct)
func IsPrimitive(typeName string) bool {
	primitives := map[string]bool{
		"u8": true, "u16": true, "u32": true, "u64": true,
		"i8": true, "i16": true, "i32": true, "i64": true,
		"f32": true, "f64": true,
		"bool": true,
		"str": true,
		"bytes": true,
	}
	return primitives[typeName]
}

// FixedSize returns the wire format size for fixed-size primitives
// Returns 0 for variable-size types (str, bytes, arrays)
func FixedSize(typeName string) int {
	sizes := map[string]int{
		"u8": 1, "i8": 1, "bool": 1,
		"u16": 2, "i16": 2,
		"u32": 4, "i32": 4, "f32": 4,
		"u64": 8, "i64": 8, "f64": 8,
	}
	return sizes[typeName]
}
