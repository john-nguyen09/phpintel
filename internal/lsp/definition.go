package lsp

import (
	"context"

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
	symbol := document.SymbolAtPos(params.TextDocumentPositionParams.Position)

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
	case *analysis.ConstantAccess:
		for _, typeString := range v.Type.Resolve() {
			for _, theConst := range store.GetConsts(document.GetImportTable().GetConstReferenceFQN(typeString)) {
				locations = append(locations, theConst.GetLocation())
			}
			for _, define := range store.GetDefines(document.GetImportTable().GetConstReferenceFQN(typeString)) {
				locations = append(locations, define.GetLocation())
			}
		}
	case *analysis.FunctionCall:
		name := analysis.NewTypeString(v.Name)
		for _, function := range store.GetFunctions(document.GetImportTable().GetFunctionReferenceFQN(name)) {
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
			for _, method := range store.GetMethods(scopeType.GetFQN(), v.Name) {
				if !method.IsStatic {
					continue
				}
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
			for _, method := range store.GetMethods(scopeType.GetFQN(), v.Name) {
				locations = append(locations, method.GetLocation())
			}
		}
	}
	return locations, nil
}
