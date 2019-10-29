package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/util"

	"github.com/john-nguyen09/go-phpparser/parser"
)

func TestFunction(t *testing.T) {
	functionTest := "../cases/function.php"
	data, err := ioutil.ReadFile(functionTest)
	if err != nil {
		panic(err)
	}

	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(functionTest), string(data), rootNode)
	cupaloy.SnapshotT(t, document.Children)
}

func TestFunctionSerialiseAndDeserialise(t *testing.T) {
	functionTest := "../cases/function.php"
	data, err := ioutil.ReadFile(functionTest)
	if err != nil {
		panic(err)
	}

	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(functionTest), string(data), rootNode)
	for _, child := range document.Children {
		if function, ok := child.(*Function); ok {
			jsonData, _ := json.MarshalIndent(function, "", "  ")
			original := string(jsonData)
			serialiser := NewSerialiser()
			function.Serialise(serialiser)
			serialiser = SerialiserFromByteSlice(serialiser.GetBytes())
			deserialisedFunction := ReadFunction(serialiser)
			jsonData, _ = json.MarshalIndent(deserialisedFunction, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		}
	}
}
