package nwenc

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDecode(t *testing.T) {
	type outType struct {
		offset []int64
		err    error
	}
	tests := []struct {
		in  []byte
		out outType
	}{
		{
			[]byte{0, 0, 2},
			outType{[]int64{2}, nil},
		},
		{
			[]byte{16, 255, 5},
			outType{[]int64{0x10FF05}, nil},
		},
		{
			[]byte{16, 255, 5, 3, 241, 16},
			outType{[]int64{0x10FF05, 0x03F110}, nil},
		},
		{
			[]byte{16, 255},
			outType{[]int64{}, io.ErrUnexpectedEOF},
		},
	}

TestLoop:
	for idx, test := range tests {
		dec, _ := NewDecoder(3)
		r := bytes.NewBuffer(test.in)

		for i := 0; ; i++ {
			offset, err := dec.Decode(r)
			if err == io.EOF {
				continue TestLoop
			}
			if !reflect.DeepEqual(test.out.err, err) {
				t.Errorf("[%d] expected %v, but got %v", idx, test.out.err, err)
			}
			if err != nil {
				continue TestLoop
			}
			if test.out.offset[i] != offset {
				t.Errorf("[%d] expected %d, but got %d", idx, test.out.offset, offset)
			}
		}
	}
}

func TestDecodeString(t *testing.T) {
	type outType struct {
		s   string
		err error
	}
	tests := []struct {
		in  []byte
		out outType
	}{
		{
			[]byte{0, 0, 0},
			outType{"a", nil},
		},
		{
			[]byte{0, 0, 43},
			outType{"bcd", nil},
		},
		{
			[]byte{0, 0, 100},
			outType{"", &OffsetDecodeError{offset: 100}},
		},
	}

	// prepare OffsetDecoder
	f, err := os.Open(filepath.Join("testdata", "words.txt"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()
	om, err := NewAllReadOffsetMapper(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for idx, test := range tests {
		r := bytes.NewBuffer(test.in)
		dec, _ := NewDecoder(3)

		s, err := dec.DecodeString(r, om)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out.err, err)
		}
		if err != nil {
			continue
		}

		if test.out.s != s {
			t.Errorf("[%d] expected %#v, but got %#v", idx, test.out.s, s)
		}
	}
}
