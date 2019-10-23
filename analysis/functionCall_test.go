package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/util"

	"github.com/john-nguyen09/go-phpparser/parser"
)

func TestFunctionCall(t *testing.T) {
	functionCallTest := "../cases/functionCall.php"
	data, err := ioutil.ReadFile(functionCallTest)
	if err != nil {
		panic(err)
	}
	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(functionCallTest), string(data), rootNode)
	cupaloy.SnapshotT(t, document.Children)
}

func TestFunctionCallSerialiseAndDeserialise(t *testing.T) {
	functionCallTest := "../cases/functionCall.php"
	data, err := ioutil.ReadFile(functionCallTest)
	if err != nil {
		panic(err)
	}
	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(functionCallTest), string(data), rootNode)
	for _, child := range document.Children {
		if functionCall, ok := child.(*FunctionCall); ok {
			jsonData, _ := json.MarshalIndent(functionCall, "", "  ")
			original := string(jsonData)
			bytes := functionCall.Serialise()
			deserialisedFunctionCall := DeserialiseFunctionCall(document, bytes)
			jsonData, _ = json.MarshalIndent(deserialisedFunctionCall, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		}
	}
}
