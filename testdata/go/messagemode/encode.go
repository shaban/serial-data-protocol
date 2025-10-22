package messagemode

import (
	"math"
	"io"
	"encoding/binary"
)

// calculatePointSize calculates the wire format size for Point.
func calculatePointSize(src *Point) int {
	size := 0
	// Field: X
	size += 8
	// Field: Y
	size += 8
	return size
}

// EncodePoint encodes a Point to wire format.
// It returns the encoded bytes or an error.
func EncodePoint(src *Point) ([]byte, error) {
	size := calculatePointSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodePoint(src, buf, &offset); err != nil {
		return nil, err
	}
	return buf, nil
}

// calculateRectangleSize calculates the wire format size for Rectangle.
func calculateRectangleSize(src *Rectangle) int {
	size := 0
	// Field: TopLeft
	size += calculatePointSize(&src.TopLeft)
	// Field: Width
	size += 8
	// Field: Height
	size += 8
	return size
}

// EncodeRectangle encodes a Rectangle to wire format.
// It returns the encoded bytes or an error.
func EncodeRectangle(src *Rectangle) ([]byte, error) {
	size := calculateRectangleSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeRectangle(src, buf, &offset); err != nil {
		return nil, err
	}
	return buf, nil
}


// encodePoint is the helper function that encodes Point fields.
func encodePoint(src *Point, buf []byte, offset *int) error {
	// Field: X (f64)
	binary.LittleEndian.PutUint64(buf[*offset:], math.Float64bits(src.X))
	*offset += 8

	// Field: Y (f64)
	binary.LittleEndian.PutUint64(buf[*offset:], math.Float64bits(src.Y))
	*offset += 8

	return nil
}

// encodeRectangle is the helper function that encodes Rectangle fields.
func encodeRectangle(src *Rectangle, buf []byte, offset *int) error {
	// Field: TopLeft (Point)
	if err := encodePoint(&src.TopLeft, buf, offset); err != nil {
		return err
	}
	// Field: Width (f64)
	binary.LittleEndian.PutUint64(buf[*offset:], math.Float64bits(src.Width))
	*offset += 8

	// Field: Height (f64)
	binary.LittleEndian.PutUint64(buf[*offset:], math.Float64bits(src.Height))
	*offset += 8

	return nil
}


// EncodePointMessage encodes a Point to self-describing message format.
// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// This format is suitable for persistence, network transmission, and cross-service communication.
func EncodePointMessage(src *Point) ([]byte, error) {
	// Encode payload
	payload, err := EncodePoint(src)
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

// EncodeRectangleMessage encodes a Rectangle to self-describing message format.
// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// This format is suitable for persistence, network transmission, and cross-service communication.
func EncodeRectangleMessage(src *Rectangle) ([]byte, error) {
	// Encode payload
	payload, err := EncodeRectangle(src)
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



// EncodePointToWriter encodes a Point to wire format and writes it to the provided io.Writer.
// This enables streaming I/O without baked-in compression.
//
// Users can compose with any io.Writer implementation:
//   - File I/O: os.File
//   - Compression: gzip.Writer, zstd.Writer, etc.
//   - Network: net.Conn, http.ResponseWriter
//   - Encryption: custom crypto.Writer
//   - Metrics: custom byte counting wrappers
func EncodePointToWriter(src *Point, w io.Writer) error {
	size := calculatePointSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodePoint(src, buf, &offset); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}

// EncodeRectangleToWriter encodes a Rectangle to wire format and writes it to the provided io.Writer.
// This enables streaming I/O without baked-in compression.
//
// Users can compose with any io.Writer implementation:
//   - File I/O: os.File
//   - Compression: gzip.Writer, zstd.Writer, etc.
//   - Network: net.Conn, http.ResponseWriter
//   - Encryption: custom crypto.Writer
//   - Metrics: custom byte counting wrappers
func EncodeRectangleToWriter(src *Rectangle, w io.Writer) error {
	size := calculateRectangleSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeRectangle(src, buf, &offset); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}
