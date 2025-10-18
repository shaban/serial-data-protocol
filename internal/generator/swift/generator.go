package swift

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/shaban/serial-data-protocol/internal/parser"
)

// Generate generates Swift code from a schema
func Generate(schema *parser.Schema, outputDir string, verbose bool) error {
	if schema == nil {
		return fmt.Errorf("schema is nil")
	}

	if outputDir == "" {
		return fmt.Errorf("output directory is empty")
	}

	// Create output directory structure
	sourcesDir := filepath.Join(outputDir, "Sources")
	if err := os.MkdirAll(sourcesDir, 0755); err != nil {
		return fmt.Errorf("failed to create Sources directory: %w", err)
	}

	// Get package name from the last component of outputDir
	packageName := filepath.Base(outputDir)
	if packageName == "." || packageName == "/" {
		packageName = "SDP"
	}

	packageDir := filepath.Join(sourcesDir, packageName)
	if err := os.MkdirAll(packageDir, 0755); err != nil {
		return fmt.Errorf("failed to create package directory: %w", err)
	}

	if verbose {
		fmt.Printf("Generating Swift code in %s\n", outputDir)
	}

	// Generate Package.swift
	if err := generatePackageSwift(outputDir, packageName, verbose); err != nil {
		return fmt.Errorf("failed to generate Package.swift: %w", err)
	}

	// Generate Types.swift
	if err := generateTypes(schema, packageDir, verbose); err != nil {
		return fmt.Errorf("failed to generate Types.swift: %w", err)
	}

	// Generate Encode.swift
	if err := generateEncode(schema, packageDir, verbose); err != nil {
		return fmt.Errorf("failed to generate Encode.swift: %w", err)
	}

	// Generate Decode.swift
	if err := generateDecode(schema, packageDir, verbose); err != nil {
		return fmt.Errorf("failed to generate Decode.swift: %w", err)
	}

	if verbose {
		fmt.Println("Swift code generation completed successfully")
	}

	return nil
}

// generatePackageSwift generates the Package.swift manifest
func generatePackageSwift(outputDir, packageName string, verbose bool) error {
	content := fmt.Sprintf(`// swift-tools-version:5.9
import PackageDescription

let package = Package(
    name: "%s",
    products: [
        .library(
            name: "%s",
            targets: ["%s"]),
    ],
    targets: [
        .target(
            name: "%s",
            dependencies: []),
    ]
)
`, packageName, packageName, packageName, packageName)

	path := filepath.Join(outputDir, "Package.swift")

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write Package.swift: %w", err)
	}

	if verbose {
		fmt.Printf("Generated %s\n", path)
	}

	return nil
}

// generateTypes generates Types.swift with struct definitions
func generateTypes(schema *parser.Schema, packageDir string, verbose bool) error {
	content, err := GenerateStructs(schema)
	if err != nil {
		return err
	}

	path := filepath.Join(packageDir, "Types.swift")

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write Types.swift: %w", err)
	}

	if verbose {
		fmt.Printf("Generated %s\n", path)
	}

	return nil
}

// generateEncode generates Encode.swift with encoding methods
func generateEncode(schema *parser.Schema, packageDir string, verbose bool) error {
	content, err := GenerateEncode(schema)
	if err != nil {
		return err
	}

	path := filepath.Join(packageDir, "Encode.swift")

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write Encode.swift: %w", err)
	}

	if verbose {
		fmt.Printf("Generated %s\n", path)
	}

	return nil
}

// generateDecode generates Decode.swift with decoding methods
func generateDecode(schema *parser.Schema, packageDir string, verbose bool) error {
	content, err := GenerateDecode(schema)
	if err != nil {
		return err
	}

	path := filepath.Join(packageDir, "Decode.swift")

	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write Decode.swift: %w", err)
	}

	if verbose {
		fmt.Printf("Generated %s\n", path)
	}

	return nil
}
