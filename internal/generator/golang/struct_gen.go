package golang

import (
	"fmt"
	"strings"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// GenerateStructs generates Go struct definitions from a schema.
// It converts all struct definitions to idiomatic Go code with:
//   - PascalCase struct and field names
//   - Exported fields (start with capital letter)
//   - Doc comments preserved from schema
//   - Proper Go type mappings
//
// Example output:
//
//	// Device represents an audio device.
//	type Device struct {
//	    // ID is the unique identifier.
//	    ID uint32
//	    // Name is the device name.
//	    Name string
//	}
//
// Returns an error if type mapping fails or schema is invalid.
func GenerateStructs(schema *parser.Schema) (string, error) {
	if schema == nil {
		return "", fmt.Errorf("schema is nil")
	}

	if len(schema.Structs) == 0 {
		return "", fmt.Errorf("schema has no structs")
	}

	var buf strings.Builder

	for i, s := range schema.Structs {
		// Add blank line between structs (except before first)
		if i > 0 {
			buf.WriteString("\n")
		}

		// Generate doc comment for struct
		if s.Comment != "" {
			buf.WriteString("// ")
			buf.WriteString(ToGoName(s.Name))
			buf.WriteString(" ")
			buf.WriteString(s.Comment)
			buf.WriteString("\n")
		}

		// Generate struct declaration
		buf.WriteString("type ")
		buf.WriteString(ToGoName(s.Name))
		buf.WriteString(" struct {\n")

		// Generate fields
		for _, field := range s.Fields {
			// Field doc comment
			if field.Comment != "" {
				buf.WriteString("\t// ")
				buf.WriteString(ToGoName(field.Name))
				buf.WriteString(" ")
				buf.WriteString(field.Comment)
				buf.WriteString("\n")
			}

			// Field declaration
			buf.WriteString("\t")
			buf.WriteString(ToGoName(field.Name))
			buf.WriteString(" ")

			// Map field type
			goType, err := mapFieldType(&field.Type)
			if err != nil {
				return "", fmt.Errorf("struct %q, field %q: %w", s.Name, field.Name, err)
			}
			buf.WriteString(goType)
			buf.WriteString("\n")
		}

		buf.WriteString("}\n")
	}

	return buf.String(), nil
}

// mapFieldType converts a field's type expression to Go type string,
// applying name conversion to named types.
func mapFieldType(typeExpr *parser.TypeExpr) (string, error) {
	if typeExpr == nil {
		return "", fmt.Errorf("type expression is nil")
	}

	// First get the base type
	var baseType string

	switch typeExpr.Kind {
	case parser.TypeKindPrimitive:
		// Primitives use direct mapping
		goType, ok := primitiveTypeMap[typeExpr.Name]
		if !ok {
			return "", fmt.Errorf("unknown primitive type: %q", typeExpr.Name)
		}
		baseType = goType

	case parser.TypeKindNamed:
		// Named types need PascalCase conversion
		baseType = ToGoName(typeExpr.Name)

	case parser.TypeKindArray:
		if typeExpr.Elem == nil {
			return "", fmt.Errorf("array type has no element type")
		}
		elemType, err := mapFieldType(typeExpr.Elem)
		if err != nil {
			return "", fmt.Errorf("array element type error: %w", err)
		}
		baseType = "[]" + elemType

	default:
		return "", fmt.Errorf("unknown type kind: %v", typeExpr.Kind)
	}

	// Wrap in pointer if optional (Option<T> → *T)
	if typeExpr.Optional {
		baseType = "*" + baseType
	}

	return baseType, nil
}
