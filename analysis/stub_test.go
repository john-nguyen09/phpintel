package analysis

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestPhpStormStub(t *testing.T) {
	store, err := NewStore("", "testData/stub_test")
	if err != nil {
		panic(err)
	}
	functions := store.GetFunctions("\\preg_match")
	cupaloy.SnapshotT(t, functions)
}