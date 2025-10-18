package optional

import (
	"encoding/binary"
	"io"
)

// calculateRequestSize calculates the wire format size for Request.
func calculateRequestSize(src *Request) int {
	size := 0
	// Field: Id
	size += 4
	// Field: Metadata
	size += 1 // presence flag
	if src.Metadata != nil {
		size += calculateMetadataSize(src.Metadata)
	}
	return size
}

// EncodeRequest encodes a Request to wire format.
// It returns the encoded bytes or an error.
func EncodeRequest(src *Request) ([]byte, error) {
	size := calculateRequestSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeRequest(src, buf, &offset); err != nil {
		return nil, err
	}
	return buf, nil
}

// calculateMetadataSize calculates the wire format size for Metadata.
func calculateMetadataSize(src *Metadata) int {
	size := 0
	// Field: UserId
	size += 8
	// Field: Username
	size += 4 + len(src.Username)
	return size
}

// EncodeMetadata encodes a Metadata to wire format.
// It returns the encoded bytes or an error.
func EncodeMetadata(src *Metadata) ([]byte, error) {
	size := calculateMetadataSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeMetadata(src, buf, &offset); err != nil {
		return nil, err
	}
	return buf, nil
}

// calculateConfigSize calculates the wire format size for Config.
func calculateConfigSize(src *Config) int {
	size := 0
	// Field: Name
	size += 4 + len(src.Name)
	// Field: Database
	size += 1 // presence flag
	if src.Database != nil {
		size += calculateDatabaseConfigSize(src.Database)
	}
	// Field: Cache
	size += 1 // presence flag
	if src.Cache != nil {
		size += calculateCacheConfigSize(src.Cache)
	}
	return size
}

// EncodeConfig encodes a Config to wire format.
// It returns the encoded bytes or an error.
func EncodeConfig(src *Config) ([]byte, error) {
	size := calculateConfigSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeConfig(src, buf, &offset); err != nil {
		return nil, err
	}
	return buf, nil
}

// calculateDatabaseConfigSize calculates the wire format size for DatabaseConfig.
func calculateDatabaseConfigSize(src *DatabaseConfig) int {
	size := 0
	// Field: Host
	size += 4 + len(src.Host)
	// Field: Port
	size += 2
	return size
}

// EncodeDatabaseConfig encodes a DatabaseConfig to wire format.
// It returns the encoded bytes or an error.
func EncodeDatabaseConfig(src *DatabaseConfig) ([]byte, error) {
	size := calculateDatabaseConfigSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeDatabaseConfig(src, buf, &offset); err != nil {
		return nil, err
	}
	return buf, nil
}

// calculateCacheConfigSize calculates the wire format size for CacheConfig.
func calculateCacheConfigSize(src *CacheConfig) int {
	size := 0
	// Field: SizeMb
	size += 4
	// Field: TtlSeconds
	size += 4
	return size
}

// EncodeCacheConfig encodes a CacheConfig to wire format.
// It returns the encoded bytes or an error.
func EncodeCacheConfig(src *CacheConfig) ([]byte, error) {
	size := calculateCacheConfigSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeCacheConfig(src, buf, &offset); err != nil {
		return nil, err
	}
	return buf, nil
}

// calculateDocumentSize calculates the wire format size for Document.
func calculateDocumentSize(src *Document) int {
	size := 0
	// Field: Id
	size += 4
	// Field: Tags
	size += 1 // presence flag
	if src.Tags != nil {
		size += calculateTagListSize(src.Tags)
	}
	return size
}

// EncodeDocument encodes a Document to wire format.
// It returns the encoded bytes or an error.
func EncodeDocument(src *Document) ([]byte, error) {
	size := calculateDocumentSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeDocument(src, buf, &offset); err != nil {
		return nil, err
	}
	return buf, nil
}

// calculateTagListSize calculates the wire format size for TagList.
func calculateTagListSize(src *TagList) int {
	size := 0
	// Field: Items
	size += 4
	for i := range src.Items {
		size += 4 + len(src.Items[i])
	}
	return size
}

// EncodeTagList encodes a TagList to wire format.
// It returns the encoded bytes or an error.
func EncodeTagList(src *TagList) ([]byte, error) {
	size := calculateTagListSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeTagList(src, buf, &offset); err != nil {
		return nil, err
	}
	return buf, nil
}


// encodeRequest is the helper function that encodes Request fields.
func encodeRequest(src *Request, buf []byte, offset *int) error {
	// Field: Id (u32)
	binary.LittleEndian.PutUint32(buf[*offset:], src.Id)
	*offset += 4

	// Field: Metadata (Metadata)
	if src.Metadata == nil {
		buf[*offset] = 0 // presence = 0 (absent)
		*offset += 1
	} else {
		buf[*offset] = 1 // presence = 1 (present)
		*offset += 1
		if err := encodeMetadata(src.Metadata, buf, offset); err != nil {
			return err
		}
	}
	return nil
}

// encodeMetadata is the helper function that encodes Metadata fields.
func encodeMetadata(src *Metadata, buf []byte, offset *int) error {
	// Field: UserId (u64)
	binary.LittleEndian.PutUint64(buf[*offset:], src.UserId)
	*offset += 8

	// Field: Username (str)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Username)))
	*offset += 4
	copy(buf[*offset:], src.Username)
	*offset += len(src.Username)
	return nil
}

// encodeConfig is the helper function that encodes Config fields.
func encodeConfig(src *Config, buf []byte, offset *int) error {
	// Field: Name (str)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Name)))
	*offset += 4
	copy(buf[*offset:], src.Name)
	*offset += len(src.Name)
	// Field: Database (DatabaseConfig)
	if src.Database == nil {
		buf[*offset] = 0 // presence = 0 (absent)
		*offset += 1
	} else {
		buf[*offset] = 1 // presence = 1 (present)
		*offset += 1
		if err := encodeDatabaseConfig(src.Database, buf, offset); err != nil {
			return err
		}
	}
	// Field: Cache (CacheConfig)
	if src.Cache == nil {
		buf[*offset] = 0 // presence = 0 (absent)
		*offset += 1
	} else {
		buf[*offset] = 1 // presence = 1 (present)
		*offset += 1
		if err := encodeCacheConfig(src.Cache, buf, offset); err != nil {
			return err
		}
	}
	return nil
}

// encodeDatabaseConfig is the helper function that encodes DatabaseConfig fields.
func encodeDatabaseConfig(src *DatabaseConfig, buf []byte, offset *int) error {
	// Field: Host (str)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Host)))
	*offset += 4
	copy(buf[*offset:], src.Host)
	*offset += len(src.Host)
	// Field: Port (u16)
	binary.LittleEndian.PutUint16(buf[*offset:], src.Port)
	*offset += 2

	return nil
}

// encodeCacheConfig is the helper function that encodes CacheConfig fields.
func encodeCacheConfig(src *CacheConfig, buf []byte, offset *int) error {
	// Field: SizeMb (u32)
	binary.LittleEndian.PutUint32(buf[*offset:], src.SizeMb)
	*offset += 4

	// Field: TtlSeconds (u32)
	binary.LittleEndian.PutUint32(buf[*offset:], src.TtlSeconds)
	*offset += 4

	return nil
}

// encodeDocument is the helper function that encodes Document fields.
func encodeDocument(src *Document, buf []byte, offset *int) error {
	// Field: Id (u32)
	binary.LittleEndian.PutUint32(buf[*offset:], src.Id)
	*offset += 4

	// Field: Tags (TagList)
	if src.Tags == nil {
		buf[*offset] = 0 // presence = 0 (absent)
		*offset += 1
	} else {
		buf[*offset] = 1 // presence = 1 (present)
		*offset += 1
		if err := encodeTagList(src.Tags, buf, offset); err != nil {
			return err
		}
	}
	return nil
}

// encodeTagList is the helper function that encodes TagList fields.
func encodeTagList(src *TagList, buf []byte, offset *int) error {
	// Field: Items ([]str)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Items)))
	*offset += 4

	for i := range src.Items {
		binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Items[i])))
		*offset += 4
		copy(buf[*offset:], src.Items[i])
		*offset += len(src.Items[i])
	}
	return nil
}


// EncodeRequestMessage encodes a Request to self-describing message format.
// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// This format is suitable for persistence, network transmission, and cross-service communication.
func EncodeRequestMessage(src *Request) ([]byte, error) {
	// Encode payload
	payload, err := EncodeRequest(src)
	if err != nil {
		return nil, err
	}

	// Allocate message buffer (header + payload)
	messageSize := MessageHeaderSize + len(payload)
	message := make([]byte, messageSize)

	// Write header
	copy(message[0:3], MessageMagic)  // Magic bytes 'SDP'
	message[3] = MessageVersion       // Protocol version '2'
	binary.LittleEndian.PutUint16(message[4:6], 1)  // Type ID
	binary.LittleEndian.PutUint32(message[6:10], uint32(len(payload)))  // Payload length

	// Copy payload
	copy(message[10:], payload)

	return message, nil
}

// EncodeMetadataMessage encodes a Metadata to self-describing message format.
// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// This format is suitable for persistence, network transmission, and cross-service communication.
func EncodeMetadataMessage(src *Metadata) ([]byte, error) {
	// Encode payload
	payload, err := EncodeMetadata(src)
	if err != nil {
		return nil, err
	}

	// Allocate message buffer (header + payload)
	messageSize := MessageHeaderSize + len(payload)
	message := make([]byte, messageSize)

	// Write header
	copy(message[0:3], MessageMagic)  // Magic bytes 'SDP'
	message[3] = MessageVersion       // Protocol version '2'
	binary.LittleEndian.PutUint16(message[4:6], 2)  // Type ID
	binary.LittleEndian.PutUint32(message[6:10], uint32(len(payload)))  // Payload length

	// Copy payload
	copy(message[10:], payload)

	return message, nil
}

// EncodeConfigMessage encodes a Config to self-describing message format.
// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// This format is suitable for persistence, network transmission, and cross-service communication.
func EncodeConfigMessage(src *Config) ([]byte, error) {
	// Encode payload
	payload, err := EncodeConfig(src)
	if err != nil {
		return nil, err
	}

	// Allocate message buffer (header + payload)
	messageSize := MessageHeaderSize + len(payload)
	message := make([]byte, messageSize)

	// Write header
	copy(message[0:3], MessageMagic)  // Magic bytes 'SDP'
	message[3] = MessageVersion       // Protocol version '2'
	binary.LittleEndian.PutUint16(message[4:6], 3)  // Type ID
	binary.LittleEndian.PutUint32(message[6:10], uint32(len(payload)))  // Payload length

	// Copy payload
	copy(message[10:], payload)

	return message, nil
}

// EncodeDatabaseConfigMessage encodes a DatabaseConfig to self-describing message format.
// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// This format is suitable for persistence, network transmission, and cross-service communication.
func EncodeDatabaseConfigMessage(src *DatabaseConfig) ([]byte, error) {
	// Encode payload
	payload, err := EncodeDatabaseConfig(src)
	if err != nil {
		return nil, err
	}

	// Allocate message buffer (header + payload)
	messageSize := MessageHeaderSize + len(payload)
	message := make([]byte, messageSize)

	// Write header
	copy(message[0:3], MessageMagic)  // Magic bytes 'SDP'
	message[3] = MessageVersion       // Protocol version '2'
	binary.LittleEndian.PutUint16(message[4:6], 4)  // Type ID
	binary.LittleEndian.PutUint32(message[6:10], uint32(len(payload)))  // Payload length

	// Copy payload
	copy(message[10:], payload)

	return message, nil
}

// EncodeCacheConfigMessage encodes a CacheConfig to self-describing message format.
// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// This format is suitable for persistence, network transmission, and cross-service communication.
func EncodeCacheConfigMessage(src *CacheConfig) ([]byte, error) {
	// Encode payload
	payload, err := EncodeCacheConfig(src)
	if err != nil {
		return nil, err
	}

	// Allocate message buffer (header + payload)
	messageSize := MessageHeaderSize + len(payload)
	message := make([]byte, messageSize)

	// Write header
	copy(message[0:3], MessageMagic)  // Magic bytes 'SDP'
	message[3] = MessageVersion       // Protocol version '2'
	binary.LittleEndian.PutUint16(message[4:6], 5)  // Type ID
	binary.LittleEndian.PutUint32(message[6:10], uint32(len(payload)))  // Payload length

	// Copy payload
	copy(message[10:], payload)

	return message, nil
}

// EncodeDocumentMessage encodes a Document to self-describing message format.
// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// This format is suitable for persistence, network transmission, and cross-service communication.
func EncodeDocumentMessage(src *Document) ([]byte, error) {
	// Encode payload
	payload, err := EncodeDocument(src)
	if err != nil {
		return nil, err
	}

	// Allocate message buffer (header + payload)
	messageSize := MessageHeaderSize + len(payload)
	message := make([]byte, messageSize)

	// Write header
	copy(message[0:3], MessageMagic)  // Magic bytes 'SDP'
	message[3] = MessageVersion       // Protocol version '2'
	binary.LittleEndian.PutUint16(message[4:6], 6)  // Type ID
	binary.LittleEndian.PutUint32(message[6:10], uint32(len(payload)))  // Payload length

	// Copy payload
	copy(message[10:], payload)

	return message, nil
}

// EncodeTagListMessage encodes a TagList to self-describing message format.
// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// This format is suitable for persistence, network transmission, and cross-service communication.
func EncodeTagListMessage(src *TagList) ([]byte, error) {
	// Encode payload
	payload, err := EncodeTagList(src)
	if err != nil {
		return nil, err
	}

	// Allocate message buffer (header + payload)
	messageSize := MessageHeaderSize + len(payload)
	message := make([]byte, messageSize)

	// Write header
	copy(message[0:3], MessageMagic)  // Magic bytes 'SDP'
	message[3] = MessageVersion       // Protocol version '2'
	binary.LittleEndian.PutUint16(message[4:6], 7)  // Type ID
	binary.LittleEndian.PutUint32(message[6:10], uint32(len(payload)))  // Payload length

	// Copy payload
	copy(message[10:], payload)

	return message, nil
}



// EncodeRequestToWriter encodes a Request to wire format and writes it to the provided io.Writer.
// This enables streaming I/O without baked-in compression.
//
// Users can compose with any io.Writer implementation:
//   - File I/O: os.File
//   - Compression: gzip.Writer, zstd.Writer, etc.
//   - Network: net.Conn, http.ResponseWriter
//   - Encryption: custom crypto.Writer
//   - Metrics: custom byte counting wrappers
func EncodeRequestToWriter(src *Request, w io.Writer) error {
	size := calculateRequestSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeRequest(src, buf, &offset); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}

// EncodeMetadataToWriter encodes a Metadata to wire format and writes it to the provided io.Writer.
// This enables streaming I/O without baked-in compression.
//
// Users can compose with any io.Writer implementation:
//   - File I/O: os.File
//   - Compression: gzip.Writer, zstd.Writer, etc.
//   - Network: net.Conn, http.ResponseWriter
//   - Encryption: custom crypto.Writer
//   - Metrics: custom byte counting wrappers
func EncodeMetadataToWriter(src *Metadata, w io.Writer) error {
	size := calculateMetadataSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeMetadata(src, buf, &offset); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}

// EncodeConfigToWriter encodes a Config to wire format and writes it to the provided io.Writer.
// This enables streaming I/O without baked-in compression.
//
// Users can compose with any io.Writer implementation:
//   - File I/O: os.File
//   - Compression: gzip.Writer, zstd.Writer, etc.
//   - Network: net.Conn, http.ResponseWriter
//   - Encryption: custom crypto.Writer
//   - Metrics: custom byte counting wrappers
func EncodeConfigToWriter(src *Config, w io.Writer) error {
	size := calculateConfigSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeConfig(src, buf, &offset); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}

// EncodeDatabaseConfigToWriter encodes a DatabaseConfig to wire format and writes it to the provided io.Writer.
// This enables streaming I/O without baked-in compression.
//
// Users can compose with any io.Writer implementation:
//   - File I/O: os.File
//   - Compression: gzip.Writer, zstd.Writer, etc.
//   - Network: net.Conn, http.ResponseWriter
//   - Encryption: custom crypto.Writer
//   - Metrics: custom byte counting wrappers
func EncodeDatabaseConfigToWriter(src *DatabaseConfig, w io.Writer) error {
	size := calculateDatabaseConfigSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeDatabaseConfig(src, buf, &offset); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}

// EncodeCacheConfigToWriter encodes a CacheConfig to wire format and writes it to the provided io.Writer.
// This enables streaming I/O without baked-in compression.
//
// Users can compose with any io.Writer implementation:
//   - File I/O: os.File
//   - Compression: gzip.Writer, zstd.Writer, etc.
//   - Network: net.Conn, http.ResponseWriter
//   - Encryption: custom crypto.Writer
//   - Metrics: custom byte counting wrappers
func EncodeCacheConfigToWriter(src *CacheConfig, w io.Writer) error {
	size := calculateCacheConfigSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeCacheConfig(src, buf, &offset); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}

// EncodeDocumentToWriter encodes a Document to wire format and writes it to the provided io.Writer.
// This enables streaming I/O without baked-in compression.
//
// Users can compose with any io.Writer implementation:
//   - File I/O: os.File
//   - Compression: gzip.Writer, zstd.Writer, etc.
//   - Network: net.Conn, http.ResponseWriter
//   - Encryption: custom crypto.Writer
//   - Metrics: custom byte counting wrappers
func EncodeDocumentToWriter(src *Document, w io.Writer) error {
	size := calculateDocumentSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeDocument(src, buf, &offset); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}

// EncodeTagListToWriter encodes a TagList to wire format and writes it to the provided io.Writer.
// This enables streaming I/O without baked-in compression.
//
// Users can compose with any io.Writer implementation:
//   - File I/O: os.File
//   - Compression: gzip.Writer, zstd.Writer, etc.
//   - Network: net.Conn, http.ResponseWriter
//   - Encryption: custom crypto.Writer
//   - Metrics: custom byte counting wrappers
func EncodeTagListToWriter(src *TagList, w io.Writer) error {
	size := calculateTagListSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeTagList(src, buf, &offset); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}
