package nwenc

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// AllReadPosMapper is one of the implementation of PosMapper.
// It reads all of io.Reader in advance in order to map fast.
type AllReadPosMapper struct {
	sToPos map[string]int64
	posToS map[int64]string
}

// NewAllReadPosMapper returns an AllReadPosMapper.
// This function reads all of io.Reader in advance in order to map fast.
func NewAllReadPosMapper(r io.Reader) (*AllReadPosMapper, error) {
	m := &AllReadPosMapper{
		posToS: map[int64]string{},
		sToPos: map[string]int64{},
	}
	pos := 0
	line := []byte{}

	for {
		buf := make([]byte, 1)
		_, err := r.Read(buf)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		pos++

		if buf[0] == '\n' {
			if !utf8.Valid(line) {
				return nil, fmt.Errorf("invalid string at %d", pos)
			}

			s := string(line)
			first := int64(pos - len(line) - 1)
			m.posToS[first] = s
			m.sToPos[s] = first

			line = []byte{}
			continue
		}

		line = append(line, buf[0])
	}

	return m, nil
}

// PosEncode is the implementation of PosEncoder. It works fast.
// When s is not found, it will return PosEncodeError.
func (m *AllReadPosMapper) PosEncode(s string) (pos int64, err error) {
	var ok bool
	pos, ok = m.sToPos[s]
	if !ok {
		err = &PosEncodeError{s: s}
		return
	}
	return
}

// PosDecode is the implementation of PosDecode. It works fast.
// When pos is not found, it will return PosDecodeError.
func (m *AllReadPosMapper) PosDecode(pos int64) (s string, err error) {
	var ok bool
	s, ok = m.posToS[pos]
	if !ok {
		err = &PosDecodeError{pos: pos}
		return
	}
	return
}

// SeekPosMapper is the implementation of PosMapper.
// It seeks file each time when PosEncode or PosDecode are called, so it works slowly.
type SeekPosMapper struct {
	r    io.ReaderAt
	size int64
}

// NewSeekPosMapper returns an NewSeekPosMapper. The size is the total bytes of r.
// It seeks file each time when PosEncode or PosDecode are called, so it works slowly.
func NewSeekPosMapper(r io.ReaderAt, size int64) *SeekPosMapper {
	return &SeekPosMapper{r: r, size: size}
}

// PosEncode is the implementation of PosEncoder.
// This function works slow because it needs io.ReadAt seeking each time.
func (pm *SeekPosMapper) PosEncode(s string) (pos int64, err error) {
	pos, ok, err := readerAtBinSearch(pm.r, s, 0, pm.size)
	if err != nil {
		return
	}
	if !ok {
		err = &PosEncodeError{s: s}
		return
	}
	return
}

// readerAtBinsearch searches s from r. The left and right is the offset which
// seeks from and to. When s is found, ok will be true.
func readerAtBinSearch(r io.ReaderAt, s string, left, right int64) (pos int64, ok bool, err error) {
	var midS string
	for left+1 < right {
		midS, pos, err = searchMidPos(r, left, right)
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

// searchMidPos finds s and pos which exists r from left to right offsets.
func searchMidPos(r io.ReaderAt, left, right int64) (s string, pos int64, err error) {
	pos, err = findBeginOfLine(r, left+(right-left)/2)
	if err != nil && err != io.EOF {
		return
	}

	s, err = readLine(r, pos)
	if err != nil && err != io.EOF {
		return
	}

	if err == io.EOF {
		err = nil
	}
	return
}

// findBeginOfLine finds the beginning of line which the line contains pos.
// It returns head offset int64.
func findBeginOfLine(r io.ReaderAt, pos int64) (first int64, err error) {
	for i := pos; i >= 0; i-- {
		if i == 0 {
			first = 0
			break
		}

		buf := make([]byte, 1)
		if _, err = r.ReadAt(buf, i); err != nil {
			return
		}

		if i == pos && rune(buf[0]) == '\n' {
			continue
		}
		if rune(buf[0]) == '\n' {
			first = i + 1
			break
		}
	}

	return
}

// readLine reads a line from pos to '\n' ('\n' is not included).
func readLine(r io.ReaderAt, pos int64) (s string, err error) {
	sBuf := []byte{}

	var bufLen int64 = 32
	var i int64
	for i = 0; ; i++ {
		buf := make([]byte, bufLen)
		if _, err = r.ReadAt(buf, pos+bufLen*i); err != nil {
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

// PosDecode is the implementation of PosDecoder.
// This function works slow because it needs io.ReadAt seeking each time.
func (pm *SeekPosMapper) PosDecode(pos int64) (s string, err error) {
	if pos >= pm.size {
		err = &PosDecodeError{pos: pos}
		return
	}

	var bufLen int64 = 32
	line := []byte{}
	var i int64
	for ; ; i++ {
		buf := make([]byte, bufLen)
		if _, err = pm.r.ReadAt(buf, pos+bufLen*i); err != nil {
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
		err = &PosDecodeError{pos: pos}
		return
	}
	if len(line) == 0 || rune(line[0]) == '\n' {
		err = &PosDecodeError{pos: pos}
		return
	}

	s = strings.Split(string(line), "\n")[0]
	err = nil
	return
}

// CachedSeekPosMapper is the implementation of PosMapper.
// It seeks io.ReaderAt when methods are called, but caches the results.
// It's takes shorter time rather than SeekPosMapper.
type CachedSeekPosMapper struct {
	r     io.ReaderAt
	size  int64
	cache *posNode
}

// NewCachedSeekPosMapper returns a CachedSeekPosMapper. The size is total bytes of r.
// It seeks io.ReaderAt when methods are called, but caches the results.
// It's takes shorter time rather than SeekPosMapper.
func NewCachedSeekPosMapper(r io.ReaderAt, size int64) *CachedSeekPosMapper {
	return &CachedSeekPosMapper{
		r:     r,
		size:  size,
		cache: nil,
	}
}

// PosEncode is the implementation of PosEncoder.
// This method makes cache when it is called.
func (pm *CachedSeekPosMapper) PosEncode(s string) (pos int64, err error) {
	pos, left, right, ok := pm.cache.searchString(s, 0, pm.size)
	if ok {
		return
	}

	// binary search
	var midS string
	for left+1 < right {
		midS, pos, err = searchMidPos(pm.r, left, right)
		if err != nil {
			return
		}

		pm.cache = pm.cache.add(midS, pos)

		if s == midS {
			return
		} else if s < midS {
			right = left + (right-left)/2
		} else {
			left = left + (right-left)/2
		}
	}

	err = &PosEncodeError{s: s}
	return
}

// PosDecode is the implementation of PosDecoder.
// This method doesn't make cache even when it is called.
func (pm *CachedSeekPosMapper) PosDecode(pos int64) (s string, err error) {
	s, _, _, ok := pm.cache.searchPos(pos, 0, pm.size)
	if ok {
		return
	}

	spm := NewSeekPosMapper(pm.r, pm.size)
	s, err = spm.PosDecode(pos)
	if err != nil {
		return
	}

	// NOTE: Don't cache in order to make balanced tree.
	// pm.cache.add(s, pos)
	return
}

// PosNode is the node which consists a binary search tree.
type posNode struct {
	s           string
	pos         int64
	left, right *posNode
}

// add adds new node into pn. The caller must update pn by root.
func (pn *posNode) add(s string, pos int64) (root *posNode) {
	if pn == nil {
		root = &posNode{s: s, pos: pos}
		return
	}

	if s < pn.s {
		pn.left = pn.left.add(s, pos)
	} else if s > pn.s {
		pn.right = pn.right.add(s, pos)
	}
	return pn
}

// searchString searches s from the range of inLeft to inRight offsets.
func (pn *posNode) searchString(s string, inLeft, inRight int64) (pos, left, right int64, ok bool) {
	f := func(node *posNode) int {
		if s < node.s {
			return -1
		} else if s > node.s {
			return 1
		}
		return 0
	}

	var node *posNode
	node, left, right = pn.search(f, inLeft, inRight)
	if node != nil {
		pos = node.pos
		ok = true
	}
	return
}

// searchPos searches pos from the range of inLeft to inRight offsets.
func (pn *posNode) searchPos(pos int64, inLeft, inRight int64) (s string, left, right int64, ok bool) {
	f := func(node *posNode) int { return int(pos - node.pos) }

	var node *posNode
	node, left, right = pn.search(f, inLeft, inRight)
	if node != nil {
		s = node.s
		ok = true
	}
	return
}

// search searches the node which pred returns true. The node will be nil when not found.
func (pn *posNode) search(pred func(*posNode) int, inLeft, inRight int64) (node *posNode, left, right int64) {
	if pn == nil {
		left = inLeft
		right = inRight
		return
	}

	if pred(pn) < 0 {
		return pn.left.search(pred, inLeft, pn.pos)
	} else if pred(pn) > 0 {
		return pn.right.search(pred, pn.pos, inRight)
	}

	node = pn
	left = inLeft
	right = inRight
	return
}
