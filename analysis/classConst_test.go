package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/go-phpparser/parser"
	"github.com/john-nguyen09/phpintel/util"
)

func TestClassConst(t *testing.T) {
	classConstTest := "../cases/classConst.php"
	data, err := ioutil.ReadFile(classConstTest)
	if err != nil {
		panic(err)
	}

	rootNode := parser.Parse(string(data))
	document := newDocument(util.PathToUri(classConstTest), string(data), rootNode)
	cupaloy.SnapshotT(t, document.Children)
}
