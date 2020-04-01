package analysis

import (
	"path/filepath"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestCompletionIndex(t *testing.T) {
	testMethodClassTest, _ := filepath.Abs("../cases/method.php")
	store := setupStore("test", "TestCompletionIndex")
	indexDocument(store, testMethodClassTest, "test1")

	classes, _ := store.SearchClasses("T", NewSearchOptions())
	cupaloy.SnapshotT(t, classes)
}
