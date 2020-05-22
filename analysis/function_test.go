package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

type testFunction struct {
	location protocol.Location

	Name        TypeString `json:"Name"`
	Params      []*Parameter
	returnTypes TypeComposite
	description string
}

func toTestFunction(function *Function) testFunction {
	return testFunction{
		location:    function.location,
		Name:        function.Name,
		Params:      function.Params,
		returnTypes: function.returnTypes,
		description: function.GetDescription(),
	}
}

func TestFunction(t *testing.T) {
	functionTest := "../cases/function.php"
	data, err := ioutil.ReadFile(functionTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", data)
	document.Load()
	results := []testFunction{}
	tra := newTraverser()
	tra.traverseDocument(document, func(tra *traverser, s Symbol) {
		if f, ok := s.(*Function); ok {
			results = append(results, toTestFunction(f))
		}
	})
	cupaloy.SnapshotT(t, results)
}

func TestFunctionSerialiseAndDeserialise(t *testing.T) {
	functionTest := "../cases/function.php"
	data, err := ioutil.ReadFile(functionTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", data)
	document.Load()
	for _, child := range document.Children {
		if function, ok := child.(*Function); ok {
			jsonData, _ := json.MarshalIndent(function, "", "  ")
			original := string(jsonData)
			e := storage.NewEncoder()
			function.Serialise(e)
			d := storage.NewDecoder(e.Bytes())
			deserialisedFunction := ReadFunction(d)
			jsonData, _ = json.MarshalIndent(deserialisedFunction, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		}
	}
}
