package valid_crlf

import (
	"io"
	"encoding/binary"
)

// calculateExampleSize calculates the wire format size for Example.
func calculateExampleSize(src *Example) int {
	size := 0
	// Field: Field
	size += 4
	return size
}

// EncodeExample encodes a Example to wire format.
// It returns the encoded bytes or an error.
func EncodeExample(src *Example) ([]byte, error) {
	size := calculateExampleSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeExample(src, buf, &offset); err != nil {
		return nil, err
	}
	return buf, nil
}


// encodeExample is the helper function that encodes Example fields.
func encodeExample(src *Example, buf []byte, offset *int) error {
	// Field: Field (u32)
	binary.LittleEndian.PutUint32(buf[*offset:], src.Field)
	*offset += 4

	return nil
}


// EncodeExampleMessage encodes a Example to self-describing message format.
// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// This format is suitable for persistence, network transmission, and cross-service communication.
func EncodeExampleMessage(src *Example) ([]byte, error) {
	// Encode payload
	payload, err := EncodeExample(src)
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



// EncodeExampleToWriter encodes a Example to wire format and writes it to the provided io.Writer.
// This enables streaming I/O without baked-in compression.
//
// Users can compose with any io.Writer implementation:
//   - File I/O: os.File
//   - Compression: gzip.Writer, zstd.Writer, etc.
//   - Network: net.Conn, http.ResponseWriter
//   - Encryption: custom crypto.Writer
//   - Metrics: custom byte counting wrappers
func EncodeExampleToWriter(src *Example, w io.Writer) error {
	size := calculateExampleSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeExample(src, buf, &offset); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}
