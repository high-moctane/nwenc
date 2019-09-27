package nwenc

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		in  []int64
		out []byte
	}{
		{
			[]int64{},
			nil,
		},
		{
			[]int64{2},
			[]byte{0, 0, 2},
		},
		{
			[]int64{0x10FF05},
			[]byte{16, 255, 5},
		},
		{
			[]int64{0x10FF05, 0x03F110},
			[]byte{16, 255, 5, 3, 241, 16},
		},
	}

	for idx, test := range tests {
		buf := new(bytes.Buffer)
		enc, _ := NewEncoder(3)

		for _, offset := range test.in {
			enc.Encode(buf, offset)
		}

		if !reflect.DeepEqual(test.out, buf.Bytes()) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out, buf.Bytes())
		}
	}
}

func TestEncodeString(t *testing.T) {
	type outType struct {
		buf []byte
		err error
	}
	tests := []struct {
		in  string
		out outType
	}{
		{
			"a",
			outType{[]byte{0, 0, 0}, nil},
		},
		{
			"bcd",
			outType{[]byte{0, 0, 43}, nil},
		},
		{
			"z",
			outType{nil, &OffsetEncodeError{s: "z"}},
		},
	}

	// prepare OffsetEncoder
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
		buf := new(bytes.Buffer)
		enc, err := NewEncoder(3)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		err = enc.EncodeString(buf, om, test.in)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out.err, err)
		}
		if err != nil {
			continue
		}

		if !reflect.DeepEqual(test.out.buf, buf.Bytes()) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out.buf, buf.Bytes())
		}
	}
}
