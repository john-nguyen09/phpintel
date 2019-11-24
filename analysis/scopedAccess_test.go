package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/util"
)

func TestScopedAccess(t *testing.T) {
	scopedAccessTest := "../cases/completion/scopedAccess.php"
	data, _ := ioutil.ReadFile(scopedAccessTest)
	document := NewDocument(util.PathToUri(scopedAccessTest), string(data))
	document.Load()

	cupaloy.SnapshotT(t, document.Children)
}
