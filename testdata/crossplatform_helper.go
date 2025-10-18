// crossplatform_helper.go - Helper for cross-language wire format tests
// Usage:
//   go run crossplatform_helper.go encode-primitives
//   go run crossplatform_helper.go decode-primitives <file>

package main

import (
	"fmt"
	"os"

	primitives "github.com/shaban/serial-data-protocol/testdata/primitives/go"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <command> [args]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Commands:\n")
		fmt.Fprintf(os.Stderr, "  encode-primitives - Encode primitives and output binary to stdout\n")
		fmt.Fprintf(os.Stderr, "  decode-primitives <file> - Decode primitives from file\n")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "encode-primitives":
		encodePrimitives()
	case "decode-primitives":
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Error: decode-primitives requires filename argument\n")
			os.Exit(1)
		}
		decodePrimitives(os.Args[2])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func encodePrimitives() {
	// Create test data
	data := primitives.AllPrimitives{
		U8Field:   255,
		U16Field:  65535,
		U32Field:  4294967295,
		U64Field:  18446744073709551615,
		I8Field:   -128,
		I16Field:  -32768,
		I32Field:  -2147483648,
		I64Field:  -9223372036854775808,
		F32Field:  3.14159,
		F64Field:  2.718281828459045,
		BoolField: true,
		StrField:  "Hello from Go!",
	}

	// Encode
	encoded, err := primitives.EncodeAllPrimitives(&data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Encode error: %v\n", err)
		os.Exit(1)
	}

	// Write binary to stdout
	os.Stdout.Write(encoded)
}

func decodePrimitives(filename string) {
	// Read binary file
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Read error: %v\n", err)
		os.Exit(1)
	}

	// Decode
	var decoded primitives.AllPrimitives
	if err := primitives.DecodeAllPrimitives(&decoded, data); err != nil {
		fmt.Fprintf(os.Stderr, "Decode error: %v\n", err)
		os.Exit(1)
	}

	// Verify values (these should match what Rust encodes)
	ok := true
	ok = ok && decoded.U8Field == 255
	ok = ok && decoded.U16Field == 65535
	ok = ok && decoded.U32Field == 4294967295
	ok = ok && decoded.U64Field == 18446744073709551615
	ok = ok && decoded.I8Field == -128
	ok = ok && decoded.I16Field == -32768
	ok = ok && decoded.I32Field == -2147483648
	ok = ok && decoded.I64Field == -9223372036854775808
	ok = ok && (decoded.F32Field-3.14159) < 0.0001 && (decoded.F32Field-3.14159) > -0.0001
	ok = ok && (decoded.F64Field-2.718281828459045) < 0.0000001 && (decoded.F64Field-2.718281828459045) > -0.0000001
	ok = ok && decoded.BoolField == true
	ok = ok && decoded.StrField == "Hello from Rust!"

	if !ok {
		fmt.Fprintf(os.Stderr, "Validation failed\n")
		fmt.Fprintf(os.Stderr, "Decoded: %+v\n", decoded)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "✓ Go successfully decoded Rust data\n")
	os.Exit(0)
}
