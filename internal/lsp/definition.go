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
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			for _, classConst := range q.GetClassConsts(scopeType.GetFQN(), v.Name) {
				locations = append(locations, classConst.GetLocation())
			}
		}
	case *analysis.ScopedMethodAccess:
		var scopeName string
		if hasName, ok := v.Scope.(analysis.HasName); ok {
			scopeName = hasName.GetName()
		}
		currentClass := document.GetClassScopeAtSymbol(v.Scope)
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			methods := analysis.EmptyInheritedMethods()
			for _, class := range q.GetClasses(scopeType.GetFQN()) {
				methods.Merge(q.GetClassMethods(class, v.Name, methods.SearchedFQNs))
			}
			for _, m := range methods.ReduceStatic(currentClass, scopeName) {
				locations = append(locations, m.Method.GetLocation())
			}
		}
	case *analysis.ScopedPropertyAccess:
		var scopeName string
		if hasName, ok := v.Scope.(analysis.HasName); ok {
			scopeName = hasName.GetName()
		}
		currentClass := document.GetClassScopeAtSymbol(v.Scope)
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			properties := analysis.EmptyInheritedProps()
			for _, class := range q.GetClasses(scopeType.GetFQN()) {
				properties.Merge(q.GetClassProps(class, v.Name, properties.SearchedFQNs))
			}
			for _, ps := range properties.ReduceStatic(currentClass, scopeName) {
				locations = append(locations, ps.Prop.GetLocation())
			}
		}
	case *analysis.PropertyAccess:
		var scopeName string
		if n, ok := v.Scope.(analysis.HasName); ok {
			scopeName = n.GetName()
		}
		currentClass := document.GetClassScopeAtSymbol(v.Scope)
		types := v.ResolveAndGetScope(resolveCtx)
		properties := analysis.EmptyInheritedProps()
		for _, scopeType := range types.Resolve() {
			for _, class := range q.GetClasses(scopeType.GetFQN()) {
				properties.Merge(q.GetClassProps(class, "$"+v.Name, properties.SearchedFQNs))
			}
		}
		for _, ps := range properties.ReduceAccess(currentClass, scopeName, types) {
			locations = append(locations, ps.Prop.GetLocation())
		}
	case *analysis.MethodAccess:
		var scopeName string
		if n, ok := v.Scope.(analysis.HasName); ok {
			scopeName = n.GetName()
		}
		currentClass := document.GetClassScopeAtSymbol(v.Scope)
		types := v.ResolveAndGetScope(resolveCtx)
		methods := analysis.EmptyInheritedMethods()
		for _, scopeType := range types.Resolve() {
			for _, class := range q.GetClasses(scopeType.GetFQN()) {
				methods.Merge(q.GetClassMethods(class, v.Name, methods.SearchedFQNs))
			}
			for _, theInterface := range q.GetInterfaces(scopeType.GetFQN()) {
				methods.Merge(q.GetInterfaceMethods(theInterface, v.Name, methods.SearchedFQNs))
			}
			for _, trait := range q.GetTraits(scopeType.GetFQN()) {
				methods.Merge(q.GetTraitMethods(trait, v.Name))
			}
		}
		for _, m := range methods.ReduceAccess(currentClass, scopeName, types) {
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
