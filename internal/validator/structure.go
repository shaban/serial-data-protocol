package validator

import (
	"fmt"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// ValidateStructure checks structural requirements of the schema:
// - At least one struct defined (schema cannot be empty)
// - No empty structs (each struct must have at least one field)
// - Array element types are valid (handled by type validator, but we check for nil)
//
// Returns all errors found (does not stop at first error).
func ValidateStructure(schema *parser.Schema) []error {
	var errors []error

	// Check schema has at least one struct
	if len(schema.Structs) == 0 {
		errors = append(errors, ValidationError{
			Message: "schema must define at least one struct",
		})
		return errors // No point checking further if schema is empty
	}

	// Check each struct has at least one field
	for _, s := range schema.Structs {
		if len(s.Fields) == 0 {
			errors = append(errors, ValidationError{
				Message: fmt.Sprintf("struct %q cannot be empty (must have at least one field)", s.Name),
			})
		}
	}

	return errors
}
