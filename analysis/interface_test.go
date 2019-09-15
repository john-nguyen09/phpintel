package analysis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/john-nguyen09/go-phpparser/parser"
	"github.com/john-nguyen09/phpintel/util"
)

func TestInterface(t *testing.T) {
	interfaceTest := "../cases/interface.php"
	data, err := ioutil.ReadFile(interfaceTest)
	if err != nil {
		panic(err)
	}

	rootNode := parser.Parse(string(data))
	document := NewDocument(util.PathToUri(interfaceTest), []rune(string(data)), rootNode)

	jsonData, err := json.MarshalIndent(document, "", " ")
	if err != nil {
		panic(err)
	}

	fmt.Println(string(jsonData))
}
