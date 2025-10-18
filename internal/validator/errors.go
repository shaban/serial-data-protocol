package validator

import "fmt"

// Standard validation error codes for consistent error reporting across
// all SDP implementations. These codes are stable and can be relied upon
// by tools and other language implementations.
const (
	// Structure validation errors
	ErrCodeEmptySchema       = "EMPTY_SCHEMA"        // Schema defines no structs
	ErrCodeEmptyStruct       = "EMPTY_STRUCT"        // Struct has no fields
	
	// Type validation errors
	ErrCodeUnknownType       = "UNKNOWN_TYPE"        // Type reference to undefined struct
	ErrCodeInvalidPrimitive  = "INVALID_PRIMITIVE"   // Invalid primitive type name
	
	// Cycle detection errors
	ErrCodeCircularReference = "CIRCULAR_REFERENCE"  // Circular struct reference detected
	
	// Naming validation errors
	ErrCodeInvalidIdentifier = "INVALID_IDENTIFIER"  // Identifier violates naming rules
	ErrCodeReservedKeyword   = "RESERVED_KEYWORD"    // Identifier is reserved in target language
	ErrCodeDuplicateStruct   = "DUPLICATE_STRUCT"    // Multiple structs with same name
	ErrCodeDuplicateField    = "DUPLICATE_FIELD"     // Multiple fields with same name in struct
)

// Error constructors for consistent error messages

func errEmptySchema() ValidationError {
	return ValidationError{
		Message: "[EMPTY_SCHEMA] schema must define at least one struct",
	}
}

func errEmptyStruct(structName string) ValidationError {
	return ValidationError{
		Message: fmt.Sprintf("[EMPTY_STRUCT] struct %q cannot be empty (must have at least one field)", structName),
	}
}

func errUnknownType(structName, fieldName, typeName string) ValidationError {
	return ValidationError{
		Message: fmt.Sprintf("[UNKNOWN_TYPE] struct %q field %q: unknown type %q", structName, fieldName, typeName),
	}
}

func errCircularReference(cyclePath string) ValidationError {
	return ValidationError{
		Message: fmt.Sprintf("[CIRCULAR_REFERENCE] circular reference detected: %s", cyclePath),
	}
}

func errInvalidIdentifier(identifierType, name, reason string) ValidationError {
	return ValidationError{
		Message: fmt.Sprintf("[INVALID_IDENTIFIER] %s name %q is invalid: %s", identifierType, name, reason),
	}
}

func errReservedKeyword(identifierType, name string, languages []string) ValidationError {
	langList := ""
	if len(languages) > 0 {
		langList = fmt.Sprintf(" (reserved in: %v)", languages)
	}
	return ValidationError{
		Message: fmt.Sprintf("[RESERVED_KEYWORD] %s name %q is a reserved keyword%s", identifierType, name, langList),
	}
}

func errDuplicateStruct(name string) ValidationError {
	return ValidationError{
		Message: fmt.Sprintf("[DUPLICATE_STRUCT] duplicate struct name %q", name),
	}
}

func errDuplicateField(structName, fieldName string) ValidationError {
	return ValidationError{
		Message: fmt.Sprintf("[DUPLICATE_FIELD] struct %q has duplicate field name %q", structName, fieldName),
	}
}
