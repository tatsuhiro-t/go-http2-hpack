package hpack

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func TestHuffmanEncodeDecode(t *testing.T) {
	buffer := &bytes.Buffer{}
	expected := "Hello, World"
	HuffmanEncode(buffer, expected)

	decoder := NewHuffmanDecoder()

	actual, err := decoder.Decode(buffer, true)

	if err != nil {
		t.Errorf("decoder.Decode(%v) returns error %v",
			hex.EncodeToString(buffer.Bytes()), err)
	}

	if expected != actual {
		t.Errorf("decoder.Decode(%v) = %v, want %v",
			hex.EncodeToString(buffer.Bytes()), expected, actual)
	}
}

func TestHuffmanBinary(t *testing.T) {
	buffer := &bytes.Buffer{}
	expected := string([]byte{0x0, 0x1, 0x2, 0x3})
	HuffmanEncode(buffer, expected)

	decoder := NewHuffmanDecoder()

	actual, err := decoder.Decode(buffer, true)

	if err != nil {
		t.Errorf("decoder.Decode(%v) returns error %v",
			hex.EncodeToString(buffer.Bytes()), err)
	}

	if expected != actual {
		t.Errorf("decoder.Decode(%v) = %v, want %v",
			hex.EncodeToString(buffer.Bytes()), expected, actual)
	}
}

func TestHuffmanDecodeEndsPrematury(t *testing.T) {
	buffer := &bytes.Buffer{}
	HuffmanEncode(buffer, "Hello, World")

	decoder := NewHuffmanDecoder()

	buffer.Truncate(buffer.Len() - 1)

	_, err := decoder.Decode(buffer, true)

	if err == nil {
		t.Errorf("decoder.Decode(%v) must return error",
			hex.EncodeToString(buffer.Bytes()))
	}

	expected := "Huffman decode ended prematurely"

	if err.Error() != expected {
		t.Errorf("error = %v, want %v", err.Error(), err)
	}
}
