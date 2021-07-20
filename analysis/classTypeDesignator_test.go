package analysis

import (
	"testing"

	"github.com/bradleyjkemp/cupaloy"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/stretchr/testify/assert"
)

func TestStaticClassTypeDesignator(t *testing.T) {
	doc1 := NewDocument("test1", []byte(`<?php
class Class1
{
    public function test1()
	{
        $var1 = new static();
        $var1->test2();
	}

    public function test2() { }
}`))
	doc1.Load()

	hasTypes := doc1.HasTypesAtPos(protocol.Position{
		Line:      6,
		Character: 11,
	})
	assert.Equal(t, "\\Class1", hasTypes.GetTypes().ToString())
}

func TestInheritedConstructor(t *testing.T) {
	withTestStore("test", t.Name(), func(store *Store) {
		doc := NewDocument("test1", []byte(`<?php
class BaseClass
{
	public function __construct() {}
}
class ExtendedClass extends BaseClass { }`))
		doc.Load()
		store.SyncDocument(doc)

		q := NewQuery(store)
		class := q.GetClasses("\\ExtendedClass")[0]
		method := q.GetClassConstructor(class)
		assert.Equal(t, protocol.Location{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 3, Character: 1},
			End:   protocol.Position{Line: 3, Character: 33},
		}}, method.Method.location)

		types := newTypeComposite()
		types.add(NewTypeString("\\ExtendedClass"))
		typeDesignator := ClassTypeDesignator{
			Expression: Expression{
				Type: types,
			},
		}
		hasParams := typeDesignator.ResolveToHasParams(NewResolveContext(q, doc))
		assert.Equal(t, 1, len(hasParams))
		assert.IsType(t, &Method{}, hasParams[0])
		assert.Equal(t, protocol.Location{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 3, Character: 1},
			End:   protocol.Position{Line: 3, Character: 33},
		}}, hasParams[0].(*Method).location)

		doc2 := NewDocument("test2", []byte(`<?php
	class BaseClass2 { private function __construct() {} }
	class BaseClass3 { protected function __construct() {} }
	class ExtendedClass2 extends BaseClass2 { }
	class ExtendedClass3 extends BaseClass3 { }`))
		doc2.Load()
		store.SyncDocument(doc2)
		class2 := q.GetClasses("\\ExtendedClass2")[0]
		method2 := q.GetClassConstructor(class2)
		assert.Nil(t, method2.Method)

		class3 := q.GetClasses("\\ExtendedClass3")[0]
		method3 := q.GetClassConstructor(class3)
		assert.Nil(t, method3.Method)
	})
}

func TestAnonymousClassConstructor(t *testing.T) {
	doc := NewDocument("test1", []byte(`<?php
class BaseClass {
	public function __construct($view, $helper) {}
}

new class() extends BaseClass {
}`))
	doc.Load()
	withTestStore("test", t.Name(), func(store *Store) {
		store.SyncDocument(doc)
		argumentList, hasParamsResolvable := doc.ArgumentListAndFunctionCallAt(protocol.Position{
			Line:      5,
			Character: 10,
		})
		resolveCtx := NewResolveContext(NewQuery(store), doc)
		hasParams := hasParamsResolvable.ResolveToHasParams(resolveCtx)
		cupaloy.SnapshotT(t, struct {
			argumentList *ArgumentList
			hasParams    []HasParams
		}{
			argumentList,
			hasParams,
		})
	})
}
