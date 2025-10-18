package golang

import (
	"fmt"
	"strings"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// GenerateWriterEncoder generates EncodeXToWriter functions for each struct in the schema.
// These functions enable streaming I/O by writing directly to io.Writer interfaces.
//
// Design Philosophy:
//   - Provide stdlib stream interfaces (io.Writer), NOT baked-in compression
//   - Users compose with their preferred libraries (gzip, files, network, etc.)
//   - Zero dependencies in generated code
//   - Language-idiomatic Go pattern (same as encoding/json)
//
// For each struct type, it generates:
//   - EncodeStructNameToWriter(src *StructName, w io.Writer) error
//
// The encoder:
//  1. Calculates exact buffer size (reuses existing calculateXSize function)
//  2. Allocates buffer once
//  3. Encodes to buffer (reuses existing encodeX helper)
//  4. Writes buffer to io.Writer
//
// Example output:
//
//	func EncodeDeviceToWriter(src *Device, w io.Writer) error {
//	    size := calculateDeviceSize(src)
//	    buf := make([]byte, size)
//	    offset := 0
//	    if err := encodeDevice(src, buf, &offset); err != nil {
//	        return err
//	    }
//	    _, err := w.Write(buf)
//	    return err
//	}
//
// Usage examples (user composition):
//
//	// File I/O
//	f, _ := os.Create("device.sdp")
//	defer f.Close()
//	EncodeDeviceToWriter(&device, f)
//
//	// Compression (user composes with gzip)
//	var buf bytes.Buffer
//	gzWriter := gzip.NewWriter(&buf)
//	EncodeDeviceToWriter(&device, gzWriter)
//	gzWriter.Close()
//
//	// Network
//	conn, _ := net.Dial("tcp", "server:8080")
//	EncodeDeviceToWriter(&device, conn)
func GenerateWriterEncoder(schema *parser.Schema) (string, error) {
	if schema == nil {
		return "", fmt.Errorf("schema is nil")
	}

	if len(schema.Structs) == 0 {
		return "", fmt.Errorf("schema has no structs")
	}

	var buf strings.Builder

	for i, s := range schema.Structs {
		// Add blank line between functions (except before first)
		if i > 0 {
			buf.WriteString("\n")
		}

		structName := ToGoName(s.Name)
		writerFunc := "Encode" + structName + "ToWriter"
		sizeFunc := "calculate" + structName + "Size"
		helperFunc := "encode" + structName

		if err := generateWriterEncoderFunction(&buf, structName, writerFunc, sizeFunc, helperFunc); err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}

// generateWriterEncoderFunction generates the writer-based encoder function.
func generateWriterEncoderFunction(buf *strings.Builder, structName, funcName, sizeFunc, helperFunc string) error {
	// Doc comment
	buf.WriteString("// ")
	buf.WriteString(funcName)
	buf.WriteString(" encodes a ")
	buf.WriteString(structName)
	buf.WriteString(" to wire format and writes it to the provided io.Writer.\n")
	buf.WriteString("// This enables streaming I/O without baked-in compression.\n")
	buf.WriteString("//\n")
	buf.WriteString("// Users can compose with any io.Writer implementation:\n")
	buf.WriteString("//   - File I/O: os.File\n")
	buf.WriteString("//   - Compression: gzip.Writer, zstd.Writer, etc.\n")
	buf.WriteString("//   - Network: net.Conn, http.ResponseWriter\n")
	buf.WriteString("//   - Encryption: custom crypto.Writer\n")
	buf.WriteString("//   - Metrics: custom byte counting wrappers\n")
	buf.WriteString("func ")
	buf.WriteString(funcName)
	buf.WriteString("(src *")
	buf.WriteString(structName)
	buf.WriteString(", w io.Writer) error {\n")

	// Calculate size (reuse existing function)
	buf.WriteString("\tsize := ")
	buf.WriteString(sizeFunc)
	buf.WriteString("(src)\n")

	// Allocate buffer
	buf.WriteString("\tbuf := make([]byte, size)\n")
	buf.WriteString("\toffset := 0\n")

	// Encode to buffer (reuse existing helper)
	buf.WriteString("\tif err := ")
	buf.WriteString(helperFunc)
	buf.WriteString("(src, buf, &offset); err != nil {\n")
	buf.WriteString("\t\treturn err\n")
	buf.WriteString("\t}\n")

	// Write to io.Writer
	buf.WriteString("\t_, err := w.Write(buf)\n")
	buf.WriteString("\treturn err\n")
	buf.WriteString("}\n")

	return nil
}
