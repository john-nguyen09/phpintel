package analysis

import (
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

func TestLineOffset(t *testing.T) {
	classTest := "../cases/class.php"
	data, err := ioutil.ReadFile(classTest)
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", string(data))
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
	document := NewDocument("test1", string(data))
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
	document := NewDocument("test1", string(data))
	document.Load()
	symbol := document.SymbolAt(14)
	if _, ok := symbol.(*ClassAccess); !ok {
		t.Errorf("symbolAt(14) is not *ClassAccess but %T", symbol)
	}
	symbol = document.SymbolAt(20)
	if _, ok := symbol.(*ScopedPropertyAccess); !ok {
		t.Errorf("symbolAt(20) is not *ScopedPropertyAccess but %T", symbol)
	}
	symbol = document.SymbolAt(19)
	if symbol != nil {
		t.Errorf("symbolAt(19) is not nil but %T", symbol)
	}
}

func TestApplyChanges(t *testing.T) {
	document := NewDocument("test1", "<?php\necho 'Hello world';")
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
	document := NewDocument("test1", string(data))
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}
