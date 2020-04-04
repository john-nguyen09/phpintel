package analysis

import (
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/stretchr/testify/assert"
)

func TestParenthesisedExpression(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/parenthesised.php")
	assert.NoError(t, err)
	doc := NewDocument("test1", data)
	doc.Load()

	cupaloy.SnapshotT(t, doc.hasTypesSymbols())
}

func TestExpressionsInsideParenthesisedExpression(t *testing.T) {
	doc := NewDocument("test1", []byte(`<?php
(new DateTime)->modify();
if (empty($data)) { }`))
	doc.Load()

	assert.Equal(t, "*analysis.ClassTypeDesignator", reflect.TypeOf(doc.HasTypesAtPos(protocol.Position{
		Line:      1,
		Character: 8,
	})).String())
	assert.Equal(t, "*analysis.FunctionCall", reflect.TypeOf(doc.HasTypesAtPos(protocol.Position{
		Line:      2,
		Character: 7,
	})).String())
	assert.Equal(t, "*analysis.Variable", reflect.TypeOf(doc.HasTypesAtPos(protocol.Position{
		Line:      2,
		Character: 11,
	})).String())
}
