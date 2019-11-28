package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/util"
)

func TestScopedMemberAccess(t *testing.T) {
	scopedPropertyAccessTest := "../cases/memberAccess.php"
	data, err := ioutil.ReadFile(scopedPropertyAccessTest)
	if err != nil {
		panic(err)
	}
	document := NewDocument(util.PathToUri(scopedPropertyAccessTest), string(data))
	document.Load()
	cupaloy.SnapshotT(t, document.Children)
}

func TestScopedAccess(t *testing.T) {
	scopedAccessTest := "../cases/completion/scopedAccess.php"
	data, _ := ioutil.ReadFile(scopedAccessTest)
	document := NewDocument(util.PathToUri(scopedAccessTest), string(data))
	document.Load()

	cupaloy.SnapshotT(t, document.Children)
}
