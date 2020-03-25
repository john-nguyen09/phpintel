package analysis

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClone(t *testing.T) {
	store, err := setupStore("test", "TestClone")
	assert.NoError(t, err)
	doc := NewDocument("test1", []byte(`<?php
$var1 = new DateTime();
$var2 = clone $var1;
$var3 = clone $var2;`))
	doc.Load()
	store.SyncDocument(doc)
	assert.Equal(t, "*analysis.Variable", reflect.TypeOf(doc.hasTypesSymbols[2]).String())
	assert.Equal(t, "\\DateTime", doc.hasTypesSymbols[2].GetTypes().ToString())

	var3 := doc.hasTypesSymbols[4]
	var3.Resolve(NewResolveContext(store, doc))
	assert.Equal(t, "\\DateTime", var3.GetTypes().ToString())
}

func TestInstanceOf(t *testing.T) {
	doc := NewDocument("test1", []byte(`<?php
if ($var1 instanceof DateTime) {
}`))
	doc.Load()
	assert.Equal(t, "*analysis.Variable", reflect.TypeOf(doc.hasTypesSymbols[0]).String())
	assert.Equal(t, "\\DateTime", doc.hasTypesSymbols[0].GetTypes().ToString())
}
