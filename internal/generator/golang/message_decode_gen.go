package golang

import (
	"fmt"
	"strings"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// GenerateMessageDecoders generates DecodeXMessage functions for self-describing messages.
// Each function validates the 10-byte header and then decodes the payload.
func GenerateMessageDecoders(schema *parser.Schema) (string, error) {
	if schema == nil {
		return "", fmt.Errorf("schema is nil")
	}

	var buf strings.Builder

	// Generate a message decoder for each struct
	for i, s := range schema.Structs {
		typeID := uint16(i + 1) // Type IDs start at 1

		if err := generateMessageDecoder(&buf, &s, typeID); err != nil {
			return "", fmt.Errorf("struct %q: %w", s.Name, err)
		}

		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// generateMessageDecoder generates a DecodeXMessage function for a single struct.
func generateMessageDecoder(buf *strings.Builder, s *parser.Struct, typeID uint16) error {
	structName := ToGoName(s.Name)
	funcName := "Decode" + structName + "Message"
	decoderFunc := "Decode" + structName

	// Function doc comment
	buf.WriteString("// ")
	buf.WriteString(funcName)
	buf.WriteString(" decodes a ")
	buf.WriteString(structName)
	buf.WriteString(" from self-describing message format.\n")
	buf.WriteString("// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]\n")
	buf.WriteString("// Returns an error if the header is invalid or the payload cannot be decoded.\n")
	buf.WriteString("func ")
	buf.WriteString(funcName)
	buf.WriteString("(data []byte) (*")
	buf.WriteString(structName)
	buf.WriteString(", error) {\n")

	// Check minimum size
	buf.WriteString("\t// Check minimum message size\n")
	buf.WriteString("\tif len(data) < MessageHeaderSize {\n")
	buf.WriteString("\t\treturn nil, ErrUnexpectedEOF\n")
	buf.WriteString("\t}\n\n")

	// Validate magic bytes
	buf.WriteString("\t// Validate magic bytes\n")
	buf.WriteString("\tif string(data[0:3]) != MessageMagic {\n")
	buf.WriteString("\t\treturn nil, ErrInvalidMagic\n")
	buf.WriteString("\t}\n\n")

	// Validate version
	buf.WriteString("\t// Validate protocol version\n")
	buf.WriteString("\tif data[3] != MessageVersion {\n")
	buf.WriteString("\t\treturn nil, ErrInvalidVersion\n")
	buf.WriteString("\t}\n\n")

	// Validate type ID
	buf.WriteString("\t// Validate type ID\n")
	buf.WriteString("\ttypeID := binary.LittleEndian.Uint16(data[4:6])\n")
	buf.WriteString(fmt.Sprintf("\tif typeID != %d {\n", typeID))
	buf.WriteString("\t\treturn nil, ErrUnknownMessageType\n")
	buf.WriteString("\t}\n\n")

	// Extract payload length
	buf.WriteString("\t// Extract payload length\n")
	buf.WriteString("\tpayloadLength := binary.LittleEndian.Uint32(data[6:10])\n\n")

	// Validate total message size
	buf.WriteString("\t// Validate total message size\n")
	buf.WriteString("\texpectedSize := MessageHeaderSize + int(payloadLength)\n")
	buf.WriteString("\tif len(data) < expectedSize {\n")
	buf.WriteString("\t\treturn nil, ErrUnexpectedEOF\n")
	buf.WriteString("\t}\n\n")

	// Extract payload
	buf.WriteString("\t// Extract payload\n")
	buf.WriteString("\tpayload := data[MessageHeaderSize:expectedSize]\n\n")

	// Decode payload
	buf.WriteString("\t// Decode payload\n")
	buf.WriteString("\tvar result ")
	buf.WriteString(structName)
	buf.WriteString("\n")
	buf.WriteString("\tif err := ")
	buf.WriteString(decoderFunc)
	buf.WriteString("(&result, payload); err != nil {\n")
	buf.WriteString("\t\treturn nil, err\n")
	buf.WriteString("\t}\n\n")
	buf.WriteString("\treturn &result, nil\n")
	buf.WriteString("}\n")

	return nil
}

// GenerateMessageDispatcher generates a DecodeMessage function that dispatches
// to the appropriate decoder based on the type ID in the header.
func GenerateMessageDispatcher(schema *parser.Schema) (string, error) {
	if schema == nil {
		return "", fmt.Errorf("schema is nil")
	}

	var buf strings.Builder

	// Function doc comment
	buf.WriteString("// DecodeMessage decodes a message and returns the struct type based on the type ID in the header.\n")
	buf.WriteString("// This is the main entry point for decoding self-describing messages.\n")
	buf.WriteString("// Returns the decoded struct as an interface{} which can be type-asserted to the specific type.\n")
	buf.WriteString("func DecodeMessage(data []byte) (interface{}, error) {\n")

	// Check minimum size
	buf.WriteString("\t// Check minimum message size\n")
	buf.WriteString("\tif len(data) < MessageHeaderSize {\n")
	buf.WriteString("\t\treturn nil, ErrUnexpectedEOF\n")
	buf.WriteString("\t}\n\n")

	// Validate magic bytes
	buf.WriteString("\t// Validate magic bytes\n")
	buf.WriteString("\tif string(data[0:3]) != MessageMagic {\n")
	buf.WriteString("\t\treturn nil, ErrInvalidMagic\n")
	buf.WriteString("\t}\n\n")

	// Validate version
	buf.WriteString("\t// Validate protocol version\n")
	buf.WriteString("\tif data[3] != MessageVersion {\n")
	buf.WriteString("\t\treturn nil, ErrInvalidVersion\n")
	buf.WriteString("\t}\n\n")

	// Extract type ID
	buf.WriteString("\t// Extract type ID\n")
	buf.WriteString("\ttypeID := binary.LittleEndian.Uint16(data[4:6])\n\n")

	// Switch on type ID
	buf.WriteString("\t// Dispatch to specific decoder\n")
	buf.WriteString("\tswitch typeID {\n")

	// Generate case for each struct
	for i, s := range schema.Structs {
		typeID := i + 1
		structName := ToGoName(s.Name)
		decoderFunc := "Decode" + structName + "Message"

		buf.WriteString(fmt.Sprintf("\tcase %d:\n", typeID))
		buf.WriteString("\t\treturn ")
		buf.WriteString(decoderFunc)
		buf.WriteString("(data)\n")
	}

	// Default case for unknown type ID
	buf.WriteString("\tdefault:\n")
	buf.WriteString("\t\treturn nil, ErrUnknownMessageType\n")
	buf.WriteString("\t}\n")
	buf.WriteString("}\n")

	return buf.String(), nil
}
