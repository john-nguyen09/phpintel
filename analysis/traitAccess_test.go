package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestTraitAccess(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/Controller.php")
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", data)
	document.Load()
	symbol := document.HasTypesAtPos(document.positionAt(312))
	class := document.Children[0].(*Class)
	cupaloy.SnapshotT(t, toTestClass(class), symbol)
}
