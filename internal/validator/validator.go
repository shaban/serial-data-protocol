package validator

import (
	"fmt"
	"strings"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// Validate runs all validators on the schema and returns a combined error
// if any validation fails. Returns nil if the schema is valid.
//
// Validators run in order:
// 1. Structure validation (empty schemas/structs)
// 2. Type reference validation (unknown types)
// 3. Cycle detection (circular references)
// 4. Naming validation (identifiers, reserved words, duplicates)
//
// All validators are run even if earlier ones fail, so that all errors
// can be reported at once.
func Validate(schema *parser.Schema) error {
	var allErrors []error

	// Run all validators
	allErrors = append(allErrors, ValidateStructure(schema)...)
	allErrors = append(allErrors, ValidateTypeReferences(schema)...)
	allErrors = append(allErrors, DetectCycles(schema)...)
	allErrors = append(allErrors, ValidateNaming(schema)...)

	// If no errors, schema is valid
	if len(allErrors) == 0 {
		return nil
	}

	// Combine all errors into a single error message
	// Format: one error per line for readability
	var messages []string
	for _, err := range allErrors {
		messages = append(messages, err.Error())
	}

	return fmt.Errorf("schema validation failed:\n  %s", strings.Join(messages, "\n  "))
}
