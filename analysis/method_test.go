package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestMethod(t *testing.T) {
	methodTest := "../cases/method.php"
	data, err := ioutil.ReadFile(methodTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", string(data))
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}

func TestMethodSerialiseAndDeserialise(t *testing.T) {
	methodTest := "../cases/method.php"
	data, err := ioutil.ReadFile(methodTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", string(data))
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

func TestMethodWithPhpDoc(t *testing.T) {
	testCase := "../cases/methodReturnPhpDoc.php"
	data, err := ioutil.ReadFile(testCase)
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", string(data))
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}

func TestMethodFromPhpDoc(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/methodFromPhpDoc.php")
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", string(data))
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}
