package analysis

import (
	"log"
	"strings"
	"time"

	"github.com/john-nguyen09/phpintel/analysis/filter"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/analysis/wordtokeniser"
	cmap "github.com/orcaman/concurrent-map"
)

type completionIndex struct {
	db      storage.DB
	entries cmap.ConcurrentMap
}

type completionInfo struct {
	collection string
	word       string
}

type completionEntry struct {
	filter *filter.Filter
}

func newCompletionEntry() completionEntry {
	return completionEntry{
		filter: filter.NewFilter(),
	}
}

func (e *completionEntry) search(store *Store, uri string, query searchQuery, pattern []rune, uniqueSet map[string]void) {
	readDocumentSymbols(store.greb, newEntry(documentSymbolsCollection, uri), func(symbol documentSymbol) {
		indexableName := symbol.indexableName
		if len(indexableName) == 0 || symbol.collection != query.collection {
			return
		}
		if _, ok := uniqueSet[symbol.key]; ok {
			return
		}
		if ok, _ := isMatch(indexableName, pattern); ok {
			query.onData(CompletionValue(symbol.key))
			uniqueSet[symbol.key] = empty
		}
	})
}

func newCompletionIndex(db storage.DB) *completionIndex {
	index := &completionIndex{
		db:      db,
		entries: cmap.New(),
	}
	if db != nil {
		dbEntry := newEntry(completionIndexColletion, filterCollection+KeySep)
		start := time.Now()
		count := 0
		db.PrefixStream(dbEntry.getKeyBytes(), func(it storage.Iterator) {
			count++
			keyInfo := strings.Split(string(it.Key()), KeySep)
			d := storage.NewDecoder(it.Value())
			index.entries.Set(keyInfo[2], completionEntry{
				filter: filter.FilterDecode(d),
			})
		})
		if count > 0 {
			log.Printf("Load completion index took %s", time.Since(start))
		}
	}
	return index
}

func (i *completionIndex) index(store *Store, doc *Document, batch storage.Batch, infos []completionInfo) {
	uri := doc.GetURI()
	i.entries.Upsert(uri, newCompletionEntry(), func(ok bool, curr interface{}, new interface{}) interface{} {
		var entry completionEntry
		if ok {
			entry = curr.(completionEntry)
		} else {
			entry = new.(completionEntry)
		}
		for _, searchableToken := range getSearchableTokens(infos) {
			entry.filter.Insert([]byte(searchableToken))
		}
		dbEntry := newEntry(completionIndexColletion, filterCollection+KeySep+uri)
		err := entry.filter.Commit()
		if err != nil {
			panic(err)
		}
		entry.filter.Encode(dbEntry.e)
		writeEntry(batch, dbEntry)
		return entry
	})
}

func (i *completionIndex) resetURI(store *Store, batch storage.Batch, uri string) {
	i.entries.Upsert(uri, newCompletionEntry(), func(ok bool, curr interface{}, new interface{}) interface{} {
		var entry completionEntry
		if ok {
			entry = curr.(completionEntry)
			dbEntry := newEntry(referenceIndexCollection, filterCollection+KeySep+uri)
			batch.Delete(dbEntry.getKeyBytes())
		} else {
			entry = new.(completionEntry)
		}
		return entry
	})
}

func (i *completionIndex) search(store *Store, query searchQuery) SearchResult {
	isComplete := true
	words := wordtokeniser.Tokenise(query.keyword)
	shouldStop := false
	uniqueSet := map[string]void{}
	for _, word := range words {
		if shouldStop {
			break
		}
		wordBytes := []byte(query.collection + KeySep + word)
		pattern := []rune(strings.ToLower(word))
		for tuple := range i.entries.IterBuffered() {
			entry := tuple.Val.(completionEntry)
			ok, err := entry.filter.Lookup(wordBytes)
			if err != nil {
				panic(err)
			}
			if ok {
				entry.search(store, tuple.Key, query, pattern, uniqueSet)
			}
		}
	}
	return SearchResult{isComplete}
}

func getSearchableTokens(infos []completionInfo) []string {
	results := []string{}
	uniqueSet := map[string]void{}
	for _, info := range infos {
		hasCollection := len(info.collection) > 0
		for _, ngram := range extractStringToNgram(info.word, uniqueSet) {
			if hasCollection {
				results = append(results, info.collection+KeySep+ngram)
			} else {
				results = append(results, ngram)
			}
		}
	}
	return results
}

// Extract one string to ngram list
// Note the Ngram is a uint32 for ascii code
func extractStringToNgram(str string, uniqueSet map[string]void) []string {
	if len(str) == 0 {
		return nil
	}

	var results []string
	for i := range str {
		if i == 0 {
			continue
		}
		ngram := str[:i]
		if _, ok := uniqueSet[ngram]; ok {
			continue
		}
		results = append(results, ngram)
		uniqueSet[ngram] = empty
	}
	if _, ok := uniqueSet[str]; !ok {
		results = append(results, str)
		uniqueSet[str] = empty
	}

	return results
}
