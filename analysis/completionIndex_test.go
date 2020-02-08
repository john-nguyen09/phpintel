package analysis

import (
	"path/filepath"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestCompletionIndex(t *testing.T) {
	testMethodClassTest, _ := filepath.Abs("../cases/method.php")
	store, err := setupStore("test", "TestCompletionIndex")
	if err != nil {
		panic(err)
	}
	indexDocument(store, testMethodClassTest, "test1")

	classes, _ := store.SearchClasses("T", NewSearchOptions())
	cupaloy.SnapshotT(t, classes)
}
