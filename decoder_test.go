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
	"testing"
)

func TestDecoderDecodeIndex(t *testing.T) {
	dec := NewDecoder()

	input := &bytes.Buffer{}

	// Encode: 5. :path: /index.html
	encodeIndex(input, 5-1)

	header, nread, err := dec.Decode(input.Bytes(), true)

	if err != nil {
		t.Errorf("dec.Decode(...) returned error %v", err)
	}

	if nread != input.Len() {
		t.Error("dec.Decode(...) read %v, want %v",
			nread, input.Len())
	}

	expected := staticTable[5-1].header

	if expected != header {
		t.Errorf("dec.Decode(...) returned %v, want %v",
			header, expected)
	}
}

func TestDecoderDecodeIndname(t *testing.T) {
	dec := NewDecoder()

	input := &bytes.Buffer{}

	// Encode cache-control: private, with indexing
	encodeIndname(input, 24-1, "private", true, false)
	nread1 := input.Len()

	// Encode authorization: basic aGVsbG86d29ybGQ=", with never
	// indexing
	encodeIndname(input, 23-1, "basic aGVsbG86d29ybGQ=", false, true)
	nread2 := input.Len() - nread1

	expected1 := Header{"cache-control", "private", false}
	expected2 := Header{"authorization", "basic aGVsbG86d29ybGQ=", true}

	header, nread, err := dec.Decode(input.Bytes(), true)

	if err != nil {
		t.Errorf("dec.Decode(...) returned error %v", err)
	}

	if nread1 != nread {
		t.Errorf("dec.Decode(...) read %v, want %v", nread, nread1)
	}

	if expected1 != *header {
		t.Errorf("dec.Decode(...) returned %v, want %v",
			*header, expected1)
	}

	header, nread, err = dec.Decode(input.Bytes()[nread1:], true)

	if err != nil {
		t.Errorf("dec.Decode(...) returned error %v", err)
	}

	if nread2 != nread {
		t.Errorf("dec.Decode(...) read %v, want %v", nread, nread2)
	}

	if expected2 != *header {
		t.Errorf("dec.Decode(...) returned %v, want %v",
			*header, expected2)
	}
}

func TestDecoderDecodeNewname(t *testing.T) {
	dec := NewDecoder()

	input := &bytes.Buffer{}

	// Encode cache-control: private, with indexing
	encodeNewname(input, "cache-control", "private", true, false)
	nread1 := input.Len()

	// Encode authorization: basic aGVsbG86d29ybGQ=", with never
	// indexing
	encodeNewname(input, "authorization", "basic aGVsbG86d29ybGQ=",
		false, true)
	nread2 := input.Len() - nread1

	expected1 := Header{"cache-control", "private", false}
	expected2 := Header{"authorization", "basic aGVsbG86d29ybGQ=", true}

	header, nread, err := dec.Decode(input.Bytes(), true)

	if err != nil {
		t.Errorf("dec.Decode(...) returned error %v", err)
	}

	if nread1 != nread {
		t.Errorf("dec.Decode(...) read %v, want %v", nread, nread1)
	}

	if expected1 != *header {
		t.Errorf("dec.Decode(...) returned %v, want %v",
			*header, expected1)
	}

	header, nread, err = dec.Decode(input.Bytes()[nread1:], true)

	if err != nil {
		t.Errorf("dec.Decode(...) returned error %v", err)
	}

	if nread2 != nread {
		t.Errorf("dec.Decode(...) read %v, want %v", nread, nread2)
	}

	if expected2 != *header {
		t.Errorf("dec.Decode(...) returned %v, want %v",
			*header, expected2)
	}
}

func TestDecoderDecodeStringEndPrematurely(t *testing.T) {
	dec := NewDecoder()

	input := &bytes.Buffer{}

	// Encode authorization: basic aGVsbG86d29ybGQ=", with never
	// indexing
	encodeNewname(input, "authorization", "basic aGVsbG86d29ybGQ=",
		false, true)

	_, _, err := dec.Decode(input.Bytes()[:input.Len()-1], true)

	if err == nil {
		t.Errorf("dec.Decode(...) must return error")
	}

	// Further call of Decode shall fail
	_, _, err = dec.Decode([]byte{}, true)

	if err == nil {
		t.Errorf("dec.Decode(...) must return error")
	}
}

func TestDecoderHandleIllegalContextUpdate(t *testing.T) {
	dec := NewDecoder()
	enc := NewEncoder(DEFAULT_HEADER_TABLE_SIZE)

	dec.ChangeTableSize(1024)
	// Deliberately set large size so that next context update
	// sends illegal value > 1024.
	enc.ChangeTableSize(3000)

	encoded := &bytes.Buffer{}

	enc.Encode(encoded, []*Header{})

	if encoded.Len() == 0 {
		t.Errorf("enc.Encode(...) produced nothing")
	}

	_, _, err := dec.Decode(encoded.Bytes(), true)

	if err == nil {
		t.Errorf("enc.Encode(...) must return error")
	}
}

func TestReadIntOverflow(t *testing.T) {
	encoded := &bytes.Buffer{}
	prefix := uint(7)

	encodeInteger(encoded, uint64(uint32Max)+1, prefix)

	_, _, _, _, err := readInt(encoded.Bytes(), 0, 0, prefix)

	if err == nil {
		t.Errorf("readInt(...) must return overflow error")
	}
}
