package nwenc

import "fmt"

// OffsetEncoder is the interface which can map a string to an int64 offset.
// The offset is the offset where the string appears in a text file.
type OffsetEncoder interface {
	OffsetEncode(s string) (offset int64, err error)
}

// OffsetEncodeError is returned when OffsetEncode fails.
type OffsetEncodeError struct {
	s string
}

func (e *OffsetEncodeError) Error() string {
	return fmt.Sprintf("string cannot encode: %#v", e.s)
}

// OffsetDecoder is the interface which can map an int64 offset to a string.
// The offset is the offset where the string appears in a text file.
type OffsetDecoder interface {
	OffsetDecode(offset int64) (s string, err error)
}

// OffsetDecodeError is returned when OffsetDecode fails.
type OffsetDecodeError struct {
	offset int64
}

func (e *OffsetDecodeError) Error() string {
	return fmt.Sprintf("offset cannot decode: %v", e.offset)
}

// OffsetMapper implements both OffsetEncoder and OffsetDecoder.
type OffsetMapper interface {
	OffsetEncoder
	OffsetDecoder
}
