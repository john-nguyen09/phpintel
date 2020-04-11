package analysis

import (
	"testing"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/stretchr/testify/assert"
)

func TestVariableTable(t *testing.T) {
	vt := newVariableTable(protocol.Range{
		Start: protocol.Position{Line: 0, Character: 0},
		End:   protocol.Position{Line: 8, Character: 0},
	}, 0)
	types1 := newTypeComposite()
	types1.add(NewTypeString("\\Class1"))
	types2 := newTypeComposite()
	types2.add(NewTypeString("\\Class2"))
	types3 := newTypeComposite()
	types3.add(NewTypeString("\\Class3"))
	var1 := &Variable{
		Expression: Expression{
			Type: types1,
			Name: "$var1",
		},
	}
	var2 := &Variable{
		Expression: Expression{
			Type: types2,
			Name: "$var1",
		},
	}
	var3 := &Variable{
		Expression: Expression{
			Type: types3,
			Name: "$var1",
		},
	}
	vt.add(var1, protocol.Position{Line: 0, Character: 5})
	vt.add(var3, protocol.Position{Line: 5, Character: 5})
	vt.add(var2, protocol.Position{Line: 3, Character: 8})
	actualPositions := []protocol.Position{}
	for _, v := range vt.variables["$var1"] {
		actualPositions = append(actualPositions, v.start)
	}
	assert.Equal(t, []protocol.Position{
		{Line: 0, Character: 5},
		{Line: 3, Character: 8},
		{Line: 5, Character: 5},
	}, actualPositions)

	v := vt.get("$var1", protocol.Position{Line: 1, Character: 0})
	assert.Equal(t, "\\Class1", v.GetTypes().ToString())

	var4 := &Variable{
		Expression: Expression{
			Name: "$var2",
		},
	}
	vt.add(var4, protocol.Position{Line: 7, Character: 4})

	vars := vt.GetVariables(protocol.Position{Line: 6, Character: 0})
	actualNames := []string{}
	for _, v := range vars {
		actualNames = append(actualNames, v.Name)
	}
	assert.Equal(t, []string{"$var1"}, actualNames)

	vars = vt.GetVariables(protocol.Position{Line: 7, Character: 5})
	actualNames = []string{}
	for _, v := range vars {
		actualNames = append(actualNames, v.Name)
	}
	assert.Equal(t, []string{"$var1", "$var2"}, actualNames)
}
