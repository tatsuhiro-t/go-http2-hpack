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
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/tatsuhiro-t/go-http2-hpack"
	"os"
	"path"
)

func encode(test map[string]interface{}) (map[string]interface{}, error) {
	encoder := hpack.NewEncoder(hpack.DEFAULT_HEADER_TABLE_SIZE)
	buffer := &bytes.Buffer{}
	res := map[string]interface{}{
		"draft": 9,
		"description": "go-http2-hpack",
	}
	resCases := []map[string]interface{}{}

	testCases := test["cases"].([]interface{})

	for seqno, tc := range testCases {

		testCase := tc.(map[string]interface{})

		resCase := map[string]interface{}{"seqno": seqno}

		headers := []*hpack.Header{}

		for _, hm := range testCase["headers"].([]interface{}) {
			singleMap := hm.(map[string]interface{})
			for k, v := range singleMap {
				headers = append(headers,
					&hpack.Header{k, v.(string), false})
			}
		}

		encoder.Encode(buffer, headers)

		resCase["wire"] = hex.EncodeToString(buffer.Bytes())
		resCase["headers"] = testCase["headers"]

		buffer.Reset()

		resCases = append(resCases, resCase)
	}

	res["cases"] = resCases

	return res, nil
}

func main() {
	flag.Parse()

	for _, inpath := range flag.Args() {
		file, err := os.Open(inpath)

		if err != nil {
			panic(err)
		}

		dec := json.NewDecoder(file)

		var test map[string]interface{}

		err = dec.Decode(&test)

		if err != nil {
			panic(err)
		}

		res, err := encode(test)

		if err != nil {
			fmt.Printf("%s: FAIL\n%v\n", inpath, err)
		}

		outpath := path.Join("out", path.Base(inpath))

		outfile, err := os.Create(outpath)
		defer outfile.Close()

		if err != nil {
			panic(err)
		}

		b, err := json.MarshalIndent(res, "", "    ")

		if err != nil {
			fmt.Printf("%s: FAIL\n%s\n", inpath, err)
			continue
		}

		if _, err := outfile.Write(b); err != nil {
			fmt.Printf("%s: FAIL\n%s\n", inpath, err)
		} else {
			fmt.Printf("%s: SUCCESS\n", inpath)
		}
	}
}
