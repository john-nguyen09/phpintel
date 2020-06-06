package analysis

import (
	"testing"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/stretchr/testify/assert"
)

func TestReferenceIndex(t *testing.T) {
	store := setupStore("test", "TestReferenceIndex")
	indexDocument(store, "../cases/function.php", "test1")
	doc2 := openDocument(store, "../cases/reference/functionCall.php", "test2")
	pos := protocol.Position{
		Line:      2,
		Character: 5,
	}
	sym := doc2.HasTypesAtPos(pos)
	name := NewTypeString(sym.(*FunctionCall).Name)
	fqn := doc2.ImportTableAtPos(pos).GetFunctionReferenceFQN(NewQuery(store), name)
	assert.Equal(t, []protocol.Location{
		{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 2, Character: 9},
			End:   protocol.Position{Line: 2, Character: 21},
		}},
		{URI: "test2", Range: protocol.Range{
			Start: protocol.Position{Line: 2, Character: 0},
			End:   protocol.Position{Line: 2, Character: 12},
		}},
	}, store.GetReferences(fqn))

	doc3 := openDocument(store, "../cases/class.php", "test3")
	indexDocument(store, "../cases/reference/classesAndInterfaces.php", "test4")
	pos = protocol.Position{
		Line:      8,
		Character: 30,
	}
	sym = doc3.HasTypesAtPos(pos)
	for _, typ := range sym.GetTypes().Resolve() {
		assert.Equal(t, []protocol.Location{
			{URI: "test1", Range: protocol.Range{
				Start: protocol.Position{Line: 2, Character: 22},
				End:   protocol.Position{Line: 2, Character: 31},
			}},
			{URI: "test1", Range: protocol.Range{
				Start: protocol.Position{Line: 9, Character: 11},
				End:   protocol.Position{Line: 9, Character: 20},
			}},
			{URI: "test3", Range: protocol.Range{
				Start: protocol.Position{Line: 15, Character: 25},
				End:   protocol.Position{Line: 15, Character: 34},
			}},
			{URI: "test3", Range: protocol.Range{
				Start: protocol.Position{Line: 2, Character: 6},
				End:   protocol.Position{Line: 2, Character: 15},
			}},
			{URI: "test3", Range: protocol.Range{
				Start: protocol.Position{Line: 8, Character: 25},
				End:   protocol.Position{Line: 8, Character: 34},
			}},
			{URI: "test4", Range: protocol.Range{
				Start: protocol.Position{Line: 2, Character: 4},
				End:   protocol.Position{Line: 2, Character: 13},
			}},
			{URI: "test4", Range: protocol.Range{
				Start: protocol.Position{Line: 4, Character: 0},
				End:   protocol.Position{Line: 4, Character: 9},
			}},
		}, store.GetReferences(typ.GetFQN()))
	}

	indexDocument(store, "../cases/interface.php", "test5")
	pos = protocol.Position{
		Line:      12,
		Character: 50,
	}
	sym = doc3.HasTypesAtPos(pos)
	for _, typ := range sym.GetTypes().Resolve() {
		assert.Equal(t, []protocol.Location{
			{URI: "test3", Range: protocol.Range{
				Start: protocol.Position{Line: 12, Character: 43},
				End:   protocol.Position{Line: 12, Character: 57},
			}},
			{URI: "test3", Range: protocol.Range{
				Start: protocol.Position{Line: 15, Character: 61},
				End:   protocol.Position{Line: 15, Character: 75},
			}},
			{URI: "test4", Range: protocol.Range{
				Start: protocol.Position{Line: 6, Character: 0},
				End:   protocol.Position{Line: 6, Character: 14},
			}},
			{URI: "test5", Range: protocol.Range{
				Start: protocol.Position{Line: 7, Character: 10},
				End:   protocol.Position{Line: 7, Character: 24},
			}},
		}, store.GetReferences(typ.GetFQN()))
	}

	t.Run("Method", func(t *testing.T) {
		store := setupStore(t.Name()+"-", t.Name())
		indexDocument(store, "../cases/method.php", "method")
		indexDocument(store, "../cases/reference/methodAccess.php", "methodAccess")

		assert.Equal(t, []protocol.Location{
			{URI: t.Name() + "-method", Range: protocol.Range{
				Start: protocol.Position{Line: 11, Character: 20},
				End:   protocol.Position{Line: 11, Character: 31},
			}},
			{URI: t.Name() + "-methodAccess", Range: protocol.Range{
				Start: protocol.Position{Line: 3, Character: 7},
				End:   protocol.Position{Line: 3, Character: 18},
			}},
		}, store.GetReferences("\\TestMethodClass::testMethod3()"))

		assert.Equal(t, []protocol.Location{
			{URI: t.Name() + "-methodAccess", Range: protocol.Range{
				Start: protocol.Position{Line: 14, Character: 15},
				End:   protocol.Position{Line: 14, Character: 22},
			}},
			{URI: t.Name() + "-methodAccess", Range: protocol.Range{
				Start: protocol.Position{Line: 7, Character: 20},
				End:   protocol.Position{Line: 7, Character: 27},
			}},
		}, store.GetReferences("\\TestMethodClass2::method1()"))
	})

	t.Run("Const", func(t *testing.T) {
		store := setupStore(t.Name()+"-", t.Name())
		indexDocument(store, "../cases/reference/classConstAccess.php", "classConstAccess")
		assert.Equal(t, []protocol.Location{
			{URI: t.Name() + "-classConstAccess", Range: protocol.Range{
				Start: protocol.Position{Line: 12, Character: 16},
				End:   protocol.Position{Line: 12, Character: 22},
			}},
			{URI: t.Name() + "-classConstAccess", Range: protocol.Range{
				Start: protocol.Position{Line: 4, Character: 10},
				End:   protocol.Position{Line: 4, Character: 16},
			}},
			{URI: t.Name() + "-classConstAccess", Range: protocol.Range{
				Start: protocol.Position{Line: 8, Character: 16},
				End:   protocol.Position{Line: 8, Character: 22},
			}},
		}, store.GetReferences("\\TestConstClass::CONST1"))
	})

	t.Run("Property", func(t *testing.T) {
		store := setupStore(t.Name()+"-", t.Name())
		indexDocument(store, "../cases/reference/propertyAccess.php", "propertyAccess")
		assert.Equal(t, []protocol.Location{
			{URI: t.Name() + "-propertyAccess", Range: protocol.Range{
				Start: protocol.Position{Line: 16, Character: 7},
				End:   protocol.Position{Line: 16, Character: 12},
			}},
			{URI: t.Name() + "-propertyAccess", Range: protocol.Range{
				Start: protocol.Position{Line: 4, Character: 11},
				End:   protocol.Position{Line: 4, Character: 17},
			}},
			{URI: t.Name() + "-propertyAccess", Range: protocol.Range{
				Start: protocol.Position{Line: 9, Character: 15},
				End:   protocol.Position{Line: 9, Character: 20},
			}},
		}, store.GetReferences("\\TestPropertyClass1::$prop1"))
		assert.Equal(t, []protocol.Location{
			{URI: t.Name() + "-propertyAccess", Range: protocol.Range{
				Start: protocol.Position{Line: 10, Character: 16},
				End:   protocol.Position{Line: 10, Character: 22},
			}},
			{URI: t.Name() + "-propertyAccess", Range: protocol.Range{
				Start: protocol.Position{Line: 14, Character: 20},
				End:   protocol.Position{Line: 14, Character: 26},
			}},
			{URI: t.Name() + "-propertyAccess", Range: protocol.Range{
				Start: protocol.Position{Line: 5, Character: 18},
				End:   protocol.Position{Line: 5, Character: 24},
			}},
		}, store.GetReferences("\\TestPropertyClass1::$prop2"))
	})
}

func TestValidatingReferences(t *testing.T) {
	store := setupStore("test", t.Name())
	doc := NewDocument("test1", []byte(`<?php
function testFunction1() {}`))
	doc.Load()
	store.SyncDocument(doc)
	doc = NewDocument("test1", []byte(`<?php

function testFunction1() {}`))
	doc.Load()
	store.SyncDocument(doc)
	results := store.GetReferences("\\testFunction1")
	assert.Equal(t, []protocol.Location{
		{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 2, Character: 9},
			End:   protocol.Position{Line: 2, Character: 22},
		}},
	}, results)
}
