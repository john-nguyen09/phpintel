package analysis

import (
	"io/ioutil"

	"github.com/akrylysov/pogreb"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/stub"
	cmap "github.com/orcaman/concurrent-map"
)

func indexDocument(store *Store, filePath string, uri string) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
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

func withTestStore(uri string, name string, fn func(*Store)) {
	db := storage.NewMemory()
	stubbers := stub.GetStubbers()
	greb, err := pogreb.Open("testData/"+name, nil)
	if err != nil {
		panic(err)
	}
	store := &Store{
		uri:       uri,
		db:        db,
		greb:      greb,
		FS:        protocol.NewFileFS(),
		refIndex:  newReferenceIndex(db),
		comIndex:  newCompletionIndex(db),
		stubbers:  stubbers,
		documents: cmap.New(),

		syncedDocumentURIs: cmap.New(),
	}
	fn(store)
	store.Close()
}

func (s *Document) hasTypesSymbols() []HasTypes {
	results := []HasTypes{}
	t := newTraverser()
	t.traverseDocument(s, func(t *traverser, s Symbol, _ []Symbol) {
		if hasTypes, ok := s.(HasTypes); ok {
			results = append(results, hasTypes)
		}
	})
	return results
}
