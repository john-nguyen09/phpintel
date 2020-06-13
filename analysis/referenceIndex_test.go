package analysis

import (
	"sort"
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
	sortLocs := func(locs []protocol.Location) []protocol.Location {
		sort.Slice(locs, func(i, j int) bool {
			return locs[i].Range.String() < locs[j].Range.String()
		})
		return locs
	}
	sym := doc2.HasTypesAtPos(pos)
	assert.Equal(t, sortLocs([]protocol.Location{
		{URI: "test2", Range: protocol.Range{
			Start: protocol.Position{Line: 2, Character: 0},
			End:   protocol.Position{Line: 2, Character: 12},
		}},
		{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 2, Character: 9},
			End:   protocol.Position{Line: 2, Character: 21},
		}},
	}), sortLocs(store.GetReferences(SymToRefs(doc2, sym)[0])))

	doc3 := openDocument(store, "../cases/class.php", "test3")
	indexDocument(store, "../cases/reference/classesAndInterfaces.php", "test4")
	pos = protocol.Position{
		Line:      8,
		Character: 30,
	}
	sym = doc3.HasTypesAtPos(pos)
	assert.Equal(t, sortLocs([]protocol.Location{
		{URI: "test3", Range: protocol.Range{
			Start: protocol.Position{Line: 2, Character: 6},
			End:   protocol.Position{Line: 2, Character: 15},
		}},
		{URI: "test3", Range: protocol.Range{
			Start: protocol.Position{Line: 8, Character: 25},
			End:   protocol.Position{Line: 8, Character: 34},
		}},
		{URI: "test3", Range: protocol.Range{
			Start: protocol.Position{Line: 15, Character: 25},
			End:   protocol.Position{Line: 15, Character: 34},
		}},
		{URI: "test4", Range: protocol.Range{
			Start: protocol.Position{Line: 2, Character: 4},
			End:   protocol.Position{Line: 2, Character: 13},
		}},
		{URI: "test4", Range: protocol.Range{
			Start: protocol.Position{Line: 4, Character: 0},
			End:   protocol.Position{Line: 4, Character: 9},
		}},
		{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 2, Character: 22},
			End:   protocol.Position{Line: 2, Character: 31},
		}},
		{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 9, Character: 11},
			End:   protocol.Position{Line: 9, Character: 20},
		}},
	}), sortLocs(store.GetReferences(SymToRefs(doc3, sym)[0])))

	indexDocument(store, "../cases/interface.php", "test5")
	pos = protocol.Position{
		Line:      12,
		Character: 50,
	}
	sym = doc3.HasTypesAtPos(pos)
	assert.Equal(t, sortLocs([]protocol.Location{
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
	}), sortLocs(store.GetReferences(SymToRefs(doc3, sym)[0])))

	t.Run("Method", func(t *testing.T) {
		store := setupStore(t.Name()+"-", t.Name())
		indexDocument(store, "../cases/method.php", "method")
		indexDocument(store, "../cases/reference/methodAccess.php", "methodAccess")

		assert.Equal(t, sortLocs([]protocol.Location{
			{URI: t.Name() + "-method", Range: protocol.Range{
				Start: protocol.Position{Line: 11, Character: 20},
				End:   protocol.Position{Line: 11, Character: 31},
			}},
			{URI: t.Name() + "-methodAccess", Range: protocol.Range{
				Start: protocol.Position{Line: 3, Character: 7},
				End:   protocol.Position{Line: 3, Character: 18},
			}},
		}), sortLocs(store.GetReferences(".testMethod3()")))

		assert.Equal(t, sortLocs([]protocol.Location{
			{URI: t.Name() + "-methodAccess", Range: protocol.Range{
				Start: protocol.Position{Line: 14, Character: 15},
				End:   protocol.Position{Line: 14, Character: 22},
			}},
			{URI: t.Name() + "-methodAccess", Range: protocol.Range{
				Start: protocol.Position{Line: 7, Character: 20},
				End:   protocol.Position{Line: 7, Character: 27},
			}},
		}), sortLocs(store.GetReferences(".method1()")))
	})

	t.Run("Const", func(t *testing.T) {
		store := setupStore(t.Name()+"-", t.Name())
		indexDocument(store, "../cases/reference/classConstAccess.php", "classConstAccess")
		assert.Equal(t, sortLocs([]protocol.Location{
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
		}), sortLocs(store.GetReferences(".CONST1")))
	})

	t.Run("Property", func(t *testing.T) {
		store := setupStore(t.Name()+"-", t.Name())
		indexDocument(store, "../cases/reference/propertyAccess.php", "propertyAccess")
		assert.Equal(t, sortLocs([]protocol.Location{
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
		}), sortLocs(store.GetReferences(".$prop1")))
		assert.Equal(t, sortLocs([]protocol.Location{
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
		}), sortLocs(store.GetReferences(".$prop2")))
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
