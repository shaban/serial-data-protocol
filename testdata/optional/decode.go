package optional

import (
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
