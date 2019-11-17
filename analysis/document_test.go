package analysis

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"testing"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/john-nguyen09/phpintel/util"
)

func TestLineOffset(t *testing.T) {
	classTest := "../cases/class.php"
	data, err := ioutil.ReadFile(classTest)
	if err != nil {
		panic(err)
	}
	document := NewDocument(util.PathToUri(classTest), string(data))
	document.Load()
	line := document.lineAt(39)
	if line != 6 {
		t.Errorf("lineAt(39) != 6, got: %d", line)
	}
	line = document.lineAt(64)
	if line != 6 {
		t.Errorf("lineAt(64) != 6, got: %d", line)
	}
	line = document.lineAt(38)
	if line != 5 {
		t.Errorf("lineAt(38) != 5, got: %d", line)
	}
}

func TestPosition(t *testing.T) {
	classTest := "../cases/class.php"
	data, err := ioutil.ReadFile(classTest)
	if err != nil {
		panic(err)
	}
	document := NewDocument(util.PathToUri(classTest), string(data))
	document.Load()
	position := document.positionAt(9)
	if position.Line != 2 || position.Character != 0 {
		t.Errorf("Expect document.positionAt(9) = 2:0, got %v", position)
	}
	position = document.positionAt(174)
	if position.Line != 12 || position.Character != 2 {
		t.Errorf("Expect document.positionAt(8) = 12:2, got %v", position)
	}
}

func TestSymbolAt(t *testing.T) {
	memberAccess := "../cases/memberAccess.php"
	data, _ := ioutil.ReadFile(memberAccess)
	document := NewDocument(util.PathToUri(memberAccess), string(data))
	document.Load()
	symbol := document.SymbolAt(14)
	fmt.Printf("%T\n", symbol)
	symbol = document.SymbolAt(20)
	fmt.Printf("%T\n", symbol)
	symbol = document.SymbolAt(19)
	fmt.Printf("%T\n", symbol)
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
	log.Println(string(data))
}
