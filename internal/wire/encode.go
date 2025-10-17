// Package wire provides low-level encoding and decoding functions for the
// Serial Data Protocol wire format. All multi-byte values use little-endian
// byte order.
package wire

import (
	"encoding/binary"
	"io"
	"math"
)

// EncodeU8 writes an 8-bit unsigned integer to the buffer at the given offset.
func EncodeU8(buf []byte, offset int, val uint8) {
	buf[offset] = val
}

// EncodeU16 writes a 16-bit unsigned integer to the buffer at the given offset
// in little-endian byte order.
func EncodeU16(buf []byte, offset int, val uint16) {
	binary.LittleEndian.PutUint16(buf[offset:], val)
}

// EncodeU32 writes a 32-bit unsigned integer to the buffer at the given offset
// in little-endian byte order.
func EncodeU32(buf []byte, offset int, val uint32) {
	binary.LittleEndian.PutUint32(buf[offset:], val)
}

// EncodeU64 writes a 64-bit unsigned integer to the buffer at the given offset
// in little-endian byte order.
func EncodeU64(buf []byte, offset int, val uint64) {
	binary.LittleEndian.PutUint64(buf[offset:], val)
}

// EncodeI8 writes an 8-bit signed integer to the buffer at the given offset.
func EncodeI8(buf []byte, offset int, val int8) {
	buf[offset] = uint8(val)
}

// EncodeI16 writes a 16-bit signed integer to the buffer at the given offset
// in little-endian byte order.
func EncodeI16(buf []byte, offset int, val int16) {
	binary.LittleEndian.PutUint16(buf[offset:], uint16(val))
}

// EncodeI32 writes a 32-bit signed integer to the buffer at the given offset
// in little-endian byte order.
func EncodeI32(buf []byte, offset int, val int32) {
	binary.LittleEndian.PutUint32(buf[offset:], uint32(val))
}

// EncodeI64 writes a 64-bit signed integer to the buffer at the given offset
// in little-endian byte order.
func EncodeI64(buf []byte, offset int, val int64) {
	binary.LittleEndian.PutUint64(buf[offset:], uint64(val))
}

// EncodeF32 writes a 32-bit floating point value to the buffer at the given
// offset in little-endian byte order (IEEE 754 binary32).
func EncodeF32(buf []byte, offset int, val float32) {
	bits := math.Float32bits(val)
	binary.LittleEndian.PutUint32(buf[offset:], bits)
}

// EncodeF64 writes a 64-bit floating point value to the buffer at the given
// offset in little-endian byte order (IEEE 754 binary64).
func EncodeF64(buf []byte, offset int, val float64) {
	bits := math.Float64bits(val)
	binary.LittleEndian.PutUint64(buf[offset:], bits)
}

// EncodeBool writes a boolean value to the buffer at the given offset.
// true is encoded as 1, false as 0.
func EncodeBool(buf []byte, offset int, val bool) {
	if val {
		buf[offset] = 1
	} else {
		buf[offset] = 0
	}
}

// EncodeString writes a string to the writer in wire format.
// Format: [u32: length][utf8_bytes]
// Returns the number of bytes written and any error.
func EncodeString(w io.Writer, s string) (int, error) {
	// Write length prefix (4 bytes)
	lenBuf := make([]byte, 4)
	EncodeU32(lenBuf, 0, uint32(len(s)))
	n, err := w.Write(lenBuf)
	if err != nil {
		return n, err
	}
	
	// Write string bytes
	m, err := w.Write([]byte(s))
	return n + m, err
}

// EncodeArrayHeader writes an array count prefix to the writer.
// Format: [u32: count]
// Returns the number of bytes written and any error.
func EncodeArrayHeader(w io.Writer, count uint32) (int, error) {
	buf := make([]byte, 4)
	EncodeU32(buf, 0, count)
	return w.Write(buf)
}
