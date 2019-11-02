package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/util"
)

func TestTrait(t *testing.T) {
	traitTest := "../cases/trait.php"
	data, err := ioutil.ReadFile(traitTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument(util.PathToUri(traitTest), string(data))
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}
