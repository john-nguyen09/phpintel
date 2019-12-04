package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestReferencesInsideArray(t *testing.T) {
	testCase := "../cases/references_inside_array.php"
	data, _ := ioutil.ReadFile(testCase)
	document := NewDocument("test1", string(data))
	document.Load()

	cupaloy.SnapshotT(t, document.Children)
}
