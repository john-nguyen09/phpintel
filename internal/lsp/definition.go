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
			for _, theClass := range store.GetClasses(typeString.GetFQN()) {
				locations = append(locations, theClass.GetLocation())
			}
		}
	case *analysis.ClassAccess:
		for _, typeString := range v.Type.Resolve() {
			for _, theClass := range store.GetClasses(typeString.GetFQN()) {
				locations = append(locations, theClass.GetLocation())
			}
		}
	case *analysis.ConstantAccess:
		for _, typeString := range v.Type.Resolve() {
			for _, theConst := range store.GetConsts(typeString.GetFQN()) {
				locations = append(locations, theConst.GetLocation())
			}
			for _, define := range store.GetDefines(typeString.GetFQN()) {
				locations = append(locations, define.GetLocation())
			}
		}
	case *analysis.FunctionCall:
		for _, function := range store.GetFunctions(v.Name) {
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
