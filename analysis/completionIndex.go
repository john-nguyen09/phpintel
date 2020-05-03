package analysis

import (
	"log"
	"sync"
	"time"

	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/sahilm/fuzzy"
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
	name    string
	key     string
	uri     string
	deleted bool
}

type fuzzyEngine struct {
	db         storage.DB
	indexMutex sync.RWMutex
	entries    map[string][]fuzzyEntry
	stopSignal chan struct{}

	currentCollection string
}

func newFuzzyEngine(db storage.DB) *fuzzyEngine {
	var engine *fuzzyEngine
	if db != nil {
		if b, err := db.Get([]byte(completionDataCollection)); err == nil && len(b) > 0 {
			log.Println("Loading fuzzy engine from DB")
			d := storage.NewDecoder(b)
			engine = fuzzyEngineFromDecoder(db, d)
		}
	}
	if engine == nil {
		engine = &fuzzyEngine{
			db:      db,
			entries: map[string][]fuzzyEntry{},
		}
	}
	go func() {
		ticker := time.NewTicker(compactionDuration)
		for {
			select {
			case <-ticker.C:
				engine.compact()
			case <-engine.stopSignal:
				break
			}
		}
	}()
	return engine
}

func fuzzyEngineFromDecoder(db storage.DB, d *storage.Decoder) *fuzzyEngine {
	entriesMap := map[string][]fuzzyEntry{}
	collectionLen := d.ReadInt()
	count := 0
	for i := 0; i < collectionLen; i++ {
		collection := d.ReadString()
		entriesLen := d.ReadInt()
		entries := []fuzzyEntry{}
		for j := 0; j < entriesLen; j++ {
			entries = append(entries, fuzzyEntry{
				name:    d.ReadString(),
				key:     d.ReadString(),
				uri:     d.ReadString(),
				deleted: false,
			})
		}
		entriesMap[collection] = entries
		count++
	}
	return &fuzzyEngine{
		db:      db,
		entries: entriesMap,
	}
}

func (f *fuzzyEngine) String(i int) string {
	return f.entries[f.currentCollection][i].name
}

func (f *fuzzyEngine) Len() int {
	return len(f.entries[f.currentCollection])
}

func (f *fuzzyEngine) serialise(e *storage.Encoder) {
	f.compact()
	f.indexMutex.Lock()
	defer f.indexMutex.Unlock()
	e.WriteInt(len(f.entries))
	for collection, entries := range f.entries {
		e.WriteString(collection)
		notDeleted := entries[:0]
		for _, entry := range entries {
			if !entry.deleted {
				notDeleted = append(notDeleted, entry)
			}
		}
		e.WriteInt(len(notDeleted))
		for _, entry := range notDeleted {
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
		currEntries []fuzzyEntry
		ok          bool
	)
	if currEntries, ok = f.entries[collection]; ok {
	}
	currEntries = append(currEntries, fuzzyEntry{
		name:    indexable.GetIndexableName(),
		key:     symbolKey,
		uri:     uri,
		deleted: false,
	})
	f.entries[collection] = currEntries
}

func (f *fuzzyEngine) search(query searchQuery) SearchResult {
	f.indexMutex.RLock()
	defer f.indexMutex.RUnlock()
	f.currentCollection = query.collection
	matches := fuzzy.FindFrom(query.keyword, f)
	searchResult := SearchResult{IsComplete: true}
	for _, match := range matches {
		if f.entries[f.currentCollection][match.Index].deleted {
			continue
		}
		result := query.onData(CompletionValue(f.entries[f.currentCollection][match.Index].key))
		if result.shouldStop {
			break
		}
	}
	return searchResult
}

func (f *fuzzyEngine) compact() {
	f.indexMutex.Lock()
	defer f.indexMutex.Unlock()
	for collection, entries := range f.entries {
		newEntries := entries[:0]
		for _, entry := range entries {
			if !entry.deleted {
				newEntries = append(newEntries, entry)
			}
		}
		f.entries[collection] = newEntries
	}
}

func (f *fuzzyEngine) close() error {
	f.stopSignal <- struct{}{}
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
	for collection, entries := range engine.entries {
		for i, entry := range entries {
			if entry.deleted {
				continue
			}
			if entry.uri == uri {
				entriesToBeDeleted = append(entriesToBeDeleted, &engine.entries[collection][i])
			}
		}
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
	}
}
