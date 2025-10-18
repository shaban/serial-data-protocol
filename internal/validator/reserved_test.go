package validator

import (
	"testing"
)

func TestIsReserved_Go(t *testing.T) {
	goKeywords := []string{
		"break", "default", "func", "interface", "select",
		"case", "defer", "go", "map", "struct",
		"chan", "else", "goto", "package", "switch",
		"const", "fallthrough", "if", "range", "type",
		"continue", "for", "import", "return", "var",
	}

	for _, kw := range goKeywords {
		if !IsReserved(kw) {
			t.Errorf("Expected %q to be reserved (Go keyword)", kw)
		}

		langs := GetReservedLanguages(kw)
		if len(langs) == 0 {
			t.Errorf("Expected %q to have languages, got none", kw)
		}

		// Check that Go is in the list
		found := false
		for _, lang := range langs {
			if lang == "Go" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %q to be reserved in Go, got languages: %v", kw, langs)
		}
	}
}

func TestIsReserved_Rust(t *testing.T) {
	rustKeywords := []string{
		"as", "async", "await", "break", "const",
		"continue", "crate", "dyn", "else", "enum",
		"extern", "false", "fn", "for", "if",
		"impl", "in", "let", "loop", "match",
		"mod", "move", "mut", "pub", "ref",
		"return", "self", "Self", "static", "struct",
		"super", "trait", "true", "type", "unsafe",
		"use", "where", "while", "abstract", "become",
		"box", "do", "final", "macro", "override",
		"priv", "typeof", "unsized", "virtual", "yield",
		"try",
	}

	for _, kw := range rustKeywords {
		if !IsReserved(kw) {
			t.Errorf("Expected %q to be reserved (Rust keyword)", kw)
		}

		langs := GetReservedLanguages(kw)
		found := false
		for _, lang := range langs {
			if lang == "Rust" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %q to be reserved in Rust, got languages: %v", kw, langs)
		}
	}
}

func TestIsReserved_C(t *testing.T) {
	cKeywords := []string{
		"auto", "break", "case", "char", "const",
		"continue", "default", "do", "double", "else",
		"enum", "extern", "float", "for", "goto",
		"if", "inline", "int", "long", "register",
		"restrict", "return", "short", "signed", "sizeof",
		"static", "struct", "switch", "typedef", "union",
		"unsigned", "void", "volatile", "while", "_Alignas",
		"_Alignof", "_Atomic", "_Bool", "_Complex", "_Generic",
		"_Imaginary", "_Noreturn", "_Static_assert", "_Thread_local",
	}

	for _, kw := range cKeywords {
		if !IsReserved(kw) {
			t.Errorf("Expected %q to be reserved (C keyword)", kw)
		}

		langs := GetReservedLanguages(kw)
		found := false
		for _, lang := range langs {
			if lang == "C" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %q to be reserved in C, got languages: %v", kw, langs)
		}
	}
}

func TestIsReserved_Swift(t *testing.T) {
	// Test a sample of Swift keywords (testing all 87 would be verbose)
	swiftKeywords := []string{
		"associatedtype", "class", "deinit", "enum",
		"guard", "protocol", "subscript", "typealias",
		"inout", "fileprivate", "convenience", "dynamic",
		"lazy", "mutating", "nonmutating", "optional",
		"override", "required", "unowned", "weak",
	}

	for _, kw := range swiftKeywords {
		if !IsReserved(kw) {
			t.Errorf("Expected %q to be reserved (Swift keyword)", kw)
		}

		langs := GetReservedLanguages(kw)
		found := false
		for _, lang := range langs {
			if lang == "Swift" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %q to be reserved in Swift, got languages: %v", kw, langs)
		}
	}
}

func TestIsReserved_NotReserved(t *testing.T) {
	notReserved := []string{
		"device", "plugin", "parameter", "config",
		"name", "value", "enabled", "status",
		"Device", "Plugin", "Parameter", "Config",
	}

	for _, word := range notReserved {
		if IsReserved(word) {
			langs := GetReservedLanguages(word)
			t.Errorf("Expected %q to not be reserved, but found in: %v", word, langs)
		}

		langs := GetReservedLanguages(word)
		if langs != nil {
			t.Errorf("Expected %q to have nil languages, got: %v", word, langs)
		}
	}
}

func TestIsReserved_CaseInsensitive(t *testing.T) {
	testCases := []struct {
		word     string
		expected bool
	}{
		{"type", true},     // lowercase
		{"Type", true},     // capitalized
		{"TYPE", true},     // uppercase
		{"TyPe", true},     // mixed case
		{"struct", true},   // lowercase
		{"Struct", true},   // capitalized
		{"STRUCT", true},   // uppercase
		{"async", true},    // Rust keyword
		{"Async", true},    // capitalized
		{"ASYNC", true},    // uppercase
		{"device", false},  // not reserved
		{"Device", false},  // not reserved
		{"DEVICE", false},  // not reserved
	}

	for _, tc := range testCases {
		result := IsReserved(tc.word)
		if result != tc.expected {
			t.Errorf("IsReserved(%q) = %v, want %v", tc.word, result, tc.expected)
		}
	}
}

func TestGetReservedLanguages_MultipleLanguages(t *testing.T) {
	// Test keywords reserved in multiple languages
	testCases := []struct {
		word              string
		expectedLanguages []string
	}{
		{"break", []string{"Go", "Rust", "C", "Swift"}},     // Reserved in all 4
		{"struct", []string{"Go", "Rust", "C", "Swift"}},    // Reserved in all 4
		{"return", []string{"Go", "Rust", "C", "Swift"}},    // Reserved in all 4
		{"const", []string{"Go", "Rust", "C"}},              // Not in Swift (let/var instead)
		{"async", []string{"Rust", "Swift"}},                // Rust and Swift only
		{"inline", []string{"C"}},                           // C only
		{"protocol", []string{"Swift"}},                     // Swift only
		{"impl", []string{"Rust"}},                          // Rust only
	}

	for _, tc := range testCases {
		langs := GetReservedLanguages(tc.word)
		if len(langs) != len(tc.expectedLanguages) {
			t.Errorf("GetReservedLanguages(%q): expected %d languages %v, got %d: %v",
				tc.word, len(tc.expectedLanguages), tc.expectedLanguages, len(langs), langs)
			continue
		}

		// Check all expected languages are present
		for _, expected := range tc.expectedLanguages {
			found := false
			for _, lang := range langs {
				if lang == expected {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("GetReservedLanguages(%q): expected language %q not found in %v",
					tc.word, expected, langs)
			}
		}
	}
}

func TestReservedKeywords_SampleCheck(t *testing.T) {
	// Verify a sample of keywords from each language to ensure lists are working
	testCases := []struct {
		word string
		lang string
	}{
		// Go samples
		{"type", "Go"},
		{"string", "Go"},
		{"make", "Go"},
		// Rust samples
		{"impl", "Rust"},
		{"Option", "Rust"},
		{"async", "Rust"},
		// C samples
		{"typedef", "C"},
		{"uint32_t", "C"},
		{"_Atomic", "C"},
		// Swift samples
		{"protocol", "Swift"},
		{"guard", "Swift"},
		{"available", "Swift"}, // Attribute name (without @)
	}

	for _, tc := range testCases {
		if !IsReserved(tc.word) {
			t.Errorf("Expected %q to be reserved (sample from %s)", tc.word, tc.lang)
			continue
		}

		langs := GetReservedLanguages(tc.word)
		found := false
		for _, lang := range langs {
			if lang == tc.lang {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %q to be reserved in %s, got languages: %v", tc.word, tc.lang, langs)
		}
	}
}
