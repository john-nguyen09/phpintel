package analysis

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestPhpStormStub(t *testing.T) {
	withTestStore("", "stub_test", func(store *Store) {
		store.LoadStubs()
		functions := store.GetFunctions("\\preg_match")
		cupaloy.SnapshotT(t, functions)
	})
}
