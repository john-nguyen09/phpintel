package analysis

import (
	"log"
	"strings"
	"time"

	"github.com/john-nguyen09/phpintel/analysis/filter"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
	cmap "github.com/orcaman/concurrent-map"
)

var filterCollection = "filter"

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
		ref: d.ReadString(),
		r: protocol.Range{
			Start: d.ReadPosition(),
			End:   d.ReadPosition(),
		},
	}
}

type referenceEntry struct {
	filter *filter.Filter
}

func newReferenceEntry() referenceEntry {
	return referenceEntry{
		filter: filter.NewFilter(),
	}
}

func (e referenceEntry) search(store *Store, uri string, ref string) []protocol.Range {
	var results []protocol.Range
	canonicalURI := util.CanonicaliseURI(store.uri, uri)
	dbEntry := newEntry(referenceIndexCollection, canonicalURI)
	if data, err := store.db.Get(dbEntry.getKeyBytes()); err == nil && len(data) != 0 {
		d := storage.NewDecoder(data)
		infoCount := d.ReadInt()
		for i := 0; i < infoCount; i++ {
			info := entryInfoDecode(d)
			if info.ref == ref {
				results = append(results, info.r)
			}
		}
	}
	return results
}

type referenceIndex struct {
	db      storage.DB
	entries cmap.ConcurrentMap
}

func newReferenceIndex(db storage.DB) *referenceIndex {
	index := &referenceIndex{
		db:      db,
		entries: cmap.New(),
	}
	if db != nil {
		dbEntry := newEntry(referenceIndexCollection, filterCollection+KeySep)
		start := time.Now()
		count := 0
		db.PrefixStream(dbEntry.getKeyBytes(), func(it storage.Iterator) {
			count++
			keyInfo := strings.Split(string(it.Key()), KeySep)
			d := storage.NewDecoder(it.Value())
			index.entries.Set(keyInfo[2], referenceEntry{
				filter: filter.FilterDecode(d),
			})
		})
		if count > 0 {
			log.Printf("Load reference index took %s", time.Since(start))
		}
	}
	return index
}

func (i *referenceIndex) index(store *Store, doc *Document, batch storage.Batch, infos []entryInfo) {
	canonicalURI := util.CanonicaliseURI(store.uri, doc.GetURI())
	i.entries.Upsert(canonicalURI, newReferenceEntry(), func(ok bool, curr interface{}, new interface{}) interface{} {
		var entry referenceEntry
		if ok {
			entry = curr.(referenceEntry)
		} else {
			entry = new.(referenceEntry)
		}
		dbEntry := newEntry(referenceIndexCollection, canonicalURI)
		dbEntry.e.WriteInt(len(infos))
		for _, info := range infos {
			entry.filter.Insert([]byte(info.ref))
			info.serialise(dbEntry.e)
		}
		writeEntry(batch, dbEntry)
		dbEntry = newEntry(referenceIndexCollection, filterCollection+KeySep+canonicalURI)
		err := entry.filter.Commit()
		if err != nil {
			panic(err)
		}
		entry.filter.Encode(dbEntry.e)
		writeEntry(batch, dbEntry)
		return entry
	})
}

func (i *referenceIndex) resetURI(store *Store, batch storage.Batch, uri string) {
	canonicalURI := util.CanonicaliseURI(store.uri, uri)
	i.entries.Upsert(canonicalURI, newReferenceEntry(), func(ok bool, curr interface{}, new interface{}) interface{} {
		var entry referenceEntry
		if ok {
			entry = curr.(referenceEntry)
			dbEntry := newEntry(referenceIndexCollection, canonicalURI)
			batch.Delete(dbEntry.getKeyBytes())
			dbEntry = newEntry(referenceIndexCollection, filterCollection+KeySep+canonicalURI)
			batch.Delete(dbEntry.getKeyBytes())
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
		ok, err := entry.filter.Lookup(refBytes)
		if err != nil {
			panic(err)
		}
		if ok {
			uri := util.URIFromCanonicalURI(store.uri, tuple.Key)
			for _, r := range entry.search(store, uri, ref) {
				results = append(results, protocol.Location{
					URI:   uri,
					Range: r,
				})
			}
		}
	}
	return results
}
