package analysis

import (
	"log"
	"strings"
	"sync"
	"time"

	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/junegunn/fzf/src/algo"
	"github.com/junegunn/fzf/src/util"
)

const compactionDuration = 15 * time.Second

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

type fuzzyEntry struct {
	collection string
	name       string
	key        string
	uri        string
	deleted    bool
}

type fuzzyEngine struct {
	db         storage.DB
	indexMutex sync.RWMutex
	entries    map[string][]*fuzzyEntry

	currentCollection string
	entryURIIndex     map[string][]*fuzzyEntry
	reusableEntries   map[string][]*fuzzyEntry
}

func newFuzzyEngine(db storage.DB) *fuzzyEngine {
	var engine *fuzzyEngine
	if db != nil {
		if b, err := db.Get([]byte(completionDataCollection)); err == nil && len(b) > 0 {
			log.Println("Loading fuzzy engine from DB")
			d := storage.NewDecoder(b)
			engine = fuzzyEngineFromDecoder(d)
		}
	}
	if engine == nil {
		engine = &fuzzyEngine{
			entries:         map[string][]*fuzzyEntry{},
			entryURIIndex:   map[string][]*fuzzyEntry{},
			reusableEntries: map[string][]*fuzzyEntry{},
		}
	}
	engine.db = db
	return engine
}

func fuzzyEngineFromDecoder(d *storage.Decoder) *fuzzyEngine {
	entriesMap := map[string][]*fuzzyEntry{}
	entryURIIndex := map[string][]*fuzzyEntry{}
	collectionLen := d.ReadInt()
	count := 0
	for i := 0; i < collectionLen; i++ {
		collection := d.ReadString()
		entriesLen := d.ReadInt()
		entries := []*fuzzyEntry{}
		for j := 0; j < entriesLen; j++ {
			entry := &fuzzyEntry{
				collection: collection,
				name:       d.ReadString(),
				key:        d.ReadString(),
				uri:        d.ReadString(),
				deleted:    false,
			}
			entries = append(entries, entry)
			var (
				entries []*fuzzyEntry
				ok      bool
			)
			if entries, ok = entryURIIndex[entry.uri]; ok {
			}
			entries = append(entries, entry)
			entryURIIndex[entry.uri] = entries
		}
		entriesMap[collection] = entries
		count++
	}
	return &fuzzyEngine{
		entries:         entriesMap,
		entryURIIndex:   entryURIIndex,
		reusableEntries: map[string][]*fuzzyEntry{},
	}
}

func (f *fuzzyEngine) String(i int) string {
	return f.entries[f.currentCollection][i].name
}

func (f *fuzzyEngine) Len() int {
	return len(f.entries[f.currentCollection])
}

func (f *fuzzyEngine) serialise(e *storage.Encoder) {
	f.indexMutex.Lock()
	defer f.indexMutex.Unlock()
	e.WriteInt(len(f.entries))
	for collection, entries := range f.entries {
		e.WriteString(collection)
		e.WriteInt(len(entries))
		for _, entry := range entries {
			e.WriteString(entry.name)
			e.WriteString(entry.key)
			e.WriteString(entry.uri)
		}
	}
}

func (f *fuzzyEngine) index(uri string, indexable NameIndexable, symbolKey string) {
	f.indexMutex.Lock()
	defer f.indexMutex.Unlock()
	collection := indexable.GetIndexCollection()
	var (
		currEntries []*fuzzyEntry
		ok          bool
		entry       *fuzzyEntry
	)
	if currEntries, ok = f.entries[collection]; ok {
	}
	if reusableEntries, ok := f.reusableEntries[collection]; ok && len(reusableEntries) > 0 {
		entry, reusableEntries = reusableEntries[len(reusableEntries)-1], reusableEntries[:len(reusableEntries)-1]
		f.reusableEntries[collection] = reusableEntries
		entry.collection = collection
		entry.name = indexable.GetIndexableName()
		entry.key = symbolKey
		entry.uri = uri
		entry.deleted = false
	} else {
		entry = &fuzzyEntry{
			collection: collection,
			name:       indexable.GetIndexableName(),
			key:        symbolKey,
			uri:        uri,
			deleted:    false,
		}
		currEntries = append(currEntries, entry)
	}
	f.entries[collection] = currEntries
	var (
		entries []*fuzzyEntry
	)
	if entries, ok = f.entryURIIndex[entry.uri]; ok {
	}
	entries = append(entries, entry)
	f.entryURIIndex[entry.uri] = entries
}

type match struct {
	Index int
}

func (f *fuzzyEngine) match(pattern string) []match {
	matches := []match{}
	dataLen := f.Len()
	patternRune := []rune(strings.ToLower(pattern))
	for i := 0; i < dataLen; i++ {
		chars := util.ToChars([]byte(f.String(i)))
		result, _ := algo.FuzzyMatchV2(false, true, true, &chars, patternRune, false, nil)
		if result.Score > 0 {
			matches = append(matches, match{i})
		}
	}
	return matches
}

func (f *fuzzyEngine) search(query searchQuery) SearchResult {
	f.indexMutex.RLock()
	defer f.indexMutex.RUnlock()
	f.currentCollection = query.collection
	matches := f.match(query.keyword)
	searchResult := SearchResult{IsComplete: true}
	for _, match := range matches {
		if f.entries[f.currentCollection][match.Index].deleted {
			continue
		}
		result := query.onData(CompletionValue(f.entries[f.currentCollection][match.Index].key))
		if result.shouldStop {
			searchResult.IsComplete = false
			break
		}
	}
	return searchResult
}

func (f *fuzzyEngine) close() error {
	e := storage.NewEncoder()
	f.serialise(e)
	return f.db.Put([]byte(completionDataCollection), e.Bytes())
}

type SearchResult struct {
	IsComplete bool
}

type fuzzyEngineDeletor struct {
	uri                string
	engine             *fuzzyEngine
	entriesToBeDeleted []*fuzzyEntry
}

func newFuzzyEngineDeletor(engine *fuzzyEngine, uri string) *fuzzyEngineDeletor {
	entriesToBeDeleted := []*fuzzyEntry{}
	engine.indexMutex.RLock()
	if entries, ok := engine.entryURIIndex[uri]; ok {
		entriesToBeDeleted = entries
	}
	engine.indexMutex.RUnlock()
	return &fuzzyEngineDeletor{
		uri:                uri,
		engine:             engine,
		entriesToBeDeleted: entriesToBeDeleted,
	}
}

func (d *fuzzyEngineDeletor) delete() {
	for _, entry := range d.entriesToBeDeleted {
		entry.deleted = true
		var (
			currentReusableEntries []*fuzzyEntry
			ok                     bool
		)
		if currentReusableEntries, ok = d.engine.reusableEntries[entry.collection]; ok {
		}
		currentReusableEntries = append(currentReusableEntries, entry)
		d.engine.reusableEntries[entry.collection] = currentReusableEntries
	}
}
