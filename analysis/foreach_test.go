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
	store, err := setupStore("test", "TestForeach")
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", string(data))
	document.Load()
	h := document.HasTypesAt(75)
	h.Resolve(NewResolveContext(store, document))
	cupaloy.SnapshotT(t, h.GetTypes())
}
