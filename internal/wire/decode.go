package wire

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
)

// DecodeU8 reads an 8-bit unsigned integer from the buffer at the given offset.
func DecodeU8(buf []byte, offset int) uint8 {
	return buf[offset]
}

// DecodeU16 reads a 16-bit unsigned integer from the buffer at the given offset
// in little-endian byte order.
func DecodeU16(buf []byte, offset int) uint16 {
	return binary.LittleEndian.Uint16(buf[offset:])
}

// DecodeU32 reads a 32-bit unsigned integer from the buffer at the given offset
// in little-endian byte order.
func DecodeU32(buf []byte, offset int) uint32 {
	return binary.LittleEndian.Uint32(buf[offset:])
}

// DecodeU64 reads a 64-bit unsigned integer from the buffer at the given offset
// in little-endian byte order.
func DecodeU64(buf []byte, offset int) uint64 {
	return binary.LittleEndian.Uint64(buf[offset:])
}

// DecodeI8 reads an 8-bit signed integer from the buffer at the given offset.
func DecodeI8(buf []byte, offset int) int8 {
	return int8(buf[offset])
}

// DecodeI16 reads a 16-bit signed integer from the buffer at the given offset
// in little-endian byte order.
func DecodeI16(buf []byte, offset int) int16 {
	return int16(binary.LittleEndian.Uint16(buf[offset:]))
}

// DecodeI32 reads a 32-bit signed integer from the buffer at the given offset
// in little-endian byte order.
func DecodeI32(buf []byte, offset int) int32 {
	return int32(binary.LittleEndian.Uint32(buf[offset:]))
}

// DecodeI64 reads a 64-bit signed integer from the buffer at the given offset
// in little-endian byte order.
func DecodeI64(buf []byte, offset int) int64 {
	return int64(binary.LittleEndian.Uint64(buf[offset:]))
}

// DecodeF32 reads a 32-bit floating point value from the buffer at the given
// offset in little-endian byte order (IEEE 754 binary32).
func DecodeF32(buf []byte, offset int) float32 {
	bits := binary.LittleEndian.Uint32(buf[offset:])
	return math.Float32frombits(bits)
}

// DecodeF64 reads a 64-bit floating point value from the buffer at the given
// offset in little-endian byte order (IEEE 754 binary64).
func DecodeF64(buf []byte, offset int) float64 {
	bits := binary.LittleEndian.Uint64(buf[offset:])
	return math.Float64frombits(bits)
}

// DecodeBool reads a boolean value from the buffer at the given offset.
// 0 is decoded as false, any non-zero value as true.
func DecodeBool(buf []byte, offset int) bool {
	return buf[offset] != 0
}

// DecodeString reads a string from the reader in wire format.
// Format: [u32: length][utf8_bytes]
// Returns the decoded string and any error.
func DecodeString(r io.Reader) (string, error) {
	// Read length prefix (4 bytes)
	lenBuf := make([]byte, 4)
	if _, err := io.ReadFull(r, lenBuf); err != nil {
		return "", fmt.Errorf("failed to read string length: %w", err)
	}
	
	length := DecodeU32(lenBuf, 0)
	
	// Read string bytes
	if length == 0 {
		return "", nil
	}
	
	strBuf := make([]byte, length)
	if _, err := io.ReadFull(r, strBuf); err != nil {
		return "", fmt.Errorf("failed to read string data: %w", err)
	}
	
	return string(strBuf), nil
}

// DecodeArrayHeader reads an array count prefix from the reader.
// Format: [u32: count]
// Returns the count and any error.
func DecodeArrayHeader(r io.Reader) (uint32, error) {
	buf := make([]byte, 4)
	if _, err := io.ReadFull(r, buf); err != nil {
		return 0, fmt.Errorf("failed to read array count: %w", err)
	}
	return DecodeU32(buf, 0), nil
}
