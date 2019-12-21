package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestStore(t *testing.T) {
	classTest := "../cases/class.php"
	data, err := ioutil.ReadFile(classTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", string(data))
	document.Load()
	store, err := setupStore("test", "TestStore")
	defer store.Close()
	if err != nil {
		panic(err)
	}
	store.SyncDocument(document)
	classes := store.GetClasses("\\TestClass1")
	cupaloy.Snapshot(classes)
}
