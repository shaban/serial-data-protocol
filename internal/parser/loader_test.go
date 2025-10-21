package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadSchemaFile_Valid(t *testing.T) {
	tests := []struct {
		name            string
		file            string
		expectedStructs int
	}{
		{
			name:            "basic schema",
			file:            "valid_basic.sdp",
			expectedStructs: 1,
		},
		{
			name:            "complex schema",
			file:            "valid_complex.sdp",
			expectedStructs: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join("..", "..", "testdata", "schemas", tt.file)
			schema, err := LoadSchemaFile(path)
			if err != nil {
				t.Fatalf("LoadSchemaFile() error = %v", err)
			}

			if len(schema.Structs) != tt.expectedStructs {
				t.Errorf("expected %d structs, got %d", tt.expectedStructs, len(schema.Structs))
			}
		})
	}
}

func TestLoadSchemaFile_MissingFile(t *testing.T) {
	_, err := LoadSchemaFile("nonexistent.sdp")
	if err == nil {
		t.Fatal("expected error for missing file")
	}

	if !strings.Contains(err.Error(), "failed to read schema file") {
		t.Errorf("expected 'failed to read schema file' in error, got: %v", err)
	}
	if !strings.Contains(err.Error(), "nonexistent.sdp") {
		t.Errorf("expected filename in error, got: %v", err)
	}
}

func TestLoadSchemaFile_InvalidSyntax(t *testing.T) {
	// Create a temporary file with invalid syntax
	tmpFile, err := os.CreateTemp("", "invalid_*.sdp")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	content := `struct InvalidSyntax { missing_colon str }`
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	_, err = LoadSchemaFile(tmpFile.Name())
	if err == nil {
		t.Fatal("expected error for invalid syntax")
	}

	if !strings.Contains(err.Error(), "failed to parse schema file") {
		t.Errorf("expected 'failed to parse schema file' in error, got: %v", err)
	}
}

func TestLoadSchemaFile_CRLF(t *testing.T) {
	// Create a temporary file with CRLF line endings
	tmpFile, err := os.CreateTemp("", "crlf_*.sdp")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write content with CRLF
	content := "/// Doc comment\r\nstruct Example {\r\n    field: u32,\r\n}\r\n"
	if _, err := tmpFile.WriteString(content); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	schema, err := LoadSchemaFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadSchemaFile() error = %v", err)
	}

	if len(schema.Structs) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(schema.Structs))
	}

	if schema.Structs[0].Name != "Example" {
		t.Errorf("expected struct name 'Example', got %q", schema.Structs[0].Name)
	}

	if schema.Structs[0].Comment != "Doc comment" {
		t.Errorf("expected comment 'Doc comment', got %q", schema.Structs[0].Comment)
	}
}

func TestLoadSchemaFile_PreservesDocComments(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "schemas", "valid_basic.sdp")
	schema, err := LoadSchemaFile(path)
	if err != nil {
		t.Fatalf("LoadSchemaFile() error = %v", err)
	}

	if len(schema.Structs) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(schema.Structs))
	}

	if schema.Structs[0].Comment != "A simple device with ID and name" {
		t.Errorf("expected doc comment to be preserved, got: %q", schema.Structs[0].Comment)
	}
}
