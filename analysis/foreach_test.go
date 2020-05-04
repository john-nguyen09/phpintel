package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestForeach(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/foreach.php")
	if err != nil {
		panic(err)
	}
	store := setupStore("test", "TestForeach")
	document := NewDocument("test1", data)
	document.Load()
	h := document.HasTypesAt(75)
	h.Resolve(NewResolveContext(store, document))
	cupaloy.SnapshotT(t, h.GetTypes())
}
