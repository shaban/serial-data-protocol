package parser

import (
	"fmt"
)

// Parser parses tokenized .sdp schema files into an AST.
type Parser struct {
	tokens []Token
	pos    int
}

// NewParser creates a new parser for the given tokens.
func NewParser(tokens []Token) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
	}
}

// ParseSchema parses the entire schema file.
func ParseSchema(input string) (*Schema, error) {
	// Lex the input
	lexer := NewLexer(input)
	tokens, err := lexer.Tokenize()
	if err != nil {
		return nil, err
	}
	
	// Parse tokens into AST
	parser := NewParser(tokens)
	return parser.parseSchema()
}

// parseSchema parses: Schema = { Struct }
func (p *Parser) parseSchema() (*Schema, error) {
	schema := &Schema{
		Structs: make([]Struct, 0),
	}
	
	// Skip any leading regular comments (not doc comments)
	p.skipRegularComments()
	
	for !p.isAtEnd() {
		// Parse struct definition
		s, err := p.parseStruct()
		if err != nil {
			return nil, err
		}
		schema.Structs = append(schema.Structs, s)
		
		// Skip any trailing regular comments (not doc comments)
		p.skipRegularComments()
	}
	
	return schema, nil
}

// parseStruct parses: Struct = [ DocComment ] "struct" Ident "{" [ FieldList ] "}"
func (p *Parser) parseStruct() (Struct, error) {
	s := Struct{
		Fields: make([]Field, 0),
	}
	
	// Collect doc comments
	s.Comment = p.collectDocComments()
	
	// Expect 'struct' keyword
	if !p.match(TokenStruct) {
		return s, p.error("expected 'struct'")
	}
	
	// Expect struct name
	if !p.check(TokenIdent) {
		return s, p.error("expected struct name")
	}
	name := p.advance()
	s.Name = name.Value
	
	// Expect '{'
	if !p.match(TokenLBrace) {
		return s, p.error("expected '{'")
	}
	
	// Parse fields
	for !p.check(TokenRBrace) && !p.isAtEnd() {
		// Skip any regular comments before fields (not doc comments)
		p.skipRegularComments()
		
		// Check again after skipping comments
		if p.check(TokenRBrace) || p.isAtEnd() {
			break
		}
		
		field, err := p.parseField()
		if err != nil {
			return s, err
		}
		s.Fields = append(s.Fields, field)
		
		// Expect comma (optional after last field)
		if p.match(TokenComma) {
			// Comma consumed, continue
			p.skipRegularComments()
		} else if !p.check(TokenRBrace) {
			return s, p.error("expected ',' or '}'")
		}
	}
	
	// Expect '}'
	if !p.match(TokenRBrace) {
		return s, p.error("expected '}'")
	}
	
	return s, nil
}

// parseField parses: Field = [ DocComment ] Ident ":" TypeExpr
func (p *Parser) parseField() (Field, error) {
	f := Field{}
	
	// Collect doc comments
	f.Comment = p.collectDocComments()
	
	// Expect field name
	if !p.check(TokenIdent) {
		return f, p.error("expected field name")
	}
	name := p.advance()
	f.Name = name.Value
	
	// Expect ':'
	if !p.match(TokenColon) {
		return f, p.error("expected ':'")
	}
	
	// Parse type expression
	typeExpr, err := p.parseTypeExpr()
	if err != nil {
		return f, err
	}
	f.Type = typeExpr
	
	return f, nil
}

// parseTypeExpr parses: TypeExpr = Ident | "[" "]" TypeExpr
func (p *Parser) parseTypeExpr() (TypeExpr, error) {
	// Check for array type: []T
	if p.check(TokenLBracket) {
		p.advance() // consume '['
		
		if !p.match(TokenRBracket) {
			return TypeExpr{}, p.error("expected ']' after '['")
		}
		
		// Parse element type
		elemType, err := p.parseTypeExpr()
		if err != nil {
			return TypeExpr{}, err
		}
		
		return TypeExpr{
			Kind: TypeKindArray,
			Elem: &elemType,
		}, nil
	}
	
	// Must be an identifier (primitive or named type)
	if !p.check(TokenIdent) {
		return TypeExpr{}, p.error("expected type name")
	}
	
	typeName := p.advance()
	
	// Determine if it's a primitive or named type
	typeExpr := TypeExpr{
		Name: typeName.Value,
	}
	
	if typeExpr.IsPrimitive() {
		typeExpr.Kind = TypeKindPrimitive
	} else {
		typeExpr.Kind = TypeKindNamed
	}
	
	return typeExpr, nil
}

// collectDocComments collects consecutive doc comments and returns them as a single string.
func (p *Parser) collectDocComments() string {
	// First skip any regular comments
	for p.check(TokenComment) {
		p.advance()
	}
	
	// Then collect doc comments
	var comments []string
	
	for p.check(TokenDocComment) {
		tok := p.advance()
		comments = append(comments, tok.Value)
	}
	
	if len(comments) == 0 {
		return ""
	}
	
	// Join with newlines
	result := comments[0]
	for i := 1; i < len(comments); i++ {
		result += "\n" + comments[i]
	}
	
	return result
}

// skipComments skips both doc comments and regular comments.
func (p *Parser) skipComments() {
	for p.check(TokenComment) || p.check(TokenDocComment) {
		p.advance()
	}
}

// skipRegularComments skips only regular comments, not doc comments.
func (p *Parser) skipRegularComments() {
	for p.check(TokenComment) {
		p.advance()
	}
}

// check returns true if the current token is of the given type.
func (p *Parser) check(t TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == t
}

// match consumes the current token if it matches the given type.
func (p *Parser) match(t TokenType) bool {
	if p.check(t) {
		p.advance()
		return true
	}
	return false
}

// advance consumes and returns the current token.
func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.pos++
	}
	return p.previous()
}

// peek returns the current token without consuming it.
func (p *Parser) peek() Token {
	return p.tokens[p.pos]
}

// previous returns the previous token.
func (p *Parser) previous() Token {
	return p.tokens[p.pos-1]
}

// isAtEnd returns true if we've reached EOF.
func (p *Parser) isAtEnd() bool {
	return p.pos >= len(p.tokens) || p.tokens[p.pos].Type == TokenEOF
}

// error creates a parse error with location information.
func (p *Parser) error(msg string) error {
	tok := p.peek()
	return fmt.Errorf("line %d, column %d: %s (got %s)", tok.Line, tok.Column, msg, tok.String())
}
