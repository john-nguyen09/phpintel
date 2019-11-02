package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/util"
)

func TestMethod(t *testing.T) {
	methodTest := "../cases/method.php"
	data, err := ioutil.ReadFile(methodTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument(util.PathToUri(methodTest), string(data))
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}

func TestMethodSerialiseAndDeserialise(t *testing.T) {
	methodTest := "../cases/method.php"
	data, err := ioutil.ReadFile(methodTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument(util.PathToUri(methodTest), string(data))
	document.Load()
	for _, child := range document.Children {
		if method, ok := child.(*Method); ok {
			jsonData, _ := json.MarshalIndent(method, "", "  ")
			original := string(jsonData)
			serialiser := NewSerialiser()
			method.Serialise(serialiser)
			serialiser = SerialiserFromByteSlice(serialiser.GetBytes())
			deserialisedMethod := ReadMethod(serialiser)
			jsonData, _ = json.MarshalIndent(deserialisedMethod, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		}
	}
}
