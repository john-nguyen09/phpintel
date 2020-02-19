package analysis

import (
	"io/ioutil"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/stretchr/testify/assert"
)

func indexTestCase(store *Store, uri string, path string, isOpen bool) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	document := NewDocument(uri, string(data))
	document.Load()
	if isOpen {
		document.Open()
	}
	store.saveDocOnStore(document)
	store.SyncDocument(document)
}

func TestFunctionCompletionWithNamespace(t *testing.T) {
	store, err := setupStore("test", "TestFunctionCompletionWithNamespace")
	if err != nil {
		panic(err)
	}
	indexTestCase(store, "test1", "../cases/function.php", false)
	indexTestCase(store, "test2", "../cases/completion/functionCompletionWithNamespace.php", true)
	document := store.GetOrCreateDocument("test2")
	importTable := document.GetImportTable()
	word := "testF"
	functions, _ := store.SearchFunctions(word, NewSearchOptions())
	items := []protocol.CompletionItem{}
	for _, function := range functions {
		label, textEdit := importTable.ResolveToQualified(document, function, function.Name, word)
		textEdits := []protocol.TextEdit{}
		if textEdit != nil {
			textEdits = append(textEdits, *textEdit)
		}
		items = append(items, protocol.CompletionItem{
			Kind:                protocol.FunctionCompletion,
			Label:               label,
			AdditionalTextEdits: textEdits,
		})
	}
	cupaloy.SnapshotT(t, items)
}

func TestDesignatorAndVariable(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/completion/designatorAndVariable.php")
	if err != nil {
		panic(err)
	}
	document := NewDocument("test1", string(data))
	document.Load()
	symbol := document.HasTypesAtPos(protocol.Position{
		Line:      9,
		Character: 20,
	})
	cupaloy.SnapshotT(t, symbol)
}

func TestCompletionWithScope(t *testing.T) {
	t.Run("Class", func(t *testing.T) {
		store, err := setupStore("Class", "CompletionWithScope-Class")
		defer store.Close()
		assert.NoError(t, err)
		defDoc1 := NewDocument("test1", `<?php
namespace Namespace1;
class TestClass {}`)
		defDoc1.Load()
		store.saveDocOnStore(defDoc1)
		store.SyncDocument(defDoc1)

		defDoc2 := NewDocument("test2", `<?php class TestClass {}`)
		defDoc2.Load()
		store.saveDocOnStore(defDoc2)
		store.SyncDocument(defDoc2)

		classes, _ := store.SearchClasses("\\Namespace1\\Te", NewSearchOptions())
		assert.Equal(t, 1, len(classes))
		names := []string{}
		for _, class := range classes {
			names = append(names, class.Name.GetFQN())
		}
		assert.Equal(t, []string{"\\Namespace1\\TestClass"}, names)
	})
	t.Run("ClassConst", func(t *testing.T) {
		store, err := setupStore("ClassConst", "CompletionWithScope-ClassConst")
		defer store.Close()
		assert.NoError(t, err)
		defDoc1 := NewDocument("test1", `<?php
class TestClass1 {
	const CLASS_CONST = 1;
}
class TestClass2 {
	const CLASS_CONST = 2;
}`)
		defDoc1.Load()
		store.saveDocOnStore(defDoc1)
		store.SyncDocument(defDoc1)

		classConsts, _ := store.SearchClassConsts("\\TestClass1", "CL", NewSearchOptions())
		assert.Equal(t, 1, len(classConsts))
		names := []string{}
		for _, classConst := range classConsts {
			names = append(names, classConst.Name)
		}
		assert.Equal(t, []string{"CLASS_CONST"}, names)
	})
}
