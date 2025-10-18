package golang

import (
	"fmt"
	"strings"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// GenerateReaderDecoder generates DecodeXFromReader functions for each struct in the schema.
// These functions enable streaming I/O by reading directly from io.Reader interfaces.
//
// Design Philosophy:
//   - Provide stdlib stream interfaces (io.Reader), NOT baked-in decompression
//   - Users compose with their preferred libraries (gzip, files, network, etc.)
//   - Zero dependencies in generated code
//   - Language-idiomatic Go pattern (same as encoding/json)
//
// For each struct type, it generates:
//   - DecodeStructNameFromReader(dest *StructName, r io.Reader) error
//
// The decoder:
//  1. Reads all bytes from io.Reader using io.ReadAll
//  2. Decodes from buffer (reuses existing DecodeX function)
//
// Example output:
//
//	func DecodeDeviceFromReader(dest *Device, r io.Reader) error {
//	    buf, err := io.ReadAll(r)
//	    if err != nil {
//	        return err
//	    }
//	    return DecodeDevice(dest, buf)
//	}
//
// Usage examples (user composition):
//
//	// File I/O
//	f, _ := os.Open("device.sdp")
//	defer f.Close()
//	var device Device
//	DecodeDeviceFromReader(&device, f)
//
//	// Decompression (user composes with gzip)
//	gzReader, _ := gzip.NewReader(bytes.NewReader(compressed))
//	defer gzReader.Close()
//	var device Device
//	DecodeDeviceFromReader(&device, gzReader)
//
//	// Network
//	conn, _ := listener.Accept()
//	var device Device
//	DecodeDeviceFromReader(&device, conn)
func GenerateReaderDecoder(schema *parser.Schema) (string, error) {
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
		readerFunc := "Decode" + structName + "FromReader"
		decodeFunc := "Decode" + structName

		if err := generateReaderDecoderFunction(&buf, structName, readerFunc, decodeFunc); err != nil {
			return "", err
		}
	}

	return buf.String(), nil
}

// generateReaderDecoderFunction generates the reader-based decoder function.
func generateReaderDecoderFunction(buf *strings.Builder, structName, funcName, decodeFunc string) error {
	// Doc comment
	buf.WriteString("// ")
	buf.WriteString(funcName)
	buf.WriteString(" decodes a ")
	buf.WriteString(structName)
	buf.WriteString(" from wire format by reading from the provided io.Reader.\n")
	buf.WriteString("// This enables streaming I/O without baked-in decompression.\n")
	buf.WriteString("//\n")
	buf.WriteString("// Users can compose with any io.Reader implementation:\n")
	buf.WriteString("//   - File I/O: os.File\n")
	buf.WriteString("//   - Decompression: gzip.Reader, zstd.Reader, etc.\n")
	buf.WriteString("//   - Network: net.Conn, http.Request.Body\n")
	buf.WriteString("//   - Decryption: custom crypto.Reader\n")
	buf.WriteString("//   - Metrics: custom byte counting wrappers\n")
	buf.WriteString("func ")
	buf.WriteString(funcName)
	buf.WriteString("(dest *")
	buf.WriteString(structName)
	buf.WriteString(", r io.Reader) error {\n")

	// Read all bytes from io.Reader
	buf.WriteString("\tbuf, err := io.ReadAll(r)\n")
	buf.WriteString("\tif err != nil {\n")
	buf.WriteString("\t\treturn err\n")
	buf.WriteString("\t}\n")

	// Decode from buffer (reuse existing function)
	buf.WriteString("\treturn ")
	buf.WriteString(decodeFunc)
	buf.WriteString("(dest, buf)\n")
	buf.WriteString("}\n")

	return nil
}
