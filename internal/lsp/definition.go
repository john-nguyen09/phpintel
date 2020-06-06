package lsp

import (
	"context"
	"strings"

	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

func (s *Server) definition(ctx context.Context, params *protocol.DefinitionParams) ([]protocol.Location, error) {
	locations := []protocol.Location{}
	uri := params.TextDocumentPositionParams.TextDocument.URI
	store := s.store.getStore(uri)
	if store == nil {
		return nil, StoreNotFound(uri)
	}
	document := store.GetOrCreateDocument(uri)
	if document == nil {
		return nil, DocumentNotFound(uri)
	}
	document.Load()
	q := analysis.NewQuery(store)
	resolveCtx := analysis.NewResolveContext(q, document)
	pos := params.TextDocumentPositionParams.Position
	symbol := document.HasTypesAtPos(pos)

	switch v := symbol.(type) {
	case *analysis.ClassTypeDesignator:
		for _, typeString := range v.Type.Resolve() {
			for _, theClass := range q.GetClasses(document.ImportTableAtPos(pos).GetClassReferenceFQN(typeString)) {
				constructor := q.GetClassConstructor(theClass)
				if constructor.Method == nil || constructor.Method.GetScope() != theClass.Name.GetFQN() {
					locations = append(locations, theClass.GetLocation())
				} else {
					locations = append(locations, constructor.Method.GetLocation())
				}
			}
		}
	case *analysis.ClassAccess:
		for _, typeString := range v.Type.Resolve() {
			for _, theClass := range q.GetClasses(document.ImportTableAtPos(pos).GetClassReferenceFQN(typeString)) {
				locations = append(locations, theClass.GetLocation())
			}
		}
	case *analysis.InterfaceAccess:
		for _, typeString := range v.Type.Resolve() {
			for _, theInterface := range q.GetInterfaces(document.ImportTableAtPos(pos).GetClassReferenceFQN(typeString)) {
				locations = append(locations, theInterface.GetLocation())
			}
		}
	case *analysis.TraitAccess:
		for _, typeString := range v.GetTypes().Resolve() {
			for _, trait := range q.GetTraits(typeString.GetFQN()) {
				locations = append(locations, trait.GetLocation())
			}
		}
	case *analysis.ConstantAccess:
		name := analysis.NewTypeString(v.Name)
		for _, theConst := range q.GetConsts(document.ImportTableAtPos(pos).GetConstReferenceFQN(q, name)) {
			locations = append(locations, theConst.GetLocation())
		}
		for _, define := range q.GetDefines(document.ImportTableAtPos(pos).GetConstReferenceFQN(q, name)) {
			locations = append(locations, define.GetLocation())
		}
	case *analysis.FunctionCall:
		name := analysis.NewTypeString(v.Name)
		for _, function := range q.GetFunctions(document.ImportTableAtPos(pos).GetFunctionReferenceFQN(q, name)) {
			locations = append(locations, function.GetLocation())
		}
	case *analysis.ScopedConstantAccess:
		currentClass := document.GetClassScopeAtSymbol(v)
		var classConsts []analysis.ClassConstWithScope
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			ccs := analysis.EmptyInheritedClassConst()
			for _, class := range q.GetClasses(scopeType.GetFQN()) {
				ccs.Merge(q.GetClassClassConsts(class, v.Name, ccs.SearchedFQNs))
			}
			for _, intf := range q.GetInterfaces(scopeType.GetFQN()) {
				ccs.Merge(q.GetInterfaceClassConsts(intf, v.Name, ccs.SearchedFQNs))
			}
			classConsts = analysis.MergeClassConstWithScope(classConsts, ccs.ReduceStatic(currentClass, v))
		}
		for _, c := range classConsts {
			locations = append(locations, c.Const.GetLocation())
		}
	case *analysis.ScopedMethodAccess:
		currentClass := document.GetClassScopeAtSymbol(v)
		var methods []analysis.MethodWithScope
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			ms := analysis.EmptyInheritedMethods()
			for _, class := range q.GetClasses(scopeType.GetFQN()) {
				ms.Merge(q.GetClassMethods(class, v.Name, ms.SearchedFQNs))
			}
			methods = analysis.MergeMethodWithScope(methods, ms.ReduceStatic(currentClass, v))
		}
		for _, m := range methods {
			locations = append(locations, m.Method.GetLocation())
		}
	case *analysis.ScopedPropertyAccess:
		currentClass := document.GetClassScopeAtSymbol(v.Scope)
		var props []analysis.PropWithScope
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			ps := analysis.EmptyInheritedProps()
			for _, class := range q.GetClasses(scopeType.GetFQN()) {
				ps.Merge(q.GetClassProps(class, v.Name, ps.SearchedFQNs))
			}
			props = analysis.MergePropWithScope(props, ps.ReduceStatic(currentClass, v))
		}
		for _, p := range props {
			locations = append(locations, p.Prop.GetLocation())
		}
	case *analysis.PropertyAccess:
		currentClass := document.GetClassScopeAtSymbol(v.Scope)
		var props []analysis.PropWithScope
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			ps := analysis.EmptyInheritedProps()
			for _, class := range q.GetClasses(scopeType.GetFQN()) {
				ps.Merge(q.GetClassProps(class, "$"+v.Name, ps.SearchedFQNs))
			}
			props = analysis.MergePropWithScope(props, ps.ReduceAccess(currentClass, v))
		}
		for _, p := range props {
			locations = append(locations, p.Prop.GetLocation())
		}
	case *analysis.MethodAccess:
		currentClass := document.GetClassScopeAtSymbol(v.Scope)
		var methods []analysis.MethodWithScope
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			ms := analysis.EmptyInheritedMethods()
			for _, class := range q.GetClasses(scopeType.GetFQN()) {
				ms.Merge(q.GetClassMethods(class, v.Name, ms.SearchedFQNs))
			}
			for _, theInterface := range q.GetInterfaces(scopeType.GetFQN()) {
				ms.Merge(q.GetInterfaceMethods(theInterface, v.Name, ms.SearchedFQNs))
			}
			for _, trait := range q.GetTraits(scopeType.GetFQN()) {
				ms.Merge(q.GetTraitMethods(trait, v.Name))
			}
			methods = analysis.MergeMethodWithScope(methods, ms.ReduceAccess(currentClass, v))
		}
		for _, m := range methods {
			locations = append(locations, m.Method.GetLocation())
		}
	case *analysis.TypeDeclaration:
		for _, typeString := range v.Type.Resolve() {
			classes := q.GetClasses(typeString.GetFQN())
			for _, class := range classes {
				locations = append(locations, class.GetLocation())
			}
			interfaces := q.GetInterfaces(typeString.GetFQN())
			for _, theInterface := range interfaces {
				locations = append(locations, theInterface.GetLocation())
			}
		}
	}
	filteredLocations := locations[:0]
	for _, location := range locations {
		if strings.HasPrefix(location.URI, "file://") {
			filteredLocations = append(filteredLocations, location)
		}
	}
	return filteredLocations, nil
}
