package analysis

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestPhpStormStub(t *testing.T) {
	store, err := setupStore("", "stub_test")
	store.LoadStubs()
	if err != nil {
		panic(err)
	}
	functions := store.GetFunctions("\\preg_match")
	cupaloy.SnapshotT(t, functions)
}
