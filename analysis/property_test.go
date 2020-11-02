package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/stretchr/testify/assert"
)

func TestProperty(t *testing.T) {
	propertyTest := "../cases/property.php"
	data, err := ioutil.ReadFile(propertyTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", data)
	document.Load()
	results := []Symbol{}
	tra := newTraverser()
	tra.traverseDocument(document, func(tra *traverser, s Symbol, _ []Symbol) {
		if _, ok := s.(*Property); ok {
			results = append(results, s)
		}
	})
	cupaloy.SnapshotT(t, results)
}

func TestPropertySerialiseAndDeserialise(t *testing.T) {
	propertyTest := "../cases/property.php"
	data, err := ioutil.ReadFile(propertyTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", data)
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

	document := NewDocument("test1", data)
	document.Load()
	results := []Symbol{}
	tra := newTraverser()
	tra.traverseDocument(document, func(tra *traverser, s Symbol, _ []Symbol) {
		if _, ok := s.(*Property); ok {
			results = append(results, s)
		}
	})
	cupaloy.SnapshotT(t, results)
}

func TestPropertyWithTypes(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/propWithTypes.php")
	assert.NoError(t, err)
	doc := NewDocument("test1", data)
	doc.Load()
	results := []Symbol{}
	tra := newTraverser()
	tra.traverseDocument(doc, func(tra *traverser, s Symbol, _ []Symbol) {
		if _, ok := s.(*Property); ok {
			results = append(results, s)
		}
	})
	cupaloy.SnapshotT(t, results)
}

func TestAssignPropertyTypesInConstructor(t *testing.T) {
	doc := NewDocument("test1", []byte(`<?php
class TestAssignPropertyTypesInConstructorClass
{
	public $prop1;
	public function __construct(DatabaseInterface $db)
	{
		$this->prop1 = $db;
	}
}`))
	doc.Load()
	results := []string{}
	TraverseDocument(doc, func(s Symbol) {
		if prop, ok := s.(*Property); ok {
			results = append(results, prop.Types.ToString())
		}
	}, nil)
	assert.Equal(t, []string{
		"\\DatabaseInterface",
	}, results)
}
