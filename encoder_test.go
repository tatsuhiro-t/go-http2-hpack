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
	"encoding/json"
	"reflect"
	"testing"
)

func TestEncoderChangeTableSize(t *testing.T) {
	enc := NewEncoder(DEFAULT_HEADER_TABLE_SIZE)
	dec := NewDecoder()

	nva := []*Header{
		&Header{"alpha", "bravo", false},
	}

	encoded := &bytes.Buffer{}

	enc.Encode(encoded, nva)
	_, _, err := dec.Decode(encoded.Bytes(), true)

	if err != nil {
		t.Errorf("dec.Decode(...) returns error %v", err)
	}

	if dec.ht.tablelen != 1 {
		t.Errorf("dec.ht.tablelen = %v, want %v",
			dec.ht.tablelen, 1)
	}

	// Change table size twice in a row.  Changing to 0 clears up
	// header table entry.
	enc.ChangeTableSize(0)
	enc.ChangeTableSize(8192)

	dec.ChangeTableSize(0)
	dec.ChangeTableSize(8192)

	if enc.ht.tablelen != 0 {
		t.Errorf("enc.ht.tablelen = %v, want %v",
			enc.ht.tablelen, 0)
	}

	if dec.ht.tablelen != 0 {
		t.Errorf("dec.ht.tablelen = %v, want %v",
			dec.ht.tablelen, 0)
	}

	if enc.settingsMinTableSize != 0 {
		t.Errorf("enc.settingsMinTableSize = %v, want %v",
			enc.settingsMinTableSize, 0)
	}

	// We capped max header table size to 4096 in encoder side
	if enc.ht.maxTableSize != 4096 {
		t.Errorf("enc.ht.maxTableSize = %v, want %v",
			enc.ht.maxTableSize, 4096)
	}

	if dec.settingsMaxTableSize != 8192 {
		t.Errorf("dec.settingsMaxTableSize = %v, want %v",
			dec.settingsMaxTableSize, 8192)
	}

	if dec.ht.maxTableSize != 8192 {
		t.Errorf("dec.ht.maxTableSize = %v, want %v",
			dec.ht.maxTableSize, 8192)
	}

	encoded.Reset()

	// This will encode context update to notify selected table
	// size.
	enc.Encode(encoded, nva)
	_, _, err = dec.Decode(encoded.Bytes(), true)

	if err != nil {
		t.Errorf("dec.Decode(...) returns error %v", err)
	}

	if enc.settingsMinTableSize != uint32Max {
		t.Errorf("enc.settingsMinTableSize = %v, want %v",
			enc.settingsMinTableSize, uint32Max)
	}

	if dec.ht.maxTableSize != 4096 {
		t.Errorf("dec.ht.maxTableSize = %v, want %v",
			dec.ht.maxTableSize, 4096)
	}
}

func TestEncoderEncode(t *testing.T) {
	nva1 := []*Header{
		&Header{":status", "301", false},
		&Header{"date", "Sat, 03 Nov 2012 13:04:26 GMT", false},
		&Header{"server", "Server", false},
		&Header{"location", "http://www.amazon.com/", false},
		&Header{"content-length", "230", false},
		&Header{"keep-alive", "timeout=2, max=20", false},
		&Header{"connection", "Keep-Alive", false},
		&Header{"content-type", "text/html; charset=iso-8859-1", false},
	}

	nva2 := []*Header{
		&Header{":status", "200", false},
		&Header{"content-type", "image/png", false},
		&Header{"content-length", "6577", false},
		&Header{"connection", "keep-alive", false},
		&Header{"date", "Tue, 23 Oct 2012 17:58:47 GMT", false},
		&Header{"server", "Server", false},
		&Header{"cache-control", "max-age=630720000,public", false},
		&Header{"expires", "Wed, 18 May 2033 03:33:20 GMT", false},
		&Header{"last-modified", "Fri, 19 Oct 2012 23:59:58 GMT", false},
		&Header{"age", "932740", false},
		&Header{"x-amz-cf-id", "whiC_hNmBgrO48K-Fv1AqlFY-Cig61exld9QXg99v4RwPo9kzfqE9Q==", false},
		&Header{"via", "1.0 e0361d2450a4995d92d661bf6b825ede.cloudfront.net (CloudFront)", false},
		&Header{"x-cache", "Hit from cloudfront", false},
	}

	enc := NewEncoder(DEFAULT_HEADER_TABLE_SIZE)
	dec := NewDecoder()

	encodeDecode(t, enc, dec, nva1)
	encodeDecode(t, enc, dec, nva2)
}

func encodeDecode(t *testing.T, enc *Encoder, dec *Decoder, src []*Header) {
	encoded := &bytes.Buffer{}

	enc.Encode(encoded, src)

	decoded := []*Header{}
	cur := 0

	for {
		header, nread, err := dec.Decode(encoded.Bytes()[cur:], true)
		if err != nil {
			t.Errorf("dec.Decode(...) with cur = %v returns error %v",
				cur, err)
			return
		}

		cur += nread

		if header == nil {
			break
		}

		decoded = append(decoded, header)
	}

	if !reflect.DeepEqual(src, decoded) {
		a, _ := json.MarshalIndent(decoded, "", "    ")
		b, _ := json.MarshalIndent(src, "", "    ")
		t.Errorf("Decoded %v, want %v\n", string(a), string(b))
	}
}
