package golang

import (
	"strings"
)

// GenerateDecodeContext generates the DecodeContext type and related constants.
// The DecodeContext is used during decoding to track array element counts and
// enforce size limits as specified in DESIGN_SPEC.md Section 5.5.
func GenerateDecodeContext() string {
	var buf strings.Builder

	// Generate constants
	buf.WriteString("// Size limit constants for decode validation\n")
	buf.WriteString("const (\n")
	buf.WriteString("\tMaxSerializedSize = 128 * 1024 * 1024\n")
	buf.WriteString("\tMaxArrayElements  = 1_000_000\n")
	buf.WriteString("\tMaxTotalElements  = 10_000_000\n")
	buf.WriteString(")\n\n")

	// Generate DecodeContext type
	buf.WriteString("// DecodeContext tracks state during decoding to enforce size limits.\n")
	buf.WriteString("// It maintains a count of total elements across all arrays to prevent\n")
	buf.WriteString("// excessive memory allocation from malicious or corrupted data.\n")
	buf.WriteString("type DecodeContext struct {\n")
	buf.WriteString("\ttotalElements int\n")
	buf.WriteString("}\n\n")

	// Generate checkArraySize method
	buf.WriteString("// checkArraySize validates an array count against per-array and total limits.\n")
	buf.WriteString("// It returns ErrArrayTooLarge if the count exceeds MaxArrayElements, or\n")
	buf.WriteString("// ErrTooManyElements if the cumulative total exceeds MaxTotalElements.\n")
	buf.WriteString("func (ctx *DecodeContext) checkArraySize(count uint32) error {\n")
	buf.WriteString("\tif count > MaxArrayElements {\n")
	buf.WriteString("\t\treturn ErrArrayTooLarge\n")
	buf.WriteString("\t}\n\n")
	buf.WriteString("\tctx.totalElements += int(count)\n")
	buf.WriteString("\tif ctx.totalElements > MaxTotalElements {\n")
	buf.WriteString("\t\treturn ErrTooManyElements\n")
	buf.WriteString("\t}\n\n")
	buf.WriteString("\treturn nil\n")
	buf.WriteString("}\n")

	return buf.String()
}
