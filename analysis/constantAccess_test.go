package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestConstantAccess(t *testing.T) {
	constantAccessTest := "../cases/constantAccess.php"
	data, err := ioutil.ReadFile(constantAccessTest)
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", data)
	document.Load()
	results := []Symbol{}
	tra := newTraverser()
	tra.traverseDocument(document, func(tra *traverser, s Symbol) {
		switch s.(type) {
		case *ConstantAccess:
			results = append(results, s)
		}
	})
	cupaloy.SnapshotT(t, results)
}
