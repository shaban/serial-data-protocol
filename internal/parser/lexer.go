package parser

import (
	"fmt"
	"unicode"
)

// TokenType represents the type of a lexical token.
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError

	// Literals and identifiers
	TokenIdent // field_name, MyStruct, u32, etc.

	// Keywords
	TokenStruct // struct

	// Punctuation
	TokenLBrace  // {
	TokenRBrace  // }
	TokenLBracket // [
	TokenRBracket // ]
	TokenColon   // :
	TokenComma   // ,

	// Comments
	TokenDocComment // /// documentation
	TokenComment    // // regular comment (not emitted, but tracked for testing)
)

// Token represents a lexical token.
type Token struct {
	Type   TokenType
	Value  string // Actual text (for identifiers, comments)
	Line   int    // Line number (1-indexed)
	Column int    // Column number (1-indexed)
}

// String returns a human-readable representation of the token.
func (t Token) String() string {
	switch t.Type {
	case TokenEOF:
		return "EOF"
	case TokenError:
		return fmt.Sprintf("ERROR(%s)", t.Value)
	case TokenIdent:
		return fmt.Sprintf("IDENT(%s)", t.Value)
	case TokenStruct:
		return "struct"
	case TokenLBrace:
		return "{"
	case TokenRBrace:
		return "}"
	case TokenLBracket:
		return "["
	case TokenRBracket:
		return "]"
	case TokenColon:
		return ":"
	case TokenComma:
		return ","
	case TokenDocComment:
		return fmt.Sprintf("DOC(%s)", t.Value)
	case TokenComment:
		return fmt.Sprintf("COMMENT(%s)", t.Value)
	default:
		return fmt.Sprintf("UNKNOWN(%d)", t.Type)
	}
}

// Lexer tokenizes .sdp schema files.
type Lexer struct {
	input  string
	pos    int  // Current position in input
	line   int  // Current line (1-indexed)
	column int  // Current column (1-indexed)
	tokens []Token
}

// NewLexer creates a new lexer for the given input.
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		line:   1,
		column: 1,
		tokens: make([]Token, 0),
	}
}

// Tokenize lexes the entire input and returns all tokens.
func (l *Lexer) Tokenize() ([]Token, error) {
	for {
		tok := l.nextToken()
		l.tokens = append(l.tokens, tok)
		
		if tok.Type == TokenEOF {
			break
		}
		if tok.Type == TokenError {
			return nil, fmt.Errorf("line %d, column %d: %s", tok.Line, tok.Column, tok.Value)
		}
	}
	return l.tokens, nil
}

// nextToken returns the next token from the input.
func (l *Lexer) nextToken() Token {
	l.skipWhitespace()
	
	if l.isAtEnd() {
		return l.makeToken(TokenEOF, "")
	}
	
	ch := l.peek()
	
	// Comments
	if ch == '/' && l.peekAhead(1) == '/' {
		return l.lexComment()
	}
	
	// Single-character tokens
	switch ch {
	case '{':
		return l.advance(TokenLBrace, "{")
	case '}':
		return l.advance(TokenRBrace, "}")
	case '[':
		return l.advance(TokenLBracket, "[")
	case ']':
		return l.advance(TokenRBracket, "]")
	case ':':
		return l.advance(TokenColon, ":")
	case ',':
		return l.advance(TokenComma, ",")
	}
	
	// Identifiers and keywords
	if isIdentStart(ch) {
		return l.lexIdent()
	}
	
	// Unknown character
	return l.makeToken(TokenError, fmt.Sprintf("unexpected character: %q", ch))
}

// lexComment handles both doc comments (///) and regular comments (//).
func (l *Lexer) lexComment() Token {
	line := l.line
	col := l.column
	
	l.consume() // first /
	l.consume() // second /
	
	isDoc := false
	if l.peek() == '/' {
		isDoc = true
		l.consume() // third /
	}
	
	// Skip any spaces after //[/]
	for l.peek() == ' ' || l.peek() == '\t' {
		l.consume()
	}
	
	// Read until end of line
	start := l.pos
	for !l.isAtEnd() && l.peek() != '\n' {
		l.consume()
	}
	
	text := l.input[start:l.pos]
	
	if isDoc {
		return Token{Type: TokenDocComment, Value: text, Line: line, Column: col}
	}
	return Token{Type: TokenComment, Value: text, Line: line, Column: col}
}

// lexIdent reads an identifier or keyword.
func (l *Lexer) lexIdent() Token {
	line := l.line
	col := l.column
	start := l.pos
	
	for !l.isAtEnd() && isIdentContinue(l.peek()) {
		l.consume()
	}
	
	value := l.input[start:l.pos]
	
	// Check for keywords
	var tokType TokenType
	switch value {
	case "struct":
		tokType = TokenStruct
	default:
		tokType = TokenIdent
	}
	
	return Token{Type: tokType, Value: value, Line: line, Column: col}
}

// skipWhitespace skips whitespace characters.
func (l *Lexer) skipWhitespace() {
	for !l.isAtEnd() {
		ch := l.peek()
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			l.consume()
		} else {
			break
		}
	}
}

// peek returns the current character without advancing.
func (l *Lexer) peek() rune {
	if l.isAtEnd() {
		return 0
	}
	return rune(l.input[l.pos])
}

// peekAhead returns the character n positions ahead without advancing.
func (l *Lexer) peekAhead(n int) rune {
	pos := l.pos + n
	if pos >= len(l.input) {
		return 0
	}
	return rune(l.input[pos])
}

// consume advances the position by one character.
func (l *Lexer) consume() {
	if l.isAtEnd() {
		return
	}
	
	ch := l.input[l.pos]
	l.pos++
	
	if ch == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
}

// advance consumes the current character and returns a token.
func (l *Lexer) advance(tokType TokenType, value string) Token {
	tok := l.makeToken(tokType, value)
	l.consume()
	return tok
}

// makeToken creates a token at the current position.
func (l *Lexer) makeToken(tokType TokenType, value string) Token {
	return Token{
		Type:   tokType,
		Value:  value,
		Line:   l.line,
		Column: l.column,
	}
}

// isAtEnd returns true if we've reached the end of input.
func (l *Lexer) isAtEnd() bool {
	return l.pos >= len(l.input)
}

// isIdentStart returns true if the rune can start an identifier.
func isIdentStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

// isIdentContinue returns true if the rune can continue an identifier.
func isIdentContinue(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_'
}
