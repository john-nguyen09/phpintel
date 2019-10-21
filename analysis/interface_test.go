package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/go-phpparser/parser"
	"github.com/john-nguyen09/phpintel/util"
)

func TestInterface(t *testing.T) {
	interfaceTest := "../cases/interface.php"
	data, err := ioutil.ReadFile(interfaceTest)
	if err != nil {
		panic(err)
	}

	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(interfaceTest), string(data), rootNode)
	cupaloy.SnapshotT(t, document.Children)
}
