package analysis

import (
	"io/ioutil"

	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/stub"
	cmap "github.com/orcaman/concurrent-map"
)

func indexDocument(store *Store, filePath string, uri string) {
	data, _ := ioutil.ReadFile(filePath)
	document := NewDocument(uri, data)
	document.Load()
	store.SyncDocument(document)
}

func openDocument(store *Store, filePath string, uri string) *Document {
	data, _ := ioutil.ReadFile(filePath)
	document := NewDocument(uri, data)
	document.Open()
	document.Load()
	store.SyncDocument(document)
	return document
}

func setupStore(uri string, name string) *Store {
	db := storage.NewMemory()
	stubbers := stub.GetStubbers()
	return &Store{
		uri:       uri,
		db:        db,
		fEngine:   newFuzzyEngine(db),
		stubbers:  stubbers,
		documents: cmap.New(),

		syncedDocumentURIs: cmap.New(),
	}
}

func (s *Document) hasTypesSymbols() []HasTypes {
	results := []HasTypes{}
	t := newTraverser()
	t.traverseDocument(s, func(t *traverser, s Symbol) {
		if hasTypes, ok := s.(HasTypes); ok {
			results = append(results, hasTypes)
		}
	})
	return results
}
