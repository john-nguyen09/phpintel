package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/analysis/storage"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/stretchr/testify/assert"
)

type testMethod struct {
	location protocol.Location

	Name               string
	Params             []*Parameter
	returnTypes        TypeComposite
	description        string
	Scope              TypeString
	VisibilityModifier VisibilityModifierValue
	IsStatic           bool
	ClassModifier      ClassModifierValue
}

func toTestMethod(m *Method) testMethod {
	return testMethod{
		location:           m.location,
		Name:               m.Name,
		Params:             m.Params,
		returnTypes:        m.returnTypes,
		description:        m.description,
		Scope:              m.Scope,
		VisibilityModifier: m.VisibilityModifier,
		IsStatic:           m.IsStatic,
		ClassModifier:      m.ClassModifier,
	}
}

func TestMethod(t *testing.T) {
	methodTest := "../cases/method.php"
	data, err := ioutil.ReadFile(methodTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", data)
	document.Load()
	results := []testMethod{}
	tra := newTraverser()
	tra.traverseDocument(document, func(tra *traverser, s Symbol) {
		if m, ok := s.(*Method); ok {
			results = append(results, toTestMethod(m))
		}
	})
	cupaloy.SnapshotT(t, results)
}

func TestMethodSerialiseAndDeserialise(t *testing.T) {
	methodTest := "../cases/method.php"
	data, err := ioutil.ReadFile(methodTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", data)
	document.Load()
	for _, child := range document.Children {
		if method, ok := child.(*Method); ok {
			jsonData, _ := json.MarshalIndent(method, "", "  ")
			original := string(jsonData)
			e := storage.NewEncoder()
			method.Serialise(e)
			d := storage.NewDecoder(e.Bytes())
			deserialisedMethod := ReadMethod(d)
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
	document := NewDocument("test1", data)
	document.Load()
	results := []testMethod{}
	tra := newTraverser()
	tra.traverseDocument(document, func(tra *traverser, s Symbol) {
		if m, ok := s.(*Method); ok {
			results = append(results, toTestMethod(m))
		}
	})
	cupaloy.SnapshotT(t, results)
}

func TestMethodFromPhpDoc(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/methodFromPhpDoc.php")
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", data)
	document.Load()
	results := []testMethod{}
	tra := newTraverser()
	tra.traverseDocument(document, func(tra *traverser, s Symbol) {
		if m, ok := s.(*Method); ok {
			results = append(results, toTestMethod(m))
		}
	})
	cupaloy.SnapshotT(t, results)
}

func TestReturnRelativeType(t *testing.T) {
	doc := NewDocument("test1", []byte(`<?php
class TestClass1 {
	/**
	 * @return static
	 */
	public function method1() {}

	/**
	 * @return TestClass2|$this
	 */
	public function method2() {}
}`))
	doc.Load()
	method1 := doc.Children[0].(*Class).getChildren()[1].(*Method)
	method2 := doc.Children[0].(*Class).getChildren()[4].(*Method)

	scopeTypes := newTypeComposite()
	scopeTypes.add(NewTypeString("\\TestClass1"))

	assert.Equal(t, "\\TestClass1", resolveMemberTypes(method1.GetReturnTypes(),
		&ClassAccess{
			Expression: Expression{
				Type: scopeTypes,
			},
		}).ToString())
	assert.Equal(t, "\\TestClass2|\\TestClass1", resolveMemberTypes(method2.GetReturnTypes(),
		&ClassAccess{
			Expression: Expression{
				Type: scopeTypes,
			},
		}).ToString())
}
