package analysis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/john-nguyen09/go-phpparser/parser"
	"github.com/john-nguyen09/phpintel/util"
)

func TestClass(t *testing.T) {
	classTest := "../cases/class.php"
	data, err := ioutil.ReadFile(classTest)
	if err != nil {
		panic(err)
	}

	rootNode := parser.Parse(string(data))
	document := NewDocument(util.PathToUri(classTest), []rune(string(data)), rootNode)
	jsonData, err := json.MarshalIndent(document.children, "", "  ")

	if err != nil {
		panic(err)
	}

	fmt.Println(string(jsonData))
}
