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
	"encoding/hex"
	"testing"
)

func TestHuffmanEncodeDecode(t *testing.T) {
	buffer := &bytes.Buffer{}
	expected := "Hello, World"
	HuffmanEncode(buffer, expected)

	decoder := NewHuffmanDecoder()

	output := &bytes.Buffer{}

	err := decoder.Decode(output, buffer.Bytes(), true)

	if err != nil {
		t.Errorf("decoder.Decode(%v) returns error %v",
			hex.EncodeToString(buffer.Bytes()), err)
	}

	actual := output.String()

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

	output := &bytes.Buffer{}

	err := decoder.Decode(output, buffer.Bytes(), true)

	if err != nil {
		t.Errorf("decoder.Decode(%v) returns error %v",
			hex.EncodeToString(buffer.Bytes()), err)
	}

	actual := output.String()

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

	output := &bytes.Buffer{}

	err := decoder.Decode(output, buffer.Bytes(), true)

	if err == nil {
		t.Errorf("decoder.Decode(%v) must return error",
			hex.EncodeToString(buffer.Bytes()))
	}

	expected := "Huffman decode ended prematurely"

	if err.Error() != expected {
		t.Errorf("error = %v, want %v", err.Error(), err)
	}
}
