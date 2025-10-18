package valid_basic

import (
	"io"
	"encoding/binary"
)

// calculateDeviceSize calculates the wire format size for Device.
func calculateDeviceSize(src *Device) int {
	size := 0
	// Field: Id
	size += 4
	// Field: Name
	size += 4 + len(src.Name)
	return size
}

// EncodeDevice encodes a Device to wire format.
// It returns the encoded bytes or an error.
func EncodeDevice(src *Device) ([]byte, error) {
	size := calculateDeviceSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeDevice(src, buf, &offset); err != nil {
		return nil, err
	}
	return buf, nil
}


// encodeDevice is the helper function that encodes Device fields.
func encodeDevice(src *Device, buf []byte, offset *int) error {
	// Field: Id (u32)
	binary.LittleEndian.PutUint32(buf[*offset:], src.Id)
	*offset += 4

	// Field: Name (str)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Name)))
	*offset += 4
	copy(buf[*offset:], src.Name)
	*offset += len(src.Name)
	return nil
}


// EncodeDeviceMessage encodes a Device to self-describing message format.
// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// This format is suitable for persistence, network transmission, and cross-service communication.
func EncodeDeviceMessage(src *Device) ([]byte, error) {
	// Encode payload
	payload, err := EncodeDevice(src)
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



// EncodeDeviceToWriter encodes a Device to wire format and writes it to the provided io.Writer.
// This enables streaming I/O without baked-in compression.
//
// Users can compose with any io.Writer implementation:
//   - File I/O: os.File
//   - Compression: gzip.Writer, zstd.Writer, etc.
//   - Network: net.Conn, http.ResponseWriter
//   - Encryption: custom crypto.Writer
//   - Metrics: custom byte counting wrappers
func EncodeDeviceToWriter(src *Device, w io.Writer) error {
	size := calculateDeviceSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeDevice(src, buf, &offset); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}
