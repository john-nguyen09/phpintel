package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/util"

	"github.com/john-nguyen09/go-phpparser/parser"
)

func TestConstant(t *testing.T) {
	constTest := "../cases/const.php"
	data, err := ioutil.ReadFile(constTest)
	if err != nil {
		panic(err)
	}

	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(constTest), string(data), rootNode)
	cupaloy.SnapshotT(t, document.Children)
}
