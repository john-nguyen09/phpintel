package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/go-phpparser/parser"
	"github.com/john-nguyen09/phpintel/util"
)

func TestMethod(t *testing.T) {
	methodTest := "../cases/method.php"
	data, err := ioutil.ReadFile(methodTest)
	if err != nil {
		panic(err)
	}

	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(methodTest), string(data), rootNode)
	cupaloy.SnapshotT(t, document.Children)
}
