package analysis

import "io/ioutil"

func indexDocument(store *Store, filePath string, uri string) {
	data, _ := ioutil.ReadFile(filePath)
	document := NewDocument(uri, string(data))
	document.Load()
	store.SyncDocument(document)
}

func openDocument(store *Store, filePath string, uri string) *Document {
	data, _ := ioutil.ReadFile(filePath)
	document := NewDocument(uri, string(data))
	document.Open()
	document.Load()
	store.SyncDocument(document)
	return document
}
