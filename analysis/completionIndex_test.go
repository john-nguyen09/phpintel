package analysis

import (
	"path/filepath"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestCompletionIndex(t *testing.T) {
	testMethodClassTest, _ := filepath.Abs("../cases/method.php")
	store, _ := NewStore("./testData/TestCompletionIndex")
	store.IndexDocument(testMethodClassTest)

	cupaloy.SnapshotT(t, store.SearchClasses("T"))
}
