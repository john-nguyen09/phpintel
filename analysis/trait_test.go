package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestTrait(t *testing.T) {
	traitTest := "../cases/trait.php"
	data, err := ioutil.ReadFile(traitTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", string(data))
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}
