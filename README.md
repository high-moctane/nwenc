# nwenc

[![TravisCI](https://travis-ci.org/high-moctane/nwenc.svg?branch=master)](https://travis-ci.org/high-moctane/nwenc)
[![CodeCov](https://codecov.io/gh/high-moctane/nwenc/branch/master/graph/badge.svg)](https://codecov.io/gh/high-moctane/nwenc)
[![Go Report Card](https://goreportcard.com/badge/github.com/high-moctane/nwenc)](https://github.com/high-moctane/nwenc)
[![GoDoc](https://godoc.org/github.com/high-moctane/nwenc?status.svg)](https://godoc.org/github.com/high-moctane/nwenc)

Package nwenc is the implementation of encoder and decoder for the nextword data.

## OffsetEncoder/OffsetDecoder

The OffsetEncoder and OffsetDecoder can encode/decode between an int64 offset value
and a string.
There are several implementation for OffsetMapper.
They have different performance.

OffsetEncoder/OffsetDecoder example:

```go
f, err := os.Open(filepath.Join("testdata", "words.txt"))
om, err := NewAllReadOffsetMapper(f)

om.OffsetEncode("a")   // 0 ("a" is at 0 byte in testdata/words.txt)
om.OffsetEncode("bcd") // 43 ("bcd" is at 43 bytes in testdata/words.txt)

om.OffsetDecode(0)  // "a"
om.OffsetDecode(43) // "bcd"
```

## Encoder/Decoder

Encoder and Decoder can encode/decode bwtween an int64 offset value and bytes.
Encoder/Decoder example:


```go
f, err := os.Open(filepath.Join("testdata", "words.txt"))
om, err := NewAllReadOffsetMapper(f)

buf := new(bytes.Buffer)

byteLen := 3 // It encodes/decodes 3 bytes (24 bits) data.
enc, err := NewEncoder(byteLen)

enc.Encode(buf, 0)               // writes "a" at offset == 0
enc.EncodeString(buf, om, "bcd") // writes "bcd"

dec, err := NewDecoder(byteLen)

dec.Decode(buf)           // 0 ("a")
dec.DecodeString(buf, om) // "bcd"
```