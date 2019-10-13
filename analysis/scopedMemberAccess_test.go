package analysis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/john-nguyen09/phpintel/util"

	"github.com/john-nguyen09/go-phpparser/parser"
)

func TestScopedMemberAccess(t *testing.T) {
	scopedPropertyAccessTest := "../cases/memberAccess.php"
	data, err := ioutil.ReadFile(scopedPropertyAccessTest)
	if err != nil {
		panic(err)
	}
	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(scopedPropertyAccessTest), string(data), rootNode)
	jsonData, err := json.MarshalIndent(document, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonData))
}
