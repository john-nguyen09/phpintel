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
	resolveCtx := analysis.NewResolveContext(store, document)
	pos := params.TextDocumentPositionParams.Position
	symbol := document.HasTypesAtPos(pos)

	switch v := symbol.(type) {
	case *analysis.ClassTypeDesignator:
		for _, typeString := range v.Type.Resolve() {
			for _, theClass := range store.GetClasses(document.GetImportTable().GetClassReferenceFQN(typeString)) {
				constructor := theClass.GetConstructor(store)
				if constructor == nil || constructor.GetScope().GetFQN() != theClass.Name.GetFQN() {
					locations = append(locations, theClass.GetLocation())
				} else {
					locations = append(locations, constructor.GetLocation())
				}
			}
		}
	case *analysis.ClassAccess:
		for _, typeString := range v.Type.Resolve() {
			for _, theClass := range store.GetClasses(document.GetImportTable().GetClassReferenceFQN(typeString)) {
				locations = append(locations, theClass.GetLocation())
			}
		}
	case *analysis.InterfaceAccess:
		for _, typeString := range v.Type.Resolve() {
			for _, theInterface := range store.GetInterfaces(document.GetImportTable().GetClassReferenceFQN(typeString)) {
				locations = append(locations, theInterface.GetLocation())
			}
		}
	case *analysis.TraitAccess:
		for _, typeString := range v.GetTypes().Resolve() {
			for _, trait := range store.GetTraits(typeString.GetFQN()) {
				locations = append(locations, trait.GetLocation())
			}
		}
	case *analysis.ConstantAccess:
		name := analysis.NewTypeString(v.Name)
		for _, theConst := range store.GetConsts(document.GetImportTable().GetConstReferenceFQN(store, name)) {
			locations = append(locations, theConst.GetLocation())
		}
		for _, define := range store.GetDefines(document.GetImportTable().GetConstReferenceFQN(store, name)) {
			locations = append(locations, define.GetLocation())
		}
	case *analysis.FunctionCall:
		name := analysis.NewTypeString(v.Name)
		for _, function := range store.GetFunctions(document.GetImportTable().GetFunctionReferenceFQN(store, name)) {
			locations = append(locations, function.GetLocation())
		}
	case *analysis.ScopedConstantAccess:
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			for _, classConst := range store.GetClassConsts(scopeType.GetFQN(), v.Name) {
				locations = append(locations, classConst.GetLocation())
			}
		}
	case *analysis.ScopedMethodAccess:
		name := ""
		classScope := ""
		if hasName, ok := v.Scope.(analysis.HasName); ok {
			name = hasName.GetName()
		}
		if hasScope, ok := v.Scope.(analysis.HasScope); ok {
			classScope = hasScope.GetScope().GetFQN()
		}
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			methods := []*analysis.Method{}
			for _, class := range store.GetClasses(scopeType.GetFQN()) {
				methods = append(methods, analysis.GetClassMethods(store, class, v.Name,
					analysis.StaticMethodsScopeAware(analysis.NewSearchOptions(), classScope, name))...)
			}
			for _, method := range methods {
				locations = append(locations, method.GetLocation())
			}
		}
	case *analysis.ScopedPropertyAccess:
		name := ""
		classScope := ""
		if hasName, ok := v.Scope.(analysis.HasName); ok {
			name = hasName.GetName()
		}
		if hasScope, ok := v.Scope.(analysis.HasScope); ok {
			classScope = hasScope.GetScope().GetFQN()
		}
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			properties := []*analysis.Property{}
			for _, class := range store.GetClasses(scopeType.GetFQN()) {
				properties = append(properties, analysis.GetClassProperties(store, class, v.Name,
					analysis.StaticPropsScopeAware(analysis.NewSearchOptions(), classScope, name))...)
			}
			for _, property := range properties {
				if !property.IsStatic {
					continue
				}
				locations = append(locations, property.GetLocation())
			}
		}
	case *analysis.PropertyAccess:
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			for _, class := range store.GetClasses(scopeType.GetFQN()) {
				for _, property := range analysis.GetClassProperties(store, class, "$"+v.Name,
					analysis.PropsScopeAware(analysis.NewSearchOptions(), document, v.Scope)) {
					locations = append(locations, property.GetLocation())
				}
			}
		}
	case *analysis.MethodAccess:
		for _, scopeType := range v.ResolveAndGetScope(resolveCtx).Resolve() {
			methods := []*analysis.Method{}
			for _, class := range store.GetClasses(scopeType.GetFQN()) {
				methods = append(methods, analysis.GetClassMethods(store, class, v.Name,
					analysis.MethodsScopeAware(analysis.NewSearchOptions(), document, v.Scope))...)
			}
			for _, method := range methods {
				locations = append(locations, method.GetLocation())
			}
		}
	case *analysis.TypeDeclaration:
		for _, typeString := range v.Type.Resolve() {
			classes := store.GetClasses(typeString.GetFQN())
			for _, class := range classes {
				locations = append(locations, class.GetLocation())
			}
			interfaces := store.GetInterfaces(typeString.GetFQN())
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
