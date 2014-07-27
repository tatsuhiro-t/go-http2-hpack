package hpack

import (
	"bytes"
)

// A encoder encodes header list to byte string using HPACK algorithm.
type Encoder struct {
	ht *headerTable
	// Maximum header table size this encoder supports.
	encoderMaxTableSize uint
	// Minimum header table size specified in ChangeTableSize().
	// This is necessary since we have to emit it in the next
	// context update.
	settingsMinTableSize uint
	// true if we have to emit context update in the next
	// encoding.
	contextUpdate bool
}

// NewEncoder returns new HPACK encoder.  encoderMaxTableSize
// specifies the maximum header table size this encoder supports.
func NewEncoder(encoderMaxTableSize uint) *Encoder {
	var contextUpdate bool
	var maxTableSize uint

	if encoderMaxTableSize < DEFAULT_HEADER_TABLE_SIZE {
		contextUpdate = true
		maxTableSize = encoderMaxTableSize
	} else {
		contextUpdate = false
		maxTableSize = DEFAULT_HEADER_TABLE_SIZE
	}

	encoder := &Encoder{
		newHeaderTable(maxTableSize),
		encoderMaxTableSize, uint32Max, contextUpdate,
	}

	return encoder
}

// Encode headers and write the output to dst.
func (enc *Encoder) Encode(dst *bytes.Buffer, headers []*Header) {
	if enc.contextUpdate {
		settingsMinTableSize := enc.settingsMinTableSize

		enc.contextUpdate = false
		enc.settingsMinTableSize = uint32Max

		if settingsMinTableSize < enc.ht.maxTableSize {
			encodeTableSize(dst, settingsMinTableSize)
		}

		encodeTableSize(dst, enc.ht.maxTableSize)
	}

	for _, header := range headers {
		enc.encodeHeader(dst, header)
	}
}

func (enc *Encoder) encodeHeader(dst *bytes.Buffer, header *Header) {
	idx, nameValueMatch := enc.ht.Search(header.Name, header.Value,
		header.NeverIndex)

	if nameValueMatch {
		encodeIndex(dst, idx)

		return
	}

	var indexing bool

	if enc.shouldIndexing(header) {
		indexing = true
		var entry *headerTableEntry
		if idx == -1 {
			entry = &headerTableEntry{header}
		} else {
			entry = &headerTableEntry{header}
		}
		enc.ht.PushFront(entry)
	} else {
		indexing = false
	}

	if idx == -1 {
		encodeNewname(dst, header.Name, header.Value,
			indexing, header.NeverIndex)
	} else {
		encodeIndname(dst, idx, header.Value,
			indexing, header.NeverIndex)
	}
}

func (enc *Encoder) shouldIndexing(header *Header) bool {
	return !ctstreq(header.Name, "set-cookie") &&
		!ctstreq(header.Name, "content-length") &&
		!ctstreq(header.Name, "location") &&
		!ctstreq(header.Name, "etag") &&
		!ctstreq(header.Name, ":path")
}

// Change maximum header table size to n.
func (enc *Encoder) ChangeTableSize(n uint) {
	if n > enc.encoderMaxTableSize {
		n = enc.encoderMaxTableSize
	}

	if n < enc.settingsMinTableSize {
		enc.settingsMinTableSize = n
	}

	enc.contextUpdate = true

	enc.ht.ChangeTableSize(n)
}

func encodeTableSize(dst *bytes.Buffer, tableSize uint) {
	head := dst.Len()

	encodeInteger(dst, uint64(tableSize), 5)

	dst.Bytes()[head] |= 0x20
}

func encodeIndex(dst *bytes.Buffer, idx int) {
	head := dst.Len()

	encodeInteger(dst, uint64(idx+1), 7)

	dst.Bytes()[head] |= 0x80
}

func encodeIndname(dst *bytes.Buffer, idx int, value string, indexing bool, neverIndexing bool) {
	head := dst.Len()

	var prefix uint
	if indexing {
		prefix = 6
	} else {
		prefix = 4
	}

	encodeInteger(dst, uint64(idx+1), prefix)

	dst.Bytes()[head] |= packFirstByte(indexing, neverIndexing)

	encodeString(dst, value)
}

func encodeNewname(dst *bytes.Buffer, name string, value string, indexing bool, neverIndexing bool) {
	dst.WriteByte(packFirstByte(indexing, neverIndexing))
	encodeString(dst, name)
	encodeString(dst, value)
}

func packFirstByte(indexing bool, neverIndexing bool) byte {
	if indexing {
		return 0x40
	}

	if neverIndexing {
		return 0x10
	}

	return 0
}

func encodeInteger(dst *bytes.Buffer, n uint64, prefix uint) {
	k := uint64((1 << prefix) - 1)

	if n < k {
		dst.WriteByte(byte(n))
		return
	}

	dst.WriteByte(byte(k))

	n -= k

	for {
		if n < 128 {
			dst.WriteByte(byte(n))
			break
		}

		dst.WriteByte(byte(0x80 | (n & 0x7f)))
		n >>= 7

		if n == 0 {
			break
		}
	}
}

func encodeString(dst *bytes.Buffer, src string) {
	huffmanLength := HuffmanEncodeLength(src)
	head := dst.Len()

	if huffmanLength < len(src) {
		encodeInteger(dst, uint64(huffmanLength), 7)
		HuffmanEncode(dst, src)
		dst.Bytes()[head] |= 0x80
	} else {
		encodeInteger(dst, uint64(len(src)), 7)
		dst.WriteString(src)
	}
}
