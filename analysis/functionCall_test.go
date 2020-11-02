package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestFunctionCall(t *testing.T) {
	functionCallTest := "../cases/functionCall.php"
	data, err := ioutil.ReadFile(functionCallTest)
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", data)
	document.Load()
	results := []Symbol{}
	tra := newTraverser()
	tra.traverseDocument(document, func(tra *traverser, s Symbol, _ []Symbol) {
		if _, ok := s.(*FunctionCall); ok {
			results = append(results, s)
		}
	})
	cupaloy.SnapshotT(t, results)
}
