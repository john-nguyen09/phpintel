package analysis

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/stretchr/testify/assert"
)

func TestNamespace(t *testing.T) {
	references2, _ := filepath.Abs("../cases/namespace/references2.php")
	data, err := ioutil.ReadFile(references2)
	if err != nil {
		panic(err)
	}
	document := NewDocument("references2", string(data))
	document.Load()

	cupaloy.SnapshotT(t, document.importTable)
}

type useResult struct {
	label  string
	insert string
}

type useTestCase struct {
	doc    *Document
	s      Symbol
	name   TypeString
	word   string
	result useResult
}

func TestNamespaceAndUse(t *testing.T) {
	doc1 := NewDocument("importTable1", `<?php
namespace TestNamespace1;

class TestClass1 {}

function TestFunction1() {}`)
	doc1.Load()

	doc2 := NewDocument("importTable2", `<?php
`)
	doc2.Load()

	doc3 := NewDocument("importTable3", `<?php
namespace TestNamespace2;`)
	doc3.Load()

	class := doc1.Children[0].(*Class)
	function := doc1.Children[1].(*Function)

	cases := []useTestCase{
		useTestCase{doc2, class, class.Name, "Test", useResult{"TestClass1", "use TestNamespace1\\TestClass1;"}},
		useTestCase{doc2, function, function.Name, "Function", useResult{"TestFunction1", "use function TestNamespace1\\TestFunction1;"}},
		useTestCase{doc2, function, function.Name, "TestNamespace1\\t", useResult{"TestFunction1", ""}},
		useTestCase{doc3, function, function.Name, "test", useResult{"TestFunction1", "use function TestNamespace1\\TestFunction1;"}},
		useTestCase{doc3, class, class.Name, "TestNamespace1\\T", useResult{"TestClass1", "use TestNamespace1;"}},
		useTestCase{doc3, function, function.Name, "TestNamespace1\\t", useResult{"TestFunction1", "use TestNamespace1;"}},
	}

	for _, testCase := range cases {
		label, edit := testCase.doc.GetImportTable().ResolveToQualified(testCase.doc, testCase.s, testCase.name, testCase.word)
		insertText := ""
		if edit != nil {
			insertText = strings.TrimSpace(edit.NewText)
		}
		assert.Equal(t, useResult{label, insertText}, testCase.result)
	}
}
