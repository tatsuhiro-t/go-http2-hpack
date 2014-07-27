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

package hpack_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/tatsuhiro-t/go-http2-hpack"
	"log"
)

func ExampleDecoder() {
	headers := []*hpack.Header{
		hpack.NewHeader(":method", "GET", false),
		hpack.NewHeader(":scheme", "https", false),
		hpack.NewHeader(":authority", "example.org", false),
		hpack.NewHeader(":path", "/", false),
		hpack.NewHeader("user-agent", "nghttp2", false),
	}

	enc := hpack.NewEncoder(hpack.DEFAULT_HEADER_TABLE_SIZE)
	dec := hpack.NewDecoder()

	encoded := &bytes.Buffer{}

	enc.Encode(encoded, headers)

	pos := 0

	for {
		header, nread, err := dec.Decode(encoded.Bytes()[pos:], true)

		if err != nil {
			log.Print(err)
			break
		}

		pos += nread

		if header == nil {
			break
		}

		fmt.Printf("%s: %s\n", header.Name, header.Value)
	}

	// Output:
	// :method: GET
	// :scheme: https
	// :authority: example.org
	// :path: /
	// user-agent: nghttp2
}

func ExampleEncoder() {
	headers := []*hpack.Header{
		hpack.NewHeader(":method", "GET", false),
		hpack.NewHeader(":scheme", "https", false),
		hpack.NewHeader(":authority", "example.org", false),
		hpack.NewHeader(":path", "/", false),
		hpack.NewHeader("user-agent", "nghttp2", false),
	}

	enc := hpack.NewEncoder(hpack.DEFAULT_HEADER_TABLE_SIZE)

	encoded := &bytes.Buffer{}

	enc.Encode(encoded, headers)

	fmt.Println(hex.EncodeToString(encoded.Bytes()))

	// Output:
	// 828741882f91d35d055cf64d847a85aa69d29ac5
}
