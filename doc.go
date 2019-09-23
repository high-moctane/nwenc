/*
Package nwenc is the implementation of encoder and decoder for the nextword data.

The PosEncoder and PosDecoder can encode/decode between an int64 position value
and a string. There are several implementation for PosMapper. They have different
performance.

Encoder and Decoder can encode/decode bwtween an int64 position value and bytes.

PosEncoder/PosDecoder example:

	f, err := os.Open(filepath.Join("testdata", "words.txt"))
	pm, err := NewAllReadPosMapper(f)

	pm.PosEncode("a")   // 0 ("a" is at 0 byte in testdata/words.txt)
	pm.PosEncode("bcd") // 43 ("bcd" is at 43 bytes in testdata/words.txt)

	pm.PosDecode(0)  // "a"
	pm.PosDecode(43) // "bcd"

Encoder/Decoder example:

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
*/
package nwenc
