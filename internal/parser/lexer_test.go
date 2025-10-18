package parser

import (
	"testing"
)

func TestLexStruct(t *testing.T) {
	input := `struct Device {
		id: u32,
		name: str,
	}`
	
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}
	
	expected := []TokenType{
		TokenStruct,   // struct
		TokenIdent,    // Device
		TokenLBrace,   // {
		TokenIdent,    // id
		TokenColon,    // :
		TokenIdent,    // u32
		TokenComma,    // ,
		TokenIdent,    // name
		TokenColon,    // :
		TokenIdent,    // str
		TokenComma,    // ,
		TokenRBrace,   // }
		TokenEOF,
	}
	
	if len(tokens) != len(expected) {
		t.Fatalf("Expected %d tokens, got %d", len(expected), len(tokens))
	}
	
	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("Token %d: expected %v, got %v (%s)", i, expected[i], tok.Type, tok.String())
		}
	}
	
	// Verify specific values
	if tokens[1].Value != "Device" {
		t.Errorf("Expected struct name 'Device', got %q", tokens[1].Value)
	}
	if tokens[3].Value != "id" {
		t.Errorf("Expected field name 'id', got %q", tokens[3].Value)
	}
}

func TestLexDocComment(t *testing.T) {
	input := `/// This is a device.
struct Device {
	/// Device identifier.
	id: u32,
}`
	
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}
	
	// Find doc comments
	var docComments []string
	for _, tok := range tokens {
		if tok.Type == TokenDocComment {
			docComments = append(docComments, tok.Value)
		}
	}
	
	if len(docComments) != 2 {
		t.Errorf("Expected 2 doc comments, got %d", len(docComments))
	}
	
	if len(docComments) > 0 && docComments[0] != "This is a device." {
		t.Errorf("Expected doc comment 'This is a device.', got %q", docComments[0])
	}
	
	if len(docComments) > 1 && docComments[1] != "Device identifier." {
		t.Errorf("Expected doc comment 'Device identifier.', got %q", docComments[1])
	}
}

func TestLexArray(t *testing.T) {
	input := `plugins: []Plugin`
	
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}
	
	expected := []TokenType{
		TokenIdent,    // plugins
		TokenColon,    // :
		TokenLBracket, // [
		TokenRBracket, // ]
		TokenIdent,    // Plugin
		TokenEOF,
	}
	
	if len(tokens) != len(expected) {
		t.Fatalf("Expected %d tokens, got %d", len(expected), len(tokens))
	}
	
	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("Token %d: expected %v, got %v", i, expected[i], tok.Type)
		}
	}
}

func TestLexIdentifiers(t *testing.T) {
	testCases := []struct {
		input    string
		expected []string
	}{
		{
			input:    "id name value",
			expected: []string{"id", "name", "value"},
		},
		{
			input:    "my_field _private field123",
			expected: []string{"my_field", "_private", "field123"},
		},
		{
			input:    "u8 u16 u32 u64 i8 i16 i32 i64 f32 f64 bool str",
			expected: []string{"u8", "u16", "u32", "u64", "i8", "i16", "i32", "i64", "f32", "f64", "bool", "str"},
		},
	}
	
	for _, tc := range testCases {
		lexer := NewLexer(tc.input)
		tokens, err := lexer.Tokenize()
		if err != nil {
			t.Errorf("Tokenize failed for %q: %v", tc.input, err)
			continue
		}
		
		var idents []string
		for _, tok := range tokens {
			if tok.Type == TokenIdent {
				idents = append(idents, tok.Value)
			}
		}
		
		if len(idents) != len(tc.expected) {
			t.Errorf("Input %q: expected %d identifiers, got %d", tc.input, len(tc.expected), len(idents))
			continue
		}
		
		for i, ident := range idents {
			if ident != tc.expected[i] {
				t.Errorf("Input %q: identifier %d expected %q, got %q", tc.input, i, tc.expected[i], ident)
			}
		}
	}
}

func TestLexKeyword(t *testing.T) {
	input := "struct"
	
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}
	
	if len(tokens) != 2 { // struct + EOF
		t.Fatalf("Expected 2 tokens, got %d", len(tokens))
	}
	
	if tokens[0].Type != TokenStruct {
		t.Errorf("Expected TokenStruct, got %v", tokens[0].Type)
	}
}

func TestLexComments(t *testing.T) {
	input := `// Regular comment
/// Doc comment
struct Device {}`
	
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}
	
	var regularComments, docComments int
	for _, tok := range tokens {
		if tok.Type == TokenComment {
			regularComments++
		}
		if tok.Type == TokenDocComment {
			docComments++
		}
	}
	
	if regularComments != 1 {
		t.Errorf("Expected 1 regular comment, got %d", regularComments)
	}
	if docComments != 1 {
		t.Errorf("Expected 1 doc comment, got %d", docComments)
	}
}

func TestLexLineNumbers(t *testing.T) {
	input := `struct Device {
	id: u32,
	name: str,
}`
	
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}
	
	// Check that struct is on line 1
	if tokens[0].Type == TokenStruct && tokens[0].Line != 1 {
		t.Errorf("Expected 'struct' on line 1, got line %d", tokens[0].Line)
	}
	
	// Find 'id' token (should be on line 2)
	for _, tok := range tokens {
		if tok.Type == TokenIdent && tok.Value == "id" {
			if tok.Line != 2 {
				t.Errorf("Expected 'id' on line 2, got line %d", tok.Line)
			}
			break
		}
	}
	
	// Find 'name' token (should be on line 3)
	for _, tok := range tokens {
		if tok.Type == TokenIdent && tok.Value == "name" {
			if tok.Line != 3 {
				t.Errorf("Expected 'name' on line 3, got line %d", tok.Line)
			}
			break
		}
	}
}

func TestLexError(t *testing.T) {
	input := `struct Device { @ }`
	
	lexer := NewLexer(input)
	_, err := lexer.Tokenize()
	if err == nil {
		t.Error("Expected error for invalid character '@', got nil")
	}
}

func TestLexEmpty(t *testing.T) {
	input := ""
	
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}
	
	if len(tokens) != 1 || tokens[0].Type != TokenEOF {
		t.Errorf("Expected single EOF token for empty input")
	}
}

func TestLexWhitespace(t *testing.T) {
	input := "  \t\n  struct  \n\t  Device  "
	
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize failed: %v", err)
	}
	
	// Should ignore whitespace and only return struct, Device, EOF
	expected := []TokenType{TokenStruct, TokenIdent, TokenEOF}
	
	if len(tokens) != len(expected) {
		t.Fatalf("Expected %d tokens, got %d", len(expected), len(tokens))
	}
	
	for i, tok := range tokens {
		if tok.Type != expected[i] {
			t.Errorf("Token %d: expected %v, got %v", i, expected[i], tok.Type)
		}
	}
}
