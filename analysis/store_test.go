package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/go-phpparser/parser"
	"github.com/john-nguyen09/phpintel/util"
)

func TestStore(t *testing.T) {
	classTest := "../cases/class.php"
	data, err := ioutil.ReadFile(classTest)
	if err != nil {
		panic(err)
	}

	rootNode := parser.Parse(string(data))
	document := NewDocument(util.PathToUri(classTest), string(data), rootNode)
	store, err := NewStore("./testData")
	defer store.Close()
	if err != nil {
		panic(err)
	}
	store.SyncDocument(document)
	classes := store.getClasses("TestClass1")
	cupaloy.Snapshot(classes)
}
