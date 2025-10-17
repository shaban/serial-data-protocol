// Package parser implements parsing of Serial Data Protocol schema files (.sdp).
package parser

// Schema represents a complete parsed schema file.
type Schema struct {
	Structs []Struct
}

// Struct represents a struct definition in the schema.
type Struct struct {
	Name    string
	Comment string   // Doc comment (from /// lines)
	Fields  []Field
}

// Field represents a field in a struct.
type Field struct {
	Name    string
	Type    TypeExpr
	Comment string   // Doc comment (from /// lines)
}

// TypeExpr represents a type expression (primitive, array, or named type).
type TypeExpr struct {
	Kind TypeKind
	Name string   // For Named types (e.g., "MyStruct", "u32")
	Elem *TypeExpr // For Array types, points to element type
}

// TypeKind identifies the kind of type expression.
type TypeKind int

const (
	TypeKindPrimitive TypeKind = iota // u8, u16, u32, u64, i8, i16, i32, i64, f32, f64, bool, string
	TypeKindNamed                     // User-defined struct type
	TypeKindArray                     // []T
)

// IsPrimitive returns true if this type is a primitive type.
func (t *TypeExpr) IsPrimitive() bool {
	if t.Kind != TypeKindPrimitive {
		return false
	}
	primitives := map[string]bool{
		"u8": true, "u16": true, "u32": true, "u64": true,
		"i8": true, "i16": true, "i32": true, "i64": true,
		"f32": true, "f64": true,
		"bool": true, "string": true,
	}
	return primitives[t.Name]
}

// String returns a string representation of the type.
func (t *TypeExpr) String() string {
	switch t.Kind {
	case TypeKindPrimitive, TypeKindNamed:
		return t.Name
	case TypeKindArray:
		if t.Elem != nil {
			return "[]" + t.Elem.String()
		}
		return "[]?"
	default:
		return "?"
	}
}
