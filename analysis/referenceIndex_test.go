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
	fqn := doc2.ImportTableAtPos(pos).GetFunctionReferenceFQN(store, name)
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
			{URI: "test5", Range: protocol.Range{
				Start: protocol.Position{Line: 7, Character: 10},
				End:   protocol.Position{Line: 7, Character: 24},
			}},
		}, store.GetReferences(typ.GetFQN()))
	}
}
