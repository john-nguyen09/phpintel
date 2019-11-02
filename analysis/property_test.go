package analysis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/go-phpparser/parser"
	"github.com/john-nguyen09/phpintel/util"
)

func TestProperty(t *testing.T) {
	propertyTest := "../cases/property.php"
	data, err := ioutil.ReadFile(propertyTest)
	if err != nil {
		panic(err)
	}

	rootNode := parser.Parse(string(data))
	document := NewDocument(util.PathToUri(propertyTest), string(data), rootNode)
	cupaloy.SnapshotT(t, document.Children)
}

func TestPropertySerialiseAndDeserialise(t *testing.T) {
	propertyTest := "../cases/property.php"
	data, err := ioutil.ReadFile(propertyTest)
	if err != nil {
		panic(err)
	}

	rootNode := parser.Parse(string(data))
	document := NewDocument(util.PathToUri(propertyTest), string(data), rootNode)
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
			fmt.Println(after)
		}
	}
}
