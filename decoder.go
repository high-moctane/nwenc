package nwenc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Decoder decodes bytes to an int64 or a string.
type Decoder struct {
	l int // byte length
}

// NewDecoder returns Decoder. The byteLen is the length of bytes.
// The byteLen must be 1 <= bytelen <= 8.
func NewDecoder(byteLen int) (*Decoder, error) {
	if byteLen < 1 || 8 < byteLen {
		return nil, fmt.Errorf("invalid byte length: %d", byteLen)
	}
	return &Decoder{l: byteLen}, nil
}

// Decode reads r and decodes to the offset.
func (d *Decoder) Decode(r io.Reader) (offset int64, err error) {
	buf := make([]byte, 8)
	var n int
	if n, err = r.Read(buf[8-d.l:]); err != nil {
		return
	}
	if n != d.l {
		return 0, io.ErrUnexpectedEOF
	}
	if err = binary.Read(bytes.NewBuffer(buf), binary.BigEndian, &offset); err != nil {
		return
	}
	return
}

// DecodeString reads r and decodes to s.
func (d *Decoder) DecodeString(r io.Reader, od OffsetDecoder) (s string, err error) {
	offset, err := d.Decode(r)
	if err != nil {
		return
	}
	s, err = od.OffsetDecode(offset)
	if err != nil {
		return
	}
	return
}
