package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/analysis/storage"
)

func TestFunctionCall(t *testing.T) {
	functionCallTest := "../cases/functionCall.php"
	data, err := ioutil.ReadFile(functionCallTest)
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", data)
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}

func TestFunctionCallSerialiseAndDeserialise(t *testing.T) {
	functionCallTest := "../cases/functionCall.php"
	data, err := ioutil.ReadFile(functionCallTest)
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", data)
	document.Load()
	for _, child := range document.Children {
		if functionCall, ok := child.(*FunctionCall); ok {
			jsonData, _ := json.MarshalIndent(functionCall, "", "  ")
			original := string(jsonData)
			e := storage.NewEncoder()
			functionCall.Serialise(e)
			d := storage.NewDecoder(e.Bytes())
			deserialisedFunctionCall := ReadFunctionCall(d)
			jsonData, _ = json.MarshalIndent(deserialisedFunctionCall, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		}
	}
}
