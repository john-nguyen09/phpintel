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

func setupStore(uri string, name string) (*Store, error) {
	store, err := NewStore(uri, "./testData/"+name)
	if err != nil {
		return nil, err
	}
	it := store.db.NewIterator(nil, nil)
	for it.Valid() {
		store.db.Delete(it.Key(), nil)
	}
	it.Release()
	return store, nil
}
