package analysis

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/karrick/godirwalk"
	"github.com/stretchr/testify/assert"
)

func TestStore(t *testing.T) {
	classTest := "../cases/class.php"
	data, err := ioutil.ReadFile(classTest)
	if err != nil {
		panic(err)
	}

	document := NewDocument("test1", data)
	document.Load()
	store := setupStore("test", "TestStore")
	store.SyncDocument(document)
	classes := store.GetClasses("\\TestClass1")
	cupaloy.Snapshot(classes)
}

func TestSearchNamespace(t *testing.T) {
	store := setupStore("test", "TestSearchNamespace")
	doc1 := NewDocument("test1", []byte(`<?php namespace Namespace1;`))
	doc1.Load()
	store.SyncDocument(doc1)
	doc2 := NewDocument("test2", []byte(`<?php namespace Namespace2;`))
	doc2.Load()
	store.SyncDocument(doc2)
	doc3 := NewDocument("test3", []byte(`<?php namespace AnotherNamespace3;`))
	doc3.Load()
	store.SyncDocument(doc3)

	doc4 := NewDocument("test4", []byte(`<?php namespace A\B\ class Test {}`))
	doc4.Load()
	store.SyncDocument(doc4)
	doc4.hasChanges = true
	doc4.SetText([]byte(`<?php namespace A\B; class Test {}`))
	doc4.Load()
	store.SyncDocument(doc4)

	namespaces, _ := store.SearchNamespaces("\\Name", NewSearchOptions())
	expected := []string{
		"\\Namespace1",
		"\\Namespace2",
		"\\AnotherNamespace3",
	}
	for _, e := range expected {
		assert.Contains(t, namespaces, e)
	}

	deletedKeys := []string{
		namespaceCompletionIndex + KeySep + "" + KeySep + "a" + KeySep + "\\A\\B\\class" + KeySep + "0",
		namespaceCompletionIndex + KeySep + "\\A" + KeySep + "b" + KeySep + "\\A\\B\\class" + KeySep + "1",
		namespaceCompletionIndex + KeySep + "\\A\\B" + KeySep + "class" + KeySep + "\\A\\B\\class" + KeySep + "2",
	}
	for _, key := range deletedKeys {
		b, _ := store.db.Get([]byte(key))
		assert.Equal(t, []byte(nil), b)
	}
	namespaces, _ = store.SearchNamespaces("\\A\\B", NewSearchOptions())
	assert.NotContains(t, namespaces, "\\A\\B\\class")
}

type getClassesByScopeTestCase struct {
	scope        string
	expectedFQNs []string
}

func TestGetClassesByScope(t *testing.T) {
	store := setupStore("test", "TestGetClassesByScope")
	doc1 := NewDocument("test1", []byte(`<?php
namespace Namespace1 {
	class Class1UnderNamespace1 {}
	class Class2UnderNamespace1 {}
}
namespace Namespace2 {
	class Class1UnderNamespace2 {}
	class Class2UnderNamespace2 {}
}`))
	doc1.Load()
	store.SyncDocument(doc1)

	testCases := []getClassesByScopeTestCase{
		{"\\Namespace1", []string{"\\Namespace1\\Class1UnderNamespace1", "\\Namespace1\\Class2UnderNamespace1"}},
		{"\\Namespace2", []string{"\\Namespace2\\Class1UnderNamespace2", "\\Namespace2\\Class2UnderNamespace2"}},
	}
	for _, testCase := range testCases {
		classFQNs := []string{}
		store.GetClassesByScopeStream(testCase.scope, func(class *Class) onDataResult {
			classFQNs = append(classFQNs, class.Name.GetFQN())
			return onDataResult{false}
		})
		if assert.Equal(t, len(testCase.expectedFQNs), len(classFQNs)) {
			for _, fqn := range testCase.expectedFQNs {
				assert.Contains(t, classFQNs, fqn)
			}
		}
	}
}

func TestStoreClose(t *testing.T) {
	store := setupStore("test", "TestStoreClose")
	filePaths := []string{}
	godirwalk.Walk("cases", &godirwalk.Options{
		Callback: func(path string, de *godirwalk.Dirent) error {
			if !de.ModeType().IsDir() && strings.HasSuffix(path, ".php") {
				filePaths = append(filePaths, path)
			}
			return nil
		},
		Unsorted: true,
	})
	for id, filePath := range filePaths {
		data, _ := ioutil.ReadFile(filePath)
		document := NewDocument("test"+string(id), data)
		document.Load()
		store.SyncDocument(document)
	}
	store.fEngine.close()
}
