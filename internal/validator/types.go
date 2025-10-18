package validator

import (
	"fmt"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// ValidationError represents a single validation error with context.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

// ValidateTypeReferences checks that all field types in the schema resolve to either:
// - A primitive type (u8-u64, i8-i64, f32, f64, bool, str)
// - A struct defined in the same schema
// - An array of a valid type []T
//
// Returns all errors found (does not stop at first error).
func ValidateTypeReferences(schema *parser.Schema) []error {
	var errors []error

	// Build a set of defined struct names for quick lookup
	structNames := make(map[string]bool)
	for _, s := range schema.Structs {
		structNames[s.Name] = true
	}

	// Validate each struct's fields
	for _, s := range schema.Structs {
		for _, field := range s.Fields {
			if err := validateTypeExpr(&field.Type, structNames, s.Name, field.Name); err != nil {
				errors = append(errors, err)
			}
		}
	}

	return errors
}

// validateTypeExpr recursively validates a type expression.
func validateTypeExpr(typeExpr *parser.TypeExpr, structNames map[string]bool, structName, fieldName string) error {
	// Validate optional fields - Option<T> where T must be a struct
	if typeExpr.Optional {
		if typeExpr.Kind == parser.TypeKindPrimitive {
			return ValidationError{
				Message: fmt.Sprintf("struct %q, field %q: Option<T> cannot wrap primitive types (use regular field instead)",
					structName, fieldName),
			}
		}
		if typeExpr.Kind == parser.TypeKindArray {
			return ValidationError{
				Message: fmt.Sprintf("struct %q, field %q: Option<T> cannot wrap array types (use empty array instead)",
					structName, fieldName),
			}
		}
		// Option<StructType> is valid, continue validation below
	}

	// Validate Box<T> - typically used with recursive types
	if typeExpr.Boxed {
		if typeExpr.Kind != parser.TypeKindNamed {
			return ValidationError{
				Message: fmt.Sprintf("struct %q, field %q: Box<T> can only wrap struct types",
					structName, fieldName),
			}
		}
		// Box<StructType> is valid, continue validation below
	}

	switch typeExpr.Kind {
	case parser.TypeKindPrimitive:
		// Primitive types are always valid (already validated by parser)
		return nil

	case parser.TypeKindNamed:
		// Named type must be a defined struct
		if !structNames[typeExpr.Name] {
			return errUnknownType(structName, fieldName, typeExpr.Name)
		}
		return nil

	case parser.TypeKindArray:
		// Array element type must be valid
		if typeExpr.Elem == nil {
			return ValidationError{
				Message: fmt.Sprintf("struct %q, field %q: array has no element type",
					structName, fieldName),
			}
		}
		return validateTypeExpr(typeExpr.Elem, structNames, structName, fieldName)

	default:
		return ValidationError{
			Message: fmt.Sprintf("struct %q, field %q: unknown type kind %v",
				structName, fieldName, typeExpr.Kind),
		}
	}
}
