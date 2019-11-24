package analysis

import (
	"github.com/john-nguyen09/phpintel/analysis/wordtokeniser"
	"github.com/syndtr/goleveldb/leveldb"
)

// CompletionValue holds references to uri and name
type CompletionValue string

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
	entry := newEntry(documentCompletionIndices, uri+KeySep+symbolKey)
	entry.serialiser.WriteInt(len(keys))
	for _, key := range keys {
		entry.serialiser.WriteBytes(key)
	}
	return entry
}

func deleteCompletionIndex(db *leveldb.DB, batch *leveldb.Batch, uri string) {
	entry := newEntry(documentCompletionIndices, uri)
	it := db.NewIterator(entry.prefixRange(), nil)
	defer it.Release()
	for it.Next() {
		batch.Delete(it.Key())
		serialiser := SerialiserFromByteSlice(it.Value())
		len := serialiser.ReadInt()
		if len > 0 {
			for i := 0; i < len-1; i++ {
				key := serialiser.ReadBytes()
				db.Delete(key, nil)
			}
		}
	}
}

func getCompletionKeys(uri string, indexable NameIndexable, symbolKey string) []string {
	tokens := wordtokeniser.Tokenise(indexable.GetIndexableName())
	keys := []string{}
	for _, token := range tokens {
		if indexable.GetPrefix() != "" {
			token = indexable.GetPrefix() + scopeSep + token
		}

		keys = append(keys, getCompletionKey(token, symbolKey))
	}
	return keys
}

func getCompletionKey(token string, symbolKey string) string {
	return token + KeySep + symbolKey
}

func searchCompletions(db *leveldb.DB, collection string, keyword string, prefix string) []CompletionValue {
	if prefix != "" {
		keyword = prefix + scopeSep + keyword
	}
	entry := newEntry(collection, keyword)
	it := db.NewIterator(entry.prefixRange(), nil)
	defer it.Release()
	completionValues := []CompletionValue{}
	uniqueCompletionValues := make(map[CompletionValue]bool, 0)
	for it.Next() {
		completionValue := readCompletionValue(SerialiserFromByteSlice(it.Value()))
		if _, ok := uniqueCompletionValues[completionValue]; ok {
			continue
		}
		completionValues = append(completionValues, completionValue)
		uniqueCompletionValues[completionValue] = true
	}
	return completionValues
}
