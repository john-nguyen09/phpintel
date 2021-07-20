package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"
)

func TestForeach(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/foreach.php")
	if err != nil {
		panic(err)
	}
	withTestStore("test", "TestForeach", func(store *Store) {
		document := NewDocument("test1", data)
		document.Load()
		ctx := NewResolveContext(NewQuery(store), document)
		h := document.HasTypesAt(75)
		h.Resolve(ctx)
		cupaloy.SnapshotT(t, h.GetTypes())

		h = document.HasTypesAt(120)
		h.Resolve(ctx)
		assert.Equal(t, "string", h.GetTypes().ToString())
	})
}
