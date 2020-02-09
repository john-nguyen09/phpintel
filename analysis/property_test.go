package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/analysis/storage"
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
			e := storage.NewEncoder()
			property.Serialise(e)
			d := storage.NewDecoder(e.Bytes())
			deserialisedProperty := ReadProperty(d)
			jsonData, _ = json.MarshalIndent(deserialisedProperty, "", "  ")
			after := string(jsonData)
			if after != original {
				t.Errorf("%s != %s\n", original, after)
			}
		}
	}
}

func TestPropertyPhpDoc(t *testing.T) {
	testFile := "../cases/propertyDocs.php"
	data, err := ioutil.ReadFile(testFile)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", string(data))
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}
