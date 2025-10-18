package optional

import (
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


// DecodeRequest decodes a Request from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodeRequest(dest *Request, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodeRequest(dest, data, &offset, ctx)
}

// DecodeMetadata decodes a Metadata from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodeMetadata(dest *Metadata, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodeMetadata(dest, data, &offset, ctx)
}

// DecodeConfig decodes a Config from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodeConfig(dest *Config, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodeConfig(dest, data, &offset, ctx)
}

// DecodeDatabaseConfig decodes a DatabaseConfig from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodeDatabaseConfig(dest *DatabaseConfig, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodeDatabaseConfig(dest, data, &offset, ctx)
}

// DecodeCacheConfig decodes a CacheConfig from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodeCacheConfig(dest *CacheConfig, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodeCacheConfig(dest, data, &offset, ctx)
}

// DecodeDocument decodes a Document from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodeDocument(dest *Document, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodeDocument(dest, data, &offset, ctx)
}

// DecodeTagList decodes a TagList from wire format.
// It validates the data size and delegates to the decoder implementation.
func DecodeTagList(dest *TagList, data []byte) error {
	if len(data) > MaxSerializedSize {
		return ErrDataTooLarge
	}
	ctx := &DecodeContext{}
	offset := 0
	return decodeTagList(dest, data, &offset, ctx)
}


// decodeRequest is the helper function that decodes Request fields.
func decodeRequest(dest *Request, data []byte, offset *int, ctx *DecodeContext) error {
	var (
		strLen uint32  // For string length prefix
		arrCount uint32  // For array count
		presence byte  // For optional field presence flags
		err error  // For error handling
	)
	_ = strLen  // Avoid unused variable error
	_ = arrCount  // Avoid unused variable error
	_ = presence  // Avoid unused variable error
	_ = err  // Avoid unused variable error

	// Field: Id (u32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Id = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	// Field: Metadata (optional)
	if *offset >= len(data) {
		return ErrUnexpectedEOF
	}
	presence = data[*offset]
	*offset += 1

	if presence == 0 {
		dest.Metadata = nil
	} else if presence == 1 {
			dest.Metadata = &Metadata{}
			err = decodeMetadata(dest.Metadata, data, offset, ctx)
			if err != nil {
				return err
			}
	} else {
		return ErrInvalidData
	}

	return nil
}

// decodeMetadata is the helper function that decodes Metadata fields.
func decodeMetadata(dest *Metadata, data []byte, offset *int, ctx *DecodeContext) error {
	var (
		strLen uint32  // For string length prefix
		arrCount uint32  // For array count
		err error  // For error handling
	)
	_ = strLen  // Avoid unused variable error
	_ = arrCount  // Avoid unused variable error
	_ = err  // Avoid unused variable error

	// Field: UserId (u64)
	if *offset + 8 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.UserId = binary.LittleEndian.Uint64(data[*offset:])
	*offset += 8

	// Field: Username (str)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	strLen = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	if *offset + int(strLen) > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Username = string(data[*offset:*offset+int(strLen)])
	*offset += int(strLen)

	return nil
}

// decodeConfig is the helper function that decodes Config fields.
func decodeConfig(dest *Config, data []byte, offset *int, ctx *DecodeContext) error {
	var (
		strLen uint32  // For string length prefix
		arrCount uint32  // For array count
		presence byte  // For optional field presence flags
		err error  // For error handling
	)
	_ = strLen  // Avoid unused variable error
	_ = arrCount  // Avoid unused variable error
	_ = presence  // Avoid unused variable error
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

	// Field: Database (optional)
	if *offset >= len(data) {
		return ErrUnexpectedEOF
	}
	presence = data[*offset]
	*offset += 1

	if presence == 0 {
		dest.Database = nil
	} else if presence == 1 {
			dest.Database = &DatabaseConfig{}
			err = decodeDatabaseConfig(dest.Database, data, offset, ctx)
			if err != nil {
				return err
			}
	} else {
		return ErrInvalidData
	}

	// Field: Cache (optional)
	if *offset >= len(data) {
		return ErrUnexpectedEOF
	}
	presence = data[*offset]
	*offset += 1

	if presence == 0 {
		dest.Cache = nil
	} else if presence == 1 {
			dest.Cache = &CacheConfig{}
			err = decodeCacheConfig(dest.Cache, data, offset, ctx)
			if err != nil {
				return err
			}
	} else {
		return ErrInvalidData
	}

	return nil
}

// decodeDatabaseConfig is the helper function that decodes DatabaseConfig fields.
func decodeDatabaseConfig(dest *DatabaseConfig, data []byte, offset *int, ctx *DecodeContext) error {
	var (
		strLen uint32  // For string length prefix
		arrCount uint32  // For array count
		err error  // For error handling
	)
	_ = strLen  // Avoid unused variable error
	_ = arrCount  // Avoid unused variable error
	_ = err  // Avoid unused variable error

	// Field: Host (str)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	strLen = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	if *offset + int(strLen) > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Host = string(data[*offset:*offset+int(strLen)])
	*offset += int(strLen)

	// Field: Port (u16)
	if *offset + 2 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Port = binary.LittleEndian.Uint16(data[*offset:])
	*offset += 2

	return nil
}

// decodeCacheConfig is the helper function that decodes CacheConfig fields.
func decodeCacheConfig(dest *CacheConfig, data []byte, offset *int, ctx *DecodeContext) error {
	var (
		strLen uint32  // For string length prefix
		arrCount uint32  // For array count
		err error  // For error handling
	)
	_ = strLen  // Avoid unused variable error
	_ = arrCount  // Avoid unused variable error
	_ = err  // Avoid unused variable error

	// Field: SizeMb (u32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.SizeMb = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	// Field: TtlSeconds (u32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.TtlSeconds = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	return nil
}

// decodeDocument is the helper function that decodes Document fields.
func decodeDocument(dest *Document, data []byte, offset *int, ctx *DecodeContext) error {
	var (
		strLen uint32  // For string length prefix
		arrCount uint32  // For array count
		presence byte  // For optional field presence flags
		err error  // For error handling
	)
	_ = strLen  // Avoid unused variable error
	_ = arrCount  // Avoid unused variable error
	_ = presence  // Avoid unused variable error
	_ = err  // Avoid unused variable error

	// Field: Id (u32)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	dest.Id = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	// Field: Tags (optional)
	if *offset >= len(data) {
		return ErrUnexpectedEOF
	}
	presence = data[*offset]
	*offset += 1

	if presence == 0 {
		dest.Tags = nil
	} else if presence == 1 {
			dest.Tags = &TagList{}
			err = decodeTagList(dest.Tags, data, offset, ctx)
			if err != nil {
				return err
			}
	} else {
		return ErrInvalidData
	}

	return nil
}

// decodeTagList is the helper function that decodes TagList fields.
func decodeTagList(dest *TagList, data []byte, offset *int, ctx *DecodeContext) error {
	var (
		strLen uint32  // For string length prefix
		arrCount uint32  // For array count
		err error  // For error handling
	)
	_ = strLen  // Avoid unused variable error
	_ = arrCount  // Avoid unused variable error
	_ = err  // Avoid unused variable error

	// Field: Items ([]str)
	if *offset + 4 > len(data) {
		return ErrUnexpectedEOF
	}
	arrCount = binary.LittleEndian.Uint32(data[*offset:])
	*offset += 4

	err = ctx.checkArraySize(arrCount)
	if err != nil {
		return err
	}

	dest.Items = make([]string, arrCount)
	for i := uint32(0); i < arrCount; i++ {
		if *offset + 4 > len(data) {
			return ErrUnexpectedEOF
		}
		strLen := binary.LittleEndian.Uint32(data[*offset:])
		*offset += 4
		
		if *offset + int(strLen) > len(data) {
			return ErrUnexpectedEOF
		}
		dest.Items[i] = string(data[*offset:*offset+int(strLen)])
		*offset += int(strLen)
	}

	return nil
}


// DecodeRequestMessage decodes a Request from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodeRequestMessage(data []byte) (*Request, error) {
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
	var result Request
	if err := DecodeRequest(&result, payload); err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodeMetadataMessage decodes a Metadata from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodeMetadataMessage(data []byte) (*Metadata, error) {
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
	var result Metadata
	if err := DecodeMetadata(&result, payload); err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodeConfigMessage decodes a Config from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodeConfigMessage(data []byte) (*Config, error) {
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
	var result Config
	if err := DecodeConfig(&result, payload); err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodeDatabaseConfigMessage decodes a DatabaseConfig from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodeDatabaseConfigMessage(data []byte) (*DatabaseConfig, error) {
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
	if typeID != 4 {
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
	var result DatabaseConfig
	if err := DecodeDatabaseConfig(&result, payload); err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodeCacheConfigMessage decodes a CacheConfig from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodeCacheConfigMessage(data []byte) (*CacheConfig, error) {
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
	if typeID != 5 {
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
	var result CacheConfig
	if err := DecodeCacheConfig(&result, payload); err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodeDocumentMessage decodes a Document from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodeDocumentMessage(data []byte) (*Document, error) {
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
	if typeID != 6 {
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
	var result Document
	if err := DecodeDocument(&result, payload); err != nil {
		return nil, err
	}

	return &result, nil
}

// DecodeTagListMessage decodes a TagList from self-describing message format.
// The message must include a valid 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// Returns an error if the header is invalid or the payload cannot be decoded.
func DecodeTagListMessage(data []byte) (*TagList, error) {
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
	if typeID != 7 {
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
	var result TagList
	if err := DecodeTagList(&result, payload); err != nil {
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
		return DecodeRequestMessage(data)
	case 2:
		return DecodeMetadataMessage(data)
	case 3:
		return DecodeConfigMessage(data)
	case 4:
		return DecodeDatabaseConfigMessage(data)
	case 5:
		return DecodeCacheConfigMessage(data)
	case 6:
		return DecodeDocumentMessage(data)
	case 7:
		return DecodeTagListMessage(data)
	default:
		return nil, ErrUnknownMessageType
	}
}


// DecodeRequestFromReader decodes a Request from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodeRequestFromReader(dest *Request, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodeRequest(dest, buf)
}

// DecodeMetadataFromReader decodes a Metadata from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodeMetadataFromReader(dest *Metadata, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodeMetadata(dest, buf)
}

// DecodeConfigFromReader decodes a Config from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodeConfigFromReader(dest *Config, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodeConfig(dest, buf)
}

// DecodeDatabaseConfigFromReader decodes a DatabaseConfig from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodeDatabaseConfigFromReader(dest *DatabaseConfig, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodeDatabaseConfig(dest, buf)
}

// DecodeCacheConfigFromReader decodes a CacheConfig from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodeCacheConfigFromReader(dest *CacheConfig, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodeCacheConfig(dest, buf)
}

// DecodeDocumentFromReader decodes a Document from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodeDocumentFromReader(dest *Document, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodeDocument(dest, buf)
}

// DecodeTagListFromReader decodes a TagList from wire format by reading from the provided io.Reader.
// This enables streaming I/O without baked-in decompression.
//
// Users can compose with any io.Reader implementation:
//   - File I/O: os.File
//   - Decompression: gzip.Reader, zstd.Reader, etc.
//   - Network: net.Conn, http.Request.Body
//   - Decryption: custom crypto.Reader
//   - Metrics: custom byte counting wrappers
func DecodeTagListFromReader(dest *TagList, r io.Reader) error {
	buf, err := io.ReadAll(r)
	if err != nil {
		return err
	}
	return DecodeTagList(dest, buf)
}
