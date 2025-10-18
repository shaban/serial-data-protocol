package golang

import (
	"strings"
	"testing"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

func TestGenerateMessageEncoders(t *testing.T) {
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
				if !strings.Contains(code, "func EncodePointMessage(") {
					t.Errorf("missing EncodePointMessage function")
				}
				// Check header comment
				if !strings.Contains(code, "encodes a Point to self-describing message format") {
					t.Errorf("missing function doc comment")
				}
				// Check header format comment
				if !strings.Contains(code, "[SDP:3][version:1][type_id:2][length:4][payload:N]") {
					t.Errorf("missing header format in comment")
				}
				// Check payload encoding
				if !strings.Contains(code, "payload, err := EncodePoint(src)") {
					t.Errorf("missing payload encoding call")
				}
				// Check message allocation
				if !strings.Contains(code, "messageSize := MessageHeaderSize + len(payload)") {
					t.Errorf("missing message size calculation")
				}
				if !strings.Contains(code, "message := make([]byte, messageSize)") {
					t.Errorf("missing message allocation")
				}
				// Check magic bytes
				if !strings.Contains(code, "copy(message[0:3], MessageMagic)") {
					t.Errorf("missing magic bytes write")
				}
				// Check version
				if !strings.Contains(code, "message[3] = MessageVersion") {
					t.Errorf("missing version write")
				}
				// Check type ID (should be 1 for first struct)
				if !strings.Contains(code, "PutUint16(message[4:6], 1)") {
					t.Errorf("missing or incorrect type ID write, expected type ID 1")
				}
				// Check length
				if !strings.Contains(code, "PutUint32(message[6:10], uint32(len(payload)))") {
					t.Errorf("missing payload length write")
				}
				// Check payload copy
				if !strings.Contains(code, "copy(message[10:], payload)") {
					t.Errorf("missing payload copy")
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
				if !strings.Contains(code, "func EncodePointMessage(") {
					t.Errorf("missing EncodePointMessage")
				}
				if !strings.Contains(code, "func EncodeColorMessage(") {
					t.Errorf("missing EncodeColorMessage")
				}
				if !strings.Contains(code, "func EncodeSizeMessage(") {
					t.Errorf("missing EncodeSizeMessage")
				}
				// Check type IDs are sequential (1, 2, 3)
				if !strings.Contains(code, "PutUint16(message[4:6], 1)") {
					t.Errorf("Point should have type ID 1")
				}
				if !strings.Contains(code, "PutUint16(message[4:6], 2)") {
					t.Errorf("Color should have type ID 2")
				}
				if !strings.Contains(code, "PutUint16(message[4:6], 3)") {
					t.Errorf("Size should have type ID 3")
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
				if !strings.Contains(code, "func EncodeDataPacketMessage(") {
					t.Errorf("expected EncodeDataPacketMessage, check name conversion")
				}
				if !strings.Contains(code, "payload, err := EncodeDataPacket(src)") {
					t.Errorf("expected EncodeDataPacket call")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, err := GenerateMessageEncoders(tt.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateMessageEncoders() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkFunc != nil {
				tt.checkFunc(t, code)
			}
		})
	}
}
