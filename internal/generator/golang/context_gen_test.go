package golang

import (
	"strings"
	"testing"
)

// TestGenerateDecodeContext verifies basic generation
func TestGenerateDecodeContext(t *testing.T) {
	result := GenerateDecodeContext()

	if result == "" {
		t.Fatal("GenerateDecodeContext returned empty string")
	}

	// Check for main components
	if !strings.Contains(result, "DecodeContext") {
		t.Error("missing DecodeContext type")
	}

	if !strings.Contains(result, "checkArraySize") {
		t.Error("missing checkArraySize method")
	}
}

// TestGenerateDecodeContextConstants verifies all three constants are present
func TestGenerateDecodeContextConstants(t *testing.T) {
	result := GenerateDecodeContext()

	expectedConstants := []string{
		"MaxSerializedSize",
		"MaxArrayElements",
		"MaxTotalElements",
	}

	for _, constant := range expectedConstants {
		if !strings.Contains(result, constant) {
			t.Errorf("missing constant: %s", constant)
		}
	}
}

// TestGenerateDecodeContextConstantValues verifies constant values match design spec
func TestGenerateDecodeContextConstantValues(t *testing.T) {
	result := GenerateDecodeContext()

	expectedValues := []struct {
		name  string
		value string
	}{
		{"MaxSerializedSize", "128 * 1024 * 1024"},
		{"MaxArrayElements", "1_000_000"},
		{"MaxTotalElements", "10_000_000"},
	}

	for _, ev := range expectedValues {
		// Check that constant name exists
		if !strings.Contains(result, ev.name) {
			t.Errorf("missing constant: %s", ev.name)
			continue
		}
		// Check that value exists (allowing for alignment spacing)
		if !strings.Contains(result, ev.value) {
			t.Errorf("constant %s should have value %s", ev.name, ev.value)
		}
	}
}

// TestGenerateDecodeContextTypeStructure verifies the struct definition
func TestGenerateDecodeContextTypeStructure(t *testing.T) {
	result := GenerateDecodeContext()

	// Check for type declaration
	if !strings.Contains(result, "type DecodeContext struct {") {
		t.Error("missing type declaration")
	}

	// Check for totalElements field
	if !strings.Contains(result, "totalElements int") {
		t.Error("missing totalElements field")
	}

	// Check for closing brace
	lines := strings.Split(result, "\n")
	foundStruct := false
	foundField := false
	foundClosing := false

	for _, line := range lines {
		if strings.Contains(line, "type DecodeContext struct {") {
			foundStruct = true
		} else if foundStruct && strings.Contains(line, "totalElements") {
			foundField = true
		} else if foundStruct && foundField && strings.TrimSpace(line) == "}" {
			foundClosing = true
			break
		}
	}

	if !foundStruct || !foundField || !foundClosing {
		t.Error("struct definition not properly formatted")
	}
}

// TestGenerateDecodeContextMethodSignature verifies checkArraySize signature
func TestGenerateDecodeContextMethodSignature(t *testing.T) {
	result := GenerateDecodeContext()

	// Check method signature
	expectedSignature := "func (ctx *DecodeContext) checkArraySize(count uint32) error {"
	if !strings.Contains(result, expectedSignature) {
		t.Errorf("missing or incorrect method signature, expected: %s", expectedSignature)
	}
}

// TestGenerateDecodeContextMethodLogic verifies checkArraySize implementation
func TestGenerateDecodeContextMethodLogic(t *testing.T) {
	result := GenerateDecodeContext()

	// Check for per-array limit check
	if !strings.Contains(result, "if count > MaxArrayElements {") {
		t.Error("missing per-array limit check")
	}

	if !strings.Contains(result, "return ErrArrayTooLarge") {
		t.Error("missing ErrArrayTooLarge return")
	}

	// Check for total elements accumulation
	if !strings.Contains(result, "ctx.totalElements += int(count)") {
		t.Error("missing totalElements accumulation")
	}

	// Check for total limit check
	if !strings.Contains(result, "if ctx.totalElements > MaxTotalElements {") {
		t.Error("missing total elements limit check")
	}

	if !strings.Contains(result, "return ErrTooManyElements") {
		t.Error("missing ErrTooManyElements return")
	}

	// Check for success return
	if !strings.Contains(result, "return nil") {
		t.Error("missing nil return for success case")
	}
}

// TestGenerateDecodeContextComments verifies documentation comments
func TestGenerateDecodeContextComments(t *testing.T) {
	result := GenerateDecodeContext()

	expectedComments := []string{
		"// Size limit constants",
		"// DecodeContext tracks state",
		"// checkArraySize validates",
	}

	for _, comment := range expectedComments {
		if !strings.Contains(result, comment) {
			t.Errorf("missing expected comment starting with: %q", comment)
		}
	}
}

// TestGenerateDecodeContextConstantBlock verifies const block format
func TestGenerateDecodeContextConstantBlock(t *testing.T) {
	result := GenerateDecodeContext()

	// Check for const block
	if !strings.Contains(result, "const (") {
		t.Error("missing 'const (' block")
	}

	// Verify all constants are in the block
	constBlockStart := strings.Index(result, "const (")
	constBlockEnd := strings.Index(result[constBlockStart:], "\n)")
	if constBlockEnd == -1 {
		t.Fatal("const block not properly closed")
	}

	constBlock := result[constBlockStart : constBlockStart+constBlockEnd]

	if !strings.Contains(constBlock, "MaxSerializedSize") {
		t.Error("MaxSerializedSize not in const block")
	}
	if !strings.Contains(constBlock, "MaxArrayElements") {
		t.Error("MaxArrayElements not in const block")
	}
	if !strings.Contains(constBlock, "MaxTotalElements") {
		t.Error("MaxTotalElements not in const block")
	}
}

// TestGenerateDecodeContextMatchesDesignSpec verifies exact match with DESIGN_SPEC.md
func TestGenerateDecodeContextMatchesDesignSpec(t *testing.T) {
	result := GenerateDecodeContext()

	// These exact patterns are from DESIGN_SPEC.md Section 5.5
	designSpecPatterns := []string{
		"MaxSerializedSize = 128 * 1024 * 1024",
		"MaxArrayElements  = 1_000_000",
		"MaxTotalElements  = 10_000_000",
		"type DecodeContext struct {",
		"totalElements int",
		"func (ctx *DecodeContext) checkArraySize(count uint32) error {",
		"if count > MaxArrayElements {",
		"return ErrArrayTooLarge",
		"ctx.totalElements += int(count)",
		"if ctx.totalElements > MaxTotalElements {",
		"return ErrTooManyElements",
	}

	for _, pattern := range designSpecPatterns {
		if !strings.Contains(result, pattern) {
			t.Errorf("missing pattern from design spec: %q", pattern)
		}
	}
}

// TestGenerateDecodeContextOrder verifies logical ordering of components
func TestGenerateDecodeContextOrder(t *testing.T) {
	result := GenerateDecodeContext()

	// Constants should come first
	constIndex := strings.Index(result, "const (")
	typeIndex := strings.Index(result, "type DecodeContext")
	methodIndex := strings.Index(result, "func (ctx *DecodeContext) checkArraySize")

	if constIndex == -1 || typeIndex == -1 || methodIndex == -1 {
		t.Fatal("missing required components")
	}

	if constIndex >= typeIndex {
		t.Error("constants should come before type definition")
	}

	if typeIndex >= methodIndex {
		t.Error("type definition should come before method")
	}
}

// TestGenerateDecodeContextNoExtraContent verifies clean output
func TestGenerateDecodeContextNoExtraContent(t *testing.T) {
	result := GenerateDecodeContext()

	// Should not contain package declaration
	if strings.Contains(result, "package ") {
		t.Error("should not contain package declaration")
	}

	// Should not contain import statements
	if strings.Contains(result, "import ") {
		t.Error("should not contain import statements")
	}

	// Should have exactly one type definition
	typeCount := strings.Count(result, "type DecodeContext struct")
	if typeCount != 1 {
		t.Errorf("expected 1 type definition, got %d", typeCount)
	}

	// Should have exactly one method
	methodCount := strings.Count(result, "func (ctx *DecodeContext)")
	if methodCount != 1 {
		t.Errorf("expected 1 method, got %d", methodCount)
	}
}

// TestGenerateDecodeContextMethodReturnPaths verifies all return paths in checkArraySize
func TestGenerateDecodeContextMethodReturnPaths(t *testing.T) {
	result := GenerateDecodeContext()

	// Extract just the checkArraySize method
	methodStart := strings.Index(result, "func (ctx *DecodeContext) checkArraySize")
	if methodStart == -1 {
		t.Fatal("method not found")
	}

	// Find the closing brace of the method (simplified approach)
	methodBody := result[methodStart:]

	// Should have three return statements
	returnCount := strings.Count(methodBody, "return ")
	if returnCount < 3 {
		t.Errorf("expected at least 3 return statements in checkArraySize, got %d", returnCount)
	}

	// Verify specific returns exist
	if !strings.Contains(methodBody, "return ErrArrayTooLarge") {
		t.Error("missing 'return ErrArrayTooLarge'")
	}
	if !strings.Contains(methodBody, "return ErrTooManyElements") {
		t.Error("missing 'return ErrTooManyElements'")
	}
	if !strings.Contains(methodBody, "return nil") {
		t.Error("missing 'return nil'")
	}
}

// TestGenerateDecodeContextAlignment verifies constant alignment
func TestGenerateDecodeContextAlignment(t *testing.T) {
	result := GenerateDecodeContext()

	// Extract const block
	constStart := strings.Index(result, "const (")
	constEnd := strings.Index(result[constStart:], "\n)")
	if constStart == -1 || constEnd == -1 {
		t.Fatal("const block not found")
	}

	constBlock := result[constStart : constStart+constEnd]
	lines := strings.Split(constBlock, "\n")

	equalPositions := []int{}
	for _, line := range lines {
		if strings.Contains(line, "=") {
			equalPos := strings.Index(line, "=")
			equalPositions = append(equalPositions, equalPos)
		}
	}

	if len(equalPositions) != 3 {
		t.Fatalf("expected 3 constant declarations, got %d", len(equalPositions))
	}

	// All '=' should be at the same position for alignment
	firstPos := equalPositions[0]
	for i, pos := range equalPositions {
		if pos != firstPos {
			t.Errorf("constant %d: '=' at position %d, expected %d for alignment", i, pos, firstPos)
		}
	}
}
