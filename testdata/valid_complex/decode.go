package valid_complex

import (
	"encoding/binary"
	"math"
	"io"
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


// DecodeDeviceList decodes a DeviceList from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodeDeviceList(dest *DeviceList, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodeDeviceList(dest, data, &offset, ctx)
}

// DecodeDevice decodes a Device from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodeDevice(dest *Device, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodeDevice(dest, data, &offset, ctx)
}

// DecodeParameter decodes a Parameter from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodeParameter(dest *Parameter, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodeParameter(dest, data, &offset, ctx)
}


// decodeDeviceList is the helper function that decodes DeviceList fields.
func decodeDeviceList(dest *DeviceList, data []byte, offset *int, ctx *DecodeContext) error {
	var (
		strLen uint32  // For string length prefix
		arrCount uint32  // For array count
		err error  // For error handling
	)
	_ = strLen  // Avoid unused variable error
	_ = arrCount  // Avoid unused variable error
	_ = err  // Avoid unused variable error

	// Field: Devices ([]Device)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	arrCount = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	err = ctx.checkArraySize(arrCount)
	if err != nil {
		return err
	}

	dest.Devices = make([]Device, arrCount)
	for i := uint32(0); i < arrCount; i++ {
		err = decodeDevice(&dest.Devices[i], data, offset, ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// decodeDevice is the helper function that decodes Device fields.
func decodeDevice(dest *Device, data []byte, offset *int, ctx *DecodeContext) error {
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

	// Field: Parameters ([]Parameter)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	arrCount = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	err = ctx.checkArraySize(arrCount)
	if err != nil {
		return err
	}

	dest.Parameters = make([]Parameter, arrCount)
	for i := uint32(0); i < arrCount; i++ {
		err = decodeParameter(&dest.Parameters[i], data, offset, ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

// decodeParameter is the helper function that decodes Parameter fields.
func decodeParameter(dest *Parameter, data []byte, offset *int, ctx *DecodeContext) error {
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

	// Field: Value (f64)
	if *offset + 8 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Value = math.Float64frombits(binary.LittleEndian.Uint64(data[*offset:]))
	*offset += 8

	return nil
}


// DecodeDeviceListMessage decodes a DeviceList from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodeDeviceListMessage(data []byte) (*DeviceList, error) {
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
	var result DeviceList
	if err := DecodeDeviceList(&result, payload); err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodeDeviceMessage decodes a Device from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodeDeviceMessage(data []byte) (*Device, error) {
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
	var result Device
	if err := DecodeDevice(&result, payload); err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodeParameterMessage decodes a Parameter from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodeParameterMessage(data []byte) (*Parameter, error) {
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
	var result Parameter
	if err := DecodeParameter(&result, payload); err != nil {
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
		return DecodeDeviceListMessage(data)
	case 2:
		return DecodeDeviceMessage(data)
	case 3:
		return DecodeParameterMessage(data)
	default:
		return nil, ErrUnknownMessageType
	}
}


// DecodeDeviceListFromReader decodes a DeviceList from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodeDeviceListFromReader(dest *DeviceList, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodeDeviceList(dest, buf)
}

// DecodeDeviceFromReader decodes a Device from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodeDeviceFromReader(dest *Device, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodeDevice(dest, buf)
}

// DecodeParameterFromReader decodes a Parameter from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodeParameterFromReader(dest *Parameter, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodeParameter(dest, buf)
}
