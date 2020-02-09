package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/analysis/storage"
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
			e := storage.NewEncoder()
			theClass.Serialise(e)
			d := storage.NewDecoder(e.Bytes())
			deserialisedClass := ReadClass(d)

			jsonData, _ = json.MarshalIndent(deserialisedClass, "", "  ")
			deserialise := string(jsonData)
			if original != deserialise {
				t.Errorf("%s != %s\n", original, deserialise)
			}
		}
	}
}

func TestClassDescription(t *testing.T) {
	testFile := "../cases/classDescription.php"
	data, err := ioutil.ReadFile(testFile)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", string(data))
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}
