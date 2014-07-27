package hpack

import (
	"bytes"
	"fmt"
)

func huffmanEncodeSymbol(dst *bytes.Buffer, rembits int, sym *huffmanSymbol) int {
	nbits := sym.nbits

	for {
		if rembits > nbits {
			b := uint8(sym.code << uint(rembits-nbits))
			dst.Bytes()[dst.Len()-1] |= b

			rembits -= nbits

			break
		}

		b := uint8(sym.code >> uint(nbits-rembits))
		dst.Bytes()[dst.Len()-1] |= b

		nbits -= rembits
		rembits = 8

		if nbits == 0 {
			break
		}

		dst.WriteByte(0)
	}

	return rembits
}

// Huffman-encode str and write the output to dst.
func HuffmanEncode(dst *bytes.Buffer, str string) {
	rembits := 8

	for _, c := range str {
		sym := &huffmanSymbolTable[c]

		if rembits == 8 {
			dst.WriteByte(0)
		}

		rembits = huffmanEncodeSymbol(dst, rembits, sym)
	}

	if rembits < 8 {
		sym := &huffmanSymbolTable[256]

		b := uint8(sym.code >> uint(sym.nbits-rembits))
		dst.Bytes()[dst.Len()-1] |= b
	}
}

// Return the length of bytes when str is huffman-encoded.
func HuffmanEncodeLength(str string) int {
	n := 0
	for _, c := range str {
		n += huffmanSymbolTable[c].nbits
	}
	return (n + 7) / 8
}

// A HuffmanDecoder decodes huffman-encoded byte string in streaming
// fashion.
type HuffmanDecoder struct {
	state  uint8
	accept bool
}

// NewHuffmanDecoder returns new Huffman decoder.
func NewHuffmanDecoder() *HuffmanDecoder {
	ctx := &HuffmanDecoder{0, true}
	return ctx
}

const (
	huffmanDecodeAccept = 0x1
	huffmanDecodeSymbol = 0x2
	huffmanDecodeFail   = 0x4
)

// Reset decoder state so that it can decode new input.
func (decoder *HuffmanDecoder) Reset() {
	decoder.state = 0
	decoder.accept = true
}

// Decode src and write output to dst.  The final signals the end of
// input.
func (decoder *HuffmanDecoder) Decode(dst *bytes.Buffer, src []byte, final bool) error {
	for _, c := range src {
		x := c >> 4
		for i := 0; i < 2; i++ {
			t := &huffmanDecodeTable[decoder.state][x]

			if (t.flags & huffmanDecodeFail) != 0 {
				return fmt.Errorf("Huffman decode error")
			}

			if (t.flags & huffmanDecodeSymbol) != 0 {
				dst.WriteByte(t.symbol)
			}

			decoder.state = t.state
			decoder.accept = (t.flags & huffmanDecodeAccept) != 0

			x = c & 0xf
		}
	}

	if final && !decoder.accept {
		return fmt.Errorf("Huffman decode ended prematurely")
	}

	return nil
}
