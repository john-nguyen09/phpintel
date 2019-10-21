package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/util"

	"github.com/john-nguyen09/go-phpparser/parser"
)

func TestFunction(t *testing.T) {
	functionTest := "../cases/function.php"
	data, err := ioutil.ReadFile(functionTest)
	if err != nil {
		panic(err)
	}

	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(functionTest), string(data), rootNode)
	cupaloy.SnapshotT(t, document.Children)
}
