package analysis

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompletionIndex(t *testing.T) {
	testMethodClassTest, _ := filepath.Abs("../cases/method.php")
	withTestStore("test", "TestCompletionIndex", func(store *Store) {
		indexDocument(store, testMethodClassTest, "test1")

		classes, _ := store.SearchClasses("T", NewSearchOptions())
		results := []string{}
		for _, c := range classes {
			results = append(results, c.Name.GetFQN())
		}
		expected := []string{
			"\\TestMethodClass",
			"\\TestAbstractMethodClass",
		}
		sort.Strings(expected)
		sort.Strings(results)
		assert.Equal(t, expected, results)

		defDoc1 := NewDocument("test2", []byte(`<?php
namespace Krizalys\Onedrive;
class Onedrive {}
`))
		defDoc1.Load()
		store.SaveDocOnStore(defDoc1)
		store.SyncDocument(defDoc1)

		classes, _ = store.SearchClasses("Onedrive", NewSearchOptions())
		results = []string{}
		for _, c := range classes {
			results = append(results, c.Name.GetFQN())
		}
		assert.Equal(t, []string{
			"\\Krizalys\\Onedrive\\Onedrive",
		}, results)
	})
}

func TestMultipleIndexing(t *testing.T) {
	withTestStore("test", t.Name(), func(store *Store) {
		indexDocument(store, "../cases/method.php", "method")
		indexDocument(store, "../cases/method.php", "method")
		classes, _ := store.SearchClasses("TestMethodClass", NewSearchOptions())
		results := []string{}
		for _, c := range classes {
			results = append(results, c.Name.GetFQN())
		}
		expected := []string{
			"\\TestMethodClass",
			"\\TestAbstractMethodClass",
		}
		sort.Strings(expected)
		sort.Strings(results)
		assert.Equal(t, expected, results)
	})
}

// func TestNgram(t *testing.T) {
// 	infos := []completionInfo{
// 		{
// 			collection: "cla",
// 			word:       "Onedrive",
// 		},
// 	}
// 	fmt.Println(getSearchableTokens(infos))
// }
