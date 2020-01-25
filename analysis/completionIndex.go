package analysis

import (
	"strings"

	"github.com/john-nguyen09/phpintel/analysis/wordtokeniser"
	"github.com/tecbot/gorocksdb"
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

func deleteCompletionIndex(db *gorocksdb.DB, batch *gorocksdb.WriteBatch, uri string) {
	entry := newEntry(documentCompletionIndex, uri)
	it := db.NewIterator(nil)
	defer it.Close()
	for it.Seek(entry.prefixRange()); it.ValidForPrefix(entry.prefixRange()); it.Next() {
		key := it.Key()
		value := it.Value()
		batch.Delete(key.Data())
		serialiser := SerialiserFromByteSlice(value.Data())
		len := serialiser.ReadInt()
		if len > 0 {
			for i := 0; i < len-1; i++ {
				key := serialiser.ReadBytes()
				batch.Delete(key)
			}
		}
		key.Free()
		value.Free()
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

func searchCompletions(db *gorocksdb.DB, query searchQuery) SearchResult {
	uniqueCompletionValues := make(map[CompletionValue]bool, 0)
	isComplete := true
	for _, prefix := range query.prefixes {
		name := strings.ToLower(query.keyword)
		prefix = strings.ToLower(prefix)
		if prefix != "" {
			name = prefix + scopeSep + name
		}
		entry := newEntry(query.collection, name)
		it := db.NewIterator(nil)
		for it.Seek(entry.prefixRange()); it.ValidForPrefix(entry.prefixRange()); it.Next() {
			value := it.Value()
			completionValue := readCompletionValue(SerialiserFromByteSlice(value.Data()))
			if _, ok := uniqueCompletionValues[completionValue]; ok {
				value.Free()
				continue
			}
			result := query.onData(completionValue)
			uniqueCompletionValues[completionValue] = true
			if result.shouldStop {
				isComplete = false
				value.Free()
				break
			}
			value.Free()
		}
		it.Close()
	}
	return SearchResult{isComplete}
}
