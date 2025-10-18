package valid_complex

import (
	"errors"
)

// Message mode constants for self-describing messages
const (
	MessageMagic         = "SDP"  // Magic bytes identifying SDP messages
	MessageVersion  byte = '2'     // Protocol version 0.2.0
	MessageHeaderSize    = 10      // Total header size: 3+1+2+4 bytes
)

// Error variables for decode failures
var (
	ErrUnexpectedEOF      = errors.New("unexpected end of data")
	ErrInvalidUTF8        = errors.New("invalid UTF-8 string")
	ErrDataTooLarge       = errors.New("data exceeds 128MB limit")
	ErrArrayTooLarge      = errors.New("array count exceeds per-array limit")
	ErrTooManyElements    = errors.New("total elements exceed limit")
	ErrInvalidData        = errors.New("invalid or corrupted data")
	ErrInvalidMagic       = errors.New("invalid magic bytes (expected 'SDP')")
	ErrInvalidVersion     = errors.New("unsupported protocol version")
	ErrUnknownMessageType = errors.New("unknown message type ID")
)
