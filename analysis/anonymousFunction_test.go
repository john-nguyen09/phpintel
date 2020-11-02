package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/stretchr/testify/assert"
)

func TestAnonymousFunction(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/anonymousFunction.php")
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", data)
	document.Load()
	anonFuncs := []Symbol{}
	tra := newTraverser()
	tra.traverseDocument(document, func(_ *traverser, s Symbol, _ []Symbol) {
		if anonFunc, ok := s.(*AnonymousFunction); ok {
			anonFuncs = append(anonFuncs, anonFunc)
		}
	})
	assert.Equal(t, 3, len(anonFuncs))
	ranges := []protocol.Range{
		{Start: protocol.Position{Line: 2, Character: 9}, End: protocol.Position{Line: 4, Character: 1}},
		{Start: protocol.Position{Line: 8, Character: 29}, End: protocol.Position{Line: 11, Character: 1}},
		{Start: protocol.Position{Line: 13, Character: 13}, End: protocol.Position{Line: 15, Character: 1}},
	}
	for i, anonFunc := range anonFuncs {
		assert.Equal(t, ranges[i], anonFunc.GetLocation().Range)
	}

	variableTable := document.GetVariableTableAt(protocol.Position{Line: 10, Character: 32})
	variable := variableTable.get("$func1", protocol.Position{Line: 10, Character: 32})
	assert.NotEqual(t, nil, variable)
	variable2 := variableTable.get("$var1", protocol.Position{Line: 10, Character: 32})
	assert.Equal(t, "\\Schema", variable2.GetTypes().ToString())
}
