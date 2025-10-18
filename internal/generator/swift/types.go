package swift

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// ToSwiftName converts a name to Swift PascalCase convention
// Examples: "audio_unit" -> "AudioUnit", "plugin_id" -> "PluginId"
func ToSwiftName(name string) string {
	if name == "" {
		return ""
	}

	// Split on underscores
	parts := strings.Split(name, "_")
	result := ""

	for _, part := range parts {
		if part == "" {
			continue
		}
		// Capitalize first letter, keep rest as-is
		result += strings.ToUpper(part[:1]) + part[1:]
	}

	return result
}

// toSwiftFieldName converts a name to Swift camelCase convention for fields
// Examples: "plugin_id" -> "pluginId", "max_value" -> "maxValue"
func toSwiftFieldName(name string) string {
	if name == "" {
		return ""
	}

	// Split on underscores
	parts := strings.Split(name, "_")
	if len(parts) == 0 {
		return ""
	}

	result := parts[0] // First part stays lowercase

	for i := 1; i < len(parts); i++ {
		if parts[i] == "" {
			continue
		}
		// Capitalize first letter of subsequent parts
		result += strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}

	return result
}

// MapTypeToSwift converts an SDP type to Swift type
func MapTypeToSwift(t *parser.TypeExpr) string {
	if t.Optional {
		inner := &parser.TypeExpr{
			Kind:  t.Kind,
			Name:  t.Name,
			Elem:  t.Elem,
			Boxed: t.Boxed,
		}
		return fmt.Sprintf("%s?", MapTypeToSwift(inner))
	}

	switch t.Kind {
	case parser.TypeKindPrimitive:
		return mapPrimitiveToSwift(t.Name)
	case parser.TypeKindNamed:
		return ToSwiftName(t.Name)
	case parser.TypeKindArray:
		if t.Elem != nil {
			elementType := MapTypeToSwift(t.Elem)
			return fmt.Sprintf("[%s]", elementType)
		}
		return "[UnknownType]"
	default:
		return "UnknownType"
	}
}

// mapPrimitiveToSwift maps SDP primitive types to Swift types
func mapPrimitiveToSwift(primitive string) string {
	switch primitive {
	case "u8":
		return "UInt8"
	case "u16":
		return "UInt16"
	case "u32":
		return "UInt32"
	case "u64":
		return "UInt64"
	case "i8":
		return "Int8"
	case "i16":
		return "Int16"
	case "i32":
		return "Int32"
	case "i64":
		return "Int64"
	case "f32":
		return "Float"
	case "f64":
		return "Double"
	case "bool":
		return "Bool"
	case "str", "string":
		return "String"
	default:
		return "UnknownType"
	}
}

// IsPrimitive checks if a type is a primitive
func IsPrimitive(t *parser.TypeExpr) bool {
	return t.Kind == parser.TypeKindPrimitive
}

// FixedSize returns the fixed byte size of a type, or -1 if variable size
func FixedSize(t *parser.TypeExpr) int {
	if t.Optional || t.Kind == parser.TypeKindArray {
		return -1 // Variable size
	}

	if t.Kind == parser.TypeKindPrimitive {
		switch t.Name {
		case "u8", "i8", "bool":
			return 1
		case "u16", "i16":
			return 2
		case "u32", "i32", "f32":
			return 4
		case "u64", "i64", "f64":
			return 8
		case "str":
			return -1 // Variable size
		}
	}

	return -1 // Named types and arrays are variable size
}

// GetSwiftEncodingMethod returns the Swift method name for encoding a primitive type
// Examples: "u32" -> "UInt32", "string" -> "String", "bool" -> "Bool"
func GetSwiftEncodingMethod(primitive string) string {
	return mapPrimitiveToSwift(primitive)
}

// NeedsImport checks if a type requires additional imports
func NeedsImport(t *parser.TypeExpr) bool {
	// Swift Foundation types (String, Data) are automatically available
	// No special imports needed for our use case
	return false
}

// IsSwiftKeyword checks if a name is a Swift keyword that needs escaping
func IsSwiftKeyword(name string) bool {
	keywords := map[string]bool{
		"associatedtype": true,
		"class":          true,
		"deinit":         true,
		"enum":           true,
		"extension":      true,
		"fileprivate":    true,
		"func":           true,
		"import":         true,
		"init":           true,
		"inout":          true,
		"internal":       true,
		"let":            true,
		"open":           true,
		"operator":       true,
		"private":        true,
		"protocol":       true,
		"public":         true,
		"rethrows":       true,
		"static":         true,
		"struct":         true,
		"subscript":      true,
		"typealias":      true,
		"var":            true,
		"break":          true,
		"case":           true,
		"continue":       true,
		"default":        true,
		"defer":          true,
		"do":             true,
		"else":           true,
		"fallthrough":    true,
		"for":            true,
		"guard":          true,
		"if":             true,
		"in":             true,
		"repeat":         true,
		"return":         true,
		"switch":         true,
		"where":          true,
		"while":          true,
		"as":             true,
		"Any":            true,
		"catch":          true,
		"false":          true,
		"is":             true,
		"nil":            true,
		"super":          true,
		"self":           true,
		"Self":           true,
		"throw":          true,
		"throws":         true,
		"true":           true,
		"try":            true,
		"_":              true,
	}
	return keywords[name]
}

// EscapeSwiftKeyword escapes a Swift keyword by backticks
func EscapeSwiftKeyword(name string) string {
	if IsSwiftKeyword(name) {
		return "`" + name + "`"
	}
	return name
}

// IsValidSwiftIdentifier checks if a string is a valid Swift identifier
func IsValidSwiftIdentifier(name string) bool {
	if name == "" {
		return false
	}

	// First character must be letter or underscore
	firstChar := rune(name[0])
	if !unicode.IsLetter(firstChar) && firstChar != '_' {
		return false
	}

	// Subsequent characters can be letters, digits, or underscores
	for _, ch := range name[1:] {
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			return false
		}
	}

	return true
}
