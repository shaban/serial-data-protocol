package primitives

import (
	"io"
	"encoding/binary"
	"math"
)

// calculateAllPrimitivesSize calculates the wire format size for AllPrimitives.
func calculateAllPrimitivesSize(src *AllPrimitives) int {
	size := 0
	// Field: U8Field
	size += 1
	// Field: U16Field
	size += 2
	// Field: U32Field
	size += 4
	// Field: U64Field
	size += 8
	// Field: I8Field
	size += 1
	// Field: I16Field
	size += 2
	// Field: I32Field
	size += 4
	// Field: I64Field
	size += 8
	// Field: F32Field
	size += 4
	// Field: F64Field
	size += 8
	// Field: BoolField
	size += 1
	// Field: StrField
	size += 4 + len(src.StrField)
	return size
}

// EncodeAllPrimitives encodes a AllPrimitives to wire format.
// It returns the encoded bytes or an error.
func EncodeAllPrimitives(src *AllPrimitives) ([]byte, error) {
	size := calculateAllPrimitivesSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeAllPrimitives(src, buf, &offset); err != nil {
		return nil, err
	}
	return buf, nil
}


// encodeAllPrimitives is the helper function that encodes AllPrimitives fields.
func encodeAllPrimitives(src *AllPrimitives, buf []byte, offset *int) error {
	// Field: U8Field (u8)
	buf[*offset] = src.U8Field
	*offset++

	// Field: U16Field (u16)
	binary.LittleEndian.PutUint16(buf[*offset:], src.U16Field)
	*offset += 2

	// Field: U32Field (u32)
	binary.LittleEndian.PutUint32(buf[*offset:], src.U32Field)
	*offset += 4

	// Field: U64Field (u64)
	binary.LittleEndian.PutUint64(buf[*offset:], src.U64Field)
	*offset += 8

	// Field: I8Field (i8)
	buf[*offset] = uint8(src.I8Field)
	*offset++

	// Field: I16Field (i16)
	binary.LittleEndian.PutUint16(buf[*offset:], uint16(src.I16Field))
	*offset += 2

	// Field: I32Field (i32)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(src.I32Field))
	*offset += 4

	// Field: I64Field (i64)
	binary.LittleEndian.PutUint64(buf[*offset:], uint64(src.I64Field))
	*offset += 8

	// Field: F32Field (f32)
	binary.LittleEndian.PutUint32(buf[*offset:], math.Float32bits(src.F32Field))
	*offset += 4

	// Field: F64Field (f64)
	binary.LittleEndian.PutUint64(buf[*offset:], math.Float64bits(src.F64Field))
	*offset += 8

	// Field: BoolField (bool)
	if src.BoolField {
		buf[*offset] = 1
	} else {
		buf[*offset] = 0
	}
	*offset++

	// Field: StrField (str)
	binary.LittleEndian.PutUint32(buf[*offset:], uint32(len(src.StrField)))
	*offset += 4
	copy(buf[*offset:], src.StrField)
	*offset += len(src.StrField)
	return nil
}


// EncodeAllPrimitivesMessage encodes a AllPrimitives to self-describing message format.
// The message includes a 10-byte header: [SDP:3][version:1][type_id:2][length:4][payload:N]
// This format is suitable for persistence, network transmission, and cross-service communication.
func EncodeAllPrimitivesMessage(src *AllPrimitives) ([]byte, error) {
	// Encode payload
	payload, err := EncodeAllPrimitives(src)
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



// EncodeAllPrimitivesToWriter encodes a AllPrimitives to wire format and writes it to the provided io.Writer.
// This enables streaming I/O without baked-in compression.
//
// Users can compose with any io.Writer implementation:
//   - File I/O: os.File
//   - Compression: gzip.Writer, zstd.Writer, etc.
//   - Network: net.Conn, http.ResponseWriter
//   - Encryption: custom crypto.Writer
//   - Metrics: custom byte counting wrappers
func EncodeAllPrimitivesToWriter(src *AllPrimitives, w io.Writer) error {
	size := calculateAllPrimitivesSize(src)
	buf := make([]byte, size)
	offset := 0
	if err := encodeAllPrimitives(src, buf, &offset); err != nil {
		return err
	}
	_, err := w.Write(buf)
	return err
}
