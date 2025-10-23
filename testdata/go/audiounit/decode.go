package audiounit

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

// DecodePluginRegistry decodes a PluginRegistry from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodePluginRegistry(dest *PluginRegistry, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodePluginRegistry(dest, data, &offset, ctx)
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

	// Field: Address (u64)
	if *offset + 8 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Address = binary.LittleEndian.Uint64(data[*offset:])
	*offset += 8

	// Field: DisplayName (str)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	strLen = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	if *offset + int(strLen) > len(data) {
		return ErrUnexpectedEOF
	}
	dest.DisplayName = string(data[*offset:*offset+int(strLen)])
	*offset += int(strLen)

	// Field: Identifier (str)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	strLen = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	if *offset + int(strLen) > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Identifier = string(data[*offset:*offset+int(strLen)])
	*offset += int(strLen)

	// Field: Unit (str)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	strLen = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	if *offset + int(strLen) > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Unit = string(data[*offset:*offset+int(strLen)])
	*offset += int(strLen)

	// Field: MinValue (f32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.MinValue = math.Float32frombits(binary.LittleEndian.Uint32(data[*offset:]))
	*offset += 4

	// Field: MaxValue (f32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.MaxValue = math.Float32frombits(binary.LittleEndian.Uint32(data[*offset:]))
	*offset += 4

	// Field: DefaultValue (f32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.DefaultValue = math.Float32frombits(binary.LittleEndian.Uint32(data[*offset:]))
	*offset += 4

	// Field: CurrentValue (f32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.CurrentValue = math.Float32frombits(binary.LittleEndian.Uint32(data[*offset:]))
	*offset += 4

	// Field: RawFlags (u32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.RawFlags = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	// Field: IsWritable (bool)
	if *offset + 1 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.IsWritable = data[*offset] != 0
	*offset += 1

	// Field: CanRamp (bool)
	if *offset + 1 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.CanRamp = data[*offset] != 0
	*offset += 1

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

	// Field: ManufacturerId (str)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	strLen = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	if *offset + int(strLen) > len(data) {
		return ErrUnexpectedEOF
	}
	dest.ManufacturerId = string(data[*offset:*offset+int(strLen)])
	*offset += int(strLen)

	// Field: ComponentType (str)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	strLen = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	if *offset + int(strLen) > len(data) {
		return ErrUnexpectedEOF
	}
	dest.ComponentType = string(data[*offset:*offset+int(strLen)])
	*offset += int(strLen)

	// Field: ComponentSubtype (str)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	strLen = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	if *offset + int(strLen) > len(data) {
		return ErrUnexpectedEOF
	}
	dest.ComponentSubtype = string(data[*offset:*offset+int(strLen)])
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

// decodePluginRegistry is the helper function that decodes PluginRegistry fields.
func decodePluginRegistry(dest *PluginRegistry, data []byte, offset *int, ctx *DecodeContext) error {
	var (
		strLen uint32  // For string length prefix
		arrCount uint32  // For array count
		err error  // For error handling
	)
	_ = strLen  // Avoid unused variable error
	_ = arrCount  // Avoid unused variable error
	_ = err  // Avoid unused variable error

	// Field: Plugins ([]Plugin)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	arrCount = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	err = ctx.checkArraySize(arrCount)
	if err != nil {
		return err
	}

	dest.Plugins = make([]Plugin, arrCount)
	for i := uint32(0); i < arrCount; i++ {
		err = decodePlugin(&dest.Plugins[i], data, offset, ctx)
		if err != nil {
			return err
		}
	}

	// Field: TotalPluginCount (u32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.TotalPluginCount = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	// Field: TotalParameterCount (u32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.TotalParameterCount = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

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

// DecodePluginRegistryMessage decodes a PluginRegistry from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodePluginRegistryMessage(data []byte) (*PluginRegistry, error) {
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
	var result PluginRegistry
	if err := DecodePluginRegistry(&result, payload); err != nil {
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
		return DecodePluginRegistryMessage(data)
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

// DecodePluginRegistryFromReader decodes a PluginRegistry from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodePluginRegistryFromReader(dest *PluginRegistry, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodePluginRegistry(dest, buf)
}
