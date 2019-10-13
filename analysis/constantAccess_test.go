package analysis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/john-nguyen09/phpintel/util"

	"github.com/john-nguyen09/go-phpparser/parser"
)

func TestConstantAccess(t *testing.T) {
	constantAccessTest := "../cases/constantAccess.php"
	data, err := ioutil.ReadFile(constantAccessTest)
	if err != nil {
		panic(err)
	}
	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(constantAccessTest), string(data), rootNode)
	jsonData, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonData))
}
