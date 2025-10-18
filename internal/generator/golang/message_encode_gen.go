package golang

import (
	"fmt"
	"strings"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// GenerateMessageEncoders generates EncodeXMessage functions for self-describing messages.
// Each function adds a 10-byte header: [magic:3][version:1][type_id:2][length:4][payload:N]
func GenerateMessageEncoders(schema *parser.Schema) (string, error) {
	if schema == nil {
		return "", fmt.Errorf("schema is nil")
	}

	var buf strings.Builder

	// Generate a message encoder for each struct
	for i, s := range schema.Structs {
		typeID := uint16(i + 1) // Type IDs start at 1

		if err := generateMessageEncoder(&buf, &s, typeID); err != nil {
			return "", fmt.Errorf("struct %q: %w", s.Name, err)
		}

		buf.WriteString("\n")
	}

	return buf.String(), nil
}

// generateMessageEncoder generates an EncodeXMessage function for a single struct.
func generateMessageEncoder(buf *strings.Builder, s *parser.Struct, typeID uint16) error {
	structName := ToGoName(s.Name)
	funcName := "Encode" + structName + "Message"
	encoderFunc := "Encode" + structName

	// Function doc comment
	buf.WriteString("// ")
	buf.WriteString(funcName)
	buf.WriteString(" encodes a ")
	buf.WriteString(structName)
	buf.WriteString(" to self-describing message format.\n")
	buf.WriteString("// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]\n")
	buf.WriteString("// This format is suitable for persistence, network transmission, and cross-service communication.\n")
	buf.WriteString("func ")
	buf.WriteString(funcName)
	buf.WriteString("(src *")
	buf.WriteString(structName)
	buf.WriteString(") ([]byte, error) {\n")

	// Encode payload (without header)
	buf.WriteString("\t// Encode payload\n")
	buf.WriteString("\tpayload, err := ")
	buf.WriteString(encoderFunc)
	buf.WriteString("(src)\n")
	buf.WriteString("\tif err != nil {\n")
	buf.WriteString("\t\treturn nil, err\n")
	buf.WriteString("\t}\n\n")

	// Calculate total message size
	buf.WriteString("\t// Allocate message buffer (header + payload)\n")
	buf.WriteString("\tmessageSize := MessageHeaderSize + len(payload)\n")
	buf.WriteString("\tmessage := make([]byte, messageSize)\n\n")

	// Write header
	buf.WriteString("\t// Write header\n")
	buf.WriteString("\tcopy(message[0:3], MessageMagic)  // Magic bytes 'SDP'\n")
	buf.WriteString("\tmessage[3] = MessageVersion       // Protocol version '2'\n")
	buf.WriteString(fmt.Sprintf("\tbinary.LittleEndian.PutUint16(message[4:6], %d)  // Type ID\n", typeID))
	buf.WriteString("\tbinary.LittleEndian.PutUint32(message[6:10], uint32(len(payload)))  // Payload length\n\n")

	// Copy payload
	buf.WriteString("\t// Copy payload\n")
	buf.WriteString("\tcopy(message[10:], payload)\n\n")

	buf.WriteString("\treturn message, nil\n")
	buf.WriteString("}\n")

	return nil
}
