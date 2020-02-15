package analysis

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
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

func TestNamespaceAndUse(t *testing.T) {
	doc1 := NewDocument("importTable1", `<?php
namespace TestNamespace1;

class TestClass1 {}

function TestFunction1() {}`)
	doc1.Load()

	doc2 := NewDocument("importTable2", `<?php
`)
	doc2.Load()

	class := doc1.Children[0].(*Class)
	function := doc1.Children[1].(*Function)
	importTable := doc2.GetImportTable()
	label, edit := importTable.ResolveToQualified(doc2, class, class.Name, "Test")
	if label != "TestClass1" || strings.Index(edit.NewText, "use TestNamespace1\\TestClass1;") == -1 {
		t.Errorf("Incorrect: %s %v", label, edit)
	}
	label, edit = importTable.ResolveToQualified(doc2, function, function.Name, "Function")
	if label != "TestFunction1" || strings.Index(edit.NewText, "use function TestNamespace1\\TestFunction1;") == -1 {
		t.Errorf("Incorrect: %s %v", label, edit)
	}
	label, edit = importTable.ResolveToQualified(doc2, function, function.Name, "TestNamespace1\\t")
	if label != "TestFunction1" || edit != nil {
		t.Errorf("Incorrect: %s %v", label, edit)
	}
}
