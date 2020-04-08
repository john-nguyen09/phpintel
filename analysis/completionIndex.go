package analysis

import (
	"strings"

	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/analysis/wordtokeniser"
)

// CompletionValue holds references to uri and name
type CompletionValue string

type onDataResult struct {
	shouldStop bool
}

type searchQuery struct {
	collection string
	keyword    string
	onData     func(CompletionValue) onDataResult
}

type SearchResult struct {
	IsComplete bool
}

// Serialise writes the CompletionValue
func (cv CompletionValue) Serialise(e *storage.Encoder) {
	e.WriteString(string(cv))
}

func readCompletionValue(d *storage.Decoder) CompletionValue {
	return CompletionValue(d.ReadString())
}

func createCompletionEntries(uri string, indexable NameIndexable, symbolKey string) []*entry {
	entries := []*entry{}
	keys := getCompletionKeys(uri, indexable, symbolKey)
	completionKeys := [][]byte{}
	for _, key := range keys {
		entry := newEntry(indexable.GetIndexCollection(), key)
		completionValue := CompletionValue(symbolKey)
		completionValue.Serialise(entry.e)
		entries = append(entries, entry)

		completionKeys = append(completionKeys, entry.getKeyBytes())
	}
	entries = append(entries, createEntryToReferCompletionIndex(uri, symbolKey, completionKeys))
	return entries
}

func createEntryToReferCompletionIndex(uri string, symbolKey string, keys [][]byte) *entry {
	entry := newEntry(documentCompletionIndex, uri+KeySep+symbolKey)
	entry.e.WriteInt(len(keys))
	for _, key := range keys {
		entry.e.WriteBytes(key)
	}
	return entry
}

type completionIndexDeletor struct {
	indexKeys map[string]bool
	keys      map[string]bool
}

func newCompletionIndexDeletor(db storage.DB, uri string) *completionIndexDeletor {
	indexKeys := map[string]bool{}
	keys := map[string]bool{}
	entry := newEntry(documentCompletionIndex, uri)
	db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		keys[string(it.Key())] = true
		d := storage.NewDecoder(it.Value())
		len := d.ReadInt()
		if len > 0 {
			for i := 0; i < len; i++ {
				key := d.ReadBytes()
				indexKeys[string(key)] = true
			}
		}
	})
	return &completionIndexDeletor{indexKeys, keys}
}

func (d *completionIndexDeletor) MarkNotDelete(uri string, indexable NameIndexable, symbolKey string) {
	keys := getCompletionKeys(uri, indexable, symbolKey)
	for _, key := range keys {
		entry := newEntry(indexable.GetIndexCollection(), key)
		delete(d.indexKeys, string(entry.getKeyBytes()))
	}
	entry := newEntry(documentCompletionIndex, uri+KeySep+symbolKey)
	delete(d.keys, string(entry.getKeyBytes()))
}

func (d *completionIndexDeletor) Delete(b storage.Batch) {
	for indexKey := range d.indexKeys {
		b.Delete([]byte(indexKey))
	}
	for key := range d.keys {
		b.Delete([]byte(key))
	}
}

func getCompletionKeys(uri string, indexable NameIndexable, symbolKey string) []string {
	tokens := wordtokeniser.Tokenise(indexable.GetIndexableName())
	keys := []string{}
	for _, token := range tokens {
		token = strings.ToLower(token)
		keys = append(keys, getCompletionKey(token, symbolKey))
	}
	return keys
}

func getCompletionKey(token string, symbolKey string) string {
	return token + KeySep + symbolKey
}

func searchCompletions(db storage.DB, query searchQuery) SearchResult {
	uniqueCompletionValues := make(map[CompletionValue]bool, 0)
	isComplete := true
	name := strings.ToLower(query.keyword)
	entry := newEntry(query.collection, name)
	db.PrefixStream(entry.getKeyBytes(), func(it storage.Iterator) {
		completionValue := readCompletionValue(storage.NewDecoder(it.Value()))
		if _, ok := uniqueCompletionValues[completionValue]; ok {
			return
		}
		result := query.onData(completionValue)
		uniqueCompletionValues[completionValue] = true
		if result.shouldStop {
			isComplete = false
			it.Stop()
		}
	})
	return SearchResult{isComplete}
}
