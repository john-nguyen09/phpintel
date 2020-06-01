package analysis

import (
	"testing"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/stretchr/testify/assert"
)

func TestQuery(t *testing.T) {
	store := setupStore("test", t.Name())
	indexDocument(store, "../cases/class.php", "test1")
	indexDocument(store, "../cases/classConst.php", "test2")
	indexDocument(store, "../cases/const.php", "test3")
	indexDocument(store, "../cases/function.php", "test4")
	indexDocument(store, "../cases/interface.php", "test5")
	indexDocument(store, "../cases/method.php", "test6")
	indexDocument(store, "../cases/property.php", "test7")

	t.Run("Class", func(t *testing.T) {
		q := NewQuery(store)
		classes := q.GetClasses("\\TestClass")
		assert.Equal(t, 1, len(classes))
		classes = q.GetClasses("\\TestClass")
		assert.Equal(t, 1, len(classes))
	})

	t.Run("ClassConst", func(t *testing.T) {
		q := NewQuery(store)
		classConsts := q.GetClassConsts("\\ClassConstTest1", "CLASS_CONST1")
		assert.Equal(t, 1, len(classConsts))
		classConsts = q.GetClassConsts("\\ClassConstTest1", "CLASS_CONST1")
		assert.Equal(t, 1, len(classConsts))
	})

	t.Run("Const", func(t *testing.T) {
		q := NewQuery(store)
		consts := q.GetConsts("\\TEST_CONST1")
		assert.Equal(t, 1, len(consts))
		consts = q.GetConsts("\\TEST_CONST1")
		assert.Equal(t, 1, len(consts))

		defines := q.GetDefines("\\TEST_CONST2")
		assert.Equal(t, 1, len(defines))
		defines = q.GetDefines("\\TEST_CONST2")
		assert.Equal(t, 1, len(defines))
	})

	t.Run("Function", func(t *testing.T) {
		q := NewQuery(store)
		functions := q.GetFunctions("\\testFunction")
		assert.Equal(t, 1, len(functions))
		functions = q.GetFunctions("\\testFunction")
		assert.Equal(t, 1, len(functions))
	})

	t.Run("Interface", func(t *testing.T) {
		q := NewQuery(store)
		interfaces := q.GetInterfaces("\\TestInterface")
		assert.Equal(t, 1, len(interfaces))
		interfaces = q.GetInterfaces("\\TestInterface")
		assert.Equal(t, 1, len(interfaces))
	})

	t.Run("Method", func(t *testing.T) {
		doc1 := NewDocument("test/Method1", []byte(`<?php
interface Interface111 {
	public function testMethod3();
}
class InheritedClass1 extends TestAbstractMethodClass {}
class InheritedClass2 extends TestAbstractMethodClass { private function testMethod10() {} }
class InheritedClass3 implements Interface111 {}
class InheritedClass4 extends TestMethodClass implements Interface111 {}
class InheritedClass5 extends TestMethodClass implements Interface111 {
	public function testMethod3() {}
}`))
		doc1.Load()
		store.SyncDocument(doc1)
		q := NewQuery(store)
		for _, class := range q.GetClasses("\\InheritedClass1") {
			result := q.GetClassMethods(class, "testMethod10", nil)
			assert.Equal(t, 1, result.Len())
			assert.Equal(t, protocol.Location{
				URI: "test6",
				Range: protocol.Range{
					Start: protocol.Position{Line: 41, Character: 4},
					End:   protocol.Position{Line: 41, Character: 45},
				},
			}, result.Methods[0].Method.location)
			methods := result.ReduceInherited()
			assert.Equal(t, 1, len(methods))
			assert.Equal(t, protocol.Location{
				URI: "test6",
				Range: protocol.Range{
					Start: protocol.Position{Line: 41, Character: 4},
					End:   protocol.Position{Line: 41, Character: 45},
				},
			}, methods[0].location)
		}
		for _, class := range q.GetClasses("\\InheritedClass2") {
			result := q.GetClassMethods(class, "testMethod10", nil)
			assert.Equal(t, 2, result.Len())
			var results []protocol.Location
			for _, method := range result.Methods {
				results = append(results, method.Method.location)
			}
			assert.Equal(t, []protocol.Location{
				{URI: "test/Method1", Range: protocol.Range{
					Start: protocol.Position{Line: 5, Character: 56},
					End:   protocol.Position{Line: 5, Character: 90},
				}},
				{URI: "test6", Range: protocol.Range{
					Start: protocol.Position{Line: 41, Character: 4},
					End:   protocol.Position{Line: 41, Character: 45},
				}},
			}, results)
			methods := result.ReduceInherited()
			assert.Equal(t, 1, len(methods))
			assert.Equal(t, protocol.Location{
				URI: "test/Method1", Range: protocol.Range{
					Start: protocol.Position{Line: 5, Character: 56},
					End:   protocol.Position{Line: 5, Character: 90},
				},
			}, methods[0].location)
		}
		for _, class := range q.GetClasses("\\InheritedClass3") {
			result := q.GetClassMethods(class, "testMethod3", nil)
			assert.Equal(t, 1, result.Len())
			assert.Equal(t, protocol.Location{
				URI: "test/Method1", Range: protocol.Range{
					Start: protocol.Position{Line: 2, Character: 1},
					End:   protocol.Position{Line: 2, Character: 31},
				},
			}, result.Methods[0].Method.location)
			methods := result.ReduceInherited()
			assert.Equal(t, 1, len(methods))
			assert.Equal(t, protocol.Location{
				URI: "test/Method1", Range: protocol.Range{
					Start: protocol.Position{Line: 2, Character: 1},
					End:   protocol.Position{Line: 2, Character: 31},
				},
			}, methods[0].location)
		}
		for _, class := range q.GetClasses("\\InheritedClass4") {
			result := q.GetClassMethods(class, "testMethod3", nil)
			assert.Equal(t, 2, result.Len())
			var results []protocol.Location
			for _, method := range result.Methods {
				results = append(results, method.Method.location)
			}
			assert.Equal(t, []protocol.Location{
				{URI: "test6", Range: protocol.Range{
					Start: protocol.Position{Line: 11, Character: 4},
					End:   protocol.Position{Line: 13, Character: 5},
				}},
				{URI: "test/Method1", Range: protocol.Range{
					Start: protocol.Position{Line: 2, Character: 1},
					End:   protocol.Position{Line: 2, Character: 31},
				}},
			}, results)
			methods := result.ReduceInherited()
			assert.Equal(t, 1, len(methods))
			assert.Equal(t, protocol.Location{URI: "test6", Range: protocol.Range{
				Start: protocol.Position{Line: 11, Character: 4},
				End:   protocol.Position{Line: 13, Character: 5},
			}}, methods[0].location)
		}
		for _, class := range q.GetClasses("\\InheritedClass5") {
			result := q.GetClassMethods(class, "testMethod3", nil)
			assert.Equal(t, 3, result.Len())
			var results []protocol.Location
			for _, method := range result.Methods {
				results = append(results, method.Method.location)
			}
			assert.Equal(t, []protocol.Location{
				{URI: "test/Method1", Range: protocol.Range{
					Start: protocol.Position{Line: 9, Character: 1},
					End:   protocol.Position{Line: 9, Character: 33},
				}},
				{URI: "test6", Range: protocol.Range{
					Start: protocol.Position{Line: 11, Character: 4},
					End:   protocol.Position{Line: 13, Character: 5},
				}},
				{URI: "test/Method1", Range: protocol.Range{
					Start: protocol.Position{Line: 2, Character: 1},
					End:   protocol.Position{Line: 2, Character: 31},
				}},
			}, results)
			methods := result.ReduceInherited()
			assert.Equal(t, 1, len(methods))
			assert.Equal(t, protocol.Location{URI: "test/Method1", Range: protocol.Range{
				Start: protocol.Position{Line: 9, Character: 1},
				End:   protocol.Position{Line: 9, Character: 33},
			}}, methods[0].location)
		}
		for _, class := range q.GetClasses("\\InheritedClass4") {
			methods := q.GetClassMethods(class, "testMethod6", nil).ReduceStatic("", "\\InheritedClass4")
			assert.Equal(t, 1, len(methods))
			assert.Equal(t, protocol.Location{URI: "test6", Range: protocol.Range{
				Start: protocol.Position{Line: 23, Character: 4},
				End:   protocol.Position{Line: 25, Character: 5},
			}}, methods[0].Method.location)
			methods = q.GetClassMethods(class, "testMethod6", nil).ReduceStatic("\\InheritedClass4", "parent")
			assert.Equal(t, 1, len(methods))
			assert.Equal(t, protocol.Location{URI: "test6", Range: protocol.Range{
				Start: protocol.Position{Line: 23, Character: 4},
				End:   protocol.Position{Line: 25, Character: 5},
			}}, methods[0].Method.location)
		}
	})
}

func TestDuplicatedMethodAccess(t *testing.T) {
	store := setupStore("test", t.Name())
	doc := NewDocument("test1", []byte(`<?php
class BaseClass { public function method1() {} }
class Extended1 extends BaseClass {}`))
	doc.Load()
	store.SyncDocument(doc)
	ctx := NewResolveContext(NewQuery(store), doc)
	types := newTypeComposite()
	types.add(NewTypeString("\\Extended1"))
	types.add(NewTypeString("\\BaseClass"))
	methods := InheritedMethods{}
	for _, typeString := range types.Resolve() {
		for _, class := range ctx.query.GetClasses(typeString.GetFQN()) {
			methods.Merge(ctx.query.GetClassMethods(class, "method1", methods.SearchedFQNs))
		}
	}
	assert.Equal(t, 1, methods.Len())
}
