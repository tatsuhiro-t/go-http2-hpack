// go-http2-hpack - HTTP/2 HPACK implementation in golang
//
// Copyright (c) 2014 Tatsuhiro Tsujikawa
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package hpack

import (
	"bytes"
	"fmt"
)

// A Decoder decodes HPACK encoded byte string in streaming fashion.
type Decoder struct {
	ht *headerTable
	// Buffer to store header name (optional) and value, both
	// concatenated.
	nvbuf *bytes.Buffer
	hdec  *HuffmanDecoder
	// Pointer to header table entry for indexed name.
	entName *headerTableEntry
	// Opcode for HPACK encoding; initially opcodeNone.  Input
	// contains opcode one of opcodeIndexed, opcodeNewname and
	// opcodeIndname.
	opcode int
	// Current decoding state
	state int
	// Bytes left to read string
	left uint
	// Next shift to make when reading variable integer.
	shift uint
	// The length of decoded name.  This is required since we
	// store both name and value in one buffer nvbuf.
	newnamelen int
	// Maximum header table size set by ChangeTableSize().
	settingsMaxTableSize uint
	// true if string currently decoded is huffman-encoded.
	huffmanEncoded bool
	// true if encoder requires that the current name/value pair
	// must be index.
	indexRequired bool
	// true if encoder requires that the current name/value pair
	// must never be indexed.
	neverIndex bool
	// true if decoder encountered error.
	fail bool
}

const (
	opcodeNone = iota
	opcodeIndexed
	opcodeNewname
	opcodeIndname
)

const (
	stateOpcode = iota
	stateReadTableSize
	stateReadIndex
	stateCheckNamelen
	stateReadNamelen
	stateReadNamehuff
	stateReadName
	stateCheckValuelen
	stateReadValuelen
	stateReadValuehuff
	stateReadValue
)

// NewDecoder returns new HPACK decoder.
func NewDecoder() *Decoder {
	ht := newHeaderTable(DEFAULT_HEADER_TABLE_SIZE)
	nvbuf := &bytes.Buffer{}
	hdec := NewHuffmanDecoder()

	d := &Decoder{ht, nvbuf, hdec, nil,
		opcodeNone, stateOpcode, 0, 0, 0, DEFAULT_HEADER_TABLE_SIZE,
		false, false, false, false}
	return d
}

// Return the sum of header name and value length currently being
// decoded.
func (dec *Decoder) DecodingHeaderSize() int {
	return dec.nvbuf.Len()
}

// Decode src and emit header field.  The final signals the decoder
// that this is the end of complete compressed header block.  It helps
// decoder to check prematured end of encoded sequence.  This function
// returns header field if it is decoded and the number of bytes
// processed so far.  This function returns whenever one header field
// was decoded.  The caller must repeatedly call this function until
// whole input is processed, for example, by updating src slice.  Once
// this function returns error, further call of this function shall
// fail.
func (dec *Decoder) Decode(src []byte, final bool) (*Header, int, error) {
	cur := 0

	if dec.fail {
		return nil, cur, fmt.Errorf("could not process any input due to earlier error")
	}

	for cur < len(src) {
		switch dec.state {
		case stateOpcode:
			c := src[cur]
			switch {
			case (c & 0xe0) == 0x20:
				dec.opcode = opcodeIndexed
				dec.state = stateReadTableSize
			case (c & 0x80) != 0:
				dec.opcode = opcodeIndexed
				dec.state = stateReadIndex
			default:
				if c == 0x40 || c == 0 || c == 0x10 {
					dec.opcode = opcodeNewname
					dec.state = stateCheckNamelen
				} else {
					dec.opcode = opcodeIndname
					dec.state = stateReadIndex
				}
				dec.indexRequired = (c & 0x40) != 0
				dec.neverIndex = (c & 0xf0) == 0x10

				if dec.opcode == opcodeNewname {
					cur++
				}
			}

			dec.left = 0
			dec.shift = 0
		case stateReadTableSize:
			size, sizefin, shift, nread, err :=
				readInt(src[cur:], dec.left, dec.shift, 5)

			if err != nil {
				dec.fail = true
				return nil, cur, err
			}

			cur += nread
			dec.left = size
			dec.shift = shift

			if size > dec.settingsMaxTableSize {
				dec.fail = true
				return nil, cur, fmt.Errorf(
					"header table size is too large %v > %v",
					size, dec.settingsMaxTableSize)
			}

			if !sizefin {
				return nil, cur, dec.almostOK(final)
			}

			dec.ht.ChangeTableSize(size)

			dec.state = stateOpcode
		case stateReadIndex:
			var prefixlen uint

			switch {
			case dec.opcode == opcodeIndexed:
				prefixlen = 7
			case dec.indexRequired:
				prefixlen = 6
			default:
				prefixlen = 4
			}

			index, indexfin, shift, nread, err :=
				readInt(src[cur:], dec.left, dec.shift,
					prefixlen)

			if err != nil {
				dec.fail = true
				return nil, cur, err
			}

			cur += nread
			dec.left = index
			dec.shift = shift

			if index > uint(dec.maxIndex()+1) {
				dec.fail = true
				return nil, cur, fmt.Errorf(
					"index is too large %v > %v",
					index, dec.maxIndex()+1)
			}

			if !indexfin {
				return nil, cur, dec.almostOK(final)
			}

			if index == 0 {
				dec.fail = true
				return nil, cur, fmt.Errorf("illegal index = 0")
			}

			index--

			if dec.opcode == opcodeIndexed {
				header := dec.emitIndexed(int(index))

				dec.state = stateOpcode

				if header != nil {
					return header, cur, nil
				}
			} else {
				dec.entName = dec.ht.Get(int(index))
				dec.state = stateCheckValuelen
			}

		case stateCheckNamelen:
			dec.checkHuffmanEncoded(src[cur])
			dec.state = stateReadNamelen
			dec.left = 0
			dec.shift = 0
		case stateReadNamelen:
			length, lengthfin, shift, nread, err :=
				readInt(src[cur:], dec.left, dec.shift, 7)

			if err != nil {
				dec.fail = true
				return nil, cur, err
			}

			cur += nread
			dec.left = length
			dec.shift = shift

			if !lengthfin {
				return nil, cur, dec.almostOK(final)
			}

			if dec.huffmanEncoded {
				dec.hdec.Reset()
				dec.state = stateReadNamehuff
			} else {
				dec.state = stateReadName
			}
		case stateReadNamehuff:
			nread, err :=
				readHuffman(dec.hdec, dec.nvbuf,
					src[cur:], int(dec.left))

			if err != nil {
				dec.fail = true
				return nil, cur, err
			}

			cur += nread
			dec.left -= uint(nread)

			if dec.left > 0 {
				return nil, cur, nil
			}

			dec.newnamelen = dec.nvbuf.Len()
			dec.state = stateCheckValuelen

		case stateReadName:
			nread := readString(dec.nvbuf, src[cur:], int(dec.left))

			cur += nread
			dec.left -= uint(nread)

			if dec.left > 0 {
				return nil, cur, nil
			}

			dec.newnamelen = dec.nvbuf.Len()
			dec.state = stateCheckValuelen

		case stateCheckValuelen:
			dec.checkHuffmanEncoded(src[cur])
			dec.state = stateReadValuelen
			dec.left = 0
			dec.shift = 0
		case stateReadValuelen:
			length, lengthfin, shift, nread, err :=
				readInt(src[cur:], dec.left, dec.shift, 7)

			if err != nil {
				dec.fail = true
				return nil, cur, err
			}

			cur += nread
			dec.left = length
			dec.shift = shift

			if !lengthfin {
				return nil, cur, dec.almostOK(final)
			}

			var header *Header

			if dec.left == 0 {
				if dec.opcode == opcodeNewname {
					header = dec.emitNewname()
				} else {
					header = dec.emitIndname()
				}

				dec.state = stateOpcode
				return header, cur, nil
			}

			if dec.huffmanEncoded {
				dec.hdec.Reset()
				dec.state = stateReadValuehuff
			} else {
				dec.state = stateReadValue
			}
		case stateReadValuehuff:
			nread, err :=
				readHuffman(dec.hdec, dec.nvbuf,
					src[cur:], int(dec.left))

			if err != nil {
				dec.fail = true
				return nil, cur, err
			}

			cur += nread
			dec.left -= uint(nread)

			if dec.left > 0 {
				return nil, cur, dec.almostOK(final)
			}

			var header *Header

			if dec.opcode == opcodeNewname {
				header = dec.emitNewname()
			} else {
				header = dec.emitIndname()
			}

			dec.state = stateOpcode

			return header, cur, nil
		case stateReadValue:
			nread := readString(dec.nvbuf, src[cur:], int(dec.left))

			cur += nread
			dec.left -= uint(nread)

			if dec.left > 0 {
				return nil, cur, dec.almostOK(final)
			}

			var header *Header

			if dec.opcode == opcodeNewname {
				header = dec.emitNewname()
			} else {
				header = dec.emitIndname()
			}

			dec.state = stateOpcode

			return header, cur, nil
		}
	}

	return nil, cur, dec.almostOK(final)
}

// Decoding almost successful, but if final is true, we have to make
// sure that current decoding state is right one.
func (dec *Decoder) almostOK(final bool) error {
	if final && dec.state != stateOpcode {
		dec.fail = true
		return fmt.Errorf("input ended prematurely")
	}

	return nil
}

func (dec *Decoder) emitIndexed(index int) *Header {
	return dec.ht.Get(index).header
}

func (dec *Decoder) emitNewname() *Header {
	header := &Header{
		string(dec.nvbuf.Bytes()[:dec.newnamelen]),
		string(dec.nvbuf.Bytes()[dec.newnamelen:]),
		dec.neverIndex,
	}

	dec.nvbuf.Reset()

	if dec.indexRequired {
		entry := newHeaderTableEntry(header)
		dec.ht.PushFront(entry)
	}

	return header
}

func (dec *Decoder) emitIndname() *Header {
	header := &Header{
		dec.entName.header.Name,
		string(dec.nvbuf.Bytes()),
		dec.neverIndex,
	}

	dec.entName = nil
	dec.nvbuf.Reset()

	if dec.indexRequired {
		entry := newHeaderTableEntry(header)
		dec.ht.PushFront(entry)
	}

	return header
}

func (dec *Decoder) maxIndex() int {
	return dec.ht.tablelen + staticTableLength() - 1
}

func (dec *Decoder) checkHuffmanEncoded(b byte) {
	dec.huffmanEncoded = (b & (1 << 7)) != 0
}

// Change maximum header table size to n.
func (dec *Decoder) ChangeTableSize(n uint) {
	dec.settingsMaxTableSize = n
	dec.ht.ChangeTableSize(n)
}

// Read variable integer from src.  To support streaming decoding
// capability, we pass the initial value (the result of the previous
// call of this function) and shift to make.  This function returns
// the decoded integer, boolean final indicating that a integer is
// fully decoded, shift to make in the next call and number bytes read
// from src.
func readInt(src []byte, initial uint, initialShift, prefix uint) (n uint, final bool, shift uint, nread int, err error) {
	n = initial
	shift = initialShift
	k := byte((1 << prefix) - 1)

	if initial == 0 {
		c := src[0]

		nread++

		if (c & k) != k {
			n = uint(c & k)

			final = true
			return
		}

		n = uint(k)

		if nread == len(src) {
			return
		}
	}

	for nread < len(src) {
		c := src[nread]

		add := uint(c) & 0x7f

		if (uint32Max >> shift) < add {
			err = fmt.Errorf("overflow on shift: add %v, shift %v",
				add, shift)
			return
		}

		add <<= shift

		if uint32Max-add < n {
			err = fmt.Errorf("overflow on addition: add %v, n %v",
				add, n)
			return
		}

		n += add

		if (c & (1 << 7)) == 0 {
			break
		}

		nread++
		shift += 7
	}

	if nread == len(src) {
		return
	}

	final = true
	nread++

	return
}

// Read huffman-encoded string and decode and write it to dst.  The
// read from src is at most left bytes.  This function returns number
// of bytes read.
func readHuffman(hdec *HuffmanDecoder, dst *bytes.Buffer, src []byte, left int) (int, error) {
	var final bool

	if len(src) >= left {
		final = true
	} else {
		final = false
		left = len(src)
	}

	err := hdec.Decode(dst, src[:left], final)

	if err != nil {
		return left, err
	}

	return left, nil
}

// Read string from src at most left bytes and write it to dst.  This
// function returns number of bytes read.
func readString(dst *bytes.Buffer, src []byte, left int) int {
	if len(src) < left {
		left = len(src)
	}

	dst.Write(src[:left])

	return left
}
