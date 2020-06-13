package analysis

import (
	"log"
	"time"

	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	cmap "github.com/orcaman/concurrent-map"
	cuckoo "github.com/seiflotfy/cuckoofilter"
)

var refInterned map[string]string = make(map[string]string)

func refIntern(str string) string {
	if interned, ok := refInterned[str]; ok {
		return interned
	}
	refInterned[str] = str
	return str
}

type entryInfo struct {
	ref string
	r   protocol.Range
}

func (i entryInfo) serialise(e *storage.Encoder) {
	e.WriteString(i.ref)
	e.WritePosition(i.r.Start)
	e.WritePosition(i.r.End)
}

func entryInfoDecode(d *storage.Decoder) entryInfo {
	return entryInfo{
		ref: refIntern(d.ReadString()),
		r: protocol.Range{
			Start: d.ReadPosition(),
			End:   d.ReadPosition(),
		},
	}
}

type referenceEntry struct {
	filter *cuckoo.Filter
	data   []entryInfo
}

func newReferenceEntry() referenceEntry {
	return referenceEntry{
		filter: cuckoo.NewFilter(100),
	}
}

func referenceEntryDecode(d *storage.Decoder) referenceEntry {
	filter, err := cuckoo.Decode(d.ReadBytes())
	if err != nil {
		panic(err)
	}
	entry := referenceEntry{
		filter: filter,
	}
	count := d.ReadInt()
	for i := 0; i < count; i++ {
		entry.data = append(entry.data, entryInfoDecode(d))
	}
	return entry
}

func (e referenceEntry) search(ref string) []protocol.Range {
	var results []protocol.Range
	for _, info := range e.data {
		if info.ref == ref {
			results = append(results, info.r)
		}
	}
	return results
}

func (e referenceEntry) serialise(en *storage.Encoder) {
	en.WriteBytes(e.filter.Encode())
	en.WriteInt(len(e.data))
	for _, info := range e.data {
		info.serialise(en)
	}
}

type referenceIndex struct {
	db      storage.DB
	entries cmap.ConcurrentMap
}

func referenceIndexDecode(d *storage.Decoder) *referenceIndex {
	count := d.ReadInt()
	entries := cmap.New()
	for i := 0; i < count; i++ {
		key := d.ReadString()
		entries.Set(key, referenceEntryDecode(d))
	}
	return &referenceIndex{
		entries: entries,
	}
}

func newReferenceIndex(db storage.DB) *referenceIndex {
	var index *referenceIndex
	if db != nil {
		if b, err := db.Get([]byte(referenceIndexCollection)); err == nil && len(b) > 0 {
			start := time.Now()
			d := storage.NewDecoder(b)
			index = referenceIndexDecode(d)
			log.Printf("Loading reference index from DB took %s", time.Since(start))
		}
	}
	if index == nil {
		index = &referenceIndex{
			entries: cmap.New(),
		}
	}
	index.db = db
	return index
}

func (i *referenceIndex) insert(store *Store, location protocol.Location, ref string) {
	canonicalURI := util.CanonicaliseURI(store.uri, location.URI)
	i.entries.Upsert(canonicalURI, newReferenceEntry(), func(ok bool, curr interface{}, new interface{}) interface{} {
		var entry referenceEntry
		if ok {
			entry = curr.(referenceEntry)
		} else {
			entry = new.(referenceEntry)
		}
		entry.filter.Insert([]byte(ref))
		entry.data = append(entry.data, entryInfo{
			ref: refIntern(ref),
			r:   location.Range,
		})
		return entry
	})
}

func (i *referenceIndex) resetURI(store *Store, uri string) {
	canonicalURI := util.CanonicaliseURI(store.uri, uri)
	i.entries.Upsert(canonicalURI, newReferenceEntry(), func(ok bool, curr interface{}, new interface{}) interface{} {
		var entry referenceEntry
		if ok {
			entry = curr.(referenceEntry)
			entry.data = nil
			entry.filter.Reset()
		} else {
			entry = new.(referenceEntry)
		}
		return entry
	})
}

func (i *referenceIndex) search(store *Store, ref string) []protocol.Location {
	var results []protocol.Location
	refBytes := []byte(ref)
	for tuple := range i.entries.IterBuffered() {
		entry := tuple.Val.(referenceEntry)
		if entry.filter.Lookup(refBytes) {
			uri := util.URIFromCanonicalURI(store.uri, tuple.Key)
			for _, r := range entry.search(ref) {
				results = append(results, protocol.Location{
					URI:   uri,
					Range: r,
				})
			}
		}
	}
	return results
}

func (i *referenceIndex) close() error {
	e := storage.NewEncoder()
	e.WriteInt(i.entries.Count())
	for tuple := range i.entries.IterBuffered() {
		e.WriteString(tuple.Key)
		tuple.Val.(referenceEntry).serialise(e)
	}
	return i.db.Put([]byte(referenceIndexCollection), e.Bytes())
}
