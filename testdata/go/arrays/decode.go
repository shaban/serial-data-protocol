package arrays

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


// DecodeArraysOfPrimitives decodes a ArraysOfPrimitives from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodeArraysOfPrimitives(dest *ArraysOfPrimitives, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodeArraysOfPrimitives(dest, data, &offset, ctx)
}

// DecodeItem decodes a Item from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodeItem(dest *Item, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodeItem(dest, data, &offset, ctx)
}

// DecodeArraysOfStructs decodes a ArraysOfStructs from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodeArraysOfStructs(dest *ArraysOfStructs, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodeArraysOfStructs(dest, data, &offset, ctx)
}


// decodeArraysOfPrimitives is the helper function that decodes ArraysOfPrimitives fields.
func decodeArraysOfPrimitives(dest *ArraysOfPrimitives, data []byte, offset *int, ctx *DecodeContext) error {
	var (
		strLen uint32  // For string length prefix
		arrCount uint32  // For array count
		err error  // For error handling
	)
	_ = strLen  // Avoid unused variable error
	_ = arrCount  // Avoid unused variable error
	_ = err  // Avoid unused variable error

	// Field: U8Array ([]u8)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	arrCount = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	err = ctx.checkArraySize(arrCount)
	if err != nil {
		return err
	}

	dest.U8Array = make([]uint8, arrCount)
	for i := uint32(0); i < arrCount; i++ {
		if *offset + 1 > len(data) {
			return ErrUnexpectedEOF
		}
		dest.U8Array[i] = uint8(data[*offset])
		*offset += 1
	}

	// Field: U32Array ([]u32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	arrCount = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	err = ctx.checkArraySize(arrCount)
	if err != nil {
		return err
	}

	dest.U32Array = make([]uint32, arrCount)
	for i := uint32(0); i < arrCount; i++ {
		if *offset + 4 > len(data) {
			return ErrUnexpectedEOF
		}
		dest.U32Array[i] = binary.LittleEndian.Uint32(data[*offset:])
		*offset += 4
	}

	// Field: F64Array ([]f64)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	arrCount = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	err = ctx.checkArraySize(arrCount)
	if err != nil {
		return err
	}

	dest.F64Array = make([]float64, arrCount)
	for i := uint32(0); i < arrCount; i++ {
		if *offset + 8 > len(data) {
			return ErrUnexpectedEOF
		}
		dest.F64Array[i] = math.Float64frombits(binary.LittleEndian.Uint64(data[*offset:]))
		*offset += 8
	}

	// Field: StrArray ([]str)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	arrCount = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	err = ctx.checkArraySize(arrCount)
	if err != nil {
		return err
	}

	dest.StrArray = make([]string, arrCount)
	for i := uint32(0); i < arrCount; i++ {
		if *offset + 4 > len(data) {
			return ErrUnexpectedEOF
		}
		strLen := binary.LittleEndian.Uint32(data[*offset:])
		*offset += 4
		
		if *offset + int(strLen) > len(data) {
			return ErrUnexpectedEOF
		}
		dest.StrArray[i] = string(data[*offset:*offset+int(strLen)])
		*offset += int(strLen)
	}

	// Field: BoolArray ([]bool)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	arrCount = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	err = ctx.checkArraySize(arrCount)
	if err != nil {
		return err
	}

	dest.BoolArray = make([]bool, arrCount)
	for i := uint32(0); i < arrCount; i++ {
		if *offset + 1 > len(data) {
			return ErrUnexpectedEOF
		}
		dest.BoolArray[i] = data[*offset] != 0
		*offset += 1
	}

	return nil
}

// decodeItem is the helper function that decodes Item fields.
func decodeItem(dest *Item, data []byte, offset *int, ctx *DecodeContext) error {
	var (
		strLen uint32  // For string length prefix
		arrCount uint32  // For array count
		err error  // For error handling
	)
	_ = strLen  // Avoid unused variable error
	_ = arrCount  // Avoid unused variable error
	_ = err  // Avoid unused variable error

	// Field: Id (u32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Id = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	// Field: Name (str)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	strLen = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	if *offset + int(strLen) > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Name = string(data[*offset:*offset+int(strLen)])
	*offset += int(strLen)

	return nil
}

// decodeArraysOfStructs is the helper function that decodes ArraysOfStructs fields.
func decodeArraysOfStructs(dest *ArraysOfStructs, data []byte, offset *int, ctx *DecodeContext) error {
	var (
		strLen uint32  // For string length prefix
		arrCount uint32  // For array count
		err error  // For error handling
	)
	_ = strLen  // Avoid unused variable error
	_ = arrCount  // Avoid unused variable error
	_ = err  // Avoid unused variable error

	// Field: Items ([]Item)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	arrCount = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	err = ctx.checkArraySize(arrCount)
	if err != nil {
		return err
	}

	dest.Items = make([]Item, arrCount)
	for i := uint32(0); i < arrCount; i++ {
		err = decodeItem(&dest.Items[i], data, offset, ctx)
		if err != nil {
			return err
		}
	}

	// Field: Count (u32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Count = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	return nil
}


// DecodeArraysOfPrimitivesMessage decodes a ArraysOfPrimitives from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodeArraysOfPrimitivesMessage(data []byte) (*ArraysOfPrimitives, error) {
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
	var result ArraysOfPrimitives
	if err := DecodeArraysOfPrimitives(&result, payload); err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodeItemMessage decodes a Item from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodeItemMessage(data []byte) (*Item, error) {
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
	if typeID != 2 {
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
	var result Item
	if err := DecodeItem(&result, payload); err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodeArraysOfStructsMessage decodes a ArraysOfStructs from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodeArraysOfStructsMessage(data []byte) (*ArraysOfStructs, error) {
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
	if typeID != 3 {
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
	var result ArraysOfStructs
	if err := DecodeArraysOfStructs(&result, payload); err != nil {
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
		return DecodeArraysOfPrimitivesMessage(data)
	case 2:
		return DecodeItemMessage(data)
	case 3:
		return DecodeArraysOfStructsMessage(data)
	default:
		return nil, ErrUnknownMessageType
	}
}


// DecodeArraysOfPrimitivesFromReader decodes a ArraysOfPrimitives from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodeArraysOfPrimitivesFromReader(dest *ArraysOfPrimitives, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodeArraysOfPrimitives(dest, buf)
}

// DecodeItemFromReader decodes a Item from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodeItemFromReader(dest *Item, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodeItem(dest, buf)
}

// DecodeArraysOfStructsFromReader decodes a ArraysOfStructs from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodeArraysOfStructsFromReader(dest *ArraysOfStructs, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodeArraysOfStructs(dest, buf)
}
