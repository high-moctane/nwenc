package nwenc

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// Encoder encodes an int64 to bytes.
type Encoder struct {
	l int // byte length
}

// NewEncoder returns an Encoder. The byteLen is the length of encoded bytes.
// The byteLen must be 1 <= byteLen <= 8.
func NewEncoder(byteLen int) (*Encoder, error) {
	if byteLen < 1 || 8 < byteLen {
		return nil, fmt.Errorf("invalid byte length: %d", byteLen)
	}

	return &Encoder{l: byteLen}, nil
}

// Encode encodes pos to bytes and writes it to w.
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

// EncodeString encodes s to bytes and writes it to w.
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
