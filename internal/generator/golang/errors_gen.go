package golang

import (
	"strings"
)

// GenerateErrors generates the error variable declarations for the decoder.
// These errors are used throughout the generated decode functions to signal
// various failure conditions as specified in DESIGN_SPEC.md Section 5.4.
func GenerateErrors() string {
	var buf strings.Builder

	buf.WriteString("// Error variables for decode failures\n")
	buf.WriteString("var (\n")
	buf.WriteString("\tErrUnexpectedEOF      = errors.New(\"unexpected end of data\")\n")
	buf.WriteString("\tErrInvalidUTF8        = errors.New(\"invalid UTF-8 string\")\n")
	buf.WriteString("\tErrDataTooLarge       = errors.New(\"data exceeds 128MB limit\")\n")
	buf.WriteString("\tErrArrayTooLarge      = errors.New(\"array count exceeds per-array limit\")\n")
	buf.WriteString("\tErrTooManyElements    = errors.New(\"total elements exceed limit\")\n")
	buf.WriteString("\tErrInvalidData        = errors.New(\"invalid or corrupted data\")\n")
	buf.WriteString(")\n")

	return buf.String()
}
