HTTP/2 HPACK in golang
======================

This is HTTP/2 HPACK implementation in golang.  This is direct port
from `nghttp2 <https://nghttp2.org/>`_ 's HPACK C implementation.  For
HPACK specification, see
http://http2.github.io/http2-spec/compression.html

Currently this package implements HPACK draft-08 plus propsed changes
toward coming draft-09.
The changes from draft-08 are:

* The reference set was removed.
* The static header table in front of the dynamic header table.
* No copy was made when referening entry in static header table.

This is my first golang project. Any comments and patches are welcome.

Documentation
-------------

https://godoc.org/github.com/tatsuhiro-t/go-http2-hpack

Example
-------

.. code-block:: go

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
