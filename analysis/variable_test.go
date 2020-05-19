package analysis

import (
	"sort"
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
	vt.add(var1, protocol.Position{Line: 0, Character: 5}, true)
	vt.add(var3, protocol.Position{Line: 5, Character: 5}, true)
	vt.add(var2, protocol.Position{Line: 3, Character: 8}, true)
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
	vt.add(var4, protocol.Position{Line: 7, Character: 4}, true)

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
	sort.Slice(actualNames, func(i, j int) bool {
		return actualNames[i] < actualNames[j]
	})
	assert.Equal(t, []string{"$var1", "$var2"}, actualNames)

	vars = vt.GetVariables(protocol.Position{Line: 2, Character: 0})
	for _, v := range vars {
		if v.Name == "$var1" {
			assert.Equal(t, "\\Class1", v.GetTypes().ToString())
		}
	}

	vars = vt.GetVariables(protocol.Position{Line: 4, Character: 0})
	for _, v := range vars {
		if v.Name == "$var1" {
			assert.Equal(t, "\\Class2", v.GetTypes().ToString())
		}
	}
}

func TestUnusedVariables(t *testing.T) {
	doc := NewDocument("test1", []byte(`<?php
$var1 = new Class();

$var1 = new DateTime();
$var1->modify('tomorrow');

function testFunction1()
{
	$var1 = new DateTime();
	$var1->format('U');

	$var1 = new Class();
}`))
	doc.Load()

	t.Run("TestDocumentUnusedVariables", func(t *testing.T) {
		results := []protocol.Location{}
		for _, unusedVar := range doc.UnusedVariables() {
			results = append(results, unusedVar.GetLocation())
		}
		assert.Equal(t, []protocol.Location{
			{URI: "test1", Range: protocol.Range{
				Start: protocol.Position{Line: 1, Character: 0},
				End:   protocol.Position{Line: 1, Character: 5},
			}},
			{URI: "test1", Range: protocol.Range{
				Start: protocol.Position{Line: 11, Character: 1},
				End:   protocol.Position{Line: 11, Character: 6},
			}},
		}, results)
	})
}
