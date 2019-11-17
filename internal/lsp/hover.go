package lsp

import (
	"context"

	"github.com/john-nguyen09/phpintel/analysis"
	"github.com/john-nguyen09/phpintel/internal/cmd"
	"github.com/john-nguyen09/phpintel/internal/lsp/protocol"
)

func (s *Server) hover(ctx context.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
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
	symbol := document.SymbolAtPos(params.TextDocumentPositionParams.Position)
	var hover *protocol.Hover = nil
	// log.Printf("%T %s\n", symbol, symbol.GetLocation())

	switch v := symbol.(type) {
	case *analysis.ClassTypeDesignator:
		classes := []*analysis.Class{}
		for _, typeString := range v.Type.Resolve() {
			classes = append(classes, store.GetClasses(typeString.GetFQN())...)
			if len(classes) > 0 {
				hover = cmd.ClassToHover(symbol, *classes[0])
				break
			}
		}
	case *analysis.ClassAccess:
		classes := []*analysis.Class{}
		for _, typeString := range v.Type.Resolve() {
			classes = append(classes, store.GetClasses(typeString.GetFQN())...)
			if len(classes) > 0 {
				hover = cmd.ClassToHover(symbol, *classes[0])
				break
			}
		}
	case *analysis.ConstantAccess:
		consts := []*analysis.Const{}
		defines := []*analysis.Define{}
		for _, typeString := range v.Type.Resolve() {
			consts = append(consts, store.GetConsts(typeString.GetFQN())...)
			if len(consts) > 0 {
				hover = cmd.ConstToHover(symbol, *consts[0])
				break
			}
			defines = append(defines, store.GetDefines(typeString.GetFQN())...)
			if len(defines) > 0 {
				hover = cmd.DefineToHover(symbol, *defines[0])
				break
			}
		}
	case *analysis.FunctionCall:
		functions := []*analysis.Function{}
		for _, typeString := range v.Type.Resolve() {
			functions = append(functions, store.GetFunctions(typeString.GetFQN())...)
			if len(functions) > 0 {
				hover = cmd.FunctionToHover(symbol, *functions[0])
				break
			}
		}
	case *analysis.ScopedConstantAccess:
		classConsts := []*analysis.ClassConst{}
		for _, typeString := range v.Type.Resolve() {
			for _, scopeType := range v.Scope.GetTypes().Resolve() {
				classConsts = append(classConsts, store.GetClassConsts(scopeType.GetFQN(), typeString.GetFQN())...)
				if len(classConsts) > 0 {
					hover = cmd.ClassConstToHover(symbol, *classConsts[0])
					break
				}
			}
		}
	case *analysis.ScopedMethodAccess:
		methods := []*analysis.Method{}
		for _, typeString := range v.Type.Resolve() {
			for _, scopeType := range v.Scope.GetTypes().Resolve() {
				methods = append(methods, store.GetMethods(scopeType.GetFQN(), typeString.GetFQN())...)
				if len(methods) > 0 {
					hover = cmd.MethodToHover(symbol, *methods[0])
					break
				}
			}
		}
	case *analysis.ScopedPropertyAccess:
		properties := []*analysis.Property{}
		for _, typeString := range v.Type.Resolve() {
			for _, scopeType := range v.Scope.GetTypes().Resolve() {
				properties = append(properties, store.GetProperties(scopeType.GetFQN(), typeString.GetFQN())...)
				if len(properties) > 0 {
					hover = cmd.PropertyToHover(symbol, *properties[0])
					break
				}
			}
		}
	case *analysis.Variable:
		hover = cmd.VariableToHover(v)
	}
	return hover, nil
}
