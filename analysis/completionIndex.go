package analysis

import (
	"strings"

	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/analysis/wordtokeniser"
	"github.com/kezhuw/leveldb"
)

// CompletionValue holds references to uri and name
type CompletionValue string

type onDataResult struct {
	shouldStop bool
}

type searchQuery struct {
	collection string
	prefixes   []string
	keyword    string
	onData     func(CompletionValue) onDataResult
}

type SearchResult struct {
	IsComplete bool
}

// Serialise writes the CompletionValue
func (cv CompletionValue) Serialise(serialiser *Serialiser) {
	serialiser.WriteString(string(cv))
}

func readCompletionValue(serialiser *Serialiser) CompletionValue {
	return CompletionValue(serialiser.ReadString())
}

func createCompletionEntries(uri string, indexable NameIndexable, symbolKey string) []*entry {
	entries := []*entry{}
	keys := getCompletionKeys(uri, indexable, symbolKey)
	completionKeys := [][]byte{}
	for _, key := range keys {
		entry := newEntry(indexable.GetIndexCollection(), key)
		completionValue := CompletionValue(symbolKey)
		completionValue.Serialise(entry.serialiser)
		entries = append(entries, entry)

		completionKeys = append(completionKeys, entry.getKeyBytes())
	}
	entries = append(entries, createEntryToReferCompletionIndex(uri, symbolKey, completionKeys))
	return entries
}

func createEntryToReferCompletionIndex(uri string, symbolKey string, keys [][]byte) *entry {
	entry := newEntry(documentCompletionIndex, uri+KeySep+symbolKey)
	entry.serialiser.WriteInt(len(keys))
	for _, key := range keys {
		entry.serialiser.WriteBytes(key)
	}
	return entry
}

type completionIndexDeletor struct {
	indexKeys map[string]bool
	keys      map[string]bool
}

func newCompletionIndexDeletor(db *storage.Storage, uri string) *completionIndexDeletor {
	indexKeys := map[string]bool{}
	keys := map[string]bool{}
	entry := newEntry(documentCompletionIndex, uri)
	db.PrefixStream(entry.getKeyBytes(), func(it *storage.PrefixIterator) {
		keys[string(it.Key())] = true
		serialiser := SerialiserFromByteSlice(it.Value())
		len := serialiser.ReadInt()
		if len > 0 {
			for i := 0; i < len; i++ {
				key := serialiser.ReadBytes()
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

func (d *completionIndexDeletor) Delete(batch *leveldb.Batch){
	for indexKey := range d.indexKeys {
		batch.Delete([]byte(indexKey))
	}
	for key := range d.keys {
		batch.Delete([]byte(key))
	}
}

func getCompletionKeys(uri string, indexable NameIndexable, symbolKey string) []string {
	tokens := wordtokeniser.Tokenise(indexable.GetIndexableName())
	keys := []string{}
	for _, token := range tokens {
		token = strings.ToLower(token)
		for _, prefix := range indexable.GetPrefixes() {
			tokenWithPrefix := token
			if prefix != "" {
				tokenWithPrefix = strings.ToLower(prefix) + scopeSep + token
			}

			keys = append(keys, getCompletionKey(tokenWithPrefix, symbolKey))
		}
	}
	return keys
}

func getCompletionKey(token string, symbolKey string) string {
	return token + KeySep + symbolKey
}

func searchCompletions(db *storage.Storage, query searchQuery) SearchResult {
	uniqueCompletionValues := make(map[CompletionValue]bool, 0)
	isComplete := true
	for _, prefix := range query.prefixes {
		name := strings.ToLower(query.keyword)
		prefix = strings.ToLower(prefix)
		if prefix != "" {
			name = prefix + scopeSep + name
		}
		entry := newEntry(query.collection, name)
		db.PrefixStream(entry.getKeyBytes(), func(it *storage.PrefixIterator) {
			completionValue := readCompletionValue(SerialiserFromByteSlice(it.Value()))
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
	}
	return SearchResult{isComplete}
}
