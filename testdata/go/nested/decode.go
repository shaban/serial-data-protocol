package nested

import (
	"io"
	"encoding/binary"
	"math"
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


// DecodePoint decodes a Point from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodePoint(dest *Point, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodePoint(dest, data, &offset, ctx)
}

// DecodeRectangle decodes a Rectangle from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodeRectangle(dest *Rectangle, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodeRectangle(dest, data, &offset, ctx)
}

// DecodeScene decodes a Scene from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodeScene(dest *Scene, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodeScene(dest, data, &offset, ctx)
}


// decodePoint is the helper function that decodes Point fields.
func decodePoint(dest *Point, data []byte, offset *int, ctx *DecodeContext) error {
	var (
		strLen uint32  // For string length prefix
		arrCount uint32  // For array count
		err error  // For error handling
	)
	_ = strLen  // Avoid unused variable error
	_ = arrCount  // Avoid unused variable error
	_ = err  // Avoid unused variable error

	// Field: X (f32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.X = math.Float32frombits(binary.LittleEndian.Uint32(data[*offset:]))
	*offset += 4

	// Field: Y (f32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Y = math.Float32frombits(binary.LittleEndian.Uint32(data[*offset:]))
	*offset += 4

	return nil
}

// decodeRectangle is the helper function that decodes Rectangle fields.
func decodeRectangle(dest *Rectangle, data []byte, offset *int, ctx *DecodeContext) error {
	var (
		strLen uint32  // For string length prefix
		arrCount uint32  // For array count
		err error  // For error handling
	)
	_ = strLen  // Avoid unused variable error
	_ = arrCount  // Avoid unused variable error
	_ = err  // Avoid unused variable error

	// Field: TopLeft (Point)
	err = decodePoint(&dest.TopLeft, data, offset, ctx)
	if err != nil {
		return err
	}

	// Field: BottomRight (Point)
	err = decodePoint(&dest.BottomRight, data, offset, ctx)
	if err != nil {
		return err
	}

	// Field: Color (u32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Color = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	return nil
}

// decodeScene is the helper function that decodes Scene fields.
func decodeScene(dest *Scene, data []byte, offset *int, ctx *DecodeContext) error {
	var (
		strLen uint32  // For string length prefix
		arrCount uint32  // For array count
		err error  // For error handling
	)
	_ = strLen  // Avoid unused variable error
	_ = arrCount  // Avoid unused variable error
	_ = err  // Avoid unused variable error

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

	// Field: MainRect (Rectangle)
	err = decodeRectangle(&dest.MainRect, data, offset, ctx)
	if err != nil {
		return err
	}

	// Field: Count (u32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Count = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	return nil
}


// DecodePointMessage decodes a Point from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodePointMessage(data []byte) (*Point, error) {
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
	var result Point
	if err := DecodePoint(&result, payload); err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodeRectangleMessage decodes a Rectangle from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodeRectangleMessage(data []byte) (*Rectangle, error) {
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
	var result Rectangle
	if err := DecodeRectangle(&result, payload); err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodeSceneMessage decodes a Scene from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodeSceneMessage(data []byte) (*Scene, error) {
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
	var result Scene
	if err := DecodeScene(&result, payload); err != nil {
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
		return DecodePointMessage(data)
	case 2:
		return DecodeRectangleMessage(data)
	case 3:
		return DecodeSceneMessage(data)
	default:
		return nil, ErrUnknownMessageType
	}
}


// DecodePointFromReader decodes a Point from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodePointFromReader(dest *Point, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodePoint(dest, buf)
}

// DecodeRectangleFromReader decodes a Rectangle from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodeRectangleFromReader(dest *Rectangle, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodeRectangle(dest, buf)
}

// DecodeSceneFromReader decodes a Scene from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodeSceneFromReader(dest *Scene, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodeScene(dest, buf)
}
