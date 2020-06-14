package analysis

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/stretchr/testify/assert"
)

func TestCompletionIndex(t *testing.T) {
	testMethodClassTest, _ := filepath.Abs("../cases/method.php")
	store := setupStore("test", "TestCompletionIndex")
	indexDocument(store, testMethodClassTest, "test1")

	classes, _ := store.SearchClasses("T", NewSearchOptions())
	cupaloy.SnapshotT(t, classes)
}

func TestMultipleIndexing(t *testing.T) {
	store := setupStore("test", t.Name())
	indexDocument(store, "../cases/method.php", "method")
	indexDocument(store, "../cases/method.php", "method")
	classes, _ := store.SearchClasses("TestMethodClass", NewSearchOptions())
	results := []string{}
	for _, c := range classes {
		results = append(results, c.Name.GetFQN())
	}
	assert.Equal(t, []string{
		"\\TestMethodClass",
		"\\TestAbstractMethodClass",
	}, results)
}

type indexable struct {
	collection string
	name       string
}

func (i indexable) GetIndexCollection() string {
	return i.collection
}

func (i indexable) GetIndexableName() string {
	return i.name
}

func TestFuzzyEngine(t *testing.T) {
	store := setupStore("test", t.Name())
	engine := newFuzzyEngine(nil)
	engine.index(store, "test1", indexable{"c1", "abc"}, "c1#abc")
	engine.index(store, "test1", indexable{"c1", "xyz"}, "c1#xyz")
	engine.index(store, "test2", indexable{"c1", "foobar"}, "c1#foobar")
	engine.index(store, "test3", indexable{"c1", "john_citizen"}, "c1#john_citizen")

	matches := []string{}
	engine.search(searchQuery{
		collection: "c1",
		keyword:    "abc",
		onData: func(data CompletionValue) onDataResult {
			matches = append(matches, string(data))
			return onDataResult{}
		},
	})
	assert.Equal(t, []string{"c1#abc"}, matches)

	deletor := newFuzzyEngineDeletor(engine, store, "test1")
	entriesToBeDeleted := []fuzzyEntry{}
	for _, entry := range deletor.entriesToBeDeleted {
		entriesToBeDeleted = append(entriesToBeDeleted, *entry)
	}
	sort.Slice(entriesToBeDeleted, func(i, j int) bool {
		return entriesToBeDeleted[i].key < entriesToBeDeleted[j].key
	})
	assert.Equal(t, []fuzzyEntry{
		{collection: "c1", name: "abc", key: "c1#abc", canonicalURI: "1", deleted: false},
		{collection: "c1", name: "xyz", key: "c1#xyz", canonicalURI: "1", deleted: false},
	}, entriesToBeDeleted)

	deletor.delete()
	matches = []string{}
	engine.index(store, "test1", indexable{"c1", "xyz"}, "c1#xyz")
	engine.search(searchQuery{
		collection: "c1",
		keyword:    "abc",
		onData: func(data CompletionValue) onDataResult {
			matches = append(matches, string(data))
			return onDataResult{}
		},
	})
	assert.Equal(t, []string{}, matches)

	e := storage.NewEncoder()
	engine.serialise(e)
	d := storage.NewDecoder(e.Bytes())
	newEngine := fuzzyEngineFromDecoder(d)

	matches = []string{}
	newEngine.search(searchQuery{
		collection: "c1",
		keyword:    "f",
		onData: func(data CompletionValue) onDataResult {
			matches = append(matches, string(data))
			return onDataResult{}
		},
	})
	assert.Equal(t, []string{"c1#foobar"}, matches)

	deletor = newFuzzyEngineDeletor(engine, store, "test2")
	entriesToBeDeleted = []fuzzyEntry{}
	for _, entry := range deletor.entriesToBeDeleted {
		entriesToBeDeleted = append(entriesToBeDeleted, *entry)
	}
	assert.Equal(t, []fuzzyEntry{
		{collection: "c1", name: "foobar", key: "c1#foobar", canonicalURI: "2", deleted: false},
	}, entriesToBeDeleted)
}
