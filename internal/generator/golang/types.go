// Package golang provides Go code generation for Serial Data Protocol schemas.
package golang

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// primitiveTypeMap maps SDP primitive types to Go types.
var primitiveTypeMap = map[string]string{
	// Unsigned integers
	"u8":  "uint8",
	"u16": "uint16",
	"u32": "uint32",
	"u64": "uint64",

	// Signed integers
	"i8":  "int8",
	"i16": "int16",
	"i32": "int32",
	"i64": "int64",

	// Floating point
	"f32": "float32",
	"f64": "float64",

	// Other primitives
	"bool": "bool",
	"str":  "string",
}

// MapTypeToGo converts a schema type expression to its Go type representation.
// It handles primitives, arrays, and named types (user-defined structs).
//
// Examples:
//   - u32 → uint32
//   - str → string
//   - []u8 → []uint8
//   - []Device → []Device (struct name preserved as-is)
//   - Device → Device (struct name preserved as-is)
//
// Named struct types are kept as-is; name conversion (e.g., snake_case → PascalCase)
// is handled separately by name conversion functions.
//
// Returns an error if the type expression is invalid or references an unknown primitive.
func MapTypeToGo(typeExpr *parser.TypeExpr) (string, error) {
	if typeExpr == nil {
		return "", fmt.Errorf("type expression is nil")
	}

	switch typeExpr.Kind {
	case parser.TypeKindPrimitive:
		goType, ok := primitiveTypeMap[typeExpr.Name]
		if !ok {
			return "", fmt.Errorf("unknown primitive type: %q", typeExpr.Name)
		}
		return goType, nil

	case parser.TypeKindNamed:
		// Named types (user-defined structs) are kept as-is
		// Name conversion happens separately
		return typeExpr.Name, nil

	case parser.TypeKindArray:
		if typeExpr.Elem == nil {
			return "", fmt.Errorf("array type has no element type")
		}
		elemType, err := MapTypeToGo(typeExpr.Elem)
		if err != nil {
			return "", fmt.Errorf("array element type error: %w", err)
		}
		return "[]" + elemType, nil

	default:
		return "", fmt.Errorf("unknown type kind: %v", typeExpr.Kind)
	}
}

// ToGoName converts a schema identifier to Go naming convention (PascalCase).
// This function is used for both struct names and field names in Go, as both
// need to be exported (start with capital letter).
//
// Conversion rules:
//   - snake_case → PascalCase (audio_device → AudioDevice)
//   - Preserve existing capitals (AudioDevice → AudioDevice)
//   - Single words capitalized (device → Device)
//   - Multiple underscores treated as single separator (audio__device → AudioDevice)
//   - Leading/trailing underscores removed (_device_ → Device)
//
// Examples:
//   - "device" → "Device"
//   - "audio_device" → "AudioDevice"
//   - "AudioDevice" → "AudioDevice"
//   - "myStruct" → "MyStruct"
//   - "my_struct_name" → "MyStructName"
//   - "http_response" → "HttpResponse"
//   - "HTTPResponse" → "HTTPResponse"
func ToGoName(name string) string {
	if name == "" {
		return ""
	}

	// Split by underscores
	parts := strings.Split(name, "_")
	
	var result strings.Builder
	for _, part := range parts {
		if part == "" {
			// Skip empty parts (from multiple underscores or leading/trailing)
			continue
		}
		
		// Capitalize first letter, preserve rest
		result.WriteString(capitalizeFirst(part))
	}
	
	return result.String()
}

// capitalizeFirst capitalizes the first rune of a string, preserving the rest.
// If the string is already capitalized or has mixed case, it preserves the original.
//
// Examples:
//   - "device" → "Device"
//   - "Device" → "Device"
//   - "HTTP" → "HTTP"
//   - "myName" → "MyName"
func capitalizeFirst(s string) string {
	if s == "" {
		return ""
	}
	
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
