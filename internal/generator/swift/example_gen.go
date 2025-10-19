package swift

import (
	"fmt"
	"strings"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// GenerateExample generates a crossplatform_helper executable that can be used
// for testing cross-language compatibility
func GenerateExample(schema *parser.Schema, packageName string) string {
	var b strings.Builder

	b.WriteString("// Cross-platform test helper for " + packageName + "\n")
	b.WriteString("// Usage:\n")

	// Generate usage examples for each struct
	for _, st := range schema.Structs {
		structName := toCamelCase(st.Name)
		b.WriteString(fmt.Sprintf("//   swift run -c release crossplatform_helper encode-%s > output.bin\n", structName))
		b.WriteString(fmt.Sprintf("//   swift run -c release crossplatform_helper decode-%s input.bin\n", structName))
	}

	b.WriteString("\n")
	b.WriteString("import Foundation\n")
	b.WriteString("import " + packageName + "\n")
	b.WriteString("\n")

	// Generate encode/decode functions for each struct
	for _, st := range schema.Structs {
		generateEncodeFunction(&b, &st)
		b.WriteString("\n")
		generateDecodeFunction(&b, &st)
		b.WriteString("\n")
	}

	// Generate main function
	generateMainFunction(&b, schema)

	return b.String()
}

func generateEncodeFunction(b *strings.Builder, st *parser.Struct) {
	structName := st.Name
	funcName := fmt.Sprintf("encode%s", toCamelCase(st.Name))

	b.WriteString(fmt.Sprintf("func %s() throws {\n", funcName))
	b.WriteString(fmt.Sprintf("    let data = %s(\n", structName))

	// Generate test data for each field
	for i, field := range st.Fields {
		fieldName := lowerCamelCase(field.Name)
		comma := ","
		if i == len(st.Fields)-1 {
			comma = ""
		}
		b.WriteString(fmt.Sprintf("        %s: %s%s\n", fieldName, generateTestValue(&field), comma))
	}

	b.WriteString("    )\n")
	b.WriteString("\n")
	b.WriteString("    let bytes = data.encodeToBytes()\n")
	b.WriteString("    let binaryData = Data(bytes)\n")
	b.WriteString("\n")
	b.WriteString("    FileHandle.standardOutput.write(binaryData)\n")
	b.WriteString("}\n")
}

func generateDecodeFunction(b *strings.Builder, st *parser.Struct) {
	structName := st.Name
	funcName := fmt.Sprintf("decode%s", toCamelCase(st.Name))

	b.WriteString(fmt.Sprintf("func %s(filename: String) throws {\n", funcName))
	b.WriteString("    let fileData = try Data(contentsOf: URL(fileURLWithPath: filename))\n")
	b.WriteString("    let bytes = [UInt8](fileData)\n")
	b.WriteString(fmt.Sprintf("    let decoded = try %s.decode(from: bytes)\n", structName))
	b.WriteString("\n")
	b.WriteString("    // Basic validation\n")
	b.WriteString("    var ok = true\n")

	// Generate validation for each field
	for _, field := range st.Fields {
		fieldName := lowerCamelCase(field.Name)
		validation := generateValidation(&field, "decoded."+fieldName)
		if validation != "" {
			b.WriteString(fmt.Sprintf("    ok = ok && %s\n", validation))
		}
	}

	b.WriteString("\n")
	b.WriteString("    if !ok {\n")
	b.WriteString("        fputs(\"Validation failed\\n\", stderr)\n")
	b.WriteString("        fputs(\"Decoded: \\(decoded)\\n\", stderr)\n")
	b.WriteString("        exit(1)\n")
	b.WriteString("    }\n")
	b.WriteString("\n")
	b.WriteString("    fputs(\"✓ Swift successfully decoded and validated\\n\", stderr)\n")
	b.WriteString("}\n")
}

func generateMainFunction(b *strings.Builder, schema *parser.Schema) {
	b.WriteString("// Main entry point\n")
	b.WriteString("let args = CommandLine.arguments\n")
	b.WriteString("\n")
	b.WriteString("if args.count < 2 {\n")
	b.WriteString("    fputs(\"Usage: \\(args[0]) <command> [args]\\n\", stderr)\n")
	b.WriteString("    fputs(\"Commands:\\n\", stderr)\n")

	// List all available commands
	for _, st := range schema.Structs {
		structName := toCamelCase(st.Name)
		b.WriteString(fmt.Sprintf("    fputs(\"  encode-%s - Encode %s to stdout\\n\", stderr)\n", structName, st.Name))
		b.WriteString(fmt.Sprintf("    fputs(\"  decode-%s <file> - Decode %s from file\\n\", stderr)\n", structName, st.Name))
	}

	b.WriteString("    exit(1)\n")
	b.WriteString("}\n")
	b.WriteString("\n")
	b.WriteString("let command = args[1]\n")
	b.WriteString("\n")
	b.WriteString("do {\n")
	b.WriteString("    switch command {\n")

	// Generate switch cases for each struct
	for _, st := range schema.Structs {
		structName := toCamelCase(st.Name)
		b.WriteString(fmt.Sprintf("    case \"encode-%s\":\n", structName))
		b.WriteString(fmt.Sprintf("        try encode%s()\n", structName))
		b.WriteString(fmt.Sprintf("    case \"decode-%s\":\n", structName))
		b.WriteString("        guard args.count >= 3 else {\n")
		b.WriteString(fmt.Sprintf("            fputs(\"Error: decode-%s requires filename argument\\n\", stderr)\n", structName))
		b.WriteString("            exit(1)\n")
		b.WriteString("        }\n")
		b.WriteString(fmt.Sprintf("        try decode%s(filename: args[2])\n", structName))
	}

	b.WriteString("    default:\n")
	b.WriteString("        fputs(\"Unknown command: \\(command)\\n\", stderr)\n")
	b.WriteString("        exit(1)\n")
	b.WriteString("    }\n")
	b.WriteString("} catch {\n")
	b.WriteString("    fputs(\"Error: \\(error)\\n\", stderr)\n")
	b.WriteString("    exit(1)\n")
	b.WriteString("}\n")
}

func generateTestValue(field *parser.Field) string {
	if field.Type.Optional {
		// For optional fields, generate .some(value)
		baseValue := generateBaseTestValue(&field.Type)
		return fmt.Sprintf(".some(%s)", baseValue)
	}
	return generateBaseTestValue(&field.Type)
}

func generateBaseTestValue(typeExpr *parser.TypeExpr) string {
	if typeExpr.Kind == parser.TypeKindArray && typeExpr.Elem != nil {
		// Generate array literal
		elemValue := generateScalarTestValue(typeExpr.Elem.Name)
		return fmt.Sprintf("ContiguousArray([%s, %s, %s])", elemValue, elemValue, elemValue)
	}
	return generateScalarTestValue(typeExpr.Name)
}

func generateScalarTestValue(fieldType string) string {
	swiftType := MapTypeToSwift(&parser.TypeExpr{Name: fieldType})

	switch fieldType {
	case "u8":
		return "255"
	case "u16":
		return "65535"
	case "u32":
		return "4_294_967_295"
	case "u64":
		return "18_446_744_073_709_551_615"
	case "i8":
		return "-128"
	case "i16":
		return "-32768"
	case "i32":
		return "-2_147_483_648"
	case "i64":
		return "-9_223_372_036_854_775_808"
	case "f32":
		return "3.14159"
	case "f64":
		return "2.718281828459045"
	case "bool":
		return "true"
	case "string", "str":
		return "\"Hello from Swift!\""
	default:
		// For nested types, use default initializer
		return fmt.Sprintf("%s()", swiftType)
	}
}

func generateValidation(field *parser.Field, accessor string) string {
	if field.Type.Optional {
		// For optional fields, check if some and validate inner value
		innerValidation := generateBaseValidation(&field.Type, accessor+"!")
		if innerValidation == "" {
			return ""
		}
		return fmt.Sprintf("%s != nil && %s", accessor, innerValidation)
	}
	return generateBaseValidation(&field.Type, accessor)
}

func generateBaseValidation(typeExpr *parser.TypeExpr, accessor string) string {
	if typeExpr.Kind == parser.TypeKindArray {
		// For arrays, just check that it's not empty
		return fmt.Sprintf("%s.count > 0", accessor)
	}

	switch typeExpr.Name {
	case "u8":
		return fmt.Sprintf("%s == 255", accessor)
	case "u16":
		return fmt.Sprintf("%s == 65535", accessor)
	case "u32":
		return fmt.Sprintf("%s == 4_294_967_295", accessor)
	case "u64":
		return fmt.Sprintf("%s == 18_446_744_073_709_551_615", accessor)
	case "i8":
		return fmt.Sprintf("%s == -128", accessor)
	case "i16":
		return fmt.Sprintf("%s == -32768", accessor)
	case "i32":
		return fmt.Sprintf("%s == -2_147_483_648", accessor)
	case "i64":
		return fmt.Sprintf("%s == -9_223_372_036_854_775_808", accessor)
	case "f32":
		return fmt.Sprintf("abs(%s - 3.14159) < 0.0001", accessor)
	case "f64":
		return fmt.Sprintf("abs(%s - 2.718281828459045) < 0.0000001", accessor)
	case "bool":
		return fmt.Sprintf("%s == true", accessor)
	case "string":
		// Accept from any language
		return fmt.Sprintf("(%s == \"Hello from Go!\" || %s == \"Hello from Swift!\" || %s == \"Hello from Rust!\")",
			accessor, accessor, accessor)
	default:
		// For complex types, skip validation for now
		return ""
	}
}

// toCamelCase converts snake_case to CamelCase
func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[0:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}

// lowerCamelCase converts snake_case to camelCase (lowercase first letter)
func lowerCamelCase(s string) string {
	camel := toCamelCase(s)
	if len(camel) > 0 {
		return strings.ToLower(camel[0:1]) + camel[1:]
	}
	return camel
}
