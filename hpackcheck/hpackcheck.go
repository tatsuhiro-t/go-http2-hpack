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
			headerTableSize := val.(uint)
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
