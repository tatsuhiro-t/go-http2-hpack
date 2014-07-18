package hpack

import (
	"bytes"
	"fmt"
)

func huffmanEncodeSymbol(buffer *bytes.Buffer, rembits int, sym *HuffmanSymbol) int {
	nbits := sym.nbits

	for {
		if rembits > nbits {
			b := uint8(sym.code << uint(rembits-nbits))
			buffer.Bytes()[buffer.Len()-1] |= b

			rembits -= nbits

			break
		}

		b := uint8(sym.code >> uint(nbits-rembits))
		buffer.Bytes()[buffer.Len()-1] |= b

		nbits -= rembits
		rembits = 8

		if nbits == 0 {
			break
		}

		buffer.WriteByte(0)
	}

	return rembits
}

func HuffmanEncode(buffer *bytes.Buffer, str string) {
	rembits := 8

	for _, c := range str {
		sym := &HuffmanSymbolTable[c]

		if rembits == 8 {
			buffer.WriteByte(0)
		}

		rembits = huffmanEncodeSymbol(buffer, rembits, sym)
	}

	if rembits < 8 {
		sym := &HuffmanSymbolTable[256]

		b := uint8(sym.code >> uint(sym.nbits-rembits))
		buffer.Bytes()[buffer.Len()-1] |= b
	}
}

type HuffmanDecoder struct {
	state  uint8
	accept bool
}

func NewHuffmanDecoder() *HuffmanDecoder {
	ctx := &HuffmanDecoder{0, true}
	return ctx
}

const (
	HUFFMAN_DECODE_ACCEPT = 0x1
	HUFFMAN_DECODE_SYMBOL = 0x2
	HUFFMAN_DECODE_FAIL   = 0x4
)

func (decoder *HuffmanDecoder) Decode(buffer *bytes.Buffer, final bool) (string, error) {
	out := bytes.Buffer{}

	for _, c := range buffer.Bytes() {
		x := c >> 4
		for i := 0; i < 2; i++ {
			t := &HuffmanDecodeTable[decoder.state][x]

			if (t.flags & HUFFMAN_DECODE_FAIL) != 0 {
				return "", fmt.Errorf("Huffman decode error")
			}

			if (t.flags & HUFFMAN_DECODE_SYMBOL) != 0 {
				out.WriteByte(t.symbol)
			}

			decoder.state = t.state
			decoder.accept = (t.flags & HUFFMAN_DECODE_ACCEPT) != 0

			x = c & 0xf
		}
	}

	if final && !decoder.accept {
		return "", fmt.Errorf("Huffman decode ended prematurely")
	}

	return string(out.Bytes()), nil
}
