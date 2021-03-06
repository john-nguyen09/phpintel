package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
)

func TestAssignment(t *testing.T) {
	assignmentTest := "../cases/variableAssignment.php"
	data, err := ioutil.ReadFile(assignmentTest)
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", data)
	document.Load()
	cupaloy.SnapshotT(t, document.hasTypesSymbols())
}

func TestRegressionOnErrorAssignment(t *testing.T) {
	doc := NewDocument("test1", []byte(`<?php
$var1 = !empty($var2) ? 
`))
	doc.Load()
}
