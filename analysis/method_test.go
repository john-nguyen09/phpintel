package analysis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/john-nguyen09/go-phpparser/parser"
	"github.com/john-nguyen09/phpintel/util"
)

func TestMethod(t *testing.T) {
	methodTest := "../cases/method.php"
	data, err := ioutil.ReadFile(methodTest)
	if err != nil {
		panic(err)
	}

	rootNode := parser.Parse(string(data))
	document := NewDocument(util.PathToUri(methodTest), []rune(string(data)), rootNode)
	jsonData, err := json.MarshalIndent(document, "", " ")

	if err != nil {
		panic(err)
	}

	fmt.Println(string(jsonData))
}
