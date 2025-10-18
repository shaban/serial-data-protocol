package golang

import (
	"strings"
	"testing"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

func TestGenerateMessageDecoders(t *testing.T) {
	tests := []struct {
		name      string
		schema    *parser.Schema
		wantErr   bool
		checkFunc func(t *testing.T, code string)
	}{
		{
			name:    "nil schema",
			schema:  nil,
			wantErr: true,
		},
		{
			name: "single struct",
			schema: &parser.Schema{
				Structs: []parser.Struct{
					{
						Name: "Point",
						Fields: []parser.Field{
							{Name: "x", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i32"}},
							{Name: "y", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i32"}},
						},
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, code string) {
				// Check function exists
				if !strings.Contains(code, "func DecodePointMessage(") {
					t.Errorf("missing DecodePointMessage function")
				}
				// Check header comment
				if !strings.Contains(code, "decodes a Point from self-describing message format") {
					t.Errorf("missing function doc comment")
				}
				// Check minimum size check
				if !strings.Contains(code, "if len(data) < MessageHeaderSize") {
					t.Errorf("missing minimum size check")
				}
				// Check magic bytes validation
				if !strings.Contains(code, "if string(data[0:3]) != MessageMagic") {
					t.Errorf("missing magic bytes validation")
				}
				if !strings.Contains(code, "return nil, ErrInvalidMagic") {
					t.Errorf("missing ErrInvalidMagic return")
				}
				// Check version validation
				if !strings.Contains(code, "if data[3] != MessageVersion") {
					t.Errorf("missing version validation")
				}
				if !strings.Contains(code, "return nil, ErrInvalidVersion") {
					t.Errorf("missing ErrInvalidVersion return")
				}
				// Check type ID validation
				if !strings.Contains(code, "typeID := binary.LittleEndian.Uint16(data[4:6])") {
					t.Errorf("missing type ID extraction")
				}
				if !strings.Contains(code, "if typeID != 1") {
					t.Errorf("missing or incorrect type ID validation")
				}
				if !strings.Contains(code, "return nil, ErrUnknownMessageType") {
					t.Errorf("missing ErrUnknownMessageType return")
				}
				// Check payload length extraction
				if !strings.Contains(code, "payloadLength := binary.LittleEndian.Uint32(data[6:10])") {
					t.Errorf("missing payload length extraction")
				}
				// Check total size validation
				if !strings.Contains(code, "expectedSize := MessageHeaderSize + int(payloadLength)") {
					t.Errorf("missing expected size calculation")
				}
				if !strings.Contains(code, "if len(data) < expectedSize") {
					t.Errorf("missing total size validation")
				}
				// Check payload extraction
				if !strings.Contains(code, "payload := data[MessageHeaderSize:expectedSize]") {
					t.Errorf("missing payload extraction")
				}
				// Check payload decoding
				if !strings.Contains(code, "DecodePoint(&result, payload)") {
					t.Errorf("missing payload decode call")
				}
			},
		},
		{
			name: "multiple structs with sequential type IDs",
			schema: &parser.Schema{
				Structs: []parser.Struct{
					{
						Name: "Point",
						Fields: []parser.Field{
							{Name: "x", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i32"}},
						},
					},
					{
						Name: "Color",
						Fields: []parser.Field{
							{Name: "r", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"}},
						},
					},
					{
						Name: "Size",
						Fields: []parser.Field{
							{Name: "w", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u16"}},
						},
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, code string) {
				// Check all three functions exist
				if !strings.Contains(code, "func DecodePointMessage(") {
					t.Errorf("missing DecodePointMessage")
				}
				if !strings.Contains(code, "func DecodeColorMessage(") {
					t.Errorf("missing DecodeColorMessage")
				}
				if !strings.Contains(code, "func DecodeSizeMessage(") {
					t.Errorf("missing DecodeSizeMessage")
				}
				// Check type IDs are sequential (1, 2, 3)
				if !strings.Contains(code, "if typeID != 1") {
					t.Errorf("Point should validate type ID 1")
				}
				if !strings.Contains(code, "if typeID != 2") {
					t.Errorf("Color should validate type ID 2")
				}
				if !strings.Contains(code, "if typeID != 3") {
					t.Errorf("Size should validate type ID 3")
				}
			},
		},
		{
			name: "struct with underscore name",
			schema: &parser.Schema{
				Structs: []parser.Struct{
					{
						Name: "data_packet",
						Fields: []parser.Field{
							{Name: "value", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i32"}},
						},
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, code string) {
				// Check CamelCase conversion
				if !strings.Contains(code, "func DecodeDataPacketMessage(") {
					t.Errorf("expected DecodeDataPacketMessage, check name conversion")
				}
				if !strings.Contains(code, "DecodeDataPacket(&result, payload)") {
					t.Errorf("expected DecodeDataPacket call")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := GenerateMessageDecoders(tt.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateMessageDecoders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, code)
			}
		})
	}
}

func TestGenerateMessageDispatcher(t *testing.T) {
	tests := []struct {
		name      string
		schema    *parser.Schema
		wantErr   bool
		checkFunc func(t *testing.T, code string)
	}{
		{
			name:    "nil schema",
			schema:  nil,
			wantErr: true,
		},
		{
			name: "single struct",
			schema: &parser.Schema{
				Structs: []parser.Struct{
					{
						Name: "Point",
						Fields: []parser.Field{
							{Name: "x", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i32"}},
						},
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, code string) {
				// Check function signature
				if !strings.Contains(code, "func DecodeMessage(data []byte) (interface{}, error)") {
					t.Errorf("missing DecodeMessage function signature")
				}
				// Check doc comment
				if !strings.Contains(code, "main entry point for decoding self-describing messages") {
					t.Errorf("missing function doc comment")
				}
				// Check minimum size check
				if !strings.Contains(code, "if len(data) < MessageHeaderSize") {
					t.Errorf("missing minimum size check")
				}
				// Check magic validation
				if !strings.Contains(code, "if string(data[0:3]) != MessageMagic") {
					t.Errorf("missing magic bytes validation")
				}
				// Check version validation
				if !strings.Contains(code, "if data[3] != MessageVersion") {
					t.Errorf("missing version validation")
				}
				// Check type ID extraction
				if !strings.Contains(code, "typeID := binary.LittleEndian.Uint16(data[4:6])") {
					t.Errorf("missing type ID extraction")
				}
				// Check switch statement
				if !strings.Contains(code, "switch typeID") {
					t.Errorf("missing switch on typeID")
				}
				// Check case for Point (type ID 1)
				if !strings.Contains(code, "case 1:") {
					t.Errorf("missing case 1")
				}
				if !strings.Contains(code, "return DecodePointMessage(data)") {
					t.Errorf("missing DecodePointMessage call")
				}
				// Check default case
				if !strings.Contains(code, "default:") {
					t.Errorf("missing default case")
				}
				if !strings.Contains(code, "return nil, ErrUnknownMessageType") {
					t.Errorf("missing ErrUnknownMessageType in default")
				}
			},
		},
		{
			name: "multiple structs",
			schema: &parser.Schema{
				Structs: []parser.Struct{
					{
						Name: "Point",
						Fields: []parser.Field{
							{Name: "x", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "i32"}},
						},
					},
					{
						Name: "Color",
						Fields: []parser.Field{
							{Name: "r", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u8"}},
						},
					},
					{
						Name: "Size",
						Fields: []parser.Field{
							{Name: "w", Type: parser.TypeExpr{Kind: parser.TypeKindPrimitive, Name: "u16"}},
						},
					},
				},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, code string) {
				// Check all three cases exist
				if !strings.Contains(code, "case 1:") {
					t.Errorf("missing case 1 for Point")
				}
				if !strings.Contains(code, "case 2:") {
					t.Errorf("missing case 2 for Color")
				}
				if !strings.Contains(code, "case 3:") {
					t.Errorf("missing case 3 for Size")
				}
				// Check decoder calls
				if !strings.Contains(code, "return DecodePointMessage(data)") {
					t.Errorf("missing DecodePointMessage call")
				}
				if !strings.Contains(code, "return DecodeColorMessage(data)") {
					t.Errorf("missing DecodeColorMessage call")
				}
				if !strings.Contains(code, "return DecodeSizeMessage(data)") {
					t.Errorf("missing DecodeSizeMessage call")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := GenerateMessageDispatcher(tt.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateMessageDispatcher() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, code)
			}
		})
	}
}
