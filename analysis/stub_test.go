package analysis

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestPhpStormStub(t *testing.T) {
	store := setupStore("", "stub_test")
	store.LoadStubs()
	functions := store.GetFunctions("\\preg_match")
	cupaloy.SnapshotT(t, functions)
}
