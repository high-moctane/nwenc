package nwenc

import "fmt"

// PosEncoder is the interface which can map a string to an int64 position.
// The position is the offset where the string appears in a text file.
type PosEncoder interface {
	PosEncode(s string) (pos int64, err error)
}

// PosEncodeError is returned when PosEncode fails.
type PosEncodeError struct {
	s string
}

func (e *PosEncodeError) Error() string {
	return fmt.Sprintf("string cannot encode: %#v", e.s)
}

// PosDecoder is the interface which can map an int64 position to a string.
// The position is the offset where the string appears in a text file.
type PosDecoder interface {
	PosDecode(pos int64) (s string, err error)
}

// PosDecodeError is returned when PosDecode fails.
type PosDecodeError struct {
	pos int64
}

func (e *PosDecodeError) Error() string {
	return fmt.Sprintf("pos cannot decode: %v", e.pos)
}

// PosMapper implements both PosEncoder and PosDecoder.
type PosMapper interface {
	PosEncoder
	PosDecoder
}
