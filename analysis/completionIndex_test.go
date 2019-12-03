package analysis

import (
	"path/filepath"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestCompletionIndex(t *testing.T) {
	testMethodClassTest, _ := filepath.Abs("../cases/method.php")
	store, _ := NewStore("./testData/TestCompletionIndex")
	indexDocument(store, testMethodClassTest, "test1")

	cupaloy.SnapshotT(t, store.SearchClasses("T"))
}
