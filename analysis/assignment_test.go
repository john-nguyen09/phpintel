package analysis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/john-nguyen09/phpintel/util"

	"github.com/john-nguyen09/go-phpparser/parser"
)

func TestAssignment(t *testing.T) {
	assignmentTest := "../cases/variableAssignment.php"
	data, err := ioutil.ReadFile(assignmentTest)
	if err != nil {
		panic(err)
	}
	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(assignmentTest), string(data), rootNode)
	jsonData, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonData))
}
