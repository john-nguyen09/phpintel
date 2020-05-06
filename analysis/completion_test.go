package analysis

import (
	"io/ioutil"
	"reflect"
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
	document := NewDocument(uri, data)
	document.Load()
	if isOpen {
		document.Open()
	}
	store.saveDocOnStore(document)
	store.SyncDocument(document)
}

func TestFunctionCompletionWithNamespace(t *testing.T) {
	store := setupStore("test", "TestFunctionCompletionWithNamespace")
	indexTestCase(store, "test1", "../cases/function.php", false)
	indexTestCase(store, "test2", "../cases/completion/functionCompletionWithNamespace.php", true)
	document := store.GetOrCreateDocument("test2")
	importTable := document.currImportTable()
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

func TestCompletionWithScope(t *testing.T) {
	t.Run("Class", func(t *testing.T) {
		store := setupStore("Class", "CompletionWithScope-Class")
		defDoc1 := NewDocument("test1", []byte(`<?php
namespace Namespace1;
class TestClass {}`))
		defDoc1.Load()
		store.saveDocOnStore(defDoc1)
		store.SyncDocument(defDoc1)

		defDoc2 := NewDocument("test2", []byte(`<?php class TestClassABC {}`))
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
		store := setupStore("ClassConst", "CompletionWithScope-ClassConst")
		defDoc1 := NewDocument("test1", []byte(`<?php
class TestClass1 {
	const CLASS_CONST = 1;
}
class TestClass2 {
	const CLASS_CONST_ABC = 2;
}`))
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
	t.Run("Method", func(t *testing.T) {
		store := setupStore("Method", "CompletionWithScope-Method")
		defDoc1 := NewDocument("test1", []byte(`<?php
class TestClass1 { public function methodABC(); }
class TestClass2 { public function method(); }`))
		defDoc1.Load()
		store.saveDocOnStore(defDoc1)
		store.SyncDocument(defDoc1)

		methods, _ := store.SearchMethods("\\TestClass2", "me", NewSearchOptions())
		assert.Equal(t, 1, len(methods))
		names := []string{}
		for _, method := range methods {
			names = append(names, method.Name)
		}
		assert.Equal(t, []string{"method"}, names)
	})
}

func TestMemberAccess(t *testing.T) {
	data, _ := ioutil.ReadFile("../cases/completion/memberAccess.php")
	doc := NewDocument("test1", data)
	doc.Load()

	symbol := doc.HasTypesAtPos(protocol.Position{Line: 6, Character: 16})
	assert.Equal(t, "*analysis.PropertyAccess", reflect.TypeOf(symbol).String())
}
