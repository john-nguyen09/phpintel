package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestAnonymousFunction(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/anonymousFunction.php")
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", data)
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}
