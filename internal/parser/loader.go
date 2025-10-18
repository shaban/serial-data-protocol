package parser

import (
	"fmt"
	"os"
	"strings"
)

// LoadSchemaFile reads and parses a schema file from the given path.
// It normalizes line endings (CRLF → LF) and wraps any errors with the filename.
func LoadSchemaFile(path string) (*Schema, error) {
	// Read the file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read schema file %q: %w", path, err)
	}

	// Normalize line endings (CRLF → LF)
	input := strings.ReplaceAll(string(data), "\r\n", "\n")

	// Parse the schema
	schema, err := ParseSchema(input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse schema file %q: %w", path, err)
	}

	return schema, nil
}
