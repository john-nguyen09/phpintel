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
	document.Load()
	symbol := document.HasTypesAtPos(params.TextDocumentPositionParams.Position)

	switch v := symbol.(type) {
	case *analysis.ClassTypeDesignator:
		for _, typeString := range v.Type.Resolve() {
			for _, theClass := range store.GetClasses(document.GetImportTable().GetClassReferenceFQN(typeString)) {
				locations = append(locations, theClass.GetLocation())
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
		for _, typeString := range v.Type.Resolve() {
			for _, theConst := range store.GetConsts(document.GetImportTable().GetConstReferenceFQN(store, typeString)) {
				locations = append(locations, theConst.GetLocation())
			}
			for _, define := range store.GetDefines(document.GetImportTable().GetConstReferenceFQN(store, typeString)) {
				locations = append(locations, define.GetLocation())
			}
		}
	case *analysis.FunctionCall:
		name := analysis.NewTypeString(v.Name)
		for _, function := range store.GetFunctions(document.GetImportTable().GetFunctionReferenceFQN(store, name)) {
			locations = append(locations, function.GetLocation())
		}
	case *analysis.ScopedConstantAccess:
		for _, scopeType := range v.ResolveAndGetScope(store).Resolve() {
			for _, classConst := range store.GetClassConsts(scopeType.GetFQN(), v.Name) {
				locations = append(locations, classConst.GetLocation())
			}
		}
	case *analysis.ScopedMethodAccess:
		for _, scopeType := range v.ResolveAndGetScope(store).Resolve() {
			methods := []*analysis.Method{}
			for _, class := range store.GetClasses(scopeType.GetFQN()) {
				methods = append(methods, analysis.GetClassMethods(store, class, v.Name,
					analysis.NewSearchOptions().
						WithPredicate(func(symbol analysis.Symbol) bool {
							method := symbol.(*analysis.Method)
							if !method.IsStatic {
								return false
							}
							return true
						}))...)
			}
			for _, method := range methods {
				locations = append(locations, method.GetLocation())
			}
		}
	case *analysis.ScopedPropertyAccess:
		for _, scopeType := range v.ResolveAndGetScope(store).Resolve() {
			for _, property := range store.GetProperties(scopeType.GetFQN(), v.Name) {
				if !property.IsStatic {
					continue
				}
				locations = append(locations, property.GetLocation())
			}
		}
	case *analysis.PropertyAccess:
		for _, scopeType := range v.ResolveAndGetScope(store).Resolve() {
			for _, property := range store.GetProperties(scopeType.GetFQN(), "$"+v.Name) {
				locations = append(locations, property.GetLocation())
			}
		}
	case *analysis.MethodAccess:
		for _, scopeType := range v.ResolveAndGetScope(store).Resolve() {
			methods := []*analysis.Method{}
			for _, class := range store.GetClasses(scopeType.GetFQN()) {
				methods = append(methods, analysis.GetClassMethods(store, class, v.Name,
					analysis.NewSearchOptions())...)
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
