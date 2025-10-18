package validator

import (
	"strings"
	"testing"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

func TestNoCycle(t *testing.T) {
	testCases := []struct {
		name  string
		input string
	}{
		{
			name: "single struct",
			input: `
			struct Device {
				id: u32,
				name: str,
			}
			`,
		},
		{
			name: "two independent structs",
			input: `
			struct Device {
				id: u32,
			}
			struct Config {
				enabled: bool,
			}
			`,
		},
		{
			name: "linear chain",
			input: `
			struct A {
				b: B,
			}
			struct B {
				c: C,
			}
			struct C {
				value: u32,
			}
			`,
		},
		{
			name: "tree structure",
			input: `
			struct Root {
				left: Branch,
				right: Branch,
			}
			struct Branch {
				value: u32,
			}
			`,
		},
		{
			name: "array doesn't create cycle",
			input: `
			struct List {
				items: []Item,
			}
			struct Item {
				value: u32,
			}
			`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := parser.ParseSchema(tc.input)
			if err != nil {
				t.Fatalf("ParseSchema failed: %v", err)
			}

			errors := DetectCycles(schema)
			if len(errors) != 0 {
				t.Errorf("Expected no cycles, got: %v", errors)
			}
		})
	}
}

func TestDirectSelfReference(t *testing.T) {
	input := `
	struct Node {
		next: Node,
	}
	`

	schema, err := parser.ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	errors := DetectCycles(schema)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d: %v", len(errors), errors)
	}

	errMsg := errors[0].Error()
	if !strings.Contains(errMsg, "circular reference") {
		t.Errorf("Error should mention 'circular reference', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "Node") {
		t.Errorf("Error should mention 'Node', got: %s", errMsg)
	}
}

func TestIndirectCycle(t *testing.T) {
	testCases := []struct {
		name          string
		input         string
		expectedNodes []string // Nodes that should appear in cycle path
	}{
		{
			name: "two-node cycle",
			input: `
			struct A {
				b: B,
			}
			struct B {
				a: A,
			}
			`,
			expectedNodes: []string{"A", "B"},
		},
		{
			name: "three-node cycle",
			input: `
			struct A {
				b: B,
			}
			struct B {
				c: C,
			}
			struct C {
				a: A,
			}
			`,
			expectedNodes: []string{"A", "B", "C"},
		},
		{
			name: "cycle with extra fields",
			input: `
			struct A {
				id: u32,
				b: B,
				name: str,
			}
			struct B {
				value: f64,
				a: A,
			}
			`,
			expectedNodes: []string{"A", "B"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			schema, err := parser.ParseSchema(tc.input)
			if err != nil {
				t.Fatalf("ParseSchema failed: %v", err)
			}

			errors := DetectCycles(schema)
			if len(errors) != 1 {
				t.Fatalf("Expected 1 error, got %d: %v", len(errors), errors)
			}

			errMsg := errors[0].Error()
			if !strings.Contains(errMsg, "circular reference") {
				t.Errorf("Error should mention 'circular reference', got: %s", errMsg)
			}

			// Check all expected nodes are in the cycle path
			for _, node := range tc.expectedNodes {
				if !strings.Contains(errMsg, node) {
					t.Errorf("Error should mention node %q, got: %s", node, errMsg)
				}
			}
		})
	}
}

func TestMultipleCycles(t *testing.T) {
	input := `
	struct A {
		a: A,
	}
	struct B {
		c: C,
	}
	struct C {
		b: B,
	}
	`

	schema, err := parser.ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	errors := DetectCycles(schema)
	
	// Should detect at least 2 cycles: A→A and B→C→B
	if len(errors) < 2 {
		t.Errorf("Expected at least 2 cycles, got %d: %v", len(errors), errors)
	}

	// Check that both cycle groups are mentioned
	allErrors := ""
	for _, e := range errors {
		allErrors += e.Error() + " "
	}

	if !strings.Contains(allErrors, "A") {
		t.Errorf("Expected errors to mention cycle involving A, got: %s", allErrors)
	}
	if !(strings.Contains(allErrors, "B") && strings.Contains(allErrors, "C")) {
		t.Errorf("Expected errors to mention cycle involving B and C, got: %s", allErrors)
	}
}

func TestCycleViaArray(t *testing.T) {
	// Cycles through arrays are still cycles
	input := `
	struct Node {
		children: []Node,
	}
	`

	schema, err := parser.ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	errors := DetectCycles(schema)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error for cycle via array, got %d: %v", len(errors), errors)
	}

	errMsg := errors[0].Error()
	if !strings.Contains(errMsg, "Node") {
		t.Errorf("Error should mention 'Node', got: %s", errMsg)
	}
}

func TestComplexCycle(t *testing.T) {
	// More complex scenario with multiple paths
	input := `
	struct Root {
		child: Child,
	}
	struct Child {
		parent: Parent,
	}
	struct Parent {
		root: Root,
	}
	`

	schema, err := parser.ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	errors := DetectCycles(schema)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d: %v", len(errors), errors)
	}

	errMsg := errors[0].Error()
	// Should mention all three structs in the cycle
	for _, name := range []string{"Root", "Child", "Parent"} {
		if !strings.Contains(errMsg, name) {
			t.Errorf("Error should mention %q, got: %s", name, errMsg)
		}
	}
}

func TestNoCycleWithSharedDependency(t *testing.T) {
	// Diamond pattern - not a cycle
	input := `
	struct Root {
		left: Branch,
		right: Branch,
	}
	struct Branch {
		leaf: Leaf,
	}
	struct Leaf {
		value: u32,
	}
	`

	schema, err := parser.ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	errors := DetectCycles(schema)
	if len(errors) != 0 {
		t.Errorf("Expected no cycles in diamond pattern, got: %v", errors)
	}
}

func TestEmptySchema(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{},
	}

	errors := DetectCycles(schema)
	if len(errors) != 0 {
		t.Errorf("Expected no errors for empty schema, got: %v", errors)
	}
}

func TestCycleWithMixedFields(t *testing.T) {
	// Cycle where structs have mix of primitive and struct fields
	input := `
	struct A {
		id: u32,
		name: str,
		b: B,
		count: i32,
	}
	struct B {
		enabled: bool,
		items: []u32,
		a: A,
		tag: str,
	}
	`

	schema, err := parser.ParseSchema(input)
	if err != nil {
		t.Fatalf("ParseSchema failed: %v", err)
	}

	errors := DetectCycles(schema)
	if len(errors) != 1 {
		t.Fatalf("Expected 1 error, got %d: %v", len(errors), errors)
	}

	errMsg := errors[0].Error()
	if !strings.Contains(errMsg, "A") || !strings.Contains(errMsg, "B") {
		t.Errorf("Error should mention both A and B, got: %s", errMsg)
	}
}
