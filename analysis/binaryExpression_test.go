package analysis

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInstanceOf(t *testing.T) {
	doc := NewDocument("test1", []byte(`<?php
if ($var1 instanceof DateTime) {
}`))
	doc.Load()
	assert.Equal(t, "*analysis.Variable", reflect.TypeOf(doc.hasTypesSymbols[0]).String())
	assert.Equal(t, "\\DateTime", doc.hasTypesSymbols[0].GetTypes().ToString())
}
