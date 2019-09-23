package nwenc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

type Encoder struct {
	l int // byte length
}

func NewEncoder(byteLen int) (*Encoder, error) {
	if byteLen < 1 || 8 < byteLen {
		return nil, fmt.Errorf("invalid byte length: %d", byteLen)
	}

	return &Encoder{l: byteLen}, nil
}

func (e *Encoder) Encode(w io.Writer, pos int64) error {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, pos); err != nil {
		return err
	}
	b := buf.Bytes()
	if _, err := w.Write(b[8-e.l:]); err != nil {
		return err
	}
	return nil
}

func (e *Encoder) EncodeString(w io.Writer, pd PosEncoder, s string) error {
	pos, err := pd.PosEncode(s)
	if err != nil {
		return err
	}
	if err := e.Encode(w, pos); err != nil {
		return err
	}
	return nil
}
