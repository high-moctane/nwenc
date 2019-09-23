# nwenc

[![TravisCI](https://travis-ci.org/high-moctane/nwenc.svg?branch=master)](https://travis-ci.org/high-moctane/nwenc)
[![CodeCov](https://codecov.io/gh/high-moctane/nwenc/branch/master/graph/badge.svg)](https://codecov.io/gh/high-moctane/nwenc)
[![Go Report Card](https://goreportcard.com/badge/github.com/high-moctane/nwenc)](https://github.com/high-moctane/nwenc)
[![GoDoc](https://godoc.org/github.com/high-moctane/nwenc?status.svg)](https://godoc.org/github.com/high-moctane/nwenc)

Package nwenc is the implementation of encoder and decoder for the nextword data.

## PosEncoder/PosDecoder

The PosEncoder and PosDecoder can encode/decode between an int64 position value
and a string.
There are several implementation for PosMapper.
They have different performance.

PosEncoder/PosDecoder example:

```go
f, err := os.Open(filepath.Join("testdata", "words.txt"))
pm, err := NewAllReadPosMapper(f)

pm.PosEncode("a")   // 0 ("a" is at 0 byte in testdata/words.txt)
pm.PosEncode("bcd") // 43 ("bcd" is at 43 bytes in testdata/words.txt)

pm.PosDecode(0)  // "a"
pm.PosDecode(43) // "bcd"
```

## Encoder/Decoder

Encoder and Decoder can encode/decode bwtween an int64 position value and bytes.
Encoder/Decoder example:


```go
f, err := os.Open(filepath.Join("testdata", "words.txt"))
pm, err := NewAllReadPosMapper(f)

buf := new(bytes.Buffer)

byteLen := 3 // It encodes/decodes 3 bytes (24 bits) data.
enc, err := NewEncoder(byteLen)

enc.Encode(buf, 0)               // writes "a" at pos == 0
enc.EncodeString(buf, pm, "bcd") // writes "bcd"

dec, err := NewDecoder(byteLen)

dec.Decode(buf)           // 0 ("a")
dec.DecodeString(buf, pm) // "bcd"
```