package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestGlobalVariable(t *testing.T) {
	globalVariableTest := "../cases/globalVariable.php"
	data, _ := ioutil.ReadFile(globalVariableTest)
	document := NewDocument("test1", data)
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}
