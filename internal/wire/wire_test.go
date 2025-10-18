package wire

import (
	"bytes"
	"math"
	"testing"
)

func TestEncodeDecodeU8(t *testing.T) {
	buf := make([]byte, 1)
	testCases := []uint8{0, 1, 127, 255}

	for _, val := range testCases {
		EncodeU8(buf, 0, val)
		got := DecodeU8(buf, 0)
		if got != val {
			t.Errorf("U8: encoded %d, decoded %d", val, got)
		}
	}
}

func TestEncodeDecodeU16(t *testing.T) {
	buf := make([]byte, 2)
	testCases := []uint16{0, 1, 255, 256, 32767, 65535}

	for _, val := range testCases {
		EncodeU16(buf, 0, val)
		got := DecodeU16(buf, 0)
		if got != val {
			t.Errorf("U16: encoded %d, decoded %d", val, got)
		}
	}
}

func TestEncodeDecodeU32(t *testing.T) {
	buf := make([]byte, 4)
	testCases := []uint32{0, 1, 255, 256, 65535, 65536, 0x7FFFFFFF, 0xFFFFFFFF}

	for _, val := range testCases {
		EncodeU32(buf, 0, val)
		got := DecodeU32(buf, 0)
		if got != val {
			t.Errorf("U32: encoded %d, decoded %d", val, got)
		}
	}
}

func TestEncodeDecodeU64(t *testing.T) {
	buf := make([]byte, 8)
	testCases := []uint64{
		0, 1, 255, 256, 65535, 65536,
		0x7FFFFFFF, 0xFFFFFFFF,
		0x100000000, 0x7FFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF,
	}

	for _, val := range testCases {
		EncodeU64(buf, 0, val)
		got := DecodeU64(buf, 0)
		if got != val {
			t.Errorf("U64: encoded %d, decoded %d", val, got)
		}
	}
}

func TestEncodeDecodeI8(t *testing.T) {
	buf := make([]byte, 1)
	testCases := []int8{-128, -1, 0, 1, 127}

	for _, val := range testCases {
		EncodeI8(buf, 0, val)
		got := DecodeI8(buf, 0)
		if got != val {
			t.Errorf("I8: encoded %d, decoded %d", val, got)
		}
	}
}

func TestEncodeDecodeI16(t *testing.T) {
	buf := make([]byte, 2)
	testCases := []int16{-32768, -1, 0, 1, 32767}

	for _, val := range testCases {
		EncodeI16(buf, 0, val)
		got := DecodeI16(buf, 0)
		if got != val {
			t.Errorf("I16: encoded %d, decoded %d", val, got)
		}
	}
}

func TestEncodeDecodeI32(t *testing.T) {
	buf := make([]byte, 4)
	testCases := []int32{-2147483648, -1, 0, 1, 2147483647}

	for _, val := range testCases {
		EncodeI32(buf, 0, val)
		got := DecodeI32(buf, 0)
		if got != val {
			t.Errorf("I32: encoded %d, decoded %d", val, got)
		}
	}
}

func TestEncodeDecodeI64(t *testing.T) {
	buf := make([]byte, 8)
	testCases := []int64{
		-9223372036854775808, -2147483648, -1, 0, 1,
		2147483647, 9223372036854775807,
	}

	for _, val := range testCases {
		EncodeI64(buf, 0, val)
		got := DecodeI64(buf, 0)
		if got != val {
			t.Errorf("I64: encoded %d, decoded %d", val, got)
		}
	}
}

func TestEncodeDecodeF32(t *testing.T) {
	buf := make([]byte, 4)
	testCases := []float32{
		0.0, 1.0, -1.0, 3.14159, -3.14159,
		math.MaxFloat32, -math.MaxFloat32,
		math.SmallestNonzeroFloat32,
		float32(math.Inf(1)), float32(math.Inf(-1)),
	}

	for _, val := range testCases {
		EncodeF32(buf, 0, val)
		got := DecodeF32(buf, 0)
		if got != val && !(math.IsNaN(float64(val)) && math.IsNaN(float64(got))) {
			t.Errorf("F32: encoded %v, decoded %v", val, got)
		}
	}

	// Test NaN separately (NaN != NaN)
	EncodeF32(buf, 0, float32(math.NaN()))
	got := DecodeF32(buf, 0)
	if !math.IsNaN(float64(got)) {
		t.Errorf("F32: encoded NaN, decoded %v", got)
	}
}

func TestEncodeDecodeF64(t *testing.T) {
	buf := make([]byte, 8)
	testCases := []float64{
		0.0, 1.0, -1.0, 3.14159265358979323846,
		math.MaxFloat64, -math.MaxFloat64,
		math.SmallestNonzeroFloat64,
		math.Inf(1), math.Inf(-1),
	}

	for _, val := range testCases {
		EncodeF64(buf, 0, val)
		got := DecodeF64(buf, 0)
		if got != val && !(math.IsNaN(val) && math.IsNaN(got)) {
			t.Errorf("F64: encoded %v, decoded %v", val, got)
		}
	}

	// Test NaN separately (NaN != NaN)
	EncodeF64(buf, 0, math.NaN())
	got := DecodeF64(buf, 0)
	if !math.IsNaN(got) {
		t.Errorf("F64: encoded NaN, decoded %v", got)
	}
}

func TestEncodeBool(t *testing.T) {
	buf := make([]byte, 1)

	EncodeBool(buf, 0, true)
	if buf[0] != 1 {
		t.Errorf("EncodeBool(true): expected 1, got %d", buf[0])
	}

	EncodeBool(buf, 0, false)
	if buf[0] != 0 {
		t.Errorf("EncodeBool(false): expected 0, got %d", buf[0])
	}
}

func TestDecodeBool(t *testing.T) {
	buf := make([]byte, 1)

	buf[0] = 0
	if DecodeBool(buf, 0) != false {
		t.Errorf("DecodeBool(0): expected false")
	}

	buf[0] = 1
	if DecodeBool(buf, 0) != true {
		t.Errorf("DecodeBool(1): expected true")
	}

	// Any non-zero value should decode as true
	buf[0] = 255
	if DecodeBool(buf, 0) != true {
		t.Errorf("DecodeBool(255): expected true")
	}

	buf[0] = 42
	if DecodeBool(buf, 0) != true {
		t.Errorf("DecodeBool(42): expected true")
	}
}

// TestLittleEndian verifies that multi-byte values are encoded in little-endian order
func TestLittleEndian(t *testing.T) {
	buf := make([]byte, 8)

	// U16: 0x1234 should be [0x34, 0x12]
	EncodeU16(buf, 0, 0x1234)
	if buf[0] != 0x34 || buf[1] != 0x12 {
		t.Errorf("U16 little-endian: expected [0x34, 0x12], got [0x%02x, 0x%02x]", buf[0], buf[1])
	}

	// U32: 0x12345678 should be [0x78, 0x56, 0x34, 0x12]
	EncodeU32(buf, 0, 0x12345678)
	if buf[0] != 0x78 || buf[1] != 0x56 || buf[2] != 0x34 || buf[3] != 0x12 {
		t.Errorf("U32 little-endian: expected [0x78, 0x56, 0x34, 0x12], got [0x%02x, 0x%02x, 0x%02x, 0x%02x]",
			buf[0], buf[1], buf[2], buf[3])
	}

	// U64: 0x123456789ABCDEF0 should be [0xF0, 0xDE, 0xBC, 0x9A, 0x78, 0x56, 0x34, 0x12]
	EncodeU64(buf, 0, 0x123456789ABCDEF0)
	expected := []byte{0xF0, 0xDE, 0xBC, 0x9A, 0x78, 0x56, 0x34, 0x12}
	for i := 0; i < 8; i++ {
		if buf[i] != expected[i] {
			t.Errorf("U64 little-endian: byte %d expected 0x%02x, got 0x%02x", i, expected[i], buf[i])
		}
	}
}

// TestOffsets verifies that encoding/decoding works at non-zero offsets
func TestOffsets(t *testing.T) {
	buf := make([]byte, 20)

	// Write different values at different offsets
	EncodeU8(buf, 0, 0xAA)
	EncodeU16(buf, 1, 0x1234)
	EncodeU32(buf, 3, 0x56789ABC)
	EncodeU64(buf, 7, 0xDEADBEEFCAFEBABE)
	EncodeBool(buf, 15, true)

	// Verify all values decode correctly
	if DecodeU8(buf, 0) != 0xAA {
		t.Errorf("U8 at offset 0: expected 0xAA")
	}
	if DecodeU16(buf, 1) != 0x1234 {
		t.Errorf("U16 at offset 1: expected 0x1234")
	}
	if DecodeU32(buf, 3) != 0x56789ABC {
		t.Errorf("U32 at offset 3: expected 0x56789ABC")
	}
	if DecodeU64(buf, 7) != 0xDEADBEEFCAFEBABE {
		t.Errorf("U64 at offset 7: expected 0xDEADBEEFCAFEBABE")
	}
	if !DecodeBool(buf, 15) {
		t.Errorf("Bool at offset 15: expected true")
	}
}

func TestEncodeDecodeString(t *testing.T) {
	testCases := []string{
		"",
		"hello",
		"Hello, World!",
		"Unicode: 日本語 🎉",
		"Newlines\nand\ttabs",
		string(make([]byte, 1000)), // Large string
	}

	for _, val := range testCases {
		var buf bytes.Buffer

		// Encode
		n, err := EncodeString(&buf, val)
		if err != nil {
			t.Errorf("EncodeString failed: %v", err)
			continue
		}

		expectedSize := 4 + len(val)
		if n != expectedSize {
			t.Errorf("EncodeString: expected %d bytes written, got %d", expectedSize, n)
		}

		// Decode
		got, err := DecodeString(&buf)
		if err != nil {
			t.Errorf("DecodeString failed: %v", err)
			continue
		}

		if got != val {
			t.Errorf("String: encoded %q, decoded %q", val, got)
		}
	}
}

func TestEncodeDecodeStringFormat(t *testing.T) {
	// Verify wire format: [u32 length][bytes]
	var buf bytes.Buffer

	_, err := EncodeString(&buf, "abc")
	if err != nil {
		t.Fatalf("EncodeString failed: %v", err)
	}

	data := buf.Bytes()

	// Check length prefix
	if len(data) != 7 { // 4 bytes length + 3 bytes string
		t.Errorf("Expected 7 bytes, got %d", len(data))
	}

	// Verify length is little-endian u32 = 3
	length := DecodeU32(data, 0)
	if length != 3 {
		t.Errorf("Expected length 3, got %d", length)
	}

	// Verify string bytes
	if string(data[4:]) != "abc" {
		t.Errorf("Expected 'abc', got %q", data[4:])
	}
}

func TestEncodeDecodeArrayHeader(t *testing.T) {
	testCases := []uint32{0, 1, 10, 100, 1000, 1000000}

	for _, count := range testCases {
		var buf bytes.Buffer

		// Encode
		n, err := EncodeArrayHeader(&buf, count)
		if err != nil {
			t.Errorf("EncodeArrayHeader failed: %v", err)
			continue
		}

		if n != 4 {
			t.Errorf("EncodeArrayHeader: expected 4 bytes written, got %d", n)
		}

		// Decode
		got, err := DecodeArrayHeader(&buf)
		if err != nil {
			t.Errorf("DecodeArrayHeader failed: %v", err)
			continue
		}

		if got != count {
			t.Errorf("ArrayHeader: encoded %d, decoded %d", count, got)
		}
	}
}

func TestDecodeStringErrors(t *testing.T) {
	// Test incomplete length
	buf := bytes.NewBuffer([]byte{0x01, 0x02})
	_, err := DecodeString(buf)
	if err == nil {
		t.Error("Expected error for incomplete length, got nil")
	}

	// Test incomplete string data
	buf = bytes.NewBuffer([]byte{0x05, 0x00, 0x00, 0x00, 0x61, 0x62}) // length=5, only 2 bytes
	_, err = DecodeString(buf)
	if err == nil {
		t.Error("Expected error for incomplete string data, got nil")
	}
}

func TestDecodeArrayHeaderErrors(t *testing.T) {
	// Test incomplete count
	buf := bytes.NewBuffer([]byte{0x01, 0x02})
	_, err := DecodeArrayHeader(buf)
	if err == nil {
		t.Error("Expected error for incomplete count, got nil")
	}
}
