package validator

import (
	"strings"
)

// Reserved keywords and problematic identifiers from target languages.
// Includes: language keywords, future reserved words, and common identifiers
// that would cause compilation errors, warnings, or ambiguous code.
// Source documentation in DESIGN_SPEC.md Section 3.5.1

var (
	// Go reserved keywords (25 keywords) + common problematic identifiers
	goKeywords = []string{
		// Language keywords
		"break", "default", "func", "interface", "select",
		"case", "defer", "go", "map", "struct",
		"chan", "else", "goto", "package", "switch",
		"const", "fallthrough", "if", "range", "type",
		"continue", "for", "import", "return", "var",
		// Built-in types and functions (would shadow built-ins)
		"bool", "byte", "complex64", "complex128", "error",
		"float32", "float64", "int", "int8", "int16",
		"int32", "int64", "rune", "string", "uint",
		"uint8", "uint16", "uint32", "uint64", "uintptr",
		"true", "false", "iota", "nil",
		"append", "cap", "close", "complex", "copy",
		"delete", "imag", "len", "make", "new",
		"panic", "print", "println", "real", "recover",
		// Common package names that shouldn't be shadowed
		"main", "init",
	}

	// Rust reserved keywords (strict + reserved for future use)
	rustKeywords = []string{
		// Strict keywords (cannot be used as identifiers)
		"as", "break", "const", "continue", "crate",
		"else", "enum", "extern", "false", "fn",
		"for", "if", "impl", "in", "let",
		"loop", "match", "mod", "move", "mut",
		"pub", "ref", "return", "self", "Self",
		"static", "struct", "super", "trait", "true",
		"type", "unsafe", "use", "where", "while",
		// Keywords reserved for future use
		"abstract", "async", "await", "become", "box",
		"do", "final", "macro", "override", "priv",
		"try", "typeof", "unsized", "virtual", "yield",
		// Weak keywords (contextual, but best avoided)
		"union", "dyn", "raw",
		// Common types and traits that shouldn't be shadowed
		"Option", "Result", "Some", "None", "Ok", "Err",
		"String", "Vec", "Box", "Rc", "Arc",
		"Copy", "Clone", "Send", "Sync", "Sized",
	}

	// C reserved keywords (C11 standard + C23 additions)
	cKeywords = []string{
		// C11 keywords
		"auto", "break", "case", "char", "const",
		"continue", "default", "do", "double", "else",
		"enum", "extern", "float", "for", "goto",
		"if", "inline", "int", "long", "register",
		"restrict", "return", "short", "signed", "sizeof",
		"static", "struct", "switch", "typedef", "union",
		"unsigned", "void", "volatile", "while",
		// C11 additions
		"_Alignas", "_Alignof", "_Atomic", "_Bool", "_Complex",
		"_Generic", "_Imaginary", "_Noreturn", "_Static_assert", "_Thread_local",
		// C23 additions
		"_BitInt", "_Decimal128", "_Decimal32", "_Decimal64",
		// Common standard library that shouldn't be shadowed
		"bool", "true", "false", "NULL",
		"size_t", "ptrdiff_t", "wchar_t",
		"int8_t", "int16_t", "int32_t", "int64_t",
		"uint8_t", "uint16_t", "uint32_t", "uint64_t",
		"FILE", "EOF",
	}

	// Swift reserved keywords (keywords + attributes + common types)
	swiftKeywords = []string{
		// Declaration keywords
		"associatedtype", "class", "deinit", "enum",
		"extension", "fileprivate", "func", "import",
		"init", "inout", "internal", "let",
		"open", "operator", "private", "precedencegroup",
		"protocol", "public", "rethrows", "static",
		"struct", "subscript", "typealias", "var",
		// Statement keywords
		"break", "case", "catch", "continue",
		"default", "defer", "do", "else",
		"fallthrough", "for", "guard", "if",
		"in", "repeat", "return", "switch",
		"throw", "where", "while",
		// Expression keywords
		"as", "false", "is", "nil", "self",
		"Self", "super", "throws", "true", "try",
		// Context-sensitive keywords
		"async", "await", "didSet", "get", "set",
		"willSet", "weak", "unowned",
		// Pattern keywords
		"_",
		// Special identifiers
		"Any", "Type", "Protocol",
		// Compiler control statements (pound keywords)
		"#available", "#colorLiteral", "#column", "#else",
		"#elseif", "#endif", "#error", "#file",
		"#fileLiteral", "#fileID", "#filePath", "#function",
		"#if", "#imageLiteral", "#line", "#selector",
		"#sourceLocation", "#warning",
		// Attributes (without @ prefix - we check the name part)
		"available", "objc", "nonobjc", "discardableResult",
		"dynamicCallable", "dynamicMemberLookup", "escaping",
		"autoclosure", "convention", "IBAction", "IBOutlet",
		"IBDesignable", "IBInspectable", "NSCopying", "NSManaged",
		"UIApplicationMain", "NSApplicationMain", "testable",
		"warn_unqualified_access", "frozen", "unknown",
		// Declaration modifiers
		"dynamic", "final", "lazy", "optional", "required",
		"convenience", "override", "mutating", "nonmutating",
		// Access control
		"open", "public", "internal", "fileprivate", "private",
		// Common standard library types
		"Int", "Int8", "Int16", "Int32", "Int64",
		"UInt", "UInt8", "UInt16", "UInt32", "UInt64",
		"Float", "Double", "Bool", "String", "Character",
		"Array", "Dictionary", "Set", "Optional",
		"Error", "Result",
	}

	// reservedMap maps lowercase keywords to the languages that reserve them
	reservedMap map[string][]string
)

func init() {
	// Build the reserved keyword map (case-insensitive)
	reservedMap = make(map[string][]string)

	addKeywords := func(keywords []string, language string) {
		for _, kw := range keywords {
			lower := strings.ToLower(kw)
			// Avoid duplicates in the same language
			found := false
			for _, existingLang := range reservedMap[lower] {
				if existingLang == language {
					found = true
					break
				}
			}
			if !found {
				reservedMap[lower] = append(reservedMap[lower], language)
			}
		}
	}

	addKeywords(goKeywords, "Go")
	addKeywords(rustKeywords, "Rust")
	addKeywords(cKeywords, "C")
	addKeywords(swiftKeywords, "Swift")
}

// IsReserved checks if a word is reserved in any target language.
// Comparison is case-insensitive.
func IsReserved(word string) bool {
	lower := strings.ToLower(word)
	_, found := reservedMap[lower]
	return found
}

// GetReservedLanguages returns the list of languages that reserve the given word.
// Returns nil if the word is not reserved. Comparison is case-insensitive.
func GetReservedLanguages(word string) []string {
	lower := strings.ToLower(word)
	return reservedMap[lower]
}
