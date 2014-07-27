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
