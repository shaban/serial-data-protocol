package golang

import (
	"strings"
	"testing"
)

// TestGenerateErrors verifies that all error variables are generated
func TestGenerateErrors(t *testing.T) {
	result := GenerateErrors()

	if result == "" {
		t.Fatal("GenerateErrors returned empty string")
	}

	// Check that all five error variables are present
	expectedErrors := []string{
		"ErrUnexpectedEOF",
		"ErrInvalidUTF8",
		"ErrDataTooLarge",
		"ErrArrayTooLarge",
		"ErrTooManyElements",
	}

	for _, errName := range expectedErrors {
		if !strings.Contains(result, errName) {
			t.Errorf("missing error variable: %s", errName)
		}
	}
}

// TestGenerateErrorsMessages verifies the error messages
func TestGenerateErrorsMessages(t *testing.T) {
	result := GenerateErrors()

	expectedMessages := map[string]string{
		"ErrUnexpectedEOF":   "unexpected end of data",
		"ErrInvalidUTF8":     "invalid UTF-8 string",
		"ErrDataTooLarge":    "data exceeds 128MB limit",
		"ErrArrayTooLarge":   "array count exceeds per-array limit",
		"ErrTooManyElements": "total elements exceed limit",
	}

	for errName, errMsg := range expectedMessages {
		if !strings.Contains(result, errMsg) {
			t.Errorf("missing error message for %s: %q", errName, errMsg)
		}
	}
}

// TestGenerateErrorsFormat verifies the overall format
func TestGenerateErrorsFormat(t *testing.T) {
	result := GenerateErrors()

	// Check for var block declaration
	if !strings.Contains(result, "var (") {
		t.Error("missing 'var (' declaration")
	}

	// Check for errors.New usage
	if !strings.Contains(result, "errors.New(") {
		t.Error("missing 'errors.New(' usage")
	}

	// Check for closing parenthesis
	if !strings.Contains(result, "\n)") {
		t.Error("missing closing ')' for var block")
	}

	// Check for comment
	if !strings.Contains(result, "// Error variables for decode failures") {
		t.Error("missing documentation comment")
	}
}

// TestGenerateErrorsStructure verifies the variable declaration structure
func TestGenerateErrorsStructure(t *testing.T) {
	result := GenerateErrors()

	// Each error should have the pattern: \tErrName = errors.New("message")\n
	expectedPatterns := []string{
		"\tErrUnexpectedEOF      = errors.New(",
		"\tErrInvalidUTF8        = errors.New(",
		"\tErrDataTooLarge       = errors.New(",
		"\tErrArrayTooLarge      = errors.New(",
		"\tErrTooManyElements    = errors.New(",
	}

	for _, pattern := range expectedPatterns {
		if !strings.Contains(result, pattern) {
			t.Errorf("missing expected pattern: %q", pattern)
		}
	}
}

// TestGenerateErrorsAlignment verifies that error names are aligned
func TestGenerateErrorsAlignment(t *testing.T) {
	result := GenerateErrors()

	// All error names should have consistent spacing for alignment
	// The longest name is ErrTooManyElements (18 chars), so others should pad to align '='
	lines := strings.Split(result, "\n")

	errorLines := []string{}
	for _, line := range lines {
		if strings.Contains(line, "Err") && strings.Contains(line, "errors.New") {
			errorLines = append(errorLines, line)
		}
	}

	if len(errorLines) != 6 {
		t.Fatalf("expected 6 error declaration lines, got %d", len(errorLines))
	}

	// Check that all '=' are at similar positions (allowing some variation for alignment)
	firstEqualPos := strings.Index(errorLines[0], "=")
	if firstEqualPos == -1 {
		t.Fatal("no '=' found in first error line")
	}

	for i, line := range errorLines {
		equalPos := strings.Index(line, "=")
		if equalPos == -1 {
			t.Errorf("line %d: no '=' found", i)
			continue
		}
		// All '=' should be at the same position for proper alignment
		if equalPos != firstEqualPos {
			t.Errorf("line %d: '=' at position %d, expected %d for alignment", i, equalPos, firstEqualPos)
		}
	}
}

// TestGenerateErrorsOrder verifies the errors are in the expected order
func TestGenerateErrorsOrder(t *testing.T) {
	result := GenerateErrors()

	expectedOrder := []string{
		"ErrUnexpectedEOF",
		"ErrInvalidUTF8",
		"ErrDataTooLarge",
		"ErrArrayTooLarge",
		"ErrTooManyElements",
	}

	lastIndex := -1
	for i, errName := range expectedOrder {
		index := strings.Index(result, errName)
		if index == -1 {
			t.Errorf("error %d (%s) not found", i, errName)
			continue
		}
		if index <= lastIndex {
			t.Errorf("error %d (%s) not in correct order (index %d <= %d)", i, errName, index, lastIndex)
		}
		lastIndex = index
	}
}

// TestGenerateErrorsNoExtraContent verifies there's no unexpected content
func TestGenerateErrorsNoExtraContent(t *testing.T) {
	result := GenerateErrors()

	// Should not contain package declaration (that's in the file header)
	if strings.Contains(result, "package ") {
		t.Error("should not contain package declaration")
	}

	// Should not contain import statements (those are separate)
	if strings.Contains(result, "import ") {
		t.Error("should not contain import statements")
	}

	// Should have exactly 6 error variable declarations (5 from original + ErrInvalidData for RC optional fields)
	errorCount := strings.Count(result, "errors.New(")
	if errorCount != 6 {
		t.Errorf("expected 6 errors.New() calls, got %d", errorCount)
	}
}

// TestGenerateErrorsMatchesDesignSpec verifies errors match DESIGN_SPEC.md Section 5.4
func TestGenerateErrorsMatchesDesignSpec(t *testing.T) {
	result := GenerateErrors()

	// These exact strings are from DESIGN_SPEC.md Section 5.4
	designSpecErrors := []string{
		`ErrUnexpectedEOF      = errors.New("unexpected end of data")`,
		`ErrInvalidUTF8        = errors.New("invalid UTF-8 string")`,
		`ErrDataTooLarge       = errors.New("data exceeds 128MB limit")`,
		`ErrArrayTooLarge      = errors.New("array count exceeds per-array limit")`,
		`ErrTooManyElements    = errors.New("total elements exceed limit")`,
	}

	for _, expected := range designSpecErrors {
		if !strings.Contains(result, expected) {
			t.Errorf("missing exact error from design spec: %q", expected)
		}
	}
}
