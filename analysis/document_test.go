package analysis

import (
	"encoding/json"
	"io/ioutil"
	"reflect"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/stretchr/testify/assert"
)

func TestLineOffset(t *testing.T) {
	classTest := "../cases/class.php"
	data, err := ioutil.ReadFile(classTest)
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", data)
	document.Load()
	line := document.lineAt(39)
	if line != 3 {
		t.Errorf("lineAt(39) != 3, got: %d", line)
	}
	line = document.lineAt(64)
	if line != 5 {
		t.Errorf("lineAt(64) != 5, got: %d", line)
	}
	line = document.lineAt(38)
	if line != 3 {
		t.Errorf("lineAt(38) != 3, got: %d", line)
	}
}

func TestPosition(t *testing.T) {
	classTest := "../cases/class.php"
	data, err := ioutil.ReadFile(classTest)
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", data)
	document.Load()
	position := document.positionAt(9)
	if position.Line != 2 || position.Character != 0 {
		t.Errorf("Expect document.positionAt(9) = 2:0, got %v", position)
	}
	position = document.positionAt(174)
	if position.Line != 12 || position.Character != 29 {
		t.Errorf("Expect document.positionAt(174) = 12:29, got %v", position)
	}
}

func TestSymbolAt(t *testing.T) {
	memberAccess := "../cases/memberAccess.php"
	data, _ := ioutil.ReadFile(memberAccess)
	document := NewDocument("test1", data)
	document.Load()
	symbol := document.HasTypesAt(14)
	if _, ok := symbol.(*ClassAccess); !ok {
		t.Errorf("symbolAt(14) is not *ClassAccess but %T", symbol)
	}
	symbol = document.HasTypesAt(20)
	if _, ok := symbol.(*ScopedPropertyAccess); !ok {
		t.Errorf("symbolAt(20) is not *ScopedPropertyAccess but %T", symbol)
	}
	symbol = document.HasTypesAt(19)
	if symbol != nil {
		t.Errorf("symbolAt(19) is not nil but %T", symbol)
	}
}

func TestApplyChanges(t *testing.T) {
	document := NewDocument("test1", []byte("<?php\necho 'Hello world';"))
	document.ApplyChanges([]protocol.TextDocumentContentChangeEvent{
		protocol.TextDocumentContentChangeEvent{
			Range: &protocol.Range{
				Start: protocol.Position{
					Line:      1,
					Character: 19,
				},
				End: protocol.Position{
					Line:      1,
					Character: 19,
				},
			},
			RangeLength: 0,
			Text:        "\n// This is inserted",
		},
	})
	data, _ := json.MarshalIndent(document.getLines(), "", "  ")
	cupaloy.SnapshotT(t, string(data))
}

func TestIntrinsics(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/intrinsics.php")
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", data)
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}

func TestChainedMethodCalls(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/RouteServiceProvider.php")
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", data)
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}

func TestSymbolBefore(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/chainedMethod.php")
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", data)
	document.Load()
	if reflect.TypeOf(document.HasTypesBeforePos(protocol.Position{
		Line:      1,
		Character: 32,
	})).String() != "*analysis.MethodAccess" {
		t.FailNow()
	}
}

type wordTestCase struct {
	doc      *Document
	pos      protocol.Position
	expected string
}

func TestWordAt(t *testing.T) {
	doc1 := NewDocument("test1", []byte(`<?php function1(\Modules\`))

	testCases := []wordTestCase{
		wordTestCase{doc1, protocol.Position{Line: 0, Character: 25}, "\\Modules\\"},
	}

	for _, testCase := range testCases {
		actual := testCase.doc.WordAtPos(testCase.pos)
		assert.Equal(t, testCase.expected, actual)
	}
}

func TestVarTableAt(t *testing.T) {
	doc1 := NewDocument("test1", []byte(`<?php
function func1($param1) {

	DB::transaction(function() use ($param1) {
	});
}`))
	doc1.Load()
	varTable := doc1.GetVariableTableAt(protocol.Position{
		Line:      2,
		Character: 4,
	})
	assert.NotNil(t, varTable)
	assert.Equal(t, protocol.Range{
		Start: protocol.Position{
			Line:      1,
			Character: 0,
		},
		End: protocol.Position{
			Line:      5,
			Character: 1,
		},
	}, varTable.locationRange)
}
