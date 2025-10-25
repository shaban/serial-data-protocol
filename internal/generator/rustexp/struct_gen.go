package rustexp

import (
	"fmt"
	"strings"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// GenerateStructs generates Rust struct definitions from a schema.
// It converts all struct definitions to idiomatic Rust code with:
//   - PascalCase struct names
//   - snake_case field names
//   - Proper derive macros (Debug, Clone, PartialEq)
//   - Doc comments preserved from schema
//   - Proper Rust type mappings
//
// Example output:
//
//	/// Device represents an audio device.
//	#[derive(Debug, Clone, PartialEq)]
//	pub struct Device {
//	    /// ID is the unique identifier.
//	    pub id: u32,
//	    /// Name is the device name.
//	    pub name: String,
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
			buf.WriteString("/// ")
			buf.WriteString(s.Name)
			buf.WriteString(" ")
			buf.WriteString(s.Comment)
			buf.WriteString("\n")
		}

		// Generate derive macro
		buf.WriteString("#[derive(Debug, Clone, PartialEq)]\n")

		// Generate struct declaration
		buf.WriteString("pub struct ")
		buf.WriteString(s.Name)
		buf.WriteString(" {\n")

		// Generate fields
		for _, field := range s.Fields {
			// Field doc comment
			if field.Comment != "" {
				buf.WriteString("    /// ")
				buf.WriteString(ToRustName(field.Name))
				buf.WriteString(" ")
				buf.WriteString(field.Comment)
				buf.WriteString("\n")
			}

			// Field declaration
			buf.WriteString("    pub ")
			buf.WriteString(ToRustName(field.Name))
			buf.WriteString(": ")

			// Map field type
			rustType, err := mapFieldType(&field.Type)
			if err != nil {
				return "", fmt.Errorf("struct %q, field %q: %w", s.Name, field.Name, err)
			}
			buf.WriteString(rustType)
			buf.WriteString(",\n")
		}

		buf.WriteString("}\n")
	}

	return buf.String(), nil
}

// mapFieldType converts a parser.TypeExpr to a Rust type string
func mapFieldType(t *parser.TypeExpr) (string, error) {
	if t == nil {
		return "", fmt.Errorf("type is nil")
	}

	// Get base type name
	var typeName string
	switch t.Kind {
	case parser.TypeKindPrimitive:
		typeName = t.Name
	case parser.TypeKindNamed:
		typeName = t.Name
	case parser.TypeKindArray:
		if t.Elem == nil {
			return "", fmt.Errorf("array type missing element type")
		}
		elemType, err := mapFieldType(t.Elem)
		if err != nil {
			return "", err
		}
		typeName = fmt.Sprintf("Vec<%s>", elemType)

		// Handle optional arrays
		if t.Optional {
			typeName = fmt.Sprintf("Option<%s>", typeName)
		}
		return typeName, nil
	default:
		return "", fmt.Errorf("unknown type kind: %d", t.Kind)
	}

	// Map the type
	rustType, err := MapTypeToRust(typeName, false, t.Optional)
	if err != nil {
		return "", err
	}

	return rustType, nil
}
