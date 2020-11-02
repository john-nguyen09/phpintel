package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

type testClass struct {
	description string
	Location    protocol.Location
	Modifier    ClassModifierValue
	Name        TypeString
	Extends     TypeString
	Interfaces  []TypeString
	Use         []TypeString
}

func toTestClass(class *Class) testClass {
	return testClass{
		description: class.GetDescription(),
		Location:    class.Location,
		Modifier:    class.Modifier,
		Name:        class.Name,
		Extends:     class.Extends,
		Interfaces:  class.Interfaces,
		Use:         class.Use,
	}
}

func TestClass(t *testing.T) {
	classTest := "../cases/class.php"
	data, err := ioutil.ReadFile(classTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", data)
	document.Load()
	classes := []testClass{}
	tra := newTraverser()
	tra.traverseDocument(document, func(tra *traverser, s Symbol, _ []Symbol) {
		if class, ok := s.(*Class); ok {
			classes = append(classes, toTestClass(class))
		}
	})
	cupaloy.SnapshotT(t, classes)
}

func TestClassSerialiseAndDeserialise(t *testing.T) {
	classTest := "../cases/class.php"
	data, err := ioutil.ReadFile(classTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", data)
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

	document := NewDocument("test1", data)
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}
