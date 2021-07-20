package analysis

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
	"github.com/stretchr/testify/assert"
)

type mockMemberAccess struct {
	scopeTypes TypeComposite
	scopeName  string
}

var _ MemberAccess = (*mockMemberAccess)(nil)

func (m *mockMemberAccess) GetLocation() protocol.Location {
	return protocol.Location{}
}

func (m *mockMemberAccess) GetTypes() TypeComposite {
	return newTypeComposite()
}

func (m *mockMemberAccess) Resolve(ctx ResolveContext) {

}

func (m *mockMemberAccess) ScopeName() string {
	return m.scopeName
}

func (m *mockMemberAccess) ScopeTypes() TypeComposite {
	return m.scopeTypes
}

func TestClassConstInheritance(t *testing.T) {
	withTestStore("test", t.Name(), func(store *Store) {
		data, err := ioutil.ReadFile("../cases/inheritedClassConst.php")
		if err != nil {
			panic(err)
		}
		doc := NewDocument("test1", data)
		doc.Load()
		store.SyncDocument(doc)
		q := NewQuery(store)

		test := func(memberName string, context string, typ string, scopeName string, loc protocol.Location) {
			typs := strings.Split(typ, "|")
			types := newTypeComposite()
			classes := []*Class{}
			for _, typ := range typs {
				types.add(NewTypeString(typ))
				classes = append(classes, q.GetClasses(typ)...)
			}
			ccs := EmptyInheritedClassConst()
			for _, class := range classes {
				ccs.Merge(q.GetClassClassConsts(class, memberName, ccs.SearchedFQNs))
			}
			classConsts := ccs.ReduceStatic(context, &mockMemberAccess{
				scopeTypes: types,
				scopeName:  scopeName,
			})
			assert.Equal(t, 1, len(classConsts))
			assert.Equal(t, loc, classConsts[0].Const.location)
		}

		testMultiple := func(memberName string, context string, typ string, scopeName string, locs []protocol.Location) {
			typs := strings.Split(typ, "|")
			types := newTypeComposite()
			for _, typ := range typs {
				types.add(NewTypeString(typ))
			}
			classConsts := []ClassConstWithScope{}
			for _, ts := range types.Resolve() {
				ccs := EmptyInheritedClassConst()
				for _, class := range q.GetClasses(ts.GetFQN()) {
					ccs.Merge(q.GetClassClassConsts(class, memberName, ccs.SearchedFQNs))
				}
				classConsts = MergeClassConstWithScope(classConsts, ccs.ReduceStatic(context, &mockMemberAccess{
					scopeTypes: types,
					scopeName:  scopeName,
				}))
			}
			assert.Equal(t, len(locs), len(classConsts))
			results := []protocol.Location{}
			for _, classConst := range classConsts {
				results = append(results, classConst.Const.location)
			}
			assert.Equal(t, locs, results)
		}

		test("BASE_CONST", "\\ExtendedClass", "\\BaseClass", "parent",
			protocol.Location{URI: "test1", Range: protocol.Range{
				Start: protocol.Position{Line: 4, Character: 10}, End: protocol.Position{Line: 4, Character: 30},
			}})

		test("BASE_CONST", "\\ExtendedClass", "\\ExtendedClass", "static",
			protocol.Location{URI: "test1", Range: protocol.Range{
				Start: protocol.Position{Line: 11, Character: 10}, End: protocol.Position{Line: 11, Character: 34},
			}})

		test("BASE_CONST", "\\ExtendedClass", "\\ExtendedClass", "self",
			protocol.Location{URI: "test1", Range: protocol.Range{
				Start: protocol.Position{Line: 11, Character: 10}, End: protocol.Position{Line: 11, Character: 34},
			}})

		test("PRIVATE_CONST", "\\ExtendedClass", "\\ExtendedClass", "self",
			protocol.Location{URI: "test1", Range: protocol.Range{
				Start: protocol.Position{Line: 12, Character: 18}, End: protocol.Position{Line: 12, Character: 57},
			}})

		test("PRIVATE_CONST", "\\ExtendedClass", "\\ExtendedClass", "$instance",
			protocol.Location{URI: "test1", Range: protocol.Range{
				Start: protocol.Position{Line: 12, Character: 18}, End: protocol.Position{Line: 12, Character: 57},
			}})

		test("PRIVATE_CONST", "\\ExtendedClass", "\\ExtendedClass", "$this",
			protocol.Location{URI: "test1", Range: protocol.Range{
				Start: protocol.Position{Line: 12, Character: 18}, End: protocol.Position{Line: 12, Character: 57},
			}})

		test("BASE_PROTECTED_CONST", "\\ExtendedClass", "\\BaseClass", "$base",
			protocol.Location{URI: "test1", Range: protocol.Range{
				Start: protocol.Position{Line: 5, Character: 20}, End: protocol.Position{Line: 5, Character: 65},
			}})

		test("BASE_CONST", "", "\\ExtendedClass", "ExtendedClass",
			protocol.Location{URI: "test1", Range: protocol.Range{
				Start: protocol.Position{Line: 11, Character: 10}, End: protocol.Position{Line: 11, Character: 34},
			}})

		test("EXTENDED_CONST", "", "\\ExtendedClass", "ExtendedClass",
			protocol.Location{URI: "test1", Range: protocol.Range{
				Start: protocol.Position{Line: 13, Character: 10}, End: protocol.Position{Line: 13, Character: 34},
			}})

		test("INHERITED_CONST", "", "\\ExtendedClass", "ExtendedClass",
			protocol.Location{URI: "test1", Range: protocol.Range{
				Start: protocol.Position{Line: 6, Character: 10}, End: protocol.Position{Line: 6, Character: 45},
			}})

		test("BASE_CONST", "", "\\BaseClass", "BaseClass",
			protocol.Location{URI: "test1", Range: protocol.Range{
				Start: protocol.Position{Line: 4, Character: 10}, End: protocol.Position{Line: 4, Character: 30},
			}})

		testMultiple("BASE_CONST", "", "\\BaseClass|\\ExtendedClass", "$var",
			[]protocol.Location{
				{URI: "test1", Range: protocol.Range{
					Start: protocol.Position{Line: 4, Character: 10}, End: protocol.Position{Line: 4, Character: 30},
				}},
				{URI: "test1", Range: protocol.Range{
					Start: protocol.Position{Line: 11, Character: 10}, End: protocol.Position{Line: 11, Character: 34},
				}},
			})

		testMultiple("INHERITED_CONST", "", "\\BaseClass|\\ExtendedClass", "$var",
			[]protocol.Location{
				{URI: "test1", Range: protocol.Range{
					Start: protocol.Position{Line: 6, Character: 10}, End: protocol.Position{Line: 6, Character: 45},
				}},
			})
	})
}

func TestMethodsInheritance(t *testing.T) {
	withTestStore("test", t.Name(), func(store *Store) {
		data, err := ioutil.ReadFile("../cases/inheritedMethod.php")
		if err != nil {
			panic(err)
		}
		doc := NewDocument("test1", data)
		doc.Load()
		store.SyncDocument(doc)
		q := NewQuery(store)

		test := func(memberName string, context string, typ string, scopeName string, loc protocol.Location, isReduceStatic bool) {
			typs := strings.Split(typ, "|")
			types := newTypeComposite()
			classes := []*Class{}
			for _, typ := range typs {
				types.add(NewTypeString(typ))
				classes = append(classes, q.GetClasses(typ)...)
			}
			ms := EmptyInheritedMethods()
			for _, class := range classes {
				ms.Merge(q.GetClassMethods(class, memberName, ms.SearchedFQNs))
			}
			var methods []MethodWithScope
			mockAccess := &mockMemberAccess{
				scopeTypes: types,
				scopeName:  scopeName,
			}
			if isReduceStatic {
				methods = ms.ReduceStatic(context, mockAccess)
			} else {
				methods = ms.ReduceAccess(context, mockAccess)
			}
			assert.Equal(t, 1, len(methods))
			assert.Equal(t, loc, methods[0].Method.location)
		}

		testMultiple := func(memberName string, context string, typ string, scopeName string, locs []protocol.Location, isReduceStatic bool) {
			typs := strings.Split(typ, "|")
			types := newTypeComposite()
			for _, typ := range typs {
				types.add(NewTypeString(typ))
			}
			methods := []MethodWithScope{}
			for _, ts := range types.Resolve() {
				ms := EmptyInheritedMethods()
				for _, class := range q.GetClasses(ts.GetFQN()) {
					ms.Merge(q.GetClassMethods(class, memberName, ms.SearchedFQNs))
				}
				var newMethods []MethodWithScope
				mockAccess := &mockMemberAccess{
					scopeTypes: types,
					scopeName:  scopeName,
				}
				if isReduceStatic {
					newMethods = ms.ReduceStatic(context, mockAccess)
				} else {
					newMethods = ms.ReduceAccess(context, mockAccess)
				}
				methods = MergeMethodWithScope(methods, newMethods)
			}
			assert.Equal(t, len(locs), len(methods))
			results := []protocol.Location{}
			for _, method := range methods {
				results = append(results, method.Method.location)
			}
			assert.Equal(t, locs, results)
		}

		test("method2", "", "\\ExtendedClass", "$instance", protocol.Location{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 32, Character: 4}, End: protocol.Position{Line: 44, Character: 5},
		}}, false)

		test("baseStaticMethod", "", "\\ExtendedClass", "ExtendedClass", protocol.Location{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 56, Character: 4}, End: protocol.Position{Line: 59, Character: 5},
		}}, true)

		test("baseStaticMethod", "", "\\BaseClass", "BaseClass", protocol.Location{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 16, Character: 4}, End: protocol.Position{Line: 19, Character: 5},
		}}, true)

		test("baseStaticMethod", "\\ExtendedClass", "\\ExtendedClass", "static", protocol.Location{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 56, Character: 4}, End: protocol.Position{Line: 59, Character: 5},
		}}, true)

		test("baseStaticMethod", "\\ExtendedClass", "\\ExtendedClass", "self", protocol.Location{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 56, Character: 4}, End: protocol.Position{Line: 59, Character: 5},
		}}, true)

		test("baseStaticMethod", "\\ExtendedClass", "\\BaseClass", "parent", protocol.Location{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 16, Character: 4}, End: protocol.Position{Line: 19, Character: 5},
		}}, true)

		test("baseStaticMethod", "\\ExtendedClass", "\\ExtendedClass", "$this", protocol.Location{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 56, Character: 4}, End: protocol.Position{Line: 59, Character: 5},
		}}, true)

		testMultiple("baseMethod", "", "\\ExtendedClass|\\BaseClass", "$var", []protocol.Location{
			{URI: "test1", Range: protocol.Range{
				Start: protocol.Position{Line: 61, Character: 4}, End: protocol.Position{Line: 64, Character: 5},
			}},
			{URI: "test1", Range: protocol.Range{
				Start: protocol.Position{Line: 20, Character: 4}, End: protocol.Position{Line: 23, Character: 5},
			}},
		}, false)

		testMultiple("baseMethod", "", "\\ExtendedClass|\\BaseClass", "$var", []protocol.Location{
			{URI: "test1", Range: protocol.Range{
				Start: protocol.Position{Line: 61, Character: 4}, End: protocol.Position{Line: 64, Character: 5},
			}},
			{URI: "test1", Range: protocol.Range{
				Start: protocol.Position{Line: 20, Character: 4}, End: protocol.Position{Line: 23, Character: 5},
			}},
		}, false)
	})
}

func TestPropsInheritance(t *testing.T) {
	data, err := ioutil.ReadFile("../cases/inheritedProps.php")
	if err != nil {
		panic(err)
	}
	withTestStore("test", t.Name(), func(store *Store) {
		doc := NewDocument("test1", data)
		doc.Load()
		store.SyncDocument(doc)
		q := NewQuery(store)

		test := func(memberName string, context string, typ string, scopeName string, loc protocol.Location, isReduceStatic bool) {
			typs := strings.Split(typ, "|")
			types := newTypeComposite()
			classes := []*Class{}
			for _, typ := range typs {
				types.add(NewTypeString(typ))
				classes = append(classes, q.GetClasses(typ)...)
			}
			ps := EmptyInheritedProps()
			for _, class := range classes {
				ps.Merge(q.GetClassProps(class, memberName, ps.SearchedFQNs))
			}
			var props []PropWithScope
			mockAccess := &mockMemberAccess{
				scopeTypes: types,
				scopeName:  scopeName,
			}
			if isReduceStatic {
				props = ps.ReduceStatic(context, mockAccess)
			} else {
				props = ps.ReduceAccess(context, mockAccess)
			}
			assert.Equal(t, 1, len(props))
			assert.Equal(t, loc, props[0].Prop.location)
		}

		test("$basePublicStatic", "", "\\ExtendedClass", "$instance", protocol.Location{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 16, Character: 18}, End: protocol.Position{Line: 16, Character: 68},
		}}, true)

		test("$basePublic", "", "\\ExtendedClass", "$instance", protocol.Location{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 17, Character: 11}, End: protocol.Position{Line: 17, Character: 49},
		}}, false)

		test("$basePublic2", "", "\\ExtendedClass", "$instance", protocol.Location{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 10, Character: 11}, End: protocol.Position{Line: 10, Character: 41},
		}}, false)

		test("$basePublic", "", "\\BaseClass", "$base", protocol.Location{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 9, Character: 11}, End: protocol.Position{Line: 9, Character: 39},
		}}, false)

		test("$baseProtectedStatic", "\\ExtendedClass", "\\BaseClass", "parent", protocol.Location{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 4, Character: 21}, End: protocol.Position{Line: 4, Character: 67},
		}}, true)

		test("$basePublicStatic", "\\ExtendedClass", "\\BaseClass", "parent", protocol.Location{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 5, Character: 18}, End: protocol.Position{Line: 5, Character: 58},
		}}, true)

		test("$baseProtected", "\\ExtendedClass", "\\ExtendedClass", "$this", protocol.Location{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 15, Character: 14}, End: protocol.Position{Line: 15, Character: 58},
		}}, false)

		test("$baseProtected2", "\\ExtendedClass", "\\ExtendedClass", "$this", protocol.Location{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 7, Character: 14}, End: protocol.Position{Line: 7, Character: 50},
		}}, false)

		test("$basePublicStatic", "\\ExtendedClass", "\\ExtendedClass", "static", protocol.Location{URI: "test1", Range: protocol.Range{
			Start: protocol.Position{Line: 16, Character: 18}, End: protocol.Position{Line: 16, Character: 68},
		}}, true)
	})
}
