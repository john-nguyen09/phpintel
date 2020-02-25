package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestScopedMemberAccess(t *testing.T) {
	scopedPropertyAccessTest := "../cases/memberAccess.php"
	data, err := ioutil.ReadFile(scopedPropertyAccessTest)
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", data)
	document.Load()
	cupaloy.SnapshotT(t, document.hasTypesSymbols)
}
