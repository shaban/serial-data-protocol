package arrays

import (
	"encoding/binary"
	"math"
	"io"
	"unsafe"
)

// calculateArraysOfPrimitivesSize calculates the wire format size for ArraysOfPrimitives.
func calculateArraysOfPrimitivesSize(src *ArraysOfPrimitives) int {
	size := 0
	// Field: U8Array
	size += 4
	size += len(src.U8Array)
	// Field: U32Array
	size += 4
	size += len(src.U32Array) * 4
	// Field: F64Array
	size += 4
	size += len(src.F64Array) * 8
	// Field: StrArray
	size += 4
	for i := range src.StrArray {
		size += 4 + len(src.StrArray[i])
	}
	// Field: BoolArray
	size += 4
	size += len(src.BoolArray)
	return size
}

// EncodeArraysOfPrimitives encodes a ArraysOfPrimitives to wire format.
// It returns the encoded bytes or an error.
func EncodeArraysOfPrimitives(src *ArraysOfPrimitives) ([]byte, error) {
	size := calculateArraysOfPrimitivesSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeArraysOfPrimitives(src, buf, &offset); err != nil {
		return nil, err
	}
	return buf, nil
}

// calculateItemSize calculates the wire format size for Item.
func calculateItemSize(src *Item) int {
	size := 0
	// Field: Id
	size += 4
	// Field: Name
	size += 4 + len(src.Name)
	return size
}

// EncodeItem encodes a Item to wire format.
// It returns the encoded bytes or an error.
func EncodeItem(src *Item) ([]byte, error) {
	size := calculateItemSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeItem(src, buf, &offset); err != nil {
		return nil, err
	}
	return buf, nil
}

// calculateArraysOfStructsSize calculates the wire format size for ArraysOfStructs.
func calculateArraysOfStructsSize(src *ArraysOfStructs) int {
	size := 0
	// Field: Items
	size += 4
	for i := range src.Items {
		size += calculateItemSize(&src.Items[i])
	}
	// Field: Count
	size += 4
	return size
}

// EncodeArraysOfStructs encodes a ArraysOfStructs to wire format.
// It returns the encoded bytes or an error.
func EncodeArraysOfStructs(src *ArraysOfStructs) ([]byte, error) {
	size := calculateArraysOfStructsSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeArraysOfStructs(src, buf, &offset); err != nil {
		return nil, err
	}
	return buf, nil
}


// encodeArraysOfPrimitives is the helper function that encodes ArraysOfPrimitives fields.
func encodeArraysOfPrimitives(src *ArraysOfPrimitives, buf []byte, offset *int) error {
	// Field: U8Array ([]u8)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.U8Array)))
	*offset += 4

	// Bulk copy optimization for primitive arrays
	if len(src.U8Array) > 0 {
		copy(buf[*offset:], src.U8Array)
		*offset += len(src.U8Array)
	}
	// Field: U32Array ([]u32)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.U32Array)))
	*offset += 4

	// Bulk copy optimization for primitive arrays
	if len(src.U32Array) > 0 {
		// Cast slice to bytes for bulk copy
		bytes := unsafe.Slice((*byte)(unsafe.Pointer(&src.U32Array[0])), len(src.U32Array)*4)
		copy(buf[*offset:], bytes)
		*offset += len(src.U32Array)*4
	}
	// Field: F64Array ([]f64)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.F64Array)))
	*offset += 4

	for i := range src.F64Array {
		binary.LittleEndian.PutUint64(buf[*offset:], math.Float64bits(src.F64Array[i]))
		*offset += 8
	}
	// Field: StrArray ([]str)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.StrArray)))
	*offset += 4

	for i := range src.StrArray {
		binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.StrArray[i])))
		*offset += 4
		copy(buf[*offset:], src.StrArray[i])
		*offset += len(src.StrArray[i])
	}
	// Field: BoolArray ([]bool)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.BoolArray)))
	*offset += 4

	for i := range src.BoolArray {
		if src.BoolArray[i] {
			buf[*offset] = 1
		} else {
			buf[*offset] = 0
		}
		*offset++
	}
	return nil
}

// encodeItem is the helper function that encodes Item fields.
func encodeItem(src *Item, buf []byte, offset *int) error {
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

// encodeArraysOfStructs is the helper function that encodes ArraysOfStructs fields.
func encodeArraysOfStructs(src *ArraysOfStructs, buf []byte, offset *int) error {
	// Field: Items ([]Item)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.Items)))
	*offset += 4

	for i := range src.Items {
		if err := encodeItem(&src.Items[i], buf, offset); err != nil {
			return err
		}
	}
	// Field: Count (u32)
	binary.LittleEndian.PutUint32(buf[*offset:], src.Count)
	*offset += 4

	return nil
}


// EncodeArraysOfPrimitivesMessage encodes a ArraysOfPrimitives to self-describing message format.
// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// This format is suitable for persistence, network transmission, and cross-service communication.
func EncodeArraysOfPrimitivesMessage(src *ArraysOfPrimitives) ([]byte, error) {
	// Encode payload
	payload, err := EncodeArraysOfPrimitives(src)
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

// EncodeItemMessage encodes a Item to self-describing message format.
// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// This format is suitable for persistence, network transmission, and cross-service communication.
func EncodeItemMessage(src *Item) ([]byte, error) {
	// Encode payload
	payload, err := EncodeItem(src)
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

// EncodeArraysOfStructsMessage encodes a ArraysOfStructs to self-describing message format.
// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// This format is suitable for persistence, network transmission, and cross-service communication.
func EncodeArraysOfStructsMessage(src *ArraysOfStructs) ([]byte, error) {
	// Encode payload
	payload, err := EncodeArraysOfStructs(src)
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



// EncodeArraysOfPrimitivesToWriter encodes a ArraysOfPrimitives to wire format and writes it to the provided io.Writer.
// This enables streaming I/O without baked-in compression.
//
// Users can compose with any io.Writer implementation:
//   - File I/O: os.File
//   - Compression: gzip.Writer, zstd.Writer, etc.
//   - Network: net.Conn, http.ResponseWriter
//   - Encryption: custom crypto.Writer
//   - Metrics: custom byte counting wrappers
func EncodeArraysOfPrimitivesToWriter(src *ArraysOfPrimitives, w io.Writer) error {
	size := calculateArraysOfPrimitivesSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeArraysOfPrimitives(src, buf, &offset); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}

// EncodeItemToWriter encodes a Item to wire format and writes it to the provided io.Writer.
// This enables streaming I/O without baked-in compression.
//
// Users can compose with any io.Writer implementation:
//   - File I/O: os.File
//   - Compression: gzip.Writer, zstd.Writer, etc.
//   - Network: net.Conn, http.ResponseWriter
//   - Encryption: custom crypto.Writer
//   - Metrics: custom byte counting wrappers
func EncodeItemToWriter(src *Item, w io.Writer) error {
	size := calculateItemSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeItem(src, buf, &offset); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}

// EncodeArraysOfStructsToWriter encodes a ArraysOfStructs to wire format and writes it to the provided io.Writer.
// This enables streaming I/O without baked-in compression.
//
// Users can compose with any io.Writer implementation:
//   - File I/O: os.File
//   - Compression: gzip.Writer, zstd.Writer, etc.
//   - Network: net.Conn, http.ResponseWriter
//   - Encryption: custom crypto.Writer
//   - Metrics: custom byte counting wrappers
func EncodeArraysOfStructsToWriter(src *ArraysOfStructs, w io.Writer) error {
	size := calculateArraysOfStructsSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeArraysOfStructs(src, buf, &offset); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}
