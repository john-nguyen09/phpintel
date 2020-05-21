package analysis

import (
	"strings"

	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

func createReferenceEntry(store *Store, location protocol.Location, fqn string) []*entry {
	entries := []*entry{}
	canonicalURI := util.CanonicaliseURI(store.uri, location.URI)
	key := createReferenceIndexKey(canonicalURI, location.Range, fqn)
	entries = append(entries, newEntry(referenceIndexCollection, key))
	entries = append(entries, createEntryToReferReferenceIndex(canonicalURI, key))
	return entries
}

func createReferenceIndexKey(uri string, r protocol.Range, fqn string) string {
	sb := strings.Builder{}
	sb.WriteString(fqn)
	sb.WriteString(KeySep)
	sb.WriteString(uri)
	sb.WriteString(KeySep)
	sb.WriteString(r.String())
	return sb.String()
}

func createEntryToReferReferenceIndex(uri string, key string) *entry {
	sb := strings.Builder{}
	sb.WriteString(uri)
	sb.WriteString(KeySep)
	sb.WriteString(key)
	return newEntry(documentReferenceIndex, sb.String())
}

func searchReferences(store *Store, fqn string) []protocol.Location {
	results := []protocol.Location{}
	entry := newEntry(referenceIndexCollection, fqn)
	store.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		keyInfo := strings.Split(string(it.Key()), KeySep)
		uri := util.URIFromCanonicalURI(store.uri, keyInfo[2])
		r := util.RangeFromString(keyInfo[3])
		results = append(results, protocol.Location{
			URI:   uri,
			Range: r,
		})
	})
	return results
}

type referenceIndexDeletor struct {
	indexKeys map[string]bool
	docKeys   map[string]bool
}

func newReferenceIndexDeletor(store *Store, uri string) *referenceIndexDeletor {
	indexKeys := map[string]bool{}
	docKeys := map[string]bool{}
	canonicalURI := util.CanonicaliseURI(store.uri, uri)
	entry := newEntry(documentReferenceIndex, canonicalURI)
	store.db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		docKey := string(it.Key())
		docKeys[docKey] = true
		keyInfo := strings.Split(docKey, KeySep)
		indexKeys[strings.Join(keyInfo[2:], KeySep)] = true
	})
	return &referenceIndexDeletor{
		indexKeys: indexKeys,
		docKeys:   docKeys,
	}
}

func (d *referenceIndexDeletor) MarkNotDelete(store *Store, s Symbol, fqn string) {
	canonicalURI := util.CanonicaliseURI(store.uri, s.GetLocation().URI)
	key := createReferenceIndexKey(canonicalURI, s.GetLocation().Range, fqn)
	delete(d.indexKeys, key)
	sb := strings.Builder{}
	sb.WriteString(canonicalURI)
	sb.WriteString(KeySep)
	sb.WriteString(key)
	delete(d.docKeys, sb.String())
}

func (d *referenceIndexDeletor) Delete(b storage.Batch) {
	for indexKey := range d.indexKeys {
		entry := newEntry(referenceIndexCollection, indexKey)
		b.Delete(entry.getKeyBytes())
	}
	for docKey := range d.docKeys {
		b.Delete([]byte(docKey))
	}
}
