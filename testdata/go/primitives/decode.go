package primitives

import (
	"math"
	"io"
	"encoding/binary"
)

// Size limit constants for decode validation
const (
	MaxSerializedSize = 128 * 1024 * 1024
	MaxArrayElements  = 1_000_000
	MaxTotalElements  = 10_000_000
)

// DecodeContext tracks state during decoding to enforce size limits.
// It maintains a count of total elements across all arrays to prevent
// excessive memory allocation from malicious or corrupted data.
type DecodeContext struct {
	totalElements int
}

// checkArraySize validates an array count against per-array and total limits.
// It returns ErrArrayTooLarge if the count exceeds MaxArrayElements, or
// ErrTooManyElements if the cumulative total exceeds MaxTotalElements.
func (ctx *DecodeContext) checkArraySize(count uint32) error {
	if count > MaxArrayElements {
		return ErrArrayTooLarge
	}

	ctx.totalElements += int(count)
	if ctx.totalElements > MaxTotalElements {
		return ErrTooManyElements
	}

	return nil
}


// DecodeAllPrimitives decodes a AllPrimitives from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodeAllPrimitives(dest *AllPrimitives, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodeAllPrimitives(dest, data, &offset, ctx)
}


// decodeAllPrimitives is the helper function that decodes AllPrimitives fields.
func decodeAllPrimitives(dest *AllPrimitives, data []byte, offset *int, ctx *DecodeContext) error {
	var (
		strLen uint32  // For string length prefix
		arrCount uint32  // For array count
		err error  // For error handling
	)
	_ = strLen  // Avoid unused variable error
	_ = arrCount  // Avoid unused variable error
	_ = err  // Avoid unused variable error

	// Field: U8Field (u8)
	if *offset + 1 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.U8Field = uint8(data[*offset])
	*offset += 1

	// Field: U16Field (u16)
	if *offset + 2 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.U16Field = binary.LittleEndian.Uint16(data[*offset:])
	*offset += 2

	// Field: U32Field (u32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.U32Field = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	// Field: U64Field (u64)
	if *offset + 8 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.U64Field = binary.LittleEndian.Uint64(data[*offset:])
	*offset += 8

	// Field: I8Field (i8)
	if *offset + 1 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.I8Field = int8(data[*offset])
	*offset += 1

	// Field: I16Field (i16)
	if *offset + 2 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.I16Field = int16(binary.LittleEndian.Uint16(data[*offset:]))
	*offset += 2

	// Field: I32Field (i32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.I32Field = int32(binary.LittleEndian.Uint32(data[*offset:]))
	*offset += 4

	// Field: I64Field (i64)
	if *offset + 8 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.I64Field = int64(binary.LittleEndian.Uint64(data[*offset:]))
	*offset += 8

	// Field: F32Field (f32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.F32Field = math.Float32frombits(binary.LittleEndian.Uint32(data[*offset:]))
	*offset += 4

	// Field: F64Field (f64)
	if *offset + 8 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.F64Field = math.Float64frombits(binary.LittleEndian.Uint64(data[*offset:]))
	*offset += 8

	// Field: BoolField (bool)
	if *offset + 1 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.BoolField = data[*offset] != 0
	*offset += 1

	// Field: StrField (str)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	strLen = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	if *offset + int(strLen) > len(data) {
		return ErrUnexpectedEOF
	}
	dest.StrField = string(data[*offset:*offset+int(strLen)])
	*offset += int(strLen)

	return nil
}


// DecodeAllPrimitivesMessage decodes a AllPrimitives from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodeAllPrimitivesMessage(data []byte) (*AllPrimitives, error) {
	// Check minimum message size
	if len(data) < MessageHeaderSize {
		return nil, ErrUnexpectedEOF
	}

	// Validate magic bytes
	if string(data[0:3]) != MessageMagic {
		return nil, ErrInvalidMagic
	}

	// Validate protocol version
	if data[3] != MessageVersion {
		return nil, ErrInvalidVersion
	}

	// Validate type ID
	typeID := binary.LittleEndian.Uint16(data[4:6])
	if typeID != 1 {
		return nil, ErrUnknownMessageType
	}

	// Extract payload length
	payloadLength := binary.LittleEndian.Uint32(data[6:10])

	// Validate total message size
	expectedSize := MessageHeaderSize + int(payloadLength)
	if len(data) < expectedSize {
		return nil, ErrUnexpectedEOF
	}

	// Extract payload
	payload := data[MessageHeaderSize:expectedSize]

	// Decode payload
	var result AllPrimitives
	if err := DecodeAllPrimitives(&result, payload); err != nil {
		return nil, err
	}

	return &result, nil
}



// DecodeMessage decodes a message and returns the struct type based on the type ID in the header.
// This is the main entry point for decoding self-describing messages.
// Returns the decoded struct as an interface{} which can be type-asserted to the specific type.
func DecodeMessage(data []byte) (interface{}, error) {
	// Check minimum message size
	if len(data) < MessageHeaderSize {
		return nil, ErrUnexpectedEOF
	}

	// Validate magic bytes
	if string(data[0:3]) != MessageMagic {
		return nil, ErrInvalidMagic
	}

	// Validate protocol version
	if data[3] != MessageVersion {
		return nil, ErrInvalidVersion
	}

	// Extract type ID
	typeID := binary.LittleEndian.Uint16(data[4:6])

	// Dispatch to specific decoder
	switch typeID {
	case 1:
		return DecodeAllPrimitivesMessage(data)
	default:
		return nil, ErrUnknownMessageType
	}
}


// DecodeAllPrimitivesFromReader decodes a AllPrimitives from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodeAllPrimitivesFromReader(dest *AllPrimitives, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodeAllPrimitives(dest, buf)
}
