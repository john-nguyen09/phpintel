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
	results := []Symbol{}
	tra := newTraverser()
	tra.traverseDocument(document, func(tra *traverser, s Symbol, _ []Symbol) {
		if _, ok := s.(*GlobalVariable); ok {
			results = append(results, s)
		}
	})
	cupaloy.SnapshotT(t, results)
}
