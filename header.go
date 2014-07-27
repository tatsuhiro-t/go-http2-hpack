// This package implements HTTP/2 HPACK encoder and decoder as defined
// in http://http2.github.io/http2-spec/compression.html
package hpack

type Header struct {
	// Header field name
	Name string
	// Header field value
	Value string
	// true if this header field must never be indexed.
	NeverIndex bool
}

// NewHeader returns new Header.
func NewHeader(name, value string, neverIndex bool) *Header {
	return &Header{name, value, neverIndex}
}

type headerTableEntry struct {
	header *Header
}

const (
	// Default header table size used for both encoder and
	// decoder.
	DEFAULT_HEADER_TABLE_SIZE = 4096
	headerEntryOverhead       = 32
)

const (
	uint32Max = uint(^uint32(0))
)

func ctstreq(a, b string) bool {
	if len(a) != len(b) {
		return false
	}

	c := byte(0)

	for i := 0; i < len(a); i++ {
		c |= a[i] ^ b[i]
	}

	return c == 0
}

func (ent *headerTableEntry) space() int {
	return len(ent.header.Name) + len(ent.header.Value) +
		headerEntryOverhead
}

type headerTable struct {
	table        []*headerTableEntry
	tablelen     int
	first        uint
	tableSize    uint
	maxTableSize uint
}

func newHeaderTable(maxTableSize uint) *headerTable {
	max := maxTableSize / headerEntryOverhead

	var entryNum uint
	for entryNum = 1; entryNum < max; entryNum <<= 1 {
	}

	table := make([]*headerTableEntry, entryNum)
	hdtable := &headerTable{table, 0, 0, 0, maxTableSize}

	return hdtable
}

func (ht *headerTable) ensureCapcity() {
	if ht.tablelen == len(ht.table) {
		newlen := len(ht.table) * 2
		newtable := make([]*headerTableEntry, newlen)

		for i := 0; i < ht.tablelen; i++ {
			newtable[i] = ht.dynget(i)
		}

		ht.table = newtable
		ht.first = 0
	}
}

func (ht *headerTable) evictFor(newEntry *headerTableEntry) {
	for ht.tablelen > 0 && ht.maxTableSize < ht.tableSize+uint(newEntry.space()) {
		ht.PopBack()
	}
}

func (ht *headerTable) ChangeTableSize(newSize uint) {
	ht.maxTableSize = newSize

	for ht.tablelen > 0 && ht.maxTableSize < ht.tableSize {
		ht.PopBack()
	}
}

func (ht *headerTable) PushFront(entry *headerTableEntry) {
	ht.evictFor(entry)
	ht.ensureCapcity()

	ht.first--

	ht.table[ht.first&uint(len(ht.table)-1)] = entry

	ht.tablelen++
	ht.tableSize += uint(entry.space())
}

func (ht *headerTable) PopBack() {
	entry := ht.dynmove(ht.tablelen - 1)

	ht.tableSize -= uint(entry.space())
	ht.tablelen--
}

func (ht *headerTable) dynget(idx int) *headerTableEntry {
	eidx := (ht.first + uint(idx)) & uint(len(ht.table)-1)

	return ht.table[eidx]
}

func (ht *headerTable) dynmove(idx int) *headerTableEntry {
	eidx := (ht.first + uint(idx)) & uint(len(ht.table)-1)
	entry := ht.table[eidx]
	// Assign nil so that so that entry will be garbage corrected.
	ht.table[eidx] = nil

	return entry
}

func (ht *headerTable) Get(idx int) *headerTableEntry {
	if idx >= staticTableLength() {
		return ht.dynget(idx - staticTableLength())
	}

	return &staticTable[idx]
}

func (ht *headerTable) Search(name string, value string, noNameValueMatch bool) (index int, nameValueMatch bool) {
	index = -1

	for idx, entry := range staticTable {
		if ctstreq(name, entry.header.Name) {
			if index == -1 {
				index = idx
			}

			if !noNameValueMatch &&
				ctstreq(value, entry.header.Value) {

				index = idx
				nameValueMatch = true
				return
			}
		}
	}

	if noNameValueMatch {
		return
	}

	for idx := 0; idx < ht.tablelen; idx++ {
		entry := ht.dynget(idx)

		if ctstreq(name, entry.header.Name) {
			if index == -1 {
				index = idx + staticTableLength()
			}

			if !noNameValueMatch &&
				ctstreq(value, entry.header.Value) {

				index = idx + staticTableLength()
				nameValueMatch = true
				return
			}
		}
	}
	return
}

func makeEntry(name, value string, nameHash, valueHash uint32) headerTableEntry {
	return headerTableEntry{&Header{name, value, false}}
}

var staticTable = []headerTableEntry{
	makeEntry(":authority", "", 2962729033, 0),
	makeEntry(":method", "GET", 3153018267, 70454),
	makeEntry(":method", "POST", 3153018267, 2461856),
	makeEntry(":path", "/", 56997727, 47),
	makeEntry(":path", "/index.html", 56997727, 2144181430),
	makeEntry(":scheme", "http", 3322585695, 3213448),
	makeEntry(":scheme", "https", 3322585695, 99617003),
	makeEntry(":status", "200", 3338091692, 49586),
	makeEntry(":status", "204", 3338091692, 49590),
	makeEntry(":status", "206", 3338091692, 49592),
	makeEntry(":status", "304", 3338091692, 50551),
	makeEntry(":status", "400", 3338091692, 51508),
	makeEntry(":status", "404", 3338091692, 51512),
	makeEntry(":status", "500", 3338091692, 52469),
	makeEntry("accept-charset", "", 124285319, 0),
	makeEntry("accept-encoding", "gzip, deflate", 4127597688, 1733326877),
	makeEntry("accept-language", "", 802785917, 0),
	makeEntry("accept-ranges", "", 1397189435, 0),
	makeEntry("accept", "", 2871506184, 0),
	makeEntry("access-control-allow-origin", "", 3297999203, 0),
	makeEntry("age", "", 96511, 0),
	makeEntry("allow", "", 92906313, 0),
	makeEntry("authorization", "", 2909397113, 0),
	makeEntry("cache-control", "", 4086191634, 0),
	makeEntry("content-disposition", "", 3027699811, 0),
	makeEntry("content-encoding", "", 2095084583, 0),
	makeEntry("content-language", "", 3065240108, 0),
	makeEntry("content-length", "", 3162187450, 0),
	makeEntry("content-location", "", 2284906121, 0),
	makeEntry("content-range", "", 2878374633, 0),
	makeEntry("content-type", "", 785670158, 0),
	makeEntry("cookie", "", 2940209764, 0),
	makeEntry("date", "", 3076014, 0),
	makeEntry("etag", "", 3123477, 0),
	makeEntry("expect", "", 3005803609, 0),
	makeEntry("expires", "", 2985731892, 0),
	makeEntry("from", "", 3151786, 0),
	makeEntry("host", "", 3208616, 0),
	makeEntry("if-match", "", 34533653, 0),
	makeEntry("if-modified-since", "", 2302095846, 0),
	makeEntry("if-none-match", "", 646073760, 0),
	makeEntry("if-range", "", 39145613, 0),
	makeEntry("if-unmodified-since", "", 1454068927, 0),
	makeEntry("last-modified", "", 150043680, 0),
	makeEntry("link", "", 3321850, 0),
	makeEntry("location", "", 1901043637, 0),
	makeEntry("max-forwards", "", 1619948695, 0),
	makeEntry("proxy-authenticate", "", 3993199572, 0),
	makeEntry("proxy-authorization", "", 329532250, 0),
	makeEntry("range", "", 108280125, 0),
	makeEntry("referer", "", 1085069613, 0),
	makeEntry("refresh", "", 1085444827, 0),
	makeEntry("retry-after", "", 1933352567, 0),
	makeEntry("server", "", 3389140803, 0),
	makeEntry("set-cookie", "", 1237214767, 0),
	makeEntry("strict-transport-security", "", 1153852136, 0),
	makeEntry("transfer-encoding", "", 1274458357, 0),
	makeEntry("user-agent", "", 486342275, 0),
	makeEntry("vary", "", 3612210, 0),
	makeEntry("via", "", 116750, 0),
	makeEntry("www-authenticate", "", 4051929931, 0),
}

func staticTableLength() int {
	return len(staticTable)
}
