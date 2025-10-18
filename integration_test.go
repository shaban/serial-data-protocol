package integration_test

import (
	"encoding/binary"
	"encoding/json"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/shaban/serial-data-protocol/testdata/arrays"
	"github.com/shaban/serial-data-protocol/testdata/audiounit"
	"github.com/shaban/serial-data-protocol/testdata/complex"
	"github.com/shaban/serial-data-protocol/testdata/nested"
	"github.com/shaban/serial-data-protocol/testdata/optional"
	"github.com/shaban/serial-data-protocol/testdata/primitives"
)

const (
	generatorBinary = "testdata/sdp-gen"
)

// TestMain ensures fresh generated code for every test run
func TestMain(m *testing.M) {
	// Step 1: Clean old generated packages
	cleanGeneratedPackages()

	// Step 2: Build generator binary
	if err := buildGenerator(); err != nil {
		panic("Failed to build generator: " + err.Error())
	}

	// Step 3: Generate test packages from schemas
	if err := generateTestPackages(); err != nil {
		panic("Failed to generate test packages: " + err.Error())
	}

	// Step 4: Run tests
	code := m.Run()

	// Step 5: Exit (leave generated files for inspection)
	os.Exit(code)
}

// cleanGeneratedPackages removes previously generated packages
func cleanGeneratedPackages() {
	dirs := []string{
		"testdata/primitives",
		"testdata/nested",
		"testdata/arrays",
		"testdata/complex",
		"testdata/audiounit",
		"testdata/optional",
	}

	for _, dir := range dirs {
		os.RemoveAll(dir)
	}
}

// buildGenerator compiles the sdp-gen CLI tool
func buildGenerator() error {
	cmd := exec.Command("go", "build", "-o", generatorBinary, "./cmd/sdp-gen")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

// generateTestPackages runs the generator on all test schemas
func generateTestPackages() error {
	schemas := []struct {
		schemaFile string
		outputDir  string
		pkgName    string
	}{
		{"testdata/primitives.sdp", "testdata/primitives", "primitives"},
		{"testdata/nested.sdp", "testdata/nested", "nested"},
		{"testdata/arrays.sdp", "testdata/arrays", "arrays"},
		{"testdata/complex.sdp", "testdata/complex", "complex"},
		{"testdata/audiounit.sdp", "testdata/audiounit", "audiounit"},
		{"testdata/optional.sdp", "testdata/optional", "optional"},
	}

	for _, s := range schemas {
		// Get absolute path to generator
		genPath, err := filepath.Abs(generatorBinary)
		if err != nil {
			return err
		}

		cmd := exec.Command(
			genPath,
			"-schema", s.schemaFile,
			"-output", s.outputDir,
			"-package", s.pkgName,
			"-lang", "go",
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	return nil
}

// Test that generator successfully creates packages
func TestGeneratorCreatesPackages(t *testing.T) {
	packages := []string{
		"testdata/primitives",
		"testdata/nested",
		"testdata/arrays",
		"testdata/complex",
	}

	for _, pkg := range packages {
		// Check that directory exists
		if _, err := os.Stat(pkg); os.IsNotExist(err) {
			t.Errorf("Package directory not created: %s", pkg)
		}

		// Check that expected files exist
		expectedFiles := []string{"types.go", "encode.go", "decode.go", "errors.go"}
		for _, file := range expectedFiles {
			filePath := filepath.Join(pkg, file)
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				t.Errorf("Expected file not created: %s", filePath)
			}
		}
	}
}

// TestWireFormatPrimitives tests decoding of hand-crafted binary data with all primitive types
func TestWireFormatPrimitives(t *testing.T) {
	// Build wire format manually:
	// u8_field: 42
	// u16_field: 1000
	// u32_field: 100000
	// u64_field: 10000000000
	// i8_field: -42
	// i16_field: -1000
	// i32_field: -100000
	// i64_field: -10000000000
	// f32_field: 3.14159
	// f64_field: 2.71828182845
	// bool_field: true
	// str_field: "hello"

	data := make([]byte, 0, 128)

	// u8_field: 42
	data = append(data, 42)

	// u16_field: 1000 (little-endian)
	data = append(data, 0xe8, 0x03)

	// u32_field: 100000
	buf32 := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf32, 100000)
	data = append(data, buf32...)

	// u64_field: 10000000000
	buf64 := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf64, 10000000000)
	data = append(data, buf64...)

	// i8_field: -42 (two's complement)
	i8val := int8(-42)
	data = append(data, byte(i8val))

	// i16_field: -1000
	i16val := int16(-1000)
	binary.LittleEndian.PutUint16(buf32[:2], uint16(i16val))
	data = append(data, buf32[:2]...)

	// i32_field: -100000
	i32val := int32(-100000)
	binary.LittleEndian.PutUint32(buf32, uint32(i32val))
	data = append(data, buf32...)

	// i64_field: -10000000000
	i64val := int64(-10000000000)
	binary.LittleEndian.PutUint64(buf64, uint64(i64val))
	data = append(data, buf64...)

	// f32_field: 3.14159
	binary.LittleEndian.PutUint32(buf32, math.Float32bits(3.14159))
	data = append(data, buf32...)

	// f64_field: 2.71828182845
	binary.LittleEndian.PutUint64(buf64, math.Float64bits(2.71828182845))
	data = append(data, buf64...)

	// bool_field: true
	data = append(data, 1)

	// str_field: "hello" (4-byte length prefix + string bytes)
	binary.LittleEndian.PutUint32(buf32, 5)
	data = append(data, buf32...)
	data = append(data, []byte("hello")...)

	// Decode using generated code
	var result primitives.AllPrimitives

	err := primitives.DecodeAllPrimitives(&result, data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Verify all fields
	if result.U8Field != 42 {
		t.Errorf("u8_field: got %d, want 42", result.U8Field)
	}
	if result.U16Field != 1000 {
		t.Errorf("u16_field: got %d, want 1000", result.U16Field)
	}
	if result.U32Field != 100000 {
		t.Errorf("u32_field: got %d, want 100000", result.U32Field)
	}
	if result.U64Field != 10000000000 {
		t.Errorf("u64_field: got %d, want 10000000000", result.U64Field)
	}
	if result.I8Field != -42 {
		t.Errorf("i8_field: got %d, want -42", result.I8Field)
	}
	if result.I16Field != -1000 {
		t.Errorf("i16_field: got %d, want -1000", result.I16Field)
	}
	if result.I32Field != -100000 {
		t.Errorf("i32_field: got %d, want -100000", result.I32Field)
	}
	if result.I64Field != -10000000000 {
		t.Errorf("i64_field: got %d, want -10000000000", result.I64Field)
	}

	// Float comparison with tolerance
	if math.Abs(float64(result.F32Field-3.14159)) > 0.00001 {
		t.Errorf("f32_field: got %f, want 3.14159", result.F32Field)
	}
	if math.Abs(result.F64Field-2.71828182845) > 0.00000000001 {
		t.Errorf("f64_field: got %f, want 2.71828182845", result.F64Field)
	}

	if result.BoolField != true {
		t.Errorf("bool_field: got %v, want true", result.BoolField)
	}
	if result.StrField != "hello" {
		t.Errorf("str_field: got %q, want \"hello\"", result.StrField)
	}
}

// TestWireFormatTruncatedData tests that decoder properly detects truncated data
func TestWireFormatTruncatedData(t *testing.T) {
	// Build incomplete wire format - missing most fields
	data := []byte{42} // Only u8_field

	var result primitives.AllPrimitives
	err := primitives.DecodeAllPrimitives(&result, data)

	if err == nil {
		t.Fatal("Expected error for truncated data, got nil")
	}
	if err != primitives.ErrUnexpectedEOF {
		t.Errorf("Expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestWireFormatInvalidStringLength tests detection of invalid string length
func TestWireFormatInvalidStringLength(t *testing.T) {
	data := make([]byte, 0, 128)
	buf32 := make([]byte, 4)

	// Add all fields up to str_field with dummy values
	data = append(data, 0)                      // u8
	data = append(data, 0, 0)                   // u16
	data = append(data, 0, 0, 0, 0)             // u32
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0) // u64
	data = append(data, 0)                      // i8
	data = append(data, 0, 0)                   // i16
	data = append(data, 0, 0, 0, 0)             // i32
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0) // i64
	data = append(data, 0, 0, 0, 0)             // f32
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0) // f64
	data = append(data, 0)                      // bool

	// str_field: claim length of 1000 but only provide 5 bytes
	binary.LittleEndian.PutUint32(buf32, 1000)
	data = append(data, buf32...)
	data = append(data, []byte("short")...) // Only 5 bytes

	var result primitives.AllPrimitives
	err := primitives.DecodeAllPrimitives(&result, data)

	if err == nil {
		t.Fatal("Expected error for invalid string length, got nil")
	}
	if err != primitives.ErrUnexpectedEOF {
		t.Errorf("Expected ErrUnexpectedEOF, got %v", err)
	}
}

// TestWireFormatNested tests decoding nested structures
func TestWireFormatNested(t *testing.T) {
	// Scene: {name: str, main_rect: Rectangle, count: u32}
	// Rectangle: {top_left: Point, bottom_right: Point, color: u32}
	// Point: {x: f32, y: f32}

	data := make([]byte, 0, 128)
	buf32 := make([]byte, 4)

	// Scene.name = "test"
	binary.LittleEndian.PutUint32(buf32, 4)
	data = append(data, buf32...)
	data = append(data, []byte("test")...)

	// Scene.main_rect.top_left.x = 10.5
	binary.LittleEndian.PutUint32(buf32, math.Float32bits(10.5))
	data = append(data, buf32...)

	// Scene.main_rect.top_left.y = 20.5
	binary.LittleEndian.PutUint32(buf32, math.Float32bits(20.5))
	data = append(data, buf32...)

	// Scene.main_rect.bottom_right.x = 100.5
	binary.LittleEndian.PutUint32(buf32, math.Float32bits(100.5))
	data = append(data, buf32...)

	// Scene.main_rect.bottom_right.y = 200.5
	binary.LittleEndian.PutUint32(buf32, math.Float32bits(200.5))
	data = append(data, buf32...)

	// Scene.main_rect.color = 0xFF00FF00 (green)
	binary.LittleEndian.PutUint32(buf32, 0xFF00FF00)
	data = append(data, buf32...)

	// Scene.count = 42
	binary.LittleEndian.PutUint32(buf32, 42)
	data = append(data, buf32...)

	var result nested.Scene
	err := nested.DecodeScene(&result, data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Verify nested structure
	if result.Name != "test" {
		t.Errorf("name: got %q, want \"test\"", result.Name)
	}
	if math.Abs(float64(result.MainRect.TopLeft.X-10.5)) > 0.001 {
		t.Errorf("main_rect.top_left.x: got %f, want 10.5", result.MainRect.TopLeft.X)
	}
	if math.Abs(float64(result.MainRect.TopLeft.Y-20.5)) > 0.001 {
		t.Errorf("main_rect.top_left.y: got %f, want 20.5", result.MainRect.TopLeft.Y)
	}
	if math.Abs(float64(result.MainRect.BottomRight.X-100.5)) > 0.001 {
		t.Errorf("main_rect.bottom_right.x: got %f, want 100.5", result.MainRect.BottomRight.X)
	}
	if math.Abs(float64(result.MainRect.BottomRight.Y-200.5)) > 0.001 {
		t.Errorf("main_rect.bottom_right.y: got %f, want 200.5", result.MainRect.BottomRight.Y)
	}
	if result.MainRect.Color != 0xFF00FF00 {
		t.Errorf("main_rect.color: got 0x%X, want 0xFF00FF00", result.MainRect.Color)
	}
	if result.Count != 42 {
		t.Errorf("count: got %d, want 42", result.Count)
	}
}

// TestWireFormatArrays tests decoding arrays of primitives
func TestWireFormatArrays(t *testing.T) {
	// ArraysOfPrimitives: {u8_array, u32_array, f64_array, str_array, bool_array}
	// Test just u32_array with 3 elements
	data := make([]byte, 0, 256)
	buf32 := make([]byte, 4)
	buf64 := make([]byte, 8)

	// u8_array count: 0
	binary.LittleEndian.PutUint32(buf32, 0)
	data = append(data, buf32...)

	// u32_array count: 3
	binary.LittleEndian.PutUint32(buf32, 3)
	data = append(data, buf32...)

	// u32_array elements: 100, 200, 300
	binary.LittleEndian.PutUint32(buf32, 100)
	data = append(data, buf32...)
	binary.LittleEndian.PutUint32(buf32, 200)
	data = append(data, buf32...)
	binary.LittleEndian.PutUint32(buf32, 300)
	data = append(data, buf32...)

	// f64_array count: 0
	binary.LittleEndian.PutUint32(buf32, 0)
	data = append(data, buf32...)

	// str_array count: 0
	binary.LittleEndian.PutUint32(buf32, 0)
	data = append(data, buf32...)

	// bool_array count: 0
	binary.LittleEndian.PutUint32(buf32, 0)
	data = append(data, buf32...)

	var result arrays.ArraysOfPrimitives
	err := arrays.DecodeArraysOfPrimitives(&result, data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(result.U32Array) != 3 {
		t.Fatalf("u32_array length: got %d, want 3", len(result.U32Array))
	}
	if result.U32Array[0] != 100 {
		t.Errorf("u32_array[0]: got %d, want 100", result.U32Array[0])
	}
	if result.U32Array[1] != 200 {
		t.Errorf("u32_array[1]: got %d, want 200", result.U32Array[1])
	}
	if result.U32Array[2] != 300 {
		t.Errorf("u32_array[2]: got %d, want 300", result.U32Array[2])
	}

	// Add f64_array test
	data2 := make([]byte, 0, 256)

	// u8_array count: 0
	binary.LittleEndian.PutUint32(buf32, 0)
	data2 = append(data2, buf32...)

	// u32_array count: 0
	binary.LittleEndian.PutUint32(buf32, 0)
	data2 = append(data2, buf32...)

	// f64_array count: 2
	binary.LittleEndian.PutUint32(buf32, 2)
	data2 = append(data2, buf32...)

	// f64_array elements: 3.14, 2.71
	binary.LittleEndian.PutUint64(buf64, math.Float64bits(3.14))
	data2 = append(data2, buf64...)
	binary.LittleEndian.PutUint64(buf64, math.Float64bits(2.71))
	data2 = append(data2, buf64...)

	// str_array count: 0
	binary.LittleEndian.PutUint32(buf32, 0)
	data2 = append(data2, buf32...)

	// bool_array count: 0
	binary.LittleEndian.PutUint32(buf32, 0)
	data2 = append(data2, buf32...)

	var result2 arrays.ArraysOfPrimitives
	err = arrays.DecodeArraysOfPrimitives(&result2, data2)
	if err != nil {
		t.Fatalf("Decode f64_array failed: %v", err)
	}

	if len(result2.F64Array) != 2 {
		t.Fatalf("f64_array length: got %d, want 2", len(result2.F64Array))
	}
	if math.Abs(result2.F64Array[0]-3.14) > 0.001 {
		t.Errorf("f64_array[0]: got %f, want 3.14", result2.F64Array[0])
	}
	if math.Abs(result2.F64Array[1]-2.71) > 0.001 {
		t.Errorf("f64_array[1]: got %f, want 2.71", result2.F64Array[1])
	}
}

// TestWireFormatEmptyArray tests decoding an empty array
func TestWireFormatEmptyArray(t *testing.T) {
	// All arrays empty
	data := make([]byte, 0, 32)
	buf32 := make([]byte, 4)

	// All arrays have count 0
	for i := 0; i < 5; i++ {
		binary.LittleEndian.PutUint32(buf32, 0)
		data = append(data, buf32...)
	}

	var result arrays.ArraysOfPrimitives
	err := arrays.DecodeArraysOfPrimitives(&result, data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if len(result.U8Array) != 0 {
		t.Errorf("u8_array length: got %d, want 0", len(result.U8Array))
	}
	if len(result.U32Array) != 0 {
		t.Errorf("u32_array length: got %d, want 0", len(result.U32Array))
	}
}

// TestWireFormatOversizedArray tests that decoder rejects oversized arrays
func TestWireFormatOversizedArray(t *testing.T) {
	data := make([]byte, 4)

	// u8_array count: 10,000,000 (exceeds MaxArrayElements)
	binary.LittleEndian.PutUint32(data, 10_000_000)

	var result arrays.ArraysOfPrimitives
	err := arrays.DecodeArraysOfPrimitives(&result, data)

	if err == nil {
		t.Fatal("Expected error for oversized array, got nil")
	}
	if err != arrays.ErrArrayTooLarge {
		t.Errorf("Expected ErrArrayTooLarge, got %v", err)
	}
}

// TestWireFormatComplex tests a realistic complex structure
func TestWireFormatComplex(t *testing.T) {
	// Plugin: {id: u32, name: str, manufacturer: str, version: u32, enabled: bool, parameters: []Parameter}
	// Parameter: {id: u32, name: str, value: f32, min: f32, max: f32}

	data := make([]byte, 0, 256)
	buf32 := make([]byte, 4)

	// Plugin.id = 1
	binary.LittleEndian.PutUint32(buf32, 1)
	data = append(data, buf32...)

	// Plugin.name = "compressor"
	binary.LittleEndian.PutUint32(buf32, 10)
	data = append(data, buf32...)
	data = append(data, []byte("compressor")...)

	// Plugin.manufacturer = "AudioCo"
	binary.LittleEndian.PutUint32(buf32, 7)
	data = append(data, buf32...)
	data = append(data, []byte("AudioCo")...)

	// Plugin.version = 100 (representing 1.0.0)
	binary.LittleEndian.PutUint32(buf32, 100)
	data = append(data, buf32...)

	// Plugin.enabled = true
	data = append(data, 1)

	// Plugin.parameters array count = 1
	binary.LittleEndian.PutUint32(buf32, 1)
	data = append(data, buf32...)

	// Parameter[0].id = 1
	binary.LittleEndian.PutUint32(buf32, 1)
	data = append(data, buf32...)

	// Parameter[0].name = "gain"
	binary.LittleEndian.PutUint32(buf32, 4)
	data = append(data, buf32...)
	data = append(data, []byte("gain")...)

	// Parameter[0].value = 0.75
	binary.LittleEndian.PutUint32(buf32, math.Float32bits(0.75))
	data = append(data, buf32...)

	// Parameter[0].min = 0.0
	binary.LittleEndian.PutUint32(buf32, math.Float32bits(0.0))
	data = append(data, buf32...)

	// Parameter[0].max = 1.0
	binary.LittleEndian.PutUint32(buf32, math.Float32bits(1.0))
	data = append(data, buf32...)

	var result complex.Plugin
	err := complex.DecodePlugin(&result, data)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if result.Id != 1 {
		t.Errorf("id: got %d, want 1", result.Id)
	}
	if result.Name != "compressor" {
		t.Errorf("name: got %q, want \"compressor\"", result.Name)
	}
	if result.Manufacturer != "AudioCo" {
		t.Errorf("manufacturer: got %q, want \"AudioCo\"", result.Manufacturer)
	}
	if result.Version != 100 {
		t.Errorf("version: got %d, want 100", result.Version)
	}
	if result.Enabled != true {
		t.Errorf("enabled: got %v, want true", result.Enabled)
	}
	if len(result.Parameters) != 1 {
		t.Fatalf("parameters length: got %d, want 1", len(result.Parameters))
	}
	if result.Parameters[0].Id != 1 {
		t.Errorf("parameters[0].id: got %d, want 1", result.Parameters[0].Id)
	}
	if result.Parameters[0].Name != "gain" {
		t.Errorf("parameters[0].name: got %q, want \"gain\"", result.Parameters[0].Name)
	}
	if math.Abs(float64(result.Parameters[0].Value-0.75)) > 0.0001 {
		t.Errorf("parameters[0].value: got %f, want 0.75", result.Parameters[0].Value)
	}
	if math.Abs(float64(result.Parameters[0].Min)) > 0.0001 {
		t.Errorf("parameters[0].min: got %f, want 0.0", result.Parameters[0].Min)
	}
	if math.Abs(float64(result.Parameters[0].Max-1.0)) > 0.0001 {
		t.Errorf("parameters[0].max: got %f, want 1.0", result.Parameters[0].Max)
	}
}

// TestEncoderOutputFormat tests that encoder produces correct wire format
func TestEncoderOutputFormat(t *testing.T) {
	// Create a struct with known values
	data := primitives.AllPrimitives{
		U8Field:   42,
		U16Field:  1000,
		U32Field:  100000,
		U64Field:  10000000000,
		I8Field:   -42,
		I16Field:  -1000,
		I32Field:  -100000,
		I64Field:  -10000000000,
		F32Field:  3.14159,
		F64Field:  2.71828182845,
		BoolField: true,
		StrField:  "hello",
	}

	// Encode
	encoded, err := primitives.EncodeAllPrimitives(&data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Verify wire format byte-by-byte
	offset := 0

	// u8_field: 42
	if encoded[offset] != 42 {
		t.Errorf("u8_field byte: got %d, want 42", encoded[offset])
	}
	offset += 1

	// u16_field: 1000 (0x03E8 little-endian = E8 03)
	u16val := binary.LittleEndian.Uint16(encoded[offset:])
	if u16val != 1000 {
		t.Errorf("u16_field: got %d, want 1000", u16val)
	}
	offset += 2

	// u32_field: 100000
	u32val := binary.LittleEndian.Uint32(encoded[offset:])
	if u32val != 100000 {
		t.Errorf("u32_field: got %d, want 100000", u32val)
	}
	offset += 4

	// u64_field: 10000000000
	u64val := binary.LittleEndian.Uint64(encoded[offset:])
	if u64val != 10000000000 {
		t.Errorf("u64_field: got %d, want 10000000000", u64val)
	}
	offset += 8

	// i8_field: -42
	i8val := int8(encoded[offset])
	if i8val != -42 {
		t.Errorf("i8_field: got %d, want -42", i8val)
	}
	offset += 1

	// i16_field: -1000
	i16val := int16(binary.LittleEndian.Uint16(encoded[offset:]))
	if i16val != -1000 {
		t.Errorf("i16_field: got %d, want -1000", i16val)
	}
	offset += 2

	// i32_field: -100000
	i32val := int32(binary.LittleEndian.Uint32(encoded[offset:]))
	if i32val != -100000 {
		t.Errorf("i32_field: got %d, want -100000", i32val)
	}
	offset += 4

	// i64_field: -10000000000
	i64val := int64(binary.LittleEndian.Uint64(encoded[offset:]))
	if i64val != -10000000000 {
		t.Errorf("i64_field: got %d, want -10000000000", i64val)
	}
	offset += 8

	// f32_field: 3.14159
	f32bits := binary.LittleEndian.Uint32(encoded[offset:])
	f32val := math.Float32frombits(f32bits)
	if math.Abs(float64(f32val-3.14159)) > 0.00001 {
		t.Errorf("f32_field: got %f, want 3.14159", f32val)
	}
	offset += 4

	// f64_field: 2.71828182845
	f64bits := binary.LittleEndian.Uint64(encoded[offset:])
	f64val := math.Float64frombits(f64bits)
	if math.Abs(f64val-2.71828182845) > 0.00000000001 {
		t.Errorf("f64_field: got %f, want 2.71828182845", f64val)
	}
	offset += 8

	// bool_field: true (encoded as 1)
	if encoded[offset] != 1 {
		t.Errorf("bool_field: got %d, want 1", encoded[offset])
	}
	offset += 1

	// str_field: "hello" (4-byte length + 5 bytes content)
	strLen := binary.LittleEndian.Uint32(encoded[offset:])
	if strLen != 5 {
		t.Errorf("str_field length: got %d, want 5", strLen)
	}
	offset += 4

	strVal := string(encoded[offset : offset+5])
	if strVal != "hello" {
		t.Errorf("str_field: got %q, want \"hello\"", strVal)
	}
	offset += 5

	// Verify total length
	if offset != len(encoded) {
		t.Errorf("encoded length: got %d, expected %d bytes consumed", len(encoded), offset)
	}
}

// TestEncoderArrayFormat tests array encoding format
func TestEncoderArrayFormat(t *testing.T) {
	data := arrays.ArraysOfPrimitives{
		U8Array:   []uint8{},
		U32Array:  []uint32{100, 200, 300},
		F64Array:  []float64{},
		StrArray:  []string{},
		BoolArray: []bool{},
	}

	encoded, err := arrays.EncodeArraysOfPrimitives(&data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	offset := 0

	// u8_array count: 0
	count := binary.LittleEndian.Uint32(encoded[offset:])
	if count != 0 {
		t.Errorf("u8_array count: got %d, want 0", count)
	}
	offset += 4

	// u32_array count: 3
	count = binary.LittleEndian.Uint32(encoded[offset:])
	if count != 3 {
		t.Errorf("u32_array count: got %d, want 3", count)
	}
	offset += 4

	// u32_array[0]: 100
	val := binary.LittleEndian.Uint32(encoded[offset:])
	if val != 100 {
		t.Errorf("u32_array[0]: got %d, want 100", val)
	}
	offset += 4

	// u32_array[1]: 200
	val = binary.LittleEndian.Uint32(encoded[offset:])
	if val != 200 {
		t.Errorf("u32_array[1]: got %d, want 200", val)
	}
	offset += 4

	// u32_array[2]: 300
	val = binary.LittleEndian.Uint32(encoded[offset:])
	if val != 300 {
		t.Errorf("u32_array[2]: got %d, want 300", val)
	}
	offset += 4

	// f64_array count: 0
	count = binary.LittleEndian.Uint32(encoded[offset:])
	if count != 0 {
		t.Errorf("f64_array count: got %d, want 0", count)
	}
	offset += 4

	// str_array count: 0
	count = binary.LittleEndian.Uint32(encoded[offset:])
	if count != 0 {
		t.Errorf("str_array count: got %d, want 0", count)
	}
	offset += 4

	// bool_array count: 0
	count = binary.LittleEndian.Uint32(encoded[offset:])
	if count != 0 {
		t.Errorf("bool_array count: got %d, want 0", count)
	}
	offset += 4

	if offset != len(encoded) {
		t.Errorf("encoded length: got %d, expected %d bytes consumed", len(encoded), offset)
	}
}

// TestEncoderNestedFormat tests nested struct encoding format
func TestEncoderNestedFormat(t *testing.T) {
	data := nested.Scene{
		Name: "test",
		MainRect: nested.Rectangle{
			TopLeft: nested.Point{
				X: 10.5,
				Y: 20.5,
			},
			BottomRight: nested.Point{
				X: 100.5,
				Y: 200.5,
			},
			Color: 0xFF00FF00,
		},
		Count: 42,
	}

	encoded, err := nested.EncodeScene(&data)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	offset := 0

	// Scene.name length: 4
	strLen := binary.LittleEndian.Uint32(encoded[offset:])
	if strLen != 4 {
		t.Errorf("name length: got %d, want 4", strLen)
	}
	offset += 4

	// Scene.name: "test"
	nameVal := string(encoded[offset : offset+4])
	if nameVal != "test" {
		t.Errorf("name: got %q, want \"test\"", nameVal)
	}
	offset += 4

	// Scene.main_rect.top_left.x: 10.5
	f32bits := binary.LittleEndian.Uint32(encoded[offset:])
	f32val := math.Float32frombits(f32bits)
	if math.Abs(float64(f32val-10.5)) > 0.001 {
		t.Errorf("top_left.x: got %f, want 10.5", f32val)
	}
	offset += 4

	// Scene.main_rect.top_left.y: 20.5
	f32bits = binary.LittleEndian.Uint32(encoded[offset:])
	f32val = math.Float32frombits(f32bits)
	if math.Abs(float64(f32val-20.5)) > 0.001 {
		t.Errorf("top_left.y: got %f, want 20.5", f32val)
	}
	offset += 4

	// Scene.main_rect.bottom_right.x: 100.5
	f32bits = binary.LittleEndian.Uint32(encoded[offset:])
	f32val = math.Float32frombits(f32bits)
	if math.Abs(float64(f32val-100.5)) > 0.001 {
		t.Errorf("bottom_right.x: got %f, want 100.5", f32val)
	}
	offset += 4

	// Scene.main_rect.bottom_right.y: 200.5
	f32bits = binary.LittleEndian.Uint32(encoded[offset:])
	f32val = math.Float32frombits(f32bits)
	if math.Abs(float64(f32val-200.5)) > 0.001 {
		t.Errorf("bottom_right.y: got %f, want 200.5", f32val)
	}
	offset += 4

	// Scene.main_rect.color: 0xFF00FF00
	color := binary.LittleEndian.Uint32(encoded[offset:])
	if color != 0xFF00FF00 {
		t.Errorf("color: got 0x%X, want 0xFF00FF00", color)
	}
	offset += 4

	// Scene.count: 42
	count := binary.LittleEndian.Uint32(encoded[offset:])
	if count != 42 {
		t.Errorf("count: got %d, want 42", count)
	}
	offset += 4

	if offset != len(encoded) {
		t.Errorf("encoded length: got %d, expected %d bytes consumed", len(encoded), offset)
	}
}

// TestRoundtripPrimitives tests encode→decode roundtrip for all primitive types
func TestRoundtripPrimitives(t *testing.T) {
	original := primitives.AllPrimitives{
		U8Field:   255,
		U16Field:  65535,
		U32Field:  4294967295,
		U64Field:  18446744073709551615,
		I8Field:   -128,
		I16Field:  -32768,
		I32Field:  -2147483648,
		I64Field:  -9223372036854775808,
		F32Field:  3.14159265,
		F64Field:  2.718281828459045,
		BoolField: true,
		StrField:  "Hello, World! 🎉",
	}

	// Encode
	encoded, err := primitives.EncodeAllPrimitives(&original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Decode
	var decoded primitives.AllPrimitives
	err = primitives.DecodeAllPrimitives(&decoded, encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Verify all fields match exactly
	if decoded.U8Field != original.U8Field {
		t.Errorf("U8Field: got %d, want %d", decoded.U8Field, original.U8Field)
	}
	if decoded.U16Field != original.U16Field {
		t.Errorf("U16Field: got %d, want %d", decoded.U16Field, original.U16Field)
	}
	if decoded.U32Field != original.U32Field {
		t.Errorf("U32Field: got %d, want %d", decoded.U32Field, original.U32Field)
	}
	if decoded.U64Field != original.U64Field {
		t.Errorf("U64Field: got %d, want %d", decoded.U64Field, original.U64Field)
	}
	if decoded.I8Field != original.I8Field {
		t.Errorf("I8Field: got %d, want %d", decoded.I8Field, original.I8Field)
	}
	if decoded.I16Field != original.I16Field {
		t.Errorf("I16Field: got %d, want %d", decoded.I16Field, original.I16Field)
	}
	if decoded.I32Field != original.I32Field {
		t.Errorf("I32Field: got %d, want %d", decoded.I32Field, original.I32Field)
	}
	if decoded.I64Field != original.I64Field {
		t.Errorf("I64Field: got %d, want %d", decoded.I64Field, original.I64Field)
	}
	if decoded.F32Field != original.F32Field {
		t.Errorf("F32Field: got %f, want %f", decoded.F32Field, original.F32Field)
	}
	if decoded.F64Field != original.F64Field {
		t.Errorf("F64Field: got %f, want %f", decoded.F64Field, original.F64Field)
	}
	if decoded.BoolField != original.BoolField {
		t.Errorf("BoolField: got %v, want %v", decoded.BoolField, original.BoolField)
	}
	if decoded.StrField != original.StrField {
		t.Errorf("StrField: got %q, want %q", decoded.StrField, original.StrField)
	}
}

// TestRoundtripEmptyString tests encode→decode with empty string
func TestRoundtripEmptyString(t *testing.T) {
	original := primitives.AllPrimitives{
		StrField: "", // Empty string
	}

	encoded, err := primitives.EncodeAllPrimitives(&original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	var decoded primitives.AllPrimitives
	err = primitives.DecodeAllPrimitives(&decoded, encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	if decoded.StrField != "" {
		t.Errorf("StrField: got %q, want empty string", decoded.StrField)
	}
}

// TestRoundtripNested tests encode→decode with nested structures
func TestRoundtripNested(t *testing.T) {
	original := nested.Scene{
		Name: "Main Scene",
		MainRect: nested.Rectangle{
			TopLeft: nested.Point{
				X: -100.5,
				Y: -200.5,
			},
			BottomRight: nested.Point{
				X: 500.25,
				Y: 750.75,
			},
			Color: 0xFFAA5533,
		},
		Count: 12345,
	}

	// Encode
	encoded, err := nested.EncodeScene(&original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Decode
	var decoded nested.Scene
	err = nested.DecodeScene(&decoded, encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Verify all fields
	if decoded.Name != original.Name {
		t.Errorf("Name: got %q, want %q", decoded.Name, original.Name)
	}
	if decoded.MainRect.TopLeft.X != original.MainRect.TopLeft.X {
		t.Errorf("TopLeft.X: got %f, want %f", decoded.MainRect.TopLeft.X, original.MainRect.TopLeft.X)
	}
	if decoded.MainRect.TopLeft.Y != original.MainRect.TopLeft.Y {
		t.Errorf("TopLeft.Y: got %f, want %f", decoded.MainRect.TopLeft.Y, original.MainRect.TopLeft.Y)
	}
	if decoded.MainRect.BottomRight.X != original.MainRect.BottomRight.X {
		t.Errorf("BottomRight.X: got %f, want %f", decoded.MainRect.BottomRight.X, original.MainRect.BottomRight.X)
	}
	if decoded.MainRect.BottomRight.Y != original.MainRect.BottomRight.Y {
		t.Errorf("BottomRight.Y: got %f, want %f", decoded.MainRect.BottomRight.Y, original.MainRect.BottomRight.Y)
	}
	if decoded.MainRect.Color != original.MainRect.Color {
		t.Errorf("Color: got 0x%X, want 0x%X", decoded.MainRect.Color, original.MainRect.Color)
	}
	if decoded.Count != original.Count {
		t.Errorf("Count: got %d, want %d", decoded.Count, original.Count)
	}
}

// TestRoundtripArrays tests encode→decode with arrays
func TestRoundtripArrays(t *testing.T) {
	original := arrays.ArraysOfPrimitives{
		U8Array:   []uint8{1, 2, 3, 255},
		U32Array:  []uint32{100, 200, 300, 4294967295},
		F64Array:  []float64{1.1, 2.2, 3.3, math.Pi, math.E},
		StrArray:  []string{"hello", "world", "", "test 🚀"},
		BoolArray: []bool{true, false, true, false, true},
	}

	// Encode
	encoded, err := arrays.EncodeArraysOfPrimitives(&original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Decode
	var decoded arrays.ArraysOfPrimitives
	err = arrays.DecodeArraysOfPrimitives(&decoded, encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Verify arrays
	if len(decoded.U8Array) != len(original.U8Array) {
		t.Fatalf("U8Array length: got %d, want %d", len(decoded.U8Array), len(original.U8Array))
	}
	for i := range original.U8Array {
		if decoded.U8Array[i] != original.U8Array[i] {
			t.Errorf("U8Array[%d]: got %d, want %d", i, decoded.U8Array[i], original.U8Array[i])
		}
	}

	if len(decoded.U32Array) != len(original.U32Array) {
		t.Fatalf("U32Array length: got %d, want %d", len(decoded.U32Array), len(original.U32Array))
	}
	for i := range original.U32Array {
		if decoded.U32Array[i] != original.U32Array[i] {
			t.Errorf("U32Array[%d]: got %d, want %d", i, decoded.U32Array[i], original.U32Array[i])
		}
	}

	if len(decoded.F64Array) != len(original.F64Array) {
		t.Fatalf("F64Array length: got %d, want %d", len(decoded.F64Array), len(original.F64Array))
	}
	for i := range original.F64Array {
		if decoded.F64Array[i] != original.F64Array[i] {
			t.Errorf("F64Array[%d]: got %f, want %f", i, decoded.F64Array[i], original.F64Array[i])
		}
	}

	if len(decoded.StrArray) != len(original.StrArray) {
		t.Fatalf("StrArray length: got %d, want %d", len(decoded.StrArray), len(original.StrArray))
	}
	for i := range original.StrArray {
		if decoded.StrArray[i] != original.StrArray[i] {
			t.Errorf("StrArray[%d]: got %q, want %q", i, decoded.StrArray[i], original.StrArray[i])
		}
	}

	if len(decoded.BoolArray) != len(original.BoolArray) {
		t.Fatalf("BoolArray length: got %d, want %d", len(decoded.BoolArray), len(original.BoolArray))
	}
	for i := range original.BoolArray {
		if decoded.BoolArray[i] != original.BoolArray[i] {
			t.Errorf("BoolArray[%d]: got %v, want %v", i, decoded.BoolArray[i], original.BoolArray[i])
		}
	}
}

// TestRoundtripEmptyArrays tests encode→decode with empty arrays
func TestRoundtripEmptyArrays(t *testing.T) {
	original := arrays.ArraysOfPrimitives{
		U8Array:   []uint8{},
		U32Array:  []uint32{},
		F64Array:  []float64{},
		StrArray:  []string{},
		BoolArray: []bool{},
	}

	encoded, err := arrays.EncodeArraysOfPrimitives(&original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	var decoded arrays.ArraysOfPrimitives
	err = arrays.DecodeArraysOfPrimitives(&decoded, encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Verify all arrays are empty
	if len(decoded.U8Array) != 0 {
		t.Errorf("U8Array length: got %d, want 0", len(decoded.U8Array))
	}
	if len(decoded.U32Array) != 0 {
		t.Errorf("U32Array length: got %d, want 0", len(decoded.U32Array))
	}
	if len(decoded.F64Array) != 0 {
		t.Errorf("F64Array length: got %d, want 0", len(decoded.F64Array))
	}
	if len(decoded.StrArray) != 0 {
		t.Errorf("StrArray length: got %d, want 0", len(decoded.StrArray))
	}
	if len(decoded.BoolArray) != 0 {
		t.Errorf("BoolArray length: got %d, want 0", len(decoded.BoolArray))
	}
}

// TestRoundtripStructArrays tests encode→decode with arrays of structs
func TestRoundtripStructArrays(t *testing.T) {
	original := arrays.ArraysOfStructs{
		Items: []arrays.Item{
			{Id: 1, Name: "First Item"},
			{Id: 2, Name: "Second Item"},
			{Id: 999, Name: ""},
		},
		Count: 3,
	}

	// Encode
	encoded, err := arrays.EncodeArraysOfStructs(&original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Decode
	var decoded arrays.ArraysOfStructs
	err = arrays.DecodeArraysOfStructs(&decoded, encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Verify
	if decoded.Count != original.Count {
		t.Errorf("Count: got %d, want %d", decoded.Count, original.Count)
	}
	if len(decoded.Items) != len(original.Items) {
		t.Fatalf("Items length: got %d, want %d", len(decoded.Items), len(original.Items))
	}

	for i := range original.Items {
		if decoded.Items[i].Id != original.Items[i].Id {
			t.Errorf("Items[%d].Id: got %d, want %d", i, decoded.Items[i].Id, original.Items[i].Id)
		}
		if decoded.Items[i].Name != original.Items[i].Name {
			t.Errorf("Items[%d].Name: got %q, want %q", i, decoded.Items[i].Name, original.Items[i].Name)
		}
	}
}

// TestRoundtripComplex tests encode→decode with complex nested structures
func TestRoundtripComplex(t *testing.T) {
	original := complex.AudioDevice{
		DeviceId:       42,
		DeviceName:     "USB Audio Interface",
		SampleRate:     48000,
		BufferSize:     512,
		InputChannels:  2,
		OutputChannels: 8,
		IsDefault:      true,
		ActivePlugins: []complex.Plugin{
			{
				Id:           1,
				Name:         "Compressor",
				Manufacturer: "AudioCorp",
				Version:      100,
				Enabled:      true,
				Parameters: []complex.Parameter{
					{Id: 1, Name: "threshold", Value: -20.0, Min: -60.0, Max: 0.0},
					{Id: 2, Name: "ratio", Value: 4.0, Min: 1.0, Max: 20.0},
					{Id: 3, Name: "attack", Value: 5.0, Min: 0.1, Max: 100.0},
				},
			},
			{
				Id:           2,
				Name:         "EQ",
				Manufacturer: "AudioCorp",
				Version:      200,
				Enabled:      false,
				Parameters:   []complex.Parameter{}, // Empty parameters array
			},
		},
	}

	// Encode
	encoded, err := complex.EncodeAudioDevice(&original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Decode
	var decoded complex.AudioDevice
	err = complex.DecodeAudioDevice(&decoded, encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Verify top-level fields
	if decoded.DeviceId != original.DeviceId {
		t.Errorf("DeviceId: got %d, want %d", decoded.DeviceId, original.DeviceId)
	}
	if decoded.DeviceName != original.DeviceName {
		t.Errorf("DeviceName: got %q, want %q", decoded.DeviceName, original.DeviceName)
	}
	if decoded.SampleRate != original.SampleRate {
		t.Errorf("SampleRate: got %d, want %d", decoded.SampleRate, original.SampleRate)
	}
	if decoded.BufferSize != original.BufferSize {
		t.Errorf("BufferSize: got %d, want %d", decoded.BufferSize, original.BufferSize)
	}
	if decoded.InputChannels != original.InputChannels {
		t.Errorf("InputChannels: got %d, want %d", decoded.InputChannels, original.InputChannels)
	}
	if decoded.OutputChannels != original.OutputChannels {
		t.Errorf("OutputChannels: got %d, want %d", decoded.OutputChannels, original.OutputChannels)
	}
	if decoded.IsDefault != original.IsDefault {
		t.Errorf("IsDefault: got %v, want %v", decoded.IsDefault, original.IsDefault)
	}

	// Verify plugins array
	if len(decoded.ActivePlugins) != len(original.ActivePlugins) {
		t.Fatalf("ActivePlugins length: got %d, want %d", len(decoded.ActivePlugins), len(original.ActivePlugins))
	}

	for i := range original.ActivePlugins {
		p1 := original.ActivePlugins[i]
		p2 := decoded.ActivePlugins[i]

		if p2.Id != p1.Id {
			t.Errorf("Plugin[%d].Id: got %d, want %d", i, p2.Id, p1.Id)
		}
		if p2.Name != p1.Name {
			t.Errorf("Plugin[%d].Name: got %q, want %q", i, p2.Name, p1.Name)
		}
		if p2.Manufacturer != p1.Manufacturer {
			t.Errorf("Plugin[%d].Manufacturer: got %q, want %q", i, p2.Manufacturer, p1.Manufacturer)
		}
		if p2.Version != p1.Version {
			t.Errorf("Plugin[%d].Version: got %d, want %d", i, p2.Version, p1.Version)
		}
		if p2.Enabled != p1.Enabled {
			t.Errorf("Plugin[%d].Enabled: got %v, want %v", i, p2.Enabled, p1.Enabled)
		}

		// Verify parameters
		if len(p2.Parameters) != len(p1.Parameters) {
			t.Fatalf("Plugin[%d].Parameters length: got %d, want %d", i, len(p2.Parameters), len(p1.Parameters))
		}

		for j := range p1.Parameters {
			param1 := p1.Parameters[j]
			param2 := p2.Parameters[j]

			if param2.Id != param1.Id {
				t.Errorf("Plugin[%d].Parameters[%d].Id: got %d, want %d", i, j, param2.Id, param1.Id)
			}
			if param2.Name != param1.Name {
				t.Errorf("Plugin[%d].Parameters[%d].Name: got %q, want %q", i, j, param2.Name, param1.Name)
			}
			if param2.Value != param1.Value {
				t.Errorf("Plugin[%d].Parameters[%d].Value: got %f, want %f", i, j, param2.Value, param1.Value)
			}
			if param2.Min != param1.Min {
				t.Errorf("Plugin[%d].Parameters[%d].Min: got %f, want %f", i, j, param2.Min, param1.Min)
			}
			if param2.Max != param1.Max {
				t.Errorf("Plugin[%d].Parameters[%d].Max: got %f, want %f", i, j, param2.Max, param1.Max)
			}
		}
	}
}

// TestRoundtripLargeData tests encode→decode with larger data sets
func TestRoundtripLargeData(t *testing.T) {
	// Create array with 1000 elements
	largeArray := make([]uint32, 1000)
	for i := range largeArray {
		largeArray[i] = uint32(i * 100)
	}

	original := arrays.ArraysOfPrimitives{
		U8Array:   []uint8{},
		U32Array:  largeArray,
		F64Array:  []float64{},
		StrArray:  []string{},
		BoolArray: []bool{},
	}

	// Encode
	encoded, err := arrays.EncodeArraysOfPrimitives(&original)
	if err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	// Decode
	var decoded arrays.ArraysOfPrimitives
	err = arrays.DecodeArraysOfPrimitives(&decoded, encoded)
	if err != nil {
		t.Fatalf("Decode failed: %v", err)
	}

	// Verify
	if len(decoded.U32Array) != len(original.U32Array) {
		t.Fatalf("U32Array length: got %d, want %d", len(decoded.U32Array), len(original.U32Array))
	}

	for i := range original.U32Array {
		if decoded.U32Array[i] != original.U32Array[i] {
			t.Errorf("U32Array[%d]: got %d, want %d", i, decoded.U32Array[i], original.U32Array[i])
		}
	}
}

// Benchmark encode/decode performance for primitives
func BenchmarkEncodePrimitives(b *testing.B) {
	data := primitives.AllPrimitives{
		U8Field:   255,
		U16Field:  65535,
		U32Field:  4294967295,
		U64Field:  18446744073709551615,
		I8Field:   -128,
		I16Field:  -32768,
		I32Field:  -2147483648,
		I64Field:  -9223372036854775808,
		F32Field:  3.14159265,
		F64Field:  2.718281828459045,
		BoolField: true,
		StrField:  "Hello, World! 🎉",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := primitives.EncodeAllPrimitives(&data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodePrimitives(b *testing.B) {
	data := primitives.AllPrimitives{
		U8Field:   255,
		U16Field:  65535,
		U32Field:  4294967295,
		U64Field:  18446744073709551615,
		I8Field:   -128,
		I16Field:  -32768,
		I32Field:  -2147483648,
		I64Field:  -9223372036854775808,
		F32Field:  3.14159265,
		F64Field:  2.718281828459045,
		BoolField: true,
		StrField:  "Hello, World! 🎉",
	}

	// Pre-encode once
	encoded, err := primitives.EncodeAllPrimitives(&data)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result primitives.AllPrimitives
		err := primitives.DecodeAllPrimitives(&result, encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark encode/decode performance for nested structures
func BenchmarkEncodeNested(b *testing.B) {
	data := nested.Scene{
		Name: "Main Scene",
		MainRect: nested.Rectangle{
			TopLeft: nested.Point{
				X: -100.5,
				Y: -200.5,
			},
			BottomRight: nested.Point{
				X: 500.25,
				Y: 750.75,
			},
			Color: 0xFFAA5533,
		},
		Count: 12345,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := nested.EncodeScene(&data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeNested(b *testing.B) {
	data := nested.Scene{
		Name: "Main Scene",
		MainRect: nested.Rectangle{
			TopLeft: nested.Point{
				X: -100.5,
				Y: -200.5,
			},
			BottomRight: nested.Point{
				X: 500.25,
				Y: 750.75,
			},
			Color: 0xFFAA5533,
		},
		Count: 12345,
	}

	// Pre-encode once
	encoded, err := nested.EncodeScene(&data)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result nested.Scene
		err := nested.DecodeScene(&result, encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark encode/decode performance for arrays
func BenchmarkEncodeArrays(b *testing.B) {
	data := arrays.ArraysOfPrimitives{
		U8Array:   []uint8{1, 2, 3, 255},
		U32Array:  []uint32{100, 200, 300, 4294967295},
		F64Array:  []float64{1.1, 2.2, 3.3, math.Pi, math.E},
		StrArray:  []string{"hello", "world", "", "test 🚀"},
		BoolArray: []bool{true, false, true, false, true},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := arrays.EncodeArraysOfPrimitives(&data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeArrays(b *testing.B) {
	data := arrays.ArraysOfPrimitives{
		U8Array:   []uint8{1, 2, 3, 255},
		U32Array:  []uint32{100, 200, 300, 4294967295},
		F64Array:  []float64{1.1, 2.2, 3.3, math.Pi, math.E},
		StrArray:  []string{"hello", "world", "", "test 🚀"},
		BoolArray: []bool{true, false, true, false, true},
	}

	// Pre-encode once
	encoded, err := arrays.EncodeArraysOfPrimitives(&data)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result arrays.ArraysOfPrimitives
		err := arrays.DecodeArraysOfPrimitives(&result, encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark encode/decode performance for complex structures
func BenchmarkEncodeComplex(b *testing.B) {
	data := complex.AudioDevice{
		DeviceId:       42,
		DeviceName:     "USB Audio Interface",
		SampleRate:     48000,
		BufferSize:     512,
		InputChannels:  2,
		OutputChannels: 8,
		IsDefault:      true,
		ActivePlugins: []complex.Plugin{
			{
				Id:           1,
				Name:         "Compressor",
				Manufacturer: "AudioCorp",
				Version:      100,
				Enabled:      true,
				Parameters: []complex.Parameter{
					{Id: 1, Name: "threshold", Value: -20.0, Min: -60.0, Max: 0.0},
					{Id: 2, Name: "ratio", Value: 4.0, Min: 1.0, Max: 20.0},
					{Id: 3, Name: "attack", Value: 5.0, Min: 0.1, Max: 100.0},
				},
			},
			{
				Id:           2,
				Name:         "EQ",
				Manufacturer: "AudioCorp",
				Version:      200,
				Enabled:      false,
				Parameters:   []complex.Parameter{},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := complex.EncodeAudioDevice(&data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeComplex(b *testing.B) {
	data := complex.AudioDevice{
		DeviceId:       42,
		DeviceName:     "USB Audio Interface",
		SampleRate:     48000,
		BufferSize:     512,
		InputChannels:  2,
		OutputChannels: 8,
		IsDefault:      true,
		ActivePlugins: []complex.Plugin{
			{
				Id:           1,
				Name:         "Compressor",
				Manufacturer: "AudioCorp",
				Version:      100,
				Enabled:      true,
				Parameters: []complex.Parameter{
					{Id: 1, Name: "threshold", Value: -20.0, Min: -60.0, Max: 0.0},
					{Id: 2, Name: "ratio", Value: 4.0, Min: 1.0, Max: 20.0},
					{Id: 3, Name: "attack", Value: 5.0, Min: 0.1, Max: 100.0},
				},
			},
			{
				Id:           2,
				Name:         "EQ",
				Manufacturer: "AudioCorp",
				Version:      200,
				Enabled:      false,
				Parameters:   []complex.Parameter{},
			},
		},
	}

	// Pre-encode once
	encoded, err := complex.EncodeAudioDevice(&data)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result complex.AudioDevice
		err := complex.DecodeAudioDevice(&result, encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Benchmark encode/decode for large array (1000 elements)
func BenchmarkEncodeLargeArray(b *testing.B) {
	largeArray := make([]uint32, 1000)
	for i := range largeArray {
		largeArray[i] = uint32(i * 100)
	}

	data := arrays.ArraysOfPrimitives{
		U8Array:   []uint8{},
		U32Array:  largeArray,
		F64Array:  []float64{},
		StrArray:  []string{},
		BoolArray: []bool{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := arrays.EncodeArraysOfPrimitives(&data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDecodeLargeArray(b *testing.B) {
	largeArray := make([]uint32, 1000)
	for i := range largeArray {
		largeArray[i] = uint32(i * 100)
	}

	data := arrays.ArraysOfPrimitives{
		U8Array:   []uint8{},
		U32Array:  largeArray,
		F64Array:  []float64{},
		StrArray:  []string{},
		BoolArray: []bool{},
	}

	// Pre-encode once
	encoded, err := arrays.EncodeArraysOfPrimitives(&data)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var result arrays.ArraysOfPrimitives
		err := arrays.DecodeArraysOfPrimitives(&result, encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkIncrementalConstruction simulates building complex struct incrementally
// (like enumerating plugins/parameters in Objective-C)
func BenchmarkIncrementalConstruction(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Stage 1: Create device with initial values
		device := complex.AudioDevice{
			DeviceId:       42,
			DeviceName:     "USB Audio Interface",
			SampleRate:     48000,
			BufferSize:     512,
			InputChannels:  2,
			OutputChannels: 8,
			IsDefault:      true,
			ActivePlugins:  make([]complex.Plugin, 0, 2),
		}

		// Stage 2: Incrementally build first plugin
		plugin1 := complex.Plugin{
			Id:           1,
			Name:         "Compressor",
			Manufacturer: "AudioCorp",
			Version:      100,
			Enabled:      true,
			Parameters:   make([]complex.Parameter, 0, 3),
		}

		// Stage 3: Incrementally build parameters for plugin1
		plugin1.Parameters = append(plugin1.Parameters, complex.Parameter{
			Id: 1, Name: "threshold", Value: -20.0, Min: -60.0, Max: 0.0,
		})
		plugin1.Parameters = append(plugin1.Parameters, complex.Parameter{
			Id: 2, Name: "ratio", Value: 4.0, Min: 1.0, Max: 20.0,
		})
		plugin1.Parameters = append(plugin1.Parameters, complex.Parameter{
			Id: 3, Name: "attack", Value: 5.0, Min: 0.1, Max: 100.0,
		})

		// Add plugin1 to device
		device.ActivePlugins = append(device.ActivePlugins, plugin1)

		// Stage 4: Build second plugin
		plugin2 := complex.Plugin{
			Id:           2,
			Name:         "EQ",
			Manufacturer: "AudioCorp",
			Version:      200,
			Enabled:      false,
			Parameters:   []complex.Parameter{},
		}

		device.ActivePlugins = append(device.ActivePlugins, plugin2)

		// Stage 5: Encode the complete structure
		_, err := complex.EncodeAudioDevice(&device)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkAllAtOnceConstruction compares building struct all at once
func BenchmarkAllAtOnceConstruction(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Build everything in one literal
		device := complex.AudioDevice{
			DeviceId:       42,
			DeviceName:     "USB Audio Interface",
			SampleRate:     48000,
			BufferSize:     512,
			InputChannels:  2,
			OutputChannels: 8,
			IsDefault:      true,
			ActivePlugins: []complex.Plugin{
				{
					Id:           1,
					Name:         "Compressor",
					Manufacturer: "AudioCorp",
					Version:      100,
					Enabled:      true,
					Parameters: []complex.Parameter{
						{Id: 1, Name: "threshold", Value: -20.0, Min: -60.0, Max: 0.0},
						{Id: 2, Name: "ratio", Value: 4.0, Min: 1.0, Max: 20.0},
						{Id: 3, Name: "attack", Value: 5.0, Min: 0.1, Max: 100.0},
					},
				},
				{
					Id:           2,
					Name:         "EQ",
					Manufacturer: "AudioCorp",
					Version:      200,
					Enabled:      false,
					Parameters:   []complex.Parameter{},
				},
			},
		}

		_, err := complex.EncodeAudioDevice(&device)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRealWorldAudioUnit tests with actual AudioUnit plugin data
// This is the ULTIMATE stress test - real production data from 62 plugins with 1,759 parameters
// Comparison targets:
//   - Protocol Buffers: 1,300 µs (1.3 ms) roundtrip
//   - FlatBuffers: 1,000 µs (1.0 ms) roundtrip
func BenchmarkRealWorldAudioUnit(b *testing.B) {
	// Load real AudioUnit plugin data from JSON
	jsonData, err := os.ReadFile("testdata/plugins.json")
	if err != nil {
		b.Skipf("plugins.json not found, skipping real-world benchmark: %v", err)
		return
	}

	// Parse JSON into intermediate structure
	var pluginsJSON []struct {
		Name           string `json:"name"`
		ManufacturerID string `json:"manufacturerID"`
		Type           string `json:"type"`
		Subtype        string `json:"subtype"`
		Parameters     []struct {
			Address      uint64  `json:"address"`
			DisplayName  string  `json:"displayName"`
			Identifier   string  `json:"identifier"`
			Unit         string  `json:"unit"`
			MinValue     float32 `json:"minValue"`
			MaxValue     float32 `json:"maxValue"`
			DefaultValue float32 `json:"defaultValue"`
			CurrentValue float32 `json:"currentValue"`
			RawFlags     uint32  `json:"rawFlags"`
			IsWritable   bool    `json:"isWritable"`
			CanRamp      bool    `json:"canRamp"`
		} `json:"parameters"`
	}

	if err := json.Unmarshal(jsonData, &pluginsJSON); err != nil {
		b.Fatalf("Failed to parse plugins.json: %v", err)
	}

	// Convert to SDP structs (simulating real usage pattern)
	registry := audiounit.PluginRegistry{
		Plugins:             make([]audiounit.Plugin, 0, len(pluginsJSON)),
		TotalPluginCount:    uint32(len(pluginsJSON)),
		TotalParameterCount: 0,
	}

	totalParams := uint32(0)
	for _, pJSON := range pluginsJSON {
		plugin := audiounit.Plugin{
			Name:             pJSON.Name,
			ManufacturerId:   pJSON.ManufacturerID,
			ComponentType:    pJSON.Type,
			ComponentSubtype: pJSON.Subtype,
			Parameters:       make([]audiounit.Parameter, 0, len(pJSON.Parameters)),
		}

		for _, paramJSON := range pJSON.Parameters {
			param := audiounit.Parameter{
				Address:      paramJSON.Address,
				DisplayName:  paramJSON.DisplayName,
				Identifier:   paramJSON.Identifier,
				Unit:         paramJSON.Unit,
				MinValue:     paramJSON.MinValue,
				MaxValue:     paramJSON.MaxValue,
				DefaultValue: paramJSON.DefaultValue,
				CurrentValue: paramJSON.CurrentValue,
				RawFlags:     paramJSON.RawFlags,
				IsWritable:   paramJSON.IsWritable,
				CanRamp:      paramJSON.CanRamp,
			}
			plugin.Parameters = append(plugin.Parameters, param)
			totalParams++
		}

		registry.Plugins = append(registry.Plugins, plugin)
	}

	registry.TotalParameterCount = totalParams

	b.Logf("Loaded real-world data: %d plugins, %d parameters",
		len(registry.Plugins), registry.TotalParameterCount)

	// Now benchmark the encode→decode roundtrip (excluding JSON parsing)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Encode
		encoded, err := audiounit.EncodePluginRegistry(&registry)
		if err != nil {
			b.Fatal(err)
		}

		// Decode
		var decoded audiounit.PluginRegistry
		err = audiounit.DecodePluginRegistry(&decoded, encoded)
		if err != nil {
			b.Fatal(err)
		}

		// Verify (prevent optimization)
		if decoded.TotalPluginCount != registry.TotalPluginCount {
			b.Fatal("decode mismatch")
		}
	}
}

// BenchmarkRealWorldAudioUnitEncodeOnly measures just encoding performance
func BenchmarkRealWorldAudioUnitEncodeOnly(b *testing.B) {
	jsonData, err := os.ReadFile("testdata/plugins.json")
	if err != nil {
		b.Skipf("plugins.json not found: %v", err)
		return
	}

	var pluginsJSON []struct {
		Name           string `json:"name"`
		ManufacturerID string `json:"manufacturerID"`
		Type           string `json:"type"`
		Subtype        string `json:"subtype"`
		Parameters     []struct {
			Address      uint64  `json:"address"`
			DisplayName  string  `json:"displayName"`
			Identifier   string  `json:"identifier"`
			Unit         string  `json:"unit"`
			MinValue     float32 `json:"minValue"`
			MaxValue     float32 `json:"maxValue"`
			DefaultValue float32 `json:"defaultValue"`
			CurrentValue float32 `json:"currentValue"`
			RawFlags     uint32  `json:"rawFlags"`
			IsWritable   bool    `json:"isWritable"`
			CanRamp      bool    `json:"canRamp"`
		} `json:"parameters"`
	}

	json.Unmarshal(jsonData, &pluginsJSON)

	registry := audiounit.PluginRegistry{
		Plugins:             make([]audiounit.Plugin, 0, len(pluginsJSON)),
		TotalPluginCount:    uint32(len(pluginsJSON)),
		TotalParameterCount: 0,
	}

	totalParams := uint32(0)
	for _, pJSON := range pluginsJSON {
		plugin := audiounit.Plugin{
			Name:             pJSON.Name,
			ManufacturerId:   pJSON.ManufacturerID,
			ComponentType:    pJSON.Type,
			ComponentSubtype: pJSON.Subtype,
			Parameters:       make([]audiounit.Parameter, 0, len(pJSON.Parameters)),
		}

		for _, paramJSON := range pJSON.Parameters {
			param := audiounit.Parameter{
				Address:      paramJSON.Address,
				DisplayName:  paramJSON.DisplayName,
				Identifier:   paramJSON.Identifier,
				Unit:         paramJSON.Unit,
				MinValue:     paramJSON.MinValue,
				MaxValue:     paramJSON.MaxValue,
				DefaultValue: paramJSON.DefaultValue,
				CurrentValue: paramJSON.CurrentValue,
				RawFlags:     paramJSON.RawFlags,
				IsWritable:   paramJSON.IsWritable,
				CanRamp:      paramJSON.CanRamp,
			}
			plugin.Parameters = append(plugin.Parameters, param)
			totalParams++
		}

		registry.Plugins = append(registry.Plugins, plugin)
	}

	registry.TotalParameterCount = totalParams

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := audiounit.EncodePluginRegistry(&registry)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkRealWorldAudioUnitDecodeOnly measures just decoding performance
func BenchmarkRealWorldAudioUnitDecodeOnly(b *testing.B) {
	jsonData, err := os.ReadFile("testdata/plugins.json")
	if err != nil {
		b.Skipf("plugins.json not found: %v", err)
		return
	}

	var pluginsJSON []struct {
		Name           string `json:"name"`
		ManufacturerID string `json:"manufacturerID"`
		Type           string `json:"type"`
		Subtype        string `json:"subtype"`
		Parameters     []struct {
			Address      uint64  `json:"address"`
			DisplayName  string  `json:"displayName"`
			Identifier   string  `json:"identifier"`
			Unit         string  `json:"unit"`
			MinValue     float32 `json:"minValue"`
			MaxValue     float32 `json:"maxValue"`
			DefaultValue float32 `json:"defaultValue"`
			CurrentValue float32 `json:"currentValue"`
			RawFlags     uint32  `json:"rawFlags"`
			IsWritable   bool    `json:"isWritable"`
			CanRamp      bool    `json:"canRamp"`
		} `json:"parameters"`
	}

	json.Unmarshal(jsonData, &pluginsJSON)

	registry := audiounit.PluginRegistry{
		Plugins:             make([]audiounit.Plugin, 0, len(pluginsJSON)),
		TotalPluginCount:    uint32(len(pluginsJSON)),
		TotalParameterCount: 0,
	}

	totalParams := uint32(0)
	for _, pJSON := range pluginsJSON {
		plugin := audiounit.Plugin{
			Name:             pJSON.Name,
			ManufacturerId:   pJSON.ManufacturerID,
			ComponentType:    pJSON.Type,
			ComponentSubtype: pJSON.Subtype,
			Parameters:       make([]audiounit.Parameter, 0, len(pJSON.Parameters)),
		}

		for _, paramJSON := range pJSON.Parameters {
			param := audiounit.Parameter{
				Address:      paramJSON.Address,
				DisplayName:  paramJSON.DisplayName,
				Identifier:   paramJSON.Identifier,
				Unit:         paramJSON.Unit,
				MinValue:     paramJSON.MinValue,
				MaxValue:     paramJSON.MaxValue,
				DefaultValue: paramJSON.DefaultValue,
				CurrentValue: paramJSON.CurrentValue,
				RawFlags:     paramJSON.RawFlags,
				IsWritable:   paramJSON.IsWritable,
				CanRamp:      paramJSON.CanRamp,
			}
			plugin.Parameters = append(plugin.Parameters, param)
			totalParams++
		}

		registry.Plugins = append(registry.Plugins, plugin)
	}

	registry.TotalParameterCount = totalParams

	// Pre-encode once
	encoded, _ := audiounit.EncodePluginRegistry(&registry)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var decoded audiounit.PluginRegistry
		err := audiounit.DecodePluginRegistry(&decoded, encoded)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// ============================================================================
// Optional Field Tests (RC Feature 1)
// ============================================================================

func TestRoundtripOptionalPresent(t *testing.T) {
	// Test with optional field present
	original := optional.Request{
		Id: 42,
		Metadata: &optional.Metadata{
			UserId:   12345,
			Username: "alice",
		},
	}

	// Encode
	encoded, err := optional.EncodeRequest(&original)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	// Decode
	var decoded optional.Request
	err = optional.DecodeRequest(&decoded, encoded)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	// Verify
	if decoded.Id != original.Id {
		t.Errorf("Id mismatch: got %d, want %d", decoded.Id, original.Id)
	}

	if decoded.Metadata == nil {
		t.Fatal("Metadata should not be nil")
	}

	if decoded.Metadata.UserId != original.Metadata.UserId {
		t.Errorf("UserId mismatch: got %d, want %d", decoded.Metadata.UserId, original.Metadata.UserId)
	}

	if decoded.Metadata.Username != original.Metadata.Username {
		t.Errorf("Username mismatch: got %q, want %q", decoded.Metadata.Username, original.Metadata.Username)
	}
}

func TestRoundtripOptionalAbsent(t *testing.T) {
	// Test with optional field absent (nil)
	original := optional.Request{
		Id:       99,
		Metadata: nil,
	}

	// Encode
	encoded, err := optional.EncodeRequest(&original)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	// Decode
	var decoded optional.Request
	err = optional.DecodeRequest(&decoded, encoded)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	// Verify
	if decoded.Id != original.Id {
		t.Errorf("Id mismatch: got %d, want %d", decoded.Id, original.Id)
	}

	if decoded.Metadata != nil {
		t.Errorf("Metadata should be nil, got %+v", decoded.Metadata)
	}
}

func TestRoundtripMultipleOptionals(t *testing.T) {
	// Test with multiple optional fields (some present, some absent)
	original := optional.Config{
		Name: "production",
		Database: &optional.DatabaseConfig{
			Host: "db.example.com",
			Port: 5432,
		},
		Cache: nil, // Cache is absent
	}

	// Encode
	encoded, err := optional.EncodeConfig(&original)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	// Decode
	var decoded optional.Config
	err = optional.DecodeConfig(&decoded, encoded)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	// Verify
	if decoded.Name != original.Name {
		t.Errorf("Name mismatch: got %q, want %q", decoded.Name, original.Name)
	}

	if decoded.Database == nil {
		t.Fatal("Database should not be nil")
	}

	if decoded.Database.Host != original.Database.Host {
		t.Errorf("Database.Host mismatch: got %q, want %q", decoded.Database.Host, original.Database.Host)
	}

	if decoded.Database.Port != original.Database.Port {
		t.Errorf("Database.Port mismatch: got %d, want %d", decoded.Database.Port, original.Database.Port)
	}

	if decoded.Cache != nil {
		t.Errorf("Cache should be nil, got %+v", decoded.Cache)
	}
}

func TestOptionalWireFormat(t *testing.T) {
	// Test that the wire format includes presence flags

	// Test 1: Optional present (presence flag = 1)
	withMetadata := optional.Request{
		Id: 42,
		Metadata: &optional.Metadata{
			UserId:   100,
			Username: "bob",
		},
	}

	encoded, err := optional.EncodeRequest(&withMetadata)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	// Wire format should be:
	// - Id: u32 (4 bytes)
	// - Presence flag: u8 (1 byte) = 1
	// - UserId: u64 (8 bytes)
	// - Username length: u32 (4 bytes)
	// - Username data: 3 bytes ("bob")
	expectedSize := 4 + 1 + 8 + 4 + 3
	if len(encoded) != expectedSize {
		t.Errorf("encoded size mismatch: got %d, want %d", len(encoded), expectedSize)
	}

	// Check presence flag (byte 4)
	presenceFlag := encoded[4]
	if presenceFlag != 1 {
		t.Errorf("presence flag should be 1, got %d", presenceFlag)
	}

	// Test 2: Optional absent (presence flag = 0)
	withoutMetadata := optional.Request{
		Id:       99,
		Metadata: nil,
	}

	encoded2, err := optional.EncodeRequest(&withoutMetadata)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	// Wire format should be:
	// - Id: u32 (4 bytes)
	// - Presence flag: u8 (1 byte) = 0
	expectedSize2 := 4 + 1
	if len(encoded2) != expectedSize2 {
		t.Errorf("encoded size mismatch: got %d, want %d", len(encoded2), expectedSize2)
	}

	// Check presence flag (byte 4)
	presenceFlag2 := encoded2[4]
	if presenceFlag2 != 0 {
		t.Errorf("presence flag should be 0, got %d", presenceFlag2)
	}
}

func TestOptionalWithArray(t *testing.T) {
	// Test optional field that contains an array
	original := optional.Document{
		Id: 42,
		Tags: &optional.TagList{
			Items: []string{"urgent", "reviewed", "approved"},
		},
	}

	// Encode
	encoded, err := optional.EncodeDocument(&original)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	// Decode
	var decoded optional.Document
	err = optional.DecodeDocument(&decoded, encoded)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	// Verify
	if decoded.Id != original.Id {
		t.Errorf("Id mismatch: got %d, want %d", decoded.Id, original.Id)
	}

	if decoded.Tags == nil {
		t.Fatal("Tags should not be nil")
	}

	if len(decoded.Tags.Items) != len(original.Tags.Items) {
		t.Fatalf("Tags.Items length mismatch: got %d, want %d", len(decoded.Tags.Items), len(original.Tags.Items))
	}

	for i, tag := range decoded.Tags.Items {
		if tag != original.Tags.Items[i] {
			t.Errorf("Tags.Items[%d] mismatch: got %q, want %q", i, tag, original.Tags.Items[i])
		}
	}
}

func TestOptionalWithArrayAbsent(t *testing.T) {
	// Test optional field with array when absent
	original := optional.Document{
		Id:   99,
		Tags: nil,
	}

	// Encode
	encoded, err := optional.EncodeDocument(&original)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}

	// Decode
	var decoded optional.Document
	err = optional.DecodeDocument(&decoded, encoded)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	// Verify
	if decoded.Id != original.Id {
		t.Errorf("Id mismatch: got %d, want %d", decoded.Id, original.Id)
	}

	if decoded.Tags != nil {
		t.Errorf("Tags should be nil, got %+v", decoded.Tags)
	}
}

func TestOptionalDecodeErrors(t *testing.T) {
	// Test decode error handling for optional fields

	// Test 1: Invalid presence flag (not 0 or 1)
	badData := []byte{
		0x2A, 0x00, 0x00, 0x00, // Id = 42
		0x99, // Invalid presence flag (should be 0 or 1)
	}

	var decoded optional.Request
	err := optional.DecodeRequest(&decoded, badData)
	if err == nil {
		t.Error("decode should fail with invalid presence flag")
	}

	// Test 2: Truncated data (presence flag present but no data after)
	truncated := []byte{
		0x2A, 0x00, 0x00, 0x00, // Id = 42
		0x01, // Presence flag = 1 (present)
		// Missing Metadata data
	}

	var decoded2 optional.Request
	err = optional.DecodeRequest(&decoded2, truncated)
	if err == nil {
		t.Error("decode should fail with truncated data")
	}
}
