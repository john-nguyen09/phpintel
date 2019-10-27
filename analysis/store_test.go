package analysis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/dgraph-io/badger"
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
	document := newDocument(util.PathToUri(classTest), string(data), rootNode)
	db, err := badger.Open(badger.DefaultOptions("./testData"))
	if err != nil {
		panic(err)
	}
	writeDocument(db, document)
	classes := getClass(db, "TestClass")
	jsonData, _ := json.MarshalIndent(classes, "", "  ")
	fmt.Println(string(jsonData))
}
