package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestProperty(t *testing.T) {
	propertyTest := "../cases/property.php"
	data, err := ioutil.ReadFile(propertyTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", string(data))
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}

func TestPropertySerialiseAndDeserialise(t *testing.T) {
	propertyTest := "../cases/property.php"
	data, err := ioutil.ReadFile(propertyTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", string(data))
	document.Load()
	for _, child := range document.Children {
		if property, ok := child.(*Property); ok {
			jsonData, _ := json.MarshalIndent(property, "", "  ")
			original := string(jsonData)
			serialiser := NewSerialiser()
			property.Serialise(serialiser)
			serialiser = SerialiserFromByteSlice(serialiser.GetBytes())
			deserialisedProperty := ReadProperty(serialiser)
			jsonData, _ = json.MarshalIndent(deserialisedProperty, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		}
	}
}
