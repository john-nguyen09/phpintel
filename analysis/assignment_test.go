package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
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
	cupaloy.SnapshotT(t, document.Children)
}
