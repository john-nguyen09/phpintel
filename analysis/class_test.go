package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/go-phpparser/parser"
	"github.com/john-nguyen09/phpintel/util"
)

func TestClass(t *testing.T) {
	classTest := "../cases/class.php"
	data, err := ioutil.ReadFile(classTest)
	if err != nil {
		panic(err)
	}

	rootNode := parser.Parse(string(data))
	document := NewDocument(util.PathToUri(classTest), string(data), rootNode)
	cupaloy.SnapshotT(t, document.Children)
}

func TestClassSerialiseAndDeserialise(t *testing.T) {
	classTest := "../cases/class.php"
	data, err := ioutil.ReadFile(classTest)
	if err != nil {
		panic(err)
	}

	rootNode := parser.Parse(string(data))
	document := NewDocument(util.PathToUri(classTest), string(data), rootNode)
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
