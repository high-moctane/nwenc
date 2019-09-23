package nwenc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type Decoder struct {
	l int // byte length
}

func NewDecoder(byteLen int) (*Decoder, error) {
	if byteLen < 1 || 8 < byteLen {
		return nil, fmt.Errorf("invalid byte length: %d", byteLen)
	}
	return &Decoder{l: byteLen}, nil
}

func (d *Decoder) Decode(r io.Reader) (pos int64, err error) {
	buf := make([]byte, 8)
	var n int
	if n, err = r.Read(buf[8-d.l:]); err != nil {
		return
	}
	if n != d.l {
		return 0, io.ErrUnexpectedEOF
	}
	if err = binary.Read(bytes.NewBuffer(buf), binary.BigEndian, &pos); err != nil {
		return
	}
	return
}

func (d *Decoder) DecodeString(r io.Reader, pd PosDecoder) (s string, err error) {
	pos, err := d.Decode(r)
	if err != nil {
		return
	}
	s, err = pd.PosDecode(pos)
	if err != nil {
		return
	}
	return
}
