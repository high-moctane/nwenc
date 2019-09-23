package nwenc

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNewAllReadPosMapper(t *testing.T) {
	tests := []struct {
		in  string
		out *AllReadPosMapper
	}{
		{
			"a\n",
			&AllReadPosMapper{
				posToS: map[int64]string{0: "a"},
				sToPos: map[string]int64{"a": 0},
			},
		},
		{
			"a\nbcd\nefg\nhijk\n",
			&AllReadPosMapper{
				posToS: map[int64]string{0: "a", 2: "bcd", 6: "efg", 10: "hijk"},
				sToPos: map[string]int64{"a": 0, "bcd": 2, "efg": 6, "hijk": 10},
			},
		},
	}

	for idx, test := range tests {
		r := strings.NewReader(test.in)
		m, err := NewAllReadPosMapper(r)
		if err != nil {
			t.Errorf("[%d] unexoected error: %v", idx, err)
			continue
		}

		if !reflect.DeepEqual(test.out, m) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out, m)
		}
	}
}

func TestAllReadPosMapper_PosEncode(t *testing.T) {
	type outType struct {
		pos int64
		err error
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
			outType{0, &PosEncodeError{s: "0"}},
		},
		{
			"aaaaa",
			outType{0, &PosEncodeError{s: "aaaaa"}},
		},
		{
			"ijkk",
			outType{0, &PosEncodeError{s: "ijkk"}},
		},
		{
			"z",
			outType{0, &PosEncodeError{s: "z"}},
		},
	}

	// open test data
	f, err := os.Open(filepath.Join("testdata", "words.txt"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	pm, err := NewAllReadPosMapper(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for idx, test := range tests {
		pos, err := pm.PosEncode(test.in)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out.err, err)
			continue
		}
		if err != nil {
			continue
		}
		if test.out.pos != pos {
			t.Errorf("[%d] expected %d, but got %d", idx, test.out.pos, pos)
		}
	}
}

func TestAllReadPosMapper_PosDecode(t *testing.T) {
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
			37, outType{"", &PosDecodeError{pos: 37}},
		},
		{
			65, outType{"", &PosDecodeError{pos: 65}},
		},
		{
			66, outType{"", &PosDecodeError{pos: 66}},
		},
	}

	// open test data
	f, err := os.Open(filepath.Join("testdata", "words.txt"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	pm, err := NewAllReadPosMapper(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for idx, test := range tests {
		s, err := pm.PosDecode(test.in)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out.err, err)
			continue
		}

		if test.out.s != s {
			t.Errorf("[%d] expected %#v, but got %#v", idx, test.out.s, s)
		}
	}
}

func BenchmarkAllReadPosMapper_PosEncode(b *testing.B) {
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

	pm, err := NewAllReadPosMapper(f)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < b.N; i++ {
		for _, q := range queries {
			pm.PosEncode(q)
		}
	}
}

func BenchmarkAllReadPosMapper_PosDecode(b *testing.B) {
	queries := []int64{0, 2, 38, 43, 47, 53, 57, 61}

	// open file
	f, err := os.Open(filepath.Join("testdata", "words.txt"))
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}
	defer f.Close()

	pm, err := NewAllReadPosMapper(f)
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < b.N; i++ {
		for _, q := range queries {
			pm.PosDecode(q)
		}
	}
}

func TestSeekPosMapper_PosEncode(t *testing.T) {
	type outType struct {
		pos int64
		err error
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
			outType{0, &PosEncodeError{s: "0"}},
		},
		{
			"aaaaa",
			outType{0, &PosEncodeError{s: "aaaaa"}},
		},
		{
			"ijkk",
			outType{0, &PosEncodeError{s: "ijkk"}},
		},
		{
			"z",
			outType{0, &PosEncodeError{s: "z"}},
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

	pm := NewSeekPosMapper(f, info.Size())

	for idx, test := range tests {
		pos, err := pm.PosEncode(test.in)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out.err, err)
			continue
		}
		if err != nil {
			continue
		}
		if test.out.pos != pos {
			t.Errorf("[%d] expected %d, but got %d", idx, test.out.pos, pos)
		}
	}
}

func BenchmarkSeekPosMapper_PosEncode(b *testing.B) {
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

	pm := NewSeekPosMapper(f, info.Size())
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < b.N; i++ {
		for _, q := range queries {
			pm.PosEncode(q)
		}
	}
}

func TestSeekPosMapper_PosDecode(t *testing.T) {
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
			37, outType{"", &PosDecodeError{pos: 37}},
		},
		{
			65, outType{"", &PosDecodeError{pos: 65}},
		},
		{
			66, outType{"", &PosDecodeError{pos: 66}},
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

	pm := NewSeekPosMapper(f, info.Size())

	for idx, test := range tests {
		s, err := pm.PosDecode(test.in)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out.err, err)
			continue
		}

		if test.out.s != s {
			t.Errorf("[%d] expected %#v, but got %#v", idx, test.out.s, s)
		}
	}
}

func BenchmarkSeekPosMapper_PosDecode(b *testing.B) {
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

	pm := NewSeekPosMapper(f, info.Size())
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < b.N; i++ {
		for _, q := range queries {
			pm.PosDecode(q)
		}
	}
}

func TestCachedSeekPosMapper_PosEncode(t *testing.T) {
	type outType struct {
		pos int64
		err error
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
			outType{0, &PosEncodeError{s: "0"}},
		},
		{
			"aaaaa",
			outType{0, &PosEncodeError{s: "aaaaa"}},
		},
		{
			"ijkk",
			outType{0, &PosEncodeError{s: "ijkk"}},
		},
		{
			"z",
			outType{0, &PosEncodeError{s: "z"}},
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

	pm := NewCachedSeekPosMapper(f, info.Size())

	for idx, test := range tests {
		pos, err := pm.PosEncode(test.in)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out.err, err)
			continue
		}
		if err != nil {
			continue
		}
		if test.out.pos != pos {
			t.Errorf("[%d] expected %d, but got %d", idx, test.out.pos, pos)
		}
	}
}

func BenchmarkCachedSeekPosMapper_PosEncode(b *testing.B) {
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

	pm := NewCachedSeekPosMapper(f, info.Size())
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}

	for i := 0; i < b.N; i++ {
		for _, q := range queries {
			pm.PosEncode(q)
		}
	}
}

func TestCachedSeekPosMapper_PosDecode(t *testing.T) {
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
			37, outType{"", &PosDecodeError{pos: 37}},
		},
		{
			65, outType{"", &PosDecodeError{pos: 65}},
		},
		{
			66, outType{"", &PosDecodeError{pos: 66}},
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

	pm := NewCachedSeekPosMapper(f, info.Size())

	for idx, test := range tests {
		s, err := pm.PosDecode(test.in)
		if !reflect.DeepEqual(test.out.err, err) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out.err, err)
			continue
		}

		if test.out.s != s {
			t.Errorf("[%d] expected %#v, but got %#v", idx, test.out.s, s)
		}
	}
}

func BenchmarkCachedSeekPosMapper_PosDecode(b *testing.B) {
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

	// make cache
	inputs := []string{
		"a",
		"aaaabbbbccccddddeeeeffffgggghhhhiii",
		"abcd",
		"bcd",
		"defgh",
		"deg",
		"ijk",
		"ijkl",
	}
	pm := NewCachedSeekPosMapper(f, info.Size())
	if err != nil {
		b.Fatalf("unexpected error: %v", err)
	}
	for _, in := range inputs {
		pm.PosEncode(in)
	}

	for i := 0; i < b.N; i++ {
		for _, q := range queries {
			pm.PosDecode(q)
		}
	}
}

func TestPosNode_Add(t *testing.T) {
	type inType struct {
		s   string
		pos int64
	}
	tests := []struct {
		in  []inType
		out *posNode
	}{
		{
			[]inType{{"a", 0}},
			&posNode{s: "a", pos: 0},
		},
		{
			[]inType{{"b", 2}, {"a", 0}, {"c", 4}},
			&posNode{
				s:   "b",
				pos: 2,
				left: &posNode{
					s:   "a",
					pos: 0,
				},
				right: &posNode{
					s:   "c",
					pos: 4,
				},
			},
		},
		{
			[]inType{{"d", 6}, {"b", 2}, {"a", 0}, {"e", 8}, {"c", 4}},
			&posNode{
				s:   "d",
				pos: 6,
				left: &posNode{
					s:   "b",
					pos: 2,
					left: &posNode{
						s:   "a",
						pos: 0,
					},
					right: &posNode{
						s:   "c",
						pos: 4,
					},
				},
				right: &posNode{
					s:   "e",
					pos: 8,
				},
			},
		},
	}

	for idx, test := range tests {
		var root *posNode
		for _, in := range test.in {
			root = root.add(in.s, in.pos)
		}

		if !reflect.DeepEqual(test.out, root) {
			t.Errorf("[%d] expected %v, but got %v", idx, test.out, root)
		}
	}
}

func TestPosNode_SearchString(t *testing.T) {
	type nodeInput struct {
		s   string
		pos int64
	}
	type inType struct {
		nodes           []nodeInput
		s               string
		inLeft, inRight int64
	}
	type outType struct {
		pos, left, right int64
		ok               bool
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
		var pn *posNode
		for _, in := range test.in.nodes {
			pn = pn.add(in.s, in.pos)
		}

		pos, left, right, ok := pn.searchString(test.in.s, test.in.inLeft, test.in.inRight)
		if test.out.pos != pos {
			t.Errorf("[%d] pos expected %d, but got %d", idx, test.out.pos, pos)
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

func TestPosNode_SearchPos(t *testing.T) {
	type nodeInput struct {
		s   string
		pos int64
	}
	type inType struct {
		nodes                []nodeInput
		pos, inLeft, inRight int64
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
		var pn *posNode
		for _, in := range test.in.nodes {
			pn = pn.add(in.s, in.pos)
		}

		s, left, right, ok := pn.searchPos(test.in.pos, test.in.inLeft, test.in.inRight)
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
