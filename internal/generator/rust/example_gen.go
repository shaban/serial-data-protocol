package rust

import (
	"fmt"
	"strings"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// GenerateExample generates a crossplatform_helper.rs example that can be used
// for testing cross-language compatibility
func GenerateExample(schema *parser.Schema, packageName string) string {
	var b strings.Builder

	b.WriteString("//! Cross-platform test helper for " + packageName + "\n")
	b.WriteString("//! Usage:\n")
	
	// Generate usage examples for each struct
	for _, st := range schema.Structs {
		structName := toSnakeCase(st.Name)
		b.WriteString(fmt.Sprintf("//!   cargo run --release --example crossplatform_helper encode-%s > output.bin\n", structName))
		b.WriteString(fmt.Sprintf("//!   cargo run --release --example crossplatform_helper decode-%s input.bin\n", structName))
	}
	
	b.WriteString("\n")
	b.WriteString("use std::env;\n")
	b.WriteString("use std::fs;\n")
	b.WriteString("use std::io::{self, Write};\n")
	b.WriteString("\n")
	// Use the crate name (sdp-xxx) not the snake_case package name
	crateName := fmt.Sprintf("sdp_%s", toSnakeCase(packageName))
	b.WriteString(fmt.Sprintf("use %s::*;\n", crateName))
	b.WriteString("\n")
	b.WriteString("// Type alias to avoid conflict with wire format Result\n")
	b.WriteString("type StdResult<T> = std::result::Result<T, Box<dyn std::error::Error>>;\n")
	b.WriteString("\n")

	// Generate encode/decode functions for each struct
	for _, st := range schema.Structs {
		generateEncodeFunction(&b, &st)
		b.WriteString("\n")
		generateDecodeFunction(&b, &st)
		b.WriteString("\n")
	}	// Generate main function
	generateMainFunction(&b, schema)

	return b.String()
}

func generateEncodeFunction(b *strings.Builder, st *parser.Struct) {
	structName := st.Name
	funcName := fmt.Sprintf("encode_%s", toSnakeCase(st.Name))

	b.WriteString(fmt.Sprintf("fn %s() -> StdResult<()> {\n", funcName))
	b.WriteString(fmt.Sprintf("    let data = %s {\n", structName))

	// Generate test data for each field
	for _, field := range st.Fields {
		fieldName := toSnakeCase(field.Name)
		b.WriteString(fmt.Sprintf("        %s: %s,\n", fieldName, generateTestValue(&field)))
	}

	b.WriteString("    };\n")
	b.WriteString("\n")
	b.WriteString("    let mut buf = vec![0u8; data.encoded_size()];\n")
	b.WriteString("    data.encode_to_slice(&mut buf)?;\n")
	b.WriteString("\n")
	b.WriteString("    io::stdout().write_all(&buf)?;\n")
	b.WriteString("    Ok(())\n")
	b.WriteString("}\n")
}

func generateDecodeFunction(b *strings.Builder, st *parser.Struct) {
	structName := st.Name
	funcName := fmt.Sprintf("decode_%s", toSnakeCase(st.Name))

	b.WriteString(fmt.Sprintf("fn %s(filename: &str) -> StdResult<()> {\n", funcName))
	b.WriteString("    let file_data = fs::read(filename)?;\n")
	b.WriteString(fmt.Sprintf("    let decoded = %s::decode_from_slice(&file_data)?;\n", structName))
	b.WriteString("\n")
	b.WriteString("    // Basic validation\n")
	b.WriteString("    let mut ok = true;\n")

	// Generate validation for each field
	for _, field := range st.Fields {
		fieldName := toSnakeCase(field.Name)
		validation := generateValidation(&field, "decoded."+fieldName)
		if validation != "" {
			b.WriteString(fmt.Sprintf("    ok = ok && %s;\n", validation))
		}
	}

	b.WriteString("\n")
	b.WriteString("    if !ok {\n")
	b.WriteString("        eprintln!(\"Validation failed\");\n")
	b.WriteString("        eprintln!(\"Decoded: {:?}\", decoded);\n")
	b.WriteString("        std::process::exit(1);\n")
	b.WriteString("    }\n")
	b.WriteString("\n")
	b.WriteString("    eprintln!(\"✓ Rust successfully decoded and validated\");\n")
	b.WriteString("    Ok(())\n")
	b.WriteString("}\n")
}

func generateMainFunction(b *strings.Builder, schema *parser.Schema) {
	b.WriteString("fn main() -> StdResult<()> {\n")
	b.WriteString("    let args: Vec<String> = env::args().collect();\n")
	b.WriteString("\n")
	b.WriteString("    if args.len() < 2 {\n")
	b.WriteString("        eprintln!(\"Usage: {} <command> [args]\", args[0]);\n")
	b.WriteString("        eprintln!(\"Commands:\");\n")

	// List all available commands
	for _, st := range schema.Structs {
		structName := toSnakeCase(st.Name)
		b.WriteString(fmt.Sprintf("        eprintln!(\"  encode-%s - Encode %s to stdout\");\n", structName, st.Name))
		b.WriteString(fmt.Sprintf("        eprintln!(\"  decode-%s <file> - Decode %s from file\");\n", structName, st.Name))
	}

	b.WriteString("        std::process::exit(1);\n")
	b.WriteString("    }\n")
	b.WriteString("\n")
	b.WriteString("    match args[1].as_str() {\n")

	// Generate match arms for each struct
	for _, st := range schema.Structs {
		structName := toSnakeCase(st.Name)
		b.WriteString(fmt.Sprintf("        \"encode-%s\" => encode_%s()?,\n", structName, structName))
		b.WriteString(fmt.Sprintf("        \"decode-%s\" => {\n", structName))
		b.WriteString("            if args.len() < 3 {\n")
		b.WriteString(fmt.Sprintf("                eprintln!(\"Error: decode-%s requires filename argument\");\n", structName))
		b.WriteString("                std::process::exit(1);\n")
		b.WriteString("            }\n")
		b.WriteString(fmt.Sprintf("            decode_%s(&args[2])?;\n", structName))
		b.WriteString("        }\n")
	}

	b.WriteString("        cmd => {\n")
	b.WriteString("            eprintln!(\"Unknown command: {}\", cmd);\n")
	b.WriteString("            std::process::exit(1);\n")
	b.WriteString("        }\n")
	b.WriteString("    }\n")
	b.WriteString("\n")
	b.WriteString("    Ok(())\n")
	b.WriteString("}\n")
}

func generateTestValue(field *parser.Field) string {
	if field.Type.Optional {
		// For optional fields, generate Some(value)
		baseValue := generateBaseTestValue(&field.Type)
		return fmt.Sprintf("Some(%s)", baseValue)
	}
	return generateBaseTestValue(&field.Type)
}

func generateBaseTestValue(typeExpr *parser.TypeExpr) string {
	if typeExpr.Kind == parser.TypeKindArray && typeExpr.Elem != nil {
		// Generate array literal
		elemValue := generateScalarTestValue(typeExpr.Elem.Name)
		// For dynamic arrays, use vec!
		return fmt.Sprintf("vec![%s, %s, %s]", elemValue, elemValue, elemValue)
	}
	return generateScalarTestValue(typeExpr.Name)
}

func generateScalarTestValue(fieldType string) string {
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
		return "\"Hello from Rust!\".to_string()"
	default:
		// For nested types, use Default::default()
		return "Default::default()"
	}
}

func generateValidation(field *parser.Field, accessor string) string {
	if field.Type.Optional {
		// For optional fields, check if Some and validate inner value
		innerValidation := generateBaseValidation(&field.Type, accessor+".as_ref().unwrap()")
		if innerValidation == "" {
			return ""
		}
		return fmt.Sprintf("%s.is_some() && %s", accessor, innerValidation)
	}
	return generateBaseValidation(&field.Type, accessor)
}

func generateBaseValidation(typeExpr *parser.TypeExpr, accessor string) string {
	if typeExpr.Kind == parser.TypeKindArray {
		// For arrays, just check that it's not empty
		return fmt.Sprintf("%s.len() > 0", accessor)
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
		return fmt.Sprintf("(%s - 3.14159).abs() < 0.0001", accessor)
	case "f64":
		return fmt.Sprintf("(%s - 2.718281828459045).abs() < 0.0000001", accessor)
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
