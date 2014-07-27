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
