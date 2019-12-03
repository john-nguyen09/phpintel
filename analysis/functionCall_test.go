package analysis

import (
	"encoding/json"
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
	document := NewDocument("test1", string(data))
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}

func TestFunctionCallSerialiseAndDeserialise(t *testing.T) {
	functionCallTest := "../cases/functionCall.php"
	data, err := ioutil.ReadFile(functionCallTest)
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", string(data))
	document.Load()
	for _, child := range document.Children {
		if functionCall, ok := child.(*FunctionCall); ok {
			jsonData, _ := json.MarshalIndent(functionCall, "", "  ")
			original := string(jsonData)
			serialiser := NewSerialiser()
			functionCall.Serialise(serialiser)
			serialiser = SerialiserFromByteSlice(serialiser.GetBytes())
			deserialisedFunctionCall := ReadFunctionCall(serialiser)
			jsonData, _ = json.MarshalIndent(deserialisedFunctionCall, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		}
	}
}
