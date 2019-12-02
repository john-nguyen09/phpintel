package analysis

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/util"
)

func TestNamespace(t *testing.T) {
	references2, _ := filepath.Abs("../cases/namespace/references2.php")
	data, err := ioutil.ReadFile(references2)
	if err != nil {
		panic(err)
	}
	document := NewDocument(util.PathToUri(references2), string(data))
	document.Load()

	cupaloy.SnapshotT(t, document.importTable)
}
