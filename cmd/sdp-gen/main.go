package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/shaban/serial-data-protocol/internal/generator/c"
	"github.com/shaban/serial-data-protocol/internal/generator/cpp"
	"github.com/shaban/serial-data-protocol/internal/generator/golang"
	"github.com/shaban/serial-data-protocol/internal/generator/rust"
	"github.com/shaban/serial-data-protocol/internal/generator/swift"
	"github.com/shaban/serial-data-protocol/internal/parser"
	"github.com/shaban/serial-data-protocol/internal/validator"
)

const version = "1.0.0"

func main() {
	// Define flags
	var (
		schemaPath   = flag.String("schema", "", "Path to .sdp schema file (required)")
		outputDir    = flag.String("output", "", "Output directory for generated code (required)")
		lang         = flag.String("lang", "go", "Target language: go, c")
		packageName  = flag.String("package", "", "Package name for generated code (Go only, defaults to output dir basename)")
		mode         = flag.String("mode", "normal", "Generation mode: normal (production), bench (with benchmark server)")
		validateOnly = flag.Bool("validate-only", false, "Only validate schema without generating code")
		astJSON      = flag.Bool("ast-json", false, "Output AST as JSON instead of generating code (for other language generators)")
		verbose      = flag.Bool("verbose", false, "Enable verbose output")
		showVersion  = flag.Bool("version", false, "Show version and exit")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "sdp-gen - Serial Data Protocol Code Generator v%s\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage: sdp-gen -schema <file> -output <dir> [options]\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nGeneration Modes:\n")
		fmt.Fprintf(os.Stderr, "  normal - Production code (default, no benchmark server)\n")
		fmt.Fprintf(os.Stderr, "  bench  - Includes TCP benchmark server for cross-language testing\n\n")
		fmt.Fprintf(os.Stderr, "Examples:\n")
		fmt.Fprintf(os.Stderr, "  # Generate Go code\n")
		fmt.Fprintf(os.Stderr, "  sdp-gen -schema device.sdp -output ./generated -lang go\n\n")
		fmt.Fprintf(os.Stderr, "  # Generate Rust code with benchmark server\n")
		fmt.Fprintf(os.Stderr, "  sdp-gen -schema device.sdp -output ./generated -lang rust -mode bench\n\n")
		fmt.Fprintf(os.Stderr, "  # Generate Swift code\n")
		fmt.Fprintf(os.Stderr, "  sdp-gen -schema device.sdp -output ./generated -lang swift\n\n")
		fmt.Fprintf(os.Stderr, "  # Validate schema only\n")
		fmt.Fprintf(os.Stderr, "  sdp-gen -schema device.sdp -validate-only\n\n")
	}

	flag.Parse()

	// Handle version flag
	if *showVersion {
		fmt.Printf("sdp-gen version %s\n", version)
		os.Exit(0)
	}

	// Validate required flags
	if *schemaPath == "" {
		fmt.Fprintf(os.Stderr, "Error: -schema flag is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	if !*validateOnly && !*astJSON && *outputDir == "" {
		fmt.Fprintf(os.Stderr, "Error: -output flag is required (or use -validate-only or -ast-json)\n\n")
		flag.Usage()
		os.Exit(1)
	}

	// Validate language
	validLangs := map[string]bool{
		"go":    true,
		"rust":  true,
		"swift": true,
		"c":     true,
		"cpp":   true,
	}

	if !validLangs[*lang] {
		fmt.Fprintf(os.Stderr, "Error: -lang must be 'go', 'rust', 'swift', 'c', or 'cpp', got '%s'\n", *lang)
		os.Exit(1)
	}

	// Validate mode
	if *mode != "normal" && *mode != "bench" {
		fmt.Fprintf(os.Stderr, "Error: -mode must be 'normal' or 'bench', got '%s'\n", *mode)
		os.Exit(1)
	}

	// Run the generator
	if err := run(*schemaPath, *outputDir, *lang, *packageName, *mode, *validateOnly, *astJSON, *verbose); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func run(schemaPath, outputDir, lang, packageName, mode string, validateOnly, astJSON, verbose bool) error {
	// Step 1: Load schema
	if verbose {
		fmt.Printf("Loading schema from: %s\n", schemaPath)
	}

	schema, err := parser.LoadSchemaFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to load schema: %w", err)
	}

	if verbose {
		fmt.Printf("Loaded %d struct(s)\n", len(schema.Structs))
	}

	// Step 2: Validate schema
	if verbose {
		fmt.Println("Validating schema...")
	}

	if err := validator.Validate(schema); err != nil {
		return fmt.Errorf("schema validation failed: %w", err)
	}

	if verbose {
		fmt.Println("Schema is valid ✓")
	}

	// If validate-only mode, we're done
	if validateOnly {
		fmt.Println("Schema validation passed")
		return nil
	}

	// If ast-json mode, output AST as JSON and exit
	if astJSON {
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(schema); err != nil {
			return fmt.Errorf("failed to encode AST as JSON: %w", err)
		}
		return nil
	}

	// Step 3: Determine package name for Go
	if packageName == "" && lang == "go" {
		packageName = filepath.Base(outputDir)
		// Sanitize package name: replace hyphens and invalid characters with underscores
		packageName = sanitizePackageName(packageName)
		if verbose {
			fmt.Printf("Using package name: %s\n", packageName)
		}
	}

	// Step 4: Generate code based on language
	if verbose {
		fmt.Printf("Generating %s code...\n", lang)
	}

	switch lang {
	case "go":
		var files map[string]string
		files, err = generateGo(schema, packageName)
		if err != nil {
			return fmt.Errorf("failed to generate Go code: %w", err)
		}

		// Create output directory
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}

		// Write files
		if verbose {
			fmt.Printf("Writing files to: %s\n", outputDir)
		}

		for filename, content := range files {
			filePath := filepath.Join(outputDir, filename)
			if verbose {
				fmt.Printf("  Writing %s\n", filename)
			}

			if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
				return fmt.Errorf("failed to write %s: %w", filename, err)
			}
		}

	case "rust":
		// Rust generator writes files directly
		benchMode := (mode == "bench")
		err = rust.Generate(schema, outputDir, benchMode, verbose)
		if err != nil {
			return fmt.Errorf("failed to generate Rust code: %w", err)
		}

	case "swift":
		// Swift generator writes files directly
		benchMode := (mode == "bench")
		err = swift.Generate(schema, outputDir, benchMode, verbose)
		if err != nil {
			return fmt.Errorf("failed to generate Swift code: %w", err)
		}

	case "c":
		// C generator writes files directly
		benchMode := (mode == "bench")
		err = c.Generate(schema, outputDir, benchMode, verbose)
		if err != nil {
			return fmt.Errorf("failed to generate C code: %w", err)
		}

	case "cpp":
		// C++ generator writes files directly
		benchMode := (mode == "bench")
		err = cpp.Generate(schema, outputDir, benchMode, verbose)
		if err != nil {
			return fmt.Errorf("failed to generate C++ code: %w", err)
		}

	default:
		return fmt.Errorf("unsupported language: %s", lang)
	}

	fmt.Printf("Successfully generated %s code in %s\n", lang, outputDir)
	return nil
}

// generateGo generates Go code files
func generateGo(schema *parser.Schema, packageName string) (map[string]string, error) {
	files := make(map[string]string)

	// Generate structs
	structs, err := golang.GenerateStructs(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate structs: %w", err)
	}

	// Generate encoder
	encoder, err := golang.GenerateEncoder(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate encoder: %w", err)
	}

	encodeHelpers, err := golang.GenerateEncodeHelpers(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate encode helpers: %w", err)
	}

	// Generate decoder
	decoder, err := golang.GenerateDecoder(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate decoder: %w", err)
	}

	decodeHelpers, err := golang.GenerateDecodeHelpers(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate decode helpers: %w", err)
	}

	// Generate message mode encoders
	messageEncoders, err := golang.GenerateMessageEncoders(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate message encoders: %w", err)
	}

	// Generate message mode decoders
	messageDecoders, err := golang.GenerateMessageDecoders(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate message decoders: %w", err)
	}

	// Generate message dispatcher
	messageDispatcher, err := golang.GenerateMessageDispatcher(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate message dispatcher: %w", err)
	}

	// Generate writer-based encoders (streaming I/O)
	writerEncoders, err := golang.GenerateWriterEncoder(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate writer encoders: %w", err)
	}

	// Generate reader-based decoders (streaming I/O)
	readerDecoders, err := golang.GenerateReaderDecoder(schema)
	if err != nil {
		return nil, fmt.Errorf("failed to generate reader decoders: %w", err)
	}

	// Generate errors and context
	errors := golang.GenerateErrors()
	context := golang.GenerateDecodeContext()

	// Combine encoder code (regular + helpers + message mode + writer mode)
	encodeCode := encoder + "\n\n" + encodeHelpers + "\n\n" + messageEncoders + "\n\n" + writerEncoders

	// Combine decoder code (context + regular + helpers + message mode + dispatcher + reader mode)
	decodeCode := context + "\n\n" + decoder + "\n\n" + decodeHelpers + "\n\n" + messageDecoders + "\n\n" + messageDispatcher + "\n\n" + readerDecoders

	// Determine imports based on content
	files["types.go"] = formatGoFileWithAutoImports(packageName, structs)
	files["encode.go"] = formatGoFileWithAutoImports(packageName, encodeCode)
	files["decode.go"] = formatGoFileWithAutoImports(packageName, decodeCode)
	files["errors.go"] = formatGoFileWithAutoImports(packageName, errors)

	return files, nil
}

// formatGoFile creates a complete Go source file with package and imports
func formatGoFile(packageName string, imports []string, body string) string {
	result := fmt.Sprintf("package %s\n\n", packageName)

	if len(imports) > 0 {
		result += "import (\n"
		for _, imp := range imports {
			result += fmt.Sprintf("\t%q\n", imp)
		}
		result += ")\n\n"
	}

	result += body

	return result
}

// formatGoFileWithAutoImports creates a Go file and automatically detects needed imports
func formatGoFileWithAutoImports(packageName string, body string) string {
	var neededImports []string

	// Check for common imports based on what's in the code
	importChecks := map[string][]string{
		"encoding/binary": {"binary.LittleEndian"},
		"errors":          {"errors.New"},
		"math":            {"math.Float"},
		"io":              {"io.ReadAll", "w io.Writer", "r io.Reader"}, // For streaming I/O functions
	}

	for importPath, markers := range importChecks {
		for _, marker := range markers {
			if strings.Contains(body, marker) {
				neededImports = append(neededImports, importPath)
				break // Only add the import once
			}
		}
	}

	return formatGoFile(packageName, neededImports, body)
}

// sanitizePackageName converts a directory name to a valid Go package name
func sanitizePackageName(name string) string {
	// Replace hyphens and other invalid characters with underscores
	result := ""
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			result += string(r)
		} else {
			result += "_"
		}
	}

	// Ensure it doesn't start with a digit
	if len(result) > 0 && result[0] >= '0' && result[0] <= '9' {
		result = "_" + result
	}

	return result
}
