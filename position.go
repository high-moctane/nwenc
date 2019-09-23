package nwenc

import "fmt"

type PosEncoder interface {
	PosEncode(s string) (pos int64, err error)
}

type PosEncodeError struct {
	s string
}

func (e *PosEncodeError) Error() string {
	return fmt.Sprintf("string cannot encode: %#v", e.s)
}

type PosDecoder interface {
	PosDecode(pos int64) (s string, err error)
}

type PosDecodeError struct {
	pos int64
}

func (e *PosDecodeError) Error() string {
	return fmt.Sprintf("pos cannot decode: %v", e.pos)
}

type PosMapper interface {
	PosEncoder
	PosDecoder
}
