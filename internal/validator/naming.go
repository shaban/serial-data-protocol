package validator

import (
	"fmt"
	"unicode"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// ValidateNaming checks that all struct and field names follow naming rules:
// - Valid identifier format (start with letter/underscore, alphanumeric + underscore)
// - Not reserved keywords in any target language
// - No duplicate struct names
// - No duplicate field names within a struct
//
// Returns all errors found (does not stop at first error).
func ValidateNaming(schema *parser.Schema) []error {
	var errors []error

	// Check for duplicate struct names
	structNames := make(map[string]bool)
	for _, s := range schema.Structs {
		if structNames[s.Name] {
			errors = append(errors, errDuplicateStruct(s.Name))
		}
		structNames[s.Name] = true
	}

	// Validate each struct
	for _, s := range schema.Structs {
		// Validate struct name format
		if err := validateIdentifier(s.Name, "struct"); err != nil {
			errors = append(errors, err)
		}

		// Check if struct name is reserved
		if IsReserved(s.Name) {
			langs := GetReservedLanguages(s.Name)
			errors = append(errors, errReservedKeyword("struct", s.Name, langs))
		}

		// Check for duplicate field names
		fieldNames := make(map[string]bool)
		for _, field := range s.Fields {
			if fieldNames[field.Name] {
				errors = append(errors, errDuplicateField(s.Name, field.Name))
			}
			fieldNames[field.Name] = true

			// Validate field name format
			if err := validateIdentifier(field.Name, "field"); err != nil {
				errors = append(errors, err)
			}

			// Check if field name is reserved
			if IsReserved(field.Name) {
				langs := GetReservedLanguages(field.Name)
				errors = append(errors, errReservedKeyword("field", field.Name, langs))
			}
		}
	}

	return errors
}

// validateIdentifier checks if a name follows identifier rules:
// - Must start with a letter (a-z, A-Z) or underscore (_)
// - Rest can be letters, digits (0-9), or underscores
// - Must not be empty
func validateIdentifier(name, identifierType string) error {
	if name == "" {
		return errInvalidIdentifier(identifierType, name, "cannot be empty")
	}

	// Check first character
	firstChar := rune(name[0])
	if !unicode.IsLetter(firstChar) && firstChar != '_' {
		return errInvalidIdentifier(identifierType, name, "must start with a letter or underscore")
	}

	// Check remaining characters
	for i, ch := range name {
		if i == 0 {
			continue // Already checked
		}
		if !unicode.IsLetter(ch) && !unicode.IsDigit(ch) && ch != '_' {
			reason := fmt.Sprintf("contains invalid character %q at position %d (only letters, digits, and underscores allowed)", string(ch), i)
			return errInvalidIdentifier(identifierType, name, reason)
		}
	}

	return nil
}
