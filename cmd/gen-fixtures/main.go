// Generate canonical binary fixtures for benchmarks
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/shaban/serial-data-protocol/testdata/generated/go/arrays"
)

func main() {
	// Load the JSON fixture
	jsonData, err := os.ReadFile("../testdata/data/arrays.json")
	if err != nil {
		panic(fmt.Sprintf("Failed to read arrays.json: %v", err))
	}

	// Unmarshal into struct
	var data arrays.ArraysOfPrimitives
	err = json.Unmarshal(jsonData, &data)
	if err != nil {
		panic(fmt.Sprintf("Failed to unmarshal JSON: %v", err))
	}

	// Encode to SDP format
	encoded, err := arrays.EncodeArraysOfPrimitives(&data)
	if err != nil {
		panic(fmt.Sprintf("Failed to encode: %v", err))
	}

	// Write to binary file
	err = os.WriteFile("../testdata/binaries/arrays_primitives.sdpb", encoded, 0644)
	if err != nil {
		panic(fmt.Sprintf("Failed to write binary: %v", err))
	}

	fmt.Printf("✓ Generated arrays_primitives.sdpb (%d bytes)\n", len(encoded))
	fmt.Printf("  u8_array: %d elements\n", len(data.U8Array))
	fmt.Printf("  u32_array: %d elements\n", len(data.U32Array))
	fmt.Printf("  f64_array: %d elements\n", len(data.F64Array))
	fmt.Printf("  str_array: %d elements\n", len(data.StrArray))
	fmt.Printf("  bool_array: %d elements\n", len(data.BoolArray))
}
