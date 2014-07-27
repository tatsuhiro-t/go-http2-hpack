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

package main

import (
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tatsuhiro-t/go-http2-hpack"
	"os"
	"reflect"
)

type hpackTest struct {
	Draft       int    `json:"draft,omitempty"`
	Description string `json:"description,omitempty"`
	Cases       []map[string]interface{}
}

func runTest(test *hpackTest) error {
	decoder := hpack.NewDecoder()

	for seqno, testCase := range test.Cases {
		wire := testCase["wire"].(string)
		expectedHeaders := testCase["headers"].([]interface{})

		if val, ok := testCase["header_table_size"]; ok {
			headerTableSize := uint(val.(float64))
			decoder.ChangeTableSize(headerTableSize)
		}

		input, err := hex.DecodeString(wire)

		if err != nil {
			return fmt.Errorf(
				"seqno %d: wire is not encoded as hex string",
				seqno)
		}

		headers := []map[string]string{}

		for cur := 0; cur < len(input); {
			// Decode 1 byte at a time to check streaming
			// decoder.
			header, nread, err := decoder.Decode(input[cur:cur+1],
				cur+1 == len(input))

			if err != nil {
				return fmt.Errorf("seqno %d: decode failed %s",
					seqno, err)
			}

			if header != nil {
				headers = append(headers,
					map[string]string{
						header.Name: header.Value,
					})
			}

			cur += nread
		}

		if reflect.DeepEqual(expectedHeaders, headers) {
			fmt.Errorf("seqno %d: decoded = %v, want %v",
				seqno, headers, expectedHeaders)
		}
	}

	return nil
}

func main() {
	flag.Parse()

	for _, inpath := range flag.Args() {
		file, err := os.Open(inpath)

		if err != nil {
			panic(err)
		}

		dec := json.NewDecoder(file)

		var test hpackTest

		err = dec.Decode(&test)

		if err != nil {
			panic(err)
		}

		err = runTest(&test)

		if err != nil {
			fmt.Printf("%s: FAIL\n%s\n", inpath, err)
		} else {
			fmt.Printf("%s: SUCCESS\n", inpath)
		}
	}
}
