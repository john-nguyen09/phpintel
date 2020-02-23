package analysis

import (
	"io/ioutil"
	"os"
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

func setupStore(uri string, name string) (*Store, error) {
	testDir := "./testData/"
	if _, err := os.Stat(testDir); os.IsNotExist(err) {
		os.Mkdir(testDir, os.ModePerm)
	}
	store, err := NewStore(uri, testDir+name)
	if err != nil {
		return nil, err
	}
	store.Clear()
	return store, nil
}
