package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/john-nguyen09/go-phpparser/parser"
	"github.com/john-nguyen09/phpintel/util"
)

func TestLineOffset(t *testing.T) {
	classTest := "../cases/class.php"
	data, err := ioutil.ReadFile(classTest)
	if err != nil {
		panic(err)
	}
	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(classTest), string(data), rootNode)
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
	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(classTest), string(data), rootNode)
	position := document.positionAt(9)
	if position.Line != 2 || position.Character != 0 {
		t.Errorf("Expect document.positionAt(9) = 2:0, got %v", position)
	}
	position = document.positionAt(174)
	if position.Line != 12 || position.Character != 2 {
		t.Errorf("Expect document.positionAt(8) = 12:2, got %v", position)
	}
}