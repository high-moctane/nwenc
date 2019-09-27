package nwenc

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNewAllReadOffsetMapper(t *testing.T) {
	tests := []struct {
		in  string
		out *AllReadOffsetMapper
	}{
		{
			"a\n",
			&AllReadOffsetMapper{
				offsetToS: map[int64]string{0: "a"},
				sToOffset: map[string]int64{"a": 0},
			},
		},
		{
			"a\nbcd\nefg\nhijk\n",
			&AllReadOffsetMapper{
				offsetToS: map[int64]string{0: "a", 2: "bcd", 6: "efg", 10: "hijk"},
				sToOffset: map[string]int64{"a": 0, "bcd": 2, "efg": 6, "hijk": 10},
			},
		},
	}

	for idx, test := range tests {
		r := strings.NewReader(test.in)
		m, err := NewAllReadOffsetMapper(r)
		if err != nil {
			t.Errorf("[%d] unexoected error: %v", idx, err)
			continue
		}

		if !reflect.DeepEqual(test.out, m) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out, m)
		}
	}
}

func TestAllReadOffsetMapper_OffsetEncode(t *testing.T) {
	type outType struct {
		offset int64
		err    error
	}
	tests := []struct {
		in  string
		out outType
	}{
		{
			"a",
			outType{0, nil},
		},
		{
			"aaaabbbbccccddddeeeeffffgggghhhhiii",
			outType{2, nil},
		},
		{
			"abcd",
			outType{38, nil},
		},
		{
			"bcd",
			outType{43, nil},
		},
		{
			"defgh",
			outType{47, nil},
		},
		{
			"deg",
			outType{53, nil},
		},
		{
			"ijk",
			outType{57, nil},
		},
		{
			"ijkl",
			outType{61, nil},
		},
		{
			"0",
			outType{0, &OffsetEncodeError{s: "0"}},
		},
		{
			"aaaaa",
			outType{0, &OffsetEncodeError{s: "aaaaa"}},
		},
		{
			"ijkk",
			outType{0, &OffsetEncodeError{s: "ijkk"}},
		},
		{
			"z",
			outType{0, &OffsetEncodeError{s: "z"}},
		},
	}

	// open test data
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
		offset, err := om.OffsetEncode(test.in)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out.err, err)
			continue
		}
		if err != nil {
			continue
		}
		if test.out.offset != offset {
			t.Errorf("[%d] expected %d, but got %d", idx, test.out.offset, offset)
		}
	}
}

func TestAllReadOffsetMapper_OffsetDecode(t *testing.T) {
	type outType struct {
		s   string
		err error
	}
	tests := []struct {
		in  int64
		out outType
	}{
		{
			0, outType{"a", nil},
		},
		{
			2, outType{"aaaabbbbccccddddeeeeffffgggghhhhiii", nil},
		},
		{
			38, outType{"abcd", nil},
		},
		{
			37, outType{"", &OffsetDecodeError{offset: 37}},
		},
		{
			65, outType{"", &OffsetDecodeError{offset: 65}},
		},
		{
			66, outType{"", &OffsetDecodeError{offset: 66}},
		},
	}

	// open test data
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
		s, err := om.OffsetDecode(test.in)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out.err, err)
			continue
		}

		if test.out.s != s {
			t.Errorf("[%d] expected %#v, but got %#v", idx, test.out.s, s)
		}
	}
}

func BenchmarkAllReadOffsetMapper_OffsetEncode(b *testing.B) {
	queries := []string{
		"a",
		"aaaabbbbccccddddeeeeffffgggghhhhiii",
		"abcd",
		"bcd",
		"defgh",
		"deg",
		"ijk",
		"ijkl",
	}

	// open file
	f, err := os.Open(filepath.Join("testdata", "words.txt"))
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	om, err := NewAllReadOffsetMapper(f)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < b.N; i++ {
		om.OffsetEncode(queries[i%len(queries)])
	}
}

func BenchmarkAllReadOffsetMapper_OffsetDecode(b *testing.B) {
	queries := []int64{0, 2, 38, 43, 47, 53, 57, 61}

	// open file
	f, err := os.Open(filepath.Join("testdata", "words.txt"))
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	om, err := NewAllReadOffsetMapper(f)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < b.N; i++ {
		om.OffsetDecode(queries[i%len(queries)])
	}
}

func TestSeekOffsetMapper_OffsetEncode(t *testing.T) {
	type outType struct {
		offset int64
		err    error
	}
	tests := []struct {
		in  string
		out outType
	}{
		{
			"a",
			outType{0, nil},
		},
		{
			"aaaabbbbccccddddeeeeffffgggghhhhiii",
			outType{2, nil},
		},
		{
			"abcd",
			outType{38, nil},
		},
		{
			"bcd",
			outType{43, nil},
		},
		{
			"defgh",
			outType{47, nil},
		},
		{
			"deg",
			outType{53, nil},
		},
		{
			"ijk",
			outType{57, nil},
		},
		{
			"ijkl",
			outType{61, nil},
		},
		{
			"0",
			outType{0, &OffsetEncodeError{s: "0"}},
		},
		{
			"aaaaa",
			outType{0, &OffsetEncodeError{s: "aaaaa"}},
		},
		{
			"ijkk",
			outType{0, &OffsetEncodeError{s: "ijkk"}},
		},
		{
			"z",
			outType{0, &OffsetEncodeError{s: "z"}},
		},
	}

	// open test data
	f, err := os.Open(filepath.Join("testdata", "words.txt"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	om := NewSeekOffsetMapper(f, info.Size())

	for idx, test := range tests {
		offset, err := om.OffsetEncode(test.in)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out.err, err)
			continue
		}
		if err != nil {
			continue
		}
		if test.out.offset != offset {
			t.Errorf("[%d] expected %d, but got %d", idx, test.out.offset, offset)
		}
	}
}

func BenchmarkSeekOffsetMapper_OffsetEncode(b *testing.B) {
	queries := []string{
		"a",
		"aaaabbbbccccddddeeeeffffgggghhhhiii",
		"abcd",
		"bcd",
		"defgh",
		"deg",
		"ijk",
		"ijkl",
	}

	// open file
	f, err := os.Open(filepath.Join("testdata", "words.txt"))
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	om := NewSeekOffsetMapper(f, info.Size())
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < b.N; i++ {
		for _, q := range queries {
			om.OffsetEncode(q)
		}
	}
}

func TestSeekOffsetMapper_OffsetDecode(t *testing.T) {
	type outType struct {
		s   string
		err error
	}
	tests := []struct {
		in  int64
		out outType
	}{
		{
			0, outType{"a", nil},
		},
		{
			2, outType{"aaaabbbbccccddddeeeeffffgggghhhhiii", nil},
		},
		{
			38, outType{"abcd", nil},
		},
		{
			37, outType{"", &OffsetDecodeError{offset: 37}},
		},
		{
			65, outType{"", &OffsetDecodeError{offset: 65}},
		},
		{
			66, outType{"", &OffsetDecodeError{offset: 66}},
		},
	}

	// open test data
	f, err := os.Open(filepath.Join("testdata", "words.txt"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	om := NewSeekOffsetMapper(f, info.Size())

	for idx, test := range tests {
		s, err := om.OffsetDecode(test.in)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out.err, err)
			continue
		}

		if test.out.s != s {
			t.Errorf("[%d] expected %#v, but got %#v", idx, test.out.s, s)
		}
	}
}

func BenchmarkSeekOffsetMapper_OffsetDecode(b *testing.B) {
	queries := []int64{0, 2, 38, 43, 47, 53, 57, 61}

	// open file
	f, err := os.Open(filepath.Join("testdata", "words.txt"))
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	om := NewSeekOffsetMapper(f, info.Size())
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	// make cache

	for i := 0; i < b.N; i++ {
		om.OffsetDecode(queries[i%len(queries)])
	}
}

func TestCachedSeekOffsetMapper_OffsetEncode(t *testing.T) {
	type outType struct {
		offset int64
		err    error
	}
	tests := []struct {
		in  string
		out outType
	}{
		{
			"a",
			outType{0, nil},
		},
		{
			"aaaabbbbccccddddeeeeffffgggghhhhiii",
			outType{2, nil},
		},
		{
			"abcd",
			outType{38, nil},
		},
		{
			"bcd",
			outType{43, nil},
		},
		{
			"defgh",
			outType{47, nil},
		},
		{
			"deg",
			outType{53, nil},
		},
		{
			"ijk",
			outType{57, nil},
		},
		{
			"ijkl",
			outType{61, nil},
		},
		{
			"0",
			outType{0, &OffsetEncodeError{s: "0"}},
		},
		{
			"aaaaa",
			outType{0, &OffsetEncodeError{s: "aaaaa"}},
		},
		{
			"ijkk",
			outType{0, &OffsetEncodeError{s: "ijkk"}},
		},
		{
			"z",
			outType{0, &OffsetEncodeError{s: "z"}},
		},
	}

	// open test data
	f, err := os.Open(filepath.Join("testdata", "words.txt"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	om := NewCachedSeekOffsetMapper(f, info.Size())

	for idx, test := range tests {
		offset, err := om.OffsetEncode(test.in)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out.err, err)
			continue
		}
		if err != nil {
			continue
		}
		if test.out.offset != offset {
			t.Errorf("[%d] expected %d, but got %d", idx, test.out.offset, offset)
		}
	}
}

func BenchmarkCachedSeekOffsetMapper_OffsetEncode(b *testing.B) {
	queries := []string{
		"a",
		"aaaabbbbccccddddeeeeffffgggghhhhiii",
		"abcd",
		"bcd",
		"defgh",
		"deg",
		"ijk",
		"ijkl",
	}

	// open file
	f, err := os.Open(filepath.Join("testdata", "words.txt"))
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	om := NewCachedSeekOffsetMapper(f, info.Size())

	for i := 0; i < b.N; i++ {
		om.OffsetEncode(queries[i%len(queries)])
	}
}

func TestCachedSeekOffsetMapper_OffsetDecode(t *testing.T) {
	type outType struct {
		s   string
		err error
	}
	tests := []struct {
		in  int64
		out outType
	}{
		{
			0, outType{"a", nil},
		},
		{
			2, outType{"aaaabbbbccccddddeeeeffffgggghhhhiii", nil},
		},
		{
			38, outType{"abcd", nil},
		},
		{
			37, outType{"", &OffsetDecodeError{offset: 37}},
		},
		{
			65, outType{"", &OffsetDecodeError{offset: 65}},
		},
		{
			66, outType{"", &OffsetDecodeError{offset: 66}},
		},
	}

	// open test data
	f, err := os.Open(filepath.Join("testdata", "words.txt"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	om := NewCachedSeekOffsetMapper(f, info.Size())

	for idx, test := range tests {
		s, err := om.OffsetDecode(test.in)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out.err, err)
			continue
		}

		if test.out.s != s {
			t.Errorf("[%d] expected %#v, but got %#v", idx, test.out.s, s)
		}
	}
}

func BenchmarkCachedSeekOffsetMapper_OffsetDecode(b *testing.B) {
	queries := []int64{0, 2, 38, 43, 47, 53, 57, 61}

	// open file
	f, err := os.Open(filepath.Join("testdata", "words.txt"))
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	om := NewCachedSeekOffsetMapper(f, info.Size())

	for i := 0; i < b.N; i++ {
		om.OffsetDecode(queries[i%len(queries)])
	}
}

func TestOffsetNode_Add(t *testing.T) {
	type inType struct {
		s      string
		offset int64
	}
	tests := []struct {
		in  []inType
		out *offsetNode
	}{
		{
			[]inType{{"a", 0}},
			&offsetNode{s: "a", offset: 0},
		},
		{
			[]inType{{"b", 2}, {"a", 0}, {"c", 4}},
			&offsetNode{
				s:      "b",
				offset: 2,
				left: &offsetNode{
					s:      "a",
					offset: 0,
				},
				right: &offsetNode{
					s:      "c",
					offset: 4,
				},
			},
		},
		{
			[]inType{{"d", 6}, {"b", 2}, {"a", 0}, {"e", 8}, {"c", 4}},
			&offsetNode{
				s:      "d",
				offset: 6,
				left: &offsetNode{
					s:      "b",
					offset: 2,
					left: &offsetNode{
						s:      "a",
						offset: 0,
					},
					right: &offsetNode{
						s:      "c",
						offset: 4,
					},
				},
				right: &offsetNode{
					s:      "e",
					offset: 8,
				},
			},
		},
	}

	for idx, test := range tests {
		var root *offsetNode
		for _, in := range test.in {
			root = root.add(in.s, in.offset)
		}

		if !reflect.DeepEqual(test.out, root) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out, root)
		}
	}
}

func TestOffsetNode_SearchString(t *testing.T) {
	type nodeInput struct {
		s      string
		offset int64
	}
	type inType struct {
		nodes           []nodeInput
		s               string
		inLeft, inRight int64
	}
	type outType struct {
		offset, left, right int64
		ok                  bool
	}
	tests := []struct {
		in  inType
		out outType
	}{
		{
			inType{
				[]nodeInput{},
				"a",
				0, 100,
			},
			outType{0, 0, 100, false},
		},
		{
			inType{
				[]nodeInput{{"a", 0}},
				"a",
				0, 100,
			},
			outType{0, 0, 100, true},
		},
		{
			inType{
				[]nodeInput{{"a", 0}},
				"b",
				0, 100,
			},
			outType{0, 0, 100, false},
		},
		{
			inType{
				[]nodeInput{{"b", 2}, {"a", 0}, {"c", 4}},
				"a",
				0, 100,
			},
			outType{0, 0, 2, true},
		},
		{
			inType{
				[]nodeInput{{"b", 2}, {"a", 0}, {"c", 4}},
				"d",
				0, 100,
			},
			outType{0, 4, 100, false},
		},
		{
			inType{
				[]nodeInput{{"d", 8}, {"b", 4}, {"a", 2}, {"e", 10}, {"c", 6}},
				"c",
				0, 100,
			},
			outType{6, 4, 8, true},
		},
		{
			inType{
				[]nodeInput{{"d", 8}, {"b", 4}, {"a", 2}, {"e", 10}, {"c", 6}},
				"cc",
				0, 100,
			},
			outType{0, 6, 8, false},
		},
		{
			inType{
				[]nodeInput{{"d", 8}, {"b", 4}, {"a", 2}, {"e", 10}, {"c", 6}},
				"0",
				0, 100,
			},
			outType{0, 0, 2, false},
		},
		{
			inType{
				[]nodeInput{{"d", 8}, {"b", 4}, {"a", 2}, {"e", 10}, {"c", 6}},
				"dd",
				0, 100,
			},
			outType{0, 8, 10, false},
		},
		{
			inType{
				[]nodeInput{{"d", 8}, {"b", 4}, {"a", 2}, {"e", 10}, {"c", 6}},
				"ee",
				0, 100,
			},
			outType{0, 10, 100, false},
		},
	}

	for idx, test := range tests {
		var pn *offsetNode
		for _, in := range test.in.nodes {
			pn = pn.add(in.s, in.offset)
		}

		offset, left, right, ok := pn.searchString(test.in.s, test.in.inLeft, test.in.inRight)
		if test.out.offset != offset {
			t.Errorf("[%d] offset expected %d, but got %d", idx, test.out.offset, offset)
		}
		if test.out.left != left {
			t.Errorf("[%d] left expected %d, but got %d", idx, test.out.left, left)
		}
		if test.out.right != right {
			t.Errorf("[%d] right expected %d, but got %d", idx, test.out.right, right)
		}
		if test.out.ok != ok {
			t.Errorf("[%d] ok expected %v, but got %v", idx, test.out.ok, ok)
		}
	}
}

func TestOffsetNode_Searchoffset(t *testing.T) {
	type nodeInput struct {
		s      string
		offset int64
	}
	type inType struct {
		nodes                   []nodeInput
		offset, inLeft, inRight int64
	}
	type outType struct {
		s           string
		left, right int64
		ok          bool
	}
	tests := []struct {
		in  inType
		out outType
	}{
		{
			inType{
				[]nodeInput{},
				0,
				0, 100,
			},
			outType{"", 0, 100, false},
		},
		{
			inType{
				[]nodeInput{{"a", 0}},
				0,
				0, 100,
			},
			outType{"a", 0, 100, true},
		},
		{
			inType{
				[]nodeInput{{"a", 0}},
				2,
				0, 100,
			},
			outType{"", 0, 100, false},
		},
		{
			inType{
				[]nodeInput{{"b", 2}, {"a", 0}, {"c", 4}},
				0,
				0, 100,
			},
			outType{"a", 0, 2, true},
		},
		{
			inType{
				[]nodeInput{{"b", 2}, {"a", 0}, {"c", 4}},
				8,
				0, 100,
			},
			outType{"", 4, 100, false},
		},
		{
			inType{
				[]nodeInput{{"d", 8}, {"b", 4}, {"a", 2}, {"e", 10}, {"c", 6}},
				6,
				0, 100,
			},
			outType{"c", 4, 8, true},
		},
		{
			inType{
				[]nodeInput{{"d", 8}, {"b", 4}, {"a", 2}, {"e", 10}, {"c", 6}},
				7,
				0, 100,
			},
			outType{"", 6, 8, false},
		},
		{
			inType{
				[]nodeInput{{"d", 8}, {"b", 4}, {"a", 2}, {"e", 10}, {"c", 6}},
				0,
				0, 100,
			},
			outType{"", 0, 2, false},
		},
		{
			inType{
				[]nodeInput{{"d", 8}, {"b", 4}, {"a", 2}, {"e", 10}, {"c", 6}},
				9,
				0, 100,
			},
			outType{"", 8, 10, false},
		},
		{
			inType{
				[]nodeInput{{"d", 8}, {"b", 4}, {"a", 2}, {"e", 10}, {"c", 6}},
				12,
				0, 100,
			},
			outType{"", 10, 100, false},
		},
	}

	for idx, test := range tests {
		var pn *offsetNode
		for _, in := range test.in.nodes {
			pn = pn.add(in.s, in.offset)
		}

		s, left, right, ok := pn.searchoffset(test.in.offset, test.in.inLeft, test.in.inRight)
		if test.out.s != s {
			t.Errorf("[%d] s expected %s, but got %s", idx, test.out.s, s)
		}
		if test.out.left != left {
			t.Errorf("[%d] left expected %d, but got %d", idx, test.out.left, left)
		}
		if test.out.right != right {
			t.Errorf("[%d] right expected %d, but got %d", idx, test.out.right, right)
		}
		if test.out.ok != ok {
			t.Errorf("[%d] ok expected %v, but got %v", idx, test.out.ok, ok)
		}
	}
}
