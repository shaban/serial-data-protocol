package validator

import (
	"strings"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// DetectCycles finds circular references in the schema's type graph.
// Returns all cycles found (does not stop at first cycle).
//
// Circular references are not allowed because:
// - They make serialization impossible (infinite recursion)
// - They make size calculation impossible
// - Wire format has no pointer/reference support
//
// Examples of cycles:
//   - Direct: struct Node { next: Node }
//   - Indirect: struct A { b: B } struct B { a: A }
//   - Multi-hop: struct A { b: B } struct B { c: C } struct C { a: A }
func DetectCycles(schema *parser.Schema) []error {
	var errors []error

	// Build adjacency list: struct name -> list of referenced struct names
	graph := make(map[string][]string)
	for _, s := range schema.Structs {
		graph[s.Name] = extractStructReferences(s.Fields)
	}

	// Check each struct as a potential cycle start
	visited := make(map[string]bool)
	recStack := make(map[string]bool) // Recursion stack for cycle detection
	path := []string{}                 // Current path for error reporting

	for _, s := range schema.Structs {
		if !visited[s.Name] {
			if cycle := findCycle(s.Name, graph, visited, recStack, path); cycle != nil {
				cyclePath := strings.Join(cycle, " → ")
				errors = append(errors, errCircularReference(cyclePath))
			}
		}
	}

	return errors
}

// extractStructReferences returns all struct names referenced by fields (not primitives).
func extractStructReferences(fields []parser.Field) []string {
	var refs []string
	seen := make(map[string]bool)

	for _, field := range fields {
		collectStructRefs(&field.Type, seen)
	}

	for name := range seen {
		refs = append(refs, name)
	}

	return refs
}

// collectStructRefs recursively collects struct names from a type expression.
func collectStructRefs(typeExpr *parser.TypeExpr, seen map[string]bool) {
	switch typeExpr.Kind {
	case parser.TypeKindPrimitive:
		// Primitives don't create dependencies
		return

	case parser.TypeKindNamed:
		// Named type is a struct reference
		seen[typeExpr.Name] = true

	case parser.TypeKindArray:
		// Recurse into array element type
		if typeExpr.Elem != nil {
			collectStructRefs(typeExpr.Elem, seen)
		}
	}
}

// findCycle performs DFS to detect cycles starting from the given node.
// Returns the cycle path if found, nil otherwise.
func findCycle(node string, graph map[string][]string, visited, recStack map[string]bool, path []string) []string {
	visited[node] = true
	recStack[node] = true
	path = append(path, node)

	// Check all neighbors
	for _, neighbor := range graph[node] {
		if !visited[neighbor] {
			// Recurse to unvisited neighbor
			if cycle := findCycle(neighbor, graph, visited, recStack, path); cycle != nil {
				return cycle
			}
		} else if recStack[neighbor] {
			// Found a cycle! Build the cycle path
			cycleStart := -1
			for i, n := range path {
				if n == neighbor {
					cycleStart = i
					break
				}
			}
			if cycleStart >= 0 {
				cyclePath := append(path[cycleStart:], neighbor) // Close the cycle
				return cyclePath
			}
		}
	}

	// Backtrack: remove from recursion stack
	recStack[node] = false
	return nil
}
