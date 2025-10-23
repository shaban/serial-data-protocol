package complex

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

// DecodePlugin decodes a Plugin from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodePlugin(dest *Plugin, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodePlugin(dest, data, &offset, ctx)
}

// DecodeAudioDevice decodes a AudioDevice from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodeAudioDevice(dest *AudioDevice, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodeAudioDevice(dest, data, &offset, ctx)
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

	// Field: Value (f32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Value = math.Float32frombits(binary.LittleEndian.Uint32(data[*offset:]))
	*offset += 4

	// Field: Min (f32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Min = math.Float32frombits(binary.LittleEndian.Uint32(data[*offset:]))
	*offset += 4

	// Field: Max (f32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Max = math.Float32frombits(binary.LittleEndian.Uint32(data[*offset:]))
	*offset += 4

	return nil
}

// decodePlugin is the helper function that decodes Plugin fields.
func decodePlugin(dest *Plugin, data []byte, offset *int, ctx *DecodeContext) error {
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

	// Field: Manufacturer (str)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	strLen = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	if *offset + int(strLen) > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Manufacturer = string(data[*offset:*offset+int(strLen)])
	*offset += int(strLen)

	// Field: Version (u32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Version = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	// Field: Enabled (bool)
	if *offset + 1 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Enabled = data[*offset] != 0
	*offset += 1

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

// decodeAudioDevice is the helper function that decodes AudioDevice fields.
func decodeAudioDevice(dest *AudioDevice, data []byte, offset *int, ctx *DecodeContext) error {
	var (
		strLen uint32  // For string length prefix
		arrCount uint32  // For array count
		err error  // For error handling
	)
	_ = strLen  // Avoid unused variable error
	_ = arrCount  // Avoid unused variable error
	_ = err  // Avoid unused variable error

	// Field: DeviceId (u32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.DeviceId = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	// Field: DeviceName (str)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	strLen = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	if *offset + int(strLen) > len(data) {
		return ErrUnexpectedEOF
	}
	dest.DeviceName = string(data[*offset:*offset+int(strLen)])
	*offset += int(strLen)

	// Field: SampleRate (u32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.SampleRate = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	// Field: BufferSize (u32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.BufferSize = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	// Field: InputChannels (u16)
	if *offset + 2 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.InputChannels = binary.LittleEndian.Uint16(data[*offset:])
	*offset += 2

	// Field: OutputChannels (u16)
	if *offset + 2 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.OutputChannels = binary.LittleEndian.Uint16(data[*offset:])
	*offset += 2

	// Field: IsDefault (bool)
	if *offset + 1 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.IsDefault = data[*offset] != 0
	*offset += 1

	// Field: ActivePlugins ([]Plugin)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	arrCount = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	err = ctx.checkArraySize(arrCount)
	if err != nil {
		return err
	}

	dest.ActivePlugins = make([]Plugin, arrCount)
	for i := uint32(0); i < arrCount; i++ {
		err = decodePlugin(&dest.ActivePlugins[i], data, offset, ctx)
		if err != nil {
			return err
		}
	}

	return nil
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
	var result Parameter
	if err := DecodeParameter(&result, payload); err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodePluginMessage decodes a Plugin from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodePluginMessage(data []byte) (*Plugin, error) {
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
	var result Plugin
	if err := DecodePlugin(&result, payload); err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodeAudioDeviceMessage decodes a AudioDevice from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodeAudioDeviceMessage(data []byte) (*AudioDevice, error) {
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
	var result AudioDevice
	if err := DecodeAudioDevice(&result, payload); err != nil {
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
		return DecodeParameterMessage(data)
	case 2:
		return DecodePluginMessage(data)
	case 3:
		return DecodeAudioDeviceMessage(data)
	default:
		return nil, ErrUnknownMessageType
	}
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

// DecodePluginFromReader decodes a Plugin from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodePluginFromReader(dest *Plugin, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodePlugin(dest, buf)
}

// DecodeAudioDeviceFromReader decodes a AudioDevice from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodeAudioDeviceFromReader(dest *AudioDevice, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodeAudioDevice(dest, buf)
}
