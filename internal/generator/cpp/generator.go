package cpp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// Generate creates all C++ files for a schema
func Generate(schema *parser.Schema, outputDir string, verbose bool) error {
	if schema == nil {
		return fmt.Errorf("schema is nil")
	}

	if outputDir == "" {
		return fmt.Errorf("output directory is empty")
	}

	// Create output directory
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Determine package name from schema or directory
	packageName := filepath.Base(outputDir)
	if packageName == "." || packageName == "/" {
		packageName = "sdp"
	}

	// Generate types.hpp
	if err := generateFile(outputDir, "types.hpp", GenerateTypes(schema, packageName), verbose); err != nil {
		return err
	}

	// Generate encode.hpp
	if err := generateFile(outputDir, "encode.hpp", GenerateEncodeHeader(schema, packageName), verbose); err != nil {
		return err
	}

	// Generate encode.cpp
	if err := generateFile(outputDir, "encode.cpp", GenerateEncodeImpl(schema, packageName), verbose); err != nil {
		return err
	}

	// Generate decode.hpp
	if err := generateFile(outputDir, "decode.hpp", GenerateDecodeHeader(schema, packageName), verbose); err != nil {
		return err
	}

	// Generate decode.cpp
	if err := generateFile(outputDir, "decode.cpp", GenerateDecodeImpl(schema, packageName), verbose); err != nil {
		return err
	}

	// Generate message_encode.hpp
	if err := generateFile(outputDir, "message_encode.hpp", GenerateMessageEncodeHeader(schema, packageName), verbose); err != nil {
		return err
	}

	// Generate message_encode.cpp
	if err := generateFile(outputDir, "message_encode.cpp", GenerateMessageEncodeImpl(schema, packageName), verbose); err != nil {
		return err
	}

	// Generate message_decode.hpp
	if err := generateFile(outputDir, "message_decode.hpp", GenerateMessageDecodeHeader(schema, packageName), verbose); err != nil {
		return err
	}

	// Generate message_decode.cpp
	if err := generateFile(outputDir, "message_decode.cpp", GenerateMessageDecodeImpl(schema, packageName), verbose); err != nil {
		return err
	}

	// Generate endian.hpp
	if err := generateFile(outputDir, "endian.hpp", GenerateEndianHeader(), verbose); err != nil {
		return err
	}

	// Generate CMakeLists.txt
	if err := generateFile(outputDir, "CMakeLists.txt", GenerateCMake(schema, packageName), verbose); err != nil {
		return err
	}

	return nil
}

func generateFile(outputDir, filename, content string, verbose bool) error {
	path := filepath.Join(outputDir, filename)

	if verbose {
		fmt.Printf("Generating %s...\n", path)
	}

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", filename, err)
	}

	return nil
}

// Helper functions (copied from C generator)

func toPascalCase(s string) string {
	parts := strings.Split(s, "_")
	result := ""
	for _, part := range parts {
		if len(part) > 0 {
			result += strings.ToUpper(part[0:1]) + part[1:]
		}
	}
	return result
}

func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	result := parts[0]
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			result += strings.ToUpper(parts[i][0:1]) + parts[i][1:]
		}
	}
	return result
}

func getCppType(typeName string) string {
	switch typeName {
	case "u8":
		return "uint8_t"
	case "u16":
		return "uint16_t"
	case "u32":
		return "uint32_t"
	case "u64":
		return "uint64_t"
	case "i8":
		return "int8_t"
	case "i16":
		return "int16_t"
	case "i32":
		return "int32_t"
	case "i64":
		return "int64_t"
	case "f32":
		return "float"
	case "f64":
		return "double"
	case "bool":
		return "bool"
	case "str":
		return "std::string"
	default:
		return toPascalCase(typeName)
	}
}

func getPrimitiveSize(typeName string) int {
	switch typeName {
	case "u8", "i8", "bool":
		return 1
	case "u16", "i16":
		return 2
	case "u32", "i32", "f32":
		return 4
	case "u64", "i64", "f64":
		return 8
	default:
		return 0
	}
}
