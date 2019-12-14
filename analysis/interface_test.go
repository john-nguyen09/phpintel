package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
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
			serialiser := NewSerialiser()
			theInterface.Serialise(serialiser)
			serialiser = SerialiserFromByteSlice(serialiser.GetBytes())

			deserialisedInterface := ReadInterface(serialiser)
			jsonData, _ = json.MarshalIndent(deserialisedInterface, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		}
	}
}
