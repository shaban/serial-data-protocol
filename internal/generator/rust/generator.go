// Package rust generates Rust code from SDP schemas.
// It produces high-performance, idiomatic Rust with both slice-based
// and trait-based APIs for maximum flexibility.
package rust

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// Generate creates Rust code files from a parsed schema.
// It generates a proper Cargo crate structure:
//   - Cargo.toml: Crate manifest with aggressive optimizations
//   - src/lib.rs: Module declarations and re-exports
//   - src/types.rs: Struct definitions with derive macros
//   - src/encode.rs: Slice-based encoding (fast path for IPC)
//   - src/decode.rs: Slice-based decoding
//
// The generated code uses the sdp crate's wire_slice module for
// maximum performance (4x faster than trait-based encoding).
func Generate(schema *parser.Schema, outputDir string, verbose bool) error {
	if schema == nil {
		return fmt.Errorf("schema is nil")
	}

	if outputDir == "" {
		return fmt.Errorf("output directory is empty")
	}

	// Create output directory structure (proper Cargo crate)
	srcDir := filepath.Join(outputDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		return fmt.Errorf("failed to create src directory: %w", err)
	}

	if verbose {
		fmt.Printf("Generating Rust code in %s\n", outputDir)
	}

	// Generate Cargo.toml
	if err := generateCargoToml(schema, outputDir, verbose); err != nil {
		return fmt.Errorf("failed to generate Cargo.toml: %w", err)
	}

	// Use src/ subdirectory for proper Cargo structure
	srcDir = filepath.Join(outputDir, "src")

	// Generate lib.rs
	if err := generateLib(schema, srcDir, verbose); err != nil {
		return fmt.Errorf("failed to generate lib.rs: %w", err)
	}

	// Generate types.rs
	if err := generateTypes(schema, srcDir, verbose); err != nil {
		return fmt.Errorf("failed to generate types.rs: %w", err)
	}

	// Generate encode.rs (slice API)
	if err := generateEncode(schema, srcDir, verbose); err != nil {
		return fmt.Errorf("failed to generate encode.rs: %w", err)
	}

	// Generate decode.rs (slice API)
	if err := generateDecode(schema, srcDir, verbose); err != nil {
		return fmt.Errorf("failed to generate decode.rs: %w", err)
	}

	// Generate embedded wire runtime (makes crate self-contained)
	if err := generateWireRuntime(srcDir, verbose); err != nil {
		return fmt.Errorf("failed to generate wire runtime: %w", err)
	}

	// Note: No example/benchmark server generation - benchmarking is external
	// Use shell scripts + Make targets for cross-language testing instead

	if verbose {
		fmt.Println("Rust code generation complete")
	}

	return nil
}

// generateCargoToml creates Cargo.toml with aggressive optimizations
func generateCargoToml(schema *parser.Schema, outputDir string, verbose bool) error {
	filepath := filepath.Join(outputDir, "Cargo.toml")

	// Determine package name from schema or directory
	packageName := "sdp-generated"
	if len(schema.Structs) > 0 {
		// Use first struct name as package hint
		packageName = "sdp-" + toSnakeCase(schema.Structs[0].Name)
	}
	serverName := toSnakeCase(schema.Structs[0].Name) + "_server"

	var content string
	content += "[package]\n"
	content += fmt.Sprintf("name = \"%s\"\n", packageName)
	content += "version = \"0.2.0-rc1\"\n"
	content += "edition = \"2021\"\n"
	content += "authors = [\"Serial Data Protocol Contributors\"]\n"
	content += "license = \"MIT\"\n"
	content += "description = \"Generated SDP package\"\n\n"

	content += "[dependencies]\n"
	content += "# Only external dependency: byteorder for endianness handling\n"
	content += "byteorder = \"1.5\"\n\n"

	content += "[features]\n"
	content += "# Benchmark server mode (for cross-language performance testing)\n"
	content += fmt.Sprintf("# Build with: cargo build --release --example %s --features bench-server\n", serverName)
	content += "bench-server = []\n\n"

	content += "[profile.release]\n"
	content += "# Maximum performance optimizations\n"
	content += "opt-level = 3              # Maximum optimization level\n"
	content += "lto = \"thin\"               # Link-time optimization (thin = good balance)\n"
	content += "codegen-units = 1          # Single codegen unit for maximum optimization\n"
	content += "panic = 'abort'            # Smaller binary, no unwinding\n"
	content += "strip = true               # Strip symbols from binary\n"
	content += "overflow-checks = false    # Disable integer overflow checks in release\n"
	content += "debug = false              # No debug info\n"
	content += "incremental = false        # Disable incremental compilation for max optimization\n\n"

	content += "# Aggressive optimization flags for all dependencies\n"
	content += "[profile.release.package.\"*\"]\n"
	content += "opt-level = 3\n"
	content += "codegen-units = 1\n\n"

	content += "# Development dependencies for benchmarking\n"
	content += "[dev-dependencies]\n"
	content += "criterion = { version = \"0.5\", features = [\"html_reports\"] }\n\n"

	content += "# Example binary for cross-platform testing\n"
	content += "[[example]]\n"
	content += fmt.Sprintf("name = \"%s\"\n", serverName)
	content += fmt.Sprintf("path = \"examples/%s.rs\"\n\n", serverName)

	content += "# Criterion benchmark configuration\n"
	content += "[[bench]]\n"
	content += "name = \"benchmarks\"\n"
	content += "path = \"benches/benchmarks.rs\"\n"
	content += "harness = false  # Use Criterion instead of default harness\n"

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return err
	}

	if verbose {
		fmt.Printf("  Generated Cargo.toml\n")
	}

	return nil
}

// generateLib creates lib.rs with module declarations and re-exports
func generateLib(_ *parser.Schema, outputDir string, verbose bool) error {
	filepath := filepath.Join(outputDir, "lib.rs")

	var content string
	content += "// Code generated by sdp-gen. DO NOT EDIT.\n"
	content += "// https://github.com/shaban/serial-data-protocol\n\n"
	content += "//! Self-contained SDP generated package with embedded wire format runtime.\n"
	content += "//! This crate has no dependencies on external SDP libraries.\n\n"
	content += "mod wire;         // Embedded: Read/Write trait API\n"
	content += "mod wire_slice;   // Embedded: Direct slice API (faster)\n"
	content += "mod types;\n"
	content += "mod encode;\n"
	content += "mod decode;\n\n"
	content += "pub use types::*;\n"
	content += "pub use encode::*;\n"
	content += "pub use decode::*;\n\n"
	content += "// Re-export common wire format types\n"
	content += "pub use wire::{Error, Result, Encoder, Decoder};\n"
	content += "pub use wire_slice::{SliceError, SliceResult};\n"

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return err
	}

	if verbose {
		fmt.Printf("  Generated lib.rs\n")
	}

	return nil
}

// generateTypes creates types.rs with struct definitions
func generateTypes(schema *parser.Schema, outputDir string, verbose bool) error {
	filepath := filepath.Join(outputDir, "types.rs")

	var content string
	content += "// Code generated by sdp-gen. DO NOT EDIT.\n\n"

	// Generate all struct definitions
	structs, err := GenerateStructs(schema)
	if err != nil {
		return err
	}

	content += structs

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return err
	}

	if verbose {
		fmt.Printf("  Generated types.rs (%d structs)\n", len(schema.Structs))
	}

	return nil
}

// generateEncode creates encode.rs with slice-based encoding
func generateEncode(schema *parser.Schema, outputDir string, verbose bool) error {
	filepath := filepath.Join(outputDir, "encode.rs")

	content, err := GenerateEncode(schema)
	if err != nil {
		return fmt.Errorf("failed to generate encode: %w", err)
	}

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return err
	}

	if verbose {
		fmt.Printf("  Generated encode.rs (%d structs)\n", len(schema.Structs))
	}

	return nil
}

// generateDecode creates decode.rs with slice-based decoding
func generateDecode(schema *parser.Schema, outputDir string, verbose bool) error {
	filepath := filepath.Join(outputDir, "decode.rs")

	content, err := GenerateDecode(schema)
	if err != nil {
		return fmt.Errorf("failed to generate decode: %w", err)
	}

	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return err
	}

	if verbose {
		fmt.Printf("  Generated decode.rs (%d structs)\n", len(schema.Structs))
	}

	return nil
}

// generateWireRuntime writes the embedded wire format runtime files
func generateWireRuntime(srcDir string, verbose bool) error {
	// Write wire.rs (Read/Write trait API)
	wirePath := filepath.Join(srcDir, "wire.rs")
	if err := os.WriteFile(wirePath, []byte(wireRuntime), 0644); err != nil {
		return fmt.Errorf("failed to write wire.rs: %w", err)
	}

	// Write wire_slice.rs (direct slice API)
	slicePath := filepath.Join(srcDir, "wire_slice.rs")
	if err := os.WriteFile(slicePath, []byte(wireSliceRuntime), 0644); err != nil {
		return fmt.Errorf("failed to write wire_slice.rs: %w", err)
	}

	if verbose {
		fmt.Printf("  Generated wire.rs (embedded runtime)\n")
		fmt.Printf("  Generated wire_slice.rs (embedded runtime)\n")
	}

	return nil
}

// generateExampleHelper creates the benchmark server example
func generateExampleHelper(schema *parser.Schema, outputDir string, benchMode bool, verbose bool) error {
	// Create examples directory
	examplesDir := filepath.Join(outputDir, "examples")
	if err := os.MkdirAll(examplesDir, 0755); err != nil {
		return fmt.Errorf("failed to create examples directory: %w", err)
	}

	// Determine package name and server name
	packageName := "sdp-generated"
	if len(schema.Structs) > 0 {
		packageName = toSnakeCase(schema.Structs[0].Name)
	}
	serverName := packageName + "_server"

	// Generate server content
	content := GenerateExample(schema, packageName, benchMode)

	// Write server file
	serverPath := filepath.Join(examplesDir, serverName+".rs")
	if err := os.WriteFile(serverPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write server: %w", err)
	}

	if verbose {
		fmt.Printf("  Generated examples/%s.rs\n", serverName)
	}

	return nil
}

// generateBenchmarkHelper creates Criterion benchmarks
func generateBenchmarkHelper(schema *parser.Schema, outputDir string, verbose bool) error {
	// Create benches directory
	benchesDir := filepath.Join(outputDir, "benches")
	if err := os.MkdirAll(benchesDir, 0755); err != nil {
		return fmt.Errorf("failed to create benches directory: %w", err)
	}

	// Determine package name
	packageName := "sdp-generated"
	if len(schema.Structs) > 0 {
		packageName = toSnakeCase(schema.Structs[0].Name)
	}

	// Generate benchmark content
	content := GenerateBenchmark(schema, packageName)

	// Write benchmark file
	benchPath := filepath.Join(benchesDir, "benchmarks.rs")
	if err := os.WriteFile(benchPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write benchmark: %w", err)
	}

	if verbose {
		fmt.Printf("  Generated benches/benchmarks.rs\n")
	}

	return nil
}
