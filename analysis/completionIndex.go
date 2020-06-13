package analysis

import (
	"log"
	"sort"
	"strings"
	"time"

	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/junegunn/fzf/src/algo"
	"github.com/junegunn/fzf/src/util"
	cmap "github.com/orcaman/concurrent-map"
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
	db storage.DB

	entries         cmap.ConcurrentMap
	entryURIIndex   cmap.ConcurrentMap
	reusableEntries cmap.ConcurrentMap

	currentCollection string
}

func newFuzzyEngine(db storage.DB) *fuzzyEngine {
	var engine *fuzzyEngine
	if db != nil {
		if b, err := db.Get([]byte(completionDataCollection)); err == nil && len(b) > 0 {
			start := time.Now()
			d := storage.NewDecoder(b)
			engine = fuzzyEngineFromDecoder(d)
			log.Printf("Loading fuzzy engine from DB took %s", time.Since(start))
		}
	}
	if engine == nil {
		engine = &fuzzyEngine{
			entries:         cmap.New(),
			entryURIIndex:   cmap.New(),
			reusableEntries: cmap.New(),
		}
	}
	engine.db = db
	return engine
}

func fuzzyEngineFromDecoder(d *storage.Decoder) *fuzzyEngine {
	entriesMap := cmap.New()
	entryURIIndex := cmap.New()
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
			entriesURIIndex := map[*fuzzyEntry]*fuzzyEntry{}
			if m, ok := entryURIIndex.Get(entry.uri); ok {
				entriesURIIndex = m.(map[*fuzzyEntry]*fuzzyEntry)
			}
			entriesURIIndex[entry] = entry
			entryURIIndex.Set(entry.uri, entriesURIIndex)
		}
		entriesMap.Set(collection, entries)
		count++
	}
	return &fuzzyEngine{
		entries:         entriesMap,
		entryURIIndex:   entryURIIndex,
		reusableEntries: cmap.New(),
	}
}

func (f *fuzzyEngine) serialise(e *storage.Encoder) {
	e.WriteInt(f.entries.Count())
	for tuple := range f.entries.IterBuffered() {
		e.WriteString(tuple.Key)
		entries := tuple.Val.([]*fuzzyEntry)
		var newEntries []*fuzzyEntry
		for _, entry := range entries {
			if !entry.deleted {
				newEntries = append(newEntries, entry)
			}
		}
		e.WriteInt(len(newEntries))
		for _, entry := range newEntries {
			e.WriteString(entry.name)
			e.WriteString(entry.key)
			e.WriteString(entry.uri)
		}
	}
}

func (f *fuzzyEngine) index(uri string, indexable NameIndexable, symbolKey string) {
	collection := indexable.GetIndexCollection()
	f.entryURIIndex.Upsert(uri, []*fuzzyEntry(nil), func(ok bool, curr interface{}, _ interface{}) interface{} {
		entryURIIndex := make(map[*fuzzyEntry]*fuzzyEntry)
		if ok {
			entryURIIndex = curr.(map[*fuzzyEntry]*fuzzyEntry)
		}
		f.entries.Upsert(collection, []*fuzzyEntry(nil), func(ok bool, curr interface{}, _ interface{}) interface{} {
			var (
				currEntries     []*fuzzyEntry
				reusableEntries []*fuzzyEntry
				entry           *fuzzyEntry
			)
			if ok {
				currEntries = curr.([]*fuzzyEntry)
			}
			f.reusableEntries.Upsert(collection, []*fuzzyEntry(nil), func(ok bool, curr interface{}, _ interface{}) interface{} {
				if ok {
					reusableEntries = curr.([]*fuzzyEntry)
				}
				if len(reusableEntries) > 0 {
					entry, reusableEntries = reusableEntries[len(reusableEntries)-1], reusableEntries[:len(reusableEntries)-1]
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
				entryURIIndex[entry] = entry
				return reusableEntries
			})
			return currEntries
		})
		return entryURIIndex
	})
}

type match struct {
	Key   string
	Score int
}

type byScore []match

func (m byScore) Len() int { return len(m) }

func (m byScore) Less(i, j int) bool { return m[i].Score > m[j].Score }

func (m byScore) Swap(i, j int) { m[i], m[j] = m[j], m[i] }

func isMatch(str string, pattern []rune) (bool, int) {
	chars := util.ToChars([]byte(str))
	result, _ := algo.FuzzyMatchV2(false, true, true, &chars, pattern, false, nil)
	return result.Score > 0, result.Score
}

func (f *fuzzyEngine) search(query searchQuery) SearchResult {
	f.currentCollection = query.collection
	searchResult := SearchResult{IsComplete: true}
	var entries []*fuzzyEntry
	if m, ok := f.entries.Get(f.currentCollection); ok {
		entries = m.([]*fuzzyEntry)
	}
	patternRune := []rune(strings.ToLower(query.keyword))
	var (
		matches []match
	)
	uniqueKey := map[string]struct{}{}
	for _, entry := range entries {
		if entry.deleted {
			continue
		}
		if _, ok := uniqueKey[entry.key]; ok {
			continue
		}
		var (
			matched bool
			score   int
		)
		if matched, score = isMatch(entry.name, patternRune); matched {
			matches = append(matches, match{entry.key, score})
			uniqueKey[entry.key] = struct{}{}
		}
	}
	sort.Sort(byScore(matches))
	for _, match := range matches {
		result := query.onData(CompletionValue(match.Key))
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

// SearchResult is the result of a search
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
	if m, ok := engine.entryURIIndex.Get(uri); ok {
		for entry := range m.(map[*fuzzyEntry]*fuzzyEntry) {
			entriesToBeDeleted = append(entriesToBeDeleted, entry)
		}
	}
	return &fuzzyEngineDeletor{
		uri:                uri,
		engine:             engine,
		entriesToBeDeleted: entriesToBeDeleted,
	}
}

func (d *fuzzyEngineDeletor) delete() {
	entryURIIndex := make(map[*fuzzyEntry]*fuzzyEntry)
	d.engine.entryURIIndex.Upsert(d.uri, entryURIIndex, func(ok bool, curr interface{}, _ interface{}) interface{} {
		if ok {
			entryURIIndex = curr.(map[*fuzzyEntry]*fuzzyEntry)
		}
		for _, entry := range d.entriesToBeDeleted {
			entry.deleted = true
			delete(entryURIIndex, entry)
			d.engine.reusableEntries.Upsert(entry.collection, []*fuzzyEntry(nil), func(ok bool, curr interface{}, _ interface{}) interface{} {
				var entries []*fuzzyEntry
				if ok {
					entries = curr.([]*fuzzyEntry)
				}
				entries = append(entries, entry)
				return entries
			})
		}
		return entryURIIndex
	})
}
