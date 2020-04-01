package analysis

import (
	"io/ioutil"

	"github.com/john-nguyen09/phpintel/analysis/storage"
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
	db := storage.NewMemOnly()
	initStubs()
	return &Store{
		uri:       uri,
		db:        db,
		documents: cmap.New(),

		syncedDocumentURIs: cmap.New(),
	}
}
