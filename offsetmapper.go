package nwenc

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// AllReadOffsetMapper is one of the implementation of OffsetMapper.
// It reads all of io.Reader in advance in order to map fast.
type AllReadOffsetMapper struct {
	sToOffset map[string]int64
	offsetToS map[int64]string
}

// NewAllReadOffsetMapper returns an AllReadOffsetMapper.
// This function reads all of io.Reader in advance in order to map fast.
func NewAllReadOffsetMapper(r io.Reader) (*AllReadOffsetMapper, error) {
	m := &AllReadOffsetMapper{
		offsetToS: map[int64]string{},
		sToOffset: map[string]int64{},
	}
	offset := 0
	line := []byte{}

	for {
		buf := make([]byte, 1)
		_, err := r.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		offset++

		if buf[0] == '\n' {
			if !utf8.Valid(line) {
				return nil, fmt.Errorf("invalid string at %d", offset)
			}

			s := string(line)
			first := int64(offset - len(line) - 1)
			m.offsetToS[first] = s
			m.sToOffset[s] = first

			line = []byte{}
			continue
		}

		line = append(line, buf[0])
	}

	return m, nil
}

// OffsetEncode is the implementation of OffsetEncoder. It works fast.
// When s is not found, it will return OffsetEncodeError.
func (m *AllReadOffsetMapper) OffsetEncode(s string) (offset int64, err error) {
	var ok bool
	offset, ok = m.sToOffset[s]
	if !ok {
		err = &OffsetEncodeError{s: s}
		return
	}
	return
}

// OffsetDecode is the implementation of OffsetDecode. It works fast.
// When offset is not found, it will return OffsetDecodeError.
func (m *AllReadOffsetMapper) OffsetDecode(offset int64) (s string, err error) {
	var ok bool
	s, ok = m.offsetToS[offset]
	if !ok {
		err = &OffsetDecodeError{offset: offset}
		return
	}
	return
}

// SeekOffsetMapper is the implementation of OffsetMapper.
// It seeks file each time when OffsetEncode or OffsetDecode are called, so it works slowly.
type SeekOffsetMapper struct {
	r    io.ReaderAt
	size int64
}

// NewSeekOffsetMapper returns an NewSeekOffsetMapper. The size is the total bytes of r.
// It seeks file each time when OffsetEncode or OffsetDecode are called, so it works slowly.
func NewSeekOffsetMapper(r io.ReaderAt, size int64) *SeekOffsetMapper {
	return &SeekOffsetMapper{r: r, size: size}
}

// OffsetEncode is the implementation of OffsetEncoder.
// This function works slow because it needs io.ReadAt seeking each time.
func (om *SeekOffsetMapper) OffsetEncode(s string) (offset int64, err error) {
	offset, ok, err := readerAtBinSearch(om.r, s, 0, om.size)
	if err != nil {
		return
	}
	if !ok {
		err = &OffsetEncodeError{s: s}
		return
	}
	return
}

// readerAtBinsearch searches s from r. The left and right is the offset which
// seeks from and to. When s is found, ok will be true.
func readerAtBinSearch(r io.ReaderAt, s string, left, right int64) (offset int64, ok bool, err error) {
	var midS string
	for left+1 < right {
		midS, offset, err = searchMidoffset(r, left, right)
		if err != nil {
			return
		}

		if s == midS {
			ok = true
			return
		} else if s < midS {
			right = left + (right-left)/2
		} else {
			left = left + (right-left)/2
		}
	}

	return
}

// searchMidoffset finds s and offset which exists r from left to right offsets.
func searchMidoffset(r io.ReaderAt, left, right int64) (s string, offset int64, err error) {
	offset, err = findBeginOfLine(r, left+(right-left)/2)
	if err != nil && err != io.EOF {
		return
	}

	s, err = readLine(r, offset)
	if err != nil && err != io.EOF {
		return
	}

	if err == io.EOF {
		err = nil
	}
	return
}

// findBeginOfLine finds the beginning of line which the line contains offset.
// It returns head offset int64.
func findBeginOfLine(r io.ReaderAt, offset int64) (first int64, err error) {
	for i := offset; i >= 0; i-- {
		if i == 0 {
			first = 0
			break
		}

		buf := make([]byte, 1)
		if _, err = r.ReadAt(buf, i); err != nil {
			return
		}

		if i == offset && rune(buf[0]) == '\n' {
			continue
		}
		if rune(buf[0]) == '\n' {
			first = i + 1
			break
		}
	}

	return
}

// readLine reads a line from offset to '\n' ('\n' is not included).
func readLine(r io.ReaderAt, offset int64) (s string, err error) {
	sBuf := []byte{}

	var bufLen int64 = 32
	var i int64
	for i = 0; ; i++ {
		buf := make([]byte, bufLen)
		if _, err = r.ReadAt(buf, offset+bufLen*i); err != nil {
			if err == io.EOF {
				sBuf = append(sBuf, buf...)
				break
			}
			return
		}
		sBuf = append(sBuf, buf...)
		if strings.ContainsRune(string(sBuf), '\n') {
			break
		}
	}
	s = strings.Split(string(sBuf), "\n")[0]
	return
}

// OffsetDecode is the implementation of OffsetDecoder.
// This function works slow because it needs io.ReadAt seeking each time.
func (om *SeekOffsetMapper) OffsetDecode(offset int64) (s string, err error) {
	if offset >= om.size {
		err = &OffsetDecodeError{offset: offset}
		return
	}

	var bufLen int64 = 32
	line := []byte{}
	var i int64
	for ; ; i++ {
		buf := make([]byte, bufLen)
		if _, err = om.r.ReadAt(buf, offset+bufLen*i); err != nil {
			if err == io.EOF {
				line = append(line, buf...)
				break
			}
			return
		}
		line = append(line, buf...)

		if strings.ContainsRune(string(line), '\n') {
			break
		}
	}

	if !utf8.Valid(line) {
		err = &OffsetDecodeError{offset: offset}
		return
	}
	if len(line) == 0 || rune(line[0]) == '\n' {
		err = &OffsetDecodeError{offset: offset}
		return
	}

	s = strings.Split(string(line), "\n")[0]
	err = nil
	return
}

// CachedSeekOffsetMapper is the implementation of OffsetMapper.
// It seeks io.ReaderAt when methods are called, but caches the results.
// It's takes shorter time rather than SeekOffsetMapper.
type CachedSeekOffsetMapper struct {
	r         io.ReaderAt
	size      int64
	cacheTree *offsetNode
	cacheMap  map[int64]string
}

// NewCachedSeekOffsetMapper returns a CachedSeekOffsetMapper. The size is total bytes of r.
// It seeks io.ReaderAt when methods are called, but caches the results.
// It's takes shorter time rather than SeekOffsetMapper.
func NewCachedSeekOffsetMapper(r io.ReaderAt, size int64) *CachedSeekOffsetMapper {
	return &CachedSeekOffsetMapper{
		r:         r,
		size:      size,
		cacheTree: nil,
		cacheMap:  map[int64]string{},
	}
}

// OffsetEncode is the implementation of OffsetEncoder.
// This method makes cache when it is called.
func (om *CachedSeekOffsetMapper) OffsetEncode(s string) (offset int64, err error) {
	offset, left, right, ok := om.cacheTree.searchString(s, 0, om.size)
	if ok {
		return
	}

	// binary search
	var midS string
	for left+1 < right {
		midS, offset, err = searchMidoffset(om.r, left, right)
		if err != nil {
			return
		}

		om.cacheTree = om.cacheTree.add(midS, offset)

		if s == midS {
			return
		} else if s < midS {
			right = left + (right-left)/2
		} else {
			left = left + (right-left)/2
		}
	}

	err = &OffsetEncodeError{s: s}
	return
}

// OffsetDecode is the implementation of OffsetDecoder.
func (om *CachedSeekOffsetMapper) OffsetDecode(offset int64) (s string, err error) {
	s, ok := om.cacheMap[offset]
	if ok {
		return
	}
	s, _, _, ok = om.cacheTree.searchoffset(offset, 0, om.size)
	if ok {
		return
	}

	som := NewSeekOffsetMapper(om.r, om.size)
	s, err = som.OffsetDecode(offset)
	if err != nil {
		return
	}

	om.cacheMap[offset] = s
	return
}

// offsetNode is the node which consists a binary search tree.
type offsetNode struct {
	s           string
	offset      int64
	left, right *offsetNode
}

// add adds new node into pn. The caller must update pn by root.
func (pn *offsetNode) add(s string, offset int64) (root *offsetNode) {
	if pn == nil {
		root = &offsetNode{s: s, offset: offset}
		return
	}

	if s < pn.s {
		pn.left = pn.left.add(s, offset)
	} else if s > pn.s {
		pn.right = pn.right.add(s, offset)
	}
	return pn
}

// searchString searches s from the range of inLeft to inRight offsets.
func (pn *offsetNode) searchString(s string, inLeft, inRight int64) (offset, left, right int64, ok bool) {
	f := func(node *offsetNode) int {
		if s < node.s {
			return -1
		} else if s > node.s {
			return 1
		}
		return 0
	}

	var node *offsetNode
	node, left, right = pn.search(f, inLeft, inRight)
	if node != nil {
		offset = node.offset
		ok = true
	}
	return
}

// searchoffset searches offset from the range of inLeft to inRight offsets.
func (pn *offsetNode) searchoffset(offset int64, inLeft, inRight int64) (s string, left, right int64, ok bool) {
	f := func(node *offsetNode) int { return int(offset - node.offset) }

	var node *offsetNode
	node, left, right = pn.search(f, inLeft, inRight)
	if node != nil {
		s = node.s
		ok = true
	}
	return
}

// search searches the node which pred returns true. The node will be nil when not found.
func (pn *offsetNode) search(pred func(*offsetNode) int, inLeft, inRight int64) (node *offsetNode, left, right int64) {
	if pn == nil {
		left = inLeft
		right = inRight
		return
	}

	if pred(pn) < 0 {
		return pn.left.search(pred, inLeft, pn.offset)
	} else if pred(pn) > 0 {
		return pn.right.search(pred, pn.offset, inRight)
	}

	node = pn
	left = inLeft
	right = inRight
	return
}
