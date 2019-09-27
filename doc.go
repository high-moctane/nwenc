/*
Package nwenc is the implementation of encoder and decoder for the nextword data.

The OffsetEncoder and OffsetDecoder can encode/decode between an int64 offset value
and a string. There are several implementation for OffsetMapper. They have different
performance.

Encoder and Decoder can encode/decode between an int64 offset value and bytes.

OffsetEncoder/OffsetDecoder example:

	f, err := os.Open(filepath.Join("testdata", "words.txt"))
	om, err := NewAllReadOffsetMapper(f)

	om.OffsetEncode("a")   // 0 ("a" is at 0 byte in testdata/words.txt)
	om.OffsetEncode("bcd") // 43 ("bcd" is at 43 bytes in testdata/words.txt)

	om.OffsetDecode(0)  // "a"
	om.OffsetDecode(43) // "bcd"

Encoder/Decoder example:

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
*/
package nwenc
