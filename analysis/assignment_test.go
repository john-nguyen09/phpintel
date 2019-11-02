package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/util"
)

func TestAssignment(t *testing.T) {
	assignmentTest := "../cases/variableAssignment.php"
	data, err := ioutil.ReadFile(assignmentTest)
	if err != nil {
		panic(err)
	}
	document := NewDocument(util.PathToUri(assignmentTest), string(data))
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}
