package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/util"
)

func TestStore(t *testing.T) {
	classTest := "../cases/class.php"
	data, err := ioutil.ReadFile(classTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument(util.PathToUri(classTest), string(data))
	document.Load()
	store, err := NewStore("./testData/TestStore")
	defer store.Close()
	if err != nil {
		panic(err)
	}
	store.SyncDocument(document)
	classes := store.GetClasses("TestClass1")
	cupaloy.Snapshot(classes)
}
