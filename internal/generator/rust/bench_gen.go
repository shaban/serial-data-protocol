package rust

import (
	"strings"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// GenerateBenchmark generates a Criterion benchmark file for a schema
func GenerateBenchmark(schema *parser.Schema, packageName string) string {
	var b strings.Builder

	// Construct proper crate name (sdp_xxx format)
	crateName := "sdp_" + toSnakeCase(packageName)

	// File header
	b.WriteString("//! Criterion benchmarks for " + packageName + "\n")
	b.WriteString("//!\n")
	b.WriteString("//! Run with: cargo bench\n")
	b.WriteString("//! View results: target/criterion/report/index.html\n\n")

	// Imports
	b.WriteString("use criterion::{black_box, criterion_group, criterion_main, Criterion};\n")
	b.WriteString("use " + crateName + "::*;\n\n")

	// Generate benchmark function for each struct
	for _, st := range schema.Structs {
		generateEncodeBenchmark(&b, &st)
		b.WriteString("\n")
		generateDecodeBenchmark(&b, &st)
		b.WriteString("\n")
		generateRoundtripBenchmark(&b, &st)
		b.WriteString("\n")
	}

	// Generate benchmark group
	generateBenchmarkGroup(&b, schema)

	b.WriteString("criterion_main!(benches);\n")

	return b.String()
}

func generateEncodeBenchmark(b *strings.Builder, st *parser.Struct) {
	structName := st.Name
	funcName := "bench_encode_" + toSnakeCase(structName)

	b.WriteString("fn " + funcName + "(c: &mut Criterion) {\n")
	b.WriteString("    let data = " + structName + " {\n")

	// Generate test data for fields
	for _, field := range st.Fields {
		b.WriteString("        " + toSnakeCase(field.Name) + ": ")
		b.WriteString(generateBenchTestValue(&field))
		b.WriteString(",\n")
	}

	b.WriteString("    };\n\n")
	b.WriteString("    c.bench_function(\"" + toSnakeCase(structName) + "/encode\", |bencher| {\n")
	b.WriteString("        let mut buf = vec![0u8; data.encoded_size()];\n")
	b.WriteString("        bencher.iter(|| {\n")
	b.WriteString("            black_box(data.encode_to_slice(&mut buf).unwrap());\n")
	b.WriteString("        });\n")
	b.WriteString("    });\n")
	b.WriteString("}\n")
}

func generateDecodeBenchmark(b *strings.Builder, st *parser.Struct) {
	structName := st.Name
	funcName := "bench_decode_" + toSnakeCase(structName)

	b.WriteString("fn " + funcName + "(c: &mut Criterion) {\n")
	b.WriteString("    let data = " + structName + " {\n")

	// Generate test data for fields
	for _, field := range st.Fields {
		b.WriteString("        " + toSnakeCase(field.Name) + ": ")
		b.WriteString(generateBenchTestValue(&field))
		b.WriteString(",\n")
	}

	b.WriteString("    };\n\n")
	b.WriteString("    let mut buf = vec![0u8; data.encoded_size()];\n")
	b.WriteString("    data.encode_to_slice(&mut buf).unwrap();\n\n")
	b.WriteString("    c.bench_function(\"" + toSnakeCase(structName) + "/decode\", |bencher| {\n")
	b.WriteString("        bencher.iter(|| {\n")
	b.WriteString("            black_box(" + structName + "::decode_from_slice(&buf).unwrap());\n")
	b.WriteString("        });\n")
	b.WriteString("    });\n")
	b.WriteString("}\n")
}

func generateRoundtripBenchmark(b *strings.Builder, st *parser.Struct) {
	structName := st.Name
	funcName := "bench_roundtrip_" + toSnakeCase(structName)

	b.WriteString("fn " + funcName + "(c: &mut Criterion) {\n")
	b.WriteString("    let data = " + structName + " {\n")

	// Generate test data for fields
	for _, field := range st.Fields {
		b.WriteString("        " + toSnakeCase(field.Name) + ": ")
		b.WriteString(generateBenchTestValue(&field))
		b.WriteString(",\n")
	}

	b.WriteString("    };\n\n")
	b.WriteString("    c.bench_function(\"" + toSnakeCase(structName) + "/roundtrip\", |bencher| {\n")
	b.WriteString("        let mut buf = vec![0u8; data.encoded_size()];\n")
	b.WriteString("        bencher.iter(|| {\n")
	b.WriteString("            data.encode_to_slice(&mut buf).unwrap();\n")
	b.WriteString("            black_box(" + structName + "::decode_from_slice(&buf).unwrap());\n")
	b.WriteString("        });\n")
	b.WriteString("    });\n")
	b.WriteString("}\n")
}

func generateBenchmarkGroup(b *strings.Builder, schema *parser.Schema) {
	b.WriteString("criterion_group!(\n")
	b.WriteString("    benches,\n")

	for i, st := range schema.Structs {
		structName := toSnakeCase(st.Name)
		b.WriteString("    bench_encode_" + structName + ",\n")
		b.WriteString("    bench_decode_" + structName + ",\n")
		b.WriteString("    bench_roundtrip_" + structName)
		if i < len(schema.Structs)-1 {
			b.WriteString(",\n")
		} else {
			b.WriteString("\n")
		}
	}

	b.WriteString(");\n\n")
}

func generateBenchTestValue(field *parser.Field) string {
	typeExpr := &field.Type

	// Handle optional types
	if typeExpr.Optional {
		baseValue := generateBenchBaseTestValue(typeExpr)
		return "Some(" + baseValue + ")"
	}

	// Handle arrays
	if typeExpr.Kind == parser.TypeKindArray {
		elemValue := generateBenchBaseTestValue(typeExpr.Elem)
		// Use realistic array sizes for benchmarking
		return "vec![" + elemValue + ", " + elemValue + ", " + elemValue + "]"
	}

	return generateBenchBaseTestValue(typeExpr)
}

func generateBenchBaseTestValue(typeExpr *parser.TypeExpr) string {
	// For nested structs, use Default
	if typeExpr.Kind == parser.TypeKindNamed {
		typeName := typeExpr.Name
		// Check if it's a primitive type
		switch typeName {
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
			return "\"Benchmark test string\".to_string()"
		default:
			// Assume it's a nested struct
			return typeName + "::default()"
		}
	}

	if typeExpr.Kind == parser.TypeKindArray {
		return generateBenchBaseTestValue(typeExpr.Elem)
	}

	return "Default::default()"
}
