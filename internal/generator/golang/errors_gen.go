package golang

import (
	"strings"
)

// GenerateErrors generates the error variable declarations for the decoder.
// These errors are used throughout the generated decode functions to signal
// various failure conditions as specified in DESIGN_SPEC.md Section 5.4.
func GenerateErrors() string {
	var buf strings.Builder

	// Message mode constants (RC Feature 2)
	buf.WriteString("// Message mode constants for self-describing messages\n")
	buf.WriteString("const (\n")
	buf.WriteString("\tMessageMagic         = \"SDP\"  // Magic bytes identifying SDP messages\n")
	buf.WriteString("\tMessageVersion  byte = '2'     // Protocol version 0.2.0\n")
	buf.WriteString("\tMessageHeaderSize    = 10      // Total header size: 3+1+2+4 bytes\n")
	buf.WriteString(")\n\n")

	// Error variables
	buf.WriteString("// Error variables for decode failures\n")
	buf.WriteString("var (\n")
	buf.WriteString("\tErrUnexpectedEOF      = errors.New(\"unexpected end of data\")\n")
	buf.WriteString("\tErrInvalidUTF8        = errors.New(\"invalid UTF-8 string\")\n")
	buf.WriteString("\tErrDataTooLarge       = errors.New(\"data exceeds 128MB limit\")\n")
	buf.WriteString("\tErrArrayTooLarge      = errors.New(\"array count exceeds per-array limit\")\n")
	buf.WriteString("\tErrTooManyElements    = errors.New(\"total elements exceed limit\")\n")
	buf.WriteString("\tErrInvalidData        = errors.New(\"invalid or corrupted data\")\n")
	buf.WriteString("\tErrInvalidMagic       = errors.New(\"invalid magic bytes (expected 'SDP')\")\n")
	buf.WriteString("\tErrInvalidVersion     = errors.New(\"unsupported protocol version\")\n")
	buf.WriteString("\tErrUnknownMessageType = errors.New(\"unknown message type ID\")\n")
	buf.WriteString(")\n")

	return buf.String()
}
