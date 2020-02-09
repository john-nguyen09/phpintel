package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/analysis/storage"
)

func TestInterface(t *testing.T) {
	interfaceTest := "../cases/interface.php"
	data, err := ioutil.ReadFile(interfaceTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", string(data))
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}

func TestInterfaceSerialiseAndDeserialise(t *testing.T) {
	interfaceTest := "../cases/interface.php"
	data, err := ioutil.ReadFile(interfaceTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", string(data))
	document.Load()
	for _, child := range document.Children {
		if theInterface, ok := child.(*Interface); ok {
			jsonData, _ := json.MarshalIndent(theInterface, "", "  ")
			original := string(jsonData)
			e := storage.NewEncoder()
			theInterface.Serialise(e)
			d := storage.NewDecoder(e.Bytes())

			deserialisedInterface := ReadInterface(d)
			jsonData, _ = json.MarshalIndent(deserialisedInterface, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		}
	}
}
