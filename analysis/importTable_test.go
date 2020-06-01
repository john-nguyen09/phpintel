package analysis

import (
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/stretchr/testify/assert"
)

func TestNamespace(t *testing.T) {
	references2, _ := filepath.Abs("../cases/namespace/references2.php")
	data, err := ioutil.ReadFile(references2)
	if err != nil {
		panic(err)
	}
	document := NewDocument("references2", data)
	document.Load()

	cupaloy.SnapshotT(t, document.currImportTable())
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
	doc1 := NewDocument("importTable1", []byte(`<?php
namespace TestNamespace1;

class TestClass1 {}

function TestFunction1() {}`))
	doc1.Load()

	doc2 := NewDocument("importTable2", []byte(`<?php
`))
	doc2.Load()

	doc3 := NewDocument("importTable3", []byte(`<?php
namespace TestNamespace2;`))
	doc3.Load()

	doc4 := NewDocument("importTable4", []byte(`<?php
class DateTime {}

function TestFunction2() {}`))
	doc4.Load()

	class := doc1.Children[0].(*Class)
	function := doc1.Children[1].(*Function)
	class2 := doc4.Children[0].(*Class)
	function2 := doc4.Children[1].(*Function)

	cases := []useTestCase{
		{doc2, class, class.Name, "Test", useResult{"TestClass1", "use TestNamespace1\\TestClass1;"}},
		{doc2, function, function.Name, "Function", useResult{"TestFunction1", "use function TestNamespace1\\TestFunction1;"}},
		{doc2, function, function.Name, "TestNamespace1\\t", useResult{"TestFunction1", ""}},
		{doc3, function, function.Name, "test", useResult{"TestFunction1", "use function TestNamespace1\\TestFunction1;"}},
		{doc3, class, class.Name, "TestNamespace1\\T", useResult{"TestClass1", "use TestNamespace1;"}},
		{doc3, function, function.Name, "TestNamespace1\\t", useResult{"TestFunction1", "use TestNamespace1;"}},
		{doc3, class, class.Name, "\\TestNamespace1\\Te", useResult{"TestClass1", ""}},
		{doc3, function, function.Name, "\\TestNamespace1\\Test", useResult{"TestFunction1", ""}},
		{doc2, class2, class2.Name, "Dat", useResult{"DateTime", ""}},
		{doc3, class2, class2.Name, "Da", useResult{"DateTime", "use DateTime;"}},
		{doc3, function2, function2.Name, "Test", useResult{"TestFunction2", ""}},
	}

	for i, testCase := range cases {
		label, edit := testCase.doc.currImportTable().ResolveToQualified(testCase.doc, testCase.s, testCase.name, testCase.word)
		insertText := ""
		if edit != nil {
			insertText = strings.TrimSpace(edit.NewText)
		}
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			assert.Equal(t, testCase.result, useResult{label, insertText})
		})
	}
}

type functionReferenceFQNTestCase struct {
	importTable *ImportTable
	funcCall    string
	expected    string
}

func TestFunctionReferenceFQN(t *testing.T) {
	store := setupStore("test", "TestFunctionReferenceFQN")
	doc1 := NewDocument("test1", []byte(`<?php
namespace Namespace1;

use function Namespace2\func1;`))
	doc1.Load()

	doc2 := NewDocument("test2", []byte(`<?php
function func2() {
}`))
	doc2.Load()
	store.SyncDocument(doc2)

	doc3 := NewDocument("test3", []byte(`<?php namespace Namespace3;`))
	doc3.Load()

	testCases := []functionReferenceFQNTestCase{
		{doc1.importTables[0], "func1", "\\Namespace2\\func1"},
		{doc3.importTables[0], "func2", "\\func2"},
		{doc3.importTables[0], "func3", "\\Namespace3\\func3"},
	}
	for _, testCase := range testCases {
		actual := testCase.importTable.GetFunctionReferenceFQN(NewQuery(store), NewTypeString(testCase.funcCall))
		assert.Equal(t, testCase.expected, actual)
	}
}

func TestImportTableOnEmptyFile(t *testing.T) {
	doc := NewDocument("test1", []byte(""))
	doc.Load()
	importTable := doc.ImportTableAtPos(protocol.Position{
		Line:      0,
		Character: 0,
	})
	assert.NotNil(t, importTable)
}

func TestUnusedImports(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/unused/import.php")
	if err != nil {
		panic(err)
	}
	doc := NewDocument("test1", data)
	doc.Load()
	importTable := doc.ImportTableAtPos(protocol.Position{Line: 3, Character: 0})
	var results []protocol.Range
	unusedImportItems := importTable.unusedImportItems()
	for _, item := range unusedImportItems {
		results = append(results, item.locationRange)
	}
	assert.Equal(t, []protocol.Range{
		{Start: protocol.Position{Line: 4, Character: 4}, End: protocol.Position{Line: 4, Character: 14}},
		{Start: protocol.Position{Line: 6, Character: 4}, End: protocol.Position{Line: 6, Character: 9}},
		{Start: protocol.Position{Line: 9, Character: 15}, End: protocol.Position{Line: 9, Character: 24}},
		{Start: protocol.Position{Line: 12, Character: 4}, End: protocol.Position{Line: 12, Character: 10}},
	}, results)
}
