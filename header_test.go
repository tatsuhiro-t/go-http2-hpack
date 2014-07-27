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
	"testing"
)

func TestHeaderTablePushAndPop(t *testing.T) {
	ht := newHeaderTable(128)

	// 5 + 6 + 32 = 43
	hd1 := &Header{":path", "/alpha", false}
	// 7 + 7 + 32 = 46
	hd2 := &Header{":method", "OPTIONS", false}
	// 10 + 11 + 32 = 53
	hd3 := &Header{":authority", "example.org", false}

	// total := 142

	ht.PushFront(newHeaderTableEntry(hd1))

	if ht.tableSize != 43 {
		t.Errorf("ht.tableSize = %v, want %v", ht.tableSize, 43)
	}

	ht.PushFront(newHeaderTableEntry(hd2))
	ht.PushFront(newHeaderTableEntry(hd3))

	if ht.tablelen != 2 {
		t.Errorf("ht.tablelen = %v, want %v", ht.tablelen, 2)
	}

	if ht.dynget(0).header != hd3 {
		t.Errorf("ht.Get(0).header = %v, want %v",
			ht.Get(0).header, hd3)
	}

	ht.PopBack()

	if ht.tablelen != 1 {
		t.Errorf("ht.tablelen = %v, want %v", ht.tablelen, 1)
	}

	if ht.tableSize != 53 {
		t.Errorf("ht.tableSize = %v, want %v", ht.tableSize, 53)
	}
}

func TestHeaderChangeTableSize(t *testing.T) {
	ht := newHeaderTable(128)

	// 5 + 6 + 32 = 43
	hd1 := &Header{":path", "/alpha", false}
	// 7 + 7 + 32 = 46
	hd2 := &Header{":method", "OPTIONS", false}

	ht.PushFront(newHeaderTableEntry(hd1))
	ht.PushFront(newHeaderTableEntry(hd2))

	ht.ChangeTableSize(50)

	if ht.tablelen != 1 {
		t.Errorf("ht.tablelen = %v, want %v", ht.tablelen, 1)
	}

	header := ht.dynget(0).header

	if header != hd2 {
		t.Errorf("ht.dynget(0) = %v, want %v", header, hd2)
	}
}

func TestHeaderSearch(t *testing.T) {
	ht := newHeaderTable(4096)

	hd1 := &Header{":path", "/alpha", false}
	hd2 := &Header{"bravo", "charlie", false}

	ht.PushFront(newHeaderTableEntry(hd1))
	ht.PushFront(newHeaderTableEntry(hd2))

	idx, nameValueMatch := ht.Search(":path", "/", false)

	if idx != 3 || !nameValueMatch {
		t.Errorf("(idx, nameValueMatch) = (%v, %v), want (%v, %v)",
			idx, nameValueMatch, 3, true)
	}

	idx, nameValueMatch = ht.Search(":path", "/", true)

	if idx != 3 || nameValueMatch {
		t.Errorf("(idx, nameValueMatch) = (%v, %v), want (%v, %v)",
			idx, nameValueMatch, 3, false)
	}

	idx, nameValueMatch = ht.Search(":path", "/delta", false)

	if idx != 3 || nameValueMatch {
		t.Errorf("(idx, nameValueMatch) = (%v, %v), want (%v, %v)",
			idx, nameValueMatch, 3, false)
	}

	idx, nameValueMatch = ht.Search("bravo", "charlie", false)

	if idx != staticTableLength() || !nameValueMatch {
		t.Errorf("(idx, nameValueMatch) = (%v, %v), want (%v, %v)",
			idx, nameValueMatch, staticTableLength(), true)
	}

	idx, nameValueMatch = ht.Search("bravo", "delta", false)

	if idx != staticTableLength() || nameValueMatch {
		t.Errorf("(idx, nameValueMatch) = (%v, %v), want (%v, %v)",
			idx, nameValueMatch, staticTableLength(), false)
	}

	idx, nameValueMatch = ht.Search("echo", "foxtrot", false)

	if idx != -1 || nameValueMatch {
		t.Errorf("(idx, nameValueMatch) = (%v, %v), want (%v, %v)",
			idx, nameValueMatch, -1, false)
	}
}
