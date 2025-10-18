package optional

import (
	"errors"
)

// Error variables for decode failures
var (
	ErrUnexpectedEOF      = errors.New("unexpected end of data")
	ErrInvalidUTF8        = errors.New("invalid UTF-8 string")
	ErrDataTooLarge       = errors.New("data exceeds 128MB limit")
	ErrArrayTooLarge      = errors.New("array count exceeds per-array limit")
	ErrTooManyElements    = errors.New("total elements exceed limit")
	ErrInvalidData        = errors.New("invalid or corrupted data")
)
