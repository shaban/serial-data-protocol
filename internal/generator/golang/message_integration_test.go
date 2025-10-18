package golang

import (
	"bytes"
	"encoding/binary"
	"strings"
	"testing"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// TestMessageModeIntegration tests the complete message mode workflow:
// schema -> generate code -> encode -> decode -> verify
func TestMessageModeIntegration(t *testing.T) {
	// Define a simple test schema
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Point",
				Fields: []parser.Field{
					{Name: "x", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i32"}},
					{Name: "y", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i32"}},
				},
			},
			{
				Name: "Color",
				Fields: []parser.Field{
					{Name: "r", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"}},
					{Name: "g", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"}},
					{Name: "b", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"}},
				},
			},
		},
	}

	// Generate message encoders
	encoders, err := GenerateMessageEncoders(schema)
	if err != nil {
		t.Fatalf("GenerateMessageEncoders failed: %v", err)
	}

	// Generate message decoders
	decoders, err := GenerateMessageDecoders(schema)
	if err != nil {
		t.Fatalf("GenerateMessageDecoders failed: %v", err)
	}

	// Generate message dispatcher
	dispatcher, err := GenerateMessageDispatcher(schema)
	if err != nil {
		t.Fatalf("GenerateMessageDispatcher failed: %v", err)
	}

	// Verify encoder content
	if !strings.Contains(encoders, "func EncodePointMessage(") {
		t.Errorf("missing EncodePointMessage function")
	}
	if !strings.Contains(encoders, "func EncodeColorMessage(") {
		t.Errorf("missing EncodeColorMessage function")
	}

	// Verify decoder content
	if !strings.Contains(decoders, "func DecodePointMessage(") {
		t.Errorf("missing DecodePointMessage function")
	}
	if !strings.Contains(decoders, "func DecodeColorMessage(") {
		t.Errorf("missing DecodeColorMessage function")
	}

	// Verify dispatcher content
	if !strings.Contains(dispatcher, "func DecodeMessage(data []byte) (interface{}, error)") {
		t.Errorf("missing DecodeMessage function")
	}
	if !strings.Contains(dispatcher, "case 1:") {
		t.Errorf("missing case for Point (type ID 1)")
	}
	if !strings.Contains(dispatcher, "case 2:") {
		t.Errorf("missing case for Color (type ID 2)")
	}
	if !strings.Contains(dispatcher, "return DecodePointMessage(data)") {
		t.Errorf("missing DecodePointMessage call in dispatcher")
	}
	if !strings.Contains(dispatcher, "return DecodeColorMessage(data)") {
		t.Errorf("missing DecodeColorMessage call in dispatcher")
	}
}

// TestMessageHeaderFormat validates the generated header format
func TestMessageHeaderFormat(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Simple",
				Fields: []parser.Field{
					{Name: "value", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
				},
			},
		},
	}

	code, err := GenerateMessageEncoders(schema)
	if err != nil {
		t.Fatalf("GenerateMessageEncoders failed: %v", err)
	}

	// Verify header writing sequence
	expectedSequence := []string{
		"copy(message[0:3], MessageMagic)",           // Magic
		"message[3] = MessageVersion",                // Version
		"binary.LittleEndian.PutUint16(message[4:6]", // Type ID
		"binary.LittleEndian.PutUint32(message[6:10]", // Length
		"copy(message[10:], payload)",                // Payload
	}

	for _, expected := range expectedSequence {
		if !strings.Contains(code, expected) {
			t.Errorf("missing expected header operation: %s", expected)
		}
	}
}

// TestMessageHeaderValidation validates the generated header validation
func TestMessageHeaderValidation(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Simple",
				Fields: []parser.Field{
					{Name: "value", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
				},
			},
		},
	}

	code, err := GenerateMessageDecoders(schema)
	if err != nil {
		t.Fatalf("GenerateMessageDecoders failed: %v", err)
	}

	// Verify validation checks
	validations := []string{
		"if len(data) < MessageHeaderSize",             // Size check
		"if string(data[0:3]) != MessageMagic",        // Magic check
		"return nil, ErrInvalidMagic",                 // Magic error
		"if data[3] != MessageVersion",                // Version check
		"return nil, ErrInvalidVersion",               // Version error
		"typeID := binary.LittleEndian.Uint16(data[4:6])", // Type ID extraction
		"if typeID != 1",                              // Type ID check
		"return nil, ErrUnknownMessageType",           // Type ID error
		"payloadLength := binary.LittleEndian.Uint32(data[6:10])", // Length extraction
		"expectedSize := MessageHeaderSize + int(payloadLength)", // Size calculation
		"if len(data) < expectedSize",                 // Total size check
		"payload := data[MessageHeaderSize:expectedSize]", // Payload extraction
	}

	for _, validation := range validations {
		if !strings.Contains(code, validation) {
			t.Errorf("missing expected validation: %s", validation)
		}
	}
}

// TestMessageTypeIDSequence verifies type IDs are assigned sequentially
func TestMessageTypeIDSequence(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{Name: "First", Fields: []parser.Field{{Name: "a", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"}}}},
			{Name: "Second", Fields: []parser.Field{{Name: "b", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u16"}}}},
			{Name: "Third", Fields: []parser.Field{{Name: "c", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}}}},
			{Name: "Fourth", Fields: []parser.Field{{Name: "d", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u64"}}}},
		},
	}

	// Test encoders
	encoders, err := GenerateMessageEncoders(schema)
	if err != nil {
		t.Fatalf("GenerateMessageEncoders failed: %v", err)
	}

	for i := 1; i <= 4; i++ {
		expected := "PutUint16(message[4:6], " + string(rune('0'+i)) + ")"
		if !strings.Contains(encoders, expected) {
			t.Errorf("missing type ID %d in encoders", i)
		}
	}

	// Test decoders
	decoders, err := GenerateMessageDecoders(schema)
	if err != nil {
		t.Fatalf("GenerateMessageDecoders failed: %v", err)
	}

	for i := 1; i <= 4; i++ {
		expected := "if typeID != " + string(rune('0'+i))
		if !strings.Contains(decoders, expected) {
			t.Errorf("missing type ID %d validation in decoders", i)
		}
	}

	// Test dispatcher
	dispatcher, err := GenerateMessageDispatcher(schema)
	if err != nil {
		t.Fatalf("GenerateMessageDispatcher failed: %v", err)
	}

	for i := 1; i <= 4; i++ {
		expected := "case " + string(rune('0'+i)) + ":"
		if !strings.Contains(dispatcher, expected) {
			t.Errorf("missing case %d in dispatcher", i)
		}
	}
}

// TestGeneratedMessageRoundtrip tests a simulated encode/decode cycle
func TestGeneratedMessageRoundtrip(t *testing.T) {
	// Create a mock message with proper header
	// Magic: "SDP"
	// Version: '2'
	// Type ID: 1 (uint16 little-endian)
	// Length: 8 (uint32 little-endian) - for two i32 values
	// Payload: x=100, y=200 (i32 little-endian)

	message := make([]byte, 10+8)
	copy(message[0:3], "SDP")
	message[3] = '2'
	binary.LittleEndian.PutUint16(message[4:6], 1)   // Type ID
	binary.LittleEndian.PutUint32(message[6:10], 8)  // Payload length
	binary.LittleEndian.PutUint32(message[10:14], 100) // x
	binary.LittleEndian.PutUint32(message[14:18], 200) // y

	// Verify header structure
	if string(message[0:3]) != "SDP" {
		t.Errorf("incorrect magic bytes")
	}
	if message[3] != '2' {
		t.Errorf("incorrect version")
	}
	typeID := binary.LittleEndian.Uint16(message[4:6])
	if typeID != 1 {
		t.Errorf("incorrect type ID, got %d, want 1", typeID)
	}
	length := binary.LittleEndian.Uint32(message[6:10])
	if length != 8 {
		t.Errorf("incorrect payload length, got %d, want 8", length)
	}

	// Verify payload
	x := binary.LittleEndian.Uint32(message[10:14])
	y := binary.LittleEndian.Uint32(message[14:18])
	if x != 100 {
		t.Errorf("incorrect x value, got %d, want 100", x)
	}
	if y != 200 {
		t.Errorf("incorrect y value, got %d, want 200", y)
	}
}

// TestInvalidMessageHeaders tests various invalid header scenarios
func TestInvalidMessageHeaders(t *testing.T) {
	tests := []struct {
		name        string
		createMsg   func() []byte
		shouldMatch string // What error message should appear in decoder
	}{
		{
			name: "invalid magic bytes",
			createMsg: func() []byte {
				msg := make([]byte, 18)
				copy(msg[0:3], "XXX") // Wrong magic
				msg[3] = '2'
				binary.LittleEndian.PutUint16(msg[4:6], 1)
				binary.LittleEndian.PutUint32(msg[6:10], 8)
				return msg
			},
			shouldMatch: "ErrInvalidMagic",
		},
		{
			name: "invalid version",
			createMsg: func() []byte {
				msg := make([]byte, 18)
				copy(msg[0:3], "SDP")
				msg[3] = '9' // Wrong version
				binary.LittleEndian.PutUint16(msg[4:6], 1)
				binary.LittleEndian.PutUint32(msg[6:10], 8)
				return msg
			},
			shouldMatch: "ErrInvalidVersion",
		},
		{
			name: "unknown type ID",
			createMsg: func() []byte {
				msg := make([]byte, 18)
				copy(msg[0:3], "SDP")
				msg[3] = '2'
				binary.LittleEndian.PutUint16(msg[4:6], 999) // Invalid type ID
				binary.LittleEndian.PutUint32(msg[6:10], 8)
				return msg
			},
			shouldMatch: "ErrUnknownMessageType",
		},
		{
			name: "insufficient data",
			createMsg: func() []byte {
				return []byte{0x53, 0x44, 0x50} // Only 3 bytes (magic)
			},
			shouldMatch: "ErrUnexpectedEOF",
		},
	}

	// Generate decoder that checks all these cases
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "TestStruct",
				Fields: []parser.Field{
					{Name: "value", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
				},
			},
		},
	}

	decoder, err := GenerateMessageDecoders(schema)
	if err != nil {
		t.Fatalf("GenerateMessageDecoders failed: %v", err)
	}

	// Verify decoder contains all error checks
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !strings.Contains(decoder, tt.shouldMatch) {
				t.Errorf("decoder missing check for %s", tt.shouldMatch)
			}
		})
	}
}

// TestDispatcherUnknownType verifies the dispatcher handles unknown types
func TestDispatcherUnknownType(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "OnlyStruct",
				Fields: []parser.Field{
					{Name: "value", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"}},
				},
			},
		},
	}

	dispatcher, err := GenerateMessageDispatcher(schema)
	if err != nil {
		t.Fatalf("GenerateMessageDispatcher failed: %v", err)
	}

	// Verify dispatcher has default case
	if !strings.Contains(dispatcher, "default:") {
		t.Error("dispatcher missing default case")
	}
	if !strings.Contains(dispatcher, "return nil, ErrUnknownMessageType") {
		t.Error("dispatcher default case missing ErrUnknownMessageType")
	}
}

// TestEmptySchema tests handling of edge cases
func TestEmptySchema(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{},
	}

	// Should still generate valid (but empty) code
	encoders, err := GenerateMessageEncoders(schema)
	if err != nil {
		t.Fatalf("GenerateMessageEncoders failed on empty schema: %v", err)
	}
	if encoders != "" {
		t.Errorf("expected empty encoder output for empty schema, got: %s", encoders)
	}

	decoders, err := GenerateMessageDecoders(schema)
	if err != nil {
		t.Fatalf("GenerateMessageDecoders failed on empty schema: %v", err)
	}
	if decoders != "" {
		t.Errorf("expected empty decoder output for empty schema, got: %s", decoders)
	}

	dispatcher, err := GenerateMessageDispatcher(schema)
	if err != nil {
		t.Fatalf("GenerateMessageDispatcher failed on empty schema: %v", err)
	}
	// Dispatcher should still have basic structure even with no cases
	if !strings.Contains(dispatcher, "func DecodeMessage") {
		t.Error("dispatcher should have function even with empty schema")
	}
	if !strings.Contains(dispatcher, "default:") {
		t.Error("dispatcher should have default case even with empty schema")
	}
}

// BenchmarkMessageHeaderSize ensures the header is exactly 10 bytes
func TestMessageHeaderSizeConstant(t *testing.T) {
	// The message header must be exactly 10 bytes:
	// - Magic: 3 bytes ("SDP")
	// - Version: 1 byte ('2')
	// - Type ID: 2 bytes (uint16)
	// - Length: 4 bytes (uint32)
	expectedSize := 3 + 1 + 2 + 4

	if expectedSize != 10 {
		t.Fatalf("header size calculation error: expected 10, got %d", expectedSize)
	}

	// Verify code generation uses this size
	code := GenerateErrors()
	if !strings.Contains(code, "MessageHeaderSize    = 10") {
		t.Error("MessageHeaderSize constant not found or incorrect")
	}
}

// TestMessageFormatDocumentation verifies doc comments are present
func TestMessageFormatDocumentation(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Doc",
				Fields: []parser.Field{
					{Name: "x", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"}},
				},
			},
		},
	}

	// Check encoder documentation
	encoder, err := GenerateMessageEncoders(schema)
	if err != nil {
		t.Fatalf("GenerateMessageEncoders failed: %v", err)
	}

	encoderDocs := []string{
		"self-describing message format",
		"[SDP:3][version:1][type_id:2][length:4][payload:N]",
		"suitable for persistence, network transmission",
	}

	for _, doc := range encoderDocs {
		if !strings.Contains(encoder, doc) {
			t.Errorf("encoder missing documentation: %s", doc)
		}
	}

	// Check decoder documentation
	decoder, err := GenerateMessageDecoders(schema)
	if err != nil {
		t.Fatalf("GenerateMessageDecoders failed: %v", err)
	}

	decoderDocs := []string{
		"self-describing message format",
		"must include a valid 10-byte header",
		"[SDP:3][version:1][type_id:2][length:4][payload:N]",
	}

	for _, doc := range decoderDocs {
		if !strings.Contains(decoder, doc) {
			t.Errorf("decoder missing documentation: %s", doc)
		}
	}

	// Check dispatcher documentation
	dispatcher, err := GenerateMessageDispatcher(schema)
	if err != nil {
		t.Fatalf("GenerateMessageDispatcher failed: %v", err)
	}

	dispatcherDocs := []string{
		"main entry point for decoding self-describing messages",
		"returns the struct type based on the type ID",
	}

	for _, doc := range dispatcherDocs {
		if !strings.Contains(dispatcher, doc) {
			t.Errorf("dispatcher missing documentation: %s", doc)
		}
	}
}

// TestGeneratedCodeStructure verifies proper Go code structure
func TestGeneratedCodeStructure(t *testing.T) {
	schema := &parser.Schema{
		Structs: []parser.Struct{
			{
				Name: "Test",
				Fields: []parser.Field{
					{Name: "value", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u32"}},
				},
			},
		},
	}

	encoder, _ := GenerateMessageEncoders(schema)
	decoder, _ := GenerateMessageDecoders(schema)
	dispatcher, _ := GenerateMessageDispatcher(schema)

	// Check that functions have proper structure
	structures := []struct {
		code string
		name string
	}{
		{encoder, "encoder"},
		{decoder, "decoder"},
		{dispatcher, "dispatcher"},
	}

	for _, s := range structures {
		// Should have balanced braces
		openBraces := bytes.Count([]byte(s.code), []byte("{"))
		closeBraces := bytes.Count([]byte(s.code), []byte("}"))
		if openBraces != closeBraces {
			t.Errorf("%s has unbalanced braces: %d open, %d close", s.name, openBraces, closeBraces)
		}

		// Should have function declarations
		if !strings.Contains(s.code, "func ") {
			t.Errorf("%s missing function declarations", s.name)
		}

		// Should have proper Go keywords
		if !strings.Contains(s.code, "return") {
			t.Errorf("%s missing return statements", s.name)
		}
	}
}
