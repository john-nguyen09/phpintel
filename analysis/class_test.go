package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestClass(t *testing.T) {
	classTest := "../cases/class.php"
	data, err := ioutil.ReadFile(classTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", string(data))
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}

func TestClassSerialiseAndDeserialise(t *testing.T) {
	classTest := "../cases/class.php"
	data, err := ioutil.ReadFile(classTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", string(data))
	document.Load()
	for _, child := range document.Children {
		if theClass, ok := child.(*Class); ok {
			jsonData, _ := json.MarshalIndent(theClass, "", "  ")
			original := string(jsonData)
			serialiser := NewSerialiser()
			theClass.Serialise(serialiser)
			serialiser = SerialiserFromByteSlice(serialiser.GetBytes())
			deserialisedClass := ReadClass(serialiser)

			jsonData, _ = json.MarshalIndent(deserialisedClass, "", "  ")
			deserialise := string(jsonData)
			if original != deserialise {
				t.Errorf("%s != %s\n", original, deserialise)
			}
		}
	}
}
