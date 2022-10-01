package analysis

import (
	"reflect"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/stretchr/testify/assert"
)

func TestClone(t *testing.T) {
	withTestStore("test", "TestClone", func(store *Store) {
		doc := NewDocument("test1", []byte(`<?php
$var1 = new DateTime();
$var2 = clone $var1;
$var3 = clone $var2;`))
		doc.Load()
		store.SyncDocument(doc)
		assert.Equal(t, "*analysis.Variable", reflect.TypeOf(doc.hasTypesSymbols()[2]).String())
		assert.Equal(t, "\\DateTime", doc.hasTypesSymbols()[2].GetTypes().ToString())

		var3 := doc.hasTypesSymbols()[4]
		var3.Resolve(NewResolveContext(NewQuery(store), doc))
		assert.Equal(t, "\\DateTime", var3.GetTypes().ToString())
	})
}

func TestInstanceOf(t *testing.T) {
	doc := NewDocument("test1", []byte(`<?php
if ($var1 instanceof DateTime) {
}`))
	doc.Load()
	assert.Equal(t, "*analysis.Variable", reflect.TypeOf(doc.hasTypesSymbols()[0]).String())
	assert.Equal(t, "\\DateTime", doc.hasTypesSymbols()[0].GetTypes().ToString())
}

func TestRequireOnce(t *testing.T) {
	doc := NewDocument("test1", []byte(`<?php
namespace block_purchasemanagement\payment;

class payment_service {
    private function ensure_profile_load($user) {
        global $CFG;

        require_once("$CFG->dirroot/user/profile/lib.php");
        profile_load_custom_fields($user);
        return $user;
    }
}`))
	doc.Load()
	results := []Symbol{}
	results = append(results, doc.HasTypesAtPos(protocol.Position{
		Line:      8,
		Character: 15,
	}))
	results = append(results, doc.HasTypesAtPos(protocol.Position{
		Line:      7,
		Character: 25,
	}))
	cupaloy.SnapshotT(t, results)
}

func TestErrorControl(t *testing.T) {
	doc := NewDocument("test1", []byte(`<?php
$a = '1';
@unlink($a);`))
	doc.Load()
	results := []Symbol{}
	results = append(results, doc.HasTypesAtPos(protocol.Position{
		Line:      2,
		Character: 4,
	}))
	results = append(results, doc.HasTypesAtPos(protocol.Position{
		Line:      2,
		Character: 10,
	}))
	cupaloy.SnapshotT(t, results)
}
