package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/analysis/storage"
)

func TestFunction(t *testing.T) {
	functionTest := "../cases/function.php"
	data, err := ioutil.ReadFile(functionTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", data)
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
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
